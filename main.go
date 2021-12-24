package main

import (
	"github.com/StructsNotClasses/musicplayer/song"
	"github.com/StructsNotClasses/musicplayer/instance"

	gnc "github.com/rthornton128/goncurses"

	"log"
)

const PARENT_DIRECTORY = "/mnt/music"
const SONG_LIST_FILE = "/mnt/music/musicplayer/songs.json"

//script management can be done with the following technique:
// when a script is created, a number is incremented and passed to it as a variable named ID or something
// this number is also placed in an array along with potentially information about the script like name or file and the script runtime maybe
// there is then a thread safe int holding the desired reciever script of a message, and another variable (thread protected by the other one) holding a message to the tengo script
// all scripts can call a function and pass their ID value to it to recieve either the message, true or nil, false
// this system would allow the user to write scripts that take arbitrary user input, the main intention being to allow a script to be killed or otherwise managed using a simple interface
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
    program.Run()
}
