# HANDOFF — leopardessconsulting.co.uk, 2026-08-25b (evening)

**Start a fresh session from exactly here. Supersedes `HANDOFF_2026-08-25_continue_here.md`**
(the morning file — its four owner decisions are ANSWERED, its D3 premise was stale when
written, and its §4/§6 are overtaken by the work below). Everything marked
`[MEASURED 2026-08-25]` was checked first-hand this evening.

**Site:** `leopardessconsulting.co.uk` · `site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Approved plan:** `~/.claude/plans/let-s-do-1-2-and-3-ancient-crab.md` (Part B now in build)
**Session record:** RUNNING_NOTES 2026-08-25 evening entries; owner log README_where_we_are same date.

---

## 1. The four owner rulings (2026-08-25, this session — all four decisions are TAKEN)

| ruling | state |
|---|---|
| **D1** — drop the `trust` voice rule entirely | **DONE + verified at served pages** |
| **D2** — per-page heroes for the dozen that matter, archetypes for the rest | 3 of the dozen shipped this session (about, services, contact); 7 pages now have per-page heroes |
| **D3** — this lane takes the platform fix for the regeneration damage | `bugs_open/403` filed + corrected; services RESTORED under the existing row LOCK; field-level fix still owed |
| **D4** — build the design critic in this lane | Storage grant committed `04c49f8f0` (Council-Submitted `30d5fdde`); seed NOT yet written |

## 2. Read these before running anything (new traps found this session)

1. **`page_components` row LOCKS exist and are live** — `lock_type='permanent'` +
   `locked_at=now()` + `locked_by`; `save_page_sections` excludes locked rows from its DELETE
   (`loadActiveRows`/`matchLockedRow`; predicate `datahelpers.AgentWritableSQLFor`). 51 rows /
   7 lanes used them before us. **Publish BEFORE you lock**; future edits to a locked slot:
   unlock → edit → publish → re-lock. Five leopardess rows are now locked under
   `locked_by='leopardess-403-restore'` (services info-card-grid, teaser-reveal-panel,
   hero-services; hero-about; hero-contact).
2. **A hand-edit to a `source:"llm"` field survives every RERENDER and dies at the first
   REWRITE** (content_rewrite/tone_shift/build) — that survival is the trap. LANDMINES has the
   full entry; `bugs_open/403` is the case file. The morning handoff's §4 "re-derive the
   icon↔item mapping" was unnecessary: restore wholesale from the `page_component_history`
   archive taken AT the destroying write (it holds the good state verbatim).
3. **An enumeration of KNOWN guards is not a search for ALL guards** — 403 was filed saying
   "no guard covers this" while the row lock existed. `grep -in 'lock' <the action>` first
   (WRONG_CALLS 2026-08-25).
4. The morning handoff's §1 items all still hold (unstable pc ids; updated_at bumps on no-ops;
   byte comparisons; capability probes; `$STAMP` guard; `site_specs.site_plan` is the section
   authority).

## 3. State `[MEASURED 2026-08-25 evening]`

- **Voice:** trust rule dropped (`bak_leo_voice_20260825`); "honest"/"earns its keep" copy
  cleared via `scripts/VOICE_2026-08-25_banned_phrases_v2.sql` (the 08-17 file refused itself
  on two expiries — see its header); use-cases/how-it-works/insights rerendered; served sweep
  0/0/0 with positive control 1.
- **Heroes live per-page (7):** index, how-we-work, use-cases, who-we-help, + NEW about
  (concentric rings), services (diverging paths→circle/square/triangle), contact (meeting
  arcs). All three new ones eyeballed per the site rule; the first services/contact images
  were REJECTED as near-duplicates and re-rolled (same asset_key overwrites).
- **`/services.html` fully restored (fourth restore, first under protection):** 6 cards,
  6 teaser items, 6 icon refs + its own hero, verified at the served page; content slots +
  hero LOCKED. CTA slot deliberately NOT restored (it carries the legitimate 08-25
  `__cta_minted` machine rewrite).
- **Home:** 3 × `/contact.html` CTAs still hold.
- **`bugs_open/403`** (authored values inside llm-owned fields): filed with 4 traced instances,
  corrected same-session re the lock; **090 run `c946b495` still iterating at session close —
  READ ITS VERDICT and record it in 403** (`SELECT ... FROM orchestration_states WHERE
  correlation_id::text LIKE 'c946b495%'`; artifacts under that corr).
- **Critic (D4):** `design-critique-agent` granted storage in `isStorageEnabledAgent`
  (`04c49f8f0`, Council-Submitted `30d5fdde` — check the verdict:
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`).
  Vigilant lane TOLD via CONTRIB in their dir. **The grant is INERT until a fleet build rolls.
  The seed must follow the build, never precede it.**

