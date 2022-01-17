package instance

import (
    "fmt"
)

func (i Instance) GetCharNonBlocking() rune {
    i.backgroundWindow.Timeout(0)
    return rune(i.backgroundWindow.GetChar())
}

func (i Instance) GetLineBlocking() string {
    i.backgroundWindow.Timeout(-1)
    line := ""
    ch := i.backgroundWindow.GetChar()
    for ; ch != '\n'; ch = i.backgroundWindow.GetChar() {
        line = fmt.Sprintf("%s%c", line, rune(ch))
    }

    return line
}

func (i Instance) GetCharBlocking() rune {
    i.backgroundWindow.Timeout(-1)
    ch := i.backgroundWindow.GetChar()
    return rune(ch)
}
