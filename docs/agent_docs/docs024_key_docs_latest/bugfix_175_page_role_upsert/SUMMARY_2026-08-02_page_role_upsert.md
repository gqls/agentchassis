# Summary — the page upsert that quietly overwrote the wrong page (2026-08-02)

## What we're trying to do

Stop a particular piece of SQL from silently destroying live pages, and — more to
the point — stop the *next* engineer writing it again. The SQL says "create this
page, or update it if the name is taken". It appears in several places in the
platform, it looks like the safe idempotent thing to write, and it is wrong in a
way that leaves no trace.

## Where we've come from

A week ago one instance of it was found and fixed (`bugs_closed/081`). It had been
quietly re-raising the same work item on one site since May — three months — because
the statement updated a page's content but not the column that says what *kind* of
page it is. So the page got the new content, kept the old identity, and the check
that asked for the page in the first place fired again on the next sweep, for ever.

When our reviewing council looked at that fix, one seat objected that it fixed one
arm of a shape it had seen recur, and that nothing had checked the siblings. The
fixing lane took that seriously, ran the grep, found **four more**, and filed them
as `bugs_open/175` — a finding with the census done and the fix deliberately not
attempted, because the right answer was not obvious: two of these arms want opposite
things from a collision, and "make them all consistent" would have been actively
harmful. It sat open and unowned for a day.

## What we've done

We took the class fix rather than the four patches. There is now one shared piece of
code that owns the question *"the name I want is already taken — now what?"* for any
step whose page type is fixed by what the step is (the tool deployer always makes a
tool page; the report builder always makes a report). It answers in four ways:
create; refresh, if the existing page already does this job; take the page over
completely, if it has **never been published** and is doing a different job; or —
if the page **is live** and doing a different job — **change nothing at all, and file
the decision for a person.** That last one is the answer `081` arrived at, and it is
there because deciding which of two pages should hold a role is a judgement no rule
we have can make reliably, and getting it wrong breaks a page someone is looking at.

Two things came out of the work that were not in the brief. First, `081`'s guard
asked `build_status = 'deployed'` to mean "this page is live", and that is not what
it means: 35 of our 46 `needs_rebuild` pages have already been deployed and are
still being served. That is `bugs_closed/037` all over again, and it is now written
down as a trap in its own right. Second, the council's prior-art seat asked whether a
page-upsert helper already existed before we built one. It did — the plan-sync
path's — and it has the **opposite** collision policy, quite correctly, because the
plan genuinely is the authority on what a page is. Reaching for the familiar name
would silently re-type a live page. That is now written down too.

We also added a mechanical check that fires the moment anyone writes the bad shape
again. It finds exactly the four known instances in the code as it was, and nothing
in the code as it is.

## Where we are now

Fixed, reviewed and live. The council approved it (four advisory objections, none
serious, all of them answered rather than noted — two of them turned out to be right
about something and changed what we wrote down). The chassis carrying it is running
on both replicas, verified by looking inside the running binary rather than trusting
a version number. `bugs_open/175` is closed and moved.

The honest note on severity: **nothing had actually been broken by these four arms.**
Four live pages sit on names they would claim, so the trap was loaded, but for it to
fire you would need a tool with one particular name on one particular site, and no
such tool exists. This was prevention. The bug report said its severity was
unmeasured and that someone should measure before choosing a fix — that is what the
measurement says, and we would rather record it that way than make it sound worse.

## Where we're going

One thing is deliberately left open for the owner rather than settled by us. The new
code, on a page that has never been published, will re-type it and take it over. That
is a wider authority than anything the old code had, and it is safe only because the
steps that use it have a fixed page type baked in — a rule enforced by review and a
comment, not by the compiler. Three separate council seats flagged it independently,
and the architecture seat asked that the reasoning live in the architecture review
track rather than in a bug's footnotes, so the next change of this shape has something
to cite. That is filed as **RFC 010**, with the options costed: ratify it as built, or
tighten it to an explicit per-caller opt-in, which is about four lines of work.

The other loose end is small and named: one more page-creating step has no collision
handling at all, so it fails loudly instead of quietly. That is the opposite failure
mode and outside this bug, but it is written down rather than forgotten.
