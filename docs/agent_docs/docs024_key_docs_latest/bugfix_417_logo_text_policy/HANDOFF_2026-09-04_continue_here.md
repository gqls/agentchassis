# HANDOFF — bugfix 417/420/462 lane — 2026-09-04, continue here

**Supersedes `HANDOFF_2026-09-03b_continue_here.md`** (kept; its §5 wrong-calls and §6 owner items
are still live and are carried forward here). That file's ⭐ item was 462 §7a — *ship the standalone
legibility sweep now, plan the render-audit version as the one that stays correct.*
**The sweep is built, calibrated, run fleet-wide, and committed. One decision now blocks the rest.**

**Bug files — resolve by SLUG, all three numbers are ambiguous:**
- `bugs_open/417_HANDOFF_2026-08-31_planner_logo_exemplar_licenses_a_wordmark_it_never_names_so_the_image_model_invents_a_brand.md`
- `bugs_open/420_HANDOFF_2026-08-31_order_intake_publishes_the_billing_email_as_the_sites_public_contact_and_registers_it_as_a_renderable_claim.md` ⚠ the *other* 420 is the negation gate's prose walker
- `bugs_open/462_HANDOFF_2026-09-03_a_logo_can_be_perfectly_rendered_correctly_deployed_and_illegible_and_no_check_can_see_it.md` — **§8 is today's work**

