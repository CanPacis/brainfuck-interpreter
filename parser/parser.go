package parser

import (
	"fmt"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/lexer"
)

type Statement struct {
	Type        string      `json:"type"`
	Value       uint32      `json:"value"`
	IOTarget    string      `json:"io_target"`
	Body        []Statement `json:"body"`
	DebugTarget bool
	lexer.Position
}

type Parser struct {
	FilePath string
	Program  []Statement
	Lexer    lexer.Lexer
}

func parse(tokens []lexer.Token) ([]Statement, int, lexer.Position, error) {
	statements := []Statement{}
	index := 0

	isDebug := false
	for ; index < len(tokens); index++ {
		token := tokens[index]

		switch token.Type {
		case "plus":
			statements = append(statements, Statement{Type: "Increment Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "minus":
			statements = append(statements, Statement{Type: "Decrement Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "move_right":
			statements = append(statements, Statement{Type: "Move Right Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "move_left":
			statements = append(statements, Statement{Type: "Move Left Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "dot":
			statements = append(statements, Statement{Type: "Stdout Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "comma":
			statements = append(statements, Statement{Type: "Stdin Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "star":
			statements = append(statements, Statement{Type: "Clear Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "debug":
			isDebug = true
		case "io":
			if index+3 > len(tokens) {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected end of file, expected io target")
			}

			spaceToken := tokens[index+1]

			if spaceToken.Type != "space" {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected %s token, expected whitespace", spaceToken.Type)
			}

			nextToken := tokens[index+2]

			if nextToken.Type != "keyword" {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected %s token, expected keyword", nextToken.Type)
			}

			statements = append(statements, Statement{Type: "Switch IO Statement", IOTarget: nextToken.Value, Position: token.Position})
			isDebug = false
		case "loop_open":
			if index+2 > len(tokens) {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected end of file, loop is unclose")
			}

			loopStatements, consumed, _, err := parse(tokens[index+1:])
			index += consumed

			if err != nil {
				return []Statement{}, 0, token.Position, err
			}
			statements = append(statements, Statement{Type: "Loop Statement", Body: loopStatements, Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "loop_close":
			isDebug = false
			return statements, index + 1, token.Position, nil
		}

	}

	return statements, index + 1, lexer.Position{}, nil
}

func (p *Parser) Parse(input string) bf_errors.RuntimeError {
	p.Lexer.Lex(input)

	statments, _, position, err := parse(p.Lexer.Tokens)

	if err != nil {
		return bf_errors.CreateSyntaxError(err, position, p.FilePath)
	}

	p.Program = statments
	return bf_errors.EmptyError
}

func NewParser(filePath string) Parser {
	return Parser{
		FilePath: filePath,
		Lexer: lexer.Lexer{
			CurrentPosition: lexer.Position{
				Line:   1,
				Column: 1,
			},
		},
	}
}
