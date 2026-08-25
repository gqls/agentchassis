# Where we are — the pages that don't deliver what they promise (`bugs_open/381`)

Plain-prose log, append-only, newest at the bottom. The owner maintains this too — add below,
never rewrite.

---

**2026-08-24, opening.** You looked at the finished garden-tools site and said two things: one
page was a wall of text, and the seasonal planner promised "month by month" and had no months.
The lane that built the site traced both to the same cause and filed it as bug 381 — the planner
had composed that page out of four components, none of which can render a list at all, so the
writer had nowhere to put twelve months and wrote four seasons of prose instead. Nothing anywhere
reported a problem: the sections rendered, the page deployed, every check passed.

That lane filed the bug and left it unowned. This session has picked it up.

**First job was checking the bug is still true, because other threads move things under us.** It
is. And it is not just that site: across the fleet, 741 pages were rebuilt in the last month and
**327 of them — 44% — contain no list, no table and no bold text anywhere in their content.**
Ninety-four percent of all the section slots filled in that month used a component whose template
physically cannot produce a list.

**Then I got the cause wrong, and the data corrected me.** My first explanation was that the
writer is forbidden from writing lists: the text slot it fills is declared as "plain text", and
the writer's rulebook says plain-text fields must contain no HTML. Tidy, and wrong. There is
another component, `article-body`, declared exactly the same way, under exactly the same rule —
and **it writes lists 76% of the time, against 7% for the one on your page.** The difference is
not the declaration. It is that `article-body` carries a one-line note to the writer saying "use
headings for sections, lists where things are lists", and the component on your page **carries no
note at all.**

So the fix is not a technicality about type declarations. **It is that nobody ever told the
writer what to do in that slot, and we have proof, in our own data, that telling it works.**

**What I am going to change, in plain terms.** Three things, all configuration — they take effect
the moment they are applied, with no software release needed.

1. **Tell the planner what each component can actually produce.** Today it picks from a list of
   names and descriptions that says nothing about whether a component can render a list or a
   table. It will now see that, worked out automatically from the component itself rather than
   typed in by hand — because we have a column where someone once tried typing it in by hand, and
   it is now wrong on twelve components and unread by anything.
2. **Tell the writer to use structure where the slot can hold it** — subheadings, lists, bold for
   the term a reader is scanning for, a table when the content is genuinely tabular. Copied from
   the note that already works on `article-body`. It will explicitly forbid images and figures
   inside these text blocks, at the request of the lane that is building proper image handling.
3. **Say plainly what these slots are**, so the rulebook stops contradicting the note.

**One thing I want to be straight about: this will not conjure a calendar.** I checked whether the
library really had the "34 list-capable components" the bug file counted. It does, technically —
but they are directories, calculators, quizzes and trackers, plus one pricing table and, embarrassingly,
the site footer. **There is no general-purpose checklist, comparison table or calendar component
in the library at all.** So a planner that knows what everything can express would still have had
nothing suitable to build your seasonal planner from. That is a real gap and I am recording it as
one, not quietly folding it into this fix.

**What I need from you, and it is the only thing.** To prove this worked I need to watch one build
that happens after the change. Garden-tools is off-limits — you said it can wait, and it is also
another bug's test case — and the lane that would normally build a new site has not been asked
for one, so it rightly will not invent one just to serve my bug. The options are: authorise one
new domain build; name one page on an existing site for a single rewrite; or I wait and measure
whatever gets built next anyway. **Left alone, I will do the last one** — it is free and honest,
just slower.

**Worth knowing, because it is the sort of thing that never surfaces otherwise:** before writing
any of this I asked seven other threads whether it would collide with their work. One of those
questions — asking whether a copy-checking tool would cope with the lists we are about to start
producing — **found a live bug in that tool.** It splits sentences at the end of table data cells
but not at the end of table *header* cells, so a sentence it "found" in a header could include the
table's own markup, and its repair would have pasted prose over the tags and broken the table.
That lane has fixed it. Nobody had a symptom; the question was the whole trigger.

---

**2026-08-24, later.** The fix is written, reviewed and committed, and it has not been switched on
yet. That last part is deliberate: these are configuration changes that take effect the instant
they are applied, with no software release involved, so applying them is the moment the fleet's
behaviour changes and it should be a decision rather than a side effect of me finishing.

The review council approved it first time, with three advisory notes. I acted on all three, and
one of them was genuinely useful — worth telling you because it is the kind of thing that would
otherwise have shipped quietly wrong.

