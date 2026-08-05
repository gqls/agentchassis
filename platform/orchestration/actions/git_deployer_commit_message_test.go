package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// resolveCommitMessage decides what a git_commit step's commit says. bugs_open/198:
// the template context is a fixed {domain, file_count, filename} map, so the
// css-patch-agent's template naming {{.input_data.spec.category}} rendered
// "CSS fix: <no value>" on all four incident commits — the audit trail of a live
// clobber was destroyed by its own messages. commit_message_field is the opt-in
// escape: the message is composed upstream (e.g. in a query_database RETURNING,
// where params resolve) and read verbatim from CollectedData here.
func TestResolveCommitMessageFieldWinsOverTemplate(t *testing.T) {
	log := zap.NewNop()
	cfg := map[string]interface{}{
		"commit_message":       "CSS patch: {{.filename}} ({{.domain}})",
		"commit_message_field": "css_saved.commit_msg",
	}
	data := map[string]interface{}{
		"css_saved": map[string]interface{}{
			"commit_msg": "CSS fix: contrast_failure (theme v7)",
		},
	}

	got := resolveCommitMessage(cfg, data, "relojistas.com", 1, "assets/css/styles.css", log)
	if got != "CSS fix: contrast_failure (theme v7)" {
		t.Fatalf("field did not win: got %q", got)
	}
}

// An unresolvable field must NOT produce an empty or broken message — it falls
// back to the template. This is the shipped-ahead-of-the-binary case inverted:
// a config carrying the field against a binary that reads it, but whose data is
// missing (e.g. the upstream step errored into a path that still commits).
func TestResolveCommitMessageFallsBackWhenFieldResolvesEmpty(t *testing.T) {
	log := zap.NewNop()
	cfg := map[string]interface{}{
		"commit_message":       "CSS patch: {{.filename}} ({{.domain}})",
		"commit_message_field": "css_saved.commit_msg",
	}
	data := map[string]interface{}{} // nothing saved

	got := resolveCommitMessage(cfg, data, "relojistas.com", 1, "assets/css/styles.css", log)
	if got != "CSS patch: assets/css/styles.css (relojistas.com)" {
		t.Fatalf("fallback template did not render: got %q", got)
	}
}

// No field configured → identical behaviour to before the field existed. An old
// config must not change meaning under the new binary.
func TestResolveCommitMessageUnsetFieldIsPureTemplate(t *testing.T) {
	log := zap.NewNop()
	cfg := map[string]interface{}{
		"commit_message": "Update {{.file_count}} files for {{.domain}}",
	}

	got := resolveCommitMessage(cfg, map[string]interface{}{}, "example.com", 3, "", log)
	if got != "Update 3 files for example.com" {
		t.Fatalf("template path changed: got %q", got)
	}
}

// Pins the defect mechanism itself so nobody "fixes" a workflow by writing a
// richer template: keys outside {domain, file_count, filename} render
// "<no value>" — that is WHY commit_message_field exists. If this test ever
// fails because the template context grew, commit_message_field may no longer
// be needed; retire them together, not separately.
func TestCommitMessageTemplateCannotSeeCollectedData(t *testing.T) {
	cfg := map[string]interface{}{
		"commit_message": "CSS fix: {{.input_data.spec.category}}",
	}

	got := buildCommitMessage(cfg, "example.com", 1, "styles.css")
	if !strings.Contains(got, "<no value>") {
		t.Fatalf("template context apparently grew beyond {domain, file_count, filename}: got %q — retire commit_message_field's fallback docs with it", got)
	}
}
