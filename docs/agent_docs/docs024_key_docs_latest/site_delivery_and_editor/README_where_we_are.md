# Where we are — site delivery + customer editor

Owner's plain-prose log. Append-only, newest at the bottom.

---

**14 Aug 2026.** The idea: as well as the preview, publish each finished
customer site as a real hosted copy, and give the customer an editor so they
can change their own site after handover. A reference document proposed doing
this with a new central engine and Netlify; the review concluded our platform
already IS that engine, and that the automated route is Cloudflare Pages
(Netlify's ownership hand-over needs a manual dashboard step, which fails the
"completely automated" bar; and since we keep the ability to update sites,
handing the hosting away doesn't fit anyway — what the customer owns is the
ZIP, as the £149 terms already say). The customer's editor will be one small
service on our existing box; customers get in through an emailed link after
handover; their edits flow through the same machinery our own edits use, so
the existing locking system referees between their changes and ours. Six
phases, each reviewable on its own. The first (the "usually next day"
promise on webdesign.uk) went live the same evening.

**15–16 Aug 2026.** Phase 2 is built and working: finished sites can now be
published to a real hosted copy automatically, and we have watched it happen.

How it works, in plain terms. Each site has a switch that is off by default —
nothing changes for any site until someone turns it on for that site. A
timer wakes up, picks the site that has been checked least recently, and asks
one question: has anything about the built site changed since we last
published it? If nothing has, it stops and does nothing. If something has, it
copies the site to its hosted address, then — and this is the part that
matters — it fetches the page back from the public address and checks the
bytes match the original. Only if they match does it record the publish as
done. So we never mark something as published on the strength of the copy
"looking like it worked".

It did not work first time, and that is the useful part of this story. The
first live run failed on the very first file: the storage service refuses an
upload unless it is told the size in advance, and our code was streaming files
without knowing the size. Nothing was half-copied and nothing was falsely
recorded as published — the failure was clean, which is what the design was
for. We fixed it, and the fix was proven by deliberately putting the old
mistake back to confirm the test caught it. Yesterday that same site failed on
file one; today it copied all eight files and the published page is
byte-for-byte identical to the original. Ten minutes later the timer came round
again and correctly did nothing, because nothing had changed.

Two other things worth knowing. The reviewers turned the work down on the first
pass, and they were right: the way I was taking a safety backup of a shared
setting would have quietly dropped part of it, so a restore would not have
worked. Fixed and re-submitted, approved second time. And once it was running I
noticed each check was starting a whole new worker just to answer "nothing has
changed" — about a hundred and fifty a day for one site — so I slowed the timer
from every ten minutes to every hour. A finished site changes a few times a
day at most, so nobody will notice the difference, and it is one line to put
back if you want it faster.

Where that leaves us: the publishing half of the delivery promise is done. Next
is the ZIP file the customer actually owns, then the handover step, then the
emailed login, then the editor itself. Still waiting on you for two keys: the
Stripe ones for taking payment, and a Cloudflare one if we later want the
sites hosted there rather than on our own storage — the current route needs
neither, so nothing is blocked.

**17 Aug 2026.** We spent today deciding, not building — five rounds of
discussion, and the shape of the whole delivery business came out the other
side. Written down properly in the plan file dated today; here is what it
means in ordinary words.

What a customer will experience: they ask for a site and, while they wait for
it to be built, we invite them to spend two minutes creating their own free
account at Netlify (a big site-hosting company) and clicking one Connect
button. If they do — and we expect most will — their finished site is
delivered straight into an account that belongs to them from the first
moment. Their hosting, their bandwidth, their bill if they ever outgrow the
free tier; nothing for us to run. If they skip it, no harm: the build and the
sale go ahead regardless, the site is always viewable on our preview address,
and their own domain shows a friendly "choose where your site should live"
page until they pick — their free Netlify, our own hosting, or just taking
the files away. We deliberately priced our own hosting high, because we don't
want to be a hosting company; the page always shows the free option next to
it.

