# RUNBOOK — bugfix 213, verifier/producer join

Commands that were hard to get right, with the gotcha attached. Change them HERE,
not in scrollback.

---

## Is a bug actually unowned? (the check I got wrong first)

`scripts/who-owns.py <n>` reads **commits**. On this tree that is a lagging surface —
a session mid-fix is invisible. Run the working-tree check FIRST, on the paths the
bug file names in its own §Cause:

```bash
python3 scripts/who-owns.py 213                       # commits — necessary, not sufficient
git status --short <each code path the bug cites>     # THE one that catches a live session
git status --short | grep -E '\.go$'                  # everything in flight right now
```

An untracked new file next to a bug's cited paths is the strongest ownership signal
there is — `write_site_plan_imagery_scope.go` was how I finally spotted 214's owner.

## The blast-radius query (213 §4), and the column that carries the answer

`item_type` cannot tell you who filed a row. `spec->>'audit_source'` can.

```sql
SELECT status,
       count(*) FILTER (WHERE spec->>'audit_source' = 'design-audit') AS producer_b,
       count(*) FILTER (WHERE spec->>'audit_source' IS NULL)          AS producer_a,
       count(*) AS total
FROM site_work_items WHERE handler_agent='color-variable-fixer' GROUP BY 1;
```

**Read the asymmetry, not the totals.** "11 complete" is not the finding; "11
complete and 0 that ever failed to close, while the other producer has 8 failures"
is. A route where one producer's items never fail is a route whose grader cannot see
that producer's defect.

## Find every verified item_type with more than one producer (the fleet check)

The generalisation of the above — run this before adding any producer to an existing
item_type:

```sql
SELECT item_type, count(*) AS rows,
       count(DISTINCT COALESCE(spec->>'audit_source','<none>')) AS distinct_sources,
       string_agg(DISTINCT COALESCE(spec->>'audit_source','<none>'), ', ') AS sources
FROM site_work_items
WHERE item_type IN (<the RegisterVerifier list>)
GROUP BY item_type ORDER BY distinct_sources DESC;
```

Get the item_type list from the code, not from memory — it drifts:

```bash
grep -rn "RegisterVerifier" platform/orchestration/actions/ --include=*.go \
  | grep -v _test | sed 's/.*RegisterVerifier\(WithPolicy\)\?(//'
```

**Gotcha:** `distinct_sources = 1` is only reassuring if the query can return 2. It
does — `hardcoded_section_colors` returns 2 — so the 1s are informative rather than
an artefact of the predicate.

## Prove a scope predicate partitions before you build on it

`Grades` keys on `spec.check`. Before writing that, prove it separates the producers
and could have come out otherwise:

```sql
SELECT status, (spec ? 'check') AS has_check_key, spec->>'audit_source', count(*)
FROM site_work_items WHERE item_type='hardcoded_section_colors' GROUP BY 1,2,3;
```

Then prove it holds historically, not just today — a key added last month would
disclaim every older row of the check's own:

```bash
git log --oneline -S '"check":            "hardcoded_section_colors"' -- <check file>
git show 62a79c8ac -- <check file> | grep -n '"check"'
```

## Building when the working tree is broken by another session

`go build ./...` in the live tree tells you nothing when someone else is mid-edit.
Build **committed HEAD plus only your files**:

```bash
S=<scratchpad>; rm -rf $S/ht && mkdir -p $S/ht
git -C /home/ant/projects/agentchassis archive HEAD | tar -x -C $S/ht
for f in <your files>; do cp "$f" "$S/ht/$f"; done
cd $S/ht && go build ./... && go test ./platform/orchestration/actions/... -count=1
```

**Two gotchas, both cost me a run.** `git archive HEAD` must be given the repo —
`git -C <repo> archive`, not a bare `git archive` after `cd`-ing to the scratchpad,
which is not a git repo and fails in a way that leaves you with an almost-empty
directory. And `go build ./...` inside that directory only works from its root; a
`cd` that gets reset between commands silently runs it somewhere else.

## Proving a test actually guards the fix

A passing test is not evidence. Mutate the fix away and require RED:

