# Where we are — component identity (`bugs_open/357`)

The owner's running plain-prose log. Append only, newest at the bottom. No jargon where plain words
will do.

---

## 2026-08-22, afternoon — asked for 277, found it already closed, moved to its successor

You asked me to look at bug 277. It had already been closed a few hours earlier, on your own
ruling — so rather than take the file's word for it I checked the live system: the routing it
promised is there, everything of that type is classified, and the two pages it used as proof still
serve what they are supposed to. The close is real. Its own closing note says where to go next, and
that is bug 357, which nobody is working on because the lane that filed it finished at the same
time.

## What 357 is about, in plain terms

A web page here is stored as a list of blocks. Each block says which *kind* of block it is — a
banner, a text panel, a calculator — and stores the HTML it should show. The kind matters because
everything else in the platform reasons about it: what fields the block needs, whether it can be
rebuilt, what to check it for.

On 22 live pages, the block is lying. It says it is a **banner** (the "hero" — a title, a sentence
and a button) and what it actually holds is an **entire working calculator or interactive tool**,
ten to twenty thousand characters of it. The block was born that way. Nothing overwrote it later.

## Why it happens, and it is a simple mistake with a big shadow

When the platform saves a page, it splits the HTML into blocks and works out what each one is. Tool
pages don't arrive in the shape it expects — they come as one single lump with no internal markers —
so the splitter gives up on identifying it and labels it with a placeholder meaning *"I don't know
what this is."*

Then the next step does something reasonable-sounding and wrong. It sees the placeholder, looks up
the page's plan, and fills in the name **by position**: this is the first block, the plan says the
first block is a banner, therefore this is a banner. It never looks at the thing itself. And because
every one of these pages plans a banner first, every single tool gets labelled a banner.

So the honest answer "I don't know" is quietly converted into a confident wrong one. That is the
whole bug, and it is the part worth fixing properly rather than case by case.

## Two things I found that the bug file does not say, and one of them is worse than it looks

**It is still happening.** The bug file describes nine pages, the newest from a fortnight ago. Its
own re-runnable query actually returns **22**, and the newest was created **today** — the homepage
of vetcomparison.uk. That page currently opens with a vet-practice search tool sitting where its
banner should be, and the banner text that was written for it is stored but never shown to anyone.

**Thirteen of them are already primed to be destroyed.** The bug file's reasoning is that these
pages are safe as long as nobody writes data to them, because writing data would let the platform
decide it can rebuild the block from scratch — and rebuilding it would render the *banner* it claims
to be, throwing away the calculator. That is right for the nine it looked at. But thirteen of the
other rows **already have that data**, complete and valid, sitting next to the tool. Nothing more
needs to be written. A routine rebuild of any of those pages would replace a working calculator with
a title band, and sixteen of the 22 are on the rebuild-eligible setting.

I want to be straight about one thing I could not establish: whether this has already happened to
some page. If it had, that page would look normal now and would have quietly dropped out of every
query I can run, because there is no systematic history of these blocks to check. I have candidates
but no proof, and I am not going to claim it in either direction.

## What I am proposing to do about it

The fix I want is not "relabel these 22 rows". It is to stop the platform ever again asserting what
a block is without checking. The encouraging part is that the check is cheap and I have measured it
rather than guessed: components stamp their own name into their HTML, and comparing the two across
every block on the fleet gives **1,550 that agree, none that disagree, and 27 where the stamp is
simply missing** — which is this problem plus five other rows worth looking at. The obvious cruder
test flags 158 and is useless, because it also convicts pages whose styling was legitimately edited
later. So there is a precise test available, it does not cry wolf, and the 1,550 agreements are what
tell me the test is actually doing something rather than passing everything.

Detailed plan being drafted now. I will come back to you before anything is applied, and separately
on the question of what to do with the 22 existing pages — that is a decision about changing what
live sites serve, which is yours and not mine, and 277 has just finished teaching everyone how
expensive it is to get that decision wrong.

## 2026-08-22, later — I had the danger backwards, and the truth is more interesting

I told you earlier that thirteen of these pages were primed to be destroyed: that a routine rebuild
would throw away the calculator and leave a title band. **That was wrong, and I want to correct it
plainly rather than quietly.**

What actually happens is the opposite, and it explains why this has been sitting there for months
without anyone noticing. When the platform rebuilds one of these pages it *does* generate a banner
for that slot — and then a safety net catches it. That net was built after an earlier accident where
rebuilds silently blanked working content, and it does its job: it sees that what is stored in the
slot is an interactive tool, sees that the freshly generated replacement is not, and puts the tool
back.

