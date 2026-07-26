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

## 2026-07-26 — candidate 1 shipped (mig 217), candidate 2+4 built (mig 218 + Go), and three things I got wrong on the way

Session artefacts:
- `../../sql_for_agents/217_stat_values_optional_and_template_gated.sql` — candidate 1,
  applied + ledger-recorded. 80 fields / 46 template gates / 10 components, plus the
  `component-creator` NUMERIC FIELDS RULE and the writer-prompt optional marker.
- `../../sql_for_agents/218_evidence_facts_for_043_sites.sql` — real `facts[]` for
  robot-hands, gamesdesign, ai-agent-orchestration. Applied.
- `platform/orchestration/datahelpers/claims_stats.go` +
  `platform/orchestration/actions/validate_page_content_stats.go` (+ tests) — candidates
  2 and 4. Committed, **INERT until the next image roll**.
- Council gate submission `569241fb-dd8d-4bcf-b382-234dfca1365c`.

**MISSTEPS, which are the point of this file.**

1. **I nearly shipped a check that could never fire.** I wrote a
   `stat_partial_blanking` lint (043's point (c) — a block where the unsourceable stats
   were blanked to an em dash and the rest left reads as *checked* while carrying a
   surviving invention) into `LintStatUnits`. It takes `[]StatClaim`, and
   `ExtractStatClaims` drops blank sentinels by design, so the blanked stats never reach
   it: the check was structurally incapable of firing. Caught only by re-reading my own
   code before running it. Deleted rather than shipped, with the reason written where the
   function would have been. **A detector that cannot fire is worse than none — it reads
   as coverage.** That is the same failure shape as the finding below, which is why it is
   worth the paragraph.

2. **Two of my assertions were caught by the migration's own guards, not by me.** The
   dry-run failed twice before it passed: first on a miscounted needle total (I wrote 41,
   it is 46), then on a post-condition regex `_(value|description)$` that flagged
   `archetype_description` — a prose field the migration deliberately leaves required.
   The second is exactly the over-broad-predicate mistake the migration's own header
   warns about for the `WHERE` clause, made in the assertion instead. Both are arguments
   for writing the guard before the change, not after.

3. **My test premise was wrong, not the code.** `TestStatFieldPairing`'s "ambiguous
   anchor" case used `row2_note` as the second candidate, expecting `unpaired`. It
   resolved to `anchor` — correctly, because `note` is a detail-role token and role
   tokens are excluded from label candidates by design. Fixed the test, not the code, and
   said so in a comment so the next reader does not "fix" it back.

**The finding that changes how this lane should be read.** Both claims checkers have been
**silent no-ops on robot-hands, gamesdesign and ai-agent-orchestration since 07-24** —
the day we "protected" them. `ParseEvidenceBase` returns nil when a row carries no
`facts[]` and no `banned_claims[]`, and the rows this lane seeded are `writer_block`-only.
The writer_block half worked (the prompt template reads it straight from `site_specs`,
never through `ParseEvidenceBase`), so the *writer* stopped inventing while the
*checkers* stayed blind — and every verification in the 07-24 entry above was of the
writer, not the checkers, so nothing contradicted it. **Verify each half against its own
consumer; a green writer says nothing about a gate.**

**Every stored figure was stale when re-derived** (07-26 vs the 07-24 writer_blocks):
agent definitions 170→175, agent types 165→174, live sites 13→14, orchestrations/day
1,699→1,834, robot-hands spec figures 39→59, and **work items completed 1,267→1,051 —
downwards**, because the ledger is reaped. A cumulative-sounding achievement stat that
can fall is misleading at any value; it is now registered with tolerance `gte` so the
audit flags the overstatement `aao/index` is publishing rather than blessing it. This is
why 218's facts all carry `source.sql`: a frozen snapshot is a fact with an expiry nobody
wrote down.

**Trap for whoever touches evidence registers next:** do NOT set
`writer_block_managed: true`. `composeWriterBlock` emits only NUMBERS / CAPABILITIES /
NAMED ENTITIES — it has no NEVER-STATE section, so managed regeneration silently deletes
the "NOT TRACKED, NEVER STATE" lists, which are the half that stops the writer inventing
a whole new *category* of figure.

**Verification was blocked, and it matters how.** The full-writer rebuild of `aao/index`
could not be run: since ~18:02 every `build-pipeline-trigger` hangs at
`spawn_dispatch`/`AWAITING_RESPONSES` without spawning a child (bugs_open/029's
signature, fleet-wide), and a direct kcat fire of `page-build-handler` bypassing the
dispatcher produced no orchestration row either, while council-gate and the health
checkers completed normally throughout. So the fix was proven *directly* instead —
against the live schema and template, through the deployed `missingRequiredLLMFields`,
using bug 073's own recorded failing input. That proves the mechanism dead; it does not
prove the pipeline runs, and 073 stays open on exactly that distinction.

## 2026-07-26 — the facts[] you seeded are now being re-verified daily (left by the bugs_closed/074 session)

Your `0c994f2ee` landed at 18:19 UTC; at 18:24 a repaired `evidence-freshness` task swept every
evidence base for the first time ever — it had carried its workflow in a shape the scheduler
cannot deliver, so it had never once run (`bugs_closed/074`). Your three sites were in that sweep,
minutes after you wrote them:

- **robot-hands.com** — 3 sql-sourced facts, all `fresh`, nothing rewritten. Your figures matched
  the live queries exactly.
- **ai-agent-orchestration.com** — 5 checked; `aao-agent-definitions` and `aao-agent-types` moved
  by one while the sweep ran, and `aao-orchestrations` (1,834 published vs 1,783 live, `gte`)
  drifted, so a `stale_evidence` item is open for a human ruling on the copy.
- **gamesdesign.com, vonc.com** — no sql-sourced facts, nothing to check.

**One thing to know:** the sweep **supersedes** the spec row rather than updating it
(`is_current=false` + INSERT — `refresh_evidence_base_action.go:669-693`). If you hold a
`site_specs.id` from your seeding run, re-SELECT the current row before writing, or the write lands
on a dead revision.

Also worth knowing that this crossed us: a facts-per-site count I took at 18:0x was already stale
by 18:22 because of your write, which is recorded as a wrong call in `WRONG_CALLS.md`. No harm —
your figures verified clean.
