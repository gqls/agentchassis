# Where we are — bugfix 238

Plain prose, append-only, newest at the bottom.

---

## 2026-08-11 — what this was, and what we did about it

**The complaint.** On the evening of the 9th, finetuning.uk's homepage started
showing five broken images in its case-studies block. The bug was filed the same
day by the lane that caused it (an improvement-loop run they'd fired for an
unrelated reason), and it was filed well — with the evidence, the live queries,
and three ranked fix options.

**The filed explanation was wrong, and that mattered.** The file said the
content generator had rewritten the section and, in the process, kept the bits
that look like writing (the image descriptions) and dropped the bits that look
like plumbing (the image addresses). That is a very plausible story about a
language model. It is not what happened. **The model was never shown those
fields and is explicitly forbidden from inventing them.** They belong to a
different category: values the system looks up for itself — an image from the
site's asset library, a link from the site's settings.

What actually happened is duller and worse. Those look-ups all failed, because
the data they point at doesn't exist for this site — and for two of them, has
never existed for *any* site we run. When a look-up fails, the default behaviour
is to quietly leave the field out. Then, when the rewritten section was saved,
the save **replaced the whole record** rather than updating part of it, so the
old values went with it.

**Why it had never bitten before.** There are two ways a page gets rewritten. One
of them merges — it keeps what's there and refreshes what changed. The other
replaces. Every rewrite of this page since May had been the merging kind, so the
image addresses survived untouched for three months and looked completely safe.
The first replacing rewrite deleted all of them at once. **The difference between
those two paths is the bug**, and neither of them is wrong on its own — which is
why nobody had spotted it.

**It was worse than five images.** Checking the live page properly turned up
something the report had missed: the same failure also removed the five "Read
case study" links and the block's main call-to-action button. Those didn't break
visibly — they simply vanished, because the template is written to hide a link
when it has nowhere to point. So the tidier, more defensive way of writing a
template turns out to fail *more* quietly than the careless way. That is a
genuinely useful thing to have learned, and I've written it down where people
will hit it.

**What we changed.** Two things, deliberately kept separate.

The first stops the loss. When one of these look-ups fails, the system now falls
back to the value the live page is already using, rather than dropping the field.
A fresh look-up always wins if it succeeds, so nothing goes stale; and if there's
genuinely nothing to fall back on, that now leaves a permanent record instead of
a log line that scrolls away. This is the fix the bug file itself asked for
first, and it's placed at the one point both rewrite paths pass through, so the
two can't drift apart again.

The second stops it shipping. It turns out the system was *already working out*,
at the exact moment of the broken render, which fields would come out empty
inside an image or link address — and then throwing that answer away. The
site-wide header/footer renderer uses it; the page renderer didn't. Now it does,
and can refuse to publish a section that would go out broken. That one is
switched **off** by default and ships with its switch held back until we've
confirmed the new code is actually running — because turning it on means a page
in the broken state can't be rebuilt until its data is fixed, and that's a real
cost that should be a deliberate choice rather than a surprise.

**The site itself is fixed and live.** The images are back, the five card links
are back, the call-to-action is back — verified on the actual served page rather
than in the database. That needed a judgement call I put to you: the same rewrite
had also changed *which* case studies the cards describe, so the old
image-to-card pairing no longer held. We matched them by subject instead, and the
descriptions the system itself had written for each image independently agreed
with that matching, which was reassuring. Two of the five were genuine judgement
calls and are flagged as such in the repair file. The old "read more" links all
pointed at pages that no longer exist, so rather than restore five dead links we
pointed them at the case-studies page.

**What we did not fix, on purpose.** Four other pages across three other sites
have the same damage. One is a real regression like finetuning's, but its images
don't exist to restore. The other three never had the values in the first place.
Restoring any of them would mean inventing content, so they're recorded rather
than papered over — and one of them is now the honest test of whether the fix
works, precisely because we didn't touch it.

