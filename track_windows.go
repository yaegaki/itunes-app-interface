package itunes

import (
	"fmt"
	"log"

	"github.com/yaegaki/go-ole-handler"
)

type Track struct {
	handler  *olehandler.OleHandler
	artworks *olehandler.OleHandler

	itunes *Itunes
	highID uint32
	lowID  uint32

	album  string
	artist string
	name   string
}

func createTrack(it *Itunes, handler *olehandler.OleHandler) (*Track, error) {
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

	track := &Track{
		handler:  handler,
		artworks: artworks,

		itunes: it,
		highID: highID,
		lowID:  lowID,

		album:  values[0],
		artist: values[1],
		name:   values[2],
	}

	return track, nil
}

func (t *Track) Close() {
	t.handler.Close()
}

func (t *Track) Play() error {
	return t.handler.CallMethod("Play")
}

func (t *Track) GetArtworks() (chan *Artwork, error) {
	count, err := t.artworks.GetIntProperty("Count")
	if err != nil {
		return nil, err
	}

	output := make(chan *Artwork)
	go func() {
		defer close(output)
		for i := 1; i <= count; i++ {
			var a *Artwork
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

func (t *Track) PersistentID() string {
	return fmt.Sprintf("%x%x", t.highID, t.lowID)
}
