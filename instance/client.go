package instance

import (
	gnc "github.com/rthornton128/goncurses"
)

type Client struct {
	backgroundWindow    *gnc.Window
	infoWindow          *gnc.Window
	treeWindow          *gnc.Window
	commandInputWindow  *gnc.Window
	commandOutputWindow *gnc.Window
}

func createClient(scr *gnc.Window) (Client, error) {
	totalHeight, totalWidth := scr.MaxYX()
	leftToRightRatio := 0.6
	COMMAND_INPUT_HEIGHT := 6
	COMMAND_OUTPUT_HEIGHT := 10
	BORDER_WIDTH := 1

	var topWindowHeight int = totalHeight - COMMAND_INPUT_HEIGHT - COMMAND_OUTPUT_HEIGHT - 4*BORDER_WIDTH
	var infoWindowWidth int = int(leftToRightRatio*float64(totalWidth)) - 2*BORDER_WIDTH
	var commandWindowsWidth int = infoWindowWidth
	var treeWindowWidth int = totalWidth - infoWindowWidth - 3*BORDER_WIDTH

	//create the window that displays information about the current song
	infoWindow, err := gnc.NewWindow(topWindowHeight, infoWindowWidth, BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	infoWindow.ScrollOk(true)

	//create the window that holds the song tree
	treeWindow, err := gnc.NewWindow(totalHeight-2*BORDER_WIDTH, treeWindowWidth, BORDER_WIDTH, infoWindowWidth+2*BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	//treeWindow.ScrollOk(true)

	//create the window that allows user input
	commandInputWindow, err := gnc.NewWindow(COMMAND_INPUT_HEIGHT, commandWindowsWidth, topWindowHeight+COMMAND_OUTPUT_HEIGHT+3*BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	commandInputWindow.ScrollOk(true)

	//create the window that holds command output
	commandOutputWindow, err := gnc.NewWindow(COMMAND_OUTPUT_HEIGHT, commandWindowsWidth, topWindowHeight+2*BORDER_WIDTH, BORDER_WIDTH)
	if err != nil {
		return Client{}, err
	}
	commandOutputWindow.ScrollOk(true)

	scr.Box('|', '-')
	scr.VLine(1, infoWindowWidth+BORDER_WIDTH, '|', totalHeight-2*BORDER_WIDTH)
	scr.HLine(topWindowHeight+BORDER_WIDTH, 1, '=', infoWindowWidth)
	scr.HLine(topWindowHeight+COMMAND_OUTPUT_HEIGHT+2*BORDER_WIDTH, 1, '=', infoWindowWidth)
	scr.Refresh()

	return Client{
		scr,
		infoWindow,
		treeWindow,
		commandInputWindow,
		commandOutputWindow,
	}, nil
}
