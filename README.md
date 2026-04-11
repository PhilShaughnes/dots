# Dotfiles

Ohi! Welcome to Dots! It's for dotfiles.

Dotfiles are awesome. They're where you write yourself cheatsheets, tune things just so, hack, explore, play. They make your system yours.

They also need a way to be backed up, hacked on, moved around, and kept track of. This is my take on that.

---

Most dotfile tools assume you want the same thing everywhere. Differences are edge cases, solved with templates. But templates break the vcs workflow I actually want — edit, diff, commit, sync — and they don't handle machines that genuinely differ. Sometimes I want the same file in different locations. Sometimes different files in the same location.

The answer turned out to be a simple list. *Here's the thing, here's where it lives.* A small Go CLI, a TOML manifest, no conventions about directory structure, no magic. Declare it, read it, apply it.

---

## What dots provides

1. **A declarative manifest** — a TOML file listing:
   - dotfiles (`src → dest`)
   - git repos (`url → dest`)

2. **Validation / state tracking** — it knows if:
   - `ok` → symlink or repo is correct
   - `empty` → dest missing
   - `blocked` → dest exists but conflicts
   - `src-missing` → source missing

3. **Safe application** — applies symlinks or clones repos without overwriting blocked destinations

4. **Filtering / listing** — query subsets by type, state, or name and pipe them to shell commands

Everything else is left to you — git, scripts, hooks, orchestration.

---

## How it compares

| Tool | Strengths | Weaknesses vs Dots |
|---|---|---|
| **chezmoi** | Templating, host-aware configs, secret encryption, reconciliation | Implicit path conventions, complex add/edit workflow, opinionated templating layer |
| **GNU Stow** | Pure symlink manager, handles directories elegantly | No declarative manifest, directory-structure-only, harder to filter or compose per-machine |
| **Bare scripts + git** | Maximal flexibility | No declarative overview, you handle errors, filtering, and blocked states yourself |
| **Homesick / vcsh / yadm** | Git-centric, some host awareness | Opinionated repo layout, less granular filtering, less shell-composable |
| **Nix Home Manager** | Full declarative system, clean per-machine handling | Heavyweight, requires learning Nix, breaks filesystem-first intuition |

**What makes dots different:**
- Flat declarative manifest — one file per machine, or multiple composed. No templating, no assumed directory structure
- Intentionally minimal — git, shell, CI do their own work
- Queryable — ask "which repos are blocked?" and pipe the answer anywhere
- Safe reconciliation — no blind overwriting
- No path conventions — any layout, any machine

---

## But can dots do...?

Dots is minimal and orthogonal by design. That usually means the answer is *yes, with a simple shell pipe.*

**Let me show you:**


**Check all repos for uncommitted changes:**
```sh
dots list -f base.toml -t repo -o dest | while read -r repo; do
  echo "== $repo =="
  git -C "$repo" status --short
done
```

**Warn on dirty repos:**
```sh
dots list -f work.toml -t repo -o dest |
while read -r repo; do
  git -C "$repo" diff --quiet || echo "dirty: $repo"
done
```

**Auto-pull all repos:**
```sh
dots list -f base.toml -t repo -o dest |
while read -r repo; do
  git -C "$repo" pull --rebase
done
```

**Fetch only repos that are already `ok`:**
```sh
dots list -f machine.toml -t repo -s ok -o dest |
while read -r repo; do
  git -C "$repo" fetch --quiet
done
```

The filtering + shell pipeline pattern is intentionally the primitive. If you can express it as a query on your manifest, you can script it.

---

Dots sets things up and helps you see what's what. The rest is your toolchain doing what it's good at. Have fun!

