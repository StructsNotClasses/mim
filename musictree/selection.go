package musictree

import (
    "fmt"
    "errors"
)

type AccessSequence []int

func (t *MusicTree) SelectIndex(index int) error {
    if index >= len(t.OrderedArray) {
        return errors.New(fmt.Sprintf("Cannot select index %d. There aren't this many entries.", index))
    } else if index < 0 {
        return errors.New(fmt.Sprintf("Cannot select a negative index.", index))
    } 

    access, err := t.getAccessSequenceForIndex(index)
    if err != nil {
        return err
    }

    t.CurrentIndex = index
    t.CurrentAccessSequence = access
    return nil
}

func (t *MusicTree) getAccessSequenceForIndex(i int) (AccessSequence, error) {
    path, err := t.OrderedArray[i].Path()
    if err != nil {
        return AccessSequence([]int{}), err
    }
    series, ok := recursivePathSearch(t.Root, path)
    if !ok {
        return AccessSequence([]int{}), errors.New(fmt.Sprintf("Directory or file '%s' is not present in the tree.", path))
    }
    access := AccessSequence(append([]int{0}, series...))
    return access, nil
}

func recursivePathSearch(dir Directory, path string) ([]int, bool) {
    if dir.Path == path {
        return []int{}, true
    }
    
    for i, subdir := range(dir.Subdirectories) {
        access, isPresent := recursivePathSearch(subdir, path)
        if isPresent {
            list := []int{i}
            return append(list, access...), true
        }
    }

    for i, file := range(dir.Songs) {
        if file.Path == path {
            return []int{i + len(dir.Subdirectories)}, true
        }
    }
    return []int{}, false
    
}
