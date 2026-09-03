# HANDOFF — bugfix_424_logo_transparency, continue here

Rewritten 2026-09-03 ~10:25 BST (supersedes the ~21:30 version from the day before — the roll
landed and the three broken sites were reset since then). Read this file first — it's "what to do
next", not "how we got here". For history: `NOTES_logo_transparency.md` (full chronology, every
correction), the bug file's own tail (a peer lane's CONTRIB, verbatim, with the production evidence
tables), `PLAN_2026-09-02_logo_background_transparency.md` (the design).

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

As of the last check in this session (shortly after the reset): all three back to `triaged`,
`attempt_count=1` (unchanged — this is honestly each site's second attempt, not reset to zero). A
background watch was running for completion; if this session ended before it reported a result,
the next thing to do is exactly the query above, then read what actually happened:

- **If a site is `complete` with `border_keyed` (adapter log) or a chunk-scanned high transparency
  %: verify at the served bytes** — `https://<domain>/assets/images/logo.png`, chunk-scan for
  colour type 6 or `tRNS` (RUNBOOK has the snippet), sample corner alpha. Send
  `site_delivery_and_editor` the reading — they asked for it explicitly, it feeds the owner's
  boxingonline decision.
- **If a site is `failed`**: read `error` and `result` on the row. The guard refusing IS the
  correct, designed behaviour for a bad generation now — a refusal is not itself a new problem,
  it's the mechanism working. Check `attempt_count` vs `max_attempts=3`; if it's about to exhaust,
  that's the guardian's LOW council objection from round 1 becoming concrete — see "Decisions"
  below.
- **If still `triaged` or `claimed` after a while**: the dispatch lane may just be busy (this queue
  has no stable drain rate — seen throughout this incident, runs landed anywhere from minutes to
  under an hour after triaging). Not itself a problem.

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
2. ~~What to do about the three broken sites~~ — **IN PROGRESS.** Reset executed 09:23:49 UTC with
   explicit authorisation; results were still landing as this was written (see "Live status").
3. **If any of the three exhausts its retry budget (max_attempts=3) without a good result**, does
   logo generation need a longer leash, or is "fail loud, go wherever the retry ladder's terminal
   state already goes" the right outcome? Still open — was a council LOW objection in round 1,
   became concrete once real failure-rate data existed (roughly 1 good of 4 stored, across the
   original incident's five runs).
4. **Whether boxingonline.com is the next deliberate test**, now that the three portfolio resets
   serve as the lower-stakes calibration `site_delivery_and_editor` recommended. Their own position
   (relayed 2026-09-02/03): boxingonline keeps its interim solid-colour mark until (a) the roll —
   now done — and (b) the three resets are read at the bytes. That second condition is what "Live
   status" above is for.

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
