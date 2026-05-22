package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"project-manager/src/models"
	"project-manager/src/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type cleanUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepositoryInterface interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	Search(query string) ([]models.User, error)
}

type AuthController struct {
	repo UserRepositoryInterface
}

func NewAuthController(repo UserRepositoryInterface) *AuthController {
	return &AuthController{repo: repo}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		utils.WriteError(w, http.StatusUnprocessableEntity, "Missing required fields")
		return
	}

	existingUser, err := c.repo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		utils.WriteError(w, http.StatusConflict, "Email already registered")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	newUser := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := c.repo.Create(&newUser); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to save user")
		return
	}

	userClean := cleanUser{
		ID:    newUser.ID,
		Name:  newUser.Name,
		Email: newUser.Email,
	}

	utils.WriteJSON(w, http.StatusCreated, userClean)
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := c.repo.FindByEmail(req.Email)
	if err != nil || user == nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	userClean := cleanUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userClean.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"token": tokenString,
		"user":  userClean,
	})
}

// search registered users matching a query (minimum 3 characters)
func (c *AuthController) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		utils.WriteError(w, http.StatusBadRequest, "Search query 'q' must be at least 3 characters long")
		return
	}

	users, err := c.repo.Search(query)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	cleanUsers := make([]cleanUser, 0, len(users))
	for _, u := range users {
		cleanUsers = append(cleanUsers, cleanUser{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}

	utils.WriteJSON(w, http.StatusOK, cleanUsers)
}
