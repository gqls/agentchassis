# Where we are — putting the same calculator on a page twice

Plain prose, append-only, newest at the bottom. The owner's document.

---

## 2026-08-16 — what this is about, and what happened today

**The problem, in one paragraph.** Every element on a web page can carry a name — an "id" — and the
page's own code uses those names to find things: read what the visitor typed into `loanAmount`,
write the answer into `displayMonthly`. The rule of the web is that a name must be unique on the
page. Our calculators each use fixed names. So if you put two calculators on one page, they both
claim the same names, and the browser gives every lookup to the first one. The second calculator
still looks perfect. It accepts typing. Its button works. It just reads the *first* calculator's
numbers and writes into the *first* calculator's answer box. On a site about loans, that is a page
that can show someone a repayment figure calculated from numbers they never entered — a believable
wrong answer rather than an error, which is why we treated it as a bug rather than a limitation.

**What you asked for.** That reuse should genuinely work — *"if we chose to list all the calculators
on one page we'd hope it would work"*. Not a workaround for one page, the real thing.

**Where the work had got to before today.** The previous session had built the machinery: each
placement of a component now gets its own name-prefix, and there is a checker that looks at an
assembled page and reports the three distinct ways repetition breaks it. It had gone to the
reviewer council and come back with **"revise"** — five of the twelve seats objected, and two of
the objections were right.

**What happened today.**

*First, a question the last session had left hanging: was any of that actually running?* It said it
couldn't tell, and it was honest about why — the log line that reports what a service was built
from had scrolled away hours earlier, and the other method it tried can only ever confirm a guess.
There is a way to settle it: ask the running container which exact image it has, ask the local copy
of that image whether it is the same bytes, and only then read the label that says which commit it
was built from. It had in fact been live for about fifteen hours.

*Second, the design question that everything else depended on.* The name-prefix had been built from
a component's position on the page — first section, second section, and so on. That turned out to
be wrong for a reason that only shows up later: the same calculator sits in position 0 on seven
pages and position 1 on the other sixteen, so it would answer to two different names depending on
which page you were looking at. Every test, every style rule, every piece of hand-written code that
points at a calculator would then need to know which page it was on. We changed it to name things
after the component itself — `c-mortgages-repayment` — with a numbered suffix only when the same
one genuinely appears twice. That way a calculator has the same name everywhere, and our 170
automated checks need one change each rather than a per-page map.

We could change it freely because nothing uses it yet: none of the 243 components reference the new
name. That is worth saying plainly — **the machinery is live and doing nothing**, by design. The
actual conversion of the 22 calculators has not happened, so the original bug is still there.

*Third, the two objections that were right.* One reviewer pointed out that I had built a second,
weaker way of generating the name for the paths that render a single section at a time — the same
label meaning two different things, which is precisely the trap this whole piece of work exists to
remove. Recreated one day later, inside the fix, by me. That is deleted now; those paths feed the
one rule instead. The other pointed out that patching the three places I knew about left the
underlying mechanism generic — any *other* place that renders a component would silently reproduce
the bug. That one turned into the most useful part of today's work.

*Fourth, and this is the bit I'd underline.* When I went to check "which places render components",
I searched two directories and found eleven. The real answer is fourteen, in eight files, and the
one I missed was in a directory I hadn't thought to look in. Separately, the council's own list of
five files was wrong — four of them don't do this at all — and the two files that were the actual
problem were on nobody's list. **Three attempts to list these by hand, three wrong lists.** So the
answer isn't a better list; it's a check that runs automatically and refuses a new one that forgets.
That now exists, and I proved it works by running it against yesterday's code, where it correctly
flagged four files, and today's, where it flags none.

*Fifth, an admission.* I had written in two places that a larger follow-up piece of work was
"filed". A reviewer asked where, and the honest answer was nowhere — it existed as a sentence in a
commit message. It is filed now (`RFC_032`), and it is a real question worth a proper decision:
there are three different pieces of code that build a component's rendering context and they don't
agree on what "one instance of a component" means.

**Result.** The council approved it this morning with six advisory notes, all of which are answered
or recorded, and it went live on the fleet the same morning.

**What is left, and what I need from you.** Converting the 22 calculators is the work that actually
fixes the original bug, and it needs your go-ahead for two reasons. It writes to the live database,
which I don't do unasked. And the council's architecture reviewer was explicit that the conversion —
not the machinery — is the point where this becomes a real commitment across the whole component
library, and deserves its own review. There is also a knock-on: converting changes the exact bytes
of 22 live pages, and one of our checks currently verifies that those bytes don't change. That
check needs rebaselining first, deliberately, rather than being discovered mid-run.

