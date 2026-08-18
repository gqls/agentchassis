# PLAN — `bugs_open/302`, design-repair completion verification

Design, phasing, decisions **and their reasons**. Corrections to the originating brief live here,
marked as corrections.

---

## The brief, and what it turned out to be

**As filed:** design-repair item types have no registered verifier, so `verifyBeforeComplete`
abstains and no-op "repairs" complete unverified. Its own same-day scoping update re-ordered the
candidates to make **gate-2 artefact verifiers for the design-repair family** the working one.

**As measured** (every figure in `NOTES`, every query in `RUNBOOK`): the defect is real but it is
not where the filing points, and the recommended fix is one the estate has already declined.

> **CORRECTION 1 to the brief — the registry is 13 types, not 11.** Both the filing and my own
> first grep counted `RegisterVerifier(` and missed `RegisterVerifierWithPolicy(`, which is how
> `hardcoded_section_colors` (with a `Grades` remit test) and `needs_brand_head_assets` are
> registered. So "no design-repair item type is registered" is too strong: the design **audit**
> family has none; the design **discovery** aggregate has one.

> **CORRECTION 2 — 7 of the 11 unreadable payloads were `bugs_closed/287`, fixed and rolled the
> same day this was filed.** The filing attributes the population to handlers returning analysis
> blobs. Actual split: 7 spawn records, 3 design-token blobs, 1 foreign triage decision. 287 went
> live on `v1.0.1307` at 2026-08-17 17:05Z; the spawn-record shape then goes **939 → 0** across
> 1,880 fleet completions. **The hole is latent, not leaking.**

> **CORRECTION 3 — the working candidate is already decided against, on the record.**
> `verifier_coverage_test.go`'s `itemTypesWithoutVerifiers` exists precisely so these gaps are
> decisions rather than accidents, and it classifies this whole family with reasons: a browser on
> the completion path (the standing objection recorded in three places) for
> `dark_section_audit`/`contrast_failure`, and `catJudgement` — "an LLM design opinion; nothing to
> re-run" — for `needs_design_review`/`spacing_fix`/`responsive_fix`. The mandated producer split
> adds that `needs_design_review` has **4 producers over 1,296 rows**, so one verifier there is
> `bugs_closed/213`'s defect exactly.

## The decision, and why this one

**Fix the gate-1b unreadable-payload arm — the place where the two completion gates contradict
each other — and do not build gate-2 verifiers for the family.**

The reason is not that gate-2 verifiers are expensive. It is that **an owner ruling already settled
this exact question for the sibling gate**, and applying a decided principle is a smaller and
better-founded change than inventing a policy:

| | gate 2 (verifiers) | gate 1b (no-change) |
|---|---|---|
| cannot run / cannot read | **fails CLOSED** (RFC_017, owner ruling 2026-08-08) | **completed** |
| unparseable input | fails closed, and its comment says exempting it "would leave a second silent completion path behind the one RFC_017 closed" | was that path |

Gate 1b was written on 2026-08-13 — **five days after** the ruling.

**What made it worth doing despite the latency**, and the argument that carries the submission: the
roster is opt-in, and an entry is an assertion *with a measurement attached* that for that type a
zero-change run cannot be a repair. An unreadable payload silently waived that assertion. And
[MEASURED] the waiver was not passive — **5 of the 11 abstentions completed items this gate had
already BLOCKED one attempt earlier** (`complete`, `attempt_count` 1, block sentence still in
`.error`). The arm reversed the gate's own refusal.

## Candidates, ordered by what makes the bad state unrepresentable

1. **CHOSEN — per-type declaration on `noChangeRule`, refuse for `dark_section_audit`, roster test
   forces every future entry to declare.** Unrepresentable: an entry whose unreadable semantics are
   unstated cannot ship (CI), and the zero value cannot block (runtime). Blast radius is exactly the
   opt-in roster — one type. Works regardless of payload shape, which is the property gate 1b
   lacked and the reason opting more types into the *counter* rule would have changed nothing.
2. **Complement, filed as a follow-up, deliberately NOT folded in** — the claimed-item-timeout sweep
   auto-completes a `claimed` item on orchestration evidence, bypassing gate 1b entirely, and its
   exclusion list is locked to the gate-2 registry by a parity test (`sql_for_agents/220` ↔
   `TestRegisteredVerifiersMatchClaimTimeoutExclusion`). Widening that contract to "verifier OR
   roster entry" is a second seam change and its own task. Measured: 0 occurrences for this family.
