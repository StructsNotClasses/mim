package remote

import (
	"io"
	"log"
)

type Remote struct {
	Pipe io.WriteCloser
}

func (r *Remote) SendString(s string) {
	r.Send([]byte(s))
}

func (r *Remote) Send(bs []byte) {
	if r == nil {
		log.Fatal("SendString: unable to send string using nil pointer to remote")
	}
	_, err := r.Pipe.Write(bs)
	if err != nil {
		log.Fatal(err)
	}
}
