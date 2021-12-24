package instance

import (
	"github.com/StructsNotClasses/musicplayer/remote"
	"github.com/StructsNotClasses/musicplayer/safebool"
	"github.com/StructsNotClasses/musicplayer/song"
	"github.com/StructsNotClasses/musicplayer/windowwriter"

	"github.com/d5/tengo/v2"
	gnc "github.com/rthornton128/goncurses"

	"errors"
	"log"
	"time"
    "io/ioutil"
	"math/rand"
    "fmt"
    "os/exec"
    "io"
)

const INT32MAX = 2147483647

type Client struct {
	backgroundWindow    *gnc.Window
	infoWindow          *gnc.Window
	treeWindow          *gnc.Window
	commandInputWindow  *gnc.Window
	commandOutputWindow *gnc.Window
}

type Script []byte

type InputMode int

const (
    CommandMode InputMode = iota
    CharacterMode
)

type UserStateMachine struct {
    line []byte
    lines []byte
	bindChar           gnc.Key
	onPlaybackBeingSet bool
	mode InputMode
    exit bool
}

type Instance struct {
	client             Client
	tree               song.Tree
	currentRemote      remote.Remote
	playbackInProgress safebool.SafeBool
	bindMap            map[gnc.Key]Script
	runOnNoPlayback    []byte
	state              UserStateMachine
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

func New(scr *gnc.Window, songs song.List) Instance {
    //make user input non-blocking
    scr.Timeout(0)

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
	instance.playbackInProgress = safebool.New(false)

	instance.bindMap = make(map[gnc.Key]Script)

	instance.state = UserStateMachine{
		bindChar:           0,
		onPlaybackBeingSet: false,
        mode: CommandMode,
        exit: false,
	}

	return instance
}

func (instance *Instance) GetKey() gnc.Key {
    return instance.client.backgroundWindow.GetChar()
}

func (instance *Instance) Run() {
	for !instance.state.exit {
        // run the script for when no song is playing if neccesary
        if !instance.playbackInProgress.Get() && len(instance.runOnNoPlayback) > 0 {
            instance.client.commandOutputWindow.Print("Running script: " + string(instance.runOnNoPlayback))
            instance.client.commandOutputWindow.Refresh()
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
        instance.client.commandOutputWindow.Print("Running script: " + string(script))
        instance.client.commandOutputWindow.Refresh()
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
            instance.state.exit = true
            instance.currentRemote.SendString("quit\n")
        case ":end\n":
            instance.manageByteScript(instance.state.lines)
            instance.state.lines = []byte{}
        case ":on_no_playback\n":
            instance.state.onPlaybackBeingSet = true
        case "debug_print_buffer\n":
            instance.client.commandOutputWindow.Printf("line: '%s'\nbytes: '%s'\n", string(instance.state.line), string(instance.state.lines))
            instance.client.commandOutputWindow.Refresh()
        case ":character_mode\n":
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
		instance.client.commandOutputWindow.Println("Running script: " + string(script))
		instance.client.commandOutputWindow.Refresh()
		instance.runBytesAsScript(script)
	}
}

func (i *Instance) runBytesAsScript(bs []byte) {
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
			Name:     "'send' argument",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
}

func (i *Instance) TengoSelectIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if v, ok := args[0].(*tengo.Int); ok {
		asInt := v.Value
		i.tree.Select(int32(asInt), i.client.treeWindow)
		i.tree.Draw(i.client.treeWindow)
		return nil, nil
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "'selectIndex' argument",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}
}

func (i *Instance) TengoPlaySelected(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	err := i.PlayIndex(int(i.tree.CurrentIndex))
	return nil, err
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
	i.currentRemote = playFileWithMplayer(i.tree.Songs[i.tree.CurrentIndex].Path, &i.playbackInProgress, i.client.infoWindow)
	i.playbackInProgress.Set(true)
	return nil
}

func (i *Instance) StopPlayback() {
	if i.playbackInProgress.Get() {
		i.currentRemote.SendString("quit\n")
	}
}

func pop(bytes []byte) []byte {
	if len(bytes) <= 1 {
		return []byte{}
	} else {
		return bytes[:len(bytes)-1]
	}
}

// replaceCurrentLine erases the current line on the window and prints a new one
// the new string's byte array could potentially contain a newline, which means this can replace the line with multiple lines
func replaceCurrentLine(win *gnc.Window, bs []byte) {
    s := string(bs)
    y, _ := win.CursorYX()
	_, w := win.MaxYX()
	win.HLine(y, 0, ' ', w)
    win.MovePrint(y, 0, s)
    win.Refresh()
}

func handleRuntimeError(outputWindow *gnc.Window) {
	if runtimeError := recover(); runtimeError != nil {
        outputWindow.Print(fmt.Sprintf("\nruntime error: %s\n", runtimeError))
	}
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func playFileWithMplayer(file string, playbackInProgress *safebool.SafeBool, outWindow *gnc.Window) remote.Remote {
	cmd := exec.Command("mplayer",
		"-slave", "-vo", "null", "-quiet", file)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go runWithWriter(cmd, windowwriter.New(outWindow), playbackInProgress)

	return remote.Remote{pipe}
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, boolWrapper *safebool.SafeBool) { // notifier chan int) {
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	boolWrapper.Set(false)
	w.Close()
}
