package itunes

type artwork struct {
	Format ArtworkFormat
}

// for compatibility
func (_ *artwork) Close() {
}
