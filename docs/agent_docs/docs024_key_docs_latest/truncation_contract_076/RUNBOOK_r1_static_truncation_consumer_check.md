# RUNBOOK — the truncation-contract checks (bugs_closed/076 R1)

Every command here was run on 2026-07-26 and its gotcha is attached. If one
changes, change it HERE.

## The two checks

```bash
# LIVE FLEET — the authority. Read-only, no writes, no credits, no LLM.
python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/103_LINT_truncation_consumer.py
python3 …/103_LINT_truncation_consumer.py --verbose     # + each guarded step and its guard
python3 …/103_LINT_truncation_consumer.py --self-test   # predicate vs fixtures, no cluster at all
python3 …/103_LINT_truncation_consumer.py --strict      # exit 1 on findings, for a caller that gates

# COMMIT TIME — runs itself, inside pattern-check.py via .githooks/pre-commit.
scripts/pattern-check.py                # staged changes
scripts/pattern-check.py --commit <sha> # audit one past commit
```

**Run the live lint after ANY seeding that touches agent config.**
`run-migrations.sh --apply` now names the applied files that touch the flag and
prints this command; that pointer is the reminder, not a check.

**A CLEAN RUN IS NOT EVIDENCE THE CHECK WORKS.** The fleet has been clean since
the guard shipped (37/37 guarded), so "clean" is what a broken check prints too.
`--self-test` is the evidence; the induced probes below are the stronger evidence.

## Inducing a real offender (how R1 was validated, 2026-07-26)

The three probes differ only in the downstream step, which is the point — a check
that flagged all tolerance would look identical on the offender alone.

```sql
BEGIN;
INSERT INTO agent_definitions (type, display_name, category, agent_category, default_config, is_active)
VALUES
 ('truncation-lint-probe-offender', 'R1 probe: tolerance, no reader', 'probe', 'specialist', '{
   "workflow": {"start_step": "ask", "steps": {
     "ask":  {"action": "execute_llm_prompt", "config": {"max_tokens": 16, "tolerate_truncation": true}, "next_step": "done"},
     "done": {"action": "complete_workflow", "config": {}}}}}'::jsonb, true),
 ('truncation-lint-probe-guarded', 'R1 probe: tolerance, hatch reader', 'probe', 'specialist', '{
   "workflow": {"start_step": "ask", "steps": {
     "ask":  {"action": "execute_llm_prompt", "config": {"max_tokens": 16, "tolerate_truncation": true}, "next_step": "done"},
     "done": {"action": "complete_workflow", "config": {"accepts_truncated": true}}}}}'::jsonb, true),
 ('truncation-lint-probe-inert', 'R1 probe: string-valued flag', 'probe', 'specialist', '{
   "workflow": {"start_step": "ask", "steps": {
     "ask":  {"action": "execute_llm_prompt", "config": {"max_tokens": 16, "tolerate_truncation": "true"}, "next_step": "done"},
     "done": {"action": "complete_workflow", "config": {}}}}}'::jsonb, true);
COMMIT;
```

Expected, and observed exactly:

| probe | lint says |
|---|---|
| `-offender` | `truncation-tolerance-no-reader`, exit 1 under `--strict` |
| `-guarded` | not flagged; listed under `accepts_truncated declarations` as an unverified claim |
| `-inert` | `inert-flag` — Go reads a string `"true"` as false, so no tolerance is in force |

**DELETE THEM AFTERWARDS — they are `is_active` rows in the live fleet:**

```sql
DELETE FROM agent_definitions WHERE type LIKE 'truncation-lint-probe%';
```

Gotchas, each one paid for:

- `category` and `display_name` are **NOT NULL** with no default.
- `agent_category` is constrained by `check_ad_category` to
  strategist / executor / analyst / integrator / coordinator / specialist —
  `'system'` is rejected. (`category` itself is free text; `'probe'` is fine.)
- Nothing dispatches to a type nobody sends to, so an idle probe is inert — but
  it is still a live definition. Keep the window short and delete by `type LIKE`.

## Proving 076 is still live after a chassis roll

R1 is scripts and docs — a roll cannot change it. The **guard** is in the binary,
so after any roll:

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c '
  echo -n "guard    : "; strings /app/agent-chassis | grep -c "no step in this workflow consumes the __truncated marker"
  echo -n "REFUSED  : "; strings /app/agent-chassis | grep -c "REFUSED (bugs_open/076"
  echo -n "degrade  : "; strings /app/agent-chassis | grep -c "TRUNCATION_DEGRADED_REVIEW"
  echo -n "negative : "; strings /app/agent-chassis | grep -c "this_string_should_not_exist_076"'
