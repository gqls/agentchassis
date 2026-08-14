# NOTES — dispatch/pool lane (`bugs_closed/239` → `bugs_open/259`)

Append-only, newest at the bottom. The technical log: evidence, commands, what the system
actually said, and every misstep.

> **CREATED LATE, 2026-08-14, and that is itself a finding.** This lane ran from 2026-08-10 to
> 2026-08-14 on a series of four `HANDOFF_*_continue_here.md` files and had **none** of the
> standing five that CLAUDE.md requires ("Create them at the START, not at handoff time — the
> point is that a doc exists to update while the work is happening"). The handoffs are good and
> they are not a substitute: a handoff is written for the next session and is therefore
> *current-state*, so every wrong turn this lane took before 08-14 is now only recoverable from
> git history and scrollback that no longer exists. What follows starts at 08-14. The earlier
> record is the handoff series: `HANDOFF_2026-08-11`, `-08-12`, `-08-12b`, `-08-13`.
> ⚠ `HANDOFF_2026-08-12` contains a claim that is false (corrected in 12b §2) — do not mine it.

---

## 2026-08-14 — `bugs_open/259` (slug `three_processor_paths…`): candidate 1 applied

Picked up from `HANDOFF_2026-08-13_continue_here.md`, which had the bug fully diagnosed and
named candidate 1 as the next task. Verified rather than inherited: re-located every line
number at HEAD before editing, because the handoff says several lanes edit `processor.go`.

**Baseline first.** `go build ./...` exit 0 and `go test ./platform/messaging/...` ok, recorded
*before* any edit so a pre-existing failure could not be read as mine. This paid off — see the
thunder note below.

**The handoff's line numbers all still held** at `ad945029d` (and again at `3b0ea20ff`, which
HEAD moved to mid-session as another lane committed; `platform/messaging/` stayed clean).

### Three things the handoff and the bug file did not carry

Found by re-locating rather than trusting, which is the only reason they were found at all:

1. **`stateRepo`** — a `*orchestration.StateRepository` field built in the constructor from the
   **nil** `sqlDB` and read **nowhere**. The grep that settles it:
   `grep -rn "stateRepo" --include=*.go platform/messaging/` → exactly two hits, the
   declaration (`:45`) and the assignment (`:136`). The identifier is unexported, so nothing
   outside the package could reach it either. Left behind it would have become
   `NewStateRepository(nil, logger)` — still compiling, still nonsense.
2. **A third breaking test.** The handoff named two. `processor_response_status_test.go:388`
   called the inner `sendWorkflowResponse`, which the fix deletes. The bug file's own
   correction said "exactly ONE **non-test** caller" — accurate, and the test dependency simply
   did not carry forward into the fix list. Worth noting as a class: *"no non-test callers"* is
   the right question for reachability and the wrong question for "what will stop compiling".
3. **`createSQLDB()` (`:974`)** — zero callers, a second opener of a second handle to the same
   database. Same class, same file, not in the bug file (it reads `CLIENTS_DB_*`, not
   `DATABASE_URL`). Owner asked for it to be included; it is named in the commit message so it
   does not read as an unexplained passenger.

### One piece of evidence added to site C

The bug file argued C's redundancy from the `processed_messages` write rate (449 rows / 82
writers in one hour). That establishes agentbase's *claim* is live. It does not establish that
the **release** arm — `bugs_closed/239`'s rule that a transiently-refused dispatch must give its
claim back rather than complete it — exists outside C. It does:

```
$ grep -rn "ReleaseMessageClaim" --include=*.go . | grep -v _test.go
platform/agentbase/agent.go:1199                 # live, on a.stateRepo (from a.db)
platform/orchestration/state.go:352              # the definition
platform/messaging/processor.go:1539             # inside dead site C
```

Two callers fleet-wide; one of them was C. So the deletion loses neither claim, release, nor
completion. This mattered because it was the one way C's deletion could have been a real
regression rather than a no-op, and the write-rate argument alone could not see it.

### The mutation check, and why it was not optional

Redirecting `TestSuccessResponseStatusStillComplete` from the deleted wrapper to
`sendWorkflowResponseWithStatus` leaves a **passing** test — which is exactly the shape that
hides a redirect that quietly stopped asserting anything. So the redirect was mutated, not
trusted: forcing the error-only status override to fire unconditionally (`if errInfo != nil` →
`if true`) produced

```
--- FAIL: TestSuccessResponseStatusStillComplete (0.00s)
    processor_response_status_test.go:404: IsComplete = false on a success response
    processor_response_status_test.go:407: IsError = true on a success response
```

and restoring made it pass. The file was copied to the scratchpad first and restored from that
copy, not from `git`, because this is a shared tree and a stash or checkout could have taken
another session's work with it.

### Missteps this session

- **I inserted a comment between struct fields, which broke gofmt**, and I did not notice —
  the **pre-commit pattern check** did, *after* the commit had already landed (`e37f79b65`):
  `gofmt platform/messaging/processor.go — not gofmt-clean`. Fixed forward in `f894b1a38` by
  moving the note to the type's doc comment, which keeps the field-alignment group contiguous.
  **The cheap check I skipped: `gofmt -l <files>` before committing, not after.** Note the
  build gate rejects un-gofmt'd code, so this would have reached CI as a failed gate.
- **I predicted a diff-grep would print 2 and it printed 0, and the prediction was the only
  reason I looked.** Checking the register edit, `git diff … | grep -c '^[+-][^+-]'` returned
  `0` while `git diff --numstat` said `1 1`. The register uses `- ` markdown bullets, so a
  changed bullet reads as `-- **open review question:…` in a diff — diff-marker followed by
  bullet-dash — and `[^+-]` cannot match the second character. **This is the documented
  landmine** ("`git diff | grep '^-[^-]'` cannot see a deleted markdown BULLET"), which the
  SessionStart hook had already shown me, and I wrote the blind grep anyway. What saved it was
  gating on the **count** first, which is what that landmine tells you to do. No new entry
  owed; recorded here because following half of a landmine's advice is how it still bites.

### Not mine, but found: `internal/adapters/thunder/api` is broken at HEAD

`go test ./platform/... ./internal/... ./pkg/...` surfaced one failure:

```
vet: internal/adapters/thunder/api/client_test.go:113:5: unknown field Identifier in struct literal of type Instance
```

Attributed before being dismissed: the package is **clean in the tree** (so the breakage is
committed, not someone's WIP), and it references nothing this change touched
(`grep -rn "messaging\|sqlDB\|sendWorkflowResponse" internal/adapters/thunder/api/` → no hits).
Another session's committed test breakage. Left alone — not this lane, and forward-only. Worth
someone's attention: `go build ./...` does **not** compile test files, so a baseline taken with
`build` alone would have missed it and a later session could inherit the blame.

### State at the end of the session

- `e37f79b65` — the fix. `f894b1a38` — gofmt follow-up. `c70f4565f` — bug file.
  `82b159ee9` — an unrelated declared tidy (SYS-091's index row still said `built`).
- Council: `Council-Submitted: 0ff072ef-ee02-465e-8a70-f5461c585ec9`. **Verdict owed a read.**
- `bugs_open/259` stays **OPEN**: fixed in the tree, inert until a fleet roll, so still
  reproducible in production. Post-roll verification is in the bug file.
