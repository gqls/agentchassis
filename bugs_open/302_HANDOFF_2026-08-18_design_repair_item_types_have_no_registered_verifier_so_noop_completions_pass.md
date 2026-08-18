# 302 — design-repair item types have no registered verifier, so a no-op "repair" completes unverified

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
