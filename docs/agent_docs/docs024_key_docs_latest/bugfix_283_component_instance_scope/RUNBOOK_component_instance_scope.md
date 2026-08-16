# RUNBOOK — component instance scope (`bugs_open/283`)

Every command here was hard to get right at least once. The gotcha is attached to the command,
not kept in a separate section, because the gotcha is the reason the command is written down.

---

## 1. Prove the code you committed is actually running

**Use this, not the two recipes in CLAUDE.md, when the `build provenance` line has scrolled.**

```bash
# 1. what the pods are running, by DIGEST (not by tag — tags are reused)
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}'
# -> docker.io/aqls/agent-chassis@sha256:75ae5902…

# 2. does the LOCAL image of that tag have the same bytes?
docker image inspect aqls/agent-chassis:v1.0.1304 --format '{{json .RepoDigests}}'
# -> ["aqls/agent-chassis@sha256:75ae5902…"]      MUST MATCH step 1

# 3. only now is the local label evidence about the running pod
docker image inspect aqls/agent-chassis:v1.0.1304 \
  --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
# -> 5de6cddbe6b281da97dc933d823ebe84da2bbf8a

# 4. ancestry, not equality
git merge-base --is-ancestor <your-commit> 5de6cddbe && echo LIVE || echo NOT-IN-THIS-BUILD
```

⚠ **Step 2 is the load-bearing one and it is the one you will skip.** A local tag can be rebuilt
by any session at any time; the label only describes the running pod once the digests match.

⚠ **Why not the documented alternatives.** `kubectl logs … | grep 'build provenance'` is a
*startup* line — it was already out of reach at `--tail=20000` on both chassis pods. Grepping the
binary (`grep -aq "<sha>" /proc/1/exe`) can only ever confirm a **guess**: the binary carries its
own build stamp, not its ancestors, so an older *real* commit is equally absent. It is also slow
enough to blow a 2-minute timeout.

## 2. The one query that governs this whole seam

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
 "SELECT count(*) FROM content_components WHERE is_active AND html_template LIKE '%InstanceID%';"
```

**0 means the seam is inert and its shape is still free to change.** Non-zero means: the token is
load-bearing, the RFC_022 exception it was approved under has expired (see `RFC_032` §5), and
changing the token shape now moves live element ids. **Re-run it before any change to
`InstanceToken`, and at merge time for anything in this lane** — the council's guardian seat asked
for exactly that, because a template conversion landing in the same window changes the answer.

## 3. Test when the working tree will not build

The shared tree frequently carries another session's staged-but-broken test file (2026-08-16:
`agent_definition_nullable_columns_test.go` redeclaring `stripLineComments`). Do not touch it.

```bash
SC=<scratchpad>/tree
rm -rf $SC && mkdir -p $SC && git archive HEAD | tar -x -C $SC
for f in component_instance_scope.go component_instance_scope_test.go assemble_from_library.go \
         rerender_page_sections_action.go v3_site_actions.go section_editor_actions.go component_library.go; do
  cp platform/orchestration/actions/$f $SC/platform/orchestration/actions/$f
done
cd $SC && go build ./... && go test ./platform/orchestration/actions/
```

This also proves your change does not depend on anyone's WIP, which a working-tree run cannot.

## 4. Mutation-test a guard before trusting it

A detector that reports nothing is indistinguishable from one that is inert, so every clean
assertion needs a mutation of the same input that must come out dirty.

```bash
F=$SC/platform/orchestration/actions/component_instance_scope.go; cp $F $F.bak
python3 - "$F" <<'EOF'
import sys; p=sys.argv[1]; s=open(p).read()
s=s.replace("\ttok := InstanceToken(function, c.seen[key])\n\tc.seen[key]++\n",
            "\ttok := InstanceToken(function, 0)\n\t_ = key\n")   # counter never advances
open(p,'w').write(s)
EOF
cd $SC && go test ./platform/orchestration/actions/ -run 'Instance|RenderLayer'   # MUST fail
cp $F.bak $F
```

⚠ **Restore the file in the same command block.** A mutation left in the scratch tree makes the
next run's result meaningless, and it looks exactly like a real failure.

## 5. Run the new pattern-check standalone, over the WHOLE tree

The pre-commit hook only sees changed files. To prove the check is not inert, run it over
everything, at HEAD and at the working tree:

```bash
python3 - <<'EOF'
import importlib.util, subprocess
spec = importlib.util.spec_from_file_location("pc", "scripts/pattern-check.py")
pc = importlib.util.module_from_spec(spec); spec.loader.exec_module(pc)
allgo = [p for p in subprocess.run(["git","ls-files","*.go"],capture_output=True,text=True).stdout.split()
         if not p.endswith("_test.go")]