3. **REJECTED — gate-2 verifiers for the design family** (the filing's candidate). See CORRECTION 3.
   Anyone reviving it owes an argument against a *specific recorded reason*, plus a `Grades` remit
   test per type.
4. **NOT THIS FIX — "missing verifier refuses for repair-shaped types"** (the filing's candidate 1).
   Broadest on paper, but it revisits a deliberately recorded decision (so it needs the RFC route, as
   302's own scoping says) and "repair-shaped by name" is the producer-list anti-pattern WII-013
   refuted. Left as the framed RFC question.
5. **REJECTED — rely on the audit re-filing.** 302's own "not a fix"; costs an audit cycle per miss
   and has already failed once (the 08-09 finding re-filed 08-12).

**Deliberately NOT touched:** `handlerReportedFailure`'s unknown-verdict arm, which also completes on
an input it cannot read. Its own measurement licenses that — **2,905 completed items carried no
`response.status` at all** — so inverting it would block nearly every completion on the fleet. The
distinguishing property is the per-type assertion, which that arm has no equivalent of.

## Scope: council gate, not an RFC — argued, not assumed

- **2026-07-29 ruling 1 (the guarantee test).** Byte-identical for every type absent from the roster
  (a map miss; mutation M6 is the containment control). For roster types the guarantee was always
  "per-type declared judgement with a measured licence" — this adds a declared axis and changes
  behaviour for the one entry edited, in the same commit, with its measurement. **No fleet-wide
  default flips**, which is what made RFC_017 itself architecture-scope.
- **2026-08-02 RFC_010 narrowing 2.** Complied with by construction: opt-in field, unsafe default OFF.
- **2026-08-11 RFC_022 narrowing.** (1) opt-in ✓ (2) unsafe-default-OFF ✓ (3) **fails** — a live
  consumer names the field in the same commit. So that safe harbour is **not claimed**; losing it
  returns the question to ruling 1. Precedent: WII-013's `Grades` shipped with a live consumer naming
  it and went through this gate; WII-017 — this roster's first entry, a new blocking outcome on the
  same shared path — went through this gate and was APPROVED at round 2.
- **Which side is unsafe:** REFUSING (new blocking authority on a shared path whose measured failure
  mode was shape drift). ABSTAINING is unsafe the other way for a repair type — WII-001's false green.
  Neither may be an author-time silent default, which is what the tri-state buys.

## Phasing, and the honest state of each phase

| phase | state |
|---|---|
| Measure, and correct the filing | **DONE** — three corrections contributed into `bugs_open/302` itself |
| Design, adversarially reviewed | **DONE** — fable endorsed the candidate and supplied the five-reversal finding, which I re-verified before using |
| Build + mutation-prove + register | **DONE** — `743bc1945`; M1–M6 all RED, unmutated control green, files restored byte-identical |
| Council gate | **SUBMITTED** — corr `edfef8cc-c42f-45f8-9b36-7578ffb56f6c`, `Council-Submitted:` trailer on the commit |
| Roll + prove live | **BLOCKED ON DEMAND, and stated as such** — see below |
| Follow-ups filed | the sweep bypass; `spacing_fix`/`responsive_fix`/`needs_design_review` semantics |

**The verification problem, named rather than glossed.** [MEASURED] zero `dark_section_audit` rows
touched since the roll, against 1,862 fleet completions; both carriers that dispatch the type are
`enabled=false`. So the refusal **cannot fire on its own**, exactly as WII-017 and WII-018 could not.
Presence is provable at the binary with a three-needle scan (the new literal, a long-live control, a
nonsense needle) on **both** replicas; behaviour is not, without manufacturing demand via a one-shot
design-discovery envelope — which is an owner-cost decision, not a verification convenience. Until
then the honest status is **"deployed, not behaviourally proven"**, carried by the wiring test (M5)
rather than by a live row, and `302` stays open with a dated note.

⚠ **And the cost of a refusal is not tidied up.** WII-018's silence retraction has never run, so a
refused row burns `max_attempts` and waits at `failed` for a human. That is RFC_017's accepted cost.
The valve itself is proven once (`empty_section` `8ab3a32b`), so the architecture works — it is
switched off for this producer, which is an operational gap plausibly belonging to
`bugs_open/230`'s rotation work rather than to this lane.
