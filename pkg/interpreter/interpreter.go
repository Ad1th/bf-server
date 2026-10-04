package interpreter

import (
	"bytes"
	"io"
)

// Execute compiles and runs a Brainfuck source string with the given options.
func Execute(source string, opts ...Option) (string, *ExecutionStats, error) {
	prog, err := Compile(source)
	if err != nil {
		return "", nil, err
	}

	var outBuf bytes.Buffer
	allOpts := append([]Option{WithOutput(&outBuf)}, opts...)
	vm := NewVM(allOpts...)

	stats, err := vm.Run(prog)
	return outBuf.String(), stats, err
}

// ExecuteWithIO executes a compiled program with specified input reader and output writer.
func ExecuteWithIO(prog *Program, input io.Reader, output io.Writer, opts ...Option) (*ExecutionStats, error) {
	allOpts := append([]Option{WithInput(input), WithOutput(output)}, opts...)
	vm := NewVM(allOpts...)
	return vm.Run(prog)
}
