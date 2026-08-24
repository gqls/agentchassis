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
