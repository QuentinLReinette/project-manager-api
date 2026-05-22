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

func (r *TaskRepository) FindByProjectID(projectID uint, statusFilter string) ([]models.Task, error) {
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
	var project models.Project
	err := r.db.Preload("Participants").First(&project, projectID).Error
	if err != nil {
		return false, err
	}

	if project.OwnerID == userID {
		return true, nil
	}

	for _, participant := range project.Participants {
		if participant.ID == userID {
			return true, nil
		}
	}

	return false, nil
}
