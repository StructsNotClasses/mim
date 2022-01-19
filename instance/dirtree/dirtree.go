package dirtree

import (
	"github.com/StructsNotClasses/mim/musicarray"

	gnc "github.com/rthornton128/goncurses"

	"errors"
)

type DirTree struct {
	win           *gnc.Window
	currentIndex  int
	array         musicarray.MusicArray
	currentSearch string
}

func New(win *gnc.Window, arr musicarray.MusicArray) DirTree {
	return DirTree{
		win:           win,
		currentIndex:  0,
		array:         arr,
		currentSearch: "",
	}
}

func (t *DirTree) Toggle(index int) error {
	if t.array[index].Type != musicarray.DirectoryEntry {
		return errors.New("dirtree.Toggle: can only toggle directories.")
	}
	t.array[index].Dir.ManuallyExpanded = !t.array[index].Dir.ManuallyExpanded
	return nil
}
