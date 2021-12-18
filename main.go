package main

import (
    //"bufio"
    //"strings"
    //"os"
    //"strconv"
    "os/exec"
    "log"
    "fmt"
    "time"
    "math/rand"
    "io"
    "errors"
    gnc "github.com/rthornton128/goncurses"
    tango "github.com/StructsNotClasses/tengotango"
    "github.com/StructsNotClasses/musicplayer/remote"
    "github.com/StructsNotClasses/musicplayer/scriptapi"
    //tengo "github.com/d5/tengo/v2"
)

const INT32MAX = 2147483647
const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"

type Client struct {
    backgroundWindow *gnc.Window
    infoWindow *gnc.Window
    treeWindow *gnc.Window
    commandInputWindow *gnc.Window
    commandOutputWindow *gnc.Window
}

type SongTree struct {
    songs SongList
    currentIndex int32
    currentAtTop int32
}

func main() { 
    //current behavior is to regenerate the song list each run. probably needs to change
    storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE) 

    //open the entire song list
    songs, err := createSongList(SONG_LIST_FILE)
    if err != nil {
        log.Fatal(err)
    }

    backgroundWindow, err := gnc.Init()
    if err != nil {
        log.Fatal(err)
    }
    defer gnc.End()

    gnc.CBreak(true)
    gnc.Echo(false)
    backgroundWindow.Keypad(true)

    playAll(songs, createGuiClient(backgroundWindow))
}

func createGuiClient(scr *gnc.Window) Client {
    totalHeight, totalWidth := scr.MaxYX()
    leftToRightRatio := 0.6
    COMMAND_INPUT_HEIGHT := 6
    COMMAND_OUTPUT_HEIGHT := 10
    BORDER_WIDTH := 1
    
    var topWindowHeight int = totalHeight - COMMAND_INPUT_HEIGHT - COMMAND_OUTPUT_HEIGHT - 4 * BORDER_WIDTH
    var infoWindowWidth int = int(leftToRightRatio * float64(totalWidth)) - 2 * BORDER_WIDTH
    var commandWindowsWidth int = infoWindowWidth
    var treeWindowWidth int = totalWidth - infoWindowWidth - 3 * BORDER_WIDTH

    //create the window that displays information about the current song
    infoWindow, err := gnc.NewWindow(topWindowHeight, infoWindowWidth, BORDER_WIDTH, BORDER_WIDTH)
    if err != nil {
        log.Fatal(err)
    }
    infoWindow.ScrollOk(true)

    //create the window that holds the song tree
    treeWindow, err := gnc.NewWindow(totalHeight - 2 * BORDER_WIDTH, treeWindowWidth, BORDER_WIDTH, infoWindowWidth + 2 * BORDER_WIDTH)
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

    return Client{
        scr,
        infoWindow,
        treeWindow, 
        commandInputWindow, 
        commandOutputWindow,
    }
}

func playAll(songs SongList, client Client) {
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

    user_bs := []byte{}
    exit_program := false

    st := SongTree{songs, 0, 0}

    for !exit_program {
        st.currentIndex = rand.Int31n(int32(len(songs)))
        drawTree(&st, client.treeWindow)
        notifier := make(chan int)
        scriptapi.RemoteToCurrentInstance = playFileWithMplayer(songs[st.currentIndex].Path, notifier, client.infoWindow)
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
                        scriptapi.RemoteToCurrentInstance.SendBytes([]byte("quit\n"))
                    case "current song info\n":
                        client.commandOutputWindow.Print("\nPlaying " + songs[st.currentIndex].Name)
                        client.commandOutputWindow.Refresh()
                    default:
                        if string(user_bs[0:4]) == "send" {
                            client.commandOutputWindow.Print("\nSending '" + string(user_bs[5:len(user_bs) - 1]) + "'")
                            client.commandOutputWindow.Refresh()
                            mplayerCommand := user_bs[5:]
                            scriptapi.RemoteToCurrentInstance.SendBytes(mplayerCommand)
                        } else {
                            client.commandOutputWindow.Print("Running script: " + string(user_bs))
                            client.commandOutputWindow.Refresh()
                            runBytesAsScript(user_bs, client.commandOutputWindow)
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

func runBytesAsScript(bs []byte, outputWindow *gnc.Window) {
    script := tango.NewScript(bs)

    bytecode, err := script.Compile()
    if err != nil {
        outputWindow.Print(err)
    } else {
        outputWindow.Print("Successful compilation!\n")
    }
    outputWindow.Refresh()
    if err := bytecode.Run(); err != nil {
        outputWindow.Print(err)
        outputWindow.Refresh()
    }
    x := bytecode.Get("x").Int()
    outputWindow.Print("The value of x is ", x)
    outputWindow.Refresh()
}

func drawTree(tree *SongTree, win *gnc.Window) {
    win.Erase()
    hog, width := win.MaxYX()
    height := int32(hog)
    delta := tree.currentAtTop - tree.currentIndex
    if delta > 56 {
        tree.currentAtTop = tree.currentIndex - 56
    } else if delta < 0 {
        tree.currentAtTop = tree.currentIndex
    }
    if tree.currentAtTop < 0 {
        tree.currentAtTop = 0
    }
    for i := tree.currentAtTop; i < tree.currentAtTop + height && i < int32(len(tree.songs)); i++ {
        song := tree.songs[i]
        if i == tree.currentIndex {
            win.AttrOn(gnc.A_STANDOUT)
            win.AttrOff(gnc.A_NORMAL)
            winClampPrintln(win, song.Name, width)
            win.AttrOff(gnc.A_STANDOUT)
            win.AttrOn(gnc.A_NORMAL)
        } else {
            winClampPrintln(win, song.Name, width)
        }
    }
    win.Refresh()
}

func winClampPrintln(w *gnc.Window, s string, limit int) {
    if len(s) > limit {
        w.Print(s[0:limit - 1] + "\n")
    } else {
        w.Print(s + "\n")
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
