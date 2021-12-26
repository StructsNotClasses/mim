package main

import (
	"github.com/StructsNotClasses/musicplayer/song"
	"github.com/StructsNotClasses/musicplayer/instance"

	gnc "github.com/rthornton128/goncurses"

	"log"
)

const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"
const CONFIG = "/mnt/music/musicplayer/config.mim"

func main() {
	//current behavior is to regenerate the song list each run. probably needs to change
	storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE)

	//open the entire song list
	songs, err := song.CreateList(SONG_LIST_FILE)
	if err != nil {
		log.Fatal(err)
	}

	backgroundWindow, err := gnc.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer gnc.End()

	gnc.CBreak(true)
	//gnc.Keypad(true)
	gnc.Echo(false)
	backgroundWindow.Keypad(true)

    program := instance.New(backgroundWindow, songs)
    err = program.LoadConfig(CONFIG)
    if err != nil {
        log.Fatal(err)
    }
    program.Run()
}
