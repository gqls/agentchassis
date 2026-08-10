// FILE: platform/orchestration/actions/write_site_plan_imagery_scope.go
//
// Imagery scope_ref resolution for WriteSitePlanAction (bugs_open/214).
// Split out of write_site_plan_action.go so the seam is reviewable on its own.
package actions

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ============================================================================
// Imagery scope_ref canonicalisation (bugs_open/214)
// ============================================================================
//
// WHY THIS EXISTS. WriteSitePlanAction canonicalises page identity for two of
// the three tables it writes and not for the third, inside one function about
// sixty lines apart:
//
//	site_plan_pages.name        <- datahelpers.CanonicalisePage  (canonical)
//	site_plan_sections.page_name <- the same canonical r.Name    (canonical)
//	site_plan_imagery.scope_ref  <- the planner LLM's raw map key (verbatim)
//
// CanonicalisePage deliberately collapses spellings onto one identity. The
// section-index family turns "about" into "about-index", "contact" into
// "contact-index", "news" into "news-index". So the planner emits
// imagery.pages["about"], the plan writes its page as "about-index", and the
// imagery row names a page ITS OWN PLAN DOES NOT CONTAIN. Nothing rejected it:
// buildImageryRow never inspects scopeRef.
//
// The consequence is not a broken image — it is a MISSING one. Every consumer
// selects the plan row by scope_ref and then joins the asset on
// assets.asset_key = spi.key, so an unmatched scope_ref does not produce a 404;
// it produces an asset that was planned, generated, deployed, paid for and
// referenced by nothing. That is bugs_open/114's symptom reached through this
// cause. Measured 2026-08-10: 10 such rows on current plans, 8 with an active
// asset already generated.
//
// THE GUARANTEE. After this pass, a page/section scope_ref written by this
// action either names a page the plan itself contains, or is preserved
// byte-for-byte AND leaves a durable agent_error_log row saying it does not.
// Nothing is ever silently dropped — an unresolvable ref keeps exactly today's
// behaviour, so no row can regress.
//
// WHAT THIS DELIBERATELY DOES NOT DO: rewrite the ordinal. No consumer parses
// it (the section join is `scope_ref LIKE <page> || ':%'` keyed by asset key,
// and flag_page_image_rebuild splits on the first colon and discards the rest),
// so a rewrite buys no behaviour while risking a 23505 on the unique index and
// breaking lock carry-forward. The correct ordinal is also unknowable here:
// ordinal shift happens in ValidateSitePlanAction, a different action, and the
// pre-drop array is gone by the time we run. Out-of-range ordinals are recorded
// instead, which turns them from silent into observable.

// imageryRefMiss is one anomalous scope_ref, accumulated during the pure pass
// and recorded durably after the transaction commits.
//
// Accumulate-then-record (rather than logging inline) is the recordFactCarryMisses
// pattern from v3_site_actions.go: it keeps the resolution logic pure and
// unit-testable, and it means a rolled-back plan write leaves no error rows
// claiming rows that do not exist.
type imageryRefMiss struct {
	Scope      string // "page" | "section"
	RawRef     string // as minted by the planner
	WrittenRef string // what the row actually carries
	Key        string
	Kind       string
	Reason     string // "page_not_in_plan" | "ordinal_out_of_range" | "ordinal_malformed"
}

// normalisePageKey applies the same cleanup ValidateRoles applied when it built
// planPageRow.RawName (lowercase, trim, spaces to hyphens, path and .html strip),
// so a planner's imagery map key can be matched against it.
//
// WHY IT ROUTES THROUGH CanonicalisePage RATHER THAN COPYING THE RULE.
// datahelpers.normaliseSlug is unexported, and a hand-copy in this package would
// be a second implementation of page-name normalisation — which is the exact
// drift class bugs_open/214 is an instance of, so reproducing it here to fix it
// would be absurd. CanonicalisePage's content/default branch returns its slug
// with precisely that normalisation applied and no role-dependent prefix or
// suffix work, so it is the exported route to the same answer.
//
// That is a real coupling to another function's internal branch, so it is PINNED
// BY TEST rather than left to trust: see
// TestNormalisePageKey_MatchesTheNormalisationValidateRolesApplies. If
// CanonicalisePage's default branch ever changes, that test fails here rather
// than this silently resolving page names differently from the plan writer.
//
// One deliberate extra: "home" and "index" both collapse to "index", because
// that branch runs first. That is correct for us — the plan writes the homepage
// as "index", so a planner keying imagery to "home" resolves to the real page
// instead of missing it.
func normalisePageKey(s string) string {
	name, _, _ := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
		Role: "content",
		Slug: s,
	})
	return name
}

