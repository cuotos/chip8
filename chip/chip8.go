package chip

import (
	"log"
	"math/rand"
	"os"

	"github.com/cuotos/chip8/gfx"
)

const (
	VF = 0xf
)

type randomUintFunc func() uint8

type Chip8 struct {
	OpCode         uint16
	Memory         [4096]uint8
	V              [16]uint8
	I              uint16
	PC             uint16
	DelayTimer     uint8
	SoundTimer     uint8
	Stack          [16]uint16
	SP             uint16
	DrawFlag       bool
	Keypad         [16]uint8
	GFX            gfx.GFX
	opcodes        // map of the opcode, can be replaced for testing
	randomUintFunc randomUintFunc
}

func NewDefaultChip() *Chip8 {
	return NewChip8(nil, nil)
}

func NewChip8(opcodes opcodes, randomiser randomUintFunc) *Chip8 {
	c := &Chip8{
		opcodes:        opcodes,
		randomUintFunc: randomiser,
	}

	if c.opcodes == nil {
		c.opcodes = defaultOpcodes
	}

	if c.randomUintFunc == nil {
		c.randomUintFunc = func() uint8 {
			return uint8(rand.Int31n(255))
		}
	}

	return c
}

func (c *Chip8) Initialise() {
	// program counter starts at 0x200
	c.PC = 0x200
	c.OpCode = 0
	c.I = 0
	c.SP = 0

	// Load fontset
	for i := 0; i < len(FontSet); i++ {
		c.Memory[i] = FontSet[i]
	}
}

// TODO: accept filename as var
func (c *Chip8) Load(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	fStat, fStateErr := file.Stat()
	if err != nil {
		return fStateErr
	}

	buffer := make([]byte, fStat.Size())

	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	for i := 0; i < len(buffer); i++ {
		c.Memory[i+512] = buffer[i]
	}

	return nil
}

func (c *Chip8) EmulateCycle() error {

	// Fetch opcode
	c.OpCode = uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])
	log.Printf("[DEBUG] oc:%04x pc:%x I:%02x\n", c.OpCode, c.PC, c.I)
	c.DiagDump()
	err := c.HandleOpcode()
	if err != nil {
		return err
	}

	return nil
}

func (c *Chip8) SetKeys() {}
