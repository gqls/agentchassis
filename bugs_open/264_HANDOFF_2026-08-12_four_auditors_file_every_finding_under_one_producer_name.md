# 264 — four separate auditors file every finding as `design-audit`, so no work item in the estate names the auditor that actually produced it

**Filed 2026-08-12.** Live, fleet-wide, all-history. **Root cause is the unfixed
literal-STRING half of `bugs_closed/042`** — this file is the *consequence*, measured,
and does not restate the mechanism. Read 042 first for why a string in step config is a
reference and never a literal.

---

## The defect in one line

Four agent definitions configure `audit_source` with a plain string. None of the four
values has ever reached a work item. **All four auditors' findings are stamped
`design-audit`.**

## Measured, live `clients_db`, 2026-08-12

**What the configs ask for** — every step in the fleet whose action is
`write_audit_findings`:

```sql
SELECT a.type, s.step_name, s.step->'config'->>'audit_source'
FROM agent_definitions a
CROSS JOIN LATERAL jsonb_each(a.default_config->'workflow'->'steps') AS s(step_name, step)
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.step->>'action'='write_audit_findings';
```

| agent | step | configured `audit_source` |
|---|---|---|
| `brief-fidelity-auditor` | `write_findings` | `brief-fidelity-audit` |
| `content-quality-auditor` | `write_findings` | `content-quality-audit` |
| `site-review-agent` | `write_strategic_findings` | `site-review` |
| `visual-design-auditor` | `write_findings` | `visual-design-audit` |

**What the record actually holds** — every distinct value, all-history:

```sql
SELECT spec->>'audit_source', count(*) FROM site_work_items WHERE spec ? 'audit_source' GROUP BY 1;
--  design-audit          | 265
--  tool-acceptance-tier4 |   1
```

**Zero rows carry any of the four configured values.** `[MEASURED]`

**And the runs are real, so this is not "the auditors never fire".**
`content-quality-auditor` alone: **34 orchestrations, all `COMPLETED`, most recent
2026-08-12**, every one a child of `design-audit-agent` or `site-review-agent`. A run's
own `collected_data` shows the write happening under the wrong name:

```
"findings_written": { "audit_source": "design-audit", "total_findings": 5,
                      "items_created": 1,
                      "classification_stats": { "content_rewrite": 2,
                                                "cta_improvement": 1,
                                                "needs_content_page": 2 } }
```

## Why the value never lands (mechanism, cited not guessed)

`WriteAuditFindingsInputSpec` (`write_audit_findings_action.go:41-46`) declares
`audit_source` Optional with `Defaults: {"audit_source": "design-audit"}`.
`ExtractActionInputs` (`datahelpers/action_inputs.go`) reads step config **only** as a
reference to resolve against `collectedData`:

- **Strategy 0** resolves `config[field]` only `if strings.Contains(pathStr, ".")`.
  `"content-quality-audit"` has no dot, so it is skipped.
- **Strategies 1/2** search `collectedData` for a key named `audit_source`; nothing sets one.
- **Strategy 4** treats a single-segment string as a *key name*, looking for
  `collectedData["content-quality-audit"]`, which does not exist.
- **Strategy 5** takes literals for non-string scalars **only**, and its comment says why
  strings are excluded deliberately: *"A string literal that fails to resolve is left
  alone on purpose: taking it as its own value would turn a broken reference into a
  silent literal and mask real wiring bugs."*

So the platform is behaving exactly as designed and documented. **This is a config-authoring
defect repeated in four agent definitions, not a platform regression** — and the reason it
is worth a bug file rather than four config edits is the consequence below.

> ⚠ **I nearly filed this as a platform bug.** My first write-up said "the configured
> literal is not landing" and implied the action was at fault. Reading Strategy 5's comment
> is what corrected it. The distinction matters for the fix: option 1 below is config-only.

## Why it matters (the part 042 could not have known)