// splitScopeRef splits a section scope_ref into its page part and the
// remainder INCLUDING the leading colon, on the FIRST colon.
//
// First-colon is not a style choice — it is the split every consumer already
// performs (flag_page_image_rebuild_action.go, and the `LIKE <page> || ':%'`
// join in plan_sections). Returning the remainder with its colon attached
// means a reassembled ref keeps its colon by construction, so
// chk_scope_ref_consistency ("section => scope_ref LIKE '%:%'") holds without
// a special case, and any ordinal text we do not understand survives untouched.
func splitScopeRef(ref string) (page, rest string) {
	if i := strings.Index(ref, ":"); i >= 0 {
		return ref[:i], ref[i:]
	}
	return ref, ""
}

// buildCanonicalPageNameMap maps every spelling the plan knows for a page onto
// the canonical name the plan actually contains.
//
// Built from planRows, which are ALREADY canonicalised and deduped by the time
// imagery is flattened — so there is no second canonicalisation here and no
// second implementation to drift. There could not be one anyway:
// CanonicalisePage requires a Role to decide between the tool-/guide-/game-
// prefix, the -index suffix and the homepage collapse, and an imagery ref
// carries no role. Re-deriving a canonical name from the bare ref is therefore
// impossible, which is exactly why this map is built from the plan's own rows.
//
// Two passes, in this order:
//
//  1. identity — every canonical name maps to itself, so a ref already spelt
//     canonically resolves unconditionally and takes the no-op branch. This is
//     what makes "rows that work today are untouched" true by construction
//     rather than by inspection.
//  2. aliases — the raw LLM spelling and the slug, both normalised the same way
//     ValidateRoles normalised RawName. An alias is only added if it is not
//     already taken; an alias that two different pages would claim is REMOVED
//     rather than resolved arbitrarily, because a confident wrong rewrite is
//     worse than no rewrite (the ref then falls through to the miss path and
//     keeps today's behaviour plus a durable record).
func buildCanonicalPageNameMap(pages []planPageRow, logger *zap.Logger) map[string]string {
	m := make(map[string]string, len(pages)*3)

	// Pass 1: identity. Must run first and completely — an alias may never
	// displace a real canonical name.
	for _, p := range pages {
		if p.Name != "" {
			m[p.Name] = p.Name
		}
	}

	// Pass 2: aliases. ambiguous tracks keys claimed by two different targets.
	ambiguous := make(map[string]bool)
	addAlias := func(alias, target string) {
		alias = normalisePageKey(alias)
		if alias == "" || target == "" || ambiguous[alias] {
			return
		}
		if existing, ok := m[alias]; ok {
			if existing == target {
				return // same answer twice; harmless
			}
			if _, isCanonical := m[existing]; isCanonical && existing == alias {
				return // pass-1 identity always wins over an alias
			}
			// Two different pages want this alias. Refuse to guess.
			delete(m, alias)
			ambiguous[alias] = true
			logger.Warn("WriteSitePlanAction: imagery page alias is ambiguous, will not be resolved",
				zap.String("alias", alias),
				zap.String("target_a", existing),
				zap.String("target_b", target))
			return
		}
		m[alias] = target
	}
	for _, p := range pages {
		addAlias(p.RawName, p.Name)
		addAlias(p.Slug, p.Name)
	}

	return m
}

// sectionCountByPage returns canonical page name -> number of sections that
// survived planning, for the ordinal range check.
func sectionCountByPage(pages []planPageRow) map[string]int {
	m := make(map[string]int, len(pages))
	for _, p := range pages {
		if p.Name != "" {
			m[p.Name] = len(p.Sections)
		}
	}
	return m
}

