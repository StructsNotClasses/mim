package musicarray

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type MusicArray []Entry

func New(rootPath string) (MusicArray, error) {
	arr, err := directoryToArray(rootPath, 0)
	if err != nil {
		return arr, err
	}
	return addDirectoryIndices(arr), nil
}

func directoryToArray(root string, depth int) (MusicArray, error) {
	var arr MusicArray

	entries, err := fs.ReadDir(os.DirFS(root), ".")
	if err != nil {
		return arr, err
	}

	arr = append(arr, Entry{
		Type:  DirectoryEntry,
		Name:  filepath.Base(root),
		Path:  root,
		Depth: depth,
		// previous and next directory indices will be properly initialized later
		Dir: Directory{
			ItemCount:          len(entries),
			PrevDirectoryIndex: -1,
		},
	})

	// add subdirectories in lexical order
	for _, entry := range entries {
		if entry.IsDir() {
			subdirArray, err := directoryToArray(root+"/"+entry.Name(), depth+1)
			if err != nil {
				return arr, err
			} else {
				arr = append(arr, subdirArray...)
			}
		}
	}

	// add files in lexical order
	for _, entry := range entries {
		if !entry.IsDir() {
			arr = append(arr, Entry{
				Type:  SongEntry,
				Name:  entry.Name(),
				Path:  root + "/" + entry.Name(),
				Depth: depth + 1,
				Song:  Song{},
			})
		}
	}

	return arr, nil
}

func addDirectoryIndices(arr MusicArray) MusicArray {
	_, err := arr.buildNextIndices(0)
	if err != nil {
		log.Fatal(err)
	}
	return arr
}

func (arr MusicArray) buildNextIndices(targetDirectoryIndex int) (int, error) {
	if arr[targetDirectoryIndex].Type != DirectoryEntry {
		return -1, errors.New(fmt.Sprintf("buildNextIndices: erroneously called on non-directory entry %v.", targetDirectoryIndex))
	}

	// the expected index will be incorrect if there are non-empty subdirectories
	expectedNextIndex := targetDirectoryIndex + arr[targetDirectoryIndex].Dir.ItemCount + 1

	// account for subdirs
	for i := targetDirectoryIndex + 1; ; {
		if i >= expectedNextIndex {
			break
		}
		if arr[i].Type == DirectoryEntry {
			subdirEntryCount, err := arr.buildNextIndices(i)
			if err != nil {
				return -1, err
			}
			expectedNextIndex += subdirEntryCount
			i += 1 + subdirEntryCount
		} else {
			i += 1
		}
	}

	if expectedNextIndex > len(arr) {
		return -1, errors.New(fmt.Sprintf("buildNextIndices: something fucky happened, most likely a directory has the wrong item count."))
	} else {
		arr[targetDirectoryIndex].Dir.EndDirectoryIndex = expectedNextIndex
		if expectedNextIndex != len(arr) && arr[expectedNextIndex].Type == DirectoryEntry && arr[expectedNextIndex].Depth == arr[targetDirectoryIndex].Depth {
			arr[expectedNextIndex].Dir.PrevDirectoryIndex = targetDirectoryIndex
		}
	}

	return expectedNextIndex - targetDirectoryIndex - 1, nil
}

func (arr MusicArray) Print() {
	for i, entry := range arr {
		fmt.Print(i, " ", entry.Depth, " ", entry.Path)
		if entry.Type == DirectoryEntry {
			fmt.Print(" ", entry.Dir.PrevDirectoryIndex, " ", entry.Dir.EndDirectoryIndex)
		}
		fmt.Println()
	}
}
