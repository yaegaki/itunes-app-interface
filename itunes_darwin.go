package itunes

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

type itunes struct {
}

// for compatibility
func Init() error {
	return nil
}

// for compatibility
func UnInit() {
	return
}

func CreateItunes() (*itunes, error) {
	return &itunes{}, nil
}

// for compatibility
func (_ *itunes) Close() {
	return
}

func (it *itunes) CurrentTrack() (*track, error) {
	columns, err := getColumnsByJS(`logTrack(app.currentTrack());`)
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}

func (_ *itunes) TrackCount() (int, error) {
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

func (_ *itunes) GetTrack(index int) (*track, error) {
	columns, err := getColumnsByJS(fmt.Sprintf("logTrack(app.tracks[%v]());", index))
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}

func (it *itunes) GetTracks() (chan *track, error) {
	output, err := execJS(`app.tracks().forEach(logTrack);`)
	if err != nil {
		return nil, err
	}

	result := make(chan *track)
	go func() {
		defer close(result)
		for line := range output {
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

func (it *itunes) FindTrackByPersistentID(persistentID string) (*track, error) {
	columns, err := getColumnsByJS(fmt.Sprintf(`logTrack(findTrackByPersistentId("%v"))`, persistentID))
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, errors.New(fmt.Sprintf("not found track:%v", persistentID))
	}

	return createTrack(columns)
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

func (it *itunes) Play() error {
	return callMethod("play")
}

func (it *itunes) Stop() error {
	return callMethod("stop")
}

func (it *itunes) BackTrack() error {
	return callMethod("backTrack")
}

func (it *itunes) PreviousTrack() error {
	return callMethod("previousTrack")
}

func (it *itunes) NextTrack() error {
	return callMethod("nextTrack")
}

func (it *itunes) SetPlayerPosition(pos float32) error {
	return putProperty("playerPosition", pos)
}

func (it *itunes) PlayerPosition() (float32, error) {
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

	return float32(result), nil
}

func (it *itunes) PlayerState() (PlayerState, error) {
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

func (it *itunes) PlayPause() error {
	return callMethod("playpause")
}

func (it *itunes) Pause() error {
	return callMethod("pause")
}

func (it *itunes) Resume() error {
	return callMethod("resume")
}

func (it *itunes) FastForward() error {
	return callMethod("fastForward")
}

func (it *itunes) Rewind() error {
	return callMethod("rewind")
}

func (it *itunes) SetSoundVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("volume is out of range")
	}

	return putProperty("soundVolume", volume)
}

func (it *itunes) SoundVolume() (int, error) {
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

func (it *itunes) SetMute(isMuted bool) error {
	return putProperty("mute", isMuted)
}

func (it *itunes) Mute() (bool, error) {
	v, err := getProperty("mute")
	if err != nil {
		return false, err
	}

	return v == "true", nil
}
