# RUNBOOK — provocation pipeline

Commands that were hard to get right, with the gotcha attached. Newest sections
appended.

## 1. Read the live feed and its real freshness

```bash
curl -s -w "\nHTTP:%{http_code} BYTES:%{size_download}\n" \
  https://vonc.com/data/provocations.json -o /tmp/prov.json
python3 -c "
import json; d=json.load(open('/tmp/prov.json'))
print('generated_at:', d['generated_at'])
print('today keys:', list(d['today'].keys()))
print('archive entries:', len(d['archive']['entries']))
"
```

> ~~**GOTCHA — `generated_at` is a HARDCODED LITERAL.** `build_provocations.py:226`
> sets it to the string `"2026-07-26T00:00:00Z"` … this command cannot tell a fresh
> file from a stale one.~~
>
> **CORRECTED 2026-07-31 — FIXED AND LIVE.** `builder/build_provocations.py`
> computes it, and the served file now carries a real stamp
> (`2026-07-31T15:03:31Z` at publish). **So this command IS a valid freshness
> check again**, from that timestamp onward.
>
> **Two carve-outs, both live:** the *old* builder at
> `gauntlet_dead_cta/p4_sources/build_provocations.py` still hardcodes the literal
> — run that one and you silently re-stamp the feed to 26 Jul. And a real
> `generated_at` proves the file was **rebuilt**, not that the provocation
> **changed**: a daily rebuild of an unchanged schedule moves the timestamp every
> day while the site says the same thing. For "did it rotate", compare
> `today.slug`, never `generated_at`.

## 2. Confirm nothing regenerates it

```sql
SELECT count(*) FROM scheduled_tasks
WHERE input_data::text ~* 'vonc|provocation'
   OR COALESCE(pre_query,'') ~* 'vonc|provocation'
   OR name ~* 'vonc|provocation';   -- 0 as of 2026-07-31
```

> **GOTCHA — `scheduled_tasks` has no `next_run_at` column.** The pacing columns
> are `interval_seconds`, `last_triggered_at`, `last_completed_at`. Run
> `\d scheduled_tasks` before writing the SELECT; guessing costs a round trip.

## 3. Prove the reusable publish path is alive (do this before believing §1a(ii) of the PLAN)

Two halves — the schedule, and the artefact it produced. **Both are needed; the
row alone proves only that something is configured.**

```sql
SELECT name, enabled, interval_seconds, target_agent_type,
       last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE name ~* 'feed|news|asset|render';
```

```bash
for d in dartsonline.com relojistas.com webdesign.co.uk; do
  printf "%-24s " "$d"
  curl -s -m 12 "https://$d/data/latest-news.json" \
    | python3 -c "import json,sys; print(json.load(sys.stdin)['updated_at'])"
done
```

Pass condition: the `updated_at` stamps fall inside the
`last_triggered_at`→`last_completed_at` window of the most recent run. Measured
2026-07-30: run 13:53:54→14:02:29Z, artefacts 13:56 / 13:58 / 14:01. **That is the
proof the `git_commit` → S3 path works unattended.**

## 4. Render a client-side page without playwright

`scripts/provocation_visibility.py` needs playwright, which is **not installed**
(`import playwright` fails). Chromium is present at `/snap/bin/chromium`:

```bash
cd /home/ant && timeout 90 /snap/bin/chromium --headless --disable-gpu \
  --no-sandbox --dump-dom --virtual-time-budget=9000 \
  "https://vonc.com/provocations/index.html" > /home/ant/prov_index_dom.html
```

