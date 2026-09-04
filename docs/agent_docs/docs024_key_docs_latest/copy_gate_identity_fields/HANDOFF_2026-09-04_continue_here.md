# HANDOFF 2026-09-04 — copy gate: `name` fields, and the identity/display split

**Lane:** `copy_gate_identity_fields` (path:
`docs/agent_docs/docs024_key_docs_latest/copy_gate_identity_fields/`).
**Bug:** `bugs_open/420_HANDOFF_2026-08-31_the_negation_gates_prose_walker_skips_name_fields_so_heading_tics_ship_unscanned.md`
— **resolve 420 BY SLUG**, the number collides with the delivery lane's billing-email 420.
**Commit:** `60091e140` · **Council:** `3e9e8ce8-fb9b-4f5b-a610-016b57427a27` (submitted, verdict
pending) · **Register:** CQ-037 · **Started from:** the session named `420 425`.

---

## THE TOP THREE ACTIONS, in order

1. **READ THE COUNCIL VERDICT — it is the only thing actually owed.** The code is already on the
   shared branch; a REVISE or REJECTED has to be acted on, not filed.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='3e9e8ce8-fb9b-4f5b-a610-016b57427a27' AND kind='council_report'
   ORDER BY created_at;
   SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
   ```
   ⚠ Budget ~30 minutes from submission, not 2 — the dispatch queues behind the fleet. A missing
   row is latency, not a drop. If APPROVED, nothing to do: `098` credits `60091e140` automatically
   off the `Council-Submitted:` trailer, and **do not** hand-write `Council-Reviewed:`.
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
- **`sectionAssetKeyLike` is NOT fixed here** and it is the dangerous member of this class:
  `section_text.go:45`, shared between a read-only duplication detector and
  `remove_duplicate_page_sections_action.go:297`, which executes a `DELETE`. Worth its own bug file.

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
