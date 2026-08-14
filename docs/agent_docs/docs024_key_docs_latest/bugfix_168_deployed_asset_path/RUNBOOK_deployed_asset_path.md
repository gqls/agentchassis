# RUNBOOK — bugfix 168, deployed asset path

Commands that were hard to get right, with the gotcha attached. Change them HERE.

---

## Is this bug being worked by another session?

`scripts/who-owns.py` reads **commits**, so a session mid-fix is invisible to it. Do both:

```bash
scripts/who-owns.py 168

cd /home/ant/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -maxdepth 1 -name "*.jsonl" -mmin -240); do
  echo "$(basename $f .jsonl | cut -c1-8): $(tail -c 900000 "$f" \
    | grep -oE 'bugs_open/[0-9]{3}' | sort | uniq -c | sort -rn | head -5 | tr '\n' ' ')"
done
```

⚠ **`tail -c` on some of these files panics** (`Result::unwrap() on an Err ... InvalidInput`)
— a uutils `tail` bug on certain offsets. It prints the panic and yields nothing for that
file, exit code non-zero, **and the loop carries on silently**. A session whose file panicked
looks exactly like a session holding no bugs. Four did on 2026-08-02. Re-check any file that
panicked with `grep -c` over the whole file before concluding a bug is unowned.

## The census that decides whether this defect is latent or live

```sql
SELECT purpose, COALESCE(asset_key,'<null>') AS asset_key, count(*) AS rows,
       count(*) FILTER (WHERE url LIKE '/assets/%')        AS url_is_webpath,
       count(*) FILTER (WHERE url LIKE 's3://%' OR url LIKE 'http%') AS url_is_s3
  FROM assets
 WHERE status='active'
   AND (asset_key IS NULL OR asset_key='' OR asset_key=purpose)
 GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ **The `WHERE` clause is the whole measurement** — it is the helper's skip branch
transcribed. Drop it and you get 195 rows of noise that answer a different question. The
rows this returns are the *only* ones where a purpose-derived spelling is used verbatim.

⚠ `count(col)` counts empty strings as present. Use `count(*) FILTER (WHERE ...)` for
anything conditional, not `count(nullif(...))`.

## Which writer published a given asset?

There is no column for this — **`assets` records no served path for deployer-published
rows**, which is the landmine that makes the whole class hard. The discriminator that works:

```sql
SELECT purpose, asset_key, origin_model, url FROM assets
 WHERE site_id = '<uuid>' AND status='active' AND purpose IN ('favicon','og_card');
```

`origin_model = 'derived-from-logo'` ⇒ written by `derive_brand_head_assets_action`, and its
`url` holds the **site-relative web path** it committed. Every other generated row's `url` is
an **expiring presigned S3 URL** and is not a path. Do not write a check that treats `url`
as one kind of thing.

## Proving a guard actually guards (mutate — do not trust a green run)

```bash
cp platform/storage/url_helpers.go /tmp/.../url_helpers.go.bak
# delete the brand-head branch from DeployedAssetPath, then:
go test ./platform/storage/...          # MUST fail, naming og_card
cp /tmp/.../url_helpers.go.bak platform/storage/url_helpers.go
```

⚠ The harness will report the mutated file as "modified — intentional, don't revert". That
message is about *your own* probe; restore the backup regardless. Keep the mutation window
to seconds: `make build-*` builds from **committed HEAD**, so a working-tree mutation cannot
ship — but another session reading the tree will see it.

## Council submission — the two traps that cost a whole round

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  /tmp/.../submission_168.json
```

Validate **before** submitting; both of these fail *silently at exit 0* and only surface
later as `current_step = complete_invalid`:

```python
d = json.load(open(path))
assert isinstance(d['risks'], str)            # a LIST fails to unmarshal
assert len(d['plan']['edits']) <= 8
banned = ['no code change','no change required','no change is required','no change needed',
          'no change is needed','clarifying note','clarifying comment','add a comment','comment-only']
for e in d['plan']['edits']:                  # literal Contains over the LOWERCASED sketch
    assert not [b for b in banned if b in e['sketch'].lower()]
```