**What the reviewer caught.** I had checked that the writer is *allowed* to produce lists, and that
nothing in the platform would reject them. I had not checked what the surrounding template puts
them *inside*. If those four text blocks had wrapped their content in a paragraph tag, then a list
written into them would have been a list nested inside a paragraph — which is invalid, which every
browser repairs differently, and which **nothing anywhere would have reported**: the database row
would look perfect, every check would pass, and only the actual page would be wrong. I checked:
all four use a plain container, so we are fine. But I was cleared by the reviewer's question, not
by my own thoroughness, and I have written that down.

**One thing I want to be plain about, because it is the honest limit of this work.** I can now tell
the planner what each component is capable of, and I can tell the writer to use structure. Neither
of those creates a calendar component that does not exist. If you ask for a month-by-month page
tomorrow, the planner will now *know* it has nothing suitable — which is better than choosing
blind, but it is not the same as being able to build it. Filling that gap is a separate piece of
work and I have recorded it as one rather than pretending this closes it.

**What I still need from you: one build to watch.** Nothing is proven until a page is written under
the new instructions. Garden-tools is off-limits, and the lane that builds new sites has not been
asked for one and rightly will not invent one. So: authorise a new domain build, or name one page
on an existing site for a single rewrite, or say nothing and I will measure whatever gets built
next anyway. The last option is free and honest, just slower.

**A small thing that says something about how this place works.** Before writing any of it I asked
seven other threads whether it would tread on their work. Five replied. One of those conversations
found a live bug in another thread's shipped code — a copy-checking tool that split sentences at
the end of table data cells but not table header cells, so its repair could have pasted prose over
the table's own tags. Another thread caught me reusing an identifier that was already taken.
A third refuted a guess of mine with better evidence than I had. **None of that came from anyone
having a symptom.** It came from saying out loud what I was about to change.

**2026-08-24, end of the afternoon — it is all live.** Both halves are applied and working, and
the fleet is now building pages differently from an hour ago.

What that means in practice: when the planner lays out a page it can now see, for every component,
whether that component can actually produce a list, a table, a set of cards, or only paragraphs —
worked out from the component itself rather than typed in by anyone. And the four text blocks that
carry most of the prose on the estate now tell the writer to use subheadings, lists and bold where
the content genuinely has that shape, instead of saying nothing at all and getting paragraphs by
default.

The single most satisfying number: the general-purpose text block that produced your wall of text
used to appear in the planner's list as "paragraphs only". It now appears as capable of headings,
lists and tables. That component is used on 181 pages.

**One delay, and it turned out to be an hour rather than days.** I held the writer half back because
another thread had just fixed a bug in a copy-checking tool that our change could have triggered —
their fix was written but I could not tell whether it was actually running yet. Two of the obvious
ways to check turned out to be worthless (one of them gives the same answer no matter what is true,
which is worse than no check at all). That thread then handed me the right method, I verified it
myself rather than taking their word, and released the same afternoon.

**What is still true, and I would rather say it than let it be discovered later.** The framework
still has no general-purpose calendar, checklist, comparison-table or step-list component. So a page
promising "month by month" will now be planned by something that *knows* it has nothing ideal to
hand — better than choosing blind, and the writer can now at least render a twelve-item list inside
an ordinary text block, which it could not before. But a purpose-built calendar component would be
better, and that is the next piece of work rather than something I have quietly finished.

**And the honest caveat on the evidence.** I know the mechanism works — I checked it at every layer,
including the parts that would have failed silently. What I have not yet seen is a page built under
the new instructions, because no build has happened since. That still needs one build to watch, and
it is still your call which.

---

**2026-08-24, evening — the components you asked for.** Three built: a **checklist**, a
**month-by-month calendar** (generic over any period — months, quarters, seasons), and a
**comparison table** that turns into readable stacked blocks on a phone instead of a squashed grid.

**One I did not build, and that is the useful part.** I was going to build four. Before writing
anything I read the closest existing component and found `mechanism-flow` already draws a
step-by-step process. So a "steps" component would have been a near-duplicate — the exact thing that
makes a library harder to choose from rather than easier. Three components, not four, because I
looked first.

**The thing I want to flag honestly is the comparison table.** A table is where a writer most wants
to invent a price or a star rating, and about two-thirds of our sites have no verified-facts record
for anything to be checked against. So I built it with **no price field, no rating field, no score
field — none exists to fill** — and the instructions tell the writer at the point of writing not to
state figures it has not been given. That is genuinely the most the component itself can do. It is
not a guarantee, and I have written that into the file rather than letting it read as solved. The
real answer to that problem is another thread's work, and I have said so.

