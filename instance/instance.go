package instance

import (
	"github.com/StructsNotClasses/mim/instance/dirtree"
	"github.com/StructsNotClasses/mim/instance/notification"
	"github.com/StructsNotClasses/mim/instance/playback"
	"github.com/StructsNotClasses/mim/instance/client"
	"github.com/StructsNotClasses/mim/musicarray"
	"github.com/StructsNotClasses/mim/windowwriter"
	"github.com/StructsNotClasses/mim/remote"

	"github.com/d5/tengo/v2"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io/ioutil"
    "errors"
	"math/rand"
	"time"
)

type Script struct {
    name string
    contents []byte
    bytecode *tengo.Compiled
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
    terminal    Terminal
    mp MplayerPlayer
}

func New(scr *gnc.Window, musicDirectory string) (Instance, error) {
    const int32Max = 2147483647

    var bgwin, mpwin, treewin, inwin, outwin *gnc.Window

    // seed random
	rand.Seed(time.Now().UnixNano())

	// make user input non-blocking
	scr.Timeout(0)

    // create the array for the music tree
	arr, err := musicarray.New(musicDirectory)
	if err != nil {
        return Instance{}, err
	}

    // random number generation currently produces an int32, so limit the array length to its max
	if len(arr) > int32Max {
        return Instance{}, errors.New(fmt.Sprintf("mim currently does not support playback of more than %d songs and directories at a time.", int32Max))
	}

	bgwin, mpwin, treewin, inwin, outwin, err = client.New(scr)
	if err != nil {
        return Instance{}, err
	}

    return Instance{
        backgroundWindow: bgwin,
        tree: dirtree.New(treewin, arr),
        terminal : Terminal{
            inWin: inwin,
            outWin: outwin,
            state: TerminalState{
                line: []byte{},
                lines: []byte{},
                runOnNoPlayback: Script{},
                currentSearch: "",
                bindChar:           0,
                mode:               CommandMode,
                commandBeingWritten: false,
                scriptBeingWritten: false,
                onPlaybackBeingSet: false,
            },
            bindMap: make(map[rune]Script),
            commandMap: make(map[string]string),
            aliasMap: make(map[string]string),
        },
        mp: MplayerPlayer{
            currentRemote: remote.Remote{},
            notifier: make(chan notification.Notification),
            mpOutputWindow: mpwin,
        },
    }, nil
}

func (instance *Instance) PassFileToInput(filename string) (bool, error) {
	instance.terminal.state.line = []byte{}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}
	for _, b := range bytes {
        if instance.HandleInput(rune(b)) {
            return true, nil
        }
	}
	return false, nil
}

func (i *Instance) PlayIndex(index int) error {
	i.mp.StopPlayback()
	if !i.tree.IsInRange(index) {
		return errors.New(fmt.Sprintf("instance.PlayIndex: index out of range %v.", index))
	}
	if i.tree.IsDir(index) {
		return errors.New(fmt.Sprintf("instance.PlayIndex: directories cannot be played"))
	}
	i.tree.Select(index)
	i.tree.Draw()

	i.mp.currentRemote = playFileWithMplayer(i.tree.CurrentEntry().Path, i.mp.notifier, windowwriter.New(i.mp.mpOutputWindow))

	//wait for the above function to send a signal that playback began
	i.mp.playbackState.ReceiveBlocking(i.mp.notifier)
	return nil
}

func (mp *MplayerPlayer) StopPlayback() {
	if mp.playbackState.PlaybackInProgress {
		mp.currentRemote.SendString("quit\n")
		mp.playbackState.ReceiveBlocking(mp.notifier)
	}
}
