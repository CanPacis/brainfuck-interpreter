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
	Io           RuntimeIO
	IOSourceList IOSourceList
	originalIo   RuntimeIO
	disposers    []func()
	wg           *sync.WaitGroup
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

func http_io(port string, io *RuntimeIO, wg *sync.WaitGroup) {
	wg.Add(1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.Set(RuntimeIO{
			Out: w,
			Err: os.Stderr,
			In:  os.Stdin,
		})
		wg.Done()
	})

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
			// os.Stdout.Write([]byte{byte(e.Tape[e.Cursor])})
			e.Io.Out.Write([]byte{byte(e.Tape[e.Cursor])})
		case "Stdin Statement":
			byte, err := e.Io.reader.ReadByte()
			if err != nil && err != io.EOF {
				return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
			}
			e.Tape[e.Cursor] = byte
		case "Switch IO Statement":
			switch statement.IoTarget {
			case "std":
				e.Io.Set(e.originalIo)
			case "http":
				http_io(e.IOSourceList.Http, &e.Io, e.wg)
				e.wg.Wait()
				e.Io.Out.Write([]byte("TEST"))
			case "tcp":
				e.Io.Out.Write([]byte("tcp"))
			case "file":
				io, close, err := file_io(e.IOSourceList.File)

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
				}

				e.disposers = append(e.disposers, func() {
					close()
				})

				e.Io.Set(io)
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
			e.Debugger.Error(err)
		}
		e.Debugger.Close()
	} else {
		if err.Reason != nil {
			err.Write(e.originalIo.Err)
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
}

func (io *RuntimeIO) Set(value RuntimeIO) {
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
		Name:         name,
		Path:         options.FilePath,
		Content:      string(content),
		Parser:       parser.NewParser(options.FilePath),
		Io:           std,
		wg:           &sync.WaitGroup{},
		IOSourceList: options.IOSourceList,
	}

	if len(e.IOSourceList.File) == 0 {
		e.IOSourceList.File = "io.txt"
	}

	if len(e.IOSourceList.Http) == 0 {
		e.IOSourceList.Http = ":8080"
	}

	e.Io.Set(e.Io)
	e.originalIo.Set(e.Io)

	if options.AttachDebugger {
		e.Debugger = debugger.NewDebugger(e.Io.Out)
	}

	if err != nil {
		e.originalIo.Err.Write([]byte(err.Error()))
		os.Exit(1)
	}

	return e
}
