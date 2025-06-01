package repository

type Repository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	RepoId      uint   `json:"id"`
	Status      string `json:"status"`
}
