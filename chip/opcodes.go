package chip

import (
	"fmt"
)

type opcodes map[uint16]func(*Chip8)

var defaultOpcodes = opcodes{
	0x00e0: func(c *Chip8) {
		c.GFX.Clear()
		c.DrawFlag = true
		c.PC += 2
	},

	//00EE	Flow	return;	Returns from a subroutine.
	0x00ee: func(c *Chip8) {
		c.SP -= 1
		c.PC = c.Stack[c.SP]
		c.PC += 2
	},

	0x1000: func(c *Chip8) {
		c.PC = c.OpCode & 0x0FFF
	},

	//2NNN - Calls subroutine at NNN
	0x2000: func(c *Chip8) {
		c.Stack[c.SP] = c.PC
		c.SP++

		c.PC = c.OpCode & 0x0fff
	},

	// 3XNN	Cond	if(Vx==NN)	Skips the next instruction if VX equals NN. (Usually the next instruction is a jump to skip a code block)
	0x3000: func(c *Chip8) {
		reg := (c.OpCode & 0x0F00) >> 8
		nn := (c.OpCode & 0x00FF)

		if nn == uint16(c.V[reg]) {
			c.PC += 4
		} else {
			c.PC += 2
		}
	},

	// 4XNN	Cond	if(Vx!=NN)	Skips the next instruction if VX doesn't equal NN. (Usually the next instruction is a jump to skip a code block)
	0x4000: func(c *Chip8) {
		reg := (c.OpCode & 0x0F00) >> 8
		nn := (c.OpCode & 0x00FF)

		if nn != uint16(c.V[reg]) {
			c.PC += 4
		} else {
			c.PC += 2
		}
	},

	//5XY0	Cond	if(Vx==Vy)	Skips the next instruction if VX equals VY. (Usually the next instruction is a jump to skip a code block)
	0x5000: func(c *Chip8) {
		r1 := (c.OpCode & 0x0F00) >> 8
		r2 := (c.OpCode & 0x00F0) >> 4

		if c.V[r1] == c.V[r2] {
			c.PC += 4
		} else {
			c.PC += 2
		}
	},

	//6XNN	Const	Vx = NN	Sets VX to NN.
	0x6000: func(c *Chip8) {
		// Get the registry number and shift it right so its the LSB
		registry := (c.OpCode & 0x0f00) >> 8
		c.V[registry] = uint8(c.OpCode & 0x00ff)
		c.PC += 2
	},

	//7XNN	Const	Vx += NN	Adds NN to VX. (Carry flag is not changed)
	0x7000: func(c *Chip8) {
		reg := c.OpCode & 0x0F00 >> 8
		c.V[reg] += uint8(c.OpCode & 0x00FF)
		c.PC += 2
	},

	//8XYn maths stuff...
	0x8000: func(c *Chip8) {
		regX := c.OpCode & 0x0f00 >> 8
		regY := c.OpCode & 0x00f0 >> 4

		// Just test the LSB
		switch c.OpCode & 0xf {
		case 0x0:
			c.V[regX] = c.V[regY]
		case 0x1:
			c.V[regX] = c.V[regX] | c.V[regY]
		case 0x2:
			c.V[regX] = c.V[regX] & c.V[regY]
		case 0x3:
			c.V[regX] = c.V[regX] ^ c.V[regY]
		// regX + regY, set VF if carry
		case 0x4:
			if c.V[regX] > (255 - c.V[regY]) {
				c.V[VF] = 1
			} else {
				c.V[VF] = 0
			}
			c.V[regX] += c.V[regY]
		// regX - regY, set VF if borrow
		case 0x5:
			if c.V[regX] > c.V[regY] {
				c.V[VF] = 1
			} else {
				c.V[VF] = 0
			}
			c.V[regX] = c.V[regX] - c.V[regY]
		// 8XY6[a]	BitOp	Vx>>=1	Stores the least significant bit of VX in VF and then shifts VX to the right by 1.[b]
		case 0x6:
			c.V[VF] = c.V[regX] & 0x1
			c.V[regX] = c.V[regX] >> 1
		// 8XY7[a]	Math	Vx=Vy-Vx	Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
		case 0x7:
			if c.V[regY] > c.V[regX] {
				c.V[VF] = 1
			} else {
				c.V[VF] = 0
			}
			c.V[regX] = c.V[regY] - c.V[regX]
		// 8XYE[a]	BitOp	Vx<<=1	Stores the most significant bit of VX in VF and then shifts VX to the left by 1.[b]
		case 0xe:
			c.V[VF] = (c.V[regX] & 0x80) >> 7
			c.V[regX] = c.V[regX] << 1
		}

		c.PC += 2
	},

	//9XY0	Cond	if(Vx!=Vy)	Skips the next instruction if VX doesn't equal VY. (Usually the next instruction is a jump to skip a code block)
	0x9000: func(c *Chip8) {
		r1 := (c.OpCode & 0x0F00) >> 8
		r2 := (c.OpCode & 0x00F0) >> 4

		if c.V[r1] != c.V[r2] {
			c.PC += 4
		} else {
			c.PC += 2
		}
	},

	// ANNN	MEM	I = NNN	Sets I to the address NNN.
	0xa000: func(c *Chip8) {
		c.I = c.OpCode & 0x0fff
		c.PC += 2
	},

	//BNNN	Flow	PC=V0+NNN	Jumps to the address NNN plus V0.
	0xb000: func(c *Chip8) {
		c.PC = uint16(c.V[0]) + c.OpCode & 0x0fff
	},

	//CXNN	Rand	Vx=rand()&NN	Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN.
	0xc000: func(c *Chip8) {
		c.PC += 2

		c.V[c.OpCode&0x0f00>>8] = uint8(c.OpCode&0xff) & c.randomUintFunc()
	},

	// DXYN - draw at points X, Y and sprite of N rows high
	0xd000: func(c *Chip8) {
		x := c.V[c.OpCode&0x0f00>>8]
		y := c.V[c.OpCode&0x00f0>>4]
		h := c.OpCode & 0x000f

		// set collision reg to 0
		c.V[VF] = 0

		for yLine := 0; yLine < int(h); yLine++ {
			pixel := c.Memory[int(c.I)+yLine]

			for xLine := 0; xLine < 8; xLine++ {
				// TODO: Need to fully understand this line
				//  0x80 (128 aka 1000 0000) and shift the 1 across 8 times
				if (pixel & (0x80 >> xLine)) != 0 {

					yPlusYLine := uint16(y) + uint16(yLine)
					yPlusYLine64 := yPlusYLine * 64
					xPlusXLine := uint16(x) + uint16(xLine)
					cell := xPlusXLine + yPlusYLine64

					p := c.GFX.GetPixel(cell)
					if p == 1 {
						c.V[VF] = 1
					}
					c.GFX.SetPixel(cell, p ^ 1)
				}
			}
		}
		c.DrawFlag = true

		c.PC += 2
	},

	0xe09e: func(c *Chip8) {
		if c.Keypad[c.OpCode&0x0f00>>8] != 0x0{
			c.PC += 2
		}

		c.PC += 2
	},

	0xe0a1: func(c *Chip8) {
		if c.Keypad[c.OpCode&0x0f00>>8] == 0x0{
			c.PC += 2
		}
		c.PC += 2
	},

	0xf000: func(c *Chip8) {
		switch c.OpCode & 0xff {

		case 0x07:
			c.V[c.OpCode&0x0f00>>8] = c.DelayTimer

		case 0x0a:
			panic("0x0a not implemented")

		case 0x15:
			c.DelayTimer = uint8(c.OpCode & 0x0f00 >> 8)

		case 0x1e:
			reg := c.OpCode & 0x0f00 >> 8
			add := uint16(c.V[reg])
			c.I = c.I + add

		case 0x29:
			char := c.V[c.OpCode&0x0f00>>8]
			c.I = uint16(char) * 5

		case 0x33:
			reg := c.V[c.OpCode&0xf00>>8]
			c.Memory[c.I] = reg / 100
			c.Memory[c.I+1] = (reg / 10) % 10
			c.Memory[c.I+2] = (reg % 100) % 10

		case 0x55:
			numberOfRegs := c.OpCode & 0x0f00 >> 8
			for i := 0; uint16(i) <= numberOfRegs; i++ {
				c.Memory[c.I+uint16(i)] = c.V[i]
			}

		case 0x65:
			maxReg := c.OpCode & 0xf00 >> 8
			for i := 0; i <= int(maxReg); i++ {
				c.V[i] = c.Memory[c.I+uint16(i)]
			}
		}

		c.PC += 2

	},
}

// TODO: this needs to be refactored, as the opcodes object can be passed as part of testing, but this cannot,
//  so even the mock opcodes needs to obey this logic before in order to get "found"
func (ocs *opcodes) LookupOpcode(opcode uint16) (func(c *Chip8), error) {

	var oc = func(c *Chip8) {}
	var opcodeRef uint16

	switch {
	// There are multiple opcodes under 0x0 and 0xe
	case opcode&0xffee == 0x00ee:
		opcodeRef = opcode
	case opcode&0xffef == 0x00e0:
		opcodeRef = opcode

	case opcode&0xf0ff == 0xe09e:
		opcodeRef = 0xe09e
	case opcode&0xf0ff == 0xe0a1:
		opcodeRef = 0xe0a1

	default:
		opcodeRef = opcode & 0xf000
	}

	var ok bool
	oc, ok = (*ocs)[opcodeRef]
	if !ok {
		return nil, fmt.Errorf("unable to lookup opcode: %04x", opcode)
	}

	return oc, nil
}

func (c *Chip8) HandleOpcode() error {

	f, err := c.LookupOpcode(c.OpCode)
	if err != nil {
		return err
	}

	f(c)

	return nil
}
