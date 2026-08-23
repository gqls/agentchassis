// FILE: platform/orchestration/actions/section_metadata_parity_test.go
//
// THE CHECK THAT STOPS THIS HOP EATING KEYS (RFC_046; bugs_open/357).
//
// A section's metadata is hand-built by several producers and REBUILT FROM
// SCRATCH at the compile hop. Until 2026-08-23 that rebuild copied a literal list
// of keys into a fresh map, so a key a producer set and the list did not name was
// dropped in silence. Twice:
//
//	bugs_open/189 — stored_slot_name. Fixed by adding that one key, and pinned
//	                with a test asserting THAT KEY is forwarded.
//	bugs_open/357 — rendered_template_sha, the RFC_046 identity stamp. The
//	                per-key test could not see it, because a per-key test only
//	                knows its own key. Measured after a day live: 820 rows born,
//	                0 stamped; 546 sections_metadata elements with component_id,
//	                0 with the digest.
//
// So the rule here is about the CONTRACT, not about any key: every key a producer
// puts on a section entry must be declared as carried or declared as denied, and
// every key the save reads must be one the carrier actually delivers. A key in
// neither list fails this test instead of failing production silently.
//
// WHY AST AND NOT GREP — the same reason render_seam_one_spelling_test.go gives,
// and it applies with force here: this file's own prose names every key in the
// contract, so a text-matching rule would read its own comments as evidence and
// could be worked around by rewording. A comment is not an assignment.
package actions

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"testing"

	"go.uber.org/zap"
)

// sectionMetadataProducers names every function that builds a section metadata
// entry, and the local variable it builds it in. The variable name is what makes
// the scan precise: these functions also build unrelated maps, and a rule that
// swept up every string key in the function would report those as undeclared.
//
// ⚠ IF YOU ADD A PRODUCER: add it here. A producer missing from this map is not
// checked by anything — which is exactly the hole the rerender's fresh-render
// entry sat in while it silently omitted the stamp it held in its hand.
var sectionMetadataProducers = map[string]string{
	"RenderComponentAction":      "result", // the compile path's producer (v3_site_actions.go)
	"annotateSectionNegation":    "result", // the copy-gate wrapper, which adds a key of its own
	"RerenderPageSectionsAction": "entry",  // fresh re-render (rerender_page_sections_action.go)
	"carryStoredSection":         "m",      // carry, same file
}

// basicString unwraps a string literal, and only a string literal: a key built
// from a variable is invisible to this check, which is stated in the failure
// message rather than papered over.
func basicString(e ast.Expr) (string, bool) {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

// stringKeysAssignedTo collects the string keys written into the named variable,
// both as a composite literal (`entry := map[string]interface{}{"k": v}`) and as
// index assignments (`entry["k"] = v`).
func stringKeysAssignedTo(fn *ast.FuncDecl, varName string) []string {
	seen := map[string]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range as.Lhs {
			// entry["k"] = v
			if ix, ok := lhs.(*ast.IndexExpr); ok {
				if id, ok := ix.X.(*ast.Ident); ok && id.Name == varName {
					if s, ok := basicString(ix.Index); ok {
						seen[s] = true
					}
				}
				continue
			}
			// entry := map[string]interface{}{"k": v, ...}
			if id, ok := lhs.(*ast.Ident); ok && id.Name == varName && i < len(as.Rhs) {
				if cl, ok := as.Rhs[i].(*ast.CompositeLit); ok {
					for _, elt := range cl.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if s, ok := basicString(kv.Key); ok {
								seen[s] = true
							}
						}
					}
				}
			}
		}
		return true
	})
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// stringIndexKeysRead collects every string-literal map index READ in a function.
func stringIndexKeysRead(fn *ast.FuncDecl) []string {
	seen := map[string]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ix, ok := n.(*ast.IndexExpr); ok {
			if s, ok := basicString(ix.Index); ok {
				seen[s] = true
			}
		}
		return true
	})
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sectionKeyIsDeclared(key string) bool {
	// rendered_html is the payload, not metadata: extractSectionFromMap resolves
	// it separately, from several possible nestings, and returns it on its own.
	if key == "rendered_html" {
		return true
	}
	for _, c := range sectionMetadataCarryKeys {
		if c == key {
			return true
		}
	}
	_, denied := sectionMetadataDeniedKeys[key]
	return denied
}