> **GOTCHA 1 — run it from `$HOME`, not the scratchpad.** Snap confinement blocks
> writes outside the home tree; the redirect fails silently-ish and you debug the
> wrong thing.
>
> **GOTCHA 2 — `--dump-dom` gives you the DOM, NOT what is painted.** This is the
> same trap `HANDOFF_2026-07-30_C` names. A `hidden` attribute or a `display:none`
> rule leaves the text fully present in the dump. **`grep -c` on this file
> answers "is it in the DOM", which is a different question from "can a visitor
> read it".** Always print surrounding context:
>
> ```bash
> python3 -c "
> h=open('/home/ant/prov_index_dom.html').read(); i=h.find('YOUR STRING')
> import re; print(re.sub(r'\s+',' ', h[max(0,i-600):i+200]))"
> ```
>
> This caught two errors in one sitting: an 'empty state' that grep said was
> present (it carries `hidden=""`), and a 'blank leading row' that turned out to
> be a correctly hidden `data-archive-template`.
>
> **GOTCHA 3 — a substring is not an identification.** `"Nobody actually"` matched
> the 29 Jun archive entry *"Nobody actually reads terms of service"*, not today's
> headline *"Nobody actually wants a personalised internet"*. I briefly recorded
> today's provocation as leaking onto the archive page on that basis. Match on
> enough text to be unambiguous, and read the context you matched.

## 5. Check which pieces of the 2026-06-25 plan were actually built

```sql
SELECT type, is_active FROM agent_definitions
WHERE type ~* 'provocation|feed-ingester|feed-triage|content-feed|asset-renderer'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
ORDER BY type;
```

2026-07-30: returns the five news-feed agents and **no `provocation-*` row** —
Phases 3–4 were never built. *(Per `seed-sql-is-history-live-row-is-fact`: ask
`agent_definitions`, not the seed files, which record what an agent once was.)*

## 6. Pages and components for the archive

```sql
SELECT p.name, p.url, count(pc.id) AS components
FROM pages p LEFT JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'
  AND (p.name ~* 'provocation' OR p.url ~* 'provocation')
GROUP BY p.name, p.url;
```

> **GOTCHA — read `pages.url`, never construct one.** Returns
> `provocations-index → /provocations/index.html` (1 component) and a stray
> `provocation → /blog/provocation.html` (0 components). The 2026-06-25 plan's
> claim that the index page has "zero components" is **stale** — it has one,
> `provocations-archive-list`, `build_status=deployed`.

## 7. Build and publish the feed (Phase 0, live 2026-07-31)

```bash
cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/builder
python3 build_provocations.py > /tmp/feed.json      # --date YYYY-MM-DD to test another day
python3 verify_rotation.py                          # invariants across the whole schedule
./publish_feed.sh --dry-run /tmp/feed.json          # preflight + target + sha, writes nothing
./publish_feed.sh /tmp/feed.json "vonc.com: ..."    # PUT, then polls the SERVED file
```

`publish_feed.sh` does not exit 0 until the **served** bytes match what it
pushed (or it fails after 10 minutes and says so). Measured 2026-07-31: GitHub
PUT → B2/CDN served in **~45 s**.

> **GOTCHA — `sites.github_repo` is EMPTY for vonc.com**, so the usual "read the
> deploy repo from the DB" move returns nothing and cannot confirm the target.
> The repo is `gqls/sites`, path `vonc.com/data/provocations.json`. **Prove it
> before writing** rather than trusting the runbook — fetch the blob and compare
> against the served bytes:
> ```bash
> gh api repos/gqls/sites/contents/vonc.com/data/provocations.json --jq '.content' \
>   | base64 -d > /tmp/blob.json
> curl -s https://vonc.com/data/provocations.json > /tmp/served.json
> cmp /tmp/blob.json /tmp/served.json && echo "this repo path IS what the site serves"
> ```
> This is the "wrong repo succeeds silently" landmine: a PUT to a plausible-looking
> path returns 200 and changes nothing a visitor sees.

### Rolling back

Same script — there is no separate revert path, deliberately, so publishing
forward exercises the mechanism a rollback would use.

```bash
./publish_feed.sh --rollback backups/provocations_2026-07-31_pre_phase0.json "revert"
```

