package safebool

import (
    "sync"
)

type SafeBool struct {
    protected bool
    lock sync.Mutex 
}

func New(b bool) SafeBool {
    return SafeBool{
        protected: b,
    }
}

func (sb *SafeBool) Set(value bool) {
    sb.lock.Lock()
    defer sb.lock.Unlock()
    sb.protected = value
}

func (sb *SafeBool) Get() bool {
    sb.lock.Lock()
    defer sb.lock.Unlock()
    return sb.protected
}
