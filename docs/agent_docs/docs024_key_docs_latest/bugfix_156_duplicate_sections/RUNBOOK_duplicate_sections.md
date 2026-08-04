# RUNBOOK — bugfix 156, duplicate section collapse

Commands that were hard to get right, each with its gotcha attached. Change them **here**,
not in your scrollback.

---

## R1 — The fleet census, and the two traps inside it

This is the query that decides whether the fix is safe. **Never run it without the
`count(DISTINCT md5(...))` column** — slot repetition is not the discriminator, content
identity is, and 11 of the fleet's 12 duplicate groups are legitimate.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
WITH dups AS (SELECT page_id, slot_name FROM page_components GROUP BY 1,2 HAVING count(*)>1)
SELECT s.domain, p.name, pc.slot_name, count(*) AS rows,
       count(DISTINCT md5(pc.content_data::text)) AS distinct_content
FROM page_components pc JOIN dups d USING (page_id, slot_name)
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
GROUP BY 1,2,3 ORDER BY 5,1,2;"
```

**Trap 1 — `distinct_content = 0` is NOT agreement, it is NULL on every row.**
finetuning.uk/our-position-on-ai reports 0. Two rows, both `content_data` NULL. Reading that
as "these two agree" is what makes the bug file's own candidate key delete a live section.

**Trap 2 — a census filtering `content_data IS NOT NULL` cannot see rows of that shape at
all.** The original HANDOFF D census did exactly that and the group was invisible.

**Trap 3 — `content_hash` looks like the column for this and is EMPTY.** No code path writes
it. It cannot stand in for the md5.

Result 2026-08-04: 12 groups, 11 with differing content, 1 the NULL pair, **0 content-identical**.

## R2 — Is anyone else on this bug? (do this BEFORE claiming one)

`who-owns.py` reads COMMITS, so a session mid-fix is invisible to it. Both checks, always:

```bash
python3 scripts/who-owns.py 156          # lagging: commits + workstream dirs
cd /home/ant/.claude/projects/-home-ant-projects-agentchassis && \
for f in $(find . -name "*.jsonl" -mmin -180); do
  echo "=== $f"; tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort | uniq -c | sort -rn | head -5
done
```

The second one is what actually decided this session: 093 and 071 both looked free to
`who-owns.py`-adjacent reasoning and both had live threads on them.

## R3 — Run the unit tests, then RUN THE MUTATIONS

A green suite proves nothing on its own. Back the file up first — it is untracked at that
point, so `git checkout` will not restore it.

```bash
go test ./platform/orchestration/actions/ -run 'TestCollapse|TestAdjacency' -count=1

SC=<scratchpad>; F=platform/orchestration/actions/save_sections_dedup.go
cp $F $SC/dedup_pristine.go
# ...apply one mutation, run the tests, expect the NAMED test to fail...
cp $SC/dedup_pristine.go $F        # restore after EVERY mutation
diff -q $SC/dedup_pristine.go $F   # prove the restore
```

The seven mutations and the tests each must break are listed in the test file header. Two are
load-bearing: identity reduced to `slot_name` (the rejected unique-index rule) and
`rendered_html` dropped from the key (the bug file's own candidate). If either mutation
*passes*, the suite is not distinguishing the shipped rule from a documented wrong answer.

## R4 — Post-roll verification. **The pod-grep is necessary and NOT sufficient**

The grep proves the binary carries the code. It says nothing about whether the guard works.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')

# ✓ NEW code only — expect 0 before the roll, ≥1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  "strings /app/agent-chassis | grep -c 'CONTENT_DUPLICATE_SECTIONS_COLLAPSED'"
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  "strings /app/agent-chassis | grep -c 'DUPLICATE SECTIONS COLLAPSED'"

# ✓ NEGATIVE control — a string this change did NOT add; must be 0, proving the grep discriminates
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  "strings /app/agent-chassis | grep -c 'CONTENT_DUPLICATE_SECTIONS_REFUSED'"

# ✓ POSITIVE control — an already-live sibling marker; must stay 1, proving you are
#   reading the binary you think you are and that grep itself works
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  "strings /app/agent-chassis | grep -c 'CONTENT_CLAIMS_FLOOR_DETAIL'"
```

Run it on **every replica**, not one — `logs deploy/X` and a single-pod exec both read one pod
of N.

## R5 — The induction that actually grades the fix (owed after the roll)

The bug is closed by behaviour, not by a marker. Feed the save a **doubled** section list and
confirm the page ends with the un-doubled count.

```sql
-- BEFORE: the page's current rows
SELECT position, slot_name, left(md5(rendered_html),8) AS html_md5,
       left(md5(content_data::text),8) AS cd_md5
FROM page_components WHERE page_id='<page>' ORDER BY position;

-- AFTER the induced save: expect N rows (not 2N), positions 1..N contiguous
-- and EXACTLY ONE record naming the collapse:
SELECT created_at, severity, error_message,
       context->>'adjacency_signature'  AS signature,
       context->>'arrival_count'        AS arrived,
       context->>'kept_count'           AS kept,
       context->>'sections_source'      AS source
FROM agent_error_log
WHERE error_code='CONTENT_DUPLICATE_SECTIONS_COLLAPSED'
ORDER BY created_at DESC LIMIT 5;
```

**Gotcha carried in from the 093 lane:** to exercise the *section* re-render path at all,
`spec.reason` must be one of `image_landed` / `section_data_resolved` / `cta_links_stale` —
it looks like free-text provenance and is control flow. Any other value takes the
assemble-only branch, which never reads `content_data`. Vary `item_key` for dedup, never the
reason. And insert work items as `status='triaged'`, never `'detected'` (`bugs_open/083`: a
queue with no consumer).

**The negative control for the induction**: run a save of a page with a legitimate repeated
slot (`webdesign.co.uk/index`, `info-card-grid` ×2 with differing content) and confirm the row
count is **unchanged** and **no** `agent_error_log` row appears. A guard that collapses
nothing is indistinguishable from a guard that is not running unless you also show it leaving
the legitimate case alone.

## R6 — Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <submission.json>
# save SUBMISSION_CORR. Budget ~30 min, not ~2 — the council runs in 2-5 min but the
# dispatch queues behind the fleet. A missing orchestration row is LATENCY, not a drop.
```

This lane's correlation: `1a3f4f27-a3b9-4388-b899-a36a911a976e`.

```sql
-- find the run by PAYLOAD, not by the printed id
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '1a3f4f27-a3b9-4388-b899-a36a911a976e';

-- the verdict
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='1a3f4f27-a3b9-4388-b899-a36a911a976e' AND kind='council_report'
ORDER BY created_at;
```

## R7 — Committing into a file another session is editing

`git status` at session start is a snapshot and goes stale in minutes. **Before every commit,
re-run `git diff --numstat <file>` and read the hunks.** This session found another lane's
`bugs_open/190` hunk sitting in `save_page_sections_action.go`, calling a function whose
defining file was **staged but not committed** — a pathspec commit would have put a call to an
undefined symbol into HEAD and broken every `make build-*` fleet-wide. The check that catches
it:

```bash
git diff <file> | grep -E '^@@|^\+' | grep -v '^\+\+\+'    # read every hunk, not the count
git ls-files <any-new-file-the-hunk-calls-into>            # is the callee even tracked?
git diff --cached --stat                                    # what has someone else staged?
```
