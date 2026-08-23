# HANDOFF 2026-08-23 — continue here (`bugfix_305_negation_gate`)

**Supersedes `HANDOFF_2026-08-22_continue_here.md`.** Read that one for how the gate got live and
proven at the artefact; read this one for state. Its "⏱ UPDATE 2026-08-23" section is still accurate
as far as it goes — this file continues from it.

> ## ▶ ONE-LINE STATE
> Two open defects, and they are in **opposite** states: the repair-log accounting fix is
> **APPROVED and committed but INERT until the next chassis roll**; the repair-ceiling fix (found
> today, and it was losing twice as much copy) is **LIVE IN CONFIG NOW** but **not yet
> demand-proven** — no repair call has run at the new ceiling yet.

---

## 1. What changed today, and how each was checked

| thing | state | how it was checked |
|---|---|---|
| council `f3046f0c` (§26 target accounting) | **APPROVED**, and the round was sound | `decision=approved`, `unreadable=0`, `in_body=10`, `voted=10`. 8 approve, 2 object-on-record |
| the §26 fix itself (`6e9cb411d`) | **committed, NOT live** | binary probe both replicas: `no_answer_for_target` **0**; control `no_answer_for_targez` **0**; known-present `rewrite_negations` **8**. Pods started 11:51:39Z on `v1.0.1328`; the commit is 12:09:34Z — it postdates the image |
| the two objections' substance | **both answered by the code** | see §2 — the seats reviewed a *sketch*, not line 592 |
| **§27 the repair ceiling** | **FOUND, FIXED, LIVE** | migration `569` applied 2026-08-23 ~12:38Z, `UPDATE 1`, verify `DO` passed; live row **re-read independently** as `16000` / `claude-sonnet-5` |
| `569` demand control | ⚠ **NOT YET** | `llm_call_log` for this step still shows **only `max_tokens=2000`**, latest 12:38:42Z — no repair call has run since the apply. **This is the one thing left to watch** |
| council `4829bd48` (569) | **APPROVED**, round sound | `decision=approved`, `unreadable=0`, `in_body=10`, `voted=10`. One medium objection on record — answered, §2a |

## 2. The council objections — checked, not merely accepted

An approval is not a reason to skip the objections. Three were checkable first-hand; **all three were
already answered by code the seats could not see**, because a submission shows a sketch:

1. **`normaliseSentenceKey` parity (medium)** — answered *structurally*, not by assertion. Line 592
   marks `answered[negationTargetKey(t)]` off **the matched target `t` itself**, and
   `unansweredTargetRejections` checks the same `t` from the same slice: same struct, same pure
   function. The divergence the seat feared cannot be constructed.
2. **Possibly-vacuous test (low)** — there is an explicit `if len(plan.targets) < 3 { t.Fatalf }`
   floor, and budget 0 makes every non-exempt hit a target (the `used <= budget` branch is what
   *skips* one). Both tests re-run green.
3. **"one call site of a class" (medium)** — the *ceiling* version of that census was run rather than
   asserted (§3). The accounting-loop sibling audit the seat asked for (`evidence_citations.go`,
   `revalidate_unverified_claims.go`) is **still open and unowned**.

### §2a. `569`'s own round: APPROVED, one objection, also checked

The `duplicate_active_rows` landmine — "four agent types carry TWO active definition rows and only the
higher version is loaded; the plan never enumerates versions before writing". **Checked, twice over:**
`page-content-writer` has **exactly one** row in any state (`all_rows_any_state=1`, `live_active=1`,
version 2). It is not one of the four. **And the guard covers it anyway** — a second active row would
have matched the same `WHERE`, made the verify's `count(*)` **2**, and raised rather than committing a
half-write. Nothing owed, but the seat was pointing at a real trap on a real footprint.

**They were right to object even though each was wrong on the facts** — each named a real failure
mode the sketch could not settle. Answering all three cost ~15 minutes, and walking the loop's
branches to prove (1) is what found §27.

## 3. §27 — the ceiling, in one table

`[MEASURED 2026-08-23 ~12:35Z, rolling window]`

| targets | markers | outcome |
|---|---|---|
| 1 / 2 / 3 / 5 | 43 | **all `repaired`** |
| 9 / 10 | 2 | **both `repair_unavailable`** |

