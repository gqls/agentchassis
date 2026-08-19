# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-19 09:05Z. Supersedes `HANDOFF_2026-08-17_continue_here.md`.

**The lane is at a decision point: 8 of 63 tools replaced, 55 remain. It is NOT blocked — the recipe
is routine — but finishing is 55 more repetitions.** See "Where this stands" below; the owner has been
asked whether to continue, stop, or scope it down.

Read: this file → `PLAN_2026-08-15_…` (design + three owner rulings + two corrections) →
`RUNBOOK_…` (every command) → `NOTES_…` (evidence, newest at bottom) →
`SUMMARY_2026-08-18_…` → `architecture_review/RFC_036` + `bugs_open/303` + `bugs_open/315`.

## The recipe — PROVEN, 8 tools, and now routine

1. **Read the LIVE tool's `<script>`** (fetch with `?cb=<epoch>` so you do not warm the edge with the
   page you are about to replace) and write the spec from its behaviour. Describe *intent* where the
   ported version is defective; do not add features.
2. **File with all THREE gates as pre-asserts inside the transaction**, plus the serial throttle:
   - `idx_cc_tool_function_unique` — fleet-wide on `function`, `WHERE component_level='tool' AND forked_from IS NULL AND is_active`
   - `content_components_name_key` — **`UNIQUE(name)`, NO predicate**; the generator derives
     `name = '<function>-<domain-slug>'`, so a REBUILD needs the old row **renamed** (never deleted)
   - the generator's `already_exists` probe — per-site, joins `page_components`, no `build_status` filter
   - always include the `bugs_open/303` build-constraint sentence in the description
3. **Grade the RUN** — `current_step='complete'`, `page_adopted='true'`, no `already_exists`, no
   `__step_error`. A failed build reports the ITEM as `complete` with `error` NULL.
4. **Grade the COMPONENT by locating the MECHANISM, not by grepping wording you imagined.** Enumerate
   its element ids and check the machinery each requirement needs. `{{\.` must be 0. Read the JS.
5. **Retire IMMEDIATELY** — record `page_components` id + length + md5 BEFORE filing; guarded UPDATE;
   md5 after must equal md5 before; assert one deployed slot remains **and that it is the new one**.
6. **Grade at the served page, cache-busted**, `http=200` first, with a negative and positive control,
   and include a NEGATIVE that only the old version could satisfy (an old element id) — that is what
   rules out a stale render.

## Where this stands `[MEASURED 2026-08-19 09:00Z]`

| | |
|---|---|
| chassis | `v1.0.1314`, digest `d0257576…` — real roll (verify by DIGEST, not a binary probe) |
| **replaced, live, graded** | **8** — aspect-ratio, markdown-tables, html-minifier, svg-optimizer, sri-generator, smooth-shadow, json-cleaner, seo-injector |
| owner-approved on sight | aspect-ratio, markdown-tables, html-minifier, svg-optimizer |
| **remaining** | **55** — 2 blocked (RFC_036), 13 external-script, 18 simple <8 KB, 22 larger ≥8 KB |

**Five of the nine ported tools examined were measurably broken in production** — two with a checkbox
whose implementation sat inside its own comment, one corrupting `pre`/`script` content, one silently
disabling truncation on a bad input, one destroying the user's output on a parse error. **None was
visible from the page.** Reading the live script before writing each brief is what found all of them,
and it costs ~2 minutes per tool. That has become the lane's main product.

## Three platform defects this lane filed (all still OPEN, none this lane's to fix)

- **`architecture_review/RFC_036`** — three uniqueness gates on one INSERT. One (`UNIQUE(name)`) means
  a native tool **can never be rebuilt** without renaming the old row. Owner direction recorded: keep
  the library-and-fork model ⇒ a rebuild should record `forked_from`. **Nobody has built it**, which is
  why 2 tools are parked. **Do NOT unblock by deactivating their library templates** — both have live
  forks on other sites.
- **`bugs_open/303`** — the tool-birth guard counts tag SUBSTRINGS, so any tool that mentions a
  structural tag is unbornable, refused with a truncation message its own `ends_cleanly:true`
  disproves. Worse second class: a tool whose OUTPUT is a script tag cannot be reworded out of it, and
  the workaround forces the generator to hide its own tag behind string concatenation — which makes a
  real truncation harder to detect.
- **`bugs_open/315`** — `pages.deployed_at` is stamped whether or not bytes are written (measured stale
  on three pages, including two that were serving correctly). One page was skipped by four completed
  rerenders and published itself ~6 h later.

## THE PATH TO 63/63 (owner ruling 2026-08-19: framework ownership of ALL 63 tools)

