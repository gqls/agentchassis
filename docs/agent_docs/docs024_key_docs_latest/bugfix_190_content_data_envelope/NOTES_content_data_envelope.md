# NOTES — bugs_open/190, content_data stores the raw LLM transport envelope

Running record. Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — session opens: picking the bug, and why this one

Task was "take the next bug in `bugs_open/` that no other thread is working". The picking
itself was most of the first hour, and two of the three methods I tried were bad. Recording
all three because the bad ones look reasonable:

1. **`git log --grep=<number>` over 4 days.** Useless on its own — a bare number matches
   dates, image tags (`v1.0.1181`), migration numbers and other bugs' cross-references. It
   told me `122` had 50 "commits" and `126` had 0; neither figure meant what it looked like.
2. **Grepping live session transcripts for `bugs_open/NNN_`.** Also bad, and worse because
   it looks precise: every session that runs `ls bugs_open/` has the entire bug list in its
   transcript, so *every* bug scored 2–6 "live sessions". A signal that fires for everything
   is not a signal.
3. **What actually worked** — grep the transcripts for the bug file appearing as a
   `"file_path"` in a tool call, i.e. sessions that actually *opened* it. That separated
   ~11 unclaimed bugs from ~35 claimed ones.

Then a piece of luck worth stealing: session `62bc0e8b` was itself running a session→bug
census and had left the output in its transcript — a table of every live session against the
bugs it mentions most. That is a better instrument than anything I built, and it immediately
showed `181` (which I had provisionally picked, and which is *labelled* "OPEN, UNOWNED" in
its own header) was in fact being worked by two live sessions, one of them the `163` lane
with 34 mentions.

> **MISSTEP, and the useful kind.** I had already half-committed to `181` on the strength of
> the words "OPEN, UNOWNED" written in the file on 2026-08-02. **A file's own ownership
> claim is a snapshot of the moment it was typed**, and this tree turns over fast enough
> that two days is long. `who-owns.py` is documented as lagging; what I had not internalised
> is that the *bug file itself* lags in exactly the same way and is more persuasive because
> it is phrased as a fact.

Also checked and rejected, each for a stated reason rather than a vibe:

- **`146` (ported tool pages outside every acceptance tier)** — unclaimed, and the
  structural half is **already fixed**. `discovery_checks/tool_eligibility.go` exists,
  shares one predicate between Tier 2 and Tier 4, and handles ported pages by keying on the
  page rather than `cc.function`. It shipped 2026-07-29 under commit `ac9f75a0c` crediting
  **`bugs_open/084` candidate 3**, not `146` — which is why `146` still reads as open. Worth
  a close-out by someone; not a fix task.
- **`093` (stat audit has one guarded call site)** — the code fix is built and live since
  `v1.0.1172`, and the file's own final triage says it is now blocked on `bugs_open/083`
  (the check has no scheduler). `083` is being actively worked by another session. Nothing
  for me to do that would not collide.
- **`096`, `126`** — both end in an owner/architecture call, not a code change.

Landed on **`190`**: unclaimed by every instrument above, framework-shaped, and its fix
candidate (1) is explicitly a make-the-bad-state-unrepresentable change.

## 2026-08-04 — re-validating 190, and finding two errors in it

The bug is **still live**: 2 envelope rows. I took the denominator in the same query as the
numerator, which this bug file itself recommends (its filer had been burned by an
empty-population count):

```sql
SELECT count(*) FILTER (WHERE content_data ? 'type' AND content_data ? 'result'
                          AND content_data->>'type'='text') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS denom_nonnull,
       count(*) AS denom_all
FROM page_components;   --  2 | 1054 | 1207
```

`site_components`: `0 | 54 | 54`. The bug never checked the sibling table. Clean today, but
it is a `content_data` store and belongs in the guard's scope.

**Error 1 in the filed bug — the row identity moved.** The file names
`25c73a1c-…` for gaswholesalers `how-pricing-works`. That id does not exist. The row serving
that page is `17e7739e-…`, `created_at = updated_at = 2026-08-03 22:35:17` — *after* the bug
was filed at ~21:30Z the same evening.

**Error 2 — the file's own verification recipe would miss half the population.** It says the
guard "must key on the exact two-key shape". The two live rows do not share a key set:

| row | top-level keys |
|---|---|
| `d2e9644b` finetuning | `{content, result, type}` — **three** |
| `17e7739e` gaswholesalers | `{result, type}` — two |

An "exactly two keys" predicate is silent on `d2e9644b` — the very row the file says is
fully repairable. The discriminator has to be the envelope *signature* (`type == 'text'` and
a string `result`), and `d2e9644b` is the interesting case because real content and envelope
keys coexist in one map.

## 2026-08-04 — the misstep I caught before it reached the bug file's conclusions

`page_component_history` has **65** envelope-shaped rows, all `source =
'save_page_sections_overwrite'`, latest 2026-08-03 22:35:17. My first reading was "the seam
has written 65 envelopes" and I very nearly wrote that down as the headline.

**It is not a write count.** The history INSERT
(`save_page_sections_action.go:586-601`) is `SELECT pc.content_data … WHERE pc.page_id = $1`
executed *before* the DELETE — it archives the state being **replaced**. So 65 counts
*overwrite events on pages that already carried an envelope*.

The cheap check that settled it was reading the SQL in the action, not querying harder. A
data question that is actually a code question will happily return a confident number.

Two things fell out of doing it properly:

- **Blast radius is much larger than the filed "2 rows"**: those 65 events span **25 distinct
  pages across 6 sites**. ⚠ `count(DISTINCT component_id)` returns **0** on the same query —
  a NULL trap, not an absence: the FK is `ON DELETE SET NULL`, so archived rows whose
  component was later deleted carry `component_id = NULL`. Group by `page_id`. Same shape as
  the `distinct_content = 0` trap already in `LANDMINES.md`.
- **Generation has stopped; propagation has not.** Exactly ONE envelope event since
  2026-07-18 (the 08-03 one). The three-tier parse fix worked — no new envelope has been
  minted since mid-July. What the 08-03 event shows is the save seam archiving an envelope
  and 84ms later creating a *new row carrying the same envelope forward*. The live defect is
  that **`save_page_sections` re-persists a transport envelope it is handed, indefinitely**,
  so a poisoned row survives every rebuild instead of being cleaned or refused by it.

That reframing matters for the fix: the guard's job is to stop propagation, not to stop
minting. It also means the urgency is lower than "a writer is actively producing poison" —
and I had written exactly that sentence into the bug file before correcting it there too,
visibly, rather than editing it away.

## 2026-08-04 — constraint noted before any code is written

`save_page_sections_action.go` is **dirty in the shared tree from another session** (the
`bugs_open/194` lane: +62/-2, adding `resolveSectionsMetadataField` and a
`require_sections_metadata` refusal). A pathspec commit still takes a **same-file**
passenger. So the fix must keep its footprint in that file to the smallest possible
call-site edit, with the logic in a new file in the same package — which is exactly the
precedent the 194 lane itself set with `save_sections_metadata_source.go`.
