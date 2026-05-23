package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"project-manager/src/middleware"
	"project-manager/src/models"
	"project-manager/src/utils"

	"golang.org/x/crypto/bcrypt"
)

type cleanUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	Search(ctx context.Context, query string) ([]models.User, error)
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

	_, err := c.repo.FindByEmail(r.Context(), req.Email)
	if err == nil {
		utils.WriteError(w, http.StatusConflict, "Email already registered")
		return
	} else if !errors.Is(err, models.ErrUserNotFound) {
		utils.WriteError(w, http.StatusInternalServerError, "Database query failed")
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	newUser := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := c.repo.Create(r.Context(), &newUser); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to save user")
		return
	}

	userClean := cleanUser{
		ID:    newUser.ID,
		Name:  newUser.Name,
		Email: newUser.Email,
	}

	tokenString, err := utils.GenerateToken(newUser.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to generate session token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24,
	})

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

	user, err := c.repo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to authenticate")
		return
	}

	if !checkPasswordHash(req.Password, user.Password) {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	userClean := cleanUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	tokenString, err := utils.GenerateToken(userClean.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600 * 24,
	})

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"token": tokenString,
		"user":  userClean,
	})
}

// retrieve details of the currently authenticated user
func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok || userID == 0 {
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := c.repo.FindByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			utils.WriteError(w, http.StatusUnauthorized, "User profile not found")
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve user profile")
		return
	}

	userClean := cleanUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	utils.WriteJSON(w, http.StatusOK, userClean)
}

// clear the authentication cookie and log out the user
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // deletes the cookie immediately
	})

	utils.WriteMessage(w, http.StatusOK, "Successfully logged out")
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

	users, err := c.repo.Search(r.Context(), query)
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

// hash a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// check if a password matches a hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
