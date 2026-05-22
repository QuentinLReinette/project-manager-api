package controllers

import (
	"encoding/json"
	"net/http"
	"project-manager/src/middleware"
	"project-manager/src/models"
	"project-manager/src/utils"
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
	userID, _ := r.Context().Value(middleware.UserIDKey).(uint)

	// clean up path: /api/projects/1/participants -> ["api", "projects", "1", "participants"]
	cleanPath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(cleanPath, "/")

	if len(parts) < 2 || parts[1] != "projects" {
		utils.WriteError(w, http.StatusNotFound, "Endpoint not found")
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
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if len(parts) >= 3 {
		idVal, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid project identifier")
			return
		}
		projectID := uint(idVal)

		// POST /api/projects/{id}/participants
		if len(parts) == 4 && parts[3] == "participants" {
			if r.Method != http.MethodPost {
				utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
				return
			}
			c.addParticipant(w, r, projectID, userID)
			return
		}

		if r.Method == http.MethodGet {
			c.getProject(w, projectID, userID)
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
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
}

func (c *ProjectController) listProjects(w http.ResponseWriter, userID uint) {
	projects, err := c.repo.FindAllForUser(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to retrieve workspace items")
		return
	}
	utils.WriteJSON(w, http.StatusOK, projects)
}

func (c *ProjectController) createProject(w http.ResponseWriter, r *http.Request, userID uint) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "Invalid payload. Title is required")
		return
	}

	newProject := models.Project{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     userID,
	}

	if err := c.repo.Create(&newProject); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to execute database record creation")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, newProject)
}

func (c *ProjectController) updateProject(w http.ResponseWriter, r *http.Request, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Project workspace target not found")
		return
	}

	if project.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid modification updates")
		return
	}

	if req.Title != "" {
		project.Title = req.Title
	}
	project.Description = req.Description

	if err := c.repo.Update(project); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to commit modifications")
		return
	}
	utils.WriteJSON(w, http.StatusOK, project)
}

func (c *ProjectController) deleteProject(w http.ResponseWriter, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Project workspace target not found")
		return
	}

	if project.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := c.repo.Delete(projectID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to remove item")
		return
	}
	utils.WriteMessage(w, http.StatusOK, "Project workspace wiped successfully")
}

func (c *ProjectController) addParticipant(w http.ResponseWriter, r *http.Request, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Project workspace target not found")
		return
	}

	if project.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		utils.WriteError(w, http.StatusBadRequest, "Email parameter required")
		return
	}

	if err := c.repo.AddParticipantByEmail(projectID, req.Email); err != nil {
		utils.WriteError(w, http.StatusNotFound, "Target user email not registered inside application")
		return
	}

	utils.WriteMessage(w, http.StatusOK, "Participant successfully attached to project")
}

func (c *ProjectController) getProject(w http.ResponseWriter, projectID uint, userID uint) {
	project, err := c.repo.FindByID(projectID)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, "Project workspace target not found")
		return
	}

	isMember := project.OwnerID == userID
	for _, p := range project.Participants {
		if p.ID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		utils.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	utils.WriteJSON(w, http.StatusOK, project)
}
