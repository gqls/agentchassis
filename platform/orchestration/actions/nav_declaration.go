// FILE: platform/orchestration/actions/nav_declaration.go
//
// A SITE DECLARES ITS OWN HEADER — bugs_open/407, at the owner's direction.
//
// THE DEFECT THIS ENDS. Membership of the primary nav was decided by a hardcoded,
// FLEET-WIDE list of page NAMES (`navPriorityTier`, populate_nav_tables_action.go),
// and `nav_order` only sorted WITHIN a tier. So a site's own page at nav_order 1
// sat behind every tier-2 page at nav_order 100, and the only ways to promote it
// were to rename it to a name on the list, edit the Go list for all 51 sites, or
// displace enough higher-tier pages that it arrived by elimination. The owner
// asked for finetuning.uk's £99 offer page — its primary conversion route — to go
// in the header, named four pages he was happy to displace, and got `Pricing` in
// the slot instead. Nothing warned. The page row said `in_header = true`
// throughout.
//
// His words, which are also the design: *"perhaps label slots and page names
// rather than having to search from names it considers important"*.
//
// ── WHY A LONGER FLEET LIST WAS REFUSED, AND IT IS A MEASUREMENT ───────────
//
// The previous remedy for exactly this is INSIDE the broken thing. Three
// site-specific page names — model-directory, adoption-tracker, protocol-tracker —
// were hardcoded into the fleet-wide tier-2 map, with a comment explaining that as
// tier 3 they would be "the first thing dropped when max_header_items truncates".
// [MEASURED 2026-08-26, at the served page] ai-agent-orchestration.com's header
// holds exactly 8 links and ONE of those three names is among them; its two
// siblings are not. The documented remedy is in the source today and the pages it
// was written for are still missing.
//
// ── WHY AN ORDERED LIST AND NOT A PRIORITY HINT ────────────────────────────
//
// Half the measured damage is SAME-TIER ties, not higher tiers.
// [MEASURED 2026-08-26] gaswholesalers.com's header holds why-gas-wholesalers,
// how-pricing-works and service-areas — all tier 3, all at nav_order 100 — while
// tier-3 pricing-transparency at nav_order 100 is absent; the tie is broken by
// load order. A design that only let a site RAISE a page's tier would fix
// finetuning.uk and leave gaswholesalers exactly as it is. The declaration must
// give a TOTAL ORDER, so it is an ordered array.
//
// ── WHERE IT LIVES, AND A CORRECTION THE COUNCIL FORCED ────────────────────
//
// > **CORRECTED 2026-08-26, before this shipped.** The first version of this file
// > put the declaration in `sites.settings->'nav'`, on a rationale that stated
// > "there is NO site_specs table". **That was false**, and it was false because I
// > read a `\dt site*` listing that had been truncated at twenty rows —
// > `site_specs` sorts after `site_plan_pages`. The council's `reuse_agent` and
// > `prior_art_librarian` seats both objected at medium that the platform already
// > has a purpose-built per-site declaration mechanism and the rationale never
// > weighed it. They were right, and the measurement settles it rather than the
// > argument: **`site_specs` aspect `site_config` ALREADY CARRIES PER-SITE HEADER
// > CONFIG** — `chrome.header_cta_url` and `chrome.header_cta_label` on oufe.com,
// > `chrome.compliance_lines` on two more `[MEASURED 2026-08-26]`. Putting header
// > SLOTS in a different table from header CTA would have been a second home for
// > one concern, which is this bug's own pathology one level up.
//
// So the declaration is `site_specs.data -> 'chrome' -> 'header_slots'`, an
// ordered array of page names, beside the header config already there, with
// `chrome.max_header_items` alongside it.
//
// What that home gives, which the settings column could not:
//
//	is_current / superseded_at  a new declaration SUPERSEDES rather than clobbers,
//	                            and the previous one is still readable
//	pinned                      the existing "an agent may not overwrite this" flag
//	source / created_by / notes provenance — an operator declaration should record
//	                            who asked for this header and why
//	WriteSiteSpecAction         a real writer, so this is settable through the
//	                            platform rather than only by hand SQL
//
// And it survives the two erasures a declaration must survive:
//
//	a nav REBUILD  — populate_nav_tables DELETEs every site_nav_items and
//	                 site_nav_groups row for the site and re-derives them, so the
//	                 declaration cannot live in the nav tables: they are derived.
//	a RE-PLAN      — sync_pages_to_db upserts `in_header = EXCLUDED.in_header` and
//	                 `nav_order = EXCLUDED.nav_order` (site_db_actions.go), so it
//	                 cannot live on the page rows either: the next re-plan rewrites
//	                 them from the plan and the silence returns.
//
// ⚠ `siteSpecDeepMerge` REPLACES a non-map value wholesale and only recurses into
// maps. `header_slots` is an ARRAY, so any writer that includes that key replaces
// the list entire — which is what you want from an operator re-declaring, and what
// you do NOT want from an agent writing `site_config` for some other reason. No
// writer emits the key today; `pinned` is the guard if one ever does.
//
// Keyed on page NAME, not id: `pages` has UNIQUE (site_id, name), so a name
// resolves to at most one row per site, and a name survives a re-plan that drops
// and re-creates the row. The owner also speaks in names.
//
// ── OPT-IN, UNSAFE SIDE OFF ────────────────────────────────────────────────
//
// Per the owner ruling of 2026-08-02 §2. The declared branch is unreachable
// unless `header_slots` is a non-empty array. [MEASURED 2026-08-26] 0 of 51 sites
// carry a `nav` key, so the first roll is a provable no-op for the whole fleet —
// and `classifyPagesForNav` keeps its exact signature as a wrapper, so
// nav_membership_test.go (which pins bugs_open/149 A2's invariant) passes with a
// ZERO DIFF.
//
// ── WHAT A DECLARATION OVERRIDES, AND WHAT IT DOES NOT ─────────────────────
//
// It overrides the tier table, the same-tier nav_order tie, `neverPrimaryTypes`
// and the child-URL bar. Every one of those is a fleet-wide DEFAULT, and the whole
// point of this change is that a default may not outrank the site's own word. The
// UNDECLARED path keeps all of them, unchanged.
//
// It does NOT override the system-page exclusion (404, sitemap, robots) or the
// legal group. Those are correctness rather than preference, and a site that
// declares one is told so rather than obeyed.
//
// ⚠ A DECLARED PAGE IS PLACED EVEN IF `in_header` IS FALSE, and that is
// load-bearing rather than convenient. A declaration subordinate to the flags
// would be silently defeated by the next re-plan — which is this bug, returning.
// The disagreement is REPORTED so the plan can be corrected.
//
// ── THREE FURTHER HOMES CONSIDERED AND REJECTED, because the council asked ─
//
//	site_nav_groups   the prior-art seat asked whether its group_keys could carry
//	                  the declaration. No: populate_nav_tables DELETEs every
//	                  site_nav_groups row for the site on every rebuild, alongside
//	                  site_nav_items. It is derived by exactly the same statement,
//	                  so it fails for exactly the same reason.
//	sites.settings    where the first draft put it. Rejected once site_specs was
//	                  found — see the correction above.
//	pages columns     rejected on the re-plan measurement above.
//
// ── THE READ IS OUTSIDE THE REBUILD TRANSACTION, AND THAT IS ACCEPTED ───────
//
// populate_nav_tables reads the declaration before it opens the transaction that
// deletes and rebuilds the nav rows, so a declaration changed in that window is
// applied one rebuild late. Accepted rather than overlooked: the declaration only
// REORDERS — it destroys nothing and adds nothing that a later rebuild cannot
// correct — and the next rebuild converges. The bug_historian seat raised it at
// low severity against the delete-and-re-derive shape, which is the real hazard
// there and is guarded by nav_prune_floor.go, not by this.
//
// ⚠ THE TEMPTING WRONG FIX IS AN EXHAUSTIVE READING. If `header_slots` meant "the
// header is exactly this and nothing else", a new page gaining `in_header` would
// never appear until an operator edited settings — recreating this bug's silence
// class from the other side. Declared slots LEAD; undeclared candidates fill what
// remains by the existing sort. An opt-in `exclusive` flag can be added later
// without invalidating any declaration written under this rule.