> **GOTCHA, found by dry-running the rollback rather than by reasoning:
> `--rollback` EXISTS because the first version of the preflight refused to roll
> back.** The pre-Phase-0 file has no `today.slug` — that is the defect Phase 0
> fixes — and the preflight required one, so the escape hatch was gated on the
> very improvement it exists to undo. The preflight is now two tiers:
> **safety** checks (the fields the live loader reads, `today` present, no
> duplicate slugs, today not in the archive) run on **every** path, because
> failing them is an outage — `round.go` 503s when `today` is missing, by design.
> **Quality** checks (slug/date) are skipped under `--rollback`.
> **Generalise it: any guard on a publish path must be checked against the oldest
> artefact you might need to restore, not just the newest one you intend to ship.**

### Post-publish regression check (render, do not grep)

```bash
cd /home/ant && for p in "" "provocations/index.html" "tools/gauntlet/index.html"; do
  timeout 90 /snap/bin/chromium --headless --disable-gpu --no-sandbox --dump-dom \
    --virtual-time-budget=9000 "https://vonc.com/$p" > "/home/ant/r_${p//\//_}.html" 2>/dev/null
done
```

Assert three things, because they can fail independently: home still **paints**
the headline and body; the archive still shows 8 entries with the empty state
`hidden` and **zero** occurrences of today's slug or body; the gauntlet page still
carries `gi-sealed` with **zero** occurrences of either. Verified all three
2026-07-31 after the Phase 0 publish.

---

## The pool and the publisher (added 2026-07-31, migrations 282 + 283)

### Where the schedule lives now

The Python `SCHEDULE` literal is no longer the source of truth for anything
deployable. The pool is `provocations` in `clients_db`.

```sql
-- what will be served today, and what the archive will hold
SELECT slug, publish_on, status,
       (headline <> '') AS has_today_shape, (detail_body <> '') AS has_case
FROM provocations
WHERE domain = 'vonc.com' AND status = 'approved' AND publish_on IS NOT NULL
ORDER BY publish_on DESC;
```

Selection rule, mirrored exactly in Go and Python: **today = the latest approved
row whose `publish_on` has arrived; archive = everything approved and strictly
earlier.** Nothing marks a row as published — that is a fact about dates, and
storing it would create a second copy that can disagree.

### Adding a provocation

```sql
INSERT INTO provocations
  (domain, slug, category, publish_on, status, title, teaser, headline, body, detail_body)
VALUES ('vonc.com', 'some-slug', 'general', DATE '2026-08-03', 'approved',
        'The archive title', 'One-line teaser.',
        'The <em>today</em> headline.', 'The long-form today body.',
        'The full case.' || E'\n\n' || 'Second paragraph.');
```

**GOTCHA — author BOTH shapes for anything new.** `title`/`teaser`/`detail_body`
is the archive shape; `headline`/`body` is the today shape. The eight historical
entries have only the first, so the action falls back (headline←title,
body←detail_body←teaser). The fallback is deliberate for them and a silent
downgrade for anything new.

**GOTCHA — a duplicate `publish_on` is refused by the database**, not by the
action: `idx_provocations_one_per_day` is a partial unique index over approved,
dated rows. That is on purpose — an ambiguous "latest" would decide the day's
provocation by plan order. If you get a unique violation, two rows want the same
day and you must choose.

### Enabling the schedule (NOT yet done — the image has to ship first)

```bash
# 1. Is the action actually in the running binary? A roll is NOT evidence.
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis \
        -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system "$POD" -- sh -c \
  'echo -n "target: "; strings /app/agent-chassis | grep -c render_provocation_feed;
   echo -n "control: "; strings /app/agent-chassis | grep -c render_news_section'
```

**GOTCHA — the positive control is not optional.** A bare `grep -c` returning 0
cannot distinguish "not shipped" from "my grep is wrong". Measured 2026-07-31:
target 0, control 3 — so the grep works and the answer is a real no.

