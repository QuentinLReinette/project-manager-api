package repositories

import (
	"context"
	"errors"
	"project-manager/src/models"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) FindByID(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).Preload("AssignedTo").First(&task, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindByProjectID(ctx context.Context, projectID uint, statusFilter models.TaskStatus) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID).Preload("AssignedTo")

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *TaskRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Task{}, id).Error
}

// check if a user is part of a project
func (r *TaskRepository) IsUserMember(ctx context.Context, projectID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Project{}).
		Joins("LEFT JOIN project_participants ON project_participants.project_id = projects.id").
		Where("projects.id = ? AND (projects.owner_id = ? OR project_participants.user_id = ?)", projectID, userID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