**What's outstanding.** Both code changes are committed but **not yet running** —
they only take effect when the fleet next rebuilds its images, which isn't
something this thread does. Both have been put through the review council and
those verdicts are still coming back. And the third option in the original bug
report — having something automatically re-examine sites so this gets caught
within a cycle — remains correct and remains unavailable, because those checks
were switched off last week to save credits.

## 2026-08-20 — the detection we built nine days ago was switched off the whole time, and the repair we planned turned out not to be needed

Picking this bug back up after eight days, the first thing to check was whether it was still real.
It is, but almost nothing about it is where the file said it was.

**The detector we built was never turned on.** Back on the 11th we shipped two things: a fix that
stops the loss happening, and a detector that notices when it does. The fix has been running in
production since the 12th. The detector has been sitting in the code, switched off, ever since —
because turning it on is a separate switch, and nobody threw it. It has never once fired in the
platform's history.

The reason nobody threw it is worth telling, because it is not carelessness. Our own internal
reference card still said the code was "not yet running, waiting for the next release". That
sentence was true when it was written and became false the next day, and nobody went back to
change it. So every reader since — including the sessions best placed to throw the switch — was
told, in the place we keep our most trusted notes, that it was too early. **A note that goes stale
does not just misinform; here it quietly prevented the thing it was describing.** The note is
corrected now, with the cost written next to it rather than tidied away.

**So the detector is on** — the safe half of it. There are two halves: one just records "this page
shipped a broken link or image", the other refuses to rebuild such a page at all. We turned on the
recording half today. The refusing half stays off until the known broken pages are dealt with,
because otherwise it would block those very pages from being rebuilt, which you would experience
as things mysteriously failing. That ordering is your call from earlier today and it is now written
down in the system itself, not just in a document, so the next session inherits the reasoning.

**The repair we were going to build would have done nothing.** The plan was: find the pages that
lost data, work out which could be repaired automatically, and repair them. Before building it we
measured which pages that would actually be — and the answer was none of them. In almost every
case the page never had the data to lose: the place the value was supposed to come from has never
existed on that site at all. There is nothing to restore, and a rebuild would restore nothing.
What those pages need is for someone to supply the missing information, or for the template to
stop asking for it. That is content work and a decision, not an automated repair.

I want to be plain that this was the most valuable half-hour of the session. Building the repair
first would have produced something that looked right, passed its tests, ran cleanly, and changed
nothing at all — and we would probably not have noticed for weeks.

**And a genuinely good piece of news, which we had never actually checked.** In August we fixed a
related bug where rebuilding a page silently deleted its buttons. We believed it worked; we had
never proved it on real traffic. Today we could, by asking the archive a simple question: across
every page rebuild the platform has recorded, did any of these values ever go from present to
blank? The answer: it happened 66 times, on 11 sites, all of them between the 11th and the 14th of
August — **and not once since the fix landed on the 14th**, across more than three thousand
rebuilds. That is as close to proof as this kind of thing gets, and it had been sitting in the data
waiting to be asked.

**Where that leaves the bug.** Still open, and the reason has changed again — which is the third
time on this file, so it is worth saying what the reason is now. The prevention works and is
proven. The detection is on for the visible kind of breakage. The invisible kind — where a missing
link makes the button disappear entirely rather than leaving an empty one — still has no automatic
detector, because the check that would find it deliberately ignores this class and the scans that
would run it are switched off to save money. And ten pages need a human decision about data that
was never there. None of that is a code fix, which is why the bug cannot honestly be closed by
writing code.

## 2026-08-21 — closed, and the last day of it was the useful one

The bug is closed. What made it closeable was not the code — that was done days ago — but finally
being able to say, with evidence, that nothing is left in it that anyone still needs.

**The thing that had been broken is provably not broken any more.** We could show it on real
traffic rather than a test: across every page rebuild the system has recorded, values of this kind
were lost 66 times, all of them between the 11th and 14th of August, and **not once since the fix
landed on the 14th** — across more than three thousand rebuilds. That is the strongest evidence
we have ever had for a fix on this platform, and it had been sitting in the data the whole time
waiting for someone to ask.

