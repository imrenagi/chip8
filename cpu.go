package chip8

import (
	"context"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	fonts = []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
)

const (
	// clockDuration = 2 * time.Millisecond
	clockFrequency = 500 // Hz
	timerFrequency = 60
)

func NewCPU() *CPU {
	clockDuration := math.Round(float64(1) / float64(clockFrequency) * 1000)

	cpu := &CPU{
		PC:       0x200,
		Display:  DefaultDisplay(),
		Keyboard: &Keyboard{},
		clock:    time.NewTicker(time.Duration(clockDuration) * time.Millisecond),
	}
	fontStartAddr := 0x050
	for _, b := range fonts {
		cpu.Memory[fontStartAddr] = b
		fontStartAddr++
	}
	return cpu
}

type CPU struct {
	// Memory will be accessed by The Chip-8. The Chip-8 language is capable of accessing up to 4KB (4,096 bytes) of RAM,
	// from location 0x000 (0) to 0xFFF (4095).
	// The first 512 bytes, from 0x000 to 0x1FF, are where the original interpreter was located, and should not be used
	// by programs. Most Chip-8 programs start at location 0x200 (512), but some begin at 0x600 (1536).
	// All instructions are 2 bytes long and are stored most-significant-byte first
	Memory [4096]uint8

	// V is 16 general purpose 8-bit registers, usually referred to as Vx, where x is a hexadecimal digit (0 through F)
	V [16]uint8

	// I is 16-bit registers. This register is generally used to store memory addresses
	I uint16

	// The VF register should not be used by any program, as it is used as a flag by some instructions.
	// VF uint8

	// Chip-8 also has two special purpose 8-bit registers, for the delay and sound timers.
	// When these registers are non-zero, they are automatically decremented at a rate of 60Hz
	DT uint8
	ST uint8

	// The program counter (PC) should be 16-bit, and is used to store the currently executing address
	PC uint16

	// The stack pointer (SP) can be 8-bit, it is used to point to the topmost level of the stack.
	SP uint8

	// The stack is an array of 16 16-bit values, used to store the address that the interpreter should return to when
	// finished with a subroutine.
	// Chip-8 allows for up to 16 levels of nested subroutines.
	Stack [16]uint16

	Display *Display

	Keyboard *Keyboard

	clock *time.Ticker
}

func (c *CPU) LoadProgram(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		c.Memory[c.PC+uint16(i)] = buf[i]
	}
	return nil
}

func (c *CPU) Start(ctx context.Context) {

	step := 0
	count := int(math.Round(float64(clockFrequency) / float64(timerFrequency)))

	for {
		select {
		case <-c.clock.C:
			instruction := c.Fetch()
			c.DecodeAndExecute(instruction)
			if step == count {
				if c.DT > 0 {
					c.DT--
				}
				if c.ST > 0 {
					c.ST--
				}
				step = 0
			} else {
				step++
			}
		case <-ctx.Done():
			log.Info().Msgf("stopping")
			c.clock.Stop()
			return
		}
	}
}

func (c *CPU) Fetch() uint16 {
	msb := c.Memory[c.PC]
	lsb := c.Memory[c.PC+1]
	c.PC += 2
	return uint16(msb)<<8 | uint16(lsb)
}

