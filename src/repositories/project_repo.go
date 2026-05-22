package repositories

import (
	"project-manager/src/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

// retrieve a project and preload its relations
func (r *ProjectRepository) FindByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Owner").Preload("Participants").Preload("Tasks").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// return projects where the user is the owner or a participant
func (r *ProjectRepository) FindAllForUser(userID uint) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Preload("Owner").
		Distinct().
		Joins("LEFT JOIN project_participants ON project_participants.project_id = projects.id").
		Where("projects.owner_id = ? OR project_participants.user_id = ?", userID, userID).
		Find(&projects).Error

	return projects, err
}

func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *ProjectRepository) AddParticipantByEmail(projectID uint, email string) error {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return err 
	}

	var project models.Project
	if err := r.db.First(&project, projectID).Error; err != nil {
		return err
	}

	return r.db.Model(&project).Association("Participants").Append(&user)
}
