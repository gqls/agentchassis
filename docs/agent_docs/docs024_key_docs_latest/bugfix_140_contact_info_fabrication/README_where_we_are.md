# Where we are — the contact page that invented a phone number

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning — what was wrong

Eight of our live sites had a contact page telling visitors the business is open
"Monday – Friday, 9am – 6pm". Nobody ever said that. No record anywhere holds
those hours. The component that draws the contact page simply made them up
whenever the site had not supplied any, and printed them in exactly the same
style as the real details, so there is no way for a visitor — or for us — to tell
the invented line from the true one by looking.

On vetcomparison.uk it went further. That page was built two days ago, and it
publishes a phone number: `+1234567890`. It is not a phone number at all, it is
the placeholder someone typed into the template years ago, and the site has been
serving it as a way to contact the business.

This was already written up as bug 140 back on 29 July, when it affected six
sites. Nobody had picked it up. By this morning it was eight, which is really the
whole argument for fixing the machine rather than the eight pages: two more sites
walked into it in the four days the ticket sat there.

## What turned out to be the real story

I expected this to need a decision from you. "Should a component invent a
plausible default, or show nothing?" is a judgement call, and it affects eight
sites that other work-streams own, so it did not feel like mine to make.

It is not a judgement call, and that was the useful discovery of the morning.
Every component carries a little contract describing where its data comes from
and what to do when it is missing. This component's contract **already says the
right thing** — for the phone, the hours and the address it says, in as many
words, *skip this field if you have not got it*. The template just ignored its
own contract. So the change is not a new policy; it is making the thing obey the
rule it already publishes. That needs no ruling from you.

The same contract, incidentally, has a proper way to say "here is a sensible
default" — the section heading uses it. So the system already knows the
difference between inventing a **fact** about a business and providing a
**label** for a button. Only the template had lost track of it.

## A second thing, which is arguably worse

While reading the template against its contract I found that it was looking for
the heading and the introduction under the wrong names. The sites are all
supplying them properly — "Get in Touch", "Contact Darts Online", "Reach us
directly", "Contact VetComparison.uk" — and the template was looking for
something else, finding nothing, and falling back to a generic "Contact
Information" every time. Every one of those eight bespoke headings, and every
intro paragraph, has been written and then silently thrown away.

Our own quality checker noticed this and recorded it on **18 May**. The finding
sat in a database column and nothing ever read it. That is a pattern we already
have open tickets about, so I have noted it and not gone chasing it here.

## What I changed

Three things, in order of how much they matter.

**The template now obeys its contract.** Each contact card appears only if the
site actually supplied that piece of information, and the four invented values
are deleted outright. A detail nobody supplied can no longer appear on a page.
While I was in there I pointed it at the right names, so those eight real
headings and intros will start appearing, and turned on the address field, which
was defined but never drawn.

**The detector that was supposed to catch this has been taught our own
placeholders.** We have a checker whose actual job is finding fabricated contact
details on pages — it raises alerts literally titled "Fabricated contact info".
It was looking for the textbook fakes: numbers beginning 555, `example.com`, "123
Main Street", "Lorem ipsum". It did not know a single one of the placeholders our
own component library ships. Across every page in the fleet its nine patterns
found one thing; the fabrications it could not see numbered nine. It was blind to
its own platform's output.

**And a new standing check reads the library itself.** The problem with a list of
known fakes is that the list is only ever as good as somebody's memory — a new
component with a newly invented default is invisible until a human thinks to add
it, which is exactly how this survived from the beginning. So the new check does
not use a list. It reads every live component and asks whether its fallback is
inventing a *fact* — a phone number, an address, a price, opening hours — as
opposed to supplying a *label* like "Read more", which is perfectly fine and
which the library is full of.

## Two things that were wrong in my own work, both caught before they shipped

The new check's first run accused a component of hard-coding a domain name. It
looked bad. It was wrong: that component prints "designed and built by
fundamentallyai.com", as a link when it has a URL and as plain text when it does
not — the same words either way, inventing nothing. I fixed the rule generally
rather than adding an exception.

Then, having tightened the rule, I ran it against half a dozen made-up examples
to check it still caught real problems. It missed one: a component inventing
"Weekdays 8am to 5pm" sailed through, because I had written the rule to look for
day names like "Monday". Had I not run that check, I would have shipped a
safeguard that misses the most natural way of writing invented opening hours.
Both are written up properly in the notes.

## Where this leaves the eight sites

