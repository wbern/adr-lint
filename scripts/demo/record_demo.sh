#!/usr/bin/env bash
# Record an asciinema cast + agg-rendered GIF for one or more demo sections.
#
# Outputs:
#   docs/demo-<section>.cast
#   docs/demo-<section>.gif
#
# Usage:
#   ./scripts/demo/record_demo.sh                       # records all sections
#   ./scripts/demo/record_demo.sh create lint           # just those two
#   ./scripts/demo/record_demo.sh --no-gif              # skip agg conversion
#
# Requires: asciinema, agg (brew install asciinema agg). Plus the `claude`
# CLI on PATH for the `lint` section.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCENARIO="$REPO_ROOT/scripts/demo/scenario.sh"
DOCS_DIR="$REPO_ROOT/docs"
TMP_BIN="$(mktemp -d)/bin"
mkdir -p "$TMP_BIN" "$DOCS_DIR"

# Recording geometry. Matches a comfortable README display size.
COLS=100
ROWS=30
THEME="${AGG_THEME:-monokai}"
FONT_SIZE="${AGG_FONT_SIZE:-14}"

WANT_GIF=1
SECTIONS=()
for arg in "$@"; do
    case "$arg" in
        --no-gif) WANT_GIF=0 ;;
        create|lint|branch|lifecycle|all) SECTIONS+=("$arg") ;;
        *) echo "unknown arg: $arg" >&2; exit 2 ;;
    esac
done
if [[ ${#SECTIONS[@]} -eq 0 ]]; then
    SECTIONS=(create lint branch lifecycle all)
fi

echo "==> Building adr-lint binary into $TMP_BIN"
(cd "$REPO_ROOT/go" && go build -o "$TMP_BIN/adr-lint" ./cmd/adr-lint)

# Workspace each scenario boots into. We avoid /tmp because macOS resolves
# it to /private/tmp via git rev-parse, which leaks into adr-lint's
# "Created /private/tmp/..." output and clutters the GIF.
WORKSPACE_ROOT="${HOME:?HOME must be set}/.adr-lint-demo-workspace"
rm -rf "$WORKSPACE_ROOT"
mkdir -p "$WORKSPACE_ROOT"
trap 'rm -rf "$WORKSPACE_ROOT"' EXIT

record_one() {
    local section="$1"
    local workdir="$WORKSPACE_ROOT/$section"
    mkdir -p "$workdir"

    local cast="$DOCS_DIR/demo-$section.cast"
    local gif="$DOCS_DIR/demo-$section.gif"

    # Bootstrap the workspace silently outside the recording. Every demo
    # assumes a fresh git repo named my-project; nothing is gained by making
    # the viewer watch `mkdir`, `git init`, and adr-lint version/help.
    local proj="$workdir/my-project"
    echo "==> Bootstrapping workspace for: $section (silent)"
    (
        mkdir -p "$proj"
        cd "$proj"
        git init -q
        git commit --allow-empty -q -m 'Initial commit'
    )

    # Sections that operate on an existing ADR (lint, lifecycle, branch) need
    # the `create` flow to have run. Do that silently too.
    case "$section" in
        lint|lifecycle|branch)
            echo "==> Pre-running create for: $section (silent)"
            PATH="$TMP_BIN:$PATH" \
                TYPE_SPEED=0 BEAT=0 \
                bash -c "cd '$proj' && bash '$SCENARIO' create" \
                >/dev/null 2>&1
            ;;
    esac

    # The branch demo additionally needs main to hold the committed ADR, and
    # a feature branch with a violating commit so `--branch` has a real diff
    # to lint.
    if [[ "$section" == "branch" ]]; then
        echo "==> Setting up branch state for: $section (silent)"
        (
            cd "$proj"
            git add -A
            git commit -q -m 'Add no-fmt-println ADR'
            git checkout -q -b add-handler
            cat > handler.go <<'GOFILE'
package main

import "fmt"

func handle() {
    fmt.Println("received request")
}
GOFILE
            git add handler.go
            git commit -q -m 'Add handler'
        ) >/dev/null 2>&1
    fi

    echo "==> Recording section: $section"
    # Inherit the outer env (so `claude` can find its config + auth) but
    # prepend $TMP_BIN to PATH so the freshly-built adr-lint wins.
    PATH="$TMP_BIN:$PATH" \
        TYPE_SPEED="${TYPE_SPEED:-0.025}" \
        BEAT="${BEAT:-1.2}" \
        asciinema rec \
            --overwrite \
            --window-size "${COLS}x${ROWS}" \
            --idle-time-limit 1.5 \
            --command "cd '$proj' && bash '$SCENARIO' '$section'" \
            "$cast"
    echo "    wrote $cast"

    if [[ "$WANT_GIF" -eq 1 ]]; then
        if command -v agg >/dev/null 2>&1; then
            agg --theme "$THEME" --font-size "$FONT_SIZE" "$cast" "$gif"
            echo "    wrote $gif"
        else
            echo "    agg not installed - skipping gif (brew install agg)" >&2
        fi
    fi
}

for s in "${SECTIONS[@]}"; do
    record_one "$s"
done

echo "==> Done. Outputs in $DOCS_DIR/"
