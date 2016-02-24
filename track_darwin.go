package itunes

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type Track struct {
	persistentID string

	album  string
	artist string
	name   string
}

func createTrack(values []string) (*Track, error) {
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

	track := &Track{
		persistentID: persistentID,

		album:  album,
		artist: artist,
		name:   name,
	}

	return track, nil
}

// for compatibility
func (_ *Track) Close() {
}

func (t *Track) Play() error {
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

func (t *Track) GetArtworks() (chan *Artwork, error) {
	formats, err := execAS(fmt.Sprintf(getArtworksScript, t.Name))
	if err != nil {
		return nil, err
	}

	output := make(chan *Artwork)
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

			output <- &Artwork{
				Track:  t,
				index:  index,
				format: format,
			}
			index++
		}
	}()

	return output, nil
}

func (t *Track) PersistentID() string {
	return t.persistentID
}
