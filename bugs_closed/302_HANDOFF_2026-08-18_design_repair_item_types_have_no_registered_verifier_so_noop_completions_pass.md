# 302 — design-repair item types have no registered verifier, so a no-op "repair" completes unverified

> ## ✅ CLOSED 2026-08-19 — **FIXED, LIVE and PROVEN IN PRODUCTION on `v1.0.1314`.** Council APPROVED r1 (`edfef8cc`).
>
> **What was fixed:** gate 1b's unreadable-payload arm, which silently waived the very assertion a
> roster entry exists to make — and [MEASURED] on 5 of 11 occasions completed an item this gate had
> **already refused** one attempt earlier. It is now a per-type declaration whose zero value is not a
> policy: undeclared is a build failure, and undeclared at runtime abstains.
>
> **Proven four-way, 2026-08-19 09:15Z** (induced, with the owner's authorisation — natural demand
> never arrived): **A** no result → REFUSED `handler_result_unreadable`; **B** readable all-zero →
> refused `handler_reported_no_change` (distinct); **C** non-zero counter → COMPLETES; **D**
> non-roster type, no result → COMPLETES. A alone would have proved almost nothing.
>
> **What was ANSWERED rather than fixed:** the filing's headline — should the design-audit family get
> artefact verifiers — is answered **"not by this route"**, with the reasons already on the record in
> `verifier_coverage_test.go` (a browser on the completion path; `catJudgement` with nothing to
> re-run) and a producer split showing `needs_design_review` has FOUR producers over 1,296 lifetime
> rows. That is a decision, not a gap.
>
> **What was spun out so closing this loses nothing:** `bugs_open/317` (the claimed-item-timeout
> sweep bypasses both gates — latent, 0 occurrences, reachable the moment a carrier is re-enabled).
> Two further follow-ups need a DECISION before any rule could be written honestly and are recorded
> in §"the three follow-ups" below: `needs_design_review` semantics (an analysis blob may legitimately
> BE the deliverable) and `spacing_fix`/`responsive_fix` success-envelope shapes.
>
> **Not claimed:** the refusal's cost is real — a refused row burns `max_attempts` and waits at
> `failed` for a human, because WII-018's retraction has never run and this producer's carrier is off.
> That is RFC_017's knowingly-accepted cost, not a tidy-up.
>
> Lane: `docs024_key_docs_latest/bugfix_302_design_repair_verification/`. Register: WII-017 (amended
> + proven), WII-011 (fifth `_verification.status` value).

**Filed 2026-08-18 by the `finetuning_uk_service` (merged) lane.**

## Verification statement (owner ruling 2026-07-31 compliance)

A `090` ran first — twice. Run 1 (`f60d72d6`) FAILED with NULL step errors after
five bundles. Run 2 (`361605fe`) returned **UNVERIFIABLE** — "NOT confirmed
(stopped: scope-not-narrowing)", the broad hypothesis as submitted marked
REFUTED, and the trail explicitly handed to a human. **This filing substitutes
declared first-hand verification for the loop's confirmation, per the named
escape hatch:** the narrow claim below is established by direct code reading
with quotes, and the loop's own citation pointed at the deciding arm.

## The narrow, code-verified claim

`platform/orchestration/actions/complete_work_item_verification.go`
(`verifyBeforeComplete`):

```go
verifier, policy := checks.GetVerifier(itemType)
if verifier == nil {
    return nil, true, abstained   // <- completes; records only an abstention
}
```

The verifier registry (`discovery_checks/`, `RegisterVerifier` call sites) holds
**eleven item types, all discovery-check shapes**: `revenue_shape_cta`,
`missing_conversion_path`, `content_duplication`, `decision_regression`,
`page_canonical_collision`, `orphan_element_refs`, `empty_section`,
`truncated_component`, `dead_fragment_link`, `unbuilt_internal_link`,
`literal_markdown`. **No design-repair item type is registered** — nothing in
the family handled by `webdesign-agent` / `component-template-fixer` /
`color-variable-fixer`. So for those items, gate verification abstains and
completion passes with no check that the named defect changed at the artefact.
(There is a second abstain arm just above for unknown result shapes — same
consequence.)

## The evidence this explains (measured, not asserted)