```sql
-- 2. Only once target > 0:
UPDATE scheduled_tasks SET enabled = true WHERE name = 'provocation-feed-refresh';
```

### Did it actually rotate?

```bash
curl -s https://vonc.com/data/provocations.json \
  | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['today']['slug'], '|', d['generated_at'])"
```

**GOTCHA — compare `today.slug`, never `generated_at`.** Once anything rebuilds
on a cadence the timestamp advances daily whether or not the provocation changed.
The action deliberately SKIPS the commit when only the timestamp would move, so a
static `generated_at` across days is now the CORRECT behaviour for a day with no
rotation, not a symptom.

### Regenerating the parity fixtures

Both, together, or not at all:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT json_agg(json_build_object('slug',slug,'publish_on',to_char(publish_on,'YYYY-MM-DD'),
     'title',title,'teaser',teaser,'card_desc',COALESCE(card_desc,''),
     'detail_body',COALESCE(detail_body,''),'headline',COALESCE(headline,''),
     'body',COALESCE(body,'')) ORDER BY publish_on)
   FROM provocations WHERE domain='vonc.com' AND status='approved' AND publish_on IS NOT NULL;" \
  > platform/orchestration/actions/testdata/provocation_pool.json

cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/builder
python3 build_provocations.py --date 2026-07-31 > /tmp/g.json
# then set generated_at to the literal GOLDEN and save as testdata/provocation_feed_golden.json
```

**GOTCHA — never refresh the golden from the Go side.** It exists to catch the Go
port drifting from the Python builder; regenerated from the code under test it can
only ever agree with it, which is a green test that checks nothing.

### Testing when the working tree will not compile

Several sessions share this tree, and another lane's uncommitted WIP can break the
package for everyone. Test against committed HEAD plus your own files:

```bash
SB=$(mktemp -d); git archive HEAD | tar -x -C "$SB"
cp platform/orchestration/actions/provocation_feed*.go "$SB/platform/orchestration/actions/"
mkdir -p "$SB/platform/orchestration/actions/testdata"
cp platform/orchestration/actions/testdata/provocation_*.json "$SB/platform/orchestration/actions/testdata/"
(cd "$SB" && go test ./platform/orchestration/actions/ -count=1)
```

This is also the honest check: `make build-*` builds from HEAD, so HEAD is what
ships — not your tree.

### Did the publisher actually run? (the monitorable, added after council round 2)

The action updates `last_completed_at` **only** when it publishes or deliberately
skips an unchanged feed. Every refusal path — empty pool, failed seal or engine
invariant, unreadable served feed, shrinking archive — returns an error and leaves
it alone. So staleness is the signal:

```sql
SELECT name, enabled, last_triggered_at, last_completed_at,
       (last_triggered_at > last_completed_at) AS last_run_did_not_complete,
       now() - last_completed_at AS since_success
