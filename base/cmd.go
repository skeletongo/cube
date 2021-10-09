package base

type Command interface {
	Done(o *Object) error
}

type CommandWrapper func(o *Object) error

func (cw CommandWrapper) Done(o *Object) error {
	return cw(o)
}

type NilCommand struct {
}

func (n *NilCommand) Done(o *Object) error {
	return nil
}
