package main

import (
	"github.com/StructsNotClasses/musicplayer/remote"
	"github.com/StructsNotClasses/musicplayer/song"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io"
	"log"
	"math/rand"
	"os/exec"
)

const INT32MAX = 2147483647
const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"

//script management can be done with the following technique:
// when a script is created, a number is incremented and passed to it as a variable named ID or something
// this number is also placed in an array along with potentially information about the script like name or file and the script runtime maybe
// there is then a thread safe int holding the desired reciever script of a message, and another variable (thread protected by the other one) holding a message to the tengo script
// all scripts can call a function and pass their ID value to it to recieve either the message, true or nil, false
// this system would allow the user to write scripts that take arbitrary user input, the main intention being to allow a script to be killed or otherwise managed using a simple interface
func main() {
	//current behavior is to regenerate the song list each run. probably needs to change
	storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE)

	//open the entire song list
	songs, err := song.CreateList(SONG_LIST_FILE)
	if err != nil {
		log.Fatal(err)
	}

	backgroundWindow, err := gnc.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer gnc.End()

	gnc.CBreak(true)
    //gnc.Keypad(true)
	gnc.Echo(false)
	backgroundWindow.Keypad(true)

	playAll(songs, createInstance(backgroundWindow, songs))
}

func playAll(songs song.List, instance Instance) {
	user_input_ch := make(chan gnc.Key)
	go takeUserInputIntoChannel(instance.client.backgroundWindow, user_input_ch)

	user_bs := []byte{}
	exit_program := false

	for !exit_program {
		r := rand.Int31n(int32(len(songs)))
		instance.PlayIndex(int(r))
		for instance.playbackInProgress {
			select {
			case user_b := <-user_input_ch:
                // if backspace remove last byte from slice
                if user_b == 263 {
                    if len(user_bs) == 1 {
                        user_bs = []byte{}
                    } else if len(user_bs) > 0 {
                        user_bs = user_bs[:len(user_bs) - 1]
                    }
                } else {
                    user_bs = append(user_bs, byte(user_b))
                }
                y, _ := instance.client.commandInputWindow.CursorYX()
                clearLine(y, instance.client.commandInputWindow)
                instance.client.commandInputWindow.MovePrint(y, 0, string(user_bs))
                instance.client.commandInputWindow.Refresh()
				if user_b == '\n' || user_b == '\r' {
					switch string(user_bs) {
					case "exit\n":
						exit_program = true
						instance.currentRemote.SendString("quit\n")
					default:
						instance.client.commandOutputWindow.Print("Running script: " + string(user_bs))
						instance.client.commandOutputWindow.Refresh()
						instance.runBytesAsScript(user_bs)
					}
					user_bs = []byte{}
				}
			case <-instance.notifier:
				instance.playbackInProgress = false
			}
			instance.client.backgroundWindow.Refresh()
		}
	}
}

func clearLine(y int, win *gnc.Window) {
    _, w := win.MaxYX()
    win.HLine(y, 0, ' ', w)
}

func handleRuntimeError(outputWindow *gnc.Window) {
	if runtimeError := recover(); runtimeError != nil {
		outputWindow.Print(fmt.Sprintf("%s\n", runtimeError))
	}
}

func takeUserInputIntoChannel(window *gnc.Window, ch chan gnc.Key) {
	for {
		ch <- window.GetChar()
	}
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func playFileWithMplayer(file string, notifier chan int, outWindow *gnc.Window) remote.Remote {
	cmd := exec.Command("mplayer",
		"-slave", "-vo", "null", "-quiet", file)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go runWithWriter(cmd, BasicWindowWriter{outWindow}, notifier)

	return remote.Remote{pipe}
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, notifier chan int) {
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	notifier <- 1
	w.Close()
}