```bash
cp <file> $S/f.bak
sed -i 's|"dark_section": "dark_section_audit",|"dark_section": "hardcoded_section_colors",|' <file>
go test ./platform/orchestration/actions/ -run TestDesignAuditRoutes -count=1   # expect FAIL
cp $S/f.bak <file>                                                              # ALWAYS restore
```

**Gotcha that changed what I claimed:** mutate **one half at a time and record the
matrix**. Reverting Half A alone stayed GREEN, because Half B independently covers
the route. Had I mutated both at once and stopped there, I would have reported "the
test guards the fix" when what it actually guards is "at least one half of the fix".

## Council submission schema (two rejections, both cheap to avoid)

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

- **`.plan.summary` is required** and is a *different field* from `rationale`. The
  script refuses an empty one — `ERROR: .plan.summary is empty`.
- **`operation` enum is `modify|add|remove|config_change`.** `create` is rejected; a
  new file is `add`.
- Budget ~30 minutes, not ~2 — the council runs in 2–5 min but dispatch queues behind
  the fleet. A missing orchestration row is latency, not a dropped dispatch; do not
  retry on that evidence. Find the run by payload:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## Reading YOUR verdict — the documented query is correlation-blind

CLAUDE.md gives `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1`. **On a tree this busy that is a coin toss.** It returned
another lane's REVISE (`cb547e0a`, the `save_page_sections` decision-gate round) which
landed between my submit and my read — a verdict I nearly recorded as mine. Key every
read on your own correlation:

```sql
-- the decision, keyed on YOU
SELECT metadata->>'decision', metadata->>'abstained', metadata->>'gated_by_truncation'
FROM diagnosis_artifacts
WHERE correlation_id='<YOUR CORR>' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;

-- the human-readable note, keyed on YOU (the corr is printed inside the body)
SELECT body FROM doc_notes WHERE body LIKE '%<YOUR CORR PREFIX>%' ORDER BY created_at DESC LIMIT 1;

-- the OBJECTIONS: in diagnosis_artifacts.body, NOT in metadata and NOT in the note
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='<YOUR CORR>' AND kind='council_report';
```

**Three more gotchas.** (1) The `metadata` shape is not stable between rounds — mine
carried `reviewers` (an integer) and `unreadable`; another lane's carried `reviews` (the
array). Do not build a query on one shape. (2) The `doc_notes` note **stops after the
plan summary** — it never contains the objections, so "APPROVED with 4 advisory
objections" is all you learn there; read the artifact body for what they were. (3) A
long `psql -tAc` with `jsonb_array_elements` over that body can exceed a `kubectl exec`
timeout — select the raw `body` and grep it locally instead.

## Committing before a verdict

The `commit-msg` trailer gate **blocks** `Council-Submitted: pending`, and is right
to: a non-UUID resolves to nothing in the 098 join, and forward-only forbids fixing
it with an amend. Submit first, then commit with the printed correlation. Never write
`Council-Reviewed:` on a verdict you have not read — 098 buckets that as MISMATCH.

## Verifying this fix after the roll

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✓ NEW — 0 before the roll, 1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verifier_scope_mismatch'"
# ✓ POSITIVE CONTROL — live since RFC_017, must stay 1 in the SAME exec
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'verification_unavailable'"
```

The control is what proves the grep and the binary, not just your spelling. Repeat on
**every replica** — a label greps a subset.

**Do not grade this fix by re-running the verifier.** It will pass again, for the
same correct reason (213 §6). Grade a producer-B item against its own
`spec.acceptance_test`, or at the served artefact.

---

## D3 — running the class detector (`verifier-remit-check`)

```bash
go run ./cmd/verifier-remit-check --dry-run      # census only, writes nothing
go run ./cmd/verifier-remit-check --ignore-remit # THE DISCONFIRMABILITY CHECK (writes refused)
```

**`--ignore-remit` is the one to run when you doubt the zero.** It re-runs the same
census with the `Grades` suppression turned off; today that produces a real finding
on `hardcoded_section_colors` and exit 1. A `--dry-run` that says "0 findings" and
an `--ignore-remit` that also says 0 means the census is blind, not that the fleet
is clean.

**Gotcha, and it is silent:** from a terminal there is no `PG_CLIENTS_HOST`, so the
binary takes the `kubectl exec` route and **writes nothing even without
`--dry-run`** — no work item, no doc_note. It says so in its own report. Only the
CronJob (which sets `PG_CLIENTS_HOST`) writes.

Deploy and prove it at the artefact, not the tag:

```bash
make build-verifier-remit-check IMAGE_TAG=<tag>   # committed HEAD only
make push-verifier-remit-check IMAGE_TAG=<tag>
# the overlay's newTag must ALREADY name <tag> — deploy ships nothing on its own
make deploy-verifier-remit-check
make verifier-remit-check-now
kubectl -n ai-persona-system logs -l app=verifier-remit-check --tail=60
```
```sql
-- the artefact: one row per run, clean or not. MISSING = the job did not run.
SELECT created_at, left(body,300) FROM doc_notes WHERE source='verifier-remit-check'
ORDER BY created_at DESC LIMIT 1;
-- what it filed, if anything
SELECT summary, status, handler_agent, spec->>'subject_type'
FROM site_work_items WHERE item_type='verifier_remit_gap' ORDER BY created_at DESC;
```

## Measuring producer shapes by hand (what the detector automates)

```sql
SELECT item_type, COALESCE(spec->>'audit_source','<none>') AS label,
       (SELECT string_agg(k,',' ORDER BY k) FROM jsonb_object_keys(spec) k) AS keyset,
       count(*) FROM site_work_items WHERE item_type IN (<the RegisterVerifier list>)
