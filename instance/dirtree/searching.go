package dirtree

import "strings"

func (t *DirTree) SetSearch(s string) {
	t.currentSearch = s
}

func (t DirTree) NextMatch(starting int) (int, bool) {
	for i := starting; i < len(t.array); i++ {
		if strings.Contains(t.array[i].Name, t.currentSearch) {
			return i, true
		}
	}
	return -1, false
}

func (t DirTree) PrevMatch(starting int) (int, bool) {
	for i := starting; i >= 0; i-- {
		if strings.Contains(t.array[i].Name, t.currentSearch) {
			return i, true
		}
	}
	return -1, false
}
