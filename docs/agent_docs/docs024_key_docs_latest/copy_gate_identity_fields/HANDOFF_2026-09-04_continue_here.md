# HANDOFF 2026-09-04 — copy gate: `name` fields, and the identity/display split

**Lane:** `copy_gate_identity_fields` (path:
`docs/agent_docs/docs024_key_docs_latest/copy_gate_identity_fields/`).
**Bug:** `bugs_open/420_HANDOFF_2026-08-31_the_negation_gates_prose_walker_skips_name_fields_so_heading_tics_ship_unscanned.md`
— **resolve 420 BY SLUG**, the number collides with the delivery lane's billing-email 420.
**Commit:** `60091e140` · **Council:** `3e9e8ce8-fb9b-4f5b-a610-016b57427a27` — **round 1 REVISE
(11:22Z); round 2 submitted 11:29Z and WITHHELD, NOT REVIEWED — see action 1** ·
**Register:** CQ-037 · **Started from:** the session named `420 425`.

---

## THE TOP THREE ACTIONS, in order

1. **READ THE ROUND-2 COUNCIL VERDICT — it is the only thing actually owed.** The code is already
   on the shared branch; a REVISE or REJECTED has to be acted on, not filed.

   **⚠ Round 1 was REVISE and the objection was CORRECT — read it before assuming round 2 is a
   formality.** `editquality` found that the plan claimed "add `name` to `headlineFieldRe`" while
   containing no edit showing that regex change. The change was in the committed code
   (`negation_content.go:253`); the PLAN omitted it. Implemented from that plan, `t.Headline` never
   becomes true for a `name` field — so the heading floor is never selected and the ordering fix
   guards a severity that never applies. **A fix that silently does two thirds of nothing**, which
   is this bug's own failure shape one level along. Round 2 makes it a standalone edit.
   **The lesson, if you submit anything from this lane: the plan is the artefact under review, not
   your working copy.** And **merging two edits to fit the 8-edit cap is refused server-side when
   they name different files** — one edit = one file, so merge same-file edits instead.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='3e9e8ce8-fb9b-4f5b-a610-016b57427a27' AND kind='council_report'
   ORDER BY created_at;
   SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
   ```
   ⚠ Budget ~30 minutes from submission, not 2 — the dispatch queues behind the fleet. A missing
   row is latency, not a drop. If APPROVED, nothing to do: `098` credits `60091e140` automatically
   off the `Council-Submitted:` trailer, and **do not** hand-write `Council-Reviewed:`.

   **⚠⚠ ROUND 2 WAS NEVER REVIEWED — IT WAS WITHHELD, AND NOT BECAUSE OF ANYTHING IN THIS LANE.**
   The run routed to `complete_invalid` with a spend-governor block reading *"WITHHELD at shed
   level 0 (32% of budget spent) - NOT queued; do not retry; re-trigger when
   governor_state.shed_level drops"*. **That remedy is unsatisfiable: `shed_level` is already 0,
   the floor.** And the governor's own decision in the same jsonb says `admitted: true` with
   `council-gate` mapped to `maintenance`, which sheds at L1 (70%) while spend is at 32% — so it
   should have run. **Estate-wide, not ours:** six distinct submissions from six lanes withheld
   between 11:21:58Z and 11:47:23Z, with the last approval at 11:09:35Z and the last revise at
   11:15:47Z — a clean onset. Reported to the `dispatch_throughput` lane, who own the governor
   (RFC_065 / migration 752), with the evidence; not filed as a bug to avoid competing with their
   record. **So: do NOT read the absent verdict as latency, and do NOT resubmit until they say it
   is fixed** — a resubmit now just mints another withheld run and another unresolvable trailer.
   Re-trigger `3e9e8ce8` when they confirm.
2. **AFTER THE NEXT CHASSIS ROLL, verify at the binary and then re-run the census.** Both commands
   are in `RUNBOOK_copy_gate_identity_fields.md`. The fix is INERT until a roll — `bugs_open/420`
   stays OPEN until then, per the fixed-AND-live bar.
3. **Do not "fix" the two red tests in `platform/orchestration/actions`.** They were already
   failing at HEAD before this work (`TestFindingCodeScanEveryWriteIsRegistered`,
   `TestTemplateExecutorsAreDeclared`, both about `renderFailWorkItemMessage`) and belong to
   another lane. Proven with a no-overlay control — see NOTES.

## WHAT THIS WAS, in one paragraph

The copy gate's walker skipped every `*.name` field by field name, so define-by-negation
constructions ("X, not Y") shipped unscanned on feature-card headings — the surface the code's own
comment calls the place the construction is "least forgivable" — while the sibling `description`
fields in the same array were repaired in the same run. `[MEASURED 2026-09-03]` 37 such values were
live across 15 domains and 23 pages.

## THE ONE THING TO UNDERSTAND BEFORE TOUCHING THIS CODE

**`name` means two opposite things depending on who wrote the item, and the discriminator is
structural, not lexical.**

| `url` sibling | `name` prose-shaped | n | contract |
|---|---|---|---|
| no `url` key | 752 of 825 | 825 | **DISPLAY** — directory/feature/tracker card headings |
| `url` key, non-empty | **0 of 908** | 908 | **IDENTITY** — the real `pages.name`; the item's own url is built from it |
| `url` key, empty/null | — | **0** | does not occur |

`[MEASURED 2026-09-03]`, zero crossover, reproduced independently by the components lane.
So **neither "skip `*.name`" nor "scan `*.name`" is right fleet-wide** — which is why the bug
file's own one-line fix candidate is unsafe and is left in place, struck, as the record.

⚠ **The identity slugs are skipped by the walker's value tests today BY LUCK, NOT BY PROTECTION.**
One tokeniser change from silent and estate-wide. Never let "the heuristic happens not to fire"
stand in for an exclusion.

## WHAT SHIPPED (`60091e140`, 7 files, 630 insertions)

- **Predicate split**, following the estate's own doctrine rather than inventing a seam —
  `markup_spans.go:63-74` and `resolve_internal_links_action.go:81-86` (`bugs_open/248`) both ruled
  this same question the same way; `runtime_fill.go:29-38` is the form copied.
  `isProseContentField` (SCAN, fails toward scanning) / `identityContentField` (OVERWRITE guard,
  fails toward exclusion).
- **Bare `name` left the never-prose list; the `_name` SUFFIX arm STAYS** — that arm is what still
  protects `company_name`, `current_page_name`, `cardN_client_name`, `*_author_name`, `tool_name`.
  Dropping it with the bare word would have been the quiet half of the mistake.
- **`dropped_name` in `AcceptNegationRewrite`** — the missing LOSE half. **In the judge, not the
  walker**: a filter is bypassable by any future caller that enumerates fields itself. Both
  mutating call sites inherit it.
- **`name` joined `headlineFieldRe`**, and the identical dual-purpose defect in `IsHeadlineField`
  is closed **by ORDER** — the identity exemption runs before the headline branch. A test pins it.
- **OWNER RULING 2026-09-03: a 2-word HEADING floor**, separate from the untouched 5-word sentence
  floor. Without it the sentence floor refuses 25 of 36 heading repairs as `gutted` — the fix would
  make them visible and then decline to repair them.

## WHAT IS OWED, AND WHAT IS EXPLICITLY NOT

**Owed:** the council verdict (1), the post-roll verification (2).

**NOT owed — decided, do not reopen without new information:**
- **The 37 existing instances are left to heal on rebuild. OWNER DECISION 2026-09-04.** No sweep.
- ⚠ **A `page_rerender` repairs NONE of them.** `page-rerender` has neither `rewrite_negations` nor
  `copy_gate_annotate`, and `rerender_page_sections_action.go:3-9` re-renders stored `content_data`
  "WITHOUT invoking the content writer (no LLM)". The defect lives in `content_data.name`. Only a
  `page-content-writer` rebuild regenerates it. **I offered a rerender as an option before checking
  this and the owner chose it** — see `WRONG_CALLS.md` 2026-09-04.
- ~~**`sectionAssetKeyLike` is NOT fixed here** and it is the dangerous member of this class…~~
  **⚠ CORRECTED 2026-09-04 — REFUTED at the code, do NOT go and "fix" it.** That delete groups by
  `SectionIdentityKey(slot, RAW blob)` (`remove_duplicate_page_sections_action.go:153`), not by the
  normalised text; the text is only an 80-char eligibility gate (`:148`), so widening the shared
  list yields FEWER deletions and cannot make two raw blobs collide. The consumer it actually moves
  is the READ-ONLY detector (`check_content_duplication.go:658`). **It is the estate's worked
  MITIGATION of this class, not an open hole** — the destructive path was given its own identity
  predicate after a near-miss (`section_text.go:105-124`), which is the shape to copy, and tuning
  that shared list is the one change its header warns against. I asserted the original from a
  subagent report without reading the deciding lines; see `WRONG_CALLS.md` 2026-09-04.

## WHAT TO EXPECT IN THE INSTRUMENTS AFTER THE ROLL

- `hits_before` rises on card sections.
- `exempt_reasons.identity_name_with_url` appears wherever a writer emitted listing-shaped items.
- **`dropped_name` in the rejection log is the guard working, not a fault.**
- ⚠ **Do not read a low rewrite count as "the fix did nothing" — read the rejection reasons.** The
  copy lane's `gutted` rework (`7cc16a5d0`) makes short repairs more likely to be refused; the
  heading floor is what stops that swallowing this fix.
- **Control that must NOT move:** nightly `brief-negation-check` (`40 7 * * *`), baseline
  **11 of 39** at 2026-09-03 07:41Z. Recorded N-of-M because the denominator grows as sites are
  added. That check does **not** share this walker (the bug file's scope note claimed it did and is
  struck through), so movement means something else changed.

## PEERS — who knows what, so you do not re-ask

- **`copy_quality_two_stage`** — owns `bugs_open/420`, handed it over explicitly. They confirmed
  there was no `dropped_name` and that the never-prose list was the sole protection. They also
  corrected their own scope note in `87780485a` after I refuted it. They asked for the nightly
  before/after pair as a **control**; still owed once a roll lands.
- **`components`** (`bugs_open/425`) — found the two-contract collision. Their deck class closed
  17/17 on 2026-09-03. They declined ownership of the register entry and asked that it be mine,
  crediting 425 in `sources`; done as CQ-037.
- ⚠ **`gamesdesign.co.uk` (in the affected list) is NOT `gamedesign.uk`.** Different sites, both in
  the same afternoon's cross-lane traffic. Easy to route work at the wrong one.

## WHERE THINGS ARE

`RUNBOOK_copy_gate_identity_fields.md` — every query and command, with its gotcha.
`NOTES_copy_gate_identity_fields.md` — the technical log, including four missteps.
`README_where_we_are.md` — the plain-prose account for the owner.
`bugs_open/420…skips_name_fields…md` — §RESOLUTION is the authoritative record.
Approved plan: `/home/ant/.claude/plans/zany-swimming-ritchie.md`.