The banned-phrase scan covers your **prose inside the sketch**, not just the diff — a
sentence explaining that you folded documentation into an edit will reject the whole plan.
Put that explanation in `rationale`, which is not scanned.

Watch for the invalid ending, not just for a verdict row:

```sql
SELECT current_step, status, collected_data->'__step_error'->>'message'
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

## Reading a diagnosis verdict when no `doc_notes` row appears

The `090` loop wrote **no** terminal note for corr `ae9404bd` — `doc_notes` had nothing, and
`diagnosis_artifacts` held only `kind='bundle'` rows (the loop's *input*, not its output).
The verdict lives on the orchestration:

```sql
SELECT jsonb_pretty(collected_data->'verdict') FROM orchestration_states
 WHERE correlation_id::text='<RUN_CORRELATION_ID>' AND collected_data ? 'verdict' LIMIT 1;
```

⚠ Use the **RUN** correlation the trigger prints as `RUN_CORRELATION_ID`, not the intake one.
And `diagnosis_artifacts` has no `artifact_type` column — it is `kind`.

## Post-roll verification (the fix is inert until an image ships)

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "derived asset deploy path"'   # positive: expect >=1
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "Phase 2E: derived variant deploy path"'  # negative: expect 0
```

⚠ **Run both, on every replica.** A positive control proves the pipeline shipped *something*;
only the negative control (a string the change **removed**) proves it shipped **yours**
(`bugs_open/153`). Mind the case — `grep -ic` if unsure; a mis-cased grep reads as
"not shipped".

---

## Pod-grep the REVALIDATORS after any roll (added 2026-08-11 — these were missing, and I had to re-derive them from source under time pressure)

> **⚠ REWRITTEN 2026-08-11 (evening). The recipe that stood here was wrong twice over** and is kept
> below only as the corrected form. It used `strings /app/agent-chassis` behind `2>/dev/null`,
> which CLAUDE.md now forbids outright — `strings` is absent from the debian-slim images, and behind
> that redirect its absence is indistinguishable from "the needle is not there" (three confidently
> wrong readings in one day). And it looped over `-l app=agent-chassis`, **which returns 2 pods
> while 41 run that image** — a false completeness claim the council caught.

**Step 1 — prove the fleet is ONE binary. This replaces "run it on every replica".**

```sh
kubectl -n ai-persona-system get pods --field-selector=status.phase=Running \
  -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}' \
  | grep agent-chassis | sort | uniq -c
#   41 docker.io/aqls/agent-chassis@sha256:d080ae14…      ← ONE line ⇒ one binary fleet-wide
```

One digest ⇒ a probe of **any single pod** is evidence about all of them, and it is cheaper and
stronger than grepping 26. **More than one line ⇒ a roll is mid-flight; probe each digest.**
`--field-selector=status.phase=Running` is required: a **completed job pod cannot be exec'd at
all**, and its failure reads exactly like a stale binary.

**Step 2 — probe that binary, capturing the EXIT CODE beside the count.**

```sh
P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
      --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec -i "$P" -- sh -c '
for pair in "ownergate:register moved, not the page" \
            "claims:an unbuilt page is not evidence the claims were removed" \
            "voice:voice gate: %d finding" \
            "CONTROL_pos:auto:revalidated" \
            "CONTROL_absent:zzz_needle_that_should_never_exist"; do
  label=${pair%%:*}; needle=${pair#*:}
  n=$(grep -ac -- "$needle" /proc/1/exe); rc=$?
  echo "$label=$n rc=$rc"
done'
```

