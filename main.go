package main

import (
	"github.com/StructsNotClasses/musicplayer/remote"
	"github.com/StructsNotClasses/musicplayer/safebool"
	"github.com/StructsNotClasses/musicplayer/song"

	gnc "github.com/rthornton128/goncurses"

	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	run(createInstance(backgroundWindow, songs))
}

func run(instance Instance) {
	for !instance.state.exit {
		select {
		case userByte := <-instance.userInputChannel:
            // if in character mode, run the script bound to the received key and skip the rest
            // (currently the colon key could be bound to something haha)
            if script, ok := instance.bindMap[userByte]; ok && !instance.state.mode {
                instance.client.commandOutputWindow.Print("Running script: " + string(script))
                instance.client.commandOutputWindow.Refresh()
                instance.runBytesAsScript(script)
                continue
            }

            // use the colon key to unset character mode instantly
            if userByte == ':' {
                instance.state.mode = true
            } 

            if userByte == 263 {
                // if backspace remove last byte from slice
				instance.state.line = tryPopByte(instance.state.line)
			} else {
                // for any other character add it to the line buffer
				instance.state.line = append(instance.state.line, byte(userByte))
			}

			replaceCurrentLine(instance.client.commandInputWindow, instance.state.line)
			if userByte == '\n' {
                // handle single line no argument commands
                switch string(instance.state.line) {
                case ":exit\n":
					instance.state.exit = true
					instance.currentRemote.SendString("quit\n")
                case ":end\n":
					instance.manageByteScript(instance.state.lines)
					instance.state.lines = []byte{}
                case ":on_no_playback\n":
					instance.state.onPlaybackBeingSet = true
                case "debug_print_buffer":
					instance.client.commandOutputWindow.Printf("line: '%s'\nbytes: '%s'\n", string(instance.state.line), string(instance.state.lines))
					instance.client.commandOutputWindow.Refresh()
				case ":character_mode\n":
					instance.state.mode = false
                default:
                    // handle single line commands with arguments
                    switch string(instance.state.line[0:5]) {
                    case ":load":
                        fileName := string(instance.state.line[6 : len(instance.state.line)-1])
                        bytes, err := ioutil.ReadFile(fileName)
                        if err != nil {
                            instance.client.commandOutputWindow.Printf("load: Failed to load file '%s' with error '%v'\n", fileName, err)
                            instance.client.commandOutputWindow.Refresh()
                        } else {
                            instance.manageByteScript(bytes)
                        }
                    case ":bind":
                        // these commands should follow the format
                        // :bind <character>
                        // <script>
                        // :end
                        instance.state.bindChar = gnc.Key(instance.state.line[6])
                        instance.state.lines = []byte{}
                    default:
                        instance.state.lines = append(instance.state.lines, instance.state.line...)
                    }
                } 
                //always clear the line buffer
                instance.state.line = []byte{}
			}
			instance.client.backgroundWindow.Refresh()
        default:
            // run the script for when no song is playing if neccesary
            if !instance.playbackInProgress.Get() && len(instance.runOnNoPlayback) > 0 {
                instance.client.commandOutputWindow.Print("Running script: " + string(instance.runOnNoPlayback))
                instance.client.commandOutputWindow.Refresh()
                instance.runBytesAsScript(instance.runOnNoPlayback)
            }
		}
	}
}

func tryPopByte(bytes []byte) []byte {
	if len(bytes) == 1 {
		return []byte{}
	} else if len(bytes) > 0 {
		return bytes[:len(bytes)-1]
	}
	return bytes
}

func replaceCurrentLine(win *gnc.Window, bs []byte) {
    s := string(bs)
    y, _ := win.CursorYX()
	_, w := win.MaxYX()
	win.HLine(y, 0, ' ', w)
    win.MovePrint(y, 0, s)
    win.Refresh()
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
func playFileWithMplayer(file string, playbackInProgress *safebool.SafeBool, outWindow *gnc.Window) remote.Remote {
	cmd := exec.Command("mplayer",
		"-slave", "-vo", "null", "-quiet", file)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go runWithWriter(cmd, BasicWindowWriter{outWindow}, playbackInProgress)

	return remote.Remote{pipe}
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, boolWrapper *safebool.SafeBool) { // notifier chan int) {
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	boolWrapper.Set(false)
	w.Close()
}
