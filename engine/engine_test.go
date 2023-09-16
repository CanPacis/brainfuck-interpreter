package engine

import (
	"bytes"
	"testing"
)

func TestAdd(t *testing.T) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	r := NewEngine(EngineOptions{
		FilePath: "../bf/add.bf",
		Stdout:   &stdout,
		Stderr:   &stderr,
	})

	r.Run()

	expected := []byte("7")
	found := stdout.Bytes()
	if !bytes.Equal(expected, found) {
		t.Errorf("Incorrect stdout expected %s found %s", string(expected), string(found))
	}

	err := stderr.Bytes()

	if len(err) != 0 {
		t.Errorf("Unexpected error %s", string(err))
	}
}

func TestHelloWorld(t *testing.T) {
	stdout := bytes.Buffer{}

	r := NewEngine(EngineOptions{
		FilePath: "../bf/hello_world.bf",
		Stdout:   &stdout,
	})

	r.Run()

	expected := []byte("Hello World!\n")
	found := stdout.Bytes()
	if !bytes.Equal(expected, found) {
		t.Errorf("Incorrect stdout expected %s found %s", string(expected), string(found))
	}
}