Money: two separate, simple charges. Keeping the domain name we chose for
them is £10 a month — we register it ourselves at trade price, since we're a
Nominet registrar, so this is good margin and completely automatic. Hosting
with us is the expensive option almost nobody should take. Everything is a
link in an email: the payment links, the file download, and a page run by
Stripe where customers manage their own subscriptions — we build no billing
screens at all.

One technical decision worth saying aloud: we will run our own nameservers —
the machines that answer "where does this domain point?". Because we register
every customer domain, that answering job is ours whichever company hosts the
site, and doing it ourselves means moving a customer between "our hosting",
"their Netlify" and "still choosing" is a one-line change on our own
machines, forever, with no outside company able to hold it hostage.

Parked for later, on purpose: giving paying hosting customers extra content
like news feeds (your idea, kept, not now); the big review of how everything
scales, including whether we need more clusters; and what to do about sites
that get seriously busy. None of these block anything.

Next: back to building. The ZIP file the customer downloads is the next piece
— it is also exactly what gets uploaded to their Netlify account, so one
piece of work serves both doors.

---

2026-08-17, later. The ZIP is built. This is the file a customer actually
gets: their whole site in one download. The code that makes it now exists,
is tested, and is committed — it gathers every file of the built site,
packs them into a single archive, checks its own work twice (once that the
archive really contains every file byte-for-byte, and again that the copy
that reached storage is exactly the size it should be — a shortened
download that looks fine until you open it is the failure we most wanted
to design out), and then produces a download link that works for seven
days. Very large sites raise a flag for us to look at rather than ever
producing a cut-down archive: the customer gets everything or we get told,
never something in between.

Two things worth saying plainly. First, nothing runs yet in production:
the new code must ride the next release, and then a small piece of
configuration (written and waiting, with its own checklist) switches it
on. Cutting a ZIP happens only when asked — there is no schedule, because
nobody wants a fresh archive every hour; the delivery email (next phase)
is what will ask. Second, the review council has been asked to look the
work over, as we do for platform changes; the commit carries the
submission's reference so the review trail joins up on its own.

## 2026-08-25 (evening) — the door between the internet and the admin API is now in the code, not in a config file on a box

Picked this lane up cold today. Nobody else had it open, though the webdesign.uk
threads hold pieces of it and I have stayed out of their way.

**The thing I found and fixed.** When a customer clicks the confirmation link we email
them, that click goes to a small web server on a box, which passes it into the cluster.
Until today it passed it to the *same door* that serves the admin system — the one that
holds every site's data, the job queue, the controls. The only thing keeping the admin
system off the internet was one line of pattern-matching in a config file on that box.
Widen that line by a single character and the whole admin API would have been public,
and nothing in our software would have noticed or objected.

So the customer links now go to their **own door**, which has nothing else behind it.
If somebody widens that line tomorrow, all they can reach is the customer links. The
owner asked for this on 25 August after an architecture review; this is that job done.

I also made it something the software *enforces* rather than something we remember: if
anyone ever puts a customer link back on the admin door, core-manager refuses to start.
That is deliberate. A warning in a log would be read by nobody while the hole served
traffic.

**The mistake worth telling you about.** The switch that turns this on is set in the
deployment config. It would have been ignored — silently. The configuration library we
use only reads environment settings for keys it already knows about, and this was a new
one. Everything would have looked perfect: the setting visibly there, the code correct,
the tests passing, and every customer link dead with no error anywhere to find. I caught
it because I wrote the test before trusting the wiring, and it failed. That failure is
the only reason this is a paragraph and not an outage. I have written the trap up so the
next person adding a setting like this does not repeat it.

**The review found two more things, and both were mine.** I put the change through the
review council. It came back "revise": the reviewers pointed out that my write-up never
showed the two changes that make the thing actually work — starting the new door, and
pointing the box at it. I *had* made both; I just left them out of the summary because it
has an eight-item limit. That is a fair hit and it is the second time this lane has been
caught by the same habit. They also found a genuine hole: my safety check compared exact
addresses, so a catch-all route would have slipped past it. Fixed and re-submitted.

