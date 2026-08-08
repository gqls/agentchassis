# Where we are — fragment blind spot (plain prose, append-only, newest at the bottom)

## 2026-08-06

We went looking for the next open bug nobody was working on. The picker that
worked was measuring how "warm" each bug is in the other sessions' live
transcripts, then confirming the coldest candidates by searching for the actual
function names a fixer would have to touch. Two bugs that looked free by every
other test turned out to have sessions actively working right next to them; the
function-name search is what saved us from starting a duplicate.

We landed on the last unowned piece of bug 071. The story of 071: the platform
detects broken links on every build, and for a while it detected them and then
shipped them anyway. Most of that is fixed now. The piece nobody took is the
"jump to a section" kind of link — `#pricing` at the end of a URL. Nothing in
the platform ever checks that the section a link jumps to actually exists, and
for a while nearly every such link on the estate jumped nowhere.

First finding: the damage has quietly healed since the bug was filed — today's
fragment links (66 of them) all work, mostly because pages got repaired by hand
and the writing agent got given a list of real pages to link to. But nothing
GUARDS this: the next page build can ship dead section-links again and nothing
would notice, which is exactly how it happened three times in a row on one site
in July.

The plan: teach the existing post-deploy link checker (which already runs, and
whose findings already get worked) to also check the section half of every
link, reusing the id-detection logic another check already paid to get right;
and add one sentence to the writer's instructions so it stops inventing section
links in the first place. No new check to schedule, no new queue to forget.
Deliberately NOT doing yet: blocking builds on this, auto-deleting dead
fragments, or making the platform stamp ids on every section — each is
recorded with reasons, and each needs its own decision.

Next: write the code and tests, prove on today's data that the new check would
raise zero false alarms (and one real one when we plant it), put it through the
council, commit.

## 2026-08-06 (later)

Done and committed. The check now has a fourth thing it looks for: a link whose
"#jump-to-section" half names a section that does not exist on the page it lands
on. It sits inside the link checker that already runs and whose findings already
get worked, rather than being a new check somebody would have to remember to
switch on — we have been bitten by that exact thing before. The writer has also
been told to stop inventing section links, since nobody ever gives it a list of
real ones.

Before shipping it I ran the new code over every page on every site to see what
it would complain about: nothing, on all 67 of today's section links. Then I
planted two broken ones into a copy of the same data and it caught both — a
"zero problems found" that has never been shown capable of finding a problem is
worth nothing.

The council approved it first time. One reviewer asked a question I had not
thought to ask myself: I had rearranged a shared piece of code, and it turns out
a third place uses it — the gate that refuses to publish a broken tool. Changing
that by accident would have been bad. I checked properly by running the old and
new versions side by side over 4,036 real documents; they agree everywhere. My
first attempt at that check was itself worthless (both versions agreed that
nothing was wrong with anything, which proves nothing), and only a guard I had
written into it beforehand stopped me reporting it as a pass. That is the third
time today that asking "could this test have failed?" changed the answer.

What is left: the new code only starts working when the next chassis image is
built and rolled — this is Go, so it is inert until then. After that roll,
someone needs to confirm it is in the binary and plant one broken link on a real
site to watch it get reported. The steps and exact commands are in the lane's
RUNBOOK and handoff.

Three related things are deliberately NOT done, each written down with why: the
pre-publish gate still cannot check section links (it cannot see the site's
header and footer, so it would raise false alarms); nothing repairs a broken one
yet; and no page section publishes a stable name to jump to, so these links can
only be avoided rather than made to work. That last one is the real capability
and it needs its own decision, because it changes every page on every site.

## 2026-08-06 (after the roll)

It is live and it works. The build that went out this afternoon carries it — I
checked the actual running binaries rather than trusting the version number, on
both of them, including a control string that should be there and one that should
not.

Then I proved it properly instead of just looking at an empty queue. An empty queue
would have been the *expected* result, because none of our real section links are
broken today, so it would have told me nothing. So I built a page with four section
links — two deliberately broken, two working — on an internal placeholder site that
nobody's browser can reach, and ran the real checker at it. It complained about
exactly the two broken ones and said nothing about the two working ones. Then I
fixed one of the two, left the other, and ran it again: one complaint, the right
one. That is the difference between "the code is in there" and "the code does the
thing". The test page is deleted and I confirmed the site is back as I found it.

A nice accident: the same test page had a "#" link that goes nowhere, and a
*different* checker picked that one up instead of mine. That is exactly the division
of labour I designed, so one test proved both that mine fires when it should and
that it keeps its hands off the neighbour's territory.

One loose end, written down: the bit that double-checks a fix before closing a
problem hasn't been run yet — its database queries are verified but the code around
them only runs when a real one of these gets fixed, and forcing that would have
meant kicking off a page build I didn't want on a placeholder site.

There is now a five-part summary in the lane you can read out to someone, and the
handoff is marked up so a new session can see at a glance that the roll-time checks
are done rather than repeating them.

## 2026-08-07

A correction to what I told you yesterday, and it is the kind worth hearing.

I said the new check "cannot land inert" because it rides a checker that is already
switched on — as opposed to the bug we have on the shelf whose fix is perfect and has
never once run. The switched-on part is true. What I did not check is **how often the
thing that runs it actually runs**, and the answer is: by hand, on about nine days out
of the last twenty-one, on a handful of sites each time, most recently two days ago.
The fleet-wide sweep that would drive it properly has been off since May.

