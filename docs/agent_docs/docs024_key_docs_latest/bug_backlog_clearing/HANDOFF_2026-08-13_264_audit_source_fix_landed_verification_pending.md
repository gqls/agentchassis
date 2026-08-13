# HANDOFF 2026-08-13 — bug 264 (four auditors' audit_source) fixed and
# committed; live verification + council verdict are what's left. Also carries
# forward 040's still-open time-gate.

Continues the backlog-clearing arc that picked up after
`HANDOFF_2026-08-12c_040_time_gated_signal_next_backlog_pickup.md` (040's own
next step was time-gated to ~2026-08-13 15:00Z, so that session moved to other
backlog items — see "Still open from 040" at the bottom). This file is written
at a genuine pause point, not a crisis: the code is committed and the config is
live; what remains is confirmation, not more design work.

## What was picked up, and why

Triaged the open-bugs backlog for an unowned, small-to-medium, no-hard-blocker
candidate (survey covered `bugs_open/237,254,256,257,258,263,264,265,269` —
`who-owns.py` returned "(none identified)" owning workstream for most of these).
Picked **`bugs_open/264`** — four auditor agents' `audit_source` config landing
as a plain string that can never resolve, so all four auditors' findings were
silently stamped `design-audit` fleet-wide, all-history. Real, already-proven
correctness impact (a different lane, `copy_quality_two_stage`, wasted a day on
2026-08-12 concluding the copy auditor's findings were unconsumed — they were
being consumed the whole time under the wrong name), small-to-medium scope, no
DB/cluster/billing blocker, no competing owner.

## State, precisely

1. **Root cause, fully diagnosed already by the bug file itself before this
   session started** — read `bugs_open/264` in full for the mechanism. One-line
   version: `ExtractActionInputs` treats every string in step config as a
   REFERENCE to resolve against `collected_data`, never as a literal (deliberate
   — see `write_audit_findings_action.go`'s Strategy 5 comment). Four agent
   definitions configured `audit_source` as a plain string with no dot and no
   matching top-level `collected_data` key, so all four silently fell through to
   the `Defaults` entry `"design-audit"`.

2. **Candidate 1 (config-only, no roll) — APPLIED AND VERIFIED LIVE.**
   Migration `docs/agent_docs/sql_for_agents/399_four_auditors_audit_source_resolves_to_a_real_value.sql`
   (+ `_ROLLBACK.sql` sidecar) applied against `clients_db` 2026-08-13. Adds one
   `query_database` step per agent with **no `FROM` clause**
   (`SELECT '<name>'::text AS audit_source`) whose `output_format:"object"`
   flattens the literal into a new `collected_data` field
   (`audit_source_literal`). The write step's `audit_source` config becomes the
   dot-path `audit_source_literal.audit_source`, which Strategy 0 resolves
   normally. Re-read live immediately after applying — confirmed all four
   `write_findings`/`write_strategic_findings` steps now carry the dot-path and
   all four new `set_audit_source` steps are present with the right literal.
   The migration's own guard (asserts the new step, the rewired `next_step`, the
   updated config, AND that unrelated sibling keys survived) passed on both a
   `ROLLBACK`-wrapped dry run and the real apply.

3. **Candidate 2 (make the field non-defaultable) — APPLIED, COMMITTED, INERT
   UNTIL THE NEXT ROLL.** `platform/orchestration/actions/write_audit_findings_action.go`:
   `audit_source` moved from `Optional` (with `Defaults:
   {"audit_source":"design-audit"}`) to `Required`, no default; the action
   body's own second, redundant empty-string fallback (which would have
   silently defeated the `Required` change) removed. `go build`/`go test` on
   `./platform/orchestration/...` both green. **Confirmed** (by reading the call
   site, not inferring) that this cannot affect the fifth `audit_source`
   producer, `tool-acceptance-tier4` — it sets the field directly in a
   Go-constructed spec map in `tool_acceptance_actions.go`'s
   `routeChromeFailures`, bypassing this action and `ExtractActionInputs`
   entirely.

4. **Ordering is deliberate — do not let a future revert/replay flip it.**
   Candidate 1 had to be (and was) live BEFORE candidate 2 shipped: an older
   binary resolves the new dot-path unconditionally (Strategy 0 doesn't depend
   on the Go change), but rolling the stricter binary FIRST — before all four
   configs were fixed — would have hard-failed every auditor's
   `write_audit_findings` step the moment it rolled, on "missing required
   fields: [audit_source]".

5. **Committed**: `3621ca7cf` — `platform/orchestration/actions/write_audit_findings_action.go`,
   both `399_*.sql` files, and `bugs_open/264`'s own write-up (§12), pathspec-scoped.
   The commit-scope hook flagged an advisory "architecture signal" (migration +
   platform code in one commit) but explicitly said "If it is a point fix, carry
   on" — this is a point fix (4 named agents, no shared-contract/guarantee
   change), not an RFC-scoped change, so no RFC was filed. Judgement call, not
   a gap — revisit only if a reviewer disagrees.

6. **Submitted to the advisory council-review gate** (touches `platform/`):
   `SUBMISSION_CORR=50ee4b26-2303-4304-b437-7320e1368a1d`. Committed under
   `Council-Submitted:` rather than waiting on the ~30-minute queue latency.
   **Verdict not yet read as of this handoff.**

## What to do next

