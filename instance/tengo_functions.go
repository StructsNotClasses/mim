package instance

import (
	"github.com/d5/tengo/v2"
	//"github.com/d5/tengo/v2/objects"

	"math/rand"
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
			i.commandHandling.InfoPrint(value)
		}
	}
	return nil, nil
}

func (i *Instance) TengoCurrentIndex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	return &tengo.Int{Value: int64(i.tree.CurrentIndex())}, nil
}

// TengoRandomIndex returns a random number that is a valid song index. It requires random to already be seeded.
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
