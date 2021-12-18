package remote

import (
    "io"
    "fmt"
)

type Remote struct {
    Pipe io.WriteCloser
}

func (r *Remote) SendString(s string) {
    fmt.Printf("sending '%s'", s)
    io.WriteString(r.Pipe, s)
}

func (r *Remote) SendBytes(bs []byte) {
    io.WriteString(r.Pipe, string(bs))
}
