# HANDOFF — the feature designer's plans die on a validator rule its prompt never states

**Filed:** 2026-07-26, by the `bugfix_077` thread, on the first real use of the
feature builder since it was proven end-to-end on 07-25.
**Severity:** MEDIUM. Not an outage. It silently destroys a completed, good design
and spends the full designer cost to do it — and the failure is *systematic*, not
flaky, so retrying without changing anything buys the same result.
**Status:** OPEN — **for candidate 2 only.** Candidate 1 (the one-line config
fix) IS applied, live, and re-verified after the `v1.0.1174` fleet roll on
2026-07-27; see §"Candidate 1 IS APPLIED" and §"Re-verified 2026-07-27". Anyone
arriving here from a summary that says "a one-line fix is identified but
deliberately not applied" is reading a stale pointer — that was true only on
2026-07-26 morning and the file records the reversal.

---

## The mechanism

`diagnose_persist_fix_plan` refuses a staged plan in which the same file appears
twice in one stage:

```go
// platform/orchestration/actions/diagnose_persist_fix_plan_action.go:526-532
f := strings.TrimSpace(e.File)
if f == "" { continue }
if seenPath[f] {
    problems = append(problems, fmt.Sprintf("%s: %s appears in more than one edit of this stage", tag, f))
}
seenPath[f] = true
```

The `feature-designer` prompt never states that constraint. Its plan rules cap
*quantities* — rule 4, verbatim: *"CAPS: at most 6 stages, 8 edits per stage, 24
edits total"* — and say nothing about uniqueness of `file` within a stage.
Measured 2026-07-26 against the live row:

```sql
SELECT CASE WHEN default_config::text ILIKE '%more than one edit%' THEN 'STATED'
            WHEN default_config::text ILIKE '%one edit per file%'  THEN 'stated (variant)'
            ELSE 'NOT STATED' END
FROM agent_definitions WHERE type='feature-designer'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- NOT STATED
```

So a designer that splits one file's work into two readable edits — *"add the
helper"* and *"add the entry point"* — produces a plan that is coherent, within
every stated cap, and unpersistable. It cannot know.

## Live evidence

Run `fix_correlation_id 42665066-b492-4ed8-9ea3-d95fac57bc5a`, work item
`7b89fb35-f42c-45d1-b64d-214aff56d918` (`bugs_closed/077`'s colour-fixer
capability gap).

- `check_spec_approved` → `{"condition_met": true, ...}` — **the owner-approval
  gate passed**; this is not a rejected submission.
- The designer ran to completion and produced a design AND a proposal.
- `__step_error`:
  *"step persist_plan failed: failed to execute action diagnose_persist_fix_plan:
  staged plan failed validation: stage s1 edit 2:
  platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go
  appears in more than one edit of this stage"*
- Workflow routed to `complete_refused`. **No `fix_plan` artifact was written**, so
  the design is only recoverable from `orchestration_states.collected_data->'design'`.

**The design that was lost was a good one**, which is what makes this worth
filing rather than shrugging at. Its own risk note:

> *"The plan deliberately answers none of the spec's three open design questions
> (bare #fff, semantic tints without existing palette entries, inline style=""
> attributes) — it only claims the two cases the evidence explicitly supports
> (exact matches to palette.background and palette.background_alt); a reviewer
> should reject any stage that widens further than that without an owner decision
> on those questions."*

Its s1 was: *add `derivePaletteColorMap`, add `replaceCSSColorsWithPalette`, leave
`replaceCSSColors` untouched so the tree still builds* — two edits, one file, one
stage. Exactly the shape the validator forbids and the prompt permits.

## This is the same class as `bugs_closed/077`

077 was a DETECTOR whose predicate was wider than its HANDLER's remit, filing work
the handler could never do. This is a PRODUCER whose output space is wider than its
CONSUMER accepts, emitting plans the persister can never store. Same shape, one
level up: **two ends of one contract, written at different times, with nothing
holding them in lockstep.** 016b §9's rule generalises directly — *"ask whether the
thing you are checking is the thing the other end promised to do"* — and the
remedy is the same: state the constraint where the producer can see it, or derive
one end from the other.

## Fix candidates, cheapest first

1. **Tell the designer.** Extend rule 4 in the `design` step's `prompt_template`:
   *"…and AT MOST ONE EDIT PER FILE PER STAGE — combine multiple changes to one
   file into a single edit whose sketch describes them all; a repeated path fails
   validation and the whole plan is discarded."* Config-only, live immediately, no
   image. Also worth adding to `reframe`/`repropose` if they carry their own copy
   of the rules — **check, do not assume**.
2. **Make the refusal recoverable.** `persist_plan` failing currently discards a
   completed design. Route the validation problem back into `repropose` (which
   exists) with the problem text, so the designer can merge the duplicate edits
   instead of the run dying. This is the durable fix; (1) only lowers the rate.
3. **Validate at the source.** Have the designer's own output schema check
   uniqueness before `persist_plan` sees it, so the failure names itself at the
   step that produced it.

Do **not** "fix" this by relaxing the validator: one edit per file per stage is
what makes a stage a single reviewable commit, and the caged implementer applies
edits per file.

## Candidate 1 IS APPLIED — migration `222`, 2026-07-26

> **This section previously said "WHY NOT APPLIED HERE", on the reasoning that the
> `feature-designer` row belongs to the feature-builder workstream and editing
> another workstream's live agent mid-flight is the collision CLAUDE.md warns
> about. That reasoning was sound and I changed the decision, so I am recording
> the change rather than quietly rewriting it.** What changed: the defect blocks
> the owner's standing instruction to drive a capability gap through the designer's
> gate, and it is systematic rather than flaky — the natural decomposition of that
> feature (add a helper, add an entry point, one file) reproduces it, so retrying
> without a change just spends another designer run to reach the same refusal.

