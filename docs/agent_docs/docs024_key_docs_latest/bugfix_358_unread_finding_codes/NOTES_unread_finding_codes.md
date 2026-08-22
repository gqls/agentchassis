# NOTES — bugs_open/358: agent_error_log finding codes are write-only and expire unread

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

## 2026-08-22 — session start: ownership + re-validation

- `scripts/who-owns.py 358` → no owning workstream; the `bugfix_238_regeneration_key_loss` lane
  cites it as a cross-reference only, and their NOTES (2026-08-22 entry) say 358 was filed AT
  their request with this class deliberately handed to whoever picks it up. No open
  `site_work_items` row touches the class. This lane now owns it; this directory is the record.
- **Census re-run against the live DB (see RUNBOOK), two deltas since the bug file was written
  this morning — neither invalidates the class:**
  1. `resolved` is now **48**, not 0. All 48 set today 10:40 UTC by
     `resolved_by = 'content-loss-check:healed'` / `'content-loss-check:row_gone'` on
     `CONTENT_KEY_LOSS` (40) and `STRUCTURAL_KEY_CARRY_MISS` (8) — the 238 lane's consumer
     shipped (`cba51ad1d`) and ran. So "the resolved workflow has never been used once" is now
     stale BY ONE DAY, and the first user is exactly the reader-ships-with-writer pattern 358
     holds up as the positive example. The §8 trap stands: resolving halves remaining life to
     14 days; content-loss-check extracts (heals) before resolving.
  2. New code `CONTENT_KEY_LOSS` (72 rows, all 2026-08-22, agent_type `content-loss-check`) —
     written AND consumed by the same binary. Not a new member of the unread class.
- Commit `0ce242d9c` (today) added a THIRD recorder in the validation-gate family:
  `CONTENT_VALIDATION_WARNING_DETAIL`. 0 rows yet (inert until image roll). Reader status being
  verified — if unread, the class grew by one TODAY, while the bug file was being written,
  which is §3's self-sustaining mechanism observed live.
- Headline counts still hold: 45,507 total (was 45,426), `RESOLVER_CONFLICTING_CANDIDATES`
  9,617 (was 9,615) and still the loudest, oldest row still 2026-07-23 (retention live),
  `REVIEW_SUPERSEDED_BY_PASSING_SAVE` (25 rows, 07-23 only) days from deletion,
  `TRUNCATION_DEGRADED_REVIEW` dies ~08-25.
