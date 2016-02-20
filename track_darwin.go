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

var playTrackScript = `
tell application "iTunes"
	set t to FindTrackByPersistentID("%v") of me
	if t is not null then
		play t
	end if
end tell
`

var getArtworksScript = `
tell application "iTunes"
    set t to FindTrackByName("%v") of me
    if t is not null then
        repeat with a in artworks of t
            set f to format of a
            P(f) of me
        end repeat
    end
end tell
`

func createTrack(line string) (*track, error) {
	if line == "" {
		return nil, errors.New("result is empty.")
	}

	values := decodeOutput(line)
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
			line, err = validateResult(line)
			if err != nil {
				log.Println(err)
				return
			}

			f := strings.Split(line, " ")[0]
			var format ArtworkFormat
			switch f {
			case JPEG.String():
				format = JPEG
			case PNG.String():
				format = PNG
			case BMP.String():
				format = BMP
			default:
				log.Printf("unknown format:%v", line)
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
