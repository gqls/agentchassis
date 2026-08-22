# NOTES — bugs_open/358: agent_error_log finding codes are write-only and expire unread

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

## 2026-08-22 — session start: ownership + re-validation

- `scripts/who-owns.py 358` → no owning workstream; the `bugfix_238_regeneration_key_loss` lane
  cites it as a cross-reference only, and their NOTES (2026-08-22 entry) say 358 was filed AT
  their request with this class deliberately handed to whoever picks it up. No open
  `site_work_items` row touches the class. This lane now owns it; this directory is the record.
- **Census re-run against the live DB (see RUNBOOK), two deltas since the bug file was written
  this morning — neither invalidates the class:**
  1. `resolved` is now **48**, not 0. All 48 set today 10:40 UTC by
     `resolved_by = 'content-loss-check:healed'` / `'content-loss-check:row_gone'` on
     `CONTENT_KEY_LOSS` (40) and `STRUCTURAL_KEY_CARRY_MISS` (8) — the 238 lane's consumer
     shipped (`cba51ad1d`) and ran. So "the resolved workflow has never been used once" is now
     stale BY ONE DAY, and the first user is exactly the reader-ships-with-writer pattern 358
     holds up as the positive example. The §8 trap stands: resolving halves remaining life to
     14 days; content-loss-check extracts (heals) before resolving.
  2. New code `CONTENT_KEY_LOSS` (72 rows, all 2026-08-22, agent_type `content-loss-check`) —
     written AND consumed by the same binary. Not a new member of the unread class.
- Commit `0ce242d9c` (today) added a THIRD recorder in the validation-gate family:
  `CONTENT_VALIDATION_WARNING_DETAIL`. 0 rows yet (inert until image roll). Reader status being
  verified — if unread, the class grew by one TODAY, while the bug file was being written,
  which is §3's self-sustaining mechanism observed live.
- Headline counts still hold: 45,507 total (was 45,426), `RESOLVER_CONFLICTING_CANDIDATES`
  9,617 (was 9,615) and still the loudest, oldest row still 2026-07-23 (retention live),
  `REVIEW_SUPERSEDED_BY_PASSING_SAVE` (25 rows, 07-23 only) days from deletion,
  `TRUNCATION_DEGRADED_REVIEW` dies ~08-25.

## 2026-08-22 — the census method itself is unreliable, which IS the finding

Trying to reproduce 358 §2.1's grep census produced junk on the first two attempts, and that is
worth recording rather than tidying away, because it is the argument for a registry.

- **`grep -ohE 'ErrorCode:\s*"..."'` plus a regex for `*Code = "..."` constants returned 47
  strings including `A`, `CODE`, `FIRST`, `SECOND`, `SOME_CODE`, `X_AUDIT` — test fixtures — while
  MISSING every one of `validationDetailErrorCode`, `linkRepairErrorCode`, `claimsFloorErrorCode`
  and nine more.** The recovery was to stop pattern-matching and resolve each identifier by name.
- **A fourth blindness, beyond the const one 358 §3.2 records:** several codes reach the table as
  a **positional argument** to `LogActionError(ctx, params, siteID, domain, action, code, …)` —
  `tool_birth_instance_scope_refused`, `RETRACTION_AUDIT`, `component_write_shared_blocked`,
  `PLAN_PAGE_SAME_NAME_IDENTITY_HELD`. An `ErrorCode:` grep cannot see any of them.
- **Conclusion that shaped the design:** if three careful attempts at the source census disagree,
  the source is the wrong authority. `SELECT DISTINCT error_code` cannot be got wrong this way.

## 2026-08-22 — CORRECTION 4 to the bug file: `BUILD_DISPATCH_STALLED`'s closed loop is not live

358 §2.2 lists it as one of three codes with an automated reader, evidence for the
"reader-with-writer from birth" law. Both halves are real IN THE FILE
(`214_build_dispatch_watchdog.sql:104` reads, `:120` writes) and **neither is live**:

```sql
SELECT filename FROM schema_migrations WHERE filename LIKE '214%';   -- 0 rows
SELECT name FROM scheduled_tasks WHERE name ILIKE '%watchdog%';       -- 0 rows
```

Migration 214 was never applied. Its zero row count reads as "quiet" and means "absent" — the
two are indistinguishable from the table's side, which is exactly why the check reports the
registered-minus-observed direction and never fails on it.

## 2026-08-22 — the `090` came back UNVERIFIABLE, and its REASON is evidence for the design

Run correlation `c965bfec-993a-4b2b-88ba-d44549c81df1`, filed on the five-writers claim.
Verdict: **UNVERIFIABLE (stopped: scope-not-narrowing)**. **This is not a refutation and must not
be reported as a confirmation either.** What it said it still needed:

- the `pgxpool`-vs-`*sql.DB` question could not be settled — *"the symbol search for pgxpool
  found 0 rows but the index holds no `type` kind declarations at all (kinds present: alias,
  const, func, interface, method, struct, var), so that 0 is unrepresentable-not-absent"*;
- `214_build_dispatch_watchdog.sql` was **absent from the bundle entirely** — *"SQL files are
  outside the .go-only indexed corpus and cannot be code_request'd"*;
- `cmd/content-loss-check/main.go` returned 0 rows, *"explicitly flagged as possible index
  staleness on unpushed work, not proof the binary is absent"*.

