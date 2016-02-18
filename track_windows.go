package itunes

import "github.com/go-ole/go-ole"

type track struct {
	handle *ole.IDispatch
	Name   string
	Artist string
}

func (t *track) Play() error {
	_, err := t.handle.CallMethod("Play")
	return err
}

func (t *track) Close() {
	t.handle.Release()
}
