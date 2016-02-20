package itunes

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type track struct {
	persistentID string

	Album  string
	Artist string
	Name   string
}

func createTrack(values []string) (*track, error) {
	if len(values) == 0 {
		return nil, errors.New("values is empty.")
	}

	count := len(values)

	persistentID := values[0]
	count = len(values)
	var album, artist, name string

	switch {
	case count > 1:
		album = values[1]
		fallthrough
	case count > 2:
		artist = values[2]
		fallthrough
	case count > 3:
		name = values[3]
	}

	track := &track{
		persistentID: persistentID,

		Album:  album,
		Artist: artist,
		Name:   name,
	}

	return track, nil
}

// for compatibility
func (_ *track) Close() {
}

func (t *track) Play() error {
	o, err := execAS(fmt.Sprintf(playTrackScript, t.persistentID))
	if err != nil {
		return nil
	}

	_, err = validateResult(<-o)
	if err != nil {
		return err
	}

	return nil
}

func (t *track) GetArtworks() (chan *artwork, error) {
	formats, err := execAS(fmt.Sprintf(getArtworksScript, t.Name))
	if err != nil {
		return nil, err
	}

	output := make(chan *artwork)
	go func() {
		defer close(output)
		index := 1
		for line := range formats {
			columns, err := validateResult(line)
			if err != nil {
				log.Println(err)
				return
			}

			f := strings.Split(columns[0], " ")[0]
			var format ArtworkFormat
			switch f {
			case JPEG.String():
				format = JPEG
			case PNG.String():
				format = PNG
			case BMP.String():
				format = BMP
			default:
				log.Printf("unknown format:%v", f)
				continue
			}

			output <- &artwork{
				track:  t,
				index:  index,
				Format: format,
			}
			index++
		}
	}()

	return output, nil
}

func (t *track) PersistentID() string {
	return t.persistentID
}
