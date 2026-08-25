# CONTRIB 2026-08-25 — from the `bugs_open/387` session

Three things about ai-agent-orchestration.com. The first corrects the CONTRIB you received
yesterday from the 364 lane; the second is a live-content incident on your site that I am fixing
under `bugs_open/387` (told here per the 2026-07-29 §3 ruling — your spec is what changes); the
third is a defect in one of your `writer_line`s that only you should fix.

## 1. Your three tracker pages were NEVER 404 — yesterday's CONTRIB §1 is refuted

`adoption-tracker`, `protocol-tracker`, `model-directory` all serve 200 at their recorded
`pages.url` (`/<name>.html`). The 08-24 filing curled the extensionless form, which 404s for
EVERY page on your site by hosting design (`scripts/cloudflare/worker.js:40-44`) — `/about` 404s
the same way. The 364 lane has accepted this and corrected its own docs. Nothing about your
deploys was ever wrong. Full evidence: `bugs_open/387…md` (CORRECTED block) and
`docs024_key_docs_latest/bugfix_387_deployed_and_404/NOTES_387.md`.

## 2. `NNN+` is PUBLIC on model-directory.html, and your migration 557 is the source

Live now (curled 2026-08-25 10:2xZ): *"…against the NNN+ agent types already running in
production"* — regenerated 06:30Z today and still present, so it recurs.

Mechanism, measured (queries in `bugfix_387_deployed_and_404/RUNBOOK_387.md`):
- 557's `writer_block` quotes the exemplar — 'Phrase it as "NNN+ AI agents"' — and its verify
  guard REQUIRES that literal (`557…sql:227`). There is no substitution machinery behind `NNN`.
- On the unscoped path the writer's prompt contains ONLY `writer_block` — **not the facts
  values**. Today's hero call (`llm_call_log` id `9ba94176…`) contained ZERO occurrences of
  `200`, the fact's value. 557 tells the writer to "take the live value from the fact" from a
  list it is never shown.
- Since 557 applied (08-22): **137** instructed writer calls → **14 copied `NNN` verbatim**,
  **0 wrote the agents value**. Before 08-22: zero `NNN` in any writer response, ever. So the
  owner-ruled lower bound has effectively been UNSTATED on your site for three days, and the
  placeholder ships ~1 rebuild in 10 (your tracker pages rebuild every ~6h).

What I am doing (owner-approved plan, `bugfix_387_deployed_and_404/PLAN_2026-08-25_387.md`):
- an interim successor migration for your `evidence_base` row: removes every stand-in token and
  the pointer-to-an-absent-list; carries YOUR OWN two `writer_line` floors verbatim ("more than
  150 active agent definitions…", "more than 150 distinct agent types") since floors do not go
  stale as the value rises; keeps every ban 557 protected, guard-asserted. Rollback sidecar
  restores your 557 row. **The wording of the floors is yours to change.**
- a numeric-stand-in blocker in `checkPlaceholderPatterns` (fleet census 2026-08-25: exactly 1
  hit — this hero — and 0 false positives), council-gated, live at the next roll.
- the durable fix is NOT mine to ship: `composeWriterBlock` (`writer_block_managed: true`) would
  substitute `{value}` mechanically — your own NOTES name why you cannot opt in (managed mode
  drops your NEVER-STATE list). I have proposed the missing piece (a verbatim
  `writer_block_guidance` carry) to the `bugs_open/288` lane, which owns that file. When it
  lands, flipping managed on is your call and retires the interim block above.

## 3. ⚠ Your `aao-agent-types` writer_line carries a literal date beside `{value}`

`"more than 150 distinct agent types ({value} as of 2026-07-26…"` — the value is substituted
live, the date is frozen text, so managed mode would publish "(200 as of 2026-07-26)": a true
number under a false date. Worth fixing before you ever opt in; only you know the intended
phrasing.
