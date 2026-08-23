# SUMMARY — 2026-08-23 — the divergence checker, and the fault that was not there

## What we are trying to do

When we publish a page, we want to know that the page people actually receive is the page
we sent. That sounds automatic and it is not: we build the page here, commit it, and a
separate process copies it out to storage and onto the network. Anything that goes wrong
in that last stretch is invisible from our side — the build says success, the record says
deployed, and a visitor gets something else.

So we started writing down a fingerprint of the exact bytes we publish, and built a checker
that fetches the live page every few hours, fingerprints what comes back, and compares.
A mismatch means "what we published is not what is being served".

## Where we have come from

The fingerprint and the checker were built over the previous four days and went live on
21 August. Almost immediately the checker raised an alert on one site, vetcomparison.uk,
and kept raising it — six times in a row, over about 21 hours.

That was reported as the checker's first success: a real, live, customer-facing fault that
we could now name precisely, where previously nobody would have noticed for hours. The
grading recorded in the handoff was "one true positive, no false positives, no misses".

## What we have done

Today we re-measured that alert instead of inheriting it, and it does not hold up. The page
was correct the whole time and no visitor ever saw a stale version.

The cause is that Cloudflare — the service in front of our sites — has a traffic-analytics
feature switched on for that one zone. When it is on, Cloudflare adds a small piece of its
own tracking code to the page as it goes out to the browser. It never touches the file we
published. So the page a browser receives is our page plus about 360 bytes of Cloudflare's,
and its fingerprint can never match ours.

That reframes the six alerts. They were not six independent confirmations of a fault. On
that site the checker could never match, so it raised the same unavoidable alert every time
it ran.

The reason it went unnoticed for a day is worth stating plainly, because it is the useful
part: the original session compared *fingerprints*, saw they differed, and concluded the
visitor was getting the old page. It never put the two pages side by side. When we did,
the entire difference was two lines. A fingerprint can tell you that two things differ; it
cannot tell you how — and "how" was the whole answer.

We have since:

- fixed the checker at source, so that when it sees a mismatch it asks the site once more in
  a way that tells the edge not to add anything; if that copy matches our fingerprint,
  nothing is reported. This cannot hide a genuine fault, because a site that really is
  serving an old page returns the old page either way;
- closed the false alert, with the mechanism and the evidence written onto it;
- corrected the handoff, the notes and the fleet-wide logs of past mistakes and traps, so
  the inverted grading is not quoted forward by the next person to read them;
- confirmed that two database changes the handoff still listed as outstanding had in fact
  been applied two days ago by a session that left no record, and added the missing records
  — which mattered, because without them the next routine run would have tried to reapply
  one and stopped the whole batch;
- confirmed that the 60-minute grace period added on Friday is genuinely live in the running
  software, using a table that records exactly which version of the code each service is on.

## Where we are now

The checker is live, running roughly every four to five hours across the fleet, and now has
no known way to manufacture this class of false alarm. Sixteen of the seventeen sites with
fingerprints match cleanly under every test we ran. The seventeenth is explained and
handled.

We have no evidence of any real delivery fault today. That is a weaker and more honest
statement than the one this lane made yesterday, and it is the one the measurements
support. What we have gained is not a caught fault — it is a checker we can believe when it
does catch one.

The fix is committed but not yet running in production; it needs the next build of the
service, as all code changes here do. Until then the checker will keep raising the same
alert for that one site, and it will be suppressed by the de-duplication that already
exists rather than filed again.

## Where we are going

Three things remain, in order of value:

1. **The empty-response gap.** If a site returns an empty page with a success code, the
   checker currently fingerprints the emptiness and could report a healthy page as broken.
   Two empty responses agree with each other, so the existing double-check does not filter
   them. This is small and known.
2. **Escalating on persistence.** A fault that clears itself in two minutes and one that
   lasts a day currently look the same. Distinguishing them needs the checker to notice that
   it has seen the same thing across several passes, which it does not yet do.
3. **Protecting the fingerprint from future gaps.** Every step that marks a page as
   deployed must also record what it sent; today all six do, but nothing stops a seventh
   being added that does not.

None of these need a decision. The one judgement call made today was to have the checker
stay silent rather than declare a page healthy when the edge is adding content — silence
claims nothing, and after yesterday we would rather it under-claim than over-claim.
