# HANDOFF — bugfix_424_logo_transparency, continue here

Rewritten 2026-09-02 ~21:30 BST (supersedes the ~17:30 version — the situation changed materially
in between; see NOTES for the full chronology if you want it). Read this file, then
`PLAN_2026-09-02_logo_background_transparency.md` for the design if you need it — this file is
"what to do next", not "how we got here".

## One-paragraph state — READ THIS FIRST, it leads with the urgent part

**There is an active, live incident, not just an open bug.** Three real sites —
`designblog.co.uk`, `seotools.co.uk`, `gamedesign.uk` — currently serve a logo that the platform's
own guard certified as fine and is not: 90%+ opaque where it should be transparent. This happened
because the fleet's own autonomous image queue fired the (buggy) new mechanism against real sites
before anyone had tested it — nobody "triggered" this manually, and no warning in any handoff could
have stopped it. The bug that let this through has been found (by a peer lane's live testing, not
this session), verified against the actual code, fixed, tested, mutation-proven, and **approved by
council** — but **is not deployed yet**, and **the three broken assets will not fix themselves once
it is** (see "What's left", item 2).

## Decisions that need the owner

1. **When to roll, and how urgently.** Both fixes (`b2322a203`: a prompt contradiction the council
   found round 1; `fcbe6071c`: the guard-measures-the-wrong-thing bug round 2) are committed,
   tested, and council-APPROVED. Neither is deployed. Every hour the current build stays up, the
   autonomous queue can hand out a fourth broken logo — this is not hypothetical, it already
   happened three times in one afternoon. This reads as urgent to this session; only the owner can
   actually run `make release`.
2. **What happens to the three already-broken sites.** Once the fix is live, their logo assets do
   NOT self-heal — the work items are `complete`, not `triaged`, so nothing retries them
   automatically (unlike `websitepromotion.co.uk`, whose failed *first* attempt went back to
   `triaged` and got a working result on retry). Someone needs to deliberately reset
   `designblog.co.uk`, `seotools.co.uk` and `gamedesign.uk`'s `needs_imagery:site:-:logo` items
   AFTER the roll — not before, or they'll just fail a fourth time against the still-broken build.
   This session has not done that reset; it needs the roll to happen first, and arguably needs the
   owner or whoever owns each site to know their logo is about to change again.
3. **Whether to pause the autonomous logo queue until this rolls**, given it's the thing that
   caused the incident and will keep running regardless of what any handoff says. Not this
   session's call — pausing fleet automation is a bigger action than fixing the bug that needed it.
4. **Is boxingonline.com (or any specific site) the right first DELIBERATE test**, given there's
   still no revert seam for a generated asset and the observed real-world success rate so far is
   low (1 good result of 4 stored, across the five runs the incident produced). Less urgent than
   1–3 — the incident already answered "does this need calibration" with real data; this is about
   whether to go looking for more data on purpose.
5. **The fail-closed guard's interaction with the retry ladder** (guardian's LOW objection, council
   round 1): with a real observed ~25% first-attempt success rate, sites could plausibly exhaust
   `needs_logo`'s retry budget before landing a good result. Is that acceptable (fail loud, whatever
   the retry ladder's terminal state already does), or does logo generation need a longer leash?

## What actually happened — the short version, evidence in NOTES/bug file

A peer lane (`bugfix_420_417`, via `site_delivery_and_editor`) ran live dynamic tests and found
that `MatteStats.BorderKeyed` — the number the fail-closed guard checks against 0.95 before
allowing an upload — was computed from "was this border pixel close enough to be reachable by the
flood-fill" (a wide net, `dist <= outer = 110`), not from "did this border pixel actually end up
transparent" (`dist <= inner = 48`). A ground that landed anywhere in the wide gap between those two
numbers scored `BorderKeyed ≈ 1.000` — a perfect pass — while staying ~90% opaque. Verified against
`keyground.go` before fixing anything, not taken on the report's word. Five real runs that
afternoon: 3 stored with 0% actual transparency (guard said 1.000, wrong), 1 correctly refused
(guard said 0, right), 1 stored genuinely good (87.4% transparent, guard said 1.000, right this
time by coincidence of the number, not because the guard could tell). Fixed by tracking each
pixel's real final alpha and computing the stat from that instead — `keyground.go`, commit
`fcbe6071c`, council-approved clean with zero objections (`52bd50a1-3783-4801-868a-31a0ee599e60`).

The encouraging part, from the peer's own later correction: the model CAN land a good result
(`websitepromotion` proves it), so this looks like a variance problem the fixed guard can now
correctly police, not necessarily a sign the threshold numbers themselves are wrong. Not
independently re-verified against a real post-fix run yet — a working hypothesis, stated as one.

