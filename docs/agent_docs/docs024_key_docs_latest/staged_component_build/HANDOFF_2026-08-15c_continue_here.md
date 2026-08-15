# HANDOFF — 2026-08-15 (c), fresh chat starts here: the council said REVISE, the objection is RIGHT, and the revision is scoped but NOT built — persist the two WARNs or the observation window cannot be read

**Supersedes `HANDOFF_2026-08-15b_continue_here.md`** (whose §1 "what is now true" and §3
traps all still hold — do not re-derive them). What changed since: the council verdict for
RFC_029 Phase 1 was **read**, it is **REVISE**, and the revision was scoped and design-probed
but **not implemented** — the owner asked for a handoff instead mid-build. This file is that
handoff. The seat-by-seat assessment lives in NOTES (bottom, `## 2026-08-15 (later)` +
verdict entry); the implementation record is still RFC_029 §10.

## 1. The verdict, and the one defect that gates everything

**REVISE** — run `ae2a88a7`, corr `75091072-9d65-433e-8a30-84719dc3f30f`, completed 14:10Z,
decided by a GATING objection from `reuse_agent` (HIGH). Approve: architecture, constitution,
mission, guidelines. Object: editquality, reuse_agent, tooling_provenance, guardian,
debug_historian, prior_art_librarian. 7 abstained.

