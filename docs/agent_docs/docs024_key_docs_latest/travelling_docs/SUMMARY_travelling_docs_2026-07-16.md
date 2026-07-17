# Travelling docs & self-verifying tools — summary

*Written 2026-07-16, after the P3 proof. The three-question version: what we've
done, where we are, where we're going. Companions: `STATUS_2026-07-16_where_we_are.md`
(the detailed snapshot), `OVERVIEW_self_verifying_tools.md` (the concept, for
talks), `RUNBOOK_travelling_docs(39).md` §0 (the operating manual + position
line), `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` (turn-by-turn log).*

---

## What this is

Every complex tool and pipeline on the platform now carries its own
**travelling documentation in Postgres**: a PLAN (what it's for, how it's
delivered, and its acceptance criteria — the per-tool definition of *working*)
and an append-only NOTES log (every fix, diagnosis, and verdict). The agents
write these docs themselves as a byproduct of building and fixing things, load
them before touching a subject, and **test the criteria in a real browser** —
so a tool documents what "working" means, checks itself against that, works
out who's to blame when it fails, files the right repair ticket, and
re-verifies the fix.

## What we've done (2026-07-04 → 07-16)

1. **The docs travel.** Tools write their own PLAN at birth (tool-generator);
   every fixer and the diagnosis loop append NOTES; docs can never fail the
   work they document (`error_step` containment throughout).
2. **The verification ladder is whole.** Tier 2 (static acceptance, the anchor
   rule) sweeps live sites; Tier 4 drives the deployed tool in headless
   Chromium — desktop **and** mobile, with real interactions (fill/click/select
   → assert), overflow checks, and console capture.
3. **It runs itself.** A scheduled discovery sweep finds tools due a run, the
   dispatch loop drives `tool-acceptance-agent`, the verdict lands back in the
   travelling docs, and failures become work items — zero manual triggers.
4. **It's honest about blame.** A document-level failure is *attributed*: tool
   defect → improve_tool; site-chrome defect → a template fix routed to
   component-template-fixer. One overflowing footer is one site ticket, not a
   false accusation against every tool on the page.
5. **Proven green on a real bug.** Tier-4 found vonc.com's footer overflowing
   on mobile, attributed it to the shared `footer-4-column` template (8 sites),
   caught the first fixer *lying* about success, fixed the durable template
   layer, redeployed, and re-verified the failing check green — humans only
   picked between options.
6. **Failures now carry photographic evidence (P3, proven today).** A failing
   run photographs the full live page per profile; the note carries a durable
   `s3://` URI (never a presigned URL — notes feed LLM prompts), the work item
   carries a clickable 7-day link. The proof's first run immediately caught a
   real bug: cancelled work items were silently blocking every future item for
   the same defect (`idx_swi_dedup` counted `cancelled` as open) — fixed as
   migration 157.

Along the way this arc also produced: the migrations system (numbered SQL +
runner + ledger), the OOM root-cause fix (`chunkContent` infinite loop), the
economy-simulator recreation with both behavioural bugs fixed, and a bank of
durable debugging rules (016b, handoff §4).

## Where we are (tonight)

- **Live: chassis + browser-runner-adapter v1.0.1125.** The full loop — birth
  PLAN → scheduled discovery → browser run (desktop+mobile, interactions) →
  attribution → routing → durable fix → re-verify — plus P3 screenshots,
  verified end-to-end in production today (evidence image downloaded and
  inspected).
- **Migrations: applied through 159** (157 = dedup fix, 158 = Sonnet 5,
  159 = recreate_tool → Opus 4.8), all out of band: the runner is still
  blocked at the failing `151_gripper_spec_sheet_component.sql`, with 152–156
  (other workstreams' files) pending behind it. **Next free number: 160.**
- **Models: the tool pipeline now runs the current generation** (migrations
  158 + 159, snapshots taken): all 7 Sonnet steps across tool-generator,
  tool-improver, tool-recreation-handler and component-template-fixer moved
  claude-sonnet-4-6 → **claude-sonnet-5**, and `recreate_tool` (the 64k-token
  Opus-tier rebuild step) moved claude-opus-4-6 → **claude-opus-4-8**.
  tool-acceptance-agent has no LLM steps. No rebuild was needed (alias
  pass-through; no temperature sent), and diagnose-agent had already proven
  Sonnet 5 through this chassis.
- **Fleet context:** ~31 other agents still run claude-sonnet-4-6 (44 steps)
  and 23 run claude-haiku-4-5 — untouched; upgrading them is a separate,
  bigger decision.

## Where we're going

| Next | What it is | Status |
|---|---|---|
| First *natural* P3 evidence | A genuine (non-manufactured) failure arriving with screenshots via the scheduled sweep | happens on its own; nothing gated |
| The 7 footer-4-column sites | Share the fixed template but have stale rendered HTML | self-heal on their next refresh; left to natural cadence |
| Fleet model upgrade | The other ~31 sonnet-4-6 agents (44 steps) and 23 haiku-4-5 agents | user decision; 158/159 are the template |
| Per-site override for shared-template fixes | Blast-radius control if autonomous shared edits ever feel too broad | optional design, deferred |
| ~~Migration runner unblock~~ | 151–156 turned out to be *already applied* by the empty-sections workstream, just never ledger-recorded; artifacts verified, ledger backfilled | **DONE 2026-07-16** — runner reports "Up to date" |
| `DEBUGaa` log sweep | Old coordinator debug logging serialises collected data twice per action | undone; wide sweep, its own turn |

The core mechanism is **done and proven**. Everything above is polish,
housekeeping, or scale-out.
