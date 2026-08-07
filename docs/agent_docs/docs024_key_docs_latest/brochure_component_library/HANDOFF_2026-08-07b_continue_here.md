# HANDOFF 2026-08-07b — Slice A OBSERVED; Slice B now blocked on humans, not on work

**Supersedes `HANDOFF_2026-08-07_continue_here.md`** (same day — that file's open
item 1 is DONE and its ordering is stale; its traps section still applies).
Written ~10:30 BST; liveness claims verified at write time.

## What changed since this morning's handoff

1. **Slice A observed on the acceptance site.** Replan of fundamentallyai fired
   (kcat-safe dispatch, corr `801b0732`, PUBLISH_OK seen; row appeared in 3s, no
   queue), new plan `8ee5807b`, 71/71 emitted section entries persisted, zero
   validate_plan drops, tri-state intact, `pages.sections` still strings. Both
   consumption negatives re-verified WHILE the replan's own builds ran. Verdict
   against RFC_016 §3's three questions: **sane YES, spread YES (trivially),
   complete NO** — object form on 5/24 pages (all newly-composed), 2/9 rostered
   facts assigned (both topically exact), every carried-over page unscoped —
   including index/capabilities/about, which hold the 9 overlap pairs that
   motivated 151. Full read-out: RFC_016 §3a (the Slice B round's entry
   evidence) · lane NOTES 2026-08-07 · dated note in `bugs_open/151`.
2. **The imagery scope_ref defect is filed and it is NOT latent**:
   `bugs_open/214` — 5/131 section-scope refs orphaned fleet-wide, one minted by
   this morning's replan, four (gamesdesign) with paid-for active assets
   unreachable by the build join. 016b §9 + LANDMINES entries added. Fix
   candidate 1 (validate at `flattenImageryBlock`, ~30 lines, no prompt change)
   is unowned and does not need to wait for anything.
3. **The `_HOLD` rename is now actually complete at HEAD** (`24f8ce1e0`). The
   08-06 rename commit `54f36a9ae` shipped only the COPY half — the git-mv
   landmine verbatim — so until this morning `git archive HEAD` carried
   live-named 328/330 seeds. `git ls-tree HEAD` now returns exactly one name
   per number. Disk was always safe (only `_HOLD` names existed there).
4. **The replan queued a real build cascade** (32 items, reconcile). Snapshot at
   ~10:00: 1 needs_page complete, 2 claimed, 4 triaged, 1 failed
   (digital-asset-recovery — `deploy_page`, `CHILD_ORCHESTRATION_FAILED`
   failed_transient: bug 207's class, whose fix is NOT in v1.0.1262; attempt
   1/3, do not cancel, let it retry), 1 needs_human_review
   (ai-readiness-checker-guide — the claims gate found 1 blocker; that is the
   gate working). 6 owned_page_review + 16 needs_imagery + 1 needs_rerender
   queued behind. **Neither failure implicates fact assignments** (the failed
   page's assignments are `[]` and it died at deploy delivery, past content).
   189 was pre-checked before builds ran: zero locked rows on this site.

## Open, in the order I would take them

1. **RFC_016 §5 + §3a need the human.** §5's two decisions (ratify the
   section-entry rule; approve the sliced order) PLUS §3a's new one: require
   object-form for every page (planner prompt change, same seed the round
   already touches — what the motivating case needs) or re-scope the acceptance
   to engaged pages. **Slice B does not move until these are answered.** Until
   then this lane has no unblocked platform work on 151.
2. **When Slice B unblocks**: human/compliance read of the v4 plaintext →
   un-`_HOLD` 328/330 → fresh council round citing RFC_016 §3 + §3a → apply 328
   then 330 → rebuild fundamentallyai's flagged pages → census: overlap pairs
   must fall on ENGAGED pages, five fact-blind sites must not move.
3. **`bugs_open/214` fix candidate 1** — small, self-contained, council-gated Go
   change; any thread can take it. Repair of the 5 bad rows waits for the code
   fix (a cleanup that passes the census once is not a fix; the current prompt
   re-mints orphans).
4. **Monday 08-10**: contact-sheet cron first fire (`~/acceptance_renders/refresh.log`) —
   standing item from 08-05.
5. Optional, cheap: glance at the build cascade's terminal states tomorrow;
   escalate only if digital-asset-recovery exhausts 3 attempts (then it is a
   207 datapoint, not a 151 one).

## Traps (beyond the 08-07 handoff's list, which still applies)

- **A monitor's case patterns must match the column's case.** This morning's
  orchestration watcher matched `completed*` against `COMPLETED` — it never
  exited on success. Harmless as a watcher (it kept reporting, which caught
  the build cascade), fatal as a gate. Test the match against a real row
  before arming.
- **A replan on a site with unbuilt planned pages IS a build dispatch.**
  reconcile files needs_page items and build-dispatch-loop claims them within
  minutes — there is no observe-only replan. Pre-check 189 (locked rows) and
  the claims-gate consequences before firing one anywhere less disposable.
- The morning handoff's traps (hot tree, no blanket `--apply`, live-row prompt
  truth, council practice, `bak_329` rollback) all stand.
