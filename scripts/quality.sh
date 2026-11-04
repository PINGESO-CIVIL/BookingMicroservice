#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if command -v cygpath >/dev/null 2>&1; then
    PS_SCRIPT="$(cygpath -w "${SCRIPT_DIR}/quality.ps1")"
elif command -v wslpath >/dev/null 2>&1; then
    PS_SCRIPT="$(wslpath -w "${SCRIPT_DIR}/quality.ps1")"
else
    PS_SCRIPT="${SCRIPT_DIR}/quality.ps1"
fi

ACTION="${1:-Status}"

if command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PS_SCRIPT" "$ACTION"
elif command -v pwsh >/dev/null 2>&1; then
    pwsh -NoProfile -File "$PS_SCRIPT" "$ACTION"
else
    echo "Error: Se requiere PowerShell (powershell.exe o pwsh) en el PATH." >&2
    exit 1
fi