GROUP BY 1,2,3 ORDER BY 1;
```

**Three axes that look like producer identity and are not** — each measured
2026-08-11 and each fires on a type with exactly ONE producer:
`count(DISTINCT created_by)` (2–3 on `empty_section`, `literal_markdown`),
`count(DISTINCT source)` (2 on `page_canonical_collision`), and
`count(DISTINCT spec->>'check')` (2 on the same). A raw distinct-key-set count is
just as bad: 2 on four single-producer types. **Cluster the key-sets** (overlap
coefficient ≥0.5, never Jaccard — J reads 0.167 on a real same-producer pair).

## Is a work-item status actually undispatchable?

```sql
SELECT status, count(*), count(*) FILTER (WHERE handler_agent='') AS empty_handler,
       count(DISTINCT item_type) AS types
FROM site_work_items GROUP BY 1 ORDER BY 2 DESC;
```
**`deferred` alone is NOT undispatchable** — 316 rows across 14 item_types carry it
and only 15 also carry the empty `handler_agent`. It is the PAIR that is the lock
(remit.go's double lock). Check both columns before believing a row is inert.

## grep flags that answer a different question, silently

`grep -Lq PATTERN file && echo file` does **not** list files without the pattern:
`-q` wins, so it prints the files WITH it — the exact opposite set, with no error.
Use `grep -rLE PATTERN <files>` and read the list. This produced a confident,
inverted answer to "is any discovery check fleet-scoped?" (20 files, all wrong;
the true answer is one, and it is a helpers file).

## Proving a CronJob IMAGE is the code you committed (no exec, no `strings`)

A CronJob pod is `Completed`, so you cannot exec into it, and the fleet's
`build provenance` log line does not apply — these images are on their own tag
sequence. The chain that needs no inference:

```bash
docker inspect $REG/<job>:$TAG --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
#   → must equal the commit you built from (ref_build stamps it)
kubectl -n ai-persona-system get pods -l job-name=<job-run> \
  -o jsonpath='{.items[0].status.containerStatuses[0].imageID}'
#   → must equal the sha256 digest `docker push` printed
```

Commit → image label → registry digest → running pod, each link checked. **A new tag
also removes the stale-cache hazard by construction** (the node cannot have cached a
tag that has never existed), which is why the same-tag rebuild rule bites re-rolls and
not first deploys — but say which one you are relying on.

## Gate 1b — the behavioural check owed AFTER the next chassis roll

No unit test proves the wiring (it needs a `*sql.DB`), so this is the only thing that
establishes gate 1b actually fires. Do it on the first roll carrying `96c53bc18`.

```bash
# 1. the gate is in the binary you are running (per SERVICE, not per fleet)
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 96c53bc18 <the stamp>   # exit 0 = your commit is in it
```

```sql
-- 2. the behavioural half. A dark_section_audit completion whose handler reported
--    total_fixed 0 must now land triaged/failed, NOT complete.
SELECT id, status, attempt_count, left(error, 120) AS error,
       result->'_verification'->>'status' AS gate_status
