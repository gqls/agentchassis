// FILE: cmd/config-key-audit/rendertruncation_test.go
//
// Every arm of the RENDER_AUDIT_TRUNCATED reader, proven by INDUCING the fault.
// A reader that has only ever been run against a healthy fixture is a reader
// nobody knows the alarm condition of — which is how the signal it consumes came
// to sit unread in the first place (bugs_open/394).
package main

import (
	"os"
	"strings"
	"testing"
)

func acks(keys ...string) map[string]bool {
	m := map[string]bool{}
	for _, k := range keys {
		m[k] = true
	}
	return m
}

// Rows arrive NEWEST FIRST per (domain, agent_type) — the query's ORDER BY — and
// arm 2 depends on it. These fixtures keep that order.
func TestJudgeRenderTruncation_HealthyCursorIsNotAFinding(t *testing.T) {
	runs := []renderTruncationRun{
		{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "tool-md5", OccurredAt: "2026-08-29"},
		{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "tool-html-minifier", OccurredAt: "2026-08-26"},
	}
	if got := judgeRenderTruncation(runs, nil); len(got) != 0 {
		t.Fatalf("an advancing cursor is healthy pagination, not a finding: %+v", got)
	}
}

// ARM 1. The rotating caller must rotate.
// MUTATION: drop the rotatingCallers membership test → this goes green and a
// permanent tail is invisible again.
func TestJudgeRenderTruncation_PrefixFromTheRotatingCallerIsAFinding(t *testing.T) {
	for _, mode := range []string{"prefix", "(absent)"} {
		runs := []renderTruncationRun{
			{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: mode, OccurredAt: "2026-08-29"},
		}
		got := judgeRenderTruncation(runs, nil)
		if len(got) != 1 || got[0].Arm != "prefix_from_rotating_caller" {
			t.Fatalf("coverage_mode=%q from the rotating caller must be a finding, got %+v", mode, got)
		}
		// "(absent)" is a PRE-394 BINARY, and it must not read as healthy just
		// because the key is missing — the same absence-is-not-zero rule the
		// writer follows.
		if !strings.Contains(got[0].Detail, "coverage_mode") {
			t.Fatalf("the finding must name what it read: %q", got[0].Detail)
		}
	}
}

// ARM 2. A cursor that stops moving looks healthy from every other angle.
// MUTATION: compare rows[0].WindowFirst to itself, or drop the arm → green.
func TestJudgeRenderTruncation_StalledCursorIsAFinding(t *testing.T) {
	runs := []renderTruncationRun{
		{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "tool-html-minifier", OccurredAt: "2026-08-29"},
		{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "tool-html-minifier", OccurredAt: "2026-08-26"},
	}
	got := judgeRenderTruncation(runs, nil)
	if len(got) != 1 || got[0].Arm != "stalled_cursor" {
		t.Fatalf("two consecutive runs at the same window_first must be a finding, got %+v", got)
	}
}

// An EMPTY window_first carries no information and must not read as "unchanged".
// MUTATION: drop the `WindowFirst != ""` guard → two uninformative rows become a
// false stall, and the reader cries wolf on its first day.
func TestJudgeRenderTruncation_EmptyWindowFirstIsNotAStall(t *testing.T) {
	runs := []renderTruncationRun{
		{Domain: "d", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "", OccurredAt: "2"},
		{Domain: "d", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "", OccurredAt: "1"},
	}
	for _, f := range judgeRenderTruncation(runs, nil) {
		if f.Arm == "stalled_cursor" {
			t.Fatalf("an empty window_first is missing information, not evidence of a stall")
		}
	}
}

