# HANDOFF 2026-08-09 — bugfix 226 chrome divergence guard: continue here

**For the next session picking up this lane. Read PLAN (design + blast radius)
and NOTES (evidence trail) beside this; this file is only what to DO next.**

## State in one paragraph

The mechanism is SHIPPED and LIVE end to end: mig 344 (archive trigger,
fail-closed, sole trigger on `site_components`, probe-verified) has been live
since 08-08 and has already archived four real production overwrites
(webdesign.uk + leopardessconsulting.co.uk chrome, all `unstamped`, zero
errors); the Go half (digest stamp + classify + `chrome_divergence_overwritten`
item + ledger read-back fallback) is on **v1.0.1270**, binary-verified on both
main `agent-chassis` replicas with positive AND negative greps. Commits:
`1eae32644` (round 1), `6bee2708e` (round 2), plus the round-3 commit that
carries this file. Council trail `cffbfec4-3bec-4577-8844-d17c546ded3e`:
round 1 REVISE (guardian) answered, round 2 REVISE (bug_historian) answered,
**round 3 pending** (run orch `50924d69-73a9-4af9-9649-4c3bcd74aa8f`,
submitted ~09:XX UTC 08-09).

## Do next, in order

1. ~~**Read the round-3 verdict.**~~ **DONE 2026-08-09 09:08Z: APPROVED**,
   round 3 of the trail — "approved with 2 advisory objection(s), none
   high-severity", 5 abstained, every seat including bug_historian's verdict
   recorded (its page-side advisory stands as the 229 OWNER CALL). The two
   advisories are carried in STY-054's open-review (d) and (e). The final
   docs commit carries `Council-Reviewed: cffbfec4-…` legitimately (verdict
   read before writing the trailer).
2. ~~**Run close criterion 2 (end-to-end protocol)**~~ **DONE 2026-08-09
   ~09:25Z on dartsonline.com — every required signal observed, by row
   identity.** The wave itself had stamped dartsonline at 09:08:30Z (step (a)
   free); psql patch drew a `machine_made`/`psql` ledger row; forced rebuild
   (orch `322b266e`) fired the WARN once (captured live on pod zhz2g),
   archived the patched bytes `hand_patched` (md5-identical), filed the
   digest-keyed item; negative control (orch `453b2eb6`) rewrote all 3 slots
   byte-identical with no row and no item. Probe item cancelled with a note.
   Evidence + the corrected timeline in the bug file's CLOSE CRITERIA and
   NOTES session 3.
3. **Watch the 117 wave's first pass (close criterion 3) — IN PROGRESS, the
   wave HAS started.** 3 sites done (leopardess + webdesign.co.uk pre-roll
   unstamped; dartsonline post-roll 3/3 stamped, byte-identical, no archive
   rows). 3/57 stamped, zero trigger errors, accumulation 1:1. Cadence is
   discovery-paced (one site per ~1h-overnight). Re-check: stamped count
   climbing, `agent_error_log` for `site_component_history` mentions (0 so
   far), and the guardian ratio (items per site+slot vs distinct digests).
4. **Then the bug is done in substance** — per the owner 08-06 ruling it
   STAYS in `bugs_open/` (do not move to bugs_closed); update its header and
   the MEMORY topic file `bugfix-226-chrome-divergence.md` + the
   MEMORY_workstreams line instead.

## Standing landmines for whoever touches this

- The archive is a **DB trigger** — no Go grep shows it (LANDMINES entry,
  synced). A chrome write failing with `site_component_history` in the error
  is the trigger refusing to destroy unarchived bytes — fix the ledger, never
  drop the trigger casually (ROLLBACK sidecar exists for a deliberate one).
- `rendered_html_digest` belongs to the render path ONLY. Never backfill it,
  never re-stamp it from another writer — either move silences the detector.
- Pod verification: enumerate by IMAGE (65 pods run the chassis image; the
  label selector shows 2). The deployment that matters is `agent-chassis`.
- `bugs_open/229` (page-side sibling) is UNOWNED and carries the design
  constraints — do not copy-paste the chrome trigger there (DELETE+INSERT
  family; shared-abstraction question; possible RFC per the architecture
  seat).

## Unrelated but load-bearing, left for its owners

- HEAD's test suite is red on `TestValidDocSubjectTypes_Lockstep` since
  `e1628f7df` (RFC_015: migration 340 added `'decision'`, the Go
  `validDocSubjectTypes` list was not updated — the fourth instance of the
  both-halves landmine at LANDMINES.md:646). Owning sessions were active on
  08-08 evening; if it is still red when you read this, it may be abandoned —
  check `who-owns` and live transcripts before touching.
