// queryselect.go — per-hypothesis selection of VETTED, read-only DB queries for
// the runtime-evidence gather. The data analogue of docselect.go: as the loop
// re-scopes, the next iteration pulls the domain-table evidence that fits the
// CURRENT hypothesis (page_components for a rendering hypothesis, work-items for
// a dispatch hypothesis, specs for a classification hypothesis) — the same
// "follow the evidence" motion the call graph gives for code.
//
// WHY A CATALOGUE, NOT MODEL-WRITTEN SQL (the safety boundary):
// The loop is read-only BY CONSTRUCTION. dbcontext's -rows mode will run whatever
// SELECT it is handed, so if a verdict could emit raw SQL the loop would lose
// that guarantee (an expensive query, the wrong DB, or a non-SELECT). So the
// queries are HAND-WRITTEN, parameterised, and \d-verified ONCE; the loop only
// SELECTS among them by hypothesis. The model never writes SQL. This also removes
// the schema-guessing + query-rewrite churn a manual investigation suffers: a
// catalogue query is verified against the real schema before it enters the
// catalogue, then reused.
//
// PARAMS: each query binds to the loop's EXISTING context (site_id, domain, page,
// correlation_id) — the values already in input_data / the seed. So the first
// version needs no wire-format change and no model-supplied parameters. A later
// increment can let the verdict NAME a catalogue query (still not raw SQL) for
// targeted drill-down; that needs a wire field + param validation and is NOT
// built here.
package diagnose

import (
	"encoding/json"
	"os"
	"sort"
)

// QueryRule is one vetted read-only query plus the conditions under which it is
// relevant to the current hypothesis/scope. Matching is identical to DocRule
// (Always, OR a Keyword substring of the hypothesis, OR a PathGlob substring of
// an in-scope symbol path), so the two selectors behave the same way.
type QueryRule struct {
	Name      string   `json:"name"`                 // identifier, e.g. "page_components_for_page"
	Title     string   `json:"title,omitempty"`      // human label for the bundle section
	SQL       string   `json:"sql"`                  // a SINGLE parameterised read-only SELECT; binds Params as $1..$n
	Params    []string `json:"params,omitempty"`     // loop-context keys bound in order, e.g. ["site_id","page_name"]
	Keywords  []string `json:"keywords,omitempty"`   // matched (case-insensitive substring) against the hypothesis
	PathGlobs []string `json:"path_globs,omitempty"` // matched against in-scope symbol paths
	Always    bool     `json:"always,omitempty"`     // run regardless of hypothesis/scope
}

// SelectQueries returns the queries whose rules match the current hypothesis or
// scope (plus any Always queries), de-duplicated by Name and stably ordered.
// Reuses the same matching helpers as SelectDocs (matchesAny / pathMatchesAny).
func SelectQueries(hypothesis string, scope Scope, rules []QueryRule) []QueryRule {
	h := lower(hypothesis)
	paths := scopePaths(scope)

	seen := map[string]bool{}
	var out []QueryRule
	for _, r := range rules {
		if seen[r.Name] {
			continue
		}
		if r.Always || matchesAny(h, r.Keywords) || pathMatchesAny(paths, r.PathGlobs) {
			seen[r.Name] = true
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// LoadQueryCatalogue reads a JSON array of QueryRule (the operator's vetted,
// \d-verified query set). Empty path returns nil, no error.
func LoadQueryCatalogue(path string) ([]QueryRule, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules []QueryRule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
