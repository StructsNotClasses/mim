package main

import (
    "github.com/StructsNotClasses/musicplayer/remote"
)

var RemoteToCurrentInstance Remote

func sendStringToCurrentMplayerInstance(s string) {
    RemoteToCurrentInstance.SendString(s)
} 
