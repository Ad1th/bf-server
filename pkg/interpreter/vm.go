package interpreter

import (
	"bytes"
	"context"
	"io"
	"time"
)

const (
	// DefaultMemorySize is the classic standard Brainfuck tape size (30,000 cells).
	DefaultMemorySize = 30000

	// DefaultMaxSteps is the default execution step limit (10,000,000 operations).
	DefaultMaxSteps uint64 = 10000000
)

// VM represents a Brainfuck virtual machine instance.
type VM struct {
	MemorySize uint
	MaxSteps   uint64
	Input      io.Reader
	Output     io.Writer
	Context    context.Context
}

// ExecutionStats contains telemetry and resource usage data from a VM run.
type ExecutionStats struct {
	Steps          uint64
	MaxMemoryUsed  int
	Duration       time.Duration
	ProgramCounter int
	TapeSnapshot   []byte
}

// Option configures VM settings.
type Option func(*VM)

// WithMemorySize sets the tape memory size in bytes.
func WithMemorySize(size uint) Option {
	return func(vm *VM) {
		if size > 0 {
			vm.MemorySize = size
		}
	}
}

// WithMaxSteps sets the maximum allowed instruction steps. 0 means unlimited.
func WithMaxSteps(steps uint64) Option {
	return func(vm *VM) {
		vm.MaxSteps = steps
	}
}

// WithInput sets the input stream reader for ',' operations.
func WithInput(r io.Reader) Option {
	return func(vm *VM) {
		vm.Input = r
	}
}

// WithOutput sets the output stream writer for '.' operations.
func WithOutput(w io.Writer) Option {
	return func(vm *VM) {
		vm.Output = w
	}
}

// WithContext sets the execution context for cancellation and timeouts.
func WithContext(ctx context.Context) Option {
	return func(vm *VM) {
		vm.Context = ctx
	}
}

// NewVM creates a new VM configured with the provided options.
func NewVM(opts ...Option) *VM {
	vm := &VM{
		MemorySize: DefaultMemorySize,
		MaxSteps:   DefaultMaxSteps,
		Input:      bytes.NewReader(nil),
		Output:     io.Discard,
		Context:    context.Background(),
	}

	for _, opt := range opts {
		opt(vm)
	}

	return vm
}

// Run executes a compiled Program and returns execution stats or error.
func (vm *VM) Run(prog *Program) (*ExecutionStats, error) {
	if prog == nil || len(prog.Instructions) == 0 {
		return &ExecutionStats{}, nil
	}

	startTime := time.Now()
	tape := make([]byte, vm.MemorySize)
	ptr := 0
	pc := 0
	numInstrs := len(prog.Instructions)
	var steps uint64 = 0
	maxPtr := 0

	inBuf := make([]byte, 1)
	outBuf := make([]byte, 1)

	ctx := vm.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Optimization: check context every 1024 loop/jump operations
	ctxCheckCounter := 0

	for pc < numInstrs {
		instr := prog.Instructions[pc]

		// Step accounting
		steps += instr.OrigSteps
		if vm.MaxSteps > 0 && steps > vm.MaxSteps {
			return &ExecutionStats{
				Steps:          steps,
				MaxMemoryUsed:  maxPtr + 1,
				Duration:       time.Since(startTime),
				ProgramCounter: pc,
			}, ErrStepLimitExceeded
		}

		switch instr.Type {
		case OpAdd:
			tape[ptr] = byte((int(tape[ptr]) + instr.Val) % 256)
			pc++

		case OpMove:
			ptr += instr.Val
			if ptr < 0 || ptr >= int(vm.MemorySize) {
				return &ExecutionStats{
					Steps:          steps,
					MaxMemoryUsed:  maxPtr + 1,
					Duration:       time.Since(startTime),
					ProgramCounter: pc,
				}, ErrPointerOutOfBounds
			}
			if ptr > maxPtr {
				maxPtr = ptr
			}
			pc++

		case OpSet:
			tape[ptr] = byte(instr.Val)
			pc++

		case OpJumpIfZero:
			ctxCheckCounter++
			if ctxCheckCounter&0x3FF == 0 { // Check context every 1024 jumps
				select {
				case <-ctx.Done():
					return &ExecutionStats{
						Steps:          steps,
						MaxMemoryUsed:  maxPtr + 1,
						Duration:       time.Since(startTime),
						ProgramCounter: pc,
					}, ErrExecutionCanceled
				default:
				}
			}

			if tape[ptr] == 0 {
				pc = instr.Target
			} else {
				pc++
			}

		case OpJumpIfNotZero:
			ctxCheckCounter++
			if ctxCheckCounter&0x3FF == 0 {
				select {
				case <-ctx.Done():
					return &ExecutionStats{
						Steps:          steps,
						MaxMemoryUsed:  maxPtr + 1,
						Duration:       time.Since(startTime),
						ProgramCounter: pc,
					}, ErrExecutionCanceled
				default:
				}
			}

			if tape[ptr] != 0 {
				pc = instr.Target
			} else {
				pc++
			}

		case OpOutput:
			outBuf[0] = tape[ptr]
			if _, err := vm.Output.Write(outBuf); err != nil {
				return &ExecutionStats{
					Steps:          steps,
					MaxMemoryUsed:  maxPtr + 1,
					Duration:       time.Since(startTime),
					ProgramCounter: pc,
				}, err
			}
			pc++

		case OpInput:
			n, err := vm.Input.Read(inBuf)
			if err != nil || n == 0 {
				tape[ptr] = 0 // EOF: set current cell to 0
			} else {
				tape[ptr] = inBuf[0]
			}
			pc++

		default:
			pc++
		}
	}

	return &ExecutionStats{
		Steps:          steps,
		MaxMemoryUsed:  maxPtr + 1,
		Duration:       time.Since(startTime),
		ProgramCounter: pc,
		TapeSnapshot:   tape[:maxPtr+1],
	}, nil
}
