package dirtree

import (
	"github.com/StructsNotClasses/musicplayer/musicarray"

	"errors"
)

type DirTree struct {
	currentIndex int
	array        musicarray.MusicArray
}

func New(a musicarray.MusicArray) DirTree {
	return DirTree{
		currentIndex: 0,
		array:        a,
	}
}

func (t *DirTree) Toggle(index int) error {
	if t.array[index].Type != musicarray.DirectoryEntry {
		return errors.New("dirtree.Toggle: can only toggle directories.")
	}
	t.array[index].Dir.ManuallyExpanded = !t.array[index].Dir.ManuallyExpanded
	return nil
}
