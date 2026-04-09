package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdApply(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	var o opts
	registerFlags(fs, &o)
	fs.Parse(args)

	if o.file == "" {
		fmt.Fprintln(os.Stderr, "error: -f required")
		fs.Usage()
		return 1
	}

	cfg, err := loadConfig(o.file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		return 1
	}

	names := collectNames(o.names)
	exit := 0
	for _, s := range collectStatuses(cfg, names) {
		if !matchType(s, o.types) || !matchState(s.state, o.states) {
			continue
		}
		var newState entryState
		var err error
		switch s.kind {
		case "dotfile":
			newState, err = applyDotfile(s)
		case "repo":
			newState, err = applyRepo(s)
		default:
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error %s: %v\n", s.name, err)
			exit = 1
			continue
		}
		s.state = newState
		outputEntry(s, o.fields)
	}
	return exit
}

// applyDotfile creates a symlink for the entry if needed.
// Returns the resulting state and any error.
func applyDotfile(s entryStatus) (entryState, error) {
	switch s.state {
	case stateOK:
		return stateOK, nil
	case stateBlocked, stateSrcMissing:
		return s.state, nil
	case stateEmpty:
		if err := os.MkdirAll(filepath.Dir(s.dest), 0o755); err != nil {
			return stateBlocked, err
		}
		if err := os.Symlink(s.src, s.dest); err != nil {
			return stateBlocked, err
		}
		return stateChanged, nil
	}
	return stateOK, nil
}

// applyRepo clones the repo if not already present.
// Returns the resulting state and any error.
func applyRepo(s entryStatus) (entryState, error) {
	switch s.state {
	case stateOK:
		return stateOK, nil
	case stateBlocked:
		return stateBlocked, nil
	case stateEmpty:
		if err := os.MkdirAll(filepath.Dir(s.dest), 0o755); err != nil {
			return stateBlocked, err
		}
		cmd := exec.Command("git", "clone", s.url, s.dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return stateBlocked, err
		}
		return stateChanged, nil
	}
	return stateOK, nil
}
