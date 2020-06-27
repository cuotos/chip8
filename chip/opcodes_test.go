package chip

import (
	"fmt"
	"github.com/cuotos/chip8/gfx"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Unkown opcode should kill the app
func TestErrorOnUnknownOpcode(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		ExpectErr   bool
	}{
		{uint16(0xA000), false},
		{uint16(0xB000), true},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%x", tc.InputOpcode), func(t *testing.T) {
			mockOpcodes := opcodes{
				0xA000: func(chip8 *Chip8) {},
			}
			c := NewChip8(mockOpcodes, nil)
			c.OpCode = tc.InputOpcode
			err := c.HandleOpcode()

			if tc.ExpectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLookupOpcode(t *testing.T) {

	// This is populated by the the test function with the id of the opcode found
	var opcodeTriggered uint16

	// generates a test opcode that sets the trigger above so that it can be asserted which function was called
	genTestFunc := func(i uint16) func(*Chip8) {
		return func(c *Chip8) {
			opcodeTriggered = i
		}
	}

	tcs := []struct {
		InputOpcode    uint16
		ExpectedOpcode uint16
	}{
		{0x0123, 0x0000},
		{0x00e0, 0x00e0},
		{0x00ee, 0x00ee},

		{0x1999, 0x1000},

		{0xa111, 0xa000},

		{0xe59e, 0xe09e},
		{0xefa1, 0xe0a1},

		{0x8ff0, 0x8000},

		{0xf433, 0xf000},
		{0xf265, 0xf000},
	}

	for _, tc := range tcs {
		// reset the trigger, guessing this wouldn't work with parallel testing
		opcodeTriggered = 0

		c := NewDefaultChip()
		c.OpCode = tc.InputOpcode

		ocs := opcodes{tc.ExpectedOpcode: genTestFunc(tc.ExpectedOpcode)}

		f, err := ocs.LookupOpcode(tc.InputOpcode)
		if assert.NoError(t, err) {

			// run the opcode function that was found
			f(c)

			// check that the opcode that wrote its id to "triggered" is the one we expected
			assert.Equal(t, tc.ExpectedOpcode, opcodeTriggered, "the expected function was not run")
		}
	}
}

//0NNN	Call		Calls machine code routine (RCA 1802 for COSMAC VIP) at address NNN. Not necessary for most ROMs.
func TestOpcode0NNN(t *testing.T) {
	t.Skip()
}

type mockClearingGFX struct{
	ClearCalled bool
}

func (g *mockClearingGFX) Initialise() (func(), error) {return nil, nil}
func (g *mockClearingGFX) SetPixel(pixel, value uint16) {}
func (g *mockClearingGFX) GetPixel(pixel uint16) uint16 {return 0}
func (g *mockClearingGFX) Draw(){}
func (g *mockClearingGFX) Clear(){
	g.ClearCalled = true
}

//00E0	Display	disp_clear()	Clears the screen.
func TestOpcode00E0(t *testing.T) {
	c := NewDefaultChip()
	c.GFX = &mockClearingGFX{
		false,
	}

	c.OpCode = 0x00e0

	err := c.HandleOpcode()
	if assert.NoError(t, err) {
		assert.True(t, c.GFX.(*mockClearingGFX).ClearCalled)
		assert.Equal(t, true, c.DrawFlag)
	}
}

//00EE	Flow	return;	Returns from a subroutine.
func TestOpcode00EE(t *testing.T) {
	c := NewDefaultChip()
	c.Stack[5] = 100
	c.SP = 6
	c.PC = 0

	c.OpCode = 0x00ee

	err := c.HandleOpcode()
	assert.NoError(t, err)

	// make sure PC got updated with the item from the stack. and the PC got incremented to the next instruction
	//  we are currently back at the location before the jump.
	assert.Equal(t, 102, int(c.PC))
	// make sure stack pointer got decremented
	assert.Equal(t, c.SP, uint16(0x5))
}

// 1NNN - Jumps to address NNN.
func TestOpcode1NNN(t *testing.T) {
	c := NewDefaultChip()

	c.OpCode = 0x124e
	err := c.HandleOpcode()
	assert.NoError(t, err)

	assert.Equal(t, uint16(0x24e), c.PC) // PC points to new location
	assert.Equal(t, uint16(0x0), c.SP)   // Stack hasn't been touched
}

//2NNN - Calls subroutine at NNN
func TestOpcode2NNN(t *testing.T) {
	c := NewDefaultChip()

	c.OpCode = 0x22d4

	err := c.HandleOpcode()
	assert.NoError(t, err)

	// assert that the program counter is now pointing at the subroutines start
	assert.Equal(t, 0x2d4, int(c.PC))
	// assert the first item on the stack is the address of the calling function (0x000)
	assert.Equal(t, 0x000, int(c.Stack[0]))
}

// 3XNN	Cond	if(Vx==NN)	Skips the next instruction if VX equals NN. (Usually the next instruction is a jump to skip a code block)
func TestOpcode3XNN(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		Vx          uint8
		ExpectedPC  uint16
	}{
		{0x362b, 0x2b, 0x4}, // PC skip instruction
		{0x362b, 0xff, 0x2}, // PC next instruction
	}

	for _, tc := range tcs {
		c := NewDefaultChip()

		c.V[0x6] = tc.Vx

		c.OpCode = tc.InputOpcode
		err := c.HandleOpcode()
		assert.NoError(t, err)

		assert.Equal(t, tc.ExpectedPC, c.PC)
	}

}

// 4XNN	Cond	if(Vx!=NN)	Skips the next instruction if VX doesn't equal NN. (Usually the next instruction is a jump to skip a code block)
func TestOpcode4XNN(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		V6          uint8
		ExpectedPC  uint16
	}{
		// using reg 6 each time
		{0x462b, 0x2b, 0x2}, // PC dont skip instruction
		{0x462b, 0xff, 0x4}, // PC skip next instruction
	}

	for _, tc := range tcs {
		c := NewDefaultChip()

		c.V[0x6] = tc.V6

		c.OpCode = tc.InputOpcode
		err := c.HandleOpcode()
		assert.NoError(t, err)

		assert.Equal(t, tc.ExpectedPC, c.PC)
	}
}

