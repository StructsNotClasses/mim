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
	for !instance.terminal.state.exit {
		//instance.writer.UpdateWindow()

		// check if there's a notification of playback state
		instance.mp.playbackState.Receive(instance.mp.notifier)

		// if no song is playing, run the script that has been dedicated to afformentioned scenario
		if !instance.mp.playbackState.PlaybackInProgress && len(instance.terminal.state.runOnNoPlayback.contents) > 0 {
			instance.terminal.runScript(instance.terminal.state.runOnNoPlayback)
		}
		if userByte := instance.GetKey(); userByte != 0 {
			instance.HandleKey(userByte)
		}
	}
}

func canBind(k gnc.Key) bool {
    const cantBind = ":\n"
    c := rune(k)
    for _, char := range(cantBind) {
        if c == char {
            return false
        }
    }
    return true
}

func (i *Instance) HandleKey(userByte gnc.Key) {
    // if a command is just starting to be entered then update the state accordingly
    if len(i.terminal.state.line) == 0 && userByte == ':' {
        i.terminal.state.commandBeingWritten = true
    // if no script is being written, no command is being written, and no command is just starting to be entered, run the script bound to the key
    } else if !i.terminal.state.scriptBeingWritten && !i.terminal.state.commandBeingWritten && canBind(userByte) {
        if script, ok := i.terminal.bindMap[userByte]; ok {
            i.terminal.runScript(script)
        } else {
            i.terminal.InfoPrintf("%c is not bound.\n", userByte)
        }
        return
    }

	if userByte == 263 {
		// backspace removes last byte from line buffer
        // if there is nothing but a colon, the user decided not to enter a command, so stop command entry mode
        if len(i.terminal.state.line) == 1 && i.terminal.state.line[0] == ':' {
            i.terminal.state.commandBeingWritten = false
        }
		i.terminal.state.line = pop(i.terminal.state.line)
	} else {
		// for any other character add it to the line buffer
		i.terminal.state.line = append(i.terminal.state.line, byte(userByte))
	}

    i.terminal.UpdateInput(i.terminal.state.line)
	if userByte == '\n' {
		line := string(i.terminal.state.line)
        if i.terminal.state.commandBeingWritten {
            command := line
            i.terminal.state.commandBeingWritten = false
			i.runCommand(command)
        } else if i.terminal.state.scriptBeingWritten {
			i.terminal.state.lines = append(i.terminal.state.lines, i.terminal.state.line...)
        } else if len(i.terminal.state.line) == 1 {
            // allow an empty line for clearing and formatting purposes
        } else {
            i.terminal.InfoPrintln("Error: Non-command (somehow) entered before a begin command.")
		}

        //always clear the line buffer
        i.terminal.state.line = []byte{}
    }
}

