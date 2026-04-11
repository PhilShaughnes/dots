# dots

Ohi! Welcome to dots — a dotfile manager.

Dotfiles make your system yours. They're where you write yourself cheatsheets,
tune things just so, hack, explore, play, express yourself.

They also need to be backed up, hacked on, moved around, and kept track of.
This is my take on that.

Most dotfile managers assume you want the same thing everywhere, with templates
handling the differences. That never fit how I actually work.

**A few principles I wanted:**

1. **Simple** — clear and understandable, no magic, no hidden conventions. Less is more.
2. **Trackable** — a declarative manifest of what dotfiles you have and where they live. Dotfiles are scattered — that's fine. Keep a list.
3. **Hackable** — open a file, edit it. Use git, diff, your editor, whatever you already use. No special workflow required.
4. **Flexible** — no opinions on where files are stored, where the manifest lives, or where things are going. Files or whole git repos — dots handles both.
5. **Composable** — one manifest per machine, or a base plus layers. Mix however makes sense for your setup.
6. **Safe** — don't break things. Make it clear what is and what should be, then let you handle it.

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

## Quick start

Create a manifest:

```toml
# machine.toml

[[dotfiles]]
name = "zshrc"
src  = "~/dotfiles/.zshrc"
dest = "~/.zshrc"

[[dotfiles]]
name = "jj"
src  = "~/dotfiles/.config/jj/config.toml"
dest = "~/.config/jj/config.toml"

[[repos]]
name = "nvim"
url  = "git@github.com:you/nvim.git"
dest = "~/.config/nvim"
```

Check the current state:

```sh
dots list -f machine.toml
```

```
zshrc    dotfile  empty    /home/you/.zshrc
jj       dotfile  empty    /home/you/.config/jj/config.toml
nvim     repo     empty    /home/you/.config/nvim
```

Apply it:

```sh
dots apply -f machine.toml
```

```
zshrc    dotfile  ok*      /home/you/.zshrc
jj       dotfile  ok*      /home/you/.config/jj/config.toml
nvim     repo     ok*      /home/you/.config/nvim
```

---

## Config

A manifest is a TOML file with `[[dotfiles]]` and `[[repos]]` sections.
Put it wherever makes sense — your dotfiles repo, `~/.config/dots/`, anywhere.

```toml
[[dotfiles]]
name = "zshrc"               # unique name, used for filtering
src  = "~/dotfiles/.zshrc"   # source path in your dotfiles repo
dest = "~/.zshrc"            # where the symlink is created

[[repos]]
name = "nvim"
url  = "git@github.com:you/nvim.git"   # git remote
dest = "~/.config/nvim"                # where to clone
```

Both `src` and `dest` accept `~` and absolute paths. Names must be unique
across all manifests used together.

**Composing manifests:**

```sh
dots list  -f base.toml -f work.toml -f projects.toml
dots apply -f base.toml -f coding.toml
```

---

## Commands

```
dots list  -f file [flags]   show current state of all entries
dots apply -f file [flags]   create symlinks and clone repos
dots help                    show full documentation
```

**States:**

| State | Meaning |
|---|---|
| `ok` | symlink correct or repo present |
| `ok*` | just fixed by this apply run |
| `empty` | nothing at dest — safe to apply |
| `blocked` | dest exists but wrong — needs manual intervention |
| `src-missing` | source path missing from dotfiles repo |

`apply` skips `blocked` and `src-missing` entries without error.
`list` exits non-zero if any entry is not `ok`.

**Flags:**

```
-f file    config file, repeatable
-t type    filter by type: dotfile, repo
-s state   filter by state (supports !): -s empty, -s '!ok'
-n name    filter by name
-o fields  output fields: name,kind,state,src,url,dest
```

Run `dots help` or `dots -h` for full flag reference.

---

## Examples

**Show only problems:**
```sh
dots list -f machine.toml -s '!ok'
```

**Apply only what's safe:**
```sh
dots apply -f machine.toml -s empty
```

**Apply only repos:**
```sh
dots apply -f machine.toml -t repo
```

**Show blocked entries with their paths:**
```sh
dots list -f machine.toml -s blocked -o name,dest
```

**Interactively choose what to apply:**
```sh
dots list -f machine.toml -o name | fzf --multi | dots apply -f machine.toml
```

**Check all repos for uncommitted changes:**
```sh
dots list -f machine.toml -t repo -o dest | while read -r repo; do
  echo "== $repo =="
  git -C "$repo" status --short
done
```

**Auto-pull all repos:**
```sh
dots list -f machine.toml -t repo -s ok -o dest | while read -r repo; do
  git -C "$repo" pull --rebase
done
```

**Check state after applying:**
```sh
dots apply -f machine.toml && dots list -f machine.toml -s '!ok'
```

---

## How it compares

| Tool | Strengths | Weaknesses vs dots |
|---|---|---|
| **chezmoi** | Templating, host-aware configs, secret encryption, reconciliation | Implicit path conventions, complex add/edit workflow, opinionated templating layer |
| **GNU Stow** | Pure symlink manager, handles directories elegantly | No declarative manifest, directory-structure-only, harder to filter or compose per-machine |
| **Bare scripts + git** | Maximal flexibility | No declarative overview, you handle errors, filtering, and blocked states yourself |
| **Homesick / vcsh / yadm** | Git-centric, some host awareness | Opinionated repo layout, less granular filtering, less shell-composable |
| **Nix Home Manager** | Full declarative system, clean per-machine handling | Heavyweight, requires learning Nix, breaks filesystem-first intuition |

---

Dots sets things up and helps you see what's what. The rest is your toolchain.
