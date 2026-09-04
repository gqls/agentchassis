# HANDOFF 2026-09-04 — bugs_open/332 feed display markdown — CONTINUE HERE

**Read this first, then `NOTES` (last four sections) and `RUNBOOK` §2–3.**
Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_332_feed_display_markdown/`

---

> ## ⚠ CORRECTED 2026-09-04 16:02Z — READ THIS BEFORE THE REST
>
> **Everything on this lane is now LIVE, including today's residual fix.** The roll completed
> at **16:00–16:01Z**; `service_binary_capabilities` shows three fresh `agent-chassis` pods on
> `06c0b18f2`, which contains `adef5d481`. **Item 1 of §3 and decision 1 of §5 are DISCHARGED
> — do not ask for another roll.**
>
> **And I had the earlier roll wrong, in my reasoning rather than my conclusion.** I wrote that
> the v1.0.1360 build commit `239ab3626` was dated 22:37 and therefore *later* than the
> 22:06:58Z pod start, so the tag could not settle what was running. That was a **timezone
> conflation**: `git log --date=format:` printed **BST**, `kubectl` prints **UTC**.
> `239ab3626` is `22:37:21+01:00` = **21:37 UTC**, thirty minutes BEFORE the pods started. The
> ancestry check settles it cleanly and always did. Caught by the `inter thread comms` lane.
> My conclusion (the fix is live) was right and independently proven by dartsonline; the
> **path** to it was wrong, and a wrong path published as reasoning is worth correcting.
>
> **Use the table, not the log or the binary** — CLAUDE.md gained this today (BLD-023):
> ```sql
> SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
>  WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
> ```
> then `git merge-base --is-ancestor <your-commit> <the stamp>`. ⚠ It is a **two-hour window**,
> so it answers *what is running now* and cannot date anything older.
>
> **What is still owed is only the verification**, and it cannot be run yet: at 16:01:44Z **no
> news component had re-rendered since the roll**, so the truncated-image fix is live but
> **unexercised**. Re-run §4 once `rerendered_since_roll=1` rows exist — hours, not minutes.

## 1. One paragraph of state

The fix **is live and it works** — proven at the artefact this morning, not inferred. A residual
gap and a second, worse defect (a detection pattern that was never wired) were found today by
verifying after the roll — both are now **fixed, committed and LIVE** (`06c0b18f2`, 16:00–16:01Z),
but **unexercised**, so the only code-side work left is re-running §4 once pages re-render.
Migration 758 is council-approved and deliberately **held** for a human. Two spin-off bugs are
filed and owned elsewhere.

---

## 2. What is PROVEN, and how — do not re-derive these

### 2.1 The fix is live and working ✅

The chassis rolled 2026-09-03 **22:06:58Z** (pods `agent-chassis-ffc9ddff9-*`, image
`v1.0.1360`).

**The demand control that proves it, and why the obvious check would not have:**

| site | feed column (7d) | re-rendered | page |
|---|---|---|---|
| **dartsonline.com** | **dirty** — 2 tail-links | **10:40Z today**, post-roll | **CLEAN** ✅ |
| gaswholesalers.com | clean | post-roll | clean — *proves nothing* |
| relojistas.com | clean | post-roll | clean — *proves nothing* |
| webdesign.co.uk | clean | post-roll | clean — *proves nothing* |

**dartsonline is the only discriminating case**: dirty column + post-roll re-render + clean page.
The other three are the third row of the RUNBOOK's three-way table — a clean page over a clean
column is *no evidence at all*, and reading them as success is the trap this lane keeps hitting.

`boxingonline.ugg2.com/news/index.html` has gone **5 → 0** occurrences.

### 2.2 ⚠ Two defects found TODAY — fixed, and LIVE since 16:00Z

Both are in `adef5d481`, which is in `06c0b18f2`, running on all three chassis pods since
16:00–16:01Z. **Live but UNEXERCISED**: no news component had re-rendered since the roll as of
16:01:44Z, so neither has yet been demonstrated on a served page.

**(a) A truncated IMAGE survived.** `![alt](url…` — alt text closed, URL severed — fell through
every rule: `mdImageStripRe` needs the closing paren, `mdLinkTruncatedStripRe`'s left boundary
`(^|[\s(])` **rejects the preceding `!`** (which is exactly the image marker), and
`mdFeedImageTailRe` requires no `]`. The value passed through untouched. Live on idea.uk,
re-rendered **10:49Z today** — hours after the roll, which is what made it a gap rather than a
stale binary.

**(b) ⚠⚠ `MDLinkTruncatedRe` WAS NEVER CALLED.** Declared, exported, documented across fifteen
lines, and cited to the council as *"detection AND strip single-sourced"*. `LiteralMarkdownPatterns`
never referenced it. **The scan was blind to truncated links** — the exact defect this bug is
about — so a page serving one scanned clean.

> **This is the most important thing in this handoff for anyone touching that file.** Every
> property test routes through `LiteralMarkdownPatterns`, so a declared-but-uncalled pattern is
> invisible to all of them, **and `TestStripThenScanFindsNothing` passed VACUOUSLY** — a
> fixpoint holds trivially when `Scan` cannot see the pattern. The test whose whole job is to
> stop detection and strip drifting could not see the drift, because the drift was *absence*.
> **A round-trip property cannot detect a missing arm.** Pair every one with a reachability
> assertion. → `WRONG_CALLS.md`, 2026-09-04.

Gate A re-run after wiring detection: **zero co-firing rows, control 128**. Safe.

---

## 3. WHAT IS LEFT — the whole list

| # | Item | Blocked on | Who |
|---|---|---|---|
| ~~1~~ | ~~**Roll a chassis build** carrying `adef5d481`~~ **DONE 2026-09-04 16:00–16:01Z**, pods on `06c0b18f2` | — | — |
| 2 | **Re-verify at the artefact** (§4) — the truncated-image fix is live but UNEXERCISED; wait for `rerendered_since_roll=1` | the feed cycle (hours) | next session |
| ~~3a~~ | ~~**Apply migration 758**~~ **DONE 2026-09-04 16:28:35Z**, verified at the rows; preconditions re-checked (2 rows, 0 forked); scripts balanced | — | — |
| 3b | **Look at a news page in a browser** — the published assets republish on each site's NEXT RENDER, so this cannot be done until then. The failure mode is an EMPTY list (`container.innerHTML = ""` runs before the loop) | a site render | a human |
| 4 | `bugs_open/473` — scraped nav as article text | nothing; unowned by choice | feed lane's charter |
| 5 | Close 332 | 2 | next session |

**Nothing else is owed.** Council `803f0d81` **APPROVED** (332 Go work, 4 advisory objections all
actioned). Council `17a61f16` **APPROVED** at round 3 (migration 758).

---

## 4. THE EXACT CLOSING CHECK — run once pages have re-rendered since 16:01Z

Full commands with their gotchas are in `RUNBOOK` §2. The short form, and **all three parts are
required**:

**(a) The five hosts** — expect `0 0 0` on each. Per-host 404 control mandatory; one file per
host (a shared temp file makes a failed fetch report the previous host's numbers).
Today's reading: boxingonline **0**, fundamentallyai 3, robot-hands 2, ai-agent-orchestration
2+1 img, idea.uk 1+1 img.

**(b) The demand control — without it (a) means nothing.** A page reads clean when that day's
feed carried no markdown. Pair every clean page with a **dirty column** for the same site:

```sql
SELECT s.domain,
       count(*) FILTER (WHERE cfi.source_summary ~ '\]\([^)]*$') AS col_tail_link,
       count(*) FILTER (WHERE cfi.source_summary ~ '!\[')        AS col_img
  FROM content_feed_items cfi JOIN sites s ON s.id=cfi.site_id
 WHERE cfi.created_at > now() - interval '7 days' GROUP BY 1 ORDER BY 2 DESC;
```

**(c) The falsifier, which is the one that actually decides it.** A page still dirty is only a
failure **if it re-rendered after the roll**:

```sql
SELECT s.domain, pc.slot_name, pc.updated_at::timestamp(0),
       (pc.updated_at > '<POD START>')::int AS rerendered_since_roll,
       (COALESCE(pc.rendered_html,'') ~ '\]\((https?://|/)[^)" ]{0,200}\.\.\.')::int AS still_dirty
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.component_id IN (SELECT id FROM content_components WHERE function IN ('news-listing','latest-news'))
 ORDER BY still_dirty DESC, pc.updated_at DESC;
```

**`rerendered_since_roll=1 AND still_dirty=1` is the failure.** That row is what found today's
gap. `still_dirty=1` with `rerendered_since_roll=0` is just a page waiting its turn.

**Also re-check the JSON**, which is where most of the damage was and which the old sweep could
not see: `/data/news-archive.json` and `/data/latest-news.json`, matching `": *"` **not** `":"`
(it is `MarshalIndent` output, so a pattern without the space scores zero on every file for
ever). `sweep_site_defects.sh` §1.4 now does this automatically.

---

## 5. DECISIONS THE OWNER NEEDS TO MAKE

1. ~~**Roll the chassis again?**~~ **DISCHARGED 2026-09-04 16:00–16:01Z.** Today's fixes went
   out in `06c0b18f2` on a roll another lane was already running. **No decision needed.** The
   only thing left on the code side is waiting for pages to re-render and then re-running §4.
2. ~~**Apply migration 758**~~ **APPLIED 2026-09-04 16:28:35Z.** What remains is **who looks at
   a browser, and when.** The published assets republish on each site's next render, so this
   cannot be done yet. When it can: `selector_count article.news-list-item` > 0 proves the
   script ran and built the list; `selector_exists a.news-more-link[href^="/"]` proves the
   internal link resolved — **the one behaviour nothing else in the chain can see**, and the
   regression that nearly shipped. ⚠ Those checks pass on count > 0 and **fail on zero, with no
   expect-zero form**, so every assertion must be the presence of the RIGHT state. If the list
   comes up empty, run `758_..._ROLLBACK.sql`.
3. **`bugs_open/473` — scraped navigation** (`Tennis`, `NFL`, `MLB` appearing as article
   summaries). The feed lane owns it and has explicitly **not** taken it, on the reasoning that
   a word-list nav filter would be worse than the open bug. **The news pages will still read a
   little oddly until this is done**, and no amount of markdown work will change that. Does he
   want it prioritised, or left filed?

---

## 6. TRAPS — read before touching anything here

- **A clean served surface expires the moment the projection rolls.** It then means *the strip
  ran*, nothing more: a dirty table and a clean table produce byte-identical output. Two
  questions, two instruments — *is the visitor seeing junk?* → the surface; *is ingestion
  clean?* → the column, never the surface. (The `news_feed_ingestion` lane found this on their
  own migration-746 plan.)
- **`grep -c 'stripped literal markdown'` matches SEVEN call sites**, only one of which is this
  projection. Use `queryresolve: stripped literal markdown`, and `-l app=agent-chassis` (there
  are 2 pods; `logs deploy/…` reads one).
- **A `(?m)^` pattern means BLOCK start on `rendered_html` and LINE start on `content_data`.**
  This is why list markers, bracket tails and bold tails are feed-display-strip-only and enter
  no detector. → LANDMINES.
- **The served news page is OVERWRITTEN in the browser** by `/data/news-archive.json`. A curl
  reads a document the visitor never sees in that state. → LANDMINES.
- **`grep -c` counts strings, not elements.** Six matches for `news-more-link` on three
  homepages are all **CSS rules**; there are zero rendered anchors.
- Probe the slug on `boxingonline.ugg2.com`, never `boxingonline.com` (parked, 200s everything).
- Postgres caps regex repetition at **255** — `{0,300}` errors as "invalid", which reads as a
  syntax error and is a limit.

---

## 7. Where everything is

| what | where |
|---|---|
| the bug | `bugs_open/332_HANDOFF_2026-08-19_…latent.md` — read the **2026-09-03 addendum** at the bottom |
| the code | `platform/orchestration/actions/queryresolve/feed_display_text.go` (the projection), `platform/orchestration/datahelpers/literal_markdown.go` (the patterns) |
| the migration | `docs/agent_docs/sql_for_agents/758_…_HOLD.sql` + its `_ROLLBACK` |
| spin-offs | `bugs_open/472` (unescaped innerHTML), `bugs_open/473` (scraped nav) |
| the register | CQ-019 in `docs026_concept_register/register/content-quality.md` |
| my mistakes | `WRONG_CALLS.md`, eight entries dated 09-03/09-04 |
| the traps | `LANDMINES.md`, two entries; `016b` §9, two patterns |
