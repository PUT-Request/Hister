#!/bin/sh
# Updates the bundled C files from upstream GitHub repositories.
# Usage:
#   go generate ./server/vectorstore/sqlitevec/
#   sh update.sh [sqlite-vec-go-bindings version, e.g. v0.1.6]
#
# If no version is given the script reads the version tag from the comment in
# vec.go (line starting with "// sqlite-vec source:").
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Determine the sqlite-vec-go-bindings version.
if [ -n "$1" ]; then
    SQLITE_VEC_VERSION="$1"
else
    SQLITE_VEC_VERSION=$(grep 'sqlite-vec source:' "${SCRIPT_DIR}/vec.go" | sed 's/.*(\(.*\))/\1/')
fi

if [ -z "$SQLITE_VEC_VERSION" ]; then
    echo "Could not determine sqlite-vec version. Pass it as an argument: sh update.sh v0.1.6"
    exit 0
fi

if ! command -v curl > /dev/null 2>&1; then
    echo "curl is required but not found. Install curl and retry."
    exit 0
fi

BASE_URL="https://raw.githubusercontent.com/asg017/sqlite-vec-go-bindings/${SQLITE_VEC_VERSION}/cgo"

echo "Downloading sqlite-vec ${SQLITE_VEC_VERSION} C source from GitHub..."
for FILE in sqlite-vec.c sqlite-vec.h; do
    URL="${BASE_URL}/${FILE}"
    if ! curl --silent --show-error --fail --location --output "${SCRIPT_DIR}/${FILE}" "${URL}"; then
        echo "Failed to download ${URL}. Check that version ${SQLITE_VEC_VERSION} exists."
        exit 0
    fi
    echo "  ${FILE}"
done

# sqlite3.h comes from mattn/go-sqlite3 which is in go.mod.
MATTN_DIR=$(go list -m -json github.com/mattn/go-sqlite3 2>/dev/null | grep '"Dir"' | sed 's/.*"Dir": "\(.*\)".*/\1/')

if [ -z "$MATTN_DIR" ]; then
    echo "mattn/go-sqlite3 not found in module cache. Run 'go mod download github.com/mattn/go-sqlite3' first."
    exit 0
fi

if [ ! -f "${MATTN_DIR}/sqlite3-binding.h" ]; then
    echo "Expected sqlite3-binding.h not found in ${MATTN_DIR}. The module layout may have changed."
    exit 0
fi

echo "Copying sqlite3.h from mattn/go-sqlite3..."
cp "${MATTN_DIR}/sqlite3-binding.h" "${SCRIPT_DIR}/sqlite3.h"

echo "Done. Remember to update the version comment in vec.go to ${SQLITE_VEC_VERSION}."
