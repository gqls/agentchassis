# 045 — the library's only tool-hero component is hard-wired to a Bayesian ranker, so every tool page that asks for one gets the wrong product's vocabulary

**Filed:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops` · **Status:** OPEN, not started
**Severity:** medium-high — **armed on 2 live pages right now** (both `needs_rebuild`); latent
across 37 tool pages / 10 sites
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
