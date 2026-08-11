// FILE: platform/orchestration/actions/discovery_checks/check_page_canonical_collision.go
//
// bugs_open/080, fix candidate 3: two `pages` rows on one site whose canonical
// form collides — the same logical page existing twice because two creation
// surfaces derived its identity differently. `pages` is unique on
// (site_id, name) and has NO unique index on url, so a disagreeing name does
// not conflict: it inserts a second row, and both can deploy. robot-hands.com
// served two live pairs this way (/news.html beside /news/index.html with an
// identical <title>, and /gripper-catalog.html beside
// /gripper-catalog/index.html) for weeks with nothing noticing.
//
// TWO GROUPING SIGNALS, both required — measured against the live pairs
// (2026-08-03): running each row through datahelpers.CanonicalisePage catches
// the /news pair (both rows canonicalise to name "news-index") but NOT
// /gripper-catalog, whose stray is typed `content` and canonicalises to
// itself; the URL path-key (url with /index.html | .html stripped) catches
// both. A pair can also fire BOTH signals, so rows are union-merged across the
// two and ONE finding is emitted per connected group — not one per signal.
//
// WHAT IT FILES. A collision between two ACTIVE rows is a decision — which row
// survives — not a job, so the item is `needs_human_review` with no handler,
// the bugs_closed/081 idiom, and the spec carries the decided section-index
// family convention so the human has the ruling in hand. Groups with fewer
// than two active rows (a planned stray beside a live page) are reported as
// findings only. A group a human has already ruled on (a terminal wont_fix /
// rejected item under the same key) is not re-filed: without that suppression
// the same true-in-the-DB collision would re-file on every sweep, for ever.
//
// The check mutates nothing. Retiring a live page has no mechanism today
// (bugs_open/098 → RFC_011), and re-typing a shipped row was deliberately
// declined by 081 — visibility is the whole of this check's job.
package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() {
	Register(&PageCanonicalCollisionCheck{})
	RegisterVerifier("page_canonical_collision", VerifyPageCanonicalCollisionResolved)
}

type PageCanonicalCollisionCheck struct{}

func (c *PageCanonicalCollisionCheck) Name() string { return "page_canonical_collision" }

// collisionPage is one `pages` row with its two derived signals.
type collisionPage struct {
	ID        string
	Name      string
	URL       string
	PageType  string
	Build     string
	Status    string
	PathKey   string
	CanonName string // "" — no name signal for this row
}

// pageURLPathKey and collisionCanonName are this check's two signals. Both now
// delegate to datahelpers, which holds the one definition (bugs_open/215): the
// plan/realised reconciler needed the same two keys, and a third hand-copy of a
// rule derived from CanonicalisePage is what the reuse note in IMG-070 says to
// stop at. The local names stay so this file's tests keep pinning the behaviour
// at the point of use.
func pageURLPathKey(url string) string { return datahelpers.PagePathKey(url) }

func collisionCanonName(name, pageType string) string {
	return datahelpers.PageCanonicalNameForRow(name, pageType)
}

// groupCollisions union-merges pages connected by either signal and returns
// only groups of two or more, members sorted by name. Pure — tests drive it
// directly.
func groupCollisions(pages []collisionPage) [][]collisionPage {
	parent := make([]int, len(pages))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	union := func(a, b int) { parent[find(a)] = find(b) }

	byKey := map[string]int{} // signal key → first index seen
	link := func(key string, i int) {
		if key == "" {
			return
		}
		if j, seen := byKey[key]; seen {
			union(i, j)
		} else {
			byKey[key] = i
		}
	}
	for i, p := range pages {
		link("path:"+p.PathKey, i)
		link("name:"+p.CanonName, i)
	}

	grouped := map[int][]collisionPage{}
	for i, p := range pages {
		r := find(i)
		grouped[r] = append(grouped[r], p)
	}
	var out [][]collisionPage
	for _, g := range grouped {
		if len(g) < 2 {
			continue
		}
		sort.Slice(g, func(a, b int) bool { return g[a].Name < g[b].Name })
		out = append(out, g)
	}
	sort.Slice(out, func(a, b int) bool { return out[a][0].Name < out[b][0].Name })
	return out
}

