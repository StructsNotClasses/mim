package instance

import (
	"github.com/StructsNotClasses/mim/script"

	"github.com/d5/tengo/v2"
)

func (i *Instance) Run() {
	for shouldExit := false; !shouldExit; {
		// check if there's a notification of playback state
		i.mp.playbackState.Receive(i.mp.notifier)

		// if no song is playing, run the so dedicated script
		if !i.mp.playbackState.PlaybackInProgress {
			i.terminal.TryRunNoPlaybackScript()
		}

		// process any new user input
		if ch := i.GetCharNonBlocking(); ch != 0 {
            i.terminal.InputCharacter(ch)
            if ch == '\n' {
                shouldExit = i.HandleNewline()
            }
		}
	}
}

func (i *Instance) HandleNewline() bool {
    if i.terminal.CommandBeingWritten() {
        cmd := i.terminal.CurrentLine()
        i.terminal.EndCommand()
        return i.runCommand(cmd)
    } else if i.terminal.ScriptBeingWritten() {
        i.terminal.PushLineToBuffer()
    } else if len(i.terminal.CurrentLine()) != 1 {
        i.terminal.InfoPrintf("Error: Non-command input '%s' entered before a begin command.\n", i.terminal.CurrentLine())
        i.terminal.ClearLine()
    } else {
        i.terminal.ClearLine()
    }
    return false
}

func (instance *Instance) manageScript(s script.Script) {
	if !instance.terminal.NextScriptShouldBeBound() && !instance.terminal.NextScriptIsNoPlayback() {
		instance.terminal.RunScript(s)
	} else {
		if instance.terminal.NextScriptShouldBeBound() {
			instance.terminal.InfoPrintf("Binding script: %s to character '%c'\n", s.Name(), instance.terminal.Binding())
			instance.terminal.BindCurrentToScript(s)
		}
		if instance.terminal.NextScriptIsNoPlayback() {
			instance.terminal.InfoPrintln("Setting script to run when no songs are playing: " + s.Name())
			instance.terminal.SetNoPlayback(s)
		}
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