**One thing I need to flag, because it is a promise to customers.** The register now says
the site stays live for **30 days** — that is your copy brief from today. The code keeps
serving it, and keeps the download link alive, for **42 days**. We agreed to keep that gap
deliberately: promise 30, serve 42, so nobody is cut off at the exact moment they were
told. Nothing recorded that anywhere, so the next person would have "fixed" the mismatch
and quietly removed the cushion. It is now written down in the code and defended by a test
that fails if someone closes the gap.

**Where this leaves delivery.** The confirmation page and this new door are both built and
both waiting on the same thing: the next core-manager release. Until that happens neither
is live, and no delivery email should go out — that is your ruling, and it still holds.
Nothing is at risk in the meantime: no customer has been handed over yet and no links
exist.

## 2026-08-26 (afternoon) — the delivery email explained to the owner: the decisions, the draft, and what is still open

The owner asked to see the email decision, his own rulings, and the email itself. Given in
chat and recorded here. The short version: the email is designed and its plumbing is live,
the words are deliberately unwritten, and three things gate sending — his pre-send review
(his 21 Aug ruling), the queue item that feeds that review (not built yet), and his two
open product questions from 25 Aug, which could change what the email promises.

A draft was shown assembled from his attested register lines (30 days, £10/£200, the ZIP,
keep-it-online), with empty slots marked where mechanisms do not exist yet (the ZIP
download link, the Stripe portal, the contact address, the terms page). It is a draft for
his eyes, not copy on the wire: nothing sends until the review gate exists and he approves
a real site through it.

His open product questions were listed back to him: (1) owner edits before a customer sees
a site; (2) customer self-editing during the 30-day window, possibly by voice; plus the
smaller owed calls — contact address on a domain we own, the terms page, Stripe timing,
the second Nominet TAG, and the two Cloudflare facts the DNS plan needs (the real zone
cap; the CF-for-SaaS tier).

## 2026-08-26 (late night) — everything checked again from cold: the last review verdict is in, and what remains is your three small steps

Since the afternoon entry above, a lot landed in one evening: you ruled both open
questions (you get an edit pass the customer never sees; no customer editing at launch,
voice editing is the next build), the whole delivery machinery was finished and went
through review, tonight's deployment picked it up, and the three delivery agents are
installed and switched on. The email account is fully proven — a real test message
passed all three authentication checks at Gmail itself.

A fresh session then re-checked everything from scratch tonight. The one outstanding
review verdict came back **approved** — the reviewers' single caveat (a safety helper we
added needed something to actually use it) was already dealt with by the review-filing
agent installed the same evening; I checked the live configuration to be sure, not just
the file.

Nothing else has moved, which is as expected. No customer has been handed anything, and
nothing can send yet. The whole critical path is now three things only you can do:

1. Create the mail password secret on the cluster (the one-line command in the handoff —
   you hold the password, it lives in no file).
2. Re-apply the links box configuration so the download link route works from the
   internet.
3. When we rehearse the full flow on a site of our own: edit the site if you like, then
   press **Approve** on the review item — Approve, not Resolve.

The mail settings themselves ride along automatically with the next normal deployment —
nothing for you to do there, and until then a send attempt fails loudly and harmlessly.

## 2026-08-27 (morning) — the 502 you saw, found and fixed; two more of the remaining steps done in passing

The bad gateway on preview was nothing in the cluster: the little server box that fronts
all the webdesign addresses had its web server die at 6:22 this morning. An automatic
overnight software update restarted it, and at that exact moment the private line to the
cluster didn't answer, so it gave up and stayed down — while the tunnel in front of it
stayed healthy and answered everything with "502". I started it again (everything came
straight back) and installed a guard so that if this ever happens again it retries every
fifteen seconds instead of staying dead.

While I was in there, two of the outstanding steps got done. The mail settings reached
the pods on their own — last night's routine deployment carried them, so nothing more is
needed there. And I applied the download-link route on the box myself (it turns out
sessions can do the box steps — that was corrected in the runbook last week), so the
customer download links now answer properly from the internet.

