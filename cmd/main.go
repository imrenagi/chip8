package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/imrenagi/chip8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logFile, err := os.OpenFile("chip8.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to create log file")
	}
	defer logFile.Close()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile})

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

	c := chip8.NewCPU()
	// c.LoadProgram("examples/IBM_Logo.ch8")
	c.LoadProgram("examples/c8games/TETRIS")
	c.Start(ctx)

	// c.DecodeAndExecute(0x60FF)
	// c.DecodeAndExecute(0x6101)
	// fmt.Println(c.V)
	// c.DecodeAndExecute(0x8014)
	// fmt.Println(c.V)
	// <-time.After(5 * time.Second)
}
