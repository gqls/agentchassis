# Where we are — the web admin console

The owner's running log for this lane, in plain English. Append-only, newest at the bottom.
Nobody rewrites or reorders what is already here; corrections go **below** as dated additions.

This file was started late — on 2026-08-24, after the lane had already been running two days.
That is a gap, not a design: the entries before today were reconstructed from the lane's other
documents and are marked as such.

---

## 2026-08-22 to 2026-08-24 — reconstructed, not written at the time

You asked for "a web based admin on a domain of mine that I log into", so that you could follow
and contribute to the steps of each website build.

The first finding was that most of it already existed. There was an admin web app running in the
cluster, with a login, and an API behind it with around seventy admin routes — sites, pages,
components, specs, media, work items, customers, pipelines. What it did not have was any way to
reach it from a browser, and one missing screen: the one that shows a build's steps.

Getting to it turned into the awkward part of the week. The first route tried was the VPN that
was already running in the cluster. That went badly. The config file we handed you contained a
line that told your desktop to send *all* its name lookups through the tunnel, and on Ubuntu that
left the machine with no working name resolution at all — you could still reach the internet by
number, but nothing by name, which looks exactly like "the VPN broke my connection". You had to
purge WireGuard to get back. That was our error: the line had been spotted and filed as a
preference at the bottom of a page instead of being removed. A safe config now exists, and the
runbook has the whole diagnosis.

Even after that, the VPN never actually carried traffic. Your laptop was sending handshakes every
few seconds and nothing was coming back — and we could prove the fault was not your machine and
not your network, because a packet capture showed the handshakes leaving, and a separate test
proved your connection could reach other services on similar ports. The loss is somewhere in
transit or at the cluster node. That is still unexplained, and it has been deliberately parked
rather than chased, because you chose a better route instead.

On 2026-08-22 you asked whether to keep debugging the VPN or go for a web-based admin, and chose
the web. Then you chose the address — **admin.apis.uk** — and the gate: **Cloudflare Access**, so
you get a one-time PIN by email and an unauthenticated visitor never reaches the machine at all,
let alone the cluster. That decision also settled a question that had been blocking another lane
for days, about whether the service holding every site's data may be reachable from the internet.
The answer is a narrow yes: through the admin app's own gateway, on that one hostname, only after
Cloudflare has authenticated you, and the server is configured to refuse everything if that
authentication header is missing — so deleting the Access rule cannot quietly open the door.

It works. You have logged in and can see every site. That also answered a question the earlier
documents had flagged as unverified — whether an admin account existed for you at all. It does.

## 2026-08-24 — I was sent to talk to that lane rather than repeat its work

I came into this with a two-day-old picture. I had been handed the original plan from 08-22 and
was about to ask you to choose how the console should be exposed — a question you had already
answered, about a console that was already built and that you had already logged into. You told
me to go and correspond with the thread that had built it, which was the right instruction.

Worth saying plainly, because it is the sort of thing that will keep happening: **the fault was
not that my information was old, it was that I never checked whether it was.** Listing the lane's
directory takes a second and would have shown a handoff document written twenty minutes earlier.
I asked you a question instead. That check is now written into the lane's notes.

So this session did the next useful thing instead of rebuilding anything: it went and measured the
things the next piece of work depends on, and found several of them were wrong.

The next piece of work is the screen you actually asked for — the one that shows a build's steps.
There is a plan for it, written today by the other session. Three things I found change it:

**One.** The plan says the database has no column recording which site an orchestration belongs
to, and proposes digging it out of a JSON blob instead. There is such a column, it has three
indexes on it, and one of the two JSON paths the plan proposed matches nothing at all. So that
part of the job is smaller and safer than it looked, and a hidden failure has been removed.

**Two, and this is the one that may change the screen's shape.** The API the plan builds on
returns "orchestrations" — the machinery the platform runs. Those are not website builds. Around
half of them are not attached to any site, the field that would hold a step-by-step history is
empty on every single row, and for a given site most of the ones that *are* attached turn out to
be its routine overnight checks rather than its build. What you would actually call a build is a
chain of work items with names like "needs domain research", "needs strategy", "needs briefing",
"needs site plan", "needs design", then one per page. That chain reads like a build. Your own
apis.uk build is a ~~three-line example of it: research finished at 12:26 on the 22nd, and it has
been sitting waiting on vertical research ever since.~~ **[That struck-out claim is WRONG — the
build had not stalled. It ran twelve stages in about an hour on the afternoon of the 22nd. The
correction, and how I got it wrong, are at the foot of this file.]** I have not acted on this —
it is that lane's call, not mine — but I have put it in front of them, because building the
screen on the wrong thing is a week nobody gets back.

