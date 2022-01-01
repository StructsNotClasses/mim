package musictree

import (
	gnc "github.com/rthornton128/goncurses"
)

type Song struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type MusicTree struct {
    CurrentAccessSequence AccessSequence
    Root Directory
    OrderedArray MusicArray
    CurrentIndex int
}

func New(rootPath string) (t MusicTree, err error) {
    root, err := CreateDirectory(rootPath)
    if err != nil {
        return
    }
    t.Root = root

    unsorted := getUnsortedNodes(&t.Root)
    t.OrderedArray = alphabeticalSort(unsorted)
    return
}

func (t *MusicTree) Draw(win *gnc.Window) {
    win.Erase()
    defer win.Refresh()

    win.Println(t.CurrentAccessSequence)

    linesRemaining, _ := win.MaxYX()

    t.printDirectory(t.Root, win, 0, linesRemaining - 1)
}

func (tree *MusicTree) printDirectory(d Directory, win *gnc.Window, depth int, linesRemaining int) int {
    if linesRemaining <= 0 {
        return 0
    } else {
        y, _ := win.CursorYX()
        if tree.dirIsSelected(d) {
            win.AttrOn(gnc.A_STANDOUT) 
            win.MovePrintln(y, depth, d.Name)
            win.AttrOff(gnc.A_STANDOUT) 
        } else {
            win.MovePrintln(y, depth, d.Name)
        }
        linesRemaining--
        y++
        if linesRemaining <= 0 {
            return 0
        } else if tree.isBeingAccessed(d) {
            for _, subdir := range(d.Subdirectories) {
                if linesRemaining <= 0 {
                    break
                }
                if tree.isBeingAccessed(subdir) {
                    oldLR := linesRemaining
                    linesRemaining = tree.printDirectory(subdir, win, depth + 1, linesRemaining)
                    y += oldLR - linesRemaining
                } else {
                    if tree.dirIsSelected(subdir) {
                        win.AttrOn(gnc.A_STANDOUT) 
                        win.MovePrintln(y, depth + 1, subdir.Name) 
                        win.AttrOff(gnc.A_STANDOUT)
                    } else {
                        win.MovePrintln(y, depth + 1, subdir.Name) 
                    }
                    y++
                    linesRemaining--
                }
            }
            for _, song := range(d.Songs) {
                if linesRemaining <= 0 {
                    break
                }
                if tree.songIsSelected(song) {
                    win.AttrOn(gnc.A_STANDOUT) 
                    win.MovePrintln(y, depth + 1, song.Name)
                    win.AttrOff(gnc.A_STANDOUT)
                }
                y++
                linesRemaining--
            }
            return linesRemaining
        } else {
            return linesRemaining
        }
    }
} 

func (tree *MusicTree) dirIsSelected(dir Directory) bool {
    currentEntry := tree.OrderedArray[tree.CurrentIndex]
    if currentEntry.Type != DirectoryEntry {
        return false
    }
    path, err := currentEntry.Path()
    return err == nil && path == dir.Path
}

func (tree *MusicTree) songIsSelected(song Song) bool {
    currentEntry := tree.OrderedArray[tree.CurrentIndex]
    if currentEntry.Type != SongEntry {
        return false
    }
    path, err := currentEntry.Path()
    return err == nil && path == song.Path
}

func (tree *MusicTree) isBeingAccessed(d Directory) bool {
    if len(tree.CurrentAccessSequence) < 1 {
        return false
    }
    return directoryIsPresentInAccessSequence(tree.CurrentAccessSequence[1:], tree.Root, d)
}

func directoryIsPresentInAccessSequence(seq AccessSequence, root Directory, target Directory) bool {
    if len(seq) <= 0 {
        return false
    } else if root.Path == target.Path {
        return true
    } else {
        // if the access key is targeting a subdirectory
        if seq[0] < len(root.Subdirectories) {
            return directoryIsPresentInAccessSequence(seq[1:], root.Subdirectories[seq[0]], target)
        } else { // the access key is targeting a file, which cannot be the target directory by nature
            return false
        }
    }
}
