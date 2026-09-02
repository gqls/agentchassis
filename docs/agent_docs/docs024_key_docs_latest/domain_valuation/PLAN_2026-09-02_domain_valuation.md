# PLAN — portfolio-wide domain valuation (2026-09-02)

## What the owner asked for (2026-09-02, session "domain valuation")

1. Get the full domain lists from **Dynadot, Porkbun, Nominet, Spaceship** (the
   four registrar lanes each own their own API/tooling — ask them, don't rebuild).
2. Value **every** domain — .co.uk and .uk were discussed before; **.com is new
   in scope this round**.
3. The goal is to **sell roughly the bottom 500, priced keenly**. The rest are kept.
4. **Categories stay together** (owner's words: "keep all category e.g. financial
   domains together and not isolate some that might be less good yet") — the
   sell-list must not cherry-pick weak members out of a category family that is
   otherwise being kept. Category first, then rank.
5. Prior valuations exist in earlier conversations (.co.uk/.uk) — mine them as
   the starting point, then improve.
6. Registrar-side valuations are welcome inputs (e.g. Dynadot appraisals).
   **Afternic already carries prices but the owner says they are generally
   overpriced and will be changed** — treat Afternic prices as a comparison
   column, never as the answer.

## Phasing

- **P1 — inventory.** One consolidated list: domain, TLD, registrar, expiry.
  Sources: registrar lanes write CSVs into `inbound/` here. Nominet (~1,500 .uk)
  is the big missing piece — its `walk` needs an owner-run `!` command
  (nominet lane README, 2026-09-02).
- **P2 — prior art.** Transcript mining of the earlier .co.uk/.uk valuation
  discussions (agent dispatched 2026-09-02) → `PRIOR_ART_*.md` here.
- **P3 — categorise.** Assign every domain a category (financial, design,
  games, health, …). Reuse `portfolio_positioning`'s classification work
  (`RUNBOOK_domain_inventory_and_classification.md`, `REGISTER_positioning.md`)
  before inventing anything.
- **P4 — value.** Per-domain valuation with method + confidence, blending:
  prior-session valuations, registrar appraisals where offered, keyword/TLD/
  length/commercial-intent scoring, and known comparables. Afternic's current
  ask as a column.
- **P5 — the cut.** Rank within category; propose the bottom-~500 sell list
  honouring rule 4; keen pricing per tier; owner reviews before anything is
  listed or repriced anywhere.

## Decisions taken

- 2026-09-02: lane directory created; registrar lanes asked for lists +
  valuations via cross-session messages rather than re-implementing their APIs
  here (they own the credentials and the tooling; CLAUDE.md: reuse machinery).