**Three.** There is a known trap where a build step can fail while the record still says
"completed" and the error column stays empty. The plan already builds around it, correctly. I
sized it: in the last seven days, sixty-seven completed runs carry a failed step that the obvious
reading would show as green. So it matters, and the plan's defence against it is the right one —
which I say with some feeling, because I twice talked myself into believing it was wrong before
checking properly. Both times the cause was the same: I trusted a search pattern without ever
looking at a single row it had matched. That is written up in the lane's notes and in the
fleet-wide record of wrong calls, because the check it produces is worth more than the finding.

## Two things that are yours, not ours

**The links host is not applied yet.** There is a change written and waiting for you to apply on
the box, which moves the customer confirmation links off the webdesign.uk shopfront onto their own
address. I checked today: that address does not exist yet in DNS, so it has not been done. It
matters more than it sounds. Right now the only thing keeping two cluster routes off the open
internet is a Cloudflare redirect that exists to *park* the domain — not for security, and nobody
chose it as a protection. The day webdesign.uk starts serving its own shopfront, that redirect
goes, and those routes become publicly reachable with no code change and nothing to notice.

**www.apis.uk works, but not for the reason I first said.** It does redirect to apis.uk — I
checked, and that part is real. I then told you the redirect rule you had been asked to create
had evidently been applied. **That was a guess dressed as a finding, and it was wrong.** The
lane working on the apis.uk apex read it and corrected me within the hour: the redirect comes
from a piece of code that already sits in front of that whole zone and sends every "www" address
to its plain form. Nobody needed to apply anything.

The distinction matters because of what I did. I saw the outcome you wanted, found a pending
instruction that would have produced that outcome, and joined them up — without checking whether
anything else could produce it. Something else did. So: the *result* is fine and needs nothing
from you, but you should not read "it works" as "the step was done", and if you do apply that
instruction, expect it to change nothing.

## 2026-08-24, later — correction: the apis.uk build had not stalled

Earlier today I told you your apis.uk build had been sitting for two days, stuck waiting on one
of its stages. **That was wrong, and I want to be plain about it because I said it to you twice —
here and in a message to another lane.**

It had not stalled. The stage I named finished eleven minutes after I looked at it, and the whole
build ran start to finish in about an hour on the afternoon of the 22nd: research, vertical
research, strategy, briefing, site plan, composition, page, design, imagery, re-render. There was
also an alarm saying the site could not be reached, which I reported as outstanding; it cleared
itself at 16:21 that same afternoon once the page started serving.

The lane working on that site checked and told me so. I re-ran the query myself rather than take
their word for it, and they were right.

**What went wrong is worth you knowing, because it is a failure mode rather than a typo.** I
looked at that build on the 22nd, when this session started. The session then stayed open across
two days. When I wrote it up today I reused the old result — and, worse, added a detail that was
never measured at all: "about two days now, not progressing" was me subtracting the row's
timestamp from today's date. The original reading was honest. The arithmetic I bolted onto it
turned a stale note into a false statement about a live build of yours.

The rule I broke is one this project had already written down: a measurement of *what something
was doing* goes out of date, while a measurement of *what happened at a moment* does not. So the
check is simply to re-run anything about current state before repeating it on a later day, and
never to compute an age from a timestamp without re-reading the row.

Two corrections in one afternoon, both caught by other lanes rather than by me, and both of the
same family — seeing something that fitted what I already believed and not asking what else could
explain it. I would rather you saw that written down than tidied away. The upshot for the work is
small and slightly better than before: that build is now a *complete* twelve-stage example to
design the screen against, instead of a half-finished one.


## 2026-08-24, evening — the Builds screen is built

The screen you asked for — following the steps of each website build — is written and
committed. It is not on the console yet; there is an ordering step below.

Each site's card gets a "Builds" button. It shows the build as a numbered timeline in the
order the pipeline actually runs: research, strategy, briefing, site plan, composition,
design, pages, imagery, re-render — each stage with when it started, when it finished and
how long it took. Your apis.uk build reads as about an hour, stage by stage. Everything
else the site has been doing (the periodic checks, the image jobs) is kept out of the way
in a separate list, so the build is not buried in routine noise.

