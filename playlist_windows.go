package itunes

import (
	"fmt"
	"log"

	"github.com/yaegaki/go-ole-handler"
)

type Playlist struct {
	handler *olehandler.OleHandler

	itunes *itunes
	tracks *olehandler.OleHandler
	highID uint32
	lowID  uint32
	Name   string
}

func createPlaylist(it *itunes, handler *olehandler.OleHandler) (*Playlist, error) {
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
	name, err := handler.GetStringProperty("Name")
	if err != nil {
		return nil, err
	}

	tracks, err := handler.GetOleHandler("Tracks")
	if err != nil {
		return nil, err
	}

	p := &Playlist{
		handler: handler,

		itunes: it,
		tracks: tracks,
		highID: highID,
		lowID:  lowID,
		Name:   name,
	}

	return p, nil
}

func (p *Playlist) Close() {
	p.handler.Close()
}

func (p *Playlist) TrackCount() (int, error) {
	return p.tracks.GetIntProperty("Count")
}

func (p *Playlist) GetTrack(index int) (t *track, err error) {
	err = p.tracks.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
		t, err = createTrack(p.itunes, handler)
		return err
	}, index+1)

	return t, err
}
func (p *Playlist) GetTracks() (chan *track, error) {
	count, err := p.TrackCount()
	if err != nil {
		return nil, err
	}

	output := make(chan *track)
	go func() {
		defer close(output)
		for i := 1; i <= count; i++ {
			var t *track
			err = p.tracks.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
				t, err = createTrack(p.itunes, handler)
				return err
			}, i)

			if err != nil {
				log.Println(err)
				return
			}

			select {
			case <-p.handler.Closed():
				t.Close()
			case output <- t:
			}
		}
	}()

	return output, nil
}

func (p *Playlist) PersistentID() string {
	return fmt.Sprintf("%x%x", p.highID, p.lowID)
}

func (p *Playlist) PlayFirstTrack() error {
	return p.handler.CallMethod("PlayFirstTrack")
}

func (p *Playlist) SetShuffle(isShuffle bool) error {
	return p.handler.PutProperty("Shuffle", isShuffle)
}
func (p *Playlist) Shuffle() (bool, error) {
	return p.handler.GetBoolProperty("Shuffle")
}

func (p *Playlist) AddTrack(t *track) (result *track, err error) {
	err = p.handler.GetOleHandlerWithCallbackAndArgsByMethod("AddTrack", func(handler *olehandler.OleHandler) error {
		result, err = createTrack(p.itunes, handler)
		return err
	}, t.handler.Handle)

	return result, err
}

func (p *Playlist) Delete() error {
	return p.handler.CallMethod("Delete")
}