//5XY0	Cond	if(Vx==Vy)	Skips the next instruction if VX equals VY. (Usually the next instruction is a jump to skip a code block)
func TestOpcode5XY0(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		V6          uint8
		Va          uint8
		ExpectedPC  uint16
	}{
		// using reg 6 each time
		{0x56a0, 0x2b, 0x2b, 0x4}, // PC skip instruction
		{0x56a0, 0xff, 0xaa, 0x2}, // PC dont skip instruction
	}

	for _, tc := range tcs {
		c := NewDefaultChip()

		c.V[0x6] = tc.V6
		c.V[0xa] = tc.Va

		c.OpCode = tc.InputOpcode
		err := c.HandleOpcode()
		assert.NoError(t, err)

		assert.Equal(t, tc.ExpectedPC, c.PC)
	}
}

//6NNN Sets VX to NN.
func TestOpcode6XNN(t *testing.T) {

	tcs := []struct {
		InputOpcode   uint16
		ExpectedReg   int
		ExpectedValue uint8
	}{
		{0x6A02,  0xA, 0x02},
		{0x6B01,  0xB, 0x01},
		{0x6FFF,  0xF, 0xFF},
	}

	for _, tc := range tcs {
		c := NewDefaultChip()
		c.OpCode = tc.InputOpcode

		err := c.HandleOpcode()

		if assert.NoError(t, err) {
			assert.Equal(t, c.V[tc.ExpectedReg], tc.ExpectedValue)
			assert.Equal(t, uint16(0x2), c.PC)
		}
	}
}