But it puts the tool back **under the banner's name**. So every rebuild faithfully preserves the
calculator and faithfully preserves the lie about what it is. **The mechanism that protects these
pages is the same mechanism that keeps them mislabelled.** I proved it on the vetcomparison
homepage: six rebuilds in the last four days, the tool intact every time, and the current database
rows were written *during* the most recent one.

The practical consequence is about ordering. These pages regenerate themselves, so tidying up the
existing 22 first would be wasted — the next rebuild would undo it. The producer has to be fixed
first, then the existing pages, and then the safety catch can be switched on.

## What the plan comes to

Four pieces, and only the first three are things I'd like to build now.

**Stop it happening.** When the platform meets a page it cannot identify, it should say so instead of
guessing from the plan. Today the "I don't know" placeholder gets silently upgraded to "banner"
because banner is first in the list. Instead the fragment gets its own honest type, with a template
that simply re-emits whatever is stored — which has the nice property that these pages become
genuinely rebuildable for the first time, rather than merely being protected from rebuilds.

**Check it.** Every page save funnels through one database write. Components stamp their own name
into their HTML, so at that one point we can compare what the row *claims* to be against what it
*is*, and record the disagreement. I want it to **record only** at first and not block anything —
there is a change from yesterday on the neighbouring code that made exactly this call, and it was
right to. The ability to actually refuse ships switched off, with a switch that has to be turned on
deliberately, one caller at a time.

**Look everywhere else.** Six different pieces of code write these rows and I have only fixed the
path used by one. A daily check covers the rest without touching them.

**Then repair the 22 — and that one is your call, not mine.** It changes what four live sites serve.
The repair itself is designed to change nothing a visitor sees: the same bytes, correctly labelled.
But 277 spent a fortnight learning how expensive it is to be casual about this class of change, so I
would rather show you the check-and-restore script and the before/after of one page than just run it.

## Where it stands right now

The plan has gone to the review council (reference 62aac6c2) and is being reviewed as I write. A
separate diagnostic run is still queued behind the fleet on the one question I could not settle from
the database alone. Nothing has been changed on any live site, and no code has been altered yet —
what exists so far is the diagnosis, the measurements, and the plan.

Two things I got wrong today are written down in the shared mistakes log with the checks that would
have caught them, and both corrections are recorded in the bug file itself rather than edited away.

## 2026-08-22, end of session — the council stopped me twice, both times rightly, and there is a decision for you

**Nothing has been changed.** No code, no database, no live site. What exists is a diagnosis, a set
of measurements, two rejected plans and an architecture question.

**The first plan was unsafe.** The review council pushed back and, while checking their objection, I
found something worse than the thing they raised. The tidy-up I described to you as "changes nothing
a visitor sees" would in fact have changed what four live sites serve. The reason is a detail nobody
had written down: when the platform rebuilds a page it matches the new blocks against the stored
ones **purely by name**. Rename a block to fix its label and the match silently fails — the platform
concludes the block was dropped, re-adds the tool at the end, and keeps the newly generated banner
too. The page ends up with both. So correcting the label is exactly the thing that would disturb the
page.

**The second plan was safe and useless, and they said so.** I cut it back to "watch and record,
change nothing", and the council's answer was blunt: the corruption carries on, the 22 pages get a
work item filed and nothing else, and this estate has a documented habit of detecting problems and
never blocking them. That is a fair hit and I have accepted it rather than argued.

**What they recommended instead is the useful thing.** One of the reviewers pointed out that the
platform works out what a block *is* in **five different ways** — from a marker in the HTML, from a
placeholder meaning "unknown", from its position in a plan, from fuzzy name-matching, and from
name equality during rebuilds — and it never simply **records** it at the moment the block is made.
Every bug this lane found is a symptom of that. Adding a sixth way of guessing is not the fix, and
the fact that my safe plan and my effective plan could not be the same plan is the evidence.

So I have written that up as an architecture question (RFC_046) rather than pushing a third version
through. It asks for one of three answers: **record** identity properly at the point the block is
made; **accept** that it is guessed but consolidate five guesses into one; or **knowingly leave it**,
which is defensible so long as it is a decision and not something that happens by default.

