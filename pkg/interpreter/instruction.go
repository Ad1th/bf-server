package interpreter

import "fmt"

// OpType represents the type of virtual machine instruction.
type OpType int

const (
	// OpAdd adds Val to the current cell (with 8-bit wrap-around).
	OpAdd OpType = iota
	// OpMove shifts the data pointer by Val positions (positive = right, negative = left).
	OpMove
	// OpSet sets the current cell directly to Val (optimized from [-] or [+]).
	OpSet
	// OpJumpIfZero jumps to Target if current cell value is 0.
	OpJumpIfZero
	// OpJumpIfNotZero jumps to Target if current cell value is not 0.
	OpJumpIfNotZero
	// OpOutput writes current cell value to output writer.
	OpOutput
	// OpInput reads a single byte from input reader into current cell.
	OpInput
)

func (op OpType) String() string {
	switch op {
	case OpAdd:
		return "ADD"
	case OpMove:
		return "MOVE"
	case OpSet:
		return "SET"
	case OpJumpIfZero:
		return "JZ"
	case OpJumpIfNotZero:
		return "JNZ"
	case OpOutput:
		return "OUT"
	case OpInput:
		return "IN"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", op)
	}
}

// Instruction represents a compiled VM instruction.
type Instruction struct {
	Type      OpType
	Val       int    // Value operand (add amount, move offset, set value)
	Target    int    // Jump target index for JZ / JNZ
	OrigSteps uint64 // Number of elementary Brainfuck steps this represents
	Line      int    // Source line (1-indexed)
	Col       int    // Source column (1-indexed)
}

func (i Instruction) String() string {
	switch i.Type {
	case OpAdd:
		return fmt.Sprintf("ADD %d", i.Val)
	case OpMove:
		return fmt.Sprintf("MOVE %+d", i.Val)
	case OpSet:
		return fmt.Sprintf("SET %d", i.Val)
	case OpJumpIfZero:
		return fmt.Sprintf("JZ -> %d", i.Target)
	case OpJumpIfNotZero:
		return fmt.Sprintf("JNZ -> %d", i.Target)
	case OpOutput:
		return "OUT"
	case OpInput:
		return "IN"
	default:
		return "NOP"
	}
}

// Program is a compiled list of instructions ready for execution.
type Program struct {
	Instructions []Instruction
	SourceLength int
}
