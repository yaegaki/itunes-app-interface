package itunes

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type itunes struct {
}

var baseScript = `
var app = Application("iTunes");
function p(/*...args*/) {
	console.log(Array.prototype.slice.call(arguments).map(encodeURIComponent).join(","));
}

function logTrack(track) {
	p(track.id(), track.name(), track.artist());
}

function findTrackById(id) {
	return app.tracks.byId(id);
}
`

var currentTrackScript = `
logTrack(app.currentTrack());
`

var getTracksScript = `
app.tracks().forEach(logTrack);
`

func execScript(script string) (chan string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
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

	_, err = io.WriteString(stdin, baseScript+script)
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return output, err
}

func decodeOutput(str string) []string {
	result := strings.Split(str, ",")
	for i, s := range result {
		s, err := url.QueryUnescape(s)
		if err != nil {
			s = ""
		}

		result[i] = s
	}

	return result
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

func getTrack(line string) (*track, error) {
	if line == "" {
		return nil, errors.New("result is empty.")
	}

	values := decodeOutput(line)
	count := len(values)

	id, err := strconv.Atoi(values[0])
	if err != nil {
		return nil, err
	}

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
		id:     id,
		Name:   name,
		Artist: artist,
	}

	return track, nil
}

func (it *itunes) CurrentTrack() (*track, error) {
	output, err := execScript(currentTrackScript)
	if err != nil {
		return nil, err
	}

	return getTrack(<-output)
}

func (it *itunes) GetTracks() (chan *track, error) {
	output, err := execScript(getTracksScript)
	if err != nil {
		return nil, err
	}

	result := make(chan *track)
	go func() {
		defer close(result)
		for line := range output {
			track, err := getTrack(line)
			if err == nil {
				result <- track
			}
		}
	}()

	return result, nil
}
