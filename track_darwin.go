package itunes

import (
	"errors"
	"fmt"
)

type track struct {
	persistentID string
	Name         string
	Artist       string
}

var playTrackScript = `
tell application "iTunes"
	set t to FindTrackByPersistentID("%v") of me
	if t is not null then
		play t
	end if
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
	var name, artist string

	switch {
	case count > 1:
		name = values[1]
		fallthrough
	case count > 2:
		artist = values[2]
	}

	track := &track{
		persistentID: persistentID,
		Name:         name,
		Artist:       artist,
	}

	return track, nil
}

// for compatibility
func (_ *track) Close() {
}

func (t *track) Play() error {
	o, err := execAS(fmt.Sprintf(playTrackScript, t.persistentID))
	fmt.Println(<-o)
	return err
}