So the honest position is: the check is live, it is correct, I proved it works by
planting broken links and watching it catch them — and it has still never run on a
real site. There are no complaints in the queue, and that means "it hasn't looked",
not "everything is fine". The reassuring number — all 67 real section links on the
estate work — comes from me running the code over a copy of the data offline, which is
the figure to trust and to quote.

Nothing is broken and nothing needs doing about it. I have deliberately **not** started
firing off checker runs to make the queue look busy; the check will run the next time
any lane dispatches that checker for its own reasons. What I have done is correct the
overclaim in all four places I had written it — including the platform's own debugging
guide, where I had offered it to other sessions as a lesson — and logged it as a wrong
call. It is a slightly embarrassing one: the bug I was carefully avoiding is described
in its own file as "a mechanism made correct and then guarded behind something that
never runs", and I reproduced it one level up while citing it.

Two loose ends remain, both small and both written down: the piece that double-checks a
fix before closing a problem still hasn't run, and neither has the checker itself on
real pages. The lane's handoff now leads with all of this so whoever reads it next
cannot make the same inference I did.

## 2026-08-07 (later)

Picked this back up cold. First job was to check nothing had rotted, and nothing
had: the check is still in both running copies of the service, I read the version
off the running pods rather than trusting the build file, and I used a string that
should be there and one that shouldn't so the check itself is proved honest. Still
no complaints in the queue, and still for the reason I explained yesterday — the
thing that runs it hasn't been run since two days before it shipped. So both loose
ends are unchanged, and neither is something I can *do*; they are things to watch
for.

The interesting thing today came from somebody else. Another session filed a bug
this morning about the machinery my check plugs into. The shape of it: when two
different things file the same kind of problem, but only one of them wrote the
double-checker that decides when it's fixed, the other one's problems get graded
against the wrong question — and because the double-checker is perfectly correct
about *its* question, it says "fixed" and the problem closes untouched. Four days
and no repair, in the case they found.

Since I added the newest double-checker to exactly that machinery last week, I
checked whether mine has the same hole. **It doesn't** — only one thing files my
kind of problem, and I confirmed that three ways rather than one. But I also found
that mine is safe by luck rather than by design: the label saying who filed a
problem is just a free-text setting, so anyone could start filing under my name
tomorrow from configuration alone, without touching any code, and my double-checker
would grade their problem against my question and wave it through. Same trap, one
step away.

That also undercuts one of the fixes the other session was considering — having
each double-checker list who it speaks for — because the list of who *can* file
lives in the database, where code can't see it. Worth them knowing.

I have **not** written this into their bug file. They were actively typing in it
when I looked — it isn't even saved into version control yet — and adding my bit
would likely have destroyed one of us. So it's written up in my own lane, flagged
clearly, and it's yours to route: either I hand it to them, or whoever picks this
lane up next adds it once their file has settled.

## 2026-08-08

A new build went out. I checked the running copies again — the check is still in
both of them, with the same honest-string test as before, and that is now three
builds in a row with no damage. Still no complaints in the queue, still because
nothing has run it since two days before it shipped.

Then I stopped waiting on the one loose end and went and looked at it properly,
and I am glad I did, because **the reason I gave you for it was wrong.**

I had written that the part which double-checks a fix before closing a problem
couldn't be reached, because the machine that closes problems only handles one
category of work and ours is a different category. That is not true. I read the
machine's actual settings today rather than repeating my own note, and it has **no
category filter at all** — it takes anything. What actually holds our problems back
is something much more ordinary: a newly-found problem starts life marked "spotted",
and the machine only picks up ones marked "triaged". Nothing has promoted ours,
because none exist yet.

I also had a second thing wrong in the same note: I'd said our kind of problem
always lands in the awkward category. It doesn't — if the broken link is in the
page furniture (the header or footer) rather than the page body, it lands in the
*easy* category, the one where this machinery is already proven to work. So the
cheap way to test the last loose end was available the whole time and I had talked
myself out of it.

The good news underneath all this: somebody else's lane proved yesterday that this
double-checking machinery genuinely **refuses** to close a problem that isn't fixed
— which is exactly the failure I was worried about, and it is now demonstrated in
production rather than assumed. Eleven of these double-checks have run across the
estate. None of them in our category yet, which is worth knowing.

One more thing I want to flag rather than bury. The estate-wide sweep that would
drive all of this has been described in our notes — including mine — as "switched
off since May". The record has two dates in it that disagree by three months: it was
last *started* by the scheduler in May, but it last *finished* three days ago. I
could not establish what ran it, and I am explicitly not guessing. I tried one way
of checking, got seven results that looked like proof, and they turned out to be a
different system entirely that merely mentioned the name. So: the May date is not
the whole story, and I have marked it unresolved rather than replacing one confident
claim with another.

All of the above is written into the lane's notes, the handoff, and the fleet-wide
log of wrong calls — that last one because the shape of this mistake is worth other
people knowing: a stated *reason something can't be done* is the claim least likely
to ever get re-checked, because it explains something everyone can already see and
it recommends doing nothing.
