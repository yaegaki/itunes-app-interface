package itunes

type PlayerState int

const (
	Stopped PlayerState = iota
	Playing
	FastForward
	Rewind
)

func (p PlayerState) String() string {
	switch p {
	case Stopped:
		return "Stopped"
	case Playing:
		return "Playing"
	case FastForward:
		return "FastForward"
	case Rewind:
		return "Rewind"
	}

	return ""
}

func (it *Itunes) GetAllTracks() ([]*Track, error) {
	output, err := it.GetTracks()
	if err != nil {
		return nil, err
	}

	tracks := make([]*Track, 0, 100)
	for track := range output {
		tracks = append(tracks, track)
	}

	return tracks, nil
}
