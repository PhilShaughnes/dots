package main

import "testing"

func TestMatchType(t *testing.T) {
	file := entryStatus{kind: "dotfile"}
	repo := entryStatus{kind: "repo"}

	tests := []struct {
		name  string
		entry entryStatus
		types multiFlag
		want  bool
	}{
		{"no filter matches dotfile",  file, nil,                   true},
		{"no filter matches repo",     repo, nil,                   true},
		{"repo matches repo",          repo, multiFlag{"repo"},     true},
		{"repo no match dotfile",      file, multiFlag{"repo"},     false},
		{"dotfile matches dotfile",    file, multiFlag{"dotfile"},  true},
		{"dotfile no match repo",      repo, multiFlag{"dotfile"},  false},
		{"!repo matches dotfile",      file, multiFlag{"!repo"},    true},
		{"!repo no match repo",        repo, multiFlag{"!repo"},    false},
		{"multi OR: repo or dotfile",  file, multiFlag{"repo", "dotfile"}, true},
		{"unknown type no match",      file, multiFlag{"unknown"},  false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchType(tt.entry, tt.types); got != tt.want {
				t.Errorf("matchType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchState(t *testing.T) {
	tests := []struct {
		name   string
		state  entryState
		states multiFlag
		want   bool
	}{
		{"no filter matches any",            stateOK,          nil,                          true},
		{"ok matches ok",                    stateOK,          multiFlag{"ok"},              true},
		{"ok* matches ok",                   stateChanged,     multiFlag{"ok"},              true},
		{"ok no match empty",                stateEmpty,       multiFlag{"ok"},              false},
		{"empty matches empty",              stateEmpty,       multiFlag{"empty"},           true},
		{"blocked matches blocked",          stateBlocked,     multiFlag{"blocked"},         true},
		{"src-missing matches src-missing",  stateSrcMissing,  multiFlag{"src-missing"},     true},
		{"!ok excludes ok",                  stateOK,          multiFlag{"!ok"},             false},
		{"!ok excludes ok*",                 stateChanged,     multiFlag{"!ok"},             false},
		{"!ok includes empty",               stateEmpty,       multiFlag{"!ok"},             true},
		{"!ok includes blocked",             stateBlocked,     multiFlag{"!ok"},             true},
		{"multi OR: empty or blocked",       stateEmpty,       multiFlag{"empty", "blocked"}, true},
		{"multi OR: no match ok",            stateOK,          multiFlag{"empty", "blocked"}, false},
		{"negation excludes explicitly",     stateBlocked,     multiFlag{"!blocked"},        false},
		{"negation includes others",         stateEmpty,       multiFlag{"!blocked"},        true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchState(tt.state, tt.states); got != tt.want {
				t.Errorf("matchState(%s, %v) = %v, want %v", tt.state, tt.states, got, tt.want)
			}
		})
	}
}
