# HANDOFF — bugs_open/445, the layout library gap — 2026-09-03b (evening) — continue here

**Supersedes `HANDOFF_2026-09-03_continue_here.md`.** That file's §1 liveness table is now stale in
two of three rows and its prediction 1 has been RETIRED by evidence. Read this, then
`README_where_we_are.md` (plain prose, owner's), then `NOTES_layout_fit.md` §(l) and §(m) — the
missteps and the disconfirmation.

**Everything shipped is live. Nothing is blocked. The open work is a sweep, a harness and a guard —
plus one finding that changes what this lane should be measuring.**

---

## 0. Read this paragraph if you read nothing else

445 said the layout library lacks a "content hub with embedded tools" archetype. True, and a
symptom. The mechanism: **the estate could not SEE a library gap of that shape, by construction** —
the growth signal fired only when a layout's TOTAL score was zero library-wide, while the
category/description/scheme bonuses are added independently of tag matching, so a layout matching
NONE of a site's tags still scored above zero. All of that is now fixed and running. **But the first
live classification (16:54Z tonight) showed that the metric I built to replace it — tag coverage —
can be satisfied by a site being typed in BORROWED library vocabulary rather than accurate
vocabulary.** Coverage went 13% → 100% on a site whose tags describe the wrong business. **See §5.
It is the most important thing on this page.**

## 1. LIVENESS — all green. Verified this session; re-verify only if something looks wrong.

| what | state | how it was established |
|---|---|---|
| **Phase 1 Go fix (`76db94fc7`)** | **LIVE** in `v1.0.1359` (pods `85c4984f77-nrqf7` / `-phgh2`, started 13:28:18Z / 13:28:43Z) | Binary probe, **both replicas, both controls**: `weak_tag_fit` 1, `layout_match_score` 1, `enforceListingItemSources` 2 (must-be-present), `zzq_literal_that_cannot_exist_4417` 0 (must-be-absent). On `v1.0.1358` the first two were **0** — that contrast is what makes it a measurement. `git merge-base --is-ancestor 76db94fc7 HEAD` → yes. |
| **Migration 735** (prompt honesty) | **LIVE and now EXERCISED** | Anchor phrase absent from the live row (`position(...) = 0`), **and** confirmed absent from the rendered prompt of the 16:54:07Z run. |
| **Migration 736** (the archetype) | **LIVE**, 0 sites on it | `is_active=t \| light \| 9 tags`. Its tags (`content-hub`, `editorial-guides`) confirmed present in the 16:54Z rendered prompt. |
| **734's surviving half** (peer's) | **LIVE and EXERCISED** | `input_fields` still names `layout_taxonomy`; tag-list region rendered a real JSON array, not `null`. |

⚠ **The `build provenance` recipe FAILED here, in the direction that gives a false answer.** Two
causes stacked: the startup line had rotated (pods 3h old — empty means *out of range*, never
*unstamped*), **and** the unstructured `grep 'build provenance'` returned a HIT that was
`LANDMINES.md` prose about build provenance, synced to `doc_notes`, injected into a prompt and
logged by the chassis. Use RUNBOOK **r10** — structural grep first, then the binary probe. Do not
re-derive this.

## 2. COUNCIL — one approved, one was never produced, one is in flight

| submission | correlation | state |
|---|---|---|
| Phase 5, migration 736 | `39942a14-…` | **APPROVED** round 1, "2 advisory objection(s) — none high-severity". Read in full. |
| Phase 1, commit `76db94fc7` | `34d57f60-…` | **DEAD — no verdict ever existed.** Run stalled at `council_decide` (last activity 12:05:46Z), killed by the `v1.0.1358` roll (pods up 12:06:47Z), swept `FAILED` at 16:07:33Z. `fix_plan` artifact only; **no `council_report`**. |
| Phase 1, **RESUBMITTED** | **`adfa4d03-67a8-419f-bc22-d0ef125f94ee`** | **In flight — READ IT.** Fired ~16:45Z this session. |

