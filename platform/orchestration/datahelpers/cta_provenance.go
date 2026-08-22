// FILE: platform/orchestration/datahelpers/cta_provenance.go
//
// RECORDED CTA destination provenance — the replacement for the DERIVED
// provenance of LNK-033 (`storedCTADestinationIsAuthored`), which answered
// "did a person choose this link, or did we?" by inference: *the resolver is
// incapable of producing a utility-area destination, so a stored one must be
// authored*. That inference is true today and `bugs_open/308` cannot be fixed
// without making it false — closing 308 means letting labels resolve INTO
// utility pages, at which point both keep-branches would freeze the resolver's
// own output for ever. The owner ruled on 2026-08-18 (recorded in
// `bugs_open/308`): record the provenance, then widen. This is the record.
//
// THE STAMP IS VALUE-BOUND, and that is the whole design, not a detail. It
// records WHICH url the resolver minted for a field, never a bare "the
// resolver wrote this". A boolean gets the important case backwards:
//
//   - a REPLACE writer (7 of RFC_042's 9 content_data writers) drops value and
//     stamp together, so its values read authored — correct, it IS authoring;
//   - a MERGE writer (the section editor's field update) writes a NEW value
//     while the OLD stamp survives. Value-bound, the stamp names a different
//     url, mismatches, and the human's edit reads authored. Presence-bound, the
//     surviving stamp would licence the recompute to clobber that edit — which
//     is precisely `bugs_open/248`, the bug the keeps exist to prevent.
//
// So eight of the nine writers are correct with no code change at all. The
// failure that remains — a path carrying a minted VALUE without its STAMP —
// yields a false "authored", i.e. a link frozen at its current destination.
// That is the safe direction: the frozen value is a valid page, and the label
// match runs AHEAD of every keep in both writers, so a frozen link whose copy
// names a different page is still repaired.
//
// NO BACKFILL, deliberately: an absent stamp means "not recorded", and every
// caller treats that exactly as today's derived rule did. Defaulting the other
// way would freeze every CTA in the fleet rather than the 200 findings 308 is
// about.
//
// Lives in datahelpers because BOTH writers (`setCTAField` at build time,
// `applyCTARecompute` at repair time) and the discovery checks need it, and
// actions imports datahelpers, never the reverse — the same placement reason as
// links_tel.go (LNK-034) and ctafields.go.

package datahelpers

// CTAMintedKey is the reserved `page_components.content_data` key recording the
// CTA destinations a resolver minted, as {url_field: minted_url}.
//
// The `__` prefix marks it as machinery rather than content. Two properties of
// that choice are load-bearing and were measured before it was made
// (2026-08-22): the namespace is empty fleet-wide (zero `__`-prefixed keys in
// any content_data), and an UNDECLARED key does survive into stored
// content_data — 16 rows across four component functions carry a
// `*_target_title` their input_schema does not declare.
//
// It is deliberately NOT schema-declared, which also keeps it invisible to
// `cmd/content-loss-check`'s differ (scoped to schema-declared non-LLM keys),
// so stamp churn can never be read as key loss by the RFC_042/355 work.
const CTAMintedKey = "__cta_minted"

// SetCTAMinted records that `url` was written into `field` by a resolver.
//
// A no-op for an empty field or url: there is no such thing as a stamp for a
// value that was not written, and an empty stamp entry would be
// indistinguishable from a mint of "" by CTAMintedCovers.
func SetCTAMinted(resolved map[string]interface{}, field, url string) {
	if resolved == nil || field == "" || url == "" {
		return
	}
	minted, _ := resolved[CTAMintedKey].(map[string]interface{})
	if minted == nil {
		minted = make(map[string]interface{}, 2)
		resolved[CTAMintedKey] = minted
	}
	minted[field] = url
}

// CTAMintedCovers reports whether the stamp in `contentData` names `url` as the
// value a resolver minted for `field`.
//
// FALSE for an absent stamp — that is the no-backfill guarantee: an unstamped
// row behaves exactly as it did before this mechanism existed.
//
// FALSE for a stamp naming a DIFFERENT url — that is the value-binding, and it
// is what makes a human's edit under a stale stamp read as authored.
//
// Comparison goes through NormalizePagePath — the same normaliser the keeps,
// the detector and the phantom checks already share, so no fourth spelling of
// "is this the same page" enters the estate. It equates the spellings of one
// page: "/contact/index.html", "/contact/" and "/contact" all reduce to
// "/contact", and a query or fragment is dropped.
//
// ⚠ It does NOT equate "/contact.html" with "/contact/" — those are two
// different pages here, and only ctaExcludedDestination collapses them, purely
// to decide the AREA. That boundary is deliberate and pinned by a test: a stamp
// vouching across it would let a mint of "/contact/" mark an authored
// "/contact.html" as the resolver's own and make it recomputable.
func CTAMintedCovers(contentData map[string]interface{}, field, url string) bool {
	if contentData == nil || field == "" || url == "" {
		return false
	}
	minted, _ := contentData[CTAMintedKey].(map[string]interface{})
	if minted == nil {
		return false
	}
	stamped, _ := minted[field].(string)
	if stamped == "" {
		return false
	}
	return NormalizePagePath(stamped) == NormalizePagePath(url)
}

// SeedCTAMinted copies the stored mint record into `resolved` before a section's
// CTA fields are processed, so that a field NOT rewritten this pass keeps its
// stamp.
//
// It exists because both persist paths merge SHALLOWLY — the rerender builds
// `mergedContent = stored ⊕ fresh` key by key, and the build path's
// resolved_data merges last the same way. CTAMintedKey holds a nested map, so a
// resolved_data carrying a stamp for ONE field REPLACES the whole stored record
// and silently drops the sibling's. On a two-CTA component (hero,
// call-to-action, archetype-combinations, gauntlet-cta) that is a live case
// every time one slot is re-minted and the other is kept: the kept slot's stamp
// disappears, it reads as authored next cycle, and it freezes. The bug this
// mechanism exists to fix, reintroduced by the mechanism.
//
// Seeding is safe ONLY because the stamp is value-bound, and that is worth
// stating: a seeded entry naming a url the field no longer carries does not
// cover it, so a stale seed reads as "authored" rather than as a false mint. A
// presence-bound stamp could not be seeded this way — it would licence the
// recompute against whatever value the field now holds.
//
// The copy is deliberate: writing into the stored map would mutate the caller's
// content_data in place, and both writers still read that map afterwards.
func SeedCTAMinted(resolved, stored map[string]interface{}) {
	if resolved == nil || stored == nil {
		return
	}
	storedMinted, _ := stored[CTAMintedKey].(map[string]interface{})
	if len(storedMinted) == 0 {
		return
	}
	seeded, _ := resolved[CTAMintedKey].(map[string]interface{})
	if seeded == nil {
		seeded = make(map[string]interface{}, len(storedMinted))
		resolved[CTAMintedKey] = seeded
	}
	for field, url := range storedMinted {
		if _, already := seeded[field]; !already {
			seeded[field] = url
		}
	}
}
