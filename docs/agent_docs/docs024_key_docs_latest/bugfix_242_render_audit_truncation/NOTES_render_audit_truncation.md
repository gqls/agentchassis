# NOTES — bugfix 242 (append-only, newest at the bottom)

## 2026-08-11 — lane opened, bug re-verified, mechanism traced to an already-ruled class

- Ownership: `who-owns 242` → only the filing commit (`a2110d732`) + a 016b §9 entry
  (`48da4b6fd`). Live-transcript grep across the 20 most recent sessions: passing
  citations only. Unowned; taken.
- Re-verified on the tree: `request_render_audit_action.go:157/:160/:252-259` unchanged —
  `truncated`/`pages_total`/`urls_audited` still returned only in `Metadata` on an
  `AwaitResponse:true` result.
- Re-verified on live rows (query in RUNBOOK): 7 of the last 8 rotation runs hold exactly
  `{response, response_status, response_received_at}` under BOTH `audit` (step key) and
  `render_audit` (output field). The eighth (08:02Z) was a `skipped` no-await run and
  keeps its full result — the control that non-awaiting results persist fine.
- Traced the loss first-hand: `storeActionResult` writes in-memory only
  (`coordinator.go:1873-1877`); `persistAwaitingStateWithRetry` loads fresh DB state and
  copies only `AwaitedRequests`+`Status` (`:2073-2102`); callers skip their own persist
  (`:941-948`, `:1472-1476`); reply-time merge preserve-then-adds onto nothing
  (`:2721-2748`).
- **Misstep, caught before it was asserted:** I initially treated that trace as a
  discovery and was preparing to claim it. It is RFC_012 addendum 2 (2026-08-04),
  owner-ruled 2026-08-06 (option B, DB-backed `agenterrors` writer — built and live).
  What caught it: reading bug 236's Related/§6 pointers before writing anything down —
  i.e. "grep before you file", applied to a mechanism instead of a bug. Bug 242 §4 itself
  did not cite RFC_012, which is how close the estate came to a second parallel account
  of one mechanism. Filed in WRONG_CALLS.md.
- Design consequence recorded in PLAN: artefact-visible facts must ride the adapter
  reply; durable queryable facts go through `agenterrors` before dispatch; the
  coordinator merge change is owner-gated (RFC_012 (a)/(a′)) and not this lane's.
- Also relevant prior art read: the RFC_012 reader census
  (`CENSUS_2026-08-07_rfc012_await_step_readers.md`) — request_render_audit is in the
  always-awaits table; and `extractTargetAgentType` returns `"unknown"` (non-empty), so
  every adapter response takes the `isAgentResponse` preserve-then-add branch — the
  wrapped shape is universal for these steps, which is why the reply envelope is the
  right vehicle.

## 2026-08-11 (later) — implemented, tested, mutation-proven, council-submitted

- Edits landed exactly per PLAN (action request fields + agenterrors row before dispatch;
  adapter echo with `omitempty`; findings-writer stamp; migration 392 + rollback).
- Six new tests, all run and passing (`-v` checked — a quiet pass can mean "not
  selected"): the two dispatch-time tests, the two adapter echo/skew tests, the two
  drain stamp/control tests. Full `actions` and `browserrunner` packages green.
- **The order guard was mutation-tested**: moving the `agenterrors.Write` after
  `ProduceWithValidation` fails `TestRequestRenderAuditTruncationTravelsInRequestAndLandsDurably`
  with "the truncation row must land BEFORE the dispatch". Reverted; suite green again.
- The no-op case is asserted via the writer's own guaranteed warn (an attempted write
  against the mock MUST produce "Failed to write to agent_error_log"), not via a mock's
  silence — per the mutate-to-prove-a-guard discipline.
- Council submission: `SUBMISSION_CORR = 700da63e-6c39-4617-ace8-4e450addd472`
  (2026-08-11 ~16:4xZ). Committing with `Council-Submitted:` trailer per the 2026-07-30
  rule; verdict to be read and recorded here when it lands (~30 min budget).
- Migration 392 is COMMITTED BUT NOT APPLIED — it goes through the migration runner
  (dry-run first, scoped dir). No ordering constraint against the image roll: old and
  new binaries both read `max_pages` the same way.

## 2026-08-11 (evening) — council round 1: REVISE, two real finds, both adopted; round 2 resubmitted

Verdict REVISE (decided_by: gating objection from debug_historian; 5 abstained). The two
real defects it found:

1. **reuse_agent was right — I bypassed the sanctioned door.** `log_action_error.go` is
   the actions-package wrapper family onto `agenterrors`, and its header names MY exact
   case: "For findings that must survive an AWAIT — the RFC_012 class — call
   `LogActionFindings` BEFORE the dispatch". Switched: the hand-built `agenterrors.Entry`
   (7 hand-set identity fields) became one `LogActionFindings` call; provenance/join
   columns now filled by the helper. Filed in WRONG_CALLS.
