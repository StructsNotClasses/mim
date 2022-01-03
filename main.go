package main

import (
	"github.com/StructsNotClasses/musicplayer/instance"
	"github.com/StructsNotClasses/musicplayer/musicarray"

	gnc "github.com/rthornton128/goncurses"

	"log"
)

const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"
const CONFIG = "/mnt/music/musicplayer/config.mim"

func main() {
	arr, err := musicarray.New(PARENT_DIRECTORY)
    /*
	if err != nil {
		log.Fatal(err)
	} else {
		arr.Print()
	}
    */

	backgroundWindow, err := gnc.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer gnc.End()

	gnc.CBreak(true)
	//gnc.Keypad(true)
	gnc.Echo(false)
	backgroundWindow.Keypad(true)

	program := instance.New(backgroundWindow, arr)
	err = program.LoadConfig(CONFIG)
	if err != nil {
		log.Fatal(err)
	}
	program.Run()
}