func (c *CPU) DecodeAndExecute(instruction uint16) {
	msb := (instruction & 0xF000) >> 12
	nnn := ((instruction << 4) & 0xFFFF) >> 4
	kk := instruction & 0x00FF
	n := instruction & 0x000F
	x := (instruction & 0x0F00) >> 8
	y := (instruction & 0x00F0) >> 4

	switch msb {
	case 0x0:
		switch ((instruction & 0x0FFF) << 4) >> 4 {
		case 0x0E0:
			c.clearScreen()
		case 0x0EE:
			c.ret()
		}
	case 0x1:
		c.jump(nnn)
	case 0x2:
		c.call(nnn)
	case 0x3:
		c.skipIfEqual(uint8(x), uint8(kk))
	case 0x4:
		c.skipIfNotEqual(uint8(x), uint8(kk))
	case 0x5:
		lsb := instruction & 0x000F
		if lsb == 0 {
			c.compareReg(uint8(x), uint8(y))
		}
	case 0x6:
		c.setValue(uint8(x), uint8(kk))
	case 0x7:
		c.addValue(uint8(x), uint8(kk))
	case 0x8:
		switch ((instruction & 0x000F) << 12) >> 12 {
		case 0x0:
			c.store(uint8(x), uint8(y))
		case 0x1:
			c.or(uint8(x), uint8(y))
		case 0x2:
			c.and(uint8(x), uint8(y))
		case 0x3:
			c.xor(uint8(x), uint8(y))
		case 0x4:
			c.sum(uint8(x), uint8(y))
		case 0x5:
			c.sub(uint8(x), uint8(y))
		case 0x6:
			c.shr(uint8(x))
		case 0x7:
			c.subn(uint8(x), uint8(y))
		case 0xE:
			c.shl(uint8(x))
		default:
			panic("instruction not recognized")
		}
	case 0x9:
		c.sne(uint8(x), uint8(y))
	case 0xA:
		c.setI(nnn)
	case 0xB:
		c.jumpFromV0(nnn)
	case 0xC:
		c.rnd(uint8(x), uint8(kk))
	case 0xD:
		c.drw(uint8(x), uint8(y), uint8(n))
	case 0xE:
		switch ((instruction & 0x00FF) << 8) >> 8 {
		case 0x9E:
			c.skipIfKeyPressed(uint8(x))
		case 0xA1:
			c.skipIfKeyNotPressed(uint8(x))
		default:
			panic("instruction not recognized")
		}
	case 0xF:
		switch ((instruction & 0x00FF) << 8) >> 8 {
		case 0x07:
			c.storeDelayTimerToRegister(uint8(x))
		case 0x0A:
			c.waitKeyPressedAndStoreToRegister(uint8(x))
		case 0x15:
			c.setDelayTimerFromRegister(uint8(x))
		case 0x18:
			c.setSoundTimerFromRegister(uint8(x))
		case 0x1E:
			c.addIWithV(uint8(x))
		case 0x29:
			c.setIWithSpriteLocationOfRegisterVal(uint8(x))
		case 0x33:
			c.storeBCD(uint8(x))
		case 0x55:
			c.storeVRegisterToMemory(uint8(x))
		case 0x65:
			c.loadMemoryToVRegister(uint8(x))
		default:
			panic("instruction not recognized")
		}
	default:
		panic("instruction not recognized")
	}

}

// clearScreen Clear the display.
// 00E0 - CLS
func (c *CPU) clearScreen() {
	log.Debug().Msgf("00E0 - CLS")
	c.Display.Clear()
	c.Display.Draw()
}

// ret Returns from a subroutine.
// The interpreter sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
// 00EE - RET
func (c *CPU) ret() {
	log.Debug().Msgf("00EE - RET")
	c.Stack[c.SP] = c.PC
	c.SP--
}

// Jump to location nnn.
// The interpreter sets the program counter to nnn.
// 1nnn - JP addr
func (c *CPU) jump(addr uint16) {
	log.Debug().Msgf("1nnn - JP addr")
	c.PC = addr
}

// call Calls subroutine at nnn.
// The interpreter increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then setValue to nnn.
// 2nnn - CALL addr
func (c *CPU) call(addr uint16) {
	log.Debug().Msgf("2nnn - CALL addr")
	c.SP++
	c.Stack[c.SP] = c.PC
	c.PC = addr
}

// skipIfEqual Skip next instruction if Vx = kk.
// The interpreter compares register Vx to kk, and if they are equal, increments the program counter by 2.
// 3xkk - SE Vx, byte
func (c *CPU) skipIfEqual(regAddr uint8, val uint8) {
	log.Debug().Msgf("3xkk - SE Vx, byte")
	if c.V[regAddr] == val {
		c.PC += 2
	}
}

// skipIfNotEqual Skips next instruction if Vx != kk.
// The interpreter compares register Vx to kk, and if they are not equal, increments the program counter by 2.
// 4xkk - SNE Vx, byte
func (c *CPU) skipIfNotEqual(regAddr uint8, val uint8) {
	log.Debug().Msgf("4xkk - SNE Vx, byte")
	if c.V[regAddr] != val {
		c.PC += 2
	}
}

// compareReg Skips next instruction if Vx = Vy.
// The interpreter compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
// 5xy0 - SE Vx, Vy
func (c *CPU) compareReg(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("5xy0 - SE Vx, Vy")
	if c.V[xRegAddr] == c.V[yRegAddr] {
		c.PC += 2
	}
}

// setValue Sets Vx = kk.
// The interpreter puts the value kk into register Vx.
// 6xkk - LD Vx, byte
func (c *CPU) setValue(regAddr, val uint8) {
	log.Debug().Msgf("6xkk - LD Vx, byte")
	c.V[regAddr] = val
}

// Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
// 7xkk - ADD Vx, byte
func (c *CPU) addValue(regAddr, val uint8) {
	log.Debug().Msgf("7xkk - ADD Vx, byte")
	c.V[regAddr] += val
}

// Set Vx = Vy.
// Stores the value of register Vy in register Vx.
// 8xy0 - LD Vx, Vy
func (c *CPU) store(destAddr, srcAddr uint8) {
	log.Debug().Msgf("8xy0 - LD Vx, Vy")
	c.V[destAddr] = c.V[srcAddr]
}

// Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corresponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
// 8xy1 - OR Vx, Vy
func (c *CPU) or(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy1 - OR Vx, Vy")
	c.V[xRegAddr] |= c.V[yRegAddr]
}

// Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corrseponding bits from two values, and if both bits are 1, then the same
// bit in the result is also 1. Otherwise, it is 0.
// 8xy2 - AND Vx, Vy
func (c *CPU) and(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy2 - AND Vx, Vy")
	c.V[xRegAddr] &= c.V[yRegAddr]
}

// Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values, and if the bits are not both
// the same, then the corresponding bit in the result is setValue to 1. Otherwise, it is 0.
// 8xy3 - XOR Vx, Vy
func (c *CPU) xor(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy3 - XOR Vx, Vy")
	c.V[xRegAddr] ^= c.V[yRegAddr]
}

// Set Vx = Vx + Vy, setValue VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is setValue to 1, otherwise 0.
// Only the lowest 8 bits of the result are kept, and stored in Vx.
// 8xy4 - ADD Vx, Vy
func (c *CPU) sum(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy4 - ADD Vx, Vy")
	// TODO use 8-bit cary ripple adder
	sum := uint16(c.V[xRegAddr]) + uint16(c.V[yRegAddr])
	if sum > 255 {
		c.V[xRegAddr] = 0xFF
		c.V[0xF] = 1
	} else {
		c.V[xRegAddr] = uint8(sum)
		c.V[0xF] = 0
	}
}

// Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
// 8xy5 - SUB Vx, Vy
func (c *CPU) sub(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy5 - SUB Vx, Vy")
	if c.V[xRegAddr] > c.V[yRegAddr] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[xRegAddr] -= c.V[yRegAddr]
}

// Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
// 8xy6 - SHR Vx {, Vy}
func (c *CPU) shr(addr uint8) {
	log.Debug().Msgf("8xy6 - SHR Vx {, Vy}")
	c.V[0xF] = c.V[addr] & 0x01
	c.V[addr] = c.V[addr] >> 1
}

// Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0.
// Then Vx is subtracted from Vy, and the results stored in Vx.
// 8xy7 - SUBN Vx, Vy
func (c *CPU) subn(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("8xy7 - SUBN Vx, Vy")
	if c.V[yRegAddr] > c.V[xRegAddr] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[xRegAddr] = c.V[yRegAddr] - c.V[xRegAddr]
}

// Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
// 8xyE - SHL Vx {, Vy}
func (c *CPU) shl(addr uint8) {
	log.Debug().Msgf("8xyE - SHL Vx {, Vy}")
	c.V[0xF] = (c.V[addr] & 0x80) >> 7
	c.V[addr] = c.V[addr] << 1
}

// Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
// 9xy0 - SNE Vx, Vy
func (c *CPU) sne(xRegAddr, yRegAddr uint8) {
	log.Debug().Msgf("9xy0 - SNE Vx, Vy")
	if c.V[xRegAddr] != c.V[yRegAddr] {
		c.PC += 2
	}
}

// Set I = nnn.
// The value of register I is set to nnn.
// Annn - LD I, addr
func (c *CPU) setI(addr uint16) {
	log.Debug().Msgf("Annn - LD I, addr")
	c.I = addr
}

// Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
// Bnnn - JP V0, addr
func (c *CPU) jumpFromV0(addr uint16) {
	log.Debug().Msgf("Bnnn - JP V0, addr")
	c.PC = uint16(c.V[0x0]) + addr
}

// Set Vx = random byte AND kk.
// The interpreter generates a random number from 0 to 255, which is then ANDed with the value kk.
// The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
// Cxkk - RND Vx, byte
func (c *CPU) rnd(addr uint8, val uint8) {
	log.Debug().Msgf("Cxkk - RND Vx, byte")
	num := uint8(rand.Intn(256))
	c.V[addr] = num & val
}

// Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
// The interpreter reads n bytes from memory, starting at the address stored in I.
// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
// Sprites are XORed onto the existing screen. If this causes any pixels to be erased, VF is set to 1,
// otherwise it is set to 0. If the sprite is positioned so part of it is outside the coordinates of the display,
// it wraps around to the opposite side of the screen.
// See instruction 8xy3 for more information on XOR, and section 2.4, Display, for more information on the Chip-8
// screen and sprites.
// Dxyn - DRW Vx, Vy, nibble
func (c *CPU) drw(xRegAddr, yRegAddr, nibble uint8) {
	log.Debug().Msgf("Dxyn - DRW Vx, Vy, nibble")
	maxX := c.Display.W - 1
	maxY := c.Display.H - 1

	x := c.V[xRegAddr] % maxX
	y := c.V[yRegAddr] % maxY
	c.V[0xF] = 0

	for i := 0; i < int(nibble); i++ {
		nthByte := c.Memory[c.I+uint16(i)]
		for j := 7; j >= 0; j-- {

			screenX := x + uint8(7-j)
			screenY := y + uint8(i)

			spritePixel := (nthByte >> j) & 0x01
			screenPixel := c.Display.GetPixel(screenX, screenY)

			if screenX > maxX || screenY > maxY {
				continue
			}

			newVal := spritePixel ^ screenPixel
			if spritePixel == 1 && screenPixel == 1 {
				c.V[0xF] = 1
			}
			c.Display.SetPixel(screenX, screenY, newVal)
		}
	}
	c.Display.Draw()
}

// Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently in the down position, PC is increased by 2.
// Ex9E - SKP Vx
func (c *CPU) skipIfKeyPressed(addr uint8) {
	log.Debug().Msgf("Ex9E - SKP Vx")
	if c.Keyboard.IsPressed(c.V[addr]) {
		c.PC += 2
	}
}

// Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently in the up position, PC is increased by 2.
// ExA1 - SKNP Vx
func (c *CPU) skipIfKeyNotPressed(addr uint8) {
	log.Debug().Msgf("ExA1 - SKNP Vx")
	if !c.Keyboard.IsPressed(c.V[addr]) {
		c.PC += 2
	}
}

// Set Vx = delay delayTimer value.
// The value of DT is placed into Vx.
// Fx07 - LD Vx, DT
func (c *CPU) storeDelayTimerToRegister(addr uint8) {
	log.Debug().Msgf("ExA1 - SKNP Vx")
	c.V[addr] = c.DT
}

// Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
// Fx0A - LD Vx, K
func (c *CPU) waitKeyPressedAndStoreToRegister(addr uint8) {
	log.Debug().Msgf("Fx0A - LD Vx, K")
	key := <-c.Keyboard.PressedEventCh()
	c.V[addr] = key
}

// Set delay delayTimer = Vx.
// DT is set equal to the value of Vx.
// Fx15 - LD DT, Vx
func (c *CPU) setDelayTimerFromRegister(addr uint8) {
	log.Debug().Msgf("Fx15 - LD DT, Vx")
	c.DT = c.V[addr]
}

// Set sound delayTimer = Vx.
// ST is set equal to the value of Vx.
// Fx18 - LD ST, Vx
func (c *CPU) setSoundTimerFromRegister(addr uint8) {
	log.Debug().Msgf("Fx18 - LD ST, Vx")
	c.ST = c.V[addr]
}

// Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
// Fx1E - ADD I, Vx
func (c *CPU) addIWithV(addr uint8) {
	log.Debug().Msgf("Fx1E - ADD I, Vx")
	c.I += uint16(c.V[addr])
}

// Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite
// corresponding to the value of Vx.
// See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
// Fx29 - LD F, Vx
func (c *CPU) setIWithSpriteLocationOfRegisterVal(addr uint8) {
	log.Debug().Msgf("Fx29 - LD F, Vx")
	key := c.V[addr] & 0x0F
	c.I = 0x050 + uint16(key*5)
}

// Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The interpreter takes the decimal value of Vx, and places the hundreds digit
// in memory at location in I, the tens digit at location I+1,
// and the ones digit at location I+2.
// Fx33 - LD B, Vx
func (c *CPU) storeBCD(addr uint8) {
	log.Debug().Msgf("Fx33 - LD B, Vx")
	val := c.V[addr]

	var iOffset uint16 = 2
	for val != 0 {
		modulo := val % 10
		val = val / 10

		c.Memory[c.I+iOffset] = modulo
		iOffset--
	}
}

// Store registers V0 through Vx in memory starting at location I.
// The interpreter copies the values of registers V0 through Vx into memory, starting at the address in I.
// Fx55 - LD [I], Vx
func (c *CPU) storeVRegisterToMemory(maxAddr uint8) {
	log.Debug().Msgf("Fx55 - LD [I], Vx")
	for i := 0; i <= int(maxAddr); i++ {
		c.Memory[c.I+uint16(i)] = c.V[i]
	}
}

// Read registers V0 through Vx from memory starting at location I.
// The interpreter reads values from memory starting at location I into registers V0 through Vx.
// Fx65 - LD Vx, [I]
func (c *CPU) loadMemoryToVRegister(maxAddr uint8) {
	log.Debug().Msgf("Fx65 - LD Vx, [I]")
	for i := 0; i <= int(maxAddr); i++ {
		c.V[i] = c.Memory[c.I+uint16(i)]
	}
}
