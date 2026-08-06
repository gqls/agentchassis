# Where we are — the "does the page's JavaScript actually exist?" bug

Append-only, newest at the bottom. Plain prose.

---

**2026-08-05, morning.** Picked up bug 084 from the open pile. The complaint is
simple to state: a web page can tell the browser "go and load this JavaScript
file", the file can be missing, and nothing we have would ever notice. The page
still looks finished. Every status in the database says it built and deployed
fine. It just doesn't *do* anything when someone clicks a button.

Choosing which bug to take was harder than expected. There are 45 open bugs and
about 38 other sessions working this same repository right now, so most of them
are already claimed. The tool we have for checking ownership only reads committed
history, so it can't see anyone who is halfway through a fix — it said "someone
owns this" for almost everything. What actually worked was reading the other
sessions' live logs and seeing which bug numbers they were talking about. 084 was
one of the very few nobody had touched.

**Then I nearly got it badly wrong.** To find out how bad the problem was, I
searched all the live pages for script references and checked whether each one
was really there. I found one that was missing — a page on webdesign.co.uk asking
for a file called `...`, which is what a half-finished, cut-off AI response looks
like. That would have been a good story: a broken tool, shipped to a real site,
invisible to every check we own. I was about to write it up.

It wasn't real. The text I'd matched wasn't a script reference at all — it was a
*comment inside that tool's own code*, a programmer's note describing what the
tool does. My search couldn't tell the difference between a page loading a file
and a page merely mentioning one. And the pages most likely to talk about web
code are, of course, the tool pages — which are exactly the pages this bug is
about. I've written that up as a wrong call.

**When I measured it properly, the honest answer was: there is nothing wrong
right now.** I fetched all 541 live pages and read them the way a browser does.
There are 854 script references, pointing at 96 different files, and all 96 of
them load correctly. Not one broken reference on the whole fleet.

That changes what this fix is. It isn't a repair — it's a smoke alarm. The
problem it guards against has happened before (we once published a site's shared
JavaScript to an address nothing pointed at, and separately about 60 of 63 ported
tools nearly shipped completely dead — that one was caught by luck), and it's the
kind of failure that looks exactly like success until a customer clicks something.
But today it would find nothing, and I've said so everywhere rather than letting
the work read as if it fixed live damage.

**2026-08-05, afternoon — one thing I found that I'm glad I didn't barge past.**
The bug file's first suggested fix was to change an existing check so that it
actually fetches the file instead of just looking for its name in the page. That
seemed obviously right. It turns out another team already proposed exactly that,
months ago, and the review council ruled it off-limits: that check is shared
vocabulary, and quietly changing what it *means* would change the verdict on every
document that already uses it. They accepted a real cost for that — a legitimate
check of theirs had to be marked "impossible for now" and parked. So I left it
alone and built the new capability alongside it instead. Changing it properly
needs its own architecture proposal, which somebody should write.

**What I built.** A new check that takes every script and stylesheet a live page
asks for, works out the real web address, and requests it. The important design
decision is what it does when the answer isn't a clean yes or no. It only reports
a problem for a definite "this file does not exist", and it asks twice before
believing it. Anything else — the server refusing us, rate-limiting us, timing
out, having a bad day — is recorded as "couldn't tell" and reported as nothing.

That sounds over-cautious until you know the history: someone previously tried
this against the same site with a simpler script, and the site's security layer
refused all 63 requests. Under a looser rule that's 63 false alarms about a
website where everything was actually fine. So the check can be *blinded*, and it
will say so in the logs — but it can't be *tricked into lying*, and between those
two, blindness is much the better failure.

**Proving it works when there's nothing to find.** This was the part I spent most
care on. A check that reports "no problems" is indistinguishable from a check
that's silently broken. So I deliberately broke each safety rule in turn and
confirmed a specific named test caught each one — five of them. If I can't make a
guard fail on purpose, it isn't a guard, it's decoration.