finetuning.uk, 2026-08-12 (evidence rows deliberately left `complete`; full
tables in `finetuning_uk_repair/NOTES` §"ALL FOUR REPAIRS"): four repair items
completed in 6 minutes; the served page byte-identical on every defect before
vs after; zero writes to `page_components`/`content_components` in the window;
every `result` a four-key design-token blob with no `changes_made`. Same shape
on `needs_design_review` items of 2026-08-11. The 08-09 audit's
`hardcoded_section_colors` finding was re-filed by the 08-12 audit because
nothing had changed.

## What is NOT claimed (the 090 refused the broader claim; respect that)

WHY the handlers return analysis blobs instead of performing repairs is
**undiagnosed** — the loop could not narrow it (its runtime evidence covered an
unrelated target) and marked the broad claim REFUTED as stated. Do not treat
this file as saying "the handlers are broken"; it says **the gate that would
have caught them is absent for their item types**. The blob question is real
and separate; the loop's `NextScope` pointers (completion path for
`color-variable-fixer`, which was absent from indexed scope) are the thread to
pull.

## Fix candidates, ordered by what closes the door

1. **Class fix at the gate:** for item types whose *name or category marks them
   as repair-shaped*, make a missing verifier REFUSE completion rather than
   abstain (fail-closed for repairs, abstain-open only for informational
   types). Closes the whole class, including future unregistered types — the
   current design makes every NEW repair type silently unverified by default.
2. **Instance fix:** register artefact-level verifiers for the design-repair
   family (before/after fetch of the named defect, the same discipline the
   eleven existing verifiers use). Necessary anyway for repairs to be provable;
   does not close the door for the next type.
3. **Not a fix:** relying on audits re-filing the same finding — that is the
   current de facto detector and it costs a full audit cycle per miss.

## Relations

`bugs_open/201` §symptom 2 — same class, different item family, established the
"unregistered verifier + mark_complete checks nothing" pattern from code;
OWNED by `bugfix_201_…` lane (this file does not route work at their case).
`bugs_closed/213` — verifier/producer mismatch class. Fleet memory:
"a `complete` work item is not a repaired artefact". 016b §9 entry added this
date. 090 artefact trail: correlations `f60d72d6` (failed), `361605fe`
(UNVERIFIABLE, bundles + evidence trail on the item row).

## How to verify a fix

Re-run the finetuning.uk repair items (or any design-repair item) after the
change: a no-op completion must FAIL (candidate 1) or the verifier must measure
the artefact and refuse (candidate 2). The four retained `complete` rows are
the regression fixtures.

---

## SCOPING UPDATE, same day — the gate already anticipates this class, and that changes which fix bites

Read in full: `complete_work_item_verification.go` + `complete_work_item_no_change.go`.
There are TWO gates, and the picture is finer than the filing above:

1. **Gate 1b (`noChangeGates`)** — opt-in per item type, counter-path based, from
   `bugs_open/213` D1's council-reviewed design. **`dark_section_audit` is ALREADY
   opted in** (with measured justification, 2026-08-12). But its counters
   (`response.fix_result.total_fixed`, …) are read from the handler result — and
   the token-blob results carry NO counters at all, which takes the
   **`unknownShape` arm: abstain-and-complete with a note**. So even the opted-in
   type sails through on exactly the failure shape we measured. **Opting the
   other design types into 1b would change nothing** — blob results defeat 1b by
   construction.
2. **Gate 2 (registered verifiers)** — the eleven; none for this family.

