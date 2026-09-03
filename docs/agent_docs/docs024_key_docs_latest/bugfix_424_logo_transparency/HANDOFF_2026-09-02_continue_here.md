# HANDOFF — bugfix_424_logo_transparency, continue here

Updated 2026-09-03 ~14:30 BST (supersedes the ~12:45 version — owner authorised a second retry
round for `designblog.co.uk`, now mid-ladder; `boxingonline.com` separately confirmed fixed too;
**and a real mark-legibility interaction reopened the despill-fringe item this file previously
called closed** — read that item again if you read the earlier version). Read this file first —
it's "what to do next", not "how we
got here". For history: `NOTES_logo_transparency.md` (full chronology, every correction), the bug
file's own tail (a peer lane's CONTRIB, verbatim, with the production evidence tables),
`PLAN_2026-09-02_logo_background_transparency.md` (the design).

## One-paragraph state

The fix (keyed-ground matting for logos, since Gemini cannot produce real transparency) shipped in
three rounds: the mechanism itself, a council-caught prompt contradiction, and a council-caught
guard bug found by a peer lane's live production testing — that bug let three real sites
(`designblog.co.uk`, `seotools.co.uk`, `gamedesign.uk`) get a near-opaque logo that the platform
believed was fine. **All three fixes are now confirmed live** (v1.0.1356, verified independently at
the artefact, not taken on any report alone). **The owner has authorised, and this session has
executed, a reset of all three affected sites' logo work items** so the queue retries them under
the corrected guard. Results were still landing as this was written — see "Live status" below,
which is the one section worth re-reading fresh rather than trusting as printed.

## Live status — check this again, don't just read it

The reset ran at **2026-09-03 09:23:49 UTC**. Re-check current state before acting on anything below:

```sql
SELECT s.domain, w.status, w.attempt_count, w.completed_at, w.result->'response'->'image_result'
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.id IN ('24dff15c-1989-4332-aeaa-62b0929a8a88', -- designblog.co.uk
               'b178ca1b-b1bc-411b-ae3b-d63b8424dad0', -- seotools.co.uk
               '2a4408aa-800b-443d-aa2e-32e919978ecb'); -- gamedesign.uk
```

**As of ~11:45 UTC, final state for all three original resets: 2 confirmed fixed, 1 exhausted,
awaiting an owner decision.**

- **`seotools.co.uk`: FIXED, verified at the served bytes.** Refused once (`border_keyed=0.000`,
  correctly nothing stored), succeeded on attempt 2 (`border_keyed=0.9993`). Served bytes: colour
  type 6 (RGBA), 92.21% fully transparent overall, 99.92% of the border ring transparent, 0.085%
  residual magenta fringe. First end-to-end artefact-verified confirmation the fixed pipeline
  works on a real, unplanned, fleet-triggered generation.
