package dao

import (
	"github.com/o0olele/opendeepwiki-go/internal/database"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"gorm.io/gorm"
)

type DocumentDao struct {
	db *gorm.DB
}

func NewDocumentDao() *DocumentDao {
	return &DocumentDao{db: database.GetDB()}
}

func (d *DocumentDao) CreateDocument(document *models.Document) error {
	return d.db.Create(document).Error
}

func (d *DocumentDao) CreateDocuments(documents []*models.Document) error {
	return d.db.Create(documents).Error
}

func (d *DocumentDao) GetDocumentByRepoId(repoId uint) ([]*models.Document, error) {
	var documents []*models.Document
	result := d.db.Where("repo_id = ?", repoId).Find(&documents)
	if result.Error != nil {
		return nil, result.Error
	}
	return documents, nil
}

func (d *DocumentDao) GetDocumentByRepoIdAndIndex(repoId uint, index int) (*models.Document, error) {
	var document models.Document
	result := d.db.Where("repo_id =? AND `index` =?", repoId, index).First(&document)
	if result.Error != nil {
		return nil, result.Error
	}
	return &document, nil
}
