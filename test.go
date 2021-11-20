package main

import (
    "bufio"
    "os"
    "strings"
    "os/exec"
    "log"
    "fmt"
    "syscall"
    "time"
    "math/rand"
    //"github.com/rthornton128/goncurses"
)

const CONTROL_FILE_NAME string = "/tmp/mplayer_controls.txt"
const INT32MAX = 2147483647
const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/music_player/songs.json"

type AsyncWriter chan byte

func main() { //the writing end of the fifo pipe has to be opened only after the reading end is opened
    remote := Remote{CONTROL_FILE_NAME, nil, false}

    //current behavior is to regenerate the song list each run. probably needs to change
    fmt.Println(os.Args)
    storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE) 

    //open the entire song list
    songs, err := openSongList(SONG_LIST_FILE)
    if err != nil {
        log.Fatal(err)
    }

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

    play_all(songs, &remote)
}

func createControlFile() {
    creation_err := create_fifo(CONTROL_FILE_NAME)
    if creation_err != nil {
        log.Fatal(creation_err)
    }
}

func play_all(songs SongList, remote *Remote) {
    //generate random number between 0 and len(songs)
    //play song at index r
    //  run mplayer command "mplayer -slave -input ctrl_file_string -vo null <song path>"
    //todo: add controls through the remote so you don't have to do wacky shit to quit and stuff
    rand.Seed(time.Now().UnixNano())

    if len(songs) > INT32MAX {
        log.Fatal(fmt.Sprintf("Cannot play more than %d songs at a time (the fuck are you even doing playing this many songs lol, you're maxing out the fucking range of an integer)", INT32MAX))
    }

    for !remote.exit_program {
        rand_num := rand.Int31n(int32(len(songs)))
        play_song(&songs[rand_num], remote)
    }
}

func play_song(song *Song, remote *Remote) {
    os.Remove(CONTROL_FILE_NAME)
    createControlFile()
    defer os.Remove(CONTROL_FILE_NAME)

    writer := AsyncWriter(make(chan byte))
    user_input_channel := make(chan string)
    mplayer_output := []byte{}
    value := byte(0)

    go receiveUserInput(user_input_channel)
    go playFileWithSlaveMplayer(CONTROL_FILE_NAME, song.Path, writer)

    err := remote.Open()
    if err != nil {
        log.Fatal("failed to open mplayer control file")
    }
    defer remote.Close()

    //fmt.Println("Enter Commands Below:")

    remote.SendString(" \n \n")

    keep_going := true
    for keep_going {
        select {
        case user_string := <- user_input_channel:
            switch user_string {
            case "exit\n":
                remote.exit_program = true
                remote.DirtySendString("quit", CONTROL_FILE_NAME)
            case "skip\n":
                remote.DirtySendString("quit", CONTROL_FILE_NAME)
            case "mout\n":
                fmt.Print(string(mplayer_output))
            default:
                remote.DirtySendString(user_string, CONTROL_FILE_NAME)
            }
        case value, keep_going = <- writer:
            mplayer_output = append(mplayer_output, value)
        }
    }
}

func playFileWithSlaveMplayer(ctrl string, file string, w AsyncWriter) {
    arguments := []string{
        "-slave", 
        "-input", fmt.Sprintf("file=%s", ctrl), 
        "-vo", "null", 
        "-quiet", file}
    run_mplayer(arguments, w)
}

func receiveUserInput(ch chan string) {
    r := bufio.NewReader(os.Stdin)
    for {
        bs, err := r.ReadBytes('\n')
        if err != nil {
            fmt.Println("failed to read input", err)
        }
        ch <- string(bs)
    }
}

func run_mplayer(args []string, writer AsyncWriter) {
    cmd := exec.Command("mplayer")
    cmd.Args = args
    cmd.Stdout = writer

    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }

    writer.Close()
}

func create_fifo(filename string) error {
    os.Remove(filename)
    err := syscall.Mkfifo(filename, 0666)
    if err != nil {
        log.Fatal("Unable to create named pipe because of error:", err)
    }
    return err
}

func remove(filename string) { //better be careful with this one haha
    cmd := exec.Command("rm", filename)
    err := cmd.Run()
    if err != nil {
        fmt.Printf("Failed to remove created file '%'\n", filename)
        log.Fatal(err)
    }
}

func (w AsyncWriter) Write(p []byte) (n int, err error) {
    for _, b := range p {
        w <- b
    }
    return len(p), nil
}

func (w AsyncWriter) Close() {
    close(w)
}