One more catch: the "Not active yet" notice on the shopfront had silently vanished — the
design system rebuilt the front page yesterday evening and dropped it, exactly as
predicted. I put it back and checked the live page shows it again, in both places.

So the list only you can do is now: create the mail password secret (one command), and
when you're ready, remove the Cloudflare rule that parks the site. Then we rehearse the
whole delivery on a site of our own.

## 2026-08-27 (mid-morning) — mail secret in; terraform answered; the Lovable line handled the proper way

You created the mail secret; the quiet-moment deploy you'll run is what carries the
password onto the pods. On your terraform question: no — leave it where it is. The
Stripe keys got wiped because terraform owns that particular secret wholesale; it
doesn't know this new one exists, so a release can't touch it. Moving it in would also
mean the password living in a file on your machine, which we deliberately avoided.

The Blueprint Compiler change went through the framework, not the HTML: I filed the
standard "improve this tool" instruction that the platform's own checks use (it has run
successfully three hundred–odd times). The instruction removes the Lovable mention — and
the v0 one, which is the same kind of recommendation one step earlier — and makes the
prompt deck neutral about which AI builder it's pasted into, changing nothing else about
the tool. One item covers the page you saw; a second covers the master copy in the tool
library, queued so it can't collide with a repair already waiting on that copy. I'll
confirm at the served page once they've run.

## 2026-08-27 (late morning) — you changed your mind on Lovable, and it was caught in time

