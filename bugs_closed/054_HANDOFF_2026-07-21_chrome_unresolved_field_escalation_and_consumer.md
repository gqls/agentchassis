# 054 — FOLLOW-ON: make an unresolvable chrome/render field ESCALATE, not just log

> **⚠️ NUMBER COLLISION (concurrent filing 2026-07-21).** Two different bugs were filed as `054` by
> two sessions at once — this one (`…_chrome_unresolved_field_escalation_and_consumer`) and
> `054_HANDOFF_2026-07-21_unguarded_range_items_in_list_templates_no_empty_state.md` (relojistas
> thread). Per the repo convention (`bugs_closed/README.md`, same as 016/017), the number is NOT
> reassigned — **resolve by slug.** Council submission `7152c7cf` references *this* one (the chrome
> escalation follow-on). They are unrelated.

**Filed:** 2026-07-21 · idea.uk vm site thread · **Status:** ✅ **CLOSED — FIXED, LIVE & BEHAVIOURALLY
VERIFIED** on **v1.0.1155** (2026-07-24). Council-APPROVED. Commits `524b03f03` → `0132f859b` →
`2afa6531a`; council trail `SUBMISSION_CORR=3951e2be-cf0e-4f73-901f-27bd84b3342d`. Its own scope
(chrome dead-control drop + escalate) is complete; the general unread-pile drain stays `033`'s and the
page-render paths are a noted follow-on — both explicitly out of scope (see foot). See "VERIFIED
2026-07-24" below.
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

## VERIFIED 2026-07-24 — induced the failing branch on the live binary, both halves proven

