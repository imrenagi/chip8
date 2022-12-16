package chip8

import (
	"fmt"
	"time"
)

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

func (c *CPU) Start() {

}

func (c *CPU) Execute(instruction uint16) {
	msb := (instruction & 0xF000) >> 12
	switch msb {
	case 0x0:
		switch ((instruction & 0x0FFF) << 4) >> 4 {
		case 0x0E0:
			fmt.Println("00E0 - CLS")
		case 0x0EE:
			fmt.Println("00EE - RET")
		}
	case 0x1:
		fmt.Println("1nnn - JP addr")
	case 0x2:
		fmt.Println("2nnn - CALL addr")
	case 0x3:
		fmt.Println("3xkk - SE Vx, byte")
	case 0x4:
		fmt.Println("4xkk - SNE Vx, byte")
	case 0x5:
		fmt.Println("5xy0 - SE Vx, Vy")
	case 0x6:
		fmt.Println("6xkk - LD Vx, byte")
	case 0x7:
		fmt.Println("7xkk - ADD Vx, byte")
	case 0x8:
		switch ((instruction & 0x000F) << 12) >> 12 {
		case 0x0:
			fmt.Println("8xy0 - LD Vx, Vy")
		case 0x1:
			fmt.Println("8xy1 - OR Vx, Vy")
		case 0x2:
			fmt.Println("8xy2 - AND Vx, Vy")
		case 0x3:
			fmt.Println("8xy3 - XOR Vx, Vy")
		case 0x4:
			fmt.Println("8xy4 - ADD Vx, Vy")
		case 0x5:
			fmt.Println("8xy5 - SUB Vx, Vy")
		case 0x6:
			fmt.Println("8xy6 - SHR Vx {, Vy}")
		case 0x7:
			fmt.Println("8xy7 - SUBN Vx, Vy")
		case 0xE:
			fmt.Println("8xyE - SHL Vx {, Vy}")
		default:
			panic("instruction not recognized")
		}
	case 0x9:
		fmt.Println("9xy0 - SNE Vx, Vy")
	case 0xA:
		fmt.Println("Annn - LD I, addr")
	case 0xB:
		fmt.Println("Bnnn - JP V0, addr")
	case 0xC:
		fmt.Println("Cxkk - RND Vx, byte")
	case 0xD:
		fmt.Println("Dxyn - DRW Vx, Vy, nibble")
	case 0xE:
		switch ((instruction & 0x00FF) << 8) >> 8 {
		case 0x9E:
			fmt.Println("Ex9E - SKP Vx")
		case 0xA1:
			fmt.Println("ExA1 - SKNP Vx")
		default:
			panic("instruction not recognized")
		}
	case 0xF:
		switch ((instruction & 0x00FF) << 8) >> 8 {
		case 0x07:
			fmt.Println("Fx07 - LD Vx, DT")
		case 0x0A:
			fmt.Println("Fx0A - LD Vx, K")
		case 0x15:
			fmt.Println("Fx15 - LD DT, Vx")
		case 0x18:
			fmt.Println("Fx18 - LD ST, Vx")
		case 0x1E:
			fmt.Println("Fx1E - ADD I, Vx")
		case 0x29:
			fmt.Println("Fx29 - LD F, Vx")
		case 0x33:
			fmt.Println("Fx33 - LD B, Vx")
		case 0x55:
			fmt.Println("Fx55 - LD [I], Vx")
		case 0x65:
			fmt.Println("Fx65 - LD Vx, [I]")
		default:
			panic("instruction not recognized")
		}
	default:
		panic("instruction not recognized")
	}

}
