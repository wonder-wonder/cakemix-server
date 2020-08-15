package model

//MemberInfoReq model
type MemberInfoReq struct {
	Member     string `json:"member"`
	Permission int    `json:"permission"`
}

//MemberInfoRes model
type MemberInfoRes struct {
	Member     Profile `json:"member"`
	Permission int     `json:"permission"`
}
