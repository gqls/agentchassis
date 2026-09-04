# SUMMARY — infographics, 2026-09-04

First read-out for this lane. Written at the milestone where the lane was created and its first
measurement reframed the question the estate had been asking for two weeks.

---

## What we are trying to do

The owner asked, on 31 August, looking at the first paid customer build: *why didn't we use
infographics to take the place of much of the explanatory copy?* On 3 September he decided that
infographics should be fleet-wide and framework-driven, wherever they help a reader understand a
concept.

Underneath that sits a question nobody owned: when a section of a page has something to explain,
which form should the explanation take — words, a drawn picture, or a graphic built out of real text
and real values? This lane exists to answer that and to make the framework's own rules say it
consistently.

## Where we have come from

The platform has two quite different machines for putting an explanatory graphic on a page, and
almost everything written about infographics concerns only the first.

The first machine asks an image model for a picture. The site planner writes a row asking for an
image of kind "infographic", the model draws it, and it ships as a JPEG. Every layer of that machine
is built and wired — generation, provider routing, plan admission, page rendering — and has been for
months.

The second machine builds the graphic out of the page itself: a comparison rendered as a real table,
a process as a numbered flow, a checklist as a real list, a chart whose bars are drawn from
registered figures. It is not a picture. It is a drawing made of text, and it can be selected,
translated and read aloud by a screen reader.

Since the owner's question, four separate threads have investigated why the first machine produces
nothing. They found in turn that the planner had been told to use it sparingly, that no component
could display its output, and that the sites being built had no figures to draw. The first was true
and had already been fixed; the second was wrong; the third turned out to be measured in a way that
could not have come out otherwise. Each session retracted the one before it, promptly and in
writing. The last of them reached the right conclusion: the instruction has never been tested,
because no site capable of producing an infographic has been planned since the instruction changed.

That is careful work and it is not wasted. But all four were looking at the first machine.

## What we have done

This lane read the whole corpus — two hundred and six files mentioning infographics — re-measured
every figure against the live database rather than repeating it from a document, and then ran the
one query nobody had run: how much is the second machine actually producing?

The first machine has produced **one graphic in the platform's entire history**, on one site, in
August.

The second has produced **forty-five, across seventeen different sites**, and it is accelerating —
four on 2 September, fifteen on 3 September, nine by the middle of 4 September, against roughly one
a day through August. Seventeen domains is the control that matters: this is not one thread seeding
examples by hand, it is the framework choosing them.

We verified one at the live page rather than trusting the database, because a row marked deployed is
not a rendered graphic: a launch-promotion checklist inside a blog article on websitepromotion.co.uk
really is there, with its own markup and forty-eight items, while an invented address on the same
site returns a not-found. The check could have failed and didn't.

We also found a contradiction in the rules themselves. The live planner instruction tells the
planner to use a drawn infographic *for numbers*. Our own imagery design decision says a drawn
infographic must *never* carry real numbers, because image models invent values — which is the
entire reason the code-rendered chart was built. And the same instruction, further down, tells the
planner to keep all wording out of the image. So the one job that is uniquely an infographic's under
the current wording is a job the surrounding rules forbid it from doing.

We are explicitly not claiming that this contradiction explains the count of one. Three sessions
have already made that shape of mistake on this question in a single day, and the count cannot be
explained by anything, because the instruction has never been exercised. This is a defect found by
reading the specification, and it is waiting at the far end of the test everyone agrees is next.

## Where we are now

The lane exists, with its five standing documents, its place among nine neighbouring threads written
down, and a short list of things it will not do — it will not cut a prompt migration, change the
imagery vocabulary, build a component, build a detector, or dispatch a build at another lane's site.
Those belong to the lanes that own them, and each of those lanes stopped at its own boundary
correctly. The gap was between them, not inside any of them.

The headline is that the owner's question already has a working answer, and we were not counting it.
Comparisons rendered as real tables, processes as numbered flows, checklists inside blog articles —
explanatory graphics taking the place of explanatory copy, on seventeen sites. What is missing is
not the capability. It is that the word "infographic" in our rules and our queries names only the
machine that isn't working, so the machine that is has been invisible to every measurement we took.

Two corrections are owed to the shared register, where entries are read by review seats as ground
truth: one entry still records these components as never having been used, which was true when
written eleven days ago and is not true now.

## Where we are going

The next step is small and cheap and it is the one every thread agrees on: plan a single site that
already has registered figures, and watch which form the planner reaches for on a genuinely numeric
section. Twenty-one sites qualify. Choosing one is the work — the obvious candidate holds copy the
owner has approved, and planning it would be a bigger act than it sounds. We have written down what
we expect to see before running it, so the result cannot be fitted afterwards.

If that test shows what we expect, the fix is a wording change assigning quantities and comparisons
to the graphic that can hold a number, and concepts and scenes to the picture — handed to the lane
that owns the prompt, as bytes for the owner to read, not cut by us.

And beyond that there is a check worth building, in the thread that owns checks: a section whose
subject is a sequence, a comparison or a set of thresholds, that shipped as prose anyway. That is
the owner's original complaint, stated as something a machine can find.