The template is fixed and live now — it is configuration, not code, so it took
effect immediately. But the eight contact pages already exist as finished pages,
and they keep serving what they were built with until each one is rebuilt. So
today the fabricated hours are still on the live sites even though the thing that
caused them is gone.

That is why I have left the ticket open rather than closing it. The rule here is
that a bug is only closed when it is no longer reproducible on the live fleet,
and it still is. The pages correct themselves as they rebuild, and I would rather
prove that on a real page than assert it.

**One thing worth your attention, which I have deliberately not touched:** the
same number, `+44 (0) 7934 524 911`, is published as the contact number on six
different businesses. I traced it before assuming the worst — it comes from the
site records, it is your own number, and these are your own portfolio sites, so
it is real and correctly propagated, not another fabrication. But six businesses
sharing one number is a thing you may or may not want, and that is a call for you
rather than a defect for me.

---

## 2026-08-02, evening — it's done, and the middle bit is worth reading

The eight contact pages are clean. No site claims opening hours any more, because
no site ever told us what its opening hours are. `vetcomparison.uk` no longer
publishes `+1234567890` as a phone number. I checked all eight on the live web,
not just in the database, and each page now shows exactly the contact details its
own records support — nothing more.

**But I got something wrong in the middle of this, and it is the most useful thing
that happened today.**

This morning I fixed the template and told you the eight existing pages would
repair themselves as they got rebuilt. That was wrong. The rebuilds happened — a
backlog of nearly three hundred queued jobs cleared during the afternoon, six of
the seven contact pages were rebuilt — and every one of them came back with the
invented hours still on it.

The reason is a genuine trap in how the platform works. A "rebuild" comes in two
kinds. One regenerates each section of the page from its template. The other just
re-staples the page together out of the section HTML it already has stored. The
second kind is the default, and it faithfully preserves whatever was wrong. Our
own code says so in a test message — *"deploys stale HTML"* — in a file I had
already looked at earlier in the day for a different reason.

So the pages were never going to fix themselves. I had measured the queue
carefully and let that stand in for knowing what the queue would actually do,
which is a different thing and a mistake I have written up.

Once you gave me the go-ahead, the fix was small: seven jobs of the *first* kind,
one per page. They ran between ten past and half past seven, and the count of
fabricating pages fell in step with them — seven, six, five, three, two, one,
none.

**Two things turned out better than I expected**, both of which I checked rather
than assumed:

- **finetuning.uk now shows a postal address.** The component always had an
  address field defined but never actually drew it; the site had the data sitting
  there unused. Turning the field on gave it a real card.
- **idea.uk's phone number is now backed by real data.** Its page had been showing
  a number that no longer existed in its records — a stale leftover I flagged as a
  separate concern in the ticket. The rebuild repopulated it properly from the
  site's own details, so that worry is resolved too.

No site lost anything real. The only things that disappeared were never true.

**One thing I have left deliberately open**, and it is the sharpest observation
anyone made today — it came from the review council, twice, from two reviewers
independently. This is now the *second* time we have caught a component ignoring
its own declared rules, and both times the repair was to hand-edit that one
component. Nothing in the system actually enforces those rules, so there is
nothing stopping a third component doing the same thing. I have written that up as
a proposal rather than quietly fixing it, because the sensible answer depends on a
judgement call — whether to enforce the rules when a page is drawn, which is
thorough but could break existing pages, or when a component is first created,
which is safe but only helps future ones. That is your call, not mine.

---

## 2026-08-03, midday — you asked for the 68, and they are done

The one thing I left open yesterday was a judgement call for you: 68 fields across 20
shared components that say, in their own settings, "if nobody gave us this, leave it
out" — and then render an empty space instead. You asked me to gate them. They are
gated, live, and checked.

**They were not quite what my own write-up said they were.** I had described them as
fields with no guard around them, which made it sound like the fix was to put a guard
around each one. When I actually read the twenty templates, 62 of the 68 turned out to
be the *second half of a pair*: a spec table row is switched on by the row's NAME, and
the VALUE sitting next to it in the same row is the unguarded one. That matters,
because the obvious repair is wrong in two different directions. Guard the value where
it sits inside a table cell and you get an empty cell exactly as before — the warning
light goes off and nothing has changed, which is the worse of the two. Guard the cell
itself and a four-column comparison row comes out with three cells, so every column
after the gap slides sideways. Neither is what the setting asks for. What it asks for
is "don't draw the row at all", so that is what I did for the tables, and for the rest
I removed whichever element was genuinely safe to remove.

