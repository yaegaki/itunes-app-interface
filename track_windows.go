package itunes

import "github.com/go-ole/go-ole"

type track struct {
	handle *ole.IDispatch
	Name   string
	Artist string
}
