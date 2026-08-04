# SUMMARY 2026-08-04 — `bugs_open/192`: the section plan that got buried one level down

Current state only. Chronology lives in `NOTES_…`; the plain-prose history in
`README_where_we_are.md`.

---

## What we're trying to do

Stop every page build in the fleet failing, find out why it started, and fix it at a
level that stops the *class* rather than this instance — because the thing that failed
is a pattern the platform recommends in writing and will use again.

## Where we've come from

Yesterday the `178` lane fixed a real defect: asked to make a small edit to an existing
page ("add a link to X"), the writer regenerated the whole section and dropped most of
the prose — measured at 4,439 → 1,806 characters on one page. Their fix adds a step that
hands the writer the page's current stored content first. The logic is correct and is
not in question.

That step was wired to write its answer back into the **same** slot the section plan
already lives in — a deliberate choice, documented in its seed, so that nothing
downstream needed changing. It is a pattern this platform recommends.

At 08:20 this morning the config for it went live. From 08:27 every page build failed.
Three lanes hit it inside forty minutes — a vet-comparison guide, a tool page on an
unrelated site, and webdesign.uk's own landing page, which is the shopfront for the
site-building product and could not be built at all. The `154` lane filed it as
`bugs_open/192` with the evidence it had gathered incidentally, said plainly it was not
diagnosed, and handed it off.

## What we've done

Diagnosed it, fixed it at source, and closed the class around it.

The cause: an action wired with `output_field: section_plan` returned **a wrapper around
the plan** — the plan plus two bookkeeping notes — on *every* one of its return paths,
including the eight it documents as "pass-throughs". The coordinator stores an action's
return value **wholesale** under `output_field`, so the real plan was demoted one level
on every build in every mode, not only the new edit mode. One cause, and it broke both
of the consumer's fallback routes: the second directly, the first because the
link-resolver step reads from the same slot and so was handed nothing.

Four changes, ordered by what closes the door:

1. **The action returns the plan itself** on every path. Its own header had always
   promised the plan came back "byte-for-byte unchanged"; the code did not do it.
2. **`extract_fields` gained an opt-in `required` list.** It used to omit a field whose
   configured paths all missed and still report success — so the run continued into a
   wall two steps later, under an error naming the wrong thing. Now, when asked, it fails
   where the fault is, naming the field, the paths tried and what was actually in scope.
   Default OFF fleet-wide; one step opted in.
3. **The loop's path-miss error now lists the keys it did find.** Had it done so this
   morning, this was a five-minute bug.
4. **A config seed, applied immediately and self-retiring**, so builds work on the binary
   already deployed. It goes structurally dead the moment the code ships.

Registered as **WFA-009**, with the landmine recorded where it fires — and the landmine
is deliberately about `output_field`, not about `extract_fields`, because reusing a key
as your output field is a *replacement*, not an annotation, and the wrong result looks
exactly like the right one.

## Where we are now

**The outage is over, and proven rather than assumed.** A failed job put back through the
queue ran to `complete`; the three jobs before it had died at the same step. That check
could have come out the other way.

**Still open, deliberately.** The Go half is committed and inert until a chassis image
rolls, so the wrapper is still produced on every build and the live seed is merely
stepping around it. This repo's bar for closing is *fixed and live*, and half of this is
not live yet, so the ticket stays open with a banner saying exactly what closes it.

Council review is submitted (`7afbf531`) and the commit carries `Council-Submitted:`,
never `Council-Reviewed:` on an unread verdict.

Three things are worth carrying beyond this bug:

- **The unit test asserted the wrapper**, so it passed on the code that took the fleet
  down — while its own comment three lines above described the correct behaviour. It has
  been rewritten and then *deliberately broken* to prove it now catches the fault.
- **The bug report's timing was wrong in a way that mattered.** It attributed an
  overnight failure spike to this defect; that spike is a different step, reachable only
  *after* this one succeeds. The wrong onset made yesterday's fix look innocent, and
  yesterday's fix was the cause.
- **The `090` diagnosis loop returned nothing** — three bundles, no verdict, no result.
  The diagnosis here rests entirely on first-hand verification, which is stated plainly
  rather than dressed in a run id that corroborated nothing.

## Where we're going

1. **On the next chassis roll** (whole-fleet, owner-run): pod-grep the new error string
   with a positive control on every replica; confirm the wrapper is gone at source on a
   fresh build; then a cleanup seed removing the shim path.
2. **Read the council verdict** and act on a REVISE or REJECTED — the code is already on
   the shared branch, so that obligation does not lapse.
3. **Then close it**, and hand the `178` lane the end-to-end verification this outage was
   blocking — a completed item is already waiting for them.

Two things are left open on purpose and belong to someone else: internal CTA resolution
is degraded until the roll (shimming it would have re-broken silently *after* the roll),
and the overnight `iter_N_generate_content` failures — roughly 38 runs between 21:00 and
01:00 — are a real, separate, undiagnosed defect that this bug's filing had absorbed.
Nobody is on them.
