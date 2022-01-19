package instance

import (
	"github.com/d5/tengo/v2"

	"math/rand"
	"strings"
)

func (i *Instance) TengoSend(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	if s, ok := args[0].(*tengo.String); ok {
		asString := s.String()
		cmdString := asString[1:len(asString)-1] + "\n" // tengo ".String()" returns a string value surrounded by quotes, so this needs to remove them before sending
		i.mp.currentRemote.SendString(cmdString)
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
		i.tree.Draw()
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
	err := i.PlayIndex(int(i.tree.CurrentIndex()))
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
	return &tengo.Int{Value: int64(i.tree.ItemCount())}, nil
}

func (i *Instance) TengoInfoPrint(args ...tengo.Object) (tengo.Object, error) {
	for _, item := range args {
		if value, ok := item.(*tengo.String); ok {
			s := value.String()
			i.terminal.InfoPrint(s[1 : len(s)-1])
		} else if value, ok := item.(*tengo.Int); ok {
			v := value.Value
			i.terminal.InfoPrint(v)
		}
	}
	return nil, nil
}

func (i *Instance) TengoInfoPrintln(args ...tengo.Object) (tengo.Object, error) {
	obj, err := i.TengoInfoPrint(args...)
	i.terminal.InfoPrintln()
	return obj, err
}

func (i *Instance) TengoCurrentIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	return &tengo.Int{Value: int64(i.tree.CurrentIndex())}, nil
}

// TengoRandomIndex returns a random number that is a valid index. It requires random to already be seeded.
func (i *Instance) TengoRandomIndex(args ...tengo.Object) (tengo.Object, error) {
	rnum := rand.Int31n(int32(i.tree.ItemCount()))
	return &tengo.Int{Value: int64(rnum)}, nil
}

func (i *Instance) TengoSelectUp(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectUp()
	i.tree.Draw()
	return nil, nil
}

func (i *Instance) TengoSelectDown(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectDown()
	i.tree.Draw()
	return nil, nil
}

func (i *Instance) TengoSelectEnclosing(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	i.tree.SelectEnclosing(i.tree.CurrentIndex())
	i.tree.Draw()
	return nil, nil
}

func (i *Instance) TengoToggleDirExpansion(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
		index := int(value.Value)
		err := i.tree.Toggle(index)
		if err != nil {
			return nil, err
		}
		i.tree.Draw()
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
		if i.tree.IsDir(index) {
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

func (i *Instance) TengoDepth(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
		index := int(value.Value)
		return &tengo.Int{Value: int64(i.tree.Depth(index))}, nil
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

	if i.tree.CurrentIsDir() {
		return tengo.TrueValue, nil
	} else {
		return tengo.FalseValue, nil
	}
}

func (i *Instance) TengoIsExpanded(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
		index := int(value.Value)
		if i.tree.IsExpanded(index) {
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

func (i *Instance) TengoItemCount(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	return &tengo.Int{Value: int64(i.tree.ItemCount())}, nil
}

func (i *Instance) TengoSetSearch(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if ts, ok := args[0].(*tengo.String); ok {
		ss := strings.TrimSuffix(strings.TrimPrefix(ts.String(), "\""), "\"")
		i.tree.SetSearch(ss)
		return nil, nil
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "search value",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
}

// TengoNextMatch returns the index of the next match for the current search term starting from the index provided or -1 if none were found
func (i *Instance) TengoNextMatch(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
		starting := int(value.Value)
		match, exists := i.tree.NextMatch(starting)
		if exists {
			return &tengo.Int{Value: int64(match)}, nil
		} else {
			return &tengo.Int{Value: int64(-1)}, nil
		}
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "'nextMatch' argument",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}
}

// TengoPrevMatch returns the index of the next match backwards for the current search term starting from the index provided or -1 if none were found
func (i *Instance) TengoPrevMatch(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	if value, ok := args[0].(*tengo.Int); ok {
		starting := int(value.Value)
		match, exists := i.tree.PrevMatch(starting - 1)
		if exists {
			return &tengo.Int{Value: int64(match)}, nil
		} else {
			return &tengo.Int{Value: int64(-1)}, nil
		}
	} else {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "'prevMatch' argument",
			Expected: "int",
			Found:    args[0].TypeName(),
		}
	}
}

// TengoGetLine reads characters from the user until it gets a newline, which isn't passed to getline
func (i *Instance) TengoGetLine(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	line := i.GetLineBlocking()
	return &tengo.String{Value: line}, nil
}

// TengoGetChar reads a character from the user
func (i *Instance) TengoGetChar(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	char := i.GetCharBlocking()
	return &tengo.Char{Value: char}, nil
}
