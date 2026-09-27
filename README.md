# Dots

Ohi! This is my take on dotfile management.

Git solves backup, sync, and diffing. Symlinks solve scattered files. Shell and
pipes let you quickly operate on whatever you need. What's missing is a single
place to know what and where all those things are, wire them up safely, and
access them in scripts.

That's what dots is: a declarative manifest that ties all these things together
and gets out of their way.

---

## The idea

You write a manifest describing your dotfiles: symlinks and repos, src and dest,
nothing more. Dots shows you their state, applies what's missing safely
(creating symlinks, cloning repos), and exposes everything as a queryable CLI so
your other tools can do their thing.

Dots started as a dotfile tool, but it’s really just a way to declare, track,
and wire up files and repos anywhere on your system.

---

## Example

```toml
[[dotfiles]]
name = "zshrc"
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"

[[repos]]
name = "nvim"
url  = "git@github.com:you/nvim.git"
dest = "~/.config/nvim"
```

```sh
dots list -f machine.toml
dots apply -f machine.toml
```

---

## States

| State | Meaning |
|---|---|
| `ok` | symlink correct or repo present |
| `ok*` | just applied |
| `empty` | nothing at dest — safe to apply |
| `blocked` | dest exists but conflicts — never overwritten |
| `src-missing` | source path missing |

`apply` skips `blocked` and `src-missing` without error.  
`list` exits non-zero if any entry is not `ok`.

---

## Design

**Don't get in the way of git.** Git handles backup, sync, history, and diffing
better than any dotfile tool will. So dots doesn't try. Your files are your
files: symlinked to exactly where they belong so you can edit them in place,
see diffs, commit, and push without any extra steps. No templates to render, no
build step, no editing-in-one-place and applying-somewhere-else. The moment
there's indirection between you and your files, git stops working naturally.

**Don't get in the way of shell.** Dots does one job: read a manifest, report
state, apply it. It takes stdin, writes to stdout, and has straightforward flags
and filters. The interesting workflows aren't built into dots, they're shell
pipelines. Dots should compose cleanly with whatever you're already doing
without hooks, magic, or a required TUI.

**Know what and where, nothing more.** A manifest is just a readable list of
what goes where, checkable against reality. Two primitives cover the space:
symlinks for scattered files you want centralized in a repo, repos for projects
big enough to be their own thing. Every path is fully explicit. No inferred
paths, no directory structure conventions, no single dotfiles repo you're forced
into. Because paths are explicit, machine differences are trivial: same file
different location, different file same location, entirely different manifests
per machine.

**Safe and idempotent.** Dots manages wiring, not content. It never overwrites,
moves, or clobbers. If something's in the way, it tells you and waits. Safe to
run repeatedly, it won't cause problems.

---

## Composability

```sh
# Interactively add dotfiles to manifest
find ~ -maxdepth 3 -name ".*" -type f | fzf --multi \
  | dots add -f machine.toml -r ~/dotfiles

# Migrate existing dotfiles into your repo, then apply
dots list -f machine.toml -t dotfile -s src-missing -o src,dest | while IFS=$'\t' read -r src dest; do
  mv "$dest" "$src"
done
dots apply -f machine.toml

# Back up blocked dotfiles, then apply
dots list -f machine.toml -t dotfile -s blocked -o dest | while read -r dest; do
  mv "$dest" "$dest.bak"
done
dots apply -f machine.toml

# Port a manifest to a new machine (same shape, new dest prefix).
# Repos clone on `apply`; dotfiles whose src isn't there yet report
# src-missing and are skipped until their files exist.
sed 's#~/code/personal/#/data/repos/personal/#' laptop.toml > devserver.toml
dots apply -f devserver.toml

# Interactively choose what to apply
dots list -f machine.toml -o name | fzf --multi | dots apply -f machine.toml

# Check all repos for uncommitted changes
dots list -f machine.toml -t repo -o dest | while read -r repo; do
  echo "== $repo =="
  git -C "$repo" status --short
done

# Fetch only repos that are already ok
dots list -f machine.toml -t repo -s ok -o dest | while read -r repo; do
  git -C "$repo" fetch --quiet
done
```

This isn't a workaround, it's the design.

```
dots  = structural state  (is everything wired up correctly?)
git   = content state     (what changed?)
shell = orchestration
```

---

## Commands

```
dots list  [flags]           show current state of all entries
dots apply [flags]           create symlinks and clone repos
dots add   [flags] [dests…]  add an entry to a manifest
dots help                    show full documentation
```

**Flags (list / apply):**

```
-f file    manifest file, repeatable (default: dots.toml; DOTS_ROOT sets search path)
-t type    filter by type: dotfile, repo
-s state   filter by state (supports !): -s empty, -s '!ok'
-n name    filter by name
-o fields  output fields: name,kind,state,src,url,dest
```

Names piped to stdin are treated as a name filter, one per line.

**Flags (add):**

```
-f file    manifest file, created if absent (-f or DOTS_ROOT required)
-t repo    write a repo entry (default: dotfile)
-r root    src root — dest path appended under root (default: DOTS_SRCROOT)
```

Dest paths are positional arguments or piped via stdin.
For repo entries, the URL is read from the existing git remote at dest.

**Composing manifests:**

```sh
dots list  -f base.toml -f work.toml -f projects.toml
dots apply -f base.toml -f coding.toml
```

## Environment

```
DOTS_ROOT      where manifest files live — resolves relative -f paths and
               sets the default manifest for list/apply
DOTS_SRCROOT   default src root for dots add (equivalent of -r)
```

---

## Install

```sh
go install github.com/philshaughnes/dots@latest
```

Or build from source:

```sh
git clone https://github.com/philshaughnes/dots
cd dots
go build -o dots .
```

---

## Who this is for

Dots is for people who already use git and the shell, and want a simple,
explicit way to track and apply their dotfiles. If you want templating, secrets
management, or a fully managed system, this probably isn't the right tool.

---

Write down what your dotfiles should be.  
Dots helps you check and apply that. Have fun!
