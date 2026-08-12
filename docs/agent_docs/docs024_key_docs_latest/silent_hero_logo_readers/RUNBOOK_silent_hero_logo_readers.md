# RUNBOOK — silent hero/logo readers (commission item 2)

Every command that was hard to get right, with its gotcha attached. Change it HERE, not in
scrollback.

---

## Locate the readers — never trust a doc's file:line

The commission's line numbers were already stale when this lane opened (see NOTES). Ask the
compiler's view of the world, not a document's:

```bash
grep -rn "hero_deployed\|logo_deployed" --include=*.go platform/ internal/ | grep -v _test.go
```

**Gotcha:** `grep -v _test.go` matters — the test files name these keys too, and a raw grep makes
the site count look larger than it is.

## Check nobody else owns the bug

```bash
python3 scripts/who-owns.py 236
```

**Gotcha:** `236` is one of the documented **ambiguous numbers** — two unrelated cases share it
(site-availability/522 and this hero/logo one). The script says so. Always resolve by slug:

```bash
ls bugs_open/ | grep 236
```

## Tests

```bash
go test ./platform/orchestration/actions/ -run 'TestHeroLogo|TestCollectedDataKeys' -count=1
```

**Gotcha:** `-count=1` defeats the test cache. Without it an edit to a *non-test* file can
still serve a cached PASS in some layouts, and a cached pass is indistinguishable from a real one.

**Prove the guard by mutation, not by the mock's bookkeeping.** A test that asserts "the recorder
was called" passes if the recorder is called for the wrong reason. Invert the demand condition,
re-run, and confirm the no-op test *fails*:

```bash
# after temporarily flipping the `present` check in the reader
go test ./platform/orchestration/actions/ -run TestHeroLogo -count=1   # MUST fail
```

If it still passes, the guard you think you tested is in series with another one.

**⚠ BACK UP FIRST, RESTORE IMMEDIATELY, AND `diff` — the tree is shared.** A mutation left in the
working tree can be committed by another session before you revert it; commit `038211dd8` swept
this lane's four files into HEAD minutes after a mutation window closed. A disabled guard reaching
HEAD inside someone else's commit is close to undetectable. The whole recipe, not just the middle:

```bash
SP=<scratchpad>
cp platform/orchestration/actions/deployed_image_read_audit.go $SP/audit.go.orig
# ... mutate, run the test that MUST fail, restore ...
diff $SP/audit.go.orig platform/orchestration/actions/deployed_image_read_audit.go \
  && echo "IDENTICAL to pre-mutation backup"
```

Do the `diff` in the same breath as the restore, not at the end of the session.

## Verify HEAD, not your working tree

`make build-*` builds from committed HEAD, and on this tree HEAD may contain your work under
someone else's commit — plus their work under yours. A green working tree says nothing about it:

```bash
SP=<scratchpad>; rm -rf $SP/headtree && mkdir -p $SP/headtree
git archive HEAD | tar -x -C $SP/headtree
cd $SP/headtree && go build ./... && go test ./platform/orchestration/actions/ -count=1
```

**Gotcha:** `git archive HEAD` is the only honest reproduction of what the build will contain.
Also `diff` each file you care about between `$SP/headtree` and the working tree — that is how
this lane established HEAD carried the restored helper and not the mutated one.

## Build and roll

```bash
make build-agent-chassis        # builds from committed HEAD — commit FIRST
```

**Gotcha:** releases are **whole-fleet** and the owner runs `make release`. Never
`kubectl apply -k` one service at its own tag. Bump `IMAGE_TAG` (makefile ~line 16) or a same-tag
rebuild ships the node's stale cached binary.

## Prove it shipped

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this-commit> <the-stamped-sha> && echo SHIPPED
```

**Gotcha:** the provenance line is a **startup** line and scrolls — on `agent-chassis` it was
measured absent from `--tail=3000` hours after a roll. **Empty means "not in range", not
"unstamped".** Fall back to the binary probe, and always run a control that must be ABSENT:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "<expected-sha>" /proc/1/exe && echo PRESENT
kubectl -n ai-persona-system exec $POD -- grep -aq "0000000000000000000000000000000000000000" /proc/1/exe || echo "control absent (good)"
```

Never `strings` — it is absent from the debian-slim images, and behind `2>/dev/null` its failure
is indistinguishable from "not stamped".

**This change DOES have a greppable literal, which not every change does** (the council's
`debug_historian` seat asked for exactly this). The new `error_code` is compiled in, so it dates
the binary without planting a marker — run it with a control that must be absent:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "DEPLOYED_IMAGE_RESULT_MISSING_URL" /proc/1/exe \
  && echo "PRESENT — this binary carries item 2"
