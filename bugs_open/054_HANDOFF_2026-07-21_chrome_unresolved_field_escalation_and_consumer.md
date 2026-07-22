# 054 — FOLLOW-ON: make an unresolvable chrome/render field ESCALATE, not just log

> **⚠️ NUMBER COLLISION (concurrent filing 2026-07-21).** Two different bugs were filed as `054` by
> two sessions at once — this one (`…_chrome_unresolved_field_escalation_and_consumer`) and
> `054_HANDOFF_2026-07-21_unguarded_range_items_in_list_templates_no_empty_state.md` (relojistas
> thread). Per the repo convention (`bugs_closed/README.md`, same as 016/017), the number is NOT
> reassigned — **resolve by slug.** Council submission `7152c7cf` references *this* one (the chrome
> escalation follow-on). They are unrelated.

**Filed:** 2026-07-21 · idea.uk vm site thread · **Status:** IMPLEMENTED (chrome path), committed
`524b03f03`, **inert until an image roll** — see "IMPLEMENTED 2026-07-22" at the foot. Council review
in flight, `SUBMISSION_CORR=f54a1808-51a6-4ddd-8f60-783f9b263e37`.
**Severity:** medium — it is the second half of a fix whose first half (observability) is already in
council review. On its own the gap is "a dead control is now logged loudly but nothing acts on it".
**Class:** structural — completes the FAIL-LOUD contract for the render path; requires a staged
rollout and a work-item consumer, which is why it is not bundled with the observability change.

---

## Why this exists as a separate file

The council reviewed the chrome-renderer fix (submission `7152c7cf`, `bugs_open/018` + `041`) and
returned REVISE **twice** with one recurring high/medium objection: a named log is *observability*,
not *escalation*. Per the FAIL-LOUD, NOT SILENT contract, an unresolvable render field should BLOCK
the build or FILE a work item — a log nobody consumes is "better-documented silence"
(`render_guardian`, `bug_historian`).

The objection is correct. The owner ruled (2026-07-21): **ship the observability version now, do the
escalation as this follow-on.** The split is deliberate, not a dodge — see "Why not just do it in the
original" below. The observability submission is the strict prerequisite: it already produces the
exact signal this work escalates on (edit 3's severity split — `Error` when a blanked placeholder
sits inside `href=""`/`src=""`, `Warn` otherwise).

## What "done" looks like

An active component whose declared field cannot be resolved at render time must reach a state a human
or handler will act on — **not** silently ship an empty `href`/`src` to a live page (which is exactly
how `bugs_open/018` shipped 30 dead controls on idea.uk). Two mechanisms, and this work needs BOTH,
because either alone has a known hole:

1. **A build-time gate** — refuse to deploy a page/site whose chrome renders a dead URL control.
2. **A work-item + consumer** — where a gate is too blunt, file a finding that something actually
   drains.

## The two hard constraints (these are why it is not a one-liner)

### Constraint 1 — a hard gate cannot land cold. It must be staged.
Measured 2026-07-20: **30 active components across the fleet** carry a URL-bound bare placeholder
(`(href|src)="{{.x}}"`, ungated) that can be unresolved. Flipping "dead control → blocker" in one
step would fail the next rebuild of most of the fleet. The repo already learned this exact lesson
with `phantom_internal_links` (LNK-009) and `empty_internal_href` (`bugs_open/023` landmine): stage
it **warning → work item → drain the backlog → flip to blocker**, never flip cold.

Census to re-run before flipping (the backlog that must reach zero first):
```sql
SELECT count(*) FILTER (WHERE is_active) AS active_ungated_url_placeholders
FROM content_components
WHERE html_template ~ '(href|src)="\{\{\s*\.[A-Za-z_][A-Za-z0-9_]*\s*\}\}'
  AND html_template NOT LIKE '%{{if%';
-- 2026-07-20: 30 active. This must be ~0 before any blocker flips.
```

### Constraint 2 — filing a work item is useless until a consumer exists.
`bugs_open/023` measured the trap precisely: **34 correctly-filed CTA findings**
(`unresolved_cta`, `cta_names_unknown_destination`) sit unread at `status='needs_human_review'`,
which `TriageDetectedItemsAction` never promotes, which no `handler_agent` consumes, and which
`load_work_item_actions.go:804` excludes from re-open queries. A grep of `platform/` for those item
types returns **only their emission sites — zero consumers.** Adding a 35th unread type makes the
invisible pile bigger. **The consumer is the load-bearing deliverable here, not the detection.**

This work therefore overlaps `bugs_open/023` fix #3 ("build the handler for CTA findings") and
should probably be done WITH it, or reuse whatever it builds — the two are the same delivery gap seen
from the chrome side and the page side. Grep `023` and coordinate before building a parallel handler.

## Suggested shape (not prescriptive)

1. **Consumer first.** A `handler_agent` (or an extension of the page/chrome rerender path) that
   drains chrome/CTA "unresolvable field" findings: where a real destination exists, repair; where
   none exists, **drop the control** (gate the anchor) — never point it at `/contact.html`, the
   heuristic that created the phantom-CTA bug LNK-007.
2. **Emit the finding from the render path** using edit 3's already-computed `inURLAttr` list (the
   observability submission hands it to you — `RenderTemplateReportingMissing` returns it). One
   finding per dead URL field, deduped by component + field.
3. **Stage the gate.** Start as a non-blocking discovery check over `site_components` +
   `page_components` rendered_html (the census above), let the consumer drain it, and only then flip
   to a deploy blocker. Wire it as a discovery check so the immune system's existing triage carries
   it, rather than a new bespoke sweep.

## Why not just do it in the original submission

