package dirtree

import (
	"github.com/StructsNotClasses/musicplayer/musicarray"

    "strings"

	gnc "github.com/rthornton128/goncurses"
)

type Line struct {
    contents string
    isSelected bool
    isDir bool
}

func (t *DirTree) Draw(win *gnc.Window) {
	win.Erase()
	defer win.Refresh()

	height, width := win.MaxYX()

    lines, selectedLine := t.getLines(width)

	centerLine := int(height / 2)
	if len(lines) <= height || selectedLine < centerLine {
		printLines(win, lines, selectedLine)
	} else if selectedLine >= (len(lines) - (height - centerLine)) {
		offsetIndex := selectedLine + height - len(lines) + 1
		printLines(win, lines[len(lines)-height-1:], offsetIndex)
	} else {
		printLines(win, lines[selectedLine-centerLine:selectedLine+(height-centerLine)], centerLine)
	}
}

// printLines prints the provided slice of strings one at a time. The first item in the slice will be printed at y = 0 on the window, second at y = 1, and so on until out of slice items or height reached
func printLines(win *gnc.Window, lines []Line, selectedLine int) {
	height, _ := win.MaxYX()
	for i := 0; i < len(lines) && i <= height; i++ {
		printLine(win, lines[i], i)
	}
}

func printLine(win *gnc.Window, line Line, y int) {
    const dirAttributes = gnc.A_BOLD
    const fileAttributes = gnc.A_NORMAL
    const selectedNameAttributes = gnc.A_STANDOUT
    
    if line.isSelected {
        pointerIndex := strings.Index(line.contents, "=>") + 2
        if len(line.contents) > pointerIndex {
            win.AttrOn(gnc.A_BOLD)
            win.MovePrint(y, 0, line.contents[:pointerIndex])
            win.AttrOff(gnc.A_BOLD)

            win.AttrOn(selectedNameAttributes)
            win.Println(line.contents[pointerIndex:])
            win.AttrOff(selectedNameAttributes)
        }
    } else if line.isDir {
        win.AttrOn(dirAttributes)
		win.MovePrintln(y, 0, line.contents)
        win.AttrOff(dirAttributes)
    } else {
        win.AttrOn(fileAttributes)
        win.MovePrintln(y, 0, line.contents)
        win.AttrOff(fileAttributes)
    }
}

func (t *DirTree) getLines(width int) ([]Line, int) {
    result := []Line{}
    selectedLine := 0
    for i := 0; i < len(t.array); {
        isSelected := i == t.currentIndex
        if isSelected {
            selectedLine = len(result)
        }
        if t.array[i].Type == musicarray.DirectoryEntry {
            isOpen := t.array[i].Dir.AutoExpanded || t.array[i].Dir.ManuallyExpanded
            result = append(result, Line{
                contents: dirNameToString(width, t.array[i].Depth, t.array[i].Name, isOpen, isSelected),
                isSelected: isSelected,
                isDir: true,
            })
            if isOpen {
                i++
            } else {
                i = t.array[i].Dir.EndDirectoryIndex
            }
        } else {
            result = append(result, Line{
                contents: songToString(width, t.array[i].Depth, t.array[i].Name, isSelected),
                isSelected: isSelected,
                isDir: false,
            })
            i++
        }
    }
    return result, selectedLine
}

func dirNameToString(width, indent int, name string, isOpen, isSelected bool) string {
	leadChars := "> "
	if isOpen {
		leadChars = "v "
	}
	if isSelected {
		return spaces(indent) + truncate("=>"+name, width-indent)
	} else {
		return spaces(indent) + truncate(leadChars+name, width-indent)
	}
}

func songToString(width, indent int, name string, isSelected bool) string {
	leadChar := "o "
	if isSelected {
		leadChar = "=>"
	}
	return spaces(indent) + truncate(leadChar+name, width-indent)
}

func spaces(count int) string {
	s := ""
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
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
		win.AttrOff(gnc.A_STANDOUT)
		win.AttrOn(gnc.A_NORMAL)
	} else {
		win.MovePrintln(y, x, truncate(leadChars+name, w-x))
	}
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
