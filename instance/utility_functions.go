package instance

import (
	"github.com/StructsNotClasses/mim/instance/playback"
	"github.com/StructsNotClasses/mim/remote"
	"github.com/StructsNotClasses/mim/windowwriter"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io"
	"log"
	"os/exec"
)

func windowPrintRuntimeError(outputWindow *gnc.Window) {
	if runtimeError := recover(); runtimeError != nil {
		outputWindow.Print(fmt.Sprintf("\nruntime error: %s\n", runtimeError))
		outputWindow.Refresh()
	}
}

// playFileWithMplayer runs the command "mplayer -slave -vo null <file>" and notifies upon the beginning and end of playback to notifier
// the remote returned contains a pipe to the commands stdin and can be used to send it input
func playFileWithMplayer(file string, notifier chan playback.Notification, out windowwriter.WindowWriter) remote.Remote {
	cmd := exec.Command("mplayer",
		"-slave", "-vo", "null", "-quiet", file)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go runWithWriter(cmd, out, notifier)

	return remote.Remote{pipe}
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, notifier chan playback.Notification) {
	notifier <- playback.Began
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	notifier <- playback.Ended
}
