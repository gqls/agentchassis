# README — where we are (bugfix 311: the missing calculators)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-19 — picked up on your instruction; the diagnosis got sharper and the fix got smaller

You asked for the missing-tools bug (311) to be fixed properly, framework-wide. Here is
where it stands after the research pass.

The bug is still live — the same seven calculator sections failed again through last
night, and nobody else is fixing it (the lane that found it deliberately waited for your
say-so and a council round; this is that).

One important correction to the bug file's story. It said the components can't be reused
because the selector looks them up by a column (`section_type`) they don't have. That's
only half right. The build actually tries the right lookup first — by the component's
function name, which these rows DO have — and finds them. It then **throws them away
because they look broken to it**: they are hand-written calculator templates shaped like
tools, and the build's "is this template whole?" check expects section-shaped templates,
so it reads them as truncated and moves on. That's why it then asks for a brand-new
component, and why the new component collides with the old one and gets refused, forever.
I've put this refinement through the diagnosis loop to be independently checked before
building on it.

Why this matters: the bug file's second fix idea — "fill in the missing section_type
column" — would not have fixed anything, and for these particular rows would have made
the failure quieter and worse. So the fix is the first idea from the file, and only that:
**when a site's build wants to write a component whose name is already taken by a
component other sites depend on, it must stop trying to overwrite it and instead create
its own, under a name scoped to that site.** The new component is created in a way the
library can find and reuse — so the next site that wants the same calculator reuses it
instead of failing. One site can then never block another site's build again, which is
the property that actually scales to the 50-site programme.

What I'm deliberately NOT doing in this round, and why: (a) not touching the three
hand-written calculator components — they belong to loanandmortgagecalculator.co.uk and
repairing their shape is that lane's work, through the framework; (b) not building the
"refuse to deploy a page with missing sections" gate — that's a real gap (it's how this
shipped silently) but it's a separate change to a different part of the pipeline and
deserves its own round rather than riding this one.

Next: council round on the plan, then the code, tests, and a commit so it rides the next
chassis build. Verification will re-run one of the failed calculator sections on
loanzy.uk and check both halves: the new site gets its calculator, AND the old site's
component is byte-for-byte untouched.

## 2026-08-19, later — the fix is written, tested and committed; the council is reviewing it now

The code went in as one commit (`17d883333`). What it does, in one sentence: when a site's
build tries to write a component whose name is already held by a component other sites
depend on, it now creates its own site-named copy that the library can find and reuse,
instead of failing forever — and the old site's component is never touched.

The tests prove both directions: the colliding case creates the new copy (and we checked
the test really bites by deliberately breaking the fix and watching the test fail), and
the normal case — a site regenerating its own component — behaves exactly as before.

Two hiccups worth knowing about. The independent check I sent through the diagnosis loop
failed because the fleet's AI budget cap was hit mid-morning; the cap lifted within the
hour, and I re-armed the check to run again on its own. And while I worked, two other
sessions were editing the same shared record files — their finished entries rode along in
my commit, declared in the commit message, which is how this tree handles that.

The council review of the plan is running as I write this. If it approves, nothing more
is needed until the next chassis image rolls — then the verification recipe in the
RUNBOOK proves both halves on the real loanzy.uk case. If it asks for revisions, that's
the next piece of work.

## 2026-08-19, end of session — the council APPROVED it, first time through

Twelve reviewers looked at the plan; it was approved on the first round with four advisory
notes and nothing severe. Most of the notes were "you asserted X, show it" — and each X has
now been measured and written down: the naming helper exists (the code compiles and the
tests pass against it), exactly one workflow in the fleet calls this action, the race two
simultaneous builds could cause resolves itself loudly rather than silently, and the reason
this fix differs in shape from the neighbouring tool-level proposal is now demonstrated
from the live database schema, recorded in that proposal's own file so the two are tracked
together.

The independent diagnosis check I re-armed earlier is running now. Nothing further is
needed from anyone until the next chassis image rolls; then the RUNBOOK's recipe proves the
fix on the real case. The bug file's status line now says exactly this.

