// Package auth implements a minimal authentication layer on top of the
// passenger service: users table, bcrypt-hashed passwords and HMAC-signed
// stateless tokens carrying user id + role + expiration.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cleanair/passenger/internal/usecase"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

const (
	RolePassenger = "passenger"
	RoleAdmin     = "admin"
)

// User is the authentication record linked to a passenger profile.
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         string     `json:"role" db:"role"`
	PassengerID  *uuid.UUID `json:"passenger_id,omitempty" db:"passenger_id"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Repository is the backing store for users.
type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

// Migrate creates the users table if it does not exist.
func (r *Repository) Migrate(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL DEFAULT 'passenger',
		passenger_id UUID REFERENCES passengers(id) ON DELETE SET NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	_, err := r.db.ExecContext(ctx, schema)
	return err
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	q := `INSERT INTO users (id, email, full_name, role, passenger_id, password_hash, created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.ExecContext(ctx, q, u.ID, strings.ToLower(u.Email), u.FullName, u.Role, u.PassengerID, u.PasswordHash, u.CreatedAt)
	return err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	q := `SELECT id, email, full_name, role, passenger_id, password_hash, created_at
	      FROM users WHERE email=$1`
	err := r.db.GetContext(ctx, &u, q, strings.ToLower(email))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	q := `SELECT id, email, full_name, role, passenger_id, password_hash, created_at
	      FROM users WHERE id=$1`
	err := r.db.GetContext(ctx, &u, q, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Service wraps business logic around Repository + bcrypt + token signing.
type Service struct {
	users       *Repository
	passengers  usecase.PassengerRepository
	tokenSecret []byte
	ttl         time.Duration
}

func NewService(users *Repository, passengers usecase.PassengerRepository, secret []byte, ttl time.Duration) *Service {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &Service{users: users, passengers: passengers, tokenSecret: secret, ttl: ttl}
}

// Errors returned by Service.
var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrInvalidCreds   = errors.New("invalid email or password")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token expired")
	ErrEmailFormat    = errors.New("email is not valid")
	ErrPasswordLength = errors.New("password must be at least 6 characters")
	ErrFullNameLength = errors.New("full name must have at least 2 words")
)

func validateCredentials(email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrEmailFormat
	}
	if len(password) < 6 {
		return ErrPasswordLength
	}
	return nil
}

// Register creates a new user. For passenger role a linked passenger
// record is also created (name/email carried over).
func (s *Service) Register(ctx context.Context, email, password, fullName string) (*User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	if err := validateCredentials(email, password); err != nil {
		return nil, "", err
	}
	if len(strings.Fields(fullName)) < 2 {
		return nil, "", ErrFullNameLength
	}
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil, "", ErrEmailTaken
	}

	// Hash password before creating any records, so cheap failures don't
	// leave orphaned rows.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	pax := &usecase.Passenger{
		ID:             uuid.New(),
		Name:           fullName,
		Email:          email,
		Phone:          "",
		PassportNumber: "",
		LoyaltyTier:    "STANDARD",
		MealPreference: "STANDARD",
	}
	if err := s.passengers.Create(ctx, pax); err != nil {
		return nil, "", err
	}
	paxID := pax.ID
	u := &User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		Role:         RolePassenger,
		PassengerID:  &paxID,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, u); err != nil {
		// Roll back the passenger we just created so we don't leave an
		// orphaned profile when, for example, a concurrent registration
		// claims the same email between the GetByEmail check and this insert.
		_ = s.passengers.Delete(context.Background(), pax.ID)
		return nil, "", err
	}
	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// Login verifies email+password and issues a token.
func (s *Service) Login(ctx context.Context, email, password string) (*User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCreds
	}
	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// EnsureAdmin creates the default admin user on startup if missing.
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	email = strings.ToLower(email)
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     "Администратор",
		Role:         RoleAdmin,
		PassengerID:  nil,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	return s.users.Create(ctx, u)
}

// GetByID returns the user by id.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.users.GetByID(ctx, id)
}

// tokenPayload is the body of our signed token.
type tokenPayload struct {
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
	Exp    int64     `json:"exp"`
}

func (s *Service) issueToken(u *User) (string, error) {
	p := tokenPayload{UserID: u.ID, Role: u.Role, Exp: time.Now().Add(s.ttl).Unix()}
	body, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding.EncodeToString(body)
	sig := sign(enc, s.tokenSecret)
	return enc + "." + sig, nil
}

// ParseToken validates signature/expiry and returns the embedded user id+role.
func (s *Service) ParseToken(token string) (uuid.UUID, string, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return uuid.Nil, "", ErrInvalidToken
	}
	expected := sign(parts[0], s.tokenSecret)
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return uuid.Nil, "", ErrInvalidToken
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	var p tokenPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	if time.Now().Unix() > p.Exp {
		return uuid.Nil, "", ErrExpiredToken
	}
	return p.UserID, p.Role, nil
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
