package debugger

import (
	"bufio"
	"encoding/json"
	"os"
)

type Client struct {
	target string
}

func (c *Client) Write(p []byte) (int, error) {
	return c.WriteOperation(StdOut{
		Operation: StdOutOperation,
		Value:     string(p),
	})
}

func (c *Client) WriteOperation(data interface{}) (int, error) {
	encoded, err := json.Marshal(data)

	if err != nil {
		return 0, err
	}

	encoded = append(encoded, 10)

	return os.Stdout.Write(encoded)
}

func (c *Client) Read(p []byte) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()

	if err != nil {
		return 0, err
	}

	copy(p, line)

	return len(line), nil
}
