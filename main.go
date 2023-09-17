package main

import (
	"github.com/CanPacis/brainfuck-interpreter/engine"
)

func main() {
	e := engine.NewEngine(engine.EngineOptions{
		FilePath: "./bf/io.bfi",
		IOSourceList: engine.IOSourceList{
			File: "./bf/index.html",
		},
	})

	e.Run()
}
