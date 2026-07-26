# NOTES — bugs_open/045 hero-tool component (append-only, newest at bottom)

## 2026-07-21 — investigation

- Read the handoff (bugs_open/045) and the parent (023) + robot-hands + cta
  memories. 045 was split out of 023 as class F; 023 is owned by the
  cta_link_integrity workstream — I did **not** touch 023's scope.
- Traced the resolution path in code:
  - `plan_sections_action.go` Path 1 (`components[sectionName]`) keys by
    name/function — a `hero-tool` section MISSES it (Bayesian function is
    `bayesian-ranking-hero-tool`). So it falls to Path 2, the selector.
  - `resolveSectionComponent` → `SelectComponentByType` → `queryCandidates`
    (`component_selector.go:164`): matches on **`section_type = $1`**,
    `component_level='section'`, `is_active`, `forked_from IS NULL`; scores; takes
    `candidates[0]`. **No minimum-score cutoff** — a sole candidate always wins.
    This is THE fact that makes the fix deterministic once the Bayesian is retired.
- Live DB confirmed the handoff's root cause:
  - Only row with `section_type='hero-tool'` = `bayesian-ranking-hero-tool_pre_037`
    (id `7cd0408b...`), 14 `source:static` Bayesian fields, `suitable_site_types=[]`,
    `suitable_page_types=["bayesian-ranking"]`, `is_dark_section=t`.

### Correction to the handoff's blast-radius table (caught by re-measuring)
> **CORRECTED 2026-07-21:** the handoff named **two** armed pages and said the
> Bayesian row has **"0 page_components placements."** Live measurement found
> **THREE** pages requesting `hero-tool`, and **one placement**:
> `gamesdesign.co.uk/bayesian-ranking` is **deployed** with the Bayesian hero
> (`page_components` slot `bayesian-ranking-hero-tool`, 9405 B, position 1).
> The handoff's "0 placements / four removed" counted only the 023-scope pages
> (finetuning + orchestration + leopardess); gamesdesign was outside 023 and
> never cleaned. What caught it: querying `pages.sections LIKE '%hero-tool%'`
> across the whole fleet instead of trusting the two named pages.
> **This mattered:** gamesdesign is the ONE page where a Bayesian hero is
> *correct*, so the retire step had to be a supersede (keep active), never a
> delete, and I verified its ranking function survives (separate
> `tool-bayesian-ranking` section, position 3).

### Design settled
- Generic `hero-tool`: llm labels, gated renderer CTAs, optional gated
  anti-fabrication trust stats, single-column intro, no embedded widget.
- Retire Bayesian: `section_type` → `bayesian-ranking-hero-tool`, stay active.
- Both atomic in migration 183. Council gate does NOT apply (SQL/docs are refused
  client-side; scope is platform/internal/pkg Go). No Go change → no image roll.

### Applied
- `183_generic_hero_tool_component.sql` applied clean; NOTICE printed new id
  `0bf81196-e4e7-430b-bd5d-1585703678ae`; all 5 post-conditions passed.
- Post-apply: selector simulation for `section_type='hero-tool'` returns exactly
  ONE row (the generic, score 0.69); new template greps **0** Bayesian strings.

### Next
- Rebuild-based verification: a rebuild is what arms this bug, so a clean rebuild
  is the proof. The two armed pages are already `needs_rebuild`.

### Verification — the rerender trap (a dead end I avoided by reading the action)
- `ai-agent-orchestration.com/agent-complexity-estimator` already had TWO fresh
  `triaged` `page_rerender` items (2026-07-21 10:34). Tempting to just let those
  run as the proof. **They would prove nothing about this fix.**
- `rerender_single_page_action.go` header: "assembles a page from **stored /
  pre-rendered components**." It re-renders existing `page_components`; it does
  NOT re-run `plan_sections` or re-select. The armed pages have **no** hero-tool
  placement (023 removed it), so a rerender can never create one — component
  selection is simply not on the rerender path.
- Only the full **site-build** path re-selects: `get_pages_to_build_actions.go`
  (per-site, statuses `planned`+`needs_rebuild`) → `plan_sections` →
  `resolveSectionComponent` → `SelectComponentByType` → `queryCandidates`. That
  last query is what I mirrored in SQL; it returns ONLY the generic component
  (score 0.69).
