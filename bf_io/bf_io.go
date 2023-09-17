package bf_io

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/CanPacis/brainfuck-interpreter/waiter"
)

type RuntimeIO struct {
	Out    io.Writer
	Err    io.Writer
	In     io.Reader
	Reader *bufio.Reader
	Writer *bufio.Writer
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
	io.Reader = bufio.NewReader(io.In)
	io.Writer = bufio.NewWriter(io.Out)

	return io
}

type IOSourceList struct {
	File string
	Http string
}

func FileIO(fileName string) (RuntimeIO, func() error, error) {
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

func HttpIO(port string, file_resource string, io_targets *[]RuntimeIO, waiters waiter.EngineWaiter) {
	var contentType string

	if path.Ext(file_resource) == ".json" {
		contentType = "application/json"
	} else {
		file, err := os.ReadFile(file_resource)

		if err != nil {
			contentType = http.DetectContentType(file)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(contentType) != 0 {
			w.Header().Set("content-type", contentType)
		}

		waiters.Add("write", 1)
		io := RuntimeIO{
			Out: w,
			Err: os.Stderr,
			In:  os.Stdin,
		}

		*io_targets = append(*io_targets, *io.Set(io))
		waiters.Done("http")
		waiters.Wait("write")
	})

	waiters.Add("program", 1)
	waiters.Add("http", 1)
	go http.ListenAndServe(port, nil)
}

type IOTargetType = string

var (
	Http IOTargetType = "http"
	Tcp  IOTargetType = "tcp"
	File IOTargetType = "file"
	Std  IOTargetType = "std"
)
