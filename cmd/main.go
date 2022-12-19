package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/imrenagi/chip8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)

	go func() {
		oscall := <-ch
		log.Warn().Msgf("system call:%+v", oscall)
		cancel()
	}()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	keyboard := chip8.NewKeyboard()

	c := chip8.NewCPU(
		chip8.DefaultDisplay(),
		keyboard,
		chip8.NewAudioController(),
	)
	// c.LoadProgram("examples/IBM_Logo.ch8")
	// c.LoadProgram("examples/keypad_test.ch8")
	// c.LoadProgram("examples/chiptest.ch8")
	c.LoadProgram("examples/c8games/TETRIS")
	// c.LoadProgram("examples/delay_timer_test.ch8")
	go c.Start(ctx)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				ke := event.(*sdl.KeyboardEvent)
				var pressed bool
				if ke.State == sdl.PRESSED {
					pressed = true
				}
				keyEvent := chip8.KeyEvent{
					Pressed:  pressed,
					ScanCode: ke.Keysym.Scancode,
				}
				keyboard.Accept(keyEvent)
			}
		}
	}
}
