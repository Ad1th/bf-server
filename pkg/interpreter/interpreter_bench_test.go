package interpreter

import (
	"strings"
	"testing"
)

func BenchmarkInterpreter_Compile(b *testing.B) {
	source := `++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Compile(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInterpreter_HelloWorld(b *testing.B) {
	source := `++++++++[>++++[>++>+++>+++>+<<<<-]>+>+>->>+[<]<-]>>.>---.+++++++..+++.>>.<-.<.+++.------.--------.>>+.>++.`
	prog, err := Compile(source)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM()
		_, err := vm.Run(prog)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInterpreter_NestedLoops(b *testing.B) {
	// Nested multiplication loop computing 100 * 100
	source := `>++++++++++[<++++++++++>-]<[>++++++++++[>++++++++++<-]<-]`
	prog, err := Compile(source)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM()
		_, err := vm.Run(prog)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInterpreter_EchoStream(b *testing.B) {
	source := `,[.,]`
	prog, err := Compile(source)
	if err != nil {
		b.Fatal(err)
	}
	inputData := strings.Repeat("GET /hello HTTP/1.1\r\nHost: localhost:8080\r\n\r\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM(WithInput(strings.NewReader(inputData)))
		_, err := vm.Run(prog)
		if err != nil {
			b.Fatal(err)
		}
	}
}
