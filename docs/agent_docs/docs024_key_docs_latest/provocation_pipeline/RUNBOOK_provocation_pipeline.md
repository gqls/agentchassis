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
