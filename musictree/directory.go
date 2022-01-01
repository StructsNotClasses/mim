package musictree

import (
    "os"
    "fmt"
    "strings"
)

type Directory struct {
    Name string
    Path string
    IsExpanded bool
    Subdirectories []Directory
    Songs []Song
}

func CreateDirectory(path string) (d Directory, err error) {
    d.Path = path
    split := strings.Split(path, "/")
    d.Name = split[len(split) - 1]
    entries, err := os.ReadDir(path)
    fmt.Println(entries)
    if err != nil {
        return
    }

    for _, entry := range(entries) {
        info, e := entry.Info()
        if e != nil {
            return d, e
        }
        name := info.Name()
        if entry.IsDir() {
            subdir, e := CreateDirectory(path + "/" + name)
            if e != nil {
                return d, e
            }
            d.Subdirectories = append(d.Subdirectories, subdir)
        } else {
            d.Songs = append(d.Songs, Song{
                Name: name,
                Path: path + "/" + name,
            })
        }
    }

    return
}

func (d *Directory) Print(depth int) {
    printSpaces(depth)
    fmt.Println(d.Name + ":")
    for _, file := range d.Songs {
        printSpaces(depth + 1)
        fmt.Println(file.Path)       
    }
    for _, dir := range d.Subdirectories {
        dir.Print(depth + 1)
    }
}

func printSpaces(count int) {
    for i := 0; i < count; i++ {
        fmt.Print(" ")
    }
}
