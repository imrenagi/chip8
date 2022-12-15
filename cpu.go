package chip8

import "time"

// clock chip-8 approximately has 500Hz clock. ~ 0.12s
var clock = time.Tick(120 * time.Millisecond)

type CPU struct {
	// Memory will be accessed by The Chip-8. The Chip-8 language is capable of accessing up to 4KB (4,096 bytes) of RAM,
	// from location 0x000 (0) to 0xFFF (4095).
	// The first 512 bytes, from 0x000 to 0x1FF, are where the original interpreter was located, and should not be used
	// by programs. Most Chip-8 programs start at location 0x200 (512), but some begin at 0x600 (1536).
	// All instructions are 2 bytes long and are stored most-significant-byte first
	Memory [4096]uint16

	// V is 16 general purpose 8-bit registers, usually referred to as Vx, where x is a hexadecimal digit (0 through F)
	V [16]uint8

	// I is 16-bit registers. This register is generally used to store memory addresses
	I uint16

	// The VF register should not be used by any program, as it is used as a flag by some instructions.
	VF uint8

	// Chip-8 also has two special purpose 8-bit registers, for the delay and sound timers.
	// When these registers are non-zero, they are automatically decremented at a rate of 60Hz
	Delay      uint8
	SoundTimes uint8

	// The program counter (PC) should be 16-bit, and is used to store the currently executing address
	PC uint16

	// The stack pointer (SP) can be 8-bit, it is used to point to the topmost level of the stack.
	SP uint8

	// The stack is an array of 16 16-bit values, used to store the address that the interpreter should return to when
	// finished with a subroutine.
	// Chip-8 allows for up to 16 levels of nested subroutines.
	Stack [16]uint16
}
