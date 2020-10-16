package model

// FolderList is structure for folder list
type FolderList struct {
	Folder   []Folder   `json:"folder"`
	Document []Document `json:"document"`
}

// Folder is structure for folder info
type Folder struct {
	Owner      Profile `json:"owner"`
	Updater    Profile `json:"updater"`
	Name       string  `json:"name"`
	Permission int     `json:"permission"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

// CreateFolderRes is structure for response of creation folder
type CreateFolderRes struct {
	FolderID string `json:"folder_id"`
}
