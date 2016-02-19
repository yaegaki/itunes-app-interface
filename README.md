# iTunes app interface
Cross Platform(OSX and Windows) iTunes application interface.

## Install
```
go get github.com/yaegaki/itunes-app-interface
```

## Sample
See also: [sample](https://github.com/yaegaki/itunes-app-interface/tree/master/sample)

### List all songs
```go
package main

import (
	"log"

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

	output, err := it.GetTracks()
	if err != nil {
		return err
	}

	for track := range output {
		log.Printf("name:%v artist:%v", track.Name, track.Artist)
		track.Close()
	}

	return nil
}
```

### NowPlaying
```go
package main

import (
	"errors"
	"log"

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

	t, err := it.CurrentTrack()
	if err != nil {
		return errors.New("Does not play song.")
	}
	defer t.Close()

	log.Printf("NowPlaying:%v %v", t.Name, t.Artist)

	artworks, err := t.GetArtworks()
	if err != nil {
		return err
	}

	artwork := <-artworks
	if artwork != nil {
		defer artwork.Close()
		path, err := artwork.SaveToFile("./", "nowplaying")
		if err != nil {
			return err
		}

		log.Printf("Save artwork to:%v", path)
	}

	return nil
}
```
