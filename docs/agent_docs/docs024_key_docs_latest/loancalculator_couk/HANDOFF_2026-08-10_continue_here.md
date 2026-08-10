# HANDOFF — loancalculator.co.uk · **the `bugs_open/227` PLATFORM thread** · continue here (written 2026-08-10 late morning)

**Supersedes `HANDOFF_2026-08-09_continue_here.md`** for the 227 job. That file stays worth
reading for its §1/§2 detail and its corrections, but everything it listed as owed is now
either done or listed below with its state.

> ⚠ **This directory has more than one live lane.** This file is **only** the
> `bugs_open/227` platform job. The site's COPY/VOICE thread is
> `HANDOFF_2026-08-09b_continue_here.md` — different job, not superseded by this one.

```
site         loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis      v1.0.1277 (pods rolled 2026-08-09 21:34Z; releases are whole-fleet, owner runs make release)
the site     DONE — 26/26 pages voice H, serving, calculators golden. NOT this file's business.
the job      bugs_open/227 — both defects now FIXED AND LIVE. One verification arm still owed.
```

---

## State in one paragraph

`bugs_open/227` had **two** defects. Both are now fixed, live, and config-only — no image,
no roll, and the council gate does not apply (DB config is refused client-side). **Defect 1**
(the planner's prompts hardcoded one site's diagnosis, so every other site got that site's
plan) was fixed by **migration 345**, applied 2026-08-09 ~11:50Z, and is **PROVEN LIVE in
both directions**. **Defect 2** (a plan was persisted `is_current` before the council voted,
and nothing demoted a vetoed one) was fixed by **migration 363**, applied 2026-08-10
~10:40Z on the owner's decision, and is **proven for the APPROVED arm only** — see "the one
thing owed".

---

## What is DONE, with the evidence

### Migration 345 — the brief becomes data (defect 1)

`docs/agent_docs/sql_for_agents/345_experience_planner_site_brief_becomes_data.sql`.
Adds a `load_brief` step reading the site's own brief from `doc_notes`
(`categories @> '["experience-brief"]'`), chains it ahead of `compose`, adds
`experience_brief` to compose's `input_fields`, rewrites `compose` generically, surgically
de-contaminates the four review/reframe prompts, and moves vonc's brief verbatim into the
new channel in the same transaction.

Case-insensitive census over the whole live row: **48 hits across five steps → 0.**

Proven with two real runs, opposite in both directions, keyed only on `subject_key`:

| run | corr | `no brief on file` hits | COALESCE fallback rendered | leaked | prompt |
|---|---|---|---|---|---|
| loancalculator `debt-difficulty-help` | `c3976aab-39e3-45b6-9996-7a40f90259f4` | 2 | TRUE | **FALSE** | 24,721 b |
| vonc `vonc-spark-game` (control) | `72f540d3-bb8f-47f7-948d-67cc747a9fc1` | 1 | FALSE | TRUE (correctly) | 70,427 b |

**Both councils returned `approved`** (2 advisory objections each, none high-severity) —
which retires the old finding-2 worry that de-contaminated seats might still object about a
missing feed or timer. They did not. The `debt-difficulty-help` plan of record is clean:
11,442 b, names loan and debt subjects, `body ~* 'provocation|gauntlet|arena|vonc|spark'`
**false** — the first non-vonc experience plan in this system's history that does not
describe vonc's pages.

Re-verified 2026-08-10 after a full rebuild and roll to **v1.0.1277**: still live, chain
intact, brief note intact (7,908 b).

### Migration 363 — persist only after approval (defect 2)

`docs/agent_docs/sql_for_agents/363_experience_plan_persists_only_after_the_council_approves.sql`
(+ `_ROLLBACK`). **Owner decision 2026-08-10: the config-only rewire (route a), not the
`write_doc_plan` Go seam (route b).** Dry-run proven first (guards + verify passed, rolled
back, no trace), then applied.

Six edges. `compose` / `recompose` / `reframe` now hand to `review_journeys` instead of
`persist_plan`; `check_approved.then_step` becomes `persist_plan`; `persist_plan.next_step`
becomes `complete`; and `complete_escalated.config.output_fields` drops `plan_persisted`.
So **persist is reachable only from the approved branch**, and the migration's verify block
asserts that property directly — 0 steps other than `check_approved` reach `persist_plan` —
rather than merely asserting the fields were written, because a `jsonb_set` on a wrong path
silently ADDS a key instead of failing.

