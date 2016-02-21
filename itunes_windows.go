package itunes

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-ole/go-ole"
	"github.com/yaegaki/go-ole-handler"
)

type itunes struct {
	handler *olehandler.OleHandler

	libraryPlaylist *Playlist
	librarySource   *olehandler.OleHandler
	playlists       *olehandler.OleHandler
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

	var librarySource, playlists *olehandler.OleHandler
	var libraryPlaylist *Playlist
	err = func() error {
		librarySource, err = handler.GetOleHandler("LibrarySource")
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		playlists, err = librarySource.GetOleHandler("Playlists")
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
		handler:       handler,
		librarySource: librarySource,
		playlists:     playlists,
	}

	err = handler.GetOleHandlerWithCallback("LibraryPlaylist", func(handler *olehandler.OleHandler) error {
		libraryPlaylist, err = createPlaylist(it, handler)
		return err
	})

	if err != nil {
		it.Close()
		return nil, err
	}

	it.libraryPlaylist = libraryPlaylist

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
	return it.libraryPlaylist.TrackCount()
}

func (it *itunes) GetTrack(index int) (t *track, err error) {
	return it.libraryPlaylist.GetTrack(index)
}

func (it *itunes) GetTracks() (chan *track, error) {
	return it.libraryPlaylist.GetTracks()
}

const PersistentIDSize = 16

func (it *itunes) findItemByPersistentID(collection *olehandler.OleHandler, persistentID string, fn func(*olehandler.OleHandler) error) error {
	length := len(persistentID)
	if length > PersistentIDSize || length < 0 {
		return errors.New(fmt.Sprintf("invalid persistentID:%v", persistentID))
	}

	var highID, lowID uint32
	if length <= (PersistentIDSize / 2) {
		highID = 0
		v, err := strconv.ParseUint(persistentID, 16, 32)
		if err != nil {
			return err
		}
		lowID = uint32(v)
	} else {
		highIndex := length - (PersistentIDSize / 2)
		v, err := strconv.ParseUint(persistentID[:highIndex], 16, 32)
		if err != nil {
			return err
		}
		highID = uint32(v)

		v, err = strconv.ParseUint(persistentID[highIndex:], 16, 32)
		if err != nil {
			return err
		}
		lowID = uint32(v)
	}

	return collection.GetOleHandlerWithCallbackAndArgs("ItemByPersistentID", func(handler *olehandler.OleHandler) error {
		return fn(handler)
	}, highID, lowID)
}

func (it *itunes) FindTrackByPersistentID(persistentID string) (t *track, err error) {
	err = it.findItemByPersistentID(it.libraryPlaylist.tracks, persistentID, func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	})

	return t, err
}

func (it *itunes) CurrentPlaylist() (p *Playlist, err error) {
	err = it.handler.GetOleHandlerWithCallback("CurrentPlaylist", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	})

	return p, err
}

func (it *itunes) PlaylistCount() (int, error) {
	return it.playlists.GetIntProperty("Count")
}

func (it *itunes) GetPlaylist(index int) (p *Playlist, err error) {
	err = it.playlists.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	}, index+1)

	return p, err
}

func (it *itunes) FindPlaylistByPersistentID(persistentID string) (p *Playlist, err error) {
	err = it.findItemByPersistentID(it.playlists, persistentID, func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	})

	return p, err
}

func (it *itunes) CreatePlaylist(playlistName string) (p *Playlist, err error) {
	err = it.handler.GetOleHandlerWithCallbackAndArgsByMethod("CreatePlaylist", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	}, playlistName)

	return p, err
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
