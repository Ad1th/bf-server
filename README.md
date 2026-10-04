# bf-server ⚡

A production-quality, high-performance HTTP web server runtime where routing, request inspection, and response generation are powered entirely by **Brainfuck**.

Written in Go with **zero third-party dependencies** (standard library only), `bf-server` features a bytecode-optimizing Brainfuck virtual machine, deterministic step execution limits, memory bounds sandboxing, structured telemetry logging (`log/slog`), and hot-reloading for development.

---

## 🏗 Architecture Overview

```
                           +-------------------------------------+
                           |            HTTP Client              |
                           +-------------------------------------+
                                              │ HTTP Request (e.g. GET /hello)
                                              ▼
                           +-------------------------------------+
                           |             bf-server               |
                           |  (net/http + slog + Panic Recovery) |
                           +-------------------------------------+
                                              │
                                              ▼
                           +-------------------------------------+
                           |            Router Table             |
                           |  maps URLs -> .bf files in /app     |
                           |  (e.g., /hello -> app/hello.bf)     |
                           +-------------------------------------+
                                              │
                                              ▼
                           +-------------------------------------+
                           |      Brainfuck VM Sandboxed Exec    |
                           |   - Request serialized to STDIN (,) |
                           |   - Response emitted to STDOUT (.)  |
                           |   - Max Steps Limit Enforcement     |
                           |   - Memory Bounds Isolation         |
                           +-------------------------------------+
                                              │
                                              ▼
                           +-------------------------------------+
                           |          Response Parser            |
                           |  - Parses status / custom headers   |
                           |  - Adds X-Brainfuck-Steps telemetry |
                           +-------------------------------------+
                                              │ HTTP Response
                                              ▼
                           +-------------------------------------+
                           |            HTTP Client              |
                           +-------------------------------------+
```

---

## 🧠 Brainfuck Execution Model & VM Design

`bf-server` compiles standard Brainfuck source code into an Intermediate Representation (IR) that optimizes common instruction idioms while strictly preserving 100% Brainfuck semantic compliance:

1. **Instruction Run-Length Compression**: Sequences of `+`, `-`, `>`, and `<` are coalesced into single `OpAdd(n)` and `OpMove(n)` operations.
2. **Clear Loop Optimization**: Idiomatic zeroing loops `[-]` and `[+]` are recognized at compile time and transformed into atomic `OpSet(0)` instructions.
3. **Precomputed Jump Table**: Loop bounds `[` and `]` are resolved to absolute target addresses at compile time, eliminating linear scanning during jumps.
4. **Syntax Diagnostics**: Bracket balance is verified upfront with precise line and column error reporting.
5. **Memory Model**: 8-bit unsigned integer cells (`0..255`) with standard overflow wrapping modulo 256. Tape size is configurable per server instance (default: 30,000 cells).
6. **Per-Request Sandboxing**: Every HTTP request runs on an isolated memory tape. State does not leak across requests or concurrent worker goroutines.

### Supported Instruction Set

| Opcode | Operation | Description |
| :--- | :--- | :--- |
| `+` | Increment | Increments the byte value at the memory pointer (mod 256). |
| `-` | Decrement | Decrements the byte value at the memory pointer (mod 256). |
| `>` | Shift Right | Moves the memory pointer right by 1 cell. Bounds checked against tape limit. |
| `<` | Shift Left | Moves the memory pointer left by 1 cell. Bounds checked (cannot move `< 0`). |
| `[` | Jump If Zero | Jumps past matching `]` if current cell is `0`. |
| `]` | Jump Non-Zero | Jumps back to instruction after matching `[` if current cell is non-zero. |
| `.` | Output | Emits the current cell's byte value to response output stream. |
| `,` | Input | Reads one byte from request input stream into current cell (`0` on EOF). |

---

## 🔌 HTTP-to-Brainfuck Interface

### 1. Request Input (via `,`)

When a client sends an HTTP request, `bf-server` pipes the raw HTTP request wire format directly into the Brainfuck program's input stream:

```http
METHOD /path?query HTTP/1.1\r\n
Host: localhost:8080\r\n
User-Agent: curl/8.7.1\r\n
Content-Type: text/plain\r\n
\r\n
<request body bytes>
```

Brainfuck programs can read header values, parse query parameters, or echo request bodies using `,`. Reading past EOF sets the cell to `0`.

### 2. Response Output (via `.`)

`bf-server` supports two output modes:

#### Mode A: Plain Response Body (Automatic 200 OK)
If the Brainfuck program emits raw text (e.g. `Hello, World!`), `bf-server` automatically sets:
- **Status**: `200 OK` (or `404 Not Found` if routed through `404.bf`)
- **Content-Type**: Inferred automatically (`text/html; charset=utf-8` if HTML tags are present, otherwise `text/plain; charset=utf-8`)

#### Mode B: Full HTTP Response with Custom Headers & Status
Brainfuck programs can emit custom HTTP status codes and headers by starting output with an HTTP status line or `Status:` header:

```http
HTTP/1.1 418 I'm a teapot\r\n
Content-Type: text/plain\r\n
X-Custom-Header: brainfuck\r\n
\r\n
I am a teapot!
```

### 3. Automatic Telemetry Headers

Every HTTP response includes telemetry headers added by `bf-server`:
- `X-Brainfuck-Steps`: Total elementary Brainfuck instructions executed.
- `X-Brainfuck-Memory`: Peak memory cells touched on the tape.
- `X-Brainfuck-Duration`: Execution time spent inside the Brainfuck VM.
- `Server`: `bf-server/1.0`

