# RUNBOOK — bugs_open/450 tool page shells

Every command here was needed at least once and had a gotcha attached. Change it HERE when it
changes.

## 1. The shell census (is the bug still live, and where)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA <<'EOF'
SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.page_type='tool' AND p.status='active' AND p.deployed_at IS NOT NULL
   AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
                   WHERE pc.page_id=p.id AND pc.build_status<>'removed' AND cc.component_level='tool')
 GROUP BY 1 ORDER BY 2 DESC;
EOF
```

⚠ **This is an upper bound, not a defect count.** On an adopted site a ported tool can live inline
in a non-tool-level component, so it counts as a shell here and is not one. The seotools mechanism
is proven only where `page_component_history` names the writer (query 2). 61 pages / 10 sites as of
2026-09-03.

## 2. Who wrote a suspected shell

```sql
SELECT p.name, w.item_type, count(*) FROM page_component_history h
  JOIN pages p ON p.id=h.page_id
  LEFT JOIN site_work_items w ON w.id=h.source_item_id
 WHERE p.site_id=:site AND p.name LIKE 'tool-%' GROUP BY 1,2;
```

`unbuilt_internal_link` in the `item_type` column is this bug. ⚠ `site_work_items` is a rolling
window — a closed row is archived out of it, so the LEFT JOIN yields NULL for old writes and
undercounts. Absence of the item type is not absence of the mechanism.

## 3. At the body — a tool is a FORM, never a size

```bash
for t in <slugs>; do
  b=$(curl -s "https://<domain>/tools/$t/?cb=$RANDOM")
  printf "%-34s forms=%d inputs=%d selects=%d\n" "$t" \
    "$(grep -o '<form' <<<"$b"|wc -l)" "$(grep -o '<input' <<<"$b"|wc -l)" "$(grep -o '<select' <<<"$b"|wc -l)"
done
```

⚠ **Always probe a known-real tool in the same run as the control** — advertise.co.uk's three read
1 form / up to 11 inputs. Size, status, headline and a 200 all pass on a prose shell; only the form
count discriminates. The cache-buster is not optional (the CDN will serve you a stale body and it
looks like a result).

## 4. The 090 verdict for this bug

```sql
SELECT result FROM site_work_items
 WHERE spec->>'dispatch_correlation_id' LIKE '96e97dc4%' AND item_type='needs_diagnosis';
```

⚠ **NOT in `doc_notes`** — the `doc_notes` query returns nothing for this run and reads as "no
verdict". The verdict lives in the item's `result`. Output is ~45 KB; pipe it through
`python3 -c "import json,sys; print(json.load(sys.stdin)['response']['response']['conclusion'])"`
rather than reading it raw.

## 5. Ownership, before routing anything at a bug

```bash
python3 scripts/who-owns.py <number|slug>
```

⚠ It reads COMMITS, so a session mid-fix is invisible; check the tree too, and re-run at each
phase boundary (the answer goes stale in minutes on this tree). For 450 it names the FILING lane —
which owns the instance, not the class. Read the lane's handoff before concluding it owns the fix.

## 6. Build and prove the change

```bash
scripts/verify-head-builds.sh --with <file> [--test]      # BEFORE committing
scripts/verify-head-builds.sh [targets]                   # AFTER committing
go test ./platform/orchestration/actions/ -run 'ToolShell|GenericBuild|OwnedPage'
```

⚠ Never hand-roll `git archive HEAD | tar` — that recipe is why the machine runs out of space.
⚠ `/tmp` is a 16 GB tmpfs; a full one reports as `link: mapping output file failed: no space left
on device`, which reads like a compiler fault and is not one.

## 7. Prove it shipped (after the next roll)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <commit-sha> <stamped-sha> && echo SHIPPED
```

⚠ The provenance line is a STARTUP line and scrolls out of reach on a busy service — an empty
result means "not in range", not "unstamped". Fall back to the binary probe with **both** controls
(a sha that must be present and one that must be absent):

```bash
kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe
```

Never `strings` (absent from the image) and never a discovery grep for "some 40-hex string" (it
matches Go's internal digit table and returns the same wrong answer on every service).
⚠ Read the stamp of the **service you mean** — one release can straddle several commits.

## 8. The demand control (a post-fix zero is not evidence)

After the roll, the fix's *positive* signal — not the absence of new shells:

```sql
-- a queued item at a shell target should now terminate wont_fix with a receipt
SELECT w.item_type, w.status, w.result->>'reason'
  FROM site_work_items w JOIN pages p ON p.id = w.page_id
 WHERE w.item_type='unbuilt_internal_link' AND p.page_type='tool'
 ORDER BY w.updated_at DESC LIMIT 20;
-- and the receipt itself
SELECT summary, spec->>'refusal_class' FROM site_work_items
 WHERE item_type='owned_page_review' AND spec->>'refusal_class'='tool_pending'
 ORDER BY created_at DESC LIMIT 10;
```

Positive control in the same window: a **non**-shell page building normally. Without it, a
zero means "nothing tried", not "the guard held".

## 8b. ⚠ Do NOT verify this fix with a re-render (until 9831e9ab4 rolls)

Flagged by the `bugs_open/427`/`454` lane 2026-09-03, and it would have cost this lane a day if
we had reached for it: since 2026-09-02 **every light re-render renders the page's own stored
`content_data` back at itself** with no freshly resolved data (`classifyStoredSection` dropped its
section plan). The run reports clean, the `rerendered` count is healthy, and nothing was
delivered. Both live agent-chassis builds still carried the defect at 09:54Z.

So a check of the form "re-render the page and read the result to see whether the fix took" is
reading a mirror. The verification in §8 deliberately reads **work-item terminal status +
receipt + the served body**, never a re-render, and must stay that way until a chassis image
carrying `9831e9ab4` is live. Full account: `bugs_open/454`.

## 9. Council submission

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <sub.json>
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <sub.json>
```

Save the printed `SUBMISSION_CORR`; budget ~30 minutes (dispatch queues behind the fleet — 29
minutes publish→start is normal). Find the run by payload, never by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

⚠ Do not submit within ~300 s of a chassis roll, and expect a roll to kill an in-flight council run.
