package controllers

import (
	"encoding/json"
	"net/http"
	"project-manager/src/middleware"
	"project-manager/src/models"
	"strconv"
	"strings"
)

type ProjectRepoInterface interface {
	Create(project *models.Project) error
	FindByID(id uint) (*models.Project, error)
	FindAllForUser(userID uint) ([]models.Project, error)
	Update(project *models.Project) error
	Delete(id uint) error
	AddParticipantByEmail(projectID uint, email string) error
}

type ProjectController struct {
	repo ProjectRepoInterface
}

func NewProjectController(repo ProjectRepoInterface) *ProjectController {
	return &ProjectController{repo: repo}
}

// clean up trailing slashes and handle explicit sub-routing
func (c *ProjectController) Dispatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, _ := r.Context().Value(middleware.UserIDKey).(uint)

	// clean up path: /api/projects/1/participants -> ["api", "projects", "1", "participants"]
	cleanPath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(cleanPath, "/")

	if parts[1] != "projects" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Endpoint not found"}`))
		return
	}

	if len(parts) == 2 {
		if r.Method == http.MethodGet {
			c.listProjects(w, userID)
			return
		}
		if r.Method == http.MethodPost {
			c.createProject(w, r, userID)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if len(parts) >= 3 {
		idVal, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid project identifier"}`))
			return
		}
		projectID := uint(idVal)

		// POST /api/projects/{id}/participants
		if len(parts) == 4 && parts[3] == "participants" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			c.addParticipant(w, r, projectID, userID)
			return
		}

		if r.Method == http.MethodPut {
			c.updateProject(w, r, projectID, userID)
			return
		}
		if r.Method == http.MethodDelete {
			c.deleteProject(w, projectID, userID)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (c *ProjectController) listProjects(w http.ResponseWriter, userID uint) {
	projects, err := c.repo.FindAllForUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to retrieve workspace items"}`))
		return
	}
	json.NewEncoder(w).Encode(projects)
}

func (c *ProjectController) createProject(w http.ResponseWriter, r *http.Request, userID uint) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid payload. Title is required"}`))
		return
	}

	newProject := models.Project{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}

	if err := c.repo.Create(&newProject); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to execute database record creation"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProject)
}

func (c *ProjectController) updateProject(w http.ResponseWriter, r *http.Request, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Project workspace target not found"}`))
		return
	}

	if project.OwnerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid modification updates"}`))
		return
	}

	if req.Title != "" {
		project.Title = req.Title
	}
	project.Description = req.Description

	if err := c.repo.Update(project); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to commit modifications"}`))
		return
	}
	json.NewEncoder(w).Encode(project)
}

func (c *ProjectController) deleteProject(w http.ResponseWriter, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Project workspace target not found"}`))
		return
	}

	if project.OwnerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	if err := c.repo.Delete(projectID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to remove item"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Project workspace wiped successfully"}`))
}

func (c *ProjectController) addParticipant(w http.ResponseWriter, r *http.Request, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Project workspace target not found"}`))
		return
	}

	if project.OwnerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Email parameter required"}`))
		return
	}

	if err := c.repo.AddParticipantByEmail(projectID, req.Email); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Target user email not registered inside application"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Participant successfully attached to project"}`))
}
