package musictree

import (
    "errors"
)

type EntryType int

const (
    DirectoryEntry = iota
    SongEntry
)

type Entry struct {
    Type EntryType
    DPtr *Directory
    SPtr *Song
}

type MusicArray []Entry

func (e *Entry) Name() (string, error) {
    switch e.Type {
    case DirectoryEntry:
        if e.DPtr == nil {
            return "", errors.New("entry.Name(): Expected directory but entry is likely a song. (DPtr is nil)")
        }
        return e.DPtr.Name, nil
    case SongEntry:
        if e.SPtr == nil {
            return "", errors.New("entry.Name(): Expected song but entry is likely a directory. (SPtr is nil)")
        }
        return e.SPtr.Name, nil
    default:
        return "", errors.New("entry.Name(): Unexpected value for entry type. This is likely an internal error.")
    }
}

func (e *Entry) Path() (string, error) {
    switch e.Type {
    case DirectoryEntry:
        if e.DPtr == nil {
            return "", errors.New("entry.Path(): Expected directory but entry is likely a song. (DPtr is nil)")
        }
        return e.DPtr.Path, nil
    case SongEntry:
        if e.SPtr == nil {
            return "", errors.New("entry.Path(): Expected song but entry is likely a directory. (SPtr is nil)")
        }
        return e.SPtr.Path, nil
    default:
        return "", errors.New("entry.Name(): Unexpected value for entry type. This is likely an internal error.")
    }
}

func getUnsortedNodes(root *Directory) (arr MusicArray) {
    arr = append(arr, Entry{
        Type: DirectoryEntry,
        DPtr: root,
        SPtr: nil,
    })

    for i := 0; i < len(root.Subdirectories); i++ {
        arr = append(arr, getUnsortedNodes(&root.Subdirectories[i])...)
    }
    for i := 0; i < len(root.Songs); i++ {
        arr = append(arr, Entry{
            Type: SongEntry,
            DPtr: nil,
            SPtr: &root.Songs[i],
        })
    }
    return
}

func alphabeticalSort(arr MusicArray) MusicArray {
    return arr
}