---

## 2026-08-17 — I checked the size of the job before starting it, and it is four times bigger

You said go ahead, and asked me to check nothing had moved underneath us first. Nothing had: 120
commits had landed from other work, three of them touched files this job also touches, and none of
them disturbed it. The safety check I built yesterday had run overnight on its own schedule and
correctly noticed that the component library grew by one overnight — which is the small proof that
it is actually watching rather than just installed.

Then, before writing any of the conversion, I measured what "convert the calculators" actually
means. **The file said 22 templates. It is 91.** Ninety-one stored components, appearing on 94 live
pages across 22 different sites, carrying about 1,350 hardcoded names between them. The good news
is that they are almost all independent: converting one of them changes one page, with three
exceptions. So this is a long job rather than a risky one.

**Two things I found that change how it has to be done, and I proved both rather than trusting my
reasoning.**

The first is the important one. The obvious way to do this job is in two steps: rename all the
names first, because that part a machine can do reliably, then deal with the trickier scripts
afterwards. I was about to propose exactly that. It turns out that doing the first half alone
produces a page that *passes the safety check* and is *still broken* — because both copies of the
calculator still share the name of the function that does the arithmetic, so whichever copy loads
last wins and both buttons run its sums. The page would look fixed, test clean, and still show a
visitor a figure calculated from someone else's numbers. So the two halves have to be done together,
per component. I wrote a test that demonstrates this rather than leaving it as an argument, and then
deliberately broke the test to confirm it was actually checking the thing.

The second is smaller but would have cost a wasted afternoon: the natural way to give each copy its
own function name doesn't work, because the name we generate contains hyphens and hyphens aren't
allowed in JavaScript function names. That rules out one of the two approaches, which is useful to
know before rather than after.

I also found two quiet ones: 58 of the 91 components have form labels wired to those names, and 33
style themselves using them. Neither produces any error if we rename the name and forget the label
or the style rule — the page just silently stops working properly in a small way.

**What I need from you.** I've written the proposal up as RFC_034. The question isn't whether to do
it — it's *how*, and there are three ways with real trade-offs: a deterministic converter that's
auditable but can't handle the script half; letting the AI rewrite each component, which can handle
it but has previously truncated a component and reported success; or a hybrid. I've recommended the
hybrid, done component by component with the existing safety check as the gate. But the choice
affects whether 94 live pages change over days or weeks, so it's yours.

I've deliberately not started converting. Building the converter before you pick a shape would risk
building the wrong one.

---

## 2026-08-17 (later) — you asked me to look it over once more, and the scary number was my instrument's fault

The "88 of 91 scripts need rewriting" I gave you earlier today is wrong, and the truth is much
better: **25**. Here is what happened, plainly.

The checker I built for this work — the one that will judge every converted component — turns out
to have had a blind spot. Most of our tool components start their code with a standard comment
block describing what the tool does. The checker looked at the first character of the code to
decide "is this safely wrapped?", saw a comment instead of the wrapper, and said "unsafe". Sixty-two
components were flagged that way, and every one of them was actually fine — properly wrapped, just
politely documented first. I found it because you asked for the second look: I opened one flagged
component by eye and its code visibly ended with the wrapper's closing bracket.

