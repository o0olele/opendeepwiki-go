package models

import (
	"gorm.io/gorm"
)

type RepositoryTask struct {
	gorm.Model
	RepositoryID uint   `gorm:"index" json:"repository_id"`
	GitURL       string `gorm:"uniqueIndex" json:"git_url"`
	Status       int    `json:"status"` // 0: Pending, 1: Cloning, 2: Analyzing, 3: Completed
	Errors       string `json:"errors"`
}

func (t *RepositoryTask) StatusString() string {
	switch t.Status {
	case RepositoryTaskStatusPending:
		return "Pending"
	case RepositoryTaskStatusCloned:
		return "Cloning"
	case RepositoryTaskStatusAnalyzed:
		return "Analyzing"
	case RepositoryTaskStatusCompleted:
		return "Completed"
	case RepositoryTaskStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

const (
	RepositoryTaskStatusPending = iota
	RepositoryTaskStatusCloned
	RepositoryTaskStatusAnalyzed
	RepositoryTaskStatusCompleted
	RepositoryTaskStatusFailed
)
