// FILE: platform/orchestration/datahelpers/tombstones_test.go
//
// Two halves, matching the two ways the tombstone contract can rot:
//
//  1. The VALUE test pins the constant itself to the assembler's NULL-safe
//     spelling. This is the single mutation point: change the const to a bare
//     inequality and this fails, loudly, with the reason.
//  2. The NEGATIVE SCAN walks platform/orchestration and fails on any
//     non-comment, non-test hand-spelling of the clause — NULL-safe or not,
//     because a correct copy is still a copy that can drift (council
//     89dcc04a round 1, reuse_agent: pairwise lockstep tests police
//     duplication; a shared constant makes it unrepresentable, and this scan
//     is what keeps the constant the only spelling).
//
// livespec was read before writing this (the round-1 reuse seat's ask): its
// own header scopes it to "a property of a live DB object" probed by the
// live-audit CronJob — Go-source consistency is out of its remit, so a
// package-local scan it is. Comment lines are skipped: a source scan must
// never let prose satisfy or fail it.

package datahelpers

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNotRemovedSQLIsNullSafe(t *testing.T) {
	if NotRemovedSQL != `build_status IS DISTINCT FROM 'removed'` {
		t.Fatalf("NotRemovedSQL = %q — the tombstone exclusion must stay NULL-safe (IS DISTINCT FROM): build_status is nullable, and a bare inequality makes a NULL-status row served-but-invisible to every consumer of this predicate", NotRemovedSQL)
	}
	if got := NotRemoved("pc"); got != `pc.build_status IS DISTINCT FROM 'removed'` {
		t.Fatalf("NotRemoved(\"pc\") = %q", got)
	}
	if got := NotRemoved(""); got != NotRemovedSQL {
		t.Fatalf("NotRemoved(\"\") = %q, want the bare constant", got)
	}
}

// hand-spellings in either polarity; matching on the comparison tail keeps the
// scan honest about aliases and whitespace variants it has not foreseen.
var handSpelled = []string{
	`build_status <> 'removed'`,
	`build_status != 'removed'`,
	`build_status IS DISTINCT FROM 'removed'`,
}

func TestNoHandSpelledTombstonePredicate(t *testing.T) {
	root := ".." // platform/orchestration — every known consumer lives here
	var offenders []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if filepath.Base(path) == "tombstones.go" {
			return nil // the constant's home is the one permitted spelling
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, needle := range handSpelled {
				if strings.Contains(line, needle) {
					offenders = append(offenders, path+":"+strconv.Itoa(i+1)+" hand-spells "+needle)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(offenders) > 0 {
		t.Errorf("hand-spelled tombstone predicates found — use datahelpers.NotRemoved/NotRemovedSQL so the clause cannot drift from the assembler's:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}