2. **Five seats on migration 392's row targeting.** The WHERE had no version tie-break
   against the two-active-rows landmine, and the 'audit' path was asserted rather than
   evidenced in the submission. Measured: exactly ONE definition row exists (8c99f1b6,
   v1). The file now enforces what I had merely checked: DO/RAISE pre-guard aborts unless
   exactly one active row AND `workflow.steps.audit.config ? 'max_pages'` (jsonb_set can
   never invent a path the loader ignores). Replayed live: guards pass, idempotent 60→60.

Also adopted: the no-truncation test now ends with a positive control of its own
instrument (same mock+observer, cap biting, MUST see the write-failed warn — the earlier
absence is evidence, not blindness). Answered with measurements: single-consumer check
(action registered once, one live workflow names it, RenderAuditRequest decoded only at
adapter.go:294); precedent for the B pattern (098's RETRACTION_AUDIT, live-probed
2026-08-05); class size behind the RFC_012 gap (census: 40 actions / 221 awaited steps).

Round 2 resubmitted on the SAME trail correlation (`RESUBMIT_CORR=700da63e…`), run
orchestration `a659d9cc-6041-47e7-99e4-9d30ceed8329`. All tests green after revisions;
transient shared-tree build break (another session mid-write on
`rerender_single_page_action.go`) resolved itself on retry — not ours.

## 2026-08-11 (evening, later) — round 2 APPROVED; advisory answers recorded

**APPROVED** ("2 advisory objections — none high-severity", 5 abstained), report
15:56Z on trail `700da63e…`. The advisories, answered here so they are not lost:

- **editquality (medium)** — does `LogActionFindings`' inheriting merge overwrite a
  provenance I meant to set? No: my base Entry names only SiteID/Domain/Action; AgentType
  and StepName are deliberately inherited from the running step, which is the same pair
  the round-1 hand-built Entry copied from `ExecutionContext` by hand. Nothing deliberate
  is overwritten because nothing else was set.
- **bug_historian (medium)** — 1 of 40 awaiting actions fixed, the class deferred: yes,
  by owner ruling. The general fix is RFC_012 (a)/(a′), gated behind the reader census;
  the per-action door until then is exactly what this fix uses.
- **debug_historian (low)** — verify form: both DO blocks use `IS DISTINCT FROM`
  (null-safe — the `<>`-goes-green-on-NULL landmine does not apply); no pg_dump was taken
  before the single-key mutation — the ROLLBACK pair and the one-line recoverable value
  (25) are the containment, noted as thinner than full SQL-surgery discipline.
- **architecture (low)** — RFC_022 optional-key tally on `RenderAuditRequest` for the
  future accumulation counter: 7 fields, 3 optional/opt-in (`capture_renders`,
  `pages_total`, `truncated`).

Lane state: code committed (`502b6c194` + `0e4e71674`, both on trail `700da63e`),
migration 392 applied+ledgered, register updated, council APPROVED. Remaining to CLOSE
per the bug's §7: a post-roll rotation run against a site whose page count exceeds the
configured cap (force with a small step-config `max_pages`; nothing exceeds 60 today).

## 2026-08-11 (night) — LIVE ON v1.0.1288 AND PROVEN; close criterion MET

- Deploy proven at the artefact, per service: chassis + render-audit-adapter both
  `v1.0.1288`, stamp `bb534864…` (adapter startup log; chassis binary probe with negative
  control — the log line had rotated, exactly as the landmine warns). Both fix commits are
  ancestors of the stamp; the chassis binary carries the `RENDER_AUDIT_TRUNCATED` literal.
- **Forced-truncation run** (the §7 close criterion; nothing exceeds 60 organically):
  set `max_pages=5` live, dispatched via the proven RUNBOOK_oufe §14 recipe
  (`PUBLISH_OK` seen), corr `a61ef6c5…`, orch `765512d1…`, COMPLETED in ~100s. All three
  artefacts held:
  1. summary: `pages: 5, pages_total: 26, truncated: true`
  2. findings_written: `truncated: true, pages_total: 26, pages_audited: 5`
  3. agent_error_log: one `RENDER_AUDIT_TRUNCATED` warning row, agent_type
     `render-audit-agent`, step `audit`, orchestration_id joining the run, context
     `{max_pages: 5, pages_total: 26, pages_audited: 5}` — which also empirically answers
     editquality's round-2 advisory: the inheriting merge filed the RIGHT provenance.
- Cap restored 5→60 immediately by replaying guarded migration 392 (both pre-guards +
  verify passed; read-back 60). The manual dispatch touched no rotation stamp, so
  loancalculator's weekly rotation slot is unaffected; the one deliberate truncation row
  and the 5-page sweep's zero findings are this test's residue, both honest.
- `agent_error_log.occurred_at` (NOT `created_at`) — schema-first paid off again; noted
  in RUNBOOK.
