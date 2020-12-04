package domain

// Document model
type Document struct {
	UUID             string
	Title            string
	OwnerUUID        string
	ParentFolderUUID string
	Permission       FilePerm
	CreatedAt        int64
	UpdatedAt        int64
	UpdaterUUID      string
	TagID            int
}

// DocumentRevision model
type DocumentRevision struct {
	UUID      string
	Text      string
	UpdatedAt int64
}
