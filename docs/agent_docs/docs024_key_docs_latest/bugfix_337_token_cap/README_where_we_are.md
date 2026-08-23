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

### One thing you need to know that is bigger than this bug (2026-08-22, ~18:30)

While resubmitting to the review council, the run died with this from Anthropic: **"You have
reached your specified API usage limits. You will regain access on 2026-09-01."** That is an
**account-level spending cap, not anything to do with my change**, and it is not confined to
my work — ten calls failed that way in a single hour this evening, the first ones today.

Until the cap is raised, **anything in the estate that calls the AI will fail**: the review
council, the diagnosis loop, and every site build that writes content. Nothing is broken and
nothing needs repairing; it will simply stop working until you lift the limit or 1 September
arrives. I would rather flag it now than have you find it as a wave of failed builds.

## 2026-08-23 — the rebuild went out, my change is live and working, and it has moved the problem one step rather than finishing it

**The fix is genuinely in the running system and genuinely doing its job.** I checked that
properly rather than assuming: the running service reports which version of the code it was
built from, my change is inside it, and I used a check that could have said "no" (a later
change correctly showed as absent). Before, the writer was handed nothing; on the first run
after the rebuild it was handed both things it had been missing — the list of field names it
must keep, and a 10,000-character list of the valid data sources it may use.

**And the specific mistake this bug was about has stopped.** Yesterday the same job was refused
because the writer invented a data source called "ctas" that does not exist. Today, shown the
real list, it used only real ones. That failure did not come back.

**But the page still is not fixed, and I want to be straight about that.** The job was refused
again — for a *different* reason, and by a rule that has been there since the 3rd of August,
nothing to do with my change. The writer declared forty-three pieces of content and then only
placed forty-two of them in the actual page layout. One missing. The safety check spotted it
and refused, correctly.

So: one field short of forty-three, rather than an invented data source or eighteen missing
names. Much closer, still not there.

**One thing I genuinely don't know and won't pretend to.** My change tells the writer to keep
all forty-three names. It is possible that asking it to juggle forty-three makes it more likely
to drop one than if it had invented a smaller set of its own. The evidence leans against me
having caused it — that failure has happened 58 times to 24 different jobs since the 3rd of
August, long before my change — but I have not proved it either way, and I have written down
the measurement whoever picks this up should run.

**The good news underneath it:** the other team's work now captures the exact reason for that
refusal and hands it to the next attempt, so the retry is being told precisely which field is
missing. If that works, it will be the first time we have seen prevention and correction close
a loop together. It is still queued, so I am not claiming it yet.

**A correction to something I told you yesterday.** You approved re-driving all eleven stopped
jobs. I checked before spending anything and the picture has changed overnight: there are nine
now, not eleven, and other teams have already repaired most of the pages — I confirmed that by
looking at the live sites, with working pages as a comparison. **Only three loanzy pages are
genuinely still missing their tool.** Worse, every one of those stopped jobs now has a
component that exists, so re-running them would *rewrite* components other pages are using
rather than create missing ones — nine expensive runs to fix three pages, with a real risk of
breaking six working ones. I have run one, and I am holding the rest for you rather than
following yesterday's instruction into a situation it was not written for.

## 2026-08-23 (late) — two of the three pages are fixed, and you can look at them

**Fixed and live:**
- **https://loanzy.uk/tools/is-a-loan-right-for-me/index.html** — now carries its checker (four
  input boxes and its own logic, where before it had none).
- **https://loanzy.uk/tools/eligibility-checker/index.html** — now carries the credit-health
  quiz: thirteen buttons and four thousand characters of working logic, where before it had one
  button and none.

The second one matters most, because it is the whole chain working: my change showed the writer
the list of valid data sources it had never been given → it stopped inventing one → it was
refused once for a small, different mistake → the other team's change told it exactly what that
mistake was → the retry got it right → the component was stored → the page picked it up. Every
link in that had to work.

**Not fixed: the credit-health-check page itself** — the one this bug is named after. Its
rebuild was refused by a *third* safety check, and that check is right. Rebuilding it would
have stripped well over half the layout from the top section of the page, and there is a
guard whose whole job is to stop exactly that. I have not overridden it. Forcing my repair
through a guard designed to prevent damage would be the wrong instinct even when the repair is
genuine.

