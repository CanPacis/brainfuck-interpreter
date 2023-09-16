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
	StackOverflow
)

type FileError struct {
	Type     int    `json:"type"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Reason   error  `json:"error"`
	Position lexer.Position
}

func CreateError(err error, position lexer.Position, filePath string) FileError {
	fileName := path.Base(filePath)

	return FileError{
		Reason:   err,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

func CreateSyntaxError(reason error, position lexer.Position, filePath string) FileError {
	fileName := path.Base(filePath)

	return FileError{
		Type:     SyntaxError,
		Reason:   reason,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

func CreateUncaughtError(reason error, position lexer.Position, filePath string) FileError {
	fileName := path.Base(filePath)

	return FileError{
		Type:     UncaughtError,
		Reason:   reason,
		Position: position,
		FileName: fileName,
		FilePath: filePath,
	}
}

var EmptyError = FileError{
	Reason:   nil,
	Position: lexer.Position{},
}

func (err FileError) String() string {
	result := ""

	switch err.Type {
	case UncaughtError:
		result += "Program threw an error:\n"
	case SyntaxError:
		result += "There is a syntax error:\n"
	}

	result += fmt.Sprintf("\t'%s' at line %d column %d in %s\n", err.Reason.Error(), err.Position.Line, err.Position.Column, err.FileName)
	result += fmt.Sprintf("\t%s %d:%d\n", err.FilePath, err.Position.Line, err.Position.Column)

	return result
}

func (err FileError) Write(w io.Writer) {
	w.Write([]byte(err.String()))
}
