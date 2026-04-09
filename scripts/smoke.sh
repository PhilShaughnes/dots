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

# pre-existing real file — should show as blocked
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

echo "=== list (before) ==="
"$DOTS" list -f "$DIR/machine.toml" || true

echo ""
echo "=== apply ==="
"$DOTS" apply -f "$DIR/machine.toml"

echo ""
echo "=== list (after) ==="
"$DOTS" list -f "$DIR/machine.toml" || true

echo ""
echo "=== -s blocked -o dest ==="
"$DOTS" list -f "$DIR/machine.toml" -s blocked -o dest || true

echo ""
echo "=== -o name ==="
"$DOTS" list -f "$DIR/machine.toml" -o name || true

echo ""
echo "=== -s '!ok' ==="
"$DOTS" list -f "$DIR/machine.toml" -s '!ok' || true

echo ""
echo "=== pipe names to apply ==="
"$DOTS" list -f "$DIR/machine.toml" -s empty -o name | "$DOTS" apply -f "$DIR/machine.toml" || true
