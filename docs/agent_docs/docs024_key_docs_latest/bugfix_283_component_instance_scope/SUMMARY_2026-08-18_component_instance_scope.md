# SUMMARY — component instance scope, 2026-08-18

Third in the series (previous: `SUMMARY_2026-08-17_…`). The milestone: **the mechanical
three-quarters of the programme is complete and serving**, the council trail on this case is
closed with an approval, and what remains is one well-scoped design task plus a short list of
named residuals.

---

## What we're trying to do

Unchanged: make it genuinely safe to put interactive components — calculators above all — on a
page more than once. Their element names were fixed text, so a second copy would read the first
copy's inputs and answer with them: a believable wrong number on a consumer-credit site.

## Where we've come from

Yesterday ended with the converter built and approved but running nowhere. Since then the properly
numbered release rolled, and the whole execution arc ran: a single trial conversion, then the full
mechanical batch, all through the platform's own work-tracking machinery as the owner ruled.

The trial converted perfectly — every predicted number exact — and then taught us the most
valuable thing this programme has found: **the re-render machinery had no concept of "the template
changed"**. Its thorough mode (rebuild from templates) fired only for three named causes; every
template fix the platform has ever made therefore shipped the old bytes under a green status. That
is fixed — "template changed" is a recognised cause, and the fixer now requests precisely the
affected pages rather than a site-wide sweep. En route we also hardened the migration practice
itself (SQL embedded in config is invisible to a probe run; the verify now compile-checks it) and
retired, on schedule, the daily alarm built to notice this exact moment.

## What we've done

Sixty-nine components converted and **serving** instance-scoped names — verified at the served
pages on four different domains, with zero unrendered tokens and zero duplicate names. Every
conversion carries a restore snapshot. Every non-conversion is a **guard refusing safely with the
reason named**: one id spelled like a colour code, one pre-existing duplicate inside a single
component, seventeen tool-owned pages whose own pipeline must apply the change, two pages where
the save floor refused a render that covered too little — and those last two are both "forked"
components, a pattern now flagged for one deliberate investigation rather than left as
coincidence.

The reviewer council pushed back once more mid-execution (round 4), and its gating concern — that
a config change of this shape can silently land on an unloaded copy of an agent's definition —
was worth the challenge: we answered it with measurements rather than assertion, including a
side-effect sweep across all fifty-three re-rendered pages, and round 5 came back **approved**.
The one sibling defect the council asked us not to leave untracked (other producers of
reason-less re-renders) now has a filed tracking item.

## Where we are now

The mechanical work is done, live, and reconciled to the last row: 69 converted (two since
deactivated by their owning lane's rebuilds — expected and pre-agreed, their templates still carry
the conversion), four small residuals parked with their routes named, and the audit trail runs
unbroken from the bug filing through five council rounds to the served pages. **The original
defect remains only where it always mattered most: the twenty-five components that genuinely
declare into global scope — twenty-three of them the loan-and-mortgage calculators this bug was
filed about.**

## Where we're going

One design task: the judged pipeline for those twenty-five. An AI rewrites each script (wrap,
rewire the inline handlers), the same acceptance gate passes or refuses each result, a size check
guards against the known truncation failure, and the loan-calculator site goes first because its
170-check oracle is the one independent witness we have. Before its first conversion: rebaseline
the byte-identical page check and move the oracle's selectors in lockstep. After that, 283 closes.
