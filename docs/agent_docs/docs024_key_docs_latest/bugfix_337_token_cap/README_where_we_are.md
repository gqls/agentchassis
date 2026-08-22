# Where we are — bug 337 (the tool section that always blows its token limit)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-22 — picked up the bug, confirmed it is still real, measured how big the problem actually is

Bug 337 is about one kind of page section — the "credit health check" interactive tool —
that the AI writer can never finish. The writer is allowed to produce at most 16,000
tokens of output per attempt (roughly 48,000 characters), and this particular section
needs more than that, every single time. The system correctly refuses to keep a
half-finished component (that safety rule exists for good reasons), retries twice more,
hits the same wall twice more, and gives up. Result: two live loan-industry pages have
no working calculator on them, and every future loans site that plans this section will
lose the same page the same way.

What I checked today:

- **The bug is still live.** The limit is still 16,000, the failed work items are still
  parked (plus a fourth occurrence the bug file didn't know about), and nobody else is
  working on it — the team that filed it deliberately left it for someone to take.
- **The limit is too tight even when it works.** Looking at the last two weeks of
  successful runs of this writer: a twentieth of them come within 15% of the ceiling.
  So this isn't one greedy section — the ceiling was set too low for the job generally.
- **Other parts of the system have the same disease.** Several other AI steps run within
  10% of their own ceilings, and a handful have hit them outright over the past month.
  Whatever we do here should help all of them, which matches the direction you suggested:
  raise the limit, and put proper management around these limits rather than picking
  numbers by folklore.
- **A bigger limit is safe with this AI model.** Other steps in our own system already
  run the same model with limits of 32,000 and 64,000 without trouble, including another
  step that writes tool HTML and was given 32,000 from the start.

The plan I'm preparing (details in the PLAN file) has three parts: raise this step's
limit to a measured number; teach the framework to notice "the answer was cut off" and
retry once with a higher ceiling instead of failing three times identically; and add a
small daily check that watches every AI step's real output sizes against its ceiling, so
we find out a limit is getting tight before pages start failing, not after.

## 2026-08-22 (later) — the fix is written, tested, committed, and in front of the reviewers

Everything described above is now built. The three parts became two, and that change is
worth explaining: while researching, I found the daily "watch every AI step's output
sizes" check I planned to build **already exists and had already noticed this problem**
— it has been writing a note about this exact step since the 18th of August, four days
before anyone acted. So the missing piece was never detection; it is that nobody and
nothing reads those notes. I have recorded that gap where the next person will find it,
rather than building a second watcher that would be ignored the same way.

What shipped: (1) the writer's limit goes from 16,000 to 24,000 — a number derived from
what its outputs actually measure, not a guess; (2) a new framework ability: any AI step
can be given a second, higher "ceiling" limit, and if its answer gets cut off at the
normal limit it retries once at the ceiling instead of failing three times identically —
switched off for everyone except steps that explicitly opt in, and the loans writer is
the first. The code is committed and goes live at the next fleet release; the limit
change itself is a configuration change I will apply once the review council has looked
at it (they are reviewing now — their verdict usually lands within the hour). After
that, I re-run the two broken loan pages and check the live pages actually gain their
calculators.

## 2026-08-22 (evening) — the fix was approved and applied, and then testing it showed the bug was misdiagnosed. That is the important part.

Two things happened, and the second matters more than the first.

**First, the work landed.** The review council approved the change (eight of its
thirteen reviewers had nothing to object to; four raised advisory points, two of which
were fair and which I fixed rather than argued with — one caught me stating something I
had not actually measured, the other asked a good question about reusing existing code).
The limit change is now live on the system, and the retry-at-a-higher-ceiling ability is
committed and switches on at the next fleet release.

**Then I tested it against the real pages, and the test contradicted the bug.** The
regenerations succeeded — but they came back *smaller than the old limit*, which cannot
happen if this section genuinely always overruns it. So I ran the count nobody had run:
of **82 attempts at generating this section, 73 succeeded comfortably** under the old
16,000 limit and only 9 were cut off. The limit fits about nine times in ten.

What is actually breaking these pages is a different thing entirely: the writer keeps
inventing a reference to a piece of site data that does not exist ("ctas"), a safety
check correctly refuses to store the result, and the system quietly regenerates — over
and over. The three stuck pages had generated **11, 13 and 55 times** while the system's
own counter said "3 attempts". Occasionally one of those many generations runs long and
hits the token limit, and that is the failure that happened to be recorded last — which
is why the bug was written up as a token-limit problem.

**How I got it wrong, in plain terms:** the bug file's evidence was a search of failed
jobs for the phrase "hit the token limit". That search can only ever return jobs whose
*final* error was the token limit, so it could not have shown me the 73 successes even
if I had looked harder — they were in a different table. I inherited that figure, put it
in front of the review council, and was approved on it. I have logged that mistake in
our shared record of wrong calls, and written the underlying trap into the debugging
guide so the next person hits a check instead of the same wall.

**Where that leaves things.** The limit raise and the new retry ability are still worth
having and I have not reverted them — the step genuinely was running close to its
ceiling, and 9 wasted full-price generations in 82 is a real loss. But they do **not**
fix these pages and I have not claimed they do. The bug is re-scoped to the real cause
and pointed at the team already working on that loop; I have told them, including
warning them that today's limit change will make their "before and after" numbers look
better for reasons that are mine, not theirs. One page's component did store today and
is being attached now; the other is still blocked by the invented-data-reference problem.

## 2026-08-22 (late) — one page is genuinely fixed and you can look at it; the other needs the other team's fix

**Fixed and live: https://loancalculator.co.uk/tools/credit-roadmap.html** — it now carries a
working "credit health check" quiz (click-through questions, a scoring meter, a restart
button). That page had been missing its tool since mid-August.

**Not fixed: loanzy.uk's credit-health-check page.** Its regeneration ran perfectly well
today — no token problem at all — and was then refused at the last step because the writer
referred to a piece of site data that doesn't exist. That is the real bug, it belongs to the
team already working on that loop, and I have handed it to them with the evidence.

Two things I got wrong while checking my own work, both now written down so the next person
doesn't repeat them:

- **I recorded a "before" measurement on the wrong web address.** Different sites in the
  fleet use different URL styles, and the address I guessed returns that site's "page not
  found" page. That page is real HTML, it's identical every time you fetch it, and it
  contains none of the things I was counting — so my "check it twice to be sure" step
  confirmed it happily. What I never checked was whether the server said 200 (found) or 404
  (not found). Checking that it doesn't change is not the same as checking it's the right page.
- **I was using the wrong test for success.** The bug file said to count text-entry boxes on
  the page. This particular tool is a button-based quiz with no text boxes at all — so a
  perfectly working page scores zero on the stated test. I'd inherited a test written for a
  calculator-shaped tool.

Neither changed a decision, because I caught both while verifying rather than after
reporting. But they are the same species of mistake as the big one earlier today: a
measurement that is accurate about the wrong thing.

## 2026-08-22 (evening) — the bug was mis-named twice, and the real one is that we never tell the writer the rules

Picking this up fresh, I found the note the last session left: "start at the validator, not
the token limit." That was the right steer. But the thing it pointed at turned out to be the
wrong culprit too, and so did my own first attempt — so this entry is mostly about how we
kept naming the wrong thing, because that is the part worth remembering.

**What is actually going on.** When the system writes a new section for a page, a safety
check inspects the result before storing it and refuses anything that breaks one of two
rules. The trouble is that **we never tell the writer what those rules are.** It is asked to
produce something, judged against a standard it has not been shown, and refused. It then
tries again with exactly the same information — so it fails again. Over eight days that
happened 101 times across four sites, and eleven jobs are now sitting permanently stopped,
each one a page missing its calculator.

One of the two rules is worth spelling out, because the failure is almost comic. The writer
has to say where each piece of text comes from — some are written by the AI, some are fixed
labels, and some are pulled from the site's own stored details. For that last kind it has to
name which store to read from. **It has never been shown the list of stores.** So it guessed
"ctas". The real name is "cta" — one letter — and that store contains exactly the two things
it was reaching for. The instructions do contain a list of valid names for a *different* kind
of lookup, with the words "use these exactly, do not invent new ones", and that list has
worked: nothing has invented one of those since it was added in May. The other list simply
was never written down.

**How we got the culprit wrong, twice.** The previous session blamed a specific kind of
refusal. I counted them: that kind accounts for 3 of the 101. The other 97 are the other
rule. Then I did the same thing myself — I told you 52 components were stuck, from a count
built the wrong way round. When I ran the obvious check against my own number, 21 of those 52
had worked perfectly well. The honest figure is that **none** of them is stuck today, because
a fix from another bug fixed most of it as a side effect three days ago without anyone
noticing. Both mistakes are written into our shared record of wrong calls. The pattern in all
three is the same: we each stopped counting the moment we had an explanation that fitted.

**What I have built.** The writer is now told both rules, and — this is the part that matters
— it is told them by *the same code that enforces them*, not by a second copy that would
drift apart over time. We have done this exact thing before, successfully, for a different
writer that was never told which pages existed, so this is a proven approach rather than a
new idea. I also fixed a small trap where a component could get into a state that made it
invisible to the thing that repairs it, so it could never climb out.

**What I have deliberately not claimed.** The problem I fixed is *rare* — this kind of
refusal has happened three times in four days, and no bad component has actually been stored
since May. What changed recently is not how often it happens but what it costs: a check we
switched on four days ago turned a quiet, invisible failure into one that stops the page
being built at all. That is a real improvement and this fixes its side effect. It is not a
flood, and I would rather you heard that from me than found it in the numbers.

Another team's change landed in the middle of my work and does part of the same job from the
other end — after a refusal, it now tells the writer what went wrong so the next attempt can
differ, and it has already turned one refusal into a success. That narrows what mine adds to:
the detail their message does not carry, the first attempt saved rather than spent, and the
eleven jobs that have already run out of attempts. I have said so in the review submission
rather than letting it look bigger than it is.

**A mistake I made that cost someone else time.** While testing, I used a command to undo a
one-line change of my own. On a shared machine that command does not undo *your* change — it
throws away everyone's unsaved work in that file. I destroyed about seventy-five lines
another session had written and not yet saved, and there was no way to get it back from any
backup. I told them within a minute, naming exactly what was lost so they could retype it
from memory rather than work it out again, and they had it restored in a few minutes. I have
written it up as a trap for everyone else, because the command looks completely harmless and
succeeds silently.

**Where this stands.** The code is committed and the wording change is live, but the code
part only takes effect at the next rebuild. You asked me to repair all eleven stopped jobs;
I will, but they have to wait for that rebuild, because re-running them now would prove
nothing. The review council has the change and I will act on whatever it says.
