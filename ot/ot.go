package ot

import (
	"errors"
	"unicode/utf16"
)

// OpType is enum of OT operation
type OpType int

// OpType enums
const (
	OpTypeRetain = iota
	OpTypeInsert
	OpTypeDelete
)

// Op is structure for part of OT operation
type Op struct {
	OpType OpType
	Len    int
	Text   string
}

// Ops is structure for OT operation
type Ops struct {
	User string
	Ops  []Op
}

// OT is structure for OT session
type OT struct {
	Text     string
	History  map[int]Ops
	Revision int
}

// New creates OT
func New(text string) *OT {
	return &OT{Text: text, Revision: 0, History: map[int]Ops{}}
}

// Transform converts OT operations
func (ot *OT) Transform(rev int, ops Ops) (Ops, error) {
	if rev < 0 || rev > ot.Revision {
		return Ops{}, errors.New("Revision is out of range")
	}
	ret := ops
	// Check all history after rev
	// for _, h := range ot.History[rev:] {
	for i := rev; i < ot.Revision; i++ {
		h := ot.History[i]
		//temporary new ops
		tops := Ops{User: ret.User}
		//History op counter
		j := 0
		//Remaining retain or delete
		remain := 0
		isDel := false
		// Check all Ops
		for _, op := range ret.Ops {
			// If insert, just insert
			if op.OpType == OpTypeInsert {
				tops.Ops = append(tops.Ops, op)
				continue
			}
			// Current length to process
			cur := op.Len
			for cur > 0 {
				// Get remain
				if remain == 0 {
					for j < len(h.Ops) {
						// If insert, just retain and ignore
						if h.Ops[j].OpType == OpTypeInsert {
							tops.Ops = append(tops.Ops, Op{OpType: OpTypeRetain, Len: h.Ops[j].Len})
							j++
							continue
						}
						// Get length for remain
						remain = h.Ops[j].Len
						isDel = h.Ops[j].OpType == OpTypeDelete
						j++
						break
					}
					// If remain is not left, the operation is inconsistent
					if remain == 0 {
						return Ops{}, errors.New("Operation is inconsistent")
					}
				}

				if cur > remain {
					//Use all remain
					if !isDel {
						//If retain, left it. If delete, discard it
						tops.Ops = append(tops.Ops, Op{OpType: op.OpType, Len: remain})
					}
					cur -= remain
					remain = 0
				} else {
					//Use remain and cur is finished
					if !isDel {
						//If retain, left it. If delete, discard it
						tops.Ops = append(tops.Ops, Op{OpType: op.OpType, Len: cur})
					}
					remain -= cur
					cur = 0
				}
			}
		}
		// hist ops is remain, it may insert operation
		if j < len(h.Ops) {
			// If insert, just retain and ignore
			if h.Ops[j].OpType == OpTypeInsert {
				tops.Ops = append(tops.Ops, Op{OpType: OpTypeRetain, Len: h.Ops[j].Len})
				j++
			}
			if j < len(h.Ops) {
				// If stil remain, it's inconsistent
				return Ops{}, errors.New("Operation is inconsistent")
			}
		}
		//Merge same type
		if len(tops.Ops) > 1 {
			for i := 1; i < len(tops.Ops); i++ {
				if tops.Ops[i-1].OpType == tops.Ops[i].OpType {
					tops.Ops[i-1].Len += tops.Ops[i].Len
					tops.Ops[i-1].Text += tops.Ops[i].Text
					tops.Ops = append(tops.Ops[:i], tops.Ops[i+1:]...)
					i--
				}
			}
		}
		ret = tops
	}
	return ret, nil
}

// Operate applies OT operation
func (ot *OT) Operate(rev int, ops Ops) (Ops, error) {
	opstrans, err := ot.Transform(rev, ops)
	if err != nil {
		panic(err)
	}
	loc := 0
	trune := utf16.Encode([]rune(ot.Text))
	for _, v := range opstrans.Ops {
		if v.OpType == OpTypeRetain {
			loc += v.Len
		} else if v.OpType == OpTypeInsert {
			trune = append(trune[:loc], append(utf16.Encode([]rune(v.Text)), trune[loc:]...)...)
			loc += v.Len
		} else if v.OpType == OpTypeDelete {
			trune = append(trune[:loc], trune[v.Len+loc:]...)
		}
	}
	ot.Text = string(utf16.Decode(trune))
	ot.History[ot.Revision] = opstrans
	ot.Revision++
	return opstrans, nil
}
