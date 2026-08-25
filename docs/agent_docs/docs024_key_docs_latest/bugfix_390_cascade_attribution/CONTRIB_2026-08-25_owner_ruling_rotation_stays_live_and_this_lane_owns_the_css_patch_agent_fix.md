# CONTRIB (from loanzy_uk_example_site, 2026-08-25 evening) — OWNER RULING: the render-audit rotation STAYS LIVE, and this lane is asked to fix css-patch-agent

**Context.** The owner's plan for switching off the improvement loop's "evolutionary" rewrites
(`loanzy_uk_example_site/PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md`)
named `site-render-audit-rotation` → `contrast_failure` → `css-patch-agent` as the OTHER live source
of bad renders (`[MEASURED 2026-08-25]` 239 completions in 14 days), out of that plan's scope, and
put the on/off choice to the owner as choice (c).

**The ruling, 2026-08-25 evening, verbatim in substance:** *"please ask the css thread to fix that
css-patch-agent, leave it live"* — i.e. **the rotation is NOT paused**; the repair is routed to this
lane as the css-patch-agent's owner of record (616 shipped, commits 2 and 3 of your own 390 plan
pending; the erasure half filed by you as
`bugs_open/396_HANDOFF_2026-08-25_a_design_run_erases_every_appended_css_repair_and_the_work_items_stay_complete.md`).

**What "fix" plausibly covers, from your own artefacts (not new scope from me):**
1. your commits 2/3 — the render audit records WHICH declaration wins and where it lives, and the
   blanket `!important` instruction becomes the measured requirement (616's header names both);
2. the erasure (your 396-css): `persist_css_to_theme` rewriting `css_themes.css_content`
   byte-for-byte at the next design run, taking every appended repair with it, items still `complete`;
3. the completion honesty half the owner deferred on 2026-08-25 (`complete` ≠ "the text became
   readable") — still deferred unless he says otherwise; this CONTRIB does not reopen it.

**No live session for this lane was listed tonight** (checked `ListAgents` 2026-08-25 ~21:20Z), so
this file is the notification; whoever next opens the lane inherits the ruling here. Nothing in
your tree was changed by me beyond adding this file.

**One adjacent fact you may want:** the improvement loop's model seats are bypassed as of tonight
(migration 623) and `improvement-sweep` is re-enabled — so `needs_design_review` /
`dark_section_audit` inflow from the LLM design audit is OFF, but your rotation's
`contrast_failure` inflow is unchanged, by the owner's explicit word.
