# PLAN — the completion-time consumer for `acceptance_predicate` (`bugs_open/395` candidate 1)

**Opened 2026-08-25** by the `vigilant_designer_offer_analysis` lane, continuing
`HANDOFF_2026-08-25_continue_here.md` §1.
**Scope:** ONE slice — make the platform READ, at completion time, the machine-checkable predicate
`CLM-024` already writes. Nothing about v2(a)/(b)/(c) (§3 of the handoff) is in this plan.

## 1. What is being built, in plain terms

A **work item** is one recorded defect on one site. A **handler** is the agent sent to fix it. When
the handler reports success, the platform stamps the item `complete`.

Since 2026-08-24 some items carry a small machine-checkable condition in their own spec — an
**acceptance predicate** — which the producer wrote alongside its prose acceptance test, and which
was PROVEN FALSE of the page at the moment the finding was made. Nothing reads it afterwards.

This slice adds **completion gate 1c**: before an item is stamped `complete`, if its type has opted
in, re-evaluate the item's own predicate against the page as it stands now. A predicate that still
refutes means the handler did something other than what the finding asked for.

## 2. The decisions, and why

### D1 — it is a GATE beside `handlerReportedNoChange`, not a registered verifier

Two independent reasons, and only the second is the one the handoff carried:

1. `GetVerifier` is **one verifier per item_type**, a scarce shared slot. `content_rewrite` is filed
   by many producers; occupying its slot with a check that only speaks to items carrying a predicate
   would either grade the wrong rows or have to be disclaimed back out through `policy.Grades`.
2. The gates in `complete_work_item_verification.go` are **composable and opt-in**; gate 1c is a
   third one in the same shape, running between 1b and 2.

⚠ The handoff's stated reason — *"`VerifyTarget` carries the SPEC, not the RESULT"* — is the argument
for **gate 1b**, and it does NOT transfer to this gate: gate 1c needs the SPEC (where the predicate
lives) and the CURRENT page row, both of which a verifier has. Recorded as a correction rather than
repeated, because inheriting it would have made the design look forced.

### D2 — it runs BETWEEN gate 1b and gate 2

Gate 1c grades the item's OWN stated criterion; gate 2 grades the TYPE's generic predicate. If the
item's own criterion is false, producing a generic `verified` that contradicts the item's own terms
helps nobody, so 1c is asked first. It is also the cheaper of the two for most types (one page-metadata
read, no HTTP, no browser).

### D3 — opt-in per `item_type`, with a THREE-VALUED outcome (owner ruling 2026-08-02 §2)

Mirrors `unreadableOutcome` exactly, for the same reason: the zero value must not be a policy.

| outcome | meaning |
|---|---|
| `predicateUndeclared` | zero value, **not** a policy; the roster test refuses it, and at runtime it records |
| `predicateRecords` | the verdict is stamped on the row; the completion proceeds |
| `predicateRefuses` | a still-refuting predicate BLOCKS the completion |

### D4 — cut 1 arms `content_rewrite` at `predicateRecords`, not `predicateRefuses`

This is the decision most open to challenge, so the reasons are stated in full:

1. **There is no negative control.** All three live predicates refute; no row exists where a predicate
   is satisfied after its fix (`bugs_open/395` §6). A blocking gate that has only ever seen failures
   cannot be distinguished from one that refuses everything.
2. **Blocking demands a second live change on somebody else's object.** The claimed-item-timeout sweep
   writes `site_work_items` directly and NEITHER completion gate runs for it; its only protection is
   `livespec.ClaimedItemTimeoutExclusions`, and adding `content_rewrite` there needs a migration
   amending a live shared `pre_query`. `content_rewrite` is **1,637 completions all-history
   [MEASURED 2026-08-25]** — not a type to widen a sweep's contract around as a side effect.
3. Recording produces the census `395` §5 says is the first job, on a queryable surface.

**The cost of this choice, named rather than buried:** it is a third instance of CLM-023's residual —
an arm that has never fired in production. Three things stop it becoming permanent:

