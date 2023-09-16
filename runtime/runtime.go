package runtime

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/debugger"
	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Runtime struct {
	Path     string
	Name     string
	Content  string
	Debugger debugger.Debugger
	Parser   parser.Parser
	Tape     [30000]uint32
	Band     [30000]byte
	Cursor   uint
}

func run(r *Runtime, program []parser.Statement) bf_errors.FileError {
	index := 0

	for ; index < len(program); index++ {
		statement := program[index]

		if r.Debugger.Exists && statement.DebugTarget {
			action := r.Debugger.Wait(r.CreateDebugState(statement))
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
			r.Band[r.Cursor] = byte(statement.Value)
			r.Tape[r.Cursor] = statement.Value
		case "Increment Statement":
			r.Band[r.Cursor]++
			r.Tape[r.Cursor] = uint32(r.Band[r.Cursor])
		case "Decrement Statement":
			r.Band[r.Cursor]--
			r.Tape[r.Cursor] = uint32(r.Band[r.Cursor])
		case "Clear Statement":
			r.Tape = [30000]uint32{}
			r.Band = [30000]byte{}
		case "Move Right Statement":
			if r.Cursor < 30000 {
				r.Cursor++
			}
		case "Move Left Statement":
			if r.Cursor > 0 {
				r.Cursor--
			}
		case "Loop Statement":
			for r.Band[r.Cursor] != 0 {
				err := run(r, statement.Body)
				if err.Reason != nil {
					return err
				}
			}
		case "Stdout Statement":
			switch r.Tape[r.Cursor] {
			case 257:
				fileName := string(r.Band[:r.Cursor])
				file, err := os.ReadFile(fileName)

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, r.Path)
				}
				os.Stdout.Write(file)
			default:
				os.Stdout.Write([]byte{byte(r.Band[r.Cursor])})
			}
		case "Stdin Statement":
			switch r.Tape[r.Cursor] {
			default:
				reader := bufio.NewReader(os.Stdin)
				char, _, err := reader.ReadRune()

				if err != nil {
					return bf_errors.CreateUncaughtError(err, statement.Position, r.Path)
				}

				r.Band[r.Cursor] = byte(char)
				r.Tape[r.Cursor] = uint32(r.Band[r.Cursor])
			}
		}
	}

	return bf_errors.EmptyError
}

func (r *Runtime) Run() {
	if r.Debugger.Exists {
		r.Debugger.Open(debugger.DebugMetaData{
			FileName: r.Name,
			FilePath: r.Path,
			Content:  r.Content,
		})
	}

	err := r.Parser.Parse(r.Content)
	if err.Reason != nil {
		fmt.Println(err)
		if r.Debugger.Exists {
			r.Debugger.Error(err)
			r.Debugger.Close()
		}
		os.Exit(1)
	}

	err = run(r, r.Parser.Program)
	if err.Reason != nil {
		fmt.Println(err)
		if r.Debugger.Exists {
			r.Debugger.Error(err)
			r.Debugger.Close()
		}
		os.Exit(1)
	}

	if r.Debugger.Exists {
		r.Debugger.Close()
	}
	os.Stdout.Write([]byte{10})
}

func (r *Runtime) CreateDebugState(statement parser.Statement) debugger.DebugState {
	return debugger.DebugState{
		Statement: statement,
		Cursor:    r.Cursor,
		Tape:      r.Tape[:100],
		Band:      r.Band[:100],
	}
}

type RuntimeOptions struct {
	FilePath string
	Debugger debugger.Debugger
}

func NewRuntime(options RuntimeOptions) *Runtime {
	name := path.Base(options.FilePath)
	content, err := os.ReadFile(options.FilePath)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return &Runtime{
		Name:     name,
		Path:     options.FilePath,
		Content:  string(content),
		Debugger: options.Debugger,
		Parser:   parser.NewParser(options.FilePath),
	}
}