Expected on a build carrying this lane's work: `ownergate=1 claims=1 voice=1 CONTROL_pos=2
CONTROL_absent=0`, with **rc=0 on the four present needles and rc=1 on the absent one**. Verified
on **v1.0.1279**, **v1.0.1284** and **v1.0.1288**; baseline was `0/0/0` on **v1.0.1270**, which
predates the commit.

⚠ **The `rc` is the point, not decoration.** `grep -c` exits 1 on zero matches, so the old
`n=${n:-0}` idiom was needed — and it **silently converts "I could not look" into "it is not
there"**. Reading `rc` distinguishes them: `rc=1` is a real zero, `rc=126/127` or empty output is a
failed exec. Never report a zero without it.

⚠ **`grep -a` on `/proc/1/exe`, never `strings`, and never a *discovery* grep** for "some 40-hex
string" — that matches Go's internal digit table and returns the same wrong answer on every
service. Probe **known** values only, always with both controls in the same run.

**Gotchas, each one paid for:**
- **`build provenance` (BLD-019) did NOT work here — do not assume it replaces this check.** It is a
  **startup** line, so it scrolls; and the clever fix does not rescue it. The line is at the *start*
  of the log, so `kubectl logs <pod> | head -c 300000 | grep -o 'build provenance[^}]\{0,240\}'`
  should beat `--tail` — on 2026-08-11 it returned **nothing** on a busy `agent-chassis` pod *and*
  on two quiet `agent-build-dispatch-loop` pods sharing the digest (rotation, not absence). With no
  candidate sha to verify, the binary probe is unavailable too. **When it does work it is strictly
  better** (`git merge-base --is-ancestor <your-commit> <the stamp>` answers "did my fix ship?"
  outright) — but an empty result means "not in range", **never** "unstamped".
- **Take the needles from the Go source, not from memory.** `grep -on '"[a-z][^"]\{40,90\}"'` over
  `revalidate_unverified_claims.go` / `revalidate_voice_tells.go` gives usable literals.
- **Avoid apostrophes and em-dashes in the needle** — the real strings contain both
  (`could not load this site's evidence base…`, `— so the site's evidence register moved…`). Single
  quotes in the shell break on the apostrophe, and the em-dash is an emission hazard. The
  substrings above are deliberately plain ASCII and apostrophe-free.
- **`CONTROL_absent` is NOT a true negative control.** It is a fabricated string, so it only proves
  grep can return 0. `bugs_open/153` wants a string the change *removed*, expecting 0 — this lane's
  change removed nothing distinctive, so that control does not exist and the check therefore
  **cannot distinguish v1.0.1284 from v1.0.1279**. Say so rather than implying a stronger proof.
- **Do not use `pc.locked_at IS NULL` as the negative control** (the skip that moved SQL→Go). It
  appears **17 times** across other `platform/` files and would match regardless.

## Did the API usage cap lift? Ask the SUCCESS side

```sql
SELECT date_trunc('hour',created_at), count(*) FILTER (WHERE success) AS ok,
       count(*) FILTER (WHERE NOT success) AS fail
FROM llm_call_log WHERE created_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
```
A resumed `ok` column is the only proof. **The failures stop appearing whether or not the cap
lifted**, so reasoning from the `fail` column cannot tell the two apart — and the vendor's stated
reset date is a worst case, not a forecast (2026-08-10: stated 2026-09-01, actually lifted in
~3h20m because the owner raised it). Cost me a wrong "three-week outage" claim in five files.

---

## CORRECTED 2026-08-11 — the pod-grep above understated the population, and `${n:-0}` hides a failed exec

Both faults found by the council's `debug_historian` seat, against this lane's own claims.

**1. `-l app=agent-chassis` returns 2 pods; 26 run that image.** "Both replicas verified" was a
false completeness claim. Other deployments running the identical chassis binary include
`agent-build-dispatch-loop`, `agent-color-variable-fixer`, `agent-diagnose-agent`,
`agent-content-quality-auditor` and ~16 more.

**The strong proof is the image DIGEST, not a pod count** — cheaper and better than grepping 26:

```sh
kubectl -n ai-persona-system get pods \
  -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}' \
  | grep agent-chassis | sort | uniq -c
