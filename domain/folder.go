package domain

// FilePerm is enum of file and folder permission.
type FilePerm int

// FilePerm list
const (
	FilePermPrivate FilePerm = iota
	FilePermRead
	FilePermReadWrite
)

// Folder model
type Folder struct {
	UUID             string
	Name             string
	OwnerUUID        string
	ParentFolderUUID string
	Permission       FilePerm
	CreatedAt        int64
	UpdatedAt        int64
	UpdaterUUID      string
}
