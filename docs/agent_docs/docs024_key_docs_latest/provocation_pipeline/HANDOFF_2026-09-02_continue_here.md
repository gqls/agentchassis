# HANDOFF — provocation pipeline, 2026-09-02 — COLD START

> ## ⇢ SUPERSEDED the same day by `HANDOFF_2026-09-02b_live_and_self_driving.md`
> This file was written while the change was still waiting on a fleet roll. The roll
> landed (v1.0.1354), `685` is applied, the schedules are running and six days are
> queued. **Its "THE ONE THING THAT NEEDS A HUMAN" section is DONE.** Kept for the
> trail — read the 09-02b file instead.

**Supersedes `HANDOFF_2026-08-10b_the_generator_works.md` and the 08-12 delta on it.**

## In one line

The owner has removed his own approval step and asked for a daily provocation; the code
half is committed and the config half is committed-but-held, waiting on a fleet roll.

## THE ONE THING THAT NEEDS A HUMAN

**A fleet roll (`make release`).** Nothing else in this lane can move until commit
`326370d6c` is live. After the roll, the sequence in `RUNBOOK §16f` finishes the job —
it is four steps and each one is checkable.

## State, measured not assumed

| thing | state (2026-09-02) |
|---|---|
| what the site serves | **`22 Aug`**, `generated_at 2026-08-22T04:58:04Z` — 11 days stale |
| shelf | **2** approved provocations left, both undated; nothing dated in the future |
| publisher | **healthy** — 6h schedule, completed 08-31 11:17Z. It had nothing to publish |
| generator / scheduler | **no `scheduled_tasks` row at all** — this is the actual root cause |
| `gate_verdict->>'gate_version'` | `{1,2}` — the new binary has gated nothing yet |
| `685_HOLD` | committed, **NOT applied** |
| councils | `c08d263a` (Go), `fb31e95e` (config) — **submitted, NOT read** |

## What changed, and what it turns on

**The human-approval stamp is gone from all three queries** (feed / scheduler / exemplar
selection) — owner instruction, his **third** position on this question (none 07-31,
required 08-09, none 09-02). The column and every verdict are retained; restoring it is
three predicates. **Restore it in BOTH queries or neither** — the 08-09 defect was exactly
a comment claiming a predicate the query did not have.

**The readability rail is now FATAL** (`hard_to_read` rejects instead of recording). It is
the deliberate replacement for the reader being removed, chosen over the judge because the
judge is documented-stochastic on this corpus and arithmetic cannot drift.

**Verified before the change, not after:** removing the stamp published nothing
retroactively — every `approved` row on `vonc.com` was already stamped (8 human + 15 llm,
**0 unstamped**), and the 8 unstamped `approved` rows are calibration fixtures on
`calibration.vonc.com`, which the domain-scoped queries cannot reach.

## ⚠ THE ORDERING TRAP — read before touching `685`

`685` **must not** apply before the roll. If it does, the generator banks drafts gated
while the rail is still advisory, and `loadGateCandidates` never re-gates an approved row
(by design, so model drift cannot retract a published provocation). **That batch stays
publishable for ever without the rail ever applying to it** — not self-correcting, and
invisible afterwards.

`685`'s guard enforces this mechanically: it refuses until a row carries
`gate_version = '3'`, which only the new binary writes. That is an **artefact** check, not
a tag check. **Proven in the refusing direction against the live DB.**

## ⚠ `321` NOW FAILS IF RE-RUN

`321_provocation_scheduler_operator_handle.sql` RAISEs if the scheduler ever gets a
`scheduled_tasks` row, citing the 08-09 ruling — which the owner has reversed. Expected
from today. **Do not repair the database to satisfy it, and do not edit the applied file**
(it is ledger-recorded; editing breaks its checksum). Full entry in `LANDMINES.md`.

## Owed, and NOT claimed

- **The §10.6 LIVE calibration was NOT re-run** (`PROVOCATION_LIVE_CALIBRATION=1`, real
  key, real tokens) and **two of its four bad-set fixtures changed**. Nobody should cite
  that calibration as current until it is run. This is the largest outstanding gap.
- **Both council verdicts are unread.** A `Council-Submitted:` trailer is a submission,
  never a verdict. Read them and act on a REVISE — the code is already on the shared
  branch.
- **Nothing in the config half is live.** The 14-day shelf is a target the generator has
  never been asked to sustain unattended.
- Agent types keep the `-manual` suffix while scheduled — carried in `description` rather
  than renamed, because renaming `agent_definitions.type` risks in-flight dispatch and
  breaks 321/371's own verify queries.

## Corresponded with other lanes this session (owner's ask)

- **`offer analyser benefit analyser visual designer` [4628f9]** — sent the measured vonc
  palette (six saturated hues on near-black; `--color-primary` has churned to `#7c3cff`,
  so the `#6d28d9` in older docs is stale) and the click-path measurement.
- **`experience_loop`** — `CONTRIB_2026-08-31_clicks_to_start_playing_vonc.md`. Note that
  lane was superseded 07-25; `gauntlet_dead_cta` owns the Spark build.
- Measured for both: **two clicks before a visitor can type anything**, and the home CTA
  says "File Your Position" while the page it lands on requires an "Enter the Gauntlet"
  click first — a promise/delivery mismatch `misdirected_cta` structurally cannot see.

## Read next

`PLAN §15` (the decisions and why), `NOTES` tail (the measurements and the misstep worth
not repeating), `README_where_we_are` tail (the owner's plain-prose version),
`RUNBOOK §16` (every command, including the `_HOLD` dry-run technique).
