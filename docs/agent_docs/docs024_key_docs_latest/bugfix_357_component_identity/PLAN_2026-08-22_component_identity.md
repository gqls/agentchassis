# PLAN — 2026-08-22, component identity vs stored bytes (`bugs_open/357`)

Design, phasing, decisions **and their reasons**. Corrections to the originating brief live here,
marked as corrections.

Council: `SUBMISSION_CORR=62aac6c2-996f-4b5d-8f8f-72e3daf4c82e` (Layers A/B/D + seed).
Diagnosis: `RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba`.
Plan drafted with `fable`; every mechanism claim below re-verified first-hand at the cited line.

---

## 1. The defect, in its general form

> A component row's **identity** (`component_id`, `slot_name`), its **bytes** (`rendered_html`) and
> its **content** (`content_data`) are written, carried and re-written as an atomic bundle by every
> path in the page-composition system — and **no seam anywhere asserts that the three agree.**

Where identity is *invented* rather than carried, it is invented **positionally**, from a plan,
without looking at the bytes. That is the whole bug, and it is why the fix is not a relabelling job.

## 2. The mechanism, in two halves

**The mint** (`save_page_sections_action.go`). A tool/game page is one `<div class="tool-page">`
fragment with no `<section>`, so the section regex matches nothing. The documented fallback
(:1453–1476) stores the whole fragment as ONE section named `"section"` — the sentinel for *identity
unknown*. `enrichSectionsWithPlannedNames` (:1778–1789) replaces the sentinel with
`planned[Position-1]` from `pages.sections`; `hero` is planned first on all 22 affected pages, so the
tool is named `hero` by position. `enrichSectionsWithComponentIDs` resolves that to the shared hero
UUID, and the single INSERT (:999) persists the pairing.

**The conservation loop** (the half I got wrong first — see §7). `RerenderPageSectionsAction` renders
a fresh hero title band for the hero slot. The save's **Layer 2 interactive carry-forward**
(:470–550) then finds the stored row for that slot is interactive while the incoming content is not,
and — *"Same slot, non-interactive rebuild content (e.g. the hero regenerated as plain text). Keep
the existing interactive markup in place rather than overwriting the tool with prose"* (:522–535) —
splices the **stored tool bytes and stored hero `content_data`** back into the section that still
carries the **hero identity**. The INSERT re-mints the mismatch.

> **The mechanism that PROTECTS the tool from being blanked is the same mechanism that PERPETUATES
> the mislabelling.** Layer 2 is doing exactly its job (the `bugs_closed/004`/`005` lineage); it
> preserves bytes under whatever identity the plan supplied, because nothing tells it the two
> disagree.

Consequence that sets the phasing: **the population is self-renewing.** Repairing rows without
fixing the producer is wasted work.

## 3. The discriminator, measured rather than chosen

357's fix candidate 2 rightly says a birth guard must be censused first. The naive
static-template-prefix test flags **158** rows and fires on legitimate template drift — unusable at a
refusal seam. The component's **own self-declaration** (`data-component`, which the extractor already
trusts as the identity carrier in this same file) gives:

| | rows |
|---|---|
| agree | **1,550** ← the demand control |
| both declared, disagree | **0** |
| flagged (template declares, HTML silent) | **27** = **24 defects + 3 legitimate** |

> **CORRECTION (same day):** I first reported this as "zero false positives fleet-wide" having not
> opened five of the 27. Three are legitimate `loancash.co.uk` verbatim pages. Logged in
> `WRONG_CALLS.md`.

**Limit, stated rather than glossed:** 190 of 339 components declare no `data-component`, so this is
a **refuse-on-certainty** test, not a census.

## 4. The layers, ordered by what closes the door

