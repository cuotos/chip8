package gfx

import (
	"fmt"
)

const (
	x = 64
	y = 32
)
type Terminal struct {
	Mem []uint16
}

func (t *Terminal) SetPixel(pixel, value uint16) {
	t.Mem[pixel] = value
}

func (t *Terminal) GetPixel(pixel uint16) uint16 {
	return t.Mem[pixel]
}

func NewTerminalGFX() *Terminal {
	gfx := &Terminal{
		Mem: make([]uint16, x*y),
	}

	return gfx
}

func (t *Terminal) Draw() {

	for i, p := range t.Mem {
		if (i % x) == 0 {
			fmt.Print("\n")
		}

		if p == 1 {
			fmt.Printf("%2s", "0")
		} else {
			fmt.Printf("%2s", " ")
		}
	}

	fmt.Printf("\n")
}

func (t *Terminal) Clear(){
	for _, i := range t.Mem {
		t.Mem[i] = 0
	}
}

func (t *Terminal) Initialise() (func(), error){
	return func(){}, nil
}


