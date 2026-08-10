# SUMMARY — provocation pipeline, 2026-08-10

*Fourth in the series. Previous: `SUMMARY_2026-08-02`. This one marks a real turn:
the site stopped making a false claim, and a second question — legal exposure —
arrived and reshaped the work.*

---

## What we're trying to do

vonc.com promises a new provocation every day and invites you to argue against an AI
about it. Two things have to be true for that to work. The site has to actually
change every day, and what it publishes has to be something we are willing to have
our name on. This workstream owns the first and has now been handed a large part of
the second.

## Where we've come from

The promise was false for a long time and nobody noticed the reason: there was no
mechanism at all. Not a broken job — nothing. The rotation rules existed only as a
Python script under `docs/`, which the cluster cannot run.

That was built and proven at the start of August: a pool of provocations in the
database, a scheduled job that picks today's, builds the whole feed and commits it to
the site. It went live on 2 August and worked. But the pool held nothing newer than
26 July, so the site correctly re-served a three-week-old provocation under a heading
saying "Today's". The machinery was right and the shelf was empty.

## What we've done

**We stopped writing about it and put content in.** Six provocations now exist, and
the owner re-ordered and re-dated them himself before approving. Since this morning
the site has been serving one dated today. **The daily claim is true for the first
time since 26 July** — thirteen days of a false statement, ended.

Getting there involved a decision worth recording. The first seven drafts were
written to an audience the plan had recommended back in July — people who argue about
work and technology. The owner read them and said they were boring, and he was right
for a reason that turned out to be mechanical rather than a matter of taste: nobody's
sense of who they are is tied up in having views about meetings, so there is nothing
at stake in defending one. The audience was changed to everyday culture — food,
music, film, cities, generational habits — where people genuinely enjoy the fight.
The drafts were rewritten and they are visibly better.

**Categories were designed, approved and built.** The owner wants several
provocation threads running at once, each with its own audience. That needed a
decision about how to publish more than one, and finding the answer turned up
something worth knowing: the game engine barely checks the feed at all. It confirms
the entry exists and then hands whatever it finds straight to the AI as part of the
question. So one of the two obvious designs would have failed completely silently —
no error, no alarm, just an AI arguing against a lump of data. The owner ruled for
the design that fails loudly instead. It is built, reviewed, and live on the servers
since this morning, and it changes nothing until somebody adds a second category.

**Then the owner asked a different kind of question,** and it was the right one: if a
visitor writes something insulting about a real person and our AI appears to endorse
it, and we publish that on the internet, is he liable? Working through it moved the
problem: publishing is not where the exposure begins. Saying something about someone
to one other person already counts. So the controls had to target *what gets written
about whom*, not how long the link lives — which is why his own first suggestion, a
link that expires, was the one thing recommended against. It shortens exposure
without removing it, and the screenshot outlives the link anyway.

Three gaps were found by simply looking: nothing stopped search engines indexing a
published round, nothing filtered what visitors wrote, and nothing ever deleted it.
The first has been fixed by another session. The second — a check that refuses to
publish a round making a factual allegation about a named person — was built here,
reviewed, and approved.

**And the review process paid for itself twice this week.** The reviewers told us the
allegation check had ignored machinery we already had. They were right, and better
than that, they were right about something we had not spotted: the check had no
handling for the word "not", so it would have refused *"Nolan did not steal the
script"* — a defence of the person, blocked as if it were an attack. That is fixed.

## Where we are now

The site is telling the truth daily and has content through 15 August. The categories
work is live and dormant. The allegation check is written and approved but not yet
running, because it lives in a service that ships to a different machine on a
schedule that is not ours to set.

We are also, quietly, in a slightly unusual position: three sessions have worked in
this area this week, and twice work was already half-built by someone else before we
started. Nothing was lost, but it is the main risk in how we are working — and once
this week it would have stopped the whole project building rather than merely
duplicating effort.

## Where we're going

Three things, in order of how much they matter.

**Content will run out on 15 August.** Six days. The generator that writes
provocations automatically is built and approved but has never been tested against a
real model, and that test is the gate on trusting it. Either that happens, or someone
writes more by hand.

**Two of the safety measures are still only on paper** — a way for someone to
complain and have something taken down, and making it explicit that our AI judges how
well you argued and not whether your claims are true. Both are cheap. The second is
the more important, because it is the difference between a service that rates
reasoning and one that appears to certify accusations.

**And the categories work is half a feature until the game engine catches up.** We
can publish a second thread today; nothing would read it. That half belongs to
another team and the decision has not been made.
