# RUNBOOK — register↔tool fact drift (Phase B)

Every command here had a gotcha attached the first time. Read the gotcha, not just
the command.

## Build and test on this shared tree

**`go build ./...` at the working tree may fail for reasons that are not yours.** On
2026-08-16 another session's untracked `component_name_resolver_menu.go` did not
compile. Build against a clean HEAD plus only your own files:

```bash
SC=<scratchpad>; rm -rf $SC/headtree; mkdir -p $SC/headtree
git archive HEAD | tar -x -C $SC/headtree
for f in <your changed files>; do cp $f $SC/headtree/$f; done
cd $SC/headtree && go build ./... && go test ./platform/orchestration/... -count=1
```

⚠ `go vet` reports a **pre-existing** `unreachable code` in
`load_component_library_actions.go`. It is at HEAD and not yours — check before
chasing it.

## Prove a guard can actually fail (do not skip this)

A passing test proves nothing about a guard until you have watched it fail. In the
headtree copy, delete the guard, run the one test, expect FAIL, restore:

```bash
cp $F /tmp/orig                      # keep the original
python3 - <<'PY' ... remove one guard ... PY
go test ./platform/orchestration/actions/ -run TestClassifyFactDrift_NonForkRoutesToHuman -count=1   # must FAIL
cp /tmp/orig $F                      # restore, re-run, must pass
```

Six routing guards were proven this way on 2026-08-16 (no_auto_fix, fork,
evidence-vs-value, fetch-error, baseline-present, baseline-precedence), plus P11.

## Measure the fleet — always with a positive control

The `?` jsonb operator silently returns 0 for a key you spelled wrong, which is
indistinguishable from "nothing has one". Run the control in the same breath:

```sql
SELECT count(*) FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
 WHERE ss.is_current AND ss.aspect='evidence_base' AND f->'source' ? 'artifact_check';  -- 0
SELECT count(*) FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
 WHERE ss.is_current AND ss.aspect='evidence_base' AND f->'source' ? 'citation';        -- 61 ← control
```

## Run the ladder's eligibility predicate yourself

Before assuming a tool is visible to any acceptance machinery — it very often is not.
The predicate lives in `discovery_checks/tool_eligibility.go` (`toolEligibilityWhere`);
paste it into a query scoped to the site. On 2026-08-16 it returned neither
stamp-duty tool.

## Does a tool have a PLAN, and what does its fence say?

```sql
SELECT subject_key, is_current, created_by,
       body ~ '```criteria' AS has_fence,
       body ~ 'no_auto_fix"?: ?true' AS no_auto_fix,
       body LIKE '%"facts"%' AS declares_facts
