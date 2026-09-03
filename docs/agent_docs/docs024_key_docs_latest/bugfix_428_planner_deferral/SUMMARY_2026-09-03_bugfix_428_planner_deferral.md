# SUMMARY — 2026-09-03, bugfix_428_planner_deferral

The first summary for this lane. Written at a real turn: the fix stopped being a better instruction
to the planner and became a check the framework performs. Read aloud-able; no field names unless
they are genuinely the subject.

## What we are trying to do

Every site we build starts from a strategy document that says, among other things, which kinds of
page the site ought to have — a homepage, an articles index, individual articles, a directory of
things, a profile page per thing. An AI planner then designs the site. We want a guarantee that
sounds obvious and until today was not true anywhere in the system: **if the strategy asked for a
kind of page and the finished site does not have one, somebody or something has to have written down
why.**

Not "the planner must always obey the strategy" — it is allowed to disagree, and that is
deliberate, because it is the thing looking at the actual brief. What we want is that a
disagreement is visible and checkable, rather than silent.

## Where we have come from

The original complaint was that the planner kept ignoring recommended page types — about three times
in four when the strategy named one. Investigation showed it was not ignoring them at all: it read
them, understood them, and deliberately left them out, because its own instructions gave it the
final say provided it explained itself. That was working as designed, and the design was thin.

The first fix, on 2 September, tightened the instruction: name each type you are dropping, and give
a real reason for that specific type, not one vague sentence covering everything. It also built a
review queue so a person could look at the flagged items, and it deliberately refused to build
anything that would act on them automatically — an earlier automatic-action mechanism had destroyed
live content on a customer site, and the owner had ruled against repeating that.

That fix worked. The planner now names every type it drops and gives a per-type reason, every time.

And that is how we found the real problem. On 3 September another team found the planner declining
to plan any articles for a site, with a perfectly well-formed reason: the articles would be
"satisfied by the blog infrastructure". There is a piece of software matching that description. It
is real, it is wired up, and it stopped running on 24 April. Nobody had noticed for four months. So
the site shipped with no articles and a tidy explanation, and every check we had was green.

The reason nothing caught it is simple and slightly embarrassing: **the field the planner writes its
reasoning into has no reader.** Nothing in the code ever looks at it. A true reason and an invented
one were the same artefact.

## What we have done

We gave the channel a reader. After the planner finishes and the plan has been through all its
processing, the system now compares what the strategy asked for against what the plan actually
contains, and records anything missing.

The part that took the most thought, and the part worth remembering: it looks at the pages **three
times**, not once. Because while building this we confirmed something worse than the original bug —
sometimes the planner does its job and our own validation step throws the work away. On one site the
planner produced nine pages including five real articles; validation returned four, with every
article gone. No error, no warning, nothing recorded anywhere. It had been doing that since May, and
it stayed invisible because on an established site the deleted pages get restored a moment later.
Only a site with no existing articles shows the loss.

So the check now distinguishes three different things that used to look identical: the planner never
proposed it, we deleted it after the planner proposed it, and we deliberately held it back for a
known reason. Three records, three different owners, instead of one shared silence.

It also does not cry wolf. When the planner declines a page type and points at some other mechanism
to supply it, the system checks whether that mechanism is actually running. If it is, the decision is
sound and nothing is flagged. Only a promise pointing at something that has stopped becomes a
finding. That is the difference between a check people read and a check people switch off.

Alongside it: a small honesty fix to the admin review page, which was offering a "release" button on
items that have nothing to release, and which would have failed with a confusing error the first
time anyone pressed it. And a database change that lets the planner state its decisions in a
structured way, written and deliberately **not** switched on — turning it on before the reader ships
would ask the planner to fill in a field nothing reads, which is precisely the mistake this whole
ticket is about.

## Where we are now

The code is committed and reviewed by the automatic review council, whose verdict is still pending.
It is not live: it needs the next routine rebuild of the system, which happens on its own schedule
and is not ours to trigger. When it does go live it turns itself on — there is no switch to remember,
which was a deliberate choice, because this estate has a history of building a detector and leaving
it switched off.

Two things are honestly incomplete and both are written down rather than glossed. The check works at
the level of "does this site have any pages of this kind", so on an established site that loses a few
new ones it stays quiet; we count that gap rather than close it, so it can be measured instead of
argued about. And the underlying defect that deletes pages during validation is not fixed here — it
belongs to another team, who have taken it. We built the detector for its whole class, not the fix.

## Where we are going

Three things, in order of who is waiting.

The council verdict needs reading and acting on, because the code is already on the shared branch.
After the next rebuild, the database change gets applied and we take the first real measurements —
carefully, because the other team's fix will legitimately drive one of our numbers to zero, and a
successful fix and a broken detector look the same from the outside unless the baseline is taken
first.

Then a decision that is yours. Somebody passed on your ruling that guides should be their own kind
of page. It is written down and deliberately not started, because there are two jobs inside it: adding
the type, which is cheap and changes nothing on its own, and re-labelling the 167 guide pages already
live across 20 sites, which changes what every blog and guide listing on those sites resolves to and
deserves a proper review. Tell us whether you want the cheap half now or both together.
