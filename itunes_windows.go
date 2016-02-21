package itunes

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/go-ole/go-ole"
	"github.com/yaegaki/go-ole-handler"
)

type itunes struct {
	handler *olehandler.OleHandler

	libraryPlaylist *olehandler.OleHandler
	tracks          *olehandler.OleHandler
}

func Init() error {
	return ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
}

func UnInit() {
	ole.CoUninitialize()
}

func CreateItunes() (*itunes, error) {
	handler, err := olehandler.CreateRootOleHandler("iTunes.Application")
	if err != nil {
		return nil, err
	}

	var libraryPlaylist, tracks *olehandler.OleHandler
	err = func() error {
		libraryPlaylist, err = handler.GetOleHandler("LibraryPlaylist")
		if err != nil {
			return err
		}

		tracks, err = libraryPlaylist.GetOleHandler("Tracks")
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		handler.Close()
		return nil, err
	}

	it := &itunes{
		handler:         handler,
		libraryPlaylist: libraryPlaylist,
		tracks:          tracks,
	}
	return it, nil
}

func (it *itunes) Close() {
	it.handler.Close()
}

func (it *itunes) CurrentTrack() (t *track, err error) {
	err = it.handler.GetOleHandlerWithCallback("CurrentTrack", func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	})

	return t, err
}

func (it *itunes) TrackCount() (int, error) {
	return it.tracks.GetIntProperty("Count")
}

func (it *itunes) GetTrack(index int) (t *track, err error) {
	err = it.tracks.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	}, index+1)

	return t, err
}

func (it *itunes) GetTracks() (chan *track, error) {
	count, err := it.TrackCount()
	if err != nil {
		return nil, err
	}

	output := make(chan *track)
	go func() {
		defer close(output)
		for i := 1; i <= count; i++ {
			var t *track
			err = it.tracks.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
				t, err = createTrack(it, handler)
				return err
			}, i)

			if err != nil {
				log.Println(err)
				return
			}

			select {
			case <-it.handler.Closed():
				t.Close()
			case output <- t:
			}
		}
	}()

	return output, nil
}

const PersistentIDSize = 16

func (it *itunes) FindTrackByPersistentID(persistentID string) (*track, error) {
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

	var (
		t   *track
		err error
	)

	err = it.tracks.GetOleHandlerWithCallbackAndArgs("ItemByPersistentID", func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	}, highID, lowID)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (it *itunes) Play() error {
	return it.handler.CallMethod("Play")
}

func (it *itunes) Stop() error {
	return it.handler.CallMethod("Stop")
}

func (it *itunes) BackTrack() error {
	return it.handler.CallMethod("BackTrack")
}

func (it *itunes) PreviousTrack() error {
	return it.handler.CallMethod("PreviousTrack")
}

func (it *itunes) NextTrack() error {
	return it.handler.CallMethod("NextTrack")
}

func (it *itunes) SetPlayerPosition(pos int) error {
	return it.handler.PutProperty("PlayerPosition", pos)
}

func (it *itunes) PlayerPosition() (int, error) {
	return it.handler.GetIntProperty("PlayerPosition")
}

func (it *itunes) PlayerState() (PlayerState, error) {
	v, err := it.handler.GetIntProperty("PlayerState")
	if err != nil {
		return PlayerState(0), err
	}

	return PlayerState(v), nil
}

func (it *itunes) PlayPause() error {
	return it.handler.CallMethod("PlayPause")
}

func (it *itunes) Pause() error {
	return it.handler.CallMethod("Pause")
}

func (it *itunes) Resume() error {
	return it.handler.CallMethod("Resume")
}

func (it *itunes) FastForward() error {
	return it.handler.CallMethod("FastForward")
}

func (it *itunes) Rewind() error {
	return it.handler.CallMethod("Rewind")
}

func (it *itunes) SetSoundVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("volume is out of range")
	}

	return it.handler.PutProperty("SoundVolume", volume)
}

func (it *itunes) SoundVolume() (int, error) {
	return it.handler.GetIntProperty("SoundVolume")
}

func (it *itunes) SetMute(isMuted bool) error {
	return it.handler.PutProperty("Mute", isMuted)
}

func (it *itunes) Mute() (bool, error) {
	return it.handler.GetBoolProperty("Mute")
}
