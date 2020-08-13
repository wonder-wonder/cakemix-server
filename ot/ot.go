package ot

type OpType int

const (
	OpTypeRetain = iota
	OpTypeInsert
	OpTypeDelete
)

type Op struct {
	OpType OpType
	Loc    int
	Text   string
}
type Ops struct {
	User string
	Ops  []Op
}
type OT struct {
	Text    string
	Users   []string
	History []Ops
}

func New(text string, users []string) *OT {
	return &OT{Text: text, Users: users}
}
