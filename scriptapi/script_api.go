package scriptapi

import (
    "github.com/StructsNotClasses/musicplayer/remote"
)

var RemoteToCurrentInstance remote.Remote

func SendStringToCurrentMplayerInstance(s string) {
    RemoteToCurrentInstance.SendString(s)
} 
