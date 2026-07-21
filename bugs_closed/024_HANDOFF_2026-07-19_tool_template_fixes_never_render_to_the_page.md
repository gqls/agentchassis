# BUG 024 — a tool-improver fix is written durably and NEVER reaches the live page

**Filed:** 2026-07-19 · travelling-docs thread · site `e33263f4-74f8-494f-b191-546845dbbddf` (gamesdesign.co.uk)
**Severity:** high — silently defeats the whole self-verifying fix loop for tool components.
**Status:** **CLOSED 2026-07-21 (fixed AND live) — owner-directed close.** The
headline symptom — a tool-improver fix never reaching the live page — is resolved
and **verified live**: the benchmark's `.ltb-row-grid` rule ships `display: flex;
… min-width: 0; max-width: 100%` (was the broken `grid`), `rendered_html` 10,705
bytes, `build_status='deployed'`, delivered via the sanctioned `section-editor`
path (correlation `c3828d17-cba4-4325-87b3-84b972ec9c7e`). The **remaining work is
spun out to `features_open/009`** so it is not lost: (1) Option A — wire
tool-improver's tail to the section-editor so delivery is automatic (owner
decision pending, touches the experience-loop workstream); (2) the non-tool
generic-path residuals — the `page_rerender` mode-collision fix (`cdd858402`,
committed, inert until the next image; council corr `746c7d60`) and candidate 4
(assemble-only ships stale). This file is the durable diagnosis record; the live
tracker is `features_open/009`.

**RE-SCOPED 2026-07-21 (kept for the record).** The whole defect-1…6 chain patches
the **generic** rerender path (`rerender_page_sections` → `save_page_sections`).
That path is **deliberately forbidden for tool pages** by the experience-loop's
ownership guard (migration 164, `fb89f1071`), and **every tool page is
`rebuild_policy='owned'` by definition**. The sanctioned delivery — the
`section-editor` agent (`apply_section_edit`) — **works, and delivered the
benchmark fix LIVE** (see "UPDATE 2026-07-21 (later)" immediately below). Defect 6
is moot for tools; **migration 180's request is well-formed but aimed at a path
that cannot deliver a tool.** The remaining work is a re-wire (tool-improver →
section-editor), not another patch to the generic path.

---

## UPDATE 2026-07-21 (later) — the delivery path was WRONG the whole time; the sanctioned path works and is now LIVE

**The single sentence:** every defect in this file (1–6) has been patching the
**generic** section-render→save path, but a deliberate guard from the
**experience-loop** workstream forbids that path for tool pages — and the
**sanctioned** path (`apply_section_edit` via the `section-editor` agent) delivers
the fix correctly, **proven live on the benchmark today**.

### How it was found

With defects 3 + 5 fixed and live in **v1.0.1144** (pod-verified:
`isSelfContainedSection`, `toolTemplateValid`/`componentTemplateValid`,
`loadSingleComponentSchema` all present), I drove a correctly-formed reason-bearing
`page_rerender` for the benchmark **directly** (kafka, unique item_key — bypassing
defect 6's collision; the page-rerender dispatch lane is cron-starved, ~1
completion in 6h). This is the **first proof run in this bug's history to get PAST
defect 6 and reach `save_sections`.** Result:

- `rerender_page_sections` **rendered the tool from its template and did NOT
  escalate** — so defects 3 + 5 are **sufficient**. The render logic works.
- Then `save_sections` **FAILED**:
  > `page tool-loot-table-balancer is rebuild_policy=owned (tool/widget-owned): a
  > generic section save would clobber it. Use apply_section_edit for targeted
  > edits or the tool pipeline for rebuilds. Refusing to overwrite.`

Nobody in this bug's six-defect history had ever seen this guard, because an
**earlier defect always blocked first** (T28 escalated pre-defect-3; T32 was
suppressed by defect 6 before reaching `page_rerender` at all).

### The guard is deliberate, and forbids exactly what defects 1–6 attempt

`save_page_sections_action.go:138-160` — "guard rail 1, experience loop"
(`fb89f1071`, migration **164**). `save_page_sections` DELETE-and-reinserts
`page_components`; on a tool-owned page that is "the TL-001 clobber," so it
**hard-refuses**. Migration 164's own notes:

- `UPDATE pages SET rebuild_policy='owned' WHERE page_type='tool'` — **every tool
  page fleet-wide is owned**; the benchmark is `rebuild_policy=owned`, confirmed.
- *"apply_section_edit and the tool pipeline remain the edit paths."*
- *"page_rerender/**assembly** is NOT gated — re-assembly of existing
  page_components is how owned pages deploy."* — the experience-loop **kept the
  assemble path (rerender_single_page) open on purpose** and gated ONLY the
  section-render **write**. The two workstreams were working the same page from
  opposite ends: one bolting the section-write door shut to protect owned tools,
  the other (this bug) patching that same door to push tool fixes through it.

**Consequence for migration 180 / defect 6.** 180 makes tool-improver's
`needs_rerender` carry `reason=section_data_resolved` so `page-rerender` takes the
section-render branch. That branch ends at `save_page_sections`, which now refuses
tool-owned pages. So **180's request is correct but undeliverable for a tool**, and
**defect 6 (reason-scoping the generic key) is moot for tools** — fixing it only
lets the request reach the guard that refuses it. (Defect 6 / 180 may retain value
for **non-tool** `generic` pages — the idea.uk audience-check-form reproduction
below is one — but that is a separate, lower-priority track, not tool delivery.)

### The sanctioned path works — PROVEN LIVE

The `section-editor` agent is active and its workflow is a complete delivery:
`load_edit_context → apply_section_edit → git_commit → update_page_status`.
`apply_section_edit`:
- `content_edit` re-renders the component from its **current** template (loaded
  fresh from `content_components`) with `content_data` as source of truth, then
  UPDATEs `page_components.rendered_html`, reassembles the page, and returns HTML
  for `git_commit`. It is the guard's own named path, so it is **not** gated.

I drove `section-editor` `content_edit` (`field_updates={}`, a pure re-render) for
the benchmark. Orchestration **COMPLETED**. Result, verified live:

| where | `.ltb-row-grid` rule | len |
|---|---|---|
| before | `display: grid; grid-template-columns: 2fr 1fr 1fr auto` | 9,901 |
| `page_components.rendered_html` (after) | `display: flex; flex-wrap: wrap; min-width: 0; max-width: 100%` | 10,705 |
| **live page** `curl gamesdesign.co.uk/tools/tool-loot-table-balancer.html` | `display: flex; flex-wrap: wrap; … min-width: 0; max-width: 100%` | — |

`page_components.build_status='deployed'`. **The tool-improver fix is on the live
page for the first time since this bug was filed.** Correlation
`c3828d17-cba4-4325-87b3-84b972ec9c7e`.

### The fix (owner decision pending — reverses this bug's approach, touches another workstream)

**Recommended — Option A:** wire tool-improver's post-fix delivery to the
**section-editor** (`apply_section_edit`) instead of the generic `needs_rerender`.
No new machinery; the section-editor exists and is proven; it respects the
experience-loop guard. Retire/repurpose migration 180's generic-rerender config on
tool-improver. This is a config/seed change to tool-improver's workflow tail
(swap `create_rerender_item` for a section-editor enqueue), not a Go patch.

**Option B (not recommended):** carve a self-contained-tool exemption into the
`save_page_sections` ownership guard. This re-opens a guard the experience-loop
added deliberately (TL-001) — re-litigating their guard rail; coordinate with that
workstream, don't do it unilaterally.

**Option C:** a dedicated tool-pipeline re-render+deploy action. More new code than
reusing the section-editor; only if A proves insufficient for tool-improver's
actual edit shape (a template change, which `content_edit`/`component_swap`
already cover).

**Verify green:** now that the fix is live, a Tier-4 acceptance run should pass
`mobile-fit@mobile`. Delivery is PROVEN; the green verdict is the last
confirmation.

---

## UPDATE 2026-07-21 (00:xx) — the request is finally correct, and it exposed a SIXTH defect that still blocks delivery

The whole fix (Go in v1.0.1140 + migration 180) is applied and **the re-render
request is now correctly formed for the first time in this bug's history.** A
real proof run (cloned acceptance-driven `improve_tool` item `216ea5fe`) drove
`tool-improver`, which emitted `needs_rerender` row `666619d1` carrying:

| field | value | fix that put it there |
|---|---|---|
| `item_key` | `rerender_tool_fix_gamesdesign.co.uk_3862f72f-…` | `item_key_suffix_field` (component-scoped) |
| `spec.reason` | `section_data_resolved` | `spec_literal` |
| `spec.component_id` | `3862f72f-…` | `spec_paths` |
| `status` | `triaged` (not `[unresolved after 2 attempts]`) | `recurrence_expected` |

All four of migration 180's changes proven in one row. **But the page still did
not render** — `rendered_html` is still the 9,901-char v1 artifact with the
broken `display:grid; grid-template-columns:2fr 1fr 1fr auto` (verified in the
stored bytes; `flex-wrap:wrap` is a FALSE marker — it exists elsewhere in the v1
render, so do not use it to prove delivery).

### Defect 6 — the per-page `page_rerender` item_key is not reason-scoped, so a stale reason-less request suppresses the real one

`create_rerender_items_action.go:248` builds
`itemKey := fmt.Sprintf("page_rerender_%s_%s", pageName, siteID)` — **site+page
scoped, blind to `reason` and `component_id`** — and inserts with `ON CONFLICT
DO NOTHING`. This is the same collision class as defect 4, **one layer down**,
and migration 180 does not touch it (180 scoped the `needs_rerender` key; this
is the `page_rerender` key `create_rerender_items` generates from it).

Timeline, all from `claimed_at`/`completed_at` (NOT `updated_at`, which is
unmaintained — `bugs_open/035`):

- **2026-07-20 18:50** a reason-less `page_rerender` (`b5dbd732`, born before
  migration 180 from an earlier run) is created and sits in the dispatch backlog.
- **21:45–21:46** my `needs_rerender` `666619d1` runs. `create_rerender_items`
  computes `scoped=true`, finds the dependent page, and tries to insert
  `page_rerender_tool-loot-table-balancer_<siteID>` — **collides with the open
  `b5dbd732` and inserts nothing.** Zero per-page items. `items_created: 0`.
- **22:51–22:53** `b5dbd732` is finally claimed and runs. With **no reason** it
  takes `check_rerender_mode`'s `else_step` → `rerender_single_page`
  ("Simple concatenation - no template re-rendering") → re-deploys the **stale**
  HTML and sets `build_status='deployed'`.

So the reason-less request both **blocked** the reason-bearing one (dedup) and
then **overwrote** the outcome with a stale assemble-only deploy. Net: the
correct request was produced and silently discarded.

### Why this was invisible until now

Before migration 180 the `needs_rerender` never carried a reason, so every
`page_rerender` was reason-less and the collision was between identical
requests — harmless. Making the request correct is exactly what exposed the next
collision. This is the third time this bug's chain has hidden the next link
behind the one in front of it.

### Fix candidates for defect 6

1. **Scope the `page_rerender` item_key by reason (and/or component_id)**, so a
   `section_data_resolved` request and a reason-less one are different keys and
   cannot suppress each other. Mirrors migration 180's key scoping, but in Go
   (`create_rerender_items_action.go:248`), so it is image-gated, not config.
   ⚠️ Not sufficient alone: two `page_rerender` items would then both run, and if
   the assemble-only one runs LAST it re-deploys stale. Needs a companion rule so
   a section-render is not overwritten by a later assemble-only (e.g. don't
   assemble-deploy when the component template is newer than the stored render —
   this is 024 candidate 4, and defect 6 is the strongest evidence for it).
2. **Make the reason-bearing request UPGRADE a pending reason-less one** rather
   than dedup against it: on conflict, if the incoming spec has a section-render
   reason and the existing row does not, update the existing row's spec instead
   of `DO NOTHING`.
3. **Drain/supersede stale reason-less `page_rerender` items** for a page when a
   reason-bearing one arrives.

This deserves the diagnosis/council loop, not a rushed patch — the interaction
between key scoping and last-writer-wins is exactly where a naive fix reopens the
bug. Recommended for the next thread; see the handoff §7 resume block.

### The verify query in this file is STALE

The "How to verify a fix" query below keys on `minmax(0, 2fr)`. The improver now
writes a **different, equally valid** fix (`display:flex; flex-wrap:wrap;
min-width:0` on `.ltb-row-grid`), so `minmax(0, 2fr)` will never appear. The real
delivery proof is: **the rendered `.ltb-row-grid` rule stops saying
`display:grid; grid-template-columns:2fr 1fr 1fr auto`** and matches whatever the
current template says, AND `length(rendered_html)` leaves 9,901. Match the
component's OWN rule, never a generic property (the T24/T28 trap, restated).

---

## Symptom

`tool-loot-table-balancer` has failed `mobile-fit@mobile` on every Tier-4 acceptance
run since 2026-07-17, across **three** tool-improver cycles. Each cycle wrote a
plausible, correctly-aimed fix and reported success. Every re-verification came
back RED with a byte-identical signal (`fieldset (419px)`).

This was diagnosed in `bugs_open/010` as *fix-loop non-convergence* — the fixer
aiming at the wrong element. That diagnosis drove a real improvement (the
drill-down attribution, T25), but it was **not the reason re-verification stayed
RED**.

## The actual finding

**The live page has never changed since the tool was born.** All three fixes are
present in the durable component template; none has ever been rendered.

Proof (2026-07-19):

| where | `.ltb-row-grid` rule | `fieldset` rule | len |
|---|---|---|---|
| `content_components.html_template` | `minmax(0,2fr) minmax(0,1fr) minmax(0,1fr) auto` | has `min-width:0; width:100%; max-width:100%` | 10,626 |
| `page_components.rendered_html` | `2fr 1fr 1fr auto` | bare, no constraints | 9,901 |
| live page (curl) | `2fr 1fr 1fr auto` | bare, no constraints | — |

`9,901` is the **v1 born length** (`component_versions.version_number = 1`).
`page_components.rendered_html` has not been regenerated since birth.

## Root cause — three defects in series

### (1) `update_component_html` relies on a flag nothing reads
`platform/orchestration/actions/update_component_html_action.go:248-255` writes
the new template to `content_components.html_template`, then sets
`page_components.build_status='pending'`, with the comment *"The rerender pipeline
regenerates rendered_html"*.

Nothing anywhere filters on `build_status='pending'`. The codebase says so itself
at `store_generated_component_action.go:455-458`: *"Without this the
build_status=pending flag is informational only — nothing downstream scans
page_components for pending rows."* **The flag is dead state.**

### (2) tool-improver's rerender request routes to assemble-from-stale
The page-rerender agent gates on the work item's `spec.reason` (LIVE definition):

```
condition: input_data.spec.reason == 'image_landed'
        OR input_data.spec.reason == 'section_data_resolved'
        OR input_data.spec.reason == 'cta_links_stale'
