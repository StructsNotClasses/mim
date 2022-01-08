package musicarray

import (
	"errors"
	"fmt"
	"io/fs"
    "regexp"
	"log"
	"os"
    "strings"
	"path/filepath"
)

type MusicArray []Entry

var fileRules = []string {
    `.*\.mp3`,
    `.*\.mp4`,
    `.*\.webm`,
    `.*\.mkv`,
    `.*\.flac`,
    `.*\.m4a`,
    `.*\.ogg`,
}


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

    containing := len(arr) - 1

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
            passes := false
            for _, pattern := range(fileRules) {
                matched, err := regexp.MatchString(pattern, entry.Name())
                if err == nil && matched {
                    passes = true
                    break
                } 
            }
            if passes {
                arr = append(arr, Entry{
                    Type:  SongEntry,
                    Name:  formatIfValid(entry.Name()),
                    Path:  root + "/" + entry.Name(),
                    Depth: depth + 1,
                    Song:  Song{},
                })
            } else {
                arr[containing].Dir.ItemCount--
            }
        }
    }

	return arr, nil
}

func formatIfValid(filename string) string {
    if !inFormat(filename) {
        return filename
    }

    // capitalize after the hyphen and replace "-" with " - " for readability
    atHyphens := strings.Split(filename, "-")
    for i, s := range(atHyphens) {
        atHyphens[i] = capitalizeFirst(s)
    }
    res := strings.Join(atHyphens, " - ")

    // capitalize after underscores and replace underscores with spaces
    atScores := strings.Split(res, "_")
    for i, s := range(atScores) {
        atScores[i] = capitalizeFirst(s)
    }
    res = strings.Join(atScores, " ")

    //remove file extension
    return strings.Split(res, ".")[0]
}

func inFormat(s string) bool {
    // following code basically checks if the filename is only lowercase, digits, and underscores
    // eg 10-foo_bar.mp3
    // only one hyphen is allowed, which should be used to denote number within an album
    periodCount := 0
    hyphenCount := 0
    for _, r := range(s) {
        if !isLower(r) && !isDigit(r) && r != rune('_') {
            if r == rune('.') {
                if periodCount == 1 {
                    return false
                } else {
                    periodCount++
                }
            } else if r == rune('-') {
                if hyphenCount == 1{
                    return false
                } else {
                    hyphenCount++
                }
            } else {
                // invalid character for formatting
                return false
            }
        }
    }
    return true
}

func isLower(r rune) bool {
    return r <= rune('z') && r >= rune('a')
}

func isDigit(r rune) bool {
    return r <= rune('9') && r >= rune('0')
}

func capitalizeFirst(s string) string {
    if len(s) == 0 {
        return s
    } else if len(s) == 1 {
        return strings.ToUpper(s)
    } else {
        first := s[0:1]
        return strings.ToUpper(first) + s[1:]
    }
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

func cleanUpFilename(fname string) string {
    words := strings.Split(fname, "_")
    /*
    for i := 0; i < len(words); i++ {
        if len(words[i]) >= 2 {
            words[i] = toUpperString(words[i][0]) + words[i][1:]
        } else if len(words[i]) == 0{

        } else {
            words[i] = toUpperString(words[i][0])
        }
    }
    */
    return strings.Join(words, " ")
}

func toUpperString(b byte) string {
    if b >= 'a' && b <= 'z' {
        return string([]byte{b + ('a' - 'A')})
    } else {
        return string([]byte{b})
    }
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
