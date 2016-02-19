package itunes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-ole/go-ole"
)

type artwork struct {
	handle *ole.IDispatch
	wg     *sync.WaitGroup
	parent *sync.WaitGroup

	Format ArtworkFormat
}

func (a *artwork) Close() {
	a.wg.Wait()

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
		wg:     new(sync.WaitGroup),
		parent: parent,
		Format: ArtworkFormat(int(v.Val)),
	}

	return artwork, nil
}

func (a *artwork) SaveToFile(directory, name string) (string, error) {
	a.wg.Add(1)
	defer a.wg.Done()

	directory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(directory)
	if err != nil {
		return "", err
	}

	filepath, err := filepath.Abs(fmt.Sprintf(`%v\%v%v`, directory, name, a.Format.Ext()))
	if err != nil {
		return "", err
	}

	_, err = a.handle.CallMethod("SaveArtworkToFile", filepath)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
