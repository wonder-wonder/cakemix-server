package ot

import (
	"fmt"
	"strconv"
)

// OpType is operation type of OT
type OpType int

// OpType list
const (
	OpTypeInsert OpType = iota
	OpTypeDelete
)

// Operation is operation info of OT
type Operation struct {
	User string
	Loc  int
	Op   OpType
	Text string
}

// OT is object for OT
type OT struct {
	Text    string
	Users   []string
	History []Operation
}

// New create OT object
func New(text string, users []string) *OT {
	return &OT{Text: text, Users: users, History: []Operation{}}
}

// Insert adds new string into specified location
func (ot *OT) Insert(user string, seq int, loc int, text string) int {
	// Check sequence number is correct
	if seq > len(ot.History) {
		return -1
	}

	// Check user is exists
	flg := false
	for _, v := range ot.Users {
		if v == user {
			flg = true
			break
		}
	}
	if !flg {
		return -1
	}

	op := Operation{User: user, Loc: loc, Op: OpTypeInsert, Text: text}
	// Transform parameter by history
	for i := seq; i < len(ot.History); i++ {
		h := ot.History[i]
		// If the text is modified before and it affects
		//   \/ 1. Inserting fuga (h.Loc)
		//  hoge
		//    /\ 2. Inserting piyo (affected by 1) (op.Loc)
		if h.Loc < op.Loc {
			if h.Op == OpTypeInsert {
				// Just add the number of added text into location
				// hoge         -->  hofugage
				//   /\ loc: 3             /\ loc: 3+4=7
				op.Loc += len(h.Text)
			} else if h.Op == OpTypeDelete {
				del, _ := strconv.Atoi(h.Text)
				// Subtract the number of removing text from location
				// hofugage         -->  hoge
				//       /\ loc: 7         /\ loc: 7-4=3
				op.Loc -= del
				if op.Loc < h.Loc {
					// If the loc is on the text removed, new loc is just the beginning of loc of removing operation
					// hofugage         -->  hoge
					//    /\ loc: 4           /\ loc: 2 (removing "fuga" operation loc is 2)
					op.Loc = h.Loc
				}
			}
		}
	}

	// Check location is in the range of text
	if op.Loc > len(ot.Text) {
		return -1
	}

	// Update text and history
	ot.Text = ot.Text[:op.Loc] + text + ot.Text[op.Loc:]
	ot.History = append(ot.History, op)
	return len(ot.History)
}

// Delete removes text from specified location
func (ot *OT) Delete(user string, seq int, loc int, del int) int {
	// Check sequence number is correct
	if seq > len(ot.History) {
		return -1
	}

	// Check user is exists
	flg := false
	for _, v := range ot.Users {
		if v == user {
			flg = true
			break
		}
	}
	if !flg {
		return -1
	}

	op := Operation{User: user, Loc: loc, Op: OpTypeDelete, Text: strconv.Itoa(del)}
	// Transform parameter by history
	for i := seq; i < len(ot.History); i++ {
		h := ot.History[i]
		// Omit the some explanation of algo; please refer Insert func
		if h.Loc < op.Loc {
			if h.Op == OpTypeInsert {
				op.Loc += len(h.Text)
			} else if h.Op == OpTypeDelete {
				hdel, _ := strconv.Atoi(h.Text)
				if hdel+h.Loc > op.Loc {
					// If the loc is on the text removed, new loc is just the beginning of loc of removing operation
					// del is reduced because the partial text is already removed.
					// (Already done the op DELETE loc: 2, del: 4)
					// hofugage                     -->  hoge
					//    /\ loc: 4                       /\ loc: 2 (removing "fuga" operation loc is 2)
					//       del: 3 (remove "gag")           del: 1 (3(current del)-(2(history loc)+4(history del)-4(current loc)))
					del -= h.Loc + hdel - op.Loc
					if del < 0 {
						// If all text is already removed, do nothing
						// (Already done the op DELETE loc: 2, del: 4)
						// hofugage                     -->  hoge
						//   /\ loc: 3                        /\ loc: 2 (removing "fuga" operation loc is 2)
						//      del: 2 (remove "ug")             del: 0 (2-(2+4-4)=-1 -> 0)
						del = 0
					}
					op.Text = strconv.Itoa(del)
					op.Loc = h.Loc
				} else {
					// The removed text is not overlaped and just move the loc
					// hofugage                     -->  hoge
					//       /\ loc: 7                     /\ loc: 7-4=3
					op.Loc -= del
				}
			}
		}
	}

	// Check location is in the range of text
	if op.Loc+del > len(ot.Text) {
		return -1
	}

	// Update text and history
	ot.Text = ot.Text[:op.Loc] + ot.Text[op.Loc+del:]
	ot.History = append(ot.History, op)
	return len(ot.History)
}

func (ot *OT) Test() {
	var u, t, op string
	var s, l int
	seq := 0
	ot.Users = append(ot.Users, "a")
	ot.Users = append(ot.Users, "b")
	ot.Text = "abc"
	for {
		fmt.Print("< ")
		_, err := fmt.Scanf("%d %s %s %d %s", &s, &u, &op, &l, &t)
		if err != nil {
			break
		}
		if op == "i" {
			seq = ot.Insert(u, s, l, t)
		} else if op == "d" {
			del, err := strconv.Atoi(t)
			if err == nil {
				seq = ot.Delete(u, s, l, del)
			}
		}

		for _, v := range ot.History {
			fmt.Printf("[%s %d %d %s]\n", v.User, v.Loc, v.Op, v.Text)
		}
		fmt.Printf("> %d: %s\n", seq, ot.Text)
	}
	temp := "abc"
	println(temp)

	for _, v := range ot.History {
		fmt.Printf("[%s %d %d %s]\n", v.User, v.Loc, v.Op, v.Text)
		if v.Op == OpTypeInsert {
			temp = temp[:v.Loc] + v.Text + temp[v.Loc:]
		} else if v.Op == OpTypeDelete {
			del, _ := strconv.Atoi(v.Text)
			temp = temp[:v.Loc] + temp[v.Loc+del:]
		}
		println(temp)
	}
}
