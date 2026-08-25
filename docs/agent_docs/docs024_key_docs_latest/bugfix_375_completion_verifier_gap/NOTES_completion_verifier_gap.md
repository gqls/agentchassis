# NOTES — `bugs_open/375` (append-only, newest at the bottom)

---

## 2026-08-24 ~21:00Z — session 1, taking the lane

### Claimed first, before anything else

`bugs_open/375`'s status line said `OPEN, UNOWNED`; edited to name this directory and committed
alone (`e0e80b65f`). `scripts/who-owns.py 375` had been reporting OWNED on the strength of the
*closed* `bugfix_367_router_remit` lane's mentions — the handoff warned about exactly that false
positive, and it was still firing when I checked.

### The core claim re-verified `[MEASURED 2026-08-24 ~21:00Z, live code at HEAD]`

- `grep -rn 'GetVerifier(' platform/ internal/ --include=*.go` → **6 hits: 1 definition, 1
  non-test caller** (`complete_work_item_verification.go:122`), 4 in tests. None in
  `UpdateWorkItemStatusAction`.
- `UpdateWorkItemStatusAction` is at `v3_site_actions.go:5978` today (the bug file says `:6010`,
  filed one day earlier — **re-locate by symbol**). Read end to end; the next `func` after it is
  `containsString`, so the body is bounded and contains no `GetVerifier`.
- Registered verifiers: **13** as of 2026-08-24 (`RegisterVerifier(WithPolicy)?\("…"` → 13 distinct
  types; the raw grep returns 18 LINES because the two registration functions and two comments match
  too — count types, not lines).

### The census, re-run `[MEASURED 2026-08-24 ~21:00Z, live DB]`

Identical to §3a on the headline numbers: **200** live agent definitions; **6** name
`update_work_item_status` across **22** steps; **4** of them reach `complete`, across **6** arms
(`image-build-handler`, `image-source-unsatisfiable-handler`, `image-url-404-handler`,
`required-fields-missing-handler` ×3). Those four handle **5** item types
(`needs_imagery` 183 rows/92 complete, `required_fields_missing` 64/38,
`image_source_unsatisfiable` 15/0, `needs_hero_image` 5/3, `needs_logo` 3/1) — **134** completions
all-history through the unguarded path — and **none of the five has a registered verifier**.

**The zero is controlled.** The same 13-type list run WITHOUT the handler filter returns real rows,
and none of them is handled by the four agents. So the spellings are right and the separation is
real.

### Misstep 1 — I read the control's row count as a type count, exactly as the handoff did

The handoff's §3a says the control returns "**12 of 13** types with real rows". I ran it and got
**12 rows** — and started to write that down as 12 types before grouping properly. It is **12
(item_type, handler_agent) PAIRS from 10 DISTINCT types**: `literal_markdown` alone contributes
three (page-build-handler 52, page-rerender 10, section-editor 8). Three registered types
(`orphan_element_refs`, `page_canonical_collision`, `revenue_shape_cta`) have **no rows at all**.

The control's *conclusion* is unaffected — 10 types with rows, none of them handled by the four
agents, is still a decisive positive control. But the figure was wrong in the doc I inherited and
would have been wrong again in mine. **Caught by:** grouping by `(item_type, handler_agent)`
instead of trusting the row count, which I only did because I wanted the handler names for a
different reason.

### Misstep 2 — my first census was NARROWER than the thing it was censusing

I enumerated steps with `jsonb_each(default_config->'workflow'->'steps')`, which is the query the
handoff hands you. It reads as complete and is not: a step can sit inside a **nested loop-step
config**, where that path cannot see it.

I found this by accident, checking the *other* writer. `complete_work_item` came back with **2**
agents against the handoff's **4** — and the handoff was right. Re-run recursively
(`strict $.**{0 to last} ? (@.action == "…")`) it is 4: `build-dispatch-loop` and
`site-work-orchestrator` carry it nested.

The recursive scan on `update_work_item_status` returns the same 22 steps, so **this bug's own
number was never at risk** — but I would have published a wrong neighbouring number, and the only
reason I did not is that a figure I did not need disagreed with a document. **Cheap check that
would have caught it first time:** run the recursive form once and compare, rather than assuming
the flat path is where steps live.

