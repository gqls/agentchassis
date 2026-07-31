# 160 — the report prose gate reads a recombined engineering rating as an invented model number, and fail-closed kills the whole report

**Filed** 2026-07-31 from the gripper dossier lane, after a fixture regeneration the owner
had asked for produced **no page at all**.

**Status: CLOSED 2026-07-31 ~21:10 — fixed, council-APPROVED (round 2), LIVE on chassis
v1.0.1222 and pod-verified on both replicas.** The gate was behaving as designed; the
*classifier inside it* was wrong. Nothing was mis-served — the failure was that a legitimate
report was destroyed rather than published, silently as far as any dashboard is concerned.

> **FIX RECORD — lane `bugfix_160_prose_gate_recombination`.** Council
> `926a7bea-ccb0-4e32-a410-e9e7cdbc3256`: REVISE at round 1 (compliance, HIGH), APPROVED at
> round 2. Commits `c5286fee4` (fix + tests), `377c3cb19` (round-2 revision), `e9f3d9483`
> (round-2 advisory + the real-run induction). Fix candidate 1 was taken but **not as
> written**, and round 1's own answer was wrong too — see below.
>
> **Proven against the real incident, not only fixtures.** The destroyed run's actual
> `report_prose` and `scoring` were pulled from `collected_data` and passed through the gate:
> with route 3 disabled it reproduces `summary_html names model-like token "IP54-or-better"
> not in the candidate set or fact block` — byte-identical to the live `__step_error` — and
> with the fix in place it returns **0 violations**, the summary still containing the phrase.
>
> **A fresh end-to-end `report_request` was deliberately NOT fired**, and this is a judgement,
> not an omission. It cannot demonstrate the fix: the trigger is the writer's phrasing, which
> varies per run — this very spec passed the gate on 2026-07-27 — so a green run proves
> nothing and a red one would be a different bug. It would also publish a live report page on
> robot-hands.com that this session cannot remove (the gripper lane needed an owner-run
> command to delete the last set). The real-prose induction above is strictly stronger
> evidence for this defect than a fresh generation would be.

---

## Symptom

A `report_request` with a valid spec fails at `verify_prose`, and because that step is
correctly **fail-closed**, no page is ever composed. The work item ends `failed`, and the
expected URL 404s.

```
step verify_prose failed: failed to execute action verify_report_prose:
report prose failed verification (1 violations):
summary_html names model-like token "IP54-or-better" not in the candidate set or fact block
```

Observed: work item `4ccc73d7-c467-480f-9a39-0b327b383870`
(`request_id bf3765d6-befe-43a8-b1cd-ca5c210f39e9`), 2026-07-31 08:20Z, chassis
`v1.0.1213`. `https://robot-hands.com/reports/bf3765d6-…html` → **HTTP 404**.

## Root cause — read from the code, not inferred

`platform/orchestration/actions/verify_report_prose_action.go:241`

```go
var modelNumberRe = regexp.MustCompile(`\b[A-Za-z0-9]*(?:[A-Za-z][0-9]|[0-9][A-Za-z])[A-Za-z0-9]*-[A-Za-z0-9-]+\b`)
```

Any token containing a letter-digit adjacency followed by a hyphen and more word characters
is treated as SKU-shaped. `IP54-or-better` matches: `IP54` supplies the `[A-Za-z][0-9]`
adjacency (`P5`), and `-or-better` supplies the tail.

The token is then cleared only if it is **contained verbatim** in the fact block, or
overlaps a candidate name (`:315-327`):

```go
if strings.Contains(allowedText, tok) { continue }
```

`IP54` is in the fact block; **`IP54-or-better` is not**, because the writer composed the
phrase. So a real rating, correctly used, is reported as an invented part number.

**The gate itself is right and must stay strict.** Its comment states the intent — *"never a
sibling model invented by the writer"* — and that guard is the reason this pipeline can be
trusted at all. The defect is that the SKU classifier cannot distinguish *a fabricated
sibling SKU* from *a fact-block token recombined into an English phrase*.

