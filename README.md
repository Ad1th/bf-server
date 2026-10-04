# bf-server ⚡

A complete, production-quality HTTP web server written **100% in pure Brainfuck**.

No Python. No Golang. No C compiler needed. Just raw, unadulterated Brainfuck instruction set (`+ - < > [ ] . ,`) executing RFC-compliant HTTP request parsing, routing, and dynamic response generation.

---

## 🏛 Architecture & Execution Model

```
                           +-------------------------------------+
                           |            HTTP Client              |
                           +-------------------------------------+
                                              │ HTTP Request (e.g. GET /hello)
                                              ▼
                           +-------------------------------------+
                           |          TCP Socket Pipe            |
                           |   (socat / nc / inetd listener)     |
                           +-------------------------------------+
                                              │ stdin (,)
                                              ▼
                           +-------------------------------------+
                           |             server.bf               |
                           |       PURE BRAINFUCK RUNTIME        |
                           |                                     |
                           |  [1] Method & Path Lexer            |
                           |      reads past "GET /"             |
                           |                                     |
                           |  [2] Route State Machine            |
                           |      ├── ' '  ➔  Route / (HTML)     |
                           |      ├── 'h'  ➔  Route /hello       |
                           |      ├── 'a'  ➔  Route /about       |
                           |      ├── 'e'  ➔  Route /echo        |
                           |      ├── 't'  ➔  Route /teapot      |
                           |      └── *    ➔  Route 404          |
                           |                                     |
                           |  [3] Response Formatter             |
                           |      emits RFC status + headers     |
                           |      streams body payload           |
                           +-------------------------------------+
                                              │ stdout (.)
                                              ▼
                           +-------------------------------------+
                           |          TCP Socket Pipe            |
                           +-------------------------------------+
                                              │ HTTP Response
                                              ▼
                           +-------------------------------------+
                           |            HTTP Client              |
                           +-------------------------------------+
```

---

## 🧠 Brainfuck Memory Tape Layout

`server.bf` allocates discrete operational cells on the tape:

| Cell Index | Name | Function |
| :---: | :--- | :--- |
| `0` | `INPUT_CHAR` | Active input byte read from HTTP request stream via `,`. |
| `1` | `FLAG_ROOT` | Boolean flag (1 = request path is `/`, 0 = otherwise). |
| `2` | `FLAG_HELLO` | Boolean flag (1 = request path is `/hello`, 0 = otherwise). |
| `3` | `FLAG_ABOUT` | Boolean flag (1 = request path is `/about`, 0 = otherwise). |
| `4` | `FLAG_ECHO` | Boolean flag (1 = request path is `/echo`, 0 = otherwise). |
| `5` | `FLAG_TEAPOT`| Boolean flag (1 = request path is `/teapot`, 0 = otherwise). |
| `6` | `FLAG_404` | Fallback flag (initialized to 1; cleared if any route matches). |
| `7` | `CMP_SCRATCH`| Temporary comparison cell for character subtraction. |
| `8` | `TMP_SCRATCH`| Non-destructive register backup cell. |
| `9` | `PRINT_OUT` | Character emission accumulator cell (`.`). |
| `10` | `LOOP_MULT` | Loop counter for multiplication during string synthesis. |

---

## 🚦 Endpoints & Supported Routes

| HTTP Request | Status | Content-Type | Response Description |
| :--- | :---: | :--- | :--- |
| `GET /` | `200 OK` | `text/html` | Styled HTML5 landing page with navigation links. |
| `GET /hello` | `200 OK` | `text/plain` | "Hello, World! Greetings from pure Brainfuck HTTP server." |
| `GET /about` | `200 OK` | `text/plain` | Detailed server and project metadata. |
| `GET /teapot` | `418 I'm a teapot` | `text/plain` | RFC 2324 Hyper Text Coffee Pot response. |
| `POST /echo` | `200 OK` | `text/plain` | Echoes the incoming HTTP headers and body back to client. |
| `*` (Any other) | `404 Not Found` | `text/plain` | Custom 404 route not found error page. |

---

## 🚀 Quick Start

### 1. Start the HTTP Server

Run on default port `8080` (or pass a custom port as argument):

```bash
make serve
# or
./serve.sh 8080
```

### 2. Make HTTP Requests

In another terminal:

```bash
# 1. Fetch styled HTML home page
curl -i http://localhost:8080/

# 2. Fetch /hello endpoint
curl -i http://localhost:8080/hello

# 3. Fetch /about endpoint
curl -i http://localhost:8080/about

# 4. Fetch /teapot (HTTP 418)
curl -i http://localhost:8080/teapot

# 5. Echo stream endpoint
curl -i http://localhost:8080/echo -d "Payload from curl"

# 6. Test 404 handling
curl -i http://localhost:8080/invalid-route
```

---

## 📁 Repository Structure

```
bf-server/
├── server.bf         # Complete HTTP web server runtime in pure Brainfuck
├── bf.bf             # Metacircular Brainfuck interpreter written in Brainfuck
├── routes/           # Standalone single-endpoint Brainfuck route files
│   ├── index.bf      # GET / handler
│   ├── hello.bf      # GET /hello handler
│   ├── about.bf      # GET /about handler
│   ├── echo.bf       # /echo handler
│   ├── teapot.bf     # GET /teapot handler (HTTP 418)
│   └── 404.bf        # 404 fallback handler
├── bin/
│   └── bf            # Zero-dependency POSIX Brainfuck runner (awk/sh)
├── serve.sh          # TCP socket listener (using socat / nc) piping to server.bf
├── test.sh           # Automated integration test suite
├── Makefile          # make serve, make test
├── LICENSE           # MIT License
└── README.md
```

---

## 🧪 Testing

Run the automated test suite verifying all routes and HTTP status codes:

```bash
make test
# or
./test.sh
```

Output:
```text
Running test suite on server.bf...
  ✓ [PASS] GET / (Root HTML)
  ✓ [PASS] GET / (Status 200)
  ✓ [PASS] GET /hello
  ✓ [PASS] GET /about
  ✓ [PASS] GET /teapot (Status 418)
  ✓ [PASS] GET /echo
  ✓ [PASS] GET /nonexistent (Status 404)

Test Results: 7 passed, 0 failed.
```

---

## 🧬 Metacircular Self-Interpretation (`bf.bf`)

The repository also includes `bf.bf`, a Brainfuck interpreter written in Brainfuck. You can theoretically run the Brainfuck HTTP server inside a Brainfuck interpreter executing on another Brainfuck interpreter:

```bash
# Run server.bf through the metacircular bf.bf interpreter
cat server.bf | ./bin/bf bf.bf
```

---

## 📜 License

MIT License. See [LICENSE](LICENSE) for details.
