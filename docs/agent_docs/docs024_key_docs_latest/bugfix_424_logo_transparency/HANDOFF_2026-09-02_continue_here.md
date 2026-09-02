# HANDOFF — bugfix_424_logo_transparency, continue here

Written 2026-09-02 ~17:30 BST, for a fresh session (or the owner) to pick up. Read this file, then
`PLAN_2026-09-02_logo_background_transparency.md` for the design, then `NOTES_logo_transparency.md`
for the full chronological detail if you need it — this file is the "what to do next", not the
"how we got here".

## One-paragraph state

`bugs_open/424`: Gemini's image models cannot produce a transparent background at all (confirmed
externally and by code inspection), so the logo pipeline was changed to ask for a fixed magenta key
colour instead and remove it mathematically after generation. **Code is written, tested, committed,
and went through council review (APPROVED).** The council caught a real bug in the first version
(a prompt contradiction) — that's fixed and committed too. **A chassis build deployed today (tag
`v1.0.1354`) carries the ORIGINAL fix but predates the contradiction fix, so it is currently running
with the bug the council found.** Nothing has been tested against a real logo yet. Three decisions
need the owner (below) before this lane can close.

## Decisions that need the owner (not something to decide unilaterally)

1. **Roll again before testing on a live asset, or test anyway and accept the risk?** The
   currently-deployed build (`v1.0.1354`) will hit the magenta/background contradiction the council
   found if a real `kind=logo` generation runs against it right now (see "Verification" below for
   how this was confirmed, not assumed). The fix is already committed (`b2322a203`) — it just needs
   another rebuild + roll of both `agent-chassis` and `image-generator-adapter` to be live. This is
   a fleet-wide release (`make release`, owner-run per CLAUDE.md), not something this session can
   trigger unilaterally.

