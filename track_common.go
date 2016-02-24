package itunes

func (t *Track) Name() string {
	return t.name
}

func (t *Track) Artist() string {
	return t.artist
}

func (t *Track) Album() string {
	return t.album
}
