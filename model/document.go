package model

type DocumentModel struct {
	UUID             string
	OwnerUUID        string
	ParentFolderUUID string
	Title            string
	Permission       int
	CreatedAt        int
	UpdatedAt        int
	UpdaterUUID      string
	TagID            int
}
