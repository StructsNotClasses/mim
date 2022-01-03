package instance

import (
	"github.com/StructsNotClasses/musicplayer/musicarray"

	gnc "github.com/rthornton128/goncurses"
)

type DirTree struct {
	currentIndex int
	array        musicarray.MusicArray
}

func (t *DirTree) SelectUp() {
	if t.currentIndex <= 0 {
		return
	}

	current := t.array[t.currentIndex]
	if current.Type == musicarray.DirectoryEntry && current.Dir.PrevDirectoryIndex >= 0 {
		t.Select(current.Dir.PrevDirectoryIndex)
	} else if t.array[t.currentIndex-1].Depth > current.Depth {
		i := t.currentIndex - 1
		for ; t.array[i].Depth > current.Depth; i-- {
		}
		t.Select(i)
	} else {
		t.Select(t.currentIndex - 1)
	}
}

func (t *DirTree) SelectDown() {
	if t.currentIndex >= len(t.array)-1 {
		return
	}

	current := t.array[t.currentIndex]
	if current.Dir.EndDirectoryIndex == len(t.array) {
		return
	}

	if current.Type == musicarray.DirectoryEntry && !current.Dir.ManuallyExpanded {
		t.Select(current.Dir.EndDirectoryIndex)
	} else {
		t.Select(t.currentIndex + 1)
	}
}

func (t *DirTree) Select(index int) {
	t.markAutoExpanded(false)
	t.currentIndex = index
	t.markAutoExpanded(true)
}

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

func (t *DirTree) Draw(w *gnc.Window) {
	w.Erase()
	defer w.Refresh()
	t.printEntry(0, 0, w)
}

func (t *DirTree) printEntry(index int, line int, win *gnc.Window) (nextIndex, nextLine int) {
	h, _ := win.MaxYX()
	if index >= len(t.array) || line >= h {
		return index, line
	}
	if t.array[index].Type == musicarray.DirectoryEntry {
		return t.printDirectory(index, line, win)
		return index + 1, line + 1
	} else {
		printSongName(win, line, t.array[index].Depth, t.array[index].Name, index == t.currentIndex)
		return index + 1, line + 1
	}
}

func (t *DirTree) printDirectory(startingIndex, startingLine int, w *gnc.Window) (int, int) {
	h, _ := w.MaxYX()
	if startingIndex >= len(t.array) || startingLine >= h {
		return startingIndex, startingLine
	}

	entry := t.array[startingIndex]
	dir := entry.Dir
	isOpen := dir.AutoExpanded || dir.ManuallyExpanded
	printDirName(w, startingLine, entry.Depth, entry.Name, isOpen, startingIndex == t.currentIndex)

	if isOpen {
		currentIndex := startingIndex + 1
		currentLine := startingLine + 1
		for currentIndex < dir.EndDirectoryIndex {
			currentIndex, currentLine = t.printEntry(currentIndex, currentLine, w)
			if currentIndex >= len(t.array) || currentLine >= h {
				return currentIndex, currentLine
			}
		}
		return dir.EndDirectoryIndex, currentLine + 1
	} else {
		return dir.EndDirectoryIndex, startingLine + 1
	}
}

func (t *DirTree) printSong(songIndex, line int, w *gnc.Window) {
	printSongName(w, line, t.array[songIndex].Depth, t.array[songIndex].Name, songIndex == t.currentIndex)
}

func printDirName(win *gnc.Window, y int, x int, name string, isOpen, isSelected bool) {
	_, w := win.MaxYX()
	var leadChars string
	if isOpen {
		leadChars = "v "
	} else {
		leadChars = "> "
	}
	if isSelected {
		win.AttrOn(gnc.A_STANDOUT)
		win.AttrOff(gnc.A_NORMAL)
		win.MovePrintln(y, x, truncate("=>"+name, w-x))
		//win.Println(truncate("=>" + name, w - x))
		win.AttrOff(gnc.A_STANDOUT)
		win.AttrOn(gnc.A_NORMAL)
	} else {
		win.MovePrintln(y, x, truncate(leadChars+name, w-x))
		//win.Println(truncate(leadChars + name, w - x))
	}
}

func printSongName(win *gnc.Window, y int, x int, name string, isSelected bool) {
	_, w := win.MaxYX()
	leadChar := "~ "
	if isSelected {
		win.AttrOn(gnc.A_STANDOUT)
		win.AttrOff(gnc.A_NORMAL)
		leadChar = "=>"
	}
	win.MovePrintln(y, x, truncate(leadChar+name, w-x))
	//win.Println(truncate(leadChar + name, w - x))
	if isSelected {
		win.AttrOff(gnc.A_STANDOUT)
		win.AttrOn(gnc.A_NORMAL)
	}
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