//7XNN	Const	Vx += NN	Adds NN to VX. (Carry flag is not changed)
func TestOpcode7XNN(t *testing.T) {
	c := NewDefaultChip()
	c.V[0xa] = 0xa

	c.OpCode = 0x7a0a
	err := c.HandleOpcode()
	assert.NoError(t, err)

	assert.Equal(t, uint8(0x14), c.V[0xa])
	assert.Equal(t, uint16(0x2), c.PC)
}

//8XY0 basic stuff, & | ^
func TestOpcode8XYNBasic(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		Expected    uint8
	}{
		{0x8ab0, 0xcd}, //8XY0	Assign	Vx=Vy	Sets VX to the value of VY.
		{0x8ab1, 0xef}, //8XY1	BitOp	Vx=Vx|Vy	Sets VX to VX or VY. (Bitwise OR operation)
		{0x8ab2, 0x89}, //8XY2	BitOp	Vx=Vx&Vy	Sets VX to VX and VY. (Bitwise AND operation)
		{0x8ab3, 0x66}, //8XY3[a]	BitOp	Vx=Vx^Vy	Sets VX to VX xor VY.
	}

	for _, tc := range tcs {
		c := NewDefaultChip()
		c.V[0xa] = 0xab
		c.V[0xb] = 0xcd

		c.OpCode = tc.InputOpcode

		err := c.HandleOpcode()
		if assert.NoError(t, err) {
			assert.Equal(t, uint16(0x2), c.PC)

			assert.Equal(t, tc.Expected, c.V[0xa])
		}
	}
}

//8XYn Maths, carries etc
func TestOpcode8XYNCarries(t *testing.T) {
	tcs := []struct {
		InputOpcode  uint16
		InputRegX    uint8
		InputRegY    uint8
		ExpectedRegX uint8
		ExpectedVF   uint8
	}{
		//8XY4	Math	Vx += Vy	Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there isn't
		{0x4, 0xab, 0xcd, 0x78, 1}, // ab + cd = carry
		{0x4, 0x01, 0x02, 0x03, 0}, // 1 + 2 = no carry

		//8XY5	Math	Vx -= Vy	VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
		{0x5, 0xff, 0xab, 0x54, 1},
		{0x5, 0xab, 0xff, 0xac, 0},

		// 8XY6[a]	BitOp	Vx>>=1	Stores the least significant bit of VX in VF and then shifts VX to the right by 1.[b]
		{0x6, 0x1, 0x0, 0x0, 0x1},
		{0x6, 0xa, 0x0, 0x5, 0},
		{0x6, 0xb, 0x0, 0x5, 1},
		{0x6, 0xff, 0x0, 0x7f, 1},

		//8XY7[a]	Math	Vx=Vy-Vx	Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
		{0x7, 0x1, 0xff, 0xfe, 1},
		{0x7, 0x3, 0x01, 0xfe, 0},

		//8XYE[a]	BitOp	Vx<<=1	Stores the most significant bit of VX in VF and then shifts VX to the left by 1.[b]
		{0xe, 0xff, 0x0, 0xfe, 1},
		{0xe, 0x7e, 0x0, 0xfc, 0},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("0x%x", (tc.InputOpcode|0x8000)), func(t *testing.T) {
			c := NewDefaultChip()
			// always using registers A and B for the sake of sanity
			c.OpCode = 0x8ab0 + tc.InputOpcode

			c.V[0xa] = tc.InputRegX
			c.V[0xb] = tc.InputRegY

			// Make sure that the carry bit gets set to 0 if required and not left from previous runs.
			// If I force it on here, it should go off if needed
			c.V[0xf] = 1

			err := c.HandleOpcode()
			if assert.NoError(t, err) {
				assert.Equal(t, tc.ExpectedRegX, c.V[0xa])
				assert.Equal(t, uint16(0x2), c.PC)
				assert.Equal(t, tc.ExpectedVF, c.V[VF])
			}
		})
	}
}