⚠ A second thing that query must do: `status` **defaults to `complete`** when the key is absent, so
`WHERE config->>'status'='complete'` cannot see a step that omits it. All 22 name it explicitly
today. `COALESCE(..., '(default=complete)')` is what makes that a finding rather than a silence.

### The finding the handoff did not have: `CQ-023`'s landmine is FALSE, and false *because of this bug*

The register entry `CQ-023` (`docs026_concept_register/register/content-quality.md:236`) warns:

> *"a verifier later registered for `required_fields_missing` (RegisterVerifier) would fail-closed
> the `converted` arm's completion — none exists today; whoever adds one must re-read this router's
> close paths first."*

`close_converted` is an **`update_work_item_status`** step (census above), and that action never
consults a verifier. So registering the verifier today would **not** fail-close that arm; it would
do nothing at all, silently.

This sharpens what the bug actually is. The handoff framed it as *"a trap set for the next person,
by name"*. It is worse: the trap is **signposted with the wrong warning**. The next person is told
to expect a fail-close, plans around it, and gets a silent no-op — and the coverage test goes green
while they do. Both halves of what they are told to expect are wrong.

**Consequence for the fix, and it is the deciding one:** it rules OUT making the consult automatic.
Fixing 375 that way would make `CQ-023`'s sentence true — i.e. it would break a live route as a
side effect of a guard nobody asked to be armed. Opt-in per step is not just the ruled shape
(owner 2026-08-02 §2); here it is the only shape that does not break something.

### Design reading, so the next session does not re-do it

- `verifyBeforeComplete` (`complete_work_item_verification.go:65`) is **two gates**: 1b, the
  no-change gate, which reads *the handler's own reply payload*; and 2, the registered verifier.
  `UpdateWorkItemStatusAction` has **no handler payload** — it has step config — so it must run
  **gate 2 only**. Handing gate 1b the wrong payload would grade the wrong evidence, which is the
  precise error that file's own header records (`complete_work_item_no_change.go:33-41`).
  ⚠ Gate 1b's roster (`noChangeGates`) holds exactly one type today, `dark_section_audit`, and none
  of our five — so "just call `verifyBeforeComplete`" would be inert *today*. That is the reasoning
  this whole bug is about, so it is not good enough.
- On refusal the guarded writer calls `failUnverifiedCompletion` (`:413`), which increments
  `attempt_count`, releases the claim, and lands `triaged` (retry) or `failed` (budget spent).
  Reuse it; a second refusal path is the drift `bugs_closed/284` exists to stop.
- `update_work_item_status` has **no `RegisterActionInputSpec`** — it reads `params.StepConfig.Config`
  directly. It reads **7** step-config keys today (6 in the action, 1 in the failure ladder:
  `stop_on_repeat_failure_item_types`). Adding one makes **8**, under RFC_022's ruled budget of 10.
  ⚠ But note the honest part: because the action declares no spec, `--optional-key-budget` counts it
  as **ZERO** and would not see a tenth key either. That is the same blind spot CLAUDE.md records
  for `retract_asset_files` / `publish_site`, and it is not this lane's to fix — recorded here so it
  is not discovered a third time.

### Misstep 3 — my test fixture broke a real guard, and my first instinct was the wrong fix

My first version of `update_work_item_status_verification_test.go` registered a synthetic
verifier into the process-wide registry from `init()`:

```go
func init() { checks.RegisterVerifier("test_375_verified_type", func(...) {...}) }
```

The package's own tests passed. `go test ./platform/orchestration/actions/...` then failed —
in a test I had never heard of:

```
--- FAIL: TestClaimTimeoutExclusionCoversBothCompletionGates
    item_type "test_375_verified_type" has a registered verifier (gate 2) but is NOT
    declared in livespec.ClaimedItemTimeoutExclusions.
```

**The guard was right and my fixture was wrong**, and the tempting fix — add the test type to
`livespec.ClaimedItemTimeoutExclusions` so the guard goes quiet — is *fixing the checker to
agree with the fixture*, with a production declaration as collateral. I did not take it, but I
considered it for long enough to be worth writing down.

The correct fix was to stop touching the shared registry at all: `verifierLookup` is now a
package variable defaulting to `checks.GetVerifier`, the test swaps it under `t.Cleanup`, and
`TestVerifierLookupIsNotASwitchInProduction` asserts there is exactly one assignment to it in
the package's non-test source (mutation-proven: adding a second makes it fail).

