# SUMMARY — 2026-08-01 — 164 CLOSED, live and induced; two residuals filed

*(New file, not an edit of `SUMMARY_2026-07-31_…`. That one said "committed, not live"; this
one says "live, proven, closed" and names two bugs that did not exist when it was written.)*

## What we're trying to do

Make the diagnosis loop's evidence bundle honest about its own gaps, so that absence of
evidence can never read to the verdict model as evidence of absence.

## Where we've come from

`164` was filed at the council gate's insistence after the `145` lane disclosed it and declined
to fix it. It was fixed, reviewed and approved on 2026-07-31, and held OPEN overnight because
committed-and-approved is not live. A chassis rolled overnight as `v1.0.1225`.

## What we've done

**Proved it at the binary, not the tag.** Both replicas grepped in a single exec for the four
strings the change added *plus* two pre-existing controls — because a clean grep against the
wrong pod or the wrong path looks exactly like a pass.

**Induced the size branch deliberately.** Nothing had tripped the cap since the roll, so
waiting would have been waiting for nothing. A scope naming a 204KB file first and a tiny
function second — the exact arrangement the old `break` destroyed — produced: the big body
skipped and named with its real size, **the alphabetically later function rendered in full**,
and a coverage line reading "1 of 2". The identical input pre-fix produced a heading with
nothing beneath it.

**The read-failure branch proved itself unprompted.** On an unrelated run the same morning, 12
symbols were requested, 7 read and **5 not** — now named in the text, correctly labelled a
tooling failure rather than a finding, with `truncated` correctly staying false. Pre-fix those
five vanished from both the artefact and the log.

**Negative control held**: two bundles from that same run dropped nothing and carry none of the
three markers, so the ~93% path is byte-unchanged.

**`164` is closed and moved to `/bugs_closed/`.**

## Where we are now

Live, verified on real traffic, closed. Two residuals were filed rather than absorbed:

- **`bugs_open/172`** — the fourth cap site, surfaced by this fix's own council round asking
  for the shapes my grep could not see (count-based, reslice). Latent at max 4 against a cap of
  5, and its discarded set is **non-deterministic**.
- **`bugs_open/174`** — **the more serious one, and it is live.** The instruction naming which
  code a diagnosis should examine is dropped in transit: the dispatch loop's `input_mapping` is
  an allow-list omitting `seed_scope` (and `runtime_page`), while the orchestrator behind it
  forwards both. Because the bundle assembler has a sensible fallback, a lost instruction
  becomes a **successful run against material nobody chose** — no error, no warning. Measured:
  3 of the 4 intakes that ever carried a seed scope lost it, and **two of those three are other
  lanes' completed diagnoses**, including `155`'s.

That last one is the substantive finding of the day. It was invisible until someone needed the
parameter to actually work and checked the result against what they asked for.

## Where we're going

1. **`174` first, and it is config-only** — add `seed_scope?`/`runtime_page?` to the dispatch
   loop's mapping from `claimed.*`, the same form its other ten keys already use. No image
   needed. Its candidate 2 (a lockstep test over the two mappings) closes the class rather than
   the instance; candidate 3 (report which branch of a fallback chain supplied the value)
   generalises past this bug to every fallback in the platform.
2. **`172`** is unowned and small: report the truncation, and add `ORDER BY type`.
3. **Anyone repeating `164`'s verification needs `DISPATCH=1`** until `174` lands. That is
   written into the closed bug file.

## The uncomfortable part, kept because it is the transferable part

Three self-inflicted errors on one small fix, and **every one was caught by two things that
should have agreed and didn't**:

- a **fabricated council verdict** written into the owner's log while the submission was still
  queued — caught by re-reading my own paragraph;
- an **induction that did not induce** and reported success — caught by checking the symbol
  count against what I had asked for *before* checking anything else;
- a **detector matching my own instrumentation**: I put the marker strings into the symptom
  text, and the symptom is embedded in every bundle, so my SQL matched my own prose. It failed
  in both directions at once — a false pass that would have "confirmed" the fix against the
  broken binary, and a false failure on a true positive I had already read on screen. The
  correct instrument existed in the unit tests (`inScopeSection()`); I did not carry it across
  from Go to SQL.

All three are in `WRONG_CALLS.md`. The pattern worth keeping is not any one of them: it is that
a single measurement never caught anything here — the disagreement did.
