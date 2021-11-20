package main

import (
    "os"
)

type ModifiableWriter struct {
    output *os.File
}

func (w ModifiableWriter) Write(bs []byte) (n int, err error) {
    return w.output.Write(bs)
}

func (w ModifiableWriter) Close() error {
    //don't close fptr because this causes issues when it's stdout
    //needs to be fixed before used for other outputs
    return nil
}