**➤ FIRST ACTION FOR THE NEXT SESSION: read `adfa4d03`'s verdict.** RUNBOOK **r11**. If APPROVED,
nothing needs re-committing — but `098` resolves `76db94fc7`'s trailer against the *dead*
correlation `34d57f60`, which will never approve, so **the honest record is a short doc commit
carrying `Council-Reviewed: adfa4d03-67a8-419f-bc22-d0ef125f94ee`** once you have READ an approved
verdict. If REVISE, the code is already live fleet-wide — act on it rather than defer.

⚠ **The trap that cost this the most time, now a LANDMINE.** `098`'s coverage report reads **only**
`diagnosis_artifacts`; it never asks whether the run is alive. A dead run buckets as **AWAITING**
with `no report yet (queued, or evidence cleared)` — *for ever*. `[MEASURED 2026-09-03]` 18 AWAITING
rows tonight, 2 carrying that annotation: one dead (this lane's), one genuinely executing
(`dda64bd1`). **Same bucket, opposite truths.** Ask `orchestration_states`, not the report. The roll
took **seven** runs at once and left **three** orphaned across three lanes.

**Advisory objections on 736 — one refuted, two still owed.**
- **Refuted at the code:** editquality claimed the `editorial`→`editorial-publication`
  canonicalisation was "asserted, not verified". It is verified: `canonicalTag`
  (`fork_theme_composition.go:146`) maps it, and **both** sides canonicalise (site terms `:227`,
  layout tags `:278`). The reachability count stands. The seat said itself that `layouts`/
  `site_specs` were outside its schema. RUNBOOK **r12**.
- **Still owed, and they belong to Phase 4:** the verify block tests only the literal
  `--section-text:` rather than any `--section-*` default; and it never positively asserts the CSS
  covers `faq-item` / `featured-article__*` / `category-listing` / `article-card*`.

## 3. WHAT SHIPPED

| phase | artefact | state |
|---|---|---|
| 0 | migration **735** — classifier prompt: promise removed, layout names dropped from tag examples, FORM-over-industry sentence added | applied, verified at the row **and at a rendered prompt** |
| 1 | commit **`76db94fc7`** — `layoutFit`, `TagCoverage`, `LibraryGap()` with a weak-fit arm at `lmMinTagCoverage = 0.50`, persisted as `lineage.layout_match_score` + `lineage.layout_fit` | **LIVE in `v1.0.1359`** |
| 5 | migration **736** — layout `content-hub-tools` | applied, council APPROVED, **0 sites** |
| 4 | reachability guard | **partial** — inline DO block in 736 only |
| 2, 3 | `internal/cronchecks`, `cmd/layout-fit-check` | **not started** |

## 4. THE PROOF STILL OWED, and it is cheap

**No composition has run since the roll.** `count(*) WHERE data->'lineage' ? 'layout_match_score'`
is **0** fleet-wide, and zero `resolved_composition` rows have been written since 13:28Z. So:

> **The first post-roll `resolved_composition` row must carry `lineage.layout_match_score`.** That
> is the proof the fit evidence is real. The roll is proven at the binary; the evidence is not
> proven until a composition writes one.

copyonline is the site to watch — `portfolio_positioning` has promised `layout_name`,
`lineage.layout_source`, `layout_candidates`, the `reasoning` string and
`classification.industry_tags` the moment composition runs.

## 5. ⚠ THE FINDING THAT CHANGES WHAT THIS LANE MEASURES — read before touching the threshold

**The real before/after boundary is `2026-09-03 16:54:12Z`**, the first classification ever run with
the tag library rendered. (**NOT 11:39Z** — the peer retracted that; their register step was broken
and every classification between 11:39Z and 16:01Z failed before reaching the prompt. Verified
both halves myself at the live row.)

**Result: copyonline emitted 10 tags, 10 matchable — 100%, against a ~13% (28/216) baseline.** The
previous handoff's pre-registered prediction 1 came true.

**And the tags describe the wrong site:** `marketplace, directory, community-platform, b2b,
professional-services, content-platform, creative-agency, interactive-platform,
practitioner-platform, industry-hub`, `category=hub`. Copyonline is an editorial authority on
writing commercial copy, with tools; its directory is one page and a marketplace is an explicit
must-not. **Nothing says copywriting, editorial, guides or long-form.** `content-hub-tools` will not
win it.

**The peer's formulation, which is the keeper: _matchability is not accuracy, and coverage cannot
distinguish them — a site typed entirely in borrowed vocabulary scores 100%._**

Three consequences, all binding on the next session:

1. **`TagCoverage` measures overlap with the library, not correctness of the classification.** So
   the weak-fit arm will stay **SILENT** on copyonline — a site the library genuinely serves badly.
   **The detector's first live encounter with a mis-typed site is a miss, by construction, not by
   threshold.** Do not "fix" this by lowering 0.50; the metric cannot see this failure at any cut.
2. **PREDICTION 1 IS RETIRED as a threshold test.** Re-deriving the 0.50 cut on post-16:54 data
   fits the metric to itself, because the population's vocabulary has shifted toward whatever the
   library happens to contain. It can still flag a *fall*. Any coverage figure must state which
   side of 16:54:12Z it sits.
3. **Do not treat n=1 as a rate.** The peer names a rival explanation neither lane can exclude:
   copyonline's brief leads with its marketplace heritage and its directory, so the classifier may
   be reading the brief correctly and the *brief* may be over-weighting. Testable on the next
   remake and on a re-classification of this one.

`[UNMEASURED]` **A risk in our own 735:** its form-over-industry sentence quotes four live library
tags as examples (`long-form`, `tool-portal`, `content-hub`, `comparison`), and the estate has a
standing trap that *a quoted exemplar in a prompt ships verbatim*. copyonline emitted none of the
four, so there is no evidence of copying yet. **Watch the next few classifications**; if exemplar
tags appear at rate, reword the sentence to describe the shape without naming instances.

## 6. WHAT IS NEXT, in order

1. **Read `adfa4d03`'s verdict** (§2, RUNBOOK r11).
2. **Watch for the first post-roll composition** and confirm `lineage.layout_match_score` is
   present (§4). One query; it is the last unproven link in Phase 1.
3. **Phase 2 — `internal/cronchecks`** (owner decision: before the sweep, answering the open
   `RFC_024` rather than adding a tenth un-harnessed cron check). ~120 lines: `DB()`, `Note()`,
   exit codes (0 clean / 1 findings / **2 refused, never a pass**), `SystemSiteID`, the
   `idx_swi_dedup` status predicate single-sourced with a test that parses migration 157, and a
   **schedule-collision ratchet test** (four collisions exist today: 06:50, 07:05, 07:25, 07:40).
   Adopt in the new check only; migrate `cmd/verifier-remit-check` later in its own measured
   commit; touch none of the other 17.
4. **Phase 3 — `cmd/layout-fit-check`** on that harness. Template `cmd/verifier-remit-check/`. Keys
   on `sites → style_collections → css_themes → layouts` (RUNBOOK r1), **NOT** on
   `resolved_composition` — `SelectStyleCollectionAction` (`v3_site_actions.go:67`) writes no
   lineage, and theme-kit sites never run the matcher. **Unit of a finding = cluster
   `(layout, exact matched-term set)`, not site**; digest the sorted term set into `item_key`;
   shelf at `system.internal`, `pipeline='maintenance'`, `status='deferred'`, **empty
   `handler_agent`** (migration 690's trigger); never `DO UPDATE`. Needs the scorer in Go →
   **extract `platform/orchestration/actions/layoutmatch`** first (24 references in 2 files; keep
   `resolveLayoutByTags`'s SQL and `Reason` byte-identical). Birth-commit registries:
   `liveItemTypes` + `itemTypesWithoutVerifiers`, `RELEASE_IMAGES` + `AGENT_DEPLOY_SERVICES`,
   `council-scope.sh` + `098`'s `SCOPE_PATHS`.
   **⚠ In light of §5, this check must report the matched-term SET, not only the coverage number** —
   a 100% cluster built on borrowed vocabulary is exactly what a human needs to see, and a
   coverage-threshold check alone would filter it out.
5. **Phase 4 proper** — `assert_layout_reachable(p_layout, p_min_sites)` as a migration-guard
   function + a `scripts/pattern-check.py` rule requiring any `INSERT INTO layouts` migration to
   call it. Both council scope. **Fold in 736's two unanswered objections** (§2).
6. **Two theme-kit facts to carry:** kit-chosen sites record NO fit evidence (the short-circuit
   returns before the matcher — correct by design; the sweep is what scores them), and
   `soft-editorial` is the editorial category's reachable-only-by-kit layout.

## 7. TRAPS — read before touching anything

- **The `build provenance` recipe returns a false POSITIVE on the chassis** (landmine prose in the
  logs) and a false negative once the startup line rotates. RUNBOOK **r10**, both controls, every
  replica.
- **`098` cannot tell a dead council run from a queued one.** RUNBOOK **r11**. New LANDMINES entry.
- **A jsonb-path query on `orchestration_states` by correlation TIMES OUT.** Bound it with
  `updated_at > now() - interval '10 hours' AND owner_agent_type='council-gate'`.
- **A migration verified at the live row has NOT been exercised.** 735 sat applied-and-unrun for
  five hours; 734's broken half was invisible for four. **Fire one job and read
  `orchestration_states.current_step/status`** — the peer's own stated lesson, and it is now ours.
- **Any all-time count of work items must union `site_work_items_archive`**; `site_specs` does NOT
  archive (it versions in place under `is_current`).
- **`090` on a code-only symptom needs `SEED_SCOPE`** or it fails after ~6 minutes and burns its
  only attempt.
- **The tree may not build because of someone else's dirty file.** `scripts/verify-head-builds.sh
  --with <yours> … --test`; never stash. Pre-existing failure on clean HEAD:
  `TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` — not ours.
- **The same-scheme bonus alone makes any same-scheme layout "eligible"**, so the scheme-gap arm is
  unreachable while any same-scheme layout exists (why `soft-editorial` is the permanent 0.50
  runner-up). A test fixture assuming otherwise fails.
- **Prompt text is not the prompt.** Read `llm_call_log.prompt_rendered`, anchored on a LONG
  phrase — `'Current library tags'` alone matches the JSON schema example earlier in the same
  prompt.
- **Council submission schema:** `.plan` is an object `{summary, edits[≤8], grounded_in[]}`;
  `operation ∈ modify|add|remove|config_change`. **Submit first, commit naming the correlation.**

## 8. PEERS

| lane | state |
|---|---|
| `portfolio_positioning` | Retracted migration 734's register step (broken `$1` binding, no `params`); rolled it back, kept the `input_fields` half — **I verified both halves at the live row.** They found and reported the copyonline result in §5. Their round-2 council correlation `f0ad8366` was one of the three orphaned by the roll — **told them.** They owe the five-field composition read. |
| `bugs_open/453` lane | Six commits deep in the template-`<no value>` class today. **Contributed a measured variant into their bug file** (guard reads the parent, payload reads a child; 7 of 23 `mission_brief` specs lack `.text`; the emptied block is the one migration 464 uses to license a regulated business model). Not filed as a new number. Theirs to route. |
| `theme kits` | Told them the archetype is live and about `soft-editorial`'s kit-only reachability. |
| `designblog.co.uk` | Churn in flight. **Nothing may queue work against designblog without telling them first.** |
| `site design planner` | Filed 445; owns the composition resolvers. |
| whoever owns `45ae3ad3` | A `seed_analytics_default.go` council submission, round 2, orphaned by the same roll. **Not yet told — no lane identified.** |

## 9. FILES

`docs/agent_docs/docs024_key_docs_latest/bugfix_445_layout_fit/` — `PLAN_2026-09-03_layout_fit.md`,
`RUNBOOK_layout_fit.md` (**r1–r12**), `NOTES_layout_fit.md` (append-only; **§(l)** the roll and the
dead verdict, **§(m)** the canary and the disconfirmation), `README_where_we_are.md` (owner's),
two `COUNCIL_SUBMISSION_*.json`, this file.
Migrations: `sql_for_agents/735_*`, `736_*` (+ `_ROLLBACK` each; 736's DEACTIVATES, never deletes).
New LANDMINES entry: *"A council run KILLED BY A ROLL leaves `Council-Submitted:` resolving for ever
as queued"*.
Scratch (session-local, will not survive): `scratchpad/score.py`, `simulate2.py`, `sitetags.txt`,
`layouts_live19.txt`, `prompt_after.txt`.
**The scorer must become Go (`layoutmatch`) before anyone else can rely on it.**