The graph now:

```
load_context -> load_schema_hint -> load_brief -> compose ─┐
                                                          v
   ┌──────────────────────────────────────────> review_journeys -> review_feasibility
   │                                                -> review_honesty -> review_mvp
   │                                                -> review_contracts -> council_decide
   │                                                -> append_council_note -> check_approved
   │   approved            -> persist_plan -> complete
   │   rejected            -> check_reframe -> reframe ──┘ (once)  | else complete_escalated
   │   revise, rounds left -> check_revise  -> run_checks -> recompose ──┘
   │   exhausted           -> complete_escalated      NOTHING PERSISTED
   └── (reframe / recompose re-enter the review chain)
```

**The fact that had to be checked before moving the write, and was:** persist reads
`plan_body_field: proposal.result`, and `compose`, `recompose` **and** `reframe` all declare
the same `output_field: proposal` — so the latest composition is what gets persisted. Verified
against both 08-09 runs rather than assumed: `length(collected_data->'proposal'->>'result')`
equals the length of the plan each run actually persisted (11,442 and 13,840). Had `recompose`
written to its own field, this rewire would have silently persisted the FIRST draft on every
revise round, which is a wrong-content failure, not a missing row.

**The deliberate loss, stated because it is a behaviour change:** `complete_escalated` no
longer lists `plan_persisted`, because nothing is persisted on that path. The escalated plan
is not lost — it is in the run's own `collected_data->'proposal'->>'result'` and in
`llm_call_log`, keyed by correlation id. Persisting an escalated plan *not-current* needs
route (b). Rejected alternative: persisting under a derived `subject_key` like
`'<key>:escalated'` — invents a convention nothing else knows to read.

---

## ⚠ THE ONE THING OWED — and do not let an approved run stand in for it

**363's whole purpose is the REJECTED arm, and that arm has not been observed.** Both 08-09
runs were approved, and a veto cannot be induced on demand.

- **DONE for the approved arm, but NOT by the check I predicted.** Run
  **`9150dd54-6129-464b-8600-771e0a84408a`** (fired 10:44:47Z) came back **COMPLETED /
  approved**, plan rows **5 → 6, one current**, new plan `051af223` 10,075 b and clean
  (`leaked=false`, so 345 is still holding too).
  > **⚠ CORRECTED same day — the row-count signal I wrote here and into 363's header is
  > NON-DISCRIMINATING for that run, and I nearly reported it as the proof.** It only
  > discriminates if the run takes **two or more** compose rounds: an approved run that
  > used to write N rows for N rounds now writes one. **This run was approved on round 1**
  > (compose ×1, no recompose, no reframe), and a single-round run writes exactly one row
  > under the OLD graph too. Same answer either way. **Check the round count before reading
  > the row count:**
  > ```sql
  > SELECT step_name, count(*) FROM llm_call_log WHERE correlation_id='<CID>'
  >   AND step_name IN ('compose','recompose','reframe') GROUP BY 1;
  > ```
  >
  > **What actually proved it — the ORDERING, which discriminates on any run.** The old edge
  > was `compose → persist_plan → review_journeys`, so under the old graph a row exists by
  > the time the run reaches any review step. Sampled mid-flight, the run was
  > `EXECUTING_STEP|review_journeys` with the count **still at the pre-run baseline of 5** —
  > it had passed the point where it used to persist and had written nothing. That is the
  > observation to repeat; it needs no multi-round run and no veto.
- **Still owed:** the rejected/escalated arm. Either wait for a natural veto, or seed a
  deliberately unbuildable experience and assert **no new `doc_plans` row** for it while the
  run ends `complete_refused` / `complete_escalated`.
- **Why this warning is in capitals:** this same bug already produced one
  check-that-cannot-fail (below). "The approved run wrote one row" does **not** demonstrate
  that a *vetoed* run writes none — it demonstrates the happy path. Record 363 as proven for
  the approved arm and **owed** for the other, in those words.

---

## The trap this lane walked into, so nobody repeats it

