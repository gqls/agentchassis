# HANDOFF — bugfix_424_logo_transparency, continue here

Updated 2026-09-03 ~10:40 BST (supersedes the ~10:25 version — one of the three resets is now
confirmed fixed at the served bytes). Read this file first — it's "what to do next", not "how we
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

**As of the last check in this session (~09:40 UTC, ~35 min after the reset): one confirmed
success, two still retrying.**

- **`seotools.co.uk`: SUCCESS, verified at the served bytes, not just the log.** Attempt 1 refused
  (09:28:17Z, `border_keyed=0.000`, correctly nothing stored); attempt 2 succeeded (09:30:13Z,
  `border_keyed=0.9993`). Fetched `https://seotools.co.uk/assets/images/logo.png` directly (200,
  26,975 bytes, fresh key `20260903/fe09592e-...`): PNG colour type 6 (RGBA — the exact chunk-scan
  signal absent when this bug was found), 92.21% of all pixels fully transparent, 99.92% of the
  border ring transparent, 0.085% residual magenta-fringe pixels (smaller than the earlier
  known-good example). This is the first end-to-end artefact-verified confirmation the fixed
  pipeline works on a real, unplanned, fleet-triggered generation — not a synthetic test, not a
  replay against old bytes.
- **`designblog.co.uk` and `gamedesign.uk`: both refused on their first retry attempt** (same
  `border_keyed=0.000` shape as seotools' own first try), **now on attempt 2 of 3, waiting out an
  approximately ONE-HOUR cooldown** before the queue will try again (`retry_after` ~10:28 and
  ~10:35 UTC respectively, checked at 09:39 UTC — this is longer than expected; not investigated
  further this session, just observed and worth knowing before assuming a "stuck" item needs
  intervention). A background watch (Monitor) was re-armed for up to an hour to catch the next
  attempt; if this session ended before it reported, the query above plus the read-outs below tell
  you what to do next:

- **If a site is `complete`**: don't trust the DB status alone — verify at the served bytes
  (`https://<domain>/assets/images/logo.png`, chunk-scan for colour type 6 or `tRNS`, RUNBOOK has
  the snippet; sample corner alpha; check `substring(storage_path from '.../([0-9]{8})/')` is
  TODAY's date, not a stale key — `assets.updated_at` can be bumped with no regeneration behind it,
  caught live this session on `gamedesign.uk`'s own row). Send `site_delivery_and_editor` the
  reading in the same shape as the seotools message above — they asked for it explicitly, it feeds
  the owner's boxingonline decision.
- **If a site is `failed`** (exhausted `max_attempts=3`): the guard refusing IS correct, designed
  behaviour — not itself a new problem. But three refusals with nothing ever stored means the
  guardian's LOW council objection from round 1 (the retry ladder can exhaust before landing a good
  result) has become a concrete, real outcome for that site, not just a theoretical risk — see
  "Decisions" below, item 3.
- **If still `triaged` after its `retry_after` has passed**: the dispatch lane may just be busy
  (no stable drain rate, observed throughout this incident — runs have landed anywhere from
  minutes to under an hour after becoming eligible). Not itself a problem.

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

## Decisions for the owner — most are now answered; two remain live

1. ~~When to roll~~ — **DONE.** v1.0.1356 carries everything.
2. ~~What to do about the three broken sites~~ — **IN PROGRESS, ONE CONFIRMED FIXED.**
   `seotools.co.uk` verified good at the served bytes. `designblog.co.uk` and `gamedesign.uk` each
   had a first retry correctly refused and are waiting out a ~1hr cooldown for attempt 2 (see "Live
   status" for the exact `retry_after` times and what to check).
3. **If any of the three exhausts its retry budget (`max_attempts=3`) without a good result**, does
   logo generation need a longer leash, or is "fail loud, go wherever the retry ladder's terminal
   state already goes" the right outcome? Still open — was a council LOW objection in round 1;
   seotools needing 2 of 3 attempts to land a good result is a live data point toward it being a
   real, not merely theoretical, constraint.
4. ~~Whether boxingonline.com is the next deliberate test~~ — **OVERTAKEN.** The owner separately
   authorised it (via `site_delivery_and_editor`) before the three portfolio resets had even
   finished — `needs_imagery` item `d71b7877-b42a-4019-9ede-74be363209ff`, fired 09:24:42 UTC, base
   prompt only, no interim ground clause. Its own result (not read by this session — not this
   lane's item) is `site_delivery_and_editor`'s to report.

## What's left before this lane can close

1. **Read the three reset runs at the artefact** once they land — not just the DB row status, the
   served PNG bytes (RUNBOOK has the exact commands). Send readings to `site_delivery_and_editor`.
2. **Feed those readings into decision #4** (boxingonline timing) and #3 (retry-ladder policy) —
   both are now answerable with real data rather than hypotheticals.
3. **Date-stamp the threshold constants** (`inner=48`, `outer=110`, `minBorderKeyed=0.95`,
   `dynamic_adapter.go`) from whatever this batch of real runs shows, if they show anything the
   original five didn't already establish. Currently `[UNMEASURED]` as constants.
4. **Look at the despill fringe** on a genuinely good result, if one of the three (or
   websitepromotion, already known-good) shows it clearly enough to diagnose. Recorded, not fixed.
5. **Confirm `bugs_open/421`'s status independently** — this fix does not verify single-composition
   and must not be treated as having cleared it.
6. **Low-priority, not blocking**: whether `platform/colour.ParseHex` was the best-fit existing
   helper (a council aside, premise didn't hold); clearer `grounded_in` citations next council
   submission.
7. **Separately filed, unowned**: `bugs_open/433` (mime_type gap + the JPEG-under-a-.png-name
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
