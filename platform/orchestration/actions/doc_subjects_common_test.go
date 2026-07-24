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

// TestValidDocSubjectTypes_LockstepWithMigrationCheck reads the NEWEST
// migration under docs/agent_docs/sql_for_agents that (re)creates
// doc_plans_subject_type_check and asserts its ARRAY values equal
// validDocSubjectTypes — the dedup-index/Go-list lockstep pattern
// (v1.0.1127) made mechanical for doc subjects. A migration that widens the
// CHECK without moving the Go vocabulary (184's failure mode) now fails the
// build gate instead of shipping a split contract.
func TestValidDocSubjectTypes_LockstepWithMigrationCheck(t *testing.T) {
	migrationsDir := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("cannot read migrations dir %s (test runs from the package dir; the checkout must include docs/): %v", migrationsDir, err)
	}

	// (?s) because the ADD CONSTRAINT and its CHECK span a newline.
	constraintRE := regexp.MustCompile(`(?s)ADD CONSTRAINT doc_plans_subject_type_check\s+CHECK \(subject_type = ANY \(ARRAY\[([^\]]+)\]\)\)`)
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
		var values []string
		for _, v := range valueRE.FindAllStringSubmatch(string(m[1]), -1) {
			values = append(values, v[1])
		}
		newest, newestFile, newestValues = num, name, values
	}
	if newest < 0 {
		t.Fatal("no migration recreating doc_plans_subject_type_check found — if the constraint moved, update this test's regex")
	}

	want := append([]string(nil), validDocSubjectTypes...)
	got := append([]string(nil), newestValues...)
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Fatalf("split contract: validDocSubjectTypes = %v but %s sets the CHECK to %v — move both together (bugs_open/064; checklist: docs/agent_docs/docs024_key_docs_latest/experience_register/design/subject_type_addition.md)",
			want, newestFile, got)
	}
}
