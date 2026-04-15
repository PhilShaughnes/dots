package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func cmdAdd(args []string) int {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var file, root, kind string
	fs.StringVar(&file, "f", "", "manifest file to add to (required)")
	fs.StringVar(&root, "r", os.Getenv("DOTS_SRCROOT"), "src root — dest path appended under root")
	fs.StringVar(&kind, "t", "", "entry type: dotfile (default), repo")
	fs.Usage = func() { printBrief() }
	fs.Parse(args)

	if file == "" {
		if os.Getenv("DOTS_ROOT") == "" {
			fmt.Fprintln(os.Stderr, "error: -f required (or set DOTS_ROOT)")
			fs.Usage()
			return 1
		}
		file = resolveManifestPath("dots.toml")
	}

	dests := fs.Args()
	if len(dests) == 0 {
		dests = stdinNames()
	}
	if len(dests) == 0 {
		fmt.Fprintln(os.Stderr, "error: dest path required (argument or stdin)")
		return 1
	}

	f, err := os.OpenFile(expandPath(file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening manifest: %v\n", err)
		return 1
	}
	defer f.Close()

	for _, dest := range dests {
		entrySrc := root
		if entrySrc != "" {
			entrySrc = mirrorPath(dest, root)
		}
		dest = normalizePath(dest)
		entryName := inferName(dest)

		if kind == "repo" {
			url := repoURL(dest)
			fmt.Fprintf(f, "\n[[repos]]\nname = \"%s\"\nurl  = \"%s\"\ndest = \"%s\"\n", entryName, url, dest)
		} else {
			if entrySrc == "" {
				entrySrc = dest
			}
			fmt.Fprintf(f, "\n[[dotfiles]]\nname = \"%s\"\nsrc  = \"%s\"\ndest = \"%s\"\n", entryName, entrySrc, dest)
		}
	}

	return 0
}

// repoURL tries to read the origin remote URL from an existing repo at dest.
// Falls back to a placeholder if the dest doesn't exist or has no remote.
func repoURL(dest string) string {
	out, err := exec.Command("git", "-C", expandPath(dest), "remote", "get-url", "origin").Output()
	if err != nil {
		return "<remote-url>"
	}
	return strings.TrimSpace(string(out))
}

// inferName derives an entry name from the dest path:
// basename, strip leading dot, strip extension.
func inferName(dest string) string {
	base := filepath.Base(dest)
	base = strings.TrimPrefix(base, ".")
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "" {
		base = filepath.Base(dest)
	}
	return base
}

// mirrorPath derives a src path by appending dest under root.
// dest is used as-given: ~/ and absolute paths strip the home prefix;
// relative paths are used directly, giving the caller control over depth.
//
//	~/.zshrc          + ~/dotfiles → ~/dotfiles/.zshrc
//	alacritty/a.toml  + ~/dotfiles → ~/dotfiles/alacritty/a.toml
func mirrorPath(dest, root string) string {
	var rel string
	switch {
	case strings.HasPrefix(dest, "~/"):
		rel = dest[2:]
	case strings.HasPrefix(dest, "/"):
		home, err := os.UserHomeDir()
		if err == nil {
			rel = strings.TrimPrefix(dest, home+"/")
		} else {
			rel = filepath.Base(dest)
		}
	default:
		rel = dest
	}
	return strings.TrimSuffix(root, "/") + "/" + rel
}

// normalizePath converts a relative path to an absolute one,
// then converts paths under home to ~/… form for portability.
func normalizePath(path string) string {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "/") {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return abs
	}
	if strings.HasPrefix(abs, home+"/") {
		return "~/" + abs[len(home)+1:]
	}
	return abs
}
