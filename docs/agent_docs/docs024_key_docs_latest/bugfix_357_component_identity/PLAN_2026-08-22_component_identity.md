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

---

# 10. OPTION 1 BUILD — owner ruled `RFC_046` on 2026-08-22, and the stamp already has a column

> Owner: *"Option 1 please, we can change the existing pages once option 1 has been built."*

## 10.1 The finding that shapes this: the machinery is already there and has never been wired

| | measured 2026-08-22 |
|---|---|
| `page_components.component_version_id` populated | **0 of 1,930** |
| Go code that WRITES it | **none** — the single non-test hit is a comment about a *different* table |
| Go code that READS it | **none** |
| `component_versions` rows | **369** across **203** components, 2026-04-20 → 2026-08-22 (live, growing) |
| `component_versions.html_template` present | **369 of 369** — it stores the FULL template text, plus `input_schema` |
| control (`rendered_html_digest`, a live column) | 1,623 of 1,930 populated, 65 code references |

> ⚠ **Trap for anyone re-running this: `component_version_id` is a column on TWO tables.**
> `site_plan_sections.component_version_id` is the live one and it accounts for nearly every grep
> hit. A bare `grep -rn component_version_id` conflates them and makes the dormant column look busy.
> Qualify by table.

**So Option 1 does not need a new column, a new table, or a new concept.** It needs the existing
purpose-built link to be written. That answers the `reuse_agent` seat before it asks, and it is why
this round is reuse rather than addition.

**It also serves `bugs_closed/277`'s residual.** 277 closed with 15 rows permanently unrepairable
because *"`component_versions` holds zero rows for any of the nine components involved"* — the
template that produced their bytes no longer exists. Had rows carried their version, that class
would be recoverable. Wiring this does not fix those 15, but it stops the class accruing.

## 10.2 Why `component_versions` alone is not already the answer

> ⚠ **CORRECTED 2026-08-22 (council round 3, `reuse_agent` seat) — the "one path" claim below is
> FALSE and I published it here, in register CLC-026 and in two council submissions.**
> `component_versions` has **five INSERT sites across four files as of 2026-08-22**:
> `fix_component_template_action.go` (×2, `change_source='repair_template_slots'`),
> `store_generated_component_action.go` (`snapshotComponentVersion`),
> `update_component_html_action.go` (`changed_by='update_component_html'`), and this lane's new one
> (`change_source='render_stamp'`). I ran the grep, read the first file's two hits, and stopped.
> Live `change_source` values confirm the plurality: 26 distinct values including
> `repair_template_slots`, `manual-restore` and a scatter of bare correlation ids.
> **What survives the correction, and is now better evidenced:** the reason `bugs_closed/277`'s nine
> components had zero versions is NOT that a single Go writer missed them — it is that a template
> edited by a **hand-written SQL migration** passes through no Go writer at all, so none of the four
> sees it. That is a stronger statement than the one it replaces, and it is why the stamp has to be
> taken at RENDER time rather than at edit time. Logged in `WRONG_CALLS.md`.
> **Also corrected:** each writer allocates `version_number` with its own `MAX+1`, so the race the
> `editquality` seat raised about my get-or-create exists BETWEEN writers too, not only within mine.

~~It is written by **one** path — `fix_component_template_action.go:1127`, as a pre-repair snapshot.~~
So a template edited by any other route is never versioned, which is precisely how 277's nine
components ended up with zero versions. **A stamp pointing at a version table that does not capture
every template change is a stamp that goes NULL exactly when it matters.** The build therefore has
two halves, and the second is not optional:

- **(a) stamp the link** — record which version produced these bytes;
- **(b) guarantee the version exists at render time** — resolve the rendered template text to a
  version row, creating one if that exact text is not already the component's newest.

## 10.3 Where it goes: the seam reports, the writer persists

