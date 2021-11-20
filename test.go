package main

import (
    "bufio"
    "os"
    "os/exec"
    "log"
    "fmt"
    "time"
    "math/rand"
    "io"
    //"github.com/rthornton128/goncurses"
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
    /*
    if len(os.Args) > 1 {
        switch os.Args[1] {
        default: // if nothing else arg 1 should be the song to play first
        for _, s := range songs {
            if strings.Contains(s.Path, os.Args[1]) {
                play_song(&s, &remote)
            }
        }
        }
    }
    */

    play_all(songs, &remote)
}

func play_all(songs SongList, remote *Remote) {
    //todo: add controls through the remote so you don't have to do wacky shit to quit and stuff
    rand.Seed(time.Now().UnixNano())

    if len(songs) > INT32MAX {
        log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time (the fuck are you even doing playing this many songs lol, you're maxing out the fucking range of an integer)", INT32MAX))
    }

    user_input_ch := make(chan []byte)
    go takeUserInputIntoChannel(user_input_ch)

    for !remote.exit_program {
        rand_num := rand.Int31n(int32(len(songs)))
        notify_ch := make(chan int)
        play_song(&songs[rand_num], remote, notify_ch)

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

func takeUserInputIntoChannel(ch chan []byte) {
    r := bufio.NewReader(os.Stdin)
    //needs a way to exit when play_all exits
    for {
        bs, err := r.ReadBytes('\n')
        if err != nil {
            fmt.Println("failed to read input", err)
        }
        ch <- bs
    }
}

// run mplayer command "mplayer -slave -vo null <song path>"
// the mplayer runner should send 1 to notify_ch when it completes playback. otherwise, nothing should be sent
func play_song(song *Song, remote *Remote, notify_ch chan int) {
    pipe, writer := playWithSlaveMplayer(song.Path, notify_ch)
    remote.pipe = &pipe

    go printMplayerOutput(writer)

    /*
    for {
        user_string := <-user_input_channel
        switch user_string {
        case "exit\n":
            remote.exit_program = true
            remote.SendString("quit")
        case "skip\n":
            remote.SendString("quit")
        case "mout\n":
            fmt.Print(string(mplayer_output))
        default:
            remote.SendString(user_string)
        }
    }
    */
}

func printMplayerOutput(w AsyncWriter) {
    value := byte(0)
    ok := true
    for ok {
        value, ok = <- w
        fmt.Printf("%c", value)
    }
}

func playWithSlaveMplayer(file string, notify_ch chan int) (io.WriteCloser, AsyncWriter) {
    cmd := exec.Command("mplayer", 
        "-slave", "-vo", "null", "-quiet", file)   

    pipe, err := cmd.StdinPipe()
    if err != nil {
        log.Fatal(err)
    }

    wrtr := make(chan byte)
    go runWithWriter(cmd, wrtr, notify_ch)

    return pipe, wrtr
}

func runWithWriter(cmd *exec.Cmd, w AsyncWriter, notify_ch chan int) {
    cmd.Stdout = w

    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }

    notify_ch <- 1
    w.Close()
}
