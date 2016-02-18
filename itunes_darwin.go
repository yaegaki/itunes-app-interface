package itunes

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

type itunes struct {
}

var baseScript = `
function p(/*...args*/) {
	console.log(Array.prototype.slice.call(arguments).map(encodeURIComponent).join(","));
}

function logTrack(track) {
	p(track.name(), track.artist());
}
var app = Application("iTunes");
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

func getTrack(line string) (*track, error) {
	if line == "" {
		return nil, errors.New("result is empty.")
	}

	values := decodeOutput(line)
	count := len(values)

	name := values[0]
	var artist = ""

	if count > 1 {
		artist = values[1]
	}

	track := &track{
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
