package instance

import (
	gnc "github.com/rthornton128/goncurses"

    "fmt"
)

type InputMode int

const (
	CommandMode InputMode = iota
	CharacterMode
)

type TerminalState struct {
    // buffer
	line               []byte
	lines              []byte

    runOnNoPlayback Script

    currentSearch string
    
	bindChar           rune
	mode               InputMode
    commandBeingWritten bool
    scriptBeingWritten bool
	onPlaybackBeingSet bool
}

type Terminal struct {
    inWin *gnc.Window
    outWin *gnc.Window
    state           TerminalState
    bindMap         map[rune]Script
    commandMap      map[string]string
    aliasMap        map[string]string
}

func (term *Terminal) tryRunBinding(ch rune) bool {
    if term.state.scriptBeingWritten || term.state.commandBeingWritten || !canBind(ch) {
        return false
    }

    if script, ok := term.bindMap[ch]; ok {
        term.runScript(script)
    } else {
        term.InfoPrintf("%c is not bound.\n", ch)
    }
    return true
}

func (term *Terminal) runScript(s Script) {
    if s.name != "" {
        term.InfoPrintln("Running script: " + s.name)
    } else {
        term.InfoPrint("Running script: " + string(s.contents))
    }

    defer term.InfoPrintRuntimeError()
    if err := s.bytecode.Run(); err != nil {
        term.InfoPrintln(err)
    }
}

func (term *Terminal) InputCharacter(ch rune) {
    term.UpdateCommandBeingWritten(ch)

    if term.tryRunBinding(ch) {
        return
    }

	if ch == 263 {
        term.handleBackspace()
	} else {
        term.addChar(ch)
	}
}

func (term *Terminal) AddInputToBuffer() {
    term.state.lines = append(term.state.lines, term.state.line...)
}

func (term *Terminal) handleBackspace() {
    // if there is nothing but a colon, the user decided not to enter a command, so stop command entry mode
    if len(term.state.line) == 1 && term.state.line[0] == ':' {
        term.state.commandBeingWritten = false
    }
    term.state.line = pop(term.state.line)
    term.updateInput()
}

func (term *Terminal) addChar(ch rune) {
    term.state.line = append(term.state.line, []byte(string(ch))...)
    term.updateInput()
}

func (term *Terminal) UpdateCommandBeingWritten(ch rune) {
    if len(term.state.line) == 0 && ch == ':' {
        term.state.commandBeingWritten = true
    }
}

func (term *Terminal) updateInput() {
	replaceCurrentLine(term.inWin, term.state.line)
}

func (c Terminal) InfoPrint(args ...interface{}) {
    c.outWin.Print(args...)
    c.outWin.Refresh()
}

func (c Terminal) InfoPrintln(args ...interface{}) {
    c.outWin.Println(args...)
    c.outWin.Refresh()
}

func (c Terminal) InfoPrintf(format string, args ...interface{}) {
    c.outWin.Printf(format, args...)
    c.outWin.Refresh()
}

func (c Terminal) InfoPrintRuntimeError() {
	if runtimeError := recover(); runtimeError != nil {
		c.InfoPrint(fmt.Sprintf("\nRuntime Error: %s\n", runtimeError))
	}
}


func (c Terminal) RequireArgCount(args []string, count int) bool {
    if len(args) != count {
        c.InfoPrintf("Command Error: %s takes %d arguments but recieved %d.\n", args[0], count, len(args))
        return false
    }
    return true
}

func (c Terminal) RequireArgCountGTE(args []string, count int) bool {
    if len(args) < count {
        c.InfoPrintf("Command Error: %s takes %d or more arguments but recieved %d.\n", args[0], count, len(args))
        return false
    }
    return true
}
