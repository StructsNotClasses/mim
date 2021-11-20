package main

type AsyncWriter chan byte

func (w AsyncWriter) Write(p []byte) (n int, err error) {
    for _, b := range p {
        w <- b
    }
    return len(p), nil
}

func (w AsyncWriter) Close() {
    close(w)
}