package actions

import (
	"encoding/json"
	"strings"
)

// navDeclarationSource records how the declaration was arrived at, so a caller can
// tell "the site said nothing" from "the site said something unreadable". The
// vocabulary is structureFloorSettings', deliberately: an override that is
// silently ignored is a landmine for whoever set it.
const (
	navDeclSourceDefault    = "default"     // no `nav` key — the fleet default decides
	navDeclSourceSiteConfig = "site_config" // a usable declaration
	navDeclSourceInvalid    = "invalid"     // present and unreadable — the caller WARNs
)

// siteNavDeclaration is what a SITE says about its own header.
type siteNavDeclaration struct {
	// HeaderSlots are page NAMES, lower-cased and de-duplicated, in the order the
	// site wants them. Empty means the site has not spoken.
	HeaderSlots []string
	// MaxHeaderItems is 0 when the site has not declared one, in which case the
	// step config's value applies.
	MaxHeaderItems int
	Source         string
}

// Declared reports whether the site has said anything about slot ORDER. A site
// may declare only a cap, which is useful on its own and is not this.
func (d siteNavDeclaration) Declared() bool { return len(d.HeaderSlots) > 0 }

// EffectiveCap resolves the cap: the site's own value, else the step config's.
func (d siteNavDeclaration) EffectiveCap(stepCap int) int {
	if d.MaxHeaderItems > 0 {
		return d.MaxHeaderItems
	}
	return stepCap
}