`docs/agent_docs/sql_for_agents/222_feature_designer_one_edit_per_file_per_stage.sql`
— applied and verified. It takes a `snapshot_agent` first, extends rule 4 **in
place** with `replace()` on the exact sentence rather than rewriting the template
(so a concurrent edit elsewhere in the prompt survives), and **RAISEs if the
replace did not bite** — a silent no-op would otherwise leave the bug in place
while reporting success. `reframe` and `repropose` were checked and carry no copy
of the caps rule, so there is no second edit to keep in lockstep.

**VERIFIED on the failing branch, same day.** The identical submission (work item
`7b89fb35`) was re-fired after the migration as
`FEATURE_CORR c91bb061-250e-4a1a-819f-78c625733956`. It cleared `persist_plan` and
**wrote a `fix_plan` artifact** — the exact discriminating test this file specifies
below, and the thing that did not exist on the first run. The plan it persisted is
three stages:

```
s1  Derive a palette-based color map inside the existing color-matching machinery
s2  Flip the pinned remit test cases the widening is supposed to flip, and prove
    non-matches stay unmapped
s3  Wire the handler to fetch the site's palette into the same shared function
```

Note s2: the designer read the code_pointer warning that
`check_hardcoded_section_colors_test.go`'s `withinRemit: false` fixtures pin the
CURRENT remit and are the ones a widening must deliberately flip — and made that a
gated stage of its own rather than quietly editing the test to pass. Which is also
evidence for the other half of the lesson: the plan is only as good as the spec's
`code_pointers`, and those come from whoever files the gap.

**This is still only candidate 1, and it only lowers the rate.** Candidate 2 —
routing a `persist_plan` validation failure back into `repropose` instead of
discarding a completed design — remains the durable fix and is **not done**. The
bug stays OPEN for it. A designer that hits any *other* validator rule still loses
its design silently, and there are a dozen more rules in that validator.

**For the feature-builder workstream:** this is a live config change to your agent,
made from outside your lane. It is snapshotted and reversible (rollback in the
migration header). If you re-seed `feature-designer` from a `.sql`, fold the rule
into the seed or this fix is clobbered — the "config re-seed clobber" landmine.

## Re-verified 2026-07-27 — candidate 1 SURVIVED a fleet-wide agent re-seed

Checked by the bug-backlog triage sweep. This is the "config re-seed clobber"
landmine the section above warns the feature-builder workstream about, and it
came within a minute of firing, so it is worth recording that the check was run
rather than assumed.

**All 180 live `agent_definitions` rows were rewritten at ~15:10 UTC** — one
minute before the `v1.0.1174` pods came up at 15:11:15Z, i.e. a `deploy-agents`
seed sync riding with the roll:

```sql
SELECT count(*) FILTER (WHERE updated_at > NOW() - INTERVAL '1 hour') AS updated_last_hour,
       count(*) AS total
FROM agent_definitions WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- 180 | 180
```

**The rule survived it, and it survived in the right place** — not merely
somewhere in the row's JSON, but in the `design` step's own `prompt_template`,
which is what migration 222 edited:

```sql
SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
       ILIKE '%ONE EDIT PER FILE PER STAGE%' AS in_design_step
FROM agent_definitions WHERE type='feature-designer'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- t
```

Note the difference between that query and the one in §"The mechanism", which
tests `default_config::text`. **A whole-row `::text ILIKE` cannot tell you the
rule is on the path the designer reads** — it would pass just as happily if the
string had landed in `reframe`, or in a comment, or in a step that never runs.
Use the path-qualified form when re-checking this after any re-seed.

`schema_migrations` still records
`222_feature_designer_one_edit_per_file_per_stage.sql`, applied 2026-07-26
21:36:46, `applied_by='record-only'`. **Beware the number:** `222` is used twice
in `sql_for_agents/` (this one and `222_news_components_carry_their_own_css.sql`,
applied 18:40 the same day) — resolve by filename, never by number.

## Why this stays OPEN

Candidate 2 — routing a `persist_plan` validation failure back into `repropose`
rather than discarding a completed design — is **not done**, and it is the
durable fix. Candidate 1 only lowers the rate for one rule; the validator
(`diagnose_persist_fix_plan_action.go`) has a dozen more, and a designer that
trips any of the others still loses its whole design silently, with
`orchestration_states.error` NULL. That is a Go change, so it is inert until a
roll — it did **not** ship in `v1.0.1174`.

## How to verify a fix

Re-fire the designer on the same work item and require BOTH:

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh \
  7b89fb35-f42c-45d1-b64d-214aff56d918
```

1. a `fix_plan` artifact EXISTS for the new correlation
   (`SELECT kind FROM diagnosis_artifacts WHERE correlation_id='<FEATURE_CORR>'`), and
2. the run does NOT end at `complete_refused`.

**Both, not either.** The run "completing" is not evidence — this one COMPLETED,
at `complete_refused`, with no artifact and no error column set. `error` was NULL;
the reason lived only in `collected_data->'__step_error'`.

## Landmine for whoever picks this up

`orchestration_states.error` was **empty** on a run that failed a step. The real
message is in `collected_data->>'__step_error'`, and `final_result` was NULL too.
A dashboard keyed on `error` reports this run as clean.

## References

- `platform/orchestration/actions/diagnose_persist_fix_plan_action.go:526-532` — the rule.
- `bugs_closed/077` — same class, one level down; its capability gap is the work item above.
- `016b` §9 — *"A 'did the fix work?' check must assert the HANDLER's remit…"* and
  *"A detector must PARTITION its population by the handler's remit…"*.
