package models

import (
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	}
	return false
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:191;unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Project struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Title        string    `gorm:"size:255;not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	OwnerID      uint      `gorm:"not null" json:"owner_id"`
	Owner        User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Participants []User    `gorm:"many2many:project_participants;" json:"participants,omitempty"`
	Tasks        []Task    `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"tasks,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Task struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Title        string     `gorm:"size:255;not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	Status       TaskStatus `gorm:"size:50;default:'todo';not null" json:"status"`
	ProjectID    uint       `gorm:"not null" json:"project_id"`
	AssignedToID *uint      `gorm:"default:null" json:"assigned_to_id"`
	AssignedTo   *User      `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
