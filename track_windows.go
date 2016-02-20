package itunes

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type track struct {
	itunes    *itunes
	highID    uint32
	lowID     uint32
	handle    *ole.IDispatch
	artworks  *ole.IDispatch
	wg        *sync.WaitGroup
	parent    *sync.WaitGroup
	closeChan chan bool

	Name   string
	Artist string
}

func createTrack(it *itunes, handle *ole.IDispatch, parent *sync.WaitGroup) (*track, error) {
	if handle == nil {
		return nil, errors.New("handle is nil")
	}

	parent.Add(1)

	v, err := it.app.GetProperty("ITObjectPersistentIDHigh", handle)
	if err != nil {
		return nil, err
	}
	highID := uint32(v.Val)

	v, err = it.app.GetProperty("ITObjectPersistentIDLow", handle)
	if err != nil {
		return nil, err
	}
	lowID := uint32(v.Val)

	properties := [...]string{
		"Name", "Artist",
	}
	values := make([]string, len(properties))

	for i, property := range properties {
		v, err := oleutil.GetProperty(handle, property)
		if err != nil {
			return nil, err
		}

		values[i] = v.ToString()
	}

	track := &track{
		itunes:    it,
		highID:    highID,
		lowID:     lowID,
		handle:    handle,
		wg:        new(sync.WaitGroup),
		parent:    parent,
		closeChan: make(chan bool),

		Name:   values[0],
		Artist: values[1],
	}

	return track, nil
}

func (t *track) Close() {
	close(t.closeChan)
	t.wg.Wait()

	t.handle.Release()
	if t.artworks != nil {
		t.artworks.Release()
	}

	t.parent.Done()
}

func (t *track) Play() error {
	t.wg.Add(1)
	defer t.wg.Done()
	_, err := t.handle.CallMethod("Play")
	return err
}

func (t *track) GetArtworks() (chan *artwork, error) {
	t.wg.Add(1)
	defer t.wg.Done()

	if t.artworks == nil {
		v, err := t.handle.GetProperty("Artwork")
		if err != nil {
			return nil, err
		}
		t.artworks = v.ToIDispatch()
	}

	v, err := t.artworks.GetProperty("Count")
	if err != nil {
		return nil, err
	}

	count := int(v.Val)

	output := make(chan *artwork)
	go func() {
		t.wg.Add(1)
		defer t.wg.Done()
		defer close(output)

		for i := 1; i <= count; i++ {
			v, err = t.artworks.GetProperty("Item", i)
			if err != nil {
				log.Println(err)
				return
			}
			artwork, err := createArtwork(v.ToIDispatch(), t.parent)
			if err != nil {
				log.Println(err)
				return
			}

			select {
			case <-t.closeChan:
				artwork.Close()
				return
			case output <- artwork:
			}
		}
	}()

	return output, nil
}

func (t *track) PersistentID() string {
	return fmt.Sprintf("%x%x", t.highID, t.lowID)
}
