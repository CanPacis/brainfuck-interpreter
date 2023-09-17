package engine

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/debugger"
	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Engine struct {
	Path         string
	Name         string
	Content      string
	Debugger     debugger.Debugger
	Parser       parser.Parser
	Tape         [30000]byte
	Cursor       uint
	IOTargets    []RuntimeIO
	IOSourceList IOSourceList
	ioTargetType string
	originalIO   RuntimeIO
	disposers    []func()
	waiters      map[string]*sync.WaitGroup
}

func file_io(fileName string) (RuntimeIO, func() error, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	io := RuntimeIO{}

	if err != nil {
		return io, nil, err
	}

	io.Out = file
	io.Err = file
	io.In = file

	return io, file.Close, nil
}

func http_io(port string, e *Engine) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		e.waiters["write"].Add(1)
		io := RuntimeIO{
			Out: w,
			Err: os.Stderr,
			In:  os.Stdin,
		}

		e.IOTargets = append(e.IOTargets, *io.Set(io))
		e.waiters["http"].Done()

		e.waiters["write"].Wait()
	})

	e.waiters["program"].Add(1)
	e.waiters["http"].Add(1)
	go http.ListenAndServe(port, nil)
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
			if e.ioTargetType == "http" && len(e.IOTargets) == 0 {
				e.waiters["http"].Wait()
			}
			for _, target := range e.IOTargets {
				if e.ioTargetType == "http" {
					if e.Tape[e.Cursor] != 0 {
						target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
					}
				} else {
					target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
				}
			}
			if e.ioTargetType == "http" && e.Tape[e.Cursor] == 0 {
				e.IOTargets = []RuntimeIO{}
				e.waiters["http"].Add(1)
				e.waiters["write"].Done()
			}
		case "Stdin Statement":
			target := e.IOTargets[0]
			byte, err := target.reader.ReadByte()
			if err != nil && err != io.EOF {
				return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
			}
			e.Tape[e.Cursor] = byte
		case "Switch IO Statement":
			e.ioTargetType = statement.IOTarget
			switch statement.IOTarget {
			case "std":
				e.IOTargets = []RuntimeIO{e.originalIO}
			case "http":
				e.IOTargets = []RuntimeIO{}
				http_io(e.IOSourceList.Http, e)
			case "tcp":
				e.originalIO.Out.Write([]byte("tcp"))
			case "file":
				io, close, err := file_io(e.IOSourceList.File)

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
				}

				e.disposers = append(e.disposers, func() {
					close()
				})

				io.Set(io)
				e.IOTargets = []RuntimeIO{io}
			}
		}
	}

	return bf_errors.EmptyError
}

func (e *Engine) dispose(err bf_errors.FileError) {
	for _, f := range e.disposers {
		f()
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

	e.waiters["program"].Wait()
	e.dispose(bf_errors.EmptyError)
}

func (e *Engine) CreateDebugState(statement parser.Statement) debugger.DebugState {
	return debugger.DebugState{
		Statement: statement,
		Cursor:    e.Cursor,
		Tape:      e.Tape[:100],
	}
}

type RuntimeIO struct {
	Out    io.Writer
	Err    io.Writer
	In     io.Reader
	reader *bufio.Reader
	writer *bufio.Writer
}

func (io *RuntimeIO) Set(value RuntimeIO) *RuntimeIO {
	if value.Out == nil {
		io.Out = os.Stdout
	} else {
		io.Out = value.Out
	}
	if value.Err == nil {
		io.Err = os.Stderr
	} else {
		io.Err = value.Err
	}
	if value.In == nil {
		io.In = os.Stdin
	} else {
		io.In = value.In
	}
	io.reader = bufio.NewReader(io.In)
	io.writer = bufio.NewWriter(io.Out)

	return io
}

type IOSourceList struct {
	File string
	Http string
}

type EngineOptions struct {
	FilePath       string
	AttachDebugger bool
	Stdout         io.Writer
	Stderr         io.Writer
	Stdin          io.Reader
	IOSourceList   IOSourceList
}

func NewEngine(options EngineOptions) *Engine {
	name := path.Base(options.FilePath)
	content, err := os.ReadFile(options.FilePath)

	std := RuntimeIO{
		Out: options.Stdout,
		In:  options.Stdin,
		Err: options.Stderr,
	}

	e := &Engine{
		Name:      name,
		Path:      options.FilePath,
		Content:   string(content),
		Parser:    parser.NewParser(options.FilePath),
		IOTargets: []RuntimeIO{std},
		waiters: map[string]*sync.WaitGroup{
			"program": {},
			"http":    {},
			"write":   {},
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
	e.ioTargetType = "std"

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
