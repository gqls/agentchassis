# PLAN 2026-09-02 — bugs_closed/314's residual: the fourth copy of the migration vocabulary

**Status:** implemented, controlled, awaiting council verdict before commit.
**Scope:** `scripts/pattern-check.py` only. One predicate corrected, one drift guard added.
**Predecessor:** `PLAN_2026-08-19_314_council_scope.md` (the original fix, closed 2026-08-20).

---

## 1. What is being fixed, and what is NOT

`bugs_closed/314` stays closed. This is the residual its own close-out banner named:

> **A FOURTH copy of the migration vocabulary exists and is already drifted:**
> `scripts/pattern-check.py:384-385` … **A candidate for its own small ticket.**

The banner's *other* residual — "tooling is still out of scope" — has been largely closed by two
later owner rulings (`cmd/config-key-audit/` 2026-08-23, `scripts/pattern-check.py` 2026-08-24).
**Do not quote that banner as current.** Verified 2026-09-02: `COUNCIL_SCOPE_CODE_RE` now carries
both, and `098:105`'s `SCOPE_PATHS` carries both too — the drift trap this lane documented in
August did not fire on either widening.

## 2. The defect, stated as a mechanism

`scripts/pattern-check.py` runs from `.githooks/pre-commit` on every commit in every session. One
of its checks, `check_unguarded_migration_insert`, implements the commit-time half of
`bugs_open/007` Class C: a migration whose INSERT carries no guard cannot be replayed, and a replay
dies on a raw 23505 that reads as broken SQL (migration 151 blocked the runner for **3 days**).

To run, it must decide which changed files are migrations. It did so with
`^\d{3}_[a-z0-9_]+\.sql$` — **lowercase only** — while the runner's appliable rule
(`run-migrations.sh:283` + `SIDECAR_RE:65`) accepts capitals. The file's own comment states the
contract it was breaking: *"lint_idempotency in run-migrations.sh — **SAME semantics, keep them in
step**"* (`:373-379`).

**Two failures, not one:**

1. **Coverage.** Any appliable migration with a capital in its name was skipped silently.
2. **Reasoning.** The comment `# sidecars (_ROLLBACK etc.) excluded` was *true for the wrong
   reason* — sidecars were excluded only because they happen to be uppercase. That is a proxy
   standing in for a rule, which is bug 314's own defect class, and is the same mistake the
   council's `editquality` seat caught **inside 314's own fix**. It was sitting in this file the
   whole time.

## 3. Measurements (all `[MEASURED 2026-09-02]`, corpus grew 1,132 → 1,140 during the session)

| | |
|---|---|
| appliable by the runner | 743 |
| visible to the lint (pre-fix) | 738 |
| **blind spot** | **5**, all already `DO`-guarded → **latent, not live damage** |
| `_HOLD` → plain renames, 2026-08-01..08-31 | **37 events / 26 distinct files, 26 of 26 stuck on disk** |
| newly visible after the fix | **44** (38 `_HOLD` + 6 appliable) |
| **NEW findings on the existing corpus** | **0** |
| **files lost from coverage** | **0** |

⚠ The first census said **660** and was wrong — `comm` with inputs its own collation rejected.
`WRONG_CALLS.md`, 2026-09-02.

## 4. The design decision that mattered: `_HOLD` is IN

The naive rule — mirror the runner's appliable set — **excludes `_HOLD` and would have written the
worst case out of the lint.**

`run-migrations.sh:245-250` **refuses to `--record-only` a sidecar**, by design. So a `_HOLD`
*cannot* be ledger-recorded while it carries the suffix, and the house sequence is forced:
**hand-apply → rename to drop the suffix → record.** Between the rename and someone remembering to
record, the runner sees a pending, unrecorded, appliable file and **replays it**.

So a `_HOLD` is the one category *guaranteed* to be applied out of band before the ledger can know.
The `dispatch_throughput` lane, which argued the other side and then withdrew it, put the
consequence best: an unguarded `_HOLD` is **not a maybe-risk, it is a scheduled appointment**.

**Timing follows from that:** the lint fires at write time, which for a `_HOLD` is the only useful
moment — by the rename commit the file has already been run against production and the diff is
R100 bookkeeping nobody re-reads. `pattern-check.py:378-379` states this purpose itself.

**The boundary, stated so a reviewer does not blur it:** coverage widens to `_HOLD` (a trailing
uppercase suffix that **is the change**), not to the true sidecars `_ROLLBACK` / `_VERIFY` /
`_SUPERSEDED`, which are hand-run against an already-decided state.

## 5. Why this does NOT share code with `scripts/council-scope.sh`

That file enumerates the same three suffixes and today selects the same set. It answers a
**different question** — *"is this the change, for review purposes?"* — where this one asks *"could
the runner ever execute this on replay?"*. Collapsing two questions because their answers currently
agree is the defect 314 exists to remove, and the council caught precisely that inside 314's fix. A
future suffix could be the change without ever being replayed, or the reverse. **Derived
independently, with the convergence noted in a comment.**