// navDeclarationReport is returned in the action's result, never only logged.
// A declared name that resolves to nothing, or to a page the platform will not
// promote, is exactly the silence this bug is about — so it surfaces where an
// operator already looks, beside the prune floor's numbers.
type navDeclarationReport struct {
	Placed        []string `json:"placed"`
	Missing       []string `json:"missing"`        // declared, no such page in scope
	Ineligible    []string `json:"ineligible"`     // system or legal, with the reason
	FlagDisagreed []string `json:"flag_disagreed"` // declared but in_header=false
	Truncated     []string `json:"truncated_by_cap"`
}

// siteNavDeclarationFromSiteConfig parses the `site_config` aspect's `data`
// jsonb and returns what the site said about its own header. PURE — no database,
// no logger — so every shape is a table test rather than an integration test.
//
// A malformed value never silently becomes the default: Source is
// navDeclSourceInvalid and the caller says so loudly. Being ignored quietly is the
// defect this file exists to remove, and reproducing it one level up would be the
// same mistake with better formatting.
func siteNavDeclarationFromSiteConfig(raw []byte) siteNavDeclaration {
	out := siteNavDeclaration{Source: navDeclSourceDefault}
	if len(raw) == 0 {
		return out
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return siteNavDeclaration{Source: navDeclSourceInvalid}
	}
	chromeRaw, ok := data["chrome"]
	if !ok {
		return out
	}

	var chrome struct {
		HeaderSlots    []json.RawMessage `json:"header_slots"`
		MaxHeaderItems json.RawMessage   `json:"max_header_items"`
	}
	if err := json.Unmarshal(chromeRaw, &chrome); err != nil {
		// `chrome` present but not an object, or its members are the wrong shape.
		// NOTE this is deliberately NOT fatal for the site's other chrome keys —
		// header_cta_url and friends are read elsewhere and are none of our
		// business; we only report that OUR keys were unreadable.
		return siteNavDeclaration{Source: navDeclSourceInvalid}
	}

	invalid := false
	seen := map[string]bool{}
	for _, m := range chrome.HeaderSlots {
		var name string
		if err := json.Unmarshal(m, &name); err != nil {
			// A non-string member is skipped rather than fatal — one typo must not
			// discard the whole declaration — but it is not silent either.
			invalid = true
			continue
		}
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" || seen[name] {
			if name != "" {
				invalid = true // a duplicate is a mistake worth surfacing
			}
			continue
		}
		seen[name] = true
		out.HeaderSlots = append(out.HeaderSlots, name)
	}

	if len(chrome.MaxHeaderItems) > 0 && string(chrome.MaxHeaderItems) != "null" {
		n, ok := navCapFromJSON(chrome.MaxHeaderItems)
		if !ok || n < 1 {
			invalid = true
		} else {
			out.MaxHeaderItems = n
		}
	}

	switch {
	case invalid && len(out.HeaderSlots) == 0 && out.MaxHeaderItems == 0:
		// Nothing usable survived — this is not a default, it is a broken override.
		return siteNavDeclaration{Source: navDeclSourceInvalid}
	case invalid:
		// Partly usable. Report `invalid` so the operator is told, and keep what
		// parsed — discarding a good list because one member was mistyped would
		// punish the site for the platform's strictness.
		out.Source = navDeclSourceInvalid
	case len(out.HeaderSlots) > 0 || out.MaxHeaderItems > 0:
		out.Source = navDeclSourceSiteConfig
	}
	return out
}

