# START HERE — work_item_completion_integrity

**Cold-start entry point.** Rewritten end of 2026-07-20 so a fresh chat can resume
without re-deriving anything. This file is current-state and gets rewritten; the
history lives in `NOTES` (technical, append-only), `README_where_we_are.md` (plain
prose, append-only, the owner's) and the `SUMMARY_*` series (never edited).

**Remit of this thread, in one sentence:** whether a `site_work_items` row can be
trusted to mean what it says.

---

## 1. State in 30 seconds

| | |
|---|---|
| **Branch** | `085_debug_and_feature_loops` |
| **Deployed** | **v1.0.1140**, pod `agent-chassis-5567d99bd6-5snzn` (started 2026-07-20 17:58:20 UTC) |
| **`bugs_open/017`** | ✅ **CLOSED & LIVE** — moved to `/bugs_closed/017_…` |
| **`bugs_open/021` §2** | ⚠️ **WORKED & LIVE, still OPEN** — machinery shipped, but net verifiers are **1 of 86** |
| **`bugs_open/032`** | ✅ Closed by `empty_sections_loop_integrity` (`ed1e20602`). Not ours |
| **In flight** | **nothing.** No dispatches pending, no uncommitted work, no background jobs |
| **Next** | **Submission A** (owner-assigned, unstarted) — §4 |

**Everything is committed.** Nothing is trapped in the previous chat.

## 2. What has shipped, and what it does

### `017` — a failed saga can no longer be stamped 'complete'

The root confusion is **two `status` fields one layer apart**. `result.response_status`
is **delivery** — the coordinator sets it to `'complete'` whenever *any* reply arrives
(`coordinator.go:2398-99`). `result.response.status` is the saga's own **verdict**. The
completion path read the first as if it were the second, so a workflow that never ran
was stamped complete beside the error proving it.

- `handlerReportedFailure` blocks completion on an explicit `failed`/`failure`/`error`
  verdict, routing into the **existing** attempt machinery via a generalised
  `failUnverifiedCompletion`. Runs *before* the per-item-type verifier.
- `recordUnknownVerdict` — an unfamiliar verdict still COMPLETES (a novel status is not
  evidence of failure) but is written to `agent_error_log` as
  `error_code='UNKNOWN_HANDLER_VERDICT'`, because a `zap.Warn` dies with the pod.
- `registry_parity_test.go` — the build fails if an action registers an
  `ActionInputSpec` with no `GlobalActionRegistry` entry.
- The dead `actioncheck.LocalActions` map is deleted, with the comment and two guide
  docs that told authors to "register in TWO places".
- 54 mis-stamped rows corrected to `failed`, reversible via `result._correction`.

### `021` §2 — verifiers are now *writable*, and the gap is *sensed*

`CompleteWorkItemAction` consults a per-item-type verifier before stamping complete.
`RegisterVerifier` had been called **once** for 86 item types.

- **`VerifyTarget{ItemID,SiteID,PageID,ItemType,Spec}`** replaces the bare spec.
  This was the real blocker, and `021` had mis-diagnosed it as opt-in discipline:
  only **9 of 5,514** specs carry a `site_id`, so site-scoped verifiers were
  *unwritable*, however willing the author.
- **`verifier_coverage_test.go`** — every item type must be verified or classified
  (`mechanical` / `creation` / `judgement` / `no_target`) with a reason, or the build
  fails. Two halves: a **source-scan sensor** over `ItemType: "literal"` (needs no
  refresh) and a **hand-refreshed DB snapshot** for types produced outside the package.
- **`ctaClassifyAnchor`** extracts the misdirected-CTA predicate into one definition so
  detection and verification cannot drift. Behaviour-preserving; it also gave that
  check its first test coverage of its core classification logic.

## 3. What is NOT done, stated plainly

**Net verifiers: 1 of 86.** The machinery is live; the coverage is not. This is the
council's standing objection and it is correct — `CompleteWorkItemAction` still stamps
complete with zero verification for 85 item types.

**The `page_rerender` verifier is written, tested and deliberately HELD** — see §5
trap 1. Do not simply un-hold it.

**No behavioural evidence yet.** Zero work items completed platform-wide in the first
minutes after the v1.0.1140 roll, so the widened contract is *present* and *not
observed running*. It is behaviour-neutral by construction, so what would surface is a
regression, not a success.

## 4. Next: Submission A — work-item origin provenance (**owner-assigned**)

`HANDOFF_2026-07-20_submission_A_work_item_origin_provenance.md`

One nullable `TEXT` column `site_work_items.origin_correlation_id`, populated at the
single INSERT in `write_audit_findings_action.go:657`, plus a partial index. **Three
council rounds, all REVISE, converging — "two small answers away", both drafted in the
handoff.** Confirmed genuinely unstarted (column absent; identifier nowhere in
`platform/`).

Why it matters standalone: auditors make ~15,000 LLM judgements a month that become
work items with real terminal outcomes, and **nothing links the two** — so an auditor
flagging twenty non-issues is indistinguishable in the data from one flagging twenty
real defects. The filing thread declares a secondary interest (dataset ground truth)
and states the platform motive wins if they conflict. It documents how to decline.

**Then: write a real verifier.** Start from the guard's gap map — but its categories are
`[INFERRED]` and **read the handler's remit first** (§5 trap 1).
`hardcoded_section_colors` is the flagged candidate, newly unblocked by
`VerifyTarget.SiteID`. `submission_B`
(`66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9`, 2 rounds, both REVISE) is a starting point,
not a finished plan: its `phantom_internal_link` edit is "a stub dressed as an edit",
and its `hardcoded_section_colors` verifier was unimplementable under the old contract.

## 5. Traps this thread paid for — do not re-learn these

1. **A verifier asserts the HANDLER did its job — read the handler's remit, not the
   detector's predicate.** `page_rerender` looked ideal (1,849 of 4,644 completions,
   `page_id` on 1,914 of 1,929). Its handler only rewrites CTA fields in six component
   types; a prose misdirect is *deliberately* left for two-strike escalation. A
   whole-page verifier would mark correctly-handled items unresolved and strand them in
   `failed`. **Six tests passed** — they tested the predicate I chose, not the one the
   handler implements. → `WRONG_CALLS.md` 2026-07-20.