// TestSectionMetadata_EveryProducerKeyIsDeclared is the emitter half: a producer
// may not invent a key that nothing downstream has agreed to carry or to drop.
//
// MUTATION THAT MUST BREAK IT: add `result["zz_undeclared"] = 1` inside
// RenderComponentAction — the key is in neither list and this test goes red.
// Removing a key from sectionMetadataCarryKeys is caught by the consumer half
// below, not by this one.
func TestSectionMetadata_EveryProducerKeyIsDeclared(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)

	for fnName, varName := range sectionMetadataProducers {
		fd, ok := funcs[fnName]
		if !ok {
			t.Fatalf("CONTROL FAILED: producer %q not found in the package — it was renamed or deleted, "+
				"and a scan that cannot find its target passes for the wrong reason", fnName)
		}
		keys := stringKeysAssignedTo(fd, varName)
		if len(keys) == 0 {
			t.Fatalf("CONTROL FAILED: found no string keys assigned to %q in %s. The variable was "+
				"probably renamed. This test would then pass while checking nothing — update "+
				"sectionMetadataProducers.", varName, fnName)
		}
		for _, k := range keys {
			if !sectionKeyIsDeclared(k) {
				t.Errorf("%s sets section metadata key %q, which is in NEITHER sectionMetadataCarryKeys "+
					"NOR sectionMetadataDeniedKeys.\n"+
					"Decide which, in section_metadata_keys.go, and say why if it is denied. An "+
					"undeclared key is dropped at the compile hop in total silence — that is how the "+
					"RFC_046 identity stamp shipped inert for a day (bugs_open/357), and how "+
					"stored_slot_name was lost before it (bugs_open/189).", fnName, k)
			}
		}
	}
}

// TestSectionMetadata_EverySaveReadIsCarried is the consumer half, and it is the
// one that would have caught bugs_open/357 on the day the stamp was written: the
// save read a key the carrier had no instruction to deliver, and nothing anywhere
// compared the two ends.
//
// MUTATION THAT MUST BREAK IT: delete "rendered_template_sha" from
// sectionMetadataCarryKeys — the save still reads it, and this test goes red.
func TestSectionMetadata_EverySaveReadIsCarried(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)

	fd, ok := funcs["extractSectionsFromMetadata"]
	if !ok {
		t.Fatal("CONTROL FAILED: extractSectionsFromMetadata not found")
	}
	keys := stringIndexKeysRead(fd)
	if len(keys) < 6 {
		t.Fatalf("CONTROL FAILED: extractSectionsFromMetadata appears to read only %d string keys (%v). "+
			"It reads at least six; a traversal finding fewer is broken and would pass vacuously.",
			len(keys), keys)
	}
	for _, k := range keys {
		if k == "rendered_html" {
			continue
		}
		carried := false
		for _, c := range sectionMetadataCarryKeys {
			if c == k {
				carried = true
			}
		}
		if !carried {
			t.Errorf("save_page_sections reads section metadata key %q, but the compile hop does not "+
				"carry it — sectionMetadataCarryKeys does not name it.\n"+
				"The save will read an empty value on every real run and write the absent-fact "+
				"branch, silently and for ever. Add the key to the carry list, or stop reading it.", k)
		}
	}
}

