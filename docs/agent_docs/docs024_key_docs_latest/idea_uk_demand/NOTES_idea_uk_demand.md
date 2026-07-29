# NOTES — idea.uk demand (append-only, newest at the bottom)

## §D.1 — day one: the front door is labelled "for sale" (2026-07-29)

Evidence for the four findings in the PLAN, with the commands.

**Google's index entry is the old parking page.** WebSearch `site:idea.uk`
returned `"idea.uk - Domain Name For Sale | Dan.com"` as the entry for
`https://idea.uk/` (2026-07-29, ~09:00 UTC). The domain's pre-us life. Until
that snippet is replaced, organic search actively repels buyers.

**Real Googlebot has never seen the money page.** UA strings are useless here —
of 12 distinct "Googlebot"-UA IPs in the last fortnight, exactly one reverse-
resolves to `googlebot.com` (66.249.65.197); the rest are vulnerability
scanners (digivps, GCP VMs, `osxwrong.us.com`) probing `/dump.sql`,
`/.env.bak`, `/wp-config.php.bak`. Filtering to the genuine 66.249.0.0/16
range: **253 hits in the whole log (05 Jun → 29 Jul)**, distribution:

```
107 /robots.txt   (404 every time)
100 /
  7 /favicon.ico  (404)
  ~1 each: /tools.html /tools /terms /refund-policy /privacy /index.html
  0  /report.html          ← the page that takes money
  0  /guides/anything      ← the content that could rank
```

Last genuine crawl: 23–24 Jul, homepage + assets only. `[INFERRED]` the crawler
has no discovery path to the inner pages: no sitemap, robots 404, and the
homepage is the only URL it trusts from the parked-domain era.

**Referrer data is spam.** Top "referrers" include
`google.com/search?q=site:example.com` (×2,020), `123deliverit.com`,
`binance.com` — classic referrer-spam signatures on a bare-IP box. Do not
quote any referrer-based number from this log without filtering.

**Meta descriptions.** 7 of 21 active pages empty (index, about, contact,
guides-index, news-index, privacy, tools) — query in RUNBOOK. The homepage —
the ONLY page real Googlebot crawls — serves `<meta name="description"
content="">` today.

**Mechanism note for any file we want on the box:** `/usr/local/bin/sitesync`
does `git reset --hard origin/main` + `rsync -a --delete idea.uk/
/var/www/idea.uk/` from `gqls/vm-sites` every 5 minutes. Anything not in that
repo is deleted. robots.txt/sitemap.xml therefore go through the repo, never
scp'd onto the box.

**Renderer fact that de-risked p4_33:** `rerender_single_page_action.go:357`
loads `COALESCE(p.meta_description,'')` fresh and builds the `<head>` from it
(line ~803), so an assemble-mode rerender applies a meta_description change
while carrying stored section HTML verbatim — no `section_data_resolved`
needed, no LLM-escalation exposure on pages with derived/NULL content_data.

**Not touched, deliberately:** favicon 404s (owned by the og-card lane,
`bugs_open/131`, blocked on S3 creds — do not collide); the stale
phantom_internal_link detections on index/about (another detector's, hrefs
without `.html` suggest an old vocabulary — not this lane's).