**The gating objection is correct and is a real defect in what shipped:** both Phase 1 WARNs
(`aggressive search: conflicting candidates`, `aggressive search: explicit single-segment
mapping bypassed`) are plain zap log lines, and chassis pod log retention is measured at
~90 seconds — so the 48h+ observation window that the whole Phase-1→Phase-2 plan rests on
**cannot be read after the fact**. The platform's own remedy already exists:
`agent_error_log` via `platform/orchestration/agenterrors` (RFC_012's leaf package).

## 2. THE REVISION TASK — scoped, design-probed, ready to build

One coherent council-gated task. Everything below was verified THIS session (2026-08-15
~16:00–16:45Z); trust it over re-derivation, but re-run `git log --oneline -10` first —
this tree moves in hours.

### 2a. Persist both WARNs (the gating fix)

- **The write path, already read:** `agenterrors.Write(ctx, db, logger, Entry) bool`
  (`platform/orchestration/agenterrors/agenterrors.go`) — best-effort, returns whether the
  row landed, nil-db returns false. The package is a TRUE LEAF (imports only
  context/sql/json/strings/zap), so **datahelpers CAN import it — no cycle.**
- **The missing piece is the DB handle:** `findFieldRecursive` and `ExtractActionInputs`
  have no `*sql.DB` and no `ctx` in their signatures, and threading one through every call
  site fleet-wide is not on. Design: a **package-level registered sink in datahelpers**
  (e.g. `SetResolverFindingRecorder(...)`), nil by default (log-only = today's behaviour,
  opt-in default-OFF per the 2026-08-02 §2 shape), registered ONCE at chassis startup where
  the DB pool and pod identity live. Find the registration point by grepping who calls
  `orchestration.LogAgentError` / constructs the pool — agentbase startup is the likely
  spot; the coordinator and messaging layer already call the forwarder, so the pool exists
  there.
- **Row shape:** `ErrorCode` `RESOLVER_CONFLICTING_CANDIDATES` / `RESOLVER_MAPPING_BYPASSED`
  (match the existing SCREAMING_SNAKE style, e.g. `NO_CHANGE_GATE_UNREADABLE_RESULT`);
  `Severity: "warning"`; `Context` jsonb carrying field, candidate_paths, winner_path /
  reference. Per-call identity (orchestration_id, step) is NOT reachable from the resolver
  without threading — ship without it, say so in the row's context, and let pod_name +
  agent_type (known at registration) carry attribution. **Write EVERY occurrence — no
  dedup:** frequency is the data; §9's disconfirmation clause needs the population size.
- **Keep the log lines.** The rows are the instrument; the lines stay for live tailing.
- ⚠ Two standing landmines for whoever queries these rows: `agent_error_log.domain` —
  "no domain" is `COALESCE(domain,'')=''`, never `IS NULL`; and the resolver rows will have
  empty orchestration_id by design (above).
- **Tests:** with a registered recorder, conflict + bypass each produce exactly one row per
  occurrence (assert via a fake recorder — do NOT mock the SQL); with none registered,
  behaviour is byte-identical to today (the default-OFF control). Arm-budget floor stays
  10/15 outer, 5/8 inner — the recorder writes no `result.Values`.

### 2b. The paper half of the revision (answers the other five seats)

- **doc_notes row for the Phase 1 mechanism itself** (tooling_provenance): subject the
  resolver/RFC_029, body recording what shipped (commits `927e12bd9`/`1806371ef`), the two
  ErrorCodes above, and that the window OPENS AT THE ROLL — this row doubles as the
  in-DB evidence `prior_art_librarian` said it cannot see (owner rulings are invisible to
  council seats — known gate landmine; quote §9's key lines in the body).
- **Migration 417's header gains two lines** (debug_historian): (1) the two-active-rows
  trap was MEASURED NOT APPLICABLE — `image-build-handler` has exactly 1 active row
  (checked 2026-08-15 ~16:15Z); (2) before applying, confirm which `snapshot_agent`
  overload the call hits via `pg_get_functiondef` (two overloads write to different
  tables — the 402 rollback note documents this; the check is still UNRUN).
- **Resubmission answers, no code needed:** editquality's two "missing D4" items were in
  fact SHIPPED in `1806371ef` (extractSingleField renames + migration 402's dated
  correction) — the submitted plan just failed to list them; name them as edits this time.
  guardian's "winner changes now" — that is §9 D2's explicit owner-delegated choice; cite
  it. tooling_provenance's ledger check — DONE: `schema_migrations` shows `416_auditor...`
  applied (the other lane's), 417 unclaimed; put the query result in the submission.
- **Update RFC_029 §10** with a dated revision note (verdict, what the revision adds), and
  NOTES/README as you go.

### 2c. Resubmit, then commit

Resubmit with the trail accumulating:
`RESUBMIT_CORR=75091072-9d65-433e-8a30-84719dc3f30f ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <revision.json>`
(the previous submission JSON is at the old session's scratchpad and is NOT recoverable —
rewrite from this file + RFC_029 §10; ≤8 edits). Commit per task with pathspec, trailer
`Council-Submitted: <corr the script prints>`. **Read the new verdict** — do not let this
file's successor say "verdict read owed" again without a date.

## 3. After the revision: the 15b worklist, amended in ONE place

15b §2's order stands (roll verification → window → apply 417 → Phase 2), with one
amendment: **the observation window is read from `agent_error_log` rows, not log greps** —
```sql
SELECT error_code, context->>'field' AS field, count(*),
       min(occurred_at), max(occurred_at)
FROM agent_error_log
WHERE error_code IN ('RESOLVER_CONFLICTING_CANDIDATES','RESOLVER_MAPPING_BYPASSED')
GROUP BY 1,2 ORDER BY 3 DESC;
```
(Adjust the timestamp column name to the real schema — `\d agent_error_log` first.) The
window's clock starts at the roll of the REVISED build, not Phase 1's.

## 4. Traps found reading the verdict (cheap, easy to lose)

- **`kubectl exec -i` with nothing piped to stdin HANGS until timeout.** The "postgres
  flakiness" recorded at the end of the afternoon session was OUR OWN `-i` flag — the pod
  was fine the whole time. Drop `-i` unless you are piping a heredoc.
- The payload-filter query on `orchestration_states` (the runbook's own
  `collected_data->'input_data'->>'fix_correlation_id'` form) seq-scans and times out —
  **narrow it with `owner_agent_type='council-gate' AND created_at > now()-interval '6
  hours'`** and it returns instantly.
- Council artifacts: `diagnosis_artifacts.body` (there is no `content` column);
  verdict summary at `metadata->>'decision'`, full seat-by-seat JSON in `body`, keyed by
  `correlation_id` + `kind='council_report'`.

## 5. Session-start checklist

1. `git log --oneline -10`; re-read THIS file from disk.
2. Read the council report yourself before building — it is the spec for the revision:
   `SELECT body FROM diagnosis_artifacts WHERE correlation_id='75091072-9d65-433e-8a30-84719dc3f30f' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;`
3. Build §2 top-down (2a is the gate; 2b is cheap; 2c closes the loop).
4. Everything else about this lane is DONE — 15b §1 records what is true, RFC_029 §10
   records how it got that way, NOTES records the missteps. Do not reopen them.
