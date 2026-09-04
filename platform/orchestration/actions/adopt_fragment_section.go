package actions

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

// CONSTRUCTIVE ADOPTION — the phase-2 half of RFC_046 / bugs_open/357.
//
// THE DEFECT, in one sentence: a tool page arrives as one `<div class="tool-page">`
// fragment with no `<section>`, the whole fragment is stored as a single section
// named `"section"` (the sentinel for *identity unknown*), and that sentinel is
// then replaced by `planned[Position-1]` from `pages.sections` — so a page whose
// plan lists `hero` first ends up with a row declaring itself the shared `hero`
// component while storing 15KB of interactive JavaScript. Every mechanism keyed
// on the component then reasons about a hero that is not there: the schema check
// files "missing required field: headline" about a page serving its own <h1>, and
// no repair route can act.
//
// WHAT THIS CHANGES, AND — JUST AS IMPORTANT — WHAT IT DOES NOT.
//
// It does NOT touch the slot name, and it does not touch `pages.sections`. That
// restraint is the whole design, not modesty. `slot_name` and `component_id` are
// different facts read by different consumers: `component_id` joins to
// `content_components` (schema checks, regeneration, ContentDataCanFillTemplate)
// and is where this bug's damage lives, while `slot_name` is the page-local
// positional name that Layer 2's carry-forward matches on with EXACT STRING
// EQUALITY and no fallback. Rename a slot and the match fails, the rebuild takes
// the re-append arm, and the page silently gains a section while the tool moves —
// LANDMINES, "CORRECTING a page_components.slot_name makes the NEXT rebuild
// append the tool beside a freshly generated section". This lane's first council
// round was rejected for exactly that. So the landmine is not managed here; it is
// never armed.
//
// WHY THIS IS NOT A SIXTH INFERENCE. RFC_046's finding is that identity is
// guessed five ways and recorded none, and that a sixth guess is not the fix.
// Nothing here is guessed. The fragment is bound to a component whose template is
// the identity function `{{.body}}`, with the fragment itself as `content_data.body`
// — and the binding is only made after RENDERING that template and checking the
// output is byte-identical to the bytes being stored. The component provably
// produces them. That also earns the row a genuine RFC_046 provenance stamp
// through the existing resolver, with no new stamping machinery, because the
// adoption is a real RenderTemplate call and the seam reports its digest.
//
// WHEN IT CANNOT: `component_id` stays NULL and the row is honestly unknown. That
// is the state RFC_046 asks for and it is strictly better than today's confident
// wrong answer. The page serves identically either way — assembly emits
// `rendered_html` regardless of what the row claims to be.
//
// OPT-IN, DEFAULT OFF (owner ruling 2026-08-02 §2: new authority on a shared seam
// ships as a field whose unsafe default is OFF, not as a documented contract).
// `save_page_sections` is the single INSERT every composition path flows through,
// so a behaviour change here reaches every carrier at once. The flag makes arming
// a separate, reversible decision that a reviewer of the CALLER can see.
//
// ~~with SIX live carriers as of 2026-08-23~~ — CORRECTED 2026-09-04, and the
// correction is worth more than the number. `[MEASURED 2026-09-04]` there are
// THREE, under this predicate:
//
//	SELECT ad.type, s.key FROM agent_definitions ad,
//	  LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key,value)
//	 WHERE s.value->>'action' = 'save_page_sections'
//	   AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
//	-- page-build-handler, page-rerender, tool-recreation-handler; one `save_sections` step each
//
// Six does not reconcile with ANY filtering of that query (dropping each of
// is_active, is_snapshot and deleted_at in turn still gives 3), nor with the Go
// side, which has exactly one entry point and one registry entry. So it either
// counted something this comment does not name, or three carriers have gone since
// — and NEITHER can be checked now, which is the whole point.
//
// ⚠ THE DATE WAS PRESENT AND DID NOT SAVE ANYONE. The owner's rule of 2026-08-22
// is that a count carries the date it was counted; this one did, and a lane still
// quoted it to a peer as current on 2026-09-04 (bugs_open/479). A date makes
// staleness *detectable*; only a PREDICATE makes the census *re-runnable*. Write
// the query next to the number, not just the day.
//
// ⚠ THE FLAG GOVERNS TWO COUPLED HALVES AND MUST GOVERN BOTH. Adoption alone does
// not survive a rebuild: Layer 2 splices the stored tool into the incoming
// section, which carries the PLAN's identity, so the next rebuild would re-mint
// the hero binding over the adopted row. Carrying the stored identity with the
// stored bytes is what makes adoption stick. Neither half is useful alone, which
// is why they share a key rather than having one each.
const adoptFragmentsKey = "adopt_unidentified_fragments"

// adoptedFragmentFunction is the seeded component whose template is exactly
// `{{.body}}`. Resolved by `function`, like every other component lookup in this
// file, and global — content_components lookups here carry no site filter, so one
// seeded row serves the fleet.
const adoptedFragmentFunction = "adopted-fragment"

