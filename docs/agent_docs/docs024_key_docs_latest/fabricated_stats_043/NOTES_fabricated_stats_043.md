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

## 2026-07-24 (evening) — wave-2d: the poisoned given, caught live

The final drain-verify showed aiagent/index CLOBBERED at 15:10 — a full-writer
re-render wrote "70+/8/30+/1000s" back over the grounded values, WITH migration
201 and the evidence_base live. Mechanism (fourth for the 043 file): the site's
own specs hardcode the figure list into the writer's "follow this closely"
content-direction — a spec number is a GIVEN and outranks every writer-side
rule. Truth audit flipped part of my wave-1 read: "over 70 agents" and the
"8 departments" taxonomy were TRUE (registry 171; departments = the platform's
own named structure — my evidence_base ban was over-broad, now narrowed to the
"departments served" misframing). The one untrue clause ("thousands of
concurrent agent instances") became the measured truth ("over a thousand
orchestrations a day" — 1,699 in the last 24h). All four aspects patched by
versioned supersede+insert; one differently-worded variant in briefing needed a
second pass (verify query caught it). Index stats restored and re-queued —
values recomputed live came back HIGHER than wave-1 (171 agents, 14 sites,
1,284 work items): the platform grew during the session, which is the whole
point of computed values.

Lessons: (1) verify EVERY page the writer touched after a full-writer pass, not
just the one you edited — about/case-study held (lightweight path), index did
not (writer path); (2) sweep the spec aspects for numeric claims — added to
043's candidate-1/2 scope.

## 2026-07-24 (close) — BEHAVIOURAL PROOF: the writer regenerated a stat block and did not fabricate

The wave-2d re-render (16:08) went through the FULL WRITER again — provably: the
persisted labels are fresh rewordings ("Agent Definitions", "Work Items
Completed"), not the stored ones. And it wrote **170 / 13 / 17 / 1,267** — the
exact dated snapshots listed in the evidence_base writer_block, rendered live.
So the first live exercise of the complete stack (de-poisoned spec → repointed
content_direction → Verified Facts → rule 14 v2) produced a freshly-generated,
fully-true stat block. Item 5 of the open list (behavioural proof of 201) is
DONE — earlier and more convincingly than the planned probe: same page, same
writer path, same afternoon as the 15:10 fabrication, opposite outcome.

Note the writer chose from the evidence list rather than echoing stored
content_data (it replaced my 171/8/14/1,284 restore with the LISTED 170/13/17/
1,267 snapshots) — exactly what the block licenses ("dated snapshots up to a
listed live count are fine"). On writer-path pages the evidence_base IS the
source of truth; keep IT current, not the content_data.