then_step: rerender_sections     # true template re-render
else_step: render_page           # rerender_single_page — assemble STORED html
```

`tool-improver`'s `create_rerender_item` config sets no `spec_data` at all, so
**no `reason`** → always `else_step` → `rerender_single_page`, whose own header
says *"Simple concatenation - no template re-rendering"*. It reads the stale
`rendered_html`, assembles it, deploys it, and marks the page `deployed`.

So the fix loop reports a successful deploy of an unchanged page. This is the
`complete` ≠ *the work happened* invariant, one layer deeper than usual: the work
item, the orchestration and the page status are all genuinely green.

### (3) even on the correct branch, a tool section escalates instead of rendering
Forcing the good branch (`reason=section_data_resolved`, item `478c44c9`) proved
the gate diagnosis right — `rerender_sections` ran — and then exposed the last
defect. `rerender_page_sections_action.go:186-206` refuses to render any section
whose `content_data` is empty, escalating the whole page to a full rebuild:

```go
if len(s.contentData) == 0 { reason = "no stored content_data" }
... escalateRerenderToWriter(...) ; out["escalated"] = true ; return out, nil
```

A **tool** section legitimately has `content_data = {}` — a tool is self-contained
HTML with no LLM-authored content fields (verified: `cd_type=object, cd_len=2`,
and the component has **no `input_schema`** at all).

`check_escalated` then routes `escalated==true` → `complete`, bypassing
`save_sections` — the only step that writes `rendered_html`. The render is
computed and thrown away.

That guard is correct for its original purpose (it is the fix for the blanked
article bodies, `bugs_closed/004` / `bugs_closed/005`). It simply has no concept
of a section type that has no content_data *by design*.

**Net: there is no path by which a tool template fix can reach a tool page.**

## Fix candidates

1. **(smallest, unblocks the loop)** Give tool-improver's `create_rerender_item`
   a `spec_data.reason`, and teach the guard that a section with no
   `input_schema` needs no `content_data` — render it from the template instead
   of escalating. Both are needed; either alone leaves the path dead.
2. **Escalate to the right handler.** For a tool page the escalation currently
   emits `needs_page` (full rebuild by the writer). That is the wrong remedy —
   a rebuild regenerates the tool and would destroy it (the known
   *re-plan clobbers built pages* landmine). If a tool section ever does escalate,
   it should route to the tool pipeline, not the page writer.
3. **Kill or honour `build_status='pending'`.** It reads as a working mechanism
   and is not one; it is what makes defect (1) look wired. Either make something
   scan it, or remove it and its comments.
4. **Structural:** `rerender_single_page` deploying stale HTML while reporting
   success is the load-bearing lie. It could compare the assembled output against
   the component templates it references and refuse (or flag) when a referenced
   template is newer than the stored render.

## How to verify a fix

```sql
-- component template vs stored render must agree on the fixed rule
SELECT (cc.html_template LIKE '%minmax(0, 2fr)%') AS template_has_fix,
       (pc.rendered_html LIKE '%minmax(0, 2fr)%') AS render_has_fix
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.id = '45229c85-5600-4e8c-b4ae-e8058f74b185';
```
Then `curl -s https://gamesdesign.co.uk/tools/tool-loot-table-balancer.html | grep 'ltb-row-grid' -A3`
must show the `minmax(0, …)` columns. Then a Tier-4 acceptance run should finally
go GREEN on `mobile-fit@mobile`.