// navCapFromJSON accepts a JSON number or a numeric string. Both spellings occur
// in hand-written settings, and refusing the string form would make the feature
// fail for the operator who typed the obvious thing.
func navCapFromJSON(raw json.RawMessage) (int, bool) {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		v = v*10 + int(r-'0')
	}
	return v, true
}

// ---------------------------------------------------------------------------
// Precedence
// ---------------------------------------------------------------------------

// navDeclarationEligibility says whether a declared page may be promoted, and why
// not when it may not. Only two refusals exist, and both are correctness rather
// than preference — everything else a declaration is allowed to override.
func navDeclarationEligibility(p pageNavInfo) (ok bool, reason string) {
	nameLower := strings.ToLower(p.Name)
	switch {
	case navSystemNames[nameLower]:
		return false, "system page"
	case navLegalNames[nameLower] || isLegalPage(nameLower):
		return false, "legal page — belongs in the legal group"
	}
	return true, ""
}

// applyNavDeclaration places the site's declared pages in the leading primary
// slots, in the site's order, ahead of every undeclared candidate.
//
// It is given the classifier's own buckets and returns the new primary list plus
// the pages that must be REMOVED from utility (a declared page that classification
// had sent to the footer must not also appear there). Everything it cannot place
// is reported rather than dropped.
func applyNavDeclaration(
	decl siteNavDeclaration,
	all []pageNavInfo,
	fallbackPrimary []pageNavInfo,
) (primary []pageNavInfo, promotedNames map[string]bool, report navDeclarationReport) {

	promotedNames = map[string]bool{}
	if !decl.Declared() {
		return fallbackPrimary, promotedNames, report
	}

	byName := make(map[string]pageNavInfo, len(all))
	for _, p := range all {
		byName[strings.ToLower(p.Name)] = p
	}

	for _, name := range decl.HeaderSlots {
		p, found := byName[name]
		if !found {
			// A declared name that resolves to nothing. Reported, never silent —
			// the operator who wrote it must be able to find out it did nothing.
			report.Missing = append(report.Missing, name)
			continue
		}
		if ok, reason := navDeclarationEligibility(p); !ok {
			report.Ineligible = append(report.Ineligible, name+" ("+reason+")")
			continue
		}
		if !p.InHeader {
			// Placed anyway — see the file header. Recorded so the plan that keeps
			// rewriting the flag can be corrected at its source.
			report.FlagDisagreed = append(report.FlagDisagreed, name)
		}
		primary = append(primary, p)
		promotedNames[name] = true
		report.Placed = append(report.Placed, name)
	}

	// Undeclared candidates fill what remains, in the order the tier table gives —
	// which is now exactly what that table is good at, a default for sites that
	// have not spoken.
	for _, p := range fallbackPrimary {
		if promotedNames[strings.ToLower(p.Name)] {
			continue
		}
		primary = append(primary, p)
	}
	return primary, promotedNames, report
}
