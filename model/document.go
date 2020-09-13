package model

type Document struct {
	Owner      Profile `json:"owner"`
	Updater    Profile `json:"updater"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Permission int     `json:"permission"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}
