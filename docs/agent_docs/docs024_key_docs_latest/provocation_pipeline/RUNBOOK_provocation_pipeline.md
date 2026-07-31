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

> **GOTCHA — `generated_at` is a HARDCODED LITERAL.** `build_provocations.py:226`
> sets it to the string `"2026-07-26T00:00:00Z"`. Re-running the builder today
> still emits that date. **So this command cannot currently tell a fresh file from
> a stale one** — it is only a freshness check *after* Phase 0 makes the field
> real. Until then, freshness is `gh api repos/gqls/sites/commits?path=...`.

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
