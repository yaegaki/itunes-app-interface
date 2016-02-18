package itunes

func (it *itunes) GetAllTracks() ([]*track, error) {
	output, err := it.GetTracks()
	if err != nil {
		return nil, err
	}

	tracks := make([]*track, 0, 100)
	for track := range output {
		tracks = append(tracks, track)
	}

	return tracks, nil
}
