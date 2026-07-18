// FILE: platform/orchestration/actions/refresh_evidence_base_test.go
//
// Covers the two pieces of V4 that must not be got wrong:
//   - the SQL guard (these queries come from a data column)
//   - tolerance semantics from the REGISTER's point of view, including the
//     under-claiming case the spec calls out
//
// plus whitelist composition, whose contract is "humans own the words".

package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEvidenceQueryGuard(t *testing.T) {
	// Rejected before execution — the guard is a pure string check, so the
	// refusal reasons are testable without a database.
	refused := map[string]string{
		"multi-statement":   "SELECT count(*) FROM sites; DROP TABLE sites",
		"not-select":        "UPDATE sites SET domain = 'x'",
		"cte-wrapped write": "WITH x AS (DELETE FROM sites RETURNING 1) SELECT count(*) FROM x",
		"nested delete":     "SELECT count(*) FROM (DELETE FROM sites RETURNING 1) t",
		"grant":             "SELECT count(*) FROM sites UNION SELECT 1; GRANT ALL ON sites TO public",
	}
	for name, q := range refused {
		if _, err := validateEvidenceQuery(q); err == nil || !strings.HasPrefix(err.Error(), "refused:") {
			t.Errorf("%s: expected a refusal, got %v", name, err)
		}
	}

	// The real fact queries from the live evidence base must all pass, with the
	// trailing semicolon stripped for execution.
	allowed := []string{
		"SELECT count(*) FROM business_intel.businesses WHERE verification_status = 'verified'",
		"SELECT count(*) FROM agent_definitions WHERE deleted_at IS NULL AND status = 'active'",
		"SELECT count(*) FROM content_feed_items WHERE credibility IS NOT NULL;",
	}
	for _, q := range allowed {
		got, err := validateEvidenceQuery(q)
		if err != nil {
			t.Errorf("legitimate SELECT refused: %q → %v", q, err)
			continue
		}
		if strings.HasSuffix(got, ";") {
			t.Errorf("trailing semicolon should be stripped: %q", got)
		}
	}
}

func TestEvidenceValueWithinTolerance(t *testing.T) {
	cases := []struct {
		name             string
		stored, live     float64
		tolerance        string
		wantWithin       bool
		explanationForMe string
	}{
		{"exact unchanged", 2767, 2767, "exact", true, "same number"},
		{"exact moved", 2767, 3104, "exact", false, "copy names a specific figure"},
		{"gte grew", 157, 900, "gte", true, "'more than 150' survives growth"},
		{"gte fell", 157, 120, "gte", false, "copy would now overclaim"},
		{"gte equal", 157, 157, "gte", true, "boundary holds"},
		{"approx within", 12000000, 12500000, "approx_pct:10", true, "4% move"},
		{"approx beyond", 12000000, 15000000, "approx_pct:10", false, "25% move"},
		{"unknown tolerance treated as exact", 100, 101, "wibble", false, "safe default"},
	}
	for _, c := range cases {
		if got := evidenceValueWithinTolerance(c.stored, c.live, c.tolerance); got != c.wantWithin {
			t.Errorf("%s (%s): stored=%v live=%v tol=%q → %v, want %v",
				c.name, c.explanationForMe, c.stored, c.live, c.tolerance, got, c.wantWithin)
		}
	}
}

// The under-claiming case from the spec: the site says 2,767, the database now
// says 3,104. Not a lie — still a finding, because the copy is stale.
func TestUnderclaimingCountsAsDrift(t *testing.T) {
	if evidenceValueWithinTolerance(2767, 3104, "exact") {
		t.Error("an exact-tolerance fact that grew must count as drift (copy is underclaiming but stale)")
	}
}

func TestComposeWriterBlock(t *testing.T) {
	var eb map[string]interface{}
	if err := json.Unmarshal([]byte(`{
		"writer_block_managed": true,
		"facts": [
			{"id":"a","value":2767,"writer_line":"{value} business records verified against Companies House"},
			{"id":"b","value":157,"writer_line":"more than {value} agent definitions in the catalogue. This is a CATALOGUE count — never a running fleet"},
			{"id":"c","kind":"capability","writer_line":"multi-source news pipeline with LLM credibility scoring"},
			{"id":"d","value":42,"claim":"unphrased fact with no writer_line"}
		],
		"allowed_entities": ["Companies House","New Media Age"]
	}`), &eb); err != nil {
		t.Fatalf("fixture: %v", err)
	}

	block := composeWriterBlock(eb)

	// Numbers formatted as copy states them; human caveat preserved verbatim.
	if !strings.Contains(block, "2,767 business records verified") {
		t.Errorf("expected a thousands-separated value, got:\n%s", block)
	}
	if !strings.Contains(block, "This is a CATALOGUE count — never a running fleet") {
		t.Errorf("human-authored caveat must survive regeneration, got:\n%s", block)
	}
	// Value-less facts go under capabilities; unphrased facts are omitted
	// entirely — the machine never invents phrasing.
	if !strings.Contains(block, "CAPABILITIES") || !strings.Contains(block, "credibility scoring") {
		t.Errorf("capability line missing, got:\n%s", block)
	}
	if strings.Contains(block, "unphrased fact") || strings.Contains(block, "42") {
		t.Errorf("a fact with no writer_line must be omitted, got:\n%s", block)
	}
	if !strings.Contains(block, "NAMED ENTITIES you may assert relationships with: Companies House, New Media Age.") {
		t.Errorf("entity list missing, got:\n%s", block)
	}

	// Nothing phrased at all → empty, so the caller leaves the existing block alone.
	var bare map[string]interface{}
	_ = json.Unmarshal([]byte(`{"facts":[{"id":"x","value":1}]}`), &bare)
	if composeWriterBlock(bare) != "" {
		t.Error("with no writer_line anywhere, composition must return empty (leave the human block alone)")
	}
}

func TestFormatEvidenceNumber(t *testing.T) {
	cases := map[float64]string{
		2767: "2,767", 937: "937", 90790: "90,790", 12000000: "12,000,000",
		15: "15", 0: "0", 1234.5: "1234.5",
	}
	for in, want := range cases {
		if got := formatEvidenceNumber(in); got != want {
			t.Errorf("formatEvidenceNumber(%v) = %q, want %q", in, got, want)
		}
	}
}