No exceptions either side — a hard capacity limit between 5 and 8 targets, not intermittent
truncation. The failure is **total** (the round is discarded whole), so those 2 markers held **19 of
78 targets — 24.4%**, against the **12** lost to §26. `569` raises 2000 → **16000**, anchored on the
sibling `generate_content` in the same sub-workflow. Full reasoning: bug file **§27**.

⚠ **Rolling-window figures.** `orchestration_states` is reaped; the population moved **42 → 44 during
the measurement**. Re-measure before quoting; do not carry these forward bare.

## 4. ⚠ CORRECTION carried forward — the invariant does NOT hold for every marker

§26 and the code comment claim `targets == rewritten + rejected` holds **"for every marker"**. It does
not. Five early returns (`rewrite_negations_action.go` 454, 458, 540, 559, 570) precede the
`unansweredTargetRejections` call at 665, so a `repair_unavailable` marker accounts for **none** of its
targets **by design** — its `error` field names which reason fired.

**When you run the post-roll census, segment by `status`, and do NOT tune the query until the total
reads zero** — that would hide the ceiling failures, which are the expensive half. Query and the three
traps: `RUNBOOK` §9.

## 5. What is left

1. **Watch for the first repair call at 16000.** `SELECT max_tokens, count(*) FROM llm_call_log WHERE
   step_name LIKE '%rewrite_negations%' AND created_at > '2026-08-23 12:39Z' GROUP BY 1;` — a row at
   `16000` is the demand control. Then the real proof: a marker with **≥9 targets** reading
   `status='repaired'`. Until then `569` is live-and-unexercised, and must be described that way.
2. **The roll**, for §26's accounting fix. Releases are whole-fleet and the owner runs `make release`.
   After it: binary-probe `no_answer_for_target` (with the two controls above), then re-run the
   `RUNBOOK` §9 census segmented by status.
3. ~~**Council `4829bd48`** — read the verdict~~ **DONE: APPROVED**, and its one objection is
   answered (§2a). Nothing owed.
4. **`D3` still must not be decided.** It was to be settled from the rejection log, and the log has
   been incomplete in *two* independent ways — §26's silent drops (fix inert) and §27's discarded
   rounds (fix live, unexercised). Give it a week of traffic under both.
5. **Not this lane's**: the accounting-loop sibling audit (§2.3), and the damage half's two pages
   still failing `validate_content` (their lane's claims work).

## 6. Standing cautions — added today

- **`/tmp` is a 16G tmpfs and was at 100%**, shared with every session on this tree. `go test` fails
  with `no space left on device` and it is not your code. **Do not clear it** — point `TMPDIR` at your
  own scratchpad (on `/`, 136G free).
- **`--apply` would have swept ~10 other lanes' pending files**, several `?? probe inconclusive` or
  `!! LIKELY ALREADY APPLIED`. Apply one file by hand with `psql -f`, then
  `run-migrations.sh --record-only <file> --note '<what you checked>'` — a hand-applied file that is
  never recorded stays "pending" for ever and eventually gets replayed (`bugs_open/007`).
- **Induce a migration's verify block before trusting it.** `569`'s was run with the `UPDATE` skipped
  and raised `569 FAILED: … got 0`, exit 3. A guard never seen to fail is not a guard.
- **The step is inside a SUB-WORKFLOW.** Every top-level `workflow.steps` query returns 0 rows and
  reads as "the step does not exist" — `RUNBOOK` §11.
- **The truncation census needs the LOOP-EXPANDED step name**
  (`process_sections_loop_iter_1_rewrite_negations`) and `provider='anthropic'`; there is no
  `finish_reason` column — `RUNBOOK` §10.
- Everything in the 08-22 handoff's §6 still stands (`\y` not `\b`; brief-supplied phrases are exempt
  BY DESIGN; a marker over a refused save describes a page that never changed).

**Migrations this lane owns:** `509`, `517`, `548`, **`569`** (all applied and recorded).
**Council:** `c48b7612` (gate, APPROVED r4), `a696e2a3` (truncation helper, APPROVED r1),
`f3046f0c` (target accounting, **APPROVED r1**), `4829bd48` (the ceiling, **APPROVED r1**).
