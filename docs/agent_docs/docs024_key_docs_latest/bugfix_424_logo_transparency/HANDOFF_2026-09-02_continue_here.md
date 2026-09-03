# HANDOFF — bugfix_424_logo_transparency, continue here

Updated 2026-09-03 ~15:35 BST — **the incident this lane exists to close is now closed.** Every
site caught by it has a genuine, verified-at-the-bytes fix. What's left is small, non-urgent
cleanup, not more remediation. Read this file first; for history read `NOTES_logo_transparency.md`
(full chronology, every correction, every cross-lane exchange) and the bug file's own tail (a peer
lane's evidence tables, verbatim).

## One-paragraph state

Gemini's image models cannot produce real transparency, so logos were switched to a keyed-magenta
prompt + a Go matte that removes it mathematically, with a fail-closed guard refusing anything the
model didn't key cleanly. The fix shipped in three rounds — the mechanism, a council-caught prompt
contradiction, and a council-caught guard bug (`BorderKeyed` measured flood-fill reachability
instead of final alpha, found by a peer lane's live production testing) — and all three are
confirmed live (checked independently at the artefact more than once, never taken on a report
alone). The bug that let it through affected four real sites before the fixes shipped:
`designblog.co.uk`, `seotools.co.uk`, `gamedesign.uk` (this lane's own remediation) and
`boxingonline.com` (a separate lane's, same mechanism). **All four are now genuinely fixed,
verified at the served bytes, not just at a DB status.**

## Final state of every site touched

| site | outcome | verified |
|---|---|---|
| `seotools.co.uk` | Fixed, attempt 2 of 3 | colour type 6, 92.2% transparent, 99.9% border, 0.085% fringe |
| `gamedesign.uk` | Fixed, attempt 3 of 3 (after an unrelated billing outage cleared) | colour type 6, 62.0% transparent, 100% border, 0.17% fringe |
| `boxingonline.com` | Fixed (separate lane, `site_delivery_and_editor`) | colour type 6, 80.8% transparent, 99.9% border, 0.038% fringe |
| `designblog.co.uk` | Fixed, round 2's final attempt (5 attempts total across two reset rounds) | colour type 6, 88.5% transparent, 99.9% border, 0.088% fringe, **dark-marked — no legibility risk** |

**Zero sites are left serving a silently-bad logo the platform believes is fine and isn't.** That
was the whole premise of the original incident, and it's closed.

## A real finding that came out of this remediation, not from this lane's own design work

While regenerating for an unrelated reason, a peer lane hit a genuine gap this fix's guard cannot
see by construction: a **light-coloured mark** on the magenta key can pass the transparency guard
perfectly (background correctly keyed, every transparency signal improved) while becoming nearly
invisible against a white page — because the mark's own light interior blends into the background
and the only thing left with any contrast is the despill fringe at its edges. This is not a defect
in this fix (verified: matting worked correctly on that artefact) — it's a category of failure the
transparency mechanism was never designed to catch, since it measures background removal, not
foreground legibility. Filed and owned as `bugs_open/462`; the owner has already ruled on it (see
below). `designblog.co.uk`'s own final result was checked specifically for this and is dark-marked,
so it doesn't carry the risk — but it's real, it's live on `websitepromotion.co.uk` right now by
deliberate owner decision (kept as 462's test case, not an outstanding repair), and it's worth
knowing if this mechanism ever gets extended to other flat-vector kinds (icons, sprite sheets).

## Decisions — all resolved

Every decision this lane raised has an answer:

1. When to roll — done, `v1.0.1356` then `v1.0.1358` both carry every fix.
2. What to do about the three (then four) broken sites — done, all fixed.
3. The retry-ladder / fail-closed-exhaustion policy — **resolved by the owner in `bugs_open/462`**,
   using this lane's own round-1 council objection as the deciding argument: no second fail-closed
   check is being added beside `BorderKeyed`, because two fail-closed gates sharing one retry
   ladder would mean more sites end up with NO logo, not just an imperfect one. Real numbers made
   the case: `seotools` needed 2 of 3 attempts, `gamedesign` 3 of 3, `designblog` five across two
   rounds.
4. Whether boxingonline was the right first deliberate test — overtaken; the owner authorised it
   independently and it succeeded.

## What's left — small, none of it urgent

1. **`bugs_open/462`'s legibility check** is being built by the `417`/adjacent lane, not this one —
   nothing for this lane to do beyond what's already recorded (a standalone sweep over stored
   assets first, a render-audit-anchored version for durability, informed by this lane's own
   staleness caution which the owning lane independently verified and hardened into the deciding
   argument).
2. **Confirm `bugs_open/421`'s status independently** — this fix never verified single-composition
   and must not be treated as having cleared it. Not checked this session.
3. **Update the `[UNMEASURED]` code comment on the threshold constants** (`inner=48`, `outer=110`,
   `minBorderKeyed=0.95`, `dynamic_adapter.go`) to reflect what the population of real runs
   actually showed: every observed refusal was `border_keyed` exactly `0.000`, a clean bimodal
   split with nothing landing near the threshold — so the constants are not the binding factor and
   there's no evidence left to gather by regenerating more. Cosmetic (the comment is stale, the
   behaviour is fine); not blocking anything.
4. **Low-priority, not blocking**: whether `platform/colour.ParseHex` was the best-fit existing
   helper (a council aside, premise didn't hold); clearer `grounded_in` citations in future council
   submissions so a reviewer without full-repo access doesn't misread a diff hunk as a missing
   file.
5. **Separately filed, unowned, unrelated**: `bugs_open/433` (the `assets.mime_type` gap, plus the
   finding that every sampled logo source object is actually JPEG stored under a `.png` name —
   `bugfix 417`'s contribution, already fully written up there).
6. **Separately filed, resolved same-day, no action needed**: `bugs_open/455` (the Gemini billing
   outage) — recorded for the pattern (third instance of provider-credit exhaustion this estate has
   hit, after `202` and `243`), not because anything is still broken.

## Where everything lives

- Design: `PLAN_2026-09-02_logo_background_transparency.md`
- Full chronology, every correction, every cross-lane exchange, in order: `NOTES_logo_transparency.md`
- Commands: `RUNBOOK_logo_transparency.md`
- Plain-English owner log: `README_where_we_are.md`
- Council submissions: `council_submission_424_logo_transparency.json` (round 1, APPROVED with 4
  advisories), `council_submission_424_round2_borderkeyed.json` (round 2, APPROVED clean)
- Concept register: `docs026_concept_register/register/imagery.md`, **IMG-076**
- Bug file, kept current, including a peer lane's own evidence tables verbatim:
  `bugs_open/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
- Related, not this lane's: `bugs_open/433` (mime_type/JPEG), `bugs_open/455` (billing, resolved),
  `bugs_open/462` (mark legibility, owner-ruled)

## Commits, in order (code)

- `6440ec968` — the fix (prompt policy + matte + fail-closed guard), 17 tests
- `2c7cda74b` — standing five docs, council submission, `bugs_open/433` filed
- `b2322a203` — round-1 council fix (magenta contradiction) — `Council-Reviewed: d018a48f-bd76-420a-8530-4491681d3bd4`
- `fcbe6071c` — round-2 fix (`BorderKeyed` measured the wrong thing), mutation-proven. No trailer
  (submitted after committing) — verdict **APPROVED, no objections**,
  `52bd50a1-3783-4801-868a-31a0ee599e60`, recorded here since forward-only forbids amending

Everything after these two is documentation and the two work-item resets — no further code
changes were needed once round 2 shipped. The fix, as committed, was sufficient for all four
sites; every remaining variable was attempt count, not code.
