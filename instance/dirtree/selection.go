package dirtree

import (
	"github.com/StructsNotClasses/mim/musicarray"

	"strings"
)

func (t *DirTree) SelectUp() {
	if t.currentIndex <= 0 {
		return
	}

	current := t.array[t.currentIndex]
	// start at the entry one index up and search for the first entry inside only expanded directories without increasing depth
	lowest := t.currentIndex - 1
	maxDepth := t.array[lowest].Depth
	for i := lowest; t.array[i].Depth >= current.Depth; i-- {
		if t.array[i].Depth >= maxDepth {
			continue
		}
		if t.array[i].Type == musicarray.DirectoryEntry {
			maxDepth = t.array[i].Depth
			if !t.array[i].Dir.ManuallyExpanded && !t.array[i].Dir.AutoExpanded {
				lowest = i
			}
		}
	}
	t.Select(lowest)
}

func (t *DirTree) SelectDown() {
	if t.currentIndex+1 >= len(t.array) {
		return
	}

	current := t.array[t.currentIndex]
	if current.Type == musicarray.DirectoryEntry && !current.Dir.Expanded() {
		// if the last directory at root level is selected don't do anything
		if current.Dir.EndDirectoryIndex != len(t.array) {
			t.Select(current.Dir.EndDirectoryIndex)
		}
	} else {
		t.Select(t.currentIndex + 1)
	}
}

func (t *DirTree) SelectEnclosing(index int) {
	targetDepth := t.array[index].Depth - 1
	if index != 0 {
		i := index
		for ; t.array[i].Depth != targetDepth; i-- {
		}
		t.Select(i)
	}
}

func (t *DirTree) SelectNextMatch(s string) {
	for i := t.currentIndex; i < len(t.array); i++ {
		if strings.Contains(t.array[i].Name, s) {
			t.Select(i)
			return
		}
	}
}

func (t *DirTree) Select(index int) {
	t.markAutoExpanded(false)
	t.currentIndex = index
	t.markAutoExpanded(true)
}

// markAutoExpanded traverses the directory tree upwards marking all directories containing the currently selected item as automatically expanded
func (t *DirTree) markAutoExpanded(value bool) {
	i := t.currentIndex - 1
	depth := t.array[t.currentIndex].Depth
	for ; i >= 0; i-- {
		if t.array[i].Depth == depth-1 {
			// 1 level up containing folder located
			depth--
			t.array[i].Dir.AutoExpanded = value
		}
	}
}
