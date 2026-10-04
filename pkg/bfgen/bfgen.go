package bfgen

import (
	"strings"
)

// Tape provides an emitter for generating verified Brainfuck code.
type Tape struct {
	Pos  int
	code strings.Builder
}

// Code returns the accumulated Brainfuck code.
func (t *Tape) Code() string {
	return t.code.String()
}

// MoveTo shifts the pointer to the target cell index.
func (t *Tape) MoveTo(target int) {
	diff := target - t.Pos
	if diff > 0 {
		t.code.WriteString(strings.Repeat(">", diff))
	} else if diff < 0 {
		t.code.WriteString(strings.Repeat("<", -diff))
	}
	t.Pos = target
}

// Clear zeroes out the cell at the given index.
func (t *Tape) Clear(cell int) {
	t.MoveTo(cell)
	t.code.WriteString("[-]")
}

// Set sets the cell to an exact positive byte value.
func (t *Tape) Set(cell int, val int) {
	t.Clear(cell)
	if val > 0 {
		t.code.WriteString(strings.Repeat("+", val))
	}
}

// Add adds (or subtracts if negative) n from the cell.
func (t *Tape) Add(cell int, n int) {
	t.MoveTo(cell)
	if n > 0 {
		t.code.WriteString(strings.Repeat("+", n))
	} else if n < 0 {
		t.code.WriteString(strings.Repeat("-", -n))
	}
}

// Read reads a byte into the cell from input stream.
func (t *Tape) Read(cell int) {
	t.MoveTo(cell)
	t.code.WriteString(",")
}

// Copy copies the value of src into dst, using tmp as a non-destructive intermediate.
func (t *Tape) Copy(src, dst, tmp int) {
	t.Clear(dst)
	t.Clear(tmp)
	t.MoveTo(src)
	t.code.WriteString("[")
	t.MoveTo(dst)
	t.code.WriteString("+")
	t.MoveTo(tmp)
	t.code.WriteString("+")
	t.MoveTo(src)
	t.code.WriteString("-]")
	// Restore src from tmp
	t.MoveTo(tmp)
	t.code.WriteString("[")
	t.MoveTo(src)
	t.code.WriteString("+")
	t.MoveTo(tmp)
	t.code.WriteString("-]")
	t.MoveTo(src)
}

// EmitString generates optimized Brainfuck code to print a string.
func (t *Tape) EmitString(s string, printCell, loopCell int) {
	t.Clear(printCell)
	t.Clear(loopCell)
	cur := 0
	for _, b := range []byte(s) {
		val := int(b)
		diff := val - cur
		if diff >= -5 && diff <= 5 {
			t.Add(printCell, diff)
		} else {
			bestA, bestB, bestC := 1, val, 0
			bestCost := val
			for a := 3; a <= 15; a++ {
				bVal := val / a
				c := val % a
				cost := a + bVal + c + 8
				if cost < bestCost {
					bestCost = cost
					bestA, bestB, bestC = a, bVal, c
				}
			}
			t.Clear(printCell)
			t.Clear(loopCell)
			t.Add(loopCell, bestA)
			t.MoveTo(loopCell)
			t.code.WriteString("[")
			t.MoveTo(printCell)
			t.code.WriteString(strings.Repeat("+", bestB))
			t.MoveTo(loopCell)
			t.code.WriteString("-]")
			if bestC > 0 {
				t.Add(printCell, bestC)
			}
		}
		t.MoveTo(printCell)
		t.code.WriteString(".")
		cur = val
	}
	t.Clear(printCell)
	t.Clear(loopCell)
}

// TextToBrainfuck generates Brainfuck source code that outputs the given text.
func TextToBrainfuck(text string) string {
	t := &Tape{}
	t.EmitString(text, 0, 1)
	return t.Code() + "\n"
}

// BuildNativeRouterProgram generates a Brainfuck program that inspects HTTP request line on stdin
// and routes GET / vs GET /hello vs 404 natively in Brainfuck.
func BuildNativeRouterProgram(rootMsg, helloMsg, notFoundMsg string) string {
	t := &Tape{}
	// Memory layout:
	// 0: input character after "GET /"
	// 1: is_root flag (1 if '/')
	// 2: is_hello flag (1 if '/hello')
	// 3: is_404 flag (1 if unmatched)
	// 4: cmp1 scratch
	// 5: cmp2 scratch
	// 6: tmp scratch
	// 7: printCell
	// 8: loopCell

	// Read past "GET /" (5 bytes: 'G', 'E', 'T', ' ', '/')
	for i := 0; i < 5; i++ {
		t.Read(0)
	}
	t.Clear(0)

	// Read character immediately after '/'
	t.Read(0)

	// Assume is_404 = 1 initially
	t.Set(3, 1)

	// Copy input char to cmp1 and cmp2
	t.Copy(0, 4, 6)
	t.Copy(0, 5, 6)

	// Test if cmp1 == 32 (' ' space, meaning exact "/")
	t.Add(4, -32)
	t.Set(1, 1) // assume is_root = 1
	t.MoveTo(4)
	t.code.WriteString("[")
	t.Set(1, 0)
	t.Clear(4)
	t.code.WriteString("]")

	// Test if cmp2 == 104 ('h', meaning starts with "/h")
	t.Add(5, -104)
	t.Set(2, 1) // assume is_hello = 1
	t.MoveTo(5)
	t.code.WriteString("[")
	t.Set(2, 0)
	t.Clear(5)
	t.code.WriteString("]")

	// If is_root (cell 1) is 1, clear is_404 (cell 3)
	t.Copy(1, 6, 0)
	t.MoveTo(6)
	t.code.WriteString("[")
	t.Set(3, 0)
	t.Clear(6)
	t.code.WriteString("]")

	// If is_hello (cell 2) is 1, clear is_404 (cell 3)
	t.Copy(2, 6, 0)
	t.MoveTo(6)
	t.code.WriteString("[")
	t.Set(3, 0)
	t.Clear(6)
	t.code.WriteString("]")

	// Route 1: Root Handler
	t.MoveTo(1)
	t.code.WriteString("[")
	t.EmitString(rootMsg, 7, 8)
	t.Clear(1)
	t.code.WriteString("]")

	// Route 2: Hello Handler
	t.MoveTo(2)
	t.code.WriteString("[")
	t.EmitString(helloMsg, 7, 8)
	t.Clear(2)
	t.code.WriteString("]")

	// Route 3: 404 Fallback Handler
	t.MoveTo(3)
	t.code.WriteString("[")
	t.EmitString(notFoundMsg, 7, 8)
	t.Clear(3)
	t.code.WriteString("]")

	return t.Code() + "\n"
}