## Blast radius — measured

- **The whole class is "hyphenated engineering notation", not one phrase.** The same spec's
  `mounting` field is `ISO 9409-1-50-4-M6 flange on a 6-axis arm`; `9409-1-50-4-M6` matches
  the same regex and survives only because it appears verbatim in the request context. Any
  writer paraphrase — `9409-1-50-4-M6-compatible`, `IP54-rated`, `M6-threaded` — trips it.
- **It is intermittent, which is worse than deterministic.** The *identical* spec passed this
  gate on 2026-07-27 (fixture 1, `d1a371be-…`, live and 200). The trigger is the writer's
  phrasing, which varies per run — so this will read as a flaky pipeline rather than as a
  rule.
- **Only the reports pipeline uses this gate today** (`verify_report_prose`), so nothing
  outside it is affected — but the reports pipeline has no other route to a page.

## Fix candidates — ordered by what makes the bad state unrepresentable

**None of these should be applied without design review; the trade-off is real and this is
the file where it should be argued, not in a commit message.**

1. **Test hyphen-segment provenance, not the whole token.** Clear the token when *every*
   hyphen-separated segment is either present in `allowedText` or is digit-free English
   (`or`, `better`, `rated`, `compatible`). `IP54-or-better` → `IP54` ✓ in fact block, `or`
   / `better` digit-free → allowed.
   **The hole to close first:** a *prefix* rule alone would clear an invented sibling
   (`EGP-50-X` where `EGP 40-N-S-B` is a real candidate), which is precisely what the guard
   exists to stop. Segment-wise is stronger than prefix-wise, but somebody must check
   whether a fabricated sibling can be assembled entirely from allowed segments.
2. **Normalise both sides before comparison** — strip hyphens and case from token and
   `allowedText`, so `IP54-or-better` → `ip54orbetter` and the fact block's `IP54` is found
   as a substring of it (note: *token contains fact*, the reverse of the current test).
   Cheap and mechanical; weakens the guard by exactly the amount that substring-matching
   always does.
3. **Instruct the writer not to hyphenate ratings.** Rejected: a prompt is not an
   enforcement mechanism, and this gate exists because prompts are not enforcement.
4. **Downgrade the violation to a warning.** Rejected outright — it reopens fabricated model
   names, which is the defect class the whole report pipeline was built to prevent.

## How to verify a fix

1. **Unit, on the classifier, not the pipeline** — `IP54-or-better`, `IP54-rated`,
   `9409-1-50-4-M6-compatible` must clear with `IP54` / `9409-1-50-4-M6` in the fact block;
   an invented sibling of a real candidate must still be **rejected**. A fix asserted only
   on the first half is half a fix.
2. **Mutation-check it**: revert the change and confirm the new tests fail. And note
   `LANDMINES.md` (2026-07-31) — *a second guard can absorb your mutation*: the numeric
   check at `:307` may reject the same string for a different reason, so assert on the
   **violation text**, not merely on "there was a violation".
3. **End to end**, only after the unit half: re-arm a `manual-test` `report_request` and
   require the page to serve 200 **and** the summary to still contain the phrase — a report
   that passes by dropping the rating has been made worse.

## Reproduce

```sql
-- the failing run, with the cause in the column that actually holds it
SELECT collected_data->'__step_error' FROM orchestration_states
WHERE created_at > now() - interval '1 day'
  AND collected_data::text LIKE '%bf3765d6-befe-43a8-b1cd-ca5c210f39e9%';
```

The work item's own `error` column carries only the wrapper
(`gripper dossier build failed — see the step error`); the sentence naming the token is in
`collected_data->'__step_error'` of the **child** orchestration. See `016b` on reading
`__step_error` rather than `error`.

## What was built, and two corrections to this file (2026-07-31, fixing lane)