---

## 🚀 Installation & Quick Start

### Prerequisites
- Go 1.22+ (tested on Go 1.26)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/madith/bf-server.git
cd bf-server

# Build the binary
go build -o bf-server ./cmd/bf-server
```

### Run Example Server

```bash
./bf-server --app ./examples/basic --addr :8080 --dev
```

In another terminal:

```bash
# Test root index
curl -i http://localhost:8080/

# Test hello endpoint
curl -i http://localhost:8080/hello

# Test 404 fallback
curl -i http://localhost:8080/nonexistent-route
```

---

## ⚙️ CLI Flags & Configuration

```bash
bf-server [flags]
```

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--addr` | `string` | `:8080` | TCP address to listen on (`:8080`, `127.0.0.1:3000`). |
| `--app` | `string` | `./app` | Directory containing `.bf` route files. |
| `--memory` | `uint` | `30000` | Tape size in bytes per request. |
| `--max-steps`| `uint64` | `10000000` | Max instruction steps before halting (`0` for unlimited). |
| `--dev` | `bool` | `false` | Enable hot-reloading (re-scans `.bf` files on change). |
| `--timeout` | `duration` | `5s` | Maximum execution duration per request. |
| `--log-format`| `string` | `text` | Log format: `text` or `json`. |
| `--log-level` | `string` | `info` | Log verbosity: `debug`, `info`, `warn`, `error`. |
| `--config` | `string` | `""` | Path to JSON configuration file. |

### Configuration File (`bf-server.json`)

You can also pass a JSON configuration file via `--config config.json`:

```json
{
  "addr": ":8080",
  "app_dir": "./examples/basic",
  "memory_size": 30000,
  "max_steps": 10000000,
  "dev_mode": true,
  "log_format": "json",
  "log_level": "info",
  "timeout": "5s"
}
```

---

## 📁 Routing Conventions

`bf-server` maps filesystem paths under `--app` directly to HTTP routes:

```
examples/basic/
├── index.bf        ──► GET / and GET /index
├── hello.bf        ──► GET /hello
├── about.bf        ──► GET /about
├── api/
│   ├── index.bf    ──► GET /api and GET /api/index
│   └── users.bf    ──► GET /api/users
├── 404.bf          ──► Fallback handler for unmatched routes
└── 500.bf          ──► Fallback handler for runtime errors
```

---

## 🔄 Hot-Reloading in Dev Mode (`--dev`)

With `--dev` enabled:
1. Changes to existing `.bf` files take effect immediately on the next HTTP request.
2. Newly created `.bf` files are automatically registered as new routes without restarting the server process.

---

## 🛡 Security Considerations & Sandboxing

1. **Infinite Loop Prevention**: Brainfuck programs with unbounded loops (e.g. `+[]`) are halted automatically when `max_steps` is exceeded. The server returns `508 Loop Detected` and releases resources.
2. **Memory Isolation**: Each request operates on a freshly allocated, zeroed slice. No data is shared between requests.
3. **Memory Bounds Enforcement**: Any pointer movement beyond `[0, memory_size - 1]` triggers `ErrPointerOutOfBounds` and returns HTTP 500.
4. **Context Cancellation & Timeout**: If a client disconnects or request timeout expires, execution in the VM is aborted cleanly.
5. **No System Access**: Brainfuck has no OS syscalls or file system APIs, providing an inherently sandboxed execution sandbox.

---

## 🧪 Testing & Benchmarks

```bash
# Run all unit and integration tests with race detector
go test -v -race ./...

# Run interpreter benchmarks
go test -bench=. -benchmem ./pkg/interpreter
```

### Benchmark Results (Apple M3)

```
BenchmarkInterpreter_Compile-8       568338     1942 ns/op    12272 B/op   16 allocs/op
BenchmarkInterpreter_HelloWorld-8    315475     3417 ns/op    32946 B/op    6 allocs/op
BenchmarkInterpreter_NestedLoops-8    72181    16667 ns/op    32946 B/op    6 allocs/op
BenchmarkInterpreter_EchoStream-8     16786    71628 ns/op    32978 B/op    7 allocs/op
```

---

## 📦 Example Applications

- `examples/router/`: **Native Brainfuck Router** — A single Brainfuck application (`index.bf`) that reads the raw HTTP request line on stdin (`,`), branches on the request path (`/`, `/hello`, unmatched), and emits corresponding HTTP responses and status codes entirely within Brainfuck.
- `examples/basic/`: Standard multi-file route mapping (`/`, `/hello`, `/about`, `404`)
- `examples/echo/`: Raw HTTP request streaming echo (`index.bf`) using `,[.,]`
- `examples/html/`: Full HTML5 web page with modern CSS rendered and served from Brainfuck
- `examples/custom_status/`: Emits custom HTTP 418 I'm a teapot status (`teapot.bf`)

---

## 🛠 Brainfuck Code Generator (`bfgen`)

The repository includes a native Go-based Brainfuck generator tool (`cmd/bfgen`):

```bash
# Generate Brainfuck code for any text
go run ./cmd/bfgen "Hello, World!" -o hello.bf

# Generate a complete native Brainfuck HTTP request router
go run ./cmd/bfgen -router -o router.bf
```

> [!NOTE]
> The entire project, compiler, VM, HTTP server, and generator toolchain is written strictly in Go and pure Brainfuck with **zero external dependencies and zero Python**.

---

## 📜 License

MIT License. See [LICENSE](LICENSE) for details.
