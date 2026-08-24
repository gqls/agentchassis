# NOTES — bugs_open/326, the front door cannot be re-used

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-23 ~17:00Z — session start, ownership and validity

`scripts/who-owns.py 326` says OWNED-or-recently-active, but the only commit it
finds is `ace492170`, the filing itself. The filing lane is
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/` (handoff
`HANDOFF_2026-08-19_fixing_the_one_shot_route.md`, commit `1eb446f89`). That lane
owns the ACCOUNT of the one-shot route; nobody has taken the FIX. The diagnosis
queue (`site_work_items` `needs_diagnosis` / `awaiting_diagnosis`) was **empty**
`[MEASURED 2026-08-23]` before I filed into it.

090 filed anyway, because I was about to assert a structural root cause that
CONTRADICTS the filed one: intake `df0b3d97-bb22-4b42-a85c-030ff42536cf`,
run `655c8508-a265-4926-8625-b54beff26869`.

## 2026-08-23 — the filed root cause is WRONG, and the live index says so

`bugs_open/326` states: *"`create_work_item`'s step matches the previous attempt's
item by `item_key` — regardless of that item's status, including `complete` and
`cancelled`"*. That is not what the estate does.

`\d site_work_items` on the live DB `[MEASURED 2026-08-23]`:

```
"idx_swi_dedup" UNIQUE, btree (site_id, item_key)
  WHERE item_key IS NOT NULL
    AND (status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
                              'failed','unresolved','cancelled']))
```

`complete` and `cancelled` are BOTH excluded from the index. A `complete`
predecessor cannot hold the dedup slot, and the `ON CONFLICT (site_id, item_key)
… DO NOTHING` in `writeWorkItem` (`platform/orchestration/actions/load_work_item_actions.go:1650`)
names exactly that predicate. So the index arm is innocent.

## 2026-08-23 — the real mechanism: the ANTI-CHURN block, two arms

`writeWorkItem` runs a block BEFORE the insert, for any item with a non-empty
`itemKey` and `recurrenceExpected == false`
(`load_work_item_actions.go:1516`). It counts siblings on
`(site_id, item_key)` at status `complete`/`failed` within 7 days:

- **Arm A — within-cycle suppression.** `newestAge < 3.0` hours ⇒
  `return workItemWrite{}, nil`. **No row, no error.** The caller gets
  `Inserted:false`, which `create_work_item` reports to config as
  `deduped: true` — byte-identical to a genuine open-item dedup.
- **Arm B — two-strike.** `terminalCount >= 2` ⇒ the arriving item's status is
  rewritten to `unresolved` and its summary prefixed `[unresolved after N attempts]`.
  `unresolved` is in `workItemTerminalStatuses` AND is not in
  `workItemDispatchableStatuses` (`triaged`, `approved` only), so the row is born
  **terminal, undispatchable, and outside the dedup index** — every later attempt
  repeats the same fate.

## 2026-08-23 — MISSTEP: my first live check could not have come out otherwise

I ran a query classifying every `deduped:true` result in retained
`orchestration_states` by whether an OPEN row holds that `(site_id, item_key)`
**now**, and got `open_holder = 0` for both distinct keys — and read that as
"the index cannot have been the mechanism". **It establishes nothing.** The
holder was open AT THE TIME and had since completed. Redone with timestamps
(`ev.updated_at` vs the row's `created_at`/`completed_at`), all **36** dedup
events resolve as legitimate index dedups: the newest dedup on each key is
~20 seconds after the row was created and ~5 minutes before it completed.

That is the right answer and a useful NEGATIVE CONTROL — Arm C works as designed
and is distinguishable — but I nearly banked the wrong one. Logged in
`WRONG_CALLS.md`.

## 2026-08-23 — the timing that settles it: loanzy.uk, three submissions

`domain-submitter` writes a `submission` spec BEFORE the `create_work_item` step,
so `site_specs` dates every submission even after `orchestration_states` is
reaped (retention is ~24h; the bug's correlation
`3296ac3a-30bd-4db5-84ae-16260394f3bc` is long gone).

```sql
SELECT created_at, is_current FROM site_specs
WHERE site_id='55213ded-03ec-40f7-8fc1-169de05e05c8' AND aspect='submission'
ORDER BY created_at;
```

| # | submission spec | outcome |
|---|---|---|
| 1 | **12:53:00.04Z** | filed `research_loanzy.uk` at 12:53:00.25, ran, `complete` at 13:36:36 |
| 2 | **15:21:17.23Z** | **the deduped one** — `inserted:false`, no row |
| 3 | **20:16:12.61Z** | filed a new `research_loanzy.uk` at 20:16:13.85 |

Submission 2 landed **2 hours 28 minutes** after the terminal sibling was created.
Arm A's window is **3.0 hours**. `[MEASURED 2026-08-23]` — and the measurement
could have come out otherwise: at 3h01m the insert would have succeeded and there
would have been no bug to file.

**Consequence the bug file does not have.** The 78-row hand-rename of `item_key`s
was **not** what made submission 3 possible `[INFERRED 2026-08-23 from the code
path + these timings — to be proven with a test, not asserted]`: by 20:16 the
newest terminal sibling was 7.4 hours old, `terminalCount` was 1, and `complete`
is outside `idx_swi_dedup`. Waiting until 15:53Z would have done the same job.
The documented recovery procedure is hand-surgery for a three-hour timer.

## 2026-08-23 — what IS permanent, and it is Arm B

Arm A expires. Arm B does not, and nothing in the bug file mentions it.
Two terminal attempts on one key inside 7 days ⇒ **every** further attempt is born
`unresolved`: terminal, undispatchable, invisible to the dedup index, and — for
build-pipeline stage types — drained by nothing. `reviewRevalidators`
(`revalidate_review_queue_action.go:277`) covers six types
(`unresolved_cta`, `required_fields_missing`, `needs_section_data`, `needs_page`,
`voice_tells`, unverified claims). **None** of `needs_domain_research`,
`needs_strategy`, `needs_briefing`, `needs_site_plan` is among them.

Fleet-wide `[MEASURED 2026-08-23]`:

```sql
SELECT count(*) FILTER (WHERE summary LIKE '[unresolved after%') AS born_two_strike,
       count(*) AS all_unresolved FROM site_work_items WHERE status='unresolved';
