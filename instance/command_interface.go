package instance

import (
    "fmt"
)

func (commands *CommandInterface) runScript(s Script) {
    if s.name != "" {
        commands.InfoPrintln("Running script: " + s.name)
    } else {
        commands.InfoPrint("Running script: " + string(s.contents))
    }

    defer commands.InfoPrintRuntimeError()
    if err := s.bytecode.Run(); err != nil {
        commands.InfoPrintln(err)
    }
}

func (c CommandInterface) InfoPrint(args ...interface{}) {
    c.outWin.Print(args...)
    c.outWin.Refresh()
}

func (c CommandInterface) InfoPrintln(args ...interface{}) {
    c.outWin.Println(args...)
    c.outWin.Refresh()
}

func (c CommandInterface) InfoPrintf(format string, args ...interface{}) {
    c.outWin.Printf(format, args...)
    c.outWin.Refresh()
}

func (c CommandInterface) InfoPrintRuntimeError() {
	if runtimeError := recover(); runtimeError != nil {
		c.InfoPrint(fmt.Sprintf("\nRuntime Error: %s\n", runtimeError))
	}
}

func (c CommandInterface) UpdateInput(currentLine []byte) {
	replaceCurrentLine(c.inWin, currentLine)
}

func (c CommandInterface) RequireArgCount(args []string, count int) bool {
    if len(args) != count {
        c.InfoPrintf("Command Error: %s takes %d arguments but recieved %d.\n", args[0], count, len(args))
        return false
    }
    return true
}

func (c CommandInterface) RequireArgCountGTE(args []string, count int) bool {
    if len(args) < count {
        c.InfoPrintf("Command Error: %s takes %d or more arguments but recieved %d.\n", args[0], count, len(args))
        return false
    }
    return true
}
