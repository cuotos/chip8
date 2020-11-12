package main

import (
	"github.com/cuotos/chip8/chip"
	"github.com/cuotos/chip8/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"os"
	"time"
)

// for stuck check
var prevPC uint16

func init(){
	rand.Seed(time.Now().UnixNano())
}
func main() {

	c := chip.NewDefaultChip()
	// initialise the chip
	c.Initialise()

//	gfx := gfx.NewTerminalGFX() // TODO: this will be added to the NewChip at some point along with a logger and stuff

	gfx, err := gfx.NewSDLGraphics(64, 32, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer gfx.Cleanup()

	c.GFX = gfx

	err = c.Load("roms/invaders.ch8")
	if err != nil {
		log.Fatal(err)
	}

	clock := time.NewTicker(time.Second / time.Duration(500))
	timers := time.NewTicker(time.Second / time.Duration(60))
	video := time.NewTicker(time.Second / time.Duration(60))

	for {
		select {
		case <-clock.C:
			err := c.EmulateCycle()
			if err != nil {
				c.DiagDump()
				log.Fatal(err)
			}

			//fmt.Printf("tick: oc:%04x pc:%x I:%02x, S:%x, r:%x\n", c.OpCode, c.PC-512, c.I, c.Stack, c.V)

		case <-video.C:
			if c.DrawFlag {
				c.GFX.Draw()
				c.DrawFlag = false
			}

		case <-timers.C:
			// decrement the timers if required
			if c.DelayTimer > 0 {
				c.DelayTimer -= 1
			}
			if c.SoundTimer > 0 {
				c.SoundTimer -= 1
			}
		}

		processEvents(c)
	}
}

func processEvents(c *chip.Chip8) bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		// switch case for when someone quits out of application
		case *sdl.QuitEvent:
			println("Quit") // not necessary
			// decided with os.Exit since I was having issues when I just
			//broke the game loop and window wasn't closing properly
			os.Exit(0)
		case *sdl.KeyboardEvent:
			if e.Type == sdl.KEYUP {
				c.SetKey(e.Keysym.Sym, false)
			}
			if e.Type == sdl.KEYDOWN {
				c.SetKey(e.Keysym.Sym, true)
			}
		}
	}

	return true
}