```

**Read the first three as `>= 1` and the last as `== 0`. Do NOT assert an exact
count of a common substring** — `strings` prints packed data blobs, not one
literal per line, so a count moves between builds with the source unchanged
(measured: `tolerate_truncation` gave 3 on v1.0.1171 and 4 on v1.0.1172 with an
empty `git diff` across all three 076 files). A control that fails on a correct
binary is worse than no control.

Then confirm the code behind it did not move under you, which is the half a
pod-grep cannot tell you:

```bash
git diff --stat <last-known-good-sha> HEAD -- \
  platform/orchestration/actions/ai_actions.go \
  platform/orchestration/actions/truncation_guard.go \
  platform/orchestration/actions/diagnose_council_decide_action.go
```

Unattended evidence that the consumer half is still working — this grows on its
own, and is stronger than any greppable string:

```sql
SELECT count(*), max(occurred_at) FROM agent_error_log
WHERE error_code = 'TRUNCATION_DEGRADED_REVIEW';   -- 3 on 07-26 21:15Z, 6 by 07-27
```

## Grounding queries (re-run these; they go stale in days)

```sql
-- the blast radius, per agent, with the guard verdict
WITH steps AS (
  SELECT d.type, e.k AS step, d.default_config->'workflow'->'steps' AS all_steps
  FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') e(k,v)
  WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
    AND e.v->>'action'='execute_llm_prompt'
    AND COALESCE(e.v->'config'->>'tolerate_truncation','false')='true')
SELECT type, count(*) FROM steps GROUP BY type ORDER BY type;
-- 2026-07-26: council-gate 16, fix-proposer 16, feature-designer 5 = 37, all guarded.
```

**Do NOT hand-copy the reader action list into a query.** That list is
`truncationAwareActions` in `platform/orchestration/actions/truncation_guard.go`;
`scripts/truncation_registry.py` parses it, and both checks read it from there.
The handoff's own measurement query hard-codes the two names — that copy is
already a liability, and a query is the easiest place for one to go stale
unnoticed. To see what is registered right now:

```bash
python3 scripts/truncation_registry.py
```

Two coverage questions worth re-asking, with the queries that answer them:

```sql
-- 1. does any live type have TWO active rows? (the lint keys by id because of this)
SELECT type, count(*) FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
GROUP BY type HAVING count(*)>1;                         -- 5 such types on 2026-07-26

-- 2. does the flag EVER appear outside workflow.steps? (if it does, the lint is blind to it)
SELECT sum((length(default_config::text)-length(replace(default_config::text,'tolerate_truncation','')))
           / length('tolerate_truncation')) AS occurrences
FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- 37 occurrences vs 37 steps found by the lint on 2026-07-26 => full coverage, that day.
```

## Falsifying the checks (do this before believing either)

```bash
# the parser: mutate a COPIED tree, never the real one
T=/tmp/regprobe; rm -rf $T; mkdir -p $T/platform/orchestration/actions
cp platform/orchestration/actions/{truncation_guard.go,registry.go} $T/platform/orchestration/actions/
# … rename the map / empty it / rename the const / point an entry at a non-action …
REPO_ROOT=$T python3 scripts/truncation_registry.py     # expect exit 2 (or 1 for the unknown action)
```

`REPO_ROOT` exists for exactly this: the parser normally finds the repo with
`git rev-parse`, which would happily read the real file while you think you are
testing a mutation.

For the pre-commit check, measure over the whole corpus plus controls — the
harness used on 2026-07-26 loads `pattern-check.py` by path and calls the one
function:

```python
spec = importlib.util.spec_from_file_location("pc", "scripts/pattern-check.py")
pc = importlib.util.module_from_spec(spec); spec.loader.exec_module(pc)
f=[]; pc.check_truncation_without_reader(subprocess.run(["git","ls-files","*.sql"],
      capture_output=True,text=True).stdout.split(), None, f)
```

849 tracked `.sql` files → **0 findings**; the offender control → 1, naming the
step; registry-guarded, hatch-guarded, patch-style and string-flag controls → 0;
self-certifying and nested-workflow controls → 1 each, correctly attributed.
