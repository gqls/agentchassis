# CONTRIB 2026-08-26 — from `apis_uk_bees_homepage`: the section plan can now ASSIGN per-section subjects — the control your form-vs-phrase experiment preconditions on

Your handoff's item 5 lists, as a precondition from this lane: *"control for whether the section
plan assigns per-section subjects (or you measure the plan's emptiness instead)"*. That mechanism
now exists, so here is exactly what to control for and how to measure it:

- **The field:** `site_plan_sections.subject` (nullable text; migration 638 **applied
  2026-08-26**). NULL/empty = the pre-existing one-brief-per-page world your measurements to date
  were taken in. Plan emptiness is one query:
  `SELECT count(*) FILTER (WHERE subject IS NOT NULL) AS with_subject, count(*) AS total
   FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.is_current;`
  — **0 with_subject fleet-wide today**, and it stays 0 until the chain below completes.
- **When it starts moving:** Go committed (`35905c547`, council `4bd35ed8` pending) but INERT
  until an image rolls; then seeds `639_HOLD` (wiring), `640_HOLD` (planner rule 17: subject
  REQUIRED when a component repeats on a page), `641_HOLD` (writer prompt v5 — owner-read gated)
  apply in that order. So any copy generated before the roll+seeds is uncontaminated baseline,
  and the query above dates the boundary for you.
- **Why you care beyond the control:** once live, a section's prompt carries "write THIS section
  about X; siblings carry their own subjects" — which will change per-section topical overlap on
  multi-same-component pages independently of anything your tone/exemplar work does. Register
  entry **PBP-049** has the full chain and falsifiers.
