package main

import (
	"github.com/StructsNotClasses/mim/instance"

	gnc "github.com/rthornton128/goncurses"

	"log"
)

func main() {
	const musicDirectory = "/mnt/music"
	const configFile = "/home/pugpugpugs/mim/config.mim"

	// start ncurses
	backgroundWindow, err := gnc.Init()
	if err != nil {
		log.Fatal(err)
	}
	gnc.CBreak(true)
	gnc.Echo(false)
	backgroundWindow.Keypad(true)

	// initialize program state
	program, err := instance.New(backgroundWindow, musicDirectory)
	if err != nil {
		gnc.End()
		log.Fatal(err)
	}
	shouldExit, err := program.PassFileToInput(configFile)
	if err != nil {
		gnc.End()
		log.Fatal(err)
	}
	if shouldExit {
		gnc.End()
		log.Fatal("Warning: ':exit' was called in the initial config file. This is probably in error.\n")
	}

	program.Run()
	gnc.End()
}
