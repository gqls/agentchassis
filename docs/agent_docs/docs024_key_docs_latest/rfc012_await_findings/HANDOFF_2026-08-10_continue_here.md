# HANDOFF — RFC_012 lane · **START HERE** · written 2026-08-10, final status 2026-08-11

> ## THE LANE IS CLOSED. NOTHING IS OWED. Updated 2026-08-11.
>
> Every piece of work, every council round and every verification this lane owned is finished.
> **Do not pick this file up as a work queue** — it is kept as the record of how the work went and
> as the source of the traps below, which are the part still worth reading.
>
> - **All five council rounds APPROVED**: `b0deddf7` (§7 backfill), `c88c0a84` (`decision`
>   lockstep), `c961b79e` (array index), `501ef561` (dead keys), `4739c992` (registry-parity skip,
>   landed 2026-08-11 — both its objections challenged absence claims of mine; both re-checked and
>   both held, written up in NOTES).
> - **All code LIVE and pod-verified on `v1.0.1284`**, both replicas (`fallback_url_field is
>   configured but resolved to no URL` = 1, `commit_from` = 0, absent-needle = 0). Verified on
>   1277, 1283 and 1284 — the fleet rolls often; re-verify rather than quoting these.
> - **Migration 356 APPLIED and recorded** (2026-08-10). Config half done.
> - **`RFC_019` is DECIDED** (owner: the shared ladder ships) and its acceptance evidence is
>   discharged as far as it honestly can be — see §12 and Job 1 below for why "as far as it
>   honestly can be" is the correct phrasing and not a hedge.
>
> **The live work moved to `COMMISSION_2026-08-10_owner_rulings_five_pieces.md`** — five owner
> rulings, already started by another thread (item 5 is BUILT, `5f8a326fc`, inert until it rolls).
> If you are here to do something, go there.
>
> **Still the owner's, not a thread's:** `bugs_open/236`'s design question (commission item 1) and
> the `deploy_commit` scoping (item 3).


**Supersedes `HANDOFF_2026-08-09_continue_here.md`**, which is now history: its two open items are
both resolved (one by an owner ruling, one by a measurement that turned out to be unfalsifiable).
Go back to it only for the ~36-row baseline story and the §1a correction, both still accurate.

Then: `RUNBOOK_rfc012_await_findings.md` · `NOTES_rfc012_await_findings.md` (the missteps — 18 now)
· `PLAN_2026-08-06_rfc012_execution.md` (the 08-09 "later" section is the ledger of what shipped).

---

## STATE: everything the owner commissioned is DELIVERED. Two jobs remain, and neither is coding.

| piece | id | state |
|---|---|---|
| Owner ruling — the shared ladder ships | `RFC_019` §11 | **RULED + RECORDED** |
| §7 resumed-step `RunAgentType` backfill | `58aefe282` | **APPROVED** (`b0deddf7`), **LIVE on v1.0.1277** |
| RSH-009 ladder | `1bc08d1ce` | **LIVE on v1.0.1277** |
| `decision` → `validDocSubjectTypes` | `5019bf2b7` | **APPROVED** (`c88c0a84`), live at HEAD |
| `ExtractNestedField` array index + fallback logging | `f7111f4d8`, `6cb41ae06` | **APPROVED** (`c961b79e`), **LIVE**; registered **WFA-012** |
| Dead config keys + detector opt-ins | `96f8075fb` | **APPROVED** (`501ef561`); **migration 356 NOT APPLIED** |
| Registry-parity test unblocked | `a6c0498f2` | done — all 4 packages green at clean HEAD |
| Hero/logo silent break | `bugs_open/236` | **FILED, root cause NOT established** |
| `PROCESS` trigger → "adds, changes or removes" | in `8f31ef710` | done, owner-sanctioned |

**Nothing is in flight. Nothing is broken.** Do not start coding before reading the two items below.

---

> **JOB 1 IS DONE (2026-08-10, later).** The induced test exists — `4fa9d1dec`,
> `platform/orchestration/resumed_step_provenance_test.go` — and it reproduces the production
> symptom verbatim under mutation (`ResolvedAgentType() = "generic", want "council-gate"`) on the
> real `ToResponseHeaders`→`FromResponseHeaders` resume path. **It had to be induced because the
> live case is EXTINCT, not merely dormant:** no inheriting door has filed a `generic` row anywhere
> in the fleet since **2026-07-26**, the day `RunAgentType` shipped and fixed the dispatch-sender
> half. See `RFC_019` §12's census table. **Read the rest of this section anyway** — the two traps
> in it (comment-needles, demand-vs-traffic) are why the job existed, and both will recur.
> **What is left of Job 1: nothing actionable.** The mechanism is proven on the resume path in a
> harness; whether the condition ever arises again in live traffic is not something a thread can
> force, and on the evidence it may simply not — in which case this fix is insurance, and the
> residue it was sized against was gone before it shipped. Do not open a new round for it.

