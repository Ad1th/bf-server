package interpreter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestInterpreter_HelloWorld(t *testing.T) {
	// Standard Brainfuck Hello World program
	helloWorld := `
	++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]
	>>.
	>---.
	+++++++..+++.
	>>.
	<-.
	<.
	+++.------.--------.
	>>+.
	>++.
	`

	out, stats, err := Execute(helloWorld)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello World!\n"
	if out != expected {
		t.Fatalf("expected output %q, got %q", expected, out)
	}

	if stats.Steps == 0 {
		t.Fatalf("expected positive step count, got %d", stats.Steps)
	}
	if stats.MaxMemoryUsed <= 0 {
		t.Fatalf("expected positive memory used, got %d", stats.MaxMemoryUsed)
	}
}

func TestInterpreter_BasicInstructions(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    string
		expected string
	}{
		{
			name:     "single plus and dot",
			source:   "+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++.", // 65 = 'A'
			expected: "A",
		},
		{
			name:     "plus and minus wrap around",
			source:   "+-.",
			expected: "\x00",
		},
		{
			name:     "pointer movement",
			source:   ">+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++.<.",
			expected: "A\x00",
		},
		{
			name:     "cell zeroing optimization [-]",
			source:   "+++++[-]+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++.",
			expected: "A",
		},
		{
			name:     "input and output echo",
			source:   ",.,.,.",
			input:    "XYZ",
			expected: "XYZ",
		},
		{
			name:     "input loop until EOF",
			source:   ",[.,]",
			input:    "Brainfuck HTTP",
			expected: "Brainfuck HTTP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := Execute(tt.source, WithInput(strings.NewReader(tt.input)))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}

func TestInterpreter_Loops(t *testing.T) {
	t.Run("skip loop when cell is zero", func(t *testing.T) {
		source := "[+++++++++++++++++++++++++++++++++.]+"
		out, _, err := Execute(source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "" {
			t.Fatalf("expected empty output, got %q", out)
		}
	})

	t.Run("nested loops counter", func(t *testing.T) {
		// Multiplies 3 * 4 = 12, then adds 53 = 65 ('A')
		source := "+++[>++++[>+<-]<-]>>+++++++++++++++++++++++++++++++++++++++++++++++++++++."
		out, _, err := Execute(source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "A" {
			t.Fatalf("expected %q, got %q", "A", out)
		}
	})
}

func TestInterpreter_BracketMatchingErrors(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		errMessage string
		errLine    int
		errCol     int
	}{
		{
			name:       "unclosed opening bracket",
			source:     "++[\n++",
			errMessage: "unclosed opening bracket '['",
			errLine:    1,
			errCol:     3,
		},
		{
			name:       "unexpected closing bracket",
			source:     "++\n  ++]",
			errMessage: "unexpected closing bracket ']' without matching '['",
			errLine:    2,
			errCol:     5,
		},
		{
			name:       "nested unclosed bracket",
			source:     "++[++[++]\n++[\n",
			errMessage: "unclosed opening bracket '['",
			errLine:    2,
			errCol:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Compile(tt.source)
			if err == nil {
				t.Fatalf("expected compile error, got nil")
			}

			var syntaxErr *SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("expected *SyntaxError, got %T: %v", err, err)
			}

			if syntaxErr.Line != tt.errLine {
				t.Errorf("expected error line %d, got %d", tt.errLine, syntaxErr.Line)
			}
			if syntaxErr.Column != tt.errCol {
				t.Errorf("expected error col %d, got %d", tt.errCol, syntaxErr.Column)
			}
			if !strings.Contains(syntaxErr.Message, tt.errMessage) {
				t.Errorf("expected message to contain %q, got %q", tt.errMessage, syntaxErr.Message)
			}
		})
	}
}

func TestInterpreter_ExecutionLimits(t *testing.T) {
	t.Run("step limit exceeded", func(t *testing.T) {
		infiniteLoop := "+[]"
		_, stats, err := Execute(infiniteLoop, WithMaxSteps(100))
		if !errors.Is(err, ErrStepLimitExceeded) {
			t.Fatalf("expected ErrStepLimitExceeded, got %v", err)
		}
		if stats.Steps < 100 {
			t.Fatalf("expected steps >= 100, got %d", stats.Steps)
		}
	})

	t.Run("pointer left out of bounds", func(t *testing.T) {
		source := "<+"
		_, _, err := Execute(source)
		if !errors.Is(err, ErrPointerOutOfBounds) {
			t.Fatalf("expected ErrPointerOutOfBounds, got %v", err)
		}
	})

	t.Run("pointer right out of bounds", func(t *testing.T) {
		source := ">+"
		_, _, err := Execute(source, WithMemorySize(1))
		if !errors.Is(err, ErrPointerOutOfBounds) {
			t.Fatalf("expected ErrPointerOutOfBounds, got %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		infiniteLoop := "+[]"
		_, _, err := Execute(infiniteLoop, WithContext(ctx), WithMaxSteps(0))
		if !errors.Is(err, ErrExecutionCanceled) {
			t.Fatalf("expected ErrExecutionCanceled, got %v", err)
		}
	})
}

func TestInterpreter_CleanSource(t *testing.T) {
	commented := "Hello World! +++ [->+<] # comment .,"
	clean := CleanSource(commented)
	if clean != "+++[->+<].," {
		t.Fatalf("expected clean source %q, got %q", "+++[->+<].,", clean)
	}
}
