package itunes

type ArtworkFormat int

const (
	Unknown ArtworkFormat = iota
	JPEG
	PNG
	BMP
)

func (a ArtworkFormat) String() string {
	switch a {
	case Unknown:
		return "Unknown"
	case JPEG:
		return "JPEG"
	case PNG:
		return "PNG"
	case BMP:
		return "BMP"
	}

	return ""
}
