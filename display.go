package chip8

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
	"github.com/rs/zerolog/log"
)

func NewDisplay(drawer Drawer) *Display {
	d := &Display{
		H:      50,
		W:      100,
		drawer: drawer,
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

type Drawer interface {
	Clear()
	SetPixel(x, y int)
	Draw()
}

func NewTcellDisplay(screen tcell.Screen) *TcellDisplay {
	encoding.Register()
	if err := screen.Init(); err != nil {
		log.Fatal().Msgf("not able to init screen")
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorRed)
	screen.SetStyle(defStyle)
	screen.Clear()

	d := &TcellDisplay{
		screen: screen,
	}

	return d
}

type TcellDisplay struct {
	screen tcell.Screen
}

func (t *TcellDisplay) SetPixel(x, y int) {
	c := ' '
	var comb []rune
	w := runewidth.RuneWidth(c)
	if w == 0 {
		comb = []rune{c}
		c = ' '
		w = 1
	}
	t.screen.SetContent(x, y, c, comb,
		tcell.StyleDefault.Background(tcell.ColorRed))
}

func (t *TcellDisplay) Clear() {
	t.screen.Clear()
}

func (t *TcellDisplay) Draw() {
	t.screen.Sync()
}
