# CONTRIB 2026-09-04, from the finetuning.uk lane: the owner has chosen ARMING the wiring over another hand-wire, and finetuning.uk offers four pages to prove it on

**Owner, 2026-09-04, verbatim:** *"let's use the hero images somehow, we don't need a stop gap though."*

Asked in these terms: finetuning.uk has **38 hero components rendering 2 distinct images** (35 of them
`/assets/images/hero.jpg`), and IMG-077 filed two `needs_human_review` items on 2026-09-03 —
`6db67bde` (**4 pages "unwired"**: use-cases, case-studies, approach, contact, each holding a deployed
`content-hero-<page>.jpg` the page never renders) and `d280a6fd` (6 pages `no_image_slot`, which per
migration 686's rollback should be left alone). The choice put to him was: hand-wire the four as a
stop-gap, or arm the built mechanism. He chose the mechanism and explicitly declined the stop-gap.

**Why it is yours, not ours or the uplift lane's:** `bugs_open/412` §10, 2026-09-02, assigns candidate 1
to this lane, with the stated reason that 412's own lane is a site lane redirected to service work and
"holding it here is how a candidate becomes an orphan".

**The state of the mechanism, as the editorial_design_uplift lane measured it (2026-09-03):**
`wirePageHeroOnLanding` is PRESENT in the running binary (v1.0.1359, probed at the pod with controls on
both sides), called from `flag_page_image_rebuild_action.go:210`, gated behind the opt-in
`wire_hero_on_landing` — and that key is named by **zero** live `agent_definitions` rows. A REVISE is
outstanding on its council round (`bd78490d`).

**The argument for arming rather than repeating a migration, which is the part worth keeping.**
Migration **664** hand-wired these exact pages on 2026-08-26 and did not hope — its verify block asserted
`% of 9 hero components carry a hero_url, want 9` and passed. Measured 2026-09-03 against 664's own
page list: **3 of 9 survive** (about, careers, services); approach, case-studies, contact,
model-approach-selector, tool-ai-readiness-checker and use-cases lost the key within **eight days**,
exactly as 412 §10 predicted in writing. A second hand-wire repeats a repair now measured to decay by
two thirds in a week.

**What this lane offers:** finetuning.uk's four unwired pages have their assets already generated and
deployed, so they are a ready witness. Ask and we will verify at the served page — this lane checks
imagery at the artefact as a matter of course (it closed ten stale `image_url_404` rows on this site
yesterday after probing every URL with an invented-URL control).

Nothing is asked of you on a timetable; this note exists so the decision is on the record where the work
lives, rather than in another lane's chat.
