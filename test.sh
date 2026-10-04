#!/usr/bin/env sh
# test.sh - Automated test suite for server.bf and route modules
set -eu

BF="./bin/bf"
PASS=0
FAIL=0

run_test() {
    TEST_NAME="$1"
    INPUT="$2"
    EXPECTED_SUBSTR="$3"

    OUTPUT="$(printf "%s" "$INPUT" | $BF server.bf)"

    if echo "$OUTPUT" | grep -F "$EXPECTED_SUBSTR" >/dev/null 2>&1; then
        echo "  ✓ [PASS] $TEST_NAME"
        PASS=$((PASS + 1))
    else
        echo "  ✗ [FAIL] $TEST_NAME"
        echo "    Expected: $EXPECTED_SUBSTR"
        echo "    Received: $OUTPUT"
        FAIL=$((FAIL + 1))
    fi
}

echo "Running test suite on server.bf..."

run_test "GET / (Root HTML)" "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n" "⚡ Pure Brainfuck Server"
run_test "GET / (Status 200)" "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n" "HTTP/1.1 200 OK"
run_test "GET /hello" "GET /hello HTTP/1.1\r\nHost: localhost\r\n\r\n" "Hello, World!"
run_test "GET /about" "GET /about HTTP/1.1\r\nHost: localhost\r\n\r\n" "bf-server: Pure Brainfuck"
run_test "GET /teapot (Status 418)" "GET /teapot HTTP/1.1\r\nHost: localhost\r\n\r\n" "HTTP/1.1 418 I'm a teapot"
run_test "GET /echo" "GET /echo HTTP/1.1\r\nHost: localhost\r\n\r\n" "[Echo Stream]"
run_test "GET /nonexistent (Status 404)" "GET /nonexistent HTTP/1.1\r\nHost: localhost\r\n\r\n" "HTTP/1.1 404 Not Found"

echo ""
echo "Test Results: $PASS passed, $FAIL failed."

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
