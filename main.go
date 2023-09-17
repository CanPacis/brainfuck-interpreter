package main

import (
	"github.com/CanPacis/brainfuck-interpreter/bf_io"
	"github.com/CanPacis/brainfuck-interpreter/engine"
	"github.com/alecthomas/kong"
)

type Run struct {
	Path  string `arg:"" name:"path" type:"path"`
	Debug bool   `help:"Attach a debugger (currently not working)."`
	File  string `help:"Provide an io source for file. The default is 'io.txt'."`
	Http  string `help:"Provide an io source for http. The default is ':8080'."`
}

func (r *Run) Run(ctx *kong.Context) error {
	e := engine.NewEngine(engine.EngineOptions{
		FilePath:       r.Path,
		AttachDebugger: r.Debug,
		IOSourceList: bf_io.IOSourceList{
			File: r.File,
			Http: r.Http,
		},
	})

	e.Run()
	return nil
}

var CLI struct {
	Run Run `cmd:"run"`
}

func main() {
	ctx := kong.Parse(&CLI)

	switch ctx.Command() {
	case "run <path>":
		ctx.Run()
	default:
		panic(ctx.Command())
	}
}
