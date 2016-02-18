package main

import (
	"log"
	"strings"

	"github.com/yaegaki/itunes-app-interface"
)

func main() {
	err := Test()
	if err != nil {
		log.Fatal(err)
	}
}

func Test() error {
	itunes.Init()
	defer itunes.UnInit()
	it, err := itunes.CreateItunes()
	if err != nil {
		return err
	}

	output, err := it.GetTracks()
	if err != nil {
		return err
	}

	// play track that contains the "love" in the title.
	for track := range output {
		if strings.Contains(strings.ToLower(track.Name), "love") {
			track.Play()
			break
		}
	}

	return nil
}
