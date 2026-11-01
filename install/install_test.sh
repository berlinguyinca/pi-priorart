#!/usr/bin/env bash
# Runs install.sh twice against a temp HOME; asserts no duplicate registrations.
set -euo pipefail
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAKE_HOME="$(mktemp -d)"
trap 'rm -rf "$FAKE_HOME"' EXIT

mkdir -p "$FAKE_HOME/.local/bin" "$FAKE_HOME/.codex" "$FAKE_HOME/.config/opencode" "$FAKE_HOME/.config/priorart"
printf '#!/usr/bin/env bash\necho priorart-stub\n' > "$FAKE_HOME/.local/bin/priorart"
chmod +x "$FAKE_HOME/.local/bin/priorart"
printf '[mcp_servers.existing]\ncommand = "x"\n' > "$FAKE_HOME/.codex/config.toml"
printf '{"mcp":{"existing":{"enabled":true}},"agent":{"build":{"tools":{}}}}\n' > "$FAKE_HOME/.config/opencode/opencode.json"
printf '{"mcpServers":{"existing":{}}}\n' > "$FAKE_HOME/.claude.json"

HOME="$FAKE_HOME" INSTALL_DIR="$FAKE_HOME/.local/bin" bash "$REPO_DIR/install/install.sh" >/dev/null
HOME="$FAKE_HOME" INSTALL_DIR="$FAKE_HOME/.local/bin" bash "$REPO_DIR/install/install.sh" >/dev/null

[ "$(grep -c '\[mcp_servers.priorart\]' "$FAKE_HOME/.codex/config.toml")" = "1" ]
[ "$(python3 -c "import json;print(json.load(open('$FAKE_HOME/.config/opencode/opencode.json'))['mcp']['priorart']['command'][0])")" = "$FAKE_HOME/.local/bin/priorart" ]
[ "$(python3 -c "import json;print(json.load(open('$FAKE_HOME/.claude.json'))['mcpServers']['priorart']['command'])")" = "$FAKE_HOME/.local/bin/priorart" ]
echo "install_test: OK (idempotent)"
