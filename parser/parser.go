package parser

import (
	"fmt"
	"os"
	"strconv"

	"github.com/CanPacis/brainfuck-interpreter/bf_errors"
	"github.com/CanPacis/brainfuck-interpreter/lexer"
)

type Statement struct {
	Type        string      `json:"type"`
	Value       uint32      `json:"value"`
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
		case "Plus":
			statements = append(statements, Statement{Type: "Increment Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "Minus":
			statements = append(statements, Statement{Type: "Decrement Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "MoveRight":
			statements = append(statements, Statement{Type: "Move Right Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "MoveLeft":
			statements = append(statements, Statement{Type: "Move Left Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "Dot":
			statements = append(statements, Statement{Type: "Stdout Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "Comma":
			statements = append(statements, Statement{Type: "Stdin Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "Clear":
			statements = append(statements, Statement{Type: "Clear Statement", Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "Debug":
			isDebug = true
		case "Bar":
			if index+2 > len(tokens) {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected end of file, expected number")
			}

			nextToken := tokens[index+1]

			if nextToken.Type != "Number" {
				return []Statement{}, 0, token.Position, fmt.Errorf("unexpected %s token, expected number", nextToken.Type)
			}

			value, err := strconv.Atoi(nextToken.Value)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			statements = append(statements, Statement{Type: "Push Statement", Value: uint32(value), Position: token.Position, DebugTarget: isDebug})
			isDebug = false
		case "LoopOpen":
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
		case "LoopClose":
			isDebug = false
			return statements, index + 1, token.Position, nil
		}

	}

	return statements, index + 1, lexer.Position{}, nil
}

func (p *Parser) Parse(input string) bf_errors.FileError {
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
