package main

import (
    //"bufio"
    //"strings"
    "os"
    "os/exec"
    "log"
    "fmt"
    "time"
    "math/rand"
    "io"
    gnc "github.com/rthornton128/goncurses"
)

const INT32MAX = 2147483647
const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/music_player/songs.json"

//the writing end of the fifo pipe has to be opened only after the reading end is opened
func main() { 
    //current behavior is to regenerate the song list each run. probably needs to change
    storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE) 

    //open the entire song list
    songs, err := openSongList(SONG_LIST_FILE)
    if err != nil {
        log.Fatal(err)
    }

    remote := Remote{nil, false}

    stdscr, err := gnc.Init()
    if err != nil {
        log.Fatal(err)
    }
    defer gnc.End()

    _, width := stdscr.MaxYX()

    newwin, err := gnc.NewWindow(30, width, 0, 0)
    if err != nil {
        log.Fatal(err)
    }
    err = newwin.Border('o', 'o', 'o', 'o', 'o', 'o', 'o', 'o')
    if err != nil {
        log.Fatal(err)
    }
    newwin.Refresh()

    newwinsub, err := Mkfakesub(newwin)
    if err != nil {
        log.Fatal(err)
    }
    newwinsub.ScrollOk(true)
    /*
    err = newwin.Box(' ', '=')
    if err != nil {
        log.Fatal(err)
    }
    */


    gnc.CBreak(true)
    gnc.Echo(false)
    stdscr.Keypad(true)

    /*
    if len(os.Args) > 1 {
        switch os.Args[1] {
        default: // if nothing else arg 1 should be the song to play first
        for _, s := range songs {
            if strings.Contains(s.Path, os.Args[1]) {
                ch := make(chan int)
                playback_complete := false
                play_song(&s, &remote, ch)
                for !playback_complete {
                    i := <- ch
                    if i == 1 {
                        playback_complete = true
                    }
                }
            }
        }
        }
    }*/

    play_all(songs, &remote, stdscr, newwin, newwinsub)
}

func play_all(songs SongList, remote *Remote, scr *gnc.Window, bwin *gnc.Window, mpwin *gnc.Window) {
    //todo: add controls through the remote so you don't have to do wacky shit to quit and stuff
    rand.Seed(time.Now().UnixNano())

    if len(songs) > INT32MAX {
        log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time (the fuck are you even doing playing this many songs lol, you're maxing out the fucking range of an integer)", INT32MAX))
    }

    user_input_ch := make(chan []byte)
    go takeUserInputIntoChannel(user_input_ch, scr)

    for !remote.exit_program {
        rand_num := rand.Int31n(int32(len(songs)))
        notify_ch := make(chan int)
        play_song(&songs[rand_num], remote, notify_ch, bwin, mpwin)

        playback_complete := false
        for !playback_complete {
            select {
            case user_bytes := <- user_input_ch:
                switch string(user_bytes) {
                case "exit\n":
                    remote.exit_program = true
                    remote.SendBytes([]byte("quit\n"))
                default:
                    remote.SendBytes(user_bytes)
                }
            case notification := <- notify_ch:
                if notification == 1 {
                    playback_complete = true
                }
            }
        }
    }
}

func takeUserInputIntoChannel(ch chan []byte, scr *gnc.Window) {
    bs := []byte{}
    //r := bufio.NewReader(os.Stdin)
    //needs a way to exit when play_all exits
    for {
        c := scr.GetChar()
        bs = append(bs, byte(c))
        if c == '\n' {
            ch <- bs
            bs = []byte{}
        }
    }
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func play_song(song *Song, remote *Remote, notify_ch chan int, bwin, mpwin *gnc.Window) {
    remote.pipe = playWithSlaveMplayer(song.Path, notify_ch, bwin, mpwin)
}

func playWithSlaveMplayer(file string, notify_ch chan int, bwin, mpwin *gnc.Window) io.WriteCloser {
    cmd := exec.Command("mplayer", 
        "-slave", "-vo", "null", "-quiet", file)   

    pipe, err := cmd.StdinPipe()
    if err != nil {
        log.Fatal(err)
    }

    wrtr := ModifiableWriter{os.Stdin, bwin, mpwin}
    go runWithWriter(cmd, wrtr, notify_ch)

    return pipe
}

func runWithWriter(cmd *exec.Cmd, w io.WriteCloser, notify_ch chan int) {
    cmd.Stdout = w

    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }

    notify_ch <- 1
    w.Close()
}
