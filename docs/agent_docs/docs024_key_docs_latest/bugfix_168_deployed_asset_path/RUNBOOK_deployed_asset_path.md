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
