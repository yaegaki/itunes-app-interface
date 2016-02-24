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

func (a ArtworkFormat) Ext() string {
	switch a {
	case JPEG:
		return ".jpg"
	case PNG:
		return ".png"
	case BMP:
		return ".bmp"
	}

	return ""

}

func (a *Artwork) Format() ArtworkFormat {
	return a.format
}
