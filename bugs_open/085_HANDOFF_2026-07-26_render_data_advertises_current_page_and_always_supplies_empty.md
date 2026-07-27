# 085 — the render data advertises `current_page` and the build path always supplies it empty

**Filed:** 2026-07-26, by the brochure_component_library workstream, which hit it building
the shared `evidence-chart` component. Found by measuring a rendered page, not by reading code.
**Severity:** Low-Medium — no data loss, no wrong figures. It silently removes a capability:
**no section component can know which page it is on**, so nothing can vary per page.
**Class:** structural (a key exists in the contract, is always empty on the main path, and
fails by doing nothing rather than by erroring).
**Status:** OPEN. Build path FIXED and LIVE (v1.0.1173); scoped-re-render path fixed in code
and inert until the next roll. **Read the two dated sections at the foot before anything
above them** — "the fix is one line", asserted twice in the original text below, was wrong
twice over: it was three points on one journey, and then a fourth on a sibling path that
only a live test found. The original text is left standing, uncorrected in place, because
how the sizing went wrong is the transferable part.

---

## Symptom, measured

`evidence-chart` was placed on `index` and `capabilities`. Its chart definitions declare
which page they belong to (`"pages": ["index"]`), and the template filters on `current_page`,
degrading to "show everything" when that value is empty.

**Every chart rendered on `index`, including the two marked `capabilities`.** So the filter
never ran, which happens only when `current_page` is empty.

```sql
-- the rendered section carries all three charts, not the one assigned to this page
SELECT substring(pc.rendered_html from 'data-chart="[a-z-]*"')
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'fundamentallyai.com' AND p.name = 'index' AND cc.function = 'evidence-chart';
```

## Cause, read from the code

`current_page` **is** in the template data map — twice:
`platform/orchestration/actions/component_library.go:756` and `:873` both emit
`"current_page": ctx.CurrentPage`. So a component author reasonably concludes it is available.

`RenderContext.CurrentPage` (`component_library.go:79`) is set in the multipage path
(`multipage_actions.go:206`) but **never by `BuildRenderContextAction`**
(`v3_site_actions.go:866`), which is what the page-build pipeline uses.

That action merges its configured sources through `mergeIntoRenderContextEnhanced`
(`v3_site_actions.go:1022`). The page-content-writer's step config does pass the page:

```json
"sources": { "page": "input_data.current_page", "site": "input_data.site_record", ... }
```

but the merge extracts only a fixed allowlist — domain, company_name, logo_text, tagline,
email, phone, colours, a handful of image-URL fields — and drops everything else on the
floor. It never assigns `ctx.CurrentPage`, and the page record's `name` reaches no other
key either. **The page's identity is passed in and thrown away.**

## Why it is worth fixing rather than working around

- It fails **silently and plausibly**: a component that branches on `current_page` renders
  the "no information" branch, which looks like a design choice rather than a defect.
- The contract advertises the field, so the next author will make the same assumption.
- It is the difference between a component that can be placed twice with different content
  and one that must be placed once. That is a general capability, not one component's need.

## Fix candidate (one line, plus a test)

In `BuildRenderContextAction`, after the sources are merged, set `CurrentPage` from the
already-configured page source — the value is right there in `params.CollectedData`:

```go
if pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.current_page.name"); pageName != "" {
    renderCtx.CurrentPage = strings.TrimSuffix(pageName, ".html")
}
```

Alternatively assign it inside `mergeIntoRenderContextEnhanced` when `sourceName == "page"`,
which fixes every caller at once but touches a shared function — the more invasive of the two.

**Verification that would settle it** (this is the cheap part): re-render
fundamentallyai.com's `index` and `capabilities` after the roll and check that each page
carries only the charts assigned to it —

```sql
SELECT p.name, substring(pc.rendered_html from 'data-chart="[a-z-]*"')
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'fundamentallyai.com' AND cc.function = 'evidence-chart';
```

The data to prove it with is already seeded: `evidence_base.charts` holds one chart marked
`pages: ["index"]` and two marked `pages: ["capabilities"]`.

## Containment now

The `capabilities` placement was removed and the section left on `index` only, so the site
does not publish the same three charts twice. The `pages` key stays in the data, unused and
correct, so the fix turns it on rather than requiring new data.

## What this is NOT

- Not a claims or evidence defect: every figure rendered is a fact row, and the wrong-page
  charts were accurate, merely misplaced.
- Not `bugs_open/068` (rebuild writer/link-resolver contract) — different path, different
  field, no overlap beyond both touching the rebuild pipeline.

---

## 2026-07-27 — FIXED IN CODE, STILL OPEN (inert until the roll)

Committed by the brochure_component_library thread. **Stays OPEN**: the defect is
reproducible on the live fleet until an image carrying it rolls, which is the bar
`/bugs_closed/README.md` sets.

