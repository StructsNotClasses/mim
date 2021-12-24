package windowwriter

import (
	gnc "github.com/rthornton128/goncurses"
)

type WindowWriter struct {
	win *gnc.Window
}

func New(win *gnc.Window) WindowWriter {
    return WindowWriter{win}
}

func (w WindowWriter) Write(bs []byte) (n int, err error) {
	w.win.Print(string(bs))
	w.win.Refresh()
	return len(bs), nil
}

func (w WindowWriter) Close() error {
	return nil
}
