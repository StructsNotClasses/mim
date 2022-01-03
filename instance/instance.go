package instance

import (
	"github.com/StructsNotClasses/musicplayer/instance/notification"
	"github.com/StructsNotClasses/musicplayer/instance/playback"
	"github.com/StructsNotClasses/musicplayer/musicarray"
	"github.com/StructsNotClasses/musicplayer/remote"

	"github.com/d5/tengo/v2"
	gnc "github.com/rthornton128/goncurses"

	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const INT32MAX = 2147483647

type Script []byte

type InputMode int

const (
	CommandMode InputMode = iota
	CharacterMode
)

type UserStateMachine struct {
	line               []byte
	lines              []byte
	bindChar           gnc.Key
	onPlaybackBeingSet bool
	mode               InputMode
	exit               bool
}

type Instance struct {
	client          Client
	tree            DirTree
	currentRemote   remote.Remote
	bindMap         map[gnc.Key]Script
	runOnNoPlayback []byte
	state           UserStateMachine
	playbackState   playback.PlaybackState
	notifier        chan notification.Notification
}

func New(scr *gnc.Window, array musicarray.MusicArray) Instance {
	//make user input non-blocking
	scr.Timeout(0)

	if len(array) > INT32MAX {
		log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time.", INT32MAX))
	}

	rand.Seed(time.Now().UnixNano())

	var instance Instance
	client, err := createClient(scr)
	if err != nil {
		log.Fatal(err)
	}
	instance.client = client
	instance.tree = DirTree{
		currentIndex: 0,
		array:        array,
	}
	instance.currentRemote = remote.Remote{}

	instance.bindMap = make(map[gnc.Key]Script)
	instance.notifier = make(chan notification.Notification)

	instance.state = UserStateMachine{
		bindChar:           0,
		onPlaybackBeingSet: false,
		mode:               CommandMode,
		exit:               false,
	}

	return instance
}

func (instance *Instance) GetKey() gnc.Key {
	return instance.client.backgroundWindow.GetChar()
}

func (instance *Instance) LoadConfig(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, b := range bytes {
		instance.HandleKey(gnc.Key(b))
	}
	return nil
}

func (instance *Instance) Run() {
	for !instance.state.exit {
		// check if there's a notification of playback state
		instance.playbackState.Receive(instance.notifier)

		// if no song is playing, run the script that has been dedicated to afformentioned scenario
		if !instance.playbackState.PlaybackInProgress && len(instance.runOnNoPlayback) > 0 {
			instance.runBytesAsScript(instance.runOnNoPlayback)
		}
		if userByte := instance.GetKey(); userByte != 0 {
			instance.HandleKey(userByte)
		}
	}
}

func (instance *Instance) HandleKey(userByte gnc.Key) {
	// if in character mode, run the script bound to the received key and skip the rest
	// (currently the colon key could be bound to something haha)
	if script, ok := instance.bindMap[userByte]; ok && instance.state.mode == CharacterMode {
		instance.runBytesAsScript(script)
		return
	}

	// use the colon key to unset character mode instantly
	if userByte == ':' {
		instance.state.mode = CommandMode
	}

	if userByte == 263 {
		// if backspace remove last byte from slice
		instance.state.line = pop(instance.state.line)
	} else {
		// for any other character add it to the line buffer
		instance.state.line = append(instance.state.line, byte(userByte))
	}

	replaceCurrentLine(instance.client.commandInputWindow, instance.state.line)
	if userByte == '\n' {
		// handle single line no argument commands
		switch string(instance.state.line) {
		case ":exit\n":
			instance.Exit()
		case ":end\n":
			instance.manageByteScript(instance.state.lines)
			instance.state.lines = []byte{}
		case ":on_no_playback\n":
			instance.state.onPlaybackBeingSet = true
		case "debug_print_buffer\n":
			instance.client.commandOutputWindow.Printf("line: '%s'\nbytes: '%s'\n", string(instance.state.line), string(instance.state.lines))
			instance.client.commandOutputWindow.Refresh()
		case ":char_mode\n":
			instance.state.mode = CharacterMode
		default:
			// handle single line commands with arguments
			switch string(instance.state.line[0:5]) {
			case ":load":
				fileName := string(instance.state.line[6 : len(instance.state.line)-1])
				bytes, err := ioutil.ReadFile(fileName)
				if err != nil {
					instance.client.commandOutputWindow.Printf("load: Failed to load file '%s' with error '%v'\n", fileName, err)
					instance.client.commandOutputWindow.Refresh()
				} else {
					instance.manageByteScript(bytes)
				}
			case ":bind":
				// these commands should follow the format
				// :bind <character>
				// <script>
				// :end
				instance.state.bindChar = gnc.Key(instance.state.line[6])
				instance.state.lines = []byte{}
			default:
				instance.state.lines = append(instance.state.lines, instance.state.line...)
			}
		}
		//always clear the line buffer
		instance.state.line = []byte{}
	}
	instance.client.backgroundWindow.Refresh()
}

func (instance *Instance) Exit() {
	instance.state.exit = true
	instance.StopPlayback()
}

func (instance *Instance) manageByteScript(script []byte) {
	if instance.state.bindChar != 0 {
		instance.client.commandOutputWindow.Printf("Binding script: %s to character %c\n", string(script), instance.state.bindChar)
		instance.client.commandOutputWindow.Refresh()
		instance.bindMap[instance.state.bindChar] = script
		instance.state.bindChar = 0
	} else if instance.state.onPlaybackBeingSet {
		instance.client.commandOutputWindow.Println("Setting script to run when no songs are playing: " + string(script))
		instance.client.commandOutputWindow.Refresh()
		instance.runOnNoPlayback = script
		instance.state.onPlaybackBeingSet = false
	} else {
		instance.runBytesAsScript(script)
	}
}

func (i *Instance) runBytesAsScript(bs []byte) {
	i.client.commandOutputWindow.Print("Running script: " + string(bs))
	i.client.commandOutputWindow.Refresh()

	outwin := i.client.commandOutputWindow

	script := tengo.NewScript(bs)
	script.Add("send", i.TengoSend)
	script.Add("selectIndex", i.TengoSelectIndex)
	script.Add("playSelected", i.TengoPlaySelected)
	script.Add("playIndex", i.TengoPlayIndex)
	script.Add("songCount", i.TengoSongCount)
	script.Add("infoPrint", i.TengoInfoPrint)
	script.Add("currentIndex", i.TengoCurrentIndex)
	script.Add("randomIndex", i.TengoRandomIndex)
	script.Add("selectUp", i.TengoSelectUp)
	script.Add("selectDown", i.TengoSelectDown)

	bytecode, err := script.Compile()
	if err != nil {
		outwin.Println(err)
	} else {
		outwin.Println("Successful compilation!\n")
	}
	outwin.Refresh()

	defer windowPrintRuntimeError(outwin)
	if err := bytecode.Run(); err != nil {
		outwin.Println(err)
		outwin.Refresh()
	}
}

func (i *Instance) PlayIndex(index int) error {
	i.StopPlayback()
	if index < 0 || index >= len(i.tree.array) {
		return errors.New(fmt.Sprintf("musicplayer: index out of range %v", index))
	}
	i.tree.Select(index)
	i.tree.Draw(i.client.treeWindow)

	i.currentRemote = playFileWithMplayer(i.tree.array[i.tree.currentIndex].Path, i.notifier, i.client.infoWindow)
	//wait for the above function to send a signal that playback began
	i.playbackState.ReceiveBlocking(i.notifier)
	return nil
}

func (i *Instance) StopPlayback() {
	if i.playbackState.PlaybackInProgress {
		i.currentRemote.SendString("quit\n")
		i.playbackState.ReceiveBlocking(i.notifier)
	}
}
