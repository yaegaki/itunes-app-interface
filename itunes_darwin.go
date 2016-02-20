package itunes

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type itunes struct {
}

var baseJScript = `
var app = Application("iTunes");
function p(/*...args*/) {
	console.log("!"+Array.prototype.slice.call(arguments).map(encodeURIComponent).join(","));
}

function logTrack(track) {
	if (track != null) {
		p(
			track.persistentID(),
			track.album(),
			track.artist(),
			track.name()
		);
	}
}

function findTrackById(id) {
	return app.tracks.byId(id);
}

function findTrackByPersistentId(persistentId) {
	var index = app.tracks.persistentID().indexOf(persistentId);
	if (index < 0) {
		return null;
	}

	return app.tracks[index];
}
`

var baseAScript = `
on P(o)
	log "!" & o
end

on FindTrackByPersistentID(persistentID)
    tell application "iTunes"
        try
            return some track whose persistent ID is persistentID
        on error
            return null
        end try
    end tell
end

on FindTrackByName(n)
    tell application "iTunes"
    	set l to search playlist 1 for n only songs
    	set c to count of l
    	if c is 0 then
    		return null
		else
			return item 1 of l
		end if
    end tell
end

on SaveArtworkToFile(persistentID, index, path)
    set fp to POSIX file path
	tell application "iTunes"
		set t to FindTrackByPersistentID(persistentID) of me
		if t is not null then
			set art to artworks index of t
			set d to raw data of art
			set f to open for access fp with write permission
			set eof f to 0
			write d to f
			close access f
		end
	end tell
end
`

var currentTrackScript = `
logTrack(app.currentTrack());
`

var getTracksScript = `
app.tracks().forEach(logTrack);
`

var findTrackByPersistentIdScript = `
logTrack(findTrackByPersistentId("%v"))
`

func execScript(cmd *exec.Cmd, script string) (chan string, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer stdin.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	output := make(chan string)
	go func() {
		defer close(output)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				output <- line
			}
		}
	}()

	_, err = io.WriteString(stdin, script)
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return output, err
}

func validateResult(result string) ([]string, error) {
	l := len(result)
	if l != 0 && result[0] != "!"[0] {
		return nil, errors.New(fmt.Sprintf("osascript error:%v", result))
	}

	if l != 0 {
		result = result[1:]
	}

	s := strings.Split(result, ",")
	for i, raw := range s {
		raw, err := url.QueryUnescape(raw)
		if err != nil {
			return nil, err
		}

		s[i] = raw
	}

	return s, nil
}

func execAS(script string) (chan string, error) {
	cmd := exec.Command("osascript")
	return execScript(cmd, baseAScript+script)
}

func execJS(script string) (chan string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	return execScript(cmd, baseJScript+script)
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
	o, err := execJS(currentTrackScript)
	if err != nil {
		return nil, err
	}

	result := <-o
	if len(result) == 0 {
		return nil, errors.New("CurrentTrack is nil")
	}

	columns, err := validateResult(result)
	if err != nil {
		return nil, err
	}

	return createTrack(columns)
}

func (it *itunes) GetTracks() (chan *track, error) {
	output, err := execJS(getTracksScript)
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
			if err == nil {
				result <- track
			}
		}
	}()

	return result, nil
}

func (it *itunes) FindTrackByPersistentID(persistentID string) (*track, error) {
	o, err := execJS(fmt.Sprintf(findTrackByPersistentIdScript, persistentID))
	if err != nil {
		return nil, err
	}

	result := <-o
	if result == "" {
		return nil, errors.New(fmt.Sprintf("not found track:%v", persistentID))
	}

	columns, err := validateResult(result)
	if err != nil {
		return nil, err
	}

	t, err := createTrack(columns)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func callMethod(method string) error {
	o, err := execJS(fmt.Sprintf("app.%v()", method))
	if err != nil {
		return err
	}

	result := <-o
	if _, err = validateResult(result); err != nil {
		return err
	}

	return nil
}

func getProperty(property string) (string, error) {
	o, err := execJS(fmt.Sprintf("p(app.%v())", property))
	if err != nil {
		return "", err
	}

	result := <-o
	columns, err := validateResult(result)
	if err != nil {
		return "", err
	}

	if len(columns) == 0 {
		return "", errors.New(fmt.Sprintf("%v is nil.", property))
	}

	return columns[0], nil
}

func putProperty(property string, v interface{}) error {
	o, err := execJS(fmt.Sprintf("app.%v = '%v'", property, v))
	if err != nil {
		return err
	}

	result := <-o
	if _, err = validateResult(result); err != nil {
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