So the loop's static tier is `.go`-only, index-backed, and holds no `type` declarations — and the
claim under test is precisely about writers spread across Go, SQL and `cmd/`, two of which it
cannot see. **A claim can be outside what an instrument can grade without being outside what can
be verified**: all five files were read first-hand and their type declarations quoted. Recorded
here rather than in `WRONG_CALLS.md` because nothing was claimed wrongly — but the *shape* is
worth knowing before filing a 090 about anything that is not Go.

## 2026-08-22 — council round 1: REVISE, and the objection was right

`be1fd678-0836-4f32-90a6-8927b2463fee`, gated by `editquality` at HIGH:

> *"Both this edit and edit 5 call `findingCodeRegistryCodesExcept(t, ...)` but no edit in the
> plan defines this helper… Without it these test files will not compile — the repointing this
> edit exists to do is a no-op until the helper exists."*

Correct, and it is the second time this lane has nearly shipped an edit whose stated effect
depended on something not in the plan. Auditing the plan the same way found **a second omission
of exactly the same kind**: `cmd/config-key-audit/main.go`'s mode dispatch, without which
`--finding-codes` is unreachable — a mode that exists and cannot be called. Round 2 resubmitted
on the same correlation with both, and with sketches taken from the working code rather than
composed. Cost of the round: one resubmission. Cost of not having it: a commit whose two
headline edits were inert.

## 2026-08-22 — what is built, and the controls it passes

`--finding-codes` (mode 16 of `cmd/config-key-audit`), registry of 53 entries, 19 tests.
Against the live table: **43 normalised codes observed (44 raw — the two
`tool_crosslink_not_emitted:*` variants collapse), 0 findings, 32 unruled, 10 registered but
unobserved.** Controls, all run this session (RUNBOOK has the commands):

| control | result |
|---|---|
| real live list | exit 0 |
| + `TEST_UNREGISTERED_X` | exactly one `undeclared` finding, exit 1 |
| remove it again | exit 0 — so the check can come out clean, i.e. it discriminates |
| a raw `tool_crosslink_not_emitted:<new reason>` | exit 0, collapses to the family key |
| `consumed` reader repointed at a real file that does not name its code | `reader-does-not-name-code`, exit 1 |
| the true reader | exit 0 — so the rejection was about the reader, not about any change |
| empty input | refuses: empty stdout, exit 2 from the compiled binary |

⚠ **The registry mutation was done on a COPY** (`scratchpad/reg_wrong_reader.json` via
`--registry`), never on the shipped file — `WRONG_CALLS.md` 2026-08-22 records a session mutating
a shared file in place to prove a guard and another session committing it during the window.
That is also why the roster test's predicate was split out pure: it can be mutation-proved
against a fixture instead.

⚠ **`go run` collapses the child's exit status** (LANDMINES): the empty-input control read as
`exit=1` under `go run` and `exit=2` from the compiled binary. The signal to branch on is
**empty stdout**, exactly as `audit-optional-key-budget.sh` records. First reading of that control
was logged as a failure before the landmine was recalled — no claim was published, so it is here
rather than in `WRONG_CALLS.md`.

## 2026-08-22 — council round 2: APPROVED, with three advisory objections, all three checked

`be1fd678-0836-4f32-90a6-8927b2463fee`, round 2, **approved** (13 reviewers, 4 abstained, no
gating objection). All three advisories came from `guardian` and none was waved through:

1. **[medium] a new `_test.go` in `platform/orchestration/actions` — a package essentially every
   pipeline imports — so a compile failure there blocks `go test ./...` for all of them, not just
   the two tests it fixes.** True. Exercised in isolation before and after repointing:
   `go test ./platform/orchestration/actions/` green, and green again from `git archive HEAD`.
2. **[medium] the roster read couples two previously self-contained unit tests in a core package
   to a file under `docs/`; a stage that copies only source now fails them for an unrelated
   reason.** A real new coupling. **Measured:** nothing runs `go test` in a stripped context today
   — no Dockerfile in the tree runs it, no CI workflow directory exists, and `.dockerignore`
   strips `*.md`, not this `.json`. So it is real but unexercised. **Not removed**, because the
   only way to remove it is to compile a copy of the roster into the package — the third
   hand-maintained roster, i.e. the exact drift being retired. Instead the failure message now
   names the coupling, says the fix is to carry the file, and says explicitly not to work around
   it by hard-coding the list back. Failing rather than skipping is deliberate: a skip lets a
   collision through in the one environment where nobody is watching.
3. **[low] confirm no other script parses `os.Args[1]` positionally in a way the new branch could
   shadow.** Checked all seven callers: every one passes either **no args** (`audit-config-keys.sh`,
   the default mode) or an explicit `--<mode>` flag. `--finding-codes` is a distinct literal and
   shadows nothing.

The seat's own note is worth keeping: *"No architecture-change signal here… Every production-path
edit is additive… Blast radius is real but bounded to CI/test-time coupling in a shared package
plus a new docs->test-time file dependency."* That matches the plan's §8 scope argument, so the
council-gate-not-RFC call was right.

**Round cost:** one resubmission. **Round value:** the round-1 gating objection caught two edits
that would not have compiled, and round 2's advisories caught a coupling I had created without
naming. Both rounds found something real, which is now 4 of 6 for this lane's experience of the
gate.
