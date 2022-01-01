package main

import (
	"github.com/StructsNotClasses/musicplayer/instance"
	"github.com/StructsNotClasses/musicplayer/musictree"
    

	gnc "github.com/rthornton128/goncurses"

	"log"
)

const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"
const CONFIG = "/mnt/music/musicplayer/config.mim"

func main() {
    tree, err := musictree.New(PARENT_DIRECTORY)
    if err != nil {
        log.Fatal(err)
    }

    tree.Root.Print(0)
	//current behavior is to regenerate the song list each run. probably needs to change
	//storeFileTree(PARENT_DIRECTORY, SONG_LIST_FILE)

    /*
	//open the entire song list
	songs, err := song.CreateList(SONG_LIST_FILE)
	if err != nil {
		log.Fatal(err)
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

    program := instance.New(backgroundWindow, tree)
    err = program.LoadConfig(CONFIG)
    if err != nil {
        log.Fatal(err)
    }
    program.Run()
}
