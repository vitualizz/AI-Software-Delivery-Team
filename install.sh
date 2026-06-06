#!/usr/bin/env bash
set -euo pipefail

MIN_GO_MAJOR=1
MIN_GO_MINOR=22

if ! command -v go &>/dev/null; then
  echo "Error: Go is not installed. Install Go $MIN_GO_MAJOR.$MIN_GO_MINOR+ from https://go.dev/dl/"
  exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt "$MIN_GO_MAJOR" ] || { [ "$GO_MAJOR" -eq "$MIN_GO_MAJOR" ] && [ "$GO_MINOR" -lt "$MIN_GO_MINOR" ]; }; then
  echo "Error: Go $MIN_GO_MAJOR.$MIN_GO_MINOR+ required (found $GO_VERSION)"
  exit 1
fi

echo "Installing asdt-tui..."
go install github.com/vitualizz/ai-software-delivery-team/cmd/asdt-tui@latest

echo ""
echo "Done! asdt-tui installed to $(go env GOPATH)/bin/"
echo ""
echo "Next steps:"
echo "  1. Run: asdt-tui"
echo "  2. Select your AI assistant(s) and install ASDT skills"
echo "  3. Open your AI assistant and run: /asdt:init"
