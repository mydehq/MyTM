#!/bin/env bash

# MyThemes Utility Functions
# Shared utilities for theme management scripts

#------------ Colors ------------------

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

#------------ Logging Functions ------------------

log.info() { echo -e "${BLUE}!${NC} $1" >&2; }

log.success() { echo -e "${GREEN}✔️${NC} $1" >&2; }

log.warn() { echo -e "${YELLOW}⚠️ ${NC} $1" >&2; }

log.error() { echo -e "${RED}❌${NC} $1" >&2; }

#------------ Utility Functions ------------------

has-cmd() {
  local cmd_str cmd_bin
  local missing=0

  [ "$#" -eq 0 ] && {
    log.error "No arguments provided."
    return 1
  }

  for cmd_str in "$@"; do
    # Extract binary name (first token before space)
    cmd_bin="${cmd_str%% *}"

    if ! command -v "$cmd_bin" &>/dev/null; then
      log.error "$cmd_str is required but not installed"
      missing=1
    fi
  done

  # Return 0 if all found, 1 if any were missing
  return "$missing"
}

get-conf() {
    local key json_flag=""
    local conf_file="$CONFIG_FILE" root_key=".packaging"

    # Parse flags
    while [ "$#" -gt 0 ]; do
        case "$1" in
            -j|--json)
                json_flag="-o=json"
                shift
                ;;
            *)
                break
                ;;
        esac
    done

    # Get remaining arguments
    key="$1"

    if [ -n "$2" ]; then
        conf_file="$2"
        root_key=""
    fi

    # Get config value
    value="$(yq $json_flag eval "${root_key}.${key}" "$conf_file" 2>/dev/null)" || {
        log.error "Failed to get config value for key '$key' from '$conf_file'"
        return 1
    }

    if [ "$value" == "null" ] || [ -z "$value" ]; then
        log.error "Config key '$key' not found or empty in '$conf_file'"
        return 1
    fi

    echo "$value"
}

replace-vars() {
    local raw_value="$1" final_value
    local theme_dir="$2"
    local theme_yml="$theme_dir/theme.yml"
    local theme_name theme_ver

    theme_name="$(basename "$theme_dir")" || exit 1
    theme_ver="$(get-theme-ver "$theme_yml")" || exit 1

    # Replace template variables
    final_value="${raw_value//\$\{\{name\}\}/$theme_name}"
    final_value="${final_value//\$\{\{version\}\}/$theme_ver}"

    echo "$final_value"
}

get-theme-ver() {
    local theme_yml="$1"
    local version

    version=$(yq eval '.version' "$theme_yml" 2>/dev/null | tr -d '\0') || {
        log.error "Failed to get theme version from $theme_yml"
        return 1
    }

    if [ "$version" == "null" ] || [ -z "$version" ]; then
        log.error "version is not defined in theme.yml: $theme_yml"
        return 1
    fi

    echo "$version"
}

# Format bytes to human readable size (KB, MB, GB)
format-size() {
    local bytes="$1"
    local unit="B"
    local size="$bytes"

    if [ "$bytes" -ge 1073741824 ]; then
        size=$((bytes / 1073741824))
        unit="GB"
    elif [ "$bytes" -ge 1048576 ]; then
        size=$((bytes / 1048576))
        unit="MB"
    elif [ "$bytes" -ge 1024 ]; then
        size=$((bytes / 1024))
        unit="KB"
    fi

    echo "${size}${unit}"
}

# Get file size in bytes
get-file-size() {
    local file="$1"

    if [ ! -f "$file" ]; then
        log.error "File not found: $file"
        return 1
    fi

    stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null
}

# Calculate total size of theme files and index.json
show-repo-summary() {
    local output_dir="$1"
    local total_bytes=0
    local file_count=0
    local index_size=0

    [ -z "$output_dir" ] && {
        log.error "Output directory not specified"
        return 1
    }

    [ ! -d "$output_dir" ] && {
        log.error "Output directory does not exist: $output_dir"
        return 1
    }

    # Calculate index.json size
    if [ -f "$output_dir/index.json" ]; then
        index_size=$(get-file-size "$output_dir/index.json")
        total_bytes=$((total_bytes + index_size))
    fi

    # Calculate theme archive sizes
    for archive in "$output_dir"/*.tar.gz; do
        [ -f "$archive" ] || continue
        local size=$(get-file-size "$archive")
        total_bytes=$((total_bytes + size))
        file_count=$((file_count + 1))
    done

    if [ "$file_count" -eq 0 ]; then
        log.error "No theme archives found in $output_dir"
        return 1
    fi

    # Show total first with colors
    echo -e "${GREEN}Repo Size = $(format-size "$total_bytes")${NC}"
    echo

    echo "Files:"
    # Show individual file sizes
    if [ -f "$output_dir/index.json" ]; then
        printf "   "; echo -e "${BLUE}index.json: ${GREEN}$(format-size "$index_size")${NC}"
    fi

    for archive in "$output_dir"/*.tar.gz; do
        [ -f "$archive" ] || continue
        local size=$(get-file-size "$archive")
        printf "   "; echo -e "${BLUE}$(basename "$archive"): ${GREEN}$(format-size "$size")${NC}"
    done
}