## 2026-08-19, evening — both halves approved; the precondition is code-complete

Two developments since the last entry. First, an owner ruling arrived via the portfolio
lane: the calculator fix must cover BOTH writers — the section one already done, and the
tool-level one described in RFC_036 — as one logical change, and the ~50-site build wave
does not start until it lands. Second, both halves are now through: the tool-level fix is
built (a native rebuild of a tool the library already offers is now born as a site copy of
it, which is what the fork field means everywhere else, so the unique-name gate no longer
kills the build), tested the same way as the first half including deliberately breaking it
to watch the test fail, and approved by the council on the first round — as was the first
half this morning.

Where each half stands: the section fix is LIVE in the fleet (proven at the running binary
with controls, not inferred from tags) but has had zero real exercises — the loanzy lane's
next clean-domain build is the agreed real test, with the old site's checksums pinned
beforehand so we can prove nothing was overwritten. The tool fix rides the next chassis
image. One follow-up is tracked in RFC_036 rather than left as folklore: the two different
ways a site can acquire a tool copy don't yet recognise each other's copies, which today
fails loudly rather than silently if it ever happens.

One process slip recorded in WRONG_CALLS: a commit message claimed a document edit that had
actually failed; the next commit corrected it and says so. The bug stays open until the
real-world test passes and a roll carries the tool half.

## 2026-08-19, evening — the real-world test ran, and the fix did exactly what it was built to do

Nobody else had run the real test (the loanzy team had no new domain lined up; the portfolio
sites are frozen under your halt), so I drove it from here on loanzy.uk, the site that first
showed the bug. I picked the car-finance calculator: it had failed three times at the exact
wall this fix removes. (The credit-health one fails for a different reason — the AI's answer is
too long for the limit — so it could not have tested anything here.)

What happened, in order. The build asked for a car-finance calculator. The AI chose the same
name as the existing calculator that belongs to loanandmortgagecalculator.co.uk — the
collision we were waiting for. Instead of failing, the platform noticed the name was taken by a
component another live site depends on, created loanzy's own copy under a site-specific name,
filed it in a way the library can find and reuse, and wrote a record saying exactly what it did
and why. The work item completed first time, no error. And the other site's calculator — all
eight of its calculators, in fact, which I fingerprinted before starting — is byte-for-byte
untouched.

