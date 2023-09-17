package engine

import (
	"fmt"
	"io"
	"time"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/bf_io"
	"github.com/CanPacis/brainfuck-interpreter/parser"
	"github.com/CanPacis/brainfuck-interpreter/waiter"
)

func (e *Engine) r_increment_s() {
	e.Tape[e.Cursor]++
}

func (e *Engine) r_decrement_s() {
	e.Tape[e.Cursor]--
}

func (e *Engine) r_clear_s() {
	e.Tape = [30000]byte{}
}

func (e *Engine) r_move_right_s(statement parser.Statement) bf_errors.RuntimeError {
	if e.Cursor < 30000 {
		e.Cursor++

		return bf_errors.EmptyError
	}

	return bf_errors.CreateError(fmt.Errorf("stack overflow"), statement.Position, bf_errors.StackOverflowError, e.Path)
}

func (e *Engine) r_move_left_s(statement parser.Statement) bf_errors.RuntimeError {
	if e.Cursor > 0 {
		e.Cursor--

		return bf_errors.EmptyError
	}

	return bf_errors.CreateError(fmt.Errorf("stack overflow"), statement.Position, bf_errors.StackUnderflowError, e.Path)
}

func (e *Engine) r_loop_s(statement parser.Statement) bf_errors.RuntimeError {
	for e.Tape[e.Cursor] != 0 {
		err := run(e, statement.Body)
		if err.Reason != nil {
			return err
		}
	}

	return bf_errors.EmptyError
}

func (e *Engine) r_stdout_s(statement parser.Statement) bf_errors.RuntimeError {
	if e.ioTargetType == bf_io.Http && len(e.IOTargets) == 0 {
		e.waiters.Wait(waiter.HttpConnection)
	}
	for _, target := range e.IOTargets {
		if e.ioTargetType == bf_io.Http {
			if e.Tape[e.Cursor] != 0 {
				_, err := target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
				return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
			}
		} else {
			_, err := target.Out.Write([]byte{byte(e.Tape[e.Cursor])})
			return bf_errors.CreateUncaughtError(err, statement.Position, e.Path)
		}
	}
	if e.ioTargetType == bf_io.Http && e.Tape[e.Cursor] == 0 {
		e.IOTargets = []bf_io.RuntimeIO{}
		e.waiters.Add(waiter.HttpConnection, 1)
		e.waiters.Done(waiter.Write)
	}

	return bf_errors.EmptyError
}

func (e *Engine) r_stdin_s(statement parser.Statement) bf_errors.RuntimeError {
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

	return bf_errors.EmptyError
}

func (e *Engine) r_switch_io_s(statement parser.Statement) bf_errors.RuntimeError {
	// while swtiching io methods, sometimes http server may not spin up, this waits for it
	// I know I need to solve this
	time.Sleep(time.Millisecond)
	if e.ioTargetType == bf_io.Http {
		e.waiters.Done(waiter.Program)
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

		e.IOTargets = []bf_io.RuntimeIO{*io.Set(io)}
	}

	return bf_errors.EmptyError
}