// canonicaliseImageryScopeRefs rewrites page/section scope_refs onto the page
// names the plan actually contains, and records the ones it cannot.
//
// Returns the rows in their original order, the misses, and the number of refs
// actually rewritten (which is what a caller should log — a count of zero on a
// plan known to use aliases is the tell that this pass is not wired in).
func canonicaliseImageryScopeRefs(
	rows []imageryRow,
	canonByRaw map[string]string,
	sectionCount map[string]int,
	logger *zap.Logger,
) ([]imageryRow, []imageryRefMiss, int) {
	var misses []imageryRefMiss
	rewritten := 0

	for i := range rows {
		r := &rows[i]
		if r.ScopeRef == nil || (r.Scope != "page" && r.Scope != "section") {
			continue // site scope carries no ref and needs no resolution
		}
		raw := *r.ScopeRef
		pagePart, rest := splitScopeRef(raw)

		canonical, ok := canonByRaw[normalisePageKey(pagePart)]
		if !ok {
			// The plan contains no page by this name under any spelling. Keep
			// the ref verbatim — identical to pre-fix behaviour — and record it.
			misses = append(misses, imageryRefMiss{
				Scope: r.Scope, RawRef: raw, WrittenRef: raw,
				Key: r.Key, Kind: r.Kind, Reason: "page_not_in_plan",
			})
			continue
		}

		if canonical != pagePart {
			// chk_scope_ref_consistency forbids a colon in a PAGE-scope ref.
			// Canonical names do not contain colons today; the guard is here
			// because the constraint makes it mandatory rather than stylistic
			// — violating it would abort the entire plan write.
			if r.Scope == "page" && strings.Contains(canonical, ":") {
				logger.Warn("WriteSitePlanAction: canonical page name contains a colon, leaving imagery scope_ref unrewritten",
					zap.String("scope_ref", raw),
					zap.String("canonical", canonical))
			} else {
				newRef := canonical + rest
				r.ScopeRef = &newRef
				rewritten++
				logger.Info("WriteSitePlanAction: imagery scope_ref canonicalised",
					zap.String("scope", r.Scope),
					zap.String("from", raw),
					zap.String("to", newRef),
					zap.String("key", r.Key))
			}
		}

		// Ordinal check — log only, never rewritten. See the header comment.
		if r.Scope == "section" && rest != "" {
			ordinalText := strings.TrimPrefix(rest, ":")
			ordinal, err := strconv.Atoi(ordinalText)
			switch {
			case err != nil:
				misses = append(misses, imageryRefMiss{
					Scope: r.Scope, RawRef: raw, WrittenRef: *r.ScopeRef,
					Key: r.Key, Kind: r.Kind, Reason: "ordinal_malformed",
				})
			case ordinal >= sectionCount[canonical]:
				misses = append(misses, imageryRefMiss{
					Scope: r.Scope, RawRef: raw, WrittenRef: *r.ScopeRef,
					Key: r.Key, Kind: r.Kind, Reason: "ordinal_out_of_range",
				})
			}
		}
	}

	return rows, misses, rewritten
}

// dedupeImageryRows collapses rows that now share (scope, scope_ref, key),
// returning survivors in first-appearance order and the number of merges.
//
// WHY THIS EXISTS. idx_site_plan_imagery_unique is UNIQUE on
// (plan_id, scope, COALESCE(scope_ref,''), key). Canonicalisation is a
// COLLAPSING map — "about" and "about-index" both become "about-index" — so a
// planner that keyed one page under two spellings with a shared imagery key now
// produces two rows carrying one identity, and the insert aborts THE WHOLE PLAN
// WRITE with SQLSTATE 23505. That is bugs_open/215 verbatim on the neighbouring
// table, and dedupePlanPageRows exists for exactly the same reason.
//
// Zero collisions exist on any current plan today (measured fleet-wide,
// 2026-08-10) — so this is not a live repair. It is here because the
// canonicalisation above is what creates the possibility, and shipping the
// collapse without the guard would be introducing the 215 failure mode to a
// second table.
//
// Merge rule: an entry whose ref was NOT rewritten beats one that was, because
// the planner keying a page by the plan's own real name is the better-aimed of
// the two. Otherwise first wins, so the rule is total and order-stable.
func dedupeImageryRows(rows []imageryRow, logger *zap.Logger) ([]imageryRow, int) {
	out := make([]imageryRow, 0, len(rows))
	at := make(map[string]int, len(rows))
	merges := 0

	for _, r := range rows {
		ref := ""
		if r.ScopeRef != nil {
			ref = *r.ScopeRef
		}
		k := r.Scope + "\x00" + ref + "\x00" + r.Key

		i, seen := at[k]
		if !seen {
			at[k] = len(out)
			out = append(out, r)
			continue
		}

		merges++
		incumbent := out[i]
		incumbentAimed := incumbent.RawScopeRef == ref
		challengerAimed := r.RawScopeRef == ref
		if challengerAimed && !incumbentAimed {
			out[i] = r
		}
		logger.Warn("WriteSitePlanAction: two imagery entries collapsed onto one identity after canonicalisation",
			zap.String("scope", r.Scope),
			zap.String("scope_ref", ref),
			zap.String("key", r.Key),
			zap.String("kept_raw_ref", out[i].RawScopeRef),
			zap.String("dropped_raw_ref", incumbent.RawScopeRef+"/"+r.RawScopeRef),
		)
	}

	return out, merges
}

