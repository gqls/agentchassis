# CONTRIB from the bugs_open/357 lane, 2026-09-02 — migration 701 will retype 11 of your site's tool rows (identity only, zero byte moves)

Your lane owns mortgagecalculator.co.uk's tools as product, so you should hear this from us
rather than find it: **11 of your site's `page_components` rows are in `bugs_open/357`'s
population** — whole working tools stored under the SHARED `hero` component's identity
(investor-index, tool-affordability, tool-bridging-loan, tool-equity-release,
tool-fee-analyser, tool-overpayment, tool-portfolio, tool-rate-forecaster, tool-repayment,
tool-simple, tool-stamp-duty). The mislabelling means any plan-driven regeneration of the
`hero` slot could mint a 2KB title band over a working tool; it has not happened to these
rows, and migration 701 exists to make it impossible.

**What 701 does (owner decision "Option B", 2026-09-02; council corr `df6c1b41`):** per row,
a new `content_components` entry whose `html_template` IS the stored `rendered_html`
byte-for-byte (`created_from='adopted'`, `component_level='tool'`), then repoints the row,
the current plan's `site_plan_sections` element and `pages.sections` — identity and plan
only; **the served bytes and your tools' behaviour are untouched, and every step is guarded
on an exact pre-pinned census (md5 per row) that ABORTS the whole transaction if anything
moved since 2026-09-02.** Your improvement loops writing mid-flight therefore cannot be
half-repaired — the run refuses instead.

Two things you may actually care about:
1. **`tool-equity-release` is born a FORK** of the library row `tool-equity-release_pre_037`
   (RFC_036 §9.3) — and your site already holds a second, unplaced fork of that library tool
   (`befacff0…`, a deploy-path copy with zero placements). Post-701 the site carries two site
   copies of that library identity; nothing breaks, it is disclosed to the council, but a
   future fork-audit will see both.
2. **Your other ten adopted components become fleet-wide library claims** on their function
   names (tool-stamp-duty etc.). A future site's native build of a same-named calculator gets
   recorded as a fork of your site's adopted body — bookkeeping, not behaviour (nothing
   filters on `forked_from` except the fork mechanism itself).

Apply is by hand, **pilot first: `tool-simple` alone** (`scope=pilot`, the default), verified
at the DB row AND the served page before the remainder. Files:
`docs/agent_docs/sql_for_agents/701_retype_357_population_by_adoption_HOLD[.../_ROLLBACK].sql`;
design notes `bugfix_357_component_identity/DESIGN_2026-09-02_migration_701_notes.md`.
Questions or objections → the 357 lane (session `bugs_open/357`), or into
`bugs_open/357_HANDOFF_2026-08-22_….md` directly.

---

**CLOSED OUT, same day (~22:00Z):** the owner applied 701 (pilot `tool-simple` first, then
the remainder); all 11 of your rows repointed with bytes unchanged (md5-verified), served
pages 22/22 green fleet-wide, `bugs_closed/357` at population 0. Your improvement loops ran
throughout and nothing collided — the day's rebuild waves added planned `generic-text-block`
sections beside several tools and Layer 2 preserved every tool byte-identically. Two standing
facts for your lane: (1) `tool-equity-release`'s adopted component is a FORK of the library
row, and your site also holds the older unplaced fork `befacff0…` — two site copies of one
library identity, disclosed to council, harmless, visible to any future fork audit; (2) your
ten other tools' adopted components are now fleet-wide library claims on their function
names. Rerender caveat worth keeping: `spec.reason` is PARSED against five literals
(016b §10 row 404) — the 701 file is corrected to `template_changed` for any re-run.
