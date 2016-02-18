package itunes

type track struct {
	highID int
	lowID  int
	it     *itunes
	Name   string
	Artist string
}

func (t *track) Play() error {
	v, err := t.it.tracks.GetProperty("ItemByPersistentID", t.highID, t.lowID)
	if err != nil {
		return err
	}
	handle := v.ToIDispatch()
	defer handle.Release()

	_, err = handle.CallMethod("Play")
	return err
}