//9XY0	Cond	if(Vx!=Vy)	Skips the next instruction if VX doesn't equal VY. (Usually the next instruction is a jump to skip a code block)
func TestOpcode9XY0(t *testing.T) {
	tcs := []struct {
		InputOpcode uint16
		V1          uint8
		V2          uint8
		ExpectedPC  uint16
	}{
		// using reg 6 each time
		{0x96a0, 0x2b, 0x2b, 0x2}, // PC dont skip instruction
		{0x96a0, 0xff, 0xaa, 0x4}, // PC skip instruction
	}

	for _, tc := range tcs {
		c := NewDefaultChip()

		c.V[0x6] = tc.V1
		c.V[0xa] = tc.V2

		c.OpCode = tc.InputOpcode
		err := c.HandleOpcode()
		assert.NoError(t, err)

		assert.Equal(t, tc.ExpectedPC, c.PC)
	}
}

// ANNN sets I to the value of NNN
func TestOpcodeANNN(t *testing.T) {

	tcs := []struct {
		Input    uint16
		Expected uint16
	}{
		{uint16(0xAFF0), uint16(0xFF0)},
		{uint16(0xA111), uint16(0x111)},
		{uint16(0xA000), uint16(0x000)},
	}

	for _, tc := range tcs {
		c := NewDefaultChip()
		c.OpCode = tc.Input
		err := c.HandleOpcode()
		assert.NoError(t, err)
		assert.Equal(t, tc.Expected, c.I)
		assert.Equal(t, uint16(0x2), c.PC)
	}
}

//BNNN	Flow	PC=V0+NNN	Jumps to the address NNN plus V0.
func TestOpcodeBNNN(t *testing.T) {
	c := NewDefaultChip()
	c.V[0] = 0x0002
	c.OpCode = 0xb555

	err := c.HandleOpcode()

	if assert.NoError(t, err ){
		assert.Equal(t, uint16(0x557), c.PC)
	}

}

//CXNN	Rand	Vx=rand()&NN	Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN.
func TestOpcodeCXNN(t *testing.T) {

	createControlledRandomFunc := func(alwaysReturn uint8) randomUintFunc {
		return func() uint8 {
			return alwaysReturn
		}
	}

	tcs := []struct{
		AlwaysReturnRandom uint8
		Input uint16
		Expected uint8
	}{
		{0xff, 0x12, 0x12},
		{0x00, 0x12, 0x00},
		{0x0f, 0x12, 0x02},
	}

	for _, tc := range tcs {
		//c := NewChip8(nil, createControlledRandomFunc(tc.AlwaysReturnRandom))
		c := NewChip8(nil, createControlledRandomFunc(tc.AlwaysReturnRandom))

		// TODO: could randomise the register used, then try changing the number to 14 and see how many tests fail
		// always using register 7 here
		c.OpCode = uint16(0xc7<<8) | tc.Input

		err := c.HandleOpcode()

		if assert.NoError(t, err){
			assert.Equal(t, 0x2, int(c.PC))

			assert.Equal(t, tc.Expected, c.V[7])
		}
	}
}

//DXYN Draws a sprite at coordinate (VX, VY) that has a width of 8 pixels and a height of N pixels.
/// Each row of 8 pixels is read as bit-coded starting from memory location I
func TestOpcodeDXYN(t *testing.T) {

	gfx := gfx.NewTerminalGFX()
	c := NewDefaultChip()
	c.GFX = gfx

	// Set I and the next 2 locations, these will be the pixels (8bit row)
	c.Memory[c.I] = 0x3c
	c.Memory[c.I+1] = 0xc3
	c.Memory[c.I+2] = 0xff

	// set the X and Y starting coords
	c.V[0xc] = 15
	c.V[0xd] = 3

	// Opcode says "Draw", get X coord from "c" reg and Y from "d" reg, and the sprite will be 3 rows tall.
	c.OpCode = 0xdcd3
	err := c.HandleOpcode()

	assert.NoError(t, err)

	// TODO: add assertion but i couldnt be bothered to list the entire gfx buffer to compare

	assert.Equal(t, uint16(0x2), c.PC)
}