- the block arm is written and mutation-tested;
- the promotion precondition is stated **in the roster entry**, in the `LicenceVoided` shape;
- **promotion is a BUILD FAILURE unless the exclusion is added too** —
  `TestClaimTimeoutExclusionCoversBothCompletionGates` is extended so a `predicateRefuses` entry
  must appear in `livespec.ClaimedItemTimeoutExclusions`. A `predicateRecords` entry is deliberately
  NOT added to that contract: it blocks nothing, so an exclusion it did not earn would trip the
  reverse direction (`bugs_open/006` §C churn).

### D5 — EVERY evaluated predicate leaves a verdict on the row, including `holds`

Not decoration, and this is the instrument the whole slice rests on: without a recorded `holds`, a
gate that permits is indistinguishable from a gate that never ran — which is the exact shape of the
residual in D4. Recording `holds` is what turns "the gate has only ever seen failures" from a worry
into a query.

### D6 — `inapplicable` records LOUDLY and never blocks

A stored predicate was evaluable at emission by construction. If it is not evaluable now, the page
left the surface, or the vocabulary moved, or **the gate has gone blind** — and that last one wears
the same face as the other two. This is the objection two council seats raised against the emit side
(corr `ef482d1c`, `editquality` + `debug_historian`), and the answer is the same: it goes to
`agent_error_log` as well as the row, and when the subject set is EMPTY the reason blames this gate
rather than the named page.

It does not block, because none of its three causes is evidence the handler failed, and refusing
would spend an attempt on something the handler cannot influence. ⚠ That is a **stated asymmetry with
RFC_017** ("I could not check" is not "I checked and it is fixed") and it is only defensible while the
outcome is `predicateRecords` — a type promoted to `predicateRefuses` must revisit it.

## 3. ⚠ THE TRAP THAT WOULD HAVE MADE THE WHOLE GATE SILENTLY INERT

**A STORED predicate cannot be fed to `EvaluateAcceptancePredicate`.** The evaluator enforces a CLOSED
KEY SET per type (`acceptancePredicateFields`), and the emit side STAMPS two provenance keys —
`verdict_at_emission`, `evidence_at_emission` — onto the predicate *after* evaluating it. So the
stored form carries two keys the evaluator rejects, and a naive consumer gets
`PredicateInapplicable` on **100% of live predicates**, with a reason that reads like a vocabulary
problem in the model's output.

Nothing catches this today: `TestTheFirstLiveEmittedPredicatesStillRefuteAfterTheFix` hand-writes the
predicates WITHOUT the provenance keys, so the only test over real live data uses a shape that does
not exist in the database.

**The fix is not to widen the key set** — the closed set exists so that "a key no checker reads is an
assertion the artefact appears to make and does not", and widening it would let a model write its own
`verdict_at_emission`. The fix is to single-source the stamp and the strip in the file that owns them,
and to pin them to each other with a test that fails when a THIRD provenance key is added:

    storedPredicate(pred, reason)  →  predicateForEvaluation(stored)  ==  pred

Goes in `LANDMINES.md`.

## 4. What is NOT in this slice

- Promotion to `predicateRefuses` (needs the negative control — D4).
- Anything for `update_work_item_status`, the SECOND writer of `complete`. **Re-measured 2026-08-25**:
  4 active agents reach `complete` through it across 6 arms (image-build-handler,
  image-source-unsatisfiable-handler, image-url-404-handler, required-fields-missing-handler ×3) and
  **none of them handles `content_rewrite`**, so gate 1c covers every completion of the armed type.
  That is `bugs_open/375`'s lane, not this one.
- The claimed-item-timeout sweep hole. It is real, it is pre-existing, and it is left OPEN and named,
  because at `predicateRecords` there is nothing for it to bypass.

## 5. How it will be verified

- **Positive:** a unit fed the three REAL live predicates in their STORED form (provenance keys and
  all) plus the served strings, asserting `refutes` — the case §3 says nothing tests today.
- **Negative control, at the gate rather than in the wild:** the same live predicate against a
  meta description edited to satisfy it must return `holds` and permit completion. This is the control
  `bugs_open/395` §6 says is not optional; it is available at the gate today and NOT available as a
  live row, and the plan says which is which rather than letting the unit stand in for the run.
- **Mutation:** delete the roster lookup → an unopted type must start being graded (test fails);
  delete the strip → every live predicate must go inapplicable (test fails).
- **At the artefact, after the roll:** re-fire the offer analyser and read
  `result->'_verification'->'acceptance_predicate'` on the completions.