## JOB 1 — the §7 acceptance evidence is STILL OWED, because the test as written cannot fail

**Do not read RFC_019 §7's recipe and run it. Read §12 first — it explains why §7 is wrong.**

The chassis rolled to **`v1.0.1277`** (both replicas, started 2026-08-09 21:35Z) and **the code is
in the binary**. That half is settled and does not need redoing:

```
POS "fallback_url_field is configured but resolved to no URL"  → 1   both replicas
POS "check 'url_field', 'fallback_url_field'"                  → 1   both replicas
NEG "zzz_this_phrase_is_in_no_version_of_the_binary"           → 0   both replicas
```

⚠ **Why those needles and not §7's.** §7 names two phrases that are **Go comments**
(`types/context.go:734`), so grepping for them returns 0 on a binary carrying the fix perfectly.
Neither `ResolvedAgentType` nor the backfill contributes any string literal of its own — one is pure
control flow, the other is one `if` — so the binary is dated by a **neighbouring literal from a
descendant commit**: both POS strings arrive in `f7111f4d8`, and `git merge-base --is-ancestor`
confirms `1bc08d1ce` and `58aefe282` are its ancestors. **Keep that argument if you re-verify after
a later roll; do not go back to the comment phrases.**

**What is NOT settled: whether the fix changes any row.** Measured 2026-08-10:

| query | result | reading |
|---|---|---|
| `generic` rows, the 3 residual actions, post-roll | **0** | looks like success |
| §7's control: all rows post-roll / distinct types | 288 / 20 | **passes** — and is blind |
| **rows from those 3 actions post-roll, ANY `agent_type`** | **0** | ← the actual finding |
| baseline, same 3 actions, 13d pre-roll | 33 | dominated by 07-27→08-05 |

The three producers filed **nothing at all** after the roll, so there was nothing to relabel. Worse:
their `generic` rows stop on **2026-08-05**, four days *before* the roll, and
`diagnose_council_decide` (42 of their 47 lifetime `generic` rows) last filed **2026-08-02**.
`agent_error_log` retains from 07-11, so this is dormancy, not retention. **The test could not have
returned non-zero from before the code existed.** Written up in `WRONG_CALLS.md` 2026-08-10,
because it is this lane's own §1 correction recurring one level up.

**A real positive control did survive and is worth keeping:** 3 `generic` rows *were* written
post-roll, by `process_message` (2) and `orchestrate` (1) — the coordinator paths RFC_019 §1 scoped
**out**. So the `generic` write path is alive; the silence is specific to the actions-door
producers.

### What to actually do

**Induce one.** Waiting is not a plan — the producers are dormant and may stay so. Deliberately fail
a step **that has been resumed after an await**, through one of the six RSH-008 doors, and read the
`agent_type` on the row it writes. Expected: the real agent type, not `generic`. That is the only
check whose result depends on the code rather than on whether anything happened to break this week.
The mutation matrix in the RUNBOOK shows which doors are reachable; `bugs_open/236`'s
`deploy_image_asset` path is an await-then-fail shape that already occurs naturally.

⚠ **Before quoting any count from `agent_error_log` again, ask what the DEMAND on that exact path
was in the same window.** Fleet traffic is not demand. This is the trap that ate this measurement.

> **JOB 2 IS DONE (2026-08-10, evening) — 356 is APPLIED and RECORDED.** The owner said go ahead.
> Applied by hand, scoped to the one file (never `--apply`, which takes every pending file):
> `UPDATE 7`, the `DO`/`RAISE` verify block passed, `COMMIT`, exit 0. **Verified at the artefacts,
> not at the exit code:** `commit_from` = **0** rows fleet-wide across all four workflow columns in
> any row state; the HITL `output_format` templates = **0**; all seven step configs still resolve to
> objects with **0 sibling keys lost** (each one printed and eyeballed); **7**
> `agent_definitions_backup` rows written by the two-arg `snapshot_agent`. Ledger row recorded via
> `--record-only` with those checks in the note — a hand-applied file that is not recorded gets
> replayed by the next session's `--apply`.
> **Expected steady state now:** the live detector (running since `v1.0.1279`, still present on
> `v1.0.1283`) warns on **one** step, `content-reviewer.mark_page_needs_attention` (`notes_field`,
> `validation_issues_field`) — confirmed still present in config. **That warning is the detector
> working, not a regression: do not silence it.** It is the `create_work_item`/`spec` precedent, and
> those two keys encode an intent the action never had.
> ⚠ **That "one warning" is a PREDICTION, not an observation, and it cannot currently be observed:**
> `content-reviewer` has **0 runs** in the retained `orchestration_states` window, so the step has
> not executed and the warning has had no opportunity to fire. Check demand before reading anything
> into its absence — a silent log here means the agent did not run.
> The section below is kept as the record of what the file does and why.

