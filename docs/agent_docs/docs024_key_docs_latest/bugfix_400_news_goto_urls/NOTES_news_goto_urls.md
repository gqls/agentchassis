# NOTES — bugfix_400 news goto URLs (append-only, newest at the bottom)

## 2026-09-03 — lane opens; the bug re-verified and INVERTED

Picked up as an unowned bug. Ownership: the bug's own §Who disclaims it (*"This lane consumed it
and is not fixing it"*), only the filing lane's docs mention it, no commit subject addresses it,
no live session named for it. `who-owns.py 400` returns the FILING lane as "owner" — a false
positive of mention-counting, and a good example of why that tool answers "who has talked about
this" rather than "who is fixing it".

**The re-verification changed the bug rather than confirming it**, which is the reason CLAUDE.md
says to check a bug is still valid before fixing it:

- intake **stopped 2026-08-28**, six days, ~1,300 items, zero occurrences — *with* a demand control
- the same sources are still ingesting, so it stopped **upstream**, not by our hand
- the **served** damage is live today (idea.uk 2 of 6, mortgagecalculator 1 of 6, curl'd)
- so the file's three fix candidates, all aimed at intake, would ship and leave 1,378 rows serving

**What I nearly got wrong.** My first read of the census was "1,378 rows, 393 in the last 7 days" —
which reads as a live intake and would have had me fixing the ingest path as the urgent thing. The
newest row being 08-28 is what caught it, and the per-day breakdown with the all-items column is
what proved it. **A total and a recent-count are both compatible with a stopped process**; only the
time series distinguishes them.

**Two instrument errors, both mine, both silent:**
- assumed a `content_feed_sources` table existed to join `source_id` against. It does not — the
  column is bare. The query died loudly, which is the good case.
- used `curl -L` to test whether a stored redirect still resolves, got **403**, and briefly had
  "the backlog is unrecoverable". Wrong: `-L` follows to the publisher, who blocks our agent —
  *after* Google has already returned the correct target. Without `-L` it is a clean 302 with the
  publisher in `Location`. **The weaker request was the correct one**, and the stronger one
  produced a confident wrong answer.

**Established for whoever builds it:** the token is opaque (base64 at every padding yields no
printable substring), so decode is impossible; one hop recovers the URL, 3/3; and
`idx_cfi_dedup` being a partial UNIQUE on `source_url` turns the file's `[UNMEASURED]` duplicate
question into a constraint on the repair.

Nothing committed beyond documentation. Handed off deliberately at the design boundary:
the fix is in council scope and wants its own round.
