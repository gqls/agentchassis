# HANDOFF — RFC_012 execution · **START HERE** · updated 2026-08-07

Cold-start for the next session. Read this, then `NOTES_rfc012_await_findings.md` (the
missteps — they are the point) and `RUNBOOK_rfc012_await_findings.md` (every command, with
its gotcha).

**This lane owns three owner rulings** (RFC_012 second sitting, recorded in the RFC itself
at commit `3851e90b5` — read that block, it is the authority, not this file):

1. **(d) → a STANDING check, ONLINE if the framework allows.**
2. **The §3(a) reader census — COMMISSIONED.**
3. **Option B implementation — ASSIGNED here.**

---

## STATE

| piece | commit | state |
|---|---|---|
| The rulings, recorded in the RFC | `3851e90b5` | durable — survives any session |
| Lane docs (the standing five) | the `docs(rfc012 lane)` commits | open, updated 08-07 |
| **B core**: `agenterrors` leaf package | `5f49b4cfd` | **DONE, POD-PROVEN LIVE on v1.0.1259** |
| **B rest**: 18 site conversions + tests | `f930de86b` | **DONE and POD-PROVEN LIVE on v1.0.1262** (2026-08-07, both replicas) |
| **(d) detector**: `--shared-output-fields` | `abf5e8266` | **DONE, proven, live-run green** — a `cmd/` binary, not in the chassis image, so it needs no roll |
| Concept register: RSH-008 + WFA-011 | 08-07 | **DONE** (the standing ruling's debt from the split commits, now paid) |
| Council round for the whole code set | corr `5c2bc265-84ac-452b-bd8b-22fd7b875427` | **round 2 SUBMITTED 2026-08-07 08:29Z, verdict not yet read** |
| The `domain` `NULLIF` follow-up | `8e786652b` | **CLOSED as NO — measured; adding it would be a REGRESSION** |

### ✅ FIRST THING — DONE 2026-08-07 08:29Z: round 2 is submitted. Read the verdict, do not resubmit again

`RESUBMIT_CORR=5c2bc265-…` under the same correlation; it began executing immediately
(`review_editquality | EXECUTING_STEP`), no 29-minute queue. Submission file:
`…/e62271e7-…/scratchpad/rfc012_submission_r2.json`. What changed — **the submission, not
the code**: the edits array is now declared a **REPRESENTATIVE SAMPLE** in the summary, the
scope is stated as 34 unique files across three commits, all eight entries name one distinct
SHAPE, and `validate_page_content.go` — the provenance site the verdict said I described but
did not show — is now IN the array. Claims about unshown files are given as counts with the
command that produces them.

**Preparing it turned up a second instance of the same objection:** last round's seventh edit
named `cmd/config-key-audit/shared_output_fields.go`, **which is not a file** (it is
`sharedoutputs.go`). So the array did not merely under-cover the change, it carried a path
that does not exist. **Before submitting, assert every `edits[].file` exists** — one loop over
the array against the tree.

Read the verdict with:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='5c2bc265-84ac-452b-bd8b-22fd7b875427' AND kind='council_report' ORDER BY created_at;
```
`f930de86b` carries `Council-Submitted:`, so 098 credits it automatically **if** this round is
approved. **Never write `Council-Reviewed:` on a verdict you have not read.**

<details><summary>The round-1 REVISE this answered (kept for the record)</summary>

### the verdict came back **REVISE** — resubmit, do not re-argue

Corr `5c2bc265-84ac-452b-bd8b-22fd7b875427`, gating objection from `editquality`. **Read
the NOTES entry for 2026-08-07 before touching it.** In short:

- **The objection is about the SUBMISSION and it is correct.** My prose named files that
  were not in the `edits` array (7 entries, 25 files changed), so I described work the
  council could not judge. **Fix:** declare the array a REPRESENTATIVE SAMPLE in the
  summary, name the file count, list the distinct shapes it covers, and stop referring to
  files it does not contain. Then:
  `RESUBMIT_CORR=5c2bc265-84ac-452b-bd8b-22fd7b875427 ./…/097_TRIGGER_council_review_v1.sh <file>`
- **It also caught a real miss, already acted on:** a landmine on `agent_error_log.domain`
  existed a day before I argued NULL→`''` was "measured inert", with `agenterrors.go` in
  its own footprint. Correction written onto that landmine, `WRONG_CALLS.md` entry filed,
  and **RSH-008 now states the real fix — give the shared writer a `NULLIF` on `domain` —
  which is NOT done and is the one piece of follow-up code this verdict implies.**

</details>

> **CORRECTED 2026-08-07 (`8e786652b`) — that last sentence is WRONG and the follow-up is
> CANCELLED. Adding the `NULLIF` would be a REGRESSION.** Measured post-roll, the table has
> **converged on `''`**: 0 NULL / 29 `''` / 16 real since 05:47Z, against 128 / 13,885 / 4,762
> before. All 128 NULL rows were written by sites *this conversion changed* (nine groups by
> `agent_type, action`, newest 2026-08-05, all pre-roll), so the NULL bucket is closed and the
> 14/30-day reaper empties it. A `NULLIF` now would put 100% of new rows into the shape 0.9%
> of rows use. The `site_id` `NULLIF` is a **uuid type necessity** (`''::uuid` raises), not a
> precedent. **The remedy is the reader's: `COALESCE(domain,'') = ''`.** Two findings fell out
> of measuring instead of conceding: the "nineteen copies" census **grepped `platform/` only**
> (a third live site exists at `internal/agents/contentcreator/claims_guard.go:184`, dormant,
> the last latent NULL producer), and **it cannot use the shared door at all** — `contentcreator`
> holds a `*pgxpool.Pool` while `agenterrors.Write` takes a `*sql.DB`, so "the ONE writer" is
> true of the `database/sql` half of the estate only. NOTES 08-07 misstep 9; `WRONG_CALLS.md`.

`f930de86b` carries `Council-Submitted:`, so 098 credits it automatically **if** a
resubmission under this correlation is approved; **do not** write `Council-Reviewed:`
anywhere without reading an approved verdict.

### ✅ SECOND: the conversions ARE live (v1.0.1262, 2026-08-07) — and the old acceptance test here was WRONG

Both replicas started 2026-08-07T05:47Z, after `f930de86b` (01:24Z). Proven per-replica
with a **discriminating pair**, not with the roll — one string must be ABSENT and its
near-twin PRESENT in the same binary:

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c '
  strings /app/agent-chassis | grep -cF "failed to write some discovery check error records"  # POS -> 1
  strings /app/agent-chassis | grep -cF "failed to write discovery check error record"        # NEG -> 0
  strings /app/agent-chassis | grep -cF "content_data envelope: failed to write record"       # NEG -> 0'
```
`f930de86b` changed that discovery-checks log line from singular to plural, so no stale
image and no lucky substring can satisfy both halves. Measured 1 / 0 / 0 on `-5ghft` and
`-dfk4b` independently.

> **CORRECTED 2026-08-07 — the needle this section used to publish gives the WRONG ANSWER
> on a correct binary.** It said `strings … | grep -c "INSERT INTO agent_error_log"` must
> read **2** after the roll (14 before, 2 INSERT sites in the tree). The real value is
> **1**: the two surviving sites hold **byte-identical** SQL and the Go linker
> **deduplicates identical string constants**, so two sites are one string. The old 14 was
> 14 only because the hand-copies had drifted — converging them is what collapsed the
> count. A session running the documented check would have seen 1 ≠ 2 and concluded the
> work had not shipped. **A string-literal count cannot count SITES.** Full write-up:
> NOTES 08-07 (morning), misstep 8; class filed in `LANDMINES.md`.

---

## NEXT — three pieces, in this order

### 1. The reader census (gates (a)/(a′)) — the biggest remaining piece
**Not started.** The delegated agent died on quota mid-sweep; its one partial finding worth
keeping is that `enrich_fingerprint_with_css` is *wrapper-adapted*, and it had begun a third
sweep for mid-string/condition-expression references two greps would miss.

Deliverable: `architecture_review/CENSUS_2026-08-06_rfc012_await_step_readers.md` — config
side (live `agent_definitions`, queries verbatim so it re-runs) + Go side (mechanically
findable, honestly bounded), each reader marked BREAKS/SURVIVES/UNCLEAR under
merge-not-replace. **The census does not decide (a)** — it is what (a) stays gated behind.

⚠ Two things a naive census will miss: the **`call_agent`/`spawn_agent` branch already
merges under `.response`**, so those readers already tolerate the shape; and dynamic loop
steps **derive** their output_field at runtime (`deriveOutputFieldFromLoopStepName`), so a
config-only census cannot see those keys. **This is a session's work on its own — do not
delegate it alongside anything else** (that is misstep 4 in NOTES).

### 2. The online half of (d) — the CronJob
Precedent to follow is **`component-render-check`**, NOT `single-owner-carriers-check`:
ship the Go binary as its own image rather than re-implementing in Python. RFC_006's Python
mirror exists only because a job cannot `go run` from source (262M clone); an image
dissolves that, and with it both of RFC_006's named drift risks (the `DECLARED_*` literal
and the parity test).
- copy `deployments/kustomize/services/component-render-check/` as the shape;
- take a free schedule slot (taken: 02:00, Sun 06:00, 06:20, 06:40, 06:50, 06:55) and say
  which in the manifest;
