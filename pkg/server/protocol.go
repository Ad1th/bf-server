package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// SerializeRequest formats an http.Request into standard HTTP wire protocol bytes for Brainfuck stdin.
func SerializeRequest(r *http.Request) ([]byte, error) {
	var buf bytes.Buffer

	// Request line: METHOD URI PROTOCOL\r\n
	uri := r.URL.RequestURI()
	if uri == "" {
		uri = "/"
	}
	proto := r.Proto
	if proto == "" {
		proto = "HTTP/1.1"
	}
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n", r.Method, uri, proto))

	// Host header
	if r.Host != "" {
		buf.WriteString(fmt.Sprintf("Host: %s\r\n", r.Host))
	}

	// Request headers
	for key, values := range r.Header {
		if strings.EqualFold(key, "Host") {
			continue // Already written
		}
		for _, val := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, val))
		}
	}

	// End of headers
	buf.WriteString("\r\n")

	// Request body if present
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		// Restore body reader in case downstream needs it
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		buf.Write(bodyBytes)
	}

	return buf.Bytes(), nil
}

// ParsedResponse holds the components of an executed Brainfuck program output.
type ParsedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// ParseBrainfuckOutput inspects raw Brainfuck stdout and decomposes it into status, headers, and body.
// If defaultStatus is provided, it is used when no explicit HTTP status line is emitted.
func ParseBrainfuckOutput(output []byte, defaultStatus int) (*ParsedResponse, error) {
	if defaultStatus == 0 {
		defaultStatus = http.StatusOK
	}

	resp := &ParsedResponse{
		StatusCode: defaultStatus,
		Header:     make(http.Header),
		Body:       nil,
	}

	// If output is empty, return empty response
	if len(output) == 0 {
		resp.Body = []byte{}
		return resp, nil
	}

	// Check if output starts with "HTTP/" or "Status:"
	if bytes.HasPrefix(output, []byte("HTTP/")) {
		return parseFullHTTPResponse(output)
	}

	if bytes.HasPrefix(output, []byte("Status:")) || bytes.HasPrefix(output, []byte("status:")) {
		return parseStatusHeaderResponse(output, defaultStatus)
	}

	// Plain body response
	resp.Body = output
	return resp, nil
}

// parseFullHTTPResponse parses raw "HTTP/1.1 200 OK\r\nHeader: Val\r\n\r\nBody"
func parseFullHTTPResponse(output []byte) (*ParsedResponse, error) {
	reader := bufio.NewReader(bytes.NewReader(output))

	// Status line
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP status line: %w", err)
	}
	statusLine = strings.TrimRight(statusLine, "\r\n")
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed HTTP status line: %s", statusLine)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP status code: %s", parts[1])
	}

	header := make(http.Header)
	// Read headers until empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // End of headers
		}

		colonIdx := strings.IndexByte(line, ':')
		if colonIdx > 0 {
			k := strings.TrimSpace(line[:colonIdx])
			v := strings.TrimSpace(line[colonIdx+1:])
			header.Add(k, v)
		}
	}

	// Remaining is body
	var body bytes.Buffer
	_, err = io.Copy(&body, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &ParsedResponse{
		StatusCode: statusCode,
		Header:     header,
		Body:       body.Bytes(),
	}, nil
}

// parseStatusHeaderResponse parses "Status: 404\r\nHeader: Val\r\n\r\nBody"
func parseStatusHeaderResponse(output []byte, defaultStatus int) (*ParsedResponse, error) {
	reader := bufio.NewReader(bytes.NewReader(output))
	statusCode := defaultStatus
	header := make(http.Header)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // End of headers
		}

		colonIdx := strings.IndexByte(line, ':')
		if colonIdx > 0 {
			k := strings.TrimSpace(line[:colonIdx])
			v := strings.TrimSpace(line[colonIdx+1:])
			if strings.EqualFold(k, "Status") {
				codeParts := strings.Fields(v)
				if len(codeParts) > 0 {
					if code, err := strconv.Atoi(codeParts[0]); err == nil {
						statusCode = code
					}
				}
			} else {
				header.Add(k, v)
			}
		}
	}

	var body bytes.Buffer
	_, _ = io.Copy(&body, reader)

	return &ParsedResponse{
		StatusCode: statusCode,
		Header:     header,
		Body:       body.Bytes(),
	}, nil
}