| | what | where | why there |
|---|---|---|---|
| **A** | the no-`<section>` fallback emits an **already-typed** `adopted-fragment` row (template exactly `{{.body}}` — the identity function, so the row is genuinely regenerable) instead of the `"section"` sentinel; `enrichSectionsWithPlannedNames` skips positional renaming for fallback-derived sections | `save_page_sections_action.go` + seed migration | identity is *invented* here; typing at birth is the only version where nothing downstream must distrust the row. **Refusing here was rejected**: the fallback exists because zero sections leaves the page with no components and the rerender silently skips it (:1438–1452) |
| **B** | verify identity against the bytes at the single INSERT; **record always** (Error level), **refuse only** behind `refuse_unverifiable_component_identity`, default **OFF**, seeded on nobody | new `save_sections_identity_guard.go` | the one place every composition path converges, and the file's own words for the precedent guard: *"the single INSERT every page-composition path flows through"* |
| **D** | fleet-wide detection filing the same typed item, covering the **five other** `page_components` writers without touching them | new `discovery_checks/check_component_identity.go` | five edits that each depend on the author remembering is the alternative 342 explicitly rejected |
| **C** | repair the 22 (byte-preserving re-type; `content_data.body` := the stored bytes) | `_HOLD` migration | **NOT in the council submission.** It changes what four live sites serve — the owner's decision — and must follow the A roll or the producer re-mints a repaired page |

### Decisions worth recording, with reasons

- **Guard placement varies its own precedent, deliberately.** `sectionIsUnresolvableStub` skips
  inside the insert loop, *after* the DELETE — right for a hollow stub, catastrophic here, where the
  row holds the only copy of a tool. The identity guard runs **before the snapshot and the DELETE**,
  beside the claims/shrink/floor guards, honouring the file's stated doctrine that *"a refused save
  writes nothing at all"*. It runs **after** the Layer 2 carry-forward so it judges the final bytes —
  which is precisely what breaks the conservation loop once armed.
- **No verbatim exemption at the seam — reachability, not omission.** A verbatim page is
  `rebuild_policy='owned'` by definition of the triple, and the owned-page guard (:172–243) refuses
  the whole save unconditionally before every section guard [VERIFIED: a bare block right after the
  page lookup, whose comment says it refuses *"rather than fall through to the heuristic guards
  below"*]. A copied-in exemption here would be dead code inviting someone to delete the real
  protection. **The detector, reading rows at rest, must carry the exemption explicitly** — sharing
  the canonical triple with `loadVerbatimPageHTML` rather than respelling it.
- **One definition of "what does this thing say it is".** `data-component` is already compiled
  inline **twice** (:1392, :1498) and the guard plus detector would make four spellings in two
  languages — while they must differ subtly (HTML `[^"]+`, template `[^"{]+`, so an interpolated
  attribute is never a declaration). One helper pair, plus a SQL rendering from the same source on
  the `interactiveHTMLSQL` "ONE definition, two languages" precedent (:1695–1710) with its parity
  test.
- **The seam and the detector ask different questions and must not be reconciled.** The seam:
  *could the component I am about to bind have produced the bytes I am about to store?* The
  detector: *does this stored row still agree with its component's CURRENT self-declaration?* They
  diverge on template drift — a row flagged after its template's attribute changed is a **stale
  row**, not a detector bug. Stated at both sites, on 342's precedent that a divergence like this
  gets a test with the reason in its failure message.

## 5. Why the refusal arm exists at all, given `342`

`bugs_open/342` shipped **report-only** at the adjacent render seam one day earlier, citing the
2026-08-02 §2 ruling. What ships here behaves identically — the field is seeded on nobody. Three
things separate the cases, and **342 itself names this plan's shape as its own next step**: its risks
block says *"the honest next step is a per-path decision about refusal or a work item, and it is
deliberately not in this change."*

1. 342's predicate judges content *quality* (optional fields, fleet defaults, schema-less
   components), so refusal would have had false positives **by design**. Ours is a certainty test.
2. 342's load-bearing clause — *"the two paths that DO want to refuse already do, before the
   render"* — has **no analogue**: the INSERT is the first point where bytes, component and template
   coexist. Copying its conclusion means report-only for ever, with the conservation loop still
   running.
3. 342 added the **first** refusal to a seam that had none. This action already refuses on claims,
   shrink, layout flatness, completeness and ownership; the new thing is only the predicate.

## 6. Verification — at the artefact, with the disconfirming result named

