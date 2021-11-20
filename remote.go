package main

import (
    "os"
    "fmt"
    "bufio"
    "log"
    "os/exec"
    "io"
)

type Remote struct {
    fname string
    fptr *os.File
    pipe io.WriteCloser
    exit_program bool
}


func (r *Remote) SendString(cmd string) {
    r.fptr.WriteString(cmd+"\n")
}

func (r *Remote) SendBytes(cmd []byte) {
    r.fptr.Write(cmd)
}

func (r *Remote) Open() (err error) {
    r.fptr, err = os.OpenFile(r.fname, os.O_WRONLY, os.ModeNamedPipe)
    return err
}

func (r *Remote) Close() {
    r.fptr.Close()
}

func (remote *Remote) Run(r *bufio.Reader) {
    fmt.Println("Enter Commands Below:")
    input := []byte{}
    err := error(nil)
    var keep_going bool = true
    for keep_going {
        input, err = r.ReadBytes('\n') 
        if err != nil {
            fmt.Println("error reading input:", err)
        }
        switch string(input) {
        case "exit\n":
            remote.exit_program = true
            remote.SendString("quit")
        default:
            remote.SendBytes(input)
        }
    }
}