**A mistake worth telling you about, because it is the same one three times.** While testing, I
found the templates producing a stray placeholder string when a field was missing, measured that it
never appears on any live page, and concluded our writers must simply be reliable. Wrong: it never
appears because **the platform already strips it**, which one command would have told me. I had the
number right and the story wrong — and I had made that same error twice earlier in the day. I have
withdrawn the warning I filed about it, kept the correction where the next person will find it, and
written up why the check I skipped was the cheap one.

**Where this leaves you.** The framework can now plan a month-by-month page, know it has a calendar
to build it with, and have the writer fill twelve real months instead of four seasons of prose. What
has not happened is a real build using any of it — the components are seeded and reviewed, not yet
exercised on a live site. That is still the one thing I need from you: a build to watch.

---

**2026-08-25 — the first real evidence, and one thing left that only you can start.**

The writer half works, and I can now show it rather than assert it. Looking at what the writer
actually produced since the change: of the pieces of text written into the blocks we changed,
**72% now contain a proper list, against 10% before, and every single one now has subheadings.**
That is the wall-of-text complaint answered, measured at the point where the change applies.

**A moment worth telling you about, because I nearly got it wrong.** My first look was at the
finished pages, and there the numbers barely moved — five out of forty-eight. Had I stopped there I
would have told you the fix had failed. It hadn't: most of those pages were *re-renders* of text
written before the change, which never go near the writer. Two measurements of what looked like the
same thing disagreed, and the honest answer was that they measure different things. I only knew that
after checking, and I have written down that when that happens, neither number is the answer until
you can explain the gap.

**What is left is genuinely one thing, and it is yours to start.** The three new components — the
checklist, the month-by-month calendar, the comparison table — are live and have **never been used
once**. Nothing is broken; I have confirmed all three appear in the planner's own live menu. The
problem is that the step which *chooses* components only runs when a **new site is built**, and no
new site has been built since. So the half of the fix that was the whole point of the bug is sitting
there untested.

**So: one new site build.** The command is in the handoff. What matters more than the domain is the
subject — pick something genuinely structured, a buying guide or a how-to or anything seasonal, and
all three components get exercised. A two-page brochure will exercise none of them and prove
nothing. Garden-tools stays off-limits as you asked, and it is also another thread's test case.

**Can we close this off?** Not yet, and I would rather say so. Half the bug is proven and half has
never run. One build settles it. If you would rather not commission one, the honest alternative is
to leave it open and measure whatever gets built next — that costs nothing and stays true. What I
don't want to do is mark it done because the configuration is in place; the whole of the last two
days has been about the difference between shipped and working.

---

**2026-08-25, closing.** You pointed the DNS, and that was the last thing needed. The bug is closed.

**The comparison, on the live sites, is the whole story.** Garden-tools — the site that started
this — serves exactly eight list items on every page, and every one of them is the navigation menu.
Not one content list anywhere, including on the page that promised a month-by-month guide.
Homegarden now serves forty to fifty list items per page on top of that same menu, and its front
page carries a proper twelve-month calendar, January through December, each month with its own
advice.

**I checked it the careful way round.** Before believing any of those numbers I asked the site for a
page that cannot exist. It correctly said "not found" — which matters, because two hours earlier the
same domain was answering "yes, here it is" to every address including invented ones, and I had
already told you the calendar was live when it wasn't. A number is only worth having if the
instrument could have said no.

**Three things I did not fix, so you have them plainly.** The checklist component was offered to the
planner and it never picked one — I don't know why, and that's worth someone looking at. The
comparison table never got a fair test, because the research for this subject came back with nothing
to compare. And your third complaint from the original review — too many cards stacking up on a
phone — I never touched; it needs a decision about page layout rather than a bug fix, and it has
never been written down as a job.

**What I'd take from the last two days, if it's useful to you.** Eight things I said turned out to be
wrong, and not one of them was caught by me re-reading my own work. Every single one was caught by
another thread asking a question they thought was routine. The pattern underneath most of them was
the same: I measured something one step away from the thing I actually cared about — a status column
instead of the page, a section count instead of the words, a clock instead of a timestamp — and each
time it felt exactly like measuring. That's now written up where the other threads will find it.