func (instance *Instance) runCommand(cmd string) {
	args, err := splitCommand(cmd)
    if err != nil {
        instance.terminal.InfoPrintln(err)
        return
    }

	switch args[0] {
	case "exit":
        // exit the program
        // :exit
		instance.Exit()
    case "begin":
        // marks the start of a script being written within the tui
        // normally, character input is redirected to keybinds, but after a begin and before a script finalization command, characters entered are put into a
        // buffer that is used by other commands
        // no arguments, naming and such is handled by finalization commands
        // clears the buffer
        // :begin
        instance.terminal.state.scriptBeingWritten = true
	case "end":
        // marks the end of a script being written within the tui 
        // optionally, the name of the script can be provided as an argument
        // if no name is provided, the script will be named its contents
        // be aware that this behaves differently based on the instance state
        //     if a character is being bound, the multiline script will be bound to the character, state will reflect that nothing is being bound, and the script will not be executed
        //     until the character is pressed
        //     the same goes for if the 'on_no_playback' script is being set. these two can occur simultaneously.
        //     if nothing was being set, the script will be executed once
        // this command triggers compilation of the script and always clears the buffer
        // :end <name>?
        if !instance.terminal.state.scriptBeingWritten {
            instance.terminal.InfoPrintln("end: called outside of script-writing environment")
            return
        }
        instance.terminal.state.scriptBeingWritten = false
        compiled, err := instance.compileScript(instance.terminal.state.lines)
        if err != nil {
            instance.terminal.InfoPrintln(err)
        } else {
            script := Script{
                name: "",
                contents: instance.terminal.state.lines,
                bytecode: compiled,
            }
            if len(args) > 1 {
                script.name = args[1]
            }
            instance.manageScript(script)
        }
        instance.terminal.state.lines = []byte{}
    case "cancel":
        // allows the user to discard their tui-written script without executing it
        // :cancel
        if instance.terminal.state.scriptBeingWritten {
            instance.terminal.state.lines = []byte{}
            instance.terminal.state.scriptBeingWritten = false
        } else {
            instance.terminal.InfoPrintln("cancel: cannot call outside of script-writing environment")
        }
	case "on_no_playback":
        // changes state such that the next script processed will be run whenever nothing is currently playing
        // to have no script run, simply provide an empty string as the script, such as (with an empty buffer)
        // eg :on_no_playback
        //    :end do_nothing
        // :on_no_playback
		if instance.terminal.RequireArgCount(args, 1) {
            instance.terminal.state.onPlaybackBeingSet = true
        }
	case "print_buffer":
        // print the current buffer contents to the info window
        // mostly for debugging or checking if empty or something
        // :print_buffer
		if instance.terminal.RequireArgCount(args, 1) {
            instance.terminal.InfoPrint(string(instance.terminal.state.lines))
        }
	case "load_script":
        // this (aims to) behave identically to doing
        // :begin
        // <contents of file>
        // :end <filename without extension>
        // refer to the details of :end for more info
        // this can't be called while writing a tengo script because it would unset the binding state for the enclosing script
        // :load_script <filename>
        if instance.terminal.state.scriptBeingWritten {
            instance.terminal.InfoPrintln("load_script: cannot call while writing a script.")
        } else if instance.terminal.RequireArgCount(args, 2) {
            bytes, err := ioutil.ReadFile(args[1])
            if err != nil {
                instance.terminal.InfoPrintf("load: Failed to load file '%s' with error '%v'\n", args[1], err)
            } else {
                compiled, err := instance.compileScript(bytes)
                if err != nil {
                    instance.terminal.InfoPrintln(err)
                } else {
                    instance.manageScript(Script{
                        name: strings.TrimSuffix(filepath.Base(args[1]), ".tengo"),
                        contents: bytes,
                        bytecode: compiled,
                    })
                }
            }
        }
	case "load_config":
		// this command can be thought of as replacing itself with the contents of its argument (a file)
		// for example, if shuffle.mim's contents are ':load_script shuffle.tengo' then
		// :load_config shuffle.mim (should) equal :load_script shuffle.tengo
		// this works recursively but there's no stack or anything so state changes within the loaded script are visible to the loader
		// takes the form
		// :load_config <filename>
		err := instance.LoadConfig(args[1])
		if err != nil {
			instance.terminal.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", args[1], err)
		}
	case "new_command":
        // creates a new command that runs the commands provided in a file
        // if you want to replace a current command with a new name, use :alias
        // takes the form
        // :new_command <name> <config file>
        if instance.terminal.RequireArgCount(args, 3) {
			instance.terminal.commandMap[args[1]] = args[2]
		}
	case "bind":
        // this tells the state that binding is occuring and to which character. only one character can be bound at a time.
        // it is silently used by other commands like :end or :load to bind <script> to a character INSTEAD of running it 
		// :bind <character>
        // <script>
        if instance.terminal.RequireArgCount(args, 2) {
            if len(args[1]) != 1 {
                instance.terminal.InfoPrintf("bind: %s is an invalid binding; only single character bindings are supported.", args[1])
            } else {
                instance.terminal.state.bindChar = gnc.Key(args[1][0])
            }
        }
	case "echo":
		// the echo command prints something to the info window. this can be useful for telling the user when a config was loaded and stuff
        // if no message is provided it still prints a newline
		// :echo <message>
        message := strings.TrimPrefix(cmd, ":echo ")
		instance.terminal.InfoPrintln(message)
    case "set_search":
        // sets instance state current search to the provided regexp
        // this doesn't do anything alone; the current search needs to be used first
        // :set_search <regexp>
        if instance.terminal.RequireArgCount(args, 2) {
            instance.terminal.state.currentSearch = args[1]
        }
    case "alias":
        // binds a command (and optionally some arguments) to a new name
        // when the new name is called, it will literally be replaced by the command it was bound to and run with the new arguments appended to the end
        // this allows one to alias only some of a command's arguments to a name and allow others to vary, eg :alias echo_error :echo "Error: "
        // :alias <name> <command> <argument>*
        if instance.terminal.RequireArgCountGTE(args, 3) {
            newName := args[1]
            command := strings.Join(args[2:], " ")
            instance.terminal.aliasMap[newName] = command
        }
	default:
        // user defined command
		configFile, ok := instance.terminal.commandMap[args[0]]
		if ok {
			err := instance.LoadConfig(configFile)
			if err != nil {
				instance.terminal.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", configFile, err)
			}
            return
		} 

        // alias
        aliased, ok := instance.terminal.aliasMap[args[0]]
        if ok {
            if len(args) > 1 {
                fullCommand := strings.Join(append([]string{aliased}, args[1:]...), " ")
                instance.runCommand(fullCommand)
            } else {
                instance.runCommand(aliased)
            }
            return
        }
        
        instance.terminal.InfoPrintf("Unknown command: '%s'\n", args[0])
	}
}

