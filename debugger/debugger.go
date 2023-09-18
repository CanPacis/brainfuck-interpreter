package debugger

import (
	"encoding/json"

	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Debugger struct {
	Exists bool
	Client Client
}

type ServerOperationType string

var (
	DiscloseMetaData   ServerOperationType = "disclose-meta-data"
	DiscloseDebugState ServerOperationType = "disclose-debug-state"
	ExitOperation      ServerOperationType = "exit"
	StdOutOperation    ServerOperationType = "std-out"
)

type MetaData struct {
	Operation ServerOperationType `json:"operation"`
	FileName  string              `json:"file_name"`
	FilePath  string              `json:"file_path"`
	Content   string              `json:"content"`
}

type State struct {
	Operation ServerOperationType `json:"operation"`
	Statement parser.Statement    `json:"statement"`
	Tape      []byte              `json:"tape"`
	Cursor    uint                `json:"cursor"`
}

type Exit struct {
	Operation ServerOperationType `json:"operation"`
}

type StdOut struct {
	Operation ServerOperationType `json:"operation"`
	Value     string              `json:"value"`
}

type ClientOperationType string

var (
	Resume   ClientOperationType = "resume"
	Step     ClientOperationType = "step"
	StepOut  ClientOperationType = "step-out"
	StepOver ClientOperationType = "step-over"
	Assign   ClientOperationType = "assign"
	Move     ClientOperationType = "move"
)

type ClientOperation struct {
	Operation ClientOperationType `json:"operation"`
}

type PlayerOperation struct {
	Operation ClientOperationType `json:"operation"`
}

type AssignOperation struct {
	Operation ClientOperationType `json:"operation"`
	Cell      uint                `json:"cell"`
	Value     byte                `json:"value"`
}

type MoveAction struct {
	Operation ClientOperationType `json:"operation"`
	Cell      uint                `json:"cell"`
}

func (d Debugger) Close() error {
	d.Client.WriteOperation(Exit{ExitOperation})
	return nil
}

func (d Debugger) ShareState(state State) (ClientOperation, error) {
	d.Client.WriteOperation(state)
	operation := ClientOperation{}

	response := make([]byte, 1024)
	n, err := d.Client.Read(response)
	response = response[:n]

	if err != nil {
		return operation, err
	}

	err = json.Unmarshal(response, &operation)

	if err != nil {
		return operation, err
	}

	return operation, nil
}

func NewDebugger() (Debugger, error) {
	client := &Client{"out"}

	return Debugger{
		Exists: true,
		Client: *client,
	}, nil
}
