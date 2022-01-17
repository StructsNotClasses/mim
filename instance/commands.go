package instance

import (
	"errors"
	"strings"
	"io/ioutil"
    "path/filepath"
)

func (instance *Instance) runCommand(cmd string) bool {
	args, err := splitCommand(cmd)
    if err != nil {
        instance.terminal.InfoPrintln(err)
        return false
    }

	switch args[0] {
	case "exit":
        // exit the program
        // :exit
        instance.mp.StopPlayback()
        return true
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
            return false
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
		shouldExit, err := instance.PassFileToInput(args[1])
		if err != nil {
			instance.terminal.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", args[1], err)
		}
        return shouldExit
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
                instance.terminal.state.bindChar = []rune(args[1])[0]
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
			shouldExit, err := instance.PassFileToInput(configFile)
			if err != nil {
				instance.terminal.InfoPrintf("load: Failed to load config '%s' with error '%v'\n", configFile, err)
			}
            return shouldExit
		} 

        // alias
        aliased, ok := instance.terminal.aliasMap[args[0]]
        if ok {
            if len(args) > 1 {
                fullCommand := strings.Join(append([]string{aliased}, args[1:]...), " ")
                return instance.runCommand(fullCommand)
            } else {
                return instance.runCommand(aliased)
            }
        }
        
        instance.terminal.InfoPrintf("Unknown command: '%s'\n", args[0])
	}
    return false
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
