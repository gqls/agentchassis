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
