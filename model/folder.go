package model

type FolderModel struct {
	UUID             string
	OwnerUUID        string
	ParentFolderUUID string
	Name             string
	Permission       int
	CreatedAt        int
	UpdatedAt        int
	UpdaterUUID      string
}
