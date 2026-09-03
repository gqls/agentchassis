# HANDOFF 2026-09-03 — bugs_open/400, news pipeline serves Google redirect URLs

**START HERE.** Diagnosis is done and is in `bugs_open/400_…md` §2026-09-03 (read that first —
it carries every measurement with its query). **No code is written. Nothing is committed.**

## State in one paragraph

The `news_search` pipeline stored **1,378** article links in the form
`https://www.google.com/goto?url=<opaque>` instead of the publisher's own URL, across **11 sites**.
When filed (2026-08-25) this was a live intake. It is **not any more** — the shape stopped arriving
on **2026-08-28** and has been absent for six days across ~1,300 new items. **But the stored rows
are still SERVED today**, and nothing prevents the intake resuming.

## The three things that decide what you build

1. **The intake stopped UPSTREAM, not because of us.** The same `source_id`s are still ingesting
   (6–44 items each since it stopped, `last_any` 2026-09-02) with zero goto rows, and nothing in
   this tree changed. So it is the provider's response shape. **It can resume silently and we have
   no detector** — that is the strongest reason to build anything at all.
2. **"Decode the token" is impossible** — verified, it is an opaque blob, not an encoded URL. The
   bug file offers decode-or-follow; only follow exists.
3. **One hop recovers the real URL, and the backlog is recoverable.** A single non-following
   request returns **302 + `Location` = the publisher** (3/3 tested: hpcwire, fortune, nature).
   ⚠ **Do not follow to the final page** — it 403s (the publisher blocks our agent) *after* the
   target was already captured, so a fix requiring a 200 discards recoverable rows.

## Suggested order of work

**A. The detector first.** It is the only piece whose value does not depend on the intake resuming,
and it is what makes a silent resumption visible. Daily-check convention: count goto-form rows in
the window, fail on non-zero, **with a demand control** (total new feed rows > 0) so a dead feed
cannot read as a clean run. That control is not optional here — this bug's whole re-verification
turned on it.

**B. The unwrap, at the bridge** (`feed_normalize_action.go`), not the provider. Upstream of
`source_url` and therefore of dedup, and catches every provider rather than scrapingbee only. On
resolve failure **keep the goto URL** (a working link beats no item) and count the failure.

**C. The backlog repair, last, and measure before you write.**
⚠ `idx_cfi_dedup` is a partial UNIQUE index on `source_url`. Repairing a row can collide with an
existing direct-URL row for the same story — which is the bug file's own `[UNMEASURED]` duplicate
question arriving as a constraint. **Measure the collision set before the UPDATE**, then decide
merge-vs-skip. Do not let the UPDATE discover it.

## Verification recipes that already work (copy these)

```sql
-- the demand-controlled census. A zero without the middle column is worthless.
SELECT created_at::date, count(*) AS all_new_items,
       count(*) FILTER (WHERE source_url LIKE 'https://www.google.com/goto%') AS goto
FROM content_feed_items WHERE created_at > now() - interval '12 days'
GROUP BY 1 ORDER BY 1 DESC;

-- is it the sources going quiet, or the shape stopping? (it was the shape)
SELECT source_id,
       count(*) FILTER (WHERE source_url LIKE '%google.com/goto%') AS goto_rows,
       count(*) FILTER (WHERE created_at > now() - interval '6 days') AS items_since,
       max(created_at)::date AS last_any
FROM content_feed_items GROUP BY 1
HAVING count(*) FILTER (WHERE source_url LIKE '%google.com/goto%') > 0
ORDER BY goto_rows DESC;
```

```bash
# the fix primitive: ONE hop, no body, no publisher fetch
curl -s -o /dev/null -w "%{http_code} %{redirect_url}\n" --max-time 20 -A "Mozilla/5.0" "<goto url>"
# expect: 302 <publisher url>.  NEVER add -L.

# served damage, at the artefact
curl -s https://idea.uk/data/latest-news.json | grep -c 'google\.com/goto'   # 2 of 6 on 09-03
```

## Traps banked while doing this

- **A zero needs a demand control.** "No goto rows for six days" and "the feed died" produce the
  same number. The control is what made this a re-framing rather than a false close.
- **`content_feed_sources` does not exist** — the table is `content_feed_items` with a bare
  `source_id` column and no join target. Two of my queries died on that assumption.
- **`-L` inverts the answer.** Following gives 403 and looks like an unrecoverable link; not
  following gives 302 and the publisher. The weaker request is the correct one.

## Scope and process notes

- **In council scope** (`platform/orchestration/actions/`, `internal/adapters/websearch/`) — unlike
  361 and 366 this one *does* want a council round. `DRY_RUN=1 097_TRIGGER…` tests admission free.
- Adjacent: `bugs_open/316` (news feed cap) and `scripts/initial_messages/100_news_feed_ingester/`.
  The filing lane (`idea_uk_vm_site`) explicitly is not fixing it and should be told when it lands.
- Severity is genuinely low-medium: the links *work*, via a tracking hop that names no publisher, on
  sites whose pitch includes source honesty. Do not let the 1,378 make it sound like an outage.
