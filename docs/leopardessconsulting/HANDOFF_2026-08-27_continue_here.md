# HANDOFF — leopardessconsulting.co.uk, 2026-08-27

**Start a fresh session from exactly here. Supersedes `HANDOFF_2026-08-25b_continue_here.md`**
(all of its queue is done or overtaken except items carried below). Everything
`[MEASURED 2026-08-27]` was checked first-hand that morning.

**Site:** `leopardessconsulting.co.uk` · `site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Session records:** RUNNING_NOTES 08-25→08-27 · owner log README_where_we_are same dates ·
milestone `SUMMARY_2026-08-27_the_critic_exists_and_every_page_has_a_face.md`

## 1. State `[MEASURED 2026-08-27]`

- **Heroes: 17 pages serve their own; 5 still on generic `hero.jpg`** (faq, privacy, terms,
  blog/hierarchical-…-explained, guides/tool-agent-complexity-estimator-guide) — the archetype
  remainder, per the owner's D2 ruling. Every generated image was EYEBALLED; 4 rejected +
  re-rolled (near-duplicates ×2; stray décor ×2 — add the exclusion clause to prompts).
- **The critic (018 Phase 1) is LIVE end to end**: grant `04c49f8f0` (council APPROVED
  `30d5fdde`), seed mig `645` (Council-Submitted `75be8d32`), register **SQ-003**. First run
  complete: spawned pod carried IMAGE_BUCKET (grant proven), first `design-report` doc_note
  filed — concrete findings, palette read correctly, lead finding = the hero sameness (an
  organic "before" leg). **Dispatch ONLY via `scripts/design_critique_run.sh`** — a generic
  orchestrate runs it INLINE with no storage client (the 243 trap; seed 645's header is wrong
  on this and cannot be edited — ledger checksum).
- **After-run in queue at session close: item `b36f8c63`** (leg `after_hero_batch`).
- **8 rows locked** (`locked_by='leopardess-403-restore'`): services ×3, hero-about,
  hero-contact, index call-to-action + stat-band + evidence-chart. Row-count baselines for the
  rerender-on-locked landmine: about 4, contact 5, index 6, services 4 — expect UNCHANGED
  after any rerender; +1 = the 189 duplicate trap.
- **403**: filed, twice corrected, design RULED (`__authored` spec in the file, shared with the
  395 lane), Go NOT started. **248**: still open with the fresh 08-26 instance + the
  architecture-label-match hypothesis contributed; their lane owns it.
- Mig `649` (finetuning lane, owner-directed) gave `case-studies-hero` + `hero-tool` image
  branches — case-studies' hero only rendered after a post-649 rerender (283 §13: a template
  edited by SQL ships nothing until a re-render). `hero-tool` means
  tool-automation-savings-estimator can take a hero in the archetype pass.

## 2. Work queue, in order

1. ~~Read the after-run report~~ **PARKED 2026-08-27 after four attempts — the vision half
   cannot run until `execute_vision_prompt` learns to DOWNSCALE**: post-hero full-page
   captures exceed BOTH providers' per-image limits (Anthropic's stated cap: 8,000px on a
   dimension; error ladder in RUNNING_NOTES same date). **New queue item 1: build the
   downscale/tile in `execute_vision_prompt`** (Go, council; then restore `max_images` 16 —
   migs 662/663 are the interim config, snapshots taken) and re-run the after leg; the
   discrimination test then runs in mutation form. Original item kept for context:
   **Read the after-run report** (`SELECT body FROM doc_notes WHERE categories ? 'design-report'
   AND site_id='4851f6fc…' ORDER BY created_at DESC LIMIT 1` once item `b36f8c63` completes).
   **The discrimination read: the sameness lead finding must be GONE, the rest substantially
   stable.** Record the result in NOTES + 018/SQ-003. If it still reports sameness, check WHICH
   pages its audit sampled before concluding anything (8 of 37, its choice).
2. ~~Two archetype heroes~~ **DONE 2026-09-02** — archetype-content (faq/privacy/terms),
   archetype-blog (both articles), gauge hero for the estimator tool; all eyeballed, merged,
   published; census watcher confirms zero generic hero.jpg site-wide.
3. **The critic's 8 findings** — concrete, actionable, several overlap deferred A2 pt 2
   (services carousel title hierarchy). Owner-visible quality work; pick order by effort.
4. **403 field-level fix** (`__authored`): design is ruled IN THE BUG FILE (read it first —
   spec, home file, enforcement point all pinned). Go + tests + council + register. The 395
   lane builds the column-side instance independently.
5. ~~Rotation watch~~ **CLOSED 2026-09-02, producers NAMED**: misdirected_cta rerenders
   COMPLETED vs services (08-31, 09-01) and index (08-31); every locked row and baseline held.
   The locks beat the named clobberer twice, organically.
6. Carousel A2 pt 2, after 3.

## 3. Standing cautions

All of `HANDOFF_2026-08-25b` §2 (locks: publish-before-lock; unlock→edit→publish→re-lock ·
llm-field hand-edits are loans · enumerating known guards ≠ searching all guards) plus:
- **priority sorts ASC in the dispatch loader** — lower number dispatches first.
- **Gate-check EVERY page before any rerender reaches it**, including pages OTHER work queues
  at yours (the image-landed rerenders the 384-fix files for listing pages).
- The queue is deep fleet-wide (~1,100+ triaged, ~270 claims/h measured 08-26) — a triaged
  item sitting ≥15 min is QUEUED, not stuck; check `priority ASC` position before poking it.
