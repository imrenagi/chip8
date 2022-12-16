package main

import (
	"github.com/imrenagi/chip8"
)

func main() {
	c := chip8.CPU{}
	c.Execute(0x1111)
	c.Execute(0x00E0)
}
