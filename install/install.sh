#!/usr/bin/env bash
# Idempotent priorart installer. Registers the MCP server into each present harness,
# writes the protocol fragment, seeds config, and re-enables opencode context7/webfetch.
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_PATH="${CONFIG_PATH:-$HOME/.config/priorart/priorart.json}"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$INSTALL_DIR/priorart"
PROTOCOL_SRC="$REPO_DIR/protocol/REUSE-FIRST.md"

log() { printf '[priorart] %s\n' "$*"; }

# 1. Build (or copy) the binary
if [ ! -x "$BIN" ]; then
  log "building binary -> $BIN"
  mkdir -p "$INSTALL_DIR"
  (cd "$REPO_DIR" && make build)
  cp "$REPO_DIR/dist/priorart" "$BIN"
else
  log "binary present at $BIN"
fi

# 2. Seed config if missing
if [ ! -f "$CONFIG_PATH" ]; then
  mkdir -p "$(dirname "$CONFIG_PATH")"
  cat > "$CONFIG_PATH" <<EOF
{
  "own_roots": ["$HOME/IdeaProjects"],
  "ecosystems": ["py", "js", "rs", "go"],
  "github": { "own_account": "berlinguyinca", "use_own_layer": false },
  "ledger_path": "$HOME/.config/priorart/ledger/decisions.jsonl",
  "context7": { "enabled": true },
  "block_licenses": ["GPL-3.0", "AGPL-3.0"],
  "cache_ttl_seconds": 3600
}
EOF
  log "seeded config at $CONFIG_PATH"
fi

# 3. Register MCP into each present harness (idempotent: check before write)
register_codex() {
  local f="$HOME/.codex/config.toml"
  [ -f "$f" ] || return 0
  if ! grep -q '\[mcp_servers.priorart\]' "$f"; then
    sed "s#__INSTALL_DIR__#$INSTALL_DIR#g; s#__CONFIG_PATH__#$CONFIG_PATH#g" "$REPO_DIR/install/harness/codex.toml" >> "$f"
    log "registered codex"
  fi
}
register_opencode() {
  local f="$HOME/.config/opencode/opencode.json"
  [ -f "$f" ] || return 0
  python3 - "$f" "$INSTALL_DIR" "$CONFIG_PATH" <<'PY'
import json,sys
f,idd,cp=sys.argv[1],sys.argv[2],sys.argv[3]
d=json.load(open(f))
m=d.setdefault("mcp",{})
if "priorart" not in m:
    m["priorart"]={"type":"local","command":[idd+"/priorart",cp]}
m.setdefault("context7",{})["enabled"]=True
d.setdefault("agent",{}).setdefault("build",{}).setdefault("tools",{})
d["agent"]["build"]["tools"]["context7_*"]=True
d["agent"]["build"]["tools"]["webfetch"]=True
json.dump(d,open(f,"w"),indent=2)
PY
  log "registered opencode (+context7/webfetch enabled)"
}
register_claude() {
  local f="$HOME/.claude.json"
  [ -f "$f" ] || return 0
  python3 - "$f" "$INSTALL_DIR" "$CONFIG_PATH" <<'PY'
import json,sys
f,idd,cp=sys.argv[1],sys.argv[2],sys.argv[3]
d=json.load(open(f))
s=d.setdefault("mcpServers",{})
if "priorart" not in s:
    s["priorart"]={"command":idd+"/priorart","args":[cp]}
json.dump(d,open(f,"w"),indent=2)
PY
  log "registered claude"
}
write_protocol() {
  local target="$1"
  [ -f "$target" ] || return 0
  if ! grep -q "Reuse-first" "$target"; then
    printf '\n%s\n' "$(cat "$PROTOCOL_SRC")" >> "$target"
    log "wrote protocol to $target"
  fi
}

register_codex
register_opencode
register_claude
write_protocol "$HOME/.codex/AGENTS.md"
write_protocol "$HOME/.claude/CLAUDE.md"
if [ -d "$HOME/.config/opencode/rules" ]; then
  cp -n "$PROTOCOL_SRC" "$HOME/.config/opencode/rules/REUSE-FIRST.md" 2>/dev/null || true
fi

log "done. Set GITHUB_TOKEN / CONTEXT7_API_KEY in your shell env."
