> **⚠ SUPERSEDED the same day by `HANDOFF_2026-08-25b_continue_here.md`.** Written before the owner ruled on the four open decisions; candidate 4 is now BUILT and a verifier is WRITTEN, so §4 of this file is out of date. Kept for its §5 traps and §2 verification method, both still current. Read the `b` file first.

# HANDOFF — continue here: `bugs_open/375`, after the fix shipped and rolled

**Written 2026-08-25** by the lane's first working session (took it on 2026-08-24 from
`HANDOFF_2026-08-24_start_here.md`, which is now HISTORY — read this file instead, and that one
only for its §5 traps).

> **In one line:** the gate is BUILT, council-APPROVED, LIVE on `v1.0.1337`, and **inert by
> design** — and the bug is **still OPEN and should be**, because a mechanism existing is not the
> defect being gone. Everything below §4 is what is left.

---

## 1. The one thing to understand before you touch anything

A **verifier** re-runs a defect's own predicate just before a work item is stamped `complete`, and
refuses the stamp if the defect is still there. There are **three** writers of `complete` and they
have never agreed:

| writer | consults the verifier? |
|---|---|
| `CompleteWorkItemAction` | yes, always |
| `UpdateWorkItemStatusAction` | **never did.** Since 2026-08-24: only when that STEP sets `verify_before_complete: true`. **No step does.** |
| the `claimed-item-timeout` sweep | **no** — writes the row directly; held off a type only by `livespec.ClaimedItemTimeoutExclusions` (`bugs_closed/317`) |

`verifier_coverage_test.go` goes green regardless of any of that. **That is the bug**, and it is
still true today for every completion through writer 2.

## 2. Verified state `[2026-08-25, each with how it was checked]`

| thing | state | how |
|---|---|---|
| the gate in the binary | **LIVE, `v1.0.1337`, both pods** | `grep -aq` on `/proc/1/exe` for the string literals `verify_before_complete` + `verifier_not_consulted`, with a must-be-PRESENT control (`owned_page_refusal_status`) and a must-be-ABSENT control. The Go identifier `updateStatusVerifyConfigKey` is **absent**, as predicted — never probe for that |
| steps arming the key | **0 of 200 live agents** | recursive `jsonb_path_query` over `default_config` |
| the census | **4 agents / 6 `complete` arms / 22 steps** | unchanged since 08-24, re-run 08-25 |
| item types those agents handle | **7**, none with a verifier | live **UNION `site_work_items_archive`** — see §5 trap 1 |
| completions all-history via the unguarded writer | **578** | same union |
| `verifier_not_consulted` rows | **0 — AND UNINFORMATIVE** | demand control empty: no verifier exists for any of the 7 types, so the record CANNOT fire. See §5 trap 2 |
| council | **APPROVED round 1**, `7a6add95-30e9-4576-85e5-df5bad0f7119` | 12 seats, 5 abstained, 2 medium objections, both acted on |
| the bug | **OPEN, and correctly so** | §9c of the bug file |

⚠ **Every row is a snapshot on a tree many sessions share. Re-run before acting.**

## 3. What shipped, and what each commit is

| commit | what |
|---|---|
| `c735bfd9c` | the gate (`update_work_item_status_verification.go`), the wiring in `UpdateWorkItemStatusAction`, `loadWorkItemVerifyRow` factored out of `verifyBeforeComplete` so both writers share one row read, and the tests |
| `c94212ad3` | `verifier_coverage_test.go`'s header, `CQ-023`'s corrected landmine, register `WII-030`, the index row, the ratchet line dropped, the new LANDMINES entry |
| `721465601` | the guardian objection answered: mutation proof on the shared extraction + the coverage gap it exposed |
| `35257bee2` / `b6aa4853b` | the rolling-window correction (§8 of the bug file) propagated everywhere including both Go headers |
| `e88cd0e4f` / `b0c066ac5` | post-roll: LIVE at the artefact; landmine verifier answered; bug file §9 |

All five carry `Council-Reviewed: 7a6add95-…`.

## 4. WHAT IS LEFT — three items, and only the third is a design question

### 4a. Arm the arms (blocked on somebody else, not on you)

The gate cannot do anything until a verifier exists for one of the 7 types. **Do not write one
just to exercise the gate** — see §5 trap 3. When somebody does, arming is a **per-arm** decision
with that type's close paths read first (`CQ-023`).

### 4b. Candidate 4 — the enforcing half (READY TO BUILD, design settled)

Full shape in **bug file §7c**. Copy `bugs_closed/317`'s three parts, which already solve this
exact class for the third writer:

1. a declaration in `platform/livespec` of which `(agent, step)` pairs complete through an
   **unarmed** `update_work_item_status`;
2. a lockstep test asserting **both directions**, modelled on
   `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go`: no type with a
   registered verifier may be reachable by an unarmed completer, and an entry no gate can earn
   does not linger;
3. `cmd/config-key-audit --live-declaration-drift` extended to compare that declaration to live
   `agent_definitions`.

⚠ **Do not build (1) and (2) without (3).** A hand-kept declaration goes stale by ADDITION, which
is precisely the criticism the council levelled at the first cut of `verifier_coverage_test.go`.
This buys what the shipped record cannot: the runtime record fires only **after** a live item has
already completed unverified; a lockstep fails **the build**, the moment somebody types
`RegisterVerifier(…)`.

