package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "list":
		os.Exit(cmdList(os.Args[2:]))
	case "apply":
		os.Exit(cmdApply(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: dots <command> [flags] [names...]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  list    show entry states")
	fmt.Fprintln(os.Stderr, "  apply   create symlinks and clone repos")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "flags:")
	fmt.Fprintln(os.Stderr, "  -f file    config file (required)")
	fmt.Fprintln(os.Stderr, "  -t type    filter by type: dotfile, repo (repeatable)")
	fmt.Fprintln(os.Stderr, "  -s state   filter by state: ok, empty, blocked, src-missing (repeatable, supports !)")
	fmt.Fprintln(os.Stderr, "  -n name    filter by name (repeatable)")
	fmt.Fprintln(os.Stderr, "  -o fields  output fields: name,kind,state,src,url,dest (comma-separated)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "stdin is read as names when piped")
}