**Your two decisions yesterday closed out most of what was left.** Saying the contact email should
appear let us point the contact block at the place the address actually lives — and looking at the
pages afterwards showed two things the database view had hidden. The email was *already* on the
pages, but only because it was baked into an old copy of the page: right on screen, and one rebuild
away from vanishing. And the same block was serving a **dead telephone link** on six pages across
three sites, which none of our database checks could see. Saying "do both" fixed that properly: a
site with a number now shows it, and a site without one shows nothing at all instead of a broken
link. Every dead contact link on the estate is gone.

**Nothing has been swept under the carpet, which was the condition for closing.** Five things
remained; I checked each rather than assuming. One turned out to be already fixed by another
thread — and I nearly told them it had broken again, because I was looking at a page address I had
guessed rather than looked up, and got a "page not found" that answered every question with a
confident zero. Two are now sitting in your review queue as proper items, each explaining what is
missing and the two ways to settle it. The remaining two were never this fault at all: three pages
on one site with a different problem, and one page that was copied wholesale from another website
and keeps whatever it originally had.

**What is genuinely still open is one question for you**, and it is a design question rather than a
bug: there are two ways this system rewrites a page, and they disagree about what to keep. We have
patched the route that matters and proved the patch works, but eight other places write the same
data and nobody has ever measured whether they lose anything. My recommendation is to measure
before building anything — twice this week, measuring first stopped us building something that
would have looked right and changed nothing. That is written up as RFC_042, alongside an older
paper asking the same question about the neighbouring column; they should be answered together.

---

**2026-08-22 — you asked me to scope the detector so we could measure it. Scoping it measured it.**

The question behind RFC_042 is simple to say: nine different pieces of code write the field values a
page section was built from, and only one of them is protected against quietly dropping a value that
came from somewhere else in the system — a link, an image, a contact address. We fixed the protected
one and proved it. Nobody had ever looked at the other eight.

So I went to write the handoff for building a detector, and before writing "here is what it would
cost" I wanted to know how big the problem was. It turns out the system already keeps a
before-and-after copy of every change to that field, going back to the 9th of August. That is enough
to answer the question for changes made in-place — which is what the other eight writers do.

**The answer is zero.** Across 279 changes made by the unprotected writers, not one resolved value
was lost.

I want to be careful about that zero, because a zero is the easiest wrong answer to get. My first
instinct was to check it by running the same query against the values written by the language model
instead, and that came back zero too — which looks like confirmation and is actually the same zero
twice, since both halves lean on the same joins. If the query were broken, both would be silent.
The check that settles it is to run it somewhere we already know the answer is not zero: the
protected writer's own history, where an earlier round of work had found losses by a completely
different method. It came back with 72, in the right class, on the right three days in August, and
none since the fix landed. So the query can see losses. There aren't any at the other eight.

**Then why is anything still open?** Because the instrument has four holes, and I would rather write
them down than pretend the zero is bigger than it is:

- it can't tell us *which* writer made any given change — the column that looks like it names the
  program is actually just the network connection, so even a positive result couldn't be routed to
  anyone;
- it never records a section being *created*, only changed or deleted, so a writer that makes a new
  section with something missing is invisible to this method entirely;
- for about a quarter of the changes it can no longer work out what the section was supposed to
  contain;
- and it only goes back thirteen days.

**What I'd suggest, and it is not the detector.** The cheapest useful thing in the whole file is one
line of code per writer: have each one say its own name when it writes, which the archive already has
a place to record. No new machinery, no database change. It permanently fixes the first hole and it
makes every future check of this kind possible — including ones we haven't thought of yet. It is
worth doing even if you decide the rest isn't.

One thing I've written into the file as a hard condition rather than a suggestion: if we do build the
detector, **whatever reads its output ships in the same commit.** We now have two separate warning
codes in the system, 41 records and 28, that nothing anywhere reads. A third would be a habit rather
than an accident.

