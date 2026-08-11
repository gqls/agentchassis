# PLAN — bugfix 242: a capped render audit must be distinguishable from a complete one

**Lane opened** 2026-08-11. Bug: `bugs_open/242_HANDOFF_2026-08-10_a_capped_render_audit_is_indistinguishable_from_a_complete_one.md`.
Taken after ownership checks: `who-owns 242` shows only the filing commit; live-transcript
grep shows only passing citations. Bug re-verified valid on the current tree and on live
rows dated 2026-08-11 (see NOTES).

## The decision that shapes everything: §4 was already answered

Bug 242 §4 marked the mechanism `[UNVERIFIED]`. It is neither unverified nor new:
**RFC_012 addendum 2** (2026-08-04) established it first-hand, and the owner **ruled on the
class** on 2026-08-06:

- `persistAwaitingStateWithRetry` (`coordinator.go:2067`) loads FRESH state from the DB at
  park time and copies across only `AwaitedRequests` + `Status` — every in-memory
  `CollectedData` mutation made during the step, including `storeActionResult`'s own writes
  (`:1873-1877`) and therefore the action's `Metadata`, is discarded. Both callers skip
  their own persist ("state was already persisted", `:941-948`, `:1472-1476`).
- Re-verified by this lane 2026-08-11 on live rows: all awaited render-audit runs hold
  exactly `{response, response_status, response_received_at}` under BOTH the step key
  (`audit`) and the output field (`render_audit`); the one non-awaiting run (`skipped`)
  keeps its full result. Query in RUNBOOK.
- **OWNER RULING (RFC_012, 2026-08-06): option B, DB-backed.** Findings that must survive
  an await go through the shared `agenterrors` writer (built, live:
  `platform/orchestration/agenterrors/agenterrors.go`), written BEFORE dispatch. The
  merge-not-replace coordinator change (a)/(a′) is explicitly open, gated behind the reader
  census, and is NOT this lane's to take.

Per the 2026-07-31 ruling, this lane states plainly why no `090` run was filed for the
mechanism: the claim is not novel — it is RFC_012 addendum 2's finding, declared there with
a first-hand three-artefact chain in place of a 090 run, ratified by an owner ruling, and
re-verified here against live rows with the code read in full at every link. A 090 round
would re-derive a decided question (and this exact symptom has a 4×-UNVERIFIABLE 090
history: 242 §6, 236 §5a/§5b).

**Consequence for the fix:** anything that must reach the stored artefact of an awaiting
step must ride the ADAPTER'S REPLY; anything that must be durably queryable must be an
`agenterrors` row written before dispatch. Fix candidate 4 of the bug ("fix the metadata
path") is out of scope by owner ruling; candidates 1+2 are in, implemented the framework
way; candidate 3 (raise the cap) is mitigation alongside, as the bug itself prescribes.

## The edits (≤5 files of code, one migration)

1. **`platform/orchestration/actions/request_render_audit_action.go`**
   - Add `pages_total` and `truncated` to the request body's `data` block, so the numbers
     the chassis alone knows can come back in the reply — the only artefact that survives
     the await.
   - When the cap bites, write ONE durable row before dispatch through the sanctioned
     RFC_012-B door: `agenterrors.Write` with `ErrorCode: "RENDER_AUDIT_TRUNCATED"`,
     `Severity: "warning"`, context `{pages_total, pages_audited, max_pages, domain}`.
     The existing `logger.Warn` stays (pod-log visibility) but stops being the only record.
   - Keep returning the same `Metadata` (it is correct, merely not persisted today; if
     (a)/(a′) is ever taken it becomes durable for free).

2. **`internal/adapters/browserrunner/render_audit_action.go`**
   - `RenderAuditRequest` gains `PagesTotal int` + `Truncated bool` (json
     `pages_total`/`truncated`).
   - `Summary` gains `PagesTotal int` + `Truncated bool`, both `omitempty`, echoed from the
     request in `Execute`. A reader of `…->'response'->'summary'` then sees
     `pages: 25, pages_total: 27, truncated: true` — the false green becomes
     unrepresentable in the stored artefact, which is bug candidate 1 verbatim.
   - `omitempty` keeps version skew safe in both directions: old chassis + new adapter →
     summary shape unchanged; new chassis + old adapter → extra request keys ignored
     (unknown JSON fields), summary shape unchanged. Either skew degrades to exactly
     today's behaviour, never to a wrong number.

3. **`platform/orchestration/actions/write_render_audit_findings_action.go`**
   - Read `truncated`/`pages_total` from the summary it already consumes; when truncated,
     stamp `truncated: true`, `pages_total`, `pages_audited` into its own result
     (`findings_written`) — which persists durably because this step does not await. This
     is bug candidate 2, and it is parity with this same action's own honest
     `findings_capped`/`findings_dropped` fields (bug §5b's lesson).

4. **Migration: raise the rotation's `max_pages` 25 → 60** on `render-audit-agent`'s
   `audit` step (mitigation alongside, per bug candidate 3; 60 covers the whole current
   fleet — largest site today is 31 deployed pages — and matches `max_items`'s order of
   magnitude). Config is live immediately and older binaries handle a larger cap safely,
   so this has no ordering constraint against the image roll.

Deliberately NOT done, with reasons:
- No change to `persistAwaitingStateWithRetry` / `applyResponseToState` — owner-gated
  behind the RFC_012 census; taking it inside a bug patch is the exact shape the
  guardian seat vetoed in 124.
- No random/rotating page order — the bug's own "do NOT" (converts a stable gap into an
  intermittent one).
- No refusal-to-run when truncated — filing the findings you have is right; the defect
  was silence, not action.

## Tests

- Adapter (`render_audit_action_test.go`): summary echoes `pages_total`/`truncated` from
  the request; both absent (old-shape request) → both absent from marshalled summary
  (`omitempty` pinned, the version-skew guarantee).
- Findings writer (`write_render_audit_findings_test.go`): truncated summary → result
  carries `truncated`/`pages_total`/`pages_audited`; untruncated → keys absent (shape
  unchanged for every existing consumer).
- Action: request body carries `pages_total`/`truncated`; truncation writes the
  `agenterrors` row BEFORE `Producer.Produce…` (assert relative order), and no row when
  the cap does not bite (the no-op case checked, not assumed).
- Build/test against `git archive HEAD` (shared-tree rule) before commit.

## Council + commit

Platform code → council gate before/alongside commit; commit with
`Council-Submitted: <corr>` (never `Council-Reviewed:` on an unread verdict). Commit with
explicit pathspec, docs first (claim), code second, this plan's files named individually.

## Verify live (after the next chassis + browser-runner roll)

Bug §7's query, on a >25-page site — but with the cap now 60, force the case:
run once with a step-config `max_pages` under the site's page count, or pick any site
whose page count exceeds the configured cap. Expect
`summary.pages_total > summary.pages` and `truncated: true`, plus one
`RENDER_AUDIT_TRUNCATED` row in `agent_error_log` for the run. Grade on a site that
genuinely exceeds the cap (bug §7: a 10-page site cannot distinguish a fix from no fix).