- **`gamedesign.uk`: FIXED, verified at the served bytes.** Refused twice — once on content
  (`border_keyed=0.000`), then several times against an UNRELATED billing outage (`bugs_open/455`,
  below) — succeeded on the 3rd counted attempt once the outage cleared. Served bytes: colour
  type 6, 100% of the border ring transparent, 61.98% fully transparent overall (a 400×400 square
  logo has less empty margin than seotools' 400×218 rectangle — not a quality difference), 0.174%
  residual fringe.
- **`designblog.co.uk`: EXHAUSTED (3 of 3 attempts), NOT fixed, awaiting an owner decision on
  whether to reset it a second time.** Its final counted attempt was a genuine content refusal
  (`border_keyed=0.000`), not the billing outage — the retry ladder's accounting worked correctly
  (infra failures didn't count against it). **Nothing worse happened**: verified its asset row is
  untouched (`key_date` still `20260902`), so it is still serving yesterday's original broken logo,
  not a new bad one. **This session asked the owner whether to reset it again** (same low-risk
  operation as the original three, but a new action beyond that original authorisation) — check
  whether that question was answered before assuming this is closed out.

**An unrelated billing outage was found and resolved mid-watch — `bugs_open/455`.** The Gemini
image provider returned "prepayment credits are depleted" for ~37–70 minutes (`10:31Z`–`~11:08Z`,
confirmed cleared by `11:41Z`), blocking `designblog.co.uk`'s and `gamedesign.uk`'s retries (and
`boxingonline.com`'s, a different lane's item) — nothing to do with the 424 fix itself. Resolved
same-day, matching `bugs_open/243`'s prior pattern; filed but not yet closed (no direct
confirmation of a deliberate top-up, only inferred from the traffic gap). If more image-generation
failures show `429`/`prepayment credits` in the error text, it may have recurred — check `455`
before assuming it's this fix's guard.

**The verification method that held up across all three, if `designblog.co.uk` gets reset and you
need to check it again**: don't trust the DB `status` column alone — fetch the served bytes
(`https://<domain>/assets/images/logo.png`), chunk-scan for colour type 6 or `tRNS` (RUNBOOK has
the snippet), sample corner alpha, and confirm `substring(storage_path from '.../([0-9]{8})/')` is
TODAY's date, not a stale key (`assets.updated_at` can be bumped with no regeneration behind it —
caught live this session on `gamedesign.uk`'s own row before it mattered). Send
`site_delivery_and_editor` the reading in the same shape as the seotools/gamedesign messages in
NOTES — they've been asking for it explicitly, it feeds the owner's boxingonline decision. The
dispatch lane has no stable drain rate (runs landed anywhere from minutes to over an hour after
becoming eligible throughout this incident) — a `triaged` item past its `retry_after` is not
necessarily stuck.

## What's actually live — verified 2026-09-03 ~10:00 BST, not assumed

| | image-generator-adapter / agent-chassis |
|---|---|
| stamp | `7bf1ff674021f2d57dfd0aa41324541070646c3a` (tag v1.0.1356) |
| original matting (`6440ec968`) | LIVE |
| round-1 fix, prompt contradiction (`b2322a203`) | LIVE |
| round-2 fix, guard measured the wrong thing (`fcbe6071c`) | LIVE |

Verified both by build-provenance log line (image-generator-adapter) and full positive/target/
negative control binary probe (agent-chassis), AND by `git merge-base --is-ancestor` for all three
commits against the stamp with a negative control (current HEAD, correctly absent). Two independent
methods agree.

## Decisions for the owner — one still genuinely open

1. ~~When to roll~~ — **DONE.** v1.0.1356 carries everything.
2. ~~What to do about the three broken sites~~ — **DONE for 2 of 3.** `seotools.co.uk` and
   `gamedesign.uk` both confirmed fixed at the served bytes.
3. **`designblog.co.uk` exhausted its 3 attempts without a good result — reset it again, or leave
   it on the original (broken) logo for now?** This session asked the owner directly and had not
   received an answer as of this handoff being written. A reset is the same low-risk operation as
   the original three (a bad result still cannot get stored, only refused), so the honest framing
   is "is it worth another cycle" rather than "is it safe" — but it's a new action, not covered by
   the original three-site authorisation, so don't do it without asking again if picking this up
   fresh. This is also the concrete instance of the council's round-1 LOW objection (the retry
   ladder can exhaust before landing a good result) — worth deciding on its own merits now that
   there's a real case, not a hypothetical one.
