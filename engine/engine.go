package engine

import (
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/bf_io"
	"github.com/CanPacis/brainfuck-interpreter/debugger"
	"github.com/CanPacis/brainfuck-interpreter/parser"
	"github.com/CanPacis/brainfuck-interpreter/waiter"
)

type Engine struct {
	Path         string
	Name         string
	Content      string
	Debugger     debugger.Debugger
	Parser       parser.Parser
	Tape         [30000]byte
	Cursor       uint
	IOTargets    []bf_io.RuntimeIO
	IOSourceList bf_io.IOSourceList
	ioTargetType bf_io.IOTargetType
	originalIO   bf_io.RuntimeIO
	disposers    []func()
	waiters      waiter.EngineWaiter
	httpServer   *http.Server
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
		case "Increment Statement":
			e.Tape[e.Cursor]++
		case "Decrement Statement":
			e.Tape[e.Cursor]--
		case "Clear Statement":
			e.Tape = [30000]byte{}
		case "Move Right Statement":
			if e.Cursor < 30000 {
				e.Cursor++
			}
		case "Move Left Statement":
			if e.Cursor > 0 {
				e.Cursor--
			}
		case "Loop Statement":
			for e.Tape[e.Cursor] != 0 {
				err := run(e, statement.Body)
				if err.Reason != nil {
					return err
				}
			}
		case "Stdout Statement":
			if e.ioTargetType == bf_io.Http && len(e.IOTargets) == 0 {
				e.waiters.Wait("http")
			}
			for _, target := range e.IOTargets {
				if e.ioTargetType == bf_io.Http {
					if e.Tape[e.Cursor] != 0 {
						target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
					}
				} else {
					target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
				}
			}
			if e.ioTargetType == bf_io.Http && e.Tape[e.Cursor] == 0 {
				e.IOTargets = []bf_io.RuntimeIO{}
				e.waiters.Add("http", 1)
				e.waiters.Done("write")
			}
		case "Stdin Statement":
			var target bf_io.RuntimeIO

			if len(e.IOTargets) == 0 {
				target = e.originalIO
			} else {
				target = e.IOTargets[0]
			}
			byte, err := target.Reader.ReadByte()
			if err != nil && err != io.EOF {
				return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
			}
			e.Tape[e.Cursor] = byte
		case "Switch IO Statement":
			time.Sleep(time.Millisecond)
			if e.ioTargetType == bf_io.Http {
				e.waiters.Done("program")
				if e.httpServer != nil {
					e.httpServer.Close()
				}
			}

			e.ioTargetType = statement.IOTarget
			switch statement.IOTarget {
			case "std":
				e.IOTargets = []bf_io.RuntimeIO{e.originalIO}
			case "http":
				e.IOTargets = []bf_io.RuntimeIO{}
				e.httpServer = bf_io.HttpIO(e.IOSourceList.Http, e.IOSourceList.File, &e.IOTargets, e.waiters)
			case "tcp":
				e.originalIO.Out.Write([]byte("tcp"))
			case "file":
				io, close, err := bf_io.FileIO(e.IOSourceList.File)

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
				}

				e.disposers = append(e.disposers, func() {
					close()
				})

				io.Set(io)
				e.IOTargets = []bf_io.RuntimeIO{io}
			}
		}
	}

	return bf_errors.EmptyError
}

func (e *Engine) dispose(err bf_errors.FileError) {
	for _, disposer := range e.disposers {
		disposer()
	}

	if e.Debugger.Exists {
		if err.Reason != nil {
			err.Write(e.originalIO.Err)
			e.Debugger.Error(err)
		}
		e.Debugger.Close()
	} else {
		if err.Reason != nil {
			err.Write(e.originalIO.Err)
		}
	}

	if e.httpServer != nil {
		e.httpServer.Close()
	}
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
		e.dispose(err)
		os.Exit(1)
	}

	err = run(e, e.Parser.Program)
	if err.Reason != nil {
		e.dispose(err)
		os.Exit(1)
	}

	e.waiters.Wait("program")
	e.dispose(bf_errors.EmptyError)
}

func (e *Engine) CreateDebugState(statement parser.Statement) debugger.DebugState {
	return debugger.DebugState{
		Statement: statement,
		Cursor:    e.Cursor,
		Tape:      e.Tape[:100],
	}
}

type EngineOptions struct {
	FilePath       string
	AttachDebugger bool
	Stdout         io.Writer
	Stderr         io.Writer
	Stdin          io.Reader
	IOSourceList   bf_io.IOSourceList
}

func NewEngine(options EngineOptions) *Engine {
	name := path.Base(options.FilePath)
	content, err := os.ReadFile(options.FilePath)

	std := bf_io.RuntimeIO{
		Out: options.Stdout,
		In:  options.Stdin,
		Err: options.Stderr,
	}

	e := &Engine{
		Name:      name,
		Path:      options.FilePath,
		Content:   string(content),
		Parser:    parser.NewParser(options.FilePath),
		IOTargets: []bf_io.RuntimeIO{std},
		waiters: waiter.EngineWaiter{
			Program:        &sync.WaitGroup{},
			HttpConnection: &sync.WaitGroup{},
			Write:          &sync.WaitGroup{},
		},
		IOSourceList: options.IOSourceList,
	}

	if len(e.IOSourceList.File) == 0 {
		e.IOSourceList.File = "io.txt"
	}

	if len(e.IOSourceList.Http) == 0 {
		e.IOSourceList.Http = ":8080"
	}

	e.IOTargets[0].Set(e.IOTargets[0])
	e.originalIO.Set(e.IOTargets[0])
	e.ioTargetType = bf_io.Std

	if options.AttachDebugger {
		e.Debugger = debugger.NewDebugger(e.originalIO.Out)
	}

	if err != nil {
		e.originalIO.Err.Write([]byte(err.Error()))
		e.originalIO.Err.Write([]byte{10})
		os.Exit(1)
	}

	return e
}
