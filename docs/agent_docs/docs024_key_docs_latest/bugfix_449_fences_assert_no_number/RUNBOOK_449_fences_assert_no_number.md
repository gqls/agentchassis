# RUNBOOK — bug 449, fences that assert no number

Every query and command that was hard to get right, with its gotcha attached.
Change it HERE when it changes, not in your scrollback.

---

## 1. Is the bug still true? (the two load-bearing facts)

### 1a. Do the fence-authoring agents know the correctness check type?

```sql
SELECT type,
       (default_config::text LIKE '%computed_values%') AS knows_computed_values,
       (default_config::text LIKE '%interaction%')     AS knows_interaction,
       updated_at
  FROM agent_definitions
 WHERE default_config::text LIKE '%```criteria%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
 ORDER BY type;
```

⚠ **GOTCHA — `updated_at` here is a trap.** Both rows currently read
`2026-09-03 08:56:53.045885+00`, which looks like "the prompt was revised this morning" and
is not. Prove it before you quote it:

```sql
SELECT count(*) FROM agent_definitions WHERE updated_at = '<the timestamp>';
```

**208** rows share that exact second — a bulk touch, not a content change. A timestamp is
not a diff.

⚠ **GOTCHA — the backtick fence in psql via `kubectl exec`.** The literal ` ```criteria `
must reach psql with its backticks intact. Inside a bash double-quoted heredoc-less `-c`
they are command substitution and will execute. Escape them (`\`\`\`criteria`) or use
single quotes around the whole SQL.

### 1b. The census — how many fences assert nothing?

```sql
WITH f AS (
  SELECT dp.id, dp.created_by, dp.created_at,
         substring(dp.body from '```criteria(.*?)```') AS fence
    FROM doc_plans dp
   WHERE dp.subject_type='tool' AND dp.is_current
     AND dp.body LIKE '%```criteria%'
)
SELECT created_by,
       count(*) AS fences,
       count(*) FILTER (WHERE fence NOT LIKE '%text_matches%'
                          AND fence NOT LIKE '%expect_values%')     AS assert_no_value,
       count(*) FILTER (WHERE fence LIKE '%expect_values%')          AS uses_computed_values,
       count(*) FILTER (WHERE fence LIKE '%"fill"%'
                           OR fence LIKE '%"select"%')               AS drives_inputs,
       count(*) FILTER (WHERE (fence LIKE '%"fill"%' OR fence LIKE '%"select"%')
                          AND fence NOT LIKE '%text_matches%'
                          AND fence NOT LIKE '%expect_values%')      AS drives_but_asserts_nothing,
       min(created_at)::date AS first, max(created_at)::date AS last
  FROM f
 GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **GOTCHA — measure NEWLY CREATED fences by a `created_at` window, never by the total.**
The bug's §6 says so and it is the whole verification design: the standing stock does not
change itself, so a fix that works shows up as a fall in the *new* rows and leaves the
total almost flat for weeks. Compare like this, not by re-running the totals:

```sql
... AND dp.created_at > '<the moment the prompt change was applied>'
```

⚠ **GOTCHA — `max(created_at)` is the freshness signal, and it is the one to read.**
`last = today` is what proves this is a live intake rather than a backlog. Quote it.

⚠ **GOTCHA — `substring(... from '```criteria(.*?)```')` is non-greedy on purpose.** A PLAN
body can contain a second fenced block; a greedy match swallows the rest of the document and
every `LIKE` after it silently changes meaning. This mirrors `extractCriteriaFence` in
`check_tool_acceptance.go:602`, which takes the **first** fence only.

---

## 2. Who authors a fence, and where does it land?

```sql
-- the agents that can write a PLAN at all
SELECT type FROM agent_definitions
 WHERE default_config::text LIKE '%write_doc_plan%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
 ORDER BY 1;
--  → experience-planner | experience-register-writer | tool-generator
```

