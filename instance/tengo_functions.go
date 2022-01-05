package instance

import (
    "github.com/StructsNotClasses/musicplayer/musicarray"

	"github.com/d5/tengo/v2"
	//"github.com/d5/tengo/v2/objects"

	"math/rand"
    "errors"
)

func (i *Instance) TengoSend(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if s, ok := args[0].(*tengo.String); ok {
		asString := s.String()
		cmdString := asString[1:len(asString)-1] + "\n" // tengo ".String()" returns a string value surrounded by quotes, so this needs to remove them before sending
		i.currentRemote.SendString(cmdString)
		return nil, nil
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "'send' argument",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
}

func (i *Instance) TengoSelectIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if v, ok := args[0].(*tengo.Int); ok {
		asInt := v.Value
		i.tree.Select(int(asInt))
		i.tree.Draw(i.client.treeWindow)
		return nil, nil
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "'selectIndex' argument",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}
}

func (i *Instance) TengoPlaySelected(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	err := i.PlayIndex(int(i.tree.currentIndex))
	return nil, err
}

func (i *Instance) TengoPlayIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if value, ok := args[0].(*tengo.Int); ok {
		err := i.PlayIndex(int(value.Value))
        return nil, err
	} else {
        return nil, tengo.ErrInvalidArgumentType{
            Name:     "'playIndex' argument",
            Expected: "int",
            Found:    args[0].TypeName(),
        }
    }
}

func (i *Instance) TengoSongCount(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	return &tengo.Int{Value: int64(len(i.tree.array))}, nil
}

func (i *Instance) TengoInfoPrint(args ...tengo.Object) (tengo.Object, error) {
	for _, item := range args {
		if value, ok := item.(*tengo.String); ok {
			i.client.commandOutputWindow.Print(value)
			i.client.commandOutputWindow.Refresh()
		}
	}
	return nil, nil
}

func (i *Instance) TengoCurrentIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	return &tengo.Int{Value: int64(i.tree.currentIndex)}, nil
}

// TengoRandomIndex returns a random number that is a valid song index. It requires random to already be seeded.
func (i *Instance) TengoRandomIndex(args ...tengo.Object) (tengo.Object, error) {
	rnum := rand.Int31n(int32(len(i.tree.array)))
	return &tengo.Int{Value: int64(rnum)}, nil
}

func (i *Instance) TengoSelectUp(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectUp()
	i.tree.Draw(i.client.treeWindow)
	return nil, nil
}

func (i *Instance) TengoSelectDown(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectDown()
	i.tree.Draw(i.client.treeWindow)
	return nil, nil
}

func (i *Instance) TengoSelectEnclosing(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectEnclosing(i.tree.currentIndex)
	i.tree.Draw(i.client.treeWindow)
	return nil, nil
}

func (i *Instance) TengoToggleDirExpansion(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
        index := int(value.Value)
        if i.tree.array[index].Type != musicarray.DirectoryEntry {
            return nil, errors.New("toggle: can only toggle directories.")
        }
        i.tree.array[index].Dir.ManuallyExpanded = !i.tree.array[index].Dir.ManuallyExpanded
        i.tree.Draw(i.client.treeWindow)
        return nil, nil
    } else {
        return nil, tengo.ErrInvalidArgumentType{
            Name:     "'toggle' argument",
            Expected: "int",
            Found:    args[0].TypeName(),
        }
    }
}

func (i *Instance) TengoIsDir(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
        index := int(value.Value)
        if i.tree.array[index].Type == musicarray.DirectoryEntry {
            return tengo.TrueValue, nil
        } else {
            return tengo.FalseValue, nil
        }
    } else {
        return nil, tengo.ErrInvalidArgumentType{
            Name:     "'toggle' argument",
            Expected: "int",
            Found:    args[0].TypeName(),
        }
    }
}

func (i *Instance) TengoSelectedIsDir(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

    if i.tree.array[i.tree.currentIndex].Type == musicarray.DirectoryEntry {
        return tengo.TrueValue, nil
    } else {
        return tengo.FalseValue, nil
    }
}
