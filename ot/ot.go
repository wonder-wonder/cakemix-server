package ot

import (
	"errors"
)

type OpType int

const (
	OpTypeRetain = iota
	OpTypeInsert
	OpTypeDelete
)

type Op struct {
	OpType OpType
	Len    int
	Text   string
}
type Ops struct {
	User string
	Ops  []Op
}
type OT struct {
	Text    string
	History []Ops
}

func New(text string) *OT {
	return &OT{Text: text}
}

func (ot *OT) Transform(rev int, ops Ops) (Ops, error) {
	if rev < 0 || rev > len(ot.History) {
		return Ops{}, errors.New("Revision is out of range")
	}
	ret := ops
	// Check all history after rev
	for _, h := range ot.History[rev:] {
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

func (ot *OT) Operate(rev int, ops Ops) (Ops, error) {
	opstrans, err := ot.Transform(rev, ops)
	if err != nil {
		panic("hoge")
	}
	loc := 0
	for _, v := range opstrans.Ops {
		if v.OpType == OpTypeRetain {
			loc += v.Len
		} else if v.OpType == OpTypeInsert {
			ot.Text = ot.Text[:loc] + v.Text + ot.Text[loc:]
			loc += v.Len
		} else if v.OpType == OpTypeDelete {
			ot.Text = ot.Text[:loc] + ot.Text[v.Len+loc:]
		}
		println(loc, v.Len, v.OpType, ot.Text)
	}
	ot.History = append(ot.History, opstrans)
	return opstrans, nil
}
