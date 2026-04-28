package delivery

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cleanair/passenger/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthHandler wires HTTP routes for register/login/me.
type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type registerStaffReq struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	EmployeeID string `json:"employee_id"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  *auth.User `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	u, token, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже зарегистрирован"})
		case errors.Is(err, auth.ErrEmailFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный email"})
		case errors.Is(err, auth.ErrPasswordPolicy):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен быть не меньше 8 символов и содержать буквы и цифры"})
		case errors.Is(err, auth.ErrFullNameLength):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введите ФИО (минимум имя и фамилия)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, authResponse{Token: token, User: u})
}

func (h *AuthHandler) RegisterStaff(c *gin.Context) {
	var req registerStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	u, token, err := h.svc.RegisterStaff(c.Request.Context(), req.Email, req.Password, req.FullName, req.EmployeeID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже зарегистрирован"})
		case errors.Is(err, auth.ErrEmployeeIDTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "Сотрудник с таким табельным номером уже зарегистрирован"})
		case errors.Is(err, auth.ErrEmployeeIDEmpty):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите табельный номер сотрудника"})
		case errors.Is(err, auth.ErrEmailFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный email"})
		case errors.Is(err, auth.ErrPasswordPolicy):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен быть не меньше 8 символов и содержать буквы и цифры"})
		case errors.Is(err, auth.ErrFullNameLength):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Введите ФИО (минимум имя и фамилия)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, authResponse{Token: token, User: u})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	u, token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}
	c.JSON(http.StatusOK, authResponse{Token: token, User: u})
}

// Me returns the current user by decoding the Authorization header token.
func (h *AuthHandler) Me(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token"})
		return
	}
	uid, _, err := h.svc.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return h
}
