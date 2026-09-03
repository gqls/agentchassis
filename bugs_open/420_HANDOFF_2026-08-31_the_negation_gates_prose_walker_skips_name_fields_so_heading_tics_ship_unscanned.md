# 420 — the negation gate's prose walker skips `name` fields, so heading-surface tics ship unscanned

**Filed:** 2026-08-31, `copy_quality_two_stage` lane, from the v1.0.1349 post-roll canary
(finetuning.uk approach page, rebuild item `4641e02d`, orchestration `2eecc01d-…`).

**Diagnosis-loop note (owner ruling 2026-07-31):** filed on first-hand verification instead of a
090 run, and here is the substitution — every link in the chain is measured on one artefact, in
one run, with the walker's own instrumentation as witness: the gate's marker shows sibling
`description` fields of the SAME array targeted and repaired while the `name` fields never
entered `hits_before`; the exclusion regex names `name$` on a single greppable line; and the
served page carries both skipped strings verbatim. No inference bridges the steps. A 090 run
would re-read the same three artefacts.

## Symptom

After a full rebuild through the corrected stack (v1.0.1349: seven gate shapes, empty mild list,
ruling-7 repair), the served page still carries two define-by-negation constructions **on
heading surfaces**:

- "Privacy built into the deployment, not added afterwards"
- "Facts checked against sources, not asserted"

Both are `features[N].name` values (iter_2's stored result: `features[1].name`,
`features[6].name`), rendered as card headings.

## Evidence (all 2026-08-31, canary run 15:24–15:28Z)

- `process_sections_loop_iter_2_rewrite_negations` marker (orchestration `2eecc01d`,
  collected_data): `hits_before: 3`, `exempt: 0`, targets = `features[1].description`,
  `features[5].description`, `features[6].description`. **The two comma-not `name` strings in
  the same `features` array are absent from every list** — not exempted (exempt would be ≥1 and
  `exempt_reasons` populated), not rejected, never counted. The scanner never saw them.
- The same run proves the machinery works where the walker looks: across the page, 12
  replacements proposed, 10 spliced and serving, 2 refused with recorded reasons
  (`still_negative_reveal`, `gutted`) — so this is not a wiring failure.
- Battery on the served page (`count_negation_tells.py`): 3 × "X, not Y", of which 2 are these
  headings.

## Root cause (read, not inferred)

`platform/orchestration/datahelpers/negation_content.go:46`:

```go
var nonProseFieldRe = regexp.MustCompile(`(?i)(^|_)(url|…|target|rel|name)$`)
```

`name` is on the never-holds-prose list, so `WalkContentStrings` → `prosey()` returns false for
every `*.name` field and neither consumer (the renderer's annotation, the repair action) ever
scans one. The file's own comment names the cost — "A false NEGATIVE here costs a missed tell"
— and `IsHeadlineField`'s comment calls headline surfaces the place the construction is "least
forgivable". A feature card's `name` IS its heading: the field the exclusion list treats as an
identifier is the page's highlight surface. (The Gemini G1 trial read flagged exactly this
surface — "tic ON the highlight surface" — one model trial earlier, on the same page family.)

Every other entry in the list (`url`, `hex`, `price`, `email`…) is genuinely non-prose;
`name` is the only member that collides with a heading use. `title`/`heading`/`headline` are
scanned (and get headline severity).

## Fix candidates, ordered by what closes the door

1. **Drop `name` from `nonProseFieldRe` and let `prosey()`'s value tests decide** (≥12 chars,
   contains whitespace, fails `nonProseValueRe`). A prose-bearing name is then structurally
   scannable — the bad state becomes unrepresentable. Cost: person/product names like
   "Farm Shield" mostly fail the length/space tests anyway; one that passes gets scanned, which
   is harmless — scanning only flags negation shapes, and every splice still passes the claim
   and structure guards. One-line change + walker test.
2. Scan `name` but only when `IsHeadlineField`-style sibling heuristics say the object is a
   card/feature (has `description`/`icon` siblings). Closes less: a new classifier whose gaps
   the guarantee inherits (see memory: a guarantee conditional on a classifier).
3. Prompt-side ban ("never negate in a heading"). Closes nothing — the gate exists because
   prompt rules leak.

Candidate 1 should also add `name` to the headline-severity question: a card name is a heading,
so a hit there should carry `headline: true` weighting (extend `headlineFieldRe` or special-case
the walker's path tail).

## How to verify a fix

Re-run the canary shape: rebuild any page whose section uses `features[].name` +
`description`, then assert the marker's `hits_before` counts name-field hits (a unit test on
`WalkContentStrings` with `{"features":[{"name":"X, not Y case","description":"…"}]}` is the
cheap half — it must return the name field). Control: the same test against the current binary
returns no name path (proves the test can fail).

## Scope note

> **⚠ CORRECTED 2026-09-03 by the `420 425` lane, verified independently before accepting — THE
> CLAIM BELOW IS FALSE AND IT WOULD HAVE WASTED A MEASUREMENT.** `cmd/brief-negation-check` does
> **NOT** import the walker. `grep -rn "WalkContentStrings\|nonProseFieldRe\|prosey"
> cmd/brief-negation-check/` returns **nothing**. It has its own traversal (`assessValue`,
> `flatten`) and reaches `datahelpers` only for text-taking scanners — `ScanDefineByNegation`,
> `NegationExempt`, `ScanBannedRegisterWords`, `ScanPracticeClaims`, `SplitPlainAssertionText`,
> plus the evidence/practice helpers. **None of them takes a field name, so none consults
> `nonProseFieldRe`.** It scans SPEC text; the walker scans `content_data`. Different populations,
> no shared filter.
>
> **So a before/after on the nightly is a CONTROL, not a measurement:** the prediction is NO
> movement attributable to a walker change, and movement on 09-04 against the 09-03 baseline means
> something ELSE moved and both lanes should want to know. Capture it as **N-of-M** — baseline
> 2026-09-03 07:41Z = **11 of 39** brief-supplies — because the denominator grows as sites are
> added and a bare count either side is uninterpretable.
>
> How the false claim arose: the two components sit next to each other, share the same scanners,
> and both talk about negations — so "shares the walker" was an inference from adjacency that one
> grep refutes. It went into a Scope note, which is the half of a bug file a later reader trusts
> without re-checking.

~~`cmd/brief-negation-check` imports the same walker for spec scanning — a fix here widens what
that check sees on its next rebuild; expect its suppression/finding counts to move and do not
read that movement as drift (its standing daily read is documented in the copy lane's handoff).~~
