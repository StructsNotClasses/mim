package script

import (
	"github.com/d5/tengo/v2"
)

type Script struct {
    name string
    contents []byte
    bytecode *tengo.Compiled
}

func New(name string, contents []byte, bytecode *tengo.Compiled) Script {
    return Script{
        name: name,
        contents: contents,
        bytecode: bytecode,
    }
}

// script.Name returns the name of the script if it is non-empty and the contents otherwise
func (s Script) Name() string {
    if s.name == "" {
        return string(s.contents)
    } else {
        return s.name
    }
}

func (s Script) Run() error {
    return s.bytecode.Run()
}

func (s Script) IsEmpty() bool {
    return len(s.contents) == 0
}
