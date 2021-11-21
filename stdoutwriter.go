package main

import (
    "log"
    "os"
    "errors"
    gnc "github.com/rthornton128/goncurses"
)

type EasyBorderedWindow struct {
    window *gnc.Window
    sub_window *gnc.Window
}

type ModifiableWriter struct {
    output *os.File
    bwin *gnc.Window
    mpwin *gnc.Window
}

func (bwin EasyBorderedWindow) InitWindow() {
    y, x := bwin.window.YX()
    h, w := bwin.window.MaxYX()
    err := error(nil)
    bwin.sub_window, err = gnc.NewWindow(h - 2, w - 2, y + 1, x + 1)
    if err != nil {
        log.Fatal(err)
    }
    bwin.sub_window.ScrollOk(true)
}

func Mkfakesub(win *gnc.Window) (*gnc.Window, error) {
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
}

func (w ModifiableWriter) Write(bs []byte) (n int, err error) {
    w.mpwin.Print(string(bs))
    border_err := w.bwin.Border('|', '|', '=', '=', '+', '+', '+', '+')
    if border_err != nil {
        log.Fatal(err)
    }
    w.bwin.Refresh()
    w.mpwin.Refresh()
    return len(bs), nil
}

func (w ModifiableWriter) Close() error {
    //don't close fptr because this causes issues when it's stdout
    //needs to be fixed before used for other outputs
    return nil
}
