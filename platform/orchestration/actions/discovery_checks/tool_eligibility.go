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
// NOT APPLIED TO tool_health, on purpose. That check raises an improve_tool
// work item for ANY issue including cosmetic warnings ("no @media breakpoint",
// "hardcoded colours"), so widening it would drop ~71 items into the build
// queue in one pass, most of them about styling on pages whose real defect is
// missing markup. The orphan_element_refs check covers the serious half of that
// population today with 10 precise findings. Revisit tool_health once the
// ported tools have PLANs and the noise can be judged per tool.

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
