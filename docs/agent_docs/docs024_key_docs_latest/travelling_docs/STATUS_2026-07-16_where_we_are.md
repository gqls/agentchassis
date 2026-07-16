# Where we are — self-documenting, self-verifying tools

*State of play, 2026-07-16. Supersedes `STATUS_2026-07-13_where_we_are.md`.
Companion to `OVERVIEW_self_verifying_tools.md` (the concept) and
`RUNBOOK_travelling_docs(39).md` §0 (the operating manual + position line). The
turn-by-turn blow-by-blow is the HANDOFF Turn log (T14–T17). This doc is the
snapshot: what's built, what's proven, what's next.*

---

## Headline

**The self-verifying loop is complete, and it has been proven GREEN on a real,
non-manufactured bug.** The system found a genuine mobile-layout defect on a live
site, worked out that it was the site's shared footer rather than the tool,
routed it to the right fixer, caught that fixer's first attempt LYING about
success, root-caused the problem to the durable template layer, fixed it,
deployed it, and re-verified the once-failing check now passes — with humans only
choosing between well-framed options, never touching the mechanism. Proven
2026-07-15; the redesigned durable-layer fixer went live in **v1.0.1123**
(2026-07-16).

---

## The milestone ladder (all proven in production)

| # | Milestone | Proof |
|---|---|---|
| 1 | **Tools write their own PLAN** (spec + criteria) at creation | tool-generator documented a tool it built, unaided |
| 2 | **Agents write their own fix NOTES** after a repair | broken game recreated with both bugs fixed; first machine fix-notes |
| 3 | **Diagnosis loop writes NOTES** | first machine diagnosis note |
| 4 | **Tier-2 static acceptance** runs on a live site | first sweep caught a hallucinated selector + a delivery mismatch |
| 5 | **Tier-4 browser tier self-drives** (pass) | agent drove a live tool in Chromium → first machine `acceptance-run` note |
| 6 | **Tier-4 fail path** (fail → note + fix ticket) | controlled test, then a REAL failure (below) |
| 7 | **Fully continuous** | discovery sweep → item → dispatch → browser run → verdict, zero manual triggers |
| 8 | **Tier-4 P1 (mobile) + P2 (interactions)** | 9/0/1 on xp-curve: interaction really drove the tool on desktop AND mobile |
| 9 | **Attribution: tool defect vs site chrome** | vonc footer overflow attributed to `div.footer-legal` in site-footer, NOT the tool |
| 10 | **Routing** | site-chrome defect → `responsive_fix`/component-template-fixer, not improve_tool |
| 11 | **Durable fix + re-verify GREEN** | real footer bug fixed at the template layer, redeployed, `mobile-fit@mobile` PASSED |

---

## The architecture in one picture

```
  agents build/fix tools ──► travelling docs in Postgres
                              • doc_plans  = PLAN (intent + acceptance criteria + container), versioned
                              • doc_notes  = NOTES (every fix/diagnosis/verdict), append-only

  design-discovery sweep (scheduled)
        │  tool_acceptance_due: any deployed tool with criteria, due a run?
        ▼
  acceptance_run item ──► dispatch loop ──► tool-acceptance-agent
                                              │  load PLAN criteria
                                              ▼
                                    browser-runner-adapter  (Chromium, live page, DESKTOP + MOBILE)
                                              │  status · selector · console · overflow(P1) · interaction(P2)
                                              ▼
                                    judge ──► attribute each failure: tool | chrome | unknown
                                              │
                        ┌─────────────────────┴───────────────────────┐
                        ▼                                              ▼
              scope = tool/unknown                            scope = chrome
              acceptance-fail note                            responsive_fix / chrome_overflow_fix
              + improve_tool ──► tool-improver                ──► component-template-fixer
                                                                   patches the DURABLE
                                                                   content_component template
                        └──────────────► rerender ──► re-run Tier-4 ──► GREEN ◄──────┘
```

Every arrow is machine-driven. The criteria come from a document the machine
wrote; the run is triggered by a scheduled sweep; the verdict is written back
into the same travelling docs; a failure becomes a ticket the existing fix
pipeline processes; and the fix is re-verified in the same browser tier.

---

## What's live right now (v1.0.1123)

- **Travelling docs** — PLANs at tool birth (now with a `container` selector);
  NOTES appended by every fix agent, the diagnosis loop, and every Tier-4 verdict.
- **The verification ladder** — Tier 0 (generation) · Tier 1 (structural) · Tier
  2 (static acceptance, the anchor rule) · Tier 4 (behavioural, headless Chromium,
  **desktop + mobile**, P0 checks + P1 overflow + P2 interactions). *(Tier 3, an
  LLM code-review, predates this work.)*
- **`browser-runner-adapter`** — headless Chromium via Playwright; on an overflow
  it names the widest offending element, says whether it sits inside the tool's
  container, and reports the offender's CSS selector + owning slot.
- **`tool-acceptance-agent`** — turns criteria into a browser run and the run into
  a verdict; labels results `id@profile`; attributes failures; routes tool defects
  to improve_tool and site-chrome defects to a responsive/chrome_overflow fix.
- **`component-template-fixer` / `chrome_overflow_fix`** — a targeted overflow fix
  that patches the DURABLE content_component template (survives refresh), refuses
  to run without a slot + selector, and reports its shared-site blast radius.
- **`tool_acceptance_due`** — the periodic check that makes it continuous; its
  cooldown is now scoped to Tier-4 verdicts (`source='tool-acceptance'`).
- **A migrations system** — numbered SQL + runner + ledger, applied through 151.

---

## What's next

| Item | Why | State |
|---|---|---|
| **Migration number collisions** | Other workstreams landed duplicate 149/150/151; `151_gripper_spec_sheet_component` fails on a dup and blocks the runner | **being fixed in a SEPARATE chat**; next free number is **152** |
| **The other 7 footer-4-column sites** | They share the fixed template but still have stale rendered_html | self-heal on their next refresh; not force-rerendered (7 live sites) |
| **Per-site override for shared-template fixes** | An autonomous per-site failure editing a template shared by N sites is broad blast radius; may want per-site isolation | optional design choice, deferred |
| **Screenshots on failure (P3)** | Visual evidence attached to a failing verdict | **BUILT 2026-07-16** — full-page PNG per failing (url, profile) → B2; durable `s3://` URI in the note's `Evidence:` line, presigned link in the item spec. Gated on the next chassis + adapter images + adapter storage env |
| **The 54 legacy `responsive_fix` items** | All lack `slot_name` → they defaulted to the header and were never really fixed | flagged to their owning workstream |

---

## The ideas worth repeating (for talks)

- **A tool that documents its own definition of *working*, tests itself against
  it in a real browser (desktop and mobile), works out whether a failure is its
  own fault or the site's, files the right repair ticket, and re-checks the fix** —
  keeping a written record so a later fix won't undo an earlier one.
- **"Passed the checks" is not "works" — at every layer.** We watched the gap ship
  the same bugs twice; then we watched a *fixer* claim success while changing
  nothing. Behavioural re-verification is the only thing that proved the fix real.
- **Fix the source, not the artifact.** A rendered footer is regenerated from its
  template; patching the render is wiped by the next refresh. The durable fix
  lives in the template.
- **The machine maintains the docs.** Humans never write or update travelling
  docs; the agents that change a thing write its history.
- **Criteria describe delivered reality, not aspiration.**
