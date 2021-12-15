package main

import (
    "log"
    "os"
    "path/filepath"
    "encoding/json"
    "io/ioutil"
)

type Song struct {
    Name string `json:"name"`
    Path string `json:"path"`
}

type SongList []Song

func (s *SongList) addFile(path string, info os.FileInfo, err error) error {
    if err != nil {
        log.Print(err)
        return nil
    }

    *s = append(*s, Song{info.Name(), path})

    return nil
}

func storeFileTree(parent_dir string, songListFile string) {
    file, err := os.Create(songListFile)
    if err != nil {
        log.Fatal(err)
    }

    s := SongList([]Song{})
    filepath.Walk(parent_dir, s.addFile)
    bytes, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        log.Fatal(err)
    }

    _, err = file.Write(bytes)
    if err != nil {
        log.Fatal(err)
    }
}

func createSongList(path string) (SongList, error) {
    s := SongList([]Song{})

    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        return s, err
    }

    err = json.Unmarshal(bytes, &s)
    return s, err
}
