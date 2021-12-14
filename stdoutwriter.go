package main

import (
    gnc "github.com/rthornton128/goncurses"
)

type BasicWindowWriter struct {
    outWindow *gnc.Window
}

/*func Mkfakesub(win *gnc.Window) (*gnc.Window, error) {
    if win != nil {
        y, x := win.YX()
        h, w := win.MaxYX()
        sub, err := gnc.NewWindow(h - 2, w - 2, y + 1, x + 1)
        if err != nil {
            return nil, err
        }
        return sub, nil
    }
    return nil, errors.New("nil pointer supplied to Mkfakesub")
}*/

func (w BasicWindowWriter) Write(bs []byte) (n int, err error) {
    w.outWindow.Print(string(bs))
    w.outWindow.Refresh()
    return len(bs), nil
}

func (w BasicWindowWriter) Close() error {
    //don't close fptr because this causes issues when it's stdout
    //needs to be fixed before used for other outputs
    return nil
}
