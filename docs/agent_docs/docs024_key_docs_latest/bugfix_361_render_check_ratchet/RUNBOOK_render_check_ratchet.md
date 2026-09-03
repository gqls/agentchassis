# RUNBOOK — bugfix_361 render-check ratchet

## Is the job actually red, and for how long?

**`lastSuccessfulTime` is the instrument. NOT `get jobs`** — `failedJobsHistoryLimit: 3` means
the Job list shows at most three failures however long it has been broken, which renders as
"fine for a fortnight, then broke on Thursday". (LANDMINES has the full trap.)

```bash
kubectl -n ai-persona-system get cronjob component-render-check \
  -o jsonpath='{.status.lastSuccessfulTime}{"\n"}'
```

The cheaper, retained signal — **one `doc_notes` row/day is green, two is red** (backoffLimit 1
⇒ two pod attempts ⇒ two reports):

```sql
SELECT created_at::date AS day, count(*) AS rows_that_day,
       left(split_part(max(body), E'\n', 1), 160)
FROM doc_notes WHERE source='component_render_check'
GROUP BY 1 ORDER BY 1 DESC LIMIT 14;
```

## Dump the live library to a fixture (to run the tool offline)

⚠ **The plain stream TRUNCATES** — it is ~6.5 MB now, and it fails as a JSON parse error at a
byte offset, not as a non-zero exit, so it looks like bad data rather than a short read.
Compress inside the pod:

```bash
kubectl exec -n ai-persona-system postgres-clients-0 -- sh -c \
  "psql -U clients_user -d clients_db -tAc \"SELECT COALESCE(jsonb_agg(jsonb_build_object(
     'name', name, 'function', function, 'html_template', html_template,
     'created_at', created_at)), '[]'::jsonb) FROM content_components WHERE is_active;\" \
   | gzip -9 | base64 -w0" > live.b64
base64 -d < live.b64 | gunzip > live_components.json
python3 -c "import json;print(len(json.load(open('live_components.json'))),'components')"
```

That last line is the point: **parse it before you trust it**, or a truncated dump becomes a
measurement.

## Build and run the tool

The working tree frequently does not compile (other sessions' WIP in
`platform/orchestration/actions`, which this tool imports). Build from **committed HEAD**:

```bash
./scripts/verify-head-builds.sh --with cmd/component-render-check/rendercheck.go \
    ./cmd/component-render-check/          # does my change build against HEAD?

git worktree add --detach /path/to/wt HEAD    # to get a RUNNABLE binary
cp cmd/component-render-check/rendercheck.go /path/to/wt/cmd/component-render-check/
(cd /path/to/wt && go build -o /path/to/crc ./cmd/component-render-check/)

/path/to/crc --json live_components.json --compare ; echo "EXIT=$?"
```

⚠ **Read the exit code without a pipe.** `… | head` returns `head`'s status, so `EXIT=0` after
a pipe means nothing. Exit 1 = ran, found a regression. Exit 2 = could not run. 0 = clean.

## Prove BOTH ratchet arms by mutation (361 §6 requires it)

```bash
(cd /path/to/wt && go test ./cmd/component-render-check/ -v)
```

Then mutate the guard itself and watch its own test go red — **a test that passes under the
mutation it was written to catch is zero evidence, and it looks identical to a good one**:

| mutation in `rendercheck.go` | must fail |
|---|---|
| `if covered[canonicalName(f.Component)] {` → `if false {` | Arm 1, clean-component pin, clone pin |
| same → `if true {` | Arm 2 |
| delete `legacy = true` | the legacy-baseline pin |
| `cn := canonicalName(name)` → `cn := name` | the written-coverage pin |

⚠ That last one **passed** before the test was hardened: `loadBaseline` re-canonicalises, so a
round-trip assertion has two guards in series. Assert on the emitted JSON, not the round-trip.

## Regenerating the baseline — a DEBT decision, not a code one

`--write-baseline` banks every outstanding finding as "already known". After the fix it also
writes the `components` coverage list, which closes the legacy blind spot. It refuses if any
component failed to parse, and now also refuses `--component` (a covered set of one would make
every other component read as unbaselined and pass).

## Deploying it

**`make deploy-component-render-check` ships NOTHING on its own** — the overlay pins the tag
and both make and kubectl report success anyway. `component-render-check` IS in
`RELEASE_IMAGES` (makefile:95), so a fleet release rebuilds it from committed HEAD. Read the
artefact, never the make target:

```bash
kubectl -n ai-persona-system get cronjob component-render-check \
  -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}{"\n"}'
kubectl -n ai-persona-system create job --from=cronjob/component-render-check \
  crc-manual-$(date +%Y%m%d-%H%M%S)      # then read the POD's terminated exitCode
```
