package domain

// TeamPerm is enum of team member's permission.
type TeamPerm int

// TeamPerm list
const (
	TeamPermUser TeamPerm = iota
	TeamPermAdmin
	TeamPermOwner
)

// Team model
type Team struct {
	UUID     string
	Teamname string
	Members  []TeamMember
}

// TeamMember model
type TeamMember struct {
	UUID       string
	Permission TeamPerm
	JoinedAt   int64
}