**One piece of good news, and it corrects something I told you earlier.** I said we could not know
whether this had already destroyed a working tool. That was wrong — there *is* a proper history
table, which I had missed because I truncated a listing and read the visible part as the whole. With
it, the answer is a clean no: of the slots that once held something interactive, 182 still do; 17
changed, and opening all 17 shows fifteen of them **grew** — ordinary rebuilds, not losses. None is
one of our 22 pages. One unrelated page did shrink by more than half on 15 August and I have written
it down as a lead for someone.

**What I would like from you:** the RFC_046 decision. Until then the 22 pages stay exactly as they
are — which is safe, they all serve their tools correctly today — and the label stays wrong.

I got six things wrong today, all of them written down in the shared mistakes log with the checks
that would have caught them. Three were the same mistake: describing a set of rows from its count
without opening the rows. It appears in the tally three times in one day, and the log now says
plainly that when that happens the check needs a mechanism rather than another entry.

## 2026-08-23 — the thing we built and got approved last night has been running all day doing nothing

Short version: yesterday's work was correct and it was not connected to anything. I found that,
connected it, and found on the way that connecting it naively would have made things worse rather
than better.

**What yesterday built.** A way for the platform to record, at the moment a page section is made,
*which version of which template actually produced these bytes*. That record is the thing the whole
repair depends on — you cannot safely fix a row that lies about what it is until something reliably
says what it is.

**What I checked first.** Not "what shall I build next" but "did that actually do anything?" It had
been live since 15:10 yesterday and approved by the review council at 18:02. In the day since,
**820 new sections were written and not one of them carries the record.** Zero. The check that
proves this is not a hunch is that I asked the same question of a field I *know* works: of 546 live
payloads, 546 carry the control field and 0 carry ours. So the copying step definitely runs — it
just quietly leaves ours out.

**Why.** Between the part that produces the record and the part that stores it, there is a step that
rebuilds the parcel from scratch, copying across a hand-written list of contents. Our field was not
on the list, so it was dropped in silence. No error, no warning, nothing in a log.

**This has happened here before, and the way it was fixed last time is why it happened again.** The
same step lost a different field in an earlier bug. That was fixed by adding *that one field* to the
list and writing a test that checks *that one field* survives. The test was passing the whole time
ours was being thrown away, because it only knows about its own field. Three fields were being
dropped, as it turns out, and nobody knew.

So I have not added our field to the list. I have made the list a **contract**: everything a producer
sends is either on the carry list or on a "deliberately dropped" list with a written reason, and a
test now fails the build if someone adds a field to neither — or if the storing end reads a field the
carrying end was never told to carry. That last rule is the one that would have caught this on day
one.

**The thing that would have gone wrong if I had just added the field.** There is a protective
mechanism that stops a rebuild wiping out an interactive tool: when the page rebuild produces a plain
text banner where a tool used to be, the platform puts the tool back. It swaps the *content* — but it
was leaving the *record of where that content came from* untouched. That was harmless while the
record was always empty. The moment I connected it up, the platform would have confidently recorded
"this tool was produced by the hero banner template" — which is false, and worse than recording
nothing, because it is exactly the kind of confident wrong answer that other tools then trust. That
is fixed in the same change.

**On the bug itself: it is still happening, roughly a dozen times a day.** Twelve of the twenty-two
mislabelled pages were created *today*. Nothing I shipped today stops that — today's work makes the
evidence real and honest, and the next step uses it to stop the mislabelling at birth. I want to be
plain that this is not the fix; the review council caught this lane claiming otherwise once already
and was right to.

**One small good thing:** last night's round-4 review came back **approved**, nine minutes after the
lane stopped writing, so nobody had recorded it. It is recorded now.

**Three mistakes of my own today**, all written down with the checks that would have caught them. The
one worth telling you about: I wrote a test for the protection above, watched it pass, then broke the
code deliberately to make sure the test would catch it — **and the test still passed.** Two different
failures produced the same visible result, so the test could not tell them apart. It was green,
reasonable-looking, and worthless. I rebuilt it so that deliberately breaking the code now does fail
it. That check — break your own code and confirm the test notices — is the only reason I know the
rest of today's tests mean anything.

## 2026-08-23 (later) — the mislabelling now stops at the source, with the switch off

Since the last note I have built the part that actually stops new bad pages being
made, and written the repair for the existing ones without running it.

**The idea that unlocked it.** A section has two different names: the slot it sits
in on its page, and the component that made it. Everything that goes wrong here is
about the second; everything dangerous is about the first — renaming a slot is what
would make a rebuild duplicate the tool. So the new code fixes what a row says it
IS and never touches what it is CALLED. That is why this version can be both safe
and effective, where the two earlier attempts could only be one or the other.

