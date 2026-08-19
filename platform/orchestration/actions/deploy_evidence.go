// FILE: platform/orchestration/actions/deploy_evidence.go
//
// bugs_open/315 / RFC_038 — reading what the deploy step actually DID.
//
// THE DEFECT THIS SERVES. `UpdatePageStatusAction` stamps `pages.deployed_at`
// without ever looking at the deploy step's output. Measured 2026-08-19 across
// the live fleet: of the five agents that stamp `deployed`, two did it before
// the deploy was dispatched at all (migration 491 removed those), and the
// remaining three run immediately after a `git_commit` whose result they
// discard — the string `deploy_result` appears nowhere in v3_site_actions.go.
// So a `git_commit` that returned `{"status":"skipped","skip_reason":"no files
// to commit"}` was followed, one step later, by a fresh `deployed_at`.
//
// WHY THIS IS NOT A LITERAL FIELD PATH. Two reasons, both measured:
//
//  1. The field NAME varies. The 19 live `git_commit` steps carry NINE distinct
//     `output_field` values and two set none; `deploy_result` names only three
//     of them, and `section-editor` uses `git_result`. A guard hard-coding one
//     name is blind on 16 of 19 — and, since it must fail open, would wave them
//     through in silence. So the name comes from the step's own config.
//
//  2. The field SHAPE varies. Over 744 orchestrations in 7 days, 57 (7.7%) are
//     nested one level deeper — `<field>.response.<field>.response.data.…` —
//     because the deploy was performed by a called sub-agent rather than inline.
//     Indexing a fixed path reports those as "no evidence".
//
// So resolution searches the named subtree for the keys by name rather than
// indexing a fixed path.
//
// ⚠ CORRECTED 2026-08-19, and this is why the collector below is ours rather
// than `datahelpers.ExtractFields`. An earlier version of this file claimed
// ExtractFields gives "collect-all / unique-or-nothing (RFC_029 §9)", so an
// ambiguous subtree would resolve to nothing rather than to a guess. **That
// claim was false for this build, and the council gate's prior_art_librarian
// seat caught it against a landmine entry that said the opposite.** Reading
// `findFieldRecursive` to the END settles it: the RULING is unique-or-nothing,
// but "PHASE 1 (this build — instrument first, refuse second): conflicts still
// resolve, to the STABLE shallowest-first winner, and emit the WARN below".
// Phase 2 flips conflicts to refusal and has not shipped.
//
// For this caller a guess is the worst possible outcome: a fingerprint taken
// from the WRONG git_commit is silently, permanently wrong, and every later
// comparison reports a page as diverged when it is fine. "No fingerprint" is
// recoverable; "someone else's fingerprint" is not. So we do not borrow a
// safety property that does not exist yet — we collect the candidates
// ourselves and REFUSE on conflict.
package actions

