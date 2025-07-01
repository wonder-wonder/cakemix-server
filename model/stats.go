package model

// Stats represents system statistics
type Stats struct {
	Users     int `json:"users"`
	Teams     int `json:"teams"`
	Documents int `json:"documents"`
	Folders   int `json:"folders"`
}