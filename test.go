package main

import (
    //"bufio"
    //"strings"
    //"os"
    "os/exec"
    "log"
    "fmt"
    "time"
    "math/rand"
    "io"
    "errors"
    gnc "github.com/rthornton128/goncurses"
)

const INT32MAX = 2147483647
const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/music_player/songs.json"
var width, height int

type Client struct {
    backgroundWindow *gnc.Window
    infoWindow *gnc.Window
    treeWindow *gnc.Window
    commandOutputWindow *gnc.Window
    commandInputWindow *gnc.Window
}

//the writing end of the fifo pipe has to be opened only after the reading end is opened
func main() { 
    //current behavior is to regenerate the song list each run. probably needs to change
    storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE) 

    //open the entire song list
    songs, err := openSongList(SONG_LIST_FILE)
    if err != nil {
        log.Fatal(err)
    }

    stdscr, err := gnc.Init()
    if err != nil {
        log.Fatal(err)
    }
    defer gnc.End()

    gnc.CBreak(true)
    gnc.Echo(false)
    stdscr.Keypad(true)

    var c Client;
    c.backgroundWindow = stdscr
    c.infoWindow, c.treeWindow, c.commandInputWindow, c.commandOutputWindow = tryCreateMplayerWindows(stdscr)

    play_all(songs, c)
}

func tryCreateMplayerWindows(scr *gnc.Window) (*gnc.Window, *gnc.Window, *gnc.Window, *gnc.Window) {
    totalHeight, totalWidth := scr.MaxYX()
    INFO_WINDOW_PORTION := 0.6
    COMMAND_INPUT_HEIGHT := 6
    COMMAND_OUTPUT_HEIGHT := 10
    BORDER_WIDTH := 1
    
    var topWindowHeight int = totalHeight - COMMAND_INPUT_HEIGHT - COMMAND_OUTPUT_HEIGHT - 4 * BORDER_WIDTH

    var infoWindowWidth int = int(INFO_WINDOW_PORTION * float64(totalWidth)) - 2 * BORDER_WIDTH
    var commandWindowsWidth int = infoWindowWidth
    var treeWindowWidth int = totalWidth - infoWindowWidth - 3 * BORDER_WIDTH

    //create the window that displays information about the current song
    infoWindow, err := gnc.NewWindow(topWindowHeight, infoWindowWidth, BORDER_WIDTH, BORDER_WIDTH)
    if err != nil {
        log.Fatal(err)
    }
    infoWindow.ScrollOk(true)

    //create the window that holds the song tree
    treeWindow, err := gnc.NewWindow(topWindowHeight, treeWindowWidth, BORDER_WIDTH, infoWindowWidth + 2 * BORDER_WIDTH)
    if err != nil {
        log.Fatal(err)
    }

    //create the window that allows user input
    commandInputWindow, err := gnc.NewWindow(COMMAND_INPUT_HEIGHT, commandWindowsWidth, topWindowHeight + COMMAND_OUTPUT_HEIGHT + 3 * BORDER_WIDTH, BORDER_WIDTH)
    if err != nil {
        log.Fatal(err)
    }
    commandInputWindow.ScrollOk(true)

    //create the window that holds command output
    commandOutputWindow, err := gnc.NewWindow(COMMAND_OUTPUT_HEIGHT, commandWindowsWidth, topWindowHeight + 2 * BORDER_WIDTH, BORDER_WIDTH)
    if err != nil {
        log.Fatal(err)
    }
    commandOutputWindow.ScrollOk(true)

    scr.Box('|', '-')
    scr.VLine(1, infoWindowWidth + BORDER_WIDTH, '|', totalHeight - 2 * BORDER_WIDTH)
    scr.HLine(topWindowHeight + BORDER_WIDTH, 1, '=', infoWindowWidth)
    scr.HLine(topWindowHeight + COMMAND_OUTPUT_HEIGHT + 2 * BORDER_WIDTH, 1, '=', infoWindowWidth)
    scr.Refresh()

    return infoWindow, treeWindow, commandInputWindow, commandOutputWindow
}

func play_all(songs SongList, client Client) {
    //check if windows are nil
    if client.backgroundWindow == nil || client.infoWindow == nil || client.commandOutputWindow == nil || client.treeWindow == nil || client.commandInputWindow == nil {
        log.Fatal(errors.New("Nil pointer to window provided for playback, quitting..."))
    }

    rand.Seed(time.Now().UnixNano())
    if len(songs) > INT32MAX {
        log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time.", INT32MAX))
    }

    user_input_ch := make(chan gnc.Key)
    go takeUserInputIntoChannel(client.backgroundWindow, user_input_ch)

    remote := Remote{nil}
    user_bs := []byte{}
    exit_program := false
    for !exit_program {
        rand_num := rand.Int31n(int32(len(songs)))
        notifier := make(chan int)
        remote = playFileWithMplayer(songs[rand_num].Path, notifier, client.infoWindow)

        playback_complete := false
        for !playback_complete {
            select {
            case user_b := <- user_input_ch:
                user_bs = append(user_bs, byte(user_b))
                client.commandInputWindow.AddChar(gnc.Char(user_b))
                client.commandInputWindow.Refresh()
                if user_b == '\n' || user_b == '\r' {
                    switch string(user_bs) {
                    case "exit\n":
                        exit_program = true
                        remote.SendBytes([]byte("quit\n"))
                    case "current song info\n":
                        client.commandOutputWindow.Print("\nPlaying " + songs[rand_num].Name)
                        client.commandOutputWindow.Refresh()
                    default:
                        if string(user_bs[0:4]) == "send" {
                            client.commandOutputWindow.Print("\nSending '" + string(user_bs[5:]) + "'")
                            client.commandOutputWindow.Refresh()
                            mplayerCommand := user_bs[5:]
                            remote.SendBytes(mplayerCommand)
                        } else {
                            remote.SendBytes(user_bs)
                        }
                    }
                    user_bs = []byte{}
                }
            case notification := <- notifier:
                if notification == 1 {
                    playback_complete = true
                }
            }
            client.backgroundWindow.Refresh()
        }
    }
}

func takeUserInputIntoChannel(window *gnc.Window, ch chan gnc.Key) {
    for {
        ch <- window.GetChar()
    }
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func playFileWithMplayer(file string, notifier chan int, outWindow *gnc.Window) Remote {
    cmd := exec.Command("mplayer", 
        "-slave", "-vo", "null", "-quiet", file)   

    pipe, err := cmd.StdinPipe()
    if err != nil {
        log.Fatal(err)
    }

    go runWithWriter(cmd, BasicWindowWriter{outWindow}, notifier)

    return Remote{pipe}
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
