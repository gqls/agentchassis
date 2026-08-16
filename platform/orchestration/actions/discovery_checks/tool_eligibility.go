// FILE: platform/orchestration/actions/discovery_checks/tool_eligibility.go
//
// Which components the tool-acceptance ladder is allowed to look at, and what
// it calls them.
//
// THE DEFECT THIS FIXES. Every tier of the ladder opened with
// `cc.component_level = 'tool'`. webdesign.co.uk's 63 tools are ported HTML
// blobs sitting in ONE shared content_components row at component_level
// 'section' (function 'ported-page', reused by 97 page_components across 97
// pages). So the eligibility query has never returned one of them, and 63 tools
// — twenty of them visibly broken, ten with scripts addressing elements that do
// not exist — shipped without a single alarm from a ladder built precisely to
// catch that. Recorded as bugs_open/084 candidate 3.
//
// Widening the level test alone would NOT have been enough, and this is the
// part worth reading twice: the ladder keys its subject on `cc.function`, which
// for all 63 is the literal string "ported-page". They would all have collided
// onto one PLAN. A ported tool's identity is its PAGE, not its component.
//
// THE RULE. A component is eligible when either:
//
//	(a) it is a real tool component (component_level = 'tool'), keyed by
//	    cc.function, exactly as before — no existing subject changes; or
//	(b) it is the SOLE component on a page_type='tool' page that has no
//	    tool-level component, keyed by the page name minus its 'tool-' prefix.
//
// The sole-component clause in (b) is what keeps this honest. A generated tool
// page carries four components (hero-tool, tool-guide-intro, the widget,
// tool-cta) and the widget is already covered by (a); admitting its siblings
// would key three content sections as if each were a tool. A ported page is one
// blob and there is nothing to disambiguate. Four fleet pages are multi-section
// with no tool component (idea.uk 3, leopardessconsulting 1) — genuinely
// ambiguous, and deliberately still excluded rather than guessed at.
//
// MEASURED BEFORE SHIPPING (2026-07-29, live DB):
//
//	71 components become newly eligible — webdesign.co.uk 63, gamesdesign.co.uk
//	6, vonc.com 2 — across 71 DISTINCT subject keys, with 0 keys colliding with
//	an existing tool function, and exactly 1 of the 71 already carrying a PLAN.
//
// That last number is the blast radius. Both callers are criteria-gated: no
// PLAN means Tier 4 emits nothing at all and Tier 2 writes a needs_criteria
// doc_note (30-day cooldown), never a work item. So this widening produces ONE
// acceptance run today, for the one ported tool that has been repaired and
// documented — and then grows exactly as fast as PLANs get written, which is
// the pace the per-tool repair loop sets anyway.
//
// APPLIED TO tool_health SINCE 2026-08-15 (bugs_open/281). Until then this
// paragraph read "NOT applied, on purpose": tool_health raises an improve_tool
// item for ANY issue including cosmetic warnings, so widening it would have
// dropped ~71 items into the build queue in one pass, mostly styling noise on
// pages whose real defect was missing markup, with no PLANs to judge them by.
// The owner's visual gate then found the Mind Map Studio (webdesign.co.uk)
// with illegible pale-on-pale controls and junk seeded content — exactly the
// class tool_health's hardcoded_colors check exists for — and no item had ever
// been filed, because the exclusion hid it. The noise objection is answered
// STRUCTURALLY in check_tool_health.go rather than dismissed: a ported
// instance's findings file as `ported_tool_fix` with no handler (a human
// queue, not tool-improver's), identity is per page instance so 63 instances
// cannot collapse onto one dedup key, and Tier-2 audit queueing is capped per
// run. A ported tool still gets no automated FIXER until it has a PLAN and a
// per-instance repair path exists — that half of the old objection stands.

package discovery_checks

// toolSubjectKeyExpr yields the ladder's subject_key for a row selected under
// toolEligibilityWhere. Requires `cc` and `p` to be in scope.
//
// The regexp_replace is anchored: only a LEADING 'tool-' is stripped, so a page
// named 'tool-tool-belt' becomes 'tool-belt' and not 'belt'.
const toolSubjectKeyExpr = `CASE
			WHEN cc.component_level = 'tool' THEN cc.function
			ELSE regexp_replace(p.name, '^tool-', '')
		END`

// ToolSubjectKeyExpr is toolSubjectKeyExpr exported for ONE outside reader:
// refresh_evidence_fact_drift.go's fact-drift fan-out (package actions), which
// must resolve a PLAN's subject_key to the pages carrying that tool.
//
// It is exported rather than copied because two council seats (reuse_agent,
// tooling_provenance, 2026-08-16) objected to the fan-out re-deriving an
// equivalent-looking predicate inline: a second spelling of the subject-key rule
// can drift from what the acceptance tiers actually resolve, and then the two
// disagree about which tool a finding belongs to.
//
// NOTE the deliberate asymmetry: the fan-out reuses this EXPRESSION but NOT
// toolEligibilityWhere below. Measured 2026-08-16, that predicate's
// sole-component clause admits neither SDLT tool, so keying the fan-out on it
// would make it permanently silent on the tools it exists for. Encoding a fact
// and being acceptance-eligible are different questions; deriving a subject key
// is the same question.
const ToolSubjectKeyExpr = toolSubjectKeyExpr

// toolEligibilityWhere is the shared predicate. It is appended to a query that
// has already constrained p.site_id, and it carries its own is_active/status
// tests so the two callers cannot drift apart on them.
const toolEligibilityWhere = `
		  AND cc.is_active = true
		  AND p.status = 'active'
		  AND (
		        cc.component_level = 'tool'
		     OR (
		          p.page_type = 'tool'
		          AND NOT EXISTS (
		                SELECT 1 FROM page_components pc_t
		                JOIN content_components cc_t ON cc_t.id = pc_t.component_id
		                WHERE pc_t.page_id = p.id
		                  AND cc_t.component_level = 'tool'
		                  AND cc_t.is_active = true
		              )
		          AND (
		                SELECT count(*) FROM page_components pc_n
		                JOIN content_components cc_n ON cc_n.id = pc_n.component_id
		                WHERE pc_n.page_id = p.id AND cc_n.is_active = true
		              ) = 1
		        )
		      )`
