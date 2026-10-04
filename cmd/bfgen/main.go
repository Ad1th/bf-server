package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/madith/bf-server/pkg/bfgen"
)

func main() {
	outPath := flag.String("o", "", "Output file path (default stdout)")
	isRouter := flag.Bool("router", false, "Generate an HTTP request router program in Brainfuck")
	flag.Parse()

	var code string
	if *isRouter {
		root := "HTTP/1.1 200 OK\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nWelcome to bf-server! (Routed natively in Brainfuck)\n"
		hello := "HTTP/1.1 200 OK\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello, World! Greetings from Brainfuck router.\n"
		notFound := "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n404 Not Found: Brainfuck router could not find this endpoint.\n"
		code = bfgen.BuildNativeRouterProgram(root, hello, notFound)
	} else {
		args := flag.Args()
		var input string
		if len(args) > 0 {
			input = strings.Join(args, " ")
		} else {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			input = string(bytes)
		}
		if input == "" {
			input = "Hello from Brainfuck!\n"
		}
		code = bfgen.TextToBrainfuck(input)
	}

	if *outPath != "" {
		if err := os.WriteFile(*outPath, []byte(code), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %v\n", *outPath, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s successfully.\n", *outPath)
	} else {
		fmt.Print(code)
	}
}
