package model

//SearchUserRes model
type SearchUserRes struct {
	Total int       `json:"total"`
	Users []Profile `json:"users"`
}

//SearchTeamRes model
type SearchTeamRes struct {
	Total int       `json:"total"`
	Teams []Profile `json:"teams"`
}
