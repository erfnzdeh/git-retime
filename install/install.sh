#!/usr/bin/env bash
set -euo pipefail

REPO="erfnzdeh/git-retime"
BINARY="git-retime"
INSTALL_DIR="/usr/local/bin"

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)  os="linux" ;;
        Darwin*) os="darwin" ;;
        *)       echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)  arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)             echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

main() {
    local platform version download_url tmpdir

    platform="$(detect_platform)"
    echo "Detected platform: ${platform}"

    echo "Fetching latest release..."
    version="$(get_latest_version)"
    if [ -z "$version" ]; then
        echo "Failed to determine latest version." >&2
        exit 1
    fi
    echo "Latest version: ${version}"

    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}_${platform}.tar.gz"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    echo "Downloading ${download_url}..."
    curl -fsSL "$download_url" -o "${tmpdir}/archive.tar.gz"

    echo "Extracting..."
    tar -xzf "${tmpdir}/archive.tar.gz" -C "$tmpdir"

    echo "Installing to ${INSTALL_DIR}/${BINARY}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY}"

    echo ""
    echo "git-retime ${version} installed successfully."
    echo "Run 'git retime HEAD~5' to get started."
}

main "$@"
