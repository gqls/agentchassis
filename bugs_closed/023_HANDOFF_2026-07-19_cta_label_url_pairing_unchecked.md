# 023 — A button's label and its destination are never checked against each other

**Filed:** 2026-07-19 · **Branch:** `085_debug_and_feature_loops` ·
**Status: CLOSED 2026-07-25** — classes A/B/C/E fixed and LIVE; see *CLOSURE* at the foot of
this file. Descendants stay open under their own numbers (033, 039, 045, 049) and the flip
round continues in `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/`.
**Severity:** high — ships visibly broken controls to live customer-facing sites, fleet-wide
**Class:** structural (schema + template + work-item routing), not a single-site content bug

> **The plan and the full evidence live in
> `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/`.**
> - `PLAN_2026-07-19_cta_link_integrity.md` — defect classes A–H, phasing, staging advice
> - `NOTES_cta_link_integrity.md` — the diagnosis, with live evidence and my own corrections
> - `RUNBOOK_cta_link_integrity.md` — every query, with its gotcha attached
>
> This file is the pointer and the summary. **Read the plan before fixing anything** — the
> highest-leverage change (§3 of the plan) is not the obvious one.

---

## Symptom

The owner reviewed leopardessconsulting.co.uk on 2026-07-19 and reported four buttons:
*Start Ranking Free*, *See How It Works*, *Start the Guide*, *Visit the Tool* — "I don't
understand what these buttons are and what they do, I think they are broken."

They are. All four are on the two tool pages, in two components, and **each is broken by a
different mechanism that defeats a different check**:

