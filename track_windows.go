package itunes

import (
	"fmt"
	"log"

	"github.com/yaegaki/go-ole-handler"
)

type track struct {
	handler  *olehandler.OleHandler
	artworks *olehandler.OleHandler

	itunes *itunes
	highID uint32
	lowID  uint32

	Album  string
	Artist string
	Name   string
}

func createTrack(it *itunes, handler *olehandler.OleHandler) (*track, error) {
	v, err := it.handler.GetProperty("ITObjectPersistentIDHigh", handler.Handle)
	if err != nil {
		return nil, err
	}
	highID := uint32(v.Val)

	v, err = it.handler.GetProperty("ITObjectPersistentIDLow", handler.Handle)
	if err != nil {
		return nil, err
	}
	lowID := uint32(v.Val)

	artworks, err := handler.GetOleHandler("Artwork")
	if err != nil {
		return nil, err
	}

	properties := [...]string{
		"Album", "Artist", "Name",
	}
	values := make([]string, len(properties))

	for i, property := range properties {
		v, err := handler.GetProperty(property)
		if err != nil {
			return nil, err
		}

		values[i] = v.ToString()
	}

	track := &track{
		handler:  handler,
		artworks: artworks,

		itunes: it,
		highID: highID,
		lowID:  lowID,

		Album:  values[0],
		Artist: values[1],
		Name:   values[2],
	}

	return track, nil
}

func (t *track) Close() {
	t.handler.Close()
}

func (t *track) Play() error {
	return t.handler.CallMethod("Play")
}

func (t *track) GetArtworks() (chan *artwork, error) {
	count, err := t.artworks.GetIntProperty("Count")
	if err != nil {
		return nil, err
	}

	output := make(chan *artwork)
	go func() {
		defer close(output)
		for i := 1; i <= count; i++ {
			var a *artwork
			err = t.artworks.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
				a, err = createArtwork(t, handler)
				return err
			}, i)

			if err != nil {
				log.Println(err)
				return
			}

			select {
			case <-t.handler.Closed():
				a.Close()
			case output <- a:
			}
		}
	}()

	return output, nil
}

func (t *track) PersistentID() string {
	return fmt.Sprintf("%x%x", t.highID, t.lowID)
}
