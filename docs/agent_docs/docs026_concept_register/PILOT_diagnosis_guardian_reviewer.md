# PILOT — "Diagnosis-Loop Guardian" council reviewer (stage 3, seat #6; candidate #4)

**Status: LIVE as of 2026-07-18.** Applied via
`fixloop_eg_dartsonline/0NN_fix_proposer_v13_diagnosis_guardian.sql` — gated
behind the relevance filter, added surgically (the v12 pattern). Council is now
**8 reviewers** (2 always-on + 6 gated specialists). Drift-checked first: the
live row had been updated by another thread that morning; all anchors (proviso,
`code_checks`, 7-seat state) verified intact before applying, and re-verified
after.

---

## 1. Why this seat

Candidate #4. `diagnosis-loop.md` is the **highest hot-concept-density category
in the register** (7 of 41 concepts heavily rediscovered) — and it has a
self-referential importance: the diagnosis loop is the machinery that *feeds*
this very council, and its guards were earned from real failures (benchmark
runs that produced CONFIRMED verdicts a fixer must never have acted on). A fix
that touches the loop could quietly weaken the honesty gates everything else
depends on. The loop reviewing fixes to itself is the point, not a conflict.

## 2. Charter — the disciplines it defends

| Discipline | Grounding |
|---|---|
| Read-only, cite-or-abstain; UNVERIFIABLE is an honest terminal | `DIAG-001`, `DIAG-003` |
| CONFIRMED needs BOTH a static AND a state/runtime citation | `DIAG-008` (+ the two-evidence-family guard proven live on BUG B) |
| Model-authored SQL runs under three layered guards (sqlguard, read-only role, EXPLAIN pre-flight) | `DIAG-009`, `DIAG-016` |
| Observability never costs a diagnosis (persistence degrades to a warning); notes are skip-never-guess | `DIAG-028` |
| `error_step` is CONFIG-level only — step-level is parsed but silently inert (recurring trap, found dormant in other agents too) | `DIAG-030` |
| Loop work + repo tokens stay off shared pods (spawn-wrapper + `isRepoCloningAgent` gate) | `DIAG-019`, `DIAG-022` |

**Judges:** (a) does any edit weaken a verdict guard or add an uncited-assertion
path; (b) does it touch the read-only SQL enforcement layers; (c) can
persistence now fail a diagnosis, or a note subject be guessed; (d) is
`error_step` placed outside config, or loop work/tokens moved onto shared pods.
If the fix doesn't touch the diagnosis machinery: approve.

**Verdicts: `approve | object`, no `veto`** — same advisory design as all
specialists.

## 3. Relevance footprint (gated)

```
"diagnosis": ["diagnose_","diagnosis_artifacts","diagnose-agent","diagnose-orchestrator","pkg/diagnose","verdict","cite-or-abstain","data_request","sqlguard","symptom"]
```

Fires on fixes touching the loop's actions (`diagnose_*` file prefix), its
tables, its agents, or its core vocabulary; abstains otherwise. Prompt and
exact wiring are in the migration file.