--  635 | 747
```

**635 of 747** `unresolved` rows carry the two-strike stamp. The biggest
populations are `page_rerender` (212) and `improve_tool` (205) — both ACTION
REQUESTS, which is the class `recurrenceExpected` exists to exempt.

## 2026-08-23 — the class, measured at the config

> **CORRECTED 2026-08-23, same session — the count below was 17 and it was an
> UNDERCOUNT, from a walk that only descended one level.** My first census did
> `jsonb_each(default_config->'workflow'->'steps')`, which never enters
> `sub_workflow`. A properly recursive walk gives **22** `create_work_item` steps,
> **21** with an `item_key_prefix`, **19** of those undeclared and **2** declaring
> `recurrence_expected: true`. **Eight of the nineteen are nested**, so the flat
> walk missed nearly half of them — including both `tool-auditor` steps, both
> `tool-suggester` steps, `internal-linker` and `component-quality-auditor`.
> This is `bugs_open/144`'s cost reproduced exactly, in the same session that
> then wrote a check whose header cites it. What caught it: the Go audit
> (`config-key-audit --undeclared-recurrence`, which uses `validation.WalkSteps`)
> reported **19** against my **15-ish**, and the disagreement was the tell.
> The two now agree — an independent cross-check from a different code path.

Live `create_work_item` steps carrying an `item_key_prefix`, and whether they
opt out of anti-churn `[SUPERSEDED — see the correction above; 17 was a flat walk]`:

- `recurrence_expected: true` on **2 of 17** — `improvement-loop.record_not_converging`
  and `tool-improver.create_rerender_item`.
- The entire one-shot build chain is in the other 15, unexempted:
  `domain-submitter` (`research`) → `domain-research-classifier` (`vertical_research`)
  → `vertical-exemplar-researcher` (`strategy`) → `domain-strategist` (`briefing`)
  → `build-briefing-agent` (`site_plan`).

Go-side: `insertWorkItem` has **36** non-test call sites and `writeWorkItem` **8**
`[MEASURED 2026-08-23]`; **10** places set `recurrenceExpected: true`.
(Census date recorded per the owner ruling of 2026-08-22 — re-run
`git log --since=2026-08-23 --diff-filter=A -- platform/orchestration/actions/`
before quoting these.)

## 2026-08-23 — the 090 loop produced NO VERDICT, and I am saying so rather than skipping it

Intake `df0b3d97-bb22-4b42-a85c-030ff42536cf`, run `655c8508-a265-4926-8625-b54beff26869`.
It ran to `COMPLETED` and wrote **5 bundles and nothing else** — `SELECT DISTINCT kind`
over `diagnosis_artifacts` returns `bundle` alone; no `iteration_note`, no
`council_report`, no `doc_notes` row.

This is the documented shape, and the SessionStart hook warned me about it before I
filed: **`load_work_item_actions.go` is 82,669 bytes**, and a 090 whose scope lands in
a file over ~60KB returns bundles and no verdict. The discriminating check (only
readable after the first bundle) confirmed it on iteration 2:

```
_(body omitted — 25397 chars, and 50544 of the 60000-char body budget is already
  spent. It was found; it did not fit.)_
