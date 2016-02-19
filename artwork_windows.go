package itunes

import (
	"errors"
	"sync"

	"github.com/go-ole/go-ole"
)

type ArtworkFormat int

const (
	Unknown ArtworkFormat = iota
	JPEG
	PNG
	BMP
)

func (a ArtworkFormat) String() string {
	switch a {
	case Unknown:
		return "Unknown"
	case JPEG:
		return "JPEG"
	case PNG:
		return "PNG"
	case BMP:
		return "BMP"
	}

	return ""
}

type artwork struct {
	handle *ole.IDispatch
	parent *sync.WaitGroup
}

func (a *artwork) Close() {
	a.handle.Release()
	a.parent.Done()
}

func createArtwork(handle *ole.IDispatch, parent *sync.WaitGroup) (*artwork, error) {
	if handle == nil {
		return nil, errors.New("handle is nil.")
	}
	parent.Add(1)

	artwork := &artwork{
		handle: handle,
		parent: parent,
	}
	return artwork, nil
}

func (a *artwork) Format() (ArtworkFormat, error) {
	v, err := a.handle.GetProperty("Format")
	if err != nil {
		return Unknown, err
	}
	return ArtworkFormat(int(v.Val)), nil
}