## 6. The framework half — the copy gets a watcher

> **CORRECTED 2026-09-02, after this section was written and its design already built and
> mutation-tested.** This section described a Python drift guard inside `pattern-check.py`,
> modelled on `council_scope_drift_warn()`: read the runner, assert its lines have not moved.
> **That design was wrong and is not what shipped.** The predicate was never drift — it was
> **wrong at birth** (runner gained `[A-Za-z]` 2026-07-20 `a51333fd7`; lint written lowercase-only
> 2026-07-25 `9d95e1c31`), so a guard watching the runner for CHANGE sits green for ever. What
> shipped instead compares **the two literals** and pins **the decisions**:
> `cmd/config-key-audit/migration_lint_predicate_parity_test.go` (4 tests) reached from
> `scripts/check-migration-lint-parity.sh` when either source file is staged, using the existing
> `scripts/lib/precommit-gotest.sh`. Three precedents in that package; the reuse seat has already
> objected once to a second hand-rolled guard there (corr `be252395` r2). The section below is
> kept, not deleted, because the wrong turn is the part worth reading — see NOTES and
> `WRONG_CALLS.md`.



Four copies of this vocabulary exist. Copies 2 (`council-scope.sh`) and 3 (`098`'s `SCOPE_PATHS`)
are watched. **Copy 4 was not, and copy 4 is the one that drifted** — for long enough that nobody
can say when. So: `check_migration_vocabulary_drift`, registered in `CHECKS`, asserting the
runner's two source lines are still present verbatim.

**Unconditional, not gated on "did this commit touch either file".** A gated check sees drift only
as it is *introduced*; it cannot see drift already present — which is this bug's entire state, and
why it persisted. Cost is one small file read per commit; it fires zero times while the copies
agree, and every commit while they do not, which is the right volume for a defect that silently
disables a check.

**Considered and rejected: a Go parity test** in `cmd/config-key-audit/`, the estate's other idiom
for this (`optional_budget_cron_parity_test.go` + two siblings, reachable from pre-commit via
`check-optional-key-parity.sh`). Stronger where it applies, because it blocks — rejected here only
because it puts the guard in a different language and area from the constant it guards, for a
single regex, when this file already runs on every commit in every session. **If a second constant
ever needs guarding, prefer the Go idiom and move this with it.** Recorded so the next author
inherits the reasoning rather than the conclusion.

## 7. Controls — and the one that actually proves anything

Harnesses in the session scratchpad; pre-fix baseline pinned at **`b6c4311`** (a `git show HEAD:`
baseline would have expired the moment I committed).

| control | result |
|---|---|
| REGRESSION: ordinary lowercase migration still linted | ✅ |
| POSITIVE: mid-name capital, appliable (`482_ROLLBACK_…`, `582_…_A_…`, `637_…_B_…`) | ✅ |
| POSITIVE (the widening): live `_HOLD` files | ✅ |
| NEGATIVE: `_ROLLBACK` / `_VERIFY` / `_SUPERSEDED`, incl. stacked (`_HOLD_ROLLBACK`, `_ROLLBACK_SUPERSEDED`) | ✅ refused |
| NEGATIVE: `README.md`, no-prefix, 1-digit prefix | ✅ refused |
| **DISCONFIRMING: the same synthetic unguarded files against the PRE-FIX module** | ✅ **missed by pre-fix, caught now** |
| blast radius over the live corpus | ✅ **0 new findings, 0 coverage lost** |
| **MUTATION: induce each drift, require the guard to FIRE** | ✅ 3/3 fire; control silent; runner-absent silent |
| mutations proven to be real edits, not silent no-ops | ✅ asserted in the harness |

The disconfirming and mutation rows are the load-bearing ones. The negatives only prove the check
was not deleted; **only the pre-fix comparison proves the change did anything**, and **only the
mutation proves the guard is not merely silent**. Both were run, not reasoned about — the harness
was first executed BEFORE the edit, where it correctly reported "no-op", which is what proves it
could have come out either way.

## 8. Process

- **No `090` run**, per the owner ruling of 2026-07-31's named escape hatch, and for the same
  reasons 314 §7 gave: the mechanism is two predicates read at source, the population is one
  reproducible `grep` loop, and there is no structural claim needing a loop. First-hand
  verification substituted, deliberately and stated.
- **No RFC.** Owner ruling 2026-07-29 §1: an RFC is owed when a shared mechanism's *guarantees*
  change. This corrects a detector's file predicate. Nothing about the lint's contract, output
  shape, `CHECKS` interface or advisory status changes. Owner ruling 2026-08-02 §2 (new authority
  on a shared seam ships opt-in, default OFF) does not apply — no new authority, and the check
  cannot block.
- **Council gate: yes**, and it is admitted without `FORCE=1` only because of the 2026-08-24
  widening. This lane's own August fix is what made its follow-up reviewable.
- **Consumers told, not merely measured** (owner ruling 2026-07-29 §3): `dispatch_throughput`
  messaged before any edit; the exchange corrected both of us and produced `bugs_open/426`.