2. **Is boxingonline.com the right first real test, given it's a live paying customer's site?**
   The threshold constants (`inner=48`, `outer=110`, `minBorderKeyed=0.95` in
   `dynamic_adapter.go`) are `[UNMEASURED]` — extrapolated from the interim ground-colour
   regeneration's drift, never from an actual magenta-keyed generation. Calibrating them properly
   means regenerating a real logo and looking at the result. boxingonline.com is the site this bug
   was diagnosed on and the natural test subject, but it's also (per `site_delivery_and_editor`'s
   own notes) the first paid customer, under an owner ruling that everything is fixed before the
   delivery email, and there is **no revert seam for a generated asset** — a bad regeneration
   cannot be undone, only refused (which the fail-closed guard does, but a refusal still burns a
   generation and leaves the interim logo in place, not a worse one). Options: test on boxingonline
   directly (fastest, but on the paying customer's live asset); find or build a lower-stakes test
   site first (slower, safer); or accept the interim logo as good enough for now and defer real
   calibration to the next site that needs a fresh logo. Not this session's call.

3. **The fail-closed guard's interaction with the retry ladder** (guardian's LOW objection,
   council round 1, not yet addressed in code): a site whose model reliably ignores the key-colour
   instruction could now exhaust `needs_logo`/`needs_hero_image`'s `attempt_count`/`max_attempts`
   rather than complete with a flawed-but-present asset. Is that the right failure mode (fail loud,
   go to human review via whatever the retry ladder's terminal state already does), or does this
   need explicit wiring to escalate faster / differently for a matting-specific failure? This is a
   product/policy call about how much friction is acceptable, not an engineering question with one
   right answer.

Smaller items that don't need the owner, just someone's time — see "What's left" below.

## What's actually live right now — verified, not assumed

Checked at the artefact 2026-09-02 ~17:26 BST, both services, with positive+negative controls
(commands in `RUNBOOK_logo_transparency.md`):

| | image-generator-adapter | agent-chassis |
|---|---|---|
| running build | `ebf27c603` (tag v1.0.1354) | binary-probed, same tag |
| original matting fix (`6440ec968`) | **LIVE** (`git merge-base --is-ancestor` = yes) | **LIVE** (`applyLogoBackgroundPolicy` present in `/proc/1/exe`) |
| magenta-contradiction fix (`b2322a203`) | **NOT LIVE** (postdates this build) | **NOT LIVE** (fix text absent from `/proc/1/exe`) |

**Practical consequence: do not run a real logo generation against the current build.** The prompt
will simultaneously tell the model to paint the background magenta and forbid using magenta —
council-confirmed real risk of the model refusing the key colour, which would defeat the mechanism
and likely trip the fail-closed guard (a refused-but-harmless outcome, but not a useful test).

## What's left before this lane can close

1. **Rebuild + roll both services** with `b2322a203` included (owner decision #1 above).
2. **Re-verify at the artefact** after that roll (RUNBOOK has the exact commands) — do not assume
   the roll picked it up.
3. **Run one real magenta-keyed generation** against a chosen test subject (owner decision #2) and
   read the outcome: did `BorderKeyed` clear the 0.95 threshold, what was the actual measured
   colour drift from `#FF00FF`, does the served PNG chunk-scan as colour type 6 or `tRNS` present
   (RUNBOOK has the exact Python snippet — check BOTH signals, never one alone). Use that
   measurement to set the threshold constants from evidence instead of the current extrapolation,
   and date the change.
4. **Decide and, if needed, wire the retry-ladder interaction** (owner decision #3).
5. **Confirm `bugs_open/421`'s status independently** — this fix does not verify single-composition
   (see PLAN's "Does NOT fix" note) and must not be treated as having cleared it. Check whether 421
   is still open on its own evidence before considering the logo asset fully resolved.
6. **Optional, low-priority, not blocking:** a closer look at whether `platform/colour.ParseHex`
   was really the best-fit existing helper (reuse_agent's MEDIUM, council round 1 — the premise
   that a NEW package was added doesn't hold, since it was reused, but a second look at the other
   colour-handling code it named costs little); a `grounded_in` citation next time so a reviewer
   without full-repo access doesn't misread a diff hunk as a missing file (two false positives this
   round, both harmless but noisy).
7. **Not required to close this bug, but flagged for whoever wants it:** `bugs_open/433` (the
   fleet-wide `assets.mime_type` gap this bug's verification surfaced) is separately filed and
   entirely unowned — diagnosis of the actual writer hasn't started.

None of items 1-5 are large. The honest state is: the hard part (diagnosis, design, implementation,
review) is done and approved; what's left is calibration against a real generation and two
owner-level judgement calls about risk and timing, not more engineering.

## Where everything lives

- Design: `PLAN_2026-09-02_logo_background_transparency.md`
- Full chronological detail, including every misstep and correction: `NOTES_logo_transparency.md`
- Commands (build verification, DB queries, council submission, test runner): `RUNBOOK_logo_transparency.md`
- Plain-English log for the owner: `README_where_we_are.md`
- Council submission as filed: `council_submission_424_logo_transparency.json`
- Concept register entry: `docs026_concept_register/register/imagery.md`, **IMG-076**
- The bug file itself, kept current: `bugs_open/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
- The spinoff bug: `bugs_open/433_HANDOFF_2026-09-02_assets_mime_type_is_empty_on_910_of_1277_rows_fleet_wide.md`

## Commits, in order

- `6440ec968` — the fix (prompt policy + matte + fail-closed guard), 17 tests
- `2c7cda74b` — standing five docs, council submission record, bugs_open/433 filed
- `b2322a203` — round-1 council fix (the magenta contradiction), new regression test

Council: `SUBMISSION_CORR=d018a48f-bd76-420a-8530-4491681d3bd4` — APPROVED, round 1. Both `Council-
Submitted:` (first commit) and `Council-Reviewed:` (the round-1 fix commit) trailers are on the
right commits, per CLAUDE.md's rule that the two must never be swapped.
