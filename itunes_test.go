package itunes

import "testing"

func TestItunes(t *testing.T) {
	err := Init()
	if err != nil {
		t.Errorf("Init failed.\n%v", err)
	}
	defer UnInit()

	it, err := CreateItunes()
	if err != nil {
		t.Errorf("CreateItunes failed.\n%v", err)
	}
	defer it.Close()

	// test sound volume
	for i := 0; i <= 100; i++ {
		err = it.SetSoundVolume(i)
		if err != nil {
			t.Errorf("SetSoundVolume failed.\n%v", err)
		}

		volume, err := it.SoundVolume()
		if err != nil {
			t.Errorf("SoundVolume failed.\n%v", err)
		}

		if i != volume {
			t.Errorf("set sound volume to %d, but soud volume is %d", i, volume)
		}
	}

	// test mute
	for _, b := range []bool{true, false} {
		err = it.SetMute(b)
		if err != nil {
			t.Errorf("SetMute failed.\n%v", err)
		}

		mute, err := it.Mute()
		if err != nil {
			t.Errorf("Mute failed.\n%v", err)
		}

		if b != mute {
			t.Errorf("set mute to %d, but mute is %d", b, mute)
		}
	}
}

func TestItunesControls(t *testing.T) {
	err := Init()
	if err != nil {
		t.Errorf("Init failed.\n%v", err)
	}
	defer UnInit()

	it, err := CreateItunes()
	if err != nil {
		t.Errorf("CreateItunes failed.\n%v", err)
	}
	defer it.Close()

	err = it.Play()
	if err != nil {
		t.Errorf("Play failed.\n%v", err)
	}
	testPlayerState(t, it, Playing)

	err = it.Pause()
	if err != nil {
		t.Errorf("Pause failed.\n%v", err)
	}
	testPlayerState(t, it, Stopped)

	for _, state := range []PlayerState{Playing, Stopped} {
		err = it.PlayPause()
		if err != nil {
			t.Errorf("PlayPause failed.\n%v", err)
		}
		testPlayerState(t, it, state)
	}

	testPlayerPosition(t, it, 10)
	testBackTrack(t, it)

	baseTrack, err := it.CurrentTrack()
	if err != nil {
		t.Errorf("CurrentTrack failed.\n%v", err)
	}
	defer baseTrack.Close()

	testNextTrack(t, it, baseTrack)
	testPreviousTrack(t, it, baseTrack)

	err = it.Play()
	if err != nil {
		t.Errorf("Play failed.\n%v", err)
	}
	testPlayerState(t, it, Playing)

	err = it.FastForward()
	if err != nil {
		t.Errorf("FastForward failed.\n%v", err)
	}
	testPlayerState(t, it, FastForward)

	err = it.Resume()
	if err != nil {
		t.Errorf("Resume failed.\n%v", err)
	}
	testPlayerState(t, it, Playing)

	err = it.Rewind()
	if err != nil {
		t.Errorf("Rewind failed.\n%v", err)
	}
	testPlayerState(t, it, Rewind)

	err = it.Stop()
	if err != nil {
		t.Errorf("Stop failed.\n%v", err)
	}
	testPlayerState(t, it, Stopped)
}

func testPlayerState(t *testing.T, it *itunes, expect PlayerState) {
	ps, err := it.PlayerState()
	if err != nil {
		t.Errorf("PlayerState failed.\n%v", err)
	}

	if ps != expect {
		t.Errorf("expect %v, but %v", expect, ps)
	}
}

func testPlayerPosition(t *testing.T, it *itunes, expect int) {
	err := it.SetPlayerPosition(expect)
	if err != nil {
		t.Errorf("SetPlayerPosition failed.\n%v", err)
	}

	pp, err := it.PlayerPosition()
	if err != nil {
		t.Errorf("PlayerPosition failed.\n%v", err)
	}

	if pp < expect || expect+1 < pp {
		t.Errorf("Expect %v <= PlayerPosition < %v, but %v", expect, expect+1, pp)
	}
}

func testBackTrack(t *testing.T, it *itunes) {
	err := it.BackTrack()
	if err != nil {
		t.Errorf("BackTrack failed.\n%v", err)
	}

	pp, err := it.PlayerPosition()
	if err != nil {
		t.Errorf("PlayerPosition failed.\n%v", err)
	}

	if pp < 0 || 1 < pp {
		t.Errorf("Expect %v <= PlayerPosition < %v, but %v", 0, 1, pp)
	}
}

func testNextTrack(t *testing.T, it *itunes, track *track) {
	err := it.NextTrack()
	if err != nil {
		t.Errorf("NextTrack failed.\n%v", err)
	}

	nt, err := it.CurrentTrack()
	if err != nil {
		t.Errorf("CurrentTrack failed.\n%v", err)
	}
	defer nt.Close()

	if nt.PersistentID() == track.PersistentID() {
		t.Errorf("invalid NextTrack.")
	}
}

func testPreviousTrack(t *testing.T, it *itunes, track *track) {
	err := it.PreviousTrack()
	if err != nil {
		t.Errorf("PreviousTrack failed.\n%v", err)
	}

	nt, err := it.CurrentTrack()
	if err != nil {
		t.Errorf("CurrentTrack failed.\n%v", err)
	}
	defer nt.Close()

	if nt.PersistentID() != track.PersistentID() {
		t.Errorf("invalid PreviousTrack.")
	}
}
