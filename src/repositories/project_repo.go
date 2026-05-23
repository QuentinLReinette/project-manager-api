package repositories

import (
	"context"
	"errors"
	"project-manager/src/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// retrieve a project and preload its relations
func (r *ProjectRepository) FindByID(ctx context.Context, id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).Preload("Owner").Preload("Participants").Preload("Tasks").First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrProjectNotFound
		}
		return nil, err
	}
	return &project, nil
}

// return the projects a user belongs to
func (r *ProjectRepository) FindAllForUser(ctx context.Context, userID uint) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.WithContext(ctx).Preload("Owner").
		Distinct().
		Joins("LEFT JOIN project_participants ON project_participants.project_id = projects.id").
		Where("projects.owner_id = ? OR project_participants.user_id = ?", userID, userID).
		Find(&projects).Error

	return projects, err
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Project{}, id).Error
}

func (r *ProjectRepository) AddParticipantByEmail(ctx context.Context, projectID uint, email string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.ErrUserNotFound
			}
			return err
		}

		var project models.Project
		if err := tx.First(&project, projectID).Error; err != nil {
			return err
		}

		if project.OwnerID == user.ID {
			return models.ErrUserIsOwner
		}

		var count int64
		err := tx.Table("project_participants").
			Where("project_id = ? AND user_id = ?", projectID, user.ID).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return models.ErrUserAlreadyParticipant
		}

		return tx.Model(&project).Association("Participants").Append(&user)
	})
}