# ONE line ⇒ one binary fleet-wide ⇒ a grep on any single pod is evidence about all of them.
# MORE than one line ⇒ a MIXED fleet; grep a pod per distinct digest before claiming anything.
```
Measured 2026-08-11: 21 pods, one digest `sha256:dcd256f9…`, v1.0.1286.

**2. `n=${n:-0}` converts "I could not look" into "it is not there".** A pod reported
`ownergate=0 cachemarker=0` — indistinguishable from a stale binary. It was a **completed job pod**
(`phase Succeeded`); `kubectl exec` refuses those, and the default swallowed the error.

The idiom is still required (`grep -c` exits 1 and prints nothing on zero), so **pair it with a
per-pod positive control and filter to running pods**:

```sh
kubectl -n ai-persona-system get pods --field-selector=status.phase=Running \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' \
  | grep agent-chassis | awk '{print $1}'
```
Then per pod: grep the positive control (`auto:revalidated`, expect ≥1) **first**. A control of 0
means the exec/grep is unusable on that pod — report it as *unverifiable*, never as absence.

⚠ Do not run two `strings` passes in one `exec` across many pods — it times out around 2 minutes
on a fleet this size. One pass per pod, or rely on the digest check above.

---

## Has a late-ladder GATE actually been reached? (added 2026-08-13 — the day a run answered "no")

**Do NOT answer this from the refusal counters.** All three gates read `0` while the code ran 21
times, because every item was decided by an arm *above* them. The counters cannot distinguish
"approved", "never asked", and "the ladder stopped above it". **Read the reasons.**

```sql
-- what actually DECIDED each item on a given run? this is the query that settles it
SELECT result #>> '{revalidation,verdict}' AS verdict,
       left(result #>> '{revalidation,reason}', 95) AS reason_prefix,
       count(*)
FROM site_work_items
WHERE item_type='claims_unverified' AND result #>> '{revalidation,at}' LIKE '2026-08-13%'
GROUP BY 1,2 ORDER BY 3 DESC;
```

Reasons that mean **the gates were never reached**: *"still carries N claim(s) the register does not
support"* (scan still trips) · *"page is absent, or has no component carrying rendered html or stored
content"* · *"site has no current evidence_base spec"*. All three gates sit **downstream of a clean
scan**, so anything that stops the ladder earlier makes their zero uninformative.

**Did a run happen at all, and against which binary?** Two independent facts, both cheap:

```sql
SELECT name, interval_seconds, enabled, last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE name = 'review-queue-revalidate-daily';
```
```sh
# which binary was live at that instant? compare the run time against pod start times,
# NOT against the tag in the makefile (which moves ahead of the fleet)
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  --field-selector=status.phase=Running \
  -o jsonpath='{range .items[*]}{.spec.containers[0].image}{" "}{.status.startTime}{"\n"}{end}'
```
⚠ **The `date LIKE` filter above is the only reliable way to scope to one run**, because
`result->'revalidation'->>'at'` is **last-write-wins** — grouping it without a date filter returns
"the last run that touched each row", which reads exactly like a run history and is not one.

---

## Where does the ladder STOP? (`result.revalidation.arm`, live from the roll after v1.0.1295)

The question the three verdict counters cannot answer. Before `arm` this needed a LIKE over the
prose of `reason`; now it is a key.

```sql
-- WHERE THE LADDER STOPS, per item, on the LATEST sweep
SELECT result #>> '{revalidation,arm}' AS arm, count(*)
FROM site_work_items
WHERE item_type='claims_unverified' AND result ? 'revalidation'
GROUP BY 1 ORDER BY 2 DESC;

-- DID A GATE GET REACHED at all — refusal OR closure. This is the observation
-- a refusal counter structurally cannot give: a gate that refused and a gate
-- never consulted both report nothing.
SELECT count(*) FILTER (WHERE arm LIKE 'gate\_%')                    AS refused_at_a_gate,
       count(*) FILTER (WHERE arm = 'resolved_all_gates_passed')      AS passed_all_gates,
       count(*) FILTER (WHERE arm LIKE 'unreported:%')                AS uninstrumented,
       count(*)                                                       AS total
FROM (SELECT result #>> '{revalidation,arm}' AS arm FROM site_work_items
      WHERE item_type='claims_unverified' AND result ? 'revalidation') s;
```

⚠ **THREE GOTCHAS, each of which returns a plausible wrong answer.**

1. **This is a SNAPSHOT, not a history — the word "ever" does not belong in it.**
   `applyRevalidation` replaces `result.revalidation` **whole** every sweep, so there is exactly
   one arm per item and it belongs to the most recent run. An item that reached a gate last week
   and stops at `scan_still_trips` today shows only today's rung; the earlier reach is **gone,
   not aggregated**. This lane filed that landmine on 08-12 for the `at` stamp and then walked
   into it on 08-13 for `arm` — caught by the council, not by us (`WRONG_CALLS.md`).
2. **`resolved_all_gates_passed` carries NO `gate_` prefix.** It is the one arm proving all three
   gates were reached and passed, named for the closure instead. A prefix-only reach query counts
   only the reaches where a gate **refused** and misses every closure — which inverts the reading.
   Always include the second term, as above.
3. **`LIKE 'gate_%'` treats `_` as a wildcard.** Harmless for today's arm set, wrong in principle;
   write `gate\_%`. And `arm IS NULL` finds nothing — an uninstrumented revalidator records
   `unreported:<item_type>`, deliberately, so a gap cannot read as an absence.

**For a RATE or a TREND, read the per-run surface instead** — one row per sweep, never overwritten:

```sql
SELECT created_at, jsonb_array_length(collected_data->'sweep'->'items') AS decided
FROM orchestration_states WHERE collected_data ? 'sweep' ORDER BY created_at DESC;
```

⚠ `[MEASURED 2026-08-14]` that returns **ONE row** (08-13 08:44:39Z) against 2,532 orchestration
rows going back to 07-13. The surface is structurally right and **effectively empty**; it becomes
a history as runs accumulate. **Do not read a one-row answer as "the sweep has run once."**

## Is a page the audit flagged actually SERVED? (the check that reframed the archived question)

`ScanDeployedClaims` has no page-status filter, and that is **correct** — an `archived` page can
still be serving. Do not "fix" it without this check first, and **always curl a fabricated URL on
the same domain**, or a catch-all 200 reads as a live page:

```bash
for u in "https://<domain>/<page>.html" "https://<domain>/definitely-not-a-real-page-control.html"; do
  printf "%-5s %-9s %s\n" "$(curl -s -o /dev/null -w '%{http_code}' -m 15 "$u")" \
                          "$(curl -s -o /dev/null -w '%{size_download}b' -m 15 "$u")" "$u"
done
```

Worked result 2026-08-14: `robot-hands.com/gripper-catalog.html` → **200, 30,997b while
`status='archived'`** (control 404s), so the scan was right to judge it. Meanwhile
`leopardessconsulting.co.uk/for-engineering-teams.html` → 404 with the **same byte size as its
control**, i.e. genuinely absent. `pages.status` does not discriminate; being served does.

### ⚠ Reading the arm column: `IS NULL` is a VINTAGE marker, not a gap (learned 2026-08-14, hours after the field shipped)

```sql
SELECT status, (result #>> '{revalidation,arm}') IS NULL AS arm_key_absent,
       result #>> '{revalidation,at}' AS stamp, count(*)
FROM site_work_items WHERE item_type='claims_unverified' AND result ? 'revalidation'
GROUP BY 1,2,3 ORDER BY 4 DESC;
```
Returns: 21 `needs_human_review` WITH an arm (stamp `2026-08-14T08:45:05Z`), and **9 `complete` with NO
`arm` key** (8× `2026-08-10`, 1× `2026-08-12`). **A terminal item is never re-swept, so its revalidation
block is frozen at closure.** So `arm IS NULL` = *decided before the instrument shipped*. A
"which revalidators lack arms?" query written the obvious way returns those 9 and reads as a gap.
**The gap check is `arm LIKE 'unreported:%'`.**