// TestSectionMetadata_ContractIsWellFormed keeps the two lists honest: a key in
// both would make the deny reason a lie, and an empty carry list would make every
// rule above vacuous.
func TestSectionMetadata_ContractIsWellFormed(t *testing.T) {
	if len(sectionMetadataCarryKeys) == 0 {
		t.Fatal("sectionMetadataCarryKeys is empty — every parity rule here would pass vacuously")
	}
	for _, c := range sectionMetadataCarryKeys {
		if reason, denied := sectionMetadataDeniedKeys[c]; denied {
			t.Errorf("key %q is both carried and denied (deny reason: %q). One of the two is wrong, "+
				"and while they disagree the reason field is not evidence of anything.", c, reason)
		}
	}
	for k, reason := range sectionMetadataDeniedKeys {
		if reason == "" {
			t.Errorf("denied key %q has no stated reason. The reason is the point: a drop with a "+
				"rationale is a decision, a drop without one is bugs_open/357.", k)
		}
	}
}

// TestExtractSectionFromMap_CarriesEveryDeclaredKey is the round trip: a
// producer-shaped map goes through the carrier and out the other side into
// SectionData, and every declared key survives. It covers BOTH shapes the carrier
// accepts — flat, and nested under a substep — because the nested recovery loop
// was a second hand-written copy of the same list and would have had to be fixed
// twice.
//
// MUTATION THAT MUST BREAK IT: remove "rendered_template_sha" from
// sectionMetadataCarryKeys — RenderedTemplateSHA arrives empty, which is exactly
// what production did.
func TestExtractSectionFromMap_CarriesEveryDeclaredKey(t *testing.T) {
	const (
		wantSHA     = "b8f1c0dd2e4a6c8091f3ab5d7e9c1204a6b8d0f2e4c68a0b2d4f6081a3c5e7f9"
		wantVersion = "6f1d2c3b-4a59-4c8e-9b1f-2d3e4a5b6c7d"
		wantCompID  = "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
	)
	producer := func() map[string]interface{} {
		return map[string]interface{}{
			"rendered_html":         "<section>hi</section>",
			"component_id":          wantCompID,
			"component_name":        "hero",
			"component_function":    "hero",
			"stored_slot_name":      "prose-0",
			"content_data":          map[string]interface{}{"headline": "H"},
			"rendered_template_sha": wantSHA,
			"component_version_id":  wantVersion,
		}
	}

	cases := map[string]map[string]interface{}{
		"flat": producer(),
		// The LoopAction nesting: the carrier must recover the same keys from a
		// substep's output, using the same declared list.
		"nested_under_section_output": {
			"section_output": producer(),
		},
	}

	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			html, meta := extractSectionFromMap(in, zap.NewNop())
			if html == "" {
				t.Fatalf("no html returned; meta=%v", meta)
			}
			for _, k := range sectionMetadataCarryKeys {
				if _, ok := meta[k]; !ok {
					t.Errorf("declared carry key %q did not survive extractSectionFromMap (meta=%v)", k, meta)
				}
			}

			got := extractSectionsFromMetadata([]interface{}{meta}, zap.NewNop())
			if len(got) != 1 {
				t.Fatalf("extractSectionsFromMetadata returned %d sections, want 1", len(got))
			}
			s := got[0]
			if s.RenderedTemplateSHA != wantSHA {
				t.Errorf("RenderedTemplateSHA = %q, want %q — the provenance stamp did not reach the save, "+
					"which is bugs_open/357 exactly", s.RenderedTemplateSHA, wantSHA)
			}
			if s.ComponentVersionID != wantVersion {
				t.Errorf("ComponentVersionID = %q, want %q", s.ComponentVersionID, wantVersion)
			}
			if s.ComponentID != wantCompID {
				t.Errorf("ComponentID = %q, want %q", s.ComponentID, wantCompID)
			}
			if s.ComponentName == "" || s.ContentData == nil {
				t.Errorf("pre-existing keys regressed: name=%q content_data=%v", s.ComponentName, s.ContentData)
			}
		})
	}
}
