# Where we are — self-documenting, self-verifying tools

*State of play, 2026-07-13. Companion to `OVERVIEW_self_verifying_tools.md` (the
concept) and `RUNBOOK_travelling_docs.md` (the operating manual). This doc is
the status snapshot: what's built, what's proven, what's next.*

---

## Headline

**The mechanism runs unattended, end to end.** A scheduled maintenance sweep
finds a tool due for verification, drives it in a real headless browser against
the acceptance criteria in its own travelling PLAN, and writes the verdict back
into that tool's history — with no human anywhere in the chain. Proven live on
2026-07-13.

---

## The milestone ladder (all proven in production)

| # | Milestone | Proof |
|---|---|---|
| 1 | **Tools write their own PLAN** (spec + criteria) at creation | tool-generator documented a tool it built, unaided |
| 2 | **Agents write their own fix NOTES** after a repair | broken game recreated with both bugs fixed; first machine fix-notes |
| 3 | **Diagnosis loop writes NOTES** | first machine diagnosis note (`unconfirmed-diagnosis`) |
| 4 | **Tier-2 static acceptance** runs on a live site | first sweep caught a hallucinated selector + a delivery mismatch |
| 5 | **Tier-4 browser tier self-drives** (pass) | agent drove a live tool in Chromium → first machine `acceptance-run` note |
| 6 | **Tier-4 fail path** (fail → note + fix ticket) | controlled test: `acceptance-fail` note + `improve_tool` item carrying the criteria |
| 7 | **Fully continuous** | discovery sweep → item → dispatch → browser run → verdict, zero manual triggers |

---

## The architecture in one picture

```
  agents build/fix tools ──► travelling docs in Postgres
                              • doc_plans  = PLAN (intent + acceptance criteria), versioned
                              • doc_notes  = NOTES (every fix/diagnosis/verdict), append-only

  design-discovery sweep (scheduled)
        │  tool_acceptance_due check: any deployed tool with criteria, due a run?
        ▼
  acceptance_run work item ──► dispatch loop ──► tool-acceptance-agent
                                                    │  load PLAN criteria
                                                    ▼
                                          browser-runner-adapter  (Chromium, live page)
                                                    │  page_status_ok · selector_exists · no_console_errors
                                                    ▼
                                          judge ──► acceptance-run note        (pass)
                                                └─► acceptance-fail note
                                                     + improve_tool ticket ──► tool-improver
                                                        (loads PLAN+NOTES, fixes, redeploys, re-verifies)
```

Every arrow is machine-driven. The criteria come from a document the machine
wrote; the run is triggered by a scheduled sweep; the verdict is written back
into the same travelling docs; a failure becomes a ticket the existing fix
pipeline already knows how to process.

---

## What's live right now

- **Travelling docs** — PLANs written at tool birth; NOTES appended by every fix
  agent and by the diagnosis loop.
- **The verification ladder** — Tier 0 (generation) · Tier 1 (structural) · Tier
  2 (static acceptance, the anchor rule) · Tier 4 (behavioural, headless
  Chromium). All four live. *(Tier 3, an LLM code-review, predates this work.)*
- **`browser-runner-adapter`** — a cluster service running headless Chromium via
  Playwright; three P0 checks (`page_status_ok`, `selector_exists`,
  `no_console_errors`), desktop profile.
- **`tool-acceptance-agent`** — the orchestrator that turns criteria into a
  browser run and the run into a verdict + (on failure) a fix ticket.
- **`tool_acceptance_due`** — the periodic check that makes it all continuous.
- **A migrations system** — numbered SQL + a runner + a ledger; every agent
  change snapshotted and guarded. 24 migrations applied this arc.

---

## What's next

| Item | Why | State |
|---|---|---|
| **P2 — interaction checks** | Assert a tool computes the *right answer* (fill a slider, click, check the output moved) — the tier that would have caught the economy-simulator bugs directly | in build |
| **P1 — mobile profile** | Run the criteria under a phone viewport; catch horizontal-overflow | in build |
| **Real failure → tool-improver → back** | The one link proven only with a manufactured failure so far | waits for a real break |
| **Cooldown scoping** | A stale Tier-2 verdict can currently suppress a Tier-4 run; scope the cooldown to Tier-4 verdicts | noted follow-up |
| **Screenshots on failure (P3)** | Visual evidence attached to a failing verdict | later |

---

## The ideas worth repeating (for talks)

- **A tool that documents its own definition of *working*, tests itself against
  it in a real browser, and files its own repair ticket when it falls short** —
  while keeping a written record of every decision so a later fix won't undo an
  earlier one.
- **"Passed the checks" is not "works"** until the top tier says so — we watched
  the gap ship the same bugs twice, which is why the ladder exists.
- **The machine maintains the docs.** Humans never write or update travelling
  docs; the agents that change a thing write its history. That's the only way
  documentation stays true.
- **Criteria describe delivered reality, not aspiration.**
