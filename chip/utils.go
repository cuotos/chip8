package chip

import (
	"fmt"
	"log"
)

func (c *Chip8) DiagDump() {

	for i := 0; i < len(c.Memory); i += 16 {

		if (i % 16) == 0 {
			log.Printf("[DEBUG] %08x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x\n", i,
				c.Memory[i], c.Memory[i+1],
				c.Memory[i+2], c.Memory[i+3],
				c.Memory[i+4], c.Memory[i+5],
				c.Memory[i+6], c.Memory[i+7],
				c.Memory[i+8], c.Memory[i+9],
				c.Memory[i+10], c.Memory[i+11],
				c.Memory[i+12], c.Memory[i+13],
				c.Memory[i+14], c.Memory[i+15])
		}
	}

	log.Printf("[DEBUG] PC: %04x\n", c.PC)
	log.Printf("[DEBUG] SP: %04x\n", c.SP)
	log.Printf("[DEBUG] Stck: %04x\n", c.Stack)

	log.Printf("[DEBUG] Rgst: %v\n", func() []string {
		output := []string{}
		for i, r := range c.V {
			output = append(output, fmt.Sprintf("%x:%04x", i, r))
		}
		return output
	}())
}