FROM site_work_items
WHERE item_type='dark_section_audit' AND created_at > '2026-08-13'
ORDER BY updated_at DESC;
-- expect error LIKE 'completion blocked: the handler reported it changed nothing%'
--        gate_status = 'handler_reported_no_change'
```

**⚠ The control that makes that meaningful, and it is the load-bearing half.** A gate
that never fires and a gate that is not wired look identical in that query. So assert
BOTH directions in the same window:

```sql
-- must be NON-ZERO once any dark_section_audit item completes at all:
SELECT count(*) FILTER (WHERE result->'_verification'->>'status'='handler_reported_no_change') AS blocked,
       count(*) FILTER (WHERE status='complete')                                               AS still_completing,
       count(*)                                                                                AS total
FROM site_work_items WHERE item_type='dark_section_audit' AND updated_at > '2026-08-13';
```

```sql
-- 3. the ABSTAIN arm, which is expected to fire for ~10 of every 14 on current data
SELECT created_at, error_message, context->>'item_type', context->>'declared_counters'
FROM agent_error_log WHERE error_code='NO_CHANGE_GATE_UNREADABLE_RESULT'
ORDER BY created_at DESC LIMIT 20;
```

If (3) is busy and (2) is empty, the gate is working and the 10-of-14 payload split
(`bugs_open/213` §D) is the norm rather than an anomaly — which is the answer to a
question this lane deliberately did not guess at.

## Re-firing the council submission (it did NOT dispatch on 2026-08-13)

Payload is complete and validated at `scratchpad/213_d1_gate1b_submission.json`.

```bash
kubectl -n ai-persona-system get pods >/dev/null || echo "TOKEN EXPIRED — owner must refresh; do not re-fire yet"
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <scratchpad>/213_d1_gate1b_submission.json
```

Two schema gotchas beyond the ones listed above, both paid for on 2026-08-13:

- **`.plan.risks` must be a STRING, not an array.** The script refuses an array outright:
  `ERROR: .plan.risks must be a STRING (Go: string) — join the risks into one prose block.`
- **⚠ A printed `SUBMISSION_CORR` is NOT evidence of a dispatch.** The script prints the
  correlation block *before* it publishes, so an expired kubeconfig produces a complete,
  convincing correlation printout followed by `Unauthorized` — and nothing is queued. Check
  for the run by payload before believing you have submitted anything, and **never write a
  `Council-Submitted:` trailer for a correlation you have not seen dispatch.**

## Proving gate 1b shipped — what actually worked on 2026-08-14 (v1.0.1298)

**The `build provenance` startup line was NOT reachable**, on either replica, at
`--tail=6000`. The pods were **under four hours old** and it had already scrolled — so treat
this service's provenance line as good for minutes, not hours, and do not read its absence as
"unstamped". `git merge-base --is-ancestor` needs a stamp you can actually read; with no
stamp, that route is closed.

**What worked: probe the binary for a KNOWN literal of your own change, with BOTH controls in
one pass.**

```bash
for P in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  echo "--- $P"
  kubectl -n ai-persona-system exec "$P" -- sh -c \
    "grep -aoE 'NO_CHANGE_GATE_UNREADABLE_RESULT|verification_unavailable|ZZZ_NOT_A_REAL_LITERAL_9f3a' /proc/1/exe | sort -u"