Deploy-grep alone proves deployment, not detection (this fix's own genesis, 018/041). So I built the
missing trigger and induced a real dead chrome control on a throwaway site:

- **The trigger gap:** no existing agent renders chrome without also deploying — `rerender-site` and
  `nav-updater` both chain `render_site_components → render_js_snippets → deploy_js_snippets (git_commit)
  → page deploy` (~26–37 commits to `gqls/sites` + B2 + CF per run). So I added a minimal
  **`render-site-chrome`** agent (`render_site_components → complete`, `site_id` from input_data,
  config-only/live, manual-dispatch) and a scratch-site verifier. Both in
  `054_chrome_verify/` (`01_render_site_chrome_agent.sql`, `02_verify_054_induced_fault.sh`).
- **Induced fault:** scratch site `scratch-054-verify.invalid` with the **ungated**
  `header-with-search_pre_037` on its header slot (bare `<a href="{{.nav_link_N_url}}">`,
  `src="{{.logo_url}}"`, `<a href="{{.cta_url}}">` — none resolvable). Dispatched `render-site-chrome`.
- **Result (both PASS, on v1.0.1155):**
  - **Half 1 (drop):** rendered header — `has_empty_url_attr=false`, `has_unrendered_placeholder=false`,
    header markup intact, some anchors still present (surgical, not a wipe), length 10220→9250.
  - **Half 2 (escalate):** exactly one `chrome_dead_control` row at `status=needs_human_review`,
    **no handler_agent**, `source=render-site-components`, `dead_url_fields=["cta_url","logo_url",
    "nav_link_1_url".."nav_link_4_url"]` — the 6 unresolvable URL fields, named precisely.
  - Scratch site cleaned up; the reusable trigger + verifier are committed for the next time.
- **Trap fixed along the way:** the first agent used `ensure_site_record` as start_step — it FAILED on a
  spec-less scratch site. `render_site_components` reads `site_id` from input_data and loads its own site
  data, so the step was unnecessary; dropped it. Also `psql -tA` on `INSERT … RETURNING` prints the
  `INSERT 0 1` command tag alongside the value — switched to INSERT-then-SELECT.

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

**Built across three commits (`524b03f03` → council-REVISE revision `0132f859b` → APPROVED guard
`2afa6531a`), both mechanisms gated on the already-computed `deadURLFields` set so a clean render is
never touched and its byte-identical output is preserved:**
- **Half 1 — drop the control.** `DropDeadURLControls` (new file
  `platform/orchestration/actions/drop_dead_url_controls.go`) removes any anchor whose `href`
  rendered empty, drops any `<img>` whose `src` rendered empty (whole element, not a bare `<img>`),
  and blanks any other empty `src`, from the rendered chrome before store (LNK-005
  correct-or-absent). Wired into `renderAndStoreSiteComponent` after `RenderTemplateReportingMissing`,
  and **skipped for `data-runtime-fill` shells** (their empty hrefs are hydrated client-side —
  exempt exactly as `check_dead_controls`). 17 unit cases (`drop_dead_url_controls_test.go`),
  positive + negative boundaries (`<area>`/`<abbr>` excluded, `href="#"` and non-empty attrs
  untouched).
- **Half 2 — escalate to the worked queue.** `emitChromeDeadControlItem` files ONE
  `chrome_dead_control` work item per (site, slot) at **`status='needs_human_review'` with NO
  handler**, MIRRORING the sibling `check_dead_controls`. Persisted via the shared `insertWorkItem`
  helper (correct `idx_swi_dedup`-matched `ON CONFLICT` on the shared `workItemTerminalStatuses`
  constant + the two-strike anti-churn label). It is visible in the dashboard queue the owner has
  ruled a queue (033).

> **CORRECTED at council REVISE (2026-07-22):** the *first* build routed Half 2 to
> `status='detected'` + `handler_agent='nav-link-fixer'` (the phantom-links convention). The council
> (guardian/bug_historian) caught that `nav-link-fixer`'s workflow re-renders then marks the item
> **complete without verifying the field resolved** — so a genuinely-unresolvable control would be
> silently re-dropped and marked done forever, never reaching a human. Owner confirmed the reroute to
> `needs_human_review` + no handler. The data-lag case self-heals on the next normal re-render
> anyway, so nothing is lost by not auto-handling it.

**Council: APPROVED** (trail `3951e2be`, round 2; 4 advisory objections, none high). The one acted-on
advisory was `render_guardian`'s `data-runtime-fill` exemption (the `2afa6531a` guard); the rest were
the reviewers' partial-schema inability to see `site_components` (it exists, 36 rows) and a
process preference for attached code_check bundles over manual grep. `Council-Reviewed: 3951e2be`
trailer on `2afa6531a`.

**DEPENDENCY — do not deploy this on a binary older than `78482c86b` / v1.0.1149.** The drop is gated
on `deadURLFields` (`missingBareFields`' `inURLAttr`). Until `78482c86b` that detector was
**control-flow-blind** — it flagged bare `{{.Name}}` inside `{{if}}`/`{{range}}`/`{{with}}` as
missing (false positives on ~30 components). Dropping on THAT would have removed live, correctly-gated
controls. `78482c86b` made it a scope-aware `text/template/parse` walk (root-scope, ungated only) and
is **LIVE in v1.0.1149**; it is an ancestor of these commits, so the next build from HEAD carries
both. (That fix is also the concurrent `component_library.go` edit noted below.)

**Concurrency note (not a wrong call, but logged):** `component_library.go` was under **active edit by
another session** — which turned out to be exactly `78482c86b` above. The helper was moved to its own
file `drop_dead_url_controls.go` so the commit's pathspec excludes the contended file and takes no
same-file passenger.

**DEPLOY-VERIFIED LIVE 2026-07-23 (v1.0.1150), against the running pod, not git/tag:**
```
POD=agent-chassis-99455fb79-7cbrv  image v1.0.1150
strings /app/agent-chassis | grep -c: DropDeadURLControls=2  emitChromeDeadControlItem=2
                                       chrome_dead_control=5  deadAnchorRe=1  (positive control RenderTemplateReportingMissing=2)
```
Also confirmed live: **0** `chrome_dead_control` rows filed since the roll and **0** "dropped dead URL
control" log lines — i.e. nothing on the live fleet trips it (blast radius ~0, as expected; this is
preventive). That is deployment proof, NOT detection proof.

**REMAINING closure step — INDUCE THE FAILING BRANCH (a green happy path proves deployment, not
detection; this is the rule this fix's own genesis, 018/041, was about):**
- Place a `*_pre_037` chrome component (e.g. `header-with-search_pre_037`, id
  `d44490cb-89ef-4657-bcbe-793d5861f81c` — bare `<a href="{{.nav_link_N_url}}">`) on a **scratch**
  site (`sites` needs only `domain`) with fields that will not resolve, run a chrome render, assert
  the stored `rendered_html` has **no** `href=""`/`src=""` **and** a `chrome_dead_control` row appears
  at `status='needs_human_review'` (no `handler_agent`), deduped on re-render; and a clean site is
  byte-identical with no row.
- **Trigger caveat (why not done yet):** the only found chrome-render trigger is `rerender-pages` +
  `refresh_site_components=true` (049_TRIGGER_chrome_refresh.sh), which also reassembles + **deploys
  every page** (~26–37 commits to `gqls/sites` + B2 sync + CF purge per run) — unacceptable
  side-effects for a scratch verification. A clean single-action `render_site_components` dispatch
  (no page deploy) is the right vehicle; build/find one before inducing, or piggy-back the next time
  a site with a genuinely ungated chrome component is legitimately re-rendered.

**Still genuinely out of this file's scope (unchanged):** the *general* consumer that drains the
existing unread pile (`unresolved_cta`=68, `cta_names_unknown_destination`=47, `dead_control`=6,
`empty_internal_href`=3 as of 2026-07-22) is `bugs_open/033`'s work now that the owner has ruled it a
queue. Likewise the **page-content** render paths (`rerender_page_sections`, `assemble_from_library`,
`section_editor`) call the `RenderTemplate` wrapper that *discards* `deadURLFields` and have the same
empty-href exposure — but they have their own (partial) nets (`check_dead_controls`,
`validate_page_content`, 023 gating); generalising the drop fleet-wide is a follow-on
(`bug_historian` advisory), not this file's scope.

---

## Note from the idea.uk thread — the signal you escalate on is now accurate (2026-07-22)

The observability prerequisite you build on (`RenderTemplateReportingMissing` / `missingBareFields`)
had a **control-flow-blind regex** detector when it first shipped: it flagged a bare `{{.Name}}` even
inside `{{range}}`/`{{if}}`/`{{with}}` bodies, testing it against the top-level map → **false-positive
`inURLAttr` Errors on ~30 active components fleet-wide**. If your escalation had BLOCKED a build on
`inURLAttr`, those false positives would have blocked ~30 legitimate builds.

That is fixed (commit `78482c86b`, LIVE in **v1.0.1149**): `missingBareFields` is now a scope-aware
`text/template/parse` walk — it reports only **ungated, root-scope** bare fields. So the `inURLAttr`
list you consume is now a true "dead control on this page" set, safe to gate/escalate on. Test:
`platform/orchestration/actions/missing_bare_fields_test.go`.

One thing to know: `RenderTemplateWithMap` (the contact-info render path) was the "second sibling"
`bug_historian` flagged — I routed it through the same detector, but it is **dead code today** (its
only caller `rerenderContactInfo` has no callers; linker-eliminated, `RenderTemplateWithMap`=0 in the
v1.0.1149 binary). So your escalation only needs to cover the live `RenderTemplate` path for now.
— idea.uk vm site thread
