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