**Two of the 68 were not harmless blanks at all.** One of them is an article image:
with nothing to show, the page was emitting an image tag pointing at nothing, which a
browser turns into a broken-image icon and a wasted request. Another two were button
labels — the button was drawn because it had a link, but with no words on it, so the
page carried an invisible, unclickable control. Those are the same family as a defect
we already track separately. They are now switched on only when both the link and the
label exist.

**On how much was actually broken, I nearly told you the wrong number.** Counting the
stored data, 47 pages were missing a hero subheadline, and I was ready to say so. Then
I looked at the pages themselves: only ONE of them is actually serving an empty gap
today. The other 46 have perfectly good text on them, written at some point in the past
and still sitting in the saved page, even though the underlying record has since been
emptied. Across all twenty components, three pages show the defect today, not 75. The
larger number is a fair description of *risk* — what would happen if those pages were
rebuilt — but quoting it as damage would have made this look twenty-five times more
urgent than it is. Both numbers are written down now, each labelled as what it is.

**Nothing on any live site changes until that page is next rebuilt**, and only if the
rebuild is the kind that redraws sections rather than re-stapling the saved ones. That
is the same trap this thread fell into last time, so I have not assumed otherwise.

**How I know it works.** I took all twenty templates back out of the live system after
the change and drew each one twice — once with the information present, once with it
missing — through the platform's real drawing code rather than a copy of it. Twenty out
of twenty: the element disappears when there's nothing to say, and still appears when
there is. The second half is the one that matters; a guard that is too aggressive would
sail through a test that only checks things vanish. The standing daily check now reports
the component library completely clean, where yesterday it reported 68.

**One loose end from another thread, now closed.** Overnight, another session left a
note in this thread's file saying my previous fix was not actually in the running
software, with what looked like solid evidence. It was a false alarm, and the reason is
worth knowing: they searched the running program for a sentence that only exists in a
*comment* in the source code. Comments are stripped when the program is built, so that
search returns "not found" against every version ever made, including one that has the
fix. Searching instead for text the program genuinely uses finds it on both machines.
The fix is live. I have written up both the correction and the cheaper check, and I want
to be clear that they were right to raise it — they reported only what they saw, offered
an innocent explanation, and handed it to this thread rather than acting on it.

**One thing I would still like your view on.** The daily check reports these blank
fields but deliberately does not treat them as a failure — the reasoning was that 68 of
them already existed and a warning nobody can ever clear is a warning everybody learns
to ignore. That reasoning has now expired, because the count is zero. Making it a hard
failure would stop the next one being introduced at all. The cost is that a genuinely
new component from another team could turn the daily check red until someone adds one
line to it. I have not made that change, because it alters what an existing check does
to other people's work, and that is your call rather than mine.

**2026-08-03, late evening — a second thing that needs your call, and this one is
about the review process itself, not this bug.** The change that gated the 68 blank
fields (migration 295) touched twenty shared components used across the whole fleet,
and it went live the moment it was applied. It was never reviewed by the council —
and it *could not have been*: the submission script refuses anything that doesn't
touch platform code, and this change was entirely configuration. The earlier, similar
change (287) was only reviewed because it happened to ride along with code changes.
So there is a class of fleet-wide changes — edits to the shared component library and
to agent definitions — that today cannot be put in front of the reviewers at all,
however large their reach. Two ways to go: either the submission script is widened so
these configuration migrations can be reviewed like code, or we accept the gap
knowingly and rely on the daily check plus after-the-fact reading. I'd lean towards
widening it — this lane has now shipped two fleet-wide config changes in two days and
would have welcomed the review both times — but it changes what the review process
covers for everyone, so it is your decision. Written up as item 8 in the handoff.

**Also done this evening, from the handoff's list:** the commit hook now refuses the
false-shaped "review pending" label three sessions accidentally wrote (a real id or
nothing — an absent label stays allowed); the daily check retries its flaky fetch
instead of failing on bad luck; there is now a small tool that picks *provable*
deploy-verification markers (it caught, on its first outing, both the comment trap
that caused yesterday's false alarm and a subtler one of its own — a marker crossing
a non-ASCII character greps as absent from a program that contains it); the
two pages storing a raw model reply where their content should be are now a filed
bug (190) with one proven mechanical repair; and the failed blog rebuild on
finetuning.uk turned out to be the platform *protecting* the page — the rebuild came
back with too little content and the save was refused whole — with the details handed
to the lane that owns that site.