1. **`spec->>'audit_source'` is documented as the only field that names a producer**
   (`bugs_open/213`'s landmine, and the concept register repeats it). That field is
   **wrong for four of the five producers**, so the one recorded defence against
   `item_type` hiding a producer split is itself defeated.
2. **No auditor's yield can be measured.** "How many findings does the copy auditor
   produce, and how many get fixed?" is unanswerable today for all four.
3. **It has already caused a real misdiagnosis.** The `copy_quality_two_stage` lane
   concluded on 2026-08-12 that the copy auditor's findings *"run and die — nothing
   consumes them, 0 rows all-history"* and drafted a build plan for the missing consumer.
   The findings were being consumed the whole time under another name. That is one lane,
   one day, and it was caught only because a second session ran a prior-art sweep.
4. **A `design-audit` row is not evidence a design audit ran.** Any query, report or
   council submission that has read `audit_source` as provenance has been reading a
   default. `[UNMEASURED — I have not swept for downstream consumers that filter on it;
   that sweep is the first thing the fixing thread should do.]`

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Config-only, no roll: set `audit_source` into `collectedData` in each of the four
   workflows**, then reference it by key. Smallest, live immediately, no Go change. But it
   leaves the trap armed for the fifth author, so it is a repair and not a fix.
2. **Make the field non-defaultable**: drop `design-audit` from `Defaults` and have
   `write_audit_findings` fail loudly when `audit_source` is absent. Converts a silent
   wrong attribution into a startup-visible error. Requires the four configs fixed first,
   or four auditors break. **This is the candidate that makes the bad state
   unrepresentable**, which is the ordering rule this estate uses.
3. **Resolve 042's string half** (`{"$literal": "..."}` marker, its option 3). Fixes the
   whole class, is the most invasive, and needs its own architecture round — every existing
   string config would need auditing. Do not bundle it with this.

⚠ **Whichever is chosen, `WasDefaulted("audit_source")` already exists** and is precisely
the signal that would have surfaced this: the action can log or refuse when the value it is
about to stamp came from a default. Nothing calls it here.

## How to verify a fix

```sql
-- after the fix, this must return four rows, not one:
SELECT spec->>'audit_source', count(*) FROM site_work_items
WHERE created_at > '<fix time>' AND spec ? 'audit_source' GROUP BY 1;
```
⚠ **Induce the non-zero before trusting it** — run one audit per auditor and confirm each
name appears. A single `design-audit` row proves nothing, because that is also what a
completely unfixed system produces.

## Related

- `bugs_closed/042` — the mechanism (string config is a reference; the string half is
  explicitly NOT fixed there, and this file is the measured cost of that decision).
- `bugs_open/213` — whose landmine names `spec->>'audit_source'` as the only producer
  signal. That landmine needs the caveat this file supplies.
- `docs024_key_docs_latest/copy_quality_two_stage/NOTES_two_stage_copy.md` — the
  misdiagnosis this caused, recorded as a correction.

## §12 — Fix applied 2026-08-13, both ordered candidates, verification pending

**Candidate 1 (config-only, no roll) — APPLIED AND VERIFIED LIVE.** Migration
`399_four_auditors_audit_source_resolves_to_a_real_value.sql` (+ `_ROLLBACK.sql`
sidecar), applied against `clients_db` 2026-08-13. Adds one `query_database` step
per agent with no `FROM` clause (`SELECT '<name>'::text AS audit_source`), whose
`output_format:"object"` flattens the literal into a new `collected_data` field
(`audit_source_literal`). The write step's `audit_source` config becomes the
genuine two-segment dot-path `audit_source_literal.audit_source`, which Strategy 0
resolves via `ExtractNestedField` — the same mechanism every correctly-wired
config value in this fleet already uses. `query_database` has no registered
`ActionInputSpec` (confirmed by grep — it reads `query`/`output_format` straight
off `StepConfig.Config`), so the new step needed no input-contract registration.

**`[MEASURED]` live re-check straight after applying, before any audit had run
under the new config:**
```sql
SELECT a.type, s.step_name, s.step->'config'->>'audit_source'
FROM agent_definitions a
CROSS JOIN LATERAL jsonb_each(a.default_config->'workflow'->'steps') AS s(step_name, step)
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND (s.step->>'action'='write_audit_findings' OR s.step_name='set_audit_source');
```
All four `write_findings`/`write_strategic_findings` steps now show
`audit_source_literal.audit_source`; all four `set_audit_source` steps present
with the correct literal, `next_step` and `output_field`. The migration's own
guard (a `DO` block asserting the new step, the rewired `next_step`, the updated
`audit_source`, AND that each write step's `site_id`/`findings_field` and each
predecessor's `output_field` survived untouched) passed on both a rollback-wrapped
dry run and the real apply.

**Candidate 2 (make the field non-defaultable) — APPLIED, COMMITTED, INERT UNTIL
THE NEXT ROLL.** `write_audit_findings_action.go`: `WriteAuditFindingsInputSpec`
moves `audit_source` from `Optional` (with `Defaults: {"audit_source":
"design-audit"}`) to `Required`, with no default. The action body's own
`if auditSource == "" { auditSource = "design-audit" }` fallback — a second,
code-level default that would have silently defeated the `Required` change for
any caller extracting an empty string — is removed as dead/misleading code.
`go build ./platform/orchestration/...` and `go test ./platform/orchestration/...`
both green.

**ORDERING, stated explicitly so nobody flips it on a future revert or replay:**
candidate 1 (config) had to be live BEFORE candidate 2 (code) ships, because an
older binary resolves the new dot-path unconditionally (Strategy 0 does not
depend on the Go change), but rolling the stricter binary FIRST — before all four
configs were fixed — would hard-fail every auditor's `write_audit_findings` step
at "missing required fields: [audit_source]" the moment it rolled. Candidate 1 was
applied and verified live first; candidate 2 was committed after. Candidate 3
(`bugs_closed/042`'s general string-literal-as-reference mechanism) remains
explicitly out of scope, per this file's own original fix-candidate ordering.

**Confirmed candidate-2 does not affect the fifth producer.** `tool-acceptance-tier4`
(the one non-`design-audit` value already landing correctly) sets `audit_source`
directly in a Go-constructed `spec` map in `tool_acceptance_actions.go`'s
`routeChromeFailures`, bypassing `WriteAuditFindingsAction` and
`ExtractActionInputs` entirely — read at the call site, not inferred.

**Submitted to the advisory council-review gate** (touches `platform/`):
`SUBMISSION_CORR=50ee4b26-2303-4304-b437-7320e1368a1d`. Verdict not yet read as of
this write-up; committing under `Council-Submitted:` per the standing norm rather
than holding the code for the ~30-minute queue.

**`[UNVERIFIED]` — the one thing still open.** Per this file's own "How to
verify a fix" section: nobody has yet run one audit per auditor and confirmed all
four now write their own name. Do this before closing:
```sql
SELECT spec->>'audit_source', count(*) FROM site_work_items
WHERE created_at > '2026-08-13T00:00:00Z' AND spec ? 'audit_source' GROUP BY 1;
```
must show four distinct values, not a single `design-audit` row (which is also
what a fully unfixed system would show if no audit has run since). Trigger one
run of each of `brief-fidelity-auditor`, `content-quality-auditor`,
`site-review-agent`, `visual-design-auditor` against any live site and re-run the
query. Also re-read the council verdict once queued:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='50ee4b26-2303-4304-b437-7320e1368a1d' AND kind='council_report'
ORDER BY created_at;
```

## §13 — Round 1 verdict was REVISE, answered; round 2 resubmission did NOT dispatch (kubeconfig expired mid-attempt)

**Council round 1 read (2026-08-13, same day as §12): REVISE**, gated by
`editquality`, with `guardian`/`debug_historian`/`prior_art_librarian` also
objecting (full report: `diagnosis_artifacts` kind=`council_report`, correlation
`50ee4b26-2303-4304-b437-7320e1368a1d`). Two of the three HIGH-severity
objections (the "duplicate active row / wrong version" landmine, raised by
`editquality` and `guardian` independently) were **already refuted by the
round's own embedded read-only checks** — for all four agent types,
`active_rows=1, max_version=1, min_version=1`, so the landmine's precondition
does not hold here. The other objections (fallback claim ungrounded, no
full-caller inventory, no owning-pipeline naming, migration not ledger-recorded,
no automated test) were real gaps, not refutations of the fix itself — every
reviewer who voted `object` still called the core direction sound.

**All objections answered with fresh evidence before resubmitting:**
- Removed fallback grounded: `git show 3621ca7cf~1:platform/orchestration/actions/write_audit_findings_action.go`
  shows the real pre-fix lines.
- **Full, unfiltered inventory** (no `is_active`/`is_snapshot`/`deleted_at`
  filter at all) of every `agent_definitions` row with a `write_audit_findings`
  step: still exactly the same four rows, each `version=1`, each `is_active=true`.
  No fifth caller anywhere, active or not.
- **Owning pipeline named, honestly, including a real gap it surfaced**:
  `improvement-loop` → `site-review-agent` → `content-quality-auditor`;
  `design-audit-agent` → `visual-design-auditor` and → `content-quality-auditor`.
  **`brief-fidelity-auditor` has NO live caller anywhere** — not in any active
  agent's workflow, no `scheduled_tasks` row, 0 orchestrations all-history. This
  fix is correct for it but currently inert — a pre-existing wiring gap this bug
  did not create and does not need to solve, but worth its own note if anyone
  goes looking for why brief-fidelity findings never appear at all.
- Migration 399 had been applied by hand (`psql -f -`) and was **not in the
  `schema_migrations` ledger** — recorded this session via
  `./scripts/migration/run-migrations.sh --record-only 399_four_auditors_audit_source_resolves_to_a_real_value.sql --note '...'`.
- New unit test added and committed (`29ae07500`,
  `write_audit_findings_input_spec_test.go`): reproduces the exact pre-fix
  unresolvable-string shape and asserts it now errors naming `audit_source`,
  plus a no-op-case companion confirming migration 399's actual dot-path shape
  still resolves.

**Round 2 resubmission (`RESUBMIT_CORR=50ee4b26-...`) DID NOT ACTUALLY DISPATCH.**
The trigger script prints `SAVE: SUBMISSION_CORR=... RUN_ORCH_ID=...` **before**
the kafka publish step (a `kubectl -n kafka run ... kcat -P` — confirmed by
reading the script, `097_TRIGGER_council_review_v1.sh` lines 167 vs 170-173), and
that publish failed with `error: You must be logged in to the server
(Unauthorized)`. **This is the known kubeconfig-token-expiry landmine, not a
submission-content problem**: the prod token expired at `2026-08-13 19:05:20`
(confirmed by decoding it — see `~/.claude/…/memory/kubeconfig-token-expires-every-3-days.md`
for the exact check), and only the owner can refresh it. `kubectl -n
ai-persona-system get pods` fails the same way, confirming it's total auth
loss, not a scoped permission issue. **Do not treat the printed
SUBMISSION_CORR/RUN_ORCH_ID from that attempt as real** — they were generated
client-side and never reached the council. The round-2 submission JSON is saved
at `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/submission_264_audit_source_round2.json`,
ready to fire once kubectl auth is restored:
```bash
RESUBMIT_CORR=50ee4b26-2303-4304-b437-7320e1368a1d \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/submission_264_audit_source_round2.json
```

**Everything requiring cluster/DB access is blocked until the token is
refreshed**, including: actually dispatching round 2, re-reading any verdict,
and the §12/original "How to verify a fix" live audit-run check above. Nothing
code-side is blocked — the fix itself is unaffected by this outage; it is
purely a review/verification-tooling gap.

## §14 — kubectl restored; live verification run, round 2 REVISE (answered), round 3 submitted

**Live "How to verify a fix" check — DONE, 3 of 4 confirmed by a real
site_work_item.** Dispatched one live run each of `site-review-agent`,
`visual-design-auditor`, `brief-fidelity-auditor`, `content-quality-auditor`
against `mortgagecalculator.co.uk` (all 4 orchestrations `COMPLETED`, no
error). Result:
```sql
SELECT spec->>'audit_source', count(*) FROM site_work_items
WHERE created_at > '2026-08-13T21:00:00Z' AND spec ? 'audit_source' GROUP BY 1;
--  brief-fidelity-audit  | 8
--  content-quality-audit | 4
--  visual-design-audit   | 5
```
`[MEASURED]` Three distinct, correct values — the fix works end to end for
these three. **`site-review` is absent, but NOT because the audit_source fix
failed** — `collected_data#>>'{audit_source_literal,audit_source}'` on that run
reads `site-review`, exactly right. It never reaches `site_work_items` because
a **separate, pre-existing defect** stops the write step from creating any item
at all: filed as **`bugs_open/272`** (site-review-agent's `findings_field`
config points at the LLM response's wrapping object, not the `findings` array
inside it, and `write_audit_findings`'s parse switch has no case for a raw
object — so `items_created` is silently `0` regardless of `audit_source`).
**This does not implicate this bug's fix** — confirmed by reading the exact
`collected_data`, not inferred.

**Council round 2 (resubmitted once kubectl access returned): REVISE**, gated
by `prior_art_librarian` (high), with `debug_historian`/`editquality` also
objecting (all three answered in round 3, submitted — see below).
`guardian`/`reuse_agent`/`guidelines`/`constitution`/`mission`/`improvement_guardian`/`architecture`
all approved round 2. The gating point: round 2's rebuttal to the
duplicate-active-row objection was never reconciled against the standing
LANDMINE naming "four agent types" with duplicate active rows — a fair catch,
since the coincidence of "four" was never checked.

**Settled directly, live, 2026-08-14**: the landmine
(`LANDMINES.md`, "Four agent types have TWO active definition rows…") names
its four **explicitly**: `chief-strategist`, `content-creator`,
`content-creator-contact`, `site-component-architect` (measured 2026-08-09) —
none of which is one of this bug's four auditors. Re-running the landmine's own
exact check fleet-wide, live, confirms it:
```sql
SELECT type, count(*) FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
GROUP BY 1 HAVING count(*) > 1;
--  chief-strategist=2, content-creator=2, content-creator-contact=2, site-component-architect=2
```
Exactly those four, no others — zero overlap with the audit four. Also checked
this round: no numbering collision on migration 399 (`ls
docs/agent_docs/sql_for_agents/399*` → exactly the migration + its ROLLBACK,
one matching `schema_migrations` row).

**Round 3 submitted** (`RESUBMIT_CORR=50ee4b26-2303-4304-b437-7320e1368a1d`),
answering all of round 2's objections with the evidence above. Submission JSON
saved at
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/submission_264_audit_source_round3.json`
(round 2's at `..._round2.json`). Verdict not yet read as of this write-up.
