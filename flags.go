package main

import (
	"bufio"
	"flag"
	"fmt"
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
	file   string
	types  multiFlag
	states multiFlag
	names  multiFlag
	fields multiFlag
}

// registerFlags registers all common flags onto fs and sets a usage header.
func registerFlags(fs *flag.FlagSet, o *opts) {
	fs.StringVar(&o.file, "f", "", "config file (required)")
	fs.Var(&o.types, "t", "filter by type: dotfile, repo (repeatable)")
	fs.Var(&o.states, "s", "filter by state: ok, empty, blocked, src-missing (repeatable, supports !)")
	fs.Var(&o.names, "n", "filter by name (repeatable)")
	fs.Var(&o.fields, "o", "output fields: name,kind,state,src,url,dest (comma-separated)")
	fs.Usage = func() { printBrief() }
}

// collectNames merges -n flag values and stdin into a name set.
// Returns nil if no names were specified (meaning: match all).
func collectNames(flagNames multiFlag) map[string]bool {
	names := map[string]bool{}
	for _, n := range flagNames {
		names[n] = true
	}
	for _, n := range stdinNames() {
		names[n] = true
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

// stdinNames reads names from stdin when it is a pipe, one per line.
func stdinNames() []string {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice != 0 {
		return nil
	}
	var names []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			names = append(names, line)
		}
	}
	return names
}

// matchType returns true if the entry matches the type filter.
// Multiple values are OR'd. Supports ! negation.
func matchType(s entryStatus, types multiFlag) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		neg := strings.HasPrefix(t, "!")
		val := strings.TrimPrefix(t, "!")
		var match bool
		switch val {
		case "dotfile":
			match = s.kind == "dotfile"
		case "repo":
			match = s.kind == "repo"
		default:
			continue
		}
		if neg {
			match = !match
		}
		if match {
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
