package model

//MemberInfoReq model
type MemberInfoReq struct {
	Member     string `json:"member"`
	Permission int    `json:"permission"`
}

//MemberInfoRes model
type MemberInfoRes struct {
	Total   int          `json:"total"`
	Members []MemberInfo `json:"members"`
}

//MemberInfo model
type MemberInfo struct {
	Member     Profile `json:"member"`
	Permission int     `json:"permission"`
}