// ARM 3. A caller nobody ruled on.
// MUTATION: default unknown callers to acknowledged → a new agent quietly
// deciding not to cover a site is invisible, which is this bug one level down.
func TestJudgeRenderTruncation_UnknownCallerPagesAndAnAckedOneDoesNot(t *testing.T) {
	runs := []renderTruncationRun{
		{Domain: "leopardessconsulting.co.uk", AgentType: "design-critique-agent", CoverageMode: "prefix", OccurredAt: "2026-08-26"},
		{Domain: "somewhere.uk", AgentType: "brand-new-agent", CoverageMode: "prefix", OccurredAt: "2026-08-26"},
	}
	got := judgeRenderTruncation(runs, acks("design-critique-agent"))
	if len(got) != 1 {
		t.Fatalf("exactly the unacknowledged caller should page, got %+v", got)
	}
	if got[0].AgentType != "brand-new-agent" || got[0].Arm != "unacknowledged_caller" {
		t.Fatalf("wrong finding: %+v", got)
	}
}

// A hollow ack is not an ack — an entry with an empty reason must be ignored AND
// warned about, or "acknowledged" becomes a word anyone can type.
func TestLoadRenderTruncationAcks_HollowAckIsIgnored(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/acks.json"
	if err := os.WriteFile(p, []byte(`{"_doc":"x","a":{"reason":"a real diagnosis"},"b":{"reason":"   "}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	acked, err := loadRenderTruncationAcks(p)
	if err != nil {
		t.Fatal(err)
	}
	if !acked["a"] {
		t.Fatalf("a real ack was dropped")
	}
	if acked["b"] {
		t.Fatalf("an empty-reason ack was honoured — a hollow ack is not an ack")
	}
}

// A MISSING acks file must be an error, never an empty ack set: "could not read
// the exceptions" and "there are no exceptions" have opposite meanings and only
// one of them is safe to act on.
func TestLoadRenderTruncationAcks_MissingFileIsAnErrorNotAnEmptySet(t *testing.T) {
	if _, err := loadRenderTruncationAcks(t.TempDir() + "/nope.json"); err == nil {
		t.Fatalf("a missing acks file must be an error — an empty ack set would silently page on everything, or (worse, in the sibling shape) acknowledge nothing and read clean")
	}
}

// The stdin mode must REFUSE a payload with no aliveness total. Without it, a
// clean table and a blind read are the same reading.
func TestDecodeRenderTruncationStdin_RefusesAPayloadWithNoAlivenessTotal(t *testing.T) {
	if _, err := decodeRenderTruncationStdin(strings.NewReader(`{"runs":[]}`)); err == nil {
		t.Fatalf("a payload with no error_log_rows must be refused, not reported as clean")
	}
	if _, err := decodeRenderTruncationStdin(strings.NewReader(`[]`)); err == nil {
		t.Fatalf("a bare array must be refused with a pointer at the wrapper")
	}
	n := 41000
	in, err := decodeRenderTruncationStdin(strings.NewReader(`{"runs":[],"error_log_rows":41000}`))
	if err != nil || in.ErrorLogRows == nil || *in.ErrorLogRows != n {
		t.Fatalf("a well-formed payload was refused: %v", err)
	}
}

// The summary must state SCOPE. "0 findings over 6 rows in a log holding 41,000"
// and "0 over 0" are different readings and only the first is a pass.
func TestRenderTruncationRunSummary_StatesScopeNotJustResult(t *testing.T) {
	runs := []renderTruncationRun{
		{Domain: "webdesign.co.uk", AgentType: "render-audit-agent", CoverageMode: "cursor", WindowFirst: "a"},
	}
	s := renderTruncationRunSummary(41000, "acks.json", runs, nil)
	for _, want := range []string{"1 " + renderTruncationCode, "41000", "0 finding"} {
		if !strings.Contains(s, want) {
			t.Fatalf("summary omits %q: %s", want, s)
		}
	}
}

// The registry checker OPENS this reader's file and requires it to name both the
// code and its sink. Pinning it here means a rename cannot silently break the
// `consumed` claim in finding_code_registry.json.
func TestReaderNamesItsCodeAndSink(t *testing.T) {
	raw, err := os.ReadFile("rendertruncation.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, want := range []string{"RENDER_AUDIT_TRUNCATED", "agent_error_log"} {
		if !strings.Contains(src, want) {
			t.Fatalf("the reader file must name %q — the finding-code registry's checker reads this file to verify the `consumed` claim", want)
		}
	}
}
