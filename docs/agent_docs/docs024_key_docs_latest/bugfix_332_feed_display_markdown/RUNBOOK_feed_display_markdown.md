# RUNBOOK — bugs_open/332 feed display markdown

Every command here was hard to get right at least once. The gotcha is attached to the command,
not filed at the bottom.

---

## 0. Two gates that must be re-run on the day, before shipping detection

Written down **before** they were run, so the disconfirming result is on record ahead of the
answer.

### Gate A — would a new pattern revoke the section-editor repair route?

`check_literal_markdown.go:251 transformRouteSlot` routes to `section-editor` **iff every
finding is `code_span`**. A new pattern co-firing on a ported page silently sends it to
`page-rerender`, which is inapplicable there **by construction** (migration 499).

```sql
SELECT p.url, pc.slot_name, COALESCE(cc.name,'(none)')
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  LEFT JOIN content_components cc ON cc.id=pc.component_id
 WHERE pc.locked_at IS NULL
   AND pc.rendered_html ~ '`[A-Za-z0-9][^`]{0,80}`'
   AND ( pc.rendered_html ~ '!\[[A-Za-z]'
      OR pc.rendered_html ~ '\]\((https?://|/)[^)" ]{0,200}\.\.\.'
      OR pc.content_data::text ~ '!\[[A-Za-z]' );
```

**ALWAYS RUN THE CONTROL IN THE SAME BREATH** — an empty result from a query that can never
match is not evidence:

```sql
SELECT count(*) FROM page_components
 WHERE locked_at IS NULL AND rendered_html ~ '`[A-Za-z0-9][^`]{0,80}`';
-- 2026-09-03: 121. Non-zero, so the gate query discriminates and its 0 rows are a real zero.
```

**Disconfirming: any row.** Then ship tier 1 **strip-only** (out of `LiteralMarkdownPatterns`)
until the HTML-surface transform has a sibling that can repair the new shapes.

### Gate B — promoter-floor headroom

```sql
SELECT handler_agent,
       count(*) FILTER (WHERE status IN ('complete','verified')) AS good,
       count(*) FILTER (WHERE status='failed')                   AS failed,
       count(*)                                                  AS all_rows
  FROM site_work_items WHERE item_type='literal_markdown'
 GROUP BY 1 ORDER BY 4 DESC;
-- 2026-09-03: page-build-handler 23/0/24 · page-rerender 2/0/2 · section-editor 2/0/2.
```

**Disconfirming: `page-rerender` below 25% complete+verified on ≥5 terminals.** The 444/454
promoter holds the pair, new items are never claimed, detection is inert this roll. Ship the
projection alone — it fixes the served artefact regardless.

---

## 1. Censuses

### ⚠ Postgres regex repetition counts max out at 255

`{0,300}` raises `ERROR: invalid regular expression: invalid repetition count(s)` — which
reads like a broken query and is a **limit**, not a syntax error. It cost a bisect. Use
`{0,200}`.

### ⚠ Use a quoted heredoc, not `-c "…"`

In a double-quoted `-c` string the shell eats `\$`, so a `$`-anchored regex silently becomes
unanchored and the count is wrong in the safe-looking direction. Always:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F'|' <<'SQL'
...
SQL
```

### ⚠ The served JSON is `MarshalIndent` output — match `": *"`, never `":"`

`grep -oE '"(summary|title)":"[^"]*"'` scores **0** on a file measured minutes earlier as
carrying 7 headings and 9 links, because every key is followed by a colon AND A SPACE. The
first cut of the sweep's JSON arm had exactly this bug and reported a clean site as clean for
the wrong reason. Caught only by running the new check on its own motivating case — which is
the general rule, not a note about this regex.

### The shape census

```sql
SELECT
 count(*) FILTER (WHERE source_summary ~ '\]\([^)]*$')                       AS tail_link_target,
 count(*) FILTER (WHERE source_summary ~ '\[[A-Za-z][^\]]{0,80}$')           AS tail_bracket_only,
 count(*) FILTER (WHERE source_summary ~ '!\[[A-Za-z]')                      AS img_letter_alt,
 count(*) FILTER (WHERE source_summary ~ '!\[\]\(')                          AS img_empty_alt,
 count(*) FILTER (WHERE source_summary ~ '\*\*[A-Za-z][^*]{0,200}$')         AS tail_bold,
 count(*) FILTER (WHERE source_summary ~ '(^|\n)[-*] [A-Za-z]')              AS list_marker,
 count(*) FILTER (WHERE source_summary ~ '(^|\n)#{1,6} ')                    AS heading,
 count(*) FILTER (WHERE source_summary ~ '\*\*[A-Za-z][^*\n]{0,80}\*\*')     AS bold_complete,
 count(*) FILTER (WHERE source_summary LIKE '%'||U&'\FFFD'||'%')             AS rune_split,
 count(*)                                                                    AS total
 FROM content_feed_items WHERE created_at > now() - interval '30 days';
-- 2026-09-03: 288 | 70 | 93 | 30 | 15 | 94 | 1177 | 105 | 2 | 5863
```

**A COUNT MUST CARRY THE DATE IT WAS COUNTED.** This one goes stale by ADDITION — the feed
adds rows every few hours. Before quoting it, re-run it.

### Which source type does the damage — the query that reframed the bug

```sql
SELECT cs.source_type, count(*) AS rows,
       count(*) FILTER (WHERE length(cfi.source_summary)=200) AS len200,
       count(*) FILTER (WHERE cfi.source_summary ~ '\]\([^)]*$') AS unclosed
  FROM content_feed_items cfi LEFT JOIN content_sources cs ON cs.id=cfi.source_id
 WHERE cfi.created_at > now() - interval '30 days' GROUP BY 1 ORDER BY 2 DESC;
-- 2026-09-03: news_search 4348/941/288 · rss 834/2/0 · scrape 472/0/0 · api_news 209/0/0
```

`rss` carries **zero** markdown. That is why relojistas measured clean in August and 332 read
as latent — and it is the single most useful line in this file.

### Are the affected components self-healing?

```sql
SELECT s.domain, pc.slot_name, pc.updated_at::timestamp(0), (now()-pc.updated_at) AS age
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE COALESCE(pc.rendered_html,'') ~ '\]\((https?://|/)[^)" ]{0,200}\.\.\.'
    OR COALESCE(pc.rendered_html,'') ~ '!\[[^\]]{0,80}\]\((https?://|/)'
 ORDER BY pc.updated_at DESC;
-- 2026-09-03 16:20Z: all 9 rewritten within 19h, three within the hour.
```

This is why there is **no repair campaign**: the feed cycle rewrites these slots continuously
and `queueNewsPageRerenders` re-resolves the query each refresh, so a producer-side fix
repairs every affected page on its own within about a day.

---

## 2. Probing the served artefacts

### ⚠ Probe the slug on the serving host, never the customer domain

`boxingonline.com` is a parked catch-all: an invented path returns **200**. The serving host
is `boxingonline.ugg2.com` (`sites.publish_target='b2worker'`), which 404s correctly.

### ⚠ Control first — and one file per host

```bash
S=$SCRATCH
for u in https://boxingonline.ugg2.com/news/index.html \
         https://fundamentallyai.com/news/index.html \
         https://robot-hands.com/news/index.html \
         https://ai-agent-orchestration.com/news/index.html \
         https://idea.uk/news/index.html; do
  host=$(echo "$u" | cut -d/ -f3); f="$S/p_${host}.html"; rm -f "$f"
  code=$(curl -sS -m 25 -o "$f" -w '%{http_code}' "$u" || echo 000)
  ctl=$(curl -sS -m 25 -o /dev/null -w '%{http_code}' "https://$host/news/zzz-not-a-real-page-9x.html" || echo ERR)
  if [ "$code" = 200 ] && [ -s "$f" ]; then
    printf '%-32s ctl %s  ](http=%-3s ![=%-3s **A=%-3s\n' "$host" "$ctl" \
      "$(grep -o '](http' "$f" | wc -l)" "$(grep -o '!\[' "$f" | wc -l)" "$(grep -o '\*\*[A-Za-z]' "$f" | wc -l)"
  else
    printf '%-32s ctl %s  NO MEASUREMENT (fetch failed)\n' "$host" "$ctl"
  fi
done
```

**Both guards are load-bearing and both were learned the hard way.** `rm -f "$f"` plus the
per-host filename: `curl -o` only overwrites on success, so a shared `page.tmp` makes a failed
fetch report the *previous* host's numbers. The explicit NO MEASUREMENT branch is what stops a
000 reading as a clean page. → WRONG_CALLS 2026-09-03.

2026-09-03 baseline (pre-fix): boxingonline 5/1/1 · fundamentallyai 3/0/0 · robot-hands 2/0/0
· ai-agent-orchestration 2/1/0 · idea.uk 2/1/0. Controls all 404.

### The JSON surfaces — the check the sweep cannot see, and ⚠ its expiry date

> **⚠ CORRECTED 2026-09-03 (the feed lane's catch, on their own doorstep first).** This section
> called the JSON scan "the load-bearing check". **That is true only until the projection
> rolls, and false immediately after**, so the phrase is retired here.
>
> **A clean `/data/news-archive.json` post-roll means the strip RAN. It says nothing about
> what is in `content_feed_items`.** A table full of raw markdown and a spotless one produce
> byte-identical served output once the projection is in front of them, so the check cannot
> come out any other way — an undisconfirmable measurement wearing the clothes of this
> estate's most-repeated invariant, *"judge at the served artefact"*. That invariant is what
> walks you into it.
>
> **Split the two questions and never let one answer the other:**
> - *Is the visitor seeing junk?* → the served page and the served JSON. The projection is
>   what fixes it, and a zero here is a real pass **for that question**.
> - *Is my ingestion clean?* → the **column**, never the surface:
>   `SELECT count(*) FILTER (WHERE source_summary ~ '\]\([^)]*$') FROM content_feed_items
>   WHERE site_id = '<id>' AND created_at > now() - interval '30 days';`
>
> **And the sharpest form, which is a property of moving the kill switch into the projection:**
> flipping `DISABLE_NEWS_MARKDOWN_STRIP` would fill yesterday's "verified clean" surfaces with
> junk **that was in the table the whole time**. Good behaviour from the switch; a trap for
> anyone reading an old verification and calling it a regression.

```bash
curl -s "https://boxingonline.ugg2.com/data/news-archive.json?cb=$RANDOM" -o "$S/na.json"
python3 - <<'EOF'
import json,re
d=json.load(open('/…/na.json')); items=d.get('items',[])
pats={'heading':r'^#{1,6} ','md_link':r'\]\(','bold':r'\*\*[A-Za-z]','img':r'!\[','list':r'^[-*] '}
from collections import Counter; c=Counter()
for it in items:
    for f in ('title','summary'):
        v=it.get(f) or ''
        for n,p in pats.items():
            if re.search(p,v,re.M): c[n]+=1
print('items:',len(items),'defects:',dict(c))
EOF
# 2026-09-03 baseline: news-archive 20 items -> heading 7, md_link 9, list 1, img 1, bold 1
#                      latest-news   6 items -> heading 3, md_link 4, list 3
```

### ⚠ A client-side behaviour is not absent because the page HTML lacks its code

`grep news-archive.json` over the page HTML returns **0** — and the fetch happens anyway,
because the script is an external asset. Follow the `<script src>`:

```bash
grep -o '<script[^>]*news-listing[^>]*>' "$S/p_boxingonline.ugg2.com.html"
curl -sS -o "$S/nl.js" -w '%{http_code} %{size_download}B\n' \
  https://boxingonline.ugg2.com/tools/assets/news-listing.js
grep -c 'news-archive.json' "$S/nl.js"     # 1
grep -c 'container.innerHTML = html' "$S/nl.js"
```

→ WRONG_CALLS 2026-09-03. Live on all five news hosts; `latest-news.js` is 200 on idea.uk,
ai-agent-orchestration.com and robot-hands.com (404 on boxingonline and fundamentallyai) —
each verified against a `/tools/assets/zzz-not-real.js` 404 control.

### feed.xml — still exactly one enabled site

```sql
SELECT domain FROM sites WHERE deploy_config->'rss_feed'->>'enabled'='true';
-- 2026-09-03: relojistas.com. Unchanged since the bug was filed.
```

```bash
curl -s "https://relojistas.com/feed.xml?cb=$RANDOM" > "$S/feed.xml"
sed -n 's/.*<description>\(.*\)<\/description>.*/\1/p' "$S/feed.xml" | grep -cE '\]\(|!\[|\*\*[A-Za-z]|^#{1,6} '
grep -c '<item>' "$S/feed.xml"
```

**This reads 0 today as well**, because relojistas' own rows carry no markdown. So a clean
result here is a **no-regression control, not evidence**. The real signal is the opposite
direction: **a drop in the `<item>` count, or any empty `<description>`, means the strip
emptied a live feed.**

**PRE-FIX BASELINE, recorded so the post-roll comparison has something to compare against**
(2026-09-03, before the projection shipped): `HTTP 200, 24,437 bytes, 30 items, 0 descriptions
carrying markdown, 0 empty`, and the `(Fuente: …)` attribution present. ⚠ **30, not 25** — my
plan said 25 from the action's item cap and the live feed serves 30
(`max_items` default). Read the baseline, never a remembered literal.

⚠ **Both controls, because the arm is otherwise untestable on this site.** The extractor must
actually extract (`sed -n 's/.*<description>…/p' | head -2` must print real Spanish prose, not
nothing), and an injected defect must be COUNTED:
`sed 's|<description>|<description>## [x](https://e.com/y) |' feed.xml | …` → 31, i.e. 30 items
plus the channel description. Without those two, "0 markdown" is indistinguishable from a
regex that can never match.

---

## 3. Proving the fix actually ran

### The strip is armed — check, do not assume

```bash
for d in agent-chassis core-manager; do
  kubectl -n ai-persona-system get deploy $d \
    -o jsonpath='{range .spec.template.spec.containers[*].env[*]}{.name}={.value}{"\n"}{end}' \
    | grep -i 'MARKDOWN\|STRIP' || echo "$d: unset — armed"
done
```

### Prove it ran; do not infer it from a clean page

A page can read clean because the feed had no markdown that day.

```bash
kubectl -n ai-persona-system logs deploy/agent-chassis --since=2h | grep 'stripped literal markdown'
```

**Disconfirming: zero lines while dirty rows exist for that site** ⇒ the switch is set, or the
reader executing is not the one you changed.

### ⚠ `build provenance` is a STARTUP line and scrolls

`kubectl logs -l app=agent-chassis --tail=3000 | grep 'build provenance'` returned nothing on
2026-09-03 — that means **"not in range"**, not "unstamped". Fall back to the binary probe with
a present-control and an absent-control, per LANDMINES.

### Prove the lever is ONE lever

Set `DISABLE_NEWS_MARKDOWN_STRIP=1` on one replica, re-render, re-fetch the JSON.
**If it comes back CLEAN, the switch does not reach the JSON producer** and the promise this
change makes to the guardian seat is false.

---

## 4. Tests

```bash
go test ./platform/orchestration/datahelpers/ \
        ./platform/orchestration/actions/queryresolve/ \
        ./platform/orchestration/actions/discovery_checks/
```

⚠ **`./platform/orchestration/actions/` is NOT a clean baseline** — 15 files there are dirty
from other sessions. Scope the run, and say so when reporting a pass.

```bash
scripts/verify-head-builds.sh --with platform/orchestration/datahelpers/literal_markdown.go --test
```

Never hand-roll `git archive HEAD | tar` — that recipe is why this machine runs out of space.
