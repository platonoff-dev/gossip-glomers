package main

type State struct {
	Data map[int][]int
}

func NewState() State {
	return State{
		Data: map[int][]int{},
	}
}

func (s *State) Transact(transactions []*Transaction) {
	for _, t := range transactions {
		if t.Op == "r" {
			t.Values = s.Data[t.Key]
		}

		if t.Op == "append" {
			s.Data[t.Key] = append(s.Data[t.Key], t.Values...)
		}
	}
}
