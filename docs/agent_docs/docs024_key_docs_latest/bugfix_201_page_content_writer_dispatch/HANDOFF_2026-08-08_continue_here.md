# HANDOFF — `bugs_open/201` lane, 2026-08-08 · **start here.** Supersedes `HANDOFF_2026-08-07_continue_here.md`

**Both of 201's symptoms were already fixed, live and proven before today** (see the superseded
file; nothing there needs re-deriving). Today's work was the thing that lane had left open:
**`RFC_017` — measured, ruled on by the owner, built, council-approved, and now LIVE.**

## 1. State in one table

| | what | evidence |
|---|---|---|
| `bugs_open/201` symptoms 1 & 2 | fixed, live, proven | 08-06 / 08-07, superseded handoff |
| **`RFC_017` — the missing number** | **MEASURED 08-08** | RFC § "The missing number"; `RUNBOOK` R8 |
| **`RFC_017` — the decision** | **OWNER RULED: fail-CLOSED.** Option 2 declined, option 3 explicitly deferred | RFC ✅ DECIDED banner |
| **The flip** | **built, council APPROVED r1, LIVE on `v1.0.1268`** | pod-verified both replicas, below |
| Behaviour of the fail-closed branch | ❌ **never executed in production** | see §3 — this is the only open verification |
| Register | `WII-011` + index row (1,794, uniqueness checked) | same commit as the seam |

**Commits (all on `087_towards_multiple_domains`):** `01ce2adfb` measurement · `e492c2abc` landmine ·
`1c5d9ceb5` the flip + `WII-011` · `111e5b817` the 4-of-8 correction + verdict triage.
Council corr **`a104d454-a4ff-4c95-a578-9a7e48c95100`** (APPROVED r1, read and triaged).

## 2. What the change actually is

`ItemVerifier` errors used to **fail OPEN** — "I could not check" stamped the item `complete`, for
all 8 verifiers, licensed by a doc comment. Now **fail-CLOSED by default**; fail-open survives only
as `RegisterVerifierWithPolicy(t, v, VerifierPolicy{FailOpenOnError: true})` — an **opt-in field,
unsafe default OFF**, per the owner's 2026-08-02 shared-seam ruling. **Nothing opts in today.**
`RegisterVerifier` keeps its signature, so the flip reached all 8 registrations without editing 8
files; `GetVerifier` now returns `(ItemVerifier, VerifierPolicy)`.

Two collateral fixes rode with it, both real: `blockedCompletionReason` (the old message hard-coded
*"found the defect still present"* and read a key the error payload does not have — it would have
claimed a finding the verifier never made, with an empty body), and `verificationDecision` extracted
**pure** so the policy is testable at all.

**Do not re-litigate:** whole-registry scope is the ruling, not an oversight. The unparseable-spec
branch takes the same policy deliberately. Option 3 (`Indeterminate`/park) was considered and
deferred by the owner **with the retry cost as its trigger** — it is not an oversight either.

## 3. THE ONE OPEN VERIFICATION — and the trap in it

**Deployment is proven. Behaviour is not.** Keep these apart in anything you write.

Pod-grep, both replicas, `v1.0.1268`, one exec each, positives + negative together:

```
fails closed (RFC_017) : 1   blocking completion (fail-closed) : 1
failing OPEN by explicit policy : 1   verification_unavailable : 1
failing open  (removed spelling): 0   <- negative control
```

The gate demonstrably **ran** post-roll: a `hardcoded_section_colors` item verified at 18:58:44Z, 63
seconds after the pods came up. But it returned `Resolved:true`, so **the fail-closed branch has
never executed in production.**

**What proof requires:** an absent-target case must recur. Then the item must land `triaged`/`failed`
(NOT `complete`) with `error` beginning *"completion blocked: verification could not run"* and
`_verification.fail_open = false`:

```sql
SELECT status, attempt_count, left(error,90) AS err, result->'_verification'
FROM site_work_items WHERE result->'_verification'->>'status'='error'
ORDER BY updated_at DESC;
```

⚠ **That case has fired TWICE in the registry's entire life.** Do not schedule a wait for it, and do
not manufacture one on a live customer site without asking. Waiting is the correct posture.

### Four traps, each of which yields a confident wrong answer

1. **`fail_open` dates ERROR rows ONLY.** It is written on the error branch alone, so `verified` /
   `defect_persists` rows carry no era marker in *either* era. My own census filter
   `result->'_verification' ? 'fail_open'` returned 0 post-roll rows on a day one demonstrably
   existed. Date a non-error row by `updated_at` against the roll (18:57Z), never by payload shape.
2. **A `_verification.status='error'` row means the OPPOSITE side of this roll.** Pre-roll it means
   the item COMPLETED; post-roll it means BLOCKED. Both readings are live in one table. Full entry in
   `LANDMINES.md`.
3. **`result` is OVERWRITTEN per completion attempt**, so any count off `_verification` is a floor,
   not a history. Independently corroborated by the `bugfix_071` lane's landmine the same day: the
   census *"counts surviving verdicts, not verifications performed, and systematically under-counts
   exactly the refusals"*.
4. **The pre-roll binary baseline was NOT taken** (my miss). The negative needle's "before" half is
   source-derived (`1c5d9ceb5^` lines 64, 87) not binary-derived, because the fleet is uniformly
   1268. Sound, but say which it is.

