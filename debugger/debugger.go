package debugger

import (
	"bufio"
	"encoding/json"
	"io"
	"net"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/bf_io"
	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Debugger struct {
	Exists   bool
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	listener net.Listener
	// clients  DebugClient
}

type MetaData struct {
	Type     string `json:"type"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

type State struct {
	Type      string           `json:"type"`
	Statement parser.Statement `json:"statement"`
	Tape      []byte           `json:"tape"`
	Cursor    uint             `json:"cursor"`
}

type IOData struct {
	Type    string             `json:"type"`
	Content []byte             `json:"content"`
	Target  bf_io.IOTargetType `json:"target"`
	Channel string             `json:"channel"`
}

// type DebugClient struct {
// 	connection net.Conn
// }

type Action struct {
	Operation string `json:"operation"`
}

// func (c DebugClient) Send(data interface{}) error {
// 	encoded, err := json.Marshal(data)

// 	if err != nil {
// 		return err
// 	}

// 	size := strconv.Itoa(len(encoded))
// 	sizeBuffer := bytes.Buffer{}
// 	sizeBytes := []byte(size)

// 	for i := 0; i < 10-len(sizeBytes); i++ {
// 		binary.Write(&sizeBuffer, binary.LittleEndian, []byte{0})
// 	}
// 	binary.Write(&sizeBuffer, binary.LittleEndian, sizeBytes)
// 	c.connection.Write(sizeBuffer.Bytes())
// 	c.connection.Write(encoded)

// 	return nil
// }

// func (c DebugClient) Receive() (Action, error) {
// 	buffer := make([]byte, 1024)
// 	n, err := c.connection.Read(buffer)
// 	action := Action{}

// 	if err != nil {
// 		return action, err
// 	}

// 	buffer = buffer[:n]

// 	err = json.Unmarshal(buffer, &action)

// 	return action, err
// }

func (d Debugger) Error(err bf_errors.RuntimeError) {
	// d.client.Send(err)
}

func (d Debugger) Share(data interface{}) Action {
	encoded, _ := json.Marshal(data)
	d.Writer.Write(encoded)
	return Action{}
}

// func (d Debugger) Wait(state State) DebugAction {
// 	err := d.client.Send(state)

// 	if err != nil {
// 		panic(err)
// 	}

// 	action, err := d.client.Receive()

// 	if err != nil {
// 		panic(err)
// 	}

// 	return action
// }

// func (d Debugger) Open(data MetaData) {
// d.client.Send(data)
// }

func (d Debugger) Close() {
	// d.client.Send(map[string]bool{"exit": true})
	// d.client.connection.Close()
	// d.listener.Close()
}

type Server struct {
	Host string
	Port string
	Type string
}

func NewDebugger(w io.Writer, r io.Reader) Debugger {
	// server := Server{
	// 	Host: "127.0.0.1",
	// 	Port: "0",
	// 	Type: "tcp",
	// }
	// listener, err := net.Listen(server.Type, server.Host+":"+server.Port)

	// if err != nil {
	// 	panic(err)
	// }

	// port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	// w.Write([]byte(port))
	// connection, err := listener.Accept()

	// if err != nil {
	// 	panic(err)
	// }

	return Debugger{
		Exists: true,
		Writer: bufio.NewWriter(w),
		Reader: bufio.NewReader(r),
		// listener: listener,
		// client:   DebugClient{connection},
	}
}
