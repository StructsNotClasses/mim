package playback

import (
    "github.com/StructsNotClasses/mim/remote"
)

type Notification int

const (
    Began Notification = iota
    Ended
)

type PlaybackState struct {
    PlaybackInProgress bool
    mplayerRemote remote.Remote
}

func (pbs *PlaybackState) Remote() (r remote.Remote, ok bool) {
    r = pbs.mplayerRemote
    ok = pbs.PlaybackInProgress
    return
}

func (pbs *PlaybackState) Receive(signals chan Notification) {
    select {
    case signal := <-signals:
        pbs.setFromNotification(signal)
    default:
        return
    }
}

func (pbs *PlaybackState) ReceiveBlocking(signals chan Notification) {
    n := <-signals
    pbs.setFromNotification(n)
}

func (pbs *PlaybackState) setFromNotification(n Notification) {
    switch n {
    case Began:
        pbs.PlaybackInProgress = true
    case Ended:
        pbs.PlaybackInProgress = false
    }
}
