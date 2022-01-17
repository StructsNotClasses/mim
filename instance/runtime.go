package instance

import (
	"github.com/d5/tengo/v2"
)

func (instance *Instance) Run() {
    for shouldExit := false; !shouldExit; {
		// check if there's a notification of playback state
		instance.mp.playbackState.Receive(instance.mp.notifier)

		// if no song is playing, run the script that has been dedicated to afformentioned scenario
		if !instance.mp.playbackState.PlaybackInProgress && len(instance.terminal.state.runOnNoPlayback.contents) > 0 {
			instance.terminal.runScript(instance.terminal.state.runOnNoPlayback)
		}

        // process any new user input
		if ch := instance.GetCharNonBlocking(); ch != 0 {
            shouldExit = instance.HandleInput(ch)
		}
	}
}

func (i *Instance) HandleInput(ch rune) bool {
    i.terminal.InputCharacter(ch)

	if ch == '\n' {
        if i.terminal.state.commandBeingWritten {
            i.terminal.state.commandBeingWritten = false
            cmd := string(i.terminal.state.line)
            i.terminal.state.line = []byte{}
            return i.runCommand(cmd)
        } else if i.terminal.state.scriptBeingWritten {
            i.terminal.AddInputToBuffer()
        } else if len(i.terminal.state.line) != 1 {
            i.terminal.InfoPrintln("Error: Non-command (somehow) entered before a begin command.")
		}

        //always clear the line buffer
        i.terminal.state.line = []byte{}
    }
    return false
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