I did find something useful for the team that owns that guard, and passed it to them: it has
refused five loanzy pages today, always the same section, and always collapsing to exactly the
same size. That pattern says one component's rebuild is systematically thinner than the stored
version — which is a fixable thing, and different from "these pages are all degrading".

**One thing I got wrong and caught, worth telling you because it is the third of its kind.** I
briefly reported that neither repaired page had actually gained its tool. They both had. I had
searched the page for the wrong kind of HTML tag — I looked for the shape I expected rather
than for the thing itself. What caught it was checking the database before writing it up. That
is now three times in this lane I have used a test that would have called a working page
broken, and I have written up the common cause rather than the three incidents.

**Where this leaves the bug:** everything it was filed for is done except one page, and that
page is waiting on another team's guard.

---

## 2026-08-23, evening — the bug was already fixed and we hadn't looked

Short version: bug 337 is finished and closed. The last page repaired itself this afternoon,
about twenty minutes after we last checked on it, and nobody went back to look. I found it by
starting the next job rather than by working on 337 at all.

Here is what happened, in order.

This morning's work went live and did what it was supposed to. The page-building agent now gets
told, up front, what names it is allowed to use — which is the whole fix. Three pages on
loanzy.uk needed rebuilding as a result. Two of them rebuilt cleanly.

The third one, the credit health check, was refused by a safety guard. That guard exists to stop
a rebuild quietly making a page thinner than it was, and it was doing its job, so the earlier
session correctly refused to switch it off. It wrote a handoff saying "this page is blocked,
here is who needs to fix the guard", and stopped.

But the refusal wasn't permanent. It was the kind of failure that clears if you simply try
again. The system did try again, on its own, twenty minutes later — and it worked. The page has
been live and working since twenty past three this afternoon. The handoff saying it was blocked
was written at quarter past five, nearly three hours after the page fixed itself.

I checked the page on the actual website before believing either story: it loads, it's about
6,000 bytes bigger than before, and it has the full working quiz on it with the right questions.
So that's three pages out of three, and the bug closes.

The second thing is more useful than the first.

The handoff also left instructions for the next person: go and find out what is deleting content
from these pages. That sounds like a serious bug and it would have been someone's whole day.

**There is nothing deleting anything.** The hero section at the top of these tool pages has room
for three optional statistics. Sometimes the writing agent fills in all three, sometimes two,
sometimes none. When it fills in fewer, the page has fewer bits of styling on it — and the
safety guard counts bits of styling, so it sees a thinner page and refuses.

The proof that nothing is being deleted is simple: the number of statistics goes **up** as often
as it goes down. Three of the four pages I could check went from zero statistics to two or three
on their next rebuild. Something that deletes content cannot also add it.

So I've cancelled that investigation and written down why, and I've been careful to say what
evidence would have proved me wrong.

The part I'm least comfortable with is that this was already known. Another team wrote exactly
this conclusion into the same bug file yesterday — and our own note was added underneath theirs,
today, saying the opposite. We read the file and missed the correction sitting in it. I've logged
that, because it's a mistake that will happen again unless it's visible.

Two other near-misses today, both caught before they went anywhere.

I nearly reported that two loanzy pages were showing the wrong tool, because two pages share the
same underlying component. They aren't — the component is a shared template and each page fills
it with its own questions. Sharing components is how the estate is meant to work; I was reading
the label instead of the contents.

And I nearly reported this morning's fix as *not deployed*. The usual way of checking had gone
stale, so I checked the running program directly — and got "not found". The reason that isn't a
result is that I also ran a control that was supposed to come back "not found", and it did too.
When the check can't tell the two apart, the check is broken, not the thing being checked. I
threw the measurement away and confirmed the fix a different way, by watching it actually do its
job twenty minutes ago.

Nothing needs a decision from you. 337 is closed, 253 has one fewer open question than it had
this morning, and I've left the next session a corrected handoff rather than the one that would
have sent them looking for a bug that isn't there.
