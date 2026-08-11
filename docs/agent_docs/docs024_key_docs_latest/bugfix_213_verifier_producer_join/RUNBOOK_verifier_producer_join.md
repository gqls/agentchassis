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
