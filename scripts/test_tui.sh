#!/usr/bin/env bash
# test_tui.sh — runs TUI unit tests and prints a human-readable summary.
#
# Usage:
#   ./scripts/test_tui.sh           # normal run
#   ./scripts/test_tui.sh -v        # verbose (shows every subtest)
#   ./scripts/test_tui.sh -run Foo  # filter to tests matching "Foo"
#
# Exit code: 0 = all passed, 1 = failures or build error.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "$REPO_ROOT"

VERBOSE=""
FILTER=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -v|--verbose) VERBOSE="-v" ;;
    -run) shift; FILTER="-run $1" ;;
    *) echo "Unknown arg: $1" >&2; exit 1 ;;
  esac
  shift
done

echo "========================================"
echo "  otdfctl TUI Test Suite"
echo "  $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================"
echo ""

# Compile check first — fast feedback before waiting for test runner.
echo "▶ Build check..."
if ! go build ./... 2>&1; then
  echo ""
  echo "✘ BUILD FAILED — fix compilation errors above before running tests."
  exit 1
fi
echo "  ✔ Build OK"
echo ""

# Run tests, capture output.
echo "▶ Running TUI tests..."
echo ""

RAW=$(go test ./tui/... ${VERBOSE} ${FILTER} -count=1 2>&1) || true
EXIT_CODE=$?

echo "$RAW"
echo ""

# Parse summary counts from the raw output.
PASS_COUNT=$(echo "$RAW" | grep -cF 'PASS:' || true)
FAIL_COUNT=$(echo "$RAW" | grep -cF 'FAIL:' || true)

echo "========================================"
if [[ $EXIT_CODE -eq 0 ]]; then
  echo "  RESULT : ✔ ALL PASSED"
else
  echo "  RESULT : ✘ FAILURES DETECTED"
fi
echo "  Passed : $PASS_COUNT"
echo "  Failed : $FAIL_COUNT"
echo "========================================"

# Print a focused failure digest (just the FAIL lines) for easy triage.
if [[ $FAIL_COUNT -gt 0 ]]; then
  echo ""
  echo "── Failed tests ────────────────────────"
  echo "$RAW" | grep -F 'FAIL:' | sed 's/^/  /'
  echo ""
  echo "Re-run a single test:"
  echo "  ./scripts/test_tui.sh -v -run '<TestName>'"
fi

exit $EXIT_CODE
