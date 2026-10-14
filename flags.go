package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// multiFlag accepts repeatable flags (-t repo -t dotfile) and comma-separated values (-o name,dest).
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	for _, s := range strings.Split(v, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			*m = append(*m, s)
		}
	}
	return nil
}

// opts holds the parsed flags common to all commands.
type opts struct {
	files  multiFlag
	types  multiFlag
	states multiFlag
	names  multiFlag
	fields multiFlag
}

// registerFlags registers all common flags onto fs and sets a usage header.
func registerFlags(fs *flag.FlagSet, o *opts) {
	fs.Var(&o.files, "f", "config file, repeatable: -f base.toml -f work.toml (required)")
	fs.Var(&o.types, "t", "filter by type: dotfile, repo (repeatable)")
	fs.Var(&o.states, "s", "filter by state: ok, empty, blocked, src-missing (repeatable, supports !)")
	fs.Var(&o.names, "n", "filter by name (repeatable, - reads names from stdin)")
	fs.Var(&o.fields, "o", "output fields: name,kind,state,src,url,dest (comma-separated)")
	fs.Usage = func() { printBrief() }
}

// collectNames merges -n flag values into a name set. "-n -" reads names
// from stdin; stdin is never read otherwise.
// Returns nil if no -n was given (meaning: match all). A non-nil empty set
// matches nothing, so an empty pipe into "-n -" selects no entries.
func collectNames(flagNames multiFlag) map[string]bool {
	if len(flagNames) == 0 {
		return nil
	}
	names := map[string]bool{}
	for _, n := range flagNames {
		if n == "-" {
			for _, line := range readLines(os.Stdin) {
				names[line] = true
			}
			continue
		}
		names[n] = true
	}
	return names
}

// readLines reads non-empty, trimmed lines from r.
func readLines(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// matchType returns true if the entry matches the type filter.
// Multiple values are OR'd.
func matchType(s entryStatus, types multiFlag) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if t == s.kind {
			return true
		}
	}
	return false
}

// matchState returns true if the state matches the filter.
// "ok" matches both stateOK and stateChanged (ok*). Supports ! negation.
// Positive values are OR'd; negations exclude regardless.
func matchState(state entryState, states multiFlag) bool {
	if len(states) == 0 {
		return true
	}
	// ok* is treated as ok for filtering
	stateStr := state.String()
	if stateStr == "ok*" {
		stateStr = "ok"
	}
	hasPositive := false
	for _, s := range states {
		if !strings.HasPrefix(s, "!") {
			hasPositive = true
		}
	}
	for _, s := range states {
		neg := strings.HasPrefix(s, "!")
		val := strings.TrimPrefix(s, "!")
		if neg && stateStr == val {
			return false
		}
		if !neg && stateStr == val {
			return true
		}
	}
	return !hasPositive
}

// outputEntry prints an entry. No -o: fixed-width table. With -o: tab-separated fields.
func outputEntry(s entryStatus, fields multiFlag) {
	if len(fields) == 0 {
		fmt.Printf("%-20s %-8s %-12s %s\n", s.name, s.kind, s.state, s.dest)
		return
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		switch f {
		case "name":
			parts[i] = s.name
		case "kind":
			parts[i] = s.kind
		case "state":
			parts[i] = s.state.String()
		case "src":
			parts[i] = s.src
		case "url":
			parts[i] = s.url
		case "dest":
			parts[i] = s.dest
		}
	}
	fmt.Println(strings.Join(parts, "\t"))
}
