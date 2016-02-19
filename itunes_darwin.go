package itunes

import (
	"bufio"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

type itunes struct {
}

var baseJSScript = `
var app = Application("iTunes");
function p(/*...args*/) {
	console.log(Array.prototype.slice.call(arguments).map(encodeURIComponent).join(","));
}

function logTrack(track) {
	p(track.persistentID(), track.name(), track.artist());
}

function findTrackById(id) {
	return app.tracks.byId(id);
}
`

var baseASScript = `
on FindTrackByPersistentID(persistentID)
    tell application "iTunes"
        try
            return some track whose persistent ID is persistentID
        on error
            return null
        end try
    end tell
end FindTrackByPersistentID

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
`

var currentTrackScript = `
logTrack(app.currentTrack());
`

var getTracksScript = `
app.tracks().forEach(logTrack);
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

func execAS(script string) (chan string, error) {
	cmd := exec.Command("osascript")
	return execScript(cmd, baseASScript+script)
}

func execJS(script string) (chan string, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript")
	return execScript(cmd, baseJSScript+script)
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

func (it *itunes) CurrentTrack() (*track, error) {
	output, err := execJS(currentTrackScript)
	if err != nil {
		return nil, err
	}

	return createTrack(<-output)
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
			track, err := createTrack(line)
			if err == nil {
				result <- track
			}
		}
	}()

	return result, nil
}