```

**Demand control, so the silence means something:** `doc_notes` took **21 rows** in the
same two-hour window. The recorder was alive; it simply had nothing to say about my run.

Iteration 2 *did* render both deciding arms (`newestAge < 3.0` and `terminalCount >= 2`
both present in the bundle body), so the diagnoser saw the mechanism — it just never
wrote a conclusion.

**So I take the owner ruling's escape hatch EXPLICITLY** (2026-07-31: a `bugs_open/`
file asserting a cross-cutting root cause is not filed until it has been through the
loop, *or the filing session states plainly why it substituted equivalent first-hand
verification*). Substituted here, and this is the substitute: the live index predicate
read from `\d site_work_items`; the whole anti-churn block read in place; the
`site_specs` submission timings; the 36-event negative control on the dedup arm; and
the five mutations below. Every one is in this file with its query.

## 2026-08-23 — mutation proofs, because a passing test proves nothing on its own

Run against a `git archive HEAD` extract with only this lane's files overlaid (the
shared tree does not compile — see below). All five caught:

| # | mutation | caught by |
|---|---|---|
| M1 | remove the `retry_after` column append | arg-count mismatch, 4 tests |
| M2 | restore the legacy DROP on arm A | `..._WithinCycleWindow_DefersRatherThanDrops` |
| M3 | restore the `unresolved` BRAND on arm B | `..._TwoCompletedPredecessors_DeferRatherThanPoison` |
| M4 | delete the kill switch | both `TestAntiChurn_KillSwitch_...` subtests |
| M5 | flat 3h interval instead of the window REMAINDER | 4 tests, incl. all three boundary cases |

## 2026-08-23 — three shared-tree hazards, all verified first-hand

1. **`load_work_item_actions.go` is dirty by a THIRD session** — `+4/−1` at
   `FailWorkItemAction:1307`, `bugs_open/345` candidate 2. Flagged by the loanzy.uk
   lane; confirmed with `git diff --numstat` rather than taken on report.
2. **The working tree does not compile**, and this is worse than the first item.
   `go build ./platform/orchestration/actions/` exits **1**: `applyCTARecompute` at
   `rerender_page_sections_action.go:550,552` is mid-signature-change by another
   session. `git archive HEAD` of the same package exits **0**. So `make build-*`
   (committed HEAD) is fine and **anyone who builds an image today will not notice**,
   while anyone running `go test` on this package is blocked — an asymmetry that reads
   exactly like "my test setup is broken". Worked around with a HEAD-overlay script;
   see the RUNBOOK.
   - The 345 hunk is also **half-written**: the caller passes an 8th argument to
     `applyWorkItemFailureLadder`, whose callee is still the 7-argument HEAD version. So
     overlaying my own copy of the file onto HEAD *also* fails to build until their hunk
     is stripped from the build copy. Stripped there, never in the tree.
3. **`cmd/config-key-audit` tests are broken at HEAD+tree too** — `livedeclarations_test.go`
   references `livespec.DeferredDeclarations`, which another session's uncommitted
   `platform/livespec/livespec.go` no longer defines. `go vet` on a clean HEAD extract
   exits 0, so it is theirs, not mine. My own audit tests were run on an overlay.

## 2026-08-23 — the council submission was DISPATCHED TWICE, and the first one never left

`097` printed its `SAVE: SUBMISSION_CORR=94c196fa-…` summary and then failed:

```
Error from server (AlreadyExists): pods "kcat-cgate-1787507905" already exists
```

The publish is at `097:263`, **after** the summary at `:260`, so the script prints a
convincing receipt and then drops the message. The pod name is
`kcat-cgate-$(date +%s)` — **one-second resolution** — and another session ran `097` in
the same second.

**This is NOT the "a missing orchestration row is latency, do not retry" case**, and the
difference is worth stating because the standing guidance points the other way: there I
had a *positive failure signal* — `kubectl run` returned non-zero, so the container never
started and `kcat` never ran. Retrying cost nothing because nothing had been dispatched.
Re-run correlation `f610741f-5054-41e8-b0b7-54915d79ba92`, confirmed live at
`review_editquality` by querying `orchestration_states` on `fix_correlation_id`.
`94c196fa` should be treated as never having existed.

## 2026-08-23 — from the loanzy.uk lane (live greenfield build, garden-tools.uk)

- `research_garden-tools.uk`: `created_at` **17:17:15.482Z**, `completed_at`
  **17:44:59.593Z** — **27m44s apart**. The brake keys on `MAX(created_at)`, so the
  window closes at **20:17:15Z**, not 20:45. **An operator reasoning from `completed_at`
  gets the boundary wrong in the UNSAFE direction on a long-running item** — on a page
  build that gap could be hours. Worth its place in the bug file.
- Time-to-first-agent was **24m52s** (submit 17:17:18Z → claim 17:42:10Z), which is
  queue depth ÷ tick rate.
- `build-pipeline-trigger`'s `find_dispatchable_site` is **FIFO by work-item
  `created_at`**, one site per ~90s tick, and **a site with ANY `claimed` item is
  invisible** until it clears. So a re-submission staged while the site has something
  claimed will not run and **will look exactly like suppression**. Any Arm A
  verification must snapshot whether the site had a claimed item at that instant, or
  the negative result is uninterpretable.
- **REFUTED, by that lane, about its own earlier claim to me:** the picker does *not*
  walk sites in ascending `site_id`. That was 14 consecutive ordered samples that
  stopped being ordered twenty minutes later; the selector never mentions `site_id`
  except as a final tie-break. I had already written the `site_id` version into my plan
  and have removed it. Recording it here because I nearly carried a refuted mechanism
  into a bug file on a peer's say-so — **a subagent's or a peer's report is another
  doc**, and this one was corrected by its own author before I could ground it.

## 2026-08-23 — COUNCIL: REJECTED, guardian hard veto. The veto is right and I am not contesting it

Corr `f610741f-5054-41e8-b0b7-54915d79ba92`, round 1. 14 seats: 11 approve, 2 object, 1 veto.

**The veto, in its own words:** *"The customer-facing bug is already fully closed by edit 4
alone… Edit 1 is therefore not required to fix the filed bug; it is a separate, fleet-wide
architecture decision bundled into an urgent point-fix submission. That is the veto-criterion
pattern: architecture change dressed as a point fix."* And on shape: the deferral *"flips
default behaviour for everyone and only offers a global env-var kill switch"*, where the owner
ruling of 2026-08-02 §2 asks for opt-in-per-caller with the unsafe side OFF.

**Both halves are correct, and the first one is the one I should have seen myself.** Migration
572 alone closes `bugs_open/326`. I had the deferral evidence in hand and bundled it, which is
exactly the pattern the criterion exists to catch. CLAUDE.md: *"A veto on SCOPE is not answered
by resubmitting with better measurements."* So I have not resubmitted.

**What I did instead**, which is the guardian's own named alternative:
- Landed edits 3/4/5 + docs (`d0930af6f`, `74c527f56`) — it said it would approve those alone.
- **Applied 572.** Census 19 → 14 findings; **no build-chain step is undeclared any more.**
- Routed the seam to `architecture_review/RFC_048_…`, with the patch beside it, three options
  costed, and my own view marked as a view.
- **Reverse-applied my own patch out of the shared tree** so a vetoed change cannot ship on
  another session's roll. Verified their `bugs_open/345` hunk survived byte-for-byte (`+4/−1`,
  the `stop_on_repeat_failure_item_types` line) — reverting a shared file is exactly where you
  clobber somebody.

**Objections acted on, not banked:**

| seat | objection | what I did |
|---|---|---|
| debug_historian (med) | unconditional `snapshot_agent` before a fenced UPDATE — a re-run takes a second snapshot labelled "pre-update" over an already-updated row | both migrations now gate the snapshot on the same pre-state marker that drives the UPDATE; a re-run is a true no-op |
| debug_historian (med) | version-blind `WHERE` on a table where four types carry TWO active rows | **checked: none of my five is in that state** (one active row each, version 1) — but both migrations now REFUSE outright if any target has duplicates, because "it does not fire today" is not a property a file can rely on |
| bug_historian (low) | is `bugs_open/091` the same root cause? | **No, and it is not open.** `bugs_closed/091` (CLOSED 2026-08-03, live v1.0.1237) is the INDEX arm dropping a second, *different* finding while an earlier item is OPEN — remedied by `refreshOnConflict`. Mine is the anti-churn brake above it. Sibling arm, same family, already fixed; the class continues in `bugs_open/184`. The seat flagged it from the title alone and said so. |
| architecture (med) | `retry_after` now carries two causal meanings; no test exercises the three readers against a deferred-but-never-failed row | **True, and I did not write it.** Recorded as owed work in RFC_048 §4 and pointed at RFC_043, whose contract it extends |
| tooling_provenance (low) | no `doc_notes` row for the design decision under a stable subject_key | written, subject_key `create_work_item` |
| reuse_agent (low) | confirm the audit extends the existing binary rather than forking one | it does — a new mode in `cmd/config-key-audit`, dispatched from the same `main.go` |
| editquality (low) | 572 assumes four targets share the step name `create_next_item` | true, and self-protecting: the verify block RAISEs if a targeted step is not a `create_work_item` step |

**The one I could not act on**, recorded because it bounds what the verdict is worth: the
guardian noted *"No SQL exists in this schema to inspect index definitions (idx_swi_dedup) or Go
status-list constants — the plan's central claim about which arms the index excludes cannot be
checked by this council tier and must be taken on the author's word."* The council could not
verify my central correction. It is in this file with its query, and `prior_art_librarian`
independently attached a `pg_indexes` check that agrees.

## 2026-08-23 19:23Z — PROVEN LIVE, on a real build, by two observers

`garden-tools.uk` died at its second hop on an unrelated defect (`bugs_open/376`). The lane
re-submitted — the exact operator move this bug is about — and the row appeared:

```
07b589a9-… | complete | 17:17:15.482481+00      <- the terminal sibling
3921bde4-… | triaged  | 19:23:06.330863+00      <- 2h05m51s later, INSIDE the 3.0h window
```

`[MEASURED 2026-08-23]` independently by me minutes afterwards, where it had already moved to
**`claimed`** — dispatched, not merely filed, which is the stronger claim. `retry_after` NULL on
both rows, correctly: the deferral is vetoed and unshipped, so this proves the classification
alone.

**Attribution control:** `recurrence_expected` re-read on the live definition in the same breath
and still `true`.

> **CORRECTED same day.** I first wrote that without this control "an insert at 2h05m is also
> consistent with 'the window elapsed'". **That is arithmetically false** — the threshold is 3.0h
> and the offset 2h05m51s, so elapse was never a rival, by 54 minutes. The control is still worth
> having, for a different reason: it rules out that something OTHER than the declaration stopped
> the brake — a retuned threshold, a disabled block, or 572 rolled back between the snapshots.
> **Attribution, not elapse.** Caught by the loanzy.uk lane before it spread.
>
> The habit was right and the stated reason was invented afterwards, which is its own failure
> mode: reaching for a control reflexively is good, narrating a purpose you have not checked is
> how a correct check ends up discrediting the block it appears in.

**Outcome meanings were agreed BEFORE the result existed** (new row ⇒ the fix works, not a
revision of the account; no row ⇒ a live defect investigated as such). That mattered: I had
already invalidated their original prediction by applying 572, and fixing the reasoning
afterwards would have been choosing an interpretation to fit a number.

**Three instrument checks the measuring lane volunteered, all of which sharpen the result:**
1. The `claimed`-items snapshot (0 both sides) is an **unused control** — it would only have
   discriminated on a null result. Recorded so nobody later reads it as corroboration.
2. **The key was free by ~40 seconds.** `needs_vertical_research` reached `failed` at 19:22:13Z.
   On my own earlier instruction ("re-submit whenever, whatever the offset") the test would have
   run while that item was still `triaged` — non-terminal, inside `idx_swi_dedup` — and the
   classifier's `create_next_item` would have conflicted, giving a **false negative on the fix's
   first live test**. Their timing discipline caught what my instruction would have broken. That
   is twice in one evening my guidance to that lane would have produced a wrong reading.
3. A re-submission **supersedes** the prior `submission` spec (`is_current` t→f) and writes a
   second. So `aspect='submission'` is not one-row-per-site; anything reading it must not assume.

**And the residual this proves is still open:** the row inserted because migration 572 declared
this step. The 14 still-undeclared steps and the 36 Go call sites would have had the request
destroyed exactly as before. That is what RFC_048 is for, and this measurement is evidence for
its premise, not against it.

## 2026-08-23 — the classifier is reproducible enough to be a FIXTURE (from the loanzy.uk lane)

Incidental to 326 but useful to anyone testing on this estate. The re-submission re-ran
`domain-research-classifier` from scratch on the same bare domain, giving two independent runs
to compare `[MEASURED 2026-08-23]`:

| field | run 1 (17:44) | run 2 (19:26) |
|---|---|---|
| category | hub | hub |
| site_type | content | content |
| confidence | **0.82** | **0.82** |
| suggested_style | modern-light | modern-light |
| page_count_estimate | 12 | 12 |
| recommended_builder | pageflow-builder | pageflow-builder |

**Every structured decision field identical, confidence to two decimal places.** Only the
free-text `industry_tags` varies, and in wording rather than meaning (8 tags → 10;
`buying-guide-platform` → `buying-guides`).

**Why it matters beyond this bug:** a before/after test that re-runs this classifier is
comparing like with like on the structured fields — not something to assume of an LLM step with
no temperature pinned, and worth knowing before anyone builds a harness that avoids re-running
it. It does NOT license treating the free-text fields as stable.

## 2026-08-23 — what the fix ENABLED, and the two lanes' evidence propping each other up

Recording this because it is an argument for fixing retry paths generally, not just this one.

The `loanzy.uk` lane could not test `bugs_open/376`'s central claim — *"retry cannot help,
because sampling permutes the exemplar set rather than re-drawing it"* — without re-running the
front door, which is precisely what 326 made impossible. **Migration 572 bought them the
control.** Their fourth attempt ran off freshly re-derived specs and returned the same three
organisations in a fourth permutation, turning a conditional claim into an unconditional one.

**And their reproducibility finding is what made that control interpretable.** Because the
classifier's *structured* verdict was identical across both runs while only the free text moved,
"the pool is fixed" can be separated from "the input never really changed". Had the structured
verdict drifted too, the control would have proved nothing. Neither lane designed that; it fell
out of re-submitting for my test.

### What I checked rather than took, and it was stronger than reported

Their inference rests on the exemplar step's input having genuinely varied. Verified at the
config `[MEASURED 2026-08-23]`:

- `select_exemplars` reads `{{.site_specs}}` — the **whole** blob, so `industry_tags` is in
  scope. ✅
- **`identity.competitors_found` ALSO changed** between the runs (5 → 6 entries,
  `sgs-engineering.com` added), which they had not noticed. So **two** independent site-specific
  inputs moved and the pool still did not.

### The finding that came out of checking — and it is theirs to own, not mine

The prompt says: *"Prefer sites named in `identity.competitors_found` when they are genuinely
strong; otherwise use well-known leaders of the vertical."*

**Not one selected exemplar is in `competitors_found`, in any of the four attempts** —
gardenersworld / thespruce / which.co.uk against six UK garden-tool retailers, zero overlap. The
`competitors_found` branch has **never fired**. Every selection came from the fallback, i.e.
model priors with no site-specific input at all.

So the pool is not fixed *despite* fresh specs; it is fixed **because the fresh specs never
reach the decision**. That separates two different remedies (an exclusion list / `on_error`
versus an upstream identity question). Sent to that lane; not filed in theirs by me.

> **⚠ CORRECTED 2026-08-23, within the hour — I wrote "a prompt branch with a live input that
> has never once been taken" and called it a DEFECT. Both halves overreach.**
>
> **On "never":** my 0/4 is 0 out of the four runs `orchestration_states` still holds, on ONE
> site, in ONE vertical, on ONE afternoon — a population the other lane generated for my test.
> The loanzy.uk lane caught this and bounded it correctly in their own file: it licenses
> *"the branch is unexercised and worth investigating"*, never *"the branch cannot fire"*. At
> n=4 those are indistinguishable.
>
> **And the denominator is worse than either of us said**, which I found checking their bound
> and which corrects them too. They wrote "four runs in its **entire history**". That is the
> ~24h retention window wearing a measurement's clothes — the same trap I logged in
> `WRONG_CALLS.md` this morning, hit again by two people in one evening. The durable evidence
> says otherwise `[MEASURED 2026-08-23]`:
>
> ```sql
> SELECT count(*), min(created_at)::date, count(DISTINCT site_id) FROM (
>   SELECT created_at, site_id FROM site_work_items         WHERE item_type='needs_strategy'
>   UNION ALL
>   SELECT created_at, site_id FROM site_work_items_archive WHERE item_type='needs_strategy') q;
> --  32 | 2026-04-02 | 27
> ```
>
> **32 `needs_strategy` items across 27 sites since April.** So this agent has run many times;
> the four visible runs are what survived reaping. The true denominator for "has the branch ever
> fired" is **not knowable from `orchestration_states` at all** — the historical selections are
> gone. My claim was not merely under-powered, it was measured against a table that cannot
> answer the question.
>
> **On "defect":** the other lane's point, and it is the better one. `identity` found RETAILERS
> for a domain the classifier called `hub`/`content`. "Not genuinely strong" as exemplars for a
> content site is a **defensible reading of the prompt**, not a bug — so *the branch not firing*
> and *the branch working correctly on unsuitable input* produce the identical observation, and
> nothing here separates them. Their §4a says that instead of asserting the upstream defect I
> sketched, and they are right to.
>
> **The discriminator, for whoever wants it:** one greenfield build in a vertical whose
> competitors genuinely ARE content properties.
>
> **⚠ SETTLED 2026-08-23 20:0xZ, AND AGAINST BOTH OF US — the branch FIRED.** A fifth draw, 30
> minutes after the four that looked settled, dropped the refused host and picked
> `burgonandball.com`, which **is** in `competitors_found`. So `0/4` was a run, exactly as the
> bound above said it might be, and the mechanism I proposed — *"the pool is fixed because the
> fresh specs never reach the decision"* — **is false**: the specs reach the decision, at least
> sometimes. The honest statement is the count: **4 of 5 draws contained the refused host.**
> Biased, not fixed. Retracted by the `loanzy.uk` lane, verified here at the artefact.
>
> **Nothing caught this except the system continuing to run.** Not a check, not either session.
> That is worth more than the finding: a run of identical observations became a mechanism, and
> only a free counter-example broke it.

**The transferable bit for this lane:** I nearly recorded "the pool is a property of the
vertical" as received. Checking a peer's inference — not their measurement, their *inference* —
is what turned a true-but-shallow mechanism into an actionable one. `[a-subagent-report-is-another-doc]`
applies to peers, and it applies to the reasoning as much as the numbers.

## 2026-08-23 20:09Z — the fix's strongest evidence, arriving unasked: the re-submitted build is at HOP FOUR

Verified at the artefact, this lane's own query:

```
needs_domain_research   | research_garden-tools.uk          | complete | 17:17:15   <- submission 1
needs_vertical_research | vertical_research_garden-tools.uk | failed   | 17:44:56   <- died here (376)
needs_domain_research   | research_garden-tools.uk          | complete | 19:23:06   <- submission 2, THE FIX
needs_vertical_research | vertical_research_garden-tools.uk | complete | 19:26:57
needs_strategy          | strategy_garden-tools.uk          | claimed  | 20:05:55
needs_briefing          | briefing_garden-tools.uk          | triaged  | 20:09:29
```

Plus a `vertical_landscape` spec at 20:05:45Z, `is_current`.

**The 19:23 re-submission did not merely queue a row — it produced a build that has since cleared
four pipeline stages**, including the hop that killed submission 1. That is the fix demonstrated
end to end rather than at its first insert, and I did not design the test: it fell out of another
lane letting a failed build take its natural course.

**Two claims this kills, one of them mine**, both retracted above: the exemplar pool is biased
(4 of 5), not fixed; and the `competitors_found` branch fires.

**The lesson I am taking, from the lane that made the error and named it better than I would
have:** *if the evidence is a count, the claim must contain the count.* "4 of 5 draws contained
the refused host" was always what was actually known, needs no hedge, and would have survived the
fifth draw intact. Twice today that lane turned a run of identical observations into a mechanism
(14 ordered dispatches → a `site_id` walk theory, broken in 20 minutes; 4 identical draws →
"structurally incapable", broken in 30), and **neither was caught by a check or by me — the
system kept running and contradicted them.** That is luck standing in for method, and it is the
most useful thing either of us learned today.

## 2026-08-24 — I burned 5.0 GB on a recipe that already had a script, and taught it in my own RUNBOOK

The entries above describe using `git archive HEAD` + an overlay to build around a
non-compiling shared tree. **The technique was right and the implementation was harmful. Do not
copy those lines** — `scripts/verify-head-builds.sh` already does it properly:

```bash
scripts/verify-head-builds.sh                          # does committed HEAD still build?
scripts/verify-head-builds.sh --with <file> [--test]   # build YOUR change against HEAD
```

**Measured 2026-08-24:** each extract is ~450 MB, and the `rm -rf` in the pasted recipe is the
*setup* half — it clears the tree that run is about to use, so it only reclaims a tree of the
**same name**, and every variant picks a new one. This lane left **eleven**: `ov`, `ov2`, `base`,
`mut`, `verify`, `headonly`, `headtest`, `trio`, `all3`, `pairtest`, `headkfTT` — **5.0 GB**,
reaped down to 108 KB once CLAUDE.md's new note pointed at the script. `/` was at 75%.

**The part that stings, and is the reason this is a NOTES entry and not just a tidy-up:** I wrote
the harmful recipe into `RUNBOOK_326_retry_the_front_door.md` as a helper function for the next
session, complete with a rationale for why it was necessary. CLAUDE.md says **73 documents** still
spell it out and **66 never delete anything** — I authored one of the 73 yesterday, in a lane
whose whole subject is a mechanism that quietly destroys things. The RUNBOOK section is now
replaced with the script and a correction banner.

**What I did not know and could not have inferred:** the script existed the whole time. CLAUDE.md
gained the note between my last read and this morning. **The check that would have found it is one
`ls scripts/ | grep -i head` before writing a build helper** — the same shape as "grep LANDMINES
for the symbol you are about to trust", and the same failure: I reached for a technique without
asking whether the estate already owned it. The `reuse_agent` council seat approved this lane's
plan on exactly that ground, which is a little sharp in hindsight — it was reviewing the change,
not my scaffolding.

**Also settled 2026-08-24:** the tree compiles again (`go build ./platform/orchestration/actions/`
exits 0 — `applyCTARecompute` landed), so the workaround is not needed for compilation at all any
more. The only reason left to isolate from HEAD is the honest one the script exists for: checking
that a commit does not lean on another session's untracked work.

## 2026-08-24 — OWNER RULED: "D + E now, census alongside". E is BUILT and mutation-proven; D queued behind 333

**E — eight `recurrenceExpected: true` declarations across seven Go files**, each with its
one-line justification at the site: `apply_adoption_plan` (adoption→classifier handoff),
`emit_design_items` ×2 (`needs_composition`, `needs_design` — the build chain's Go half),
`emit_imagery_items`, `flag_page_image_rebuild`, `reconcile_section_data`,
`seed_build_queue` (the canonical retry), `validate_composition_inputs`.

**Read before flagged, and three sites were deliberately REFUSED:**
- `create_tool_cross_link_items.go` — its own comment already decided FALSE, with reasons. Respected.
- `emit_content_card_derive.go` — an action request, but its item_key is shared with a discovery
  CHECK (`ContentImageItemKey`); the coupling's lane decides, not this one.
- `rerender_page_sections_action.go` escalation — genuinely ambiguous (re-request vs
  remedy-did-not-work). Listed, not guessed.

**And a correction that shrank E (brief updated, v3):** the `page_rerender` "bleed" I had
assigned to E comes from two discovery checks (`check_misdirected_cta`,
`check_contact_form_undeliverable`) — DETECTORS filing a re-render as the remedy. When the
re-render completes and the CTA is still misdirected, the brake is RIGHT to two-strike. That
population is `352`'s class. E does not stop it and nothing should.

**Proof:** `action_request_producers_recurrence_test.go` — a comment-stripped source ratchet over
all eight sites (with its own anti-prose mutation guard) + one end-to-end effect test
(`EmitDesignItems` driven through sqlmock in `nav_rebuild_request_test.go`'s worked pattern: a
2-strike history registered so a dropped flag ACTUALLY brands, INSERT pinned to `$12='triaged'`).
**Mutation-proven:** removing the `needs_design` flag fails BOTH tests, the ratchet naming the
site. One existing test corrected (`page_section_satisfiability_test.go` scripted the brake probe
E now skips — helper split so the two still-braked `page-rerender` callers keep the original).
Package green.

**Sequencing:** E's tests call `expectWorkItemDoorStandsDown` — `bugs_open/333`'s helper, in the
tree but uncommitted — so **the E commit waits for 333 to land** (they answered: within the
hour, option 1, I rebase D on top). D is written next against their described final state.
Their two design notes are adopted: a parked row sets `recurrenceExpected` itself so D can never
fire on it, and D's deferral must stay out of `deferred`-status ambiguity.

**Census alongside:** 14 undeclared config steps (list above from today's run). Decision 3 —
who tells their lanes — remains with the owner; not acting on other lanes' steps from here.

## 2026-08-24 — D LANDED (f16c87beb). The ruling is executed: D + E both at HEAD

**D:** the within-cycle arm defers (`retry_after` = window remainder, row at the caller's own
status, dedup slot held) instead of dropping. Two-strike arm verbatim-unchanged — the asymmetry
is the ruling. One sub-case changed and pinned: a two-striker INSIDE the window used to vanish
via the early return; it now falls through to the existing brand. Kill switch
`DISABLE_ANTI_CHURN_DEFERRAL` armed, restores the drop exactly, both sub-cases tested.

**Sequencing worked as agreed:** 333 landed their door first (`6ab0b3434`), I wrote D against
their landed state, and their two constraints held — a parked row sets `recurrenceExpected`
itself so it can never reach the deferral arm, and `deferred`+`retry_after` cannot co-occur from
their path. Their vacuity warning shaped the test header the other way round: `baseItem` has no
`pageID`, so the door never fires in D's tests, and the header says so to stop anyone adding the
helper "for safety" and then chasing an unmet expectation.

**Proof:** five mutations, all caught by NAMED tests — drop the column append; restore the
unconditional drop; delete the kill switch; flat interval instead of remainder; **defer
two-strikers too (option A by the back door)**. Gate: `verify-head-builds.sh` on HEAD
`fb1b8a9be` + only D's four files, full package `ok`. Council round `74d4fa7d`, dispatch
verified at the orchestration row, live at `review_editquality` at commit time.

**Named residual (RFC_048 §4, and in the submission's risks):** `retry_after` now carries two
causal meanings — RFC_043's post-failure cooldown and this deferral-at-birth. The three readers
need not distinguish them (both mean "not before T") and the single-renderer test pins the
predicate at all three statements; a live-DB skip-then-serve test is the part sqlmock
structurally cannot supply.

**Go changes are inert until an image is rebuilt and rolled** — D and E ride the NEXT chassis
build after `f16c87beb`. If the fresh chassis the owner mentioned was built before that sha,
it carries neither; check the `build provenance` stamp per service, never infer from the roll.

**Still open, unchanged:** decision 2 (`on_dedup`/573 — separable, unruled; 573 stays `_HOLD`
and its gate grep still correctly returns empty, since D deliberately did not add the key) and
the 14 undeclared config steps (their lanes; census recorded).
