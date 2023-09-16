package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/debugger"
	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Engine struct {
	Path       string
	Name       string
	Content    string
	Debugger   debugger.Debugger
	Parser     parser.Parser
	Tape       [30000]uint32
	Band       [30000]byte
	Cursor     uint
	Io         RuntimeIo
	OriginalIo RuntimeIo
}

func run(e *Engine, program []parser.Statement) bf_errors.FileError {
	index := 0

	for ; index < len(program); index++ {
		statement := program[index]

		if e.Debugger.Exists && statement.DebugTarget {
			action := e.Debugger.Wait(e.CreateDebugState(statement))
			switch action.Operation {
			case "step":
				if len(program) > index+1 {
					program[index+1].DebugTarget = true
				}
			case "step-over":
				index++
			case "step-out":
				return bf_errors.EmptyError
			}
		}

		switch statement.Type {
		case "Push Statement":
			e.Band[e.Cursor] = byte(statement.Value)
			e.Tape[e.Cursor] = statement.Value
		case "Increment Statement":
			e.Band[e.Cursor]++
			e.Tape[e.Cursor] = uint32(e.Band[e.Cursor])
		case "Decrement Statement":
			e.Band[e.Cursor]--
			e.Tape[e.Cursor] = uint32(e.Band[e.Cursor])
		case "Clear Statement":
			e.Tape = [30000]uint32{}
			e.Band = [30000]byte{}
		case "Move Right Statement":
			if e.Cursor < 30000 {
				e.Cursor++
			}
		case "Move Left Statement":
			if e.Cursor > 0 {
				e.Cursor--
			}
		case "Loop Statement":
			for e.Band[e.Cursor] != 0 {
				err := run(e, statement.Body)
				if err.Reason != nil {
					return err
				}
			}
		case "Stdout Statement":
			switch e.Tape[e.Cursor] {
			default:
				e.Io.Out.Write([]byte{byte(e.Band[e.Cursor])})
			}
		case "Stdin Statement":
			switch e.Tape[e.Cursor] {
			default:
				reader := bufio.NewReader(os.Stdin)
				char, _, err := reader.ReadRune()

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
				}

				e.Band[e.Cursor] = byte(char)
				e.Tape[e.Cursor] = uint32(e.Band[e.Cursor])
			}
		case "Switch IO Statement":
			fmt.Println(statement.IoTarget)
		}
	}

	return bf_errors.EmptyError
}

func (e *Engine) Run() {
	if e.Debugger.Exists {
		e.Debugger.Open(debugger.DebugMetaData{
			FileName: e.Name,
			FilePath: e.Path,
			Content:  e.Content,
		})
	}

	err := e.Parser.Parse(e.Content)
	if err.Reason != nil {
		err.Write(e.Io.Err)
		if e.Debugger.Exists {
			e.Debugger.Error(err)
			e.Debugger.Close()
		}
		os.Exit(1)
	}

	err = run(e, e.Parser.Program)
	if err.Reason != nil {
		err.Write(e.Io.Err)
		if e.Debugger.Exists {
			e.Debugger.Error(err)
			e.Debugger.Close()
		}
		os.Exit(1)
	}

	if e.Debugger.Exists {
		e.Debugger.Close()
	}
}

func (e *Engine) CreateDebugState(statement parser.Statement) debugger.DebugState {
	return debugger.DebugState{
		Statement: statement,
		Cursor:    e.Cursor,
		Tape:      e.Tape[:100],
		Band:      e.Band[:100],
	}
}

type RuntimeIo struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

type EngineOptions struct {
	FilePath       string
	AttachDebugger bool
	Stdout         io.Writer
	Stderr         io.Writer
	Stdin          io.Reader
}

func NewEngine(options EngineOptions) *Engine {
	name := path.Base(options.FilePath)
	content, err := os.ReadFile(options.FilePath)

	std := RuntimeIo{
		Out: options.Stdout,
		In:  options.Stdin,
		Err: options.Stderr,
	}

	r := &Engine{
		Name:    name,
		Path:    options.FilePath,
		Content: string(content),
		Parser:  parser.NewParser(options.FilePath),
		Io:      std,
	}

	if r.Io.Out == nil {
		r.Io.Out = os.Stdout
	}

	if r.Io.Err == nil {
		r.Io.Err = os.Stderr
	}

	if r.Io.In == nil {
		r.Io.In = os.Stdin
	}

	r.OriginalIo = r.Io

	if options.AttachDebugger {
		r.Debugger = debugger.NewDebugger(r.Io.Out)
	}

	if err != nil {
		r.Io.Err.Write([]byte(err.Error()))
		os.Exit(1)
	}

	return r
}
