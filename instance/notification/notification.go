package notification

type Notification int

const (
    PlaybackBegan Notification = iota
    PlaybackEnded
)