## 4. Work queue, in order

1. ~~**Read the 090 verdict** → record in 403.~~ **DONE, same evening after this file was cut:**
   verdict UNVERIFIABLE (scope-not-narrowing); it challenged the attribution as timing
   correlation and the re-check CONFIRMED it from `llm_call_log`'s verbatim generate_content
   replies (403 §"The 090 verdict" has the full exchange). Nothing left here.
2. **Council verdict on `30d5fdde`** → 098 credits the commit automatically on approval;
   REVISE gets fixed forward.
3. **Hero batch — the remaining 9 of the dozen** (D2 answered; canary proven ×3):
   how-it-works, tools, insights, engagement-model, technical-architecture,
   ai-readiness-quiz, careers, + blog `why-most-ai-agent-projects-never-reach-production`,
   `can-you-trust-ai-with-your-data`. Route A per hero: dispatch scope-less `needs_imagery`
   (worked payloads in RUNNING_NOTES/HANDOFF.md ~504), EYEBALL each, merge
   `background_image`, gate-check, safe-rerender, verify served. Then archetype heroes for
   the remainder + **redeploy `hero_case_studies`** (asset stuck on a presigned URL — RUNBOOK
   O5 §4 defect).
4. **The critic seed** (after the next fleet roll ships the grant — verify at the binary:
   capability probe with both controls): Phase 2 workflow shape in the approved plan Part B;
   sql_for_agents migration = council scope; register the agent in the concept register in
   the same commit; the before/after discrimination control needs redesigning (the plan's A0
   control expired when A0 shipped).
5. **403 field-level fix design** (the row lock is too coarse where automation should keep
   improving unlocked fields): candidate 1 in 403 — informed by the 090 verdict. Council +
   register when built.
6. **Carousel (A2 part 2)** — cards now have real images only on services; revisit once the
   hero batch lands.

## 5. Commands (unchanged from the morning handoff §7, plus)

```bash
# lock / unlock a row (publish BEFORE locking)
UPDATE page_components SET locked_at=now(), locked_by='<lane>', lock_type='permanent' WHERE id='…';
UPDATE page_components SET locked_at=NULL, locked_by=NULL, lock_type=NULL WHERE id='…';

# check what the diagnosis loop said
SELECT current_step, status FROM orchestration_states WHERE correlation_id::text LIKE 'c946b495%';
```

**Backups this session:** `bak_leo_voice_20260825`, `bak_leo_portfolio_voice_20260825`,
`bak_leo_insights_pc_20260825`, `bak_leo_about_hero_pc_20260825`,
`bak_leo_svc_contact_hero_pc_20260825`, `bak_leo_services_content_pc_20260825`.

---

## Addendum 2026-08-26 (next morning)

- **The design-discovery rotation restarted 09:20Z** (cross-session heads-up from
  webdesign-tool-rebuilds; `bugs_open/401` covers why the 08-11 pause was never surfaced).
  Leopardess's visit lands within ~2-3 days; findings auto-promote to dispatch. Surprise
  design items = the rotation. **After the visit: check every locked slot survived and record
  the producer it survived** — that is the locks' behavioural proof arriving organically.
- **Home CTA was clobbered a third time overnight** (`misdirected_cta:index` completed
  02:03:58Z) — re-authored (4th), published, verified served, and the row is now **LOCKED**,
  along with `stat-band` and `evidence-chart` (8 locked rows site-wide,
  `locked_by='leopardess-403-restore'`). CONTRIB with the pre-write stamp evidence + the
  label-match hypothesis appended to `bugs_open/248` — including that the lock CONFOUNDS their
  re-author discriminator; coordinate before unlocking.
- Hero on index deliberately left unlocked (its keep holds; preserves 248's natural surface).
- **Critic dispatch, corrected same day:** `orchestrate_safe.sh design-critique-agent` runs it INLINE
  on the shared chassis → `no storage client` (`complete_no_critique`, corr `95f6b328`). Use
  **`scripts/design_critique_run.sh <site_id> <domain> [leg]`** — a `design_critique_run` work item
  the dispatch loop SPAWNS (the 243 lesson, already written in `tool_acceptance_run.sh`'s header).
  Seed 645's header is wrong on this and cannot be edited (ledger checksum); SQ-003 carries the
  correction. First spawned run: item `4f1fb87b`, leg `before_hero_batch`.
