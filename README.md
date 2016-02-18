# iTunes app interface
Cross Platform(OSX and Windows) iTunes application interface.

## Install
```
go get github.com/yaegaki/itunes-app-interface
```

## Example
See also: [example](https://github.com/yaegaki/itunes-app-interface/tree/master/example)
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

	output, err := it.GetTracks()
	if err != nil {
		return err
	}

	for track := range output {
		log.Printf("name:%v artist:%v", track.Name, track.Artist)
	}

	return nil
}
```