`RenderTemplate` (`component_library.go:1007`) is the estate's ONE render spelling — **15** non-test
call sites as of 2026-08-22 (owner ruling 2026-08-21; the count matches `bugs_open/342`'s). It has
no DB handle and no site identity, which is exactly why `AbsentRequiredFields` is an **out-field**
that callers read rather than something the seam acts on. Follow that division precisely:

- `RenderContext` gains an input `ComponentID` (set by any caller that has the component in hand —
  NIL MEANS UNKNOWN, never inferred) and an **output** `RenderedTemplateSHA`, which the seam sets to
  the digest of the template text it actually executed. The seam reports a fact only it knows; it
  performs no I/O.
- `save_page_sections` — which already has the DB handle, and is the single INSERT every composition
  path flows through — resolves `(ComponentID, RenderedTemplateSHA)` to a `component_versions` row,
  creating one when that text is not the newest, and writes `component_version_id` in the **same
  statement** as the bytes, on the `rendered_html_digest` / IMP-052 precedent.
- **Carry paths carry the stamp**: `carryStoredSection` already forwards `rendered_html`,
  `stored_slot_name`, `component_id` and `content_data` as one bundle (:1150–1169); the stamp joins
  that bundle. A carried row keeps the version that produced its bytes, which is the whole point —
  the bytes did not change, so neither does their provenance.

## 10.4 Phasing, which is the direct answer to the council's two rejections

| phase | what | why it is safe / effective |
|---|---|---|
| **1 (this round)** | write the stamp | **Provably inert.** Nothing reads `component_version_id` — 0 writers, 0 readers, measured. Writing a column nobody reads cannot change what any page serves, which is what round 1 failed on |
| **2 (separate round)** | readers stop guessing: `enrichSectionsWithPlannedNames` may not name an unstamped section from `planned[Position-1]` | This is where behaviour changes and where the authority gate belongs. It is what round 2 was rejected for lacking, and it is only defensible once the stamp exists to read |
| **3 (owner-authorised, after 2)** | repair the 22 | Ruled in principle 2026-08-22. Still blocked on the slot-name matching hazard (§7 / LANDMINES), which phase 2 must resolve — not this round |

**Phase 1 does not fix `bugs_open/357`, and the submission must say so in those words.** It builds
the thing the fix is made of. Claiming otherwise is precisely the "detected but never blocked"
dishonesty the council caught in round 2.

## 10.5 Verification — at the artefact, disconfirming result named

- **After the roll:** rows created after the roll carry a non-NULL `component_version_id`, while
  rows created before it stay NULL (**the control** — a migration or backfill would break it, and
  none is proposed). *Disconfirming:* still 0 after the roll ⇒ the seam is not reporting or the
  writer is not persisting; the count going up on OLD rows ⇒ something is backfilling, which this
  design does not do.
- **The stamp is true, not merely present:** for a stamped canary row, the `html_template` at the
  stamped version must re-render to the stored bytes **byte-identically** — the same round-trip gate
  `cmd/content-data-recover` (CQ-029) already uses. *Disconfirming:* a byte difference means the
  stamp names a template that did not produce those bytes, which is worse than no stamp.
- **Version churn is bounded:** `component_versions` should gain roughly one row per component whose
  template text is not already its newest, then settle. *Disconfirming:* row count climbing on every
  render ⇒ the resolve is comparing wrongly and minting a version per call.

---

# 11. CORRECTIONS 2026-08-23 — three things §10 asserted that turned out to be false or incomplete

Recorded here rather than edited into §10, because §10 is what we believed when phase 1 shipped and
the difference is the useful part.

## 11.1 ⚠ "Carry paths carry the stamp" was ONE-THIRD true

§10.3 says *"Carry paths carry the stamp"* and lists `carryStoredSection`. That path was wired
correctly. **Two others were not:**

- `RerenderPageSectionsAction`'s **fresh-render** entry (`rerender_page_sections_action.go:746`) never
  emitted `rendered_template_sha` at all, though `:662` calls `RenderTemplate` and holds the digest
  85 lines earlier. `bbe178309` touched only the carry half of that file.
- **`extractSectionFromMap` dropped it outright** — the compile hop rebuilds the metadata map from a
  hand-written six-key allow-list. This is why the stamp wrote **0 of 820** rows in its first day
  live. §10.5's own disconfirming result ("still 0 after the roll") fired and nobody was looking.

The plan verified the **two ends** of the chain and asserted the middle. The check that settles it
measures the middle: 546 live `sections_metadata` elements, **0** with the digest, **546** with
`component_id` as the demand control.

## 11.2 ⚠ Phase 1 was not merely inert — once un-severed it would have written FALSE stamps

§10.4 calls phase 1 *"provably inert"*, and while the digest never arrived that was true. It is not a
property of the code; it is a property of the breakage. Layer 2's splice
(`save_page_sections_action.go:532`) replaces a section's HTML with stored interactive bytes while
leaving the incoming section's `RenderedTemplateSHA` — the digest of the *discarded* render — in
place. Deliver the digest and `resolveComponentVersionID` resolves it against `hero` and stamps the
hero version onto a whole interactive tool. **"Worse than no stamp", by §10.5's own standard, on
exactly the population this lane exists for.** Fixed by `adoptCarriedProvenance` at both carry arms
(phase 0, `a2e2fbac2`).

Generalised: **before repairing a severed field, enumerate every place its payload is REPLACED
without it.** A missing fact becomes a false one at the moment you fix the plumbing.

## 11.3 §10.4's phase 2 ("readers stop guessing") is wrong as written, and Layer A inherits the landmine

§10.4 states phase 2 as *"`enrichSectionsWithPlannedNames` may not name an unstamped section from
`planned[Position-1]`"*. Two problems:

1. **Unstamped is the common, honest case** — the whole HTML-parse path, plus 190 of 339 components
   that declare no `data-component`. As written this deletes positional naming wholesale, which §8 of
   this same plan explicitly declines to do.
2. **Suppressing the slot rename desynchronises the row from `pages.sections`** and arms the
   carry-forward landmine this lane itself filed. §4's Layer A has the same defect for the same
   reason: an `adopted-fragment`-*named* slot on a page whose plan lists `hero` first re-appends on
   every rebuild.

**The correct cut, and it is the one that makes safe and effective the same plan:** slot name and
component identity are different facts. The damage is entirely on the **component** axis
(`component_id` → hero's schema, hero's template, `ContentDataCanFillTemplate`); the landmine is
entirely on the **slot** axis (Layer 2's match key; **420** Go references as of 2026-08-23). So type
the component and **never touch the slot name or `pages.sections`** — the landmine is not managed, it
is never armed. Constructive adoption onto a `{{.body}}` component, verified byte-identical by a real
`RenderTemplate` round trip (`text/template`, so no escaping — checked), earns a genuine stamp
through existing machinery and needs no new inference.

## 11.4 And Layer B should not be built

§4's Layer B (a `data-component` verification guard at the INSERT) is the sixth inference RFC_046
declines by name, is silent for 190 of 339 components, and cannot vouch for the repair's own output
(`{{.body}}` declares no attribute). With the stamp now actually flowing, it answers the same
question as fact. Layer D (an at-rest detector for the other five `page_components` writers) keeps
its value, as a separate lane.

---

## DECISION 2026-09-02 (OWNER): Phase 3 ships as OPTION B — per-tool adoption, not 578's shared retype

The owner ruled between two repair shapes for the 22 (both fully specified in NOTES
2026-09-02): **Option B — the lendzy-693 shape transferred** — each row gets its own
content_components row whose html_template IS the stored bytes (created_from='adopted',
component_level='tool'), with the PLAN repointed in the same transaction at BOTH levels
(site_plan_sections.component_name for durability — pages.sections is a derived copy that
sync overwrites — and the pages.sections element for immediate effect).

**Reason:** 578's rebuild-safety rests on the armed Layer 2 identity-carry, which is exactly
what precondition 4 exists to prove and could not be tested (the cv1 canary is structurally
the wrong vehicle — 26b handoff Finding 2). Option B makes regeneration safe BY CONSTRUCTION
— template = bytes, plan no longer names hero — removing the dependency instead of testing
through it, which is this estate's own fix-ranking rule. 578 stays on disk untouched until B
is applied and verified, then gets a SUPERSEDED marker.

**Pre-design measurements (all 2026-09-02, this session):** drift reconciler compares
built_from_plan_version (a site_plans.id) to the current plan id — plan-element edits create
no plan row, no drift storm; all 22 bodies binding-free (zero '{{'); all 22 pass
toolTemplateValid through the real function (both-way controls); exactly one hero row at
ordering 0 per page in each site's single current plan; ONE function collision
(tool-equity-release vs library row tool-equity-release_pre_037 a5236dec) → that one adopted
row forks per RFC_036 §9.3; section resolution tries stored component_id FIRST, so repointed
rows resolve directly. Pilot: mortgagecalculator/tool-simple first, then the remainder.
Migration: **700_retype_357_population_by_adoption_HOLD.sql** (by hand, never the runner).
