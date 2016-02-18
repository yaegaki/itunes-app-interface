package itunes

import (
	"errors"
	"log"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type itunes struct {
	unknwon  *ole.IUnknown
	app      *ole.IDispatch
	playlist *ole.IDispatch
	tracks   *ole.IDispatch
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

	return &itunes{unknwon: obj, app: handle}, nil
}

func (it *itunes) Close() {
	it.unknwon.Release()
	it.app.Release()

	if it.tracks != nil {
		it.tracks.Release()
		it.tracks = nil
	}

	if it.playlist != nil {
		it.playlist.Release()
		it.playlist = nil
	}
}

func (it *itunes) CurrentTrack() (*track, error) {
	v, err := it.app.GetProperty("CurrentTrack")
	if err != nil {
		return nil, err
	}
	handle := v.ToIDispatch()
	if handle == nil {
		return nil, errors.New("CurrentTrack is none.")
	}
	defer handle.Release()

	return createTrack(it, handle)
}

func createTrack(it *itunes, handle *ole.IDispatch) (*track, error) {
	v, err := it.app.GetProperty("ITObjectPersistentIDHigh", handle)
	if err != nil {
		return nil, err
	}
	highID := int(v.Val)

	v, err = it.app.GetProperty("ITObjectPersistentIDLow", handle)
	if err != nil {
		return nil, err
	}
	lowID := int(v.Val)

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
		it:     it,
		highID: highID,
		lowID:  lowID,
		Name:   values[0],
		Artist: values[1],
	}

	return track, nil
}

func (it *itunes) GetTracks() (chan *track, error) {
	var (
		v   *ole.VARIANT
		err error
	)

	if it.tracks == nil {
		v, err = it.app.GetProperty("LibraryPlaylist")
		if err != nil {
			return nil, err
		}
		it.playlist = v.ToIDispatch()

		v, err = it.playlist.GetProperty("Tracks")
		if err != nil {
			return nil, err
		}
		it.tracks = v.ToIDispatch()
	}

	v, err = it.tracks.GetProperty("Count")
	if err != nil {
		return nil, err
	}

	count := int(v.Val)

	output := make(chan *track)
	go func() {
		defer close(output)
		for i := 1; i <= count; i++ {
			v, err = it.tracks.GetProperty("Item", i)
			if err != nil {
				log.Println(err)
				return
			}
			handle := v.ToIDispatch()

			track, err := createTrack(it, handle)
			handle.Release()
			if err != nil {
				log.Println(err)
				return
			}

			output <- track
		}
	}()

	return output, nil
}