One honesty rule shaped it: on this platform a build step can fail and still be recorded
as complete. The screen never trusts that label — where a step left an error behind, a red
panel says so at the top, whatever the status column claims.

Two smaller things came along because they were cheap and worth more than they cost.

First: when someone hand-edits a section through the console and does not lock it, the
next rebuild silently throws their edit away. That has already happened 25 times on one
site. Sections that have suffered this now carry a warning badge, red if the section is
still unlocked — that is, still set to lose the next edit too.

Second: the Direction editor now tells you which kind of thing you are editing. Most of
those documents are instructions a writer reads — typing "never say X" there is a wish,
nothing enforces it. One of them, the evidence register, is enforced. They now carry
labels saying which is which. And I closed a genuinely nasty hole: a save to the register
that was valid JSON but the wrong shape used to silently switch the site's claims checking
off while reporting success. The console now refuses that save unless you explicitly
confirm it, and after every save it tells you the counts it actually stored — "40 banned
claims, 12 facts" — because a zero there is the one warning a broken save cannot fake.

The ordering step: the backend half must reach the cluster before the console half is
deployed, or the new screen would show misleading data against the old backend. The
backend rides the next core-manager roll; the console image gets built and deployed after
that. Nothing for you to do — whichever session deploys next follows the note in NOTES.

Also closed today: the reviewers approved the confirmation-link guard from the 23rd, first
round. And one correction to something I told you this morning has already been recorded —
the www redirect works but not because your redirect rule was applied; a piece of code in
front of the whole zone does it. If you apply that rule anyway, expect no visible change.

## 2026-08-24, night — the links address is live, and it held up to prodding

You applied the box steps and the DNS record this evening, and I tested it from the
outside rather than taking the clean terminal output on trust. Everything behaves as
designed: random paths get nothing, malformed links get nothing, a properly-shaped link
travels the whole way through to the cluster and comes back with our page, and hammering
the address gets you blocked by Cloudflare after a handful of requests — I watched the
blocking kick in. The admin console rode out the restart without trouble, and the old
shopfront address still parks exactly as before.

So the customer-links address exists, is deliberately public, and is wrapped in every
layer we planned. Nothing can be confirmed through it yet — there are no live links, and
the button-press page you asked for gets built before the first email goes out.

Two things wait on other events: the new Builds screen still needs the cluster's next
routine software roll before it appears in your console, and the reviewers owe us a pass
over the overall exposure picture now that this second public door exists — that review
fires once the cluster access key is refreshed (it expired on schedule this evening; the
usual three-day cycle, yours to renew whenever convenient).

## 2026-08-24, late — your new software roll delivered the Builds screen

The roll you ran this evening carried both halves of the new screen — the cluster side
and the console side — in the right order, so there is nothing left to deploy. I checked
it the thorough way: asked the running service which code it was built from, and pulled
the actual page bundle the console is serving to confirm the new screen is inside it.
It is. Open the console, pick a site, press "Builds" — your apis.uk build should read as
about an hour, stage by stage.

One small favour while you are in there: open any months-old stuck workflow and press
Terminate. It should say it worked. Until today that button always failed with a server
error — it was writing to a database table that never existed — and your click is the
last proof the fix behaves in real life.

I also filed the exposure review the architecture reviewers asked us to trigger once a
second public door existed. It puts three questions to you when you have a moment,
written in plain terms in the review file — the biggest one is whether the customer-links
door should get its own dedicated entrance in the service itself, so that even a
misconfigured relay box could never expose anything beyond the links.

Next piece of building work is the button-press confirmation page — the thing you ruled
must exist before any customer email goes out.

## 2026-08-25, morning — re-checked after the overnight roll; still live

Another software roll went out about half an hour ago, so last night's proof had expired
and I ran it again rather than repeat it. Both halves are still there: the console is
serving image v1.0.1337 and the page bundle inside it still contains all five of the new
screen's fingerprints, and the cluster service is running code that contains the Builds
backend. I also ran each check with a deliberate dud alongside it — a phrase that must
not be found, and the ancestry test run backwards — so a check that always says yes
would have been caught.

