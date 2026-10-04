package interpreter

import (
	"strings"
	"unicode/utf8"
)

type rawToken struct {
	char rune
	line int
	col  int
}

type bracketPos struct {
	irIndex int
	line    int
	col     int
}

// Compile parses Brainfuck source code into an optimized Program.
// It checks for balanced brackets and returns a SyntaxError with line and column info if mismatched.
func Compile(source string) (*Program, error) {
	// First pass: extract tokens with accurate line & column positions
	var tokens []rawToken
	line := 1
	col := 1

	for len(source) > 0 {
		r, size := utf8.DecodeRuneInString(source)
		source = source[size:]

		if r == '\n' {
			line++
			col = 1
			continue
		}

		switch r {
		case '+', '-', '<', '>', '[', ']', '.', ',':
			tokens = append(tokens, rawToken{char: r, line: line, col: col})
		}
		col++
	}

	// Validate brackets and build IR
	var instrs []Instruction
	var stack []bracketPos

	i := 0
	numTokens := len(tokens)

	for i < numTokens {
		tok := tokens[i]

		switch tok.char {
		case '+':
			// Count consecutive '+' and '-'
			netAdd := 0
			count := uint64(0)
			for i < numTokens && (tokens[i].char == '+' || tokens[i].char == '-') {
				if tokens[i].char == '+' {
					netAdd++
				} else {
					netAdd--
				}
				count++
				i++
			}
			// Wrap to 8-bit range
			netAdd = ((netAdd % 256) + 256) % 256
			if netAdd != 0 {
				instrs = append(instrs, Instruction{
					Type:      OpAdd,
					Val:       netAdd,
					OrigSteps: count,
					Line:      tok.line,
					Col:       tok.col,
				})
			}
			continue

		case '-':
			netAdd := 0
			count := uint64(0)
			for i < numTokens && (tokens[i].char == '+' || tokens[i].char == '-') {
				if tokens[i].char == '+' {
					netAdd++
				} else {
					netAdd--
				}
				count++
				i++
			}
			netAdd = ((netAdd % 256) + 256) % 256
			if netAdd != 0 {
				instrs = append(instrs, Instruction{
					Type:      OpAdd,
					Val:       netAdd,
					OrigSteps: count,
					Line:      tok.line,
					Col:       tok.col,
				})
			}
			continue

		case '>':
			netMove := 0
			count := uint64(0)
			for i < numTokens && (tokens[i].char == '>' || tokens[i].char == '<') {
				if tokens[i].char == '>' {
					netMove++
				} else {
					netMove--
				}
				count++
				i++
			}
			if netMove != 0 {
				instrs = append(instrs, Instruction{
					Type:      OpMove,
					Val:       netMove,
					OrigSteps: count,
					Line:      tok.line,
					Col:       tok.col,
				})
			}
			continue

		case '<':
			netMove := 0
			count := uint64(0)
			for i < numTokens && (tokens[i].char == '>' || tokens[i].char == '<') {
				if tokens[i].char == '>' {
					netMove++
				} else {
					netMove--
				}
				count++
				i++
			}
			if netMove != 0 {
				instrs = append(instrs, Instruction{
					Type:      OpMove,
					Val:       netMove,
					OrigSteps: count,
					Line:      tok.line,
					Col:       tok.col,
				})
			}
			continue

		case '[':
			// Check for clear cell optimization: [-] or [+]
			if i+2 < numTokens && (tokens[i+1].char == '-' || tokens[i+1].char == '+') && tokens[i+2].char == ']' {
				instrs = append(instrs, Instruction{
					Type:      OpSet,
					Val:       0,
					OrigSteps: 3,
					Line:      tok.line,
					Col:       tok.col,
				})
				i += 3
				continue
			}

			// Regular loop start
			irIndex := len(instrs)
			stack = append(stack, bracketPos{irIndex: irIndex, line: tok.line, col: tok.col})
			instrs = append(instrs, Instruction{
				Type:      OpJumpIfZero,
				Target:    -1, // filled on matching ']'
				OrigSteps: 1,
				Line:      tok.line,
				Col:       tok.col,
			})
			i++

		case ']':
			if len(stack) == 0 {
				return nil, &SyntaxError{
					Message: "unexpected closing bracket ']' without matching '['",
					Line:    tok.line,
					Column:  tok.col,
				}
			}

			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			irIndex := len(instrs)
			// JZ jumps to instruction right after JNZ when zero
			instrs[open.irIndex].Target = irIndex + 1
			// JNZ jumps to instruction right after JZ when non-zero
			instrs = append(instrs, Instruction{
				Type:      OpJumpIfNotZero,
				Target:    open.irIndex + 1,
				OrigSteps: 1,
				Line:      tok.line,
				Col:       tok.col,
			})
			i++

		case '.':
			instrs = append(instrs, Instruction{
				Type:      OpOutput,
				OrigSteps: 1,
				Line:      tok.line,
				Col:       tok.col,
			})
			i++

		case ',':
			instrs = append(instrs, Instruction{
				Type:      OpInput,
				OrigSteps: 1,
				Line:      tok.line,
				Col:       tok.col,
			})
			i++
		}
	}

	if len(stack) > 0 {
		unclosed := stack[len(stack)-1]
		return nil, &SyntaxError{
			Message: "unclosed opening bracket '['",
			Line:    unclosed.line,
			Column:  unclosed.col,
		}
	}

	return &Program{
		Instructions: instrs,
		SourceLength: numTokens,
	}, nil
}

// Validate checks Brainfuck source code for syntax errors without compiling.
func Validate(source string) error {
	_, err := Compile(source)
	return err
}

// CleanSource returns a string containing only valid Brainfuck instructions.
func CleanSource(source string) string {
	var sb strings.Builder
	for _, r := range source {
		switch r {
		case '+', '-', '<', '>', '[', ']', '.', ',':
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