**And the failure was worth more than the inconvenience.** It is how I learned there is a
THIRD writer of `complete` — the `claimed-item-timeout` sweep, which writes the row directly
so neither gate runs, and is held off a type only by a declaration plus a lockstep test
(`bugs_closed/317`). That is **the same class as 375, already solved once, by the shape
candidate 4 should copy.** It is now written up as §7c of the bug file. Nothing in `375`, in
the handoff, or in my own reading had pointed at it.

**Cheap check that would have caught the fixture problem first time:** before registering
anything into a package-level registry from a test, grep for who READS that registry —
`grep -rn 'RegisteredVerifierItemTypes()' platform/` returns two callers, and one of them is a
cross-package contract test. One command.

### Where the change ended up `[2026-08-24 ~22:00Z]`

- `c735bfd9c` — the gate (`update_work_item_status_verification.go`), the wiring in
  `UpdateWorkItemStatusAction`, the shared row read factored into `loadWorkItemVerifyRow`,
  and the tests. Four mutations, each failing the right tests: **M1** wiring removed → 4 fail;
  **M2** `mayComplete` forced true → only the defect-persists test; **M3** unarmed payload
  nulled → only the bypass-record test; **M4** seam re-pointed in production source → only the
  seam guard. Run against **committed HEAD** via `scripts/verify-head-builds.sh`, because
  another session's uncommitted WIP in `discovery_checks/check_page_list_stale.go` did not
  compile for about ten minutes in the middle of this — a working-tree `go test` failure that
  was not mine and would have read as mine.
- `c94212ad3` — `verifier_coverage_test.go`'s header, `CQ-023`'s corrected landmine,
  `WII-030`, the index row, the `102_coverage_ratchet.txt` line dropped (it explicitly asked
  to be, once candidate 1 shipped with a register entry), and the new LANDMINES entry.
- Council: `7a6add95-30e9-4576-85e5-df5bad0f7119`, dispatched 20:26:42Z and executing within
  minutes rather than the ~29 the runbook budgets for.

### Council verdict: APPROVED round 1 — and the two medium objections were both worth acting on

`7a6add95-30e9-4576-85e5-df5bad0f7119`, dispatched 20:26:42Z, `complete_approved` 20:40:52Z.
**12 reviewers, 5 abstained, "approved with 2 advisory objection(s) — none high-severity".**
Seats: editquality, bug_historian, reuse_agent, guidelines, guardian (object), diagnosis_guardian,
improvement_guardian, debug_historian, constitution, mission, prior_art_librarian (object),
architecture.

**An APPROVED verdict is not a verdict with nothing in it.** The two objecting seats each named
something real, and one of them found a false measurement I had already published to six places.

| seat | severity | what it said | what I did |
|---|---|---|---|
| `prior_art_librarian` | **medium** | the blast-radius figures may be drawn from `site_work_items`, a rolling window, so "all-history" is likely an undercount | **CONFIRMED and corrected** — §8 of the bug file, and see misstep 4 below |
| `guardian` | **medium** | factoring `loadWorkItemVerifyRow` touches the widest-used existing path; mutation-prove the existing callers | **DONE, and it found a real gap** — see misstep 5 |
| `guardian` | low | confirm `failUnverifiedCompletion` assumes nothing about `CompleteWorkItemAction`'s call context | **CHECKED**: it takes db/itemID/agentType/resultJSON/errorMsg/reason/logger explicitly and issues one UPDATE. The only context it reads is the row's own `retry_after` (`workItemRetryNotPendingSQL`), which is correct from either caller — it stops an attempt being double-charged when the ladder has already ruled |
| `architecture` + `guidelines` | low | confirm `verify_before_complete` does not push the action past RFC_022's N=10 optional-key budget | **MEASURED**: `update_work_item_status` reads **7** step-config keys today (6 in the action + `stop_on_repeat_failure_item_types` in the ladder) → **8**. Under N=10. ⚠ And the honest part: the action declares **no `RegisterActionInputSpec`**, so `--optional-key-budget` counts it as **ZERO** and would not see a tenth key either — the same blind spot CLAUDE.md records for `retract_asset_files`/`publish_site`. Not this lane's to fix, recorded so it is not found a third time |
| `prior_art_librarian` | low | the "gate 1b cannot apply" argument was asserted from behaviour, not an existence check | **ANSWERED**: `noChangeGates` holds exactly one type, `dark_section_audit`, and none of the seven. Read, not inferred |
| `editquality` + `bug_historian` | low/missing | edit 5 is comment-only; and `CQ-023` itself is not corrected, so a reader consulting it directly still inherits the wrong warning | **ALREADY DONE** before the verdict landed (`c94212ad3`). The submission's edit list is council-scope files only, and register prose is refused by the gate client-side — so the correction was invisible to the reviewers. Worth knowing: **a seat can only object to what the submission can carry** |
| `bug_historian` | low | nothing arms it; the mechanism may rot unexercised | the named trade-off. The bypass record is the answer to it, and §7c is the enforcing half |