## 4. Open, and NOT this lane's to close

- **`empty_sections_loop_integrity` — the cheap fix.** `bugs_closed/032`'s own "stronger option":
  ask whether the page still declares the slot, return `Resolved:false`. Correct on 2 of 2 observed
  cases, and it converts **three futile rebuilds into one true verdict** now that fail-closed is
  live. Told twice; their call.
- **⚠ THE NAMED COST IS NOW ARMED.** With fail-closed live and that verifier unchanged, an
  absent-target case burns up to `max_attempts` (3) **page rebuilds** before a human sees it. Applies
  to **4 of 8** verifiers (`empty_section`, `truncated_component`, `content_duplication` = page gone,
  `page_canonical_collision` = site gone). **This is RFC_017 option 3's trigger** — if it shows in
  the numbers, that is the case for building the parking outcome.
- ~~**Two live pages still serving empty sections**, recorded `complete` on 08-03:
  `ai-guides` + `insights`, site `1368e337-dd1d-4799-bbb3-8221a1b79bcc`, slot `featured-content`,
  334-byte shells. **Unfiled**, and no `featured-content` item has ever existed fleet-wide although
  `findEmptySections`' predicate matches both right now. **Why detection never re-filed is NOT
  established** — dedup and `bugs_closed/041` ruled out by measurement. **This wants a `090` run**
  and is the single most valuable next step in the area.~~
  > **CORRECTED 2026-08-08 (late) — two slot names conflated, and the answer is now established.**
  > The live slot is **`featured-content`**; the work items name **`featured-article`**. Items for
  > those two pages have existed **twice** (April → `unresolved`; 08-03 → `complete`), so
  > "never existed fleet-wide" was an artefact of grepping `item_key` for the wrong spelling.
  > The site is **`finetuning.uk`**, held by two other lanes.
  > **Both halves are now measured** (NOTES, 2026-08-08 late): (1) the 08-03 items closed via the
  > absent-target fail-OPEN path — they ARE §3's "fired twice in the registry's entire life" cases,
  > and `RFC_017` now makes this same case fail closed; (2) `featured-content` is unfiled because
  > **site discovery has no recurring driver** — 0 of 5 `scheduled_tasks` rows targeting
  > `quality`/`completeness`/`design-discovery-agent` are enabled, all five are `oneshot-*`, and no
  > CronJob fires them. Detection follows whichever site a lane is hand-driving.
  > **CLOSED OUT 2026-08-09.** Filed as **`bugs_open/230`**; transferable pattern in **016b §9**;
  > the two finetuning lanes have been told in their shared cold-start
  > (`finetuning/HANDOFF_2026-08-08_continue_here.md` §6) and **nothing was dispatched at their
  > site** — owner's call.
  > ⚠ **The `090` returned NO VERDICT** (intake `c5778c3e-…`, run
  > **`2ccc7551-76d3-40d2-ac2a-01d8120ea0fb`**): `COMPLETED`, work item `complete`, **5 bundles and
  > no verdict artifact**, no `doc_notes` row — `diagnose_route.max_iterations` is 5, so it
  > exhausted its budget while still re-scoping. **NOT the known ~60KB body-omission mode**: the
  > final bundle is 72,310 bytes, carries `findEmptySections` whole, no truncation marker. So `230`
  > rests on first-hand verification and **says so in its §7**, per the owner ruling of 2026-07-31.
  > A run that ends without a verdict is evidence for nothing — do not cite it either way.
- **`bugfix_071` lane's landmine** — the success UPDATE never clears `site_work_items.error`, so a
  row refused then completed keeps the refusal text. Fail-closed makes it *commoner*, not caused by
  us. Plausibly a one-line fix (`error = NULL`) on the success path; **theirs, not ours.**
- **`RFC_017` option 3** stays open in the RFC.

## 5. Lane files

`PLAN` · `RUNBOOK` (**R8** = the verifier-error census + the three ways it lies) · `NOTES` (evidence
and every misstep, newest at bottom — including today's three) · `README_where_we_are` (owner's plain
prose) · `SUMMARY_2026-08-06`, `SUMMARY_2026-08-08` · `SUBMISSION_RFC017_fail_closed_r1.json`.
Outside the lane: `architecture_review/RFC_017` (the decision record),
`docs026_concept_register/register/work-item-integrity.md` **`WII-011`**,
`work_item_completion_integrity/HANDOFF_2026-08-08_fail_open_measured…` (consumer primary),
`WRONG_CALLS.md` (the 2-of-8 error), `LANDMINES.md` (the inverted fact).

## 6. The thing worth carrying out of this lane

**Five times in one day the check that caught an error was one already written down** — R8 trap 3
(wrong table), the same trap's sibling (grepped a spelling, not a class — published "2 of 8" when it
was 4, into an owner ruling and a council round), the pre-roll baseline lesson another lane had
logged hours earlier, and my own `fail_open` marker answering a narrower question than I asked it.
Four of the five were caught by someone or something else, after the fact. **The council seat that
found the biggest one did not spot the error — it refused the FORM of the claim** ("asserted rather
than enumerated"), and the enumeration it asked for took one command.
