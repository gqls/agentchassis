# 160 — the report prose gate reads a recombined engineering rating as an invented model number, and fail-closed kills the whole report

**Filed** 2026-07-31 from the gripper dossier lane, after a fixture regeneration the owner
had asked for produced **no page at all**.

**Status: OPEN, unowned.** The gate is behaving as designed; the *classifier inside it* is
wrong. Nothing is mis-served — the failure is that a legitimate report is destroyed rather
than published, silently as far as any dashboard is concerned.

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

## Context, so nobody re-derives it

This was found while regenerating a fixture to confirm an unrelated, already-approved chart
fix (`f8e7c31ce`, council `60d05267`). That fix is **live and verified on the pod** — it is
not implicated here, and the chart is rendered at `compose_page`, downstream of the step
that failed. Lane notes: `docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/`.
