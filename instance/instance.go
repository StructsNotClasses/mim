package instance

import (
	"github.com/StructsNotClasses/mim/instance/dirtree"
	"github.com/StructsNotClasses/mim/instance/notification"
	"github.com/StructsNotClasses/mim/instance/playback"
	"github.com/StructsNotClasses/mim/instance/client"
	"github.com/StructsNotClasses/mim/musicarray"
	"github.com/StructsNotClasses/mim/remote"

	"github.com/d5/tengo/v2"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const INT32MAX = 2147483647

type Script struct {
    name string
    contents []byte
    bytecode *tengo.Compiled
}

type InputMode int

const (
	CommandMode InputMode = iota
	CharacterMode
)

type CommandState struct {
    // buffer
	line               []byte
	lines              []byte

    runOnNoPlayback Script

    currentSearch string
    
    // mode settings
	bindChar           gnc.Key
	mode               InputMode
    commandBeingWritten bool
    scriptBeingWritten bool
	onPlaybackBeingSet bool
	exit               bool
}

type CommandInterface struct {
    inWin *gnc.Window
    outWin *gnc.Window
    state           CommandState
    bindMap         map[gnc.Key]Script
    commandMap      map[string]string
    aliasMap        map[string]string
}

type MplayerPlayer struct {
    // mplayer management
	currentRemote   remote.Remote
	playbackState   playback.PlaybackState
	notifier        chan notification.Notification
    mpOutputWindow *gnc.Window
}

type Instance struct {
    backgroundWindow *gnc.Window
	tree            dirtree.DirTree
    terminal    CommandInterface
    mp MplayerPlayer
}

func New(scr *gnc.Window, array musicarray.MusicArray) Instance {
	if len(array) > INT32MAX {
		log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time.", INT32MAX))
	}
	rand.Seed(time.Now().UnixNano())

	//make user input non-blocking
	scr.Timeout(0)

	var instance Instance

    var mpwin, treewin, inwin, outwin *gnc.Window
    var err error
	instance.backgroundWindow, mpwin, treewin, inwin, outwin, err = client.New(scr)
	if err != nil {
		log.Fatal(err)
	}

	instance.tree = dirtree.New(treewin, array)

    instance.terminal = CommandInterface{
        inWin: inwin,
        outWin: outwin,
        state: CommandState{
            line: []byte{},
            lines: []byte{},
            runOnNoPlayback: Script{},
            currentSearch: "",
            bindChar:           0,
            mode:               CommandMode,
            commandBeingWritten: false,
            scriptBeingWritten: false,
            onPlaybackBeingSet: false,
            exit:               false,
        },
        bindMap: make(map[gnc.Key]Script),
        commandMap: make(map[string]string),
        aliasMap: make(map[string]string),
    }

    instance.mp = MplayerPlayer{
        currentRemote: remote.Remote{},
        notifier: make(chan notification.Notification),
        mpOutputWindow: mpwin,
    }

	return instance
}

func (instance *Instance) LoadConfig(filename string) error {
	instance.terminal.state.line = []byte{}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, b := range bytes {
		instance.HandleKey(gnc.Key(b))
	}
	return nil
}

func (instance *Instance) Exit() {
	instance.terminal.state.exit = true
	instance.mp.StopPlayback()
}

func (mp *MplayerPlayer) StopPlayback() {
	if mp.playbackState.PlaybackInProgress {
		mp.currentRemote.SendString("quit\n")
		mp.playbackState.ReceiveBlocking(mp.notifier)
	}
}

func (i Instance) GetKey() gnc.Key {
	return i.backgroundWindow.GetChar()
}

func (i Instance) GetLineBlocking() string {
    // set blocking
    i.backgroundWindow.Timeout(-1)
    line := ""
    ch := i.backgroundWindow.GetChar()
    for ; ch != '\n'; ch = i.backgroundWindow.GetChar() {
        line = fmt.Sprintf("%s%c", line, rune(ch))
        i.terminal.InfoPrintf("%c", rune(ch))
    }
    i.terminal.InfoPrintln()

    // unset blocking
    i.backgroundWindow.Timeout(0)
    return line
}

func (i Instance) GetCharBlocking() rune {
    i.backgroundWindow.Timeout(-1)
    ch := i.backgroundWindow.GetChar()
    i.backgroundWindow.Timeout(0)
    return rune(ch)
}

// replaceCurrentLine erases the current line on the window and prints a new one
// the new string's byte array can contain a newline, which means this can replace the line with multiple lines
func replaceCurrentLine(win *gnc.Window, bs []byte) {
	s := string(bs)
	y, _ := win.CursorYX()
	_, w := win.MaxYX()
	win.HLine(y, 0, ' ', w)
	win.MovePrint(y, 0, s)
	win.Refresh()
}