// collisionGroupKey derives the order-independent discriminator for a group:
// the shared path-key when the group has exactly one, else the sorted distinct
// canonical names. Stable no matter which row is encountered first.
func collisionGroupKey(group []collisionPage) string {
	paths := map[string]bool{}
	names := map[string]bool{}
	for _, p := range group {
		paths[p.PathKey] = true
		if p.CanonName != "" {
			names[p.CanonName] = true
		}
	}
	if len(paths) == 1 {
		for k := range paths {
			return k
		}
	}
	var ns []string
	for n := range names {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	if len(ns) > 0 {
		return strings.Join(ns, "+")
	}
	// Degenerate: no name signal and several paths — fall back to sorted names.
	var raw []string
	for _, p := range group {
		raw = append(raw, p.Name)
	}
	sort.Strings(raw)
	return strings.Join(raw, "+")
}

func (c *PageCanonicalCollisionCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT id::text, name, COALESCE(url, ''), COALESCE(page_type, ''),
		       COALESCE(build_status, ''), COALESCE(status, '')
		FROM pages
		WHERE site_id = $1
		  AND COALESCE(url, '') <> ''
		ORDER BY name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("load pages: %w", err)
	}
	defer rows.Close()

	var pages []collisionPage
	for rows.Next() {
		var p collisionPage
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.PageType, &p.Build, &p.Status); err != nil {
			return nil, err
		}
		p.PathKey = pageURLPathKey(p.URL)
		p.CanonName = collisionCanonName(p.Name, p.PageType)
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &CheckResult{}
	openKeys := map[string]bool{} // item_keys this run still stands behind

	for _, group := range groupCollisions(pages) {
		key := collisionGroupKey(group)
		itemKey := "page_canonical_collision:" + key

		members := make([]map[string]interface{}, 0, len(group))
		pathKeys := map[string]bool{}
		canonNames := map[string]bool{}
		activeCount := 0
		for _, p := range group {
			members = append(members, map[string]interface{}{
				"page_id":      p.ID,
				"name":         p.Name,
				"url":          p.URL,
				"page_type":    p.PageType,
				"build_status": p.Build,
				"status":       p.Status,
			})
			pathKeys[p.PathKey] = true
			if p.CanonName != "" {
				canonNames[p.CanonName] = true
			}
			if p.Status == "active" {
				activeCount++
			}
		}

		finding := map[string]interface{}{
			"check":        "page_canonical_collision",
			"group_key":    key,
			"members":      members,
			"active_count": activeCount,
		}
		result.Findings = append(result.Findings, finding)

		// A collision needs two claimants: a planned or archived stray beside
		// one live page is a finding, not a decision item.
		if activeCount < 2 {
			continue
		}

		// A human has already ruled on this exact group ("keep both") —
		// re-filing it every sweep would train them to ignore the type.
		var ruled bool
		if err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND item_key = $2
				  AND status IN ('wont_fix', 'rejected')
			)
		`, dctx.SiteID, itemKey).Scan(&ruled); err != nil {
			return nil, fmt.Errorf("prior-ruling lookup for %s: %w", itemKey, err)
		}
		if ruled {
			finding["suppressed_by_prior_ruling"] = true
			continue
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":           "page_canonical_collision",
			"group_key":       key,
			"members":         members,
			"path_keys":       sortedKeys(pathKeys),
			"canonical_names": sortedKeys(canonNames),
			"active_count":    activeCount,
			"bug":             "bugs_open/080",
			"convention": "The decided shape is the section-index family convention " +
				"(page_canonical.go, doc 029 Phase 0, bugs_closed/015): one row per logical page at " +
				"(name=<section>-index, url=/<section>/index.html) with the flavour preserved as page_type. " +
				"Decide which row survives; retiring the other from the deployed site has no mechanism yet " +
				"(bugs_open/098 -> RFC_011), so record the decision here either way.",
		})

		names := make([]string, 0, len(group))
		for _, p := range group {
			names = append(names, p.Name+" @ "+p.URL)
		}
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "page_canonical_collision",
			Severity: "medium",
			Summary: fmt.Sprintf("%d live pages claim one canonical identity (%s) — needs a human decision: %s",
				activeCount, key, strings.Join(names, " vs ")),
			SpecJSON:     string(specJSON),
			Priority:     40,
			HandlerAgent: "", // a decision, not a job — the bugs_closed/081 idiom
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		})
		openKeys[itemKey] = true
	}

	// Retraction (RFC_010 seam, opt-in): an open collision item whose group no
	// longer has two active claimants was fixed out of band — close it with
	// the reason recorded rather than leaving it to sit for ever.
	staleRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT item_key FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'page_canonical_collision'
		  AND status NOT IN ('complete', 'failed', 'verified', 'rejected', 'wont_fix', 'unresolved', 'cancelled')
		  AND item_key IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("load open collision items: %w", err)
	}
	defer staleRows.Close()
	for staleRows.Next() {
		var k string
		if err := staleRows.Scan(&k); err != nil {
			return nil, err
		}
		if !openKeys[k] {
			result.Resolved = append(result.Resolved, ResolvedFinding{
				// ItemType is REQUIRED by the runner's resolveWorkItems — an
				// entry without it matches nothing, silently. Proven live
				// 2026-08-05: the first real retraction left both items open
				// while this arm reported clean; the check-side test was green
				// because it pins only this side of a two-sided contract.
				ItemType: "page_canonical_collision",
				ItemKey:  k,
				Reason:   "re-ran page_canonical_collision: the group behind this item no longer has two active claimants",
			})
		}
	}
	if err := staleRows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// verifier
// ---------------------------------------------------------------------------

// VerifyPageCanonicalCollisionResolved re-runs the check's own predicate for
// the item's group signature and reports resolved only when no signature key
// still has two or more active rows.
//
// Unlike content_duplication, rows disappearing IS this item's expected fix
// shape — the remedy is retiring one claimant. What stays ambiguous is the
// whole SITE vanishing: count=0 would then be vacuous, so that is an error
// (the caller fails open and records it), mirroring the existence guard
// convention.
func VerifyPageCanonicalCollisionResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	var siteExists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM sites WHERE id = $1)`, target.SiteID,
	).Scan(&siteExists); err != nil {
		return VerifyResult{}, fmt.Errorf("site existence check: %w", err)
	}
	if !siteExists {
		return VerifyResult{}, fmt.Errorf("site %s no longer exists — cannot distinguish a fix from a deleted site", target.SiteID)
	}

	pathKeys := specStringSet(target.Spec, "path_keys")
	canonNames := specStringSet(target.Spec, "canonical_names")
	if len(pathKeys) == 0 && len(canonNames) == 0 {
		return VerifyResult{}, fmt.Errorf("item %s spec carries no group signature (path_keys / canonical_names)", target.ItemID)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT name, COALESCE(url, ''), COALESCE(page_type, '')
		FROM pages
		WHERE site_id = $1
		  AND COALESCE(url, '') <> ''
		  AND COALESCE(status, '') = 'active'
	`, target.SiteID)
	if err != nil {
		return VerifyResult{}, err
	}
	defer rows.Close()

	pathCounts := map[string]int{}
	nameCounts := map[string]int{}
	for rows.Next() {
		var name, url, pageType string
		if err := rows.Scan(&name, &url, &pageType); err != nil {
			return VerifyResult{}, err
		}
		if pk := pageURLPathKey(url); pathKeys[pk] {
			pathCounts[pk]++
		}
		if cn := collisionCanonName(name, pageType); cn != "" && canonNames[cn] {
			nameCounts[cn]++
		}
	}
	if err := rows.Err(); err != nil {
		return VerifyResult{}, err
	}

	for k, n := range pathCounts {
		if n >= 2 {
			return VerifyResult{
				Resolved: false,
				Detail:   fmt.Sprintf("%d active pages still claim path %s", n, k),
			}, nil
		}
	}
	for k, n := range nameCounts {
		if n >= 2 {
			return VerifyResult{
				Resolved: false,
				Detail:   fmt.Sprintf("%d active pages still canonicalise to name %q", n, k),
			}, nil
		}
	}
	return VerifyResult{
		Resolved: true,
		Detail:   "no signature key of this group has two active claimants any more",
	}, nil
}

func specStringSet(spec map[string]interface{}, key string) map[string]bool {
	out := map[string]bool{}
	arr, _ := spec[key].([]interface{})
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out[s] = true
		}
	}
	return out
}
