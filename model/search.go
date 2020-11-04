package model

//SearchUserTeamRes model
type SearchUserTeamRes struct {
	Total int       `json:"total"`
	Users []Profile `json:"users"`
}