4. ~~Whether boxingonline.com is the next deliberate test~~ — **OVERTAKEN.** The owner separately
   authorised it (via `site_delivery_and_editor`) before the three portfolio resets had even
   finished — `needs_imagery` item `d71b7877-b42a-4019-9ede-74be363209ff`, fired 09:24:42 UTC, base
   prompt only, no interim ground clause. Its own result (not read by this session — not this
   lane's item) is `site_delivery_and_editor`'s to report.

## What's left before this lane can close

1. ~~Read the three reset runs at the artefact~~ — **DONE.**
2. **`designblog.co.uk` retry round 2 IN PROGRESS** — owner authorised, reset with a fresh
   3-attempt budget at `12:50:50 UTC`, currently running. This is the last genuinely open action
   item; check "Live status" (below, once updated) or `NOTES` for the outcome.
3. ~~Date-stamp the threshold constants~~ — **EFFECTIVELY SETTLED, not by tuning.** Independent
   read by the `bugfix 417` lane across every run so far: every refusal is `border_keyed` exactly
   `0.000`, a clean bimodal split with nothing landing near `0.95` on either side. **The threshold
   is not the binding factor — the model either keys the ground correctly or does not key it at
   all, and retuning `inner`/`outer` would not have saved any observed failure.** Retry budget
   (attempts), not the constants, is the real lever — feeds decision #3-below directly. Constants
   themselves still carry `[UNMEASURED]` in the code comment; a future session could update that
   comment to reflect this finding, but there is no evidence left to gather by regenerating more.
4. **REOPENED — the despill fringe is a real, potentially serious problem, not a cosmetic one.**
   This morning's "effectively closed" call (0.01%–0.05% on two dark-marked examples) was
   **retracted by the same reporting session** after a third real generation: a LIGHT-marked
   result (`websitepromotion.co.uk`'s second regeneration) passed the guard cleanly (transparency
   84.3%→93.4%, genuinely correct matting) but is now close to invisible against a white header
   (median contrast `1.43:1 → 1.01:1`) — because most of a light mark's own opaque pixels are
   themselves near-white (content, not a matte defect) and the small remainder that has any
   contrast is 63% leftover magenta despill fringe. **The fringe percentage barely changed
   (0.62%→0.48%); what changed is that a light mark has nothing else visible to dilute it with.**
   So severity depends on the mark's own lightness, which this fix's prompt never constrains —
   only "no magenta/pink in the artwork" is asked for, not "avoid near-white". **This is
   structurally invisible to `BorderKeyed` by construction** — the guard measures whether the
   BACKGROUND became transparent, not whether the FOREGROUND remains legible once composited; a
   perfectly keyed ground is exactly what produces this failure. Not a defect in this lane's own
   design, but a real gap this design cannot close on its own — squarely `bugs_open/462`'s
   territory (mark legibility), with a concrete architectural note for it: a contrast check must
   run AFTER matting and against the real deployment background, never pre-matte (pre-matte a
   white-mark-on-magenta image reads as high-contrast and would pass happily). **Owner ruling,
   same day: `websitepromotion.co.uk`'s illegible logo is NOT being restored** — it's a deliberate
   decision, kept live as 462's own motivating test case. Do not read it as an outstanding repair
   or a regression if it surfaces in a future sweep from this lane; it is correct-per-424 (matting
   genuinely worked) and known-bad-per-462, on purpose.
5. **Confirm `bugs_open/421`'s status independently** — this fix does not verify single-composition
   and must not be treated as having cleared it.
6. ~~Decide whether the retry-ladder policy needs revisiting now that `bugs_open/462` exists~~ —
   **RESOLVED, by the owner, in 462.** The owner chose 462's fix candidate 2 (a post-hoc sweep)
   over candidate 1 (a second fail-closed statistic on this SAME retry ladder) **specifically
   because of this lane's own round-1 LOW council objection** — a second fail-closed check
   compounds exhaustion risk (more sites with NO logo, not just an imperfect one), and today's
   numbers (`seotools` 2/3, `gamedesign` 3/3, `designblog` five attempts across two rounds with
   nothing stored) made that concrete rather than hypothetical. No second gate is being added
   beside `BorderKeyed`. Nothing for this lane to decide further here.
7. **Consider whether `bugs_open/455` (the billing outage) warrants a prevention conversation** —
   third instance of provider-credit/quota exhaustion counting `202` and `243`, all resolved
   same-day by adding credit, none of them prevented from recurring. Not this lane's call to make
   unilaterally, just worth surfacing given the pattern.
8. **Low-priority, not blocking**: whether `platform/colour.ParseHex` was the best-fit existing
   helper (a council aside, premise didn't hold); clearer `grounded_in` citations next council
   submission.
9. **Separately filed, unowned**: `bugs_open/433` (mime_type gap + the JPEG-under-a-.png-name
   finding) — not blocking this lane.

## Where everything lives

- Design: `PLAN_2026-09-02_logo_background_transparency.md`
- Full chronology, every correction, the peer's CONTRIB with the production evidence tables:
  `NOTES_logo_transparency.md` and the bug file's own tail
- Commands: `RUNBOOK_logo_transparency.md`
- Plain-English owner log: `README_where_we_are.md`
- Council submissions: `council_submission_424_logo_transparency.json` (round 1),
  `council_submission_424_round2_borderkeyed.json` (round 2)
- Concept register: `docs026_concept_register/register/imagery.md`, **IMG-076**
- Bug file, kept current: `bugs_open/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
- Spinoff bug: `bugs_open/433_HANDOFF_2026-09-02_assets_mime_type_is_empty_on_910_of_1277_rows_fleet_wide.md`

## Commits, in order

- `6440ec968` — the fix (prompt policy + matte + fail-closed guard), 17 tests
- `2c7cda74b` — standing five docs, council submission, bugs_open/433 filed
- `b2322a203` — round-1 council fix (magenta contradiction) — `Council-Reviewed: d018a48f-bd76-420a-8530-4491681d3bd4`
- `fcbe6071c` — round-2 fix (BorderKeyed measured the wrong thing), mutation-proven. No trailer (submitted after committing) — verdict **APPROVED, no objections**, `52bd50a1-3783-4801-868a-31a0ee599e60`, recorded here since forward-only forbids amending the commit
- (docs-only commits since: incident record, round-2 validation against real production data, this handoff)

## Work item IDs, for reference

- `24dff15c-1989-4332-aeaa-62b0929a8a88` — designblog.co.uk, `needs_imagery:site:-:logo`
- `b178ca1b-b1bc-411b-ae3b-d63b8424dad0` — seotools.co.uk, `needs_imagery:site:-:logo`
- `2a4408aa-800b-443d-aa2e-32e919978ecb` — gamedesign.uk, `needs_imagery:site:-:logo`

Landmine for next time: these use `item_type='needs_imagery'` with `item_key` ending
`:site:-:logo`, not `item_type='needs_logo'` — that type exists in the schema but wasn't what these
rows used. Query on `item_key ILIKE '%site:-:logo%'` or the domain join, not the type name alone.