Both filed instructions were cancelled before the platform picked either up — nothing
ran, and the page never changed: the v0 and Lovable mentions are still there, verified
on the live preview. Your reasoning (we're a different service, and may positively
recommend such tools when we don't suit) is recorded so no future session "tidies"
those references away.

## 2026-08-27 — the shopfront is live, and delivery's turn is next

The site went live late morning (details in the webdesign lane's log — including the
one wobble, a timed-out first minute caused by address records that had never been
exercised). For this lane it means the last external thing delivery was waiting on has
happened. What's left: your quiet-moment deployment to carry the mail password onto the
pods, then the rehearsal.

## 2026-08-31 (evening) — your review of boxingonline, and what each thing turned out to be

Your address is off the site's database everywhere, and I rebuilt every page so the footer
stops carrying it. The public copy of the site is a mirror that refreshes on the hour, so
there is a gap of up to an hour where the old page is still out there — the other session
working this site is watching that through and will confirm when the served page is clean.
I would rather tell you that than say "done". The underlying cause is worth knowing: the
email you ordered with and the email the site publishes are the same database column, so
this would happen to the next customer too. That is being filed as a bug.

On the copy reading like a description of the site rather than the site: I think I found
where the words came from. Before we build, we research the best sites in the field and
write down what makes them good — for this one it said things like "vague fight previews
that say 'this could be a great fight' add no value". That research is meant to be
instructions to us. It came out as sentences on the about page. So the writer was handed a
list of rules about writing and published the rules.

The articles page having nothing on it but our own standards is the same emptiness from a
different angle: the plan never produced the six articles your brief asked for, and rather
than the page looking broken, the writer filled the space with policy and the link checker
quietly repointed the dead links. Everything passed. That is the part I find most worth
saying out loud — nothing on this site failed. It all passed.

You asked whether we have a quality auditor. We do, it is alive, it ran forty-nine times
recently, and it ran zero times on this site. It is not part of building a new site.

On imagery: every page has exactly one picture and it is the logo. We can generate
infographics — the whole path is built and works — and across every site we have ever
planned, an infographic has been asked for once. You also asked for pictures inside
articles on the 13th of August; there is a good plan for it written on the 14th that still
says "nothing implemented".

And your question about being best in the vertical: the research knew, and it was good. It
said this site's opening should be owning "what to watch this weekend" — a weekly guide
pulling every card together. The strategy document repeated it. The plan then built six
pages and none of them was that. The phrase "best in class" appears in none of this site's
specifications. There is already a plan to fix that, written six days ago after your last
ruling, and it is waiting on your go-ahead.

Two things I need you to decide: whether this particular site gets fixed before its
delivery email or delivered and then improved, and whether that best-in-class plan goes
ahead. Full detail, with the measurements, is in
OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md
in this directory.

## 2026-08-31 (later) — your address is off the site, and I was wrong twice on the way there

It is gone. Every one of the nineteen live pages checked one by one, each with a word that
had to be there so a blank answer couldn't fool us, and the delivery session checked the
same nineteen separately and got the same result. It is also gone from all four places it
was stored in the database, and nothing will put it back.

Two things I told you that were not true when I said them, both worth writing down.

The first: I said the database was clean. I had checked three places it could live and
there was a fourth — the footer, which every page shares. The rebuild I ran refreshed the
top of the page and the page header and quietly skipped the footer, so my check passed and
your address kept being published for about forty minutes.

The second: once I had fixed that, I set a watcher on one page. It said clean. A check of
all nineteen pages, ten seconds later, found six still carrying it. The single page was
the flattering one.

Both mistakes are the same mistake. I decided what to look at, looked at all of it, and
read "all of what I chose" as "all of it".

There is a third thing I want to record because it is the one I would most like to avoid
repeating. For that whole forty minutes my simple check kept telling me the address was
still there, and I kept explaining it away as the public copy lagging behind. So did the
other session, independently, which made it feel like agreement rather than two people
making the same guess. The check was right the entire time. The evidence that would have
settled it was sitting in the page's own headers, showing the file had just been written
and still contained the address — which is the opposite of lagging.

Two genuine faults came out of it and are now filed: the rebuild that skips the footer,
and the fact that changing something every page shares does not make any page look
changed, so a whole-site rebuild does nothing and still reports success.

Two things still open on this: the footer is currently a hand-patch and should be rebuilt
properly before anything is sent to a customer, and the contact page is now a form that
does not go anywhere, on a page you never asked for. Delete it or give it somewhere to go
— that one is your call.

## 2026-08-31 (evening, later still) — two more things on boxingonline, one of them a trap

The logo. I downloaded the image the site is actually serving and looked at it. It is
lettered "BOXING NEWS". Your site is Boxing Online — that is the company name, the order,
the domain, and even the alt text on the image itself. Nothing anywhere calls it Boxing
News. The instruction we gave the image generator told it to letter a wordmark and never
told it what the wordmark should say, so it invented one. That fault already had a bug
file open from another site, where it produced "Farm Shield Info"; yours is the second
confirmed case and the first on a site someone paid for.

There is a second thing wrong with the same file, which I think is worse in its way. It is
not a logo at all. It is a two-panel presentation board — the mark on a dark background,
then the same mark again on a light background with the lettering beside it — and it is
being squeezed into the header slot as though it were a single logo. Nothing between the
image generator and the page noticed. That now has its own bug file.

The instruction has been corrected, so regenerating will produce a clean, text-free mark.

And the trap, which matters more than either. We must not rebuild this site from scratch
until a fix has rolled out, because the order record still holds your email address, and
the code that copies it across only does so when the field is empty — which it now is,
because we emptied it. So a rebuild would put it straight back. The nasty part is that
every check we ran tonight would say the site was clean right up until the moment someone
pressed rebuild. It is written down in the traps file and both other sessions know.

I mention it because "just rebuild it" is the obvious thing to reach for while we sort out
the remaining items, and for the next little while it is the one thing that would undo the
work.

## 2026-08-31 (close of evening) — where boxingonline actually is

Measured in one pass at the end, not carried forward from earlier: every published page is
clean of your address, every page has the Fight Calendar in its menu, both listings carry
your six articles and no explainer guides, and the logo is the regenerated text-free mark.

What got fixed tonight, in the order it happened: your address off the site and out of all
four places it was stored; the six articles your brief asked for, built; those articles
actually linked from the news page and the home page; the logo regenerated after it turned
out to say "BOXING NEWS"; and the fight calendar put into the menu, which it had never
been in.

What is still yours to decide, and none of it is urgent tonight: what to do with the four
explainer guides now that nothing links to them; whether the site's name should sit beside
the logo; whether the contact page should go or be given somewhere to send messages; the
colour scheme; and whether the guide-typing question gets settled across the whole estate
or just here.

Two things that constrain anyone working on this site. It must not be rebuilt from scratch
until a fix has shipped, or your address comes back. And the footer cannot currently be
regenerated at all — the reason was found tonight by watching it fail live: the page
builds fine and then the database rejects it because a character has been cut in half
somewhere, and the code notes this and reports success anyway. Until that is fixed the
footer on your site is the one I edited by hand, which is the only version that exists.

The honest summary of the evening is that almost nothing here was broken in a way that
announced itself. Every page passed every check, all night, including while it was
publishing your personal email address. What found things was looking at the actual served
pages, one at a time, with something in each check that had to be true so a blank answer
could not pass for a clean one.

## 2026-09-02 — your fourteen points, and the one number that explains most of them

All fourteen have an owner and are moving. The short version of what turned out to be true.

The articles genuinely have no news in them, and the reason is worse than bad writing: the
site already had the real story. Your news page carries Hrgovic stopping Itauma for the IBF
heavyweight title, dated 31 August — which is an underdog beating a champion — while the
article titled "Last night's result: the underdog shocks the champion" was written about
Buster Douglas in 1990.

The pictures for your articles already exist. Six of them were generated on 1 September and
are sitting on the server right now. The component that builds an article page has one slot
and it holds text, so it cannot display an image whatever we generate. That is why the
cards have pictures and the articles they lead to do not.

The cards were missing their summary line because the summary was already in the row under
a different name. Two pieces of code that both fill the same card disagree about what to
call things, so the template asked for something nobody writes while a perfectly good
sentence sat unused beside it.

The guides index we built for you today listed your six articles instead of your four
guides. That is the third time in four days that a list has been filled with the wrong kind
of thing — and this one was caused by the fix for the first one. Fixed, without undoing the
earlier fix.

And the number that explains the calendar, the comparison tool and a good deal else. The
system has a store for checked facts about a site. Twenty of our fifty-four sites have one
at all. Yours holds three facts. Forty-two of the fifty-four hold five or fewer, and one
site holds most of what exists. Three pieces of code write to that store: one puts a stub in
when an order arrives, one refreshes facts that are already there, one records sources for
things already written. None of them goes and finds out what is happening in the world.

So when you said the research agent should have researched what is on, and that this is what
should have appeared on the calendar — that is exactly right, and there is no such agent.
It is filed as bugs_open/427 and it is the one piece of work with nobody assigned to it.

One correction of mine. I told you that fact store was used by 444 sites. That was a count
of database rows including old versions, repeated from another thread without my checking
it. The real figure is twenty sites. It was caught by the session that went and re-derived
it instead of using my number.

---

2 September, evening. The guides index now lists the four guides. This was the item we
told you was stuck until the next software release, and that turned out to be wrong — the
capability we needed has been in the running system all along, under a general name rather
than the specific one we went looking for. What was genuinely true is that the page's list
definition is shared with fourteen other pages across seven sites, so we could not simply
edit it. Instead we made this one page its own private copy of the definition — identical
in every respect except which kind of page it lists — and rebuilt the page. The build came
back with exactly the four guides, each with its one-line summary underneath and a clean
title, and the heading now matches what is beneath it. So the bridge you chose — people
reach the guides through the guides index — is in place, without waiting for a release.

Two things worth saying honestly alongside that. First, the rebuilt page also settled a
question another thread was chasing: the card-summary fix does work on the current system;
the earlier pages that missed it were built in a window where one of the two servers may
have been running stale code. Re-rendering the homepage would confirm that cheaply. Second,
while filing the rebuild I found that a hand-filed job can sit in the queue for ever,
silently, if two bookkeeping fields are not set — mine did, for seven minutes, until I read
the claiming code. That trap is now written down where the next person will meet it.

The list of what still holds the delivery email back is shorter now: the transparent logo
(waiting on the next release, genuinely this time — verified at the running system), the
contact page returning a proper "not found" (bugs_open/429, unowned), and the calendar
having real events to show (bugs_open/427).

Correction to the entry above, same evening: the contact-page "not found" work
(bugs_open/429) is no longer unowned — you've routed it to its own thread, which is
now running. I've sent that thread what closes our half (a real "not found" served at
the public address) and a warning about the hourly publishing cycle we're both using
tonight, so neither of us misreads the other's publish.

