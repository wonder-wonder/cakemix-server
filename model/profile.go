package model

//Profile model
type Profile struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Bio       string    `json:"bio"`
	IconURI   string    `json:"icon_uri"`
	CreatedAt int64     `json:"created_at"`
	Attr      string    `json:"attr"`
	IsTeam    bool      `json:"is_team"`
	Teams     []Profile `json:"teams"`
	Lang      string    `json:"lang"`
}

//ProfileReq model
type ProfileReq struct {
	Name    *string `json:"name"`
	Bio     *string `json:"bio"`
	IconURI *string `json:"icon_uri"`
	Lang    *string `json:"lang"`
}
