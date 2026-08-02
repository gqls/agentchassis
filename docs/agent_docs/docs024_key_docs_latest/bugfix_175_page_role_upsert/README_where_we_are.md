# Where we are — the page upsert that quietly overwrote the wrong page

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, evening — what this is about

There is a piece of SQL that appears in several places in the platform. It says,
roughly: *"create this page; and if a page of that name already exists, update it
instead."* It looks like the safe, idempotent thing to write, and four different
sessions wrote it in four different files over the last few months.

It is not safe, and the reason is a single missing word. When the statement
updates an existing page instead of creating one, it updates *some* of the
columns — the title, the address, the content — but not the column that says what
kind of page this is. So if a tool deployment lands on a name already used by, say,
an ordinary article, the article's content is replaced by the tool's, and the page
still calls itself an article. Nothing errors. The code that asked for the page
gets back a page id and has no way of telling whether it created something or
flattened something. The machinery that looks after tool pages never sees it,
because as far as the database is concerned it is not a tool page.

We already knew this — one instance of it was fixed last week (`bugs_closed/081`),
after it was found looping on one site for three months. What is new is that when
the reviewing council looked at that fix, one of the reviewers said, in effect,
*"you have fixed one of these; how do you know there aren't others?"* There were
four others, and the note recording that (`bugs_open/175`) has been sitting open
and unowned since yesterday.

## What I did about it

The obvious repair is to add the missing word in four places. I did not do that,
because it fixes today's four and does nothing about the fifth someone writes next
month — which is exactly what happened to get us here.

Instead there is now one shared piece of code that every one of these page-creating
steps calls, and it answers the "that name is taken" question once, in four
explicit ways:

- the name is free → create the page, as before;
- the name belongs to a page already doing this job → refresh it, exactly as
  before;
- the name belongs to a page that has **never been published** and is doing a
  different job → take it over completely, including correcting what kind of page
  it says it is;
- the name belongs to a page that **is live** and is doing a different job →
  **stop, change nothing, and file the decision for a person.**

That last one is the important one, and it is not a new idea — it is the answer
last week's fix arrived at, for a good reason: deciding which of two pages should
hold a role is a judgement no rule we have can make reliably, and getting it wrong
breaks a page a visitor is looking at right now. Better to leave it alone and ask.

## How bad was it, really

I want to be straight about this, because the temptation is to make it sound
worse than it is. I checked the live database for pages sitting on names these
steps would claim: there are four, all published — two on the robot-hands site,
one on idea.uk, one on lendzy. So the trap is real and loaded.

But nothing has actually stepped on it. For the collision to happen you would need
a tool with one particular name to be deployed on one particular site, and no such
tool exists today. The report-page version cannot collide at all, because it names
its pages with a random identifier.

So: this is prevention, not a fire. The note that filed it said its severity was
unmeasured and that somebody should measure before choosing a fix. That is what
the measurement says.

## The part I got wrong, and it is slightly embarrassing

Alongside the fix I added an automatic check that will spot this pattern if anyone
writes it again — it runs whenever a commit touches a Go file, and it warns rather
than blocks.

I ran it across the whole codebase to see how noisy it would be. It reported zero
problems across all 1,120 files, and I was about to write that down as evidence
that it was well behaved. It was not evidence of anything: I had run it against a
copy of the code that already had my fix in it, so zero was the only answer it
could have given. A check that has never been seen to go off is not a check.

Run properly, against the code as it was before the fix, it finds exactly the four
places the bug report lists, and nothing else — including none of the five other
places that use similar SQL deliberately and correctly. That is the number worth
having, and it is now recorded with the command that produces it.

## Where it stands

The code is committed, so it will go out with the next chassis build (someone else
rolled one an hour before I finished, so it missed that one). It has been
submitted to the reviewing council; I will act on whatever comes back, and the
verdict is what decides whether this is finished or has another round in it.

## 2026-08-02, later — approved, live, and two of the objections were right

The council came back **approved**, with four advisory objections and none of them
serious. Two of them were right about something and changed what we wrote down,
which is the point of having them:

- One seat noticed that our claim "this work-item type has one producer and nothing
  automated consumes it" came from a **comment in a test file**, not from asking the
  database. Fair hit. Asked properly: no live agent definition mentions it anywhere
  in its configuration, and there are no rows. The claim was true — but it was
  folklore until somebody ran the query, and the seat exists precisely to catch
  comments being quoted as current fact.
- Another asked whether a page-upsert helper already existed before we built one.
  One did, on the plan-sync path — and it does the **opposite** thing on a
  collision: it re-types whatever it hits. That is correct for it (a plan is
  entitled to say what a page is) and wrong for ours. Nobody would notice picking
  the wrong one, because both compile and both return a page id. That is now
  written down as a trap.

Meanwhile another session rolled the chassis, which killed our review run
mid-flight — a known hazard here, and the reason we resubmitted before rolling
anything ourselves. Their *next* build picked up our commit, so the fix went live
on their roll rather than ours; we verified it by looking inside the running binary
on both replicas rather than trusting the version number.

The one thing not settled: the new code will take over and re-type a page that has
never been published. Three seats independently flagged that as the widest new
power in the change, and it is safe only because of a rule a comment states and a
human enforces. That is now **RFC 010**, for the owner: leave it as is, or make each
caller opt in explicitly — about four lines. The bug itself is closed.
