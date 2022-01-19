package instance

import (
	"github.com/StructsNotClasses/mim/instance/dirtree"
	"github.com/StructsNotClasses/mim/instance/playback"
	"github.com/StructsNotClasses/mim/instance/terminal"
	"github.com/StructsNotClasses/mim/musicarray"
	"github.com/StructsNotClasses/mim/remote"
	"github.com/StructsNotClasses/mim/windowwriter"

	gnc "github.com/rthornton128/goncurses"

	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

type MplayerPlayer struct {
	// mplayer management
	currentRemote  remote.Remote
	playbackState  playback.PlaybackState
	notifier       chan playback.Notification
	mpOutputWindow *gnc.Window
}

type Instance struct {
	bg *gnc.Window
	tree             dirtree.DirTree
	terminal         terminal.Terminal
	mp               MplayerPlayer
}

func New(scr *gnc.Window, musicDirectory string) (Instance, error) {
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
	const int32Max = 2147483647
	if len(arr) > int32Max {
		return Instance{}, errors.New(fmt.Sprintf("mim currently does not support playback of more than %d songs and directories at a time.", int32Max))
	}

    // create windows
	var bgwin, mpwin, treewin, inwin, outwin *gnc.Window
	bgwin, mpwin, treewin, inwin, outwin, err = CreateWindows(scr)
	if err != nil {
		return Instance{}, err
	}

	return Instance{
		bg: bgwin,
		tree:             dirtree.New(treewin, arr),
		terminal:         terminal.New(inwin, outwin),
		mp: MplayerPlayer{
			currentRemote:  remote.Remote{},
			notifier:       make(chan playback.Notification),
			mpOutputWindow: mpwin,
		},
	}, nil
}

func (instance *Instance) PassFileToInput(filename string) (bool, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, err
	}
	for _, ch := range string(bytes) {
        instance.terminal.InputCharacter(ch)
        if ch == '\n' && instance.HandleNewline() {
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

func CreateWindows(scr *gnc.Window) (backgroundWindow *gnc.Window, infoWindow *gnc.Window, treeWindow *gnc.Window, commandInputWindow *gnc.Window, commandOutputWindow *gnc.Window, err error) {
	totalHeight, totalWidth := scr.MaxYX()
	leftToRightRatio := 2.0/3.0

	var terminalWidth int = int(leftToRightRatio*float64(totalWidth)) - 2
	var treeWidth int = totalWidth - terminalWidth - 3

    var mpHeight int = (totalHeight - 4)/3
    var inputHeight int = (totalHeight - mpHeight - 4)/2
    var outputHeight int = totalHeight - mpHeight - inputHeight - 4

    backgroundWindow = scr

	//create the window that displays information about the current song
	infoWindow, err = gnc.NewWindow(mpHeight, terminalWidth, 1, 1)
	if err != nil {
		return
	}
	infoWindow.ScrollOk(true)

	//create the window that holds the song tree
	treeWindow, err = gnc.NewWindow(totalHeight-2, treeWidth, 1, terminalWidth+2)
	if err != nil {
		return
	}

	//create the window that allows user input
	commandInputWindow, err = gnc.NewWindow(inputHeight, terminalWidth, mpHeight + outputHeight + 3, 1)
	if err != nil {
		return 
	}
	commandInputWindow.ScrollOk(true)

	//create the window that holds command output
	commandOutputWindow, err = gnc.NewWindow(outputHeight, terminalWidth, mpHeight+2, 1)
	if err != nil {
		return
	}
	commandOutputWindow.ScrollOk(true)

	scr.Box('|', '-')
    scr.VLine(1, terminalWidth+1, '|', totalHeight-2)
	scr.HLine(mpHeight + 1, 1, '=', terminalWidth)
	scr.HLine(mpHeight+outputHeight+2, 1, '=', terminalWidth)
	scr.Refresh()

	return
}
