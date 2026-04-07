package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	file := flag.String("f", "", "config file (required)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: dots -f machine.toml <command> [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  status   show state of all entries")
		fmt.Fprintln(os.Stderr, "  apply    create symlinks and clone repos")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := loadConfig(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "status":
		os.Exit(cmdStatus(cfg, args[1:]))
	case "apply":
		os.Exit(cmdApply(cfg, args[1:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		flag.Usage()
		os.Exit(1)
	}
}
