package model

// Document is structure for document info
type Document struct {
	UUID           string  `json:"uuid"`
	Owner          Profile `json:"owner"`
	Updater        Profile `json:"updater"`
	Title          string  `json:"title"`
	Body           string  `json:"body"`
	Permission     int     `json:"permission"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
	Editable       bool    `json:"editable"`
	ParentFolderID string  `json:"parentfolderid"`
}

// CreateDocumentRes is structure for response of document creation
type CreateDocumentRes struct {
	DocumentID string `json:"doc_id"`
}

// DocumentModifyReqModel is structure for request document info modification
type DocumentModifyReqModel struct {
	OwnerUUID  string `json:"owneruuid"`
	Permission int    `json:"permission"`
}