//EX9E	KeyOp	if(key()==Vx)	Skips the next instruction if the key stored in VX is pressed. (Usually the next instruction is a jump to skip a code block)
func TestOpcodeEX9E(t *testing.T) {
	tcs := []struct{
		InputKey uint8
		Pressed uint8 // 1 or 0
		ExpectPC uint16
	}{
		{0xa, 0x0, 0x2},
		{0xa, 0x1, 0x4},
		{0x2, 0x0, 0x2},
		{0x2, 0x1, 0x4},
	}

	for _, tc := range tcs {
		c := NewDefaultChip()
		c.OpCode = 0xe09e | uint16(tc.InputKey) << 8

		c.Keypad[tc.InputKey] = tc.Pressed

		err := c.HandleOpcode()
		if assert.NoError(t, err) {
			assert.Equal(t, tc.ExpectPC, c.PC)
		}
	}
}

//EXA1	KeyOp	if(key()!=Vx)	Skips the next instruction if the key stored in VX isn't pressed. (Usually the next instruction is a jump to skip a code block)
func TestOpcodeEXA1(t *testing.T) {

	tcs := []struct{
		InputKey uint8
		Pressed uint8 // 1 or 0
		ExpectPC uint16
	}{
		{0xa, 0x1, 0x2},
		{0xa, 0x0, 0x4},
		{0x2, 0x1, 0x2},
		{0x2, 0x0, 0x4},
	}

	for _, tc := range tcs {
		c := NewDefaultChip()
		c.OpCode = 0xe0a1 | uint16(tc.InputKey) << 8

		c.Keypad[tc.InputKey] = tc.Pressed

		err := c.HandleOpcode()
		if assert.NoError(t, err) {
			assert.Equal(t, tc.ExpectPC, c.PC)
		}
	}
}

//FX07	Timer	Vx = get_delay()	Sets VX to the value of the delay timer.
func TestOpcodeFX07(t *testing.T) {
	c := NewDefaultChip()

	c.OpCode = 0xfa07
	c.DelayTimer = 0xc

	err := c.HandleOpcode()

	if assert.NoError(t, err) {
		assert.Equal(t, 2, int(c.PC))

		assert.Equal(t, uint8(0xc), c.V[0xa])
	}
}

//FX0A	KeyOp	Vx = get_key()	A key press is awaited, and then stored in VX. (Blocking Operation. All instruction halted until next key event)
func TestOpcodeFX0A(t *testing.T) {
	TODO(t)
}

//FX15	Timer	delay_timer(Vx)	Sets the delay timer to VX.
func TestOpcodeFX15(t *testing.T) {
	c := NewDefaultChip()

	c.OpCode = 0xfa15

	err := c.HandleOpcode()

	if assert.NoError(t, err) {
		assert.Equal(t, 2, int(c.PC))

		assert.Equal(t, uint8(0xa), c.DelayTimer)
	}
}

//FX18	Sound	sound_timer(Vx)	Sets the sound timer to VX.
func TestOpcodeFX18(t *testing.T) {
	c := NewDefaultChip()
	c.V[0xa] = 0x99
	c.OpCode = 0xfa10

	err := c.HandleOpcode()

	if assert.NoError(t, err){
		assert.Equal(t, uint16(0x2), c.PC)
	}
}

//FX1E	MEM	I +=Vx	Adds VX to I. VF is not affected.[c]
func TestOpcodeFX1E(t *testing.T) {
	c := NewDefaultChip()
	c.I = 0x2
	c.V[0xa] = 0x5

	c.OpCode = 0xfa1e

	err := c.HandleOpcode()

	if assert.NoError(t, err){
		assert.Equal(t, uint16(0x7), c.I)
		assert.Equal(t, uint16(0x2), c.PC)
	}
}