### Misstep 4 — I published a rolling-window figure to six files, with a control that shared its blind spot

Full account in `WRONG_CALLS.md`. In short: `site_work_items` is a rolling window, the archive
holds 25,281 rows, and over the union the blast radius is **7 item types / 578 completions**, not
5 / 134. Two types (`unfulfilled_hero_variant`, `image_url_404`) had completed **entirely** into
the archive. **The landmine for this is in my own auto-loaded memory index** and I did not apply it.

Three things I want the next session in this lane to take:
- **The positive control did not help, and I had been relying on it.** It queried the same table,
  so it tested my spelling and not my window. A control drawn from the same source as the
  measurement cannot see a source-shaped error.
- **The conclusion survived — all seven types are still unverified — and that is luck, not
  method.** Had either archived-only type carried a verifier, the census would have printed the
  same reassuring zero and an `RFC_022` scope claim to a review board would have rested on it.
- **A `[MEASURED]` marker travels faster than its correction.** The figure reached two Go headers,
  where a stale marker outlives every doc. Both are corrected at source.

### Misstep 5 — the guardian's mutation request found a gap, and my first fix for it was vacuous

Mutating the extracted `loadWorkItemVerifyRow` to return an empty `ItemType` failed six tests,
**three of them pre-existing `TestVerifyBeforeComplete_*` ones** — so the extraction is guarded and
the guardian's objection is answered.

But a second mutation — replacing the `spec` column with a literal `'{}'` — **failed nothing.** No
test in the package had ever asserted that a verifier receives the item's real spec.

**My first attempt to close that was itself vacuous**, and instructively so. I wrote a test
asserting the values that arrive at `VerifyTarget` — and it *passed the mutation*, because sqlmock
returns whatever rows the test queued regardless of what the statement says. **A mock cannot assert
anything about SQL TEXT**; that is the "a mock's own bookkeeping cannot assert a NEGATIVE" trap one
level along, and I walked into it while explicitly trying to avoid it. The fix is to put the column
list in the **expectation** (`verifyRowReadSQL`), which sqlmock matches against the real statement.
Dropping either `spec` or `page_id` now fails six tests.

## 2026-08-25 — post-roll verification

Chassis rolled to **`v1.0.1337`** overnight (both pods, started 09:27Z). Verified at the artefact,
never at the tag:

- `build provenance` had already scrolled out of `--tail=300` — the expected shape on this
  service. An empty result there means "not in range", **not** "unstamped".
- Binary probe on **both** pods, four literals plus two controls:
  `verify_before_complete` PRESENT · `verifier_not_consulted` PRESENT ·
  `owned_page_refusal_status` PRESENT (must-be-present control) ·
  `verify_before_complete_THIS_MUST_BE_ABSENT` absent (must-be-absent control) ·
  **`updateStatusVerifyConfigKey` absent** — the Go const identifier, exactly as `WII-030`'s
  verify-later predicted. Probing for the identifier would have read as "not shipped" while the
  feature works.

Runtime state, re-measured `[2026-08-25]`: **0** steps arm the key · census unchanged at 4 agents /
6 `complete` arms / 22 steps · **0** rows with `result._verification.status='verifier_not_consulted'`.

⚠ **That last zero was recorded WITH its demand control, because on its own it is worthless.** The
intersection of the 13 registered verifier types with the 7 reachable types is still `∅`, checked
mechanically at HEAD — so the record **cannot** fire. It is neither a pass nor a fail. Written into
`WII-030`'s verify-later and bug §9b in those words, because the next reader will meet the zero
before they meet the reasoning.

