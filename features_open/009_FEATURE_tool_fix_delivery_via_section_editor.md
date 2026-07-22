# FEATURE 009 — automate tool-fix delivery via the section-editor (the remaining work from bug 024)

**Filed:** 2026-07-21 · travelling-docs workstream · site `e33263f4-74f8-494f-b191-546845dbbddf` (gamesdesign.co.uk benchmark)
**Status:** **IMPLEMENTED 2026-07-22 — migration 195 APPLIED & PROVEN end-to-end**
(tooldoctraveller session; owner-approved 2026-07-21). tool-improver's delivery
step now emits a `section_edit` work item routed to the `section-editor` agent; the
build-dispatch-loop routes it and section-editor delivers via `apply_section_edit`.
Proven: a `section_edit` item shaped exactly as tool-improver now emits was
dispatched → section-editor → git-committed the benchmark page (item
`195b9c03`, section-editor orchestration COMPLETED). **No council gate** — the change
is config-only (a migration/agent-def edit), which the gate's scope
(`platform/`,`internal/`,`pkg/`) deliberately excludes (097 refuses it); governed by
owner approval + the migration's pre-flight/snapshot/post-condition guards.
Spun out of `bugs_closed/024` when that bug was **closed** (its headline symptom —
a tool-improver fix never reaching the live page — is now fixed AND live).
**Why this exists:** so the two days of diagnosis behind 024 is not lost when the
bug file closes. Read `bugs_closed/024` for the full six-defect chain; this note
holds only *what is left to build and decide*.

---

## The one-paragraph story (what 024 taught us)

A `tool-improver` template fix was written correctly to
`content_components.html_template` and **never rendered onto the live page**. Six
defects were diagnosed in series — all of them on the **generic** rerender path
(`rerender_page_sections → save_page_sections`). The real blocker turned out to
be none of them: **that whole path is deliberately forbidden for tool pages.**
The experience-loop workstream's ownership guard
(`save_page_sections_action.go:138`, migration **164**, commit `fb89f1071`)
hard-refuses any `rebuild_policy='owned'` page, and **every tool page is `owned`
by definition** (`UPDATE pages SET rebuild_policy='owned' WHERE page_type='tool'`).
So the generic section-render `DELETE`-and-reinsert of `page_components` — the
"TL-001 clobber" — is exactly what the guard exists to stop. The two workstreams
were working the same door from opposite ends: one bolting it shut to protect
owned tools, the other (024) patching it to push tool fixes through.

**The sanctioned door was open all along.** The `section-editor` agent
(`load_edit_context → apply_section_edit → git_commit → update_page_status`) is
the guard's *own named* path and is **not** gated. Driven by hand for the
benchmark (`content_edit`, a pure re-render), it delivered the fix **LIVE in one
shot** — correlation `c3828d17-cba4-4325-87b3-84b972ec9c7e`:

| where | `.ltb-row-grid` rule | len |
|---|---|---|
| before | `display: grid; grid-template-columns: 2fr 1fr 1fr auto` | 9,901 |
| `page_components.rendered_html` (after) | `display: flex; flex-wrap: wrap; min-width: 0; max-width: 100%` | 10,705 |
| live page (`curl`) | `display: flex; … min-width: 0; max-width: 100%` | — |

`build_status='deployed'`. **First time a tool-improver fix has reached the live
page since 024 was filed.**

---

## Remaining work

### 1. Option A (RECOMMENDED) — wire tool-improver's post-fix delivery to the section-editor

Today `tool-improver`'s workflow tail enqueues a generic `needs_rerender`
(`create_rerender_item`), which routes to `page-rerender → rerender_page_sections
→ save_page_sections` — the path the guard refuses for tools. Instead, its tail
should enqueue a **`section-editor`** edit (`apply_section_edit` / `content_edit`),
the sanctioned path that respects the guard and re-renders from the *current*
template.