**Also load-bearing:** the file's own doc comment records that an absent
item_type abstaining is a DELIBERATE, reviewed decision ("the handler changed
nothing is a legitimate SUCCESS for other handlers"). So candidate 1
(fail-closed for repair-shaped types) is not just architecture-scope — it
revisits a recorded ruling, and MUST go the RFC route, not a bug patch.

**Fix candidates, re-ordered by this scoping:**
1. **(now the working candidate) Gate-2 artefact verifiers for the design-repair
   family** — measure the named defect at the served/stored artefact, the same
   discipline as the existing eleven. Works regardless of result shape, which is
   the property 1b lacks. Per-type semantics must be read first: the four
   evidence types differ (`needs_design_review` is arguably a review, not a
   repair — what completion MEANS per type needs stating before a verifier can
   grade it). `section_edit` is EXPLICITLY OUT of this file's scope — it is a
   fleet-wide content type owned by other campaigns; a zero-change run may be a
   legitimate success there, and the noChangeGates design demands a per-type
   measured justification we do not have for it.
2. **(RFC question, not a patch) unknownShape-blocks-for-opted-in-types and/or
   missing-verifier-refuses-for-repair-shaped-types** — both change what the
   shared gate guarantees; both touch a council-reviewed design. File as an RFC
   with this bug as the motivating case.
3. Registering 1b opt-ins alone: **ruled out** (defeated by blob results, above).

---

## MEASUREMENT PASS 2026-08-18, later the same day — by the fixing thread (`bugfix_302_design_repair_verification`)

Contributed into this file rather than into a parallel account. **The defect is real and I am
fixing it — but three of the statements above do not survive measurement, and one of them is the
recommended fix.** Everything below names the query or the file it came from.

### 1. The registry holds THIRTEEN item types, not eleven

The list above (and my own first grep) counted `RegisterVerifier(` and missed
**`RegisterVerifierWithPolicy(`**:

```bash
grep -rn "RegisterVerifier(\|RegisterVerifierWithPolicy(" platform/ --include=*.go | grep -v _test.go | grep -v 'func Register'
```

Also registered: **`hardcoded_section_colors`** (with a `Grades` remit test) and
**`needs_brand_head_assets`**. So *"No design-repair item type is registered"* is too strong — the
design **discovery** aggregate has a verifier; it is the design **audit** family
(`dark_section_audit`, `needs_design_review`, `spacing_fix`, `responsive_fix`) that has none.

### 2. The eleven unreadable payloads are mostly `bugs_closed/287`, which was fixed and rolled the day this was filed

§"The evidence this explains" attributes the population to handlers returning analysis blobs. The
actual shapes of the 11 `NO_CHANGE_GATE_UNREADABLE_RESULT` rows (`agent_error_log`, note the column
is `occurred_at`):

| result top-level keys | rows | what it is |
|---|---|---|
| `agent_id,agent_type,role,topics` | **7** | a **spawn record** — `bugs_closed/287` |
| `color_scheme,design_notes,spacing,typography` | 3 | the design-token blob described above |
| `add_to_page,approach,new_page,…` | 1 | an unrelated child-page triage decision |

`bugs_closed/287` is fixed, live and proven on `v1.0.1307` (rolled **2026-08-17 17:05Z**). All 11
abstentions **predate that roll** (latest 12:44Z). Fleet-wide, split at the roll:

| era | completions | spawn-record shaped |
|---|---|---|
| 08-14 → roll | 2,694 | **939** |
| after the roll | **1,880** | **0** |

939 → 0 against 1,880 completions of demand, and the 67 post-roll completions with no handler
envelope are **all legitimate non-handler closes** (47 retraction, ~10 revalidation, 4 owner
decisions, 2 bookkeeping) — not one is a malformed handler reply.

**So the hole is LATENT, not leaking.** It is real by construction and worth closing; it has no
current traffic. `dark_section_audit` has had **zero** rows touched since the roll (against 1,862
fleet completions in the same window), so no post-fix rate will be measurable without an induced
case. Argue the fix as a door being closed, not a leak being stemmed — and note the items are still
being FILED (7 on 08-14, 5 on 08-15, 2 on 08-17), so this is not a dead type.

### 3. The re-ordered "working candidate" (gate-2 verifiers for the family) is already decided AGAINST, on the record

`discovery_checks/verifier_coverage_test.go`'s `itemTypesWithoutVerifiers` — the guard that exists
precisely so these gaps are decisions rather than accidents — already classifies this family, with
reasons:

- `dark_section_audit` → `catMechanical`, *"verification needs a browser — pass condition is
  `spec.acceptance_test` free text over computed styles… candidate verifier is `criteria_check`
  (RFC_002) over `acceptance_test`"*;
- `contrast_failure` → *"**Deliberately NOT a verifier candidate** — verification needs a browser,
  i.e. an outbound probe on the completion path, the same standing objection as `image_url_404`,
  `backend_entry_orphaned` and `asset_reference_404`"*;
- `needs_design_review`, `spacing_fix`, `responsive_fix` → **`catJudgement`**, *"an LLM design
  opinion; nothing to re-run"*.

And the producer split (mandatory before registering any verifier — `LANDMINES.md`, archive-inclusive
so it is a lifetime count and not a 7-day one):

| item_type | producers | which | rows (live+archive) |
|---|---|---|---|
| `needs_design_review` | **4** | brief-fidelity-audit, design-audit, visual-design-audit, `<none>` | 1,296 |
| `responsive_fix` | 3 | + tool-acceptance-tier4 | 341 |
| `hardcoded_section_colors` | 2 | design-audit, `<none>` | 564 |
| `spacing_fix` | 2 | design-audit, visual-design-audit | 450 |
| `dark_section_audit` | 2 | design-audit, visual-design-audit | 30 |

A single verifier over a 4-producer population is `bugs_closed/213`'s defect exactly. So candidate 1
is not merely expensive — **it contradicts a reasoned classification in the guard built to keep this
on the record, and it needs a `Grades` remit test per type on top.** Whoever still wants it owes an
argument against a specific recorded reason, not a fresh start.

### 4. What I am fixing instead, and the honest price

The **gate 1b unreadable-payload arm**, which is where the two gates contradict each other. Gate 2
fails CLOSED when it cannot run (RFC_017, owner ruling 08-08) and its code refuses to exempt even an
unparseable spec because that "would leave a second silent completion path behind the one RFC_017
closed". Gate 1b — written five days *after* that ruling — is exactly such a path: an opted-in type
whose payload cannot be read is silently exempted from **an assertion it made about itself**, for
ever, and every future repair type inherits the exemption by default.

⚠ **The price, stated rather than glossed:** a refused row burns `max_attempts` rebuilds and lands
in `failed` for human review. **It is NOT released by WII-018's silence retraction** — that
mechanism is deployed but has **never run** (zero rows carry `result.retraction`; the design audit's
carrier `site-discovery-rotation-design` has been `enabled=false` since 08-11). The
refuse→attempts→retraction valve IS proven live once, on `empty_section` (`8ab3a32b`: gate errored
08-09, detector retracted it 08-14) — the architecture works, it is just not switched on for this
producer. That is the same cost the owner knowingly accepted for RFC_017.

⚠ **Not touched, deliberately:** `handlerReportedFailure`'s unknown-verdict arm, which also completes
on an unreadable input. Its header's measurement licenses that — 2,905 completed items carried no
`response.status` at all — and inverting it would block nearly every completion on the fleet. The
distinguishing property is that gate 1b's roster carries a **per-type assertion with a measurement
attached**; that arm does not.

**Plan and scope argument (RFC vs council gate):**
`docs024_key_docs_latest/bugfix_302_design_repair_verification/` — `NOTES` carries every figure above
with its query, `RUNBOOK` the six queries and their gotchas, `README_where_we_are` the plain-prose
account. `WRONG_CALLS.md` 2026-08-18 carries my own wrong call from this pass (I called WII-018
"live" from the code being merged).

---

## FIXED 2026-08-18 (one arm of it) — and the three follow-ups it does NOT cover

**Commit `743bc1945`.** Gate 1b's unreadable-payload arm is now a **per-type declaration** whose
zero value is not a policy, and `dark_section_audit` declares **refuse**. `_verification.status`
gains a fifth value `handler_result_unreadable` with its own operator message and reason code.
Registered in the same commit (WII-017 amended; its abstain LANDMINE corrected in place and its
"wiring proven only behaviourally" gap CLOSED by a new sqlmock wiring test). Council
**APPROVED round 1** — corr `edfef8cc-c42f-45f8-9b36-7578ffb56f6c`, *"approved with 2 advisory
objection(s) — none high-severity"*, 10 reviewers, 7 abstained, `unreadable: 0`. **All four
objections acted on, not filed** (`24235e990`) — including `editquality`'s catch of a FALSE CLAIM in
my own submission (I asserted WII-011's landmine was amended in the same commit; I had amended
WII-017, which merely mentions it — now fixed and logged in `WRONG_CALLS.md`), and `guardian`'s
by-query blast radius, which surfaced a second-hand effect nobody had enumerated: the live
`detected-item-promoter`'s `floor_ok` door-closer needs a pair ≥25% good, and this pair reads
**26 complete / 4 failed = 86.7%** today — an artefact of exactly the false greens this fix removes.
**75 further refusals** would cross the floor (~16 days if every filing were refused), and when it
does that is the correct outcome. `architecture` returned `point_fix` with a watch item, now in the
register: **at a SECOND type declaring `unreadableRefuses` this stops being a point fix and the
accumulated surface is worth its own RFC.**
Mutation-proven M1–M6, exit-status gated, unmutated control, both files restored byte-identical
(matrix in the lane NOTES; M4 is the instructive one — deleting the fifth reason arm does not
error, it falls through to `verification_failed`, a finding no gate made).

> ### ✅ ROLLED 2026-08-18 18:00Z on `agent-chassis` `v1.0.1310` — PRESENT on both replicas, BEHAVIOURALLY UNPROVEN
> Proven at the artefact, not inferred from the roll: image label revision `0b185bad2`, two-sided
> ancestry (`743bc1945` **in**, HEAD `01770302d` **not in**, so the check discriminates), running
> `imageID` digest `sha256:9ca35bac…` identical on both pods and matching the tag inspected, and a
> per-replica binary probe returning `handler_result_unreadable` **present**, long-live control
> `NO_CHANGE_GATE_UNREADABLE_RESULT` **present**, two nonsense needles **absent**. The
> `build provenance` log line had already scrolled on 38-minute-old pods — "could not look", not
> "unstamped".
> ⚠ **[MEASURED 18:38Z] 0 rows of this type touched since the roll, 0 carrying the new status, and 0
> fresh abstain records — but that last zero is VACUOUS**: this file's own plan nominated it as the
> "the refusal is not wired" control, and with no demand it could not have come out otherwise. Fleet
> completions in the window: 18. **Status: deployed, not behaviourally proven**, carried by the
> wiring test rather than a live row.

⚠ **NOT proven live, and it cannot be without manufactured demand.** Zero `dark_section_audit`
rows touched since the `v1.0.1307` roll (1,862 fleet completions in the same window) and both
carriers that dispatch the type are `enabled=false`. After a roll the honest status is **"deployed,
not behaviourally proven"**. Plan, phasing and the scope argument:
`docs024_key_docs_latest/bugfix_302_design_repair_verification/PLAN_2026-08-18_*`.

### Follow-up 1 — the claimed-item-timeout sweep bypasses gate 1b entirely, and its lockstep cannot see the roster

[VERIFIED in the live `pre_query`, not inferred] `scheduled_tasks.claimed-item-timeout`
auto-completes a `claimed` item on generic orchestration evidence, and its own comment says the
exclusion list is *"the LOCKSTEP TWIN of the `RegisterVerifier()` calls"* — i.e. it keys on **gate 2
only**. A type with a gate-1b roster entry and no registered verifier — which is exactly
`dark_section_audit` — is therefore **not excluded**, so the sweep can complete it 15 minutes after
a claim without either gate running.

[MEASURED, archive-inclusive] it has **never happened**: 0 of 30 `dark_section_audit` and 0 of 564
`hardcoded_section_colors` rows carry the sweep's `completed_by_step` shape. So this is **latent**,
and deliberately left out of `743bc1945`: widening the lockstep contract from "has a verifier" to
"has a verifier OR a roster entry" changes a shared scheduler predicate AND the parity test that
enforces it both ways (`TestRegisteredVerifiersMatchClaimTimeoutExclusion`), which is its own
coherent task with its own blast radius. **Do not "just add the type" to `220`** — the parity test
enforces excluded ⇔ verifier in both directions and will fail.

### Follow-up 2 — three design types are NOT covered by this fix, and each needs a decision first

The four finetuning.uk evidence rows this file was filed on are **not all one type**:
`needs_design_review` ×2, `spacing_fix` ×1, `dark_section_audit` ×1. Only the last is covered.

- **`needs_design_review`** — needs a *semantics ruling* before any gate: an analysis blob may
  legitimately BE the deliverable for a review type, in which case "the handler changed nothing" is
  a success and a no-change rule would be wrong. It is also the worst producer population in the
  estate for this (4 producers, 1,296 lifetime rows).
- **`spacing_fix` / `responsive_fix`** — need their handlers' success-envelope shape MEASURED before
  `CounterPaths` and `OnUnreadable` can be declared honestly. The roster's bar is a measurement, not
  a guess about somebody else's handler.

### Follow-up 3 — the operational gap, which is probably not this lane's

The refuse→attempts→**retraction** release valve is proven live once (`empty_section` `8ab3a32b`:
gate errored 08-09, the detector retracted it 08-14) but is unavailable to this family because the
design audit's carrier has been `enabled=false` since 08-11 and WII-018 has never run. Re-enabling a
carrier is a cost decision, and the rotation work is `bugs_open/230`'s. **This file should not be
read as claiming refused rows get tidied up.**

### Status of this bug

**STAYS OPEN.** One arm fixed and unproven-live; three follow-ups above; and the filing's headline
question — should the design-audit family have artefact verifiers at all — is answered "not by this
route" with the recorded reasons, which is a decision on the record rather than a fix.
