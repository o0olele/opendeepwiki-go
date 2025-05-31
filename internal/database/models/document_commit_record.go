package models

import (
	"gorm.io/gorm"
)

// DocumentCommitRecord Document commit record model.
type DocumentCommitRecord struct {
	gorm.Model
	WarehouseId   string `json:"warehouse_id"`
	Title         string `json:"title"`
	CommitMessage string `json:"commit_message"`
	Author        string `json:"author"`
}