## Consequences for other records

- **`bugs_open/010` needs re-reading.** Its non-convergence evidence (two
  identical RED re-verifications, two "materially identical" fixes) is explained
  by an unchanged page, not by a fixer that cannot aim. Its candidate (a)
  (drill-down) was still a genuine improvement; candidate (b) (convergence guard)
  is still worth having, but it would have fired here on a loop that was never
  actually being given a chance.
- **T24's claim that "the durable fix reached the live page (`max-width` present
  ×10)" is wrong.** `max-width` appears 10× on that page in unrelated site-chrome
  rules; the tool's own `fieldset` rule has none. Verifying a specific fix by
  grepping a *generic* CSS property is the trap — match the specific rule.
  (This thread made the identical mistake once today before catching it: see
  NOTES 2026-07-19.)

## Scope

Not tool-specific in principle: **any** component whose fix is written via
`update_component_html` and whose section carries no `content_data` is affected.
Tools are simply the population where that is the normal shape.
Adjacent to `bugs_open/021` (the durable write guard covers one path only) — same
`page_components.rendered_html` surface, different failure.

---

## Second, independent reproduction — 2026-07-20, idea.uk, and the scope is WIDER than tool components

The idea.uk thread hit defect **(2)** (rerender routes to assemble-from-stale) from a completely
different direction: no tool-improver, no `update_component_html`, just a hand-edited
`content_components.html_template` plus a plain `rerender-pages` run. Same outcome.

