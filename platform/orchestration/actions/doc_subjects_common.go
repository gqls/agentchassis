// FILE: platform/orchestration/actions/doc_subjects_common.go   (NEW FILE)
//
// Centralised subject_type vocabulary for the travelling-docs substrate
// (doc_plans/doc_notes), in the same mould as work_items_common.go — the
// dedup-index/terminal-status lockstep lesson (v1.0.1127) applied to doc
// subjects. Before this file the contract had FOUR enforcement points (two DB
// CHECKs + two independently hard-coded Go gates) and every addition moved a
// subset of them: migration 163 (+'experience') missed the
// persist_diagnosis_note gate; migration 184 (+'action') moved the DB CHECKs
// only, leaving its own seeded action docs unreachable through every doc
// action (bugs_open/064). Now there are two points: the DB CHECKs and this
// list.

package actions

import (
	"fmt"
	"strings"
)

// validDocSubjectTypes is the canonical set of subject_type values a
// travelling doc (doc_plans/doc_notes row) may carry. It MUST agree with the
// DB CHECK constraints doc_plans_subject_type_check /
// doc_notes_subject_type_check — a value the DB accepts but a Go gate
// rejects, or vice versa, is a split contract; move both together.
// History: 'tool','pipeline' (base schema); +'experience' (migration 163,
// experience loop); +'action' (migration 184, shared fix-loop actions);
// +'experience-pattern' (migration 218, experience register — one travelling
// doc per register entry, keyed by the pattern name, carrying its direction,
// provenance and council trail exactly as a tool's does);
// +'component' (migration 273, staged component build — a section component
// carries a PLAN with its criteria fence and NOTES with its per-site verdicts,
// keyed by content_components.function);
// +'landmine' (migration 270, doc_notes ONLY — D10's LANDMINES.md sync. Found
// missing here 2026-07-31 by landmine-verifier's first live run: the DB
// accepted subject_type='landmine' on doc_notes for two days while this gate
// rejected it, the exact split bugs_closed/064 exists to prevent. It slipped
// past TestValidDocSubjectTypes_LockstepWithMigrationCheck because that test
// only scans migrations recreating doc_plans_subject_type_check — 270 never
// touches doc_plans (landmine rows never belong there), so the test had
// nothing to catch. See the sibling test below, added for the same reason).
//
// WHY 'component' IS KEYED BY function, AND WHY THE PLAN CARRIES NO SITE.
// A component's TEMPLATE is fleet-shared — one content_components row serves
// 11 sites for info-card-grid — so its contract is a fleet-wide fact and
// doc_plans (which has no site_id) is the right home for it. Whether that
// component actually works is a per-site, per-page question, and those
// verdicts land in doc_notes, which does have site_id. PLAN = the shared
// contract; NOTES = the per-site verdicts. Do not add a site_id to doc_plans
// to "fix" this: the asymmetry is the design.
//
// SUBJECT KEYS ARE IMMUTABLE ONCE WRITTEN — supersede under a new key rather
// than renaming. RekeyTravellingDocs is wired only to rename_tool_identity, so
// a renamed component function would orphan its docs silently. This is not
// hypothetical: bugs_open/136 is a half-landed *_domain -> *_pipeline rename.
//
// When adding a subject type, follow the full checklist in
// docs/agent_docs/docs024_key_docs_latest/experience_register/design/subject_type_addition.md
// — image BEFORE migration, or the widened CHECK just recreates 184's split.
// TestValidDocSubjectTypes_LockstepWithMigrationCheck reads the newest
// migration that sets the CHECK and fails on drift, so this list and the
// migration must land in the SAME commit.
var validDocSubjectTypes = []string{"tool", "pipeline", "experience", "action", "experience-pattern", "component", "landmine"}

// isValidDocSubjectType reports membership in validDocSubjectTypes.
func isValidDocSubjectType(subjectType string) bool {
	for _, t := range validDocSubjectTypes {
		if subjectType == t {
			return true
		}
	}
	return false
}

// docSubjectTypesQuoted renders the vocabulary for error and log messages:
// 'tool', 'pipeline', 'experience', 'action'.
func docSubjectTypesQuoted() string {
	quoted := make([]string, len(validDocSubjectTypes))
	for i, t := range validDocSubjectTypes {
		quoted[i] = "'" + t + "'"
	}
	return strings.Join(quoted, ", ")
}

// docSubjectGateReason classifies a (subject_type, subject_key) pair for the
// skip-don't-guess gates: "" means proceed, otherwise the skip reason. The
// two reasons are deliberately distinct (bugs_open/064: an explicit
// 'experience' subject used to be skipped as "no explicit subject" — the
// subject was explicit, only its type fell outside a stale private
// allowlist).
func docSubjectGateReason(subjectType, subjectKey string) string {
	if subjectType == "" || subjectKey == "" {
		return "no explicit subject"
	}
	if !isValidDocSubjectType(subjectType) {
		return fmt.Sprintf("unsupported subject_type %q (valid: %s)", subjectType, docSubjectTypesQuoted())
	}
	return ""
}
