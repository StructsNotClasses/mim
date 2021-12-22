package main

import (
	"github.com/StructsNotClasses/musicplayer/remote"
	"github.com/StructsNotClasses/musicplayer/song"

	"github.com/d5/tengo/v2"
	gnc "github.com/rthornton128/goncurses"

	"errors"
	"fmt"
	"log"
    "math/rand"
    "time"
)

type Client struct {
	backgroundWindow    *gnc.Window
	infoWindow          *gnc.Window
	treeWindow          *gnc.Window
	commandInputWindow  *gnc.Window
	commandOutputWindow *gnc.Window
}

type Instance struct {
	client             Client
	tree               song.Tree
	currentRemote      remote.Remote
	notifier           chan int
	playbackInProgress bool
}

func createClient(scr *gnc.Window) (Client, error) {
	totalHeight, totalWidth := scr.MaxYX()
	leftToRightRatio := 0.6
	COMMAND_INPUT_HEIGHT := 6
	COMMAND_OUTPUT_HEIGHT := 10
	BORDER_WIDTH := 1

	var topWindowHeight int = totalHeight - COMMAND_INPUT_HEIGHT - COMMAND_OUTPUT_HEIGHT - 4*BORDER_WIDTH
	var infoWindowWidth int = int(leftToRightRatio*float64(totalWidth)) - 2*BORDER_WIDTH
	var commandWindowsWidth int = infoWindowWidth
	var treeWindowWidth int = totalWidth - infoWindowWidth - 3*BORDER_WIDTH

	//create the window that displays information about the current song
	infoWindow, err := gnc.NewWindow(topWindowHeight, infoWindowWidth, BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	infoWindow.ScrollOk(true)

	//create the window that holds the song tree
	treeWindow, err := gnc.NewWindow(totalHeight-2*BORDER_WIDTH, treeWindowWidth, BORDER_WIDTH, infoWindowWidth+2*BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}

	//create the window that allows user input
	commandInputWindow, err := gnc.NewWindow(COMMAND_INPUT_HEIGHT, commandWindowsWidth, topWindowHeight+COMMAND_OUTPUT_HEIGHT+3*BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	commandInputWindow.ScrollOk(true)

	//create the window that holds command output
	commandOutputWindow, err := gnc.NewWindow(COMMAND_OUTPUT_HEIGHT, commandWindowsWidth, topWindowHeight+2*BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	commandOutputWindow.ScrollOk(true)

	scr.Box('|', '-')
	scr.VLine(1, infoWindowWidth+BORDER_WIDTH, '|', totalHeight-2*BORDER_WIDTH)
	scr.HLine(topWindowHeight+BORDER_WIDTH, 1, '=', infoWindowWidth)
	scr.HLine(topWindowHeight+COMMAND_OUTPUT_HEIGHT+2*BORDER_WIDTH, 1, '=', infoWindowWidth)
	scr.Refresh()

	return Client{
		scr,
		infoWindow,
		treeWindow,
		commandInputWindow,
		commandOutputWindow,
	}, nil
}

func createInstance(scr *gnc.Window, songs song.List) Instance {
	if len(songs) > INT32MAX {
		log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time.", INT32MAX))
	}

	rand.Seed(time.Now().UnixNano())

	var instance Instance
	client, err := createClient(scr)
	if err != nil {
		log.Fatal(err)
	}
	instance.client = client
	instance.tree = song.Tree{
		Songs:        songs,
		CurrentIndex: 0,
		CurrentAtTop: 0,
	}
	instance.currentRemote = remote.Remote{}
	instance.notifier = make(chan int)
	return instance
}

func (i *Instance) runBytesAsScript(bs []byte) {
	outwin := i.client.commandOutputWindow

	script := tengo.NewScript(bs)
	script.Add("send", i.TengoSend)
	script.Add("playIndex", i.TengoPlayIndex)
    script.Add("songCount", i.TengoSongCount)
    script.Add("infoPrint", i.TengoInfoPrint)
    script.Add("currentIndex", i.TengoCurrentIndex)
    script.Add("randomIndex", i.TengoRandomIndex)

	bytecode, err := script.Compile()
	if err != nil {
		outwin.Println(err)
	} else {
		outwin.Println("Successful compilation!\n")
	}
	outwin.Refresh()

	defer handleRuntimeError(outwin)
	if err := bytecode.Run(); err != nil {
		outwin.Println(err)
		outwin.Refresh()
	}
}

func (i *Instance) TengoSend(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if s, ok := args[0].(*tengo.String); ok {
		asString := s.String()
		cmdString := asString[1:len(asString)-1] + "\n" // tengo ".String()" returns a string value surrounded by quotes, so this needs to remove them before sending
		i.currentRemote.SendString(cmdString)
		return nil, nil
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "string argument",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
}

func (i *Instance) TengoPlayIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if value, ok := args[0].(*tengo.Int); ok {
		err := i.PlayIndex(int(value.Value))
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Instance) TengoSongCount(args ...tengo.Object) (tengo.Object, error) {
    if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
    }
    return &tengo.Int{Value: int64(len(i.tree.Songs))}, nil
}

func (i *Instance) TengoInfoPrint(args ...tengo.Object) (tengo.Object, error) {
    for _, item := range args {
        if value, ok := item.(*tengo.String); ok {
            i.client.commandOutputWindow.Print(value) 
            i.client.commandOutputWindow.Refresh()
        }
    }
    return nil, nil
}

func (i *Instance) TengoCurrentIndex(args ...tengo.Object) (tengo.Object, error) {
    if len(args) != 0 {
        return nil, tengo.ErrWrongNumArguments
    }
    return &tengo.Int{Value: int64(i.tree.CurrentIndex)}, nil
}

// TengoRandomIndex returns a random number that is a valid song index. It requires random to already be seeded.
func (i *Instance) TengoRandomIndex(args ...tengo.Object) (tengo.Object, error) {
    rnum := rand.Int31n(int32(len(i.tree.Songs)))
    return &tengo.Int{Value: int64(rnum)}, nil
}

func (i *Instance) PlayIndex(index int) error {
	i.StopPlayback()
	if index < 0 || index >= len(i.tree.Songs) {
		return errors.New(fmt.Sprintf("musicplayer: song index out of range %v", index))
	}
	i.tree.Select(int32(index), i.client.treeWindow)
	i.tree.Draw(i.client.treeWindow)
	i.currentRemote = playFileWithMplayer(i.tree.Songs[i.tree.CurrentIndex].Path, i.notifier, i.client.infoWindow)
	i.playbackInProgress = true
	return nil
}

func (i *Instance) StopPlayback() {
	if i.playbackInProgress {
		i.currentRemote.SendString("quit\n")
		<-i.notifier
		i.playbackInProgress = false
	}
}