**The verification prescribed for 345 in three separate documents could not come out false,
and I ran it before noticing.** `prompt_rendered LIKE '%no brief on file%'` was the check.
That phrase is *also* in the static `compose` template 345 installs — the instruction
covering the no-brief case — so it is TRUE on every run of every site, **including one where
`load_brief` was never wired**, which is precisely the silent failure it existed to catch.

It came out TRUE for the vonc control, where all three documents demand FALSE, and **the
first reading of that is "the fix is broken" when it is the check that is broken.**

- Disconfirmable forms: the **count** (2 = template + rendered fallback, 1 = template only),
  or the substring only the `COALESCE` emits.
- **Better than either — skip sentinel reasoning entirely:**
  `collected_data->'experience_brief'->>'text'` on the orchestration row shows what the
  loader actually returned. vonc's opens with its 2026-07-17 diagnosis, 7,908 b;
  loancalculator's is the sentinel. That query is what I would run first next time.
- Corrected in `bugs_open/227`, in 345's VERIFY header and in the 08-09 handoff; logged to
  `WRONG_CALLS.md`; filed as a landmine on `llm_call_log.prompt_rendered` (**a rendered
  prompt always contains its own template**).

---

## Two facts about this estate that came out of the job

1. **A roll does NOT replay seeds over `default_config` — config-only fixes survive a
   rebuild.** Measured, not assumed: a fleet-wide column-by-column fingerprint diff across
   the v1.0.1277 roll gave `image_tag` **189** rows changed, `updated_at` **189**,
   `default_config` **4** (all attributable to other lanes' migrations — `domain-strategist`,
   `image-build-handler`, `quality-discovery-agent`, `section-editor`), `usage_count` **0**.
   A seed replay would have changed ~190 configs, not 4.
2. **The "something mass-rewrites agent definitions and leaves no trail" mystery in the 08-09
   handoff is SOLVED: it is the deploy stamping `image_tag`** —
   `scripts/deploy/update-agent-images.sh`, described in
   `platform/orchestration/actions/agent_image.go:21-24` and owned by `bugs_open/066`. A
   spawned agent pod takes its image from its `agent_definitions` row, so the deploy syncs
   those rows. No `schema_migrations` row because it is not a migration. It fires ~40–60s
   before the chassis pods report started.
   **And the reading trap:** `now()` is transaction start time, so rows sharing one
   microsecond mark one **transaction**, not one statement — which is what made a deploy sync
   plus concurrent migrations look like a single 188-row fleet rewrite. **A fresh `updated_at`
   on an agent row after a roll is not evidence your config changed** — check the fields.

---

## Housekeeping already done — do not redo

- **`bugs_open/227` stays in `bugs_open/`** (owner direction 08-06). Its top correction block
  and its "how to verify" section are both current as of 08-10.
- **Concept register:** the new channel is registered as **DOC-076** in
  `docs026_concept_register/register/documentation-system.md`, with an index row in
  `000_concept_index.md`. Do not add a concept count — counts were retired 08-09.
  **363 is NOT separately registered**, deliberately: it rewires one agent's own graph and
  adds no callable mechanism. If route (b) is ever built, *that* owes a register entry and a
  council round.
- **Work items:** both `needs_experience_plan` intake rows were closed so `idx_swi_dedup`
  releases the keys (`debt-difficulty-help` → `complete`; `vonc-spark-game` → `cancelled`,
  it was a test artefact). Run `9150dd54` will have created a fresh `debt-difficulty-help`
  row — close it when the verification above is read.
- **vonc's plan of record was restored by hand** to `b6fdbc09` (2026-07-25) after the
  positive control displaced it; the new council-approved plan `2ec02a7e` is kept and demoted
  with the reason in its `notes` column, should the vonc lane want to promote it. The restore
  held across the roll. **Running the planner mutates the plan of record — there is no
  dry-run mode.** That is worth remembering before firing it to "just check something", and
  it is a gap route (b) would also close.
- Commits: `e42000bb0` (345 live + the inert-assertion correction), `62a7a91e8` (post-roll
  verification + the bulk-writer identification), `255783c21` (363).

## If you are starting cold

Read this file, then `bugs_open/227` (its top correction block first), then 363's header —
which carries the design, the checked assumption, and the stated loss. `NOTES_loancalculator_couk.md`
has the full session record including the missteps; `README_where_we_are.md` has the owner's
plain-prose version.
