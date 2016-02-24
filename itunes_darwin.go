package itunes

import (
	"errors"
	"fmt"
	"strconv"
)

type Itunes struct {
}

// for compatibility
func Init() error {
	return nil
}

// for compatibility
func UnInit() {
	return
}

func CreateItunes() (*Itunes, error) {
	return &Itunes{}, nil
}

// for compatibility
func (_ *Itunes) Close() {
	return
}

func (it *Itunes) CurrentTrack() (*Track, error) {
	columns, err := getColumnsByJS(`logTrack(app.currentTrack());`)
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}

func (_ *Itunes) TrackCount() (int, error) {
	columns, err := getColumnsByJS("p(app.tracks.length);")
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(columns[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (_ *Itunes) GetTrack(index int) (*Track, error) {
	columns, err := getColumnsByJS(fmt.Sprintf("logTrack(app.tracks[%v]());", index))
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}

func (it *Itunes) GetTracks() (chan *Track, error) {
	p, err := it.GetPlaylist(0)
	if err != nil {
		return nil, err
	}

	return p.GetTracks()
}

func findTrackByPersistentID(persistentID string) (*Track, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logTrack(findTrackByPersistentId("%v"))`, persistentID))
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, errors.New(fmt.Sprintf("not found track:%v", persistentID))
	}

	return createTrack(columns)
}

func (it *Itunes) FindTrackByPersistentID(persistentID string) (*Track, error) {
	return findTrackByPersistentID(persistentID)
}

func (_ *Itunes) CurrentPlaylist() (p *Playlist, err error) {
	columns, err := getColumnsByJS(`logPlaylist(app.currentPlaylist());`)
	if err != nil {
		return nil, err
	}

	return createPlaylist(columns)
}

func (_ *Itunes) PlaylistCount() (int, error) {
	columns, err := getColumnsByJS(`p(app.playlists.length);`)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(columns[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (_ *Itunes) GetPlaylist(index int) (*Playlist, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logPlaylist(app.playlists[%d]())`, index))
	if err != nil {
		return nil, err
	}

	return createPlaylist(columns)
}

func findPlaylistByPersistentID(persistentID string) (*Playlist, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logPlaylist(findPlaylistByPersistentId("%v"))`, persistentID))
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, errors.New(fmt.Sprintf("not found playlist:%v", persistentID))
	}

	return createPlaylist(columns)
}

func (_ *Itunes) FindPlaylistByPersistentID(persistentID string) (*Playlist, error) {
	return findPlaylistByPersistentID(persistentID)
}

func (_ *Itunes) CreatePlaylist(name string) (*Playlist, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logPlaylist(createPlaylist("%v"));`, name))
	if err != nil {
		return nil, err
	}

	return createPlaylist(columns)
}

func callMethod(method string) error {
	_, err := getColumnsByJS(fmt.Sprintf("app.%v()", method))
	if err != nil {
		return err
	}

	return nil
}

func getProperty(property string) (string, error) {
	columns, err := getColumnsByJS(fmt.Sprintf("p(app.%v())", property))
	if err != nil {
		return "", err
	}

	if len(columns) == 0 {
		return "", errors.New(fmt.Sprintf("%v is nil.", property))
	}

	return columns[0], nil
}

func putProperty(property string, v interface{}) error {
	_, err := getColumnsByJS(fmt.Sprintf("app.%v = '%v'", property, v))
	if err != nil {
		return err
	}

	return nil
}

func (it *Itunes) Play() error {
	return callMethod("play")
}

func (it *Itunes) Stop() error {
	return callMethod("stop")
}

func (it *Itunes) BackTrack() error {
	return callMethod("backTrack")
}

func (it *Itunes) PreviousTrack() error {
	return callMethod("previousTrack")
}

func (it *Itunes) NextTrack() error {
	return callMethod("nextTrack")
}

func (it *Itunes) SetPlayerPosition(pos int) error {
	return putProperty("playerPosition", pos)
}

func (it *Itunes) PlayerPosition() (int, error) {
	v, err := getProperty("playerPosition")
	if err != nil {
		return 0, err
	}

	if v == "null" {
		return 0, errors.New("PlayerPosition is nil")
	}

	result, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0, err
	}

	return int(result), nil
}

func (it *Itunes) PlayerState() (PlayerState, error) {
	v, err := getProperty("playerState")
	if err != nil {
		return PlayerState(0), err
	}

	var ps PlayerState
	switch v {
	case "playing":
		ps = Playing
	case "paused", "stopped":
		ps = Stopped
	case "fast forwarding":
		ps = FastForward
	case "rewinding":
		ps = Rewind
	default:
		return PlayerState(0), errors.New(fmt.Sprintf("unknown player state:%v", v))
	}

	return ps, nil
}

func (it *Itunes) PlayPause() error {
	return callMethod("playpause")
}

func (it *Itunes) Pause() error {
	return callMethod("pause")
}

func (it *Itunes) Resume() error {
	return callMethod("resume")
}

func (it *Itunes) FastForward() error {
	return callMethod("fastForward")
}

func (it *Itunes) Rewind() error {
	return callMethod("rewind")
}

func (it *Itunes) SetSoundVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("volume is out of range")
	}

	return putProperty("soundVolume", volume)
}

func (it *Itunes) SoundVolume() (int, error) {
	v, err := getProperty("soundVolume")
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(result), nil
}

func (it *Itunes) SetMute(isMuted bool) error {
	return putProperty("mute", isMuted)
}

func (it *Itunes) Mute() (bool, error) {
	v, err := getProperty("mute")
	if err != nil {
		return false, err
	}

	return v == "true", nil
}
