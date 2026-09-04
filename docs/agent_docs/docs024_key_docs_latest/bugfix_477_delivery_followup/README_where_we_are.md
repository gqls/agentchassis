# Where we are — the confirm button, and the reminder nobody sends

Plain-prose log for the owner. Append-only, newest at the bottom.

---

**2026-09-04, late morning.**

You asked, in passing, whether we might repeat the hosting instructions in a follow-up email a week
or so after delivery. Somebody went to look at how that would fit, and found that the follow-up
email you were proposing to add to doesn't exist at all — and worse, that we already tell customers
it does.

Here is the shape of it. When a site is delivered, the customer gets a link. Opening it shows a page
with a button, and the page says: *"Pressing the button below tells us you have moved everything
across. You will not get any more reminders about it."* They press it, and a second page says the
same thing back to them: no more reminders.

There are no reminders. Nothing in the system sends one. I checked this morning rather than taking
it on trust: exactly one thing in the whole estate can send an email, nothing is scheduled to run
it, and the date the button records — the fact that this customer has finished moving — is written
into the database and then read by nothing at all. So the button asks the customer to do something,
tells them what it saves them from, and the thing it saves them from was never built. Pressing it
changes nothing except a date nobody looks at.

Nobody has pressed it yet, so no customer has actually been told this to their face. That is only
because we have delivered exactly one site so far, and that was last night's rehearsal.

**What I have done today, and what I have deliberately not done.**

I have deleted the false sentence. Both pages now say only what pressing the button actually does:
it tells us you have moved, and nothing is recorded until you press. I did not write a replacement
reason for pressing, and that was on purpose — the reason was the false part, and inventing a new
one is how you end up with the same problem in different words. Today's other two bugs are exactly
that shape: an email that says the download contains instructions when it doesn't, and instructions
that say you can open the site by double-clicking when you can't. All three were found by a person
reading the words and asking whether they were true. Nothing automatic found any of them.

I also left a test behind that fails if anybody puts the sentence back while there is still nothing
to send. That is not paranoia about a colleague; it is so that when we *do* build the follow-up, the
person restoring the wording has to delete the test on purpose and therefore has to notice that they
are making a promise.

One correction to my own expectations, worth you knowing because it changes the cost: this fix is
not live yet and cannot be until the next release, because this particular page's words are baked
into the program rather than kept in the database. The delivery *email's* words are editable without
a release; this page's are not. Releases are yours to run.

**What I need from you, and one warning.**

The follow-up email itself is the real fix, and it is the thing you actually asked for. It is also
the piece that can go wrong quietly, because anything scheduled that emails customers can, with one
mistake, email somebody every night. So I want to build it carefully and I need three answers:

How long after delivery should it go — "a week or so" needs to be a number. Which address should it
go to: we hold a couple of different email addresses per site, and the one published on the website
is not necessarily the one the customer ordered with. And thirdly, a warning rather than a question:
the only site in the entire estate that a follow-up would currently select is idea.uk from last
night's rehearsal, and the address on it is your own. The first time I test this thing end to end,
**you will get the email.** I would rather you heard that here than found it in your inbox.

Until you have answered, I will build it switched off: the code written, the schedule seeded but
disabled, so that turning it on is one deliberate flick rather than a surprise.

---

**2026-09-04, early afternoon.**

Three things since the note above, one of which you should know about before you next look at a
customer email.

**The same untrue sentence was in a third place, and it is the one that actually reached somebody.**
The delivery email itself — the one that went out with the rehearsal last night — said *"press the
button here so we stop reminding you"*. So it was not only on the two web pages I fixed this morning;
it was in the letter. That one was editable without a release, and the delivery lane rewrote it within
the hour. It now says *"press the button here to tell us you have moved"*, which is true. Nothing
false about reminders is now live anywhere.

**I have built the follow-up email you asked for, and deliberately left it switched off.** The code
is written and committed, the schedule is seeded but disabled, and turning it on is one command once
you have answered two questions. The care went into one thing above all: making it impossible for it
to email somebody twice. The way it works is that a site is *claimed* before anything is sent, in a
single database statement that can only succeed once — so even if the scheduler ran twice at the same
instant, or ran every minute for a week, exactly one email can go. I proved that against the real
database rather than a mock, including the case that matters most: a customer who presses the confirm
button in the gap between us choosing to email them and the email being sent does not get it.

**And then it turned out it cannot email anyone at all, for a reason worth your attention.**

We do not have a durable record of who any site was delivered to.

The rule in our own documentation is that a customer's address comes from the order they placed. That
is right, and for a site that came through an order it works. But idea.uk — the only site we have
ever delivered — was our own rehearsal, sent to your address typed in by hand, so there is no order
and there is no recorded address. The one other place it exists is the log of the delivery run
itself, and I checked how long those are kept: **under twenty-four hours.** A follow-up due in a week
would be looking for it six days after it was thrown away.

I only found this because I ran the query with the calendar deliberately relaxed, so that it *had* to
return idea.uk, and it returned nothing. Without that, the follow-up would have sat there switched
on, selecting nobody, looking perfectly healthy for as long as anyone cared to watch it. That is the
kind of failure I would rather find in an afternoon than in a month.

The fix is small and belongs in the delivery step rather than in mine: record the address on the site
at the moment we hand it over. I have handed that to the lane that owns delivery rather than reaching
into their code. Until it is done, the follow-up can reach ordinary customers who came through an
order, and cannot reach anything we deliver by hand.

**So, two questions and one warning, unchanged from this morning and now more concrete.**

How long after delivery should the follow-up go? The code refuses to run without a number rather than
picking one for you, so nothing happens until you say.

And the warning stands: idea.uk is the only site this thing can currently see, and its address is
yours. The first time it is switched on for real, **you get the email**. Better you read that here.

**One thing that went wrong on our side, for completeness.** The review council that checks work like
this was dead for about forty minutes this morning — the account had run out of credit. It has been
topped up and everything is running again. It matters only because a dead review looks exactly like a
finished one unless you go and look, so I have written down how to tell the difference.
