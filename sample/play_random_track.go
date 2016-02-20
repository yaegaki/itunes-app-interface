package main

import (
	"log"
	"math/rand"
	"time"

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
	defer it.Close()

	c, err := it.TrackCount()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())
	t, err := it.GetTrack(rand.Intn(c) + 1)
	if err != nil {
		return err
	}
	defer t.Close()

	log.Printf("Play: %v %v %v", t.Name, t.Artist, t.Album)
	t.Play()

	return nil
}