**Working docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/`
Plain-prose version first: `SUMMARY_2026-09-04_logo_text_policy.md`.

---

## 1. WHAT EXISTS NOW — `scripts/audit-logo-legibility.py` (`01160bafc`), register **IMG-080**

```bash
./scripts/audit-logo-legibility.py                 # the fleet, ~90s
./scripts/audit-logo-legibility.py --site <domain>
./scripts/audit-logo-legibility.py --self-test     # both arms, offline, no cluster, no network
./scripts/audit-logo-legibility.py --json out.json # one record per site, shaped for a filer
```
Live the moment it is run — a script, no image, no migration, nothing to roll. **Council gate not
submitted:** `scripts/` is outside the scope regex except `pattern-check.py` (checked
`scripts/council-scope.sh`, not assumed).

**It takes BOTH verdict arms or neither.** Arm A = `max < 3:1` (no pixel anywhere clears the floor).
Arm B = under 15% of the mark's ink clears it. **A max-only rule PASSES the artefact 462 was filed
about** — post-regeneration websitepromotion reaches 20.75:1 off a magenta despill fringe while 86%
of it is white on a white header. The RUNBOOK's own "read max, not median" has been corrected in
place for exactly this.

## 2. THE FLEET RESULT `[MEASURED 2026-09-04 11:42Z]` — and the headline is not the finding count

| bucket | n | |
|---|---|---|
| **FINDING** | **2** | `websitepromotion.co.uk` (arm B — the motivating case, ruled to stay), `mortgagecalculator.co.uk` (arm A, **new**) |
| measured legible | 5 | designblog 88.1%, relojistas 75.5%, boxingonline 50.0%, seotools 29.3%, gamedesign 26.4% |
| **not judged — baked background** | **22** | no alpha at all; the header is not their backdrop. SITE_DEFECT_CATEGORIES 4.5 |
| **not displayed** | **3** | an active logo asset the served page never loads (2 render `class="logo-text"`) |
| BLIND | 2 | `noted.co.uk`, `loanandmortgagecalculator.co.uk` — not on the token-based theme |

**Only 7 of 34 logos can be judged at all**, because 22 of the 29 fetched have a background baked in
(the pre-`bugs_closed/424` estate). Those marks mostly read fine — **verified by eye, not assumed**;
judging them by this rule would have produced ~20 false findings. So "how widespread is 462" is
answerable for 7 sites today and grows only as old logos are remade.

⚠ **`mortgagecalculator` is deliberately under-claimed and should stay that way.** No pixel reaches
3:1 against its header (max 2.39:1) — but it was opened and looked at: a gold house-and-key mark on
cream that a person *can* see. "The whole mark sits below the WCAG non-text floor" is the true
sentence; "invisible" is not. Whether it earns a regeneration is the owner's call.

## 3. WHAT IS LEFT, ORDERED

1. **⭐ ROUTING — the only thing blocking 462, and it needs a decision, not a build.** The sweep
   *reports*; nothing files. `write_render_audit_findings_action.go:12-13` files `contrast_failure`
   at `css-patch-agent`, which repaints a CSS class and **cannot fix a pale PNG**. A logo finding
   needs a handler that can regenerate or replace an image. Per the owner ruling of 2026-08-02, the
   new work-item type owes its **producer set and `item_key` shape in the concept register in the
   commit that ships it**. `--json` is already shaped for that filer.
2. **462 §7a option (a), the render-audit version — still the destination, and the reason is
   staleness not coverage.** What shipped trusts the DECLARED theme token. Colour churn is live here
   (`generic_theme` landmine; `bugs_open/396` rewrites the theme row byte-for-byte), so a pass
   recorded today decays into a **false pass** — the one direction this bug is already about. The
   thresholds sit in one constant block so (a) reuses them rather than inventing a second rule.
   Every row already records `header_bg` + `measured_at` so a reader can tell "passed against a
   palette that no longer exists" from "passed".
3. **A NAMED BLIND SPOT, new and unowned:** nothing measures whether a **baked-background** mark
   reads against its own box. The sweep reports that measurement (`baked_bg`, `baked_max`,
   `baked_legible_frac`) and takes **no verdict**, because there is no known-bad artefact to
   calibrate one against. 22 of 29 logos sit there. Do not close this by picking a number.
4. ⚠ **IF THE FENCE IS EVER BUILT, ITS FINDINGS NEED A DESTINATION THAT IS NOT `needs_human_review`.**
   Contributed by the `bug 462` lane, **re-measured here, and their number is right while their
   framing needs one correction that makes the warning STRONGER** `[MEASURED 2026-09-04 ~13:4xZ]`:
   - the **`item_type`** `needs_human_review` is small — **40 rows all-time**, 35 open, oldest
     2026-07-22, **0 in the archive**. On that reading the queue looks alive;
   - but the same string is **also a STATUS**, and `status='needs_human_review'` holds **1,440 rows
     across ~dozens of item types**, oldest **2026-03-15** — `owned_page_review` 176,
     `cta_names_unknown_destination` 111, `image_source_unsatisfiable` 93, `unresolved_cta` 72,
     `voice_tells` 69, and so on. `revalidate_review_queue_action.go:3-5` records **370** on
     2026-07-25 and `bugs_open/033`'s auto-drain has closed 390 since, so it has still grown ~4x.
   **So it is not one dead queue you can route around — it is a dead STATE that almost any item type
   falls into.** A 417 finding does not have to be typed `needs_human_review` to end up parked with
   1,440 others; it only has to reach that status. Whatever the fence files, check where its type
   parks, and do not treat "filed for review" as having told anyone anything.
   ⚠ **And do not repeat the 1,439 figure as a count of that item_type** — it is the status. Both
   numbers are real and they answer different questions; the type/status collision is what makes
   the wrong one look right.

5. **The fence (417) — sample is STILL n=1 and the fence stays UNBUILT.** Terms unchanged: any
   lettered logo that carried the clause builds it. ⚠ **Check regeneration by the STORAGE KEY's date
   prefix, never `updated_at`** — see §5. A session wanting to close this deliberately should
   regenerate one of the 12 licence-carrying sites on purpose rather than wait; nothing schedules it.
6. **`not displayed` (3 sites) belongs to someone else** — `ai-agent-orchestration.com`, `cookly.uk`,
   `webdesign.co.uk` hold an active logo asset their header never loads. That is the 417 RUNBOOK's
   "a site has a logo asset but the header still shows TEXT" case, not a 462 finding. Nobody owns it.
7. **417, 420 and 462 all stay OPEN.** ⚠ **462 TRANSFERRED 2026-09-04** to a dedicated session
   (lane docs `docs024_key_docs_latest/bugfix_462_logo_legibility/`, transfer recorded in 462 §5).
   **This lane keeps 417 and 420.** Do not edit `bugs_open/462_*`, `scripts/audit-logo-legibility.py`
   or the 462 sections of this lane's RUNBOOK/NOTES; tell them first if you must. 462 on items 1–3; 417 on the fence residual; 420 on §C.

## 4. STILL THE OWNER'S — carried forward from the 09-03b handoff §6, unchanged except where noted

1. **RFC_058 (identity model) — RULED: Option C**, plus the two additions (a fifth *selling party*,
   deferred; **more than one contact per identity, NOT deferred**). ⚠ Do not decouple them: the
   relation forced by the second is what makes the first cheap to defer. **Still owed: §5.4's READER
   census** — writers refreshed 09-03 (still 4); the **14 readers remain dated 2026-08-31** and each
   must learn which identity it reads.
2. **The 420 §C residual** — does the narrow ruling extend to *derived* contacts (28 specs carry
   one)? Re-framed by RFC_058: the live question is what CONSENT STATE a classifier may write on a
   contact row, and the honest answer is likely a **third** state (*recorded, not published, never
   asked*). ⚠ A two-state answer designs the fill-only-if-empty inversion back in at row level.
   Timing agreed with `site_delivery_and_editor` 2026-09-03: they carry it to the owner when a
   DELIVERY decision is already in front of him, and hand the answer back verbatim. **Ownership
   stays here.**
3. **Ordering cannot reopen until the intake chat asks the contact question** — box-side.
4. **`bugs_open/421` still has no owner.**
5. **NEW, and it is item 1 of §3:** where a logo-legibility finding routes. Everything else about
   462 is decided.

## 5. TRAPS RECORDED TODAY — read these before re-running anything above

- **LANDMINE — `assets.updated_at` says a logo regenerated when it did not.** 8 of 34 rows were
  touched in 20 hours, six since midnight, about one an hour — the exact shape of a rolling
  regeneration. **None was one.** Every `storage_path` still carried its original date prefix
  (`loanzy` 20260818, `cv1` 20260825). Two of the eight are on 417's list of 13, so believing the
  column would have moved the fence sample **n=1 → a claimed n=3** on artefacts made before the
  override existed. A regeneration mints a **fresh key** (462 §6): the key's date is the artefact's
  birthday. Also: `assets.file_size` disagrees with the served file (12,325 recorded vs 70,156 served).
- **LANDMINE — a logo check that finds no logo in the `<header>` falls through to `assets.url` and
  measures a picture the page never loads.** The first `<header>` in the document is often a
  *content* heading (3 of 34 sites open with `info-card-grid__header`). And "no logo in the markup"
  is a **third state** — an active asset plus a text header. Every disciplined control (404 control,
  byte count, magic bytes) passes on the wrong file.
- The `assets.url` presigned-expiry landmine (dartsonline lane, 2026-07-30) now carries a dated
  population update: **1 of 34** today, not "half the rows" — thinner, and therefore harder to catch
  by sampling.

## 6. MY OWN CLAIMS AND ACTIONS THAT WERE WRONG TODAY

- **I wrote a landmine whose first half already existed** (`assets.url` presigned expiry, filed
  2026-07-30 — whose own *fires when* line names "writing a check that fetches an asset"). Caught by
  a *formatting* failure in `landmines-sync.py --check`, which printed the existing slug two lines
  above mine; correct formatting would have shipped the duplicate silently. `WRONG_CALLS.md` has the
  row. **The transferable half is not "grep before you file" — it is that the rule applies when you
  are WRITING an entry, not only when reading one. A duplicate is invisible from the inside.**
- **Three false measurements in the sweep's first run, all the same shape: careful, correct, and of
  the wrong image.** Trusted `assets.url` (a presigned link expired since 08-17) → false BLIND on a
  site whose logo is fine. Took the first `<header>` → confident statistics about images three pages
  never load. Read a CDN flake as an unmeasurable logo → 2 spurious BLINDs, fixed with retries.
  **None was visible in the numbers.** All three were caught by opening the artefact.
- **A fourth, caught before it shipped:** the first cut was ready to judge all 29 fetched logos.
  Compositing `farmerinsurance` and `apis` onto their headers and *looking* showed a legible mark
  inside a baked box — the statistic was measuring the box.
- **462 §1's control figure was wrong and is corrected in §8c:** seotools is recorded at 7.64:1 "vs
  the white header"; its header is `#faf8f3` and the real figure is **6.98:1**. The control still
  passes. §4's own warning not to assume white was not applied to §4's own control.
- **Housekeeping note:** my LANDMINES rewrite was swept into another session's commit
  (`bdb846972`) between writing and committing — CLAUDE.md's "your uncommitted work is not safe",
  working as documented. Nothing lost. And commit `0bd155668` shows **1 line removed** from
  LANDMINES: that is the `added:` attribution line of the presigned entry, replaced by a two-line
  version that keeps the original attribution and adds the update. No entry was deleted.

## 7. IF YOU READ ONE THING

**A measurement can be careful, correctly executed, properly dated — and taken of the wrong
object.** That happened four times today in one afternoon's work, and the marker discipline
(`[MEASURED]`, a stated method, a control) caught **none** of them, because every one of those
controls passes on the wrong file. What caught all four was opening the artefact and asking *is this
the thing a visitor actually sees?* — which is the same lesson as 09-03's, arriving from a new
direction. The one number that would have been most damaging, `updated_at`, was the one that looked
least like a judgement call.
