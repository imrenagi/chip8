package chip8

import "fmt"

func DefaultDisplay() *Display {
	d := &Display{
		H: 32,
		W: 64,
	}
	d.init()

	return d
}

type Display struct {
	H, W uint8
	data [][]uint8
}

func (d *Display) init() {
	row := make([][]uint8, d.H)
	for r := 0; r < int(d.H); r++ {
		colums := make([]uint8, d.W)
		for c := 0; c < int(d.W); c++ {
			colums[c] = 0
		}
		row[r] = colums
	}
	d.data = row
}

func (d *Display) Clear() {
	d.init()
}

func (d *Display) SetPixel(x, y uint8, val uint8) {
	d.data[y][x] = val
}

func (d *Display) GetPixel(x, y uint8) uint8 {
	return d.data[y][x]
}

func (d *Display) Draw() {
	for _, r := range d.data {
		for _, c := range r {
			if c > 0 {
				fmt.Print("o")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}