## 2026-09-02 (late evening, fresh session) — the release landed, but I am NOT regenerating the logo tonight, and here is why

The release you built this evening is running, and I checked it the proper way: I asked
each of the two services involved what code they were built from, and both answered with
the same commit, which contains every fix we were waiting on. So the contact page's "not
found", the card summaries on the home page, and the analytics tag are all now able to
arrive. Each arrives on its own schedule without anyone pushing it, and I am watching the
live site for all three. The contact page should flip at the site's next publishing tick,
which falls a few minutes before the hour.

The transparent logo is different, and the previous note from this thread got it wrong.
It said the logo regeneration was now unblocked and ours to fire. The release condition
is met, but another thread had already run the new logo mechanism on three other sites
this afternoon and read the results pixel by pixel. What they found: the mechanism works
about one time in four, and the safety check that is supposed to refuse a bad result
cannot tell a good one from a bad one. It gave the same perfect score to a logo that came
out properly transparent and to one that came out entirely opaque. Three sites are now
serving an unusable logo that the system believes is fine.

Nothing about tonight's release changes that. The fix that went out corrects the wording
of the instruction to the image model; it does not touch the safety check, which is
unchanged in the code. And a generated logo cannot be undone: there is no "put the old one
back". Boxingonline's current logo, the solid dark mark whose shape you approved, is
correct and serving. A regeneration tonight would have roughly a three-in-four chance of
replacing it with something worse, on the site of our first paying customer, while
reporting success. So I have not fired it, and I have corrected the note that said to.