- **Decision (2026-07-21):** did NOT trigger a full site build to verify. It is
  per-site (would rebuild all 38-of-fleet / this site's pending pages), costs real
  credits, and risks colliding with other sessions' active work on finetuning.uk /
  ai-agent-orchestration.com (both have live voice_tells + CTA items). The fix is
  proven deterministically (verbatim selector query → sole candidate; template
  greps 0 Bayesian strings; migration post-conditions green) and the live build
  path runs that identical query. The artefact-level proof lands naturally when
  the platform drains these `needs_rebuild` pages; RUNBOOK documents the exact
  confirmation query + live-page curl. 045 stays OPEN until that lands.
- Pod age 176m (safe re the ~300s dispatch-drop caveat); build-dispatch-loop +
  page-rerender confirmed alive (COMPLETED rows within the last ~15 min) — so the
  drain WILL happen, it is a scheduling/backlog question, not a broken loop.

### Pattern recorded
- Added the transferable pattern to 016b §9 ("A generic section name resolves to a
  product-specific component") including the rerender-does-not-re-select landmine.

## 2026-07-26 — the proof landed on its own; case CLOSED

### The 07-21 deferral was right, and the platform produced the evidence

On 07-21 this lane declined to force a full site build (per-site, real credits,
collision risk with sessions live on both named sites) and predicted the artefact
proof would arrive when the platform drained its own `needs_rebuild` queue. **It
arrived on 2026-07-25 02:08 UTC** — but from a page this file never named.

`fundamentallyai.com/llm-cost-calculator` was built through the real path. All four
of its `page_components` rows carry `created_at == updated_at` at that instant,
which is the discriminator that matters: those rows were **written** by that build,
so `plan_sections` → `SelectComponentByType` genuinely re-ran. `hero-tool` resolved
to `0bf81196-…`, the generic component. Live: 200, 70,162 B, **0** Bayesian strings,
headline *"Compare LLM provider costs before you commit"*.

Two design intentions were confirmed by what the hero did **not** emit. Styles
stripped, the whole section is 615 bytes: **zero anchors** — CTAs are gated
`{{if .x_url}}` and no url was supplied, so the degraded state is a missing button
rather than a dead one — and **zero trust stats**, the optional anti-fabrication
gate declining to invent figures. Both are easy to claim from a template and only
observable in a real render.

Fleet sweep on `rendered_html ~* '(Start Ranking Free|Calculate Rankings|Try the
Bayesian Ranker)'` returns **one row, fleet-wide**: `gamesdesign.co.uk/bayesian-ranking`,
where it is correct. The supersede-don't-delete call from 07-21 holds up.

### MISSTEP — the runbook's definitive live test was a false green, and I nearly closed on it

The RUNBOOK's final check (written by this lane on 07-21) was:

```bash
curl -s https://finetuning.uk/tools/ai-agent-roi-estimator/ | grep -ciE '…Bayesian'  # expect 0
```

**That URL 404s.** The fleet serves `/tools/<name>.html`; the trailing-slash form
returns a 304-byte B2 error JSON, which contains no Bayesian strings, so the command
prints `0` and the check *passes against a page that does not exist*. It would have
printed `0` before the fix, after it, and against a misspelled hostname — nothing
about it was ever conditional on the thing it claimed to test.

What caught it: curling with `-w '%{http_code}'` out of habit while collecting
closure evidence, and noticing a 404 sitting beside a "clean" result. The runbook
does not tell you to do that. This was luck, not method.

**The class, which is the transferable part:** a negative assertion is satisfied by
*nothing existing*, so every way of breaking the check itself yields "pass". Same
shape as `ON CONFLICT DO NOTHING` returning `err == nil` having inserted nothing.
The fix is a positive control over the same fetch — `data-component="hero-tool"`
must be present — which returns 1 on the real page and 0 on the 404, verified both
directions before writing it into the runbook. Recorded in 016b §9 and
`WRONG_CALLS.md`; the runbook command is corrected in place with the old one shown
as a marked correction rather than quietly deleted.

### What is deliberately NOT claimed in the closure

- The two pages this case named have **still not rebuilt**. Both are clean live
  today and can now only select the generic component, but closure rests on one real
  rebuild plus a deterministic selector, not on all three pages. Said so in the file.
- The proof page was a **fresh** tool page, never Bayesian-damaged — so "a damaged
  page rebuilds clean" was not exercised. There is no mechanism behind that
  distinction (the Bayesian row is no longer a `hero-tool` candidate at all), but it
  was not measured, so it is not asserted.

### Residuals handed on, not dropped

- **Candidate 4** (build-time selection-sanity check) was to ride with `bugs_open/039`;
  039 has since closed, leaving it homeless. Contributed into
  `features_open/017` (component-adoption check) — 017 is already the mechanical,
  no-LLM health report over `content_components`, and candidate 4 is its inverse
  (a dormant component is *never* selected; this is a section name whose *only*
  candidate is product-specific). Contributed in rather than forked.
- **Two stale review-queue items** (`11dd56f1…`, `ba28ba8d…`) still name
  `bayesian-ranking-hero-tool` CTAs on leopardess pages that no longer request a
  hero-tool and hold no such placement — an extinct defect still sitting in the
  human queue. That is `bugs_open/033` / the `review_queue_drain` lane's remit
  (`revalidate_review_queue` built, inert until a roll). Cited there as evidence,
  not fixed here — `scripts/who-owns.py` says 033 is active elsewhere.

### Where this closure's commits actually are — the trail is SPLIT, by accident

The case commit is `8feb5dd27` (`close(045): …`), seven files, verified clean against
the commit-scope block: the moved case file, the four standing-five docs,
`features_open/017` and the `bugs_open/033` contribution.

**But the two 016b entries and the WRONG_CALLS entry are NOT in it.** `016b` and
`WRONG_CALLS.md` are fleet-wide append-only ledgers that several sessions write at
once, and a pathspec commit takes the working tree **whole-file** — so there was no
way to commit my entries without also taking another thread's uncommitted `052` work
sitting in both files. I prepared a second commit that disclosed exactly that. Between
the check and the commit, **that thread committed first** (`fe00304bd`,
`fix(bugs_open/052): blog listing derivation carries the build-state floor`) and swept
my three entries in with theirs — the mirror of the case CLAUDE.md warns about, from
the receiving end. Nothing was lost and forward-only holds, so it was left alone.

**Consequence worth knowing:** `git log` for the §9 *"a negative assertion over an
unguarded fetch is vacuous"* entry, the §10 `045` CLOSED row, and the WRONG_CALLS row
all resolve to a commit titled about **052**. Anyone bisecting or running the `098`
coverage join on this closure will find the doc trail under two messages, one of which
does not mention 045. Hence this paragraph.

**The transferable bit:** for the two shared ledgers specifically, the window between
"I checked the file" and "I committed the file" is where this happens, and it is not
closable — the fix is not tighter timing but *writing down where your entry landed*,
because the commit message will not say. Do not attempt to re-add a swept entry: it is
already in HEAD, and a second copy is the real damage.
