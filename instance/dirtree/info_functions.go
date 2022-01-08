package dirtree

import (
	"github.com/StructsNotClasses/mim/musicarray"
)

func (t DirTree) IsInRange(index int) bool {
	return index >= 0 && index < len(t.array)
}

func (t DirTree) Depth(index int) int {
	return t.array[index].Depth
}

func (t DirTree) CurrentEntry() musicarray.Entry {
	return t.array[t.currentIndex]
}

func (t DirTree) CurrentIndex() int {
	return t.currentIndex
}

func (t DirTree) ItemCount() int {
	return len(t.array)
}

func (t DirTree) CurrentIsDir() bool {
	return t.IsDir(t.currentIndex)
}

func (t DirTree) IsDir(index int) bool {
	return t.array[index].Type == musicarray.DirectoryEntry
}
