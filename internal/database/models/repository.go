package models

import (
	"gorm.io/gorm"
)

// Repository Repository model.
type Repository struct {
	gorm.Model
	GitURL             string `gorm:"uniqueIndex" json:"git_url"`
	Name               string `json:"name"`
	Path               string `json:"path"` // local path to the repository database, default is {repoDir}/{name}
	Description        string `json:"description"`
	Status             int    `json:"status"`
	Branch             string `json:"branch"` // default branch name, default is maste
	Overview           string `json:"overview"`
	Readme             string `json:"readme"` // readme file content, default is README.md
	StructedCatalogue  string `json:"structured_catalogue"`
	StructedCodePath   string `json:"structured_code_path"`   // path to the structured code database, default is {repoDir}/{name}.db
	StructedVectorPath string `json:"structured_vector_path"` // path to the structured vector database, default is {repoDir}/{name}.db
	Language           string `json:"language"`
}

func (r *Repository) StatusString() string {
	return getStatusString(r.Status)
}

func (r *Repository) WebStatus() string {
	switch r.Status {
	case RepositoryStatusPending, RepositoryStatusCloned, RepositoryStatusAnalyzed:
		return "pending"
	case RepositoryStatusCompleted:
		return "success"
	case RepositoryStatusFailed:
		return "failed"
	default:
		return "pending"
	}
}

func getStatusString(status int) string {
	switch status {
	case RepositoryStatusPending:
		return "Pending"
	case RepositoryStatusCloned:
		return "Cloning"
	case RepositoryStatusAnalyzed:
		return "Analyzing"
	case RepositoryStatusCompleted:
		return "Completed"
	case RepositoryStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

const (
	RepositoryStatusPending = iota
	RepositoryStatusCloned
	RepositoryStatusAnalyzed
	RepositoryStatusCompleted
	RepositoryStatusFailed
)