> **CORRECTED 2026-07-27 — "the fix is one line" (above, twice) was wrong, and wrong
> in the same shape as the bug.** Following the value end to end instead of reading
> the one function found it dropped at **three** points on one journey, each of which
> looks complete in isolation:
>
> 1. `BuildRenderContextAction` never assigns `CurrentPage` — the cause as filed.
> 2. **`renderCtxToMap` (`v3_site_actions.go:1316`) does not emit `current_page` at
>    all**, so a correctly-set struct field would not survive the step boundary into
>    `collected_data`.
> 3. **`mergeIntoRenderContext` (`:1428`) does not restore it** on the render side.
>    Its catch-all copies every key into `ContentData`, which is enough for the
>    html/template path (ContentData wins in `contextToInterfaceMap`) but **not** for
>    the regex fallback: `contextToMap` skips any `ContentData` key the base map
>    already holds, and the base map holds an empty `CurrentPage`.
>
> Fixing only (1) — the filed one-liner — would have left the field empty at every
> template and looked exactly like a failed fix. What caught it was measuring the
> serialised context rather than trusting the struct.

**The proposed candidate would also have missed.** It read
`input_data.current_page.name`; the live payloads use `name` on the
page-content-writer envelope and `page_name` on the rerender/page-build ones, and one
observed shape carries both. The shipped `resolveCurrentPageName` reads the path the
step config designates (`sources.page`) and tries both keys, stripping `.html` to match
`buildHeaderConfig`.

**Live evidence, re-measured 2026-07-27** (the original was inferred from a rendered
page; this is the mechanism itself):

```sql
SELECT collected_data ? 'render_context'                    AS has_rc,      -- t
       collected_data->'render_context'->>'domain'          AS rc_domain,   -- populated
       collected_data->'render_context' ? 'current_page'    AS rc_has_cp    -- FALSE, every row
  FROM orchestration_states
 WHERE COALESCE(owner_agent_type,'') = 'page-content-writer'
   AND jsonb_typeof(collected_data->'input_data'->'current_page') = 'object'
 ORDER BY created_at DESC LIMIT 6;
```

The key is **absent**, not merely empty — with `domain` and `company_name` alongside it
as the positive control. And `build_render_context` has exactly **one** caller
fleet-wide (surveyed unfiltered across all active `agent_definitions`), so the blast
radius is the page-build path.

The two-render-path asymmetry was measured directly, not reasoned: with only the
`ContentData` catch-all in place, `html/template path -> "capabilities"`,
`regex fallback path -> ""`.

**Test:** `platform/orchestration/actions/render_context_current_page_test.go`. It
follows the value through the whole chain rather than asserting each function, and both
round-trip halves were proven load-bearing by deleting each in turn and watching it fail
with a distinct message.

**Owed at the next roll**, in order:

1. Pod-grep a symbol the change *created*, with a negative control:
   `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c resolveCurrentPageName'`
2. Restore the `capabilities` placement of `evidence-chart` (removed as containment —
   see below). The data is already correct: one chart marked `pages: ["index"]`, two
   marked `pages: ["capabilities"]`.
3. Re-render both pages and run the verification query above this section. **Induce the
   failing branch too** — a page that carries the section but matches no chart must
   render nothing rather than everything.

---

## 2026-07-27 (later) — LIVE on v1.0.1173 for the BUILD path, and a FOURTH drop point found by testing it

Chassis `v1.0.1173` carries the fix. Pod-verified with controls
(`agent-chassis-5f85dff548-8d2tq`): `resolveCurrentPageName` = 6, the unique log
string = 1, an invented control = 0, a pre-existing string = 1.

**And the feature still did not work.** A scoped section re-render of
fundamentallyai.com/`index` at 14:08 UTC on that binary re-rendered the section
(`page_components.updated_at 14:08:17`) and it **still carried all three charts**,
two of which declare `pages: ["capabilities"]`.

> **Cause: `RerenderPageSectionsAction` never calls `BuildRenderContextAction`.**
> It assembles its own ambient base in `buildRerenderBaseData` (`:496`) — seeded
> with `domain` and `year` only, plus keys from `sites.content_data` — and merges
> it via `mergeIntoRenderContext`. This round's fix made that merge *restore*
> `current_page` from a map, so the plumbing works; nothing ever put the key in the
> map on this path. 016b §9: *a fix applied to one branch of a two-branch router
> reads as done, and the other branch keeps the bug.*

**The failing run was HEALTHY, which is the half of this evidence that matters.** The
orchestration finished `COMPLETED / complete`, `error` NULL, no `__step_error` — so
the section really was re-rendered by the new binary, not left stale by a run that
died before writing. A failed run produces the same symptom for an entirely different
reason and the diagnosis would have been wrong. Confirmed on the SERVED page too,
because a stored render is not proof of what a visitor gets:

```
$ curl -fsS https://fundamentallyai.com/index.html | grep -oE 'data-chart="[a-z-]+"' | sort -u
data-chart="council-review-outcomes"      <- declared pages: ["capabilities"]
data-chart="news-pipeline-credibility"    <- declared pages: ["capabilities"]
data-chart="relojistas-feed-restoration"  <- the only one that belongs on this page
```

