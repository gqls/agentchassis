# HANDOFF 2026-08-11b — fact-assignment front: the census is DONE and PROVEN; two council rounds in flight. Cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-11_…` as the entry point.** Written late afternoon 2026-08-11,
after one session executed the entire 08-11 job queue: the census ran and passed, the 012
round is built+submitted, and the two small seeds are built+submitted. What remains is
**verdict-gated application and one follow-up measurement** — nothing else.

Site id: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**. Both verdicts by CORRELATION, never
`doc_notes … LIMIT 1`.

## 1. The census result (full evidence: NOTES 2026-08-11 entries; bugs_open/151 dated note)

Replan corr `e74974b3`, new current plan **`40a66d3a`**: 18/18 singular built pages
preserved by name+order (362 proven; zero carries — the mechanism, not the net); all 9
writer-visible facts assigned one-place-each; rebuilds through the new path state assigned
facts in register-mandated form (floors "10+"/"12+"/"0") and NOTHING unassigned.
Whole-site fact-overlap pairs **34 → 9** (the recorded "9 pre-round" had staled to 34 —
re-derived same-morning with a faithful instrument port, `fact_overlap_census.py` in the
session scratchpad + pinned inputs). All 9 residual pairs are non-writer: 3× the same
evidence-chart data on three pages (a COMPOSITION question for the owner), 2× portfolio
card resolver metrics, 4× stale copy on production-backend-engineering (drains at its next
rebuild). Fact-blind sites: zero writer/build orchestrations all round. **151's candidate-1
arc is measured-done** (file stays in `bugs_open/` per the 08-06 ruling).

## 2. IN FLIGHT — two council rounds, apply gated on verdicts

1. **012 round** (seed `385_build_site_planner_recompose_pages_visible.sql`), corr
   **`62d2463f-b269-41fb-8f25-078983ffceab`**, commit `fc01fdbc2`.
   On APPROVED: run the seed (`psql < sql_for_agents/385_…`), then verify
   (`REDESIGN REQUESTED` present, post-length 20,445), then a live recompose run when one
   is next wanted proves the field; after that proof, the registered follow-up retires the
   prose escape's load-bearing status (a further seed) + updates the LANDMINES 2026-08-10
   recompose entry + `features_open/012`. On REVISE: objections come back with checks
   answered; resubmit with `RESUBMIT_CORR=62d2463f…`.
2. **Seeds 386+387 round**, corr **`d1e8c36e-6c48-4025-9e6a-f24deabb9896`**, commit
   `28436b190`. On APPROVED: apply 386 (writer rule-5 commitments ban; verify post-length
   13,974), apply 387 (guidelines seat, fix-proposer; verify post-length 8,695), then
   **run `099_SYNC_gate_roster.py` dry then `--apply`** so council-gate mirrors 387 —
   never hand-patch the gate. Record all three in the migration ledger `--record-only`.
   All three seeds' drift guards RAISE on anchor-count<>1; if one fires, the live row
   drifted — regenerate the seed from a fresh dump (the generator recipe is in NOTES).

Both rounds went in with FORCE=1 (pure-seed rounds read as docs to 097's path filter; the
edited artefacts are live agent_definitions rows — reason stated in each rationale).

## 3. Owner items surfaced by the census (recorded, NOT actioned — his calls)

- **The 215 revisit trigger TRIPPED**: `PLAN_PAGE_MERGE_LOSSY` = 2 (both composed-vs-composed
  guide-pair merges, all four page rows live). Noted in `bugs_open/215` + lane README. The
  underlying duplicate-page condition (one page, two names, both live) has no owner yet.
- **The evidence-chart composition question**: the same chart serves on index, capabilities
  and digital-asset-recovery — 3 of the 9 residual pairs. Does it belong on all three?
- Compliance findings 3.5/3.6 remain recorded, unscheduled.

## 4. Cross-lane notices already delivered (do not redo)

- **Sweep front** (its 08-09 handoff, addendum 2026-08-11): the replan re-planned its three
  archived phantoms; `ai-readiness-checker-guide` auto-built AND deployed (needs file
  retraction this time), the other two parked at `owned_page_review`. Hand-archive pass is
  theirs, per ruling 1.
- **bugfix-235** (`bugs_open/235`, addendum): my index regeneration partially reverted
  their logo patch — relojistas + idea.uk back to origin `.jpg` (the 238 family on the
  resolver path). All URLs serve 200 today, but their deletion gate is NOT met; their
  durable fix must land at the resolver's source.
- **The success-result delivery failures** (3 builds today "failed" after the work
  persisted; "message validation failed (code: CHILD_ORCHESTRATION_FAILED)" on a SUCCESS
  result): evidence pinned in NOTES with orch ids (pruned ~24h). Not squarely 207/216/217.
  Deliberately NOT filed — no root cause diagnosed; a future filer should run 090 or
  declare the substitution.

## 5. Standing traps (unchanged from 08-10b §5, all still live)

`orchestration_states` prunes ~24h · verdicts by correlation · a roll is not evidence
(re-verify detection literals per SERVICE after any roll) · the tree is shared, expect
passengers both directions · `who-owns` is blind to uncommitted sessions.

## 6. Commit trail (this session)

`c3b424c8e` census complete (docs + 3 bug-file notes) · `fc01fdbc2` 012 round built+submitted
(Council-Submitted 62d2463f) · `28436b190` seeds 386+387 built+submitted (Council-Submitted
d1e8c36e) · this file's commit.