**The audit question is settled and it is NOT extra work.** Reading the live script is step 1 of every
rebuild, so rebuilding all 55 audits all 55 by construction. A separate audit pass would duplicate it.
What IS worth having is the cheap prevalence sweep already run `[MEASURED 2026-08-19]`, as a
prioritiser and a brief-writing aid — across the 55 remaining ported tools:
`onclick=` **42** · `alert(` **25** · `parseInt`/`parseFloat` **14** · `innerHTML =` **18** ·
`localStorage` **3** · a copy button **26** · external `<script src>` **13**.
⚠ **These are PATTERN PREVALENCE, not confirmed defects.** Inline `onclick` is a code-quality smell,
not a bug; `alert()` is a real UX defect; the `parseInt` 14 are candidates for the NaN-guard class that
silently disabled truncation in `json-cleaner`; the 26 with a copy button are candidates for the
lying-copy class the owner reported. **Each still has to be read.** The value of the sweep is that it
says where to look first, and it sets an expectation: the defects found so far are not incidental.

### Phase A — the 18 simple, self-contained tools (<8 KB, no external script)
The proven path, unchanged. ~5–10 min of attention each; wall-clock is queue depth, not work.
Smallest first (the RUNBOOK's "Scope the batch correctly" query orders them). Serial — the item key
enforces it. **Expect roughly half to have a real defect**, on the run rate so far (5 of 9).

### Phase B — the 22 larger self-contained tools (≥8 KB)
Same recipe, longer briefs. **The rich hand-built apps live here** (mind-map, meme studio, logic
architect, micro-CMS, pasteboard) and by the owner's 2026-08-16 ruling they are reimplementations, not
preservations. **Owner's standing instruction: these go LAST and one at a time, each seen at the served
page.** For these the grade is a feature list checked in a browser, not a tag count — a raw-tag count
cannot tell you a mind-map lost its export.

### Phase C — the 13 external-`<script src>` tools
The page is not self-describing: the logic lives in S3 assets the DB-side checks cannot read (TL-032).
So the brief **must** come from the tool's behaviour in a browser, and **the external asset must be
retired with the slot** or the page keeps fetching a file nothing serves. Do these after Phase A has
made the recipe boring, and expect the spec work to dominate.

### Phase D — the 2 blocked tools (`tool-ab-test-calculator`, `tool-meme-generator`)
**Cannot be reached by any amount of lane effort.** They need `RFC_036 §9`: a ~10-line change in
`create_tool_component_action.go` to set `forked_from` when a library entry already claims the
function, then council + a chassis roll. **There is no config-only interim — proved in §9.1**: the
platform's own definition of a library tool (`forked_from IS NULL AND is_active`) is exactly the index
predicate, so "forkable by other sites" and "blocks a rebuild" are the same condition. Do not spend
another cycle looking for a way round it.

### Ordering, and why
A before C before B: Phase A keeps the recipe warm and cheap; Phase C's cost is spec-writing, not
mechanism; Phase B spends the owner's review attention, so it goes last when everything else is known
to work. Phase D runs whenever RFC_036 lands — it is not a sequencing dependency for A/B/C.

### What "done" means, stated now so it cannot drift
**63/63 replaced, each graded at the served bytes with a cache-buster and a negative control.** If
RFC_036 is never built, the honest terminal claim is **"61 of 63, 2 blocked on RFC_036"** — not "done".

## Next actions

1. **Phase A, next tool by size.**
2. If continuing: next by size from the 18 simple ones; run the six steps; ~5–10 min of attention each,
   wall-clock dominated by queue depth. **Do not file a rebuild you cannot attend** — the retire race
   has been as short as 2 minutes and was lost once at 96.
3. Rich apps last and one at a time (owner ruling 2026-08-16 put them in scope as reimplementations).

## Traps that each cost a real cycle (full detail in NOTES / WRONG_CALLS / LANDMINES)

- A written precondition saying "must be empty" is a STOP, not an input to a judgement.
- **Elapsed time comes from the ROW** (`now()-created_at`, `now()-claimed_at`), never the session clock
  or a `min(created_at)` column. I asserted a stall from a misread timestamp **twice**.
- A negative control alone cannot license a negative finding — something must MATCH in the same run.
- **Grade a requirement by its mechanism, not by a phrase you guessed** — two false negatives on
  correct fixes in one session.
- **Cache vs origin produce the identical symptom:** `?cb=` defeats a stale cache, not a stale origin;
  `last-modified` vs your rerender's `completed_at` tells you which. A PASS can be faked by neither.
- `handler_agent` is NOT NULL on `site_work_items`; for rerenders it is `page-rerender`.
- This lane's own rebuilds are its main queue competitor — each tool spawns a guide page and rerenders.
