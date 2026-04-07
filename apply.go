package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdApply(cfg *Config, args []string) int {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	fs.Parse(args)

	filter := map[string]bool{}
	for _, a := range fs.Args() {
		filter[a] = true
	}

	statuses := collectStatuses(cfg, filter)
	exit := 0

	for _, s := range statuses {
		switch s.kind {
		case "file":
			if err := applyDotfile(s); err != nil {
				fmt.Fprintf(os.Stderr, "  error %s: %v\n", s.name, err)
				exit = 1
			}
		case "repo":
			if err := applyRepo(cfg, s); err != nil {
				fmt.Fprintf(os.Stderr, "  error %s: %v\n", s.name, err)
				exit = 1
			}
		}
	}

	return exit
}

func applyDotfile(s entryStatus) error {
	switch s.state {
	case stateOK:
		fmt.Printf("  ok       %s\n", s.name)
		return nil
	case stateConflict:
		fmt.Printf("  conflict %s → %s (real file exists, skipping)\n", s.name, s.dest)
		return nil
	case stateWrongTarget:
		fmt.Printf("  conflict %s → %s (symlink points elsewhere, skipping)\n", s.name, s.dest)
		return nil
	case stateSrcMissing:
		return fmt.Errorf("src %s does not exist", s.src)
	case stateMissing:
		if err := os.MkdirAll(filepath.Dir(s.dest), 0755); err != nil {
			return err
		}
		fmt.Printf("  linking  %s → %s\n", s.src, s.dest)
		return os.Symlink(s.src, s.dest)
	}
	return nil
}

func applyRepo(cfg *Config, s entryStatus) error {
	switch s.state {
	case stateCloned:
		fmt.Printf("  ok       %s\n", s.name)
		return nil
	case stateConflict:
		fmt.Printf("  conflict %s → %s (exists but not a repo, skipping)\n", s.name, s.dest)
		return nil
	case stateNotCloned:
		// find url from cfg
		for _, r := range cfg.Repos {
			if r.Name == s.name {
				dest := expandPath(r.Dest)
				fmt.Printf("  cloning  %s → %s\n", r.URL, dest)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return err
				}
				cmd := exec.Command("git", "clone", r.URL, dest)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}
		}
	}
	return nil
}