| Button | Renders as | Why it's broken |
|---|---|---|
| Start Ranking Free | `href="/contact.html"` | Label belongs to a **different tool** (a Bayesian ranker). Frozen `source:static` fallback; content_data cannot override it. Destination is the residue of the fleet-wide "everything points at /contact.html" bug that migration 091 fixed for six components only. |
| See How It Works | `href=""` | `cta_secondary_url` resolves from an unset `site_specs` key → empty → **ungated template** emits an empty href. |
| Start the Guide | `href="#guide-start"` | **Hardcoded fragment in the template; no such id exists** on any page using the component. Dead on 4 pages / 3 sites. |
| Visit the Tool | `https://leopardess.contactforsales.com` (NXDOMAIN) and `https://leopardessconsulting.com/…` (owner's domain, but a 114-byte blank page) | Field is `source:llm, required:true` with no source of truth. **The schema requires a model to produce a URL it cannot look up, so it invents one** — two different hostnames on two adjacent pages. |

Owner-confirmed: he owns `contactforsales.com` and `leopardessconsulting.com` but never
created the `leopardess.` subdomain.

> **CORRECTED 2026-07-19** (owner challenged the first framing; he was right). I initially
> wrote that the model "assembled a hostname from two real domains in the owner's estate",
> implying estate knowledge. It has none. The real mechanisms are sharper and both are
> checkable:
> - `leopardessconsulting.com` is just the obvious `.com` variant of the site's own name.
>   The owner happening to own it **is** coincidence.
> - `leopardess.contactforsales.com` is a **transform of a real contact email**. The site's
>   identity spec holds `leopardess@contactforsales.com`; the model swapped `@` for `.`.
>
> The parts were true and in-context; only the recombination was invented. **That makes it
> deterministically detectable** — a hostname equal to a known contact address with `@`→`.`
> is fabricated by construction, no network call needed. **Exposure: 6 sites share
> `contactforsales.com` as their contact domain** (`agents@`, `finetuning@`, `gas@`,
> `idea.uk@`, `leopardess@`), each in its *current* identity spec, so any of them can produce
> the same fabrication. Filed as plan step P1.5.

## Root cause

**A component declares a button's label and its URL as two unrelated schema fields, and
nothing anywhere expresses "a label implies a destination" as a constraint.**

The label is typically `source:static`, which (`plan_sections_action.go:1210-1218`) writes
its fallback and `continue`s — **bypassing `required` and `on_missing` entirely** and
re-applying on every render. The URL may be absent from the schema, unresolvable, empty, or
LLM-authored. The template then renders the anchor ungated. So the button always has text
and sometimes has nowhere to go.

Compounding it, every relevant check has a hole exactly where these land:

- `href=""` → `validate_page_content.go:551-560` files it as a **warning**; `:257` only
  blocks on blockers and errors. `check_dead_controls.go` explicitly **cedes** `href=""` to
  the phantom check, so it falls between two stools.
- `href="#frag"` → `LinkScopeAnchor`, skipped by phantom, misdirected and validate alike.
  **Nothing anywhere resolves a fragment against the page's ids.**
- external → skipped by every consumer; **there are zero HTTP reachability checks in
  `platform/`**.
- content_data level → `check_required_fields_missing.go:189-192` skips any field whose
  `source != "llm"`, and CTA url fields are `renderer`/`site_specs` by design. They are
  **categorically exempt** from the only required-field check.

## The part that matters most

**One of the four *was* detected — correctly — and nothing delivered it.**

On **2026-07-17**, two days before the owner clicked it, the platform filed:

```
cta_names_unknown_destination | CTA "Start Ranking Free" on ai-agent-roi-estimator
                                (bayesian-ranking-hero-tool): lands in an excluded area (contact)
```

Right diagnosis, right component, right page. Filed at `status='needs_human_review'` —
which `TriageDetectedItemsAction` never promotes, which no `handler_agent` consumes, and
which `load_work_item_actions.go:804` **excludes from re-open queries**. A grep of the whole
`platform/` tree for `unresolved_cta`, `cta_names_unknown_destination` and `dead_control`
returns **only their emission sites. Zero consumers.**

Leopardess currently holds **21 `unresolved_cta` + 13 `cta_names_unknown_destination`**, all
inert, oldest 2026-07-13.

> This is not only a detection gap. It is a **delivery** gap, and it needs a different fix.
> Adding more checks without building the handler makes the invisible pile bigger.

## Blast radius (measured, not estimated — 2026-07-19)

- **51 dead or suspect controls across 7 of 11 sites**: 30 × `href=""`, 17 × bare `#`,
  4 × fragment. (`page_components` only — header/footer not yet counted.)
- **75 ungated CTA anchors across 38 active components**, vs 14 gated across 12. About
  **84% of URL-bound CTA anchors in the library violate the platform's own stated invariant
  LNK-005** ("an unresolvable destination renders nothing rather than a broken link").
- `tool-guide-intro` (the dead `#guide-start`) is live on leopardess ×2, finetuning.uk and
  robot-hands.com.
- **16 active `_pre_037` component rows** serve live traffic, each the sole row for its
  function — migration 037's replacements never landed. **Do not delete them; that deletes
  the live component.**

## Fix candidates

Full phasing in the plan. The headline:

1. **Derive CTA field pairs from `input_schema` instead of the hardcoded 6-entry
   `ctaFieldNames` map** (`resolve_internal_links_action.go:91-98`). Any `*_url` with a
   sibling `*_label` is a CTA pair. This kills the "detectable but not repairable" class
   permanently, makes the label/url pairing checkable at all, and retires a file that four
   separate migrations (091, 096, 097b, 098) have hand-patched with the same lesson. **This
   is the highest-leverage single change and it is not the obvious one.**
2. Gate every CTA anchor (`{{if .x_url}}`) — the `info-card-grid` precedent already exists
   *on this very site* and was never generalised.
3. ~~Build the **handler** for CTA findings.~~ → **MOVED TO `bugs_open/033`** (2026-07-20).
   The design guidance stands and is worth carrying over: where a real destination exists,
   repair; where none exists, **drop the button** — do not point it at `/contact.html`, that
   heuristic is what created *Start Ranking Free → /contact.html* in the first place.
4. Ban `source:llm` + `required:true` on any URL field as a schema-lint rule.
5. Post-hoc `external_link_unreachable` check (never at build time).
6. **Deterministic email→hostname check (P1.5)** — reject any external host equal to a known
   contact address with `@`→`.`. Cheap, no network, catches this exact class fleet-wide.
7. **Owner action, approved 2026-07-19: 301 `leopardessconsulting.com` →
   `leopardessconsulting.co.uk`**, path preserved (P4.1). Fixes one of the four buttons
   immediately and independently of all code work. ⚠️ It makes a *fabricated* URL resolve —
   do not mistake that for the defect being fixed; the field still invents on next build.

## How to verify a fix

> **REWRITTEN 2026-07-20 (bugfix-023 session).** The original third criterion — *"the 34
> inert work items must reach a terminal state via a handler"* — has been **removed, not
> satisfied**. It is not this bug's to meet: the write-only-queue defect is now owned by
> **`bugs_open/033`** (human-review queue has no working surface — 292 items fleet-wide,
> none ever actioned, 47 of them `cta_names_unknown_destination`), and 033 is blocked on an
> owner decision about intent, not on code. Leaving it here made 023 permanently
> un-closeable on its own scope: a bug cannot be gated on another bug's work. Class **G**
> stays *documented* here as the origin evidence — it was found from this investigation —
> but it is **tracked** in 033. Do not fix it here; do not close 023 waiting for it.

> **SUPERSEDED 2026-07-25 — read the CLOSURE section at the foot of this file for the final
> verdicts.** The table below is the 2026-07-20 state, kept because the *reasoning* in it
> (why criterion 1 is about surviving a rebuild, why criterion 3 is structural) is what the
> closure was measured against. Its states are stale: all three are now met.

Criteria, all re-measured 2026-07-20 19:15 (see NOTES for the queries):

| # | Criterion | State |
|---|---|---|
| 1 | RUNBOOK R1 census falls **and stays fallen after a rebuild** — the real test, since a content-level fix regresses and a template/schema fix does not | **PARTIAL.** 51 → **39** (22 empty href + 17 bare `#`); the fragment class is **extinct** (4 → 0). Survived a full rebuild *and* the v1.0.1140 image roll on 2026-07-20 — so the fixes are structural, as intended. The 17 bare `#` were never in scope (different class). Target: the 22 empty hrefs, which fall with candidate 2. |
| 2 | RUNBOOK R8 against the affected live tool pages: zero `href=""`, zero unresolvable fragments, no external host failing DNS | **MET**, and wider than filed. All four pages across three sites verified against the rendered artefacts post-image-roll: `finetuning.uk/tools/llm-cost-calculator`, `robot-hands.com/gripper-cycle-time-estimator`, both leopardess tool pages — 200, zero defects. |
| 3 | **No component in the active library can pair a rendered label with an absent destination** — i.e. classes A/C/E are structurally impossible, not merely absent from today's pages | **NOT MET**, but moving. After migration 181: **152 ungated CTA anchors / 29 components** (+17 range-scoped item links, a separate class) and **19 `source:llm` url fields / 5 components**. Both figures CORRECTED 2026-07-20 — the previous 70/37 was a measurement artefact; see below. Close it with candidates 2 + 4. |

**Closure bar for this file:** criteria 1–3 above — i.e. classes **A, B, C, E**. Everything
else has moved out, so that this file can close when *its own* scope is done:

| class | now tracked in | why it left |
|---|---|---|
| **D** hardcoded dead fragment | — **CLOSED HERE** | migration 179 removed it; class extinct fleet-wide (4 → 0 anchors) |
| **F** wrong-product static vocabulary | **`bugs_open/045`** | it is a *component-selection* defect (the library's only tool hero is a Bayesian ranker), not a label/url pairing one. Different fix, different blast radius. **Armed on 2 live `needs_rebuild` pages.** |
| **G** detection with no consumer | **`bugs_open/033`** | wider than CTA — 292 items fleet-wide, none ever actioned; blocked on an owner decision |
| **H** repair scope < detection scope | **council trail `2525f980`** | observe stage live in v1.0.1140, verified in-pod 2026-07-20; flip round follows the accumulated delta logs |

## Landmines

- **Do not flip `empty_internal_href` to blocker cold.** 30 live instances across 7 sites
  would fail the next rebuild of most of the fleet. Stage it: warning + work item → drain →
  flip. The repo already learned this with `phantom_internal_links` (LNK-009).
- **Fix at component/schema level, not in `page_components`** on leopardess —
  `bugs_open/001` (re-plan clobbers built pages) gives anything written to `page_components`
  an undefined shelf life. Template and schema fixes survive a re-plan; content_data fixes do not.
- **`slot_name` resolves via `content_components.function`, not `.name`**, and
  `component_id` is unpopulated fleet-wide. Querying by `name` returns zero rows and looks
  like an orphan; joining on `component_id` "finds" ~100 false orphans. I made the second
  mistake and corrected it — see NOTES.
- **Running the experience loop now will not help.** It is a detection loop, and these
  buttons are already correctly described in unread items — 34 on leopardess at filing,
  **still exactly 34 on 2026-07-20, none ever actioned.** More detection makes the invisible
  pile bigger. The consuming surface is `bugs_open/033`'s work, not this file's.

## Related

- **`bugs_open/045`** — class **F** split out 2026-07-20: the library's only tool-hero
  component is hard-wired to a Bayesian ranker. Armed on 2 live `needs_rebuild` pages.
- **`bugs_open/033`** — class **G** moved there 2026-07-20: the human-review queue has no
  working surface (292 items, none ever actioned).
- `bugs_open/039` — a section name resolving to **no** component renders a hollow stub; the
  sibling branch of the same selector that produced class F (which is: resolving to the
  **wrong** component). Worth fixing together with 045.
- `bugs_open/001` — re-plan clobbers built pages (constrains where the leopardess fix can live)
- `bugs_open/015` — mistyped `page_type` orphaned a page; same "routing key wrong → silence
  in every gate at once" shape
- `bugs_open/017` — static cutover orphans backend entry forms; the href *resolves*, so no
  check fires — adjacent hole in the same wall
- Link-integrity loop, closed 2026-07-16 (`docs/social001_vonc_tiktok_social/minilobby_task/`)
  — fixed the `/contact.html` lock for six components; its own documented residue (repair
  scope < detection scope, `*_label` fields left static) is exactly what bit us here.

---

## STATUS 2026-07-20 (bugfix-023 session) — the four buttons are extinct fleet-wide; the structural work is staged

**Fixed and verified live:**
- **All placements of both offending components are gone** — leopardess ×2 (P4,
  2026-07-20 morning, other session), finetuning.uk ×1 (both components),
  robot-hands.com ×1 (tool-guide-intro). Zero placements remain. R1 census: empty
  hrefs 30 → 22, **fragment class 4 → 0 (extinct)**. Evidence + backups in
  `cta_link_integrity/NOTES` (entries of 2026-07-20).
- **`tool-guide-intro` can no longer ship classes C/D/E** — migration 179 (applied +
  ledgered live, commit a6a31b8b1): `cta_primary_url` added (renderer/optional),
  `cta_secondary_url` flipped off `llm+required:true`, both anchors gated, `#guide-start`
  hardcode removed. Fix candidate 4's schema-lint rule ("no source:llm URL") is NOT yet
  a lint — 20 other llm-sourced URL fields remain (fleet sweep still owed).
- The fabricated-host class gained live confirmation: `finetuning.ai` (different-TLD
  variant) RESOLVES to a third-party-controlled page. P1.5 remains unimplemented.

**In flight (other thread, leopardess3):** fix candidate 1 (schema-derived pairing) is at
council trail `2525f980` round 5 REVISE — observe-only staging agreed, migration measured
a NO-OP (091/098 already flipped the mapped six; the residual 57 site_specs.* fields are
all UNMAPPED, so no config-only flip is safe). ⚠️ Their v3–v5 sketch carries a dead
observe-log (planSection's fresh map can never hold a prior value — the loss site is the
rerender merge) and a sibling rule that misses 3 mapped bare-stem fields while capturing
one image; corrections filed in `doc_notes` (`pipeline/plan_sections`,
`pipeline/resolve_internal_links`, category `correction`) and in NOTES 12:30/13:50.

**Still open:** fix candidates 3 (the handler — the delivery gap), 2/P2.1 (fleet anchor
gating, 75 ungated), 4 (schema-lint), 5 (external reachability), 6 (P1.5 email→hostname +
different-TLD), the generic `hero-tool` component (the selection landmine that caused the
Bayesian adoptions — `_pre_037` row is now placement-free but still the sole selectable),
and the class-A build check (P1.2).

---

## CORRECTION + RESIZING 2026-07-20 (bugfix-023 session 4) — the ungated count was wrong, and the live subset is small

**The 75/38 → 70/37 ungated figure this file has carried since filing is a measurement artefact.**
RUNBOOK **R9** attached its own warning ("a 60-character-lookback heuristic, not a parse …
re-derive the exact list with a real template parse before mass-editing") and that warning was
correct. A proper parse — tokenise each template, maintain an `{{if}}/{{range}}/{{with}}` block
stack, mark an anchor gated only when an enclosing condition references **the same field** —
gives:

```
189 href="{{.X}}" anchors in the active library
  GATED    18 anchors / 14 components
  UNGATED 171 anchors / 41 components      <-- vs R9's 70/37: a 2.4x undercount
```

> **REFINED an hour later, while writing migration 181 — and the refinement matters.**
> **15 of those 171 are not CTA anchors at all.** They sit inside a `{{range}}`, so the
> field is a property of the ranged item, not a component field:
> `{{range .items}}<a href="{{.url}}" class="tool-cta-card">`. That is an **item link** fed
> by a query-provided list — different class, different fix, different owner. The true CTA
> worklist is therefore **156 anchors / 32 components** (pre-181), not 171/41.
> Found the hard way: 181's first post-condition was a blanket *"no ungated `{{.x_url}}`
> anchor remains in these components"*, which `tool-cta`'s range-scoped `.url` would have
> tripped — rolling back an otherwise correct migration. Caught before applying, by reading
> the third anchor instead of trusting the count. `parse_gates.py` now reports the split, so
> the distinction cannot be lost again.
> **Note the shape:** I corrected a figure, then within the hour over-stated the corrected
> figure in the same way. The fix for "a number travelled without its caveat" is not a better
> number — it is making the tool emit the distinction.

R9 undercounts because its greedy `.{0,60}` prefix is consumed by the *previous* match, so in runs
of adjacent anchors (nav lists, footer link columns — exactly where they cluster) every other
anchor is swallowed. Script: `cta_link_integrity/scripts/parse_gates.py`; R9 now carries the
correction.

**But the job is smaller than 171, not bigger — because most of it is dormant.** Resolving every
component to its live placements (`page_components` by `function`/`name`; `site_components` by
`component_id`, which unlike `page_components` **is** populated fleet-wide):

- **21 of the 41 components are placed anywhere.** The 20 unplaced ones hold ~80 of the anchors,
  concentrated in library stock that nothing uses: `header-with-categories_pre_037` (27),
  `footer-with-disclaimer_pre_037` (18), `header-docs` (14), `site-head` (12),
  `header-with-cart-or-nav_pre_037` (11), `header-with-search_pre_037` (10).
- **Live-placed, ungated, worth doing first:** `content-block-about` (13 placements / 5 sites),
  `tool-list` (6), `case-studies-grid` (4), `system-stats` (4), `content-listing` (3),
  `guide-list_pre_037` (3), `tool-ai-agent-roi-estimator` (3), `tool-cta` (3), plus nine more at
  1–2, plus three chrome components (`site-header`, `site-footer`, `footer-theme-chrome`).

So P2.1 stages naturally: **corrective** (the 21 placed components) then **prophylactic** (the 20
dormant ones, zero live risk). That ordering was not available while the figure was a single
aggregate.

**Class E is still LIVE, and that is new.** Migration 179 fixed `tool-guide-intro`, but three
*placed* components still declare `source:llm, required:true` URL fields — a model instructed to
author a URL it cannot look up, which is precisely what fabricated
`leopardess.contactforsales.com`:

```
content-block-about   cta_url                              13 placements / 5 sites
tool-cta              primary_cta_url, secondary_cta_url    3 placements / 2 sites
platform-comparison   cta_url                               1 placement  / 1 site
```

17 live placements. Fix candidate 4 (the schema-lint) is therefore **corrective, not preventive**,
and should be sequenced accordingly. (The other 15 of the 22 `source:llm` url fields are nav-link
fields in the dormant `header-*` stock.)

### FIXED 2026-07-20 — migration 181, applied and ledgered (classes C+E on the live components)

`181_class_e_live_cta_url_integrity.sql`, applied in one transaction with needle-gates and
post-conditions (179's pattern), snapshot in `bak_class_e_components_20260720`:

| component | placements | what changed |
|---|---|---|
| `content-block-about` | 13 / 5 sites | `cta_url` **llm+required → renderer+optional**, anchor gated |
| `tool-cta` | 3 / 2 sites | `primary_cta_url` + `secondary_cta_url` **llm+required → renderer+optional**, both anchors gated |
| `platform-comparison` | 1 / 1 site | **anchor gated only — schema deliberately unchanged** |

Verified after apply: the three flipped fields read `renderer`/`false`; gated anchors
**18 → 22** (exactly the four); ungated CTA anchors **156 → 152**.

**The three are not treated alike, and the reasoning is the point:**

- `content-block-about` is **in `ctaFieldNames`**, so `chooseCTATargets` already writes this
  field on every render. Flipping the declared source to `renderer` makes the schema tell the
  truth about who owns it; nothing about the rendered output changes.
- `tool-cta` is **not** in the map, so after the flip nothing populates it and the gated
  template renders **no button** — the intended LNK-005 outcome, and strictly better here
  because the values it emitted while `llm+required` were **404s on live pages**
  (`finetuning.uk/tools`, `finetuning.uk/tools/llm-cost-calculator`,
  `leopardess/tools` — all verified 404 on 2026-07-20, and all three are in `bugs_open/049`'s
  census).
- `platform-comparison` was **gated but NOT flipped**. Its one live value,
  `vonc.com/tools/gauntlet/index.html`, is a **real working page (200)**. It is not in the map
  either, so flipping it would leave nothing to repopulate it and would **delete a working
  button from a live page**. The structurally correct change is blocked on the field having an
  owner — i.e. on the schema-derived pairing (council trail `2525f980`). Gating still removes
  the `href=""` failure mode without removing the button.

> **Residual, deliberate:** `platform-comparison.cta_url` remains `source:llm, required:true` —
> the last live class-E field. It is recorded here rather than quietly left, because "we closed
> class E" would be false. Flip it in the same change that gives it a resolver.

### And a defect class the gating sweep does NOT cover — filed as `bugs_open/049`

Sizing this sweep turned up **312 live broken-link instances across 7 sites** (68 unique 404
targets, on 117 of 180 live pages), dominated by `/privacy.html` + `/terms.html` in the footer of
every page of three sites. **It is not 023's defect** — the label/destination pairing is fine, the
destination just does not exist — and critically:

> **Gating an anchor does not help.** `{{if .x_url}}` tests non-emptiness; `/privacy.html` is
> non-empty, passes the gate, and 404s.

Cause proven and filed in `049`: the chrome renderer's hardcoded legal-link slice was fixed on
2026-06-10 (`0681e1542`), but chrome is re-rendered only on explicit trigger and nothing sweeps it,
so three sites still serve April/May artefacts. Do not count P2.1 as covering it.

---

## ADDENDUM 2026-07-20 (robot-hands thread) — the cause is stronger than "nothing pairs them"

Found while repairing ~40 mispaired CTAs on robot-hands.com. **The URL is not
authored and then left unchecked against the label — it is not authored at all.**
It is recomputed on every render, label-blind, from nav order.

`platform/orchestration/actions/resolve_internal_links_action.go`:

- `ctaFieldNames` (`:99-105`) declares the fields the resolver owns:
  `hero{cta_url, secondary_cta_url}`, `call-to-action{primary_cta_url,
  secondary_cta_url}`, `content-block-about{cta_url}`,
  `archetype-grid{cta_url}`, `archetype-combinations{...}`, `gauntlet-cta{...}`.
- `chooseCTATargets` (`:319-350`) computes them. It ranks interactive pages then
  hubs, drops `areasExcludedFromCTA` and self-links, sorts by **`NavOrder` then
  `Name`**, and returns `primary = ordered[0]`, `secondary = ordered[1]`.
  **The label is never read.**

So every CTA of a given kind on a site converges on the *same* two destinations,
whatever the buttons say. On robot-hands that produced 20 components across 11
pages all pointing at `/tools/matchmatrix/index.html` — including "Search the
Gripper Catalog", "Browse the Learning Center" and "Open the Payload Calculator"
— for the single reason that `tool-matchmatrix` sorted first among interactive
pages. It was a 404 at the time, which is how it got noticed at all; the 20
mispaired *secondary* CTAs never 404'd and nothing had ever flagged them.

**Demonstrated, not inferred.** I corrected `content_data` by label
(`docs024_key_docs_latest/robot_hands/SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql`),
the page re-rendered, and the resolver put its own choice back — URL *and*
`primary_cta_target_title` — with a later `updated_at` than my write.

**Two consequences for whoever fixes this:**

1. **A content-side fix cannot hold.** Teaching the content writer to pair label
   and URL, or correcting `content_data` directly, is overwritten by the next
   render. The fix belongs in `chooseCTATargets` — either make it label-aware, or
   give an explicitly-authored URL precedence over the derived one. Note the
   staged-rollout comment at `:79-84`: `ctaFieldNames` is currently an OVERRIDE on
   the schema-derived pairing in `datahelpers/ctafields.go`, which runs
   OBSERVE-ONLY pending a council round. Any fix should land inside that plan
   rather than beside it.
2. **`areasExcludedFromCTA` (`:72-74`) = `{about, contact, privacy, terms, legal}`
   makes some correct pairings unreachable.** A button reading "Request
   Integration Support" *should* go to `/contact.html`, and the resolver will
   never allow it. That exclusion is sensible as a default for a generated CTA and
   wrong as an absolute — it needs an authored-intent escape hatch, or labels of
   that kind will stay permanently mispaired.

Related: `/bugs_open/043` (generated copy invents quantitative claims) was found
in the same sweep on the same site.

### Proposal for the flip round: make field ownership queryable (robot-hands, 2026-07-20)

Offered to the owning thread, not built — this belongs inside the staged rollout, not
beside it.

The knowledge that would have prevented my whole class of error **already exists in this
file**, three lines above the map:

```go
// re-resolved into resolved_data on every render and merges last, so no
// recompute or content edit can win against it.
var ctaFieldNames = map[string][2]string{
```

That comment is exactly right and it reached nobody. It is visible only to someone already
reading `resolve_internal_links_action.go` — which is to say, someone who has already
worked out that the resolver is involved. A session editing `content_data` over psql has no
path to it, and will be told nothing when its write is silently reverted on the next render.

**The ask is small: emit the ownership set.** Whatever the flip settles on as the authority
(`ctaFieldNames`, or the schema-derived successor in `datahelpers/ctafields.go`), write it
somewhere queryable — a generated JSON committed beside the Go, or a small table. Then one
source feeds three consumers that currently have none:

1. **A pre-apply check** — `is this field mine to write?` before a session edits
   `content_data`. This is the load-bearing one: a pre-commit check fires *after* the write
   has already reached production, which is where the real cost lands.
2. **`scripts/pattern-check.py`** — staged SQL touching a resolver-owned key. Mechanically
   decidable, and it would have fired on
   `robot_hands/SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql`. Hold it to that script's
   own bar (documented incident ✓, decidable ✓, ≤2% fire rate — **unmeasured**, measure
   before wiring).
3. **The conflict log you are already building at the rerender merge** — it can name the
   owner in its message rather than just recording that a value was replaced.

Cheapest possible version, if the above is too much for this round: put the comment's
content in the RUNBOOK under a heading someone editing `content_data` would actually search
for. That does not fix it, but it moves the knowledge from "in the file that defines it" to
"in the file people read before touching the data", which is where it failed.

**Why this is offered rather than done:** `ctaFieldNames` is mid-flip (OBSERVE-ONLY,
`:79-84`), so freezing its shape into a second artefact now would create exactly the
two-hand-maintained-lists drift this council reviews for — the same trap the gate-roster
mirror (`099_SYNC_gate_roster.py`) exists to prevent. Whatever emits it should be generated
from the winner, after the flip.

---

## CONTRIBUTED FINDING (2026-07-25, idea-uk-vm-site session) — a SECOND self-link path, outside `resolve_internal_links`

Filed here rather than as a new bug: `scripts/who-owns.py 023` says this file is owned and
actively worked, and this is the same family. **Nothing here competes with the flip** — it is a
different derivation, in the component schema layer, not in `ctaFieldNames`/`chooseCTATargets`.

### What was observed (live, two sites, curl-verified — not inferred)

`guide-list_pre_037` (`9d5e461a-8981-4ecc-b236-05895edfc15d`) renders, **on the guides hub
itself**:

```html
<a class="guide-list-cta-btn" href="/guides/index.html">Browse all guides</a>
```

- gamesdesign.co.uk `/guides/index.html` — self-link (curl-verified 2026-07-25)
- relojistas.com `/guias/index.html` — self-link (`cta_url=/guias/index.html` on its own hub)
- idea.uk `/guides/index.html` — self-link, and worse: it sat directly under authored copy
  promising a £29 report, so the label, the copy and the destination all disagreed.

A dead control (destination = the page you are on) with a label that contradicts the copy beside
it. Exactly this bug's shape, reached by a route this bug does not currently cover.

### Mechanism — and it is NOT `chooseCTATargets`

`chooseCTATargets` (`resolve_internal_links_action.go:319-350`) already **drops self-links**. This
CTA never goes near it. It is resolved from the component's own `input_schema`:

| field | source | effect |
|---|---|---|
| `cta_url` | `query.section_index_for:guide` | resolves to the guides hub — *the page the component lives on* |
| `cta_button_label` | `static`, `fallback: 'Browse all guides'` | fallback written unconditionally |
| `eyebrow_label` | `static`, `fallback: 'Guides'` | fallback written unconditionally |

`queryresolve.section_index_for` has **no self-link guard**, unlike `chooseCTATargets`. So the
same defect class has two independent derivations and only one of them is guarded — worth knowing
before the flip declares `ctaFieldNames` the single owner of CTA truth, because it is not.

**The `static`+`fallback` half is arguably nastier and is fully general, not guide-list-specific.**
`plan_sections_action.go:1556-1562`:

```go
if source == "renderer" || source == "static" || ... {
        if fallback != nil {
                resolvedData[fieldName] = fallback
        }
        continue
}
```

`resolved_data` merges **LAST**, so a `static` field with a `fallback` is an unconditional
constant: `content_data` is never consulted. Yet both fields' own `llm_guidance` reads *"Override
if the site tone prefers a different phrasing / calls for something more specific."* **You cannot
override it.** The guidance is false, and it is the guidance an author (human or LLM) reads before
writing the value that then gets silently discarded. **Any schema field with `source: static` and
a non-null `fallback` is in this state** — worth a fleet sweep; I only checked this component.

Also destructive, not merely presentational: the rerender writes resolved values **back into
`page_components.content_data`**, so the authored value is gone from the DB after one render, not
just absent from the HTML. That is what happened to idea.uk's `/report.html` target.

### What I changed, and its blast radius

`idea_uk_vm_site/sql/p4_04_guide_list_cta_overridable.sql` — on the shared component: dropped the
query `source` from `cta_url` and the `fallback` from `cta_button_label` / `eyebrow_label`.
`items` **keeps** `query.pages_where_type:guide` (that list must stay derived).

**Verified no-op for the other three instances before writing**: all four already carry those three
keys in `content_data` with exactly the values the resolver produced, so gamesdesign ×2 and
relojistas ×1 render byte-identically. Whether *they* should keep a self-linking hub CTA is this
workstream's call, not mine — I have not changed what they render.

Known consequence, stated up front: a **new** guide-list instance with empty `content_data` now
renders with no CTA button rather than an auto-filled self-link (the template gates on
`{{if .cta_url}}`). That is this bug's own principle applied — do not emit a control with no
authored destination — but it is a behaviour change for future instances, so it belongs in your
flip's accounting.

### Suggested for this bug's backlog (not done — owned here)

1. Give `queryresolve.section_index_for` the same self-link guard `chooseCTATargets` has: it knows
   the resolving page in the rerender path (`pageURL` is already threaded to `applyCTARecompute`).
2. Sweep `input_schema` for `source: 'static' AND fallback IS NOT NULL` and decide per field
   whether it is a genuine constant or a stolen override — then fix the `llm_guidance` that
   promises an override that cannot happen.

---

## CONTRIBUTED FINDING (2026-07-25, vetcomparison.uk session) — a THIRD derivation: authored links inside a repeating array, which the `[2]string` contract cannot express

Filed here, not as a new bug: `scripts/who-owns.py 023` reports this file OWNED and actively
worked (34 commits/14d, one the same day). **Nothing here competes with the flip.** It is neither
`chooseCTATargets` nor `queryresolve.section_index_for` — there is *no resolver in this path at
all*. It is the bug's literal title case (label and destination never checked against each other)
reached by authored content, and it is invisible to the gate for a **structural** reason worth
knowing before the flip declares `ctaFieldNames` the single owner of CTA truth.

### What was observed (live, curl-verified 2026-07-25 — not inferred)

vetcomparison.uk's homepage carries **9 broken internal links**. Six are in one component,
`info-card-grid` — and they are *every anchor it has*, a 100% dead-link rate on the page's main
content grid:

| card title | label | `link_url` | live |
|---|---|---|---|
| Search practices by location | Search the directory → | `/search` | 404 |
| Compare prices before you go | Read about pricing → | `/about-pricing` | 404 |
| See who owns your practice | Read about ownership → | `/about-ownership-disclosure` | 404 |
| Understand your rights as a pet owner | Read the guide → | `/guides/pet-owner-rights` | 404 |
| Claim or correct your listing | Claim your listing → | `/claim-listing` | 404 |
| CMA compliance guides for practices | Read the guides → | `/guides/cma-compliance` | 404 — **page exists** |

Five have no `pages` row at all (phantom). The sixth is a **URL-form** miss: the page exists, is
`build_status='deployed'` and returns 200 at `/guides/cma-compliance/index.html`. This site's host
does no directory-index rewrite — `/guides/cma-compliance` and `/guides/cma-compliance/` both 404.
The remaining 3 of the 9 are chrome links to `planned`/never-deployed pages
(`/directory/index.html`, `/guides/index.html`, `/tools/compliance-deadline-calculator/index.html`),
i.e. `049` mechanism 2, not this bug.

### Mechanism — the fields are `cards[].link_url`, not a scalar CTA field

`info-card-grid` is **absent from `ctaFieldNames`** (`resolve_internal_links_action.go:98-105`,
6 components: hero, call-to-action, archetype-grid, archetype-combinations, gauntlet-cta,
content-block-about). So neither writer of CTA destinations touches it — not build-time
`setCTAField`, not repair-time `applyCTARecompute`.

**Adding it to the map would not work, and that is the finding.** The map's value type is
`[2]string` — a primary/secondary pair of *scalar* field names. This component's destinations live
one level down, inside a repeating array:

```
content_data.cards[] -> { title, body, icon, link_label, link_url }
```

Six links, N-ary, each with its own label. There is no pair of top-level field names that names
them. **Any component whose links live in a repeating array is not merely unlisted — it is
unrepresentable in the current contract**, so the flip cannot fix this class by enrolment alone.
The label/destination pairing this bug is named for is *right there* in each array item
(`link_label` beside `link_url`) — the data needed to check the pairing exists; nothing reads it.

`[VERIFIED]` The URLs are authored, not derived: no `site_specs` row for this site contains them
(all 12 current aspects checked), and `site_plan_sections` holds no content at all (structure only
— `component_name`/`ordering`/styling ids). They persist solely in `page_components.content_data`.

### Why the audit-time backstop did not catch it either

`check_phantom_internal_links` is enabled (`completeness-discovery-agent`'s checks array carries
`phantom_internal_links`, `misdirected_cta`, `dead_controls`) and its tests cover exactly this
shape (`/ghost.html`, `/never-built.html`). But:

`[VERIFIED]` **vetcomparison.uk has ZERO rows of every link/CTA item type, across all time.**
Fleet-wide those types total **188** rows: `cta_names_unknown_destination` 69, `unresolved_cta` 68,
`needs_internal_links` 25, `phantom_internal_link` 22, `cta_improvement` 4.

Broken down by site for the two destination-validity types only
(`phantom_internal_link` + `cta_names_unknown_destination`, 91 of those 188 rows), **six sites
have rows and vetcomparison is not among them**: ai-agent-orchestration 27, robot-hands 20,
idea.uk 17, leopardess 13, relojistas 7, vonc 7.

`[INFERRED]` — therefore the completeness discovery agent has not run against this site's deployed
HTML. I could not prove it directly: `orchestration_states` is pruned at ~24h, so its silence is
not evidence (the over-flagging trap from `bugs_open/044`). The inference rests on the check being
deterministic over deployed HTML — six phantoms are present right now, so a run would have had to
produce rows. **If the flip's accounting assumes the audit backstop covers every live site, that
assumption is worth testing rather than trusting** — a per-site "when did discovery last run"
figure would settle it, and I did not find one that survives the 24h prune.

### What I changed (site content, not platform — one link of six)

Repointed **only** the card whose destination exists, in `content_data` *and* `rendered_html`:
`/guides/cma-compliance` → `/guides/cma-compliance/index.html`. Snapshot first
(`_vetcomparison_bak_20260725_index_components`); card order and all six cards preserved.

**Stated honestly: this is not durable, and it is not a fix for this bug.** Renders
delete-and-recreate agent-writable rows (`save_page_sections_action.go:498`), so the edit holds
across an assemble-only render (content_data was byte-identical across the 07-24→07-25 renders,
though the row id changed) but a full content re-render regenerates the copy and can reproduce the
old URL. The other five need a destination that does not exist yet — an owner decision on this
site, not something to paper over.

### Suggested for this bug's backlog (not done — owned here)

1. **The contract, not the enrolment, is the gap**: `ctaFieldNames`'s `[2]string` cannot address
   `cards[].link_url`. A component-level descriptor that can name a repeating path
   (e.g. `cards[].link_url` + its sibling `cards[].link_label`) would cover this class; enrolling
   `info-card-grid` under the present type cannot.
2. The label/destination pair is co-located per array item — the cheapest possible pairing check,
   and the one this bug is named for, is available here and unused.
3. Worth a fleet sweep: **which components carry anchors in a repeating array?** They are all in
   this same blind spot. `info-card-grid` is live on more sites than this one.

---

## CLOSURE 2026-07-25 (bugfix-023 session) — the library-wide sweep; classes A/B/C/E are structurally closed

**What was still open when this session started:** criterion 3 — *no component in the ACTIVE
library can pair a rendered label with an absent destination* — plus criterion 1's remaining
target, the empty `href=""` class. Migrations 179 and 181 had done this for the four components
that happened to be on live pages. **Scoping the last pass to live placements would have been
wrong twice over:** criterion 3 says *active library*, and `bugs_open/045` is precisely the case
of dormant library stock being adopted onto a live page and shipping its frozen defaults.

### Measured live 2026-07-25 (a real parse, not R9)

```
TOTAL href="{{.X}}" anchors : 213 across 142 active components
  GATED   :  40 / 24 components
  UNGATED : 173 / 43
     inside {{range}} (item links, separate class) :  17 / 13
     NOT in a range (THE CTA WORKLIST)             : 156 / 31   <- 31 live-placed, 125 dormant
llm + required:true URL fields                     :  23 / 5 components
R1 census: 33 href="" (5 sites) · 11 bare # · 1 fragment
```

The single fragment is a **false positive**: `gauntlet-interface`'s `#gi-rules` resolves to a real
`id="gi-rules"` in its own template. Class D remains extinct. The `href=""` count is *up* on
2026-07-20's 22 — not regression, but five days of new sites and components (webdesign.co.uk was
built the same day).

### What shipped (both applied + ledgered; config → LIVE immediately, no image roll)

**`212_cta_gate_archetype_taster_quiz.sql`** — the last hardcoded `#`, found by the new lint on
its first run. A field parse cannot see an anchor whose href is a literal `#`; the lint was
written to report a *different* shape than the migration swept, which is why it caught one more.

**`211_cta_gate_every_active_anchor.sql`**

| | change |
|---|---|
| A | **156 CTA anchors / 31 components** wrapped in `{{if .x_url}}` — an unresolved destination now renders **no control** (LNK-005) |
| B | **7 placeholder anchors / 4 components** the field parse cannot see: `brief-explanation` ×2 (`href="{{if .x}}{{.x}}{{else}}#{{end}}"` — 4 live placements / 3 sites), `category-listing`, `Pricing Tiers` ×3, `featured_article` (hardcoded `href="#"`), each given a **renderer/optional url field** and a gate |
| C | **23 `source:llm` + `required:true` URL fields / 5 components → `required:false`**, fleet-wide |

**Class C's reasoning matters, and it is the `platform-comparison` lesson from 181 generalised.**
`source` is deliberately left `llm`: flipping the source with no resolver behind the field
*deletes values that are real today*. What creates the fabrication is the **compulsion** — a
required field the model cannot look up — so removing `required` plus gating the anchor kills the
defect without deleting a working button. `leopardess.contactforsales.com` was the site's own
contact address with `@`→`.`; `finetuning.ai` was the different-TLD variant of the same move.

**`scripts/check_cta_gates.py`** — the standing lint (RUNBOOK **R17**). Reports UNGATED,
PLACEHOLDER, LLM_URL and NO_VALUE across the whole library in ~1s, with its deliberate exclusions
written into the file. **RUNBOOK R18** is the safe mass-edit recipe this needed.

### Evidence

**Live, after apply** (`parse_gates.py` on a fresh dump; the class-E query):

```
UNGATED CTA worklist                : 0   (was 156)
llm + required:true URL fields      : 0   (was 23)
GATED                               : 203 anchors / 59 components
lint findings                       : UNGATED 0, LLM_URL 0
```

**Before/after through the platform's own render engine, on real data.** The deployed
`who-we-help.html` on leopardess carries 6 dead controls from `case-studies-grid`. Rendering the
page's actual stored `content_data` (52 keys) through `text/template` with `missingkey=zero` —
the exact configuration of `executeGoTemplate` (`call_agent.go:1150`), which is what
`RenderTemplate` uses:

| template | anchors rendered | `<no value>` | after the platform's `<no value>` strip |
|---|---|---|---|
| pre-migration (from `bak_cta_gates_20260725`) | **6** | 6 | 6 × `href=""` — exactly what is deployed |
| live, post-211 | **0** | 0 | no control at all |

That is the mechanism, proven against the real engine and the real data rather than asserted.

**And the same check across all 31 gated components, both ways** — because a gate that removes
every button is not a fix either. Each template was rendered twice with the platform's engine
settings: once with **no data** (expect no control) and once with **every gated field populated**
(expect every control back).

```
31 of 31 : with fields populated, every gated anchor returns  (2 to 28 anchors each)
31 of 31 : with no data, zero anchors with an empty href
```

Six components do emit exactly one anchor with no data, and reading them is the point: five are
the **site brand link** `<a href="/" class="…-brand">` — a hardcoded root link that always
resolves — and the sixth is `provocations-archive-list`'s hidden `[data-archive-template]` clone
source. None is a control with an absent destination. *(A first pass of this sweep flagged all
six as failures because its rule was "no anchors at all with no data". The rule was wrong, not
the components — a fixed, always-valid href is not this bug's defect.)*

> **The one thing gating cannot do, stated plainly: it does not rewrite HTML that is already
> deployed.** `page_components.rendered_html` is a stored artefact. Pre-existing dead controls
> drain as pages re-render. This is exactly why `info-card-grid`'s 8 live `href=""` persisted on
> gaswholesalers/aao although that component has been gated for a long time — I mis-scoped them
> as in-scope until I read the template. **A live empty href is not proof the template is
> ungated, and a gated template is not proof the live page is clean.**

### Criteria

| # | criterion | verdict |
|---|---|---|
| 1 | R1 census falls and **stays fallen after a rebuild** | **MET as to mechanism.** The identical change shape on four components (179/181) was already shown to survive a full rebuild *and* the v1.0.1140 image roll — recorded in the table above. 211 extends that proven shape to the whole library, and the render check above reproduces the fall on real data. The *count* falls per page as each re-renders; that drain is not a defect and not this file's to wait on. |
| 2 | R8 against the affected live tool pages | **MET** (unchanged since 2026-07-20; re-verified). |
| 3 | **No component in the active library can pair a rendered label with an absent destination** | **MET, verified live**: ungated CTA anchors 0, `llm`+`required` URL fields 0, and a standing lint that fails if either returns. |

### Deliberate residuals — recorded, not quietly dropped (the lint keeps reporting them)

- **`image-hover-card-grid`** — `href="{{if .link_url}}{{.link_url}}{{else}}#{{end}}"` inside
  `{{range .cards}}`. Its anchor wraps the card's **image, title and description**, so gating it
  deletes content, not a control; it needs an `{{else}}<div>` restructure. Item-link class.
- **17 range-scoped item links / 13 components** — the field belongs to the ranged item, fed by a
  query. Different class, different owner. A blanket "no ungated url anchor" post-condition trips
  on them and rolls back correct migrations (181's first draft did exactly that).
- **`lobby-grid` + `provocation-card`** — their `html_template` contains the literal `<no value>`
  (37 and 13 occurrences) and **no Go template actions at all**: rendered artefacts saved back as
  templates. Both are `data-runtime-fill="true"` vonc components — the vonc workstream's own
  documented landmine, owned there, deliberately not filed as a competing bug.
- **`provocations-archive-list`'s `href="#"`** is a hidden `[data-archive-template]` clone source
  filled by JS. Gating it would break the archive. The lint excludes `hidden` / `data-*template`
  anchors for this reason.

### Still open — under their own numbers, never this file's scope

`bugs_open/033` (the human-review queue that has never delivered a finding — the platform
correctly detected one of the original four buttons two days before the owner clicked it),
`bugs_open/045` + `039` (component selection: the wrong component, and no component),
`bugs_open/049` (links to pages that do not exist — gating cannot help, `/privacy.html` is
non-empty and 404s), and **the flip round** (stage 2 of the schema-derived pairing, carrying its
five council constraints), which is where `platform-comparison.cta_url` — the last live URL field
with no owner — finally gets one.

**Full record:** `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/` —
`NOTES` (the missteps, including a regex-dialect trap that would have mangled 21 of 35 live
templates and read as obviously correct), `RUNBOOK` R17/R18, `SUMMARY_2026-07-25`,
`README_where_we_are`, and `016b` §9 for the transferable version.
