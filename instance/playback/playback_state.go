package playback

import (
    "github.com/StructsNotClasses/mim/remote"
    "github.com/StructsNotClasses/mim/instance/notification"
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

func (pbs *PlaybackState) Receive(notificationChannel chan notification.Notification) {
    select {
    case n := <-notificationChannel:
        pbs.setFromNotification(n)
    default:
        return
    }
}

func (pbs *PlaybackState) ReceiveBlocking(notificationChannel chan notification.Notification) {
    n := <-notificationChannel
    pbs.setFromNotification(n)
}

func (pbs *PlaybackState) setFromNotification(n notification.Notification) {
    switch n {
    case notification.PlaybackBegan:
        pbs.PlaybackInProgress = true
    case notification.PlaybackEnded:
        pbs.PlaybackInProgress = false
    }
}
