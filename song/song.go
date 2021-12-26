package song

import (
    "os"
    "log"
    "errors"
    "fmt"
    "io/ioutil"
	"encoding/json"
    gnc "github.com/rthornton128/goncurses"
)

type Song struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type List []Song

type Tree struct {
    Songs List
    CurrentIndex int32
    CurrentAtTop int32
}

type Directory struct {
    Subdirectories []Directory
    Files []Song
}

/* this is the goal
type Album struct {
    Songs []Song
}

type Artist struct {
    Albums []Album
}

type Tree struct {
    Artists []Artist
}
*/

func CreateList(path string) (List, error) {
	s := List([]Song{})

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	err = json.Unmarshal(bytes, &s)
	return s, err
}

func CreateDirectory(path string) (d Directory, err error) {
    entries, err := os.ReadDir(path)
    if err != nil {
        return
    }

    for _, entry := range(entries) {
        info, err := entry.Info()
        if entry.IsDir() {
            subdir, err := CreateDirectory(path)
            d.Directories = append(d.Directories, )
        } else {
        }
    }
}

func (s *List) AddFile(path string, info os.FileInfo, err error) error {
    if err != nil {
        log.Print(err)
        return nil
    }

    *s = append(*s, Song{info.Name(), path})

    return nil
}

func (tree *Tree) Draw(win *gnc.Window) {
    win.Erase()
    height, width := win.MaxYX()
    for i := int(tree.CurrentAtTop); i < int(tree.CurrentAtTop) + height && i < len(tree.Songs); i++ {
        song := tree.Songs[i]
        if i == int(tree.CurrentIndex) {
            win.AttrOn(gnc.A_STANDOUT)
            win.AttrOff(gnc.A_NORMAL)
            winClampPrintln(win, song.Name, width)
            win.AttrOff(gnc.A_STANDOUT)
            win.AttrOn(gnc.A_NORMAL)
        } else {
            winClampPrintln(win, song.Name, width)
        }
    }
    win.Refresh()
}

//currently need to pass window for height info. this should probably be in the tree struct.
func (tree *Tree) Select(index int32, win *gnc.Window) error {
    if int(index) >= len(tree.Songs) {
        return errors.New(fmt.Sprintf("Cannot play song number %d. There aren't this many songs.", index))
    } else if index < 0 {
        return errors.New(fmt.Sprintf("Cannot play a song at a negative index.", index))
    } 
    tree.CurrentIndex = index

    heightint, _ := win.MaxYX()
    height := int32(heightint)
    delta := tree.CurrentAtTop - tree.CurrentIndex
    if delta > height - 1 {
        tree.CurrentAtTop = tree.CurrentIndex - height - 1
    } else if delta < 0 {
        tree.CurrentAtTop = tree.CurrentIndex
    }
    if tree.CurrentAtTop < 0 {
        tree.CurrentAtTop = 0
    }
    return nil
}

func winClampPrintln(w *gnc.Window, s string, limit int) {
    if len(s) > limit {
        w.Print(s[0:limit - 1] + "\n")
    } else {
        w.Print(s + "\n")
    }
}
