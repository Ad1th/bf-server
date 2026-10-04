.PHONY: all serve test clean

all: test

# Start the pure Brainfuck HTTP server on port 8080 (or specify PORT=...)
serve:
	./serve.sh $(PORT)

# Run the automated test suite against server.bf
test:
	./test.sh

clean:
	rm -f *.tmp *.out coverage.html
