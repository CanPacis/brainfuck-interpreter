package bf_errors

import (
	"fmt"
	"io"
	"path"

	"github.com/CanPacis/brainfuck-interpreter/lexer"
)

const (
	UncaughtError = iota
	SyntaxError
	StackOverflowError
	StackUnderflowError
)

type RuntimeError struct {
	Type     int    `json:"type"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Reason   error  `json:"error"`
	Position lexer.Position
}

func CreateError(err error, position lexer.Position, typ int, filePath string) RuntimeError {
	fileName := path.Base(filePath)

	return RuntimeError{
		Reason:   err,
		Type:     typ,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

func CreateSyntaxError(reason error, position lexer.Position, filePath string) RuntimeError {
	fileName := path.Base(filePath)

	return RuntimeError{
		Type:     SyntaxError,
		Reason:   reason,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

func CreateUncaughtError(reason error, position lexer.Position, filePath string) RuntimeError {
	fileName := path.Base(filePath)

	return RuntimeError{
		Type:     UncaughtError,
		Reason:   reason,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

var EmptyError = RuntimeError{
	Reason:   nil,
	Position: lexer.Position{},
}

func (err RuntimeError) String() string {
	result := ""

	switch err.Type {
	case UncaughtError:
		result += "Program threw an error:\n"
	case SyntaxError:
		result += "There is a syntax error:\n"
	case StackOverflowError:
		result += "Stack overflow:"
	case StackUnderflowError:
		result += "Stack underflow:"
	}

	result += fmt.Sprintf("\t'%s' at line %d column %d in %s\n", err.Reason.Error(), err.Position.Line, err.Position.Column, err.FileName)
	result += fmt.Sprintf("\t%s %d:%d\n", err.FilePath, err.Position.Line, err.Position.Column)

	return result
}

func (err RuntimeError) Write(w io.Writer) {
	w.Write([]byte(err.String()))
}
