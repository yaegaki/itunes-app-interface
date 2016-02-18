package main

import (
	"github.com/yaegaki/itunes-app-interface"
	"log"
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

	for track := range output {
		log.Printf("name:%v artist:%v", track.Name, track.Artist)
	}

	return nil
}