This needs you on two points. First, the thread that owns the logo mechanism
(bugs_open/424) needs to fix the safety check so it measures transparency rather than
whether the background was merely reached, and that means another release. Second,
whether boxingonline should be the first real test of the repaired mechanism at all, or
whether we prove it on one of our own sites first. Their own handoff already lists that as
your decision. Until one of those happens the interim logo stays, and delivery stays held
on your fix-everything ruling as before.

One practical thing: the cluster credentials expired partway through this session, so I
can watch the site itself but cannot read the database until you refresh them.

Correction, twenty minutes later (you refreshed the cluster credentials, thank you). Two of
the things I said above have moved. The card summaries on the home page will NOT arrive from
this release on their own after all: the components thread re-ran its test against the new
release at 21:22 and the home page's stored data still has no summaries, exactly as before.
The one route that does work is a full rebuild of the home page, and that thread now says
such a rebuild is safe for its investigation and is the best next step. It holds the
sequencing, so I have not fired it. And the logo safety check I described as needing a fix
has now been fixed and approved by review, about half an hour ago. It still needs another
release before it can reach the site, and after that your decision on where to test it
first still stands. The contact page's "not found" and the analytics tag are unchanged:
still expected at the site's next publishing tick, a few minutes before the hour.

## 2026-09-03 (morning) — the contact page is properly gone; the home-page rebuild you approved was refused by a safety check; and the analytics tag was never on its way

Overnight the contact page finally answers "not found" at the public address, with a normal
page still answering beside it, and the thread that owned that fix has closed it. That one is
done.

