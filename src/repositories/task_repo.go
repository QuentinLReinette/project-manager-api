package repositories

import (
	"project-manager/src/models"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) FindByID(id uint) (*models.Task, error) {
	var task models.Task
	err := r.db.Preload("AssignedTo").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindByProjectID(projectID uint, statusFilter models.TaskStatus) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.Where("project_id = ?", projectID).Preload("AssignedTo")

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&models.Task{}, id).Error
}

// check if a user is part of a project
func (r *TaskRepository) IsUserMember(projectID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Project{}).
		Joins("LEFT JOIN project_participants ON project_participants.project_id = projects.id").
		Where("projects.id = ? AND (projects.owner_id = ? OR project_participants.user_id = ?)", projectID, userID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
