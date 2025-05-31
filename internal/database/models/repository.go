package models

import (
	"gorm.io/gorm"
)

// Repository Repository model.
type Repository struct {
	gorm.Model
	GitURL      string           `gorm:"uniqueIndex" json:"git_url"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Tasks       []RepositoryTask `gorm:"foreignKey:RepositoryID" json:"-"`
}
