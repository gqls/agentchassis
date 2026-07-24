# NOTES — fabricated_stats_043 (bug 043 fixing lane)

Running record, append-only, newest at the bottom. **The account of record is
`bugs_open/043_HANDOFF_2026-07-20_generated_page_copy_invents_quantitative_claims.md`**
— this file holds only what doesn't belong there (working detail, missteps).
The robot-hands-specific half of the story lives in
`../robot_hands/NOTES_robot_hands_site_fixes.md` (Turns 10–13).

---

## 2026-07-24 — sweep run, wave-1/2 fixes, migration 201, evidence_base

Session artefacts, in order:
- `SQL_2026-07-24_vonc_index_stats_de_fabricated.sql` — wave 1, applied.
- `SQL_2026-07-24_gamesdesign_index_stats_traced.sql` — wave 1, applied.
- `SQL_2026-07-24_aiagent_casestudy_stats_grounded.sql` — wave 1, applied.
- `SQL_2026-07-24_wave2_sweep_findings_fixed.sql` — sweep findings (robot-hands
  index REGRESSION, vonc about/index extras, aiagent index/about, leopardess
  clone fallbacks), applied. NOTE the two `UPDATE 0`s in its first apply: the
  vonc index components had been REPLACED WITH NEW IDs by the in-flight wave-1
  re-render — **page_components.id is not stable across re-renders; key content
  edits on (page, component, label), never on pc.id.** The residual fix went in
  by label after the render completed (wave-2b, in-session SQL, recorded here).
- `../sql_for_agents/201_content_writers_never_invent_numbers.sql` — candidate 3,
  applied + ledger-recorded (`--record-only`).
- `SQL_2026-07-24_evidence_base_four_sites.sql` — writer_blocks for the four
  fixed sites, applied; **carries an in-file CORRECTION**: the vonc seed
  superseded migration 166's structured row (experience-loop `banned_claims`)
  unread — caught on the `UPDATE 1` count, fixed with a MERGE row. WRONG_CALLS'd.

Verifications were against rendered pages, not statuses (all pass; the one
residual "70%" grep hit on aiagent was a CSS gradient stop, not a stat).

**Open when picking this lane back up** (also in 043 itself):
1. Candidate 1 — bind stat fields to provenance in component schemas; the
   `stat_N_value` llm_guidance examples ('2.4M') are the invention's seed shape.
2. Candidate 2 — post-generation numeric audit → needs_human_review.
3. finetuning.uk/about — fabricated "Clients Served 11+ / Satisfaction 100%";
   needs the owner's real story, do NOT invent a replacement.
4. Prose numbers outside stat fields were not audited (the robot-hands 9-vs-42
   ratio says stat blocks are the tip); 201 guards new writes only.
5. Behavioural proof of 201: the next routine `needs_page` full-writer render on
   any of the four seeded sites should keep the corrected stats (evidence_base
   supplies them). Probe: compare stat values before/after the next
   `render_news_section` item on robot-hands/index.

## 2026-07-24 (later) — wave-2c: the prose tail on vonc

The last grep residual ("4h 12m") led past the stat fields into ordinary prose:
fabricated countdowns, "real time"/"clock is live" theatre, and an entire
component (archetype-combinations) built on six archetypes the site does not
have. Fixed in `SQL_2026-07-24_wave2c_…` (5 UPDATEs, verify query returned 0
remaining, 3 re-renders queued). Scope line drawn and recorded in 043: the
present-tense product-VISION copy (arena guide article, conceptual
differentiators) belongs to the experience-loop/vonc-spark thread — their 166
banned_claims routes it to review by design; not touched from this lane.