**The rule shipped:** a SKU-shaped token also clears when it splits into a **head** that traces
by the two routes that already existed (verbatim in the allowed text; overlapping a candidate
name) and a **tail** whose every hyphen segment is digit-free, ≥2 characters and lower-case.
`modelNumberRe` puts the letter-digit adjacency in segment 0, so the code-bearing part is
always inside the head — the change relaxes which *suffixes* are tolerated, never whether the
model number itself was published. Purely additive: nothing cleared before is affected, and the
numeric gate, vendor gate, no-match contract, truncation guard and fail-closed behaviour are
untouched.

> **CORRECTION 1 — fix candidate 1 as filed opens the hole it names.** *"…or is digit-free
> English (`or`, `better`, `rated`, `compatible`)"* clears `X` and `XL` too, so
> `2F-85-X` / `2F-85-XL` — invented siblings of the indexed `Robotiq 2F-85` — would pass. The
> three clauses above exist precisely to reject those; each is mutation-checked against its own
> counterexample.
>
> **CORRECTION 2 — `EGP-50-X` is not matched by `modelNumberRe` at all, so it could not have
> been the hole.** The regex needs the letter-digit adjacency in the segment **before the first
> hyphen**, and `EGP` has no digit. The same is true of the real candidate `Festo EHPS-20-A-LK`.
> Verified by running the classifier, not by reading it. A related surprise, same source: for a
> paraphrase of the ISO flange code the regex does not match the whole token, it extracts the
> **sub-token `M6-compatible`** — a fresh `\b` after each hyphen. Any counterexample used to
> test this gate must therefore start with a mixed letter-digit segment, or it tests nothing.

> **CORRECTION 3 — the shape rule above was ROUND 1, and the council was right to refuse it.**
> Admitting a tail segment because it is lower-case, digit-free and ≥2 characters is
> **orthographic, not semantic**: `not`, `instead`, `unless` and `lower` pass exactly the test
> that `or` and `better` do. `IP54-not-rated` would have cleared a gate whose whole purpose is
> to stop a report asserting what the facts do not say — the fabrication merely relocated from
> the model number to the qualifier (compliance seat, HIGH). `editquality` found the same hole
> from the numeral side (`2F-eighty-five`). **What shipped is a closed vocabulary**, compared
> case-insensitively, under three admission rules: connectives that assert nothing
> (and/of/or/to), strengtheners that ask for at least the stated value
> (above/better/greater/higher/min/minimum/over), and attributives that predicate the traced
> code and add no fact (rated, compliant, compatible, class, threaded, mounted, …). No
> negation, no inversion, no numeral word, and no word naming a family beyond the code
> (`series`/`style`/`type` were admitted at round 2 and removed after the round-2 advisory —
> `2F-85-series` asserts a product line that may not exist).
>
> My round-1 justification for *not* building the vocabulary — "no instance of either has been
> observed" — was the error worth recording. The instance could not have been observed,
> because the class did not exist until my own change created it. **A rule that admits by
> shape has to be argued against the space of strings it admits, not against the strings that
> have turned up.**

**Residual, stated not absorbed:** a legitimate qualifier absent from the vocabulary is still
rejected. The cost is one report, which the writer usually clears on retry with different
phrasing, and adding a word is a one-line change plus a test — deliberately preferred to a
rule that cannot tell a connective from a negation.

**Verification done:** unit, both halves, asserting on the violation **text** (5 recombinations
clear including title case, 8 fabrications still rejected — invented siblings, negation,
inversion, numeral word, family name); every vocabulary exclusion individually mutation-pinned
(admitting `not` fails `IP54-not-rated` alone, `lower` fails `IP54-or-lower` alone, `series`
fails `2F-85-series` alone); the real destroyed run induced both ways; both pods grepped for
`skuTokenTraces`/`qualifierWords` with a positive control in the same exec. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_160_prose_gate_recombination/`.

## Context, so nobody re-derives it

This was found while regenerating a fixture to confirm an unrelated, already-approved chart
fix (`f8e7c31ce`, council `60d05267`). That fix is **live and verified on the pod** — it is
not implicated here, and the chart is rendered at `compose_page`, downstream of the step
that failed. Lane notes: `docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/`.