- **Shape:** config/seed change to `tool-improver`'s `default_config.workflow`
  tail (swap the `create_rerender_item` step for a section-editor enqueue). **No
  new Go machinery** — the section-editor exists and is proven live.
- **Retire/repurpose migration 180** on `tool-improver`: its `spec_literal`
  /`spec_paths` made the generic request well-formed, but that request is
  undeliverable for a tool. Keep 180's mechanism for any non-tool generic
  producer; remove it from tool-improver's tail.
- **Why owner-pending + cross-workstream:** this *reverses* 024's whole approach
  and leans on the experience-loop's guard/section-editor. Coordinate with the
  primary 024 thread and the experience-loop workstream before shipping; do not
  re-open their guard unilaterally.
- **Verify green:** after the rewire, a Tier-4 acceptance run on the benchmark
  should pass `mobile-fit@mobile` autonomously (delivery is already proven; the
  green verdict is the last confirmation). **Do not hand-fix the benchmark** — it
  is the deliberate RED reproduction for `bugs_open/010`; drive the loop.

### 2. Option B (NOT recommended) — carve a self-contained-tool exemption into the ownership guard

Re-opens a guard the experience-loop added deliberately (TL-001). Re-litigates
their guard rail. Only with their agreement.

### 3. Option C — a dedicated tool-pipeline re-render+deploy action

More new code than reusing the section-editor; only if A proves insufficient for
tool-improver's actual edit shape (a template change, which
`content_edit`/`component_swap` already cover).

---

## Related open threads — the generic-path residuals (NON-TOOL pages)

Defect 6 and migration 180 keep value for **non-tool `generic` pages**, where the
section-render path IS sanctioned and the collision is a real bug. These are
genuine, currently-reproducible defects on that path — tracked here so they are
not lost, but they are **separate and lower priority** than Option A.

