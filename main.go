package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
)

//go:embed help.txt
var helpText string

func main() {
	if len(os.Args) < 2 {
		printBrief(os.Stderr)
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
		printBrief(os.Stdout)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "run 'dots help' for usage")
		os.Exit(1)
	}
}

// printBrief writes the short usage to w: stdout when asked for (-h),
// stderr on errors so usage never lands in a pipeline.
func printBrief(w io.Writer) {
	fmt.Fprintln(w, "Dots is a minimalist dotfile manager that is simple, orthogonal, and composable.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "usage: dots <command> [flags] [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  list    show current state of all entries")
	fmt.Fprintln(w, "  apply   create symlinks and clone repos")
	fmt.Fprintln(w, "  add     add an entry to a manifest")
	fmt.Fprintln(w, "  help    show full documentation")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "flags (list / apply):")
	fmt.Fprintln(w, "  -f file    manifest file, repeatable (default: dots.toml; DOTS_ROOT sets search path)")
	fmt.Fprintln(w, "  -t type    filter by type: dotfile, repo (repeatable)")
	fmt.Fprintln(w, "  -s state   filter by state: ok, empty, blocked, src-missing (repeatable, supports !)")
	fmt.Fprintln(w, "  -n name    filter by name (repeatable, - reads names from stdin)")
	fmt.Fprintln(w, "  -o fields  output fields: name,kind,state,src,url,dest (comma-separated)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "flags (add):")
	fmt.Fprintln(w, "  -f file    manifest file (default: dots.toml; DOTS_ROOT sets search path)")
	fmt.Fprintln(w, "  -t repo    write a repo entry (default: dotfile)")
	fmt.Fprintln(w, "  -r root    src root — dest appended under root (default: DOTS_SRCROOT)")
}