kubectl -n ai-persona-system exec $POD -- grep -aq "DEPLOYED_IMAGE_RESULT_MISSING_URL_NOT" /proc/1/exe \
  || echo "control absent (good — the grep is not saying yes to everything)"
```

## The check that actually matters — a durable row at the moment of failure

```sql
SELECT created_at, agent_type, step_name, action, error_code, severity,
       context->>'wanted_key'      AS wanted,
       context->>'container_key'   AS container,
       context->'keys_present'     AS keys_the_map_held
FROM agent_error_log
WHERE error_code = 'DEPLOYED_IMAGE_RESULT_MISSING_URL'
ORDER BY created_at DESC
LIMIT 20;
```

**Gotcha — this needs DEMAND, not just time.** A zero here is meaningless unless a site build
that deploys a hero or logo actually ran in the window. Establish the demand in the same breath,
or the zero has two readings (nothing broke / nothing was tried):

```sql
-- did anything even try? (the 4-hour window means run this SOON after a build)
SELECT count(*) FILTER (WHERE collected_data ? 'hero_deployed') AS hero,
       count(*) FILTER (WHERE collected_data ? 'logo_deployed') AS logo,
       count(*) AS retained
FROM orchestration_states;
```

**Gotcha:** `hero_deployed`/`logo_deployed` live on rows in `AWAITING_RESPONSES`, which
`database-cleanup` prunes after **4 hours** (hourly job, live row — the repo seed says 24h and
disagrees). Read `pre_query FROM scheduled_tasks WHERE name='database-cleanup'` for the live
truth; a query run the next morning is guaranteed to return 0 and that is not evidence of absence.

## Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/silent_hero_logo_readers/SUBMISSION_2026-08-11_silent_readers.json
```

**Gotcha:** budget ~30 minutes, not ~2 — the council takes 2–5 min but the dispatch queues behind
the fleet (measured 29 min). A missing orchestration row is latency, not a dropped dispatch; do
not retry. Find the run by payload, not by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## Read a 090 verdict — it is NOT in `diagnosis_artifacts`

```sql
-- the verdict lives on the RUN's own orchestration row
SELECT jsonb_pretty(collected_data->'verdict') FROM orchestration_states
WHERE correlation_id::text='<RUN_CORRELATION_ID>' AND collected_data ? 'verdict' LIMIT 1;

-- status, and whether it really finished
SELECT status, current_step, updated_at FROM orchestration_states
WHERE correlation_id::text='<RUN_CORRELATION_ID>' ORDER BY updated_at DESC;
```

**Gotcha, and it cost two polls here:** `diagnosis_artifacts` for a 090 correlation holds only
`kind='bundle'` rows — one per iteration. There is **no `diagnosis_report` row and no
`metadata->>'verdict'`**, so a poll shaped like the council's reports `NOT YET` forever on a run
that finished hours ago. A poll that looks in the wrong place is indistinguishable from a run still
in flight; always print a **connectivity control** (`SELECT count(*) FROM orchestration_states`)
beside it so "cannot see" cannot masquerade as "not there".

**Gotcha:** the 090 row is found by `correlation_id`, **not** by
`collected_data->'input_data'->>'fix_correlation_id'` — that path is the *council's* key shape.
Using the council's predicate on a 090 returns zero rows and looks like a dropped dispatch.

**Gotcha:** `status='COMPLETED'` is not success — per `bugs_open/099`, a failed step can show
COMPLETED with `error` NULL. Read the verdict payload itself.

## Is the code tier actually giving the loop what it needs?

Before believing "the loop could not read function X", check which half is broken — the index or
the bundle. They fail differently and only one of them is fixed by reindexing:

```sql
SELECT symbol, kind, length(COALESCE(body,'')) AS body_len, line_start, line_end
FROM code_symbols WHERE symbol IN ('<yours>', '(*Receiver).<yours>');
```

**Gotcha:** a healthy row here proves **nothing** about what the bundle rendered — that was this
lane's wrong call (`WRONG_CALLS.md`, 2026-08-12). The index held four bodies and the bundle rendered
one. To see what the verdicter actually received, read the artefact:

```sql
SELECT left(body, 4000) FROM diagnosis_artifacts
WHERE correlation_id='<RUN_CORRELATION_ID>' AND kind='bundle' ORDER BY created_at DESC LIMIT 1;
```

**Gotcha:** methods are stored **receiver-qualified** (`(*SagaCoordinator).applyResponseToState`),
so a bare-name lookup returns 0 rows and no error — an existing LANDMINE, and the loop's
cite-or-abstain rule acts on that absence.
