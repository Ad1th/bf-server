#!/usr/bin/env sh
# serve.sh - Runs server.bf as a live HTTP server over TCP
set -eu

PORT="${1:-${PORT:-8080}}"
BF_INTERPRETER="${BF_INTERPRETER:-./bin/bf}"
SERVER_FILE="${SERVER_FILE:-server.bf}"

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║          ⚡ PURE BRAINFUCK HTTP WEB SERVER ⚡             ║"
echo "║ Listening on http://localhost:$PORT                       ║"
echo "║ Routes: / , /hello , /about , /echo , /teapot , (404)     ║"
echo "╚═══════════════════════════════════════════════════════════╝"

if command -v socat >/dev/null 2>&1; then
    exec socat TCP-LISTEN:"$PORT",fork,reuseaddr EXEC:"$BF_INTERPRETER $SERVER_FILE"
elif command -v nc >/dev/null 2>&1; then
    while true; do
        "$BF_INTERPRETER" "$SERVER_FILE" | nc -l "$PORT" || true
    done
else
    echo "Error: neither socat nor nc found to listen on port $PORT" >&2
    exit 1
fi