**What was edited:** `audience-check-form`'s template (added a `<script src>` ref and an
`#ac-result` div) and `tool-list`'s resolved URL, via `sql/p3_03`.

**What the platform reported:** total success. **9/9** `page_rerender` items `complete`; a real
commit per page to `gqls/vm-sites`; and the JS asset genuinely published —
`/tools/assets/audience-check-form.js` → **HTTP 200, 1469B**.

**What actually changed on the page:** nothing. No `<script>` ref, no `#ac-result`, cards still
`href="/audience-check"`.

**Why it matters for this bug file:**

1. **The trigger is not tool-specific.** `rerender-pages` creates its `page_rerender` items with
   **no `spec.reason` at all** — verified: `{"domain","page_id","filename","page_name"}` and nothing
   else. So *every* item it produces takes `else_step: render_page`. This is not "the tool-improver
   sends the wrong reason"; it is **the general-purpose site rerender entry point that cannot ever
   re-render a template**, for any component, on any site. Anyone editing an `html_template` and
   reaching for the obvious rerender gets a green run and no change.
2. **`rerender-pages` exposes no way to set the reason.** Its input takes `site_id`, `domain`,
   `refresh_site_components` — there is no reason parameter to pass through. The only route to
   `rerender_sections` found was inserting `page_rerender` items by hand with
   `reason='section_data_resolved'` (`sql/p3_04`). That worked: both pages re-rendered from the
   template and deployed correctly, verified live.