### The lane's own status lines were the thing most likely to mislead next

Three documents said **"INERT until the next chassis roll"**. That sentence was true when written
and became actively misleading this morning — it is the *"a stale status line prevents the thing it
describes"* shape, where a correct next action reads as premature. `WII-030`, the concept index row
and the workstreams memory line are all updated to **LIVE at the artefact**, each keeping the
separate and still-true statement that the gate is **inert on every live path by design**. Those
two are not the same claim and collapsing them is how "live" would come to mean "working".

### Landmine verifier: `NEEDS_HUMAN_REVIEW`, and it is a SCOPE LIMIT

It confirmed the Go footprint and could not reach three things: `verifier_coverage_test.go`
("0 rows"), and `CQ-023`/`WII-030`, which are register prose a `.go`-only index cannot hold. **The
0 rows are the index's staleness, not the file's absence** — its answers describe indexed commit
`e347c5ad` of **2026-08-23 12:21Z**, which predates the file's header edit and the entry itself, and
it has not been recut since. Answered in the entry, per the precedent set by the previous landmine's
author. Do not read it as an open objection.

### What I did NOT do, and why

**I did not close the bug.** The gate is live; the defect is not gone. Every completion through
`update_work_item_status` is still unverified, and a verifier registered for any of the 7 types is
still consulted by nothing. What shipped is the mechanism plus a tripwire, and CLAUDE.md's bar is
**fixed AND live**. Closing on "the gate shipped" would be `bugs_open/021` §INSTANCE 2's own error
one level along — mistaking *a mechanism exists* for *the defect is gone*. Bug §9c/§9d state the
reasoning and name what would actually let a future session close it.

## 2026-08-25 (evening) — the owner's four rulings, and what each cost

### Candidate 4: built, and the guard fired on its motivating case in real use

Built in the order that made the guard testable: the declaration + lockstep FIRST (passing, with
nothing registered), then the verifier — so registering it exercised the guard for real rather than
via a synthetic mutation. It fired three times, once per arm of `CQ-023`'s router, each naming the
step and the two ways out. That is the "re-run your detector on the motivating case" rule paying off:
a guard proven only by a hand-made mutation is proven against a case you invented.

### Misstep 6 — I wrote the declaration into `livespec.go` and had to take it back out

`ClaimedItemTimeoutExclusions` lives in `livespec.go`, so appending beside it was the obvious move
and I did it. Then `git diff --numstat` on that file read **186 insertions / 16 deletions** — and my
append had deleted nothing. **Another session had four hunks of an in-flight rename in the same
file, and it did not compile** (`livespec.DeferredDeclarations` undefined). A pathspec commit would
have shipped their half-written rename inside my commit.

Moved my block into `platform/livespec/unarmed_completers.go` — same package, no shared file — and
verified `livespec.go` went back to showing only their changes (92/16). **Caught by:** running
`--numstat` before committing and noticing a deletion count I could not account for. **The cheap
check, and it is the one to keep:** *my* diff of *my* file should have a deletion count I can name.
An unexplained deletion is somebody else's work.

### Misstep 7 — my first attempt to close a coverage gap was vacuous, again, for a new reason

Registering the verifier surfaced that no test asserted a verifier receives the item's real spec.
I wrote one asserting the values arriving at `VerifyTarget` — and it **passed the mutation** that
dropped the column, because sqlmock returns whatever rows the test queued regardless of the
statement. Same trap as yesterday's, one level along: **a mock cannot assert SQL text.** The column
list has to live in the *expectation*. This is now the second time in two days, so it went into the
handoff's trap list rather than only into NOTES.

### The council found one thing I had genuinely overstated, and one it had wrong itself

`prior_art_librarian`, MEDIUM: my rationale said `bugs_closed/317`'s protection is "declaration +
build-time lockstep + a live-drift auditor", and quoted my own `grounded_in` back at me — a sentence
in `claim_timeout_exclusion_lockstep_test.go` saying the phase-2 auditor had not shipped.

I went and checked rather than defending. **Phase 2 HAS shipped**: `livespec.Declarations` carries
`scheduled_task.claimed-item-timeout.exclusions` with `ProbeSQL: SELECT pre_query FROM
scheduled_tasks WHERE name = 'claimed-item-timeout'`, and `compareAllDeclarations` iterates **every**
declaration with no `Phase` filter (`livedeclarations.go:129`). So my claim was right and the comment
was stale.