So the checker is fixed (with tests that prove both that it now sees through the comment and that a
comment can't be used to smuggle genuinely unsafe code past it), it has gone back to the reviewer
council as a third round on the same case, and the honest count is: **66 components need only the
mechanical renaming; 25 need the careful script work.** And those 25 are almost exactly the
calculators on the loan-and-mortgage site that this whole bug was filed about — the original "22
templates" estimate was close to right all along, arrived at from the other side.

Two things worth saying straight. First, this flips yesterday evening's framing back: the AI-rewrite
exposure is 25 components, 23 of them on the one site with an independent test harness watching, so
the hybrid plan is comfortable rather than uncomfortable. Second, if we had started converting
without this second look, the broken checker would have rejected 62 perfectly good results somewhere
in the middle of the programme — the exact moment when the temptation is to loosen the checker to
make the problem go away. Finding it before the programme, with the fix reviewed, is the cheap
version of that discovery.

The decision in front of you is unchanged in shape and easier in substance: the hybrid approach,
loan-calculator site first, now covers 73% of the estate with the reliable mechanical pass and
reserves the careful work for the 25 that genuinely need it.

---

## 2026-08-17 (evening) — your ruling is in effect, the converter is built, and today's redeploy didn't actually ship it

You chose the hybrid, loan-calculator first, everything through the framework. That last part
shaped the build: the converter is a platform action, so every conversion will be a recorded,
reviewable work item with a before-snapshot — not a hand edit.

It's built and tested, and the review board passed it with no objections. Two things from the
build worth telling you. The tests use real stored components rather than examples I invent, and
that habit paid for itself again: one real component's copy-buttons store the name of the element
they act on in a side attribute the code reads at runtime — a kind of reference none of my renaming
rules covered and no invented example would have contained. It's covered now. And the converter's
best feature is a refusal: any component whose script would still clash after renaming is rejected
untouched, because a half-converted page passes every check while still giving visitors the wrong
calculator's answer.

One thing to know about today's redeploy: it restarted the machines but served them yesterday's
software. The rebuild reused the old version number, and machines keep a cached copy per number —
a known trap here, which another team hit the same day. Nothing is lost; the converter simply rides
the next properly numbered release. Until then nothing has been converted and the original bug is
still live.

Next after that release: one end-to-end trial conversion, then the 66 easy components in batches
while we design the careful pipeline for the 25 calculators.

---

## 2026-08-18 — the trial conversion is done end to end, it found one more real gap, and the batch is now running

The trial finished, and it is worth telling as a story because the middle of it is the valuable
part.

The conversion itself was perfect — every one of the numbers I predicted in advance came out
exactly. The system then re-rendered and redeployed the page, reported success everywhere… and the
live page still had the old names. Nothing had lied, exactly: the "re-render" machinery has two
modes — a cheap one that reassembles the page from stored building blocks, and a thorough one that
rebuilds the blocks from their templates — and it picks the thorough mode only for a short list of
known causes. **"The template changed" was not on the list.** So every template fix this platform
has ever made shipped the old bytes with a green tick, and nobody had noticed because nobody had
checked the served page against a predicted change this precisely before.

That's fixed now — "template changed" is a recognised cause, and the fixer now asks for exactly the
affected pages to be rebuilt rather than all 111 pages of the site. The fix went to the reviewer
council as round four on the same case. (En route I also shipped a one-line mistake of my own — a
column name that doesn't exist, hidden inside a stored query where no check could see it — caught
minutes later, corrected, and turned into a permanent check so the next migration can't repeat it.)

With the fix in, the trial page went live properly: thirty-four new instance-scoped names on the
served page, zero old ones, and every copy-button correctly wired to its renamed target. The daily
alarm we built fired this morning right on schedule — that's the "architecture exception has
expired" notice doing its job, and it can be retired now.

**The batch is released.** Seventy components queued for the mechanical conversion — each one
converts, snapshots its previous state, and asks for precisely its own pages to be rebuilt. A
monitor is watching the queue drain. The twenty-five calculators that need the careful script work
remain untouched; designing that pipeline, with the loan-calculator site's test harness watching,
is the next piece of work.

---

**2026-08-18, evening.** The plan for the last quarter — the twenty-five calculators that need
real script surgery — is now written, and two facts found while checking it reshaped it. First:
all twenty-three loan-calculator pages are "owned" pages, which the rebuild machinery we proved
last week deliberately skips — so those conversions will be delivered through the section
editor, the same door the tool-fixing agent has used successfully forty-two times. Second: the
obvious shortcut (reuse that tool-fixing agent wholesale) turns out to be wrong, because its
safety fence would refuse one of our own calculators — so the careful work stays inside the
same agent that did the mechanical batch, which simply hands the hard cases to the language
model and then checks its work mechanically before anything is written.

The checking is the heart of it: the machine does all the renaming it is already proven at, the
language model only wraps and rewires the scripts, and a gate re-renders two copies of the
result and refuses to save anything that is half-done, altered outside its brief, or cut off
mid-generation. Refusals go to a human; nothing gets written on a doubt. The loan-calculator
site's own test harness — 170 arithmetic checks against the live pages — is the independent
witness, moved one tool at a time in step with each conversion, and that lane has been given
notice and a chance to object before the first small calculator goes through. Next session:
build it, put it through the review council, and run the first one.

---

**2026-08-19.** Building the careful-conversion pipeline turned up something worse and more
useful than the pipeline itself: yesterday's "finished" mechanical batch quietly broke a third
of what it converted. The converter renamed every element name and every direct lookup, and
checked its own work by searching for the patterns it had just renamed — but many of these
little tools pass their element names around through lists and variables before looking them
up, and those travelling names were never renamed. The page looks perfect; the calculator
underneath is dead. Thirty-two of the sixty-nine converted components have the fault and
fourteen are live right now, including a demo tool on our own web-design shopfront. Every
check we ran yesterday was green because every check measured what the converter changed, not
what the script still refers to — that lesson is now written into the permanent guides.

The good news: the fix is mostly mechanical and it is built. A new detector answers the right
question ("does anything still refer to the old name?"), it now sits inside the acceptance
gate so this can never ship silently again, a repair pass fixes twenty-seven of the thirty-two
by machine, and the five genuinely tricky ones join the language-model queue we designed
yesterday — which is also now built, gates and all. Everything is submitted for review and
waits only for the next release to roll out; the repair job is written so the broken-and-live
pages go first.

**2026-08-19, evening.** The release rolled and it carries everything we were waiting for —
checked three separate ways down to the bytes running in the pods, because we have been burned
before by "new" deploys serving old code. The review council had already approved the whole
package on its ninth round that afternoon. So both switches were flipped tonight: the careful
LLM branch is live in the fixer, and the repair batch was seeded — sixty-seven jobs, one per
converted component, with the pages currently serving broken bytes pushed to the front of the
queue. One honest surprise in the numbers: forty-two components are now serving converted bytes,
against fourteen when we measured this morning. That is not new damage — pages have simply kept
re-rendering onto the converted templates all day, which is exactly why the repair could not
wait, and most of those forty-two are sound anyway (the counter can't tell "serving the new
style" from "serving broken" — only the repair pass can, and it no-ops politely on the sound
ones).

The fleet is working through the queue on its own as I write: real repairs are landing with the
right shape, sound components are being waved through, and each genuine fix files the re-render
that carries it to visitors. One first: a fix just went out to an owner-managed page through the
new hand-off we built for exactly that case — the one link in the chain we had never seen run
for real. If that page re-renders and deploys cleanly, the last unverified box gets ticked.

**2026-08-20.** A very full day, and the pipeline earned its keep three times over. First, the
repair batch from last night finished: of sixty-seven jobs, twenty-eight components were
genuinely repaired, thirty-five were checked and waved through as already sound, and four were
refused by the safety gate and parked for a person to look at — which is exactly the shape we
designed for. Every repaired page we checked, we checked at the actual bytes being served, not
at the job status; one live page was fetched from the internet and read by hand. The one loose
end from that batch: three pages on the AI-consultancy sites still carry the broken
savings-estimator, because its scripts pass names around in a way no machine rewrite could
prove safe. Those three wait on your decision — roll the component back to its pre-conversion
snapshot, or let a person fix its script. Nothing else is serving broken tools.

Then the canary. The first loan calculator went through the full careful pipeline: the language
model rewrote its script, the gate checked the rewrite eighteen ways, the section editor
re-rendered and republished the owner-managed page, and the arithmetic referee — 170 checks,
run before and after, with a deliberate-sabotage control both times — came back all green both
sides. The calculator behaves identically to the penny; only its internal plumbing changed. On
the strength of that, the remaining twenty-two calculators were queued through the same pipeline
this evening and are converting as I write.

Two things went wrong today, neither ours and both instructive. Another team armed a new
feature this morning that accidentally broke every page-publishing job on the fleet for
half an hour — they caught it, rolled it back, fixed it properly and re-armed it by mid
afternoon; our canary's publish was collateral and simply retried. And when they revived the
jobs their outage had killed, one of ours came back looking queued but was actually
unrunnable for ever — a subtle bookkeeping trap (the retry counter was spent) that we
diagnosed, fixed in one line, and wrote into the shared trap register so nobody hits it again.
The queue picked our job up within a minute of the fix.

**2026-08-20, late.** The last three tools went through tonight and the planned conversion is
done. Every one of the twenty-four components the careful pipeline touched is live and proven:
the arithmetic referee reads identical-to-the-penny across the loan site, and on the three
game-design tools — which have no referee — we drove the live pages in a real browser and
watched them respond: type into the ranking tool and eight figures recompute; press the clash
calculator's button and the results panel fills. That responding-at-all is exactly what the
broken batch from last week couldn't do, so it is the right thing to have watched. What's left:
six components the gate refused now wait on people — three of those are the broken savings
estimator that needs your rollback-or-hand-fix call — and the estate keeps minting new tools
faster than anyone converts them, which needs a decision about whether the front door enforces
the rule at birth or a sweeper converts arrivals on a schedule.

**2026-08-21.** You asked the right question — why do unconverted tools keep appearing — and
the answer was embarrassing enough to write into the permanent mistakes log: we fixed every
tool that existed and never once asked what keeps making new ones. The tool-writing agent had
simply never been told the naming rule; it was writing tools the way every web tutorial does.
Twenty-three arrived in three days, seven on the day we finished the backlog.

All three of your rulings are done. The generator's instructions now teach the two habits that
matter (keep the script self-contained, name elements plainly). The stronger half is in the
code, per your preference: the save-path now runs the same proven machine that converted the
estate over every newborn tool — a well-shaped tool is silently converted before it is saved,
and one the machine can't prove safe is rejected so the generator simply tries again; nothing
half-done can be born. And a small daily sweeper now walks the whole library each morning and
files conversion work for anything that slipped in by any other door; its very first run found
twenty-eight, filed twenty-six, and the pipeline converted twenty-five of them within the
hour. The twenty-sixth is an old friend: a game tool that has always carried the same label
twice inside itself, which the safety gate rightly refuses to paper over — it joins the short
repair list a person needs to look at. The whole package is with the review council now.

**2026-08-21, night.** Your two decisions are done and the story has a good ending. The shared
savings estimator was retired, and each of its two sites got a brand-new tool born through the
full pipeline — and this was the first real test of the new birth guard, which passed both
ways: one site's tool came out correctly named on the first try and the guard converted it as
it was saved; the other site's first two attempts wrote their code in a style the machine
can't prove safe, the guard refused them both — nothing broken was ever saved — and the third
attempt, asked more firmly to name elements plainly, sailed through. Both pages are live and
correct. **No page anywhere serves a broken tool any more.**

One small tail: the finetuning page works but is still on its old naming, because its
re-publish was stopped by the honesty checker — an empty content field would have rendered as
"0% reduction", and the system refuses to publish a claim like that. That is two safety nets
catching each other's cases, which is what they are for. A content refresh is queued; the
page re-publishes itself after that. Tomorrow: the two loan-site calculators rebuild once
their team's window closes, the morning sweep should report a clean estate, and then both bug
files can finally close.

**2026-08-22, close of play.** Both bug files are closed. The week's arc, in one breath: we
found that no calculator on the estate could ever appear twice on a page; we converted every
existing one (proving the loan site identical to the penny with an independent referee); we
found and repaired the damage our own converter had quietly caused; we taught the tool factory
the rule, put a guard on both of its doors, and set a daily sweeper behind every other door —
and today the machinery ran without us: the sweeper's own escalation rebuilt three stubborn
tools, the guard refused the bad drafts and accepted the good ones, and the rebuilt debt
consolidation calculator came out computing to the exact penny the referee expects. Two small
things wait on people, both harmless meanwhile, both filed where the right person will find
them. Along the way we broke another team's build for three hours by removing a setting we
wrongly thought was ours — owned, attributed, and written into the permanent mistakes log with
the check that prevents it. The referee's expected score is now 166, the loan team has been
told why, and the lane is done.

**2026-08-22, late afternoon (a second session, running alongside the one that closed it).**
You asked me to look at 283 fresh, check nobody else was on it, and check it was still real.
Two of those answers turned out interesting, so here they are plainly.

**Somebody else finished it while I was still reading.** I checked for other sessions working
this bug and found none — no session was even *named* after it. That check was not good enough:
another session closed the bug at ten past four, and I only found out because I re-read the
project history before touching a file. The close is right and I have not argued with it. The
thing it fixed — calculators whose element names are fixed text, so two on one page fight over
them — really is fixed and really is live. The lesson worth keeping is that "no session is
working on this" is not something you can establish by looking at session names.

**The bug was still real, but not in the place it was filed.** The estate has *two* systems for
naming an element uniquely, and 283 only ever fixed one of them. The older one is still running
and it does not work: it hands every copy of a component the same name. Today, eighteen live
pages serve duplicate names, one of them six times over, and eleven more serve a name that is
literally empty. I fetched the pages and looked, rather than trusting the database — and I
fetched two pages that *should* be clean as a check on my own method, because a test that
cannot come out the other way is not a test.

Nothing on those pages is broken for a visitor. They are all text sections, so nobody sees a
wrong number. It is the same shape the original bug had when it was filed a week ago: a wall
rather than a fire — it quietly stops us putting two of something on one page, and it will
become a fire the first time someone tries.

**You made a decision and I have written it down.** Rather than patching the old system, we are
retiring it: the five templates that still use it move onto the new one that already works, and
then the old one is deleted. One way of naming things, not two. That is now recorded in the
architecture file where the question was raised.

**Then the fix turned out to be impossible, which is the useful part of today.** The plan was
to run those five templates through the same conversion machine that has already done a hundred
and twenty-four. It would not have worked, and — this is the bit worth pausing on — it would
not have *told us* it had not worked. The machine looks for element names written as plain
text, and these five write theirs as a placeholder, so the machine sees no names at all and
politely reports "nothing to do here". Five jobs would have completed, the queue would have gone
green, and not one template would have changed. Nobody would have had a reason to look.

So what I have actually built and committed today is the smallest thing: teaching that machine
to see a placeholder. It also refuses, now, in a case it used to get half-right — where it would
convert some names, silently skip the placeholder, and produce a page whose safety check passes
while the page is still broken. That one cannot happen today; it is there for the next one.

I proved the six new tests by breaking the code on purpose and watching every one of them fail,
then putting it back. A test that has never failed is not evidence of anything.

**Where that leaves us.** Nothing a visitor sees has changed yet, and I want to be exact about
that: the eighteen pages still serve duplicate names this evening. What exists is the capability
to fix them, sitting in the code waiting for the next release. The conversion itself, the
deletion of the old system, and the re-render of those pages are the next steps and they are not
done. The review council has the change and has not reported back yet.

**2026-08-23.** Picking up where yesterday evening left off. Four things happened, and one of
them is a mistake of mine worth reading before the good news.

**The review council approved the change, and then caught something real.** Yesterday I taught
the conversion machine to see a name written as a placeholder rather than as plain text. The
council approved it but asked a pointed question: that machine is also used by the two guards
that sit on the door where new calculators are born — did anything downstream depend on the old
refusal message? I went and looked instead of reasoning about it, and yes, something did. One
piece of code decides what to do by reading the *words* of the refusal, and it had a branch
meaning "nothing here to worry about, save it as-is". A template of exactly the shape we are
fixing would have taken that branch and been saved, broken, by the guard built to prevent it.
Fixed, with tests that fail if either half is removed.

**Then the council caught me making the same mistake again, in the fix.** My repair still worked
by reading words in a sentence — so if anyone ever rewrites that sentence, the hole reopens
silently. The reviewer pointed out I had a better option sitting right there for free. What
stings is that I had spent the morning writing that exact lesson into our permanent traps file:
*route on a fact, not a sentence.* I wrote the rule and then did not follow it an hour later. It
is fixed properly now — the two pieces of code pass a plain yes/no fact between them, and if
anyone renames it the build breaks instead of a live page. The traps file now records that
episode too, because the gap between writing a lesson down and applying it is the actual lesson.

**The conversion itself is running, and the system is doing it for us.** All four of the
templates in use are converted. Better than that: converting a template makes the system queue
up a refresh for every page using it, so two hundred and nineteen page refreshes went into the
queue by themselves and have been working through steadily. I checked the first finished page —
apis.uk's homepage, which was serving the same name six times over — and it now serves six
distinct ones with all its text intact. One template, `pricing`, is not converted: it is not used
on any page, and the way we file this work needs a page to hang it on. Harmless today, but it
must be done before we delete the old naming system, or the first page to use it comes out blank.

**And the mistake.** We have been saying "eighteen pages are affected". I repeated it in a plan
and in a commit message. Then I actually fetched all eighteen, and it is twelve. Three of them
were never broken — they happen to supply their own names — and three are not reachable at all
(two are missing, one is a parked domain that redirects). The number came from asking the
database what it *would* build rather than asking the pages what they *are serving*, which are
different questions. Nobody was careless: it was written down, dated, and corrected once
already. It was just counting the wrong thing, and being dated made it look more checked than it
was. It is in the mistakes log with the ninety-second check that would have caught it.

**Where that leaves us.** Twelve pages needed fixing; the first is confirmed fixed and the rest
are working through the queue. Nothing a visitor sees has got worse. Still outstanding, and
deliberately not done today: the unused template above, deleting the old naming system (which
must come last), and one deeper issue — when a single section is re-rendered on its own rather
than as part of a whole page, the system assumes it is the only copy on the page. For almost
everything that is true. For exactly these templates it is not, so editing one section of one of
those twelve pages could reintroduce a clash. It is visible when it happens rather than silent,
and it is written up as the next piece of work rather than quietly absorbed into this one.

**2026-08-23, evening.** You asked me to convert the unused `pricing` template and then sort out the
bindings. Both are done, and the second half turned up something that changes the picture.

**Pricing, and two others I hadn't seen.** When I went to convert `pricing` I widened the search to
every place in the database that could hold a template, rather than the one table I'd checked
before. There were three of these old-style templates, not one: `pricing`, plus a `header` and a
`footer` left over from before the current chrome system. The latter two are switched off and used
nowhere, but I converted them anyway — once the old naming is gone, switching one back on would
produce a blank name that none of our checks can see. All three are converted, and the count of
templates still using the old style is now zero.

**A trap I nearly walked into.** Two of the files that originally created these templates are not
recorded as having been run, and they overwrite on re-run. So the next time anyone runs our database
update script, it would have quietly put the old naming back into two of the templates we fixed this
morning. Fixed by correcting those two files.

**The bindings: two of three removed.** The third lives in a file that another session is currently
working in, and their work-in-progress calls three functions from a file they haven't committed yet.
Because of how our commit rules work, committing that file would have taken their half-finished work
along with mine and broken the build for everyone. So I left that one, and it's written down in
three places so it gets finished. It's harmless meanwhile: nothing uses the old naming any more, so
a leftover connection to it changes nothing.

**And the thing you should actually know.** The fix does not stay fixed. Of the twelve pages that
were serving duplicate names this morning, nine are fixed and three have already gone back to
serving duplicates — not through anything I did, but because another team's routine content update
ran over them at ten to six this evening. When any part of our system re-renders a single section on
its own, rather than a whole page at once, it assumes that section is the only copy on the page. For
almost every component that's true. For these it isn't.

So the honest position: the underlying template work is done and will stay done, but individual
pages will keep flipping back until that assumption is fixed, and nothing tells us when they do. I
have written it up as the next piece of work with a live example anyone can check in ten seconds.

**One more thing worth flagging, because it would have fooled a status report.** Every summary
number looks healthy: all four templates converted, 244 of 275 page sections carrying the new
naming. Those numbers stay healthy while pages break, because two sections with the *same* new name
still count as two sections with the new naming. The only check that sees the problem is asking a
served page whether its names are actually different from each other.

**2026-08-24, morning.** The new build is out and I checked it the hard way: the running system
provably no longer contains the old naming machinery — the retired phrase is absent from the
binary itself, with a check that would have caught me if I'd been wrong. Everything we shipped
this week is live.

**The review came back against us, and the reason is embarrassing but instructive.** The
reviewers rejected Saturday evening's cleanup — not for the cleanup, which they called sound,
but because my submitted change carried a chunk of *someone else's* half-finished work without
saying so. Another team was editing the same file at the same time; I'd checked the file was
clean at midday and didn't re-check before committing in the evening. Their code rode along
into my commit, into my review, and into production — and it compiled and worked, which is
exactly why nothing flagged it. The reviewers read it as me sneaking in an undisclosed feature
and vetoed the lot. Fair, on what they could see.

What I've done about it: told the other team their code shipped under my name and that a review
panel discussed their mechanism without them; written the mistake into the permanent log with
the ten-second check that prevents it; closed the one loose end the reviewers were right about
on the merits (two old setup files that could have silently undone Monday's fix are now
formally marked as done, so they can never re-run); and resubmitted the cleanup for review as
exactly what it is, nothing more. That verdict is pending.

**One piece of housekeeping keeps refusing to die.** The third and last connection to the old
naming still exists in one file. It's harmless — nothing uses the old naming any more — but I
can't remove it yet, because that file *again* has someone else's unfinished work sitting in it,
and committing it would repeat Saturday's exact mistake the day after it cost us a veto. On top
of that, the removal I'd drafted on Saturday was wiped out overnight when another team
overwrote the file — a known hazard of how we share this codebase. It stays on the list with
precise instructions for whoever finds the file clean.
