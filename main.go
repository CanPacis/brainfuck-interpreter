package main

import (
	"github.com/CanPacis/brainfuck-interpreter/runtime"
)

func main() {
	r := runtime.NewRuntime(runtime.RuntimeOptions{
		FilePath: "./bf/syntax_error.bf",
	})

	r.Run()
}
