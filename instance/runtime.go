package instance

import (
	"github.com/StructsNotClasses/mim/windowwriter"

	"github.com/d5/tengo/v2"
	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"errors"
	"strings"
	"io/ioutil"
    "path/filepath"
)

func (instance *Instance) Run() {
	for !instance.commandHandling.state.exit {
		//instance.writer.UpdateWindow()

		// check if there's a notification of playback state
		instance.playbackState.Receive(instance.notifier)

		// if no song is playing, run the script that has been dedicated to afformentioned scenario
		if !instance.playbackState.PlaybackInProgress && len(instance.commandHandling.state.runOnNoPlayback.contents) > 0 {
			instance.runScript(instance.commandHandling.state.runOnNoPlayback)
		}
		if userByte := instance.GetKey(); userByte != 0 {
			instance.HandleKey(userByte)
		}
	}
}

func (instance *Instance) HandleKey(userByte gnc.Key) {
	// if in character mode, run the script bound to the received key and skip the rest
	// (currently the colon key could be bound to something haha)
	if script, ok := instance.commandHandling.bindMap[userByte]; ok && instance.commandHandling.state.mode == CharacterMode {
		instance.runScript(script)
		return
	}

	// use the colon key to unset character mode instantly
	if userByte == ':' {
		instance.commandHandling.state.mode = CommandMode
	}

	if userByte == 263 {
		// if backspace remove last byte from slice
		instance.commandHandling.state.line = pop(instance.commandHandling.state.line)
	} else {
		// for any other character add it to the line buffer
		instance.commandHandling.state.line = append(instance.commandHandling.state.line, byte(userByte))
	}

    instance.commandHandling.UpdateInput(instance.commandHandling.state.line)
	if userByte == '\n' {
		line := string(instance.commandHandling.state.line)
		if command := strings.TrimPrefix(line, ":"); command != line {
			command = strings.TrimSuffix(command, "\n")
			instance.runCommand(command)
		} else {
			instance.commandHandling.state.lines = append(instance.commandHandling.state.lines, instance.commandHandling.state.line...)
		}
		//always clear the line buffer
		instance.commandHandling.state.line = []byte{}
	}
	instance.backgroundWindow.Refresh()
}

func (instance *Instance) runCommand(cmd string) {
	args := strings.Split(cmd, " ")
	// handle single line no argument commands
	switch args[0] {
	case "exit":
		instance.Exit()
	case "end":
        // this command marks the end of a script being written across multiple lines
        // optionally, the name of the script can be provided as an argument
        // if no name is provided, the script will be unnamed
        // takes the form
        // :end <name>?
        script := Script{
            name: "",
            contents: instance.commandHandling.state.lines,
        }
        if len(args) > 1 {
            script.name = args[1]
        }
		instance.manageScript(script)
		instance.commandHandling.state.lines = []byte{}
	case "on_no_playback":
		instance.commandHandling.state.onPlaybackBeingSet = true
	case "debug_print_buffer":
		instance.commandHandling.InfoPrintf("line: '%s'\nbytes: '%s'\n", cmd, string(instance.commandHandling.state.lines))
	case "char_mode":
		instance.commandHandling.state.mode = CharacterMode
	case "load_script":
		bytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			instance.commandHandling.InfoPrintf("load: Failed to load file '%s' with error '%v'\n", args[1], err)
		} else {
            instance.manageScript(Script{
                name: strings.TrimSuffix(filepath.Base(args[1]), ".tengo"),
                contents: bytes,
            })
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
			instance.commandHandling.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", args[1], err)
		}
	case "new_command":
		if len(args) != 3 {
			instance.commandHandling.InfoPrintln(":new_command: usage ':new_command <name> <config file>'")
		} else {
			instance.commandHandling.commandMap[args[1]] = args[2]
		}
	case "bind":
		// these commands should follow the format
		// :bind <character>
		// <script>
		// :end
		instance.commandHandling.state.bindChar = gnc.Key(args[1][0])
		instance.commandHandling.state.lines = []byte{}
	case "echo":
		// the echo command prints something to the info window. this can be useful for telling the user when a config was loaded and stuff
		// takes the form
		// :echo <message>
		message := strings.TrimPrefix(cmd, "echo ")
		instance.commandHandling.InfoPrintln(message)
	default:
		configFile, ok := instance.commandHandling.commandMap[args[0]]
		if ok {
			err := instance.LoadConfig(configFile)
			if err != nil {
				instance.commandHandling.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", configFile, err)
			}
		} else {
			instance.commandHandling.InfoPrintf("unknown command: '%s'\n", args[0])
		}
	}
}

func (instance *Instance) manageScript(script Script) {
    whatToPrint := string(script.contents)
    if script.name != "" {
        whatToPrint = script.name
    }

	if instance.commandHandling.state.bindChar != 0 {
        instance.commandHandling.InfoPrintf("Binding script: %s to character '%c'\n", whatToPrint, instance.commandHandling.state.bindChar)
		instance.commandHandling.bindMap[instance.commandHandling.state.bindChar] = script
		instance.commandHandling.state.bindChar = 0
	} else if instance.commandHandling.state.onPlaybackBeingSet {
		instance.commandHandling.InfoPrintln("Setting script to run when no songs are playing: " + whatToPrint)
		instance.commandHandling.state.runOnNoPlayback = script
		instance.commandHandling.state.onPlaybackBeingSet = false
	} else {
		instance.runScript(script)
	}
}

func (i *Instance) runScript(s Script) {
    if s.name != "" {
        i.commandHandling.InfoPrintln("Running script: " + s.name)
    } else {
        i.commandHandling.InfoPrint("Running script: " + string(s.contents))
    }

	script := tengo.NewScript(s.contents)
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
	script.Add("isExpanded", i.TengoIsExpanded)
	script.Add("itemCount", i.TengoItemCount)

	bytecode, err := script.Compile()
	if err != nil {
		i.commandHandling.InfoPrintln(err)
    } else {
        defer i.commandHandling.InfoPrintRuntimeError()
        if err := bytecode.Run(); err != nil {
            i.commandHandling.InfoPrintln(err)
        }
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
	i.tree.Draw()

	i.currentRemote = playFileWithMplayer(i.tree.CurrentEntry().Path, i.notifier, windowwriter.New(i.mpOutputWindow))

	//wait for the above function to send a signal that playback began
	i.playbackState.ReceiveBlocking(i.notifier)
	return nil
}