One thing I got wrong earlier and have corrected in the notes: I had said the parked failures
would "converge by themselves" once the fix was live, because a background check keeps
re-flagging the pages. The flag is real, but nothing acts on it — a page only gets rebuilt when
something actually files a rebuild job for it. So the fix makes the next build succeed; it does
not start one. For loanzy that means each of the seven calculator pages needs a rebuild job
filed (or the loanzy lane's full rebuild, which files them all). I have filed one now, for the
car-finance page, to prove the last step: that the page really picks up the new calculator and
serves it. The page as served right now has zero input fields — that is the "before".

## 2026-08-19, later that evening — the page picked up its calculator; the test is passed end to end

The rebuild I filed for loanzy's car-finance page ran through cleanly. The page now carries the
new calculator as its second section, and the live page at loanzy.uk/tools/car-finance-calculator
has gone from a page with no controls at all to one with four — price, deposit, term and
interest-rate — with its own script loading. No stray template code on the page, and the other
site's eight calculators are still exactly as they were before I started.

So for the section half of this bug: fixed, live, and now proven on the real case it was built
for, with the "before" measured so nobody has to take the "after" on trust.

Two things keep the bug open, and neither is a fault in the fix. First, your ruling: the
calculator fix is a pair, and the second half (the tool-level writer) is written and approved but
is still waiting for the next chassis image — nothing to do until that rolls. Second, loanzy's
other six calculator pages are still hollow; each needs its own rebuild filed (the recipe is
written down), or the loanzy lane's full rebuild does them all. I stopped at one on purpose —
that is their site and their planned run — but it is a ten-minute job per page if you'd rather
it just happened. One of the six (credit-health-check) will still fail for an unrelated reason:
the AI's answer is longer than the limit allows, which is a separate small bug to raise.

## 2026-08-19, night — the new build carries both halves; the second half now needs its first real run

You said a fresh build had rolled, and it has: both chassis pods are on v1.0.1316, started at
17:13 UTC, and I proved at the running binary — with a positive and a negative control — that
it contains both the section fix and the tool fix. So the pair you ruled a precondition for the
portfolio wave is live in full.

One correction against myself: in the earlier entries I said the chassis was "still on
v1.0.1315" at 20:17. That was a reading from 16:15 repeated without looking again; the roll had
happened at 17:13. The car-finance page rebuild actually ran on the new build. It does not
change the result (both builds carry the section fix), but the label was wrong, and I have
corrected it everywhere it appeared and logged it as a wrong call.

What remains for the tool half is the same thing the section half needed this afternoon: one
real run. The obvious case is webdesign's A/B-test calculator, which has been parked for days
on exactly the wall this fix removes. That site is being actively worked by its own lane right
now, so rather than fire a build into the middle of their work I have told them, in their own
folder, that the wall is down, what to assert, and the fingerprints to compare against. When
their run passes, this bug is done. If you would rather not wait on them, say so and I will
run it myself once their site goes quiet.

## 2026-08-20, morning — the tool half's real run came back clean, and loanzy's five broken calculators are being repaired

Two things closed overnight and this morning.

First, the webdesign lane ran the test I handed them, and it passed. Their A/B-test calculator was
rebuilt natively instead of dying on the wall this bug was about; the new component correctly
records that it was forked from the shared library copy, and the library copy itself is
byte-for-byte what it was before. This morning I finished the one piece they had left hanging —
grading the actual live page. It serves five working inputs and no stray template code. I also
checked *which* of the three components that page has carried is the one actually being served,
because two of them look similar in the markup: only the new forked one contains the ids the live
page has. So it is provably the new component, not the old one still hanging around.

Second, both halves of the fix are still live on the newest build. The fleet rolled again at 22:26
last night to v1.0.1317, which means every "it is live on v1.0.1316" statement in my notes was
about an image that no longer exists. I re-proved it on the running binary rather than assuming it
carried forward — and this time with a control that shows the check can actually come out
negative, which the earlier ones did not.

Then, on your instruction, I repaired loanzy's five broken calculator pages. All five went through
cleanly on the first attempt, and all five did the thing the fix exists to do: instead of trying to
overwrite the other site's calculator, each one created its own copy under its own name. The other
site's eight calculators are untouched, byte for byte — including one of them that a completely
different piece of the system had quietly rewritten at 7am this morning, which I only noticed
because I re-took the fingerprints just before starting instead of trusting yesterday's. Page
rebuilds for all five are queued now; they take about ten minutes each and run one at a time.

Three things I found while measuring that change what is actually left on that site. The list I had
been carrying said "six hollow pages", and it was wrong in three ways: one of the six only needs a
page rebuild and no new component at all (so re-driving it would have been wasted money); a page I
had not counted, the eligibility checker, is a second victim of a completely different problem; and
two more pages have never planned a calculator section in the first place, so nothing in this fix
will ever help them. The different problem is worth a line: two pages both want the same
credit-health-check calculator, and the AI writing it produces about 48,000 characters when the
step is only allowed 16,000, so it is cut off and the whole thing fails. That is a limit to be
raised or a prompt to be shortened, not a bug in the store, and I have left both pages alone.

The last piece of this bug that is genuinely unfinished is the one you originally saw: a page whose
calculator never resolved still gets built and published, with a placeholder where the calculator
should be. I read the code for that this morning rather than take the earlier note's word for it,
and it is real — the builder inserts a stub and carries on, and the only trace is one warning line
in a log. Twelve pages across six sites are live in that state today. Detection for it exists and
even flags the page for rebuild; the flag just has nothing listening to it. That deserves its own
bug file rather than living on inside this one, and that is what I am writing next.

## 2026-08-20, midday — three of the five pages are fixed, and the two that are not are both something else stopping them

The repairs are done as far as they can go. Every one of the five calculators was created
correctly — that is the part this bug was about, and it worked five times out of five, first time,
without touching the other site's originals. Three of the five pages now serve a real, working
calculator where they served none: loan comparison, overpayment and settlement, with six, five and
five input controls respectively. Loanzy now has four working calculator pages, against one this
time yesterday.

The two that did not land are worth understanding, because neither is this fix failing.

The loan repayment page was built perfectly and then refused publication at the last step, because
the page is **archived**. That is a guard doing exactly its job — it will not publish a page
somebody retired — but I should have checked that before spending the work, and I did not. It cost
one generation and one page build. I have written the missing check into the runbook. Un-archiving
that page is the loanzy lane's decision, not mine.

The interest-rate stress-test page is blocked by a different protection. Re-rendering a page
rewrites all of its sections, and on this page the writer produces a much plainer version of the
banner at the top — twelve styled elements down to five. There is a guard that refuses to save when
a section loses more than half its layout that way, and it refused, twice, with identical numbers.
Nothing was written either time, so the page is exactly as it was. The identical numbers matter:
they mean this is not bad luck that a third attempt would clear, so I stopped. The proper fix,
according to the note left by whoever built that guard, is to tell the writer what the site's
layout vocabulary is, rather than to lower the bar — but that is a change to how every page on
loanzy is written, so it is their call and not a repair I should make from here.

One thing I got wrong twice today and want on the record. I predicted that retry would work, and I
had written the exact diagnostic that would have told me it would not — four hours earlier, in this
same session, into the debugging guide. Writing a check down is not the same as using it.

## 2026-08-21, midday — the pilot site is unblocked, and the calculator fix worked; something else is now in the way

I have lifted the halt on remortgagecalculator.uk, and only that site. The other portfolio site,
adversecreditmortgage.co.uk, is still held exactly as it was, with its forty-one queued items
untouched — you named one site and one row moved. I wrote the old lock values down in the
portfolio lane's own folder so they can put the halt back verbatim if they think I have read it
wrongly. For what it is worth, both of the things the halt was waiting for look settled in their
own notes: the builder flow was ruled on the 19th, and the classifier question was superseded on
the 20th. What they still list as outstanding is a piece of software that has not been written
yet, which is not the same as a decision.

Before lifting it I listed what would be let loose, because a new site build is exactly what the
halt existed to prevent. Five jobs, all filed by automated agents — two contrast fixes, a brand
asset, an unpublished image, and a directory page re-render that arrived at twenty past midnight.
No site build among them.

**The good news, and it is the point of the whole exercise.** The calculator that has been missing
from that home page since the 17th — the one you originally complained about — went through the
fix correctly. The system recognised that the calculator template it wanted to write over belonged
to a different site, and it created its own copy under its own name instead. That is precisely
what this bug was about, and it now works on the very page that started it.

**The bad news.** The new calculator was then refused at the last moment by a different check, and
that check is right. The AI wrote the template asking for a "currency symbol" from a place in the
site's settings that does not exist — anywhere, on any site. I checked the entire codebase; there
is no such setting. So the guard that stopped it did its job. I could have made the error go away
by inventing that setting, and I have deliberately not done that: it would be changing the test to
agree with the wrong answer.

What makes it worse is that it will not correct itself. It tried twice and made the **identical**
mistake both times, and when I looked at why, the regeneration step is handed exactly the same
information as the first attempt — it is never told what was wrong with the last one. Across the
estate that has now happened ninety-nine times on three sites, and every single repeat produced
the same complaint as its predecessor. One job on another site burned fifty-two attempts over
three and a half hours that way. I have written it up as its own bug, and I cancelled our third
attempt rather than pay for a result I already knew.

So the page is still missing its calculator, and I want to be plain that this is now the third
separate thing standing between a working fix and a working page — the other two being a page
somebody had archived, and a guard protecting an unrelated banner. The collision fix itself is
sound and proven seven times over. What is left is a queue of unrelated gates, each reasonable on
its own, none of them this bug.