```bash
# every production Go writer of a PLAN BODY — the seam, enumerated not assumed
grep -rn "INSERT INTO doc_plans\|UPDATE doc_plans" --include='*.go' . | grep -v _test
```

⚠ **GOTCHA — three of the hits are not the door.** `travelling_docs_rekey.go` only rewrites
`subject_key`; the copy under `docs/agent_docs/docs024_key_docs_latest/travelling_docs/` is a
doc artefact and is not compiled. The door is
`platform/orchestration/actions/write_doc_plan_action.go` alone.

⚠ **GOTCHA — the door is not total.** Operator scripts (`install_fences.py` in the mcalc
lane and its ancestor in `loanandmortgagecalculator_couk`) write `doc_plans` over `psql`
directly and never touch the Go action. So a guard at the door governs every *generated*
fence and no hand-installed one. Do not describe it as covering all fences.

---

## 3. Dump an agent's live prompt to read it properly

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -tA -c "SELECT default_config::text FROM agent_definitions
          WHERE type='tool-generator' AND is_active
            AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > tool-generator.json

python3 -c "
import json; d=json.load(open('tool-generator.json'))
print(d['workflow']['steps']['compose_plan']['config']['prompt_template'])"
```

⚠ **GOTCHA — do not grep the JSON for the vocabulary and stop there.** The criteria
instructions are one 2,766-char `prompt_template` inside
`workflow.steps.compose_plan.config`; `experience-planner` has **six** prompt templates
across `compose` / `reframe` / `recompose` / `review_journeys` / `review_contracts`. A grep
that hits one tells you nothing about the other five. Walk the JSON and print the paths.

---

## 4. Induce the red — the bar bug 449 §6 sets

> *"Take a passing tool, change one constant in its JS so it computes a wrong figure,
> re-run acceptance. Today it still passes. After the fix it must fail. A fix that only
> adds checks which pass is indistinguishable from no fix."*

The existing worked machinery, to reuse rather than rebuild:

- `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/acceptance/verify_criteria.py`
  — re-derives every pinned value from a non-page source at three labelled strengths
  (DEFINITION / REGISTER / CONVENTION) and carries `--mutate <fact-id>=<value>`, which
  perturbs one registered fact in memory and re-runs. **It exists because 80 agreements
  prove nothing about whether the register is load-bearing** — only the mutation
  distinguishes a real read from a hard-coded fallback.
- `docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail_tool_fuel_cost_estimator.go`
  — the existence proof that a fence CAN be made to fail.

⚠ **GOTCHA — a detector step needs a demand control in the same run.** A zero from a new
sweep means "nothing found" and "the sweep is blind" equally. Pair every clean result with
something that must still be found (e.g. the 55 known-blind fences), or the zero grades
nothing.

---

## 5. Where the code lives

| what | path |
|---|---|
| the write door (only Go writer of a PLAN body) | `platform/orchestration/actions/write_doc_plan_action.go` |
| criteria validator, rules P1–P11 | `platform/orchestration/actions/experience_criteria.go` |
| …its only production caller today | `platform/orchestration/actions/write_experience_pattern_action.go:217` |
| Tier 2 static sweep + `needs_criteria` notes | `platform/orchestration/actions/discovery_checks/check_tool_acceptance.go` |
| Tier 4 browser runner, `runComputedValues` | `internal/adapters/browserrunner/run_checks_action.go` (~760+) |
| who is eligible for the ladder at all | `platform/orchestration/actions/discovery_checks/tool_eligibility.go` |
| Tier 4 routing / `no_auto_fix` | `platform/orchestration/actions/tool_acceptance_actions.go` |

⚠ **GOTCHA — the two `run_checks_action.go` files are different files.** The runner is
`internal/adapters/browserrunner/run_checks_action.go`. There is no
`platform/orchestration/actions/run_checks_action.go`; bug 449's §1 cites line numbers
against the browserrunner one.

---