- The observability change is safe and immediate and improves today's situation (count-only `Warn` →
  field-named `Error` on dead controls). Holding it hostage to the bigger piece helps no live site.
- A gate bundled with it would have failed 30 components' next rebuild — the observability change is
  precisely what lets you SEE those 30 before you gate them.
- The consumer is real engineering (it is `023`'s open fix #3), needs its own review, and must not be
  a stub that files into the same unread pile.

## How to verify (when built)

- The 30-active-ungated census above trends to ~0 as the consumer drains it, and **stays** at ~0
  after a rebuild (a content-level fix regresses; a template/schema/consumer fix does not).
- A newly-onboarded chrome component with an unresolvable required URL field does NOT reach
  `build_status='deployed'` with an empty `href`/`src` — it is either repaired, gated, or blocked.
- The findings it emits reach a **terminal** state via the consumer, not by hand and not by rotting
  at `needs_human_review`.

## Related
- `bugs_open/018` — the chrome renderer ignores `input_schema` (the observability fix, in council review as `7152c7cf`).
- `bugs_open/041` — chrome component JS never published (fixed in the same submission).
- `bugs_open/023` — CTA label/URL pairing unchecked; **its fix #3 is the same consumer this needs — coordinate.**
- `docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES §X.6–X.7` — the two council rounds and the measurements.
- Council submission `7152c7cf` round 3 rationale — the owner ruling that created this file.

---

## IMPLEMENTED 2026-07-22 (bugfix o54 session) — chrome path done, committed, inert until roll

**Two owner rulings unblocked this** (AskUserQuestion, 2026-07-22):
1. **Mechanism = render-path drop-the-control.** Make the site-chrome renderer itself refuse to
   ship a dead URL control, rather than gate templates one-by-one or hard-block a deploy.
2. **`bugs_open/033` IS a queue** — a human works these escalations. So the finding goes to a
   **draining** pathway, not the `needs_human_review` void that Constraint 2 warned against. (This
   is a decision `033`'s own thread needs — recorded there.)

**What the acute state actually is (grounded live 2026-07-22, and it changes the calculus):**
- **0** live `site_components` render an empty `href`/`src`; all **10** live-placed chrome
  components are gated. The 7 ungated chrome components in Constraint 1's census
  (`*_pre_037`, `site-head`, `header-docs`) have **0 placements** — dormant library stock.
- So the acute idea.uk fire (30 dead controls) is **out**, and a chrome-scoped mechanism has
  **~0 live blast radius**. Constraint 1's "can't gate cold" fear is largely moot for this path;
  the change is **preventive** (stops a recurrence at source) and inert on today's fleet.

**Built (commit `524b03f03`), both gated on the already-computed `deadURLFields` set so a clean
render is never touched and its byte-identical output is preserved:**
- **Half 1 — drop the control.** `DropDeadURLControls` (new file
  `platform/orchestration/actions/drop_dead_url_controls.go`) removes any anchor whose `href`
  rendered empty and blanks any empty `src` from the rendered chrome before store (LNK-005
  correct-or-absent). Wired into `renderAndStoreSiteComponent` after
  `RenderTemplateReportingMissing`. 16 unit cases (`drop_dead_url_controls_test.go`), positive +
  negative boundaries (`<area>`/`<abbr>` excluded, `href="#"` and non-empty attrs untouched).
- **Half 2 — escalate to the worked queue.** `emitChromeDeadControlItem` files ONE
  `chrome_dead_control` work item per (site, slot) at `status='detected'` +
  `handler_agent='nav-link-fixer'` (`build` pipeline) — the `phantom_internal_links`
  site_component convention. Triage promotes every `detected` item; the dispatch loop claims it;
  a re-render fixes the common data-lag case, a genuinely-unresolvable field exhausts
  `max_attempts` → `failed` (bounded, no storm) and surfaces to the human queue. Deduped by
  `item_key` against `idx_swi_dedup`.

**Why NOT the `needs_human_review` void:** `check_dead_controls` (page_components) files at
`needs_human_review` with no handler *by design* ("picking a fixer automatically would guess"), and
the machinery map confirmed triage cannot even reach those rows — that is the 124-item unread pile.
The owner's 033=queue ruling + the phantom-links `detected`+handler convention is the deliberate
alternative: a chrome dead control routes to a re-render first (which fixes the idea.uk-shape
data-lag case) and only a persistently-dead one reaches a human.

**Concurrency note (not a wrong call, but logged):** `component_library.go` was under **active edit
by another session** (a template-scanning feature) while I worked — the helper was moved to its own
file `drop_dead_url_controls.go` so the commit's pathspec excludes the contended file and takes no
same-file passenger.

**How to verify once it rolls:**
- Induce the failing branch (per the "verify the failing branch" rule — a green happy path proves
  deployment, not detection). Place a `*_pre_037` chrome component on a test site with an
  unresolvable nav/CTA field and re-render; assert the rendered chrome has **no** `href=""`/`src=""`
  and a `chrome_dead_control` row appears at `status='detected'`, `handler_agent='nav-link-fixer'`.
- Confirm it drains: the item is claimed and either re-render resolves the field (data present) or
  it exhausts attempts to `failed`.
- Pod-grep the roll for the CREATED symbols (`DropDeadURLControls`, `emitChromeDeadControlItem`,
  `chrome_dead_control`) — not a string the change merely uses.

**Still genuinely out of this file's scope (unchanged):** the *general* consumer that drains the
existing 124-item unread pile (`unresolved_cta`=68, `cta_names_unknown_destination`=47,
`dead_control`=6, `empty_internal_href`=3) is `bugs_open/033`'s work now that the owner has ruled it
a queue — this change only wires the **chrome** finding into a draining path; it does not build the
human-facing drain for the pre-existing pile.