done
```

Read it as three answers, not one: the **first** must be PRESENT (gate 1b's own error code);
`verification_unavailable` must be PRESENT (a long-live RFC_017 literal — proves the probe and
the binary, not your spelling); the nonsense needle must be ABSENT (proves the grep is not
matching everything). Result on v1.0.1298: needles 1 and 2 present, needle 3 absent, **on
both replicas**.

⚠ **Two gotchas that cost a run each.** A loop doing one `grep -aq` per needle **times out at
2 minutes** — each invocation rescans a ~100MB binary, so a three-needle loop over two pods
cannot finish. One `grep -aoE 'a|b|c' … | sort -u` scans once and answers all three. And
**never grep for your COMMIT SHA to prove your code shipped**: the provenance stamp is a
single sha — the build's HEAD — so unless the build was cut at exactly your commit, your sha
is absent while your code is present. That is the wrong question, and it fails in the
direction that looks like bad news.

## ⚠ The behavioural check CANNOT be run today, and this is why

Gate 1b is in the binary and **has never executed**. [MEASURED 2026-08-14] zero
`dark_section_audit` rows touched since the 08:58Z roll; zero
`NO_CHANGE_GATE_UNREADABLE_RESULT` records; zero `_verification` keys.

The cause is not the gate: **`improvement-sweep` is `enabled = false`** (last triggered
2026-08-12 16:16Z, switched off by the `bugfix_122` lane after a 3.2x cost surprise), and it
is the only triage carrier that dispatches these items. `site-discovery-rotation-design` is
off too.

```sql
SELECT name, target_agent_type, enabled, last_triggered_at FROM scheduled_tasks
WHERE name IN ('improvement-sweep','site-discovery-rotation-design');
-- improvement-sweep | improvement-loop | f | 2026-08-12 16:16:22+00
```

So one `detected` row sits waiting (`mortgagecalculator.co.uk`, filed 2026-08-13 22:03,
`attempt_count` 0) and nothing will pick it up. **Waiting does not obtain this proof.** The
three ways to get it are in the handoff's decisions section; two of them need the owner.

**The reframing this forces, and it should be said plainly rather than buried:** the
false-green bleed has been **paused since 2026-08-12 by the sweep being off for unrelated
cost reasons**, not by gate 1b. The gate's value is that it makes re-enabling that sweep safe.
There is no urgency, and any claim that the gate "stopped the bleed" would be false.

---

## D1 HALF TWO — the commands that were hard to get right (2026-08-15)

### 1. Before writing ANY periodic update to a work-item row: enumerate the readers

The one that nearly cost a fix. **Do not assume there is no reader** — and read the `pre_query`
rather than trusting the `LIKE`, because three tasks match and only one genuinely reads this
column.

```sql
SELECT name, enabled, (pre_query ILIKE '%updated_at%') AS reads_updated_at
FROM scheduled_tasks WHERE pre_query ILIKE '%site_work_items%'
ORDER BY enabled DESC, name;
-- 2026-08-15: stale-work-item-reaper is the ONLY real reader.
-- claimed-item-timeout keys on claimed_at + orchestration_states.updated_at;
-- improvement-sweep reads sites.updated_at. Both are false positives of the LIKE.

SELECT name, enabled, interval_seconds, pre_query FROM scheduled_tasks
WHERE name='stale-work-item-reaper';   -- read the predicate, do not summarise it

SELECT tgname FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid
WHERE c.relname='site_work_items' AND NOT tgisinternal;
-- trg_site_work_items_updated_at (BEFORE UPDATE) — this is WHY no column list saves you
```

### 2. The mutation matrix harness, with the three things that make it worth running

Gate on **exit status** — never `grep '^\s+--- FAIL'`, which matches SUBtests only, so a
top-level failure reads as "mutation not caught". Always include an **unmutated control**, and
always **verify byte-identity** of every mutated file afterwards.

```python
# one mutation at a time; restore from a pristine copy between each
p = subprocess.run(["go","test","./platform/orchestration/actions/","-run",TEST,"-count=1"],
                   cwd=ROOT, capture_output=True, text=True)
verdict = "RED (caught)" if p.returncode != 0 else "GREEN (NOT CAUGHT)"
if "build failed" in out or "declared and not used" in out:
    verdict = "BUILD ERROR — worthless, fix the mutation"   # a compile error prints FAIL