The file is `bugs_open/355_HANDOFF_2026-08-22_eight_of_nine_content_data_writers_cannot_be_observed_losing_keys.md`.
It says plainly that it can be closed with **no code at all** if you'd rather take the zero and put
the census on a monthly schedule — that is a legitimate answer, and it's cheaper than the rest.

---

**2026-08-22, afternoon.** You ruled: build it — option (c), the detector. So the order of work is
the one already written in the bug file: first the one-line-per-writer change where every writer says
its own name when it writes (so anything we ever find can be routed to someone), then the detector
itself together with the thing that reads its output, in the same commit. The stricter "refuse the
write" mode stays unbuilt until the detector has actually seen a loss — no point arming a gun at a
population of zero. One thing you didn't decide and I haven't assumed: the sister question about the
*rendered page text* column (RFC_008) is still open; today's ruling covers the structured content
only. Each piece of code goes through the review council before it ships.

---

**2026-08-22, later — you said ship it, and it's live.**

You asked two things this morning: fix the fact that two warning codes were being written and never
read, and ship the detector together with the thing that reads its output. Both are done, and the
second one ran for the first time this afternoon, on the real database, before I wrote this.

What now exists: a small program that runs every morning and asks three questions. Did any part of
the system quietly drop a value it was supposed to keep? Is any page serving right now with a
required value missing? And — for every warning already on file — is it still true, or has the page
since healed? Warnings that have healed get marked off. That marking-off had never happened in the
system's entire history: forty-eight warnings were closed today as verified-healed, the first ever.

Two safety habits are built into it, learned the hard way this week. It never says "all clear"
without first proving it can still see — every run it re-finds a set of old, known losses, and if it
ever can't, it refuses to report at all rather than hand you a false clean bill. And it writes a
one-page note of what it found every single day, including the quiet days — so if the note is ever
missing, you know the check didn't run, rather than assuming nothing was wrong.

It also has an honest red light: today it reports 32 required values missing across 13 pages. Ten of
those are the ai-agent-orchestration homepage cards we already parked for the imagery work; six are
a leopardess page from the same family; one is the gamesdesign contact page that genuinely has no
email. The rest are new sightings on dartsonline, finetuning and gaswholesalers pages — nobody knew
about them this morning. The daily job will show as "failed" until those pages get owners, and
that's deliberate: a red light you have to look at beats a green one that lies.

Two supporting pieces ride with it: every part of the system that edits page content now signs its
name when it writes (so the next investigation can say WHO, not just WHAT — that starts working at
the next release), and the archive now keeps a copy even when only the content changes and not the
visible page (a gap we only noticed while building this). Both are with the council now; the
archive change waits for their verdict before it's switched on.

The parallel session you spoke to about "option c" filed the companion piece: a register of ALL
sixteen warning types that nothing reads — ours were just the two we tripped over. That's
bugs_open/358, and it's the to-do list for applying today's pattern more widely, one type at a time.

---

**2026-08-23 — closed, and the system told us so itself.**

This morning's evidence arrived without anyone asking for it, which is the best kind. The daily
check ran on its own schedule and wrote its note. Overnight, the imagery work supplied the missing
pictures on the ai-agent-orchestration homepage — and the check noticed: its tally dropped from 32
missing values to 27, and the warnings for that page now name only the links that are still owed,
not the images that arrived. Nobody told it; it saw. And at 07:18 someone's ordinary section edit
became the first write in the system's history to be recorded with its author's name instead of a
network address.

So everything you commissioned is not just built but observed working: the detector, its reader,
the archive that now sees every kind of change, and the signatures. The bug file has moved to the
closed shelf.

One honest footnote: the second half of the advisory review can't report until the 1st of
September, because the fleet has used up its AI budget for the month — the reviewers are AI calls
and the meter is at its cap until then. The review can't block anything and the paperwork resolves
itself when it lands; if it comes back with a real objection, we fix it forward like anything else.

What the daily note will keep showing you, for whoever picks each up: five link values still owed
on the ai-agent-orchestration homepage, a leopardess page from the same family, and the new
sightings on dartsonline's two shop pages — each named every morning until someone owns them.
