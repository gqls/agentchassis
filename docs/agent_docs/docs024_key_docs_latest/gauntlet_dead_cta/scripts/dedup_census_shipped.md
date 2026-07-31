# Re-deriving the duplication census with the SHIPPED rule (not a restatement of it)

**Why this exists.** On 2026-07-31 this lane told its own handoff that enabling
`content_duplication` would "file zero items and delete zero rows", having measured
content identity as `md5(content_data)` — byte identity — while the shipped check
compared *normalised prose*. Wrong ruler, undercounting by construction. Measured
properly, the rule as it then stood would have **deleted a live section from
vonc.com's home page**. See `WRONG_CALLS.md` (2026-07-31) and `RUNBOOK` §16b.

**The rule this encodes.** When the claim is about what Go code will do, execute the
Go code. A SQL or Python restatement is a second definition of "identical", which is
exactly the drift `datahelpers/section_text.go` exists to prevent — so reimplementing
it in order to check it rebuilds the bug you are checking for.

**This is the independent-verification path** the council's editquality seat asked for
(round 2, corr `da3f2d9b-ae6f-492d-ad3b-748323b66367`): the figures below are not to be
taken on any thread's word, they are to be re-run.

## Step 1 — dump every section, one JSON object per line

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -A <<'SQL' > /tmp/pc_dump.jsonl
SELECT json_build_object(
         'pc_id', pc.id, 'page_id', pc.page_id, 'position', pc.position,
         'slot', COALESCE(pc.slot_name,''), 'comp', COALESCE(pc.component_id::text,''),
         'domain', s.domain, 'url', p.url,
         'raw', COALESCE(pc.content_data::text,'{}')
       )::text
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites  s ON s.id = p.site_id
WHERE pc.page_id IS NOT NULL;
SQL
```

`raw` MUST come straight from the jsonb column as `::text` — that form is canonical
(Postgres sorts keys), which is what makes byte comparison a true equality test.

## Step 2 — a scratch module that compiles the REAL functions

Outside the repo, so nothing lands in the shared tree:

```bash
mkdir -p /tmp/census && cd /tmp/census
cat > go.mod <<'MOD'
module census
go 1.21
require github.com/gqls/agentchassis v0.0.0
replace github.com/gqls/agentchassis => /home/ant/projects/agentchassis
MOD
cp <this dir>/dedup_census_shipped.go main.go
go mod tidy && go run . < /tmp/pc_dump.jsonl
```

## Step 3 — what it should say

Measured 2026-07-31 ~12:00Z, after the fix in `43492ec94`:

```
sections=1023  eligible(>=80 chars prose)=815
SHIPPED RULE (page + slot + byte-identical blob): groups=0  rows_deleted=0
```

Run it with `-legacy` to reproduce the pre-fix rule and the false positive it found:

```
LEGACY RULE (page + identical normalised text): groups=1  rows_deleted=1
  vonc.com/index.html  provocation-card@2  <->  lobby-grid@5   (raw blobs byte-identical)
```

**If the post-fix number is not 0, do not enable anything — read the group first.**
And re-run rather than requoting: any session hand-fixing a page moves these figures,
which is how the original claim went stale within hours.
