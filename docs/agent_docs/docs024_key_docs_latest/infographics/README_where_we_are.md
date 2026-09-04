# README — where we are, infographics

Plain-prose running log for the owner. Append-only, newest at the bottom.
Full path: `docs/agent_docs/docs024_key_docs_latest/infographics/README_where_we_are.md`

---

## 2026-09-04 — the lane opens, and the first thing it found was that we have been counting the wrong thing

You asked me to take charge of infographics. Before proposing anything I read everything the
estate has written about them and re-measured the numbers myself against the live database.

**The short version: there are two entirely different ways this platform can put an explanatory
graphic on a page, and every conversation we have had for the last two weeks has been about the
one that does not work, while the one that does work has quietly gone from nothing to forty-five
graphics across seventeen sites — nine of them since yesterday morning.**

### The two routes

The first route is a **generated picture**. The site planner writes a row asking for an image of
kind `infographic`, an image model draws it, and it is deployed as a JPEG. This is the one every
document means when it says "infographic".

The second route is a **component**: a piece of the page built out of real HTML and CSS from real
values — a comparison table, a numbered process flow, a checklist, a chart whose bars are drawn
from registered figures. It is not a picture at all. It is a drawing made of text.

### What the numbers say

Counted this morning against the live database:

| route | graphics on the fleet | sites |
|---|---|---|
| generated picture (`kind='infographic'`) | **1** | 1 |
| code-rendered component | **45** | 17 |

The single generated infographic was planned on 2 August, on mortgagecalculator.co.uk, and is the
only one in the platform's entire history.

The forty-five are spread across seventeen different sites — advertise, copyonline, cv1,
dartsonline, designblog, farmerinsurance, fundamentallyai, homegarden, lendzy, leopardess, loanzy,
mortgagecalculator, oufe, remortgagecalculator, robot-hands, seotools and websitepromotion. That
spread matters: it is not one thread seeding examples by hand, it is the framework choosing them.

And they are recent. One or two a day through late August; then **four on 2 September, fifteen on 3
September, nine so far on 4 September.** Something changed in the last three days and the curve
turned upward.

I checked one at the served page rather than trusting the database: the launch-promotion checklist
article on websitepromotion.co.uk returns HTTP 200 and really does contain a rendered checklist —
its own markup, forty-eight list items — inside a blog article. An invented URL on the same domain
returns 404, so the check could have failed and didn't.

### Why this matters for the thing you actually asked for

On 31 August, looking at the first paid customer build, you said there was not enough imagery and
asked *why we didn't use infographics to take the place of much of the explanatory copy*. On 3
September you decided infographics should be fleet-wide and framework-driven.

Four separate threads have since gone looking for why the planner never asks for one. They found,
in order: that the prompt told it to use them sparingly (true, then fixed); that no component could
display one (false); that the sites being built had no registered figures to draw (true but it
explained nothing); and finally — correctly — that no site capable of producing one has been planned
since the instruction changed, so the instruction has simply never been tested. Each of those
sessions did careful work and each retracted the one before it.

**Every one of them was measuring route one.** Meanwhile route two answered your question. A
comparison of tools rendered as a real table on seotools.co.uk, a regulation process drawn as a
numbered flow on advertise.co.uk, a launch checklist inside a blog post on websitepromotion — those
are explanatory graphics taking the place of explanatory copy. That is what you asked for. It is
happening. Nobody counted it because the word "infographic" is attached to the other route.

### The contradiction I want to flag before we spend anything

The live planner instruction, changed on 2 September, tells the planner to use a generated
`infographic` "for **numbers**, comparisons or steps".

Two of our own written rules say the opposite. The imagery design decision recorded last May says a
generated infographic "must **never** carry real numbers", because image models invent values —
that is why the code-rendered chart exists at all. And the same planner instruction, a few
paragraphs later, tells the planner to keep **all wording out of the image**.

So the instruction asks for a picture of a number, with no text in it, drawn by a model that cannot
be trusted with numbers. The one thing that is uniquely an infographic's job under the current
wording is the one thing the surrounding rules forbid it from doing.

I am not claiming this is why the count is one — that mistake has already been made three times on
this question and I am not making it a fourth. Nothing can explain the count, because the
instruction has never been exercised on a site that could answer. What I am saying is that when we
finally do run that test, this contradiction is waiting at the other end of it, and it is cheap to
resolve first.

### What I propose to do, and what I need from you

I would like to keep the fleet-wide decision you made and **route it to the answer that is already
working.** Concretely: when a section's point is a quantity, a comparison or an ordered sequence,
the framework should reach for a code-rendered component — the table, the flow, the chart — and when
the point is a concept, a process or a scene, it should reach for a drawn picture. That is close to
what the words already say; it is the assignment of "numbers" that is on the wrong side.

That is a change to the planner's wording, which the prompts lane owns and which you read as bytes
before it applies. I will write the specification; I will not cut the migration.

Nothing needs a decision from you today. The next cheap step is to plan one of the twenty-one sites
that already has registered figures and simply watch what the planner picks. I will say what I find.

— the infographics lane, first entry