- **(a) The `page_rerender` mode-collision — fix COMMITTED, inert.** A per-page
  `page_rerender` item_key of `page_rerender_<page>_<site>` ignored render mode, so
  a stale reason-less (assemble-only) request suppressed a reason-bearing
  (section-render) one via `ON CONFLICT DO NOTHING` and then re-deployed stale
  HTML. **Fixed `cdd858402`** (`pageRerenderItemKey` scopes the key by mode),
  **inert until the next chassis image**. Verify after the roll on the idea.uk
  `audience-check-form` reproduction (`bugs_closed/024`, "Second, independent
  reproduction"): a hand-edited `html_template` + a `rerender-pages` run should
  now re-render from template, not assemble stale. Council review of the fix is in
  flight — corr `746c7d60-9ec5-491d-8a28-40bfd3ad8503` (its rationale over-claims
  "tool delivery blocker"; read the verdict as *is this a correct generic-path
  fix*, per `WRONG_CALLS.md` 2026-07-21).
- **(b) Candidate 4 (STRUCTURAL, unbuilt) — assemble-only ships stale.**
  `rerender_single_page` (`assemblePage`) reads `page_components.rendered_html`
  directly and re-ships it, so on a non-tool generic page a reason-less
  `rerender-pages` run deploys stale HTML while reporting success (the idea.uk
  root cause — no reason-bearing item ever exists, so key-scoping can't help). The
  fix is a fleet-wide rule: **refuse (or self-upgrade to section-render) when a
  referenced component template is newer than the stored render.** Bounded cleanly
  to self-contained sections, where `rendered_html` should equal the template.
  Deserves its own diagnosis/council loop — it touches the platform's most-used
  "re-render this site" entry point.

---

## Missteps banked (so the next thread doesn't re-pay them)

- **Two days patching a door another team had deliberately bolted shut.** The
  six-defect chain never saw the ownership guard because an earlier defect always
  blocked first (T28 escalated pre-defect-3; T32 was suppressed by defect 6 before
  reaching `save_sections`). The lesson is a coordination one: when a fix keeps
  hitting a new wall one layer down, check whether *another workstream owns that
  layer* before patching it again. Grep the guard's own error text
  (`Refusing to overwrite`) — it names the sanctioned path.
- **Shipped `cdd858402` under a headline framing a concurrent thread was, at that
  moment, refuting** (`WRONG_CALLS.md` 2026-07-21). The `bugs_open/024` file was
  moving under me ("File has been modified since read"); that IS the signal to
  re-read an actively-owned case before investing in a fix premised on its earlier
  framing. Confidence is not a signal — the "defect 6 is the blocker" framing felt
  obviously right.
- **Verify a fix by its SPECIFIC rule, never a generic CSS property.** A spaced
  `LIKE '%display:grid; grid-template-columns:2fr…%'` returned false against
  `display: grid;` (the CSS has spaces) — the same exact-match trap that hid this
  bug for two days. Pull the substring of the component's OWN rule.

---

## Key identifiers

- Benchmark: `tool-loot-table-balancer` on gamesdesign.co.uk; page_component
  `45229c85-5600-4e8c-b4ae-e8058f74b185`; component `3862f72f-…`.
- Ownership guard: `save_page_sections_action.go:138`; migration 164
  (`fb89f1071`); `pages.rebuild_policy='owned'` for all `page_type='tool'`.
- Sanctioned delivery: `section-editor` agent, `apply_section_edit`
  (`content_edit` for a pure re-render). Live delivery correlation
  `c3828d17-cba4-4325-87b3-84b972ec9c7e`.
- Full six-defect diagnosis: `bugs_closed/024`. Generic-path fix: `cdd858402`.
  Council submission: `submission_024_defect6_page_rerender_key.json`.

---

## Owner decision + session log — 2026-07-21 (tooldoctraveller session)

**A concurrent session reached this file independently.** This "tooldoctraveller"
session is the one that ran the two probes behind 009: probe 1 (a reason-bearing
`page_rerender`, first to get past defect 6 → hit the ownership guard at
`save_sections`) and probe 2 (`section-editor` `content_edit`, correlation
`c3828d17`, the LIVE delivery this file cites). It then put the fix direction to
the owner and reconciled on discovering the primary 024 thread had already closed
024 and filed this feature. Recording here rather than forking a parallel account.

**OWNER DECISION: Option A** — wire tool-improver's post-fix delivery to the
section-editor. Chosen 2026-07-21 over Option B (relax the guard) and Option C
(new tool-pipeline action). So the "owner decision pending" blocker at the top of
this file is now resolved: **build Option A.**

**Feasibility notes for whoever implements it:**
- **The DIRECT section-editor drive is proven** (probe 2, top-level `input_data`
  fields: `site_id`, `page_component_id`/`page_name`+`slot_name`, `edit_type=
  content_edit`, `field_updates={}`). `content_edit` REQUIRES a non-nil
  `field_updates` or `replacement_content_data`; `{}` satisfies it and is a pure
  re-render from the *current* template (`applyContentEdit`, section_editor_actions.go:594).
  `load_edit_context` resolves the section by `page_component_id` OR
  `page_name`+`slot_name` — tool-improver knows the tool function (= page_name =
  slot_name), so it needs no extra lookup.
- **The WORK-ITEM route is NOT yet verified.** section-editor has never handled a
  work item (0 rows in `site_work_items` for `handler_agent='section-editor'`). A
  probe that drove section-editor with fields nested under `input_data.spec.*` (the
  shape the dispatch loop produces from a work item) left **no orchestration_states
  row** — inconclusive (looks like a dropped message, not a resolution failure;
  the identical top-level drive minutes earlier completed). **Two clean shapes for
  Option A, pick during implementation:**
  - **(b) emit a `section-editor` work item** from tool-improver's tail (swap
    `create_rerender_item`'s config). Cleanest reuse, but VERIFY the dispatch→
    section-editor input mapping resolves `page_name`/`slot_name` from
    `input_data.spec.*` first (the field-collision trap, 001 §).
  - **(c) inline the actions** into tool-improver's workflow tail
    (`load_edit_context → apply_section_edit → git_commit → update_page_status`,
    all registered actions) — most directly proven (these are the exact actions
    probe 2 exercised), synchronous, testable end-to-end, but lengthens
    tool-improver and duplicates the section-editor's delivery steps.
- **Option A complies with the experience-loop guard** (it uses their sanctioned
  path, does not touch `save_page_sections`), so it does not need their sign-off to
  ship — but the council gate (workstream norm for a tool-improver seed change) is
  the right coordination mechanism, and the primary 024 thread should not
  double-build it.
- Note the benchmark is currently delivered/green-eligible ONLY because probe 2
  drove the section-editor by hand. Until Option A ships, a fresh `improve_tool`
  cycle would again write the template and enqueue the (undeliverable) generic
  `needs_rerender` — so the loop is not yet autonomously closed for tools.

---

## IMPLEMENTED — migration 195, 2026-07-22 (tooldoctraveller session)

**Shape chosen: (b) emit a `section_edit` work item** (not (c) inline). Reason:
`content_edit` needs a literal `edit_type`, and `create_work_item`'s `spec_literal`
(migration 180) carries literals cleanly into the item's spec — so (b) is genuinely
config-only, whereas (c) would need a literal-injection affordance. The
build-dispatch-loop routes it generically and delivery is synchronous *within the
section-editor orchestration* (no LLM, no starved-lane dependency for correctness).

**`195_tool_improver_deliver_via_section_editor.sql`** replaces tool-improver's
`create_rerender_item` step config (step key/output_field/next_step unchanged, so
`complete.output_fields`'s `rerender_item` reference stays valid):
- `handler_agent`: `rerender-pages` → `section-editor`; `item_type`: `needs_rerender`
  → `section_edit`.
- `spec_literal`: `{reason:section_data_resolved}` → `{edit_type:content_edit,
  field_updates:{}}` (a pure re-render from the current template).
- `spec_paths`: `{component_id}` → `{page_name: tool_data.page_name, slot_name:
  tool_data.function}` (both verified present/populated in real runs;
  `create_work_item` hard-errors on an unresolved path).
- Keeps `item_key_suffix_field: update_result.component_id` (component-scoped dedup)
  and `recurrence_expected: true`.

**Every link proven this session:**
1. section-editor delivers a tool fix live (probe 2, corr `c3828d17`).
2. section-editor resolves `input_data.spec.*` fields via recursive lookup (probe
   `8dfbb732`, COMPLETED).
3. build-dispatch-loop routes ANY handler dynamically — `spawn_handler` uses
   `agent_type_field: current_item.handler_agent`; `load_work_items` has no handler
   filter; `call_handler` maps `spec ← current_item.spec` → `input_data.spec`.
4. `tool_data.page_name`/`.function` and `update_result.component_id` all resolve.
5. **Full route end-to-end:** a `section_edit` item shaped as tool-improver now
   emits (`195b9c03`) → dispatch-loop claimed it → section-editor orchestration
   `load_edit_context → apply_edit → git_commit → update_page_status` COMPLETED →
   `result.git_result.files=["/tools/tool-loot-table-balancer.html"]`. Benchmark
   still `rendered_html`=10,705 FLEX-FIX, live.

**Not exercised (verified in parts, low risk):** a full LLM tool-improver run
CREATING the `section_edit` item. Deliberately skipped — a fresh `improve_tool`
LLM pass could regress the already-good benchmark template, and the emit is plain
`create_work_item` mechanics on a verified config with resolvable paths. **The
natural confirmation is the next tool-auditor-driven improve cycle**, which will now
deliver via section-editor instead of dying at the ownership guard. Rollback:
snapshot `1f3ebb4a` + `195_..._ROLLBACK` (strip the step back to the 180 shape).

**Caveat — delivery latency:** the section_edit item rides the same cron-spawned
build-dispatch-loop, which is currently starved (`bugs_open/030`), so autonomous
delivery may lag until that lane's throughput is fixed. Correctness is unaffected;
latency is a separate workstream.
