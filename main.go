package main

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed help.txt
var helpText string

func main() {
	if len(os.Args) < 2 {
		printBrief()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "list":
		os.Exit(cmdList(os.Args[2:]))
	case "apply":
		os.Exit(cmdApply(os.Args[2:]))
	case "add":
		os.Exit(cmdAdd(os.Args[2:]))
	case "help":
		fmt.Print(helpText)
		os.Exit(0)
	case "-h", "--help":
		printBrief()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "run 'dots help' for usage")
		os.Exit(1)
	}
}

func printBrief() {
	fmt.Println("Dots is a minimalist dotfile manager that is simple, orthogonal, and composable.")
	fmt.Println()
	fmt.Println("usage: dots <command> [flags] [args]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  list    show current state of all entries")
	fmt.Println("  apply   create symlinks and clone repos")
	fmt.Println("  add     add an entry to a manifest")
	fmt.Println("  help    show full documentation")
	fmt.Println()
	fmt.Println("flags (list / apply):")
	fmt.Println("  -f file    manifest file, repeatable (default: dots.toml; DOTS_ROOT sets search path)")
	fmt.Println("  -t type    filter by type: dotfile, repo (repeatable)")
	fmt.Println("  -s state   filter by state: ok, empty, blocked, src-missing (repeatable, supports !)")
	fmt.Println("  -n name    filter by name (repeatable)")
	fmt.Println("  -o fields  output fields: name,kind,state,src,url,dest (comma-separated)")
	fmt.Println()
	fmt.Println("flags (add):")
	fmt.Println("  -f file    manifest file (default: dots.toml; DOTS_ROOT sets search path)")
	fmt.Println("  -t repo    write a repo entry (default: dotfile)")
	fmt.Println("  -r root    src root — dest appended under root (default: DOTS_SRCROOT)")
}