...
print("restored byte-identical:", open(f,'rb').read() == open(BAK,'rb').read())
```

**Reading a GREEN result — it means one of three things, and they need opposite responses:**
1. **The test is vacuous.** Rewrite the test. (Our `UnrecognisedReplyIsNotSilence` — it was
   asserting a negative through sqlmock's refusal, not through the guard.)
2. **You hit a guard in SERIES.** Something behind it caught the mutation. Pass the real
   condition down instead of short-circuiting above it, so each guard can be mutated singly.
3. **The code is unreachable by construction.** **Delete it.** Do not write a test to justify it.

### 3. Testing a guard that PREVENTS work — do not do it through the action

sqlmock cannot express "this call must not happen": an expectation you intend not to be met
fails `ExpectationsWereMet`. And driving the action with no `ExpectBegin` passes either way,
because the mock's refusal becomes an error the action swallows into a log line.

```go
// Call the FUNCTION directly, with a mock carrying NO expectations at all,
// and assert on the ERROR RETURN — now a stray DB call is visible.
db, _, _ := sqlmock.New()
out, err := retractSilentAuditFindings(ctx, db, siteID, src, batch, map[string]bool{}, false, log)
if err != nil { t.Fatalf("must touch no database at all, got: %v", err) }
if out != nil { t.Fatalf("must report no retraction, got %#v", out) }
```
Always pair it with the opposite-direction test, or "guard always on" passes too.

### 4. Building the test's ActionParams for write_audit_findings

Config values are **dot-paths into `collected_data`, never literals** (bugs_open/264 made
`audit_source` Required with no default). Getting this wrong fails with
`missing required fields: [audit_source]`:

```go
StepConfig: models.Step{Config: map[string]interface{}{
    "site_id":      "site_record.site_id",
    "audit_source": "audit_source_literal.audit_source",
}},
CollectedData: map[string]interface{}{
    "site_record":          map[string]interface{}{"site_id": siteID.String()},
    "audit_source_literal": map[string]interface{}{"audit_source": "visual-design-audit"},
    "audit_result":         map[string]interface{}{"result": findings},
},
```

### 5. Is the producer-scope guard releasing the rows, or sealing them shut?

The council's best objection. **Re-run this whenever the `audit_source` literal changes** — the
guard compares a row's STORED source against the run's CURRENT one, and a rename orphans rows.

```sql
-- (a) what the open rows carry
SELECT s.domain, w.status, w.spec->>'audit_source'
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='dark_section_audit'
  AND w.status NOT IN ('complete','verified','rejected','wont_fix','cancelled');

-- (b) what a run would produce — the literal is hardcoded in the auditor's own step
SELECT jsonb_pretty(jsonb_path_query_array(default_config->'workflow'->'steps',
       '$.* ? (@.output_field == "audit_source_literal")'))
FROM agent_definitions WHERE type='visual-design-auditor'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- 2026-08-15: both read 'visual-design-audit' — they MATCH, so the guard releases.
```

### 6. Has the auditor ever been truncated? (silence must not be a cut-off reply)

```sql
SELECT agent_type, count(*) calls, count(*) FILTER (WHERE output_tokens >= max_tokens) at_cap,
       max(output_tokens) max_out, min(max_tokens) min_cap
FROM llm_call_log WHERE agent_type IN ('visual-design-auditor','design-audit-agent') GROUP BY 1;
-- 2026-08-15: 4,088 calls, 0 at_cap, max_out 1,869 vs a 4,000 cap. Re-run if either moves.
```

### 7. Is the account's LLM cap actually still biting? — check the SUCCESS side

A stated reset date in the 400 body is the vendor's worst case, **not a forecast**; on this
estate it has been raised within hours all three times. Failures simply stop appearing whether
or not capability returned, so a "no failures" query proves nothing.

```sql
SELECT date_trunc('hour',created_at) hr,
       count(*) FILTER (WHERE success) ok,
       count(*) FILTER (WHERE error_message LIKE '%usage limits%') capped
FROM llm_call_log WHERE created_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
-- A resumed `ok` column is the only proof. 2026-08-14's "outage until 2026-09-01" was ~90 min.
```

### 8. Verifying half two once a carrier is re-enabled

Neither carrier is on: `improvement-sweep` (`enabled=false`) and
`site-discovery-rotation-design` (`enabled=false`). Until one flips, this is unexercised.

```sql
-- the streak accumulating across runs IS the mechanism working
SELECT s.domain, w.status, w.result->'retraction'
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='dark_section_audit'
  AND w.status NOT IN ('complete','verified','rejected','wont_fix','cancelled');

-- a retraction, when one happens
SELECT item_key, result->>'resolved_by', result->>'reason', result->>'resolved_at'
FROM site_work_items WHERE item_type='dark_section_audit' AND result ? 'resolved_at';
```
⚠ **The correct expected outcome on the 4 live `failed` rows is that they do NOT retract.**
They are genuinely unrepaired, so a run that closes them is the bug, not the proof. And check
the pod carries the code before trusting any of it — the binary states its own commit.
