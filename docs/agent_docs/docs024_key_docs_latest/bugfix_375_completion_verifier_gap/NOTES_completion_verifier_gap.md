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