### A. Read the council verdict (should be resolvable within ~30 min of the submit, i.e. shortly after 2026-08-13 ~15:40 BST — check the actual submit time in git log if you need precision)

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='50ee4b26-2303-4304-b437-7320e1368a1d' AND kind='council_report'
ORDER BY created_at;
```
- **APPROVED** → nothing to do; `098`'s coverage report resolves the
  `Council-Submitted:` trailer automatically, no amend needed (forward-only).
- **REVISE** → read the objections (they come with the reviewers' own read-only
  checks already answered), fix, resubmit with `RESUBMIT_CORR=50ee4b26-...`.
- **REJECTED** → read the guardian's veto note and its named safest contained
  alternative:
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;`
  — the code is already on the shared branch (forward-only, no revert), so a
  REJECTED verdict means "fix it forward," not "undo it."

### B. Run the live verification the bug file itself calls for (not yet done)

`bugs_open/264`'s own "How to verify a fix" section, and its own warning: **a
single `design-audit` row proves nothing — that's also what a fully unfixed
system produces.** Trigger one audit run of each of `brief-fidelity-auditor`,
`content-quality-auditor`, `site-review-agent`, `visual-design-auditor` against
any live site, then:
```sql
SELECT spec->>'audit_source', count(*) FROM site_work_items
WHERE created_at > '2026-08-13T00:00:00Z' AND spec ? 'audit_source' GROUP BY 1;
```
Must show **four distinct new values**, not a repeat of the single
`design-audit` row. Write the result into `bugs_open/264` §12 (or a new §13)
either way — a null result (still one row) is itself a finding, not a
non-event, per this codebase's own disconfirmable-check norms.

Once both A and B are clean, `264` is a fixed-AND-live candidate for
`bugs_closed/` per the CLAUDE.md 2026-08-12 bar (though re-check that section —
it has moved more than once).

### C. Then: resume general backlog clearing

Standing four before touching anything new (unchanged from the 040 series):
`who-owns.py`, `git log` on the bug's own named file, live `.jsonl` transcript
grep, `site_work_items` queue check. The earlier triage in this session (see
"What was picked up, and why" above) surfaced two close seconds that were NOT
picked up and remain live candidates: `bugs_open/269` (sibling-signatures bare
method handles — the hard part, a `CanonicalSymbolName` helper, already exists
from a sibling bug; trivial-to-small scope) and `bugs_open/257` (provider
`max_tokens` options-map fix, capped short of its full architectural version on
purpose to stay one-sitting).

## Still open from 040 (carried forward, unchanged)

`bugs_open/040` (kafka dial timeouts) — fix is committed, council-approved, and
proven live across two chassis rolls. Its own next action is time-gated: **do
not re-read the `refused` counter before 2026-08-13 ~15:00Z** (24h into the
fix's actual runtime), and even then treat one more reading as a data point,
not a verdict — see
`HANDOFF_2026-08-12c_040_time_gated_signal_next_backlog_pickup.md` for the full
query and the disconfirming-shape reasoning. **Check the current time before
acting on this** — this handoff does not know whether that gate has passed.

## Files touched this session (all committed, pathspec per task)

- `platform/orchestration/actions/write_audit_findings_action.go`
- `docs/agent_docs/sql_for_agents/399_four_auditors_audit_source_resolves_to_a_real_value.sql`
- `docs/agent_docs/sql_for_agents/399_four_auditors_audit_source_resolves_to_a_real_value_ROLLBACK.sql`
- `bugs_open/264_HANDOFF_2026-08-12_four_auditors_file_every_finding_under_one_producer_name.md` (§12 added)
- `docs/agent_docs/docs024_key_docs_latest/bug_backlog_clearing/` — this file

No `bugs_closed/` move yet — verification (§ B above) has not run.

## UPDATE (same day, later) — round 1 REVISE, answered; round 2 blocked by expired kubeconfig

Read the council verdict per §A above: **REVISE**, not APPROVED — gated by
`editquality`, with `guardian`/`debug_historian`/`prior_art_librarian` also
objecting. Full detail is in `bugs_open/264` §13, not repeated here; short
version: two of the three HIGH-severity objections (a "duplicate active
row/wrong version" landmine) were already refuted by the same review round's
own embedded checks, and the rest (fallback claim ungrounded, no full-caller
inventory, no owning-pipeline naming, migration not ledger-recorded, no
automated test) were real gaps, all closed with fresh evidence — see
`bugs_open/264` §13 for exactly what was checked and how.

**The round-2 resubmission did NOT actually dispatch.** Mid-session the prod
kubeconfig token expired (`2026-08-13 19:05:20` — confirmed by decoding it, see
`~/.claude/…/memory/kubeconfig-token-expires-every-3-days.md`), which broke the
council trigger script's kafka-publish step (`kubectl -n kafka run ... kcat -P`)
with `Unauthorized`. The script prints `SAVE: SUBMISSION_CORR=...` **before**
that publish step, so the correlation it echoed was never real — nothing
reached the council. `kubectl -n ai-persona-system get pods` fails the same
way, confirming total auth loss, not a scoped issue.

**This blocks everything cluster/DB-side**, not just the resubmission: the
live audit-run verification in §B above, re-reading any verdict, and any
further `psql`/`kubectl` work on `040` or any other backlog item. **Nothing
code-side is blocked** — this is a credentials outage, not a defect.

**Only the owner can clear it** (download a fresh kubeconfig from the Rackspace
Spot console — no self-refresh mechanism is configured; `kubectl oidc-login`
would work but the `kubelogin` plugin isn't installed). Once refreshed, re-fire
the round-2 submission with the saved JSON:
```bash
RESUBMIT_CORR=50ee4b26-2303-4304-b437-7320e1368a1d \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/submission_264_audit_source_round2.json
```
then run the §B live verification, then decide on `bugs_closed/`.
