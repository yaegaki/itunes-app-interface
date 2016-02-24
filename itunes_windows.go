package itunes

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-ole/go-ole"
	"github.com/yaegaki/go-ole-handler"
)

type Itunes struct {
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

func CreateItunes() (*Itunes, error) {
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

	it := &Itunes{
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

func (it *Itunes) Close() {
	it.handler.Close()
}

func (it *Itunes) CurrentTrack() (t *Track, err error) {
	err = it.handler.GetOleHandlerWithCallback("CurrentTrack", func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	})

	return t, err
}

func (it *Itunes) TrackCount() (int, error) {
	return it.libraryPlaylist.TrackCount()
}

func (it *Itunes) GetTrack(index int) (t *Track, err error) {
	return it.libraryPlaylist.GetTrack(index)
}

func (it *Itunes) GetTracks() (chan *Track, error) {
	return it.libraryPlaylist.GetTracks()
}

const PersistentIDSize = 16

func (it *Itunes) findItemByPersistentID(collection *olehandler.OleHandler, persistentID string, fn func(*olehandler.OleHandler) error) error {
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

func (it *Itunes) FindTrackByPersistentID(persistentID string) (t *Track, err error) {
	err = it.findItemByPersistentID(it.libraryPlaylist.tracks, persistentID, func(handler *olehandler.OleHandler) error {
		t, err = createTrack(it, handler)
		return err
	})

	return t, err
}

func (it *Itunes) CurrentPlaylist() (p *Playlist, err error) {
	err = it.handler.GetOleHandlerWithCallback("CurrentPlaylist", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	})

	return p, err
}

func (it *Itunes) PlaylistCount() (int, error) {
	return it.playlists.GetIntProperty("Count")
}

func (it *Itunes) GetPlaylist(index int) (p *Playlist, err error) {
	err = it.playlists.GetOleHandlerWithCallbackAndArgs("Item", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	}, index+1)

	return p, err
}

func (it *Itunes) FindPlaylistByPersistentID(persistentID string) (p *Playlist, err error) {
	err = it.findItemByPersistentID(it.playlists, persistentID, func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	})

	return p, err
}

func (it *Itunes) CreatePlaylist(playlistName string) (p *Playlist, err error) {
	err = it.handler.GetOleHandlerWithCallbackAndArgsByMethod("CreatePlaylist", func(handler *olehandler.OleHandler) error {
		p, err = createPlaylist(it, handler)
		return err
	}, playlistName)

	return p, err
}

func (it *Itunes) Play() error {
	return it.handler.CallMethod("Play")
}

func (it *Itunes) Stop() error {
	return it.handler.CallMethod("Stop")
}

func (it *Itunes) BackTrack() error {
	return it.handler.CallMethod("BackTrack")
}

func (it *Itunes) PreviousTrack() error {
	return it.handler.CallMethod("PreviousTrack")
}

func (it *Itunes) NextTrack() error {
	return it.handler.CallMethod("NextTrack")
}

func (it *Itunes) SetPlayerPosition(pos int) error {
	return it.handler.PutProperty("PlayerPosition", pos)
}

func (it *Itunes) PlayerPosition() (int, error) {
	return it.handler.GetIntProperty("PlayerPosition")
}

func (it *Itunes) PlayerState() (PlayerState, error) {
	v, err := it.handler.GetIntProperty("PlayerState")
	if err != nil {
		return PlayerState(0), err
	}

	return PlayerState(v), nil
}

func (it *Itunes) PlayPause() error {
	return it.handler.CallMethod("PlayPause")
}

func (it *Itunes) Pause() error {
	return it.handler.CallMethod("Pause")
}

func (it *Itunes) Resume() error {
	return it.handler.CallMethod("Resume")
}

func (it *Itunes) FastForward() error {
	return it.handler.CallMethod("FastForward")
}

func (it *Itunes) Rewind() error {
	return it.handler.CallMethod("Rewind")
}

func (it *Itunes) SetSoundVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("volume is out of range")
	}

	return it.handler.PutProperty("SoundVolume", volume)
}

func (it *Itunes) SoundVolume() (int, error) {
	return it.handler.GetIntProperty("SoundVolume")
}

func (it *Itunes) SetMute(isMuted bool) error {
	return it.handler.PutProperty("Mute", isMuted)
}

func (it *Itunes) Mute() (bool, error) {
	return it.handler.GetBoolProperty("Mute")
}