// splitCommand parses a command into its name and arguments
func splitCommand(cmd string) ([]string, error) {
    // current rules:
    // a command is :<name> (argument*)\n
    // an argument is <string start><stuff><whitespace>, <whitespace><stuff><whitespace>, <whitespace><stuff><string end>
    // where whitespace inside of a pair of quotation marks is considered stuff
    if len(cmd) == 0 {
        return []string{}, errors.New("Empty string cannot be a command.")
    }

    trimmed := strings.TrimPrefix(strings.TrimSuffix(cmd, "\n"), ":")

    args := []string{}
    insideString := false
    stringBegin := 0
    for i, r := range(trimmed) {
        if r == rune('"') {
            insideString = !insideString
        } else if r == rune(' ') && !insideString {
            args = append(args, trimmed[stringBegin:i])
            stringBegin = i + 1
        }
    }
    args = append(args, trimmed[stringBegin:len(trimmed)])
    if insideString {
        return args, errors.New("Unterminated string detected.")
    }

    return args, nil
}

func (instance *Instance) manageScript(script Script) {
    whatToPrint := string(script.contents)
    if script.name != "" {
        whatToPrint = script.name
    }

    bound := false
	if instance.terminal.state.bindChar != 0 {
        instance.terminal.InfoPrintf("Binding script: %s to character '%c'\n", whatToPrint, instance.terminal.state.bindChar)
		instance.terminal.bindMap[instance.terminal.state.bindChar] = script
		instance.terminal.state.bindChar = 0
        bound = true
	} 
    if instance.terminal.state.onPlaybackBeingSet {
		instance.terminal.InfoPrintln("Setting script to run when no songs are playing: " + whatToPrint)
		instance.terminal.state.runOnNoPlayback = script
		instance.terminal.state.onPlaybackBeingSet = false
        bound = true
	} 
    if !bound {
		instance.terminal.runScript(script)
	}
}

func (i *Instance) compileScript(bs []byte) (*tengo.Compiled, error) {
    script := tengo.NewScript(bs)
	script.Add("send", i.TengoSend)
	script.Add("selectIndex", i.TengoSelectIndex)
	script.Add("playSelected", i.TengoPlaySelected)
	script.Add("playIndex", i.TengoPlayIndex)
	script.Add("songCount", i.TengoSongCount)
	script.Add("infoPrint", i.TengoInfoPrint)
	script.Add("infoPrintln", i.TengoInfoPrintln)
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
	script.Add("setSearch", i.TengoSetSearch)
	script.Add("nextMatch", i.TengoNextMatch)
	script.Add("prevMatch", i.TengoPrevMatch)
    script.Add("getLine", i.TengoGetLine) 
    script.Add("getChar", i.TengoGetChar) 

	return script.Compile()
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
