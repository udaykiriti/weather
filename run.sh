#!/usr/bin/env bash
set -e

BINARY_CLI="weather-cli"
BINARY_WEB="weather-web"

usage() {
  cat <<EOF
Usage: ./run.sh <command> [options]

Commands:
  cli            Build the CLI binary
  web            Build the web server binary
  build          Build both binaries
  run-cli [city] Build and run the CLI  (default city: London)
  run-web        Build and run the web server
  fmt            Format all Go source files
  vet            Run go vet
  clean          Remove built binaries
  help           Show this message

Examples:
  ./run.sh cli
  ./run.sh web
  ./run.sh run-cli London
  ./run.sh run-cli "New York"
  ./run.sh run-web
  ./run.sh clean
EOF
}

cmd_cli() {
  echo "==> Building CLI..."
  go build -o "$BINARY_CLI" ./cmd/cli/
  echo "    Built: $BINARY_CLI"
}

cmd_web() {
  echo "==> Building web server..."
  go build -o "$BINARY_WEB" .
  echo "    Built: $BINARY_WEB"
}

cmd_build() {
  cmd_cli
  cmd_web
}

cmd_run_cli() {
  cmd_cli
  city="${*:-London}"
  echo "==> Running CLI for: $city"
  echo ""
  ./"$BINARY_CLI" "$city"
}

cmd_run_web() {
  cmd_web
  echo "==> Starting web server..."
  ./"$BINARY_WEB"
}

cmd_fmt() {
  echo "==> Formatting Go source files..."
  gofmt -w .
  echo "    Done."
}

cmd_vet() {
  echo "==> Running go vet..."
  go vet ./...
  echo "    Done."
}

cmd_clean() {
  echo "==> Cleaning binaries..."
  rm -f "$BINARY_CLI" "$BINARY_WEB"
  echo "    Removed: $BINARY_CLI $BINARY_WEB"
}

case "${1:-help}" in
  cli)        cmd_cli ;;
  web)        cmd_web ;;
  build)      cmd_build ;;
  run-cli)    shift; cmd_run_cli "$@" ;;
  run-web)    cmd_run_web ;;
  fmt)        cmd_fmt ;;
  vet)        cmd_vet ;;
  clean)      cmd_clean ;;
  help|--help|-h) usage ;;
  *)
    echo "Unknown command: $1"
    echo ""
    usage
    exit 1
    ;;
esac
