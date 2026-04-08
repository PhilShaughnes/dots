#!/usr/bin/env bash
set -euo pipefail

DOTS="${1:-./dots}"
DIR=$(mktemp -d)
trap "rm -rf $DIR" EXIT

# fake dotfiles repo
mkdir -p "$DIR/repo/.config/jj"
echo "# zshrc"        > "$DIR/repo/.zshrc"
echo "# aliases"      > "$DIR/repo/.aliases"
echo "[user]"         > "$DIR/repo/.config/jj/config.toml"

# fake home
mkdir -p "$DIR/home"

# pre-existing real file — should show as conflict
echo "# existing"     > "$DIR/home/.aliases"

cat > "$DIR/machine.toml" <<TOML
[[dotfiles]]
name = "zshrc"
src  = "$DIR/repo/.zshrc"
dest = "$DIR/home/.zshrc"

[[dotfiles]]
name = "aliases"
src  = "$DIR/repo/.aliases"
dest = "$DIR/home/.aliases"

[[dotfiles]]
name = "jj"
src  = "$DIR/repo/.config/jj/config.toml"
dest = "$DIR/home/.config/jj/config.toml"
TOML

echo "=== status (before) ==="
"$DOTS" -f "$DIR/machine.toml" status || true

echo ""
echo "=== apply ==="
"$DOTS" -f "$DIR/machine.toml" apply

echo ""
echo "=== status (after) ==="
"$DOTS" -f "$DIR/machine.toml" status || true

echo ""
echo "=== --conflicts ==="
"$DOTS" -f "$DIR/machine.toml" status --conflicts

echo ""
echo "=== --name-only ==="
"$DOTS" -f "$DIR/machine.toml" status --name-only
