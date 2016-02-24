package itunes

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type Playlist struct {
	persistentID string

	name string
}

func createPlaylist(values []string) (*Playlist, error) {
	persistentID := values[0]

	count := len(values)
	var name string

	switch {
	case count > 1:
		name = values[1]
	}

	p := &Playlist{
		persistentID: persistentID,

		name: name,
	}

	return p, nil
}

// for compatibility
func (p *Playlist) Close() {
}

func (p *Playlist) TrackCount() (int, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`p(findPlaylistByPersistentId("%v").tracks.length);`, p.persistentID))
	if err != nil {
		return 0, err
	}
	if len(columns) == 0 {
		return 0, errors.New("TrackCound is nil.")
	}

	count, err := strconv.ParseInt(columns[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (p *Playlist) GetTrack(index int) (t *Track, err error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logTrack(findPlaylistByPersistentId("%v").tracks[%v]());`, p.persistentID, index))
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}
func (p *Playlist) GetTracks() (chan *Track, error) {
	log.Println(p.GetTrack(0))
	o, err := execJS(fmt.Sprintf(`findPlaylistByPersistentId("%v").tracks().forEach(logTrack);`, p.persistentID))
	if err != nil {
		return nil, err
	}

	result := make(chan *Track)
	go func() {
		defer close(result)
		for line := range o {
			columns, err := validateResult(line)
			if err != nil {
				log.Println(err)
				return
			}

			track, err := createTrack(columns)
			if err != nil {
				log.Println(err)
				return
			}

			result <- track
		}
	}()

	return result, nil
}

func (p *Playlist) PersistentID() string {
	return p.persistentID
}

func (p *Playlist) PlayFirstTrack() error {
	_, err := getColumnsByJS(fmt.Sprintf(`findPlaylistByPersistentId("%v")`, p.persistentID))
	return err
}

func (p *Playlist) SetShuffle(isShuffle bool) error {
	return errors.New("SetShuffle is not support on OSX.")
}

func (p *Playlist) Shuffle() (bool, error) {
	return false, errors.New("Shuffle is not support on OSX.")
}

func (p *Playlist) AddTrack(t *Track) (result *Track, err error) {
	columns, err := getColumnsByAS(fmt.Sprintf(`P(AddTrackToPlaylist("%v", "%v"))`, t.persistentID, p.persistentID))
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, errors.New("AddTrackToPlaylist is failed.")
	}

	return findTrackByPersistentID(columns[0])
}

func (p *Playlist) Delete() error {
	_, err := getColumnsByJS(fmt.Sprintf(`findPlaylistByPersistentId("%v").delete()`, p.persistentID))

	return err
}