FROM scheduled_tasks WHERE name = 'provocation-feed-refresh';
```

**GOTCHA — a healthy day looks identical to a broken one in `generated_at`, but
NOT here.** On a day with no rotation the action skips the commit, which still
counts as success and still moves `last_completed_at`. So `since_success` growing
beyond the 6h interval means the action is *failing*, not merely quiet. Nothing
alerts on this yet — a council reviewer flagged exactly that gap.

To see why it refused, read the orchestration rather than guessing:

```sql
SELECT current_step, status, error, updated_at
FROM orchestration_states
WHERE current_step = 'publish_feed'
ORDER BY updated_at DESC LIMIT 5;
```

---

## §N — going live, and proving both paths (2026-08-02)

### Prove the action is in the running binary (never trust the tag)

```bash
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  kubectl exec -n ai-persona-system ${p#pod/} -- sh -c '
    strings /app/agent-chassis | grep -c "render_provocation_feed"          # expect >0
    strings /app/agent-chassis | grep -c "deploy_image_asset"               # positive control
    strings /app/agent-chassis | grep -c "render_provocation_feed_NOT_REAL" # negative, expect 0'
done
```

**Gotcha.** This change was purely additive, so there is no string it REMOVED and
the ideal negative control does not exist. A synthetic near-miss proves the grep
discriminates but NOT that the image postdates your commit. Get provenance
separately: `git merge-base --is-ancestor <your-commit> HEAD`.

### Predict the run before you enable it

Run the oracle for today and diff against what is served. If they match apart
from `generated_at`, a live run must skip — which makes it a safe first firing.

```bash
python3 docs/.../provocation_pipeline/builder/build_provocations.py --date $(date -u +%F) > /tmp/oracle.json
curl -s https://vonc.com/data/provocations.json > /tmp/served.json
python3 -c "
import json
a=json.load(open('/tmp/served.json')); b=json.load(open('/tmp/oracle.json'))
a.pop('generated_at'); b.pop('generated_at'); print('will skip:', a==b)"
```

**Gotcha.** This comparison is STRUCTURAL — it unmarshals first. It cannot see an
encoding or key-order difference, which is exactly how the escaped-markup defect
shipped. To compare what is actually written, diff the bytes:
`diff <(python3 build_provocations.py --date …) <(gh api …/contents/… --jq .content | base64 -d)`.

### Enable, and watch it fire

```sql
UPDATE scheduled_tasks SET enabled = true, updated_at = now()
WHERE name = 'provocation-feed-refresh';
-- both timestamp columns NULL ⇒ due immediately; it fires within ~60s
SELECT last_triggered_at, last_completed_at FROM scheduled_tasks
WHERE name = 'provocation-feed-refresh';
```

Read what it DID, not that it completed:

```sql
SELECT jsonb_pretty(collected_data->'complete'->'result'->'publish_feed'), status, error
FROM orchestration_states
WHERE collected_data->'input_data'->>'task_name' = 'provocation-feed-refresh'
ORDER BY created_at DESC LIMIT 1;
```

`committed: false` + `reason: "no change since the served feed"` is the skip path.

### Induce the commit path deliberately

Only do this on a day the content is provably identical, so the sole possible
change is the timestamp.

```sql
UPDATE scheduled_tasks
SET input_data = input_data || '{"force_commit": true}'::jsonb,
    last_triggered_at = NULL, last_completed_at = NULL, updated_at = now()
WHERE name = 'provocation-feed-refresh';
-- ... wait for it to fire, confirm, then ALWAYS restore:
UPDATE scheduled_tasks SET input_data = input_data - 'force_commit', updated_at = now()
WHERE name = 'provocation-feed-refresh';
SELECT (input_data ? 'force_commit') AS still_set FROM scheduled_tasks
WHERE name = 'provocation-feed-refresh';   -- must be f
```

### Verify at the artefact, in the repo

```bash
gh api "repos/gqls/sites/commits?path=vonc.com/data/provocations.json&per_page=3" \
  --jq '.[] | "\(.sha[0:9])  \(.commit.author.date)  \(.commit.message | split("\n")[0])"'
gh api "repos/gqls/sites/commits/<sha>" --jq '.files[] | "\(.filename)  +\(.additions) -\(.deletions)"'
```

**A +N/−N diff where N is the whole file means the writer changed, not the
content.** That is the signal that found the encoding defect.

### The monitorable: a refusal is a STALE timestamp, not an error row

`last_completed_at` moves only on the success paths (published, or skipped as
unchanged). A refusal — unreachable served feed, failed `checkFeed`, shrink guard
— leaves it where it was. So staleness is the signal; there is no error row to
look for.

```sql
SELECT name, enabled, last_triggered_at, last_completed_at,
       now() - last_completed_at AS since_success
FROM scheduled_tasks WHERE name = 'provocation-feed-refresh';
-- interval is 21600s (6h); since_success much beyond that ⇒ it is refusing
```
