package itunes

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type track struct {
	handle *ole.IDispatch
	Name   string
	Artist string
}

func (t *track) Play() error {
	_, err := oleutil.CallMethod(t.handle, "Play")
	return err
}
