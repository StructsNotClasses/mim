package instance

import (
	"github.com/StructsNotClasses/musicplayer/instance/notification"
	"github.com/StructsNotClasses/musicplayer/remote"
	"github.com/StructsNotClasses/musicplayer/windowwriter"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io"
	"log"
	"os/exec"
)

func pop(bytes []byte) []byte {
	if len(bytes) <= 1 {
		return []byte{}
	} else {
		return bytes[:len(bytes)-1]
	}
}

// replaceCurrentLine erases the current line on the window and prints a new one
// the new string's byte array could potentially contain a newline, which means this can replace the line with multiple lines
func replaceCurrentLine(win *gnc.Window, bs []byte) {
	s := string(bs)
	y, _ := win.CursorYX()
	_, w := win.MaxYX()
	win.HLine(y, 0, ' ', w)
	win.MovePrint(y, 0, s)
	win.Refresh()
}

func windowPrintRuntimeError(outputWindow *gnc.Window) {
	if runtimeError := recover(); runtimeError != nil {
		outputWindow.Print(fmt.Sprintf("\nruntime error: %s\n", runtimeError))
	}
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func playFileWithMplayer(file string, notifier chan notification.Notification, outWindow *gnc.Window) remote.Remote {
	cmd := exec.Command("mplayer",
		"-slave", "-vo", "null", "-quiet", file)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go runWithWriter(cmd, windowwriter.New(outWindow), notifier)

	return remote.Remote{pipe}
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, notifier chan notification.Notification) { // notifier chan int) {
	notifier <- notification.PlaybackBegan
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	notifier <- notification.PlaybackEnded
	w.Close()
}
