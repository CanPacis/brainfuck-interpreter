package lexer

type Position struct {
	Line   uint `json:"line"`
	Column uint `json:"column"`
}

type Token struct {
	Type     string
	Value    string
	Position Position
}

type Lexer struct {
	Tokens          []Token
	CurrentPosition Position
}

func (l *Lexer) CreateToken(t, value string) Token {
	token := Token{
		Type:  t,
		Value: value,
		Position: Position{
			Line:   l.CurrentPosition.Line,
			Column: l.CurrentPosition.Column,
		},
	}

	if t == "NewLine" {
		l.CurrentPosition.Line++
		l.CurrentPosition.Column = 1
	} else {
		l.CurrentPosition.Column += uint(len(value))
	}

	return token
}

func (l *Lexer) Lex(input string) {
	index := 0

	for ; index < len(input); index++ {
		char := input[index]

		switch char {
		case '+':
			l.Tokens = append(l.Tokens, l.CreateToken("Plus", "+"))
		case '-':
			l.Tokens = append(l.Tokens, l.CreateToken("Minus", "-"))
		case '.':
			l.Tokens = append(l.Tokens, l.CreateToken("Dot", "."))
		case ',':
			l.Tokens = append(l.Tokens, l.CreateToken("Comma", ","))
		case '>':
			l.Tokens = append(l.Tokens, l.CreateToken("MoveRight", ">"))
		case '<':
			l.Tokens = append(l.Tokens, l.CreateToken("MoveLeft", "<"))
		case '[':
			l.Tokens = append(l.Tokens, l.CreateToken("LoopOpen", "["))
		case ']':
			l.Tokens = append(l.Tokens, l.CreateToken("LoopClose", "]"))
		case '|':
			l.Tokens = append(l.Tokens, l.CreateToken("Bar", "|"))
		case '*':
			l.Tokens = append(l.Tokens, l.CreateToken("Clear", "*"))
		case 10:
			l.CreateToken("NewLine", "\n")
		case '\\':
			l.CreateToken("Escape", "\\\\")
			index++
		default:
			if char < 58 && char > 47 {
				consumed := l.LexNumber(input[index:])
				index += consumed
			} else if char == 'd' {
				consumed := l.LexDebug(input[index:])
				index += consumed
			} else {
				l.CurrentPosition.Column++
			}
		}

	}
}

func (l *Lexer) LexDebug(input string) int {
	if input[:5] == "debug" {
		l.Tokens = append(l.Tokens, l.CreateToken("Debug", "debug"))
		return 4
	} else {
		l.CurrentPosition.Column++
		return 0
	}
}

func (l *Lexer) LexNumber(input string) int {
	index := 0
	consumed := 0
	number := ""

	for index < len(input) {
		char := input[index]

		index++
		consumed++
		if char < 58 && char > 47 {
			number += string(char)
		} else {
			consumed--
			break
		}
	}

	l.Tokens = append(l.Tokens, l.CreateToken("Number", number))

	return consumed - 1
}
