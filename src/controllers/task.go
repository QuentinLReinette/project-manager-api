package controllers

import (
	"encoding/json"
	"net/http"
	"project-manager/src/middleware"
	"project-manager/src/models"
	"strconv"
	"strings"
)

type TaskRepoInterface interface {
	Create(task *models.Task) error
	FindByID(id uint) (*models.Task, error)
	FindByProjectID(projectID uint, statusFilter models.TaskStatus) ([]models.Task, error)
	Update(task *models.Task) error
	Delete(id uint) error
	IsUserMember(projectID uint, userID uint) (bool, error)
}

type TaskController struct {
	repo TaskRepoInterface
}

func NewTaskController(repo TaskRepoInterface) *TaskController {
	return &TaskController{repo: repo}
}

func (c *TaskController) Dispatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, _ := r.Context().Value(middleware.UserIDKey).(uint)

	cleanPath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(cleanPath, "/")

	if len(parts) < 2 || parts[1] != "tasks" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// GET /api/tasks (requires ?project_id=X) or POST /api/tasks
	if len(parts) == 2 && parts[1] == "tasks" {
		if r.Method == http.MethodGet {
			c.listTasks(w, r, userID)
			return
		}
		if r.Method == http.MethodPost {
			c.createTask(w, r, userID)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// GET, PUT or DELETE /api/tasks/{id}
	if len(parts) == 3 && parts[1] == "tasks" {
		idVal, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid task identifier"}`))
			return
		}
		taskID := uint(idVal)

		if r.Method == http.MethodGet {
			c.getTask(w, taskID, userID)
			return
		}
		if r.Method == http.MethodPut {
			c.updateTask(w, r, taskID, userID)
			return
		}
		if r.Method == http.MethodDelete {
			c.deleteTask(w, taskID, userID)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (c *TaskController) listTasks(w http.ResponseWriter, r *http.Request, userID uint) {
	pIDStr := r.URL.Query().Get("project_id")
	statusFilter := models.TaskStatus(r.URL.Query().Get("status")) // optional filter: todo, in_progress, done

	if pIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Missing required project_id query parameter"}`))
		return
	}

	pID, _ := strconv.ParseUint(pIDStr, 10, 32)
	projectID := uint(pID)

	isMember, err := c.repo.IsUserMember(projectID, userID)
	if err != nil || !isMember {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied to this project workspace"}`))
		return
	}

	if statusFilter != "" && !statusFilter.IsValid() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid status query filter. Must be 'todo', 'in_progress', or 'done'"}`))
		return
	}

	tasks, err := c.repo.FindByProjectID(projectID, statusFilter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to fetch tasks"}`))
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func (c *TaskController) createTask(w http.ResponseWriter, r *http.Request, userID uint) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   uint   `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.ProjectID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid payload. Title and project_id are required"}`))
		return
	}

	isMember, err := c.repo.IsUserMember(req.ProjectID, userID)
	if err != nil || !isMember {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied to this project workspace"}`))
		return
	}

	newTask := models.Task{
		Title:       req.Title,
		Description: req.Description,
		ProjectID:   req.ProjectID,
		Status:      models.StatusTodo, // default fallback state
	}

	if err := c.repo.Create(&newTask); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to create task"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func (c *TaskController) updateTask(w http.ResponseWriter, r *http.Request, taskID uint, userID uint) {
	task, err := c.repo.FindByID(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Task not found"}`))
		return
	}

	isMember, err := c.repo.IsUserMember(task.ProjectID, userID)
	if err != nil || !isMember {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	var req struct {
		Title        string            `json:"title"`
		Description  string            `json:"description"`
		Status       models.TaskStatus `json:"status"`
		AssignedToID *uint             `json:"assigned_to_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request payload"}`))
		return
	}

	if req.Status != "" {
		if !req.Status.IsValid() {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error": "Status must be 'todo', 'in_progress', or 'done'"}`))
			return
		}
		task.Status = req.Status
	}

	if req.AssignedToID != nil {
		isAssigneeValid, _ := c.repo.IsUserMember(task.ProjectID, *req.AssignedToID)
		if !isAssigneeValid {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error": "Assignee must be an active participant of the project"}`))
			return
		}
		task.AssignedToID = req.AssignedToID
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	task.Description = req.Description

	if err := c.repo.Update(task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to update task"}`))
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (c *TaskController) deleteTask(w http.ResponseWriter, taskID uint, userID uint) {
	task, err := c.repo.FindByID(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Task not found"}`))
		return
	}

	isMember, err := c.repo.IsUserMember(task.ProjectID, userID)
	if err != nil || !isMember {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	if err := c.repo.Delete(taskID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to delete task"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Task deleted successfully"}`))
}

func (c *TaskController) getTask(w http.ResponseWriter, taskID uint, userID uint) {
	task, err := c.repo.FindByID(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Task not found"}`))
		return
	}

	isMember, err := c.repo.IsUserMember(task.ProjectID, userID)
	if err != nil || !isMember {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Access denied"}`))
		return
	}

	json.NewEncoder(w).Encode(task)
}