### 4c. Candidate 2 — unify the writers (ARCHITECTURE-SCOPE, the real fix)

`bugs_closed/284` is the precedent (duplicate writers unified with a structural
single-definition test). This lane did its **first half**: both paths now share one gate
implementation and one row read. **Not started, and not startable as a bug patch** — the owner
ruling of 2026-07-29 and `bugs_closed/124`'s REJECTED verdict are both about exactly that shape.
Nobody has read `CompleteWorkItemAction`'s call sites to judge feasibility.

## 5. Traps specific to this lane — the ones that already cost this session

1. **`site_work_items` is a ROLLING WINDOW.** `work-item-archiver` moves terminal rows to
   `site_work_items_archive` (25,281 rows), which is where COMPLETED rows go — i.e. exactly what a
   completion census counts. Measured: the live table alone reports **5 types / 134 completions**;
   the union reports **7 / 578**, and two whole types had completed *entirely* into the archive.
   **My positive control had the identical blind spot, because it queried the same table.** Caught
   by the council, not by me (`WRONG_CALLS.md`, and bug file §8).
2. **The `verifier_not_consulted` zero has no demand control.** Nothing can produce that row until
   a verifier exists for one of the 7 types. Quoting the zero as "the mechanism is fine" is the
   error; so is quoting it as "the mechanism is broken".
3. **Registering a verifier to test the fix would break a live route.** `CQ-023`: the
   `required_fields_missing` one fail-closes the `converted` arm. Use a fixture, or a type with no
   live close path.
4. **A mock cannot assert SQL text.** sqlmock returns whatever rows you queue regardless of the
   statement, so a test asserting the VALUES that reach `VerifyTarget` passes even if the column is
   dropped from the SELECT. Put the column list in the **expectation** (`verifyRowReadSQL`).
   Measured: values-only, dropping `spec` failed nothing; expectation-matched, it fails 6 tests.
5. **Mutate, don't trust green.** Seven mutations are on record (M1 wiring removed, M2 gate never
   blocks, M3 bypass unrecorded, M4 seam re-pointed, M5 extracted helper drifts, M6 spec column
   dropped, M7 page_id dropped). ⚠ The sibling trap: this arm carries the terminal-decision guard
   in SERIES, so a fixture row in a terminal status makes any mutation read as covered — fixtures
   must sit in `detected`/`claimed`/`triaged`.
6. **Do NOT register a verifier from a test `init()`.** The registry has no removal and is read
   process-wide by `TestClaimTimeoutExclusionCoversBothCompletionGates`. Use the `verifierLookup`
   seam and `t.Cleanup`. (This is how the third writer was found — `WRONG_CALLS.md`.)
7. **The code index is pinned at `e347c5ad` of 2026-08-23.** Any landmine-verifier verdict on
   post-08-23 files reports "0 rows" as *staleness*, not absence. Mine did; it is answered in the
   entry and is NOT an open objection.
8. **Re-locate by symbol, never by line.** The bug file's original coordinates drifted ~30 lines in
   one day. `grep -n 'func UpdateWorkItemStatusAction'`.

## 6. What is NOT established — do not quietly inherit it

- **Whether any of the 7 types SHOULD have a verifier.** Two are on `verifier_coverage_test.go`'s
  own `catMechanical` backlog; that is an invitation, not a judgement.
- **Whether the 578 unguarded completions contain FALSE completions.** Nobody has re-run those
  predicates. `bugs_open/367` found one by accident; that is one, not a rate. Its own measurement,
  probably its own `090`.
- **Whether candidate 2 is feasible.** Cited by shape from `bugs_closed/284`, never verified.
- **The `image_url_404` undispatched population** — 42 rows, 38 `detected`, empty `handler_agent`.
  ⚠ And correct the handoff that first named it: `image-url-404-handler` has **not** handled "0
  rows ever" — it handled 3, all archived. Different defect, unfiled, not this bug's.
- **Whether `098` has credited these commits.** The trailers are present and well-formed on all
  five; the report itself is slow and was not run to completion. `098_REPORT_unreviewed_commits_v1.sh 3`.

## 7. Where everything is

| what | where |
|---|---|
| the bug | `bugs_open/375_HANDOFF_2026-08-23_update_work_item_status_completes_without_ever_consulting_the_verifier_framework.md` — **read §7c, §8, §9** |
| the gate | `platform/orchestration/actions/update_work_item_status_verification.go` |
| the wiring | `platform/orchestration/actions/v3_site_actions.go`, `UpdateWorkItemStatusAction` (locate by symbol) |
| the shared row read | `platform/orchestration/actions/complete_work_item_verification.go`, `loadWorkItemVerifyRow` |
| the tests | `platform/orchestration/actions/update_work_item_status_verification_test.go` |
| the third writer's precedent | `platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go`, `bugs_closed/317` |
| the coverage guard | `platform/orchestration/actions/discovery_checks/verifier_coverage_test.go` |
| this lane's five | `PLAN_2026-08-24_…`, `RUNBOOK_…`, `NOTES_…`, `README_where_we_are.md`, `SUMMARY_2026-08-24_…` in this directory |
| register | `WII-030` (`docs026_concept_register/register/work-item-integrity.md`); `CQ-023` (corrected) |
| the landmine | `LANDMINES.md`, "Registering a verifier protects a type only on the paths that ASK" |
| the wrong calls | `WRONG_CALLS.md`, two entries dated 2026-08-24 |
