#!/bin/bash
#
# md2wechat - Runtime locator for coding-agent skills
# Philosophy: skill executes an already-installed runtime; it does not install one implicitly
#

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================

BINARY_NAME="md2wechat"

# Cache directory (tool-specific, not Claude's cache)
CACHE_DIR="${XDG_CACHE_HOME:-${HOME}/.cache}/md2wechat"
BIN_DIR="${CACHE_DIR}/bin"
VERSION_FILE="${CACHE_DIR}/.version"

get_version() {
    if [[ -n "${MD2WECHAT_SKILL_VERSION:-}" ]]; then
        printf '%s\n' "${MD2WECHAT_SKILL_VERSION}"
        return 0
    fi

    local script_dir repo_root version_file
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    repo_root="$(cd "${script_dir}/../../.." && pwd)"
    version_file="${repo_root}/VERSION"

    if [[ -f "${version_file}" ]]; then
        tr -d '[:space:]' < "${version_file}"
        return 0
    fi

    printf '2.0.2\n'
}

VERSION="$(get_version)"

# =============================================================================
# PLATFORM DETECTION
# =============================================================================

detect_platform() {
    local os arch

    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; return 1 ;;
    esac

    case "$os" in
        darwin|linux) echo "${os}-${arch}" ;;
        msys*|mingw*|cygwin*) echo "windows-${arch}" ;;
        *) echo "Unsupported OS: $os" >&2; return 1 ;;
    esac
}

# =============================================================================
# BINARY MANAGEMENT
# =============================================================================

get_binary_path() {
    local platform=$1
    local path="${BIN_DIR}/${BINARY_NAME}-${platform}"
    [[ "$platform" == windows-* ]] && path="${path}.exe"
    echo "$path"
}

is_cache_valid() {
    local binary=$1
    [[ -x "$binary" ]] && [[ -f "$VERSION_FILE" ]] && [[ "$(cat "$VERSION_FILE" 2>/dev/null)" == "$VERSION" ]]
}

extract_version_from_json() {
    local output=$1
    printf '%s\n' "$output" | sed -n 's/.*"version":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
}

binary_matches_version() {
    local candidate=$1
    local resolved_candidate script_path output actual_version

    [[ -x "$candidate" ]] || return 1

    resolved_candidate="$(cd "$(dirname "$candidate")" && pwd)/$(basename "$candidate")"
    script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    if [[ "$resolved_candidate" == "$script_path" ]]; then
        return 1
    fi

    output="$("$candidate" version --json 2>/dev/null)" || return 1
    actual_version="$(extract_version_from_json "$output")"
    [[ -n "$actual_version" ]] || return 1
    [[ "$actual_version" == "$VERSION" ]]
}

get_release_base_url() {
    printf 'https://github.com/geekjourneyx/md2wechat-skill/releases/download/v%s\n' "${VERSION}"
}

ensure_binary() {
    local platform
    platform=$(detect_platform) || return 1

    local binary
    binary=$(get_binary_path "$platform")

    # Fast path: valid cache
    if is_cache_valid "$binary"; then
        echo "$binary"
        return 0
    fi

    # Try local development binary
    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local local_binary="${script_dir}/bin/${BINARY_NAME}-${platform}"
    [[ "$platform" == windows-* ]] && local_binary="${local_binary}.exe"

    if binary_matches_version "$local_binary"; then
        echo "$local_binary"
        return 0
    fi

    # Try an existing installed CLI on PATH.
    if command -v md2wechat >/dev/null 2>&1; then
        local path_binary
        path_binary="$(command -v md2wechat)"
        if binary_matches_version "$path_binary"; then
            echo "$path_binary"
            return 0
        fi
        echo "  Found md2wechat on PATH, but version does not match v${VERSION}" >&2
    fi

    local release_base_url
    release_base_url="$(get_release_base_url)"

    echo "md2wechat runtime not found for skill v${VERSION}." >&2
    echo "" >&2
    echo "Expected one of:" >&2
    echo "  - ${binary}" >&2
    echo "  - ${local_binary}" >&2
    echo "  - md2wechat on PATH (must report version ${VERSION})" >&2
    echo "" >&2
    echo "Install the CLI first, then rerun the skill. Recommended:" >&2
    echo "  curl -fsSL ${release_base_url}/install.sh | bash" >&2
    echo "" >&2
    return 1
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    local binary
    binary=$(ensure_binary) || exit 1
    exec "$binary" "$@"
}

main "$@"
