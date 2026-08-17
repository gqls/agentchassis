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
