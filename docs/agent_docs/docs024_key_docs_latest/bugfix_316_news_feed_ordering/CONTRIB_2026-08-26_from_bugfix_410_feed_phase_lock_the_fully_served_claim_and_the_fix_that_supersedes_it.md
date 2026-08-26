# CONTRIB from the `bugfix_410_feed_phase_lock` lane, 2026-08-26 — your "fully served" claim, and what shipped about it

*(Contribution into this lane's record per the who-owns convention — no active owner found
(last 316 commit 2026-08-22, confirmed with the idea_uk_vm_site lane by message). Nothing in
your docs is edited; this file only adds.)*

**README:138's "the seven 6-hourly sites are now fully served" was refuted on 2026-08-26** —
not because 556's capacity fix was wrong (it wasn't; the cap genuinely stopped binding) but
because a SECOND, independent mechanism kept the same sites at half cadence:
`next_fetch_at` is stamped `NOW() + fetch_interval` at FETCH time, so a 6 h interval on the
6 h trigger falls due seconds after the next pass fires. Full mechanism + census:
`bugs_open/410_HANDOFF_2026-08-26_next_fetch_at_stamped_at_fetch_time_phase_locks_every_six_hour_news_site_to_a_twelve_hour_cadence.md`
(filed by the idea_uk_vm_site lane; ⚠ the number 410 also names an unrelated scan-loss case —
resolve by slug).

**The fix shipped 2026-08-26 (commit `201236b2a`, `Council-Submitted:
04c657d2-cbee-4528-b124-b53a747d2e96`): a half-cadence due look-ahead** in both layers — a
shared Go predicate in `dispatch_feed_sources` (rides the next chassis roll) and migration
`653_content_feed_due_lookahead_HOLD.sql` for `find_news_sites` (hand-applied only after the
roll). It deliberately does NOT touch your 554 ordering or your 556 caps.

**Two things your lane's records should absorb:**

1. **Post-fix, cap hits become the NORM, by design.** ~12 news sites will be due every pass
   against your caps of 10/10; 554's due_at fair ordering makes the overflow rotate (average
   ≈ 7.2 h against the 6 h label). LCO-009 and `--capped-schedule-ordering` reporting hits on
   most passes is the demand 556 was sized against finally arriving, not a regression. If
   the owner wants the full 6 h for all 12, the next capacity step is caps 12/12 — that is
   an owner cost decision, flagged in the council submission, not taken by either lane.
2. **Your capacity read-out's "20 site-refreshes/day against 42 demanded" arithmetic changes
   shape once the look-ahead is live** — demand presented per pass roughly doubles (every
   6 h-only site, every pass), so any future capacity census should be re-derived from
   post-fix passes, not carried forward.

Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_410_feed_phase_lock/`
(PLAN has the full decision record; RUNBOOK has the deploy sequencing and acceptance queries).
