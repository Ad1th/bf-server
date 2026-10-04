package interpreter

import (
	"errors"
	"fmt"
)

var (
	// ErrStepLimitExceeded is returned when a program exceeds the maximum execution steps.
	ErrStepLimitExceeded = errors.New("execution step limit exceeded")

	// ErrPointerOutOfBounds is returned when the memory pointer moves beyond tape bounds.
	ErrPointerOutOfBounds = errors.New("memory pointer out of bounds")

	// ErrExecutionCanceled is returned when context is canceled during execution.
	ErrExecutionCanceled = errors.New("execution canceled by context")
)

// SyntaxError represents a bracket mismatch or parser syntax error with location.
type SyntaxError struct {
	Message string
	Line    int
	Column  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}