The home-page rebuild you approved last night did not go through. It ran three times and was
refused each time by a guard that stops a page losing more than half of a section's visible
text in one save. The reason is worth knowing: the "call to action" block on the home page
currently holds about a thousand characters of running news teaser, and the writer now
produces a short call to action of about a hundred and seventy. The guard cannot tell whether
that is the fix or the damage, so it refuses. I have not lowered the guard, because the knob
for that is fleet-wide. So the cards on the home page still have no summaries, and the only
real route left is the rendering defect the components thread is chasing; their next
experiment runs on the guides index instead, where it does not need a rebuild. The same
attempt also showed that the "featured content" block on the home page asks for data by a
name the system does not know, which is why that block has been empty; that needs whoever
owns that component.

The analytics tag and the cookie banner had not arrived because the refresh that carries them
was never going to run. The system re-detects that a site's header and footer need refreshing
whenever their inputs change, and a rule meant to stop a broken fix being retried for ever
treats a refresh that WORKED as a failed attempt. After two in a week the third is parked as
"unresolved" the moment it is filed. Boxingonline was parked on 2 September at 06:19, and
eleven other sites are in the same state, which also stalls the consent banner rollout
fleet-wide. I have filed this as a bug with the fix shape, told the analytics thread, and put
the refresh in by hand for boxingonline this morning. It is queued behind about three hours
of other work; I will report when the tag is on the live pages.

Nothing else moved: no new release overnight, so the logo position is exactly as I left it,
and delivery is still held.

## 2026-09-03 (mid-morning) — you said go on the logo, and it is running

The new release is live and carries the repaired safety check for logo backgrounds; I checked
that on every pod of both services. On your word I have asked the system to regenerate
boxingonline's logo with a transparent background. It went in at 09:24 and will take a little
while. The three other sites that got bad logos yesterday are being retried at the same time
by the thread that owns that fix, so we get four readings from one window. I will judge the
result by the image bytes and the safety check's own score together, never one alone, and by
eye for a single clean mark. If the check refuses the result, the current solid mark stays, and
that is still useful information rather than a failure.

## 2026-09-03 (late morning) — the guard that refused the home-page rebuild was protecting the copy you rejected; on your word I opened a one-run window

The boxingonline thread read the block the safety guard was defending. It is the home page's
"call to action": about eleven hundred characters that describe the news feed and then walk the
reader through four tools one sentence at a time. That is your first complaint from the
review, word for word. The rebuilt version is a hundred and sixty-seven characters, which is
why the guard refused it as a loss of text; it was a repair.

You chose the scoped override. I have written it as a recorded, reversible change: the floor on
the page BUILD step only, lowered from half to a tenth, taken at 09:53, with the rerender path
untouched and no other builds queued anywhere at that moment. The rebuild went back in a few
seconds later, and a monitor restores the floor automatically the moment the rebuild finishes,
whichever way it goes. When it lands you get the cards with their summaries and the shorter
call to action in one change, and the empty "featured content" block drops off the page
because its data source does not exist. I will read the new copy against your review before
calling it done.

The logo test is still queued behind three sites the logo thread is retrying. One of those,
seotools, has already come back right: a genuine transparent background, measured at the image
itself, after the repaired safety check correctly refused a first attempt. That is the first
real success of the mechanism.

## 2026-09-03 (late morning) — one thing needs you: the image model's prepaid credits are gone

Boxingonline's turn in the queue came at 10:42, and the logo regeneration you approved failed
before it could even try: Google's image model answered that the project's prepayment credits
are depleted. Until that is topped up in AI Studio, no logo can be generated anywhere on the
fleet, including the three portfolio retries. The platform parked the attempt without counting
it, so nothing is lost; it just waits for the credit. The logo thread has filed this as
bugs_open/455.

The rest of the turn: the site's header and footer were regenerated and the head now carries
the analytics tag and the consent banner, but the refresh then failed on a step that tries to
re-insert an unchanged blog listing, so the twenty page rebuilds it should have queued never
appeared. I have queued them by hand. The home-page rebuild failed its first attempt after all,
by ten seconds: the platform copies an agent's settings into a run at the moment it starts, and
my floor change landed ten seconds after that moment. The floor is now in place for the second
attempt at about 11:15, with an automatic reset at noon if nothing else closes it. That
ten-second lesson is written down where the next person will meet it.