## JOB 2 — migration 356 ~~is written, verified, and DELIBERATELY NOT APPLIED~~ **APPLIED 2026-08-10**

`docs/agent_docs/sql_for_agents/356_retire_dead_config_keys_commit_from_and_hitl_output_format.sql`
(+ `_ROLLBACK`). Strips `commit_from` from the 6 live agents that carry it and the dead
`output_format` map from `simple-content-writer-with-approval`. Council **APPROVED** (`501ef561`).

- **Still pending as of 2026-08-10:** `SELECT count(*) FROM agent_definitions WHERE is_active AND
  COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND default_config::text LIKE
  '%commit_from%';` → **6**.
- **Either order is safe.** The Go opt-in is live on v1.0.1277 and will warn on 7 steps until the
  migration lands; the keys were dead before and after. No image-ordering constraint, so it is a
  plain pending file, not a `_HOLD`.
- Its verify block uses `DO`/`RAISE` and **was induced to fail** against the live DB before commit
  (psql exit 3), so it can actually stop a `COMMIT` — unlike a block of bare `SELECT`s.
- ⚠ Numbering raced **four times** in one afternoon (352–355 all taken while the work was in
  progress). Re-check `ls docs/agent_docs/sql_for_agents/ | grep -E '^35'` before assuming 356 is
  still unique.

---

## THINGS THAT WILL MISLEAD YOU (all learned the hard way this week)

- **The top-level step census sees 3 of 6.** Three of the six `commit_from` steps are nested inside a
  loop `sub_workflow`. `s.value->'config' ? 'commit_from'` over `jsonb_each` of `workflow->steps`
  returns **3** and looks authoritative. Use a depth-aware walk, or `default_config::text LIKE` as a
  coarse cross-check. I ran the naive one as a "verification" of the agent's figure and it
  disagreed; **the agent was right and my check was the broken one.**
- **The live HITL step spells its action `process_data`, not `process_approval_decision`.**
  Registering only the canonical name opts in **zero** live steps while looking like a working fix.
  The deprecated-alias registration is load-bearing.
- **`bugs_open/236` names a REFUTED theory on purpose.** The obvious cause — "the awaited-response
  merge overwrites the key" — is contradicted by `coordinator.go:2719-2748`, which preserves
  existing keys and adds to them. Do not re-derive it and do not quote it as the cause.
- **Both 090 runs on the hero/logo symptom returned UNVERIFIABLE, for different reasons** — the
  first (`dce40cf4`) because no evidence existed, the second (`074beb8a`) because the harness could
  not reach evidence that does. **Neither is a refutation** and the pair must not be quoted as one.
- **The diagnosis loop cannot address `orchestration_states`.** It is absent from the bundle's
  "Schema (live tables)" section, so the loop queried it by `id` when the column is
  `orchestration_id` and failed with 42703. Any hypothesis whose evidence lives in the platform's
  central state table is currently unfalsifiable by the loop **for reasons unrelated to the
  hypothesis**. Flagged for whoever owns that loop; `[UNVERIFIED]` what populates `runtime.schema`.
- **A change with no string literal cannot be pod-grepped for.** Date it by a neighbouring literal
  from a descendant commit and prove ancestry. Say so in the acceptance section rather than
  inventing a phrase from a doc comment.

## Other threads, still NOT this lane's

- **`bugs_open/236`** — hero/logo. Contained fix (make the three silent readers say something when
  they come up empty) is safe and unclaimed; the real fix may land in the `(a)`/`(a′)` merge design,
  which remains an **open owner decision**.
- **Dead config keys, the residue:** `content-reviewer`'s `notes_field` and `validation_issues_field`
  are dead the same way and were **left standing as declared true positives** (the
  `create_work_item`/`spec` precedent), documented in the spec comment.
- **`search_results.results.0.url` is FIXED** (WFA-012) — but ~10 near-clone path walkers remain
  map-only, so "dot paths support indexes now" is false as a general statement.

## Milestone read-outs
`SUMMARY_2026-08-08c_the_seam_is_strict_now.md` · `SUMMARY_2026-08-09_the_last_job_and_the_reason_that_was_wrong.md`
— the current one. **No new summary was written for 08-09/08-10**: by the five-headings test the
read-out would say much the same as 08-09's, and rarity is the design. The next real inflection is
Job 1 returning a result that could have come out otherwise.
