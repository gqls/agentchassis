package actions

import "fmt"

// ONE definition of what travels with a section's bytes (RFC_046; bugs_open/357).
//
// THE DEFECT THIS EXISTS TO END. A section's metadata is hand-built by three
// producers and rebuilt from scratch at the compile hop, where
// extractSectionFromMap copied a literal list of keys into a fresh map. A key the
// producer set and the list did not name was dropped in silence — no error, no
// log, no failing test. That is how bugs_open/357's identity stamp came to be
// built, tested, council-approved, rolled, and inert: RenderComponentAction set
// rendered_template_sha, the carrier did not copy it, save_page_sections read an
// empty string and wrote NULL. Measured 2026-08-23, after the stamp had been live
// for a day: 820 page_components rows born, 0 stamped; 546 sections_metadata
// elements carrying component_id, 0 carrying the digest.
//
// THIS IS THE SECOND TIME. bugs_open/189 lost stored_slot_name at this same hop.
// The remedy then was to add that one key and pin it with a test asserting THAT
// KEY is forwarded — a good test that could not see the next key, because a
// per-key test only knows its own key. So the remedy here is the contract rather
// than the key: a section metadata key must be named as carried or named as
// denied, and a key in neither list fails section_metadata_parity_test.go.
//
// WHY NOT CARRY EVERYTHING. sections_metadata is persisted into
// orchestration_states.collected_data, which already averages 210KB per run and
// peaks at 2.6MB [MEASURED 2026-08-23], and 7 live agent types consume it. A
// default-carry rule would turn every future producer key into a silent size and
// shape change on a persisted artefact, and would remove the moment where a new
// key's carriage is actually decided. Declaring beats defaulting here.

// sectionMetadataCarryKeys are the keys that travel from a section producer
// (RenderComponentAction, RerenderPageSectionsAction, carryStoredSection) through
// CompilePageSectionsAction to save_page_sections. rendered_html is not in this
// list because it is the payload rather than metadata: extractSectionFromMap
// resolves it first, from several possible nestings, and returns it separately.
var sectionMetadataCarryKeys = []string{
	"component_id",
	"component_name",
	"component_function",
	"content_data",
	"stored_slot_name",
	// ── Provenance (RFC_046). Both are read by extractSectionsFromMetadata and
	// resolved to page_components.component_version_id at the single INSERT.
	"rendered_template_sha",
	"component_version_id",
}

// sectionMetadataDeniedKeys are keys a producer sets that deliberately do NOT
// travel, each with the reason it does not. The reason is the point: a dropped
// key with a stated rationale is a decision, and a dropped key with no entry
// anywhere is the bug above. Adding a producer key without choosing one list or
// the other fails the parity test.
var sectionMetadataDeniedKeys = map[string]string{
	"copy_gate_findings": "per-render annotation attached by the copy gate wrapper; " +
		"no reader at the save, and carrying it would persist finding arrays into " +
		"collected_data a second time",
	"stripped_markdown_fields": "whole-run diagnostic reported at the action result, " +
		"not a per-section fact about these bytes",
}

// adoptCarriedProvenance makes a section's provenance describe the bytes it has
// just been HANDED, rather than the bytes it arrived with (RFC_046,
// bugs_open/357).
//
// Layer 2 carries stored markup into a section the rebuild produced — the splice
// that keeps an interactive tool alive when the fresh composition would have
// replaced it with prose. What it replaces is the HTML; what it used to leave
// behind was the fresh render's digest, which describes markup that has just been
// thrown away. Resolve that digest and it matches the component that rendered the
// discarded band, stamping THAT version onto the tool now stored in the row: not
// a missing answer but a false one, and the estate's own rule is that a stamp
// naming a template which did not produce the bytes is worse than no stamp.
//
// So the digest is cleared and the stored row's own stamp — which does describe
// these bytes — is adopted. Empty stored stamp means the row was never stamped,
// and the section stays honestly unknown, which is the state every row in
// bugs_open/357's population is in today.
//
// THIS IS A FUNCTION RATHER THAN TWO LINES AT THE SPLICE because two lines at the
// splice could not be tested: the resolver swallows its own query errors and
// returns "no stamp" on failure, so at the level of the whole action a section
// carrying a stale digest and a section carrying none write the SAME nil bind.
// The mutation that removes the clearing therefore passes an action-level test —
// measured, on the first attempt at exactly that test. A guard in series
// (resolver error handling) stood in for the property under test. Pinning the
// decision here makes it directly observable, and TestSplice_UsesAdoptCarried…
// keeps the seam calling it.
func adoptCarriedProvenance(s *SectionData, storedVersionID string) {
	if s == nil {
		return
	}
	s.RenderedTemplateSHA = ""
	s.ComponentVersionID = storedVersionID
}

// sectionMetaComplete reports whether every carried key has been found, so the
// nested-substep search can stop looking. It is an optimisation and nothing else:
// a section that legitimately has no content_data, or bytes nobody rendered and
// so no provenance, is never "complete" and simply costs a scan of the two
// remaining substep shapes.
func sectionMetaComplete(meta map[string]interface{}) bool {
	for _, key := range sectionMetadataCarryKeys {
		if _, ok := meta[key]; !ok {
			return false
		}
	}
	return true
}

// carrySectionMetaKey copies one declared key from a producer's map into the
// metadata map the save will read, preserving exactly the per-key behaviour the
// hand-written blocks had before this list existed:
//
//   - component_id accepts any non-nil value and is rendered with %v, because
//     producers supply it as both a string and a uuid.UUID;
//   - content_data is taken whole when non-nil, as a map or as anything else a
//     producer chose to put there;
//   - every other key is a string and is carried only when non-empty, so an
//     absent fact stays absent rather than becoming an empty one. That
//     distinction is load-bearing for provenance: "" means unknown, and unknown
//     must reach the database as NULL rather than as a stamp nobody earned.
//
// It never overwrites a value already present, so a top-level key wins over the
// same key recovered from a nested substep — the precedence extractSectionFromMap
// has always had.
func carrySectionMetaKey(dst, src map[string]interface{}, key string) {
	if dst == nil || src == nil {
		return
	}
	if _, already := dst[key]; already {
		return
	}
	v, ok := src[key]
	if !ok || v == nil {
		return
	}

	switch key {
	case "component_id":
		dst[key] = fmt.Sprintf("%v", v)
	case "content_data":
		dst[key] = v
	default:
		if s, isStr := v.(string); isStr && s != "" {
			dst[key] = s
		}
	}
}