//FX29	MEM	I=sprite_addr[Vx]	Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
// TODO: currently test the starting location of the fonts in memory, but not if they are actually loaded
func TestOpcodeFX29(t *testing.T) {
	tcs := []struct {
		InputChar            uint8
		ExpectStartingMemLoc uint16
	}{
		{0x0, 0x0},
		{0x1, 0x5},
		{0x2, 0xa},
		{0x3, 0xf},
		{0x4, 0x14},
		{0x5, 0x19},
		{0x6, 0x1e},
		{0x7, 0x23},
		{0x8, 0x28},
		{0x9, 0x2d},
		{0xa, 0x32},
		{0xb, 0x37},
		{0xc, 0x3c},
		{0xd, 0x41},
		{0xe, 0x46},
		{0xf, 0x4b},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%x", tc.InputChar), func(t *testing.T) {
			c := NewDefaultChip()

			// set the character to print in register A
			c.V[0xa] = tc.InputChar

			c.OpCode = 0xfa29

			err := c.HandleOpcode()
			if assert.NoError(t, err) {
				assert.Equal(t, 2, int(c.PC))
				assert.Equal(t, tc.ExpectStartingMemLoc, c.I)
			}
		})
	}

}

//FX33	BCD	set_BCD(Vx); *(I+0)=BCD(3); *(I+1)=BCD(2); *(I+2)=BCD(1); //Stores the binary-coded decimal representation of VX, with the most significant of three digits at the address in I, the middle digit at I plus 1, and the least significant digit at I plus 2. (In other words, take the decimal representation of VX, place the hundreds digit in memory at location in I, the tens digit at location I+1, and the ones digit at location I+2.)
func TestOpcodeFX33(t *testing.T) {
	c := NewDefaultChip()
	c.V[2] = 0x7b //Vx = 111

	c.OpCode = 0xf233
	c.I = 5

	err := c.HandleOpcode()

	if assert.NoError(t, err) {

		assert.Equal(t, uint16(2), c.PC)

		assert.Equal(t, uint8(1), c.Memory[5])
		assert.Equal(t, uint8(2), c.Memory[6])
		assert.Equal(t, uint8(3), c.Memory[7])
	}
}

//FX55	MEM	reg_dump(Vx,&I)	Stores V0 to VX (including VX) in memory starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified.[d]
func TestOpcodeFX55(t *testing.T) {
	howManyRegistersToCheck := 11

	c := NewDefaultChip()

	// load up the registers in decreasing numbers
	for i := 0; i < howManyRegistersToCheck; i++ {
		c.V[i] = 0xf - uint8(i)
	}

	//set the starting memory location
	c.I = 0xf0

	// make sure the opcode contains the correct number of registers that need to be copied (the X in fX55)
	c.OpCode = 0xf055 | (uint16(howManyRegistersToCheck) << 8)

	err := c.HandleOpcode()

	if assert.NoError(t, err) {
		assert.Equal(t, 2, int(c.PC))

		for i := 0; i < howManyRegistersToCheck; i++ {
			assert.Equal(t, 15-i, int(c.Memory[0xf0+uint16(i)]))
		}
	}
}

//FX65	MEM	reg_load(Vx,&I)	Fills V0 to VX (including VX) with values from memory starting at address I. The offset from I is increased by 1 for each value written, but I itself is left unmodified.
func TestOpcodeFX65(t *testing.T) {
	c := NewDefaultChip()
	c.I = 0

	c.Memory[c.I] = 1
	c.Memory[c.I+1] = 2
	c.Memory[c.I+2] = 3
	c.Memory[c.I+3] = 4
	c.Memory[c.I+4] = 5
	c.Memory[c.I+5] = 6

	c.OpCode = 0xf565

	err := c.HandleOpcode()

	if assert.NoError(t, err) {
		assert.Equal(t, uint16(2), c.PC)

		assert.Equal(t, uint8(1), c.V[0])
		assert.Equal(t, uint8(2), c.V[1])
		assert.Equal(t, uint8(3), c.V[2])
		assert.Equal(t, uint8(4), c.V[3])
		assert.Equal(t, uint8(5), c.V[4])
	}
}

func TODO(t *testing.T) {
	t.Skip("not yet implemented")
}
