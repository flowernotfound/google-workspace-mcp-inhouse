#!/usr/bin/env bash
set -euo pipefail

REPO="flowernotfound/google-workspace-mcp-inhouse"
BINARY="google-workspace-mcp-inhouse"
INSTALL_DIR="${HOME}/bin"
CONFIG_DIR="${HOME}/.config/google-workspace-mcp-inhouse"

# --------------------------------------------------------------------------
# Detect OS and architecture
# --------------------------------------------------------------------------
detect_platform() {
  local os arch

  case "$(uname -s)" in
    Darwin) os="darwin" ;;
    Linux)  os="linux"  ;;
    *)
      echo "Unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64)          arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *)
      echo "Unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac

  # linux/arm64 is not distributed
  if [ "$os" = "linux" ] && [ "$arch" = "arm64" ]; then
    echo "linux/arm64 is not supported. Please build from source." >&2
    exit 1
  fi

  echo "${os}_${arch}"
}

# --------------------------------------------------------------------------
# Download helper (curl or wget)
# --------------------------------------------------------------------------
http_get() {
  local url="$1"
  if command -v curl > /dev/null 2>&1; then
    curl -fsSL "$url"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO- "$url"
  else
    echo "curl or wget is required to install ${BINARY}." >&2
    exit 1
  fi
}

http_download() {
  local url="$1" dest="$2"
  if command -v curl > /dev/null 2>&1; then
    curl -fsSL -o "$dest" "$url"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO "$dest" "$url"
  else
    echo "curl or wget is required to install ${BINARY}." >&2
    exit 1
  fi
}

# --------------------------------------------------------------------------
# Resolve download URL from GitHub Releases API
# --------------------------------------------------------------------------
resolve_download_url() {
  local platform="$1"
  local asset_name="${BINARY}_${platform}"
  local api_url="https://api.github.com/repos/${REPO}/releases/latest"

  local url
  url=$(http_get "$api_url" \
    | grep '"browser_download_url"' \
    | grep "\"${asset_name}\"" \
    | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')

  if [ -z "$url" ]; then
    echo "Could not find asset '${asset_name}' in the latest release." >&2
    exit 1
  fi

  echo "$url"
}

# --------------------------------------------------------------------------
# Main
# --------------------------------------------------------------------------
main() {
  echo "Installing ${BINARY}..."

  local platform
  platform=$(detect_platform)

  local download_url
  download_url=$(resolve_download_url "$platform")

  # Create install directory
  mkdir -p "$INSTALL_DIR"

  local tmp_file
  tmp_file=$(mktemp)
  trap 'rm -f "$tmp_file"' EXIT

  echo "Downloading ${BINARY} (${platform})..."
  http_download "$download_url" "$tmp_file"
  chmod +x "$tmp_file"
  mv "$tmp_file" "${INSTALL_DIR}/${BINARY}"

  # Create config directory
  mkdir -p "$CONFIG_DIR"

  echo ""
  echo "✓ Installed to ${INSTALL_DIR}/${BINARY}"
  echo ""

  # PATH guidance
  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    echo "Add ~/bin to your PATH by adding the following to ~/.bashrc or ~/.zshrc:"
    echo ""
    echo "  export PATH=\"\$HOME/bin:\$PATH\""
    echo ""
    echo "Then reload your shell:"
    echo ""
    echo "  source ~/.zshrc   # or source ~/.bashrc"
    echo ""
  fi

  # Next steps
  echo "──────────────────────────────────────────────────────────"
  echo "Next steps:"
  echo ""
  echo "1. Place credentials.json in the config directory:"
  echo "   mv ~/Downloads/credentials.json ${CONFIG_DIR}/credentials.json"
  echo ""
  echo "2. Authenticate with your Google account:"
  echo "   ${BINARY} auth"
  echo ""
  echo "3. Register with Claude Code:"
  echo "   claude mcp add ${BINARY} ${INSTALL_DIR}/${BINARY}"
  echo "──────────────────────────────────────────────────────────"
}

main
