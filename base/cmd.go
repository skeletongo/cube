package base

type Command interface {
	Done(*Object) error
}

type CommandWrapper func(*Object) error

func (cw CommandWrapper) Done(o *Object) error {
	return cw(o)
}

type NilCommand struct {
}

func (n *NilCommand) Done(o *Object) error {
	return nil
}
