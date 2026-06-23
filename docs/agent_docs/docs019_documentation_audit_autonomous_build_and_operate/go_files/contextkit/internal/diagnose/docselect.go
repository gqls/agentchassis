// docselect.go — per-hypothesis selection of authored context docs for the
// bundle. The loop pastes the 003 contract / 016 §9 entry / dev-guide section
// that matches the CURRENT hypothesis or scope into each iteration's bundle,
// rather than every doc into every bundle. Two reasons:
//   - the always-on rules already ride in via -constitution (the thin-slice
//     constitution is in every bundle); these are the TASK-SPECIFIC standards
//     the constitution itself says to "paste alongside when a task touches them";
//   - including everything dilutes the verdict (the B4a lesson: irrelevant
//     context buries the signal).
//
// Selection is DETERMINISTIC and testable — the same property the guards and
// re-scope have — so it can be exercised without a model. It is a HEURISTIC
// (keyword/path substring), so its failure mode is benign: a false match adds a
// contract doc (not misinformation), and a miss degrades to "no extra doc"
// (today's behaviour). A future extension can let the verdict NAME a needed doc
// (a `needed_docs` field, mirroring `needed_evidence`/`next_scope`) and ADD it to
// this set — not built here.
//
// REUSE: this is a pure function in the diagnose package, so BOTH realisations of
// the loop call it — the standalone BundleGatherer (this module) and the chassis
// diagnose_assemble_bundle action (which can take a catalogue as config).
package diagnose

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
)

// DocRule maps one authored doc to the conditions under which it is relevant to
// the current hypothesis/scope. A rule matches if Always, OR any Keyword is a
// substring of the hypothesis, OR any PathGlob is a substring of an in-scope
// symbol's path. A rule with no conditions and Always=false never matches.
type DocRule struct {
	Doc       string   `json:"doc"`                  // path to the doc, forwarded verbatim via bundle -doc
	Keywords  []string `json:"keywords,omitempty"`   // matched (case-insensitive substring) against the hypothesis
	PathGlobs []string `json:"path_globs,omitempty"` // matched (case-insensitive substring) against in-scope symbol paths
	Always    bool     `json:"always,omitempty"`     // include regardless of hypothesis/scope
}

// SelectDocs returns the docs whose rules match the current hypothesis or scope
// (plus any Always docs), de-duplicated and stably (alphabetically) ordered.
// Matching is case-insensitive substring — dependency-free and consistent with
// the analyser's substring excludes.
func SelectDocs(hypothesis string, scope Scope, rules []DocRule) []string {
	h := strings.ToLower(hypothesis)

	// in-scope paths (strip ":Symbol" to the file path), lowercased
	paths := make([]string, 0, len(scope.Symbols))
	for _, s := range scope.Symbols {
		p := s
		if i := strings.Index(p, ":"); i >= 0 {
			p = p[:i]
		}
		paths = append(paths, strings.ToLower(p))
	}

	seen := map[string]bool{}
	var out []string
	add := func(doc string) {
		if doc != "" && !seen[doc] {
			seen[doc] = true
			out = append(out, doc)
		}
	}

	for _, r := range rules {
		if r.Always {
			add(r.Doc)
			continue
		}
		if matchesAny(h, r.Keywords) || pathMatchesAny(paths, r.PathGlobs) {
			add(r.Doc)
		}
	}
	sort.Strings(out)
	return out
}

func matchesAny(hypothesis string, keywords []string) bool {
	for _, kw := range keywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" && strings.Contains(hypothesis, kw) {
			return true
		}
	}
	return false
}

func pathMatchesAny(paths, globs []string) bool {
	for _, g := range globs {
		g = strings.ToLower(strings.TrimSpace(g))
		if g == "" {
			continue
		}
		for _, p := range paths {
			if strings.Contains(p, g) {
				return true
			}
		}
	}
	return false
}

// LoadDocCatalogue reads a JSON array of DocRule from a file (the operator's doc
// index — the 003 contracts, the 016 §9 entries, the dev-guide sections, each
// keyed to its keywords/paths). An empty path returns nil with no error, so the
// catalogue is optional.
func LoadDocCatalogue(path string) ([]DocRule, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules []DocRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
