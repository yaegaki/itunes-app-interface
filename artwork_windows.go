package itunes

import (
	"errors"
	"sync"

	"github.com/go-ole/go-ole"
)

type artwork struct {
	handle *ole.IDispatch
	parent *sync.WaitGroup

	Format ArtworkFormat
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

	v, err := handle.GetProperty("Format")
	if err != nil {
		parent.Done()
		return nil, err
	}

	artwork := &artwork{
		handle: handle,
		parent: parent,
		Format: ArtworkFormat(int(v.Val)),
	}

	return artwork, nil
}