2. **The pod-grep passes on a string your change merely USES.** `grep -c
   fix_forced_text_colors` returned 1 *before* the fix too. Grep a literal from the
   changed line (`"Strip forced child-text colours"`, the widened SELECT), plus a
   positive control. → 016b §9.
3. **A queued orchestration is indistinguishable from a dropped one.** No
   `orchestration_state_audit` rows meant *queued* — ~18 min under backlog vs ~10 s
   quiet. Resubmitting cost three redundant council runs. Ask when **other**
   orchestrations started. → 016b §9.
4. **"It rests on an author-run audit" from a reviewer is a defect report**, not
   box-ticking. I waved that off twice; the claim held by luck, not method.
5. **An inbound handoff is a claim about the PAST.** Verify its state before acting on
   it *and* before forwarding it — three times this session a doc I was about to write
   was already stale.
6. **"The check passed" is not "I complied".** `check_append_only_docs` fires at ≥20
   lines lost; I broke the never-overwrite SUMMARY rule by 7 and 1, and it caught
   neither. Same shape as trusting green tests that exercised the wrong rule.

## 6. Verification commands (full set in `RUNBOOK_…md`)

```sql
-- 017's defining query. MUST be 0.
SELECT count(*) FROM site_work_items
WHERE status='complete' AND result->'response'->>'status'='failed';

-- has the 017 guard ever blocked in production? (0 as of the v1.0.1140 roll)
SELECT id, item_type, status, attempt_count FROM site_work_items
WHERE error LIKE 'completion blocked: handler saga reported failure%';

-- an unfamiliar handler verdict appeared → widen the allowlist
SELECT * FROM agent_error_log WHERE error_code='UNKNOWN_HANDLER_VERDICT';

-- verifier coverage today (2026-07-20: 4,644 / 5)
SELECT count(*) FILTER (WHERE status='complete') AS complete_total,
       count(*) FILTER (WHERE status='complete' AND result ? '_verification') AS verified
FROM site_work_items;
```

```bash
# coverage report by category — never fails, prints the shape of the gap
go test ./platform/orchestration/actions/discovery_checks/ -run TestVerifierCoverage -v

# deploy check: a literal from a CHANGED line + a positive control
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "COALESCE(spec, .{}.::jsonb), site_id, page_id"'
```

> Note: `RegisteredVerifierItemTypes` greps **0** in the binary and that is CORRECT —
> it is called only from a test, so the linker strips it. Do not read it as a failed
> deploy.

## 7. Council state

| run | scope | verdict |
|---|---|---|
| `319e23f6-b333-42ba-88ef-069b4426c057` | 017 | r1 REVISE → r2 **REVISE, 8 approve / 2 object** |
| `9f7bd637-081f-45c4-bf10-1f9645424ce8` | 021 §2 | **REVISE, 9 approve / 2 object** — deciding objection acted on in `c46e57bea` |

**No `Council-Reviewed:` trailer is claimed on any commit** — the trailer is earned by
APPROVED only. Surviving objections on 021 §2: net verifiers still 1, and the
blast-radius claim rests on my own grep rather than an independent check.

## 8. Reading order for a cold start

1. this file
2. `PLAN_…md` — design decisions **and their reasons**, plus corrections to the
   original bug reports
3. `NOTES_…md` — technical log, **five recorded missteps**, newest at the bottom
4. `SUMMARY_2026-07-18` then `SUMMARY_2026-07-20` — read in order; the series shows how
   the understanding moved and is never edited
5. `HANDOFF_2026-07-20_submission_A_…` — the next piece of work
6. `RUNBOOK_…md` — commands, each with its gotcha
7. `/bugs_closed/017_…` and `bugs_open/021` — the cases, with evidence inline
