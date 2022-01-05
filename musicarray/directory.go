package musicarray

type Directory struct {
	ManuallyExpanded   bool
	AutoExpanded       bool
	PrevDirectoryIndex int
	EndDirectoryIndex  int
	ItemCount          int
}

func (d Directory) Expanded() bool {
    return d.ManuallyExpanded || d.AutoExpanded
}
