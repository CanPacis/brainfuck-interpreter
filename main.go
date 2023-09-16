package main

import (
	"os"
	"slices"

	"github.com/CanPacis/brainfuck-interpreter/debugger"
	"github.com/CanPacis/brainfuck-interpreter/runtime"
)

func main() {
	if len(os.Args) > 1 {
		options := runtime.RuntimeOptions{
			FilePath: os.Args[1],
		}

		if slices.Contains(os.Args, "--debug") {
			options.Debugger = debugger.NewDebugger()
		}

		eso := runtime.NewRuntime(options)
		eso.Run()
	}
}