Stored row and served page agree exactly (`page_components.updated_at
2026-07-27 14:08:17`), so this is the current state of the live site, not a stale
artefact of an older render.

`pageName` is already local at the call site — passed to
`newSourceResolver(siteID, params.DB, logger, pageName)` on the line above — so the
identity was available and simply never reached the render base. Fix: pass it and
set `base["current_page"] = strings.TrimSuffix(pageName, ".html")`. Regression added
to `rerender_page_sections_base_data_test.go` and proven to detect its own defect.
**Council round 3: APPROVED** (2 advisory, none high-severity) on the same
correlation `b64141e5-b95c-418d-a20d-e917f050ed75`; committed with the trailer.
**Needs the NEXT roll.** All three rounds are on that one correlation, so the trail
reads in order: REVISE → APPROVED (build path, shipped v1.0.1173) → APPROVED
(scoped path, inert).

### The complete survey (do this instead of trusting a caller count)

| path | set `CurrentPage`? |
|---|---|
| `multipage_actions.go:206` | yes, always did |
| `rerender_pages_actions.go:190` | yes, always did |
| `section_editor_actions.go:489` | yes, always did |
| `BuildRenderContextAction` | **no** → fixed, LIVE v1.0.1173 |
| `buildRerenderBaseData` | **no** → fixed, inert |

Two of five, no sixth. *"`build_render_context` has exactly one caller fleet-wide"*
was true and verified twice — and is not the same question as *"what else builds a
`RenderContext` without calling it"*. That is the question to ask next time.

### Owed at the next roll (supersedes the checklist above)

1. Pod-grep the new warning string with a negative control.
2. Scoped re-render `index` → assert **one** chart (`relojistas-feed-restoration`),
   not three. This query is already known to FAIL before the fix, so it
   discriminates:
   `SELECT (SELECT string_agg(DISTINCT m[1],',') FROM regexp_matches(pc.rendered_html,'data-chart="([a-z-]+)"','g') m) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id WHERE s.domain='fundamentallyai.com' AND cc.function='evidence-chart';`
3. Restore the `capabilities` placement (plan level + page_components), re-render,
   assert it carries the **two** charts declared for it.
4. Induce the empty case — a page carrying the section whose charts all name other
   pages must render **nothing**, not everything.

---

## 2026-07-27 (post-roll triage sweep) — the scoped-path fix is LIVE; steps 2–4 deliberately NOT run

**The roll happened.** `v1.0.1174` at **15:11:15Z**, binary dated **14:58 UTC**; the last
Go commit before it is `e96d42226` at 14:52:33 UTC, so `c447d34a6` and `32a55597e` are
both in the image. (UTC throughout: this machine is BST, and comparing BST `git log`
against UTC `kubectl` makes live fixes look un-shipped.)

**Step 1 — PASS.** `agent-chassis-5994dc6d6c-pt8v9`: `currentPageName` resolves **2**,
`resolveCurrentPageName` resolves **1**, invented control **0**.

**Steps 2, 3 and 4 — NOT RUN, on purpose, and this is a coordination call not a
verification one.** The `brochure_component_library` workstream **owns this bug and that
site, and was actively working both while this sweep ran** — untracked
`sql/085b`–`085f` and `scripts/deploy_stylesheet_direct.sh` were written minutes
earlier, and `085f_queue_rerenders_after_contrast_and_imagery_fix.sql` had already
queued the work. Firing a competing scoped re-render at `fundamentallyai.com/index`
would have raced their palette/imagery re-render through the same section rows.
(Their `085b`–`085f` numbering is their own SQL seed sequence and is unrelated to this
bug number — easy to misread.)

**Live state at 15:58Z, as the baseline they should diff against.** The stored render is
unchanged from the 14:08 measurement in the section above:

```
page  index   updated_at 2026-07-27 14:08:17.952+00
charts: council-review-outcomes, news-pipeline-credibility, relojistas-feed-restoration
```

Still all three, two of which declare `pages: ["capabilities"]` — i.e. **the defect is
still visible on the live site**, because nothing has re-rendered that section on the new
binary yet. That is expected, not a failed fix.

**And their re-renders were queued but not dispatching, for an unrelated reason.** Six
`page_rerender` items sat `triaged` and unclaimed on `fundamentallyai.com`
(`site_id 199733a8-…`) while `build-pipeline-trigger` lost nine consecutive
`spawn_dispatch` requests between 15:24 and 15:44 — `bugs_open/029`'s post-roll degraded
window, not anything to do with this bug. Dispatch recovered at 15:58:27 (first claim by
`build-dispatch-loop`). **So the natural way to close this case is to let their queued
re-render land and then run the step-2 query** — no separate induction needed, and no
race.

> **Next action (owner: brochure_component_library):** after the queued re-render of
> `index` completes on v1.0.1174, run the step-2 query and assert **one** chart
> (`relojistas-feed-restoration`), not three. Then steps 3 and 4. The query is already
> known to fail pre-fix, so it discriminates.
