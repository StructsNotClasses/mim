package terminal

import (
	"github.com/StructsNotClasses/mim/script"

	gnc "github.com/rthornton128/goncurses"

    "fmt"
)

type InputMode int

const (
	CommandMode InputMode = iota
	CharacterMode
)

type OptionalScript struct {
    exists bool
    s script.Script
}

type TerminalState struct {
    // buffer
	line               []byte
	lines              []byte

    commandBeingWritten bool
    scriptBeingWritten bool

	onPlaybackBeingSet bool
	bindChar           rune
}

type Terminal struct {
    inWin *gnc.Window
    outWin *gnc.Window

    State           TerminalState
    onNoPlayback OptionalScript

    BindMap         map[rune]script.Script
    CommandMap      map[string]string
    AliasMap        map[string]string
}

func New(inwin, outwin *gnc.Window) Terminal {
    return Terminal{
        inWin: inwin,
        outWin: outwin,

        State: TerminalState{
            line: []byte{},
            lines: []byte{},
            bindChar:           0,
            commandBeingWritten: false,
            scriptBeingWritten: false,
            onPlaybackBeingSet: false,
        },
        onNoPlayback: OptionalScript{
            exists: false,
        },

        BindMap: make(map[rune]script.Script),
        CommandMap: make(map[string]string),
        AliasMap: make(map[string]string),
    }
}

func (term *Terminal) TryRunNoPlaybackScript() {
    if term.onNoPlayback.exists {
        term.RunScript(term.onNoPlayback.s)
    }
}

func (term *Terminal) ClearLine() {
    term.State.line = []byte{}
}

func (term *Terminal) SetNoPlayback(s script.Script) {
    if s.IsEmpty() {
        term.onNoPlayback.exists = false
    } else {
        term.onNoPlayback.exists = true
    }
    term.onNoPlayback.s = s
    term.State.onPlaybackBeingSet = false
}

func (term *Terminal) NextScriptShouldBeBound() bool {
    return term.State.bindChar != 0
}

func (term *Terminal) NextScriptIsNoPlayback() bool {
    return term.State.onPlaybackBeingSet
}

func (term *Terminal) BindCurrentToScript(s script.Script) {
    term.BindMap[term.State.bindChar] = s
    term.State.bindChar = 0
}

func (term *Terminal) Binding() rune {
    return term.State.bindChar
}

func (term *Terminal) SetBinding(ch rune) {
    term.State.bindChar = ch
}

func (term *Terminal) ClearBinding() {
    term.State.bindChar = 0
}

func (term Terminal) CurrentLine() string {
    return string(term.State.line)
}

func (term *Terminal) BeginCommand() {
    term.State.commandBeingWritten = true
}

func (term Terminal) CommandBeingWritten() bool {
    return term.State.commandBeingWritten
}

func (term *Terminal) EndCommand() {
    term.State.commandBeingWritten = false
    term.ClearLine()
}

func (term *Terminal) BeginScript() {
    term.State.scriptBeingWritten = true
}

func (term Terminal) ScriptBeingWritten() bool {
    return term.State.scriptBeingWritten
}

func (term *Terminal) EndScript() {
    term.State.scriptBeingWritten = false
    term.ClearLine()
}

func (term *Terminal) WrittenScript() []byte {
    return term.State.lines
}

func (term *Terminal) ClearWrittenScript() {
    term.State.lines = []byte{}
}

func (term *Terminal) BindNextScriptToNoPlayback() {
    term.State.onPlaybackBeingSet = true
}

func (term *Terminal) RunBinding(ch rune) {
    if script, ok := term.BindMap[ch]; ok {
        term.RunScript(script)
    } else {
        term.InfoPrintf("%c is not bound.\n", ch)
    }
}

func (term *Terminal) RunScript(s script.Script) {
    term.InfoPrintln("Running script: " + s.Name())

    defer term.InfoPrintRuntimeError()
    if err := s.Run(); err != nil {
        term.InfoPrintln(err)
    }
}

func (term *Terminal) InputCharacter(ch rune) {
    term.UpdateCommandBeingWritten(ch)

    if !term.CommandBeingWritten() && !term.ScriptBeingWritten() && validBinding(ch) {
        term.RunBinding(ch) 
    } else if ch == 263 {
        term.handleBackspace()
	} else {
        term.addChar(ch)
	}
}

func (term *Terminal) PushLineToBuffer() {
    term.State.lines = append(term.State.lines, term.State.line...)
    term.ClearLine()
}

func (term *Terminal) handleBackspace() {
    // if there is nothing but a colon, the user decided not to enter a command, so stop command entry mode
    if len(term.State.line) == 1 && term.State.line[0] == ':' {
        term.State.commandBeingWritten = false
    }
    term.State.line = pop(term.State.line)
    term.updateInput()
}

func (term *Terminal) addChar(ch rune) {
    term.State.line = append(term.State.line, []byte(string(ch))...)
    term.updateInput()
}

func (term *Terminal) UpdateCommandBeingWritten(ch rune) {
    if len(term.State.line) == 0 && ch == ':' {
        term.BeginCommand()
    }
}

func (term *Terminal) updateInput() {
	replaceCurrentLine(term.inWin, term.State.line)
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

func validBinding(ch rune) bool {
    const cantBind = ":\n"
    for _, cant := range(cantBind) {
        if ch == cant {
            return false
        }
    }
    return true
}

func pop(bytes []byte) []byte {
    if len(bytes) >= 1 {
        return bytes[:len(bytes)-1]
    }
    return bytes
}

// replaceCurrentLine erases the current line on the window and prints a new one
// the new string's byte array can contain a newline, which means this can replace the line with multiple lines
func replaceCurrentLine(win *gnc.Window, bs []byte) {
	s := string(bs)
	y, _ := win.CursorYX()
	_, w := win.MaxYX()
	win.HLine(y, 0, ' ', w)
	win.MovePrint(y, 0, s)
	win.Refresh()
}
