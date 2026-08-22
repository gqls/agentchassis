# PLAN — bugs_open/316, ordering fairness in the news-feed trigger

Design, phasing, decisions **and their reasons**. Corrections live here, marked as corrections.

---

## Status: 2026-08-22, evidence complete, plan being drafted

Everything in this section is settled and measured; see `NOTES_316_news_feed_ordering.md` for the
queries and `RUNBOOK_316_news_feed_ordering.md` for how to re-run them.

### The two defects are separable and only one is mine to fix

1. **Unfairness — a defect.** `content-feed-trigger.find_news_sites` ends `ORDER BY s.domain LIMIT 5`.
   Fix by ordering on the schedule. Config-only, live on apply, no image roll.
2. **Capacity — an owner spend decision, NOT a defect fix.** 42 fetches/day demanded against 20 supplied
   (2.10x); removing the cap entirely still leaves 36 vs 42 (1.17x). [MEASURED 2026-08-22, reproduces the
   filed figure exactly.] **Out of scope. Present the arithmetic; change nothing.**

### Decisions taken, with reasons

- **Order on the due-time, not on `random()`.** `random()` is what the sibling `model-directory-trigger`
  uses and the bug file calls it "the cheap version of the same idea". It is cheaper and it is worse:
  it makes starvation unbiased rather than absent, and a site can still lose several draws running. The
  platform's own Go layer already orders this exact kind of work by `next_fetch_at ASC NULLS FIRST` in
  two places, so due-ordering is the **existing convention**, applied to the one layer that skipped it.
- **The fix must answer the `NOT EXISTS(active sources)` arm explicitly.** The bug file's fix candidate 1
  (`ORDER BY min_next_fetch_at NULLS FIRST`) taken literally creates a permanent head-of-queue squatter
  — see NOTES. This is a **correction to the originating bug file's proposed remedy**, recorded here
  rather than silently worked around.
- **Prefer a framework-level answer over the one-line fix, but size it against the evidence.** The
  dangerous shape has exactly **one** live member today. The argument for building a detector anyway is
  `query_row_cap.go`'s own, in its header: *"one function, and the whole class becomes visible at once,
  including caps nobody has written yet."* The argument against is over-engineering for n=1. This is the
  central open design question the plan must settle, and it must be settled with reasons, not taste.
- **Reuse `cmd/config-key-audit`, do not build a new service shape.** It already hosts ~10 live-config
  audits behind flags, already has a direct-Postgres route for CronJobs (`fleetdb.go`), and already
  traverses with `validation.WalkSteps`. A new check that is a new binary would be new machinery for a
  question the existing binary is built to ask.
- **Sequencing is a correctness property, not housekeeping.** Any detector must be shown FIRING on the
  motivating case before the migration lands. Once the migration applies, the live row can no longer
  supply the positive control — which is why the pre-fix query text is already captured verbatim to
  `PREFIX_find_news_sites_query_2026-08-22.sql`.

### Open at the time of writing
- The exact corrected SQL and its NULL semantics.
- Whether to build the detector, and at what size.
- Council submission.
