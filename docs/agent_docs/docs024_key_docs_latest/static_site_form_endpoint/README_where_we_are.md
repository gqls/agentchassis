# Where we are — the form endpoint for static sites

The owner's running log. Plain prose, append-only, newest at the bottom. Add below; don't rewrite
what's above.

---

## 2026-09-03 — picking the thread up, and finding that most of the job was already built

You asked me to find this thread if it existed. It did: a pre-plan written yesterday that
deliberately decided nothing, a review from the team that owns the publishing side, and a note
that arrived this morning from the copyonline lane asking for exactly this and naming themselves
as the first customer.

The first thing I did was re-check the pre-plan's facts, because they were a day old and this
repository moves fast. Most of them didn't survive.

**The pre-plan said we had no machinery for handling forms. We have quite a lot of it.** Back in
late July, a different piece of work built two things: a package for sending email, and a package
for guarding public web endpoints — rate limiting, and a "is this a bot" check designed
specifically for forms. The email package's own notes say what it was built for: idea.uk's paid
report first, the robot-hands dossier next, and *contact forms after that*. Yesterday's search
missed both because it searched for words like "form submission" and neither package is named
anything like that.

**More importantly, we already have a live public endpoint that static pages post to.** It's at
tools.apis.uk, it's been running for over a month, and two of our sites — vonc.com and
robot-hands.com — already send data to it from published pages. I probed it today and it answered
correctly. So the big question the pre-plan spent most of its length on, "where should the receiver
live and is it safe to open a new public door", turns out to be already answered: the door is open,
it has locks on it, and we just need to add a room behind it.

**And one claim needs correcting outright.** The pre-plan said every form we've ever built is a
decoration that submits nowhere. That was measured on what the writing system *wrote*, not on what
the visitor actually *receives* — and in between, the page-building code quietly repairs it. Of 27
contact forms, 21 currently serve a working "email us" link, using the approach you chose back in
July. Only 6 are still dead, and they're all on sites that have no email address on file, which is
the one case the repair deliberately refuses to guess at.

I want to be clear that this isn't a criticism of yesterday's work — the query was written out,
dated and labelled as measured. It just couldn't have come out any other way, because it was
pointed at the wrong layer. That's worth recording precisely because it looked careful.

**While counting, I found a broken form nobody knew about.** gamesdesign.co.uk has a page whose
form posts to an address that returns "not found". I checked it three ways to be sure — the rest of
the site is up, a made-up address on the same site correctly fails, and the same test on two sites
that *do* have working forms comes back differently. So it's real. Our automatic checker can't see
it, and that's structural rather than an oversight: the checker holds a list of known-bad
destinations, and this one isn't on the list. It knows whether an address looks wrong; it never
asks whether anyone answers.

**What you decided today.** Build it properly rather than just design it. A submission gets stored
first and then emailed, with the recipient held in the database — so moving where the leads go is a
settings change, not a rebuild, which is what the copyonline lane asked for. Leave the 21 working
forms alone, but fix the 6 dead ones and the gamesdesign one. And give each site its own secret
token stamped into its form.

That last one is worth explaining. The existing endpoint works out which site a request came from
by reading a header the browser sends — and anyone can set that header to anything. For the two
tools we run today that only affects which bucket we count the request in, so it doesn't much
matter. But for something that forwards a message to a real person's inbox, it means a stranger
could send us anything and have it delivered as though it came from any of our sites. So each site
gets a token instead, and the header goes back to being just a courtesy check.

**What happens next.** I'll tell the team that owns the endpoint what I'm adding to it, create the
database tables, build the receiver, then change the one function that decides where a form points.
Copyonline is the pilot. Nothing changes for any site that hasn't opted in — a site with no entry in
the new table behaves exactly as it does today.

One thing to know for later: committing this and it being *live* are two separate events. That
endpoint runs on a separate machine that doesn't pick up our normal releases automatically, so I'll
confirm how it gets updated before I tell you it's working.