// recordImageryRefMisses writes one agent_error_log row per anomalous ref.
//
// Called AFTER the transaction commits. Two reasons: agenterrors.Write takes a
// *sql.DB and cannot join the tx anyway, and recording post-commit means a
// rolled-back plan write leaves no rows describing imagery that was never
// persisted. Best-effort throughout — a failed log write must never change the
// action's disposition.
//
// The durable record is the whole point of the miss path. Pod logs rotate
// (chassis keeps under a second) and the orchestration's collected_data is
// pruned at ~24h, so without this row a scope_ref that matches nothing produces
// EXACTLY the output a working one produces, for ever.
func recordImageryRefMisses(
	ctx context.Context,
	params ActionParams,
	siteID, planID uuid.UUID,
	misses []imageryRefMiss,
	logger *zap.Logger,
) {
	for _, m := range misses {
		code := "IMAGERY_SCOPE_REF_ORDINAL_ANOMALY"
		message := fmt.Sprintf(
			"imagery %s scope_ref %q: ordinal is %s for page %q — the reference still reaches the page, but its section placement is meaningless",
			m.Scope, m.WrittenRef, strings.TrimPrefix(m.Reason, "ordinal_"), splitPageOf(m.WrittenRef))
		remedy := "no consumer parses the ordinal today, so this is observable rather than damaging; " +
			"the planner keyed a section position that the surviving section list does not have. " +
			"Fix belongs in the planner prompt or in RFC_016's move of imagery inside the section entry (bugs_open/214 candidate 2)"

		if m.Reason == "page_not_in_plan" {
			code = "IMAGERY_SCOPE_REF_UNRESOLVED"
			message = fmt.Sprintf(
				"imagery %s scope_ref %q names a page this plan does not contain under any spelling — the row was written verbatim and no build will reference it",
				m.Scope, m.WrittenRef)
			remedy = "compare scope_ref against site_plan_pages.name for this plan_id; the asset (if generated) is " +
				"reachable by asset_key but no page will pick it up. Either the planner named a page it did not plan, " +
				"or the page was dropped after the imagery block was written (bugs_open/214)"
		}

		LogActionError(ctx, params, siteID.String(), "", "write_site_plan",
			code, "warning", message,
			map[string]interface{}{
				"plan_id":           planID.String(),
				"scope":             m.Scope,
				"scope_ref_raw":     m.RawRef,
				"scope_ref_written": m.WrittenRef,
				"key":               m.Key,
				"kind":              m.Kind,
				"reason":            m.Reason,
				"remedy":            remedy,
			}, logger)
	}
}

// countMissReason counts misses of one reason, for the completion log.
func countMissReason(misses []imageryRefMiss, reason string) int {
	n := 0
	for _, m := range misses {
		if m.Reason == reason {
			n++
		}
	}
	return n
}

// splitPageOf returns just the page part of a scope_ref, for message text.
func splitPageOf(ref string) string {
	p, _ := splitScopeRef(ref)
	return p
}

// sortedAnyKeys returns a map's keys in a stable order. (The package already
// has sortedKeys for map[string]string, in execution_context_params.go.)
//
// flattenImageryBlock walks Go maps, whose iteration order is randomised per
// run. That is invisible while every row is independent, but the dedupe guard
// above resolves a collision by first-appearance — so without a stable order
// the survivor of a collision would differ between two runs over identical
// input, and no test could pin it.
func sortedAnyKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