**How it decides.** When a page arrives as a single blob that nothing can identify,
the platform now attaches it to a component that is literally "the content, as
given" — but only after rendering that component and checking the result is
byte-for-byte what it was about to store. If that check fails, it records nothing
rather than guessing. Honestly unlabelled beats confidently wrong.

**It is switched off.** It ships behind a flag that defaults to off, so the next
release changes nothing on any site. Turning it on is a separate decision, and a
reversible one.

**The repair for the twenty-two is written and deliberately not run**, because your
ruling was that it may only happen once the record is real and readable on a live
page — which needs the release first.

**Something you may want to decide.** Six of the twenty-two are pages marked as
claimed by a human. That is why they have sat unchanged since June while the others
are rewritten constantly. They are still mislabelled and still generating false
warnings, but I do not think automation should quietly rewrite a page someone has
claimed, so the repair now prints them by name and skips them. If you would like
them included, it is a one-line change.

**Two more of my own mistakes today, both the same shape as the first.** A test
fixture was one column short, which made the thing under test silently not run at
all — the test passed while checking nothing. And the first version of the repair
quietly selected 16 of the 22 pages, missing the six the bug was originally
reported about, because I had written my own narrower version of a rule the
platform already had. Both were caught by running the check against the real
database instead of trusting what I had written, which is now the only way I am
willing to accept one of these.

## 2026-08-24 — it went live overnight and it works, and I can show you rather than tell you

Yesterday I said the fix would go out with the next release and then had to be
*checked* rather than assumed, because the whole lesson of the day before was that
approved-and-shipped does not mean working. The release went out overnight. I
checked. **It works.**

**How I know, rather than believe it.** Sections written before about nine this
morning carry no record. Every hour after it is at or near everything — forty-six
out of forty-eight, then a hundred and seventeen of a hundred and nineteen, then
twenty-four of twenty-four, then fifty-eight of fifty-eight. The step is the
evidence, not the total. And the control holds in the other direction: of the
nine-hundred-odd sections written *before* the release, **none** has a record,
which is exactly right — nothing goes back and fills them in, by design. A number
that only moves forward from a known moment is hard to explain any other way.

**Something better than a passing test: the mechanism earned its keep on day one.**
Six sections looked wrong at first glance — they name a template that is not what
their component says today. Opening them up: the version was created at 10:17, the
sections at 10:55, and **somebody edited that component at 11:15**. So the sections
are right and the component moved afterwards. Before this change those six were
indistinguishable from sections built with the new version. That is precisely the
confusion this was built to end, and it turned up within two hours of going live.

**One thing I nearly recorded as good news and is not.** The check that would tell
us the worst case had happened reads zero — no mislabelled page has picked up a
*wrong* record. But no mislabelled page has been rebuilt since the release either,
so the check has had nothing to judge. **A zero with nothing to count is not a
pass.** It is still pending, and I have written down the version of the query that
shows both numbers together so the next person cannot make the same mistake I
nearly did.

