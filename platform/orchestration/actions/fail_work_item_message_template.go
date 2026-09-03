// FILE: platform/orchestration/actions/fail_work_item_message_template.go
//
// The OPT-IN half of fail_work_item's refusal message — bugs_open/440, RFC_062
// phase 3, owner ruling D1 ("a refused item routes to needs_human_review; the
// message names the bad key AND the vocabulary").
//
// WHY THIS EXISTS AT ALL. `fail_work_item`'s `error_message` is a config
// LITERAL, resolved through no path and no template — the comment at its read
// site says so deliberately, to keep a message from being mistaken for a data
// reference. `[MEASURED 2026-09-03]` seven live steps across six agents
// configure it and NONE contains `{{`. That is fine for every message it
// carries today, because each of them describes a fixed situation
// ("tool-improver produced a structurally-collapsed component"). It is NOT
// enough for a refusal whose whole job is to report the OFFENDING VALUE: a
// static string can name the field `spec.routing_reason` and the vocabulary,
// but not the key that was actually wrong, and D1 asks for the key.
//
// ── WHY OPT-IN, AND WHY THAT IS THE WHOLE SAFETY ARGUMENT ─────────────────
//
// This is a shared action: six live agents' refusal steps run through it. Per
// the owner ruling of 2026-08-02 §2, new authority on a shared seam ships as a
// field whose unsafe side is the DEFAULT OFF, not as a documented contract —
// "a comment is not a control on a tree this many sessions share". So:
//
//   - absent `error_message_template` → this file does nothing at all, and the
//     literal is used exactly as it is today. Byte-inert for all seven live
//     steps, which is checked by execution in the test beside this file, not
//     asserted here.
//   - present → it renders over collected_data and REPLACES the literal.
//
// Under RFC_022's three-condition test (opt-in; the unsafe side is the default;
// zero live consumers name it — enumerable with the query in the test header)
// that makes this NOT architecture-scope: it goes through the ordinary council
// gate as part of phase 3.
//
// ── THE TWO FAILURE MODES, BOTH LOUD ON PURPOSE ───────────────────────────
//
// A refusal message is read by a human deciding what to do with a parked item,
// so a message that renders WRONG is worse than one that renders plainly.
//
//  1. The template does not parse, or execution fails. `missingkey=error` is
//     set, so a path that is not there is an ERROR rather than a shrug.
//  2. It parses, executes, and emits `<no value>` — text/template's rendering
//     of a key that is PRESENT but nil. `missingkey=error` does not catch this
//     one, and it is the trap git_deployer already carries a regression test
//     for (git_deployer_commit_message_test.go: "<no value>" is WHY
//     commit_message_field exists). A refusal reading `routing_reason =
//     '<no value>'` would send a human hunting for a key that is not the
//     problem.
//
// Both fall back to the `error_message` literal — which is why a step that
// configures a template MUST also configure a literal, and why the fallback is
// reported through LogActionEntry rather than a log line: an operator needs to
// find out that the message they are reading is the second-best one. Falling
// back SILENTLY would make a degraded message indistinguishable from the
// intended one, which is this bug's own shape.

package actions

import (
	"fmt"
	"strings"
	"text/template"
)

// failWorkItemNoValue is text/template's rendering of a present-but-nil key.
// Detected rather than tolerated — see mode 2 in the header.
const failWorkItemNoValue = "<no value>"

// renderFailWorkItemMessage renders an `error_message_template` over the run's
// collected data.
//
// The error is the load-bearing return: every caller must fall back to the
// literal and SAY it did. Returning a best-effort string with a nil error
// would put a half-rendered refusal in front of a human with nothing marking
// it as degraded.
func renderFailWorkItemMessage(tmplText string, collectedData map[string]interface{}) (string, error) {
	tmpl, err := template.New("fail_work_item_error_message").
		Option("missingkey=error").
		Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("error_message_template does not parse: %w", err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, collectedData); err != nil {
		return "", fmt.Errorf("error_message_template failed to render: %w", err)
	}

	rendered := out.String()
	if strings.Contains(rendered, failWorkItemNoValue) {
		return "", fmt.Errorf(
			"error_message_template rendered %q — a key it names is present but nil, so the "+
				"message would report the wrong thing to whoever reads the parked item; "+
				"rendered text was: %s", failWorkItemNoValue, rendered)
	}
	if strings.TrimSpace(rendered) == "" {
		return "", fmt.Errorf("error_message_template rendered empty text")
	}

	return rendered, nil
}