## What's actually live right now — verified at the artefact where possible

| | image-generator-adapter / agent-chassis |
|---|---|
| original matting fix (`6440ec968`) | **LIVE** (verified ~17:26 BST, binary probe + build provenance, both services) |
| round-1 fix, the prompt contradiction (`b2322a203`) | confirmed **LIVE** as of a peer's provenance read at 20:56:58Z (stamp `0d2feee2f`) — independently confirmed via `git merge-base --is-ancestor`, no cluster access needed |
| round-2 fix, the guard bug (`fcbe6071c`) | **NOT YET DEPLOYED** — committed and approved after that stamp |

kubectl access dropped mid-session (expired token, matches the peer's report) and came back on its
own later — if it's down when you pick this up, git-based checks (`merge-base --is-ancestor`
against a provenance stamp someone else reports, or `verify-head-builds.sh` locally) still work
without it.

## What's left before this lane can close

1. **Roll the fleet** with `fcbe6071c` included (owner decision #1).
2. **Re-verify at the artefact** after that roll — RUNBOOK has the exact commands (build
   provenance, or the binary probe with positive/negative controls if the log line scrolls out of
   range on a busy service).
3. **Reset the three broken sites' logo work items** after the roll, not before (owner decisions
   #2–3 first). Then watch what the queue produces — this IS the "run one real generation" step
   the previous version of this handoff asked for as a separate, optional thing; it is no longer
   optional or separate, it is the fix for a live problem.
4. **Use those (and any further) real runs to date-stamp the threshold constants** (`inner=48`,
   `outer=110`, `minBorderKeyed=0.95` in `dynamic_adapter.go`) from evidence — currently
   `[UNMEASURED]` as constants, extrapolated from an unrelated regeneration. The incident already
   supplied five real data points; use them rather than waiting for more.
5. **Look at the despill fringe** on `websitepromotion`'s good result once someone can see the
   actual image (this session couldn't) — recorded, not fixed.
6. **Decide and, if needed, wire the retry-ladder interaction** (owner decision #5).
7. **Confirm `bugs_open/421`'s status independently** — this fix does not verify single-composition
   and must not be treated as having cleared it.
8. **Low-priority, not blocking:** whether `platform/colour.ParseHex` was the best-fit existing
   helper (a council reviewer's aside, premise didn't hold — it was reused, not new); clearer
   `grounded_in` citations next submission so a reviewer without full-repo access doesn't misread a
   diff hunk as a missing file (two harmless false positives this round).
9. **Separately filed, unowned:** `bugs_open/433` (fleet-wide `mime_type` gap) — not blocking this
   lane, flagged for whoever wants it. Note also (from the peer's investigation): every logo source
   object sampled is actually JPEG under a `.png` name/URL/extension, and the adapter discards the
   provider's real MIME type at upload — a second, independent reason the `mime_type` column can't
   be trusted, and a second, independent argument that a colour-distance matte is fighting JPEG
   chroma subsampling on top of everything else. Detail in `bugs_open/433`'s own update.

## Where everything lives

- Design: `PLAN_2026-09-02_logo_background_transparency.md`
- Full chronological detail, every misstep and correction, the peer's CONTRIB verbatim: `NOTES_logo_transparency.md` and the bug file's own tail
- Commands: `RUNBOOK_logo_transparency.md`
- Plain-English log for the owner: `README_where_we_are.md`
- Council submissions: `council_submission_424_logo_transparency.json` (round 1), `council_submission_424_round2_borderkeyed.json` (round 2)
- Concept register: `docs026_concept_register/register/imagery.md`, **IMG-076**
- The bug file, kept current, including the peer's own CONTRIB with full evidence tables: `bugs_open/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
- The spinoff bug: `bugs_open/433_HANDOFF_2026-09-02_assets_mime_type_is_empty_on_910_of_1277_rows_fleet_wide.md`

## Commits, in order

- `6440ec968` — the fix (prompt policy + matte + fail-closed guard), 17 tests
- `2c7cda74b` — standing five docs, council submission record, bugs_open/433 filed
- `b2322a203` — round-1 council fix (the magenta contradiction), new regression test — `Council-Reviewed: d018a48f-bd76-420a-8530-4491681d3bd4`
- `fcbe6071c` — round-2 fix (BorderKeyed measured the wrong thing), new regression test, mutation-proven. **No `Council-Submitted:` trailer** (submitted after committing, not alongside — the bug was found and fixed fast). Verdict: **APPROVED, no objections**, `52bd50a1-3783-4801-868a-31a0ee599e60`. Since forward-only forbids amending, this correlation is recorded here rather than on the commit — join them by hand if `098`'s report doesn't already do it by file-overlap.