**Both halves are now approved by the review council.** The second one came back
for revision first, and both of its objections were right: one part of my change
reached wider than the problem it was for (narrowed), and I had claimed a page
"serves the same either way" without actually checking one of the two paths it
could go down (it doesn't, and that limit is now written down instead of glossed).

**What has not changed: the mislabelling is still happening.** The half that stops
it is live but switched off, and turning it on is your decision, not mine. The
procedure is written out in order now — seed first, then arm one pipeline rather
than all eight, then three specific things to watch for, each of which is invisible
to the obvious "is the tool still there?" check. It is a decision you can take in
one sitting rather than a piece of research.

**Two other lanes came to us today and both exchanges were worth having.**

One is replacing a component-quality score and asked whether to use our new record
to filter out our bad rows. I said yes and **I was wrong** — it excludes our
twenty-two, but at four hours old it excludes about seventeen hundred honest ones
too, so it would have measured *how recently something was rebuilt* rather than how
sound it is. They declined it, correctly. I had justified a filter by what it
throws away and never counted what survives it.

The other told me something I would never have found: **a change of mine went out
inside their commit**, twelve minutes before my own, because we were both editing
the same file. Nothing was lost — the code in production is exactly what I wrote —
but my own check for "have I committed everything?" had said yes for the wrong
reason: the file looked clean because *they* had committed my work, not because I
had. Their message also surfaced four serious objections about our mechanism that
had been sitting in a rejected review under someone else's name, addressed to
nobody. They are answered, and the answer was already on record from an earlier
approved round.

And a question I asked them came back at me: I warned them their change might upset
a stored checksum, and it turned out **the same question was unanswered about our
own repair migration** — which I had been calling "changes no bytes" all day without
checking what reads that checksum. It is safe, now verified at the source rather
than assumed. The useful part was not the answer; it was that neither of us had
asked.

**What I would like from you, unchanged from yesterday and now the only thing
holding it up:** whether to turn the second half on, and whether the six pages
marked as claimed by a human should be repaired along with the rest.

## 2026-08-24 (later) — it is switched on

You said arm it, so it is armed. The component it needs was created first, then the
switch was turned on, and I read the result back independently rather than trusting
the script's own "done".

**One thing I did differently from my own written plan, and I want to be upfront
about it.** This morning I wrote that we should turn it on for one pipeline first and
watch — the obvious one, the one that rebuilds tools. Before doing that I listed
which pipelines can actually produce these bad rows, and the answer was **five of the
six, not one**. The route that creates them is reached by any page-builder whose
normal data comes back empty, not just the tool rebuilder. Turning on the obvious one
would have left four others still making the problem, which is a pattern this
platform's reviewers keep catching — fix one door properly and leave the rest open.
So I turned it on everywhere it can matter, and wrote down why, and marked my own
earlier advice as wrong rather than quietly ignoring it.

**Turning it on cannot affect ordinary pages.** The new behaviour only triggers on a
page that arrives as one unlabelled blob with no sections in it at all. A normal page
has sections, so the code path is unreachable for it. That is checkable in one line
rather than something you have to take on trust.

**What is true now:** new tool pages should stop being mislabelled. **What is not yet
true:** I have not seen one come through. Arming is a setting; the proof is a real
page landing correctly, and I have a watch running for it along with three specific
things that would mean stop. I am being careful about the difference because this lane
already lost a day to a mechanism that was approved, shipped and doing nothing.

**The twenty-two existing bad pages are untouched.** Their repair is written and
still deliberately not run — it checks its own preconditions and will refuse until a
real page has come through correctly. And six of those twenty-two are the ones marked
as claimed by a human, which is still a decision I would like from you rather than
one I should take.

## 2026-08-24 — correction: those six pages are not "yours", and that changes the answer

You said you did not mark those six pages as owned. **You are right, and I was wrong
in a way worth spelling out, because I told you the same thing three times.**

I read a field called `rebuild_policy = 'owned'` and took "owned" to mean a person
had claimed the page. It does not. The platform's own guard says such a page
*"belongs to a tool/widget or is a runtime-fill shell"* — it is a category the code
assigns to pages that a tool produces, and the code writes it outright when it
creates one. **172 of the 704 pages on the estate carry it.** Nobody chose it page by
page, and certainly not you.

What I did was infer a meaning from a field's name, never check what writes it, and
then bring you a decision built on that inference. The decision I offered you was not
a real one.

**And the correction turns the conclusion round completely.** I had treated those six
as the pages I was least willing to touch. They are in fact **the only six that
cannot possibly fix themselves.** The producer fix we armed this afternoon lives
inside the save routine — and for a page in this category that routine refuses at the
door, two hundred lines before the new code is reached. Every other bad page will be
repaired by an ordinary rebuild now. These six never will be. A migration is their
only route, which makes them the *most* deserving of the repair rather than the
least.

They are also the safest to repair, once you see it that way: because the pipeline
refuses them, a row fixed here stays fixed. There is no rebuild waiting to undo it.

So: **they are now included.** The repair targets all twenty-two, and it prints each
of these six by name as it goes with the reason it is touching them, rather than
skipping them behind a condition nobody would read.

One thing I want to be straight about: repairing them does not defeat the guard. That
guard exists to stop the generic page pipeline deleting and rebuilding a tool page's
contents. The repair does not delete or rebuild anything — it corrects three fields
and leaves the actual markup untouched. It is not the operation the guard is there to
prevent.

## 2026-08-25 — the safety check we could not test has now been tested, and it held

Yesterday I flagged one check as **pending rather than passed**: the thing that stops
the platform recording a *false* origin for a tool. It read "all clear", but nothing
had happened that could have made it read otherwise, so the all-clear was worthless.

**This morning it got its test, and it passed.**

At 09:08 one of the twenty-two bad pages was rebuilt. The platform did what it always
does — put the tool back where a rebuild would have replaced it with a plain banner —
and at that exact moment it had to decide what to record about where those bytes came
from. The honest answer is "I don't know", and the tempting wrong answer is "the
banner template made them", which is false and worse than saying nothing.

**It said nothing.** And the reason that means something: of **571** page-sections
saved since we switched this on, **570** were recorded confidently. Exactly one was
not — and it is that row. So the machinery is not simply silent; it is silent in the
one place where speaking would have been a lie.

**What has not changed: all twenty-two pages are still wrong**, and that one was
re-created wrong this morning. That is expected — the half that stops NEW pages going
wrong is separate from the half that repairs the existing ones, and the repair has not
run.

**The honest gap, which I want to be plain about.** The part we switched on yesterday
— the bit that should stop new pages being mislabelled — **has still never actually
done anything.** No page has come through the route it watches. That is not evidence
it is broken, but it is not evidence it works either, and this lane already lost a day
to something that was approved, shipped and quietly doing nothing. I am not going to
call it fixed on the strength of it being switched on.

There is also a hint that new bad pages may arrive by a route we have not
instrumented: the affected pages have their sections written at *different times*,
which does not fit the one route we have covered. That is the first thing the next
session should chase, and it is written at the top of the handoff.

**The repair is written and will refuse to run**, deliberately, until we have seen the
new behaviour work once on a real page. I would rather it refused than have it rewrite
twenty-two live pages on the strength of a mechanism nobody has watched do its job.

**Can we close this?** No — not yet. The thing the bug complains about is still true of
every one of the twenty-two pages. What is finished is the foundation: the recording
works, at volume, and its most dangerous failure mode has now been shown not to happen.

## 2026-08-25 (afternoon) — we ran the adoptions, and they found the thing that has been blocking us

You asked me to run an adoption, and offered cv1.co.uk and lampenkap.com. I ran both,
cv1 first, as you chose. Both went through cleanly. Neither produced the result we were
after — **and the reason why is the most useful thing this lane has learned all week.**

**First, a piece of luck worth owning up to.** I told you lampenkap was the surer bet,
because its front page has a working calculator on it and cv1's front page has no
interactive code at all. **I had that exactly backwards.** The platform decides what
counts as an interactive page by reading the crawled site with a model, not by looking
for code — and it judged cv1's two pages interactive and lampenkap's calculator page
not. Had you taken my recommendation and run only lampenkap, we would have got nothing
at all. You said run both, and that is the only reason we have a result.

**What happened.** Both cv1 pages went to the tool rebuilder, exactly as we needed. It
wrote two complete tool pages — twenty-six thousand and twenty-one thousand characters.
I checked those against every condition the new mechanism needs, and **all of them were
perfect.** This was the moment the mechanism was built for.

**And then the save was refused, by something else entirely.**

There is a separate safety rule that stops a rebuild replacing a good page with a
thinner one. It works by comparing how many sections the rebuild produced against how
many the page is *planned* to have. A tool page is one single block by its nature — so
it produced one. The plan for those pages said four and three. One out of four is 25%,
and the rule refuses anything under 50%. So it threw the whole save away.

**Here is why that is a fault and not the rule doing its job.** The same piece of
software makes both decisions, moments apart: it decides "this page is a tool, send it
to the tool rebuilder", and it writes down "this page has four sections". Those two
statements cannot both be true. The tool rebuilder can only ever produce one section.
So the page was sent down a road that the rule guarantees will be blocked.

**And this explains the original bug.** Our twenty-two bad pages should then almost all
be pages that were planned with one or two sections — because those are the only ones
whose one-section save could ever have got through. I checked. **Twenty-one of the
twenty-two.** So this rule has quietly been deciding which tool pages get built at all:
plan it with two sections and it saves (badly labelled, which is our bug); plan it with
three and it is refused outright and the page stays empty.

**It is not just us.** There are thirty-two of these refusals sitting waiting for a
human, going back to the end of July, across fourteen different sites — and several are
named tool pages on sites you know: webdesign.co.uk, fundamentallyai.com,
mortgagecalculator.co.uk. Nobody has been acting on them.

**So the honest position.** The mechanism we switched on last week still has not been
seen to work — but we now know that its zero was never evidence about the mechanism. It
was evidence about a gate further down the road that nothing ever got past. That is a
much better place to be than "it is armed and silent and we do not know why".

**The repair of the twenty-two is still refusing to run, and it is right to.** Its
condition is that the new shape has been seen working in production once. It has not, so
it stops. I have not touched that check.

**I have written the fault up properly and put it through the diagnosis process** rather
than just asserting it, because it is a claim about how the system works rather than a
one-off.

**What I need from you is which way to unblock it** — there are three ways and they are
genuinely different in risk. I have set them out in the message to you rather than
choosing one myself, because two of them involve either relaxing a safety rule across
the fleet or changing what a live page says about itself.

## 2026-08-25 (later) — it worked. The thing we switched on last week has now done its job on a real page

You chose to correct the two page plans rather than relax the safety rule or wait for the
proper code fix. That was the right call and it took about three minutes to land.

**What I changed.** Those two pages had been told, by the platform itself, that they were
tools — and told, by the same piece of software in the same breath, that they had four and
three sections. A tool page is one thing, not four. I corrected the plans to say one, which
is what the platform's own notes on those pages already said: its analysis recorded
"self_contained: true" for one of them while writing it a three-section plan. I did not touch
the safety rule, any setting, or any other site. Both pages were empty, so there was nothing
to lose, and I wrote the exact undo alongside it before running anything.

**Then both pages rebuilt and both were recorded correctly.**

The important one is the front page. It now holds a seventeen-and-a-half-thousand-character
interactive tool, sitting in a slot called "hero" — **which is precisely the situation this
whole bug was filed about.** Before, the platform would have written down "this is the shared
banner component", which is false, and everything downstream would then have reasoned about a
banner that is not there. Now it says "these bytes are my content", it can prove it — the
stored content regenerates the page exactly — and it carries a proper record of where it came
from.

So the mechanism is no longer something we believe works because it is switched on. **It has
been watched doing its job, twice, on real pages.**

**I checked the three things that would have meant stop, rather than assuming them.** None of
them fired. One of them looked at first as though it might: a page on a different site had a
component with no identity recorded. It turned out to be from half an hour before I started,
on another lane's site, and it is a different situation entirely — I have written down why,
because on the surface it looks exactly like our mechanism failing and it is not.

**What is still open.** The repair of the original twenty-two has not been run. Its condition
is now genuinely met — but there is a second condition, written down as a sentence rather than
enforced by the code: that an adopted page survives being rebuilt without losing what it just
gained. I am testing that now on one of the two pages and deliberately leaving the other
alone, so that if something changes I can tell whether the rebuild caused it. I would rather
spend twenty minutes on that than re-type twenty-two live pages on the strength of a shape
that has existed for four minutes.

**And the bigger thing this turned up is still there.** Thirty-two pages across fourteen sites
were refused in exactly this way and are sitting waiting for a human, going back to the end of
July. The plan correction I did today fixes two pages. The software that writes the
contradiction is untouched, and it will keep writing it.

## 2026-08-26 — can we close it? No, and it is worth being precise about why

You asked whether this can be closed. **No — and the reason is a single number: twenty-two.**

The bug says twenty-two live pages have a record claiming to be a shared banner while actually
holding a whole interactive tool. I re-measured this morning with the bug's own test, and it is
still true of **all twenty-two**. Our rule here is that a bug closes when it is fixed *and*
live, not when the fix exists — and every one of those pages would still show the fault today.

**What has finished is the machinery, and that is genuinely most of the work.** Three of the
four parts are done: the record of where a page's markup came from, the guard that stops it
recording a false origin, and — proven yesterday — the part that stops new pages being
mislabelled in the first place. That last one is no longer something we believe because it is
switched on; it has been watched working on two real pages, and it has now survived two
platform restarts without losing anything.

**What is left is the fourth part: repairing the existing twenty-two.** It is written, it
checks its own conditions, and it is one step away. That step is proving an adopted page
survives being rebuilt — because re-labelling twenty-two live pages into a shape we cannot then
rebuild would be worse than the fault we are fixing.

**And right now that step cannot be taken, for a reason outside this work: the platform cannot
make any AI calls at all.** I checked properly rather than trusting the call log — every one of
the last hundred and twenty-six attempts failed, none produced any output. The message is that
the Anthropic account is out of credit. It started at about a quarter to midnight and it is
worse this morning, not better. Everything remaining on this lane needs a page rebuilt, and
every rebuild writes content first, so nothing here can move until that is sorted.

**One caution I want to flag, because I nearly fell for it myself.** The log of AI calls
records that a call was *attempted*, not that it *worked*. Looking at it quickly, it shows
plenty of recent activity and looks like a healthy system. It is not. If you or anyone else
checks whether the platform is back up, the number that matters is how many calls returned
anything.

**The most valuable thing to come out of this is not on this lane at all.** Running the
adoptions turned up a genuine fault in how new sites are built: the software decides a page is
a tool and sends it to the tool builder, and in the same breath writes down that the page has
four sections — which the tool builder can never produce. The save is then refused and the page
is left empty. Thirty-two pages across fourteen sites are sitting refused because of it, going
back to the end of July. That deserves its own file and its own review, and it will keep
producing casualties until someone fixes it.

**2026-08-26, evening.** The API credit came back this morning and held all day, so that
blocker is gone. A fresh build (v1.0.1345) rolled tonight, but nobody has fixed the crash bug
(408) yet, so rebuilding the adopted page is still off-limits.

Before carrying on I re-checked the whole plan, and found a problem with the test we were
queueing up: once the crash is fixed, that test would come back green without proving
anything — every step in the chain quietly skips, nothing touches the row, and our checklist
reads that as a pass. Worse, the cv1 pages can never prove the thing the repair (578) actually
depends on, because their build plans point at the adopted component itself, so the build
always skips before the preservation machinery gets a turn.

The 22 mislabelled pages are the shape we actually need to test — their plans point at "hero",
so a rebuild really does generate new content and the preservation machinery really does have
to fire and keep the new label. So the suggestion for the next session: fix the crash bug
first, then trial the repair on ONE of the 22 (the simplest mortgage-calculator page), rebuild
it, and check the database row AND the live page. If both hold, run the full repair. One thing
to watch on that trial: the pipeline publishes the rebuilt page to the site one step before
the preservation logic runs, so the live page needs reading afterwards, not just the database.
Full detail in the new handoff (2026-08-26b).

**2026-09-02.** The crash bug (408) is fixed in code and committed. The broken function used
to try two "helpful" path rewrites that were exact opposites of each other, so when neither
worked it bounced between them forever until the pod died. It now tries a fixed, finite list
of path spellings once each and gives up cleanly — and the page is skipped instead of the pod
crashing. Fifteen tests prove it, including one that runs the exact input that used to kill
the pod (we also proved the test catches the old bug by running it against the old code — it
fails in under four seconds there). The fix went to the review council and is committed, so it
rides the next fleet build; the bug stays open until we see it working on the live cluster.

Two things worth telling: the fix recipe written in the bug file turned out to be subtly wrong
(it would have quietly stopped some unusual paths resolving) — caught during planning, before
any code was written, and logged. And when we went to edit the file, another session had
half-finished work sitting in it that would have broken everyone's build if we'd committed
blindly — we messaged them, they committed their half within minutes, and both changes now
coexist cleanly. Next: the fleet build, then the live check, then back to the main 357 repair
via the one-page trial we proposed on the 26th.

**2026-09-02, close of day.** Three things landed together. The repair migration (701) is
approved — it took three review rounds, and every round made it genuinely better: round one
made the misconfiguration case countable and forced a controlled re-census that caught my own
"all pages have a plan row" error; round two made us read two existing mechanisms we'd have
otherwise ignored and prove — at the running binary, not in the source — that the fork logic
the future depends on is actually deployed. The crash bug (408) is closed properly: the fix
is live, and we re-ran the exact input that used to kill the pod — it now completes in
minutes with a polite "nothing to do" instead of a dead pod and a four-hour wedge. And the
warning file that used to say "rebuilding an adopted page crashes everything" now says what's
actually true today: it doesn't crash, it quietly does nothing, and quietly-nothing is the
thing to check for. What's left is yours: run the one-page pilot (the command is in the
notes), check the page still works, then run the other twenty-one.

**2026-09-02, night — the bug is closed.** You applied the repair yourself in two stages and
every guard passed; all twenty-two pages now own their tools under honest names, and the
served sites never moved a byte. Then the system handed us the proof we couldn't manufacture:
the evening news refresh rebuilt the vetcomparison homepage through the entire new machinery
— and rebuilt the tool perfectly. That was the exact scenario this whole lane existed to make
safe, happening on its own, hours after the repair landed. One honest asterisk is in the
notes: the migration's own "verify" rerenders turned out to prove less than they seemed (a
neighbouring lane caught why, and the file is corrected for any future run) — the organic
rebuild is the proof that counts. 357 is in the closed pile. Still open and handed on: the
save-refusal defect (406), and the vetcomparison contrast tidy-up now unblocked for the
design pass.