**2026-08-05, evening.** The review council approved it first time round, with
five advisory comments. One of them was a genuine catch and I've fixed it: my
database query used a home-made test for "is this page live", where we have a
shared one specifically because home-made versions keep missing pages — 28 live
pages, in fact. My version would have quietly skipped them, and a check that
skips pages reports "all clear" for the wrong reason. That's the exact failure
this whole thing exists to prevent, so it was an embarrassing one to have made.
Now fixed and, more usefully, wired so that anyone who re-makes the same mistake
breaks 21 tests immediately.

And a second wrong call, right at the end, which is the one I'd most want
remembered. When I ran the experiment to prove that new safeguard worked, it
reported "nothing failed" — meaning the safeguard was useless. Except it hadn't
actually run: my script's edit silently didn't apply, so I'd tested nothing and
read the blank result as an answer. Same shape as the mistake I'd spent all day
building a check to prevent, made by me, in my own test tooling, on the same day.
The experiment now announces whether it actually changed anything before it
reports a result.

**One objection I could not answer, and haven't pretended to.** A reviewer pointed
out that this adds yet another kind of alert that nobody is assigned to act on —
we already have several, and we're creating them faster than we're draining them.
That's fair. Our own rules say this kind of alert is acceptable because the repair
is a human judgement, not something a program can decide. But the reviewer is
right that the pile is growing, and that's a call for a person, not for me to
argue away. It's recorded in the bug file.

**Where it stands.** The code is written, tested, reviewed and committed. It is
deliberately switched off: if you name a check the running software doesn't know
about, the whole scan fails, so it can only be turned on after a build carrying it
goes out. I checked the build that went out this evening and confirmed it does
*not* contain this yet — expected, since the commit came after it. So the bug
stays open until it ships, gets switched on, and is proven to catch a fault we
deliberately introduce. Two of the bug's five original suggestions are also still
untouched, which is the other reason it can't simply be closed.

---

**2026-08-06.** Done, and proven.

The build carrying the new check went out overnight. I checked the running
software actually contained it before changing any settings — that order isn't
fussiness: if you switch on a check the software doesn't recognise, the entire
scan fails, so config has to follow the build, never lead it. It was there, along
with a deliberately-wrong control string that wasn't, which is what tells you
you've tested the pipeline and not your own spelling. Switched on.

Then the part I'd flagged as needing your say-so, and it turned out to need less
than I thought. The documented way to make a check run fires the whole
improvement loop, which also sends automated fixers at a live customer site — a
real side effect for the sake of a test. But reading the definitions showed the
discovery agent on its own is only three steps and does no dispatching at all. So
I aimed the same message at that instead. Discovery ran, nothing was dispatched,
nobody's site was edited by a robot.

To prove the check can actually catch something I had to break something first,
because there is nothing broken to find. I picked the quietest page on the
quietest site — a guide page nobody had touched in three weeks — checked first
that a missing file there really does report "not found", took a checksummed
backup, and added one reference to a file that doesn't exist. Worth saying: this
changes the stored copy, not the page a visitor sees.

It caught it. One alert, naming the exact address, the exact page, the exact
status. And — the half I care about just as much — **it raised nothing else**:
eight other pages on that site, every real script and stylesheet they load, no
false alarms. A detector that flags your deliberate fault *and* a pile of
innocent things hasn't been proved, it's been flattered.

Then I put the page back, verified by checksum that it is byte-for-byte what it
was, and re-ran. The alert stayed open — which is exactly what I'd written down
it would do, because the check only withdraws an alert when it can see the thing
working again, and I'd *deleted* the reference rather than fixed it. Pleasing to
have a documented limitation turn out to be true rather than a guess with good
grammar. Finally I cancelled the test alert with a note explaining where it came
from: leaving a fake fault sitting in the queue for ever, while complaining in
the same breath about alerts nobody acts on, would have been a poor joke.

Bug closed and moved. Two of its five original suggestions were never this job's
to do and now point at where they actually live, and one still needs a proper
architecture proposal that I haven't written. The one thing I'd want you to know
if you only remember a sentence: **this check finding nothing is not evidence
that it works** — there's a two-minute procedure written down for proving it
still bites, and it should be run before anyone quotes a clean result.
