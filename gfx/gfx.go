package gfx

type GFX interface {
	SetPixel(pixel, value uint16)
	GetPixel(pixel uint16) uint16
	Draw()
	Clear()
}