- **report to `doc_notes`** — one row per run **including clean**, so a missing row means
  THE JOB DID NOT RUN, which must not be indistinguishable from "nothing is wrong";
- exit 1 on NEW findings so the Job shows failed;
- the ack file is in-repo, so it ships with the build.

### 3. The last INSERT site — STILL BLOCKED (re-checked 2026-08-07)
`git status --porcelain` on it is **still ` M`** — one line, another session's
`PageWantedLivePredicateFor` change at `:872`, untouched by the INSERT at `:1353`
(`git diff -U0 <file> | grep -c agent_error_log` → 0). Converting now would take their line
as a same-file passenger, so it stays blocked. Re-check with the two commands above; the
status goes stale within minutes.

`store_generated_component_action.go:1353` is the **one** site left hand-copied, on three
independent grounds (all in NOTES 08-07): a standing council objection naming it directly,
it already writes the canonical 13 columns, and it was dirty with another session's
`PageWantedLivePredicateFor` change (a pathspec commit still takes a same-file passenger).

**When `git status --porcelain` on that file is clean:** convert with `LogActionEntry` and
set `AgentType:"component-creator"`, `StepName:"store_component"`,
`Action:"store_generated_component"` **explicitly**. Those three literals are exactly what
the guardian/edit-quality seats warned would be misfiled fleet-wide, and no test in the
package would catch the slip.