⚠ **I nearly got this backwards.** The Declaration is marked `Phase: PhaseGoSide`, which reads like
"Go only, not live-audited" — and I briefly took it that way, which would have made the seat right.
Reading the constants settled it: `PhaseGoSide` is the **checked** state and `PhaseLiveAudit` is the
**inert** one ("nothing can check this until the phase-2 live auditor exists"). The names are the
opposite way round from their meanings. That is now in the corrected comment and in `WII-031`.

**What the staleness cost, which is why the correction is not cosmetic:** a competent seat spent a
MEDIUM objection on a correct claim, and a reader of that file would have inherited the same belief.
Corrected in place with the cost recorded (`08a44365f`).

Two other objections were **factually wrong**, and answered with evidence rather than argument: two
seats doubted `checks.RegisteredVerifierItemTypes()` exists (it is `verifiers.go:198`, and the test
compiles and runs), and `reuse_agent` asked whether an existing audit mode already covers this —
grep says no mode mentions `update_work_item_status` or `verify_before_complete` at all. One fair low
objection stands unactioned and is recorded as such: a generic exclusion struct covering writers 2
and 3 was not considered.

### The verifier: written, and stopped one line short on purpose

Writing the one-line `init()` failed **five** build guards. Four were bookkeeping. The fifth was not:
the `claimed-item-timeout` sweep writes `site_work_items` directly, so until the **live** `pre_query`
excludes this type it would complete items straight past the verifier — `bugs_closed/317`
reintroduced *by adding a guard*. That step edits `livespec.go`, which was still another session's,
so I stopped at the file boundary and wrote the sequence into the verifier's own header instead.

**I am not comfortable with an unregistered verifier** — it is the "helper with no callers looks like
a finished refactor" shape — so it is exercised directly by tests (Grades both ways, all five
positive-absence arms, the still-failing case, the negative control, the fail-closed case) and three
mutations confirm the guards are load-bearing. The honest position is in `WII-032`'s status line and
in the file header, not implied.

### The `image_url_404` bug: not filed, premise refuted

Came to file it and found the empty `handler_agent` is deliberate and documented three times in the
detector, and that the handler had handled 3 rows rather than 0 (archived). The useful half: those 3
were hand-assigned and the handler **escalated all three back to `needs_human_review`**, which
refutes the "give it a dispatch route" remedy on this type by direct evidence. Contributed into
`bugs_open/033`, whose header already poses exactly that contract question and records that two
council seats disagreed about it.

**Two sessions made the same rolling-window mistake about the same table on consecutive days** —
mine yesterday, the morning handoff's "0 rows ever" today. That is not carelessness twice; it is a
table whose name promises history and holds a window.

### Misstep 8 — the council REVISE found a real defect my own mutation suite had passed over

Full account in `WRONG_CALLS.md`. Short version: `SchemaContentFields` returns `ok=true` with a
**zero-length** fields map for a v2 schema whose `fields` object is empty
(`component_schema_fields.go:78` — the key is present, so the assertion succeeds and it returns
early). `missingRequiredValueFields` then finds nothing missing, and my verifier said **Resolved**.
An emptied schema — the silent-loss class — would have been certified as repaired by the guard added
to catch it.

**Why my own testing missed it.** I had written the fail-closed branch for unparseable JSON and for
`ok == false`, and cited RFC_017 while doing it. `ok == true, len == 0` sat in the gap between the
two conditions I had already thought about. **The lesson is to enumerate what the helper can RETURN,
not what you expect it to return** — and to mutate per shape: three of the four shapes in the new
test were already caught by `!ok`, and only `{"fields":{}}` was the hole, which an aggregate mutation
would have hidden.

**How the seat got there.** Through the LANDMINES entry whose footprint names
`findResolvedRequiredFields` — a symbol that returns **zero** grep hits. The pointer had rotted while
the mechanism stayed live. So the entry worked *despite* being stale, because the reviewer read the
mechanism rather than chasing the symbol. I have corrected that footprint and added the half the
entry never carried: **it fires hardest on a VERIFIER**, because a detector's `continue` on an
unreadable declaration is correct while a verifier's identical arithmetic certifies a repair.

