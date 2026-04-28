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
	RoleStaff     = "staff"
	RoleAdmin     = "admin"
)

// User is the authentication record linked to either a passenger profile
// (RolePassenger) or a staff record (RoleStaff). Admins have neither.
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         string     `json:"role" db:"role"`
	PassengerID  *uuid.UUID `json:"passenger_id,omitempty" db:"passenger_id"`
	EmployeeID   string     `json:"employee_id,omitempty" db:"employee_id"`
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
		employee_id VARCHAR(40) DEFAULT '',
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return err
	}
	// Backwards-compatible migration for installations created before the
	// staff role was introduced. Must run BEFORE the partial unique index.
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS employee_id VARCHAR(40) DEFAULT '';`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS users_employee_id_unique ON users(employee_id) WHERE employee_id <> '';`); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	q := `INSERT INTO users (id, email, full_name, role, passenger_id, employee_id, password_hash, created_at)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, q, u.ID, strings.ToLower(u.Email), u.FullName, u.Role, u.PassengerID, u.EmployeeID, u.PasswordHash, u.CreatedAt)
	return err
}

const userColumns = `id, email, full_name, role, passenger_id,
	COALESCE(employee_id, '') AS employee_id,
	password_hash, created_at`

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	q := `SELECT ` + userColumns + ` FROM users WHERE email=$1`
	err := r.db.GetContext(ctx, &u, q, strings.ToLower(email))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	q := `SELECT ` + userColumns + ` FROM users WHERE id=$1`
	err := r.db.GetContext(ctx, &u, q, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByEmployeeID(ctx context.Context, employeeID string) (*User, error) {
	var u User
	q := `SELECT ` + userColumns + ` FROM users WHERE employee_id=$1`
	err := r.db.GetContext(ctx, &u, q, strings.TrimSpace(employeeID))
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
	ErrEmailTaken      = errors.New("email already registered")
	ErrEmployeeIDTaken = errors.New("employee id already registered")
	ErrEmployeeIDEmpty = errors.New("employee id is required")
	ErrInvalidCreds    = errors.New("invalid email or password")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("token expired")
	ErrEmailFormat     = errors.New("email is not valid")
	ErrPasswordPolicy  = errors.New("password does not satisfy policy")
	ErrFullNameLength  = errors.New("full name must have at least 2 words")
)

// validatePassword enforces the system-wide password policy:
//   - minimum 8 characters
//   - at least one letter
//   - at least one digit
//
// Symbols are allowed but not required, to keep the demo accessible.
func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordPolicy
	}
	hasLetter := false
	hasDigit := false
	for _, r := range password {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		case r >= 0x0400 && r <= 0x04FF: // Cyrillic block
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrPasswordPolicy
	}
	return nil
}

func validateCredentials(email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrEmailFormat
	}
	return validatePassword(password)
}

// validateAdminPassword skips the policy because the default admin/admin
// account is provisioned at boot for the demo. Real admin passwords still
// have to satisfy validatePassword when set via Register.
func validateAdminBootstrap(password string) bool {
	return password == "admin"
}

// RegisterStaff creates a new staff user with a unique employee_id. The
// employee_id is intended to be a corporate identifier (in production it
// would be cross-checked against an HR registry; for the demo any unique
// non-empty value is accepted).
func (s *Service) RegisterStaff(ctx context.Context, email, password, fullName, employeeID string) (*User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	employeeID = strings.TrimSpace(employeeID)
	if employeeID == "" {
		return nil, "", ErrEmployeeIDEmpty
	}
	if err := validateCredentials(email, password); err != nil {
		return nil, "", err
	}
	if len(strings.Fields(fullName)) < 2 {
		return nil, "", ErrFullNameLength
	}
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil, "", ErrEmailTaken
	}
	if existing, _ := s.users.GetByEmployeeID(ctx, employeeID); existing != nil {
		return nil, "", ErrEmployeeIDTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	u := &User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		Role:         RoleStaff,
		EmployeeID:   employeeID,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err
	}
	token, err := s.issueToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
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

// EnsureAdmin creates the default admin user on startup if missing. The
// boot-time admin/admin password bypasses the password policy on purpose;
// the admin user can change it via a future password-reset flow.
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	email = strings.ToLower(email)
	if existing, _ := s.users.GetByEmail(ctx, email); existing != nil {
		return nil
	}
	if !validateAdminBootstrap(password) {
		if err := validatePassword(password); err != nil {
			return err
		}
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
