package protector

var _ Protector = &empty{}

type empty struct{}

func (e empty) Entry(login string) (locked bool) {
	return false
}

func (e empty) Failure(login string) {
}
