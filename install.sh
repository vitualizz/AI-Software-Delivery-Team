#!/usr/bin/env bash
set -euo pipefail

REPO="vitualizz/ai-software-delivery-team"
BIN_NAME="asdt-tui"
INSTALL_DIR="${HOME}/.local/bin"

# ── helpers ────────────────────────────────────────────────────────────────────

fatal()   { echo "  [error] $*" >&2; exit 1; }
success() { echo "  [ok]    $*" >&2; }
info()    { echo "  [...]   $*" >&2; }

# ── platform detection ─────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       fatal "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)          echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *)               fatal "Unsupported architecture: $(uname -m)." ;;
  esac
}

# ── download ───────────────────────────────────────────────────────────────────

fetch_latest_version() {
  if command -v curl &>/dev/null; then
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' \
      | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
  else
    fatal "curl is required but not installed."
  fi
}

download_binary() {
  local version="$1" os="$2" arch="$3"
  local asset="${BIN_NAME}_${os}_${arch}"
  local url="https://github.com/${REPO}/releases/download/${version}/${asset}"
  local tmp
  tmp=$(mktemp)

  info "Downloading ${asset} (${version})..."
  if ! curl -fsSL "$url" -o "$tmp"; then
    rm -f "$tmp"
    fatal "Download failed. Check that ${version} has a release for ${os}/${arch}:\n  https://github.com/${REPO}/releases"
  fi

  echo "$tmp"
}

# ── install ────────────────────────────────────────────────────────────────────

install_binary() {
  local tmp="$1"
  mkdir -p "$INSTALL_DIR"
  chmod +x "$tmp"
  mv "$tmp" "${INSTALL_DIR}/${BIN_NAME}"
  success "${BIN_NAME} installed to ${INSTALL_DIR}/${BIN_NAME}"
}

ensure_path() {
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo ""
    echo "  [warn]  ${INSTALL_DIR} is not in your PATH."
    echo "          Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "            export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
  fi
}

# ── main ───────────────────────────────────────────────────────────────────────

main() {
  echo ""
  echo "  asdt-tui"
  echo "  ─────────────────────────────────────────"
  echo ""

  local os arch version tmp

  os=$(detect_os)
  arch=$(detect_arch)
  info "Platform: ${os}/${arch}"

  version=$(fetch_latest_version)
  [ -n "$version" ] || fatal "Could not determine latest release version."

  tmp=$(download_binary "$version" "$os" "$arch")
  install_binary "$tmp"
  ensure_path

  echo ""
  echo "  ─────────────────────────────────────────"
  success "Done!"
  echo ""
  echo "  Next steps:"
  echo "    1. Run:  asdt-tui"
  echo "    2. Select your AI assistant(s) and install ASDT skills"
  echo "    3. In your AI assistant, run:  /asdt-init"
  echo ""
}

main "$@"
