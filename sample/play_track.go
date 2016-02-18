package main

import (
	"log"
	"os"
	"strings"

	"github.com/yaegaki/itunes-app-interface"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal(`usage: go run example/play_track.go track_name`)
	}

	err := Test()
	if err != nil {
		log.Fatal(err)
	}
}

func s(str string) string {
	return strings.ToLower(strings.Replace(str, " ", "", -1))
}

func Test() error {
	itunes.Init()
	defer itunes.UnInit()
	it, err := itunes.CreateItunes()
	if err != nil {
		return err
	}
	defer it.Close()

	output, err := it.GetTracks()
	if err != nil {
		return err
	}

	// play track that contains `word` in the title.
	word := s(strings.Join(os.Args[1:], ""))
	for track := range output {
		name := track.Name
		if strings.Contains(s(name), word) {
			log.Printf("Play: %v", name)
			err = track.Play()
			track.Close()
			return err
		}
		track.Close()
	}

	return nil
}