⚠ **Ledger note:** that correction shows as 3 insertions / 1 deletion on `LANDMINES.md`, and the
pattern check flagged the deletion — correctly, since the file is fleet-wide append-only. It was an
in-place correction of the one footprint line, with the old text preserved under strike-through
inside the replacement (verified: the struck symbol still appears in the new text). Nothing of any
other session's was touched. Recording it here because I did not say so in that commit message, and
the check's own guidance asks for exactly that.

### Round 2 resubmitted on the same trail

`RESUBMIT_CORR=c8ed18c1…`, so the whole trail accumulates under one correlation. Round 2 carries the
fix, plus answers to the other five objections — two of which were factually wrong about the code
(`checks.RegisteredVerifierItemTypes()` exists at `verifiers.go:198`; no existing `config-key-audit`
mode mentions `update_work_item_status` or `verify_before_complete` at all), one asked for the census
query in-submission (given), one asked whether I had checked the two shared guard-test files for
in-flight work before editing (I had — `git status --porcelain` empty on both, and both edits are
pure appends), and one asked whether guard 3 could desync the coverage-gap list (it cannot;
`itemTypesWithoutVerifiers` still lists the type, which is correct while unregistered, and the
coverage test fails on exactly that pairing the moment it registers).

**The gating objection is worth stating plainly rather than only rebutting:** `editquality` said, at
HIGH, that an unregistered verifier leaves the defect unchanged. That is *true as a description*.
The answer is that registering it without the migration is strictly worse than inert — the
claimed-item-timeout sweep would complete items straight past it, which is a false green rather than
a missing one. And the REVISE round is itself the argument for this order: a predicate that would
have certified an emptied schema as repaired was caught **before it ever graded a live claim.**

## 2026-08-25 (19:13Z) — the `v1.0.1339` roll, and the best piece of evidence this lane has

Chassis rolled to **`v1.0.1339`** (both pods, 19:07Z). Verified at the artefact, and this time the
pod was minutes old so the `build provenance` line was still in range rather than scrolled:
`git_commit a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`, dated 08-25 17:38.

**`WII-030`'s gate survived the roll:** `verify_before_complete` and `verifier_not_consulted` both
PRESENT on the new image, with a must-be-present control (`owned_page_refusal_status`) and a
must-be-absent control both behaving. Re-probed rather than assumed, per "per SERVICE, not per fleet".

### The finding: the LINKER proves the unregistered verifier cannot run

`git merge-base --is-ancestor` says all four of this lane's commits — `c735bfd9c`, `75c6919ac`,
`64645d05e`, `43277271a` — are **IN** build `a7459a44b`. And yet the verifier's own error literal
`must not be read as a repair` probes **ABSENT** from `/proc/1/exe`.

I did not guess at why. **Control, same package, same build, same breath:** the REGISTERED detector's
literal `schema declares these fields required with source llm` → **PRESENT**. Same file family, one
reachable and one not. So the difference is reachability: **nothing references
`VerifyRequiredFieldsMissingResolved`, and the Go linker removed it.**

**Why this is the best evidence this lane has produced.** The deferral argument — that leaving the
verifier unregistered is safer than arming it without the migration — was until now an *argument*,
and the council's `editquality` seat objected to it at HIGH precisely on those grounds. It is now a
*measurement*: the code is not in the running binary at all. The linker is a more credible witness
than my reading of the source, and it says the deferral carries no runtime risk whatsoever rather
than a small one.

⚠ **And the corollary is a trap I nearly walked into while establishing it.** My first reading of the
absent literal was "the build is stale / the deploy failed" — the same conclusion `bugs_open/153` is
about. Probing for a symbol you have deliberately left unwired returns "absent", which is
indistinguishable from a failed roll and means the *opposite*. It also refines a rule I had written
into `WII-030` myself: *"probe the string literal, never the Go identifier"* is **necessary and not
sufficient** — the literal must also sit in code something calls. Recorded fleet-wide as a **third
cause** on the existing `strings`/marker landmine (the other two being non-ASCII splitting and a
same-tag rebuild), with the discriminating control.

### Sweep: I did the thing I had spent the day warning three lanes about

