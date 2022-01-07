package instance

import (
	"github.com/d5/tengo/v2"

	gnc "github.com/rthornton128/goncurses"

	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

func (instance *Instance) Run() {
	for !instance.state.exit {
		//instance.writer.UpdateWindow()

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
		line := string(instance.state.line)
		if command := strings.TrimPrefix(line, ":"); command != line {
			command = strings.TrimSuffix(command, "\n")
			instance.runCommand(command)
		} else {
			instance.state.lines = append(instance.state.lines, instance.state.line...)
		}
		//always clear the line buffer
		instance.state.line = []byte{}
	}
	instance.client.backgroundWindow.Refresh()
}

func (instance *Instance) runCommand(cmd string) {
	args := strings.Split(cmd, " ")
	// handle single line no argument commands
	switch args[0] {
	case "exit":
		instance.Exit()
	case "end":
		instance.manageByteScript(instance.state.lines)
		instance.state.lines = []byte{}
	case "on_no_playback":
		instance.state.onPlaybackBeingSet = true
	case "debug_print_buffer":
		instance.client.commandOutputWindow.Printf("line: '%s'\nbytes: '%s'\n", cmd, string(instance.state.lines))
		instance.client.commandOutputWindow.Refresh()
	case "char_mode":
		instance.state.mode = CharacterMode
	case "load_script":
		bytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			instance.client.commandOutputWindow.Printf("load: Failed to load file '%s' with error '%v'\n", args[1], err)
			instance.client.commandOutputWindow.Refresh()
		} else {
			instance.manageByteScript(bytes)
		}
	case "load_config":
		// this command can be thought of as replacing itself with the commands in its argument
		// for example, if shuffle.mim's contents are ':load_script shuffle.tengo' then
		// :load_config shuffle.mim (should) equal :load_script shuffle.tengo
		// this works recursively
		// takes the form
		// :load_config <filename>
		err := instance.LoadConfig(args[1])
		if err != nil {
			instance.client.commandOutputWindow.Printf("load: Failed to load config '%s' with error '%v'\n", args[1], err)
			instance.client.commandOutputWindow.Refresh()
		}
	case "new_command":
		if len(args) != 3 {
			instance.client.commandOutputWindow.Println(":new_command: usage ':new_command <name> <config file>'")
			instance.client.commandOutputWindow.Refresh()
		} else {
			instance.commandMap[args[1]] = args[2]
		}
	case "bind":
		// these commands should follow the format
		// :bind <character>
		// <script>
		// :end
		instance.state.bindChar = gnc.Key(args[1][0])
		instance.state.lines = []byte{}
	case "echo":
		// the echo command prints something to the info window. this can be useful for telling the user when a config was loaded and stuff
		// takes the form
		// :echo <message>
		message := strings.TrimPrefix(cmd, "echo ")
		instance.client.commandOutputWindow.Println(message)
		instance.client.commandOutputWindow.Refresh()
	default:
		configFile, ok := instance.commandMap[args[0]]
		if ok {
			err := instance.LoadConfig(configFile)
			if err != nil {
				instance.client.commandOutputWindow.Printf("load: Failed to load config '%s' with error '%v'\n", configFile, err)
				instance.client.commandOutputWindow.Refresh()
			}
		} else {
			instance.client.commandOutputWindow.Printf("unknown command: '%s'\n", args[0])
		}
	}
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
	script.Add("selectEnclosing", i.TengoSelectEnclosing)
	script.Add("toggle", i.TengoToggleDirExpansion)
	script.Add("isDir", i.TengoIsDir)
	script.Add("selectedIsDir", i.TengoSelectedIsDir)
	script.Add("depth", i.TengoDepth)

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
	if !i.tree.IsInRange(index) {
		return errors.New(fmt.Sprintf("instance.PlayIndex: index out of range %v.", index))
	}
	if i.tree.IsDir(index) {
		return errors.New(fmt.Sprintf("instance.PlayIndex: directories cannot be played"))
	}
	i.tree.Select(index)
	i.tree.Draw(i.client.treeWindow)

	i.currentRemote = playFileWithMplayer(i.tree.CurrentEntry().Path, i.notifier, i.client.infoWindow)

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
