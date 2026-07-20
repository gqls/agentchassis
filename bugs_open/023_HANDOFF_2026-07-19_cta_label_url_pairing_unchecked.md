# 023 — A button's label and its destination are never checked against each other

**Filed:** 2026-07-19 · **Branch:** `085_debug_and_feature_loops` · **Status:** OPEN, not started
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

Criteria, all re-measured 2026-07-20 19:15 (see NOTES for the queries):

| # | Criterion | State |
|---|---|---|
| 1 | RUNBOOK R1 census falls **and stays fallen after a rebuild** — the real test, since a content-level fix regresses and a template/schema fix does not | **PARTIAL.** 51 → **39** (22 empty href + 17 bare `#`); the fragment class is **extinct** (4 → 0). Survived a full rebuild *and* the v1.0.1140 image roll on 2026-07-20 — so the fixes are structural, as intended. The 17 bare `#` were never in scope (different class). Target: the 22 empty hrefs, which fall with candidate 2. |
| 2 | RUNBOOK R8 against the affected live tool pages: zero `href=""`, zero unresolvable fragments, no external host failing DNS | **MET**, and wider than filed. All four pages across three sites verified against the rendered artefacts post-image-roll: `finetuning.uk/tools/llm-cost-calculator`, `robot-hands.com/gripper-cycle-time-estimator`, both leopardess tool pages — 200, zero defects. |
| 3 | **No component in the active library can pair a rendered label with an absent destination** — i.e. classes A/C/E are structurally impossible, not merely absent from today's pages | **NOT MET** — this is the real remaining bar. Current: **171 ungated anchors / 41 components** (CORRECTED 2026-07-20, see below — the 70/37 figure was a measurement artefact) and **22 `source:llm` url fields / 6 components** (was 21). Close it with candidates 2 + 4. |

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
