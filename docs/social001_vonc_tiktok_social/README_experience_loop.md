I've diagnosed all three broken surfaces against live artifacts, and written up a workflow proposal that folds both of your directions into one loop. The plan is committed at
docs/agent_docs/docs024_key_docs_latest/experience_loop/PLAN_experience_loop.md. Here's the substance.

What's actually broken (three different failure classes, all verified)

Provocations archive — the entries are runtime-filled into a template whose href="#" never gets a destination. Per-provocation detail pages were never planned or built (that's the deliberately parked
needs_page:provocation item from 07-12). Clicking "returns to the same page" because # is a self-link. This is a journey dead-end: nobody ever specified where clicking should go.

Arena — the current stored component (23KB, from 07-14) contains no fetch of /data/provocations.json at all, even though the feed itself is live and healthy. The page cannot load a provocation — "Loading… DAY 0" is
the shell's hardcoded state. Worse: the Arena has a travelling-docs PLAN with acceptance criteria, but it's orphaned under the old key tool-arena-interface — when the page was renamed to tool-arena during the
generator-drift reconcile, the doc didn't move, so the acceptance sweep has never once covered the live page. The defect was invisible to the whole self-verifying ladder.

Gauntlet — this one isn't broken; it was built this way. Both CTAs are literally href="#", and its 3.9KB of JavaScript only does the strikethroughs, the timer, and the stat counters. It's a marketing mockup shipped
as a product — including fabricated numbers (12,847 competitors, a named leaderboard), which violates your standing anti-fabrication rule. It predates travelling docs, so it has no PLAN and no criteria.

The common thread: every check we have verifies a page or tool in isolation; nothing owns the experience. The link-integrity loop we closed yesterday guarantees a button reaches a real page — it cannot know the
page is a mock.

The proposed workflow: the Experience Loop

It's the travelling-docs machinery lifted one level up, with your "several loops of discussion and challenge" made mechanical:

- Phase A — Experience spec. A new experience-planner agent writes an EXPERIENCE_PLAN (a doc_plans row, subject_type='experience'): journeys as first-class objects (page → control → action → observable outcome), a
  promise ledger (every CTA's copy → what its destination must deliver), data contracts (what provocations.json must contain and who writes it), and an explicit MVP cut with the hard rule today's mess teaches: a
  not-yet feature is absent or labelled coming-soon — never simulated.
- Phase B — Challenge council. Four critics with distinct lenses (journey-completeness, feasibility, honesty/anti-fabrication, MVP referee) attack the spec; the planner revises; loop until convergence or max
  rounds. This reuses the council pattern already live in the concept register. Every round is machine-logged in RUNNING_NOTES — no human writes.
- Phase C — Contract-first build. Spec becomes pages/tools with acceptance criteria attached at birth — the thing tool-generator already does for tools, extended to pages and journeys.
- Phase D — Journey acceptance + self-heal. The existing Tier-4 browser tier runs the criteria, extended to multi-page journeys (its interaction machinery plus navigation). Failures attribute to tool/chrome as
  today, plus two new scopes — data and plan-gap — where a needs_experience_replan item routes back to Phase A instead of a code fixer.
- Phase E — Feature rounds. MVP green end-to-end → next slice from the LATER list, repeat B→D. Copy and imagery are rounds, not blockers.

Most of this is reuse: doc_plans/notes, the acceptance ladder with screenshots, the council pattern, the work-item pipeline, and the claims machinery for the fabricated numbers. The genuinely new pieces are small
and listed in §6 of the plan — plus four guard rails today's incident demands regardless: a page-ownership marker so generic rebuilds refuse to touch tool-owned pages (the arena clobber class), doc re-keying on
rename (the orphaned-criteria class), a Tier-2 dead-control check (href="#" fails acceptance), and quantitative-copy-must-trace-to-data.

The pilot and your three open decisions

Round 0 runs the loop on the Spark game itself — the spec phase will rediscover all three defects and fix them to a spec rather than patching. Only three product calls surface, each with a default so you can just
say go: (1) Gauntlet in MVP: minimal-real playable round (default) or honest demotion to coming-soon; (2) provocation detail pages: static pages emitted daily by the pipeline (default) or client-side rendering; (3)
full autonomy on vonc with artifact-verified checkpoints instead of approval gates (default: yes).
