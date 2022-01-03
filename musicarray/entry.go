package musicarray

type EntryType int

const (
	DirectoryEntry = iota
	SongEntry
)

type Entry struct {
	Type  EntryType
	Name  string
	Path  string
	Depth int
	Dir   Directory
	Song  Song
}
