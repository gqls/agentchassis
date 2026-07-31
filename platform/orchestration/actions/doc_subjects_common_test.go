// FILE: platform/orchestration/actions/doc_subjects_common_test.go
//
// Pure-parts tests for the doc-subject vocabulary (bugs_open/064): the
// table-driven gate test over every subject type × both Go gates, and the
// migration-lockstep regression test that fails when the newest migration's
// CHECK and validDocSubjectTypes drift — migration 184 widened the CHECK
// without the Go gates and nothing failed; now this does.

package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestDocSubjectGates_TableDriven(t *testing.T) {
	// Every value the vocabulary carries must pass BOTH Go gates.
	for _, st := range validDocSubjectTypes {
		if _, _, err := docResolveSubject(
			map[string]interface{}{"subject_type": st, "subject_key": "k"},
			map[string]interface{}{}); err != nil {
			t.Errorf("docResolveSubject must accept %q: %v", st, err)
		}
		if reason := docSubjectGateReason(st, "k"); reason != "" {
			t.Errorf("persist gate must accept %q, got skip reason %q", st, reason)
		}
	}
	// Values outside the vocabulary must be rejected by both. The list
	// self-maintains: a value the vocabulary later grows to include (e.g.
	// 'experience-pattern' in the experience-register P2) is skipped here
	// rather than failing the test.
	for _, st := range []string{"", "site", "component", "Tool", "experience-pattern"} {
		if isValidDocSubjectType(st) {
			continue
		}
		if _, _, err := docResolveSubject(
			map[string]interface{}{"subject_type": st, "subject_key": "k"},
			map[string]interface{}{}); err == nil {
			t.Errorf("docResolveSubject must reject %q", st)
		}
		if reason := docSubjectGateReason(st, "k"); reason == "" {
			t.Errorf("persist gate must skip %q", st)
		}
	}
}

func TestDocSubjectGateReason_DistinctReasons(t *testing.T) {
	// The two skip reasons are deliberately distinct (bugs_open/064: an
	// explicit 'experience' subject used to be logged as "no explicit
	// subject" when only its TYPE fell outside a stale allowlist).
	if got := docSubjectGateReason("", ""); got != "no explicit subject" {
		t.Errorf("absent subject: want %q, got %q", "no explicit subject", got)
	}
	if got := docSubjectGateReason("experience", ""); got != "no explicit subject" {
		t.Errorf("empty key must read as absent subject, got %q", got)
	}
	if got := docSubjectGateReason("site", "k"); !strings.Contains(got, `unsupported subject_type "site"`) {
		t.Errorf("unsupported type must name itself, got %q", got)
	}
}

// newestConstraintValues reads the NEWEST migration under
// docs/agent_docs/sql_for_agents that (re)creates the named CHECK constraint
// (doc_plans_subject_type_check or doc_notes_subject_type_check) and returns
// its ARRAY values. Shared by both lockstep tests below — one copy of the
// migration-scanning logic, not two that could themselves drift apart.
func newestConstraintValues(t *testing.T, constraintName string) (file string, values []string, found bool) {
	t.Helper()
	migrationsDir := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("cannot read migrations dir %s (test runs from the package dir; the checkout must include docs/): %v", migrationsDir, err)
	}

	// (?s) because the ADD CONSTRAINT and its CHECK span a newline.
	constraintRE := regexp.MustCompile(`(?s)ADD CONSTRAINT ` + regexp.QuoteMeta(constraintName) + `\s+CHECK \(subject_type = ANY \(ARRAY\[([^\]]+)\]\)\)`)
	valueRE := regexp.MustCompile(`'([a-z_-]+)'`)

	newest := -1
	var newestFile string
	var newestValues []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		numEnd := strings.IndexByte(name, '_')
		if numEnd <= 0 {
			continue
		}
		num, err := strconv.Atoi(name[:numEnd])
		if err != nil || num <= newest {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		m := constraintRE.FindSubmatch(raw)
		if m == nil {
			continue
		}
		var vals []string
		for _, v := range valueRE.FindAllStringSubmatch(string(m[1]), -1) {
			vals = append(vals, v[1])
		}
		newest, newestFile, newestValues = num, name, vals
	}
	if newest < 0 {
		return "", nil, false
	}
	return newestFile, newestValues, true
}

// TestValidDocSubjectTypes_LockstepWithMigrationCheck asserts validDocSubjectTypes
// equals the UNION of the newest doc_plans_subject_type_check and
// doc_notes_subject_type_check values — the dedup-index/Go-list lockstep
// pattern (v1.0.1127) made mechanical for doc subjects.
//
// UNION, not equality-with-either-alone: migration 184 widened doc_plans'
// CHECK without moving the Go vocabulary (bugs_open/064, fixed by the
// original form of this test, which watched doc_plans only). Migration 270
// then widened doc_notes' CHECK the same way — the DB accepted
// subject_type='landmine' on doc_notes for two days while this package's own
// gate rejected it — and this test, watching only doc_plans, had nothing to
// catch it (landmine rows never belong on doc_plans, so no migration ever
// touched that constraint). Found 2026-07-31 by landmine-verifier's first
// live run failing at append_doc_note. The two tables' vocabularies are
// allowed to diverge (doc_notes now accepts 'landmine', doc_plans does not —
// a landmine has no shared-contract shape to put in doc_plans), so
// validDocSubjectTypes is a UNION gate, slightly more permissive than either
// table's own CHECK alone: an out-of-place value passes this Go gate but is
// still caught by the target table's own CHECK at insert time, which is the
// same safe direction of imprecision the original test already accepted
// (a rejected write, never a silently wrong one).
func TestValidDocSubjectTypes_LockstepWithMigrationCheck(t *testing.T) {
	plansFile, plansValues, plansFound := newestConstraintValues(t, "doc_plans_subject_type_check")
	if !plansFound {
		t.Fatal("no migration recreating doc_plans_subject_type_check found — if the constraint moved, update this test's regex")
	}
	notesFile, notesValues, notesFound := newestConstraintValues(t, "doc_notes_subject_type_check")
	if !notesFound {
		t.Fatal("no migration recreating doc_notes_subject_type_check found — if the constraint moved, update this test's regex")
	}

	union := map[string]bool{}
	for _, v := range plansValues {
		union[v] = true
	}
	for _, v := range notesValues {
		union[v] = true
	}
	var got []string
	for v := range union {
		got = append(got, v)
	}

	want := append([]string(nil), validDocSubjectTypes...)
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Fatalf("split contract: validDocSubjectTypes = %v but %s sets doc_plans to %v and %s sets doc_notes to %v (union = %v) — move both together (bugs_open/064; checklist: docs/agent_docs/docs024_key_docs_latest/experience_register/design/subject_type_addition.md)",
			want, plansFile, plansValues, notesFile, notesValues, got)
	}
}
