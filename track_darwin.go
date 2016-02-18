package itunes

import (
	"fmt"
)

type track struct {
	id     int
	Name   string
	Artist string
}

var playTrackScript = `
findTrackById(%d).play();
`

func (t *track) Play() error {
	o, err := execScript(fmt.Sprintf(playTrackScript, t.id))
	fmt.Println(<-o)
	return err
}
