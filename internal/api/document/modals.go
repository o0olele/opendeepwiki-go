package document

import "github.com/o0olele/opendeepwiki-go/internal/database/models"

type Document struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	Title   string `json:"title"`
}

type Overview struct {
	ID       uint       `json:"id"`
	Content  string     `json:"content"`
	Catalogs []*Catalog `json:"catalogs"`
}

type Catalog struct {
	Index    int        `json:"index"`
	Title    string     `json:"title"`
	Children []*Catalog `json:"children"`
	ParentId int        `json:"-"`
}

func NewOverview(id uint, content string, docs []*models.Document) *Overview {
	tmp := &Overview{
		ID:      id,
		Content: content,
	}
	var catalogMap = make(map[int]*Catalog)
	for _, doc := range docs {
		catalog := &Catalog{
			Index:    doc.Index,
			Title:    doc.Title,
			ParentId: int(doc.ParentId),
		}
		catalogMap[doc.Index] = catalog
	}

	for _, catalog := range catalogMap {
		if catalog.ParentId == 0 {
			continue
		}
		parentCatalog, ok := catalogMap[catalog.ParentId]
		if !ok {
			continue
		}
		parentCatalog.Children = append(parentCatalog.Children, catalog)
	}

	for _, catalog := range catalogMap {
		if catalog.ParentId == 0 {
			tmp.Catalogs = append(tmp.Catalogs, catalog)
		}
	}

	if len(tmp.Catalogs) == 1 && tmp.Catalogs[0].Title == "" {
		tmp.Catalogs = tmp.Catalogs[0].Children
	}
	return tmp
}