import (
	"context"
	"database/sql"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// deployEvidence is what the deploy step reported about its own work.
//
// The three states are deliberately distinct, because they call for opposite
// responses: Skipped means do NOT stamp (nothing was sent); a resolved commit
// means stamp AND record what was sent; and "unreadable" — signalled by the
// bool return, not by a field — means fall through to the old behaviour and say
// so, because refusing every stamp on a resolution failure would freeze deploys
// fleet-wide over a config typo.
type deployEvidence struct {
	// Skipped is set when the deploy step declined to do anything. The reason
	// carries the guard prefixes (OWNED_PAGE_GUARD / ARCHIVED_PAGE_GUARD) when
	// one of those refused it, which the caller dispatches on.
	Skipped    bool
	SkipReason string

	// CommitSHA and FilesSHA256 come from the git-adapter's reply (RFC_038).
	// Both are EMPTY against a git-adapter image older than that change, which
	// is the normal state during a partial roll — the chassis and the adapter
	// are separate images. That case is "unreadable", not "failed".
	CommitSHA   string
	FilesSHA256 map[string]string
}

// deployEvidenceKeys are the reply keys worth resolving. Deliberately specific:
// "status" would collide with a dozen unrelated keys in a collected-data tree,
// whereas none of these three appears anywhere else in the estate's vocabulary.
var deployEvidenceKeys = []string{"skip_reason", "commit_sha", "files_sha256"}

// resolveDeployEvidence reads the deploy step's output field.
//
// ok=false means "no usable evidence here" and the caller must fail OPEN. It
// covers a genuinely absent field, a field holding something that is not a map,
// and — importantly — a reply from a git-adapter that predates RFC_038 and
// therefore carries no commit identity at all. The caller records that rather
// than swallowing it: a silent fail-open is how bugs_open/315 survived four
// completed rerenders.
func resolveDeployEvidence(collected map[string]interface{}, field string, logger *zap.Logger) (deployEvidence, bool) {
	field = strings.TrimSpace(field)
	if field == "" || collected == nil {
		return deployEvidence{}, false
	}

	subtree, ok := collected[field].(map[string]interface{})
	if !ok {
		return deployEvidence{}, false
	}

	found := make(map[string]interface{}, len(deployEvidenceKeys))
	for _, k := range deployEvidenceKeys {
		v, ok, conflict := collectUniqueValue(subtree, k, 0)
		if conflict {
			// Ambiguous: two git_commit results under one field. Refuse the whole
			// resolution rather than fingerprint the page from an arbitrary one.
			logger.Warn("deploy evidence: CONFLICTING candidates in the named subtree — refusing rather than guessing",
				zap.String("deploy_result_field", field),
				zap.String("key", k))
			return deployEvidence{}, false
		}
		if ok {
			found[k] = v
		}
	}

	var ev deployEvidence
	if reason, _ := found["skip_reason"].(string); strings.TrimSpace(reason) != "" {
		ev.Skipped = true
		ev.SkipReason = reason
		// A skip is complete evidence on its own: nothing was committed, so
		// there is no sha and no hash to look for.
		return ev, true
	}

	ev.CommitSHA, _ = found["commit_sha"].(string)
	ev.FilesSHA256 = stringMap(found["files_sha256"])

	// Neither half resolved => this reply predates RFC_038, or is not a git
	// commit reply at all. Not a failure, and emphatically not a skip: unreadable.
	if ev.CommitSHA == "" && len(ev.FilesSHA256) == 0 {
		return deployEvidence{}, false
	}
	return ev, true
}

// hashForPageFile returns the fingerprint of the file this page is served as.
//
// The key is the path the CHASSIS sent, which is what the adapter keys the map
// by — `pages.url` with its leading slash removed, via the one shared
// definition (`PageFilePathFromURL`, DGH-006/125). That helper REFUSES a url
// carrying a fragment or query, which is correct here and not merely defensive:
// idea.uk has a live page row whose url is "/tools.html#audience-check" while a
// DIFFERENT page owns "/tools.html", so a stripped fragment would fingerprint
// one page with another page's bytes.
func hashForPageFile(files map[string]string, pageURL string) string {
	if len(files) == 0 {
		return ""
	}
	path, ok := datahelpers.PageFilePathFromURL(pageURL)
	if !ok {
		return ""
	}
	return files[path]
}

// stringMap narrows the JSON-decoded map[string]interface{} the resolver returns.
func stringMap(v interface{}) map[string]string {
	m, ok := v.(map[string]interface{})
	if !ok || len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, val := range m {
		if s, ok := val.(string); ok && s != "" {
			out[k] = s
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// pageURLForHash reads the page's url so the fingerprint can be looked up by the
// path the file was committed under. A failed read yields "" and therefore no
// fingerprint — never a wrong one.
func pageURLForHash(ctx context.Context, db *sql.DB, pageID uuid.UUID) string {
	if db == nil {
		return ""
	}
	var url string
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(url, '') FROM pages WHERE id = $1`, pageID).Scan(&url); err != nil {
		return ""
	}
	return url
}

// collectUniqueValue walks a decoded-JSON subtree for every occurrence of a key
// and returns its value only if every occurrence AGREES.
//
// It exists because the shared whole-tree search resolves conflicts in this
// build instead of refusing them (see the file header). It is deliberately
// small and deliberately dumber than that search: no unwrap patterns, no
// aliases, no ranking — just "find this key, and tell me if the answers
// disagree". Ranking is only needed when you intend to PICK a winner, and the
// whole point here is that we will not.
//
// conflict=true beats found=true: a caller that ignored the conflict flag would
// get the shallowest value, which is precisely the guess this avoids.
func collectUniqueValue(node interface{}, key string, depth int) (value interface{}, found bool, conflict bool) {
	if depth > 12 {
		return nil, false, false
	}
	switch v := node.(type) {
	case map[string]interface{}:
		// Sorted keys so a conflict is reported identically on every run — an
		// intermittent refusal would be worse than a consistent one.
		names := make([]string, 0, len(v))
		for k := range v {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			child := v[k]
			if k == key {
				value, found, conflict = merge(value, found, conflict, child, true, false)
				// Do NOT descend into a matched key: its own sub-tree cannot
				// hold a more authoritative copy of itself.
				continue
			}
			cv, cf, cc := collectUniqueValue(child, key, depth+1)
			value, found, conflict = merge(value, found, conflict, cv, cf, cc)
		}
	case []interface{}:
		for _, item := range v {
			cv, cf, cc := collectUniqueValue(item, key, depth+1)
			value, found, conflict = merge(value, found, conflict, cv, cf, cc)
		}
	}
	return value, found, conflict
}

// merge folds one candidate into the running answer.
func merge(cur interface{}, curFound, curConflict bool, next interface{}, nextFound, nextConflict bool) (interface{}, bool, bool) {
	if curConflict || nextConflict {
		return cur, curFound || nextFound, true
	}
	if !nextFound {
		return cur, curFound, curConflict
	}
	if !curFound {
		return next, true, false
	}
	if !reflect.DeepEqual(cur, next) {
		return cur, true, true
	}
	return cur, true, false
}