---

## What B actually changed, for anyone reviewing it cold

- **19 hand-copied INSERTs → one writer.** 18 converted (the 19th was left, above); the
  count is 19 not 18 because `plan_sections_action.go` gained a **NEW** hand-copy
  (`FACT_SCOPING_EMPTY_COMPOSITION`, `ff515351e`) from another session *after* this lane's
  census and *before* it landed. That is the argument for the seam, not for a tidy-up.
- **Nine sites gain `orchestration_id`** — the run join. `reconcile_superseded_reviews` also
  gains `step_name`.
- **`domain` goes `''` where nine sites wrote NULL.** The filtering reader
  (`diagnose_load_runtime_action.go:267`) is genuinely indifferent — but a landmine filed
  the day before already showed this makes the `domain IS NULL` census undercount WORSE,
  and I missed it. **The real fix, NOT done: give `agenterrors.Write` a `NULLIF` on
  `domain`.** See RSH-008, `WRONG_CALLS.md` 08-07, and NOTES 08-07.
- **12 test arity pins updated to 13.** Two of them asserted a code/severity against SQL
  **literals** that are now bind parameters — both assertions were MOVED to the argument
  and pinned by value, not relaxed to `AnyArg`.

## Landmines this lane found (all now in `LANDMINES.md`, synced to `doc_notes`)
- **`LogActionEntry`'s merge fills a provenance you meant to set, and the whole suite stays
  green** — the package's tests pin codes and messages, never `agent_type`. Full entry +
  the `agent_type` NOT NULL and `domain` corollaries are in `LANDMINES.md`.
- **A nil `Context` map marshals to `null`, not `{}`** — the old writer's nil-guard never
  fired; byte-compatibility means pinning `"null"` in tests, not `"{}"`.
- **The RFC's own 13-key list is one key short** — trust the live query, not the prose.
- **A green live run of the census proves the FLEET, never the DETECTOR** — the mutation
  test (sever the `config.then_step` edge, lose the finding) is what proves the detector.
- **The concept index's uniqueness check earned its place**: my first `WFA-007` collided
  with a live relay-gaps entry. Renumbered to WFA-011 before landing. Always run the
  id-uniqueness check, not just the row count — the count cannot see a collision.

## Milestone read-outs
`SUMMARY_2026-08-06_two_of_three_built.md` — the series' first. **A second is owed once the
conversions go live** (the read-out genuinely changes then: "built" → "shipped and proven").
Do not write one before that; it would say much the same as the first, which is the test
the standing-five rules set.
