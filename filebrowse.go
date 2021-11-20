package main

import (
    "fmt"
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

func storeFileTree(parent_dir string, song_list_file string) {
    file, err := os.Create(song_list_file)
    if err != nil {
        fmt.Println("unable to open file", song_list_file)
        log.Fatal(err)
    }

    song_list := SongList([]Song{})
    filepath.Walk(parent_dir, song_list.addFile)
    bytes, err := json.MarshalIndent(song_list, "", "  ")
    if err != nil {
        fmt.Println("unable to convert song list to json")
        log.Fatal(err)
    }

    _, err = file.Write(bytes)
    if err != nil {
        fmt.Println("unable to write song list to file", song_list_file)
        log.Fatal(err)
    }
}

func openSongList(path string) (SongList, error) {
    s := SongList([]Song{})

    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        fmt.Println("unable to open or read file", path)
        return s, err
    }

    err = json.Unmarshal(bytes, &s)
    if err != nil {
        fmt.Println("unable to convert json to song list")
        return s, err
    }

    return s, nil
}