for label, ref in (("HEAD", ("HEAD~1","HEAD")), ("working tree", None)):
    f=[]; pc.check_unscoped_component_render(allgo, ref, f)
    print(label, len(f), [x[1] for x in f])
EOF
```

⚠ **`git ls-files '*.go'`, never a directory list.** Scoping the sweep to `platform/ internal/` is
how `cmd/component-render-check` was missed — a census whose scope is a list of directories you
chose is a census of your own expectations (`WRONG_CALLS.md`, 2026-08-16).

## 6. Which live workflows execute an action — BOTH queries, always

```sql
-- structured: executes it. MUST include the sub_workflow arm or you will under-report.
WITH s AS (SELECT ad.type, st.value AS step
           FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') st
           WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT DISTINCT type FROM s WHERE step->>'action' = '<action>'
UNION
SELECT DISTINCT s.type FROM s, LATERAL jsonb_each(s.step->'config'->'sub_workflow'->'steps') sw
 WHERE sw.value->>'action' = '<action>';

-- text control: mentions it. Over-reports (footprint maps, next_step labels) and that is the point.
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%<action>%';
```

⚠ **A top-level-only census told me `render_component` was executed by no live workflow.** It runs
inside a loop step — specifically `page-content-writer`'s `process_sections_loop`, whose
`config.sub_workflow.steps.render_section.action` the shallow query cannot see. Fleet-wide that blind spot hides **80 invocations across 19 action names**. Reconcile
the two counts before publishing either; a difference is a question, not noise. Full trap in
`LANDMINES.md`.

## 7. Council: submit, resubmit, read the verdict

```bash
RESUBMIT_CORR=<previous corr> \
 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

```sql
SELECT created_at, metadata->>'decision', metadata->>'decided_by'
FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;
-- the report body is large; read it in slices
SELECT left(body, 5000)                FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;
SELECT substring(body from 5000 for 5500) FROM diagnosis_artifacts WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;
```

⚠ **Round 2 came back in ~20 minutes**, not the ~30 the runbook budgets — but do not plan on that.
⚠ **`RESUBMIT_CORR` keeps one trail**: both rounds' reports sit under the same correlation, so
`ORDER BY created_at` and take the newest.

## 8. Landmines: append, then ARM — in that order

```bash
./scripts/landmines-verify-dispatch.sh                       # syncs AND arms
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>' # if the status was already consumed
```

⚠ **If another session sweeps your LANDMINES edit into their commit and runs the sync, the
verifier will never check your entry** — the dispatch will report a different entry as the one
needing verification. That happened on 2026-08-16. Recover the slug from the DB:

```sql
SELECT subject_key FROM doc_notes WHERE subject_key LIKE 'LANDMINES.md#%' ORDER BY created_at DESC;
```

Slugs are a truncated kebab-case of the heading, with punctuation dropped —
`` `{{.ComponentID}}` is the estate's… `` becomes
`componentid-is-the-estate-s-per-instance-id-convention-on-one-render-path-and-th`.

## 9. The RFC_022 expiry tripwire

```bash
# deploy / redeploy (no makefile target yet — see NOTES for why)
kubectl apply -k deployments/kustomize/services/instance-token-adoption-check/overlays/production/uk_001

# run it now rather than waiting for 07:40 UTC
kubectl -n ai-persona-system create job ita-now --from=cronjob/instance-token-adoption-check
kubectl -n ai-persona-system logs job/ita-now
kubectl -n ai-persona-system delete job ita-now      # one-off jobs are not garbage-collected

# what it has been saying
```
```sql
SELECT created_at, left(body, 200) FROM doc_notes
WHERE subject_key='instance-token-adoption' ORDER BY created_at DESC LIMIT 5;
```

⚠ **A MISSING row is the alarming case, not a quiet one.** The job writes a row on every run,
including when it finds nothing, precisely so that "the job did not run" cannot look like "the
exception still holds".

⚠ **Exercise both branches before trusting a change to it** — the tripped branch is the one that
never runs in practice, so it is the one that rots:

```bash
cd deployments/kustomize/services/instance-token-adoption-check/base
echo '{"adopters":0,"control":5,"active_total":243,"adopter_names":[]}' | python3 check.py --stdin; echo "exit=$?"   # 0
echo '{"adopters":2,"control":5,"active_total":243,"adopter_names":["x","y"]}' | python3 check.py --stdin; echo "exit=$?"  # 1
```

⚠ **If it ever reports `REFUSING TO RUN: the {{.ComponentID}} control matched 0 templates`, do not
"fix" it by removing the control.** That message means the pattern matching itself stopped working,
and every zero this job has reported since is worthless. Find out why the control stopped firing.
