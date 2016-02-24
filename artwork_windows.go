package itunes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaegaki/go-ole-handler"
)

type Artwork struct {
	handler *olehandler.OleHandler

	format ArtworkFormat
}

func (a *Artwork) Close() {
	a.handler.Close()
}

func createArtwork(t *Track, handler *olehandler.OleHandler) (*Artwork, error) {
	v, err := handler.GetIntProperty("Format")
	if err != nil {
		return nil, err
	}

	artwork := &Artwork{
		handler: handler,

		format: ArtworkFormat(v),
	}

	return artwork, nil
}

func (a *Artwork) SaveToFile(directory, name string) (string, error) {
	directory, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(directory)
	if err != nil {
		return "", err
	}

	filepath, err := filepath.Abs(fmt.Sprintf(`%v\%v%v`, directory, name, a.format.Ext()))
	if err != nil {
		return "", err
	}

	err = a.handler.CallMethod("SaveArtworkToFile", filepath)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
