package instance

import (
	"fmt"
)

func (i Instance) GetCharBlocking() rune {
	i.bg.Timeout(-1)
	return rune(i.bg.GetChar())
}

func (i Instance) GetCharNonBlocking() rune {
	i.bg.Timeout(0)
	return rune(i.bg.GetChar())
}

func (i Instance) GetLineBlocking() string {
	i.bg.Timeout(-1)
	line := ""
	ch := i.bg.GetChar()
	for ; ch != '\n'; ch = i.bg.GetChar() {
		line = fmt.Sprintf("%s%c", line, rune(ch))
	}

	return line
}
