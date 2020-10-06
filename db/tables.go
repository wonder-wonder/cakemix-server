package db

// TeamPerm is enum of team member's permission.
type TeamPerm int

// TeamPerm list
const (
	TeamPermOwner TeamPerm = iota
	TeamPermAdmin
	TeamPermUser
)

// FilePerm is enum of file and folder permission.
type FilePerm int

// FilePerm list
const (
	FilePermPrivate FilePerm = iota
	FilePermRead
	FilePermReadWrite
)

// Auth table model
type Auth struct {
	UUID     string
	Email    string
	Password string
	Salt     string
}

// Profile table model
type Profile struct {
	UUID     string
	Name     string
	Bio      string
	IconURI  string
	CreateAt int64
	Attr     string
	Lang     string
}

// TeamMember table model
type TeamMember struct {
	TeamUUID   string
	UserUUID   string
	Permission TeamPerm
	JoinAt     int64
}

// Folder table model
type Folder struct {
	UUID             string
	OwnerUUID        string
	ParentFolderUUID string
	Name             string
	Permission       FilePerm
	CreatedAt        int64
	UpdatedAt        int64
	UpdaterUUID      string
}

// Document table model
type Document struct {
	UUID             string
	OwnerUUID        string
	ParentFolderUUID string
	Title            string
	Permission       FilePerm
	CreatedAt        int64
	UpdatedAt        int64
	UpdaterUUID      string
}