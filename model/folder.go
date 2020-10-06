package model

type FolderList struct {
	Folder   []Folder   `json:"folder"`
	Document []Document `json:"document"`
}

type Folder struct {
	Owner      Profile `json:"owner"`
	Updater    Profile `json:"updater"`
	Name       string  `json:"name"`
	Permission int     `json:"permission"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

type CreateFolderRes struct {
	FolderID string `json:"folder_id"`
}