- **A:** after the roll, one tool recreation → the new row is `adopted-fragment`-typed with
  `rendered_html = content_data->>'body'`, and the page serves `class="tool-page"` plus its
  `<script>`. *Disconfirming:* a new `slot_name='hero'` row with attribute-silent bytes; a page with
  zero components; a page missing its tool markup. Then the population query on a cadence —
  *disconfirming:* the count grows with post-roll `created_at`.
- **B (record):** census reproduces 1,550 / 0 / 27 before, and N / 0 / 2 after the repair.
  *Disconfirming:* any of the 1,550 agreeing rows flagged, or a seeded known-bad fixture missed.
- **B (refusal):** integration test — a save carrying hero-typed tool bytes is refused and the
  page's rows are **unchanged**, asserted by row count AND per-row md5. *Disconfirming:* the row
  count drops after a refusal, which is the placement bug this design varies its precedent to avoid.
  **Mutation test:** move the guard below the DELETE and it must FAIL.
- **C:** per page — md5(`rendered_html`) identical pre/post; `rendered_html = content_data->>'body'`;
  the row leaves the population; `curl` each URL asserting the tool's own markup, controls and
  `<script>` (`bugs_closed/287`: a `complete` item proves nothing).
- **Conservation re-check:** after the repair, let one rerender run on a repaired page — the fresh
  rows must still be `adopted-fragment`-typed with identical bytes. *Disconfirming:* the slot reverts
  to `hero`, meaning a positional re-derivation path this plan missed.

## 7. Corrections to this plan's own originating brief

> **CORRECTED 2026-08-22:** I briefed the planner that the 13 rows carrying a full hero
> `content_data` were "already armed" — one rebuild from having the tool replaced by a title band.
> **Refuted**: `vetcomparison.uk` `index` has six completed rerenders 08-19→08-22 with the tool
> serving throughout, and its rows were re-created *inside* the 08-22 rerender window. The rerender
> is the **author** of the mismatched row, not a threat to it. Nothing in this plan rests on the
> retracted claim, and the corrected mechanism (§2) is what produced the phasing in §4.

> **CORRECTED 2026-08-22:** `ContentDataCanFillTemplate` is **not** a gate on any rebuild path — it
> has exactly one non-test caller, `discovery_checks/check_literal_markdown.go:429`, a detector's
> routing classifier. It is `bugs_closed/277`'s arming mechanism, and I imported it here without
> testing its reach. The repair's safety argument rests instead on the `{{.body}}` template being the
> identity function, so a regeneration reproduces the stored bytes **by construction**.

## 8. Deliberately not done, and why

- The static-prefix guard as 357 wrote it (158 hits, fires on drift).
- Backfilling `content_data` onto the 9 NULL rows or "completing" the 13 — any repair leaving hero
  identity in place while touching `content_data` polishes the lie.
- Deleting positional enrichment wholesale — it serves genuine multi-section alignment.
- Flipping the pages to `deploy_mode='verbatim'` — these are chrome-less fragments on assembled
  pages, several multi-component, and the verbatim/assembled flip is the **row count**.
- A second hand-rolled decomposer — port `loancalculator_couk/decompose/load_decomposition.py`'s
  conventions (back up every affected page first, per-page transactions, `_provenance`), not its job.
- A DB constraint on identity — the predicate needs the joined template, and a constraint fails
  workflows with no typed item and no route.
- Blanket locks on the 22 — precaution without a measured threat, and each rerender of a locked row
  files a `lock_blocked_change` item (six rerenders in three days on one page).
- Widening the guard to the 190 undeclared components — refuse-on-suspicion, the 158-row mistake in
  new clothes.

## 9. Follow-ups filed, not folded in

- **A zero-byte `rendered_html` with a `component_id` passes `sectionIsUnresolvableStub`** (it
  requires `component_id IS NULL`). `idea.uk`'s homepage carries such a row. Its own finding, and it
  deserves its own census first.
- Whether `mortgagecalculator.co.uk`'s ten tool pages should be `owned` rather than `generic`.
- Whether `vetcomparison.uk`'s homepage *should* serve an embedded tool in slot 1 — it does today,
  and the repair preserves that truthfully rather than deciding it.
