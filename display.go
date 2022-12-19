package chip8

import (
	"fmt"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

func DefaultDisplay() *Display {
	d := &Display{
		H:      64,
		W:      128,
		drawer: NewSDLDisplay(),
	}
	d.init()
	return d
}

type Display struct {
	drawer Drawer

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
	d.drawer.Clear()
}

func (d *Display) SetPixel(x, y uint8, val uint8) {
	d.data[y][x] = val
}

func (d *Display) GetPixel(x, y uint8) uint8 {
	return d.data[y][x]
}

func (d *Display) Draw() {
	d.drawer.Clear()
	for y, r := range d.data {
		for x, c := range r {
			if c > 0 {
				d.drawer.SetPixel(x, y)
			}
		}
	}
	d.drawer.Draw()
}

func (d *Display) Stop() {
	d.drawer.Stop()
}

type Drawer interface {
	Clear()
	SetPixel(x, y int)
	Draw()
	Stop()
}

func NewSDLDisplay() *SDLDisplay {
	window, err := sdl.CreateWindow("chip-8 emulator",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		1280, 640, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	return &SDLDisplay{
		window:  window,
		surface: surface,
	}
}

type SDLDisplay struct {
	sync.Mutex
	window  *sdl.Window
	surface *sdl.Surface
}

func (s *SDLDisplay) Clear() {
	s.surface.FillRect(nil, 0)
}

func (s *SDLDisplay) SetPixel(x, y int) {
	rect := sdl.Rect{int32(x) * 10, int32(y) * 10, 10, 10}
	s.surface.FillRect(&rect, 0xffff0000)
}

func (s *SDLDisplay) Draw() {
	s.window.UpdateSurface()
}

func (s *SDLDisplay) Stop() {
	fmt.Println("stop sdl display")
	s.window.Destroy()
}
