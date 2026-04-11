# Dots - The Simple Dotfile Manager

Ohi! Welcome to Dots!  
It's for dotfiles.

Dotfiles are awesome.  
They're where you write yourself cheatsheets,  
tune things just so, hack, explore, play.  
They make your system yours.

They also need a way to be backed up, hacked on, moved around, and kept track of.

_This is my take on that._

---

**A few principles I wanted:**

1. **Simple** — clear and understandable, no magic, no hidden conventions. Less is more.
2. **Trackable** — a declarative manifest of what dotfiles you have and where they live. Dotfiles are scattered — that's fine. Keep a list.
3. **Hackable** — open a file, edit it. Use git, diff, your editor, whatever you already use. No special workflow required.
4. **Flexible** — no opinions on where files are stored, where the manifest lives, or where things are going. Files or whole git repos — dots handles both.
5. **Composable** — one manifest per machine, or a base plus layers. Mix however makes sense for your setup.
6. **Safe** — don't break things. Make it clear what is and what should be, then let you handle it.

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

Create a manifest (doesn't matter where!):

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

If your dotfiles already exist at the dest locations, you'll see them as `blocked`:

```
zshrc    dotfile  blocked  /home/you/.zshrc
jj       dotfile  blocked  /home/you/.config/jj/config.toml
nvim     repo     empty    /home/you/.config/nvim
```

That's because the files are already there — dots won't overwrite them. Move them
to the src locations first, then list again:

```
zshrc    dotfile  empty    /home/you/.zshrc
jj       dotfile  empty    /home/you/.config/jj/config.toml
nvim     repo     empty    /home/you/.config/nvim
```

See the [examples](#but-can-it-do) section for a script to move blocked files automatically.

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

Names piped to stdin are used as a name filter, one per line:

```sh
dots list -f machine.toml -o name | fzf --multi | dots apply -f machine.toml
```

Run `dots help` or `dots -h` for full flag reference.

---

## But can it do...?

Dots is minimal and orthogonal by design. That usually means the answer is *yes, with a simple shell pipe.*
Let me show you:


**Migrate existing dotfiles into your repo, then apply:**
```sh
dots list -f machine.toml -t dotfile -s blocked -o src,dest | while IFS=$'\t' read -r src dest; do
  mv "$dest" "$src"
done
dots apply -f machine.toml
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

The filtering + shell pipeline pattern is intentionally the primitive. If you can express it as a query on your manifest, you can script it.

---

Dots sets things up and helps you see what's what. The rest is your toolchain doing what it's good at. Have fun!

