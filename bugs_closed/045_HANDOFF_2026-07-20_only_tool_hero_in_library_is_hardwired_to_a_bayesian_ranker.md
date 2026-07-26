# 045 — the library's only tool-hero component is hard-wired to a Bayesian ranker, so every tool page that asks for one gets the wrong product's vocabulary

**Filed:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops`
**Status:** **CLOSED 2026-07-26 — fixed, live, and proven at the artefact level by a real
rebuild.** Fix applied 2026-07-21 (migration `183_generic_hero_tool_component.sql`, DB config,
live immediately, no image roll); the rebuild proof it was held open for landed unprompted on
2026-07-25. Closure evidence below.
**Severity when open:** medium-high — armed on 2 live pages; latent across 37 tool pages / 10 sites

---

## CLOSURE — 2026-07-26 (workstream `hero_tool_component_045`)

The case was held open for one thing: **an artefact-level rebuild proof.** A rebuild is what
arms this bug, and `page_rerender` does not re-select components, so only a full build path
(`get_pages_to_build` → `plan_sections` → `SelectComponentByType`) could prove the fix. On
2026-07-21 this lane deliberately did **not** force one (per-site, costs credits, would collide
with other sessions live on both named sites) and recorded that the proof would land naturally
when the platform drained its own `needs_rebuild` queue. **It did.**

### The proof — `fundamentallyai.com/llm-cost-calculator`, built 2026-07-25 02:08 UTC

Not a page this file named, and not a repair: a **fresh tool page** built through the real path,
which is exactly the latent-fleet case (37 tool pages). Selection genuinely re-ran — all four
`page_components` rows carry `created_at == updated_at` at that instant, the discriminator that
separates a real build from a rerender:

```sql
SELECT pc.slot_name, pc.component_id::text, pc.created_at = pc.updated_at AS fresh_select,
       (pc.rendered_html ~* '(Bayesian|Ranking Free|Calculate Rankings)') AS has_bayes
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='fundamentallyai.com' AND p.name='llm-cost-calculator';
--  hero-tool | 0bf81196-e4e7-430b-bd5d-1585703678ae | t | f     ← the generic component
```

The live artefact (200, 70,162 B) greps **0** Bayesian strings and carries a headline about the
page's own subject. The whole rendered hero, styles stripped, is 615 bytes:

```html
<section class="hero-tool-section" data-component="hero-tool"><div class="htl-container">
<span class="htl-badge">Free calculator</span>
<h1 class="htl-headline">Compare LLM provider costs before you commit</h1>
<p class="htl-subheadline">Token pricing varies significantly across providers…</p>
</div></section>
```

**Two design guarantees held, visible in that markup:** it emits **zero anchors** — the CTAs are
gated `{{if .x_url}}` and no url was supplied, so the failure mode is a missing button, never a
dead one (LNK-005 by construction, `023` class H) — and **zero trust stats**, the optional
anti-fabrication gate (`043` class) declining to invent figures.

### The supersede was safe — fleet-wide, one placement, on the one page where it is right

```sql
SELECT s.domain, p.name, pc.slot_name FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ~* '(Start Ranking Free|Calculate Rankings|Try the Bayesian Ranker)';
--  gamesdesign.co.uk | bayesian-ranking | bayesian-ranking-hero-tool     ← the only row, fleet-wide
```

`gamesdesign.co.uk/bayesian-ranking` is still deployed with its Bayesian hero at position 1 and
still serves that vocabulary live — correct there, and the reason retiring had to be a supersede
rather than a delete. **No live page anywhere renders the Bayesian vocabulary from any other slot.**

**One known, accepted consequence — re-verified 2026-07-26.** gamesdesign keeps that hero only
while it is not rebuilt. Its `pages.sections` asks for the **generic** `"hero-tool"` (the
`bayesian-ranking-hero-tool` you see in `page_components.slot_name` is the *component's function*,
not the requested section name — 023 R2), so on any rebuild it will now resolve to the generic
component like every other tool page. That is the trade this lane took on 2026-07-21 and it still
looks right: the page's actual ranking calculator is a **separate** `tool-bayesian-ranking` section
at position 3 and is untouched, so the page loses a product-specific banner and gains one written
for its own subject. Recorded because the page is `deployed` today and its appearance *will* change
when it next builds — that is expected, not a regression.

### What this closure does NOT prove — stated so nobody infers more than was measured

- The two pages this file named (`finetuning.uk/ai-agent-roi-estimator`,
  `ai-agent-orchestration.com/agent-complexity-estimator`) **have not rebuilt.** Both are still
  `needs_rebuild`, both are clean live today (200, 0 Bayesian strings), and both can now only
  select the generic component. Closure rests on one real rebuild plus a deterministic selector
  (sole candidate, no score threshold), not on all three pages.
- The proof page was never Bayesian-damaged, so "a previously-damaged page rebuilds clean" was
  not exercised. There is no mechanism behind that distinction — the Bayesian row is no longer a
  candidate for `hero-tool` at all — but it was not measured, so it is not claimed.

**If you see either named page rebuild**, run the query above against it and append the result
here; that is a free strengthening of this closure, not a reopening condition.

### Residuals — owned elsewhere, do not start a competing fix

- **Fix candidate 4** (build-time selection-sanity check) was never built. Its intended sibling
  `039` has since closed, so it had no home; handed to **`features_open/017`**
  (component-adoption check) on 2026-07-26 as a dated contribution — 017 is already the
  mechanical no-LLM health report over `content_components`, and candidate 4 is its inverse.
- **Two stale review-queue items** (`11dd56f1…`, `ba28ba8d…`, both `needs_human_review` since
  2026-07-17) still name `"Start Ranking Free" … (bayesian-ranking-hero-tool)` on
  `leopardessconsulting.co.uk` pages whose `sections` no longer request a hero-tool and which
  hold no such placement. **They describe an extinct defect.** They belong to `bugs_open/033` /
  the `review_queue_drain` lane, whose `revalidate_review_queue` is built and inert until an
  image roll — cite them there as evidence for it, not here.

---

## FIX LOG — 2026-07-21 (workstream `hero_tool_component_045`)

**Done (candidates 1 + 2, atomically in migration 183):**
1. **Built a generic `hero-tool` component** (id `0bf81196-e4e7-430b-bd5d-1585703678ae`,
   `function=section_type=hero-tool`, `component_level=section`). Every visible label is
   `source:llm` — **zero `source:static` product fallbacks**. CTA anchors **gated**
   `{{if .x_url}}` with `*_url` fields `source:renderer` (LNK-005 by construction, 023 class
   H). Trust stats **optional + gated** with anti-fabrication guidance (bugs_open/043). No
   embedded widget — the tool is a separate `tool-<slug>` section. Same shape as
   `tool-guide-intro` after migration 179.
2. **Retired the Bayesian row from the pool, did NOT delete it** — `section_type`
   `hero-tool` → `bayesian-ranking-hero-tool`, kept `is_active=true`, function unchanged.
   The generic component is now the **sole** `hero-tool` selector candidate (score 0.69);
   `SelectComponentByType` has no score threshold, so selection is deterministic.

**Verified (DB/template level):** selector simulation returns exactly one `hero-tool`
candidate (the generic); new template greps **0** Bayesian strings; all 5 migration
post-conditions green. Touches **no deployed page** (page_components bake `component_id` +
`rendered_html`).

> **CORRECTION to this file's own blast-radius section (2026-07-21):** there are **THREE**
> pages requesting `hero-tool`, not two, and the Bayesian row is **NOT** placement-free —
> `gamesdesign.co.uk/bayesian-ranking` is **deployed** with the Bayesian hero (position 1),
> and it is the one page where that is *correct*. The "0 placements / four removed" figure
> counted only the 023-scope pages. This is exactly why the retire step is a supersede, not
> a delete; gamesdesign's ranking function survives (separate `tool-bayesian-ranking` section,
> position 3). Caught by re-querying `pages.sections LIKE '%hero-tool%'` fleet-wide.

**Not done (deliberately):** candidate 4 (build-time selection-sanity check) — it is Go
(needs an image roll) and is the shared branch with `bugs_open/039`; left for a dedicated
change. **Remaining to close:** rebuild one armed page and confirm the hero renders generic
(RUNBOOK in `docs024_key_docs_latest/hero_tool_component_045/`).
**Class:** library gap (a missing component), NOT a planner defect — that distinction is the
whole point of this file

> **Split out of `bugs_open/023` on 2026-07-20** (owner instruction). It was class **F**
> there — "static fallbacks carry another tool's vocabulary". 023's headline symptom (four
> broken buttons) is fixed fleet-wide and its remaining classes are about *label/URL
> pairing*; this one is about *component selection* and has a different fix, a different
> blast radius and a different owner. It was getting buried in a bug that reads as done.

---

## Symptom

A page plan asks for a generic `hero-tool` section — a perfectly sensible request. It
resolves to a hero panel for **a Bayesian ranking tool**, whatever the page is actually
about. On leopardess this put *"Start Ranking Free"*, *"Calculate Rankings"* and *"Try the
Bayesian Ranker"* on an **LLM cost calculator** and an **ROI estimator**; the same component
did it again on finetuning.uk. Two of the four buttons the owner reported on 2026-07-19
came from this.

## Root cause — measured, not inferred

**The active component library contains exactly one component that can serve a `hero-tool`
section request, and it is hard-wired to a different product.**

```sql
SELECT name, function, is_active FROM content_components
WHERE is_active AND (function LIKE '%hero%tool%' OR name LIKE '%hero%tool%');
--  bayesian-ranking-hero-tool_pre_037 | bayesian-ranking-hero-tool | t     ← the ONLY row
```

Its vocabulary is frozen onto any page that adopts it: **14 `source:static` fields with
Bayesian-specific fallbacks**, and a `static` fallback re-applies on every render while
bypassing `required`/`on_missing` entirely (`plan_sections_action.go:1210-1218`), so
`content_data` **cannot override them**:

```
badge_label          Free Ranking Tool        rank_btn_label       Calculate Rankings
cta_primary_label    Start Ranking Free       result_label         Bayesian Rankings
cta_secondary_label  See How It Works         tool_panel_title     Try the Bayesian Ranker
input_item_label     Item name & total votes  input_prior_label    Prior belief (0–1)
input_positives_label Positive votes          add_item_btn_label   Add Item
+ 4 placeholder fields (e.g. Product A, Positives, e.g. 0.5, Total votes)
```

So the selector is behaving correctly and the planner is behaving correctly. **The library
is missing a neutral tool hero**, and selection resolves the only thing that matches.

> **This is why it matters that it is NOT a planner bug.** The obvious reading —
> "the planner proposed a Bayesian ranker for an LLM cost page" — is wrong and would send
> the next thread to diagnose the planner. `pages.sections` asked for the string
> `"hero-tool"`. Nothing about Bayes was ever planned. Corrected in `cta_link_integrity/
> NOTES` on 2026-07-20 after checking what the section name actually resolved to.

## Blast radius (measured 2026-07-20 19:20)

**Armed right now — two live pages request `hero-tool` and are flagged for rebuild:**

| site | page | build_status | placed today | live now |
|---|---|---|---|---|
| finetuning.uk | `ai-agent-roi-estimator` (`/tools/…`) | **`needs_rebuild`** | `tool-ai-agent-roi-estimator` only | 200, 34,726 B, 0 Bayesian strings |
| ai-agent-orchestration.com | `agent-complexity-estimator` (`/tools/…`) | **`needs_rebuild`** | `tool-agent-complexity-estimator` only | 200, 34,848 B, 0 Bayesian strings |

Both `pages.sections` read `["hero-tool", "tool-guide-intro", "<the real tool>", "tool-cta"]`.
The Bayesian sections are currently **absent from the page** but **still requested by the
plan**, so the next rebuild re-adopts them. They are clean today only because the sections
were never (re)built after the 023 cleanup — not because anything prevents it.

- **Latent fleet exposure:** 37 active `page_type='tool'` pages across 10 sites; any new or
  re-planned tool page asking for a tool hero hits this.
- **Aspect layer is clear:** zero `site_plan_sections` rows name a hero-tool, so the two
  above are `pages.sections`-only (see 023's RUNBOOK R12 — authority differs per site).
- `bayesian-ranking-hero-tool_pre_037` currently has **0 `page_components` placements**
  (the 023 cleanup removed all four).

## Fix candidates

1. **Build a generic `hero-tool` component** (the actual fix). No product-specific `static`
   fallbacks — labels `llm`-sourced or genuinely generic; **CTA anchors gated**
   (`{{if .x_url}}`) and url fields `source:renderer`, so it satisfies LNK-005 by
   construction and derives cleanly under the schema-derived CTA pairing (023 class H).
   Cheaper than it looks and fixes every future tool page at once.
   **Precedent to copy:** `tool-guide-intro` after migration 179 (023) — same shape of fix.
2. **Then retire the Bayesian row properly.** ⚠️ **Do NOT delete it** — it is the sole
   active row for its function; deleting it deletes the live component (023 RUNBOOK R10, and
   it is one of 16 active `_pre_037` rows in the same state). Rename/supersede deliberately,
   and only once a `hero-tool` component exists to take the selection.
3. **Disarm the two named pages** in the meantime — either rebuild them after (1), or strip
   `hero-tool`/`tool-guide-intro` from their `pages.sections` as was done for the four
   already-broken pages (023 NOTES 2026-07-20; SQL pattern in that session's scratchpad).
4. Optional, wider: a **selection-sanity check** — a section resolving to a component whose
   static vocabulary names a different product is exactly the kind of thing a build-time
   check could flag. Cheap version: warn when the *only* candidate for a generic section
   name is a product-specific component.

## How to verify a fix

- A `hero-tool` section on a tool page renders **no** Bayesian string:
  `grep -cE 'Start Ranking Free|Calculate Rankings|Bayesian|Free Ranking Tool'` → 0 on the
  rendered artefact (not `page_components.rendered_html` alone — check the live page).
- The two `needs_rebuild` pages above rebuild **and stay** correct — a rebuild is the test,
  since the defect only fires on rebuild.
- `SELECT … WHERE is_active AND function LIKE '%hero%tool%'` returns a **generic** component,
  and the Bayesian row is superseded rather than deleted.

## Landmines

- **Do not diagnose the planner.** It asked for `hero-tool` and that was right. See above.
- **Do not delete `_pre_037` rows.** 16 of them are the sole active row for their function
  — deletion removes the live component (023 R10).
- **A rebuild is what arms this**, so "the page looks fine" is not evidence the bug is
  absent. Both pages above look fine today and are still armed.
- `slot_name` resolves via `content_components.function`, not `.name` (023 R2).
- **The "expect 0 Bayesian strings" live check passes on a 404** (found 2026-07-26 while
  closing this). This file's own verification URLs are `/tools/<name>.html`, **not**
  `/tools/<name>/` — the trailing-slash form returns a 304-byte B2 error JSON, which of
  course contains no Bayesian strings, so `curl | grep -c` reports `0` and the check
  "passes" against a page that does not exist. Any negative assertion over an unguarded
  fetch is vacuous by construction: assert `-w '%{http_code}'` first, or assert a positive
  marker the page must contain. See 016b §9 and `WRONG_CALLS.md`.

## Related

- **`bugs_open/023`** — parent; classes A/B/C/E still open there, G moved to 033, H is the
  council trail. 023's migration 179 is the template for fix candidate 1.
- **`bugs_open/039`** — the *sibling branch of the same selector*: a section name resolving
  to **no** component renders a hollow stub. This file is a section name resolving to the
  **wrong** component. Same mechanism, opposite failure, different fix — worth fixing
  together if anyone touches selection.
- `bugs_open/033` — human-review queue has no working surface (where 023's class G went).
- `bugs_open/001` — re-plan clobbers built pages; constrains `page_components`-level
  workarounds (fix candidate 3).

---

## Contribution from the 040-partial-build workstream (2026-07-24) — your two rebuild candidates are queued, and a third page family surfaced

While decomposing 040's section-drop causes: the two 1-of-4 tool pages
(`finetuning.uk/ai-agent-roi-estimator`, `ai-agent-orchestration.com/agent-complexity-estimator`)
each hold ONLY the tool widget row (`tool-<name>`, `build_status='pending'`); the planned
`hero-tool` / `tool-guide-intro` / `tool-cta` trio has no `page_components` row at all. When a
deploy path tried to stamp them `deployed`, 040's shortfall guard (live v1.0.1146+) refused and
flipped both to `needs_rebuild` with the plan stamp cleared — so **your rebuild-arms-the-fix test
is already queued by the platform itself**; when those rebuilds fire, the section trio should now
resolve (generic `hero-tool` exists post-2ba9d3d50; `tool-cta`'s `query.pages_where_type:tool`
has data; `tool-guide-intro` is llm-sourced). If a rebuilt page comes back still 1-of-4, that is
evidence of a second defect in the tool-page build path (does it run `plan_sections` at all?),
not of this one. No fix forked here — observation only. Full context:
`bugs_open/040…partial_build…md` CORRECTED 2026-07-24 block; diagnosis of the
skip-not-recorded interaction is corr `65103331`.
