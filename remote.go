package main

import (
    "io"
    "fmt"
)

type Remote struct {
    pipe io.WriteCloser
    exit_program bool
}


func (r *Remote) SendString(s string) {
    fmt.Printf("sending '%s'", s)
    io.WriteString(r.pipe, s)
}

func (r *Remote) SendBytes(bs []byte) {
    io.WriteString(r.pipe, string(bs))
}
