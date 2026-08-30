# agritec.uk — continue here (updated 2026-08-30, evening)

Cold start for a fresh session. Read this, then `SUBJECT_LEDGER.md`, then `NOTES_agritec_uk.md`
from the bottom. Background and the fuller mechanism maps are in
`HANDOFF_2026-08-25_continue_here.md` (§5's five-seam palette map and §6's traps still hold).
Every figure below was measured on the date stated next to it, not remembered.

---

## 1. What this lane is

Rebuilding `agritec.uk` — a hand-built static site of calculators and technical articles —
inside the framework (owner ruling 2026-08-04). `SUBJECT_LEDGER.md` is the completeness
contract: 26 original pages, depth floors, gate results.

## 2. State — measured 2026-08-30 at the SERVED ARTEFACT

**Everything on the 08-26 dispatch list landed and is verified at the served site:**

- **Light palette is complete INCLUDING imagery.** All 17 assets regenerated under the light
  guide and deployed 08-27 11:01. Measured, not eyeballed: mean luminance 224–236/255 on
  sampled hero/logo/icon, dark-pixel fraction ≤5%.
- **The agreement-cap garble is FIXED on the served page** — "100,000 agreements" gone; the
  copy now states the registered £100,000 per-agreement value cap correctly
  (`CIT-86c4010f7cdf820d`).
- **The SFI26 tool is intact** through four days of waves: £224 in, £382 out, rate table
  present, no dark hex. **GTM is back** in the head (bugs_open/397's wave). **Guides hub
  6/6.** Favicon serves at the URL the head names (`/assets/images/favicon.png`); the bare
  `/favicon.png` 404 is browser-fallback noise, not a defect.
- **The 24-fact reconciliation is DONE** (08-26): register vs tool data array, 24/24 match,
  items closed with rulings (`SEED_2026-08-26b`). The stale_evidence and citation_unverified
  reviews are closed with rulings in the same seed.
- **The companion guide is KEPT — the 08-25 "retire the stub" item is DISSOLVED by
  measurement** (08-30): it grew into a ~6,000-word methods-companion, hub-listed,
  cross-linking the explainer and the tool, stating zero £ figures (so the cite-every-figure
  ruling is satisfied vacuously and deliberately). Do not retire it; there is no
  stub-pointing CTA anywhere to repoint.

## 3. ⚠ BLOCKED: the kubeconfig token expired 2026-08-27 19:11:20Z

Every kubectl/DB read fails `Unauthorized` until the owner refreshes (3-day cycle; the
dispatch_throughput lane notified the owner, commit `68f4fd1bd`). **First action once it
refreshes — read these, and stamp each read with its honest date:**

1. **The `acceptance_run` result** for `tool-sfi26-revenue-stacker` (filed by discovery
   2026-08-26 00:24, the FIRST Tier-4 run this site has ever had). The artefact looks right;
   the run's verdict on the ARITHMETIC is the one thing still unproven. If it never ran,
   RUNBOOK §14 (retry backoff — ask the row).
2. Statuses of the 17 `needs_imagery` (batch `af3e9ffa-…`) and the `content_rewrite` — expect
   complete; glance at `result`/`__step_error` anyway (a `complete` item is not the artefact).
3. Did `claims_unverified` self-close (revalidator arm `resolved_all_gates_passed`)?
4. The open queue generally — four days of discovery/improvement waves are unread. The
   rotation now visits ~1 site/3h (re-enabled 08-26), so expect routine findings.

## 4. Open items, in priority order

1. **Read the blocked results above** (§3) the moment the token refreshes.
2. **The five remaining Phase 1 calculators** (vertical energy, VPD, nutrient dosing, BSF
   converter, blue carbon — ledger T1–T5). T1 evidence is part-registered; specs not written.
   The SFI stacker's build pattern (scoped brief, fence, artifact_checks, acceptance) is the
   template; its guard rails are all live and proven now.
3. **Phase 2 of the ledger:** the IoT cluster — 7 tools, 7 deep dives.
4. **Later:** news, editorial, directory (directory needs a new `directory_entities` kind).

## 5. Standing cautions (unchanged, still live)

- `unresolved_cta` ×9 are BENIGN (gated template renders no button) and self-resolve as tools
  get built — leave them.
- `chrome_divergence_overwritten` belongs to bugs_open/397 (analytics_gtm lane) — theirs.
- The `undeployed_asset` check disagrees with the artefact (assets serve 200) — its meter
  reads a stamp, not the URL; framework-owned, do not "fix" the site for it.
- Two discovery carriers visit this site (improvement-loop children AND the fair rotation) —
  same-shape items from both is expected, not double-filing; item_key dedup handles it.
- All of `HANDOFF_2026-08-25` §6 (derived `pages.sections`, rerender-vs-rebuild, retry
  backoffs, no backticks in commit messages…) and the DO-NOT-TIDY list in its §4.

## 6. Owner decisions on record

D1–D8 in `PLAN_2026-08-21_agritec_uk.md`; cite-every-figure (08-24); light palette as default
(08-25); no cannabis content; no unsourced figure anywhere; every site through the framework.