// adoptFragmentSection binds a fallback-adopted fragment to the adopted-fragment
// component, having PROVEN that component reproduces the fragment's bytes.
//
// Returns true only when the row now states something true about itself. On every
// other path it returns false having changed nothing, and the caller leaves the
// section unidentified rather than falling back to a guess — falling back to a
// guess is the bug.
func adoptFragmentSection(ctx context.Context, db *sql.DB, s *SectionData, logger *zap.Logger) bool {
	if db == nil || s == nil || s.HTML == "" {
		return false
	}

	var componentID, template string
	err := db.QueryRowContext(ctx, `
		SELECT id::text, html_template FROM content_components
		WHERE function = $1 AND is_active = true
		LIMIT 1
	`, adoptedFragmentFunction).Scan(&componentID, &template)
	if err != nil {
		// Not seeded (or unreadable). The honest outcome is an unidentified row,
		// which is what happens with this whole mechanism switched off.
		logger.Info("adopt fragment: no adopted-fragment component available; leaving the section unidentified",
			zap.String("function", adoptedFragmentFunction), zap.Error(err))
		return false
	}

	// THE PROOF, and the reason this is adoption rather than assertion. Render the
	// component with the fragment as its body and require the output to be the
	// fragment. It is byte-identical by construction — the template is the identity
	// function and RenderTemplate uses text/template, which does not escape — but
	// "by construction" is an argument, and this is a check. It also means that if
	// anyone ever edits the seeded template into something that wraps or escapes,
	// adoption stops rather than silently storing bytes the component would not
	// reproduce.
	// No instance token is bound, deliberately (bugs_open/283's check flags this
	// shape and is right to): this render's output is COMPARED and thrown away, it
	// never reaches page_components.rendered_html and is never served. The template
	// is the identity function, so it contains no {{.InstanceID}} to render empty —
	// and if one were ever introduced, the byte comparison below would REFUSE the
	// adoption rather than store a collision. Recorded in pattern-check.py's
	// INSTANCE_TOKEN_ALLOWED with that reason.
	// No instance token is bound here, deliberately — bugs_open/283's pattern check
	// flags this shape and is right to ask. This render's output is COMPARED and
	// thrown away: it never reaches page_components.rendered_html and is never
	// served, which is the same standing as cmd/component-render-check's offline
	// lint. The template is the identity function, so it holds no {{.InstanceID}}
	// to render empty; and if one were ever introduced, the byte comparison below
	// REFUSES the adoption rather than storing a collision. Recorded with that
	// reason in pattern-check.py's INSTANCE_TOKEN_ALLOWED.
	rc := &RenderContext{ContentData: map[string]interface{}{"body": s.HTML}}
	rendered, _, _, renderErr := RenderTemplate(template, rc, logger)
	if renderErr != nil {
		logger.Warn("adopt fragment: the adopted-fragment template failed to render; leaving the section unidentified",
			zap.Error(renderErr))
		return false
	}
	if rendered != s.HTML {
		logger.Warn("adopt fragment: the adopted-fragment template did not reproduce the fragment byte-for-byte "+
			"— refusing to bind bytes to a component that would not regenerate them",
			zap.Int("fragment_bytes", len(s.HTML)),
			zap.Int("rendered_bytes", len(rendered)))
		return false
	}

	s.ComponentID = componentID
	if s.ContentData == nil {
		s.ContentData = map[string]interface{}{}
	}
	s.ContentData["body"] = s.HTML
	// Earned, not asserted: the digest describes the template that just produced
	// these exact bytes, so the existing resolver stamps the row at the INSERT.
	s.RenderedTemplateSHA = rc.RenderedTemplateSHA

	logger.Info("adopt fragment: bound an unidentified fragment to the adopted-fragment component "+
		"(slot name and pages.sections untouched — RFC_046 / bugs_open/357)",
		zap.String("slot_name", s.ComponentName),
		zap.String("component_id", componentID),
		zap.Int("bytes", len(s.HTML)))
	return true
}

// carriedIdentity returns the stored row's component when the carry is armed AND
// that component is one adoption created, and the empty string otherwise — so a
// carried section states the identity of the bytes it is made of, or states
// nothing.
//
// ⚠ NARROWED after council round 1, where THREE seats (editquality, guardian,
// bug_historian) independently made the same point about the first version: it
// carried the stored identity for EVERY interactive section, which is far broader
// than the diagnosed bug and would silently keep a legitimately-typed component at
// its old identity when a plan intended to swap it. The first version defended
// that as "a fix rather than a regression" — by argument, not by measurement,
// which is exactly the move this estate's review exists to catch. Restricting the
// carry to `adopted-fragment` rows delivers the causal fix (adoption must survive
// a rebuild, or the next one re-mints the plan's identity over it) and changes
// carry semantics for nothing else.
//
// It stays a helper rather than an inline condition because the two Layer 2 arms
// must agree: the splice and the re-append are the same decision about the same
// bytes, and the last time they diverged nobody noticed for a day.
func carriedIdentity(armed bool, storedComponentID, storedComponentFunction string) string {
	if !armed || storedComponentID == "" {
		return ""
	}
	if storedComponentFunction != adoptedFragmentFunction {
		return ""
	}
	return storedComponentID
}
