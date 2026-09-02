# HANDOFF — leopardessconsulting.co.uk, 2026-09-02 (evening)

**Start a fresh session from exactly here. Supersedes `HANDOFF_2026-08-27_continue_here.md`.**
Everything `[MEASURED 2026-09-02]` was checked first-hand that day, including the evening
build probe. Site: `leopardessconsulting.co.uk` · `site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732`.
Records: RUNNING_NOTES through 09-02 evening · README_where_we_are same · milestone
`SUMMARY_2026-08-27_the_critic_exists_and_every_page_has_a_face.md`.

## 1. State `[MEASURED 2026-09-02]`

- **Imagery: DONE, census zero.** No page serves generic `hero.jpg` (sitemap-wide sweep,
  positive control passed). 18 per-page heroes + `archetype-content` (faq/privacy/terms) +
  `archetype-blog` (2 articles) + the estimator's gauge hero. Every image eyeballed.
  ⚠ Hero wirings are 403-class loans (`background_image` re-resolves on a full rebuild) —
  durable protection = the `__authored` work (item 3 below).
- **Locks: PROVEN against the named clobberer.** `misdirected_cta` rerenders COMPLETED vs
  services (08-31 23:24, 09-01 14:53) and index (08-31 23:24); all 8 locked rows
  (`locked_by='leopardess-403-restore'`) and baselines held (about 4/1, contact 5/1,
  index 6/3, services 4/3; services 6 cards; home CTA `/contact.html`).
- **THE DOWNSCALE IS LIVE** on the fresh chassis (replicaset `744cfb4bf`, 09-02 ~15:40Z):
  probe `downscaleVisionImage: image scaled for provider limits` PRESENT +
  `max_image_dimension` PRESENT + positive control PRESENT + invented string absent.
  `execute_vision_prompt` now scales any image whose long edge exceeds
  `max_image_dimension` (default 7900) to fit, JPEG q85; legal images byte-identical
  (pinned by test); `max_image_dimension: 0` opts out (pinned).
- **Critic (018 Phase 1)**: LIVE; ONE production report exists (2026-08-26 15:14,
  `doc_notes` categories `['design-report']`) — the "before" leg, lead finding = hero
  sameness. Interim config from the blocked era still applied: provider anthropic/
  claude-sonnet-5 (mig 662) + `max_images: 8` (mig 663, snapshots taken).
  **Dispatch ONLY via `scripts/design_critique_run.sh <site> <domain> [leg]`** — the
  generic topic runs it inline with NO storage client (243 trap; seed 645's header is
  wrong on this, correction in register SQ-003).
- **Council, 4 verdicts pending** `[as of 09-02 evening]`: `e5a664d9` (the downscale) +
  resubmitted rounds 2 on `52c9a201` (662), `c6046171` (663), `75be8d32` (645).
  Round-1 lesson already logged: a sketch must carry the file's text verbatim.
- **403** (`bugs_open/403_...authored_values...`): design RULED in the file (`__authored`
  spec: key/shape/enforcement/home all pinned; the 395 lane builds the column-side
  independently; their no-writer caveat recorded). Go NOT started.
- **248** open (their lane); classification spec is the legacy 04-18 shape (RUNNING_NOTES
  09-02: guard before any needs_composition dispatch here).

## 2. Work queue, in order

1. **Finish the critic chain (everything is unblocked):**
   a. Migration `701_design_critique_restores_max_images_16.sql` (or next free —
      RE-CHECK the number, it moved twice under me; grep the dir first): snapshot_agent,
      jsonb_set `{workflow,steps,critique,config,max_images}` → 16, DO/RAISE guard.
      Council scope; apply SCOPED (`MIGRATIONS_DIR=<tmp>` — the shared dir carries other
      lanes' pending files; `--apply` takes all).
   b. `./docs/leopardessconsulting/scripts/design_critique_run.sh 4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk after_hero_batch_downscaled`
      — then bump the item to priority 70 if queued behind an 80-batch (**the loader sorts
      priority ASC** — lower first).
   c. On completion: `images_downscaled >= 1` in the run's output = the wiring's first live
      proof (0 on captures known too tall = dead wiring, STOP and read the pod log).
      Then the report: `SELECT body FROM doc_notes WHERE categories ? 'design-report' AND
      site_id='4851f6fc…' ORDER BY created_at DESC LIMIT 1` (expect a SECOND note).
      **Discrimination read: the hero-sameness lead finding must be GONE, the rest
      substantially stable.** If sameness persists, check WHICH 8 of 40 pages the audit
      sampled before concluding. Record in NOTES + SQ-003 (its verify-later names this).
2. **The critic's 8 findings** (the 08-26 report; several overlap deferred carousel A2 pt 2
   — services carousel title hierarchy). Owner-visible quality work.
3. **403 `__authored` build**: read the bug file FIRST — spec is pinned there
   (key `__authored`, field→true, whole-field; enforcement at save_page_sections/
   planSection; home `datahelpers/authored_provenance.go`; columns share the convention
   not the storage). Go + tests (mutate to prove the guard) + council + register. This is
   what converts every hero wiring and hand-repair from loan to owned.
4. **Carousel A2 pt 2**, after 2.
5. Housekeeping: check the 4 council verdicts (REVISE → fix forward on the same
   correlation); 098 credits approved trailers automatically.

## 3. Standing cautions (all previously bitten)

- **Dispatch loader sorts `priority ASC`** — lower number first. A triaged item ≥15 min is
  usually QUEUED (fleet ~1,100+ triaged, ~270 claims/h measured 08-26) — check position.
- **Gate-check EVERY page before any rerender reaches it** (both branches, worked SQL in
  RUNNING_NOTES / L9 §4) — including pages OTHER queues aim at yours (image-landed
  rerenders for listing pages).
- **Locks: publish BEFORE locking; unlock → edit → publish → re-lock.** Row-count check
  after any rerender on a locked page: counts above; +1 = the 189 duplicate trap.
- **A hand-edit to a `source:"llm"` field survives rerenders and dies at the first
  rewrite** — LANDMINES has the entry; keep repairs as re-runnable SQL until 403 ships.
- **Probe capabilities, not commits/tags** — positive AND negative control in the same
  breath (worked examples: NOTES 08-26 + 09-02).
- **A sketch in a council submission must be the file's text verbatim** — abbreviations
  earned two REVISEs.
- **Migration numbers move under you** (650 was taken between ls and write; dir was at 66x
  on 09-02) — re-check, and NEVER edit an applied file (ledger checksum): supersede or
  correct in the register.
- Backups this arc: `bak_leo_*_2026082[5-7]`, `bak_leo_archetype_hero_pc_20260902`;
  snapshots via `snapshot_agent` for migs 662/663.