Worth knowing for next time: the roll changed the console's version number but not the
page bundle itself, which still has the same name as yesterday. That is expected — the
name is derived from the contents, so an unchanged screen keeps its name through a new
release. The version number moving is not evidence the screen changed, and the name
staying is not evidence it didn't ship.

Nothing else moved overnight. Still owed: the button-press confirmation page, and your
three questions on the exposure review.

## 2026-08-25, midday — the button page is built

The thing you ruled must exist before any customer email goes out is written, tested and
sent for review. It works the way you asked: the link in the email now opens a page with a
single button and changes nothing at all, and only pressing the button records that the
customer has moved. A mail scanner opening the link cannot confirm anything, because
scanners follow links and do not press buttons.

Two small choices you made this morning shaped it. The page does not look the link up at
all, so opening it cannot even tell someone whether a link is real, and the page names no
site or customer, so a forwarded email gives nothing away. The cost, which you accepted, is
that someone whose link has already been used finds out when they press rather than before.

One thing worth telling you because it nearly slipped through. I test these guards by
deliberately breaking the code and checking the tests notice. The first break I tried — make
the page quietly do the confirmation anyway — did **not** get caught, because the page looks
identical either way, so nothing a browser sees could reveal it. I rewrote the test to check
the thing directly, and the same break now fails it. The tests were green before that and
would have stayed green while the hole was open.

It is not live yet. It goes out with the next routine software roll, and the first customer
email stays blocked until I have checked it running, not merely committed.

## 2026-08-25, afternoon — a suspected hazard checked out, and it is not one

Yesterday's notes left a warning for whoever picked the work up next: that one of the ways
agents write a site's "evidence register" — the list of facts they may state and claims they
must never make — could wipe that list wholesale by accident. It looked serious enough to fix.

I measured it before building anything, which yesterday's note also asked for, and the answer
is that it cannot happen the way described. Nothing in the system actually writes the evidence
register through that route. Not one of the twenty places that could, and none historically
either. And the nightly refresher that touches those registers most — two hundred and
twenty-two times in the record — has never once removed anything from one. Its own code says
it never will, and the history agrees with its code.

There is exactly one occasion in the whole record where a register was emptied. It lasted
fifty-nine seconds, and the same session that emptied it put it back. It was somebody doing a
change in two steps, not a fault.

So I have closed that item rather than build a guard nobody needs, and written down why, with
the numbers, so it does not get re-raised in three weeks.

One real thing did come out of it, and it is not what the warning said. Eight of our nineteen
sites are not covered by the refresher at all — including apis.uk and webdesign.uk, the two
with the largest registers. Nothing rebuilds those if they are ever damaged. And the only way
a register has ever actually been emptied is somebody running SQL by hand, which no amount of
code can prevent. That is a habit to keep rather than a thing to build: hand-written database
changes touching those registers should check themselves before they finish.

## 2026-08-25, evening — your release went out, and both things are live

The confirmation page is running. I checked it the thorough way rather than trusting the
version number: asked the service which code it was built from, confirmed all three pieces are
in there, and then actually opened the page from inside the cluster. It shows the heading, the
one button, and nothing else, and it does not record anything when opened. I also pressed the
button's route with a made-up link and got the "no longer active" page, which proves the whole
path works end to end without anything being confirmed.

The other thread's containment change went out in the same release, and the two fit together
properly: the customer links now live on their own port, and the admin API is genuinely absent
from it. I checked that as a pair, because "the admin API is not on this port" only means
something if you also show it IS on the other one. It is.

Your Builds screen changes are live too. Refresh and you should see the shorter orchestration
list, with a link to show the older ones, and a line telling you when nothing is running, which
is why no Terminate button appears.

**One thing is left and it is yours:** the links host config needs applying on the box, so it
points at the new port. Until you do, a customer link gets nothing. I confirmed that is the
current state rather than assuming it — and it is worth saying how, because it nearly fooled
me: every test I ran from outside returned the same "not found", including the one that is
*supposed* to. The difference was in the text of the reply, not the code. One came from our own
software, which told me the request is still reaching the old port; the other came from the
relay box itself. Same number, different fact.

Nothing is at risk while it waits, because no customer links exist yet.

**And one thing I have deliberately not done.** I have not tested an actual working link end to
end, because that means creating a real one and pressing it, which stamps a real site as
"customer has moved" and feeds the schedule that eventually takes their site down. Everything
around it is proven; that last step is yours to authorise, and it takes about a minute whenever
you want it.
