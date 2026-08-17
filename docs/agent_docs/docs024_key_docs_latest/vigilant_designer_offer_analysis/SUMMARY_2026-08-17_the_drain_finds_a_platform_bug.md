# SUMMARY — 2026-08-17: the analyser starts running itself, and its drain finds a platform fault

## What we are trying to do

Give every site a written, ranked answer to "what should this page lead with, and what should it
never lead with" — derived from the site's own recorded strategy rather than from anyone's taste —
and then turn the gaps between that answer and the live site into work the platform does by itself.

## Where we have come from

Two weeks ago this was an argument about copy. The owner's complaint — *we should not talk about
ourselves unless it is to the reader's benefit* — became a premise record per site, then an
analyser that reads it, then findings that flow into the ordinary work queue. On 15 August the
owner approved enrolling it into the automatic improvement sweep, and we watched one full cycle
work end to end. What we had **not** seen was the machine choosing a site we had never analysed and
doing the whole thing unprompted.

## What we have done

**It now does that.** In a twenty-four-minute window the sweep picked two sites of its own accord,
analysed both, wrote their ranked records and filed the findings — taking us from three sites with
a ranked record to five, with nobody touching anything. The judgements are good: on the games
design site it worked out, unprompted, that the page should lead with *"calculate the exact gravity
your platformer needs… without signing up"* and should **not** lead with the site's own tool count.

**And the findings' drain turned up a platform fault, which is the more valuable half.** A quarter
of our pages belong to tools and widgets, and the platform correctly refuses to let a generic
rewrite clobber one. But on the route our findings actually take, that refusal left no trace: the
task died, and the reason expired within a day. Three parts of the platform guard this; two left a
note for a human and the third did not. We added the missing six lines, put it through the review
council — approved first time by all thirteen reviewers — and it is now live and **proven**: eight
refusals across four sites have left notes where there had been none in the system's entire
history.

The check worth mentioning is the one that could have embarrassed us. Eight notes proves nothing by
itself, because a broken fix that filed a note *every* time would produce the same eight and more.
So we checked the opposite: six ordinary pages were rewritten successfully in the same period and
left no notes at all. That is what makes it proven rather than probably fine.

## Where we are now

Five of twenty-three sites carry the ranked record; the rest need the sweep opened again, which is
a cost decision and stays the owner's — it is about one site per fifteen minutes and every visit is
the expensive kind. The platform fix is closed. Along the way we measured that **twenty-six**
content tasks have died silently on tool-owned pages in the past week, from **six** different parts
of the system — the automated checkers, the design auditor, the cross-linker and others. We can
prove three were this exact fault; the rest are past the point where evidence survives. All of it
is visible from now on.

We also learned, twice in one day, that a deployment can look complete and ship no new code at all —
the first build carried none of the day's work because its version label never changed. Everything
now gets checked at the binary, with a control that is capable of coming out negative.

## Where we are going

Three things, in order. **One:** the analyser writes a test alongside each finding describing what
"fixed" would look like, and nothing currently reads it — we found a case where the repair
reintroduced the exact fault it was filed to remove, and a survey says about a third of those tests
could be made machine-checkable. That is the next build. **Two:** claim-checking the premise records
themselves, which the owner approved on 14 August and which nothing on the estate does today.
**Three:** the tool-owned pages still cannot be repaired, only refused — making a refusal visible is
not the same as fixing the page, and that route is still missing.
