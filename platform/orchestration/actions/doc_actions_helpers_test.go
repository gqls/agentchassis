// FILE: platform/orchestration/actions/doc_actions_helpers_test.go
//
// DRAFT for the agent-chassis repo — pure-parts tests for the travelling-docs
// action helpers, in the same mould as diagnose_route_resolver_test.go (no
// DB/HTTP; the wired paths are exercised by the actions in production).
// Covers: docResolveSubject, docCategoriesJSON, docCategoriesJSONFromList,
// extractCriteriaBlock, firstNonEmptyDoc.

package actions

import (
	"strings"
	"testing"
)

func TestDocResolveSubject_DirectAndFieldPath(t *testing.T) {
	// direct config key wins
	st, sk, err := docResolveSubject(
		map[string]interface{}{"subject_type": "tool", "subject_key": "tool-vat-calculator"},
		map[string]interface{}{})
	if err != nil || st != "tool" || sk != "tool-vat-calculator" {
		t.Fatalf("direct resolution wrong: %v %q %q", err, st, sk)
	}
	// field path fallback
	st, sk, err = docResolveSubject(
		map[string]interface{}{"subject_type": "pipeline", "subject_key_field": "input_data.subject_key"},
		map[string]interface{}{"input_data": map[string]interface{}{"subject_key": "build"}})
	if err != nil || st != "pipeline" || sk != "build" {
		t.Fatalf("field-path resolution wrong: %v %q %q", err, st, sk)
	}
}

func TestDocResolveSubject_RejectsBadTypeAndMissingKey(t *testing.T) {
	if _, _, err := docResolveSubject(
		map[string]interface{}{"subject_type": "site", "subject_key": "x"},
		map[string]interface{}{}); err == nil {
		t.Fatal("subject_type outside validDocSubjectTypes must error")
	}
	// 'experience' is accepted (migration 163; experience loop). The full
	// vocabulary × gate matrix lives in doc_subjects_common_test.go.
	if st, sk, err := docResolveSubject(
		map[string]interface{}{"subject_type": "experience", "subject_key": "vonc-spark-game"},
		map[string]interface{}{}); err != nil || st != "experience" || sk != "vonc-spark-game" {
		t.Fatalf("experience subject must resolve: %v %q %q", err, st, sk)
	}
	if _, _, err := docResolveSubject(
		map[string]interface{}{"subject_type": "tool"},
		map[string]interface{}{}); err == nil {
		t.Fatal("missing subject_key must error")
	}
}

func TestDocCategoriesJSON_Sources(t *testing.T) {
	// direct config list takes precedence
	got := docCategoriesJSON(
		map[string]interface{}{"note_categories": []interface{}{"diagnosis", "seam"}},
		map[string]interface{}{})
	if got != `["diagnosis","seam"]` {
		t.Fatalf("direct list wrong: %s", got)
	}
	// collected-data field, non-strings skipped
	got = docCategoriesJSON(
		map[string]interface{}{"note_categories_field": "cats"},
		map[string]interface{}{"cats": []interface{}{"migration", 42, ""}})
	if got != `["migration"]` {
		t.Fatalf("field list wrong: %s", got)
	}
	// single string coerces to one-element array
	got = docCategoriesJSON(
		map[string]interface{}{"note_categories_field": "cats"},
		map[string]interface{}{"cats": "acceptance-fail"})
	if got != `["acceptance-fail"]` {
		t.Fatalf("string coercion wrong: %s", got)
	}
	// absent -> empty array (jsonb-safe)
	got = docCategoriesJSON(map[string]interface{}{}, map[string]interface{}{})
	if got != `[]` {
		t.Fatalf("absent must be []: %s", got)
	}
}

func TestDocCategoriesJSONFromList_Quoting(t *testing.T) {
	got := docCategoriesJSONFromList([]string{"diagnosis", "unconfirmed-diagnosis"})
	if got != `["diagnosis","unconfirmed-diagnosis"]` {
		t.Fatalf("quoting wrong: %s", got)
	}
	if got := docCategoriesJSONFromList(nil); got != `[]` {
		t.Fatalf("nil must be []: %s", got)
	}
}

func TestExtractCriteriaBlock(t *testing.T) {
	body := "# PLAN\n\n## Acceptance criteria\n```criteria\n{ \"checks\": [] }\n```\n\n## Delivery mechanism\n"
	got := extractCriteriaBlock(body)
	if got != `{ "checks": [] }` {
		t.Fatalf("extraction wrong: %q", got)
	}
	// absent fence -> empty
	if got := extractCriteriaBlock("# PLAN with no criteria"); got != "" {
		t.Fatalf("absent fence must be empty, got %q", got)
	}
	// unclosed fence -> empty (never return a half block)
	if got := extractCriteriaBlock("```criteria\n{ \"checks\": ["); got != "" {
		t.Fatalf("unclosed fence must be empty, got %q", got)
	}
	// only the FIRST criteria block is read
	two := "```criteria\nFIRST\n```\ntext\n```criteria\nSECOND\n```"
	if got := extractCriteriaBlock(two); !strings.Contains(got, "FIRST") || strings.Contains(got, "SECOND") {
		t.Fatalf("must return first block only, got %q", got)
	}
}

func TestFirstNonEmptyDoc(t *testing.T) {
	if got := firstNonEmptyDoc("", "second", "third"); got != "second" {
		t.Fatalf("want second, got %q", got)
	}
	if got := firstNonEmptyDoc("", ""); got != "" {
		t.Fatalf("all empty must be empty, got %q", got)
	}
}
