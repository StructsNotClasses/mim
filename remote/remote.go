package remote

import (
    "io"
    "github.com/d5/tengo/v2"
)

type Remote struct {
    Pipe io.WriteCloser
}

func (r *Remote) SendString(s string) {
    io.WriteString(r.Pipe, s)
}

func (r *Remote) SendBytes(bs []byte) {
    io.WriteString(r.Pipe, string(bs))
}

func (remote *Remote) Send(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments

	}
	s, ok := args[0].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "string argument",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	asString := s.String()
	cmdString := asString[1:len(asString)-1] + "\n"
	remote.SendString(cmdString)
	return nil, nil
}