3. **A partial success is the worst signal.** The asset published while the tag referencing it did
   not, because `collectJSAssets` reads `content_components.js_content` **directly** rather than the
   rendered HTML. Result: a file that exists, returns 200, and nothing loads. Anyone spot-checking
   "did my change ship?" by curling the asset gets a false green.
4. **I nearly mis-diagnosed it the other way.** Seeing stale `page_components.rendered_html`
   timestamps I first concluded "reported complete, did nothing" — wrong; the deploys were real. The
   work item's own `result` JSON lists the deployed files verbatim and is what settles it. *A stale
   `page_components.rendered_html` is not evidence that a rerender did nothing.*

**Correction to this file's quoted gate.** The condition above is right (3 reasons — the LIVE
`page-rerender` definition confirms `image_landed OR section_data_resolved OR cta_links_stale`), but
note the **source comment in the Go is stale**: `rerender_page_sections_action.go:47-51` documents
only two (`image_landed OR section_data_resolved`) and omits `cta_links_stale`. Read the agent
definition, not the comment.

**Suggested addition to the fix candidates:** whatever else changes, `rerender-pages` should either
pass a reason through or default to the section-rerender path. As it stands the platform's most
obvious "re-render this site" button is the one that cannot apply a template change, and it reports
success while doing so.
