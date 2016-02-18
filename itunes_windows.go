package itunes

import (
	"errors"
	"sync"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type itunes struct {
	unknwon   *ole.IUnknown
	app       *ole.IDispatch
	playlist  *ole.IDispatch
	tracks    *ole.IDispatch
	wg        sync.WaitGroup
	closeChan chan bool
}

func Init() error {
	return ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
}

func UnInit() {
	ole.CoUninitialize()
}

func CreateItunes() (*itunes, error) {
	obj, err := oleutil.CreateObject("iTunes.Application")
	if err != nil {
		return nil, err
	}

	handle, err := obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}

	it := &itunes{}
	it.unknwon = obj
	it.app = handle
	it.closeChan = make(chan bool)
	return it, nil
}

func (it *itunes) Close() {
	it.unknwon.Release()
	it.app.Release()
	if it.playlist != nil {
		it.playlist.Release()
	}

	if it.tracks != nil {
		it.tracks.Release()
	}

	close(it.closeChan)

	it.wg.Wait()
}

func (it *itunes) CurrentTrack() (*track, error) {
	v, err := it.app.GetProperty("CurrentTrack")
	if err != nil {
		return nil, err
	}

	return createTrack(v.ToIDispatch())
}

func createTrack(handle *ole.IDispatch) (*track, error) {
	if handle == nil {
		return nil, errors.New("handle is nil")
	}
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
		handle: handle,
		Name:   values[0],
		Artist: values[1],
	}

	return track, nil
}

func (it *itunes) GetTracks() (chan *track, error) {
	if it.playlist == nil {
		v, err := it.app.GetProperty("LibraryPlaylist")
		if err != nil {
			return nil, err
		}
		it.playlist = v.ToIDispatch()
	}

	if it.tracks == nil {
		v, err := it.playlist.GetProperty("Tracks")
		if err != nil {
			return nil, err
		}
		it.tracks = v.ToIDispatch()
	}

	v, err := it.tracks.GetProperty("Count")
	if err != nil {
		return nil, err
	}

	count := int(v.Val)

	output := make(chan *track)
	it.wg.Add(1)
	go func() {
		defer it.wg.Done()
		defer close(output)
		for i := 1; i <= count; i++ {
			v, err = it.tracks.GetProperty("Item", i)
			if err != nil {
				return
			}

			track, err := createTrack(v.ToIDispatch())
			if err != nil {
				return
			}

			select {
			case <-it.closeChan:
				track.Close()
				return
			case output <- track:
			}
		}
	}()

	return output, nil
}
