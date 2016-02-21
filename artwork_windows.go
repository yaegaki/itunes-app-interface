package itunes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaegaki/go-ole-handler"
)

type artwork struct {
	handler *olehandler.OleHandler

	Format ArtworkFormat
}

func (a *artwork) Close() {
	a.handler.Close()
}

func createArtwork(t *track, handler *olehandler.OleHandler) (*artwork, error) {
	v, err := handler.GetIntProperty("Format")
	if err != nil {
		return nil, err
	}

	artwork := &artwork{
		handler: handler,

		Format: ArtworkFormat(v),
	}

	return artwork, nil
}

func (a *artwork) SaveToFile(directory, name string) (string, error) {
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

	err = a.handler.CallMethod("SaveArtworkToFile", filepath)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
