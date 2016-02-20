package itunes

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type itunes struct {
	unknwon   *ole.IUnknown
	app       *ole.IDispatch
	playlist  *ole.IDispatch
	tracks    *ole.IDispatch
	wg        *sync.WaitGroup
	pm        *sync.Mutex
	tm        *sync.Mutex
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

	it := &itunes{
		unknwon:   obj,
		app:       handle,
		closeChan: make(chan bool),
		wg:        new(sync.WaitGroup),
		pm:        new(sync.Mutex),
		tm:        new(sync.Mutex),
	}
	return it, nil
}

func (it *itunes) Close() {
	close(it.closeChan)
	it.wg.Wait()

	it.unknwon.Release()
	it.app.Release()
	if it.playlist != nil {
		it.playlist.Release()
	}

	if it.tracks != nil {
		it.tracks.Release()
	}
}

func (it *itunes) CurrentTrack() (*track, error) {
	it.wg.Add(1)
	defer it.wg.Done()

	v, err := it.app.GetProperty("CurrentTrack")
	if err != nil {
		return nil, err
	}

	return createTrack(it, v.ToIDispatch(), it.wg)
}

func (it *itunes) initPlaylist() error {
	it.pm.Lock()
	defer it.pm.Unlock()
	if it.playlist != nil {
		return nil
	}

	v, err := it.app.GetProperty("LibraryPlaylist")
	if err != nil {
		return err
	}
	it.playlist = v.ToIDispatch()

	return nil
}

func (it *itunes) initTracks() error {
	it.tm.Lock()
	defer it.tm.Unlock()

	var err error
	if it.playlist == nil {
		err := it.initPlaylist()
		if err != nil {
			return err
		}
	}

	if it.tracks != nil {
		return nil
	}

	v, err := it.playlist.GetProperty("Tracks")
	if err != nil {
		return err
	}
	it.tracks = v.ToIDispatch()

	return nil
}

func (it *itunes) GetTracks() (chan *track, error) {
	it.wg.Add(1)
	defer it.wg.Done()

	if it.tracks == nil {
		err := it.initTracks()
		if err != nil {
			return nil, err
		}
	}

	v, err := it.tracks.GetProperty("Count")
	if err != nil {
		return nil, err
	}

	count := int(v.Val)

	output := make(chan *track)
	go func() {
		it.wg.Add(1)
		defer it.wg.Done()
		defer close(output)

		for i := 1; i <= count; i++ {
			v, err = it.tracks.GetProperty("Item", i)
			if err != nil {
				return
			}

			track, err := createTrack(it, v.ToIDispatch(), it.wg)
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

const PersistentIDSize = 16

func (it *itunes) FindTrackByPersistentID(persistentID string) (*track, error) {
	it.wg.Add(1)
	defer it.wg.Done()

	length := len(persistentID)
	if length > PersistentIDSize || length < 0 {
		return nil, errors.New(fmt.Sprintf("invalid persistentID:%v", persistentID))
	}

	var highID, lowID uint32
	if length <= (PersistentIDSize / 2) {
		highID = 0
		v, err := strconv.ParseUint(persistentID, 16, 32)
		if err != nil {
			return nil, err
		}
		lowID = uint32(v)
	} else {
		highIndex := length - (PersistentIDSize / 2)
		v, err := strconv.ParseUint(persistentID[:highIndex], 16, 32)
		if err != nil {
			return nil, err
		}
		highID = uint32(v)

		v, err = strconv.ParseUint(persistentID[highIndex:], 16, 32)
		if err != nil {
			return nil, err
		}
		lowID = uint32(v)
	}

	if it.tracks == nil {
		err := it.initTracks()
		if err != nil {
			return nil, err
		}
	}

	v, err := it.tracks.GetProperty("ItemByPersistentID", highID, lowID)
	if err != nil {
		return nil, err
	}

	t, err := createTrack(it, v.ToIDispatch(), it.wg)
	if err != nil {
		return nil, err
	}

	return t, nil
}
