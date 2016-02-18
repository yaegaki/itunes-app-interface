package itunes

import (
	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"errors"
)

type itunes struct {
	handle *ole.IDispatch
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

	handle, err :=obj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, err
	}

	return &itunes{handle}, nil
}

func (it *itunes) CurrentTrack() (*track, error) {
	v, err := oleutil.GetProperty(it.handle, "CurrentTrack")
	if err != nil {
		return nil, err
	}

	return getTrack(v.ToIDispatch())
}

func getTrack(handle *ole.IDispatch) (*track, error) {
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
		Name: values[0],
		Artist: values[1],
	}

	return track, nil
}

func (it *itunes) GetTracks() (chan *track, error) {
	v, err := oleutil.GetProperty(it.handle, "LibraryPlaylist")
	if err != nil {
		return nil, err
	}

	v, err = oleutil.GetProperty(v.ToIDispatch(), "Tracks")
	if err != nil {
		return nil, err
	}

	trackHandle := v.ToIDispatch()

	v, err = oleutil.GetProperty(trackHandle, "Count")
	if err != nil {
		return nil, err
	}

	count := int(v.Val)

	output := make(chan *track)
	go func () {
		defer close(output)
		for i := 1; 1 <= count; i++ {
			v, err = oleutil.GetProperty(trackHandle, "Item", i)
			if err != nil {
				return
			}

			track, err := getTrack(v.ToIDispatch())
			if err != nil {
				return
			}

			output <- track
		}
	}()

	return output, nil
}

