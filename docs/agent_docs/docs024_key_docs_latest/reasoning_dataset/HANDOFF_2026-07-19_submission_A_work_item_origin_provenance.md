# HANDOFF — submission A: record where a work item came from

**From:** the reasoning-dataset thread, 2026-07-19. Read-only for `platform/`;
not implementing this.
**To:** *no active owning thread found* — see §Owner below. **This needs the
owner to assign it**, which is the one thing blocking it.
**State:** three council rounds, all REVISE, converging. Two small answers away
from a plausible approval, both drafted below.
**Plan:** `reasoning_dataset/submission_A_work_item_origin_provenance.json`
(`SUBMISSION_CORR=61105914-fe50-4e23-b36f-70654ed25727`).

---

## What it does

Adds one nullable `TEXT` column, `site_work_items.origin_correlation_id`,
populated at the single INSERT that turns auditor findings into work items
(`write_audit_findings_action.go:657`), plus a partial index, a doc_notes entry,
and a verify/rollback file.

## Why it is worth doing even if nobody ever builds a dataset

**Today nobody can tell whether any judgement agent is any good.** The auditors
run constantly — `content-quality-auditor` 7,119 LLM calls, `visual-design-auditor`
4,032, `site-review-agent` 3,987 — and each call forms a verdict about a page or
a design. Those verdicts become `site_work_items`, which reach real terminal
outcomes (live 2026-07-19: 4,570 `complete`, 278 `needs_human_review`, 236
`unresolved`, 152 `failed`, 50 `wont_fix`).

Nothing connects the two. So an auditor that flags twenty non-issues and one real
defect is **indistinguishable, in the data, from one that flags twenty real
defects**. There is no per-agent precision signal, so nobody can say which
auditor earns its spend or which item types it over-fires on.

That is the standalone argument. Our interest is secondary and we declare it: the
same link makes ~15,000 judgement calls a month joinable to ground truth, which
is what a reasoning dataset needs. If those two motives ever conflict, the
platform one wins — it is the reason this is worth an owner's time.

## Owner — the honest position

`write_audit_findings_action.go` has **no active owning thread**. Its git history
is old and generic (`44cd2e35b` "forking themes", `8f6ba2ddd` "prior to fixing
audit/fix endless loop"); no workstream directory tracks it. The nearest
candidates and why each is imperfect:

- `content_quality_and_internal_linking/` — owns the auditor *agents*, but its
  docs are June-era and it is not obviously active.
- `work_item_completion_integrity/` — active and adjacent (owns work-item
  lifecycle) but this is the *creation* end, not completion.

We have **not** dumped it on either. It is filed here, in the workstream whose
motivation it serves, flagged for assignment. **Assigning it is the ask.**

## What the council said, so it is not re-spent

Three rounds, each pulling different seats as the plan changed shape.

**Round 1** — `editquality` ×3, `tooling_provenance` ×2. All accepted:
- The original UUID column with a parse-and-drop-to-NULL step *"risks silently
  nulling out legitimate correlation ids … while looking like normal unknown
  origin NULLs."* Sharpest objection of the whole exercise. **Fixed:** the column
  is `TEXT` and stores the correlation verbatim, so no parse can fail. Mirrors
  `diagnosis_artifacts.correlation_id`, which is `TEXT` for the same reason.
- Two edits were scope creep onto `run_discovery_checks_action`, a path the
  rationale never established, and one was a no-op field addition. **Fixed:**
  removed; scope is now the one path the rationale supports.
- No doc_notes commitment. **Fixed:** now an explicit deliverable.

**Round 2** — `debug_historian` ×2, both procedural, both accepted:
- No pod-verification step named. **Fixed:** the Go change is inert until an
  image rolls and must be confirmed by grepping the *running pod*, never git or
  the image tag.
- No verify/rollback files for a production migration. **Fixed:** added.

**Round 3** — `reuse_agent` ×2, firing for the first time because the plan now
touches a migration:

1. *"`site_work_items` already has a `batch_id` … show `batch_id`'s actual
   semantics/generation site before adding a parallel provenance column."*
   **We checked, and the objection does not stand — but the reviewer was right
   that the plan asserted it by omission.** `batch_id := uuid.New()`
   (`write_audit_findings_action.go:600`) is a fresh random uuid minted per
   invocation and mapped to **no** correlation or orchestration id anywhere in
   the codebase. It groups the items one audit run produced; it cannot identify
   *which* run produced them, which is the whole requirement. **Action for
   whoever takes this: paste that refutation into the plan verbatim.**
2. Two insert paths now diverge — one provenance-aware, one not — with no
   unification plan. Called *"exactly the 'fork' pattern"*, and says the
   deferral is reasonable for scope control but is architecture-level and
   *"should be flagged explicitly rather than left to a doc_notes footnote."*
   **Fair, and unaddressed.** **Action: promote it from doc_notes into the plan
   body**, stating the deliberate position — discovery-check items come from
   deterministic predicates with no LLM reasoning behind them, so there is
   nothing to join to and no provenance to record. If that reasoning is wrong,
   the fork is real and needs a plan.

Those two edits are the whole remaining delta. A fourth round would likely
approve.

## Risks the plan already carries

`ALTER TABLE ADD COLUMN` of a nullable `TEXT` with no default is metadata-only on
modern Postgres — confirm the server version and that no long transaction holds a
conflicting lock. Deliberately **not** touched: `idx_swi_dedup` and
`workItemTerminalStatuses` — no status added, no dedup key altered, so the
index/Go-list lockstep that has caused fleet-wide 42P10 is untouched. Provenance
must never enter the dedup key, or the same defect raised by two runs would stop
deduplicating. Migration numbering follows `bugs_open/007`. No backfill:
historical rows have no recoverable origin and inventing one would poison the
join.

## Reading the council reports

```sql
SELECT da.orchestration_id, r->>'reviewer', r->>'verdict', jsonb_pretty(r->'objections')
FROM diagnosis_artifacts da, jsonb_array_elements(da.body::jsonb->'reviews') r
WHERE da.kind='council_report'
  AND da.correlation_id='61105914-fe50-4e23-b36f-70654ed25727'
  AND r->>'verdict' <> 'approve'
ORDER BY da.created_at;
```
