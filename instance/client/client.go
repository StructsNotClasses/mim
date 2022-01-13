package client

import (
	gnc "github.com/rthornton128/goncurses"
)

func New(scr *gnc.Window) (backgroundWindow *gnc.Window, infoWindow *gnc.Window, treeWindow *gnc.Window, commandInputWindow *gnc.Window, commandOutputWindow *gnc.Window, err error) {
	totalHeight, totalWidth := scr.MaxYX()
	leftToRightRatio := 2.0/3.0

	var terminalWidth int = int(leftToRightRatio*float64(totalWidth)) - 2
	var treeWidth int = totalWidth - terminalWidth - 3

    var mpHeight int = (totalHeight - 4)/3
    var inputHeight int = (totalHeight - mpHeight - 4)/2
    var outputHeight int = totalHeight - mpHeight - inputHeight - 4

    backgroundWindow = scr

	//create the window that displays information about the current song
	infoWindow, err = gnc.NewWindow(mpHeight, terminalWidth, 1, 1)
	if err != nil {
		return
	}
	infoWindow.ScrollOk(true)

	//create the window that holds the song tree
	treeWindow, err = gnc.NewWindow(totalHeight-2, treeWidth, 1, terminalWidth+2)
	if err != nil {
		return
	}

	//create the window that allows user input
	commandInputWindow, err = gnc.NewWindow(inputHeight, terminalWidth, mpHeight + outputHeight + 3, 1)
	if err != nil {
		return 
	}
	commandInputWindow.ScrollOk(true)

	//create the window that holds command output
	commandOutputWindow, err = gnc.NewWindow(outputHeight, terminalWidth, mpHeight+2, 1)
	if err != nil {
		return
	}
	commandOutputWindow.ScrollOk(true)

	scr.Box('|', '-')
    scr.VLine(1, terminalWidth+1, '|', totalHeight-2)
	scr.HLine(mpHeight + 1, 1, '=', terminalWidth)
	scr.HLine(mpHeight+outputHeight+2, 1, '=', terminalWidth)
	scr.Refresh()

	return
}
