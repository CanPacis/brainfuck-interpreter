package debugger

import (
	"encoding/json"

	"github.com/CanPacis/brainfuck-interpreter/parser"
)

type Debugger struct {
	Exists      bool
	Client      Client
	ErrorClient ErrorClient
}

type ServerAction string

var (
	DiscloseMetaData   ServerAction = "disclose-meta-data"
	DiscloseDebugState ServerAction = "disclose-debug-state"
	ExitAction         ServerAction = "exit"
	StdOutAction       ServerAction = "std-out"
	StdErrAction       ServerAction = "std-err"
)

type MetaData struct {
	Operation ServerAction `json:"operation"`
	FileName  string       `json:"file_name"`
	FilePath  string       `json:"file_path"`
	Content   string       `json:"content"`
}

type State struct {
	Operation ServerAction     `json:"operation"`
	Statement parser.Statement `json:"statement"`
	Tape      []byte           `json:"tape"`
	Cursor    uint             `json:"cursor"`
}

type Exit struct {
	Operation ServerAction `json:"operation"`
	Code      int          `json:"code"`
}

type StdOut struct {
	Operation ServerAction `json:"operation"`
	Value     string       `json:"value"`
}

type ClientAction string

var (
	Resume   ClientAction = "resume"
	Step     ClientAction = "step"
	StepOut  ClientAction = "step-out"
	StepOver ClientAction = "step-over"
	Assign   ClientAction = "assign"
	Move     ClientAction = "move"
)

type ClientOperation struct {
	Operation ClientAction `json:"operation"`
}

type PlayerOperation struct {
	Operation ClientAction `json:"operation"`
}

type AssignOperation struct {
	Operation ClientAction `json:"operation"`
	Cell      uint         `json:"cell"`
	Value     byte         `json:"value"`
}

type MoveOperation struct {
	Operation ClientAction `json:"operation"`
	Cell      uint         `json:"cell"`
}

func (d Debugger) Close(code int) error {
	d.Client.WriteOperation(Exit{Operation: ExitAction, Code: code})
	return nil
}

func (d Debugger) ShareState(state State) (string, interface{}, error) {
	d.Client.WriteOperation(state)

	response := make([]byte, 1024)
	n, err := d.Client.Read(response)
	response = response[:n]

	if err != nil {
		return "", nil, err
	}

	var action map[string]interface{}
	err = json.Unmarshal(response, &action)

	if err != nil {
		return "", action, err
	}

	switch action["operation"] {
	case "resume", "step", "step-out", "step-over":
		return action["operation"].(string), PlayerOperation{Operation: ClientAction(action["operation"].(string))}, nil
	case "move":
		return "move", MoveOperation{Operation: "move", Cell: uint(action["cell"].(float64))}, nil
	case "assign":
		return "assign", AssignOperation{Operation: "assign", Cell: uint(action["cell"].(float64)), Value: byte(action["value"].(float64))}, nil
	}

	return "", action, nil
}

func NewDebugger() (Debugger, error) {
	client := &Client{}

	return Debugger{
		Exists:      true,
		Client:      *client,
		ErrorClient: ErrorClient{client},
	}, nil
}
