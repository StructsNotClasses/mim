package main

import (
	"encoding/json"
	"github.com/StructsNotClasses/musicplayer/song"
	"log"
	"os"
	"path/filepath"
)

func storeFileTree(parent_dir string, songListFile string) {
	file, err := os.Create(songListFile)
	if err != nil {
		log.Fatal(err)
	}

	s := song.List([]song.Song{})
	filepath.Walk(parent_dir, s.AddFile)
	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}
