package instance

import (
	"github.com/StructsNotClasses/musicplayer/musicarray"

    "strings"

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
	if t.currentIndex + 1 >= len(t.array) {
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
        for ; t.array[i].Depth != targetDepth; i-- {}
        t.Select(i)
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

func (t *DirTree) Draw(w *gnc.Window) {
	w.Erase()
	defer w.Refresh()
    _, width := w.MaxYX()
    buffer := t.toString(width)
    lines := strings.Split(buffer, "\n")
    //lineCount := strings.Count(buffer, "\n")
    /*
    if len(lines) > height {
        w.MovePrint(0, 0, buffer)
    } else {
        w.MovePrint(0, 0, buffer)
    }
    */
    for _, line := range(lines) {
        if strings.Contains(line, "=>") {
            w.AttrOn(gnc.A_STANDOUT)
            w.AttrOff(gnc.A_NORMAL)
            w.Println(line)
            w.AttrOn(gnc.A_NORMAL)
            w.AttrOff(gnc.A_STANDOUT)
        } else {
            w.Println(line)
        }
    }
}

func (t *DirTree) toString(width int) string {
    _, result := t.entryToString(0, width)
    return result
}

func (t *DirTree) entryToString(index, width int) (nextIndex int, result string) {
	if index >= len(t.array) {
		return index, ""
	}
	if t.array[index].Type == musicarray.DirectoryEntry {
		return t.directoryToString(index, width)
	} else {
		return index + 1, songToString(width, t.array[index].Depth, t.array[index].Name, index == t.currentIndex)
	}

}

func (t *DirTree) directoryToString(index, width int) (int, string) {
    ret := ""
	if index >= len(t.array) {
		return index, ret
	}

	entry := t.array[index]
	dir := entry.Dir
	isOpen := dir.AutoExpanded || dir.ManuallyExpanded
	ret += dirNameToString(width, entry.Depth, entry.Name, isOpen, index == t.currentIndex)

	if isOpen {
		currentIndex := index + 1
		for currentIndex < dir.EndDirectoryIndex {
            next, s := t.entryToString(currentIndex, width)
            currentIndex = next
            ret += s
			if currentIndex >= len(t.array) {
				return currentIndex, ret
			}
		}
		return dir.EndDirectoryIndex, ret
	} else {
		return dir.EndDirectoryIndex, ret
	}
}

func dirNameToString(width, indent int, name string, isOpen, isSelected bool) string {
    leadChars := "> "
	if isOpen {
		leadChars = "v "
	} 
	if isSelected {
		return spaces(indent) + truncate("=>" + name, width - indent - 1) + "\n"
	} else {
        return spaces(indent) + truncate(leadChars + name, width - indent - 1) + "\n"
	}
}

func songToString(width, indent int, name string, isSelected bool) string {
	leadChar := "~ "
	if isSelected {
		leadChar = "=>"
	}
    return spaces(indent) + truncate(leadChar + name, width - indent - 1) + "\n"
}

func spaces(count int) string {
    s := ""
    for i := 0; i < count; i++ {
        s += " "
    }
    return s
}

func (t *DirTree) printEntry(index int, line int, win *gnc.Window) (nextIndex, nextLine int) {
	h, _ := win.MaxYX()
	if index >= len(t.array) || line >= h {
		return index, line
	}
	if t.array[index].Type == musicarray.DirectoryEntry {
		return t.printDirectory(index, line, win)
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
		return dir.EndDirectoryIndex, currentLine
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
