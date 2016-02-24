package itunes

import (
	"fmt"
	"os"
	"path/filepath"
)

type Artwork struct {
	track *Track
	index int

	format ArtworkFormat
}

// for compatibility
func (_ *Artwork) Close() {
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

	filepath, err := filepath.Abs(fmt.Sprintf(`%v/%v%v`, directory, name, a.Format.Ext()))
	if err != nil {
		return "", err
	}

	o, err := execAS(fmt.Sprintf(`SaveArtworkToFile("%v", %d, "%v")`, a.track.persistentID, a.index, filepath))

	if err != nil {
		return "", err
	}

	_, err = validateResult(<-o)
	if err != nil {
		return "", err
	}

	return filepath, nil
}