FROM doc_plans WHERE subject_type='tool' AND subject_key IN ('stamp-duty','mortgages-stamp-duty');
```

⚠ **The subject key is not the page name.** mortgagecalculator's page is
`tool-stamp-duty`; its PLAN key is `stamp-duty`. LMC's page and key are both
`mortgages-stamp-duty`.

⚠ **Never hand-edit a `doc_plans` body to add `facts`.** Both lanes' `install_fences.py`
rewrite the whole body on `--apply`, so a hand-added key is lost on the next install.
Add it to the lane's criteria JSON and re-install.

⚠ **`install_fences.py` REFUSES the install for exactly the tools this mechanism serves,
and it refuses SILENTLY** (found by the mcalc lane, 2026-08-17):

```
SKIP     stamp-duty   not ladder-eligible on this site — a PLAN here would never be read
```

Its rule 2 keys on the acceptance ladder's eligibility, and a decomposed or ported tool
is not eligible — which is precisely why the fan-out does NOT key on
`toolEligibilityWhere`. So the installer's refusal rests on a premise CLM-022 made false:
since Piece 3 a declaring PLAN *is* read, by the name rule. A lane following "just
re-install" literally gets a clean-looking run, no error and no `facts` key — and the
verification query (`body LIKE '%"facts"%'`) returns `f` with nothing explaining why.

The mcalc lane added `--allow-ineligible`, fenced on BOTH: the criteria document must
actually declare `facts`, and a current `doc_plans` row must already exist under that key
(so the subject key is inherited, never constructed from a page name). **LMC's
`mortgages-stamp-duty` will hit the identical wall** — 3 components since B2.

## Fire a one-off dry run of the sweep

Not within 300 s of a chassis pod restart (the spawn is silently dropped). Publish to
`system.agent.generic.requests`, ONE line of JSON, with an inline workflow — `dry_run`
is read from STEP config, not `input_data`:

```json
{"action":"orchestrate","config":{"workflow":{"start_step":"refresh_evidence","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"refresh_evidence":{"action":"refresh_evidence_base","config":{"dry_run":true},"next_step":"complete","output_field":"refresh_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["refresh_result"]}}}}},"input_data":{"site_id":"<site>"}}
```

Read it back by payload, not by a printed id:

```sql
SELECT status, jsonb_pretty(collected_data->'refresh_result')
FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_evidence_base%'
ORDER BY created_at DESC LIMIT 1;
```

A dry run **plans** the fan-out and marks each emission `dry_run`; it writes nothing.

## Induce the fan-out (the only proof that matters)

Supersede the fact, dry-run, restore. `pinned` must be carried forward (CLM-001), and
check `writer_block_managed` first — if true, the daily sweep will regenerate the writer
block with your test number, so keep the window short and do it outside 09:00–09:10 UTC
(the sweep's own CAS window). Put the restore in a `trap … EXIT` so it fires whatever
happens, and flip `is_current` back onto the ORIGINAL row rather than inserting a copy —
the mcalc lane's shape, and it makes the restore exact.

> **⚠ CORRECTED 2026-08-17 — this step used to predict `kind: value_drift`, and that is
> UNREACHABLE on a freshly-seeded declaration.** Every (fact, tool) pair with no prior
> finding is `never_reconciled`, and that arm is tested FIRST, so a newly-declared fence
> yields `kind: unreconciled_declaration` on BOTH the baseline and the induced run. The
> mcalc lane ran this recipe as written and got a result my own text called a failure.
> **`value_drift` only becomes inducible after a REAL (non-dry) sweep has recorded the
> baselines** — a dry run writes no items, so it can never create them.

**What a PASS looks like on a fresh declaration** (measured by the mcalc lane, 16:17Z
2026-08-17, and it is a genuine proof):

| run | `fact_drift` entries | kind | `new_value` for the changed fact |
|---|---|---|---|
| baseline (register 500000) | 13 | `unreconciled_declaration` | **500000** |
| induced (register 550000) | 13 | `unreconciled_declaration` | **550000** |

The discriminator is **not** the kind — it is that `new_value` tracks the register
between the two runs. That proves the fan-out reads the register AT CHECK TIME rather
than replaying a stored number, which is the property the whole design turns on.

**To reach `value_drift`:** let one real sweep file the one-time items (they become the
baselines and self-quiet), then induce again. **DONE 2026-08-17 18:30Z — both remaining
properties are now proven:**

| check | result |
|---|---|
| self-quieting: dry run with baselines set, register unchanged | **0 entries** (13 facts checked) — the identical run before the baselines produced 13 |
| `value_drift`: baseline moved 500000→450000, register left at 500000 | **1 entry**, `kind: value_drift`, `old_value: 450000`, `new_value: 500000` |
| per-fact isolation | exactly **1** entry — the other 12 declared facts stayed silent |
| written | **nothing** — 13 items before and after, 0 of kind `value_drift`, register still 500000, item spec restored byte-equal |

**⚠ Move the BASELINE, not the register.** The comparison is symmetric
(`abs(new-baseline) > 1e-9`), so moving the item's recorded `spec.fact.new_value` exercises
the identical branch — without touching a live tax figure on a public site for even ninety
seconds. Capture the pre-image first (`SELECT spec::text`), restore in a `trap … EXIT`, and
assert the restored spec parses **equal to** the pre-image rather than eyeballing one field.

> **⚠ CORRECTED 2026-08-17 (third time this recipe was wrong — see WRONG_CALLS): the
> `reason` is `not_a_fork`, NOT `no_auto_fix`.** Both conditions are true of
> `tool-stamp-duty` (non-fork AND a `no_auto_fix` fence), and the fork guard is tested
> FIRST, so it wins. The route is the same either way; only the recorded reason differs.
> **A `switch` reports the first arm that matches, not the most important one** — if you
> want to know why something routed, read the order, not the set of true conditions.

**⚠ Where to look, because the counter next door lies.** `fact_drift` is per-site and
NESTED — `refresh_result->'results'->N->'fact_drift'`. There is no top-level key, and
top-level `total_drifted` counts CITATION drift: it reads **0** while each site carries
13 entries. Read it with a path query:

```sql
SELECT jsonb_pretty(jsonb_path_query_array(collected_data,
         '$.refresh_result.results[*].fact_drift[*]'))
FROM orchestration_states WHERE correlation_id='<corr>';
```

**A dry run that reports nothing after a real change is still the failure** — just make
sure you looked in the right place before concluding it.


## Prove the code is live

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -ac fact_drift_review /proc/1/exe   # expect >0
kubectl -n ai-persona-system exec $POD -- grep -ac stale_attestation /proc/1/exe   # positive control
```

Never `strings` (absent from the image), never a discovery grep for "some 40-hex
string", and always run the control in the same exec.
