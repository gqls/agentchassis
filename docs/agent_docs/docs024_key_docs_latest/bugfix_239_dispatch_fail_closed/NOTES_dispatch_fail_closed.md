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

### Two landmines and a wrong call filed — and the verification deliberately NOT chased

Appended two entries to `LANDMINES.md` (the `go build` baseline blindness, and the
struct-comment/gofmt one) plus a `WRONG_CALLS.md` row for the diff-grep prediction.

Two process notes, both `[UNVERIFIED]` by choice rather than omission:

- **My headings used `##`; the format is `###`.** The sync's `--check` said so
  (`heading uses '##', the format is '###'`) and the census backs it: 418 `###` entries to 79
  `##`, the latter being the file's structural sections plus a drift of entries that made the
  same mistake — including one filed by another lane the same day. Fixed. The slug is derived
  from the heading TEXT, so the level does not change it; what it changes is whether the parser
  treats the block as an entry at all.
- **I ran `landmines-sync.py --apply`, which consumed my own entries' "new" status**, exactly as
  the landmine at `LANDMINES.md#running-landmines-sync-py-apply-before-landmines-verify-dispatch`
  says it does: the wrapper `landmines-verify-dispatch.sh` should be run **instead of** `--apply`,
  never after. CLAUDE.md's own instruction ("after you append, run `--apply`") is what leads you
  into it. **I did not then fire the verifier by hand, and that is a decision, not a lapse:** the
  adjacent entry records that the verifier's index is **100% Go while 81% of footprints are not**,
  and my two entries' footprints are `go build`/`go vet`/`go test` and
  `gofmt`/`.githooks/pre-commit`/`scripts/pattern-check.py` — a shell hook and a Python script.
  A verdict on those would resolve 0 rows and be equally uninformative whether it came back
  STILL_VALID or STALE, and a STALE is the signal that argues for deleting a correct trap. So
  both entries stand on their prose and their cited commits, unverified by the harness, and
  marked as such here.

### My `WRONG_CALLS.md` row was swept into another session's commit, and this is the record of it

I appended the diff-grep row to `WRONG_CALLS.md` and, minutes later, `git status` reported the
file **clean** while `grep` found my text present. Not a lost write — the opposite: commit
`9de74ada1` (08:54:58, another session, "wrong_calls: log the CLASS — a census of what your
commit contains goes stale between measuring and committing") is a single-file, 88-insertion
commit and **my ~30 lines are inside it**. `git show 9de74ada1 -- …/WRONG_CALLS.md | grep -c` on
my own sentence returns 1.

This is the hazard CLAUDE.md names outright: committing per task stops *you* sweeping up others'
work, and cannot stop a session running `git add -A` from sweeping up *yours*. Nothing is lost
and forward-only holds, so there is **nothing to re-commit** — my content is already at HEAD,
filed under a message that describes only their entry. The row is therefore findable by content
but not by commit message, which is why it is named here.

**How it was caught, and the transferable bit:** by gating on `git diff --numstat` before
committing and noticing the file was **absent from the output entirely**. An absent path in a
numstat is not "nothing changed" — it is "git and I disagree about what I just did", and the two
explanations (my write failed / someone committed my write) look identical until you `grep` the
file and then `git log` it. Checking the file alone would have shown my text and looked fine;
checking git alone would have shown clean and looked like a failed write. It took both.

### Council: APPROVED first round, and what the objections were actually worth

`0ff072ef`, 07:57:26Z — **~15 minutes from submission to verdict**, well inside the ~30 minutes
CLAUDE.md tells you to budget. 10 reviewers, 8 approve, 2 object (guardian, prior_art_librarian),
none high-severity, `gated_by_truncation: false`. Full answers in `bugs_open/259`.

**The objections were worth more than the approval was.** Both medium ones caught the same real
weakness: the submission's `risks` block *named* agentbase's gate equivalence as the load-bearing
invariant and then **asserted** it rather than measuring it — I wrote "that is worth a reviewer's
eye" where I should have written a grep. Guardian put it under `missing` as *"asserted, not
checked"*, which is exactly right.

Measuring it took one file read and came out stronger than the assertion: `a.db` and
`a.stateRepo` are assigned on **adjacent lines inside one `if a.config.DatabaseURL != ""`**, and
`a.db` is the argument that becomes `p.db`. So `a.stateRepo != nil` ⟺ `p.db != nil` — the gate
cannot be weaker than "the processor has a live handle". `isStateless` is assigned `true` once
and never false.

**And the guardian's objection contained its own resolution, which I nearly missed:** it asked to
confirm gate equivalence *"or that this site was truly always dead regardless of gate state"*.
The second disjunct is trivially true and I had already proven it. Two claims had been running
together in my head as well as in the submission: *deleting C changes nothing* (needs only that C
never ran — a `DATABASE_URL` fact) versus *the chassis still dedupes* (a claim about agentbase,
true or false **before** my change as much as after). Separating them dissolves the objection
without any new evidence. **Lesson: when a reviewer offers a disjunctive condition, check the
cheap branch first** — I went to measure the expensive one and only then noticed the other was
already established.

**Two objections left OPEN rather than argued away**, because they are honest tooling limits:
site A's no-defer claim is a function-**body** fact and the code index stores declarations only;
site B's zero-caller grep ran against the working tree, not the pushed index. Both seats asked
for a human re-read at HEAD and both should get one. `debug_historian` also flagged something the
post-roll plan had missed — the fleet is **MIXED for hours** after a release and
`-l app=agent-chassis` selects 2 pods of the many running that binary — now carried into the
handoff's verification section.

### State at the end of the session

- `e37f79b65` — the fix. `f894b1a38` — gofmt follow-up. `c70f4565f` — bug file.
  `82b159ee9` — an unrelated declared tidy (SYS-091's index row still said `built`).
  `fea516212` — these standing docs. `f91397213` — the handoff. `d5d3ad592` — two landmines.
  `6ea444469` — the verdict and the objection answers.
- Council: **APPROVED**, `Council-Reviewed: 0ff072ef-ee02-465e-8a70-f5461c585ec9` (written only
  after reading the verdict; the code commits carry `Council-Submitted:` and `098` credits them
  automatically, with no amend).
- `bugs_open/259` stays **OPEN**: fixed in the tree, inert until a fleet roll, so still
  reproducible in production. Post-roll verification is in the bug file and the handoff.
- **One commit message is degraded and cannot be fixed:** `d5d3ad592`'s body contains
  `so a  keeps the reassuring` where I had written a backticked `| tail -N`. Backticks in
  `git commit -m` are **command substitution** — bash tried to execute `| tail -N`, failed with a
  syntax error, and substituted empty. The commit still landed. This is the documented
  `shell-tool-traps-committing` trap. Forward-only forbids an amend; the content is intact in
  `LANDMINES.md` itself (4 occurrences of `tail -N` at HEAD, verified), so nothing is lost but the
  message. **Check: single-quote the whole `-m` body, or avoid backticks in it entirely** — every
  later commit in this session used single quotes.
