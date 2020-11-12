package chip

import (
	"github.com/stretchr/testify/assert"
	"github.com/veandco/go-sdl2/sdl"
	"testing"
)

func TestCountersDecrementOnCycle(t *testing.T) {
	t.Skip()
	tcs := []struct {
		StartingDelayTimer          uint8
		StartingSoundTimer          uint8
		ExpectedDelayAfterTwoCycles uint8
		ExpectedSoundAfterTwoCycles uint8
	}{
		{0xa, 0xb, 0x8, 0x9},
		{0x5, 0x8, 0x3, 0x6},
		{0x1, 0x8, 0x0, 0x6},
		{0x1, 0x1, 0x0, 0x0},
	}

	for _, tc := range tcs {
		c := Chip8{
			DelayTimer: tc.StartingDelayTimer,
			SoundTimer: tc.StartingSoundTimer,
		}
		c.opcodes = opcodes{0x0000: func(c *Chip8) {}} //NOOP

		err := c.EmulateCycle()
		if err != nil {
			t.Error("shouldnt get here")
		}

		err = c.EmulateCycle()

		// Counters should decrement
		if assert.NoError(t, err) {
			assert.Equal(t, tc.ExpectedDelayAfterTwoCycles, uint8(c.DelayTimer))
			assert.Equal(t, tc.ExpectedSoundAfterTwoCycles, uint8(c.SoundTimer))
		}
	}
}

type mockGFX struct{}
func (g mockGFX) Initialise() (func(), error) {return nil, nil}
func (g mockGFX) SetPixel(pixel, value uint16) {}
func (g mockGFX) GetPixel(pixel uint16) uint16 {return 0}
func (mockGFX) Draw(){}
func (mockGFX) Clear(){}

func TestDrawFlagIsResetAfterADraw(t *testing.T) {
	t.Skip()
	c := NewChip8(opcodes{0x0000: func(c *Chip8) {}}, nil) //NOOP
	c.GFX = &mockGFX{}

	c.DrawFlag = true

	err := c.EmulateCycle()

	if assert.NoError(t, err) {
		assert.False(t, c.DrawFlag)
	}
}

func TestCanGetOpcodeFromMemory(t *testing.T) {
	c := NewChip8(nil, nil)
	c.opcodes = opcodes{0xa000: func(c *Chip8) {}} //NOOP

	c.Memory[0x100] = 0xab // 0d256
	c.Memory[0x101] = 0xcd // 0d257

	c.PC = 0x100

	err := c.EmulateCycle()

	if assert.NoError(t, err) {
		assert.Equal(t, uint16(0xabcd), c.OpCode)
	}
}

func TestNewChip8(t *testing.T) {

	triggered := false

	randomUintFunc := func() uint8 {
		triggered = true
		return 0x12
	}
	opcodes := opcodes{0x0001: func(c *Chip8) {}}

	var c *Chip8

	c = NewDefaultChip()
	assert.NotNil(t, c.opcodes)
	assert.NotNil(t, c.randomUintFunc)

	c = NewChip8(opcodes, nil)
	assert.NotNil(t, c.randomUintFunc)
	assert.Equal(t, opcodes, c.opcodes)

	// pass a control function as random in and call that. make sure our trigger was tripped and the returned value
	// is the same as the opcode of the function we intended to run.
	c = NewChip8(nil, randomUintFunc)
	randomUint := c.randomUintFunc()
	assert.NotNil(t, c.opcodes)
	assert.True(t, triggered)
	assert.Equal(t, uint8(0x12), randomUint)

}

//TODO this is pants
func TestSetKeys(t *testing.T) {
	t.Skip()
	c := Chip8{}

	c.SetKey(sdl.K_0, true)
	c.SetKey(sdl.K_1, true)
	c.SetKey(sdl.K_2, true)
	c.SetKey(sdl.K_3, true)
	c.SetKey(sdl.K_4, true)
	c.SetKey(sdl.K_5, true)
	c.SetKey(sdl.K_6, true)
	c.SetKey(sdl.K_7, true)
	c.SetKey(sdl.K_8, true)
	c.SetKey(sdl.K_9, true)
	c.SetKey(sdl.K_a, true)
	c.SetKey(sdl.K_b, true)
	c.SetKey(sdl.K_c, true)
	c.SetKey(sdl.K_d, true)
	c.SetKey(sdl.K_e, true)
	c.SetKey(sdl.K_f, true)

	assert.Equal(t, uint8(0x1), c.Keypad[0x0])
	assert.Equal(t, uint8(0x1), c.Keypad[0x1])
	assert.Equal(t, uint8(0x1), c.Keypad[0x2])
	assert.Equal(t, uint8(0x1), c.Keypad[0x3])
	assert.Equal(t, uint8(0x1), c.Keypad[0x4])
	assert.Equal(t, uint8(0x1), c.Keypad[0x5])
	assert.Equal(t, uint8(0x1), c.Keypad[0x6])
	assert.Equal(t, uint8(0x1), c.Keypad[0x7])
	assert.Equal(t, uint8(0x1), c.Keypad[0x8])
	assert.Equal(t, uint8(0x1), c.Keypad[0x9])
	assert.Equal(t, uint8(0x1), c.Keypad[0xa])
	assert.Equal(t, uint8(0x1), c.Keypad[0xb])
	assert.Equal(t, uint8(0x1), c.Keypad[0xc])
	assert.Equal(t, uint8(0x1), c.Keypad[0xd])
	assert.Equal(t, uint8(0x1), c.Keypad[0xe])
	assert.Equal(t, uint8(0x1), c.Keypad[0xf])

	c.SetKey(sdl.K_b, false)

	assert.Equal(t, uint8(0x0), c.Keypad[0xb])
}

//TODO: Test Initialise
func TestInitialiseTheChip(t *testing.T) {
	TODO(t)
}
