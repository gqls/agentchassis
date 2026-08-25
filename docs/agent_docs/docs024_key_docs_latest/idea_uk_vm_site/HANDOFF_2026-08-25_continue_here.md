# HANDOFF 2026-08-25 — the CSS incident is CLOSED fleet-wide and idea.uk is healed at the served page. The oldest gap standing is news/content_sources.

> **SUPERSEDED 2026-08-25 (evening) by `HANDOFF_2026-08-25b_continue_here.md`** — §4 item 1 (news) is DONE and live at the artefact; items 2–4 are answered there. Read that file first; this one stays for the CSS-incident close-out (§1) and the queue-reading trap (§2).

**Supersedes `HANDOFF_2026-08-18_continue_here.md` as the cold-start file** (its 08-19
live-incident banner is resolved — see §1). The 08-16 file remains the reference for
the honesty-arc history and the head blind spot; `HANDOFF_2026-08-11` §3 for RFC_015.

Cold-start order: **this file → `HANDOFF_2026-08-16` §4 + §7 → `README_where_we_are.md`**.

---

## 1. The invisible-text incident — RESOLVED, verified at the artefact 2026-08-25

What happened (short form; full record now in
**`bugs_closed/198_…css_patch_agent_fragment_clobbers_whole_stylesheet.md`**): on
08-17 the css-patch-agent deployed its near-empty `css_themes` row over idea.uk's real
23,650-byte stylesheet, deleting every `:root` colour variable — most text went
black-on-black / white-on-white. Six sites took the same wave; the root cause was the
DB row and the deployed file having diverged fleet-wide (the row often BORN empty —
ultimately an INSTALL-contract defect, per the owning lane's final diagnosis).

- **This lane's repair (08-19) worked exactly as designed:** `css_themes` restored to
  v6 from vm-sites `8c407a18f` + the four legitimate patches; canary item `01a4dbca`
  (a real parked finding, unparked on its own expired condition) was **claimed 27
  minutes later and completed first attempt**, deploying the full restored row.
- **The owning lane (vigilant_designer/198) then finished the class:** a third wave
  caught and restored, **fleet-wide empty-row backfill DONE**, a `stylesheet_gutted`
  detector built (fleet gate 22/0 as of 2026-08-23), deploy-side guard **DGH-016 LIVE
  on v1.0.1323**, council-approved, owner ruling closed the shared-theme case. 198
  moved to `bugs_closed/`.
- `[VERIFIED 2026-08-25]` at the served page: `/assets/css/styles.css` = **26,264
  bytes, `:root` present (4)**; `css_themes` now v15 / 26,066 chars (nine further
  legitimate appends since v6, last 08-23) — DB and file are the same document now,
  which is the fixed state, not drift. D-005's blessed clause still served; favicon
  still 200.

**Nothing remains to do on this incident from this lane.**

## 2. Queue state `[MEASURED 2026-08-25]` — and a reading trap that moved under us

| status | count | note |
|---|---|---|
| needs_human_review | **49** | the owner's queue, and GROWING — fresh audit activity 08-20→08-24 added `decision_blocked_change` (12, newest 08-23), `cta_names_unknown_destination` (6), `content_rewrite` (4), `claims_unverified`, `cta_tel_malformed`, `brief_supplies_negation` |
| deferred | 33 | 16 contrast (down from 23 — the fleet worked some through the healed pipeline), 12 `undeployed_asset` (07-31), 3 `capability_gap` |
| detected | 31 | unchanged: the handler-less `head_essentials_missing` family — `bugs_open/083`'s lane owns the promoter side; watch, don't push |
| failed | **3** | ALL fresh (08-24): 2 `empty_section`, 1 `page_rerender` — errors unread, see §4 |
| triaged/approved/claimed | 0 | nothing waiting on a turn |

⚠ **Do not compare `complete` totals across sessions**: the 08-18 handoff said 262
complete; today reads 78. Nothing was lost — `site_work_items` is a **rolling window**
(the archiver moves closed rows out, ~7 days). Lifetime claims need the history UNION
(the memory note `a-closer-census-cannot-see-what-it-succeeded-at`).

## 3. Standing rulings — unchanged, re-verified where cheap

- **Honesty arc CLOSED** (owner 08-17, migration 454); class C retired; do not touch
  the writer prompt's anti-fabrication rules (08-16 handoff banner).
- **D-005** guards the report hero; `honest assessment` served today.
- **`whether you're` stays** — built-in tell, not an owner rule.
- **The voice gate still cannot see `pages.title`/`pages.meta_description`** (08-16
  §4) — unchanged and still the trap to remember if heads are ever edited.

## 4. WHAT'S NEXT — in value order, as of 2026-08-25

1. **news at `/data/latest-news.json` still 404; `content_sources` for idea.uk still
   0.** Untouched since 08-04/05 and now the oldest live gap by three weeks. The
   dispatch mystery is §X.53 in RUNNING_NOTES. Start here: why did the 08-04/05
   content-source seeding never land, and what does the framework need to file to
   populate it? (Fleet count was 49 sources on 08-18 — re-measure before citing.)
2. **Read the 3 fresh `failed` errors (08-24)** — 2 `empty_section`, 1 `page_rerender`.
   One query: `SELECT item_type, left(error,200), attempt_count FROM site_work_items
   wi JOIN sites s ON s.id=wi.site_id WHERE s.domain='idea.uk' AND wi.status='failed';`
   Route findings; don't fix blind.
3. **Class B is STILL not filed** (8 components, 3 sites, `content_data` NULL —
   flagged unfiled in both the 08-16 and 08-18 handoffs). Either file it (cross-site
   ⇒ 090 first, per the 07-31 ruling) or write down why it doesn't merit a case.
   Stop carrying it forward unfiled.
4. **The owner's review queue (49) is worth a triage pass WITH the owner** — 12
   `decision_blocked_change` since 08-23 suggests something is repeatedly bumping
   into locked/decision-protected content; a pattern read may collapse many rows
   into one cause.
5. Older residuals, unchanged: first organic signed Stripe webhook; tools-page card
   images and tool-page heroes; the empty-kind → SDXL image-routing hole; ingress
   landmines (`ufw allow 80,443` FIRST).

## 5. Pointers out of this lane

- **Queue timing questions** → `dispatch_throughput/` is now its own ACTIVE lane:
  **N=2 concurrent dispatch is LIVE** (migrations 582–584, 08-24) and the owner's
  rulings D0a–D16 are recorded there. Do not re-derive drain-rate arithmetic here;
  their STARTER/PLAN/README are current.
- **CSS/theme anything** → `bugs_closed/198` + the two adjacent LANDMINES
  (`css_themes` footprint). The detector (`stylesheet_gutted`) and guard (DGH-016)
  are live; a recurrence should be caught fleet-wide, but the artefact check is
  still one line: `curl -s https://idea.uk/assets/css/styles.css | grep -c ':root'`.

## 6. Traps carried forward

- The 08-16 §7 set still stands (queue depth ≠ prediction; head-of-line ≠ proof;
  chassis provenance line scrolls; D-005 guard timing; `replace()` for special chars).
- `attempt_count = 0` = never tried (queue position, not failure).
- The rolling-window trap in §2 — new this handoff, and it WILL bite a totals
  comparison again.
- A completion count is not an artefact check: two of the 08-17 "42 completions" were
  the stylesheet clobber commits. Spot-check the artefact behind any batch of
  completions you are about to call healthy.