Full account in `WRONG_CALLS.md`. `4210764e9` swept **four** LANDMINES entries from three other lanes
(`idea_uk_vm_site`, `loanzy_uk_example_site`, `dispatch_throughput` ×2) under a commit message about
my own documentation habits. Nothing lost; attribution destroyed.

**The mechanism, which is the only part worth keeping:** I ran `git status --porcelain` on that file,
printed `M …LANDMINES.md` in my own terminal output, and committed by pathspec anyway. My three
earlier LANDMINES commits the same day were clean because I ran `git diff --numstat` and *read the
deletion count*. **A check you do not read is worse than one you skip, because it leaves a record that
looks like diligence.** What actually worked all day was stating the expected numbers first, so there
was something to fail — every commit after this one in these notes does that explicitly.

Second lesson, which I had written the opposite of an hour earlier: **correcting a shared append-only
ledger IN PLACE is strictly more dangerous than appending to it.** An append can go at the tail; an
in-place correction needs the file to be in the state you last read, on a file every thread writes.

All three affected lanes were notified with the commit id their entries landed in. The follow-up
`WRONG_CALLS` commit (`483b37f6d`) carried two further complete passengers and was declared as a
`sweep:` naming their lanes, per CLAUDE.md — which is the sanctioned form and strictly better than
leaving three finished entries for a less careful commit to take silently.

### Misstep 9 — my registration sequence would have broken the build, and a peer lane caught it

The `bugs_open/395` lane messaged: it needs the *same* migration (`content_rewrite` on the
`claimed-item-timeout` clause) and warned that adding a type to `ClaimedItemTimeoutExclusions` before
anything can grade it trips the lockstep's REVERSE arm.

**They were right, and it was wrong in three of my documents at once** — the verifier's own header
(which I had been calling "the runbook"), the handoff's step table, and a LANDMINE. All three said
the exclusions entry **must come first**. Verified rather than accepted: `required_fields_missing` is
in **none** of the three rosters (`RegisterVerifier` — no, and my first grep said 1 because it
matched **my own commented-out line**, the source-scan trap; `noChangeGates` — no;
`acceptancePredicateGates` — no). So the entry alone → `excluded` true, `gated` absent → the reverse
arm fires. Corrected at `9e53bb02c`.

**The correct order, and each half is placed for a different reason:** apply the **migration** first,
because it concerns the LIVE object and is what stops the sweep completing past the verifier — its
window (live clause holds a type the Go slice does not) merely makes `--live-declaration-drift` noisy
while the sweep skips the type. Then land the exclusions entry **and** the `RegisterVerifier` call in
**one commit**, because either alone breaks the build in opposite directions. The reverse window is
the one with the actual hole in it, so the noise goes on the safe side.

⚠ **Why mine went stale, and it is not carelessness:** `395`'s gate 1c added the **third** roster
(`acceptancePredicateGates`) at `69479bcf6`, **13:41 the same day** — hours after I wrote the
sequence. My text was correct against a two-roster contract and wrong against a three-roster one.
**Nothing I could have re-checked on my own side would have caught it**, because the thing that
changed was somebody else's file changing the meaning of mine. That is a different failure from the
stale-status class: not a claim that decayed, but a contract that moved underneath a correct claim.
The only defence is the one that worked here — a peer reading it against their own use.

### The composition constraint I had not thought about

`livespec.Declarations` pins the live `pre_query` to `ClaimedItemTimeoutExclusionClause()`'s rendering
of the Go slice with a **FragmentMatch Min:1/Max:1**. So two migrations written against today's
14-type clause **do not compose** — whichever applies second must render the **MERGED** slice, or the
drift auditor fires on a correct-looking change. I would have shipped a migration that made `395`'s
correct amendment fail. Now recorded in all three documents.

**Ownership settled:** this lane writes the migration (my blocker is four build guards I can clear;
theirs is a live negative control they cannot manufacture honestly, and inventing one would be the
false-green this whole class is about). `395` anchors its `content_rewrite` amendment on our tail
afterwards. House pattern: **482** — read-before-write anchor, targeted `replace()`, `DO`/`RAISE`
verify, because a bare `SELECT` cannot stop a `COMMIT`.

**Not written today, deliberately:** step (b) needs `livespec.go`, still dirty with `363`'s rename,
and step (a) opens a drift window that should not sit open overnight. Better to write and apply both
close together than leave the estate half-changed.
