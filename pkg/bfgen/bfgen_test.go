package bfgen

import (
	"strings"
	"testing"

	"github.com/madith/bf-server/pkg/interpreter"
)

func TestTextToBrainfuck(t *testing.T) {
	expected := "Hello, Brainfuck!\n"
	bfCode := TextToBrainfuck(expected)

	out, _, err := interpreter.Execute(bfCode)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	if out != expected {
		t.Fatalf("expected output %q, got %q", expected, out)
	}
}

func TestBuildNativeRouterProgram(t *testing.T) {
	rootMsg := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nRoot Endpoint"
	helloMsg := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello Endpoint"
	notFoundMsg := "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\n\r\nNot Found Endpoint"

	routerBF := BuildNativeRouterProgram(rootMsg, helloMsg, notFoundMsg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "match root route",
			input:    "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			expected: rootMsg,
		},
		{
			name:     "match hello route",
			input:    "GET /hello HTTP/1.1\r\nHost: localhost\r\n\r\n",
			expected: helloMsg,
		},
		{
			name:     "match unknown route",
			input:    "GET /something-else HTTP/1.1\r\nHost: localhost\r\n\r\n",
			expected: notFoundMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := interpreter.Execute(routerBF, interpreter.WithInput(strings.NewReader(tt.input)))
			if err != nil {
				t.Fatalf("failed to execute router: %v", err)
			}
			if out != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, out)
			}
		})
	}
}
