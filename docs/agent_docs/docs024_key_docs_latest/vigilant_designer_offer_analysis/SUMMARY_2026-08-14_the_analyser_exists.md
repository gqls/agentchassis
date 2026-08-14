# SUMMARY — 2026-08-14: the offer analyser exists

## What we are trying to do

Every site we build gets a written commercial premise at build time — who it is for, what it
offers that others do not, what a visitor needs before they will act, and why they would come
back. Then nothing ever reads it again. This programme exists to ask one question of every site,
for ever: **does this site actually answer its market's need, in a way that pays us?** Not "is it
well built" — that is the designer's question. Not "is it true" — that is the claims auditor's.
Is the *offer* right, and is the benefit to the visitor visible?

## Where we have come from

We started at the other end. The first pieces gave the platform eyes: a strategic review that
finally reads a site's own premise instead of guessing from the domain name, a way to refresh a
premise safely on a site that is already live, and two mechanical checks that hold a site against
the revenue model it recorded for itself. Then we spent a week making the inputs real — refreshing
the premise on thirteen sites, and hand-merging four fields into the two sites whose premises a
human wrote, without touching a word the human had written.

That week also taught us the thing that shaped tonight's work. One of those refreshes produced a
premise sentence that was simply invented — a twice-weekly technical blog on a site that has six
posts in four months, on other subjects. It was caught by reading it. Nothing on this estate
checks a premise record for truth, and a banned-word list cannot catch the next invention because
it is a record of the last one.

## What we have done

**The analyser is built, live, and proven.** It reads a site's own premise and the list of pages a
visitor can actually reach, and produces two things from one pass.

The first is new: **a ranked list of what that site should lead with** — six points for the first
site, each a sentence a page could actually open with, each traceable to the part of the premise it
came from, each marked for whether a competitor could say the same thing. Plus a list of what a
page should *not* open with. That ordering did not exist anywhere, on any site. The premise held
four paragraphs a human could read and no process could sort by. Another team asked us for exactly
this artefact two days ago; it now exists and they have been told, with the query to read it.

The second is findings: five per site, each aimed at a handler that already exists, each with a
test another agent can check.

**And it did something we did not ask it to.** Its "do not lead with" list opens with *"a
description of the site's page count or content inventory"*. That is, word for word, the mistake
another team hit last week when a page brief led with "23 free UK calculators" — and it is what the
owner meant when he said we should not talk about ourselves unless it is to the reader's benefit.
The analyser worked that out from the site's own premise, with no knowledge of the complaint.

**Three things we checked rather than assumed.** A bug filed this morning showed that our existing
strategic reviewer has been throwing away every finding it ever produced — silently, for weeks,
with every step reporting success — because of a mismatch between the shape its prompt asks for and
the shape the code can read. The new analyser uses the same path and would have inherited it
exactly; it does not, and we proved that by checking a pair of numbers rather than one, because
"nothing filed" is also what a genuinely clean site looks like. On the site whose premise a human
protected, the analyser wrote alongside that record rather than over it, and the protected record
came out character-for-character as it was. And the case where one premise field is deliberately
missing now announces itself in the artefact, computed from the database rather than judged by the
model, so a thinner analysis cannot pass as a full one.

## Where we are now

The analyser runs when we fire it by hand, and has run twice. It is **not** in the automatic sweep
yet — that change is written, tested against a transaction we rolled back, and deliberately held,
because enrolling it puts an extra language-model call on every site we sweep and about five new
actionable items per site, and the fleet hit its spending cap earlier the same day. By the
programme's own plan, that enrolment is the owner's call.

We also know one honest limit, found by reading the output rather than by anything breaking. The
page list the analyser reads carries page names, titles and search descriptions — not a word of
what any page actually says. So of the first five findings, three were grounded in things it could
genuinely see and two were reasoned guesses about page content. To its credit it said so itself,
inside the finding. Nobody should read "the analyser said X about this page" as "the analyser saw
X on this page", and the fix — give it the opening lines of each page — is written down.

## Where we are going

Three things, in order. The owner's decision on whether to put the analyser into every sweep. Then
the piece he approved tonight: teaching the claims audit to read premise records, so an invented
sentence in a premise can be caught by machinery rather than by someone happening to read it —
which matters more now than it did this morning, because the analyser grades every site against
those records. Then the small second version of the analyser that lets it see what a page says.

The larger open questions from the feature file are unchanged and still the owner's: whether this
becomes a council seat as well as an auditor, which of the two unrouted counterparts matters more,
and whether we ever try to grade the quality of a premise itself rather than a site's fit to it.
