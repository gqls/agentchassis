# HANDOFF — RFC_012 execution · **START HERE** · 2026-08-06

Cold-start for the next session. Read this, then `NOTES_rfc012_await_findings.md` (the
missteps) and `RUNBOOK_rfc012_await_findings.md` (every command, with its gotcha).

**This lane owns three owner rulings** (RFC_012 second sitting, recorded in the RFC itself
at commit `3851e90b5` — read that block, it is the authority, not this file):

1. **(d) → a STANDING check, ONLINE if the framework allows.**
2. **The §3(a) reader census — COMMISSIONED.**
3. **Option B implementation — ASSIGNED here.**

---

## STATE: what is DONE and committed

| piece | commit | state |
|---|---|---|
| The rulings, recorded in the RFC | `3851e90b5` | durable — survives any session |
| Lane docs (plan/notes/owner log) | `8b26e7fbb`-ish (the `docs(rfc012 lane)` commit) | open |
| **B core**: `agenterrors` leaf package | `5f49b4cfd` | **DONE, tests green** |
| **(d) detector**: `--shared-output-fields` | `abf5e8266` | **DONE, proven, live-run green** |

### B core, in detail (`5f49b4cfd`)
- **New leaf package `platform/orchestration/agenterrors`**: `Write` (byte-compatible
  13-column INSERT, best-effort, **returns whether the row landed**), `RecordFindings`
  (returns `(attempted, recorded)` so a lost row cannot read as recorded), `ClassifyError`.
  7 unit tests green.
- **The cycle is retired**: the edge is `platform/orchestration → actions`, so the writer
  moved BELOW both. `orchestration.LogAgentError` is now a thin forwarder and
  `AgentErrorEntry` is a **type alias** of `agenterrors.Entry` — agentbase, messaging and
  the coordinator compile unchanged (verified).
- **`actions` door**: `LogActionError` / `LogActionFindings` in
  `platform/orchestration/actions/log_action_error.go`, identity resolved from
  `ActionParams` the way the proven debt-5b copy did.
- **One exemplar converted**: `retract_page_deployment_action.go`'s
  `insertRetractionConditionRow` is now a call through the helper; its 13-arg + ordering
  pinning tests pass **unchanged**, which is the compatibility proof.

### (d) detector, in detail (`abf5e8266`)
- `cmd/config-key-audit --shared-output-fields` (+ `--ack <file>`), the mode pattern
  `singleowner.go` established. `scripts/audit-shared-output-fields.sh` is the hand-run form.
- Key = same `output_field` + **transitively reachable over the FULL routing graph** +
  **DIFFERENT action**. Graph = `next_step`/`error_step` + **13 config keys**.
  **I re-derived those keys against the live fleet and the RFC's list was one short** —
  there is a **config-level `error_step`** (158 occurrences) distinct from the top-level
  field. Query in the RUNBOOK.
- **Acceptance bar met**: tests fire on `bugs_open/192`'s own pre-fix shape, and the
  mutation test proves the graph is real (sever the `config.then_step` edge → finding
  disappears). Same-action retry loops and mutually-exclusive branches stay silent.
- **Live run 2026-08-06: 176 agents, findings = EXACTLY the addendum's hand census** — the
  2 known hazards, the 5 benign pairs correctly absent.
- **Ack ratchet**: `scripts/shared_output_fields_ack.txt` holds the 2 known pairs with
  reasons; the check exits 0 while only those reproduce, 1 on a NEW pair, and reports
  **stale acks** (an ack that outlives its finding is how a ratchet loosens).

---

## NEXT — three pieces, in this order

### 1. The 18 remaining hand-copied INSERTs (finishes B)
Sites, with their quirks, are enumerated in NOTES. Convert each to `LogActionError`
(or `agenterrors.Write` where there is no `ActionParams` in scope). **Preserve each site's
action/code/severity/message/context EXACTLY** — several carry them as SQL literals, which
become ordinary Go arguments. Sites that previously omitted `orchestration_id` (nine of
them) gain it — that is the point, note it in one line where it happens.
**Tests**: several pin SHORT arity (9/10/11 args) and will fail; update those `WithArgs`
to the canonical 13 (`AnyArg` for the newly-filled identity slots) **without weakening the
assertions that name codes/severities**.
⚠ **Check `git status --porcelain` per file first** — many `actions/` files carry other
sessions' WIP right now, and the package currently does not build for an unrelated reason
(`deploy_image_asset_action.go` signature mismatch, another session's in-flight change).
Use targeted `go test -run`, not the whole package, and say so honestly.

### 2. The online half of (d) — the CronJob
Precedent to follow is **`component-render-check`**, NOT `single-owner-carriers-check`:
ship the Go binary as its own image rather than re-implementing in Python. RFC_006's
Python mirror exists only because a job cannot `go run` from source (262M clone); an image
dissolves that, and with it both of RFC_006's named drift risks (the `DECLARED_*` literal
and the parity test).
- copy `deployments/kustomize/services/component-render-check/` as the shape;
- take a free schedule slot (taken: 02:00, Sun 06:00, 06:20, 06:40, 06:50, 06:55) and say
  which in the manifest;
- **report to `doc_notes`** — one row per run **including clean**, so a missing row means
  THE JOB DID NOT RUN, which must not be indistinguishable from "nothing is wrong";
- exit 1 on NEW findings so the Job shows failed;
- the ack file must reach the image (it is in-repo, so it ships with the build).

### 3. The reader census (gates (a)/(a′))
**Not started** — the delegated agent died on quota mid-sweep. Its one partial finding,
worth keeping: `enrich_fingerprint_with_css` is *wrapper-adapted*, and it had begun a
third sweep for mid-string/condition-expression references two greps would miss.
Deliverable: `architecture_review/CENSUS_2026-08-06_rfc012_await_step_readers.md` — config
side (live `agent_definitions`, queries verbatim so it re-runs) + Go side (mechanically
findable, honestly bounded), each reader marked BREAKS/SURVIVES/UNCLEAR under
merge-not-replace. **The census does not decide (a)** — it is what (a) stays gated behind.
⚠ Two things a naive census will miss: the **`call_agent`/`spawn_agent` branch already
merges under `.response`**, so those readers already tolerate the shape; and dynamic loop
steps **derive** their output_field at runtime (`deriveOutputFieldFromLoopStepName`), so a
config-only census cannot see those keys.

---

## Council + register, still owed for this lane
One council round covers the RFC_012 code set (B + the (d) detector) — **not yet
submitted**; both commits carry no council trailer. Concept-register entries owed for the
`agenterrors` seam and the `--shared-output-fields` mode (the register entry must land in
the same commit as the seam per the standing ruling — that was missed here because the
work was split across two commits; register both when the set is complete and say so).

## Landmines this lane found
- **A nil `Context` map marshals to `null`, not `{}`** — the old writer's nil-guard never
  fired; byte-compatibility means pinning `"null"`.
- **The RFC's own 13-key list is one key short** — trust the live query, not the prose.
- A green live run of the census proves the fleet, never the detector; the mutation test
  is what proves the detector.
