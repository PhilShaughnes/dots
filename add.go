package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func cmdAdd(args []string) int {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var file, src, root, name, url string
	var dests multiFlag
	fs.StringVar(&file, "f", "", "manifest file to add to (required)")
	fs.Var(&dests, "D", "dest path (repeatable, or pipe via stdin)")
	fs.StringVar(&src, "S", "", "src path (dotfile, single entry only)")
	fs.StringVar(&root, "R", "", "src root — mirrors dest path structure under root")
	fs.StringVar(&name, "N", "", "entry name (default: dest basename without extension)")
	fs.StringVar(&url, "U", "", "git remote URL (repo, single entry only)")
	fs.Usage = func() { printBrief() }
	fs.Parse(args)

	if file == "" {
		fmt.Fprintln(os.Stderr, "error: -f required")
		fs.Usage()
		return 1
	}

	if len(dests) == 0 {
		for _, n := range stdinNames() {
			dests = append(dests, n)
		}
	}

	if len(dests) == 0 {
		fmt.Fprintln(os.Stderr, "error: -D required (or pipe dest paths via stdin)")
		return 1
	}

	multi := len(dests) > 1

	if src != "" && root != "" {
		fmt.Fprintln(os.Stderr, "error: -S and -R are mutually exclusive")
		return 1
	}
	if url != "" && (src != "" || root != "") {
		fmt.Fprintln(os.Stderr, "error: -U cannot be used with -S or -R")
		return 1
	}
	if multi && src != "" {
		fmt.Fprintln(os.Stderr, "error: -S requires a single -D")
		return 1
	}
	if multi && name != "" {
		fmt.Fprintln(os.Stderr, "error: -N requires a single -D")
		return 1
	}
	if multi && url != "" {
		fmt.Fprintln(os.Stderr, "error: -U requires a single -D")
		return 1
	}

	f, err := os.OpenFile(expandPath(file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening manifest: %v\n", err)
		return 1
	}
	defer f.Close()

	for _, dest := range dests {
		entryName := name
		if entryName == "" {
			entryName = inferName(dest)
		}
		if url != "" {
			fmt.Fprintf(f, "\n[[repos]]\nname = \"%s\"\nurl  = \"%s\"\ndest = \"%s\"\n", entryName, url, dest)
		} else {
			entrySrc := src
			if entrySrc == "" && root != "" {
				entrySrc = mirrorPath(dest, root)
			}
			if entrySrc == "" {
				entrySrc = dest
			}
			fmt.Fprintf(f, "\n[[dotfiles]]\nname = \"%s\"\nsrc  = \"%s\"\ndest = \"%s\"\n", entryName, entrySrc, dest)
		}
	}

	return 0
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

// mirrorPath derives a src path by mirroring dest's structure under root.
// ~/.zshrc with root ~/dotfiles/ → ~/dotfiles/.zshrc
func mirrorPath(dest, root string) string {
	var rel string
	if strings.HasPrefix(dest, "~/") {
		rel = dest[2:]
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			rel = strings.TrimPrefix(expandPath(dest), home)
			rel = strings.TrimPrefix(rel, "/")
		} else {
			rel = filepath.Base(dest)
		}
	}
	return strings.TrimSuffix(root, "/") + "/" + rel
}
