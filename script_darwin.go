package itunes

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

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

const playTrackScript = `
tell application "iTunes"
	set t to FindTrackByPersistentID("%v") of me
	if t is not null then
		play t
	end if
end tell
`

const getArtworksScript = `
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

func getColumns(executor func(string) (chan string, error), script string) ([]string, error) {
	o, err := executor(script)
	if err != nil {
		return nil, err
	}

	result := <-o
	if result == "" {
		return []string{}, nil
	}

	columns, err := validateResult(result)
	if err != nil {
		return nil, err
	}

	return columns, nil
}

func getColumnsByAS(script string) ([]string, error) {
	return getColumns(execAS, script)
}

func getColumnsByJS(script string) ([]string, error) {
	return getColumns(execJS, script)
}
