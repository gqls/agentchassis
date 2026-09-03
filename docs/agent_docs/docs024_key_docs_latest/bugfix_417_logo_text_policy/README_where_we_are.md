# Where we are — the logo that named itself

*(Plain-prose log for the owner, append-only, newest at the bottom.)*

## 2026-08-31, evening

The short version: the site planner was telling the image model it could put words in a logo,
but never telling it what the words should say. So the model made a name up. On
farmerinsurance.uk it invented "Farm Shield Info". On boxingonline.com — your first paying
customer — it put **"BOXING NEWS"** on a site called Boxing Online, and that logo has been
sitting in the header of all 19 pages, with the favicon and the social-share card both cut
from the same picture.

Another session had already found this and fixed the planner this morning, and washed the
prompts that had copied it. Both fixes were right. Both were also, by this evening, provably
not enough, and the way they fell short is the interesting part.

**First:** boxingonline's logo instruction was written **41 seconds after** the planner was
fixed — by a job already in mid-air. You cannot fix a source and thereby fix the things that
already left it.

**Second, and worse:** the wash searched for the exact sentence the planner used. Boxingonline's
copy had been *reworded* — "no text other than the wordmark" instead of "no text outside the
wordmark". Same meaning, different words, invisible to the search. The model doesn't copy
instructions, it paraphrases them, so **searching for the exact wording will always find some of
them and never all of them.**

That pushed me to a different kind of fix. Rather than chase the wording, I moved the rule to the
one place every logo request has to pass through on its way to the image model, and made it apply
there regardless of what the instruction says. It doesn't need to *recognise* the bad wording;
it overrides whatever it finds. That also covers requests already sitting in the queue, which no
amount of fixing the planner could have reached.

One thing I got wrong and want on the record. I started from a note in our own files saying the
image provider ignores "don't do X" instructions. It turned out that was true *before* we fixed
it, months ago, and we did fix it. The proof was in the logs: for the boxingonline logo, the
model was **explicitly told "no text"** — and lettered BOXING NEWS anyway, because the same
prompt also said it could have a wordmark. A permission beats a prohibition in the same
sentence. My fix didn't change, but my reason for it did, and the old reason would have pointed
a reviewer at a cheaper repair that could never have worked.

You ruled tonight that logos carry no words. I've built that as the default everywhere, with one
deliberate exception you can use: a logo may carry lettering only if someone says **exactly what
it must say**, and the system checks that text really is that site's own name. That keeps
farmerinsurance's wordmark — which you asked for — legal and repeatable, and it makes "put a
wordmark on it, I don't care what it says" impossible to express, which is the sentence that
caused all this. Worth knowing: eight sites already have deliberately worded logos, so a blanket
no-words rule would have quietly broken them the next time they regenerated.

Where it stands: the wash for boxingonline's instruction is **live now**. The main fix is
committed but sleeps until the next chassis build goes out. It's with the review council.

Two things still owed, neither of them mine to do: boxingonline's logo needs regenerating before
you hand the site over (the delivery thread owns that), and after it lands the favicon and social
card have to be re-cut, because nothing re-does those automatically.

And one thing you should probably see for yourself. Beyond the invented name, that logo file
isn't a logo at all — it's a **two-panel presentation board**: the mark on dark navy on the left,
the mark plus lettering on light grey on the right, both squeezed into the header slot. Even with
the words gone, it's unusable. Nothing between the image model and the live page ever checks that
a logo is *one picture*. That's a separate fault from this one and I've written it up separately
rather than bolting it on here.

## 2026-09-02, afternoon

Both fixes came through the second build of the day intact — I checked the running binary rather
than trusting the deploy, and this time I had two proper controls: two strings my own commits
deleted are both absent from it, which proves the pod is running the *revision* I mean and not
just some version of the code.

The logo guard has now reached a second site. advertise.co.uk had a logo made this afternoon and
the instruction is in its prompt, same as boxingonline's. So the mechanism travels — it isn't a
one-site fluke.

I nearly got this badly wrong, and it's worth telling you because it's the same shape as the
mistake I logged yesterday. Five other logos showed a "last changed" date of today, none carrying
the instruction. Read at face value that says the guard missed five sites. It didn't — something
else is touching those rows (probably the sweep fixing a separate file-type bug), and that column
records only *that* a row was written, never what wrote it. The thing that actually settles it is
the job queue, which creates one row per dispatch and can't be nudged by an unrelated process.

So: two logos made since the fix, both governed. Three more are queued right now, which is quietly
the best news in this update — the open question I deliberately left undecided (whether to harden
the fix further) now has three scheduled chances to answer itself, rather than sitting as an
opinion nobody can settle.

One small thing still owed and it takes two minutes: nobody has actually **looked** at
advertise.co.uk's new logo. The database can tell me the instruction was sent; only a person can
tell whether the picture obeyed it. Both times we've caught this bug, it was someone opening an
image.

---

**2026-09-02, later — the logo fix is holding, and I found a problem in someone else's fix while checking it**

The two eye-checks this lane owed are done, and a third arrived free while I was working.

**All three logos are clean.** advertise.co.uk is a broadcast-signal mark — concentric arcs coming
off a mast. designblog.co.uk, which generated at five past six this evening while I was mid-check,
is a geometric star. I re-checked boxingonline's myself rather than take the delivery lane's word
for it: a fist in a square. **None of them has a single letter in it, and none is a two-panel
design comp.** That was the whole worry. Three for three.

So the guard is doing what it was built to do, on every logo the fleet has produced since it went
live. I would still not call the class closed on three, but there is nothing here suggesting a
fence is needed, and the fence decision can keep waiting for evidence rather than nerves.

**Getting the pictures was the hard part, and worth writing down.** advertise.co.uk's own domain
is not ours — it serves somebody else's Drupal site and returns "not found" for everything,
including things that do exist. Only two of our sites are actually published to the open internet
right now. So for two of the three logos I had to pull the file out of the storage bucket through
one of the running services, so that no password ever came near this session. That recipe is in
the runbook now, because the next person will hit exactly the same wall.

**The thing I did not expect.** Another lane shipped a fix today for a different logo problem —
the background behind a logo should be transparent, and the image model cannot do transparency, so
they added a step that cuts the background out afterwards. Their fix ran for the first time this
evening, on the designblog logo, **and it did not work.** The background is still there; it has
just been made half-see-through, like a veil, rather than removed.

What makes it worth reporting rather than shrugging at is *why it went unnoticed*: they wrote a
safety check that is supposed to refuse the image if the background was not properly removed, and
that check **passed with a perfect score.** It turns out it measures whether the tool *found* the
background, not whether it actually *removed* it. So it cannot detect the one failure it exists to
catch. That part is a real defect and it will not go away by adjusting settings.

I have written all of it up for their lane rather than touching their code — it is their fix and
they shipped it hours ago. **Two more logos are queued behind this one and will hit the same
thing**, so I have flagged that prominently; the queue triggers them automatically, which is worth
knowing because their own notes say "don't test this yet".

**One thing I nearly got wrong, and it is the more useful lesson.** I measured how far off the
background colour was, and their code has a comment literally asking the next person to measure
exactly that and write the number down. I drafted it as the answer. Then I checked the clock:
the run happened twenty-two minutes *before* they fixed a contradiction in the instructions sent
to the model, so the number I measured came from a run that was already known to be poisoned.
Correctly measured, correctly dated, and still the wrong answer to their question. I have marked
it as contaminated in what I sent them, and said plainly that their number is still unmeasured.

**Nothing here changes what is waiting on you.** The four decisions in the handoff — the identity
model, the derived-contacts question, the intake chat needing to ask about contact details before
ordering reopens, and bug 421 still having no owner — are all exactly where they were.

---

**2026-09-02, later still — you asked why websitepromotion has no logo. Here is the answer, and a correction to what I told you above**

**First, a correction.** Earlier this evening I wrote that "advertise.co.uk's own domain is not
ours" and that "only two of our sites are actually published to the open internet right now."
**The second part was wrong when I wrote it.** I checked a column called `publish_target`, found it
set on only two sites, and concluded it decided whether a site was live. It does not — it controls
copying a site to a *second* address. A site's own domain goes live when its DNS is pointed at our
server, which that column knows nothing about. Checked properly just now: five sites all have that
column empty and all five are live, including advertise.co.uk, which is now serving our site with
its logo in place. It was showing a stranger's page at five o'clock, so it was repointed somewhere
in between. I have corrected this where I wrote it, including in the runbook, and logged it.

It cost me some time going the long way round to fetch images, and nothing else.

**Now, the logo.** websitepromotion **does** have one. It was generated at one minute past six this
evening, it is a paper-plane-and-signal mark with no lettering, and it is sitting on the site right
now — `websitepromotion.co.uk/assets/images/logo.png` returns it. What is missing is the `<img>`
tag: the header still says the site's name in text.

The reason is a stale piece of the page rather than a missing file. The site's header is built once
and stored, and that header was built at **half past five**, half an hour *before* the logo existed
— its stored record of what it was built from lists the logo as empty. The page itself was rebuilt
at 18:01, after the logo arrived, but rebuilding the page just slots the already-built header back
in. So a fresh page keeps an old header, and the header is the only part that knows about logos.

**This is not just websitepromotion.** Of 34 sites that have a logo, **29 show it and 5 do not** —
websitepromotion, webdesign.co.uk, ai-agent-orchestration.com, loanandmortgagecalculator.co.uk and
cookly.uk. And they are not all the same fault: two had their header built before the logo arrived,
but **webdesign.co.uk's header was built afterwards, with the logo listed as an input, and still
came out as text.** That second one is a different problem and I have not diagnosed it.

I have not fixed any of them. Re-rendering headers on five live sites is a visible change to
customer-facing pages and it is not this lane's work — say the word and I will either do it or
route it to whoever owns the header pipeline.

**One caution if you do want it done quickly.** The logo that would appear on websitepromotion has
a faint magenta outline around its edges — a leftover from the background-removal step another lane
shipped today. It is much less bad than what the same step did to two other sites this evening
(there it left the whole background as a coloured veil), but it is visible. That step is unreliable
rather than broken: five attempts tonight produced one good result, three bad ones and one correct
refusal — **and its own safety check gave the good one and a bad one exactly the same perfect
score**, so nothing in the system can currently tell them apart. I have sent that lane the evidence.

---

**2026-09-02, 20:30 — the header is fixed; the last step is blocked and needs you**

**websitepromotion's header now has the logo in it.** The stored header went from 2,362 to 2,633
bytes and now carries `<img src="/assets/images/logo.png" class="logo-img">`, pointing at a file
that already serves. That part is done and verified.

**The live page has not caught up yet, and that is the step I could not finish.** Rebuilding the
chrome deliberately does not touch already-published pages — they keep serving until each one is
re-assembled. There are 11 pages. I prepared the 11 re-assembly jobs, and **the command that files
them into the queue was blocked by a permission check**, because it writes rows to the production
database. I have not tried to get around it. If you want it done, either approve that kind of write
or run it yourself and I will watch it through.

**Of the five sites, only one was ever fixable this way — that is the real finding.** I checked each
before touching anything, which was worth doing:
- **websitepromotion.co.uk** — the logo simply arrived after the header was built. Fixed.
- **webdesign.co.uk** — it has a *hand-built* header of its own, and that template has no place to
  put a logo at all: no image slot, nothing. I re-rendered it to be sure; it ran cleanly and
  changed nothing, exactly as the template says it should. Giving it a logo is a design change.
- **ai-agent-orchestration.com** and **cookly.uk** — neither site's current plan ever asked for a
  logo, and neither has a fallback, so there is nothing for the builder to find. They own a logo
  file nothing references.
- **loanandmortgagecalculator.co.uk** — its header, footer and head were **locked by a person** on
  5 August. Forced rebuilds are refused on locked pieces by design. Unlocking is your call.

So "the header shows text" turned out to be one symptom with four different causes, and only the
first was a rebuild problem.

**One thing that cost an hour and is worth knowing.** My first two attempts to trigger the rebuild
went nowhere — and left no trace at all. The message was accepted, the receipt confirmed, and no
job was ever created. The system refuses a malformed request *before* it writes anything down, so
there is nothing to find and nothing to explain it; it looks identical to a slow queue, and our own
written advice for that symptom is "be patient, don't re-send". The way I caught it was to check
whether *other* jobs were being created at the same time — 154 of them were, so mine had plainly
been rejected rather than queued. Third attempt, with the full set of routing fields, worked in ten
seconds. That is now written down as a trap, along with the check that tells the two apart.

---

**2026-09-03 — the logo is on websitepromotion, on every page. And it is too faint to see.**

You approved the last step and it worked. The 11 re-assembly jobs ran, all completed, and every
page on the site now carries the logo image instead of the site's name in text. I checked all 25
served pages, not just the front one — 25 out of 25, with the usual controls to prove the check
could have failed.

**But I want to be straight with you about what is actually on the page now.** The logo is a pale
blue-and-lavender mark on a transparent background, and the header behind it is white. I measured
the contrast rather than guess: **1.43 to 1**, where the accessibility floor for a graphic like
this is 3 to 1. It is genuinely there, and it is close to invisible. There is also a faint purple
halo around the edges, left over from the background-removal step.

So the mechanism you asked me to fix is fixed and proven, and the picture it is now delivering is
not good enough. Those are separate problems and I have not touched the second one — regenerating
the logo is a content decision, and the background-removal fix that would clean up the halo has
been written by another lane but is not yet live.

**What I would suggest, if you want it to look right:** wait for that fix to ship, then regenerate
this one logo and look at it. Doing it now would produce another one with the same halo, and
possibly another pale one, since nothing in the pipeline currently asks for a mark that reads
against a white header. If you would rather I raise that last point as a proper bug — "nothing
checks whether a logo is visible on the header it sits on" — say so and I will, once I have
checked who owns the imagery-quality side.

---

**2026-09-03 (later) — the test we were waiting for has happened, and the logos come back clean.**

You asked two things this morning and both are done or under way.

**First, the big one: the logo fix works.** Two sites have now generated a brand new logo with both
fixes switched on, and I looked at both pictures rather than trusting the green ticks.

`seotools.co.uk` came back with a magnifying glass over a woven lattice pattern. `gamedesign.uk`
came back with an abstract maze in terracotta and tan. **Neither has a single letter on it, and
neither invented a brand name** — which is the whole thing this bug was opened about.

What makes those two worth more than they look: in both cases the instruction that causes the
problem was *still in the prompt*, sitting alongside the correction we added. So the model had to
choose, and both times it chose correctly. That is the harder test, not the easier one.

**Second, a site failed, and it failed correctly.** `designblog.co.uk` tried three times and was
refused all three times — the safety check looked at what the model produced, decided the background
had not been removed properly, and threw it away rather than publishing it. So that site is still
showing its old poor logo. Nothing is broken; the guard did its job. But it means three tries were
not enough for that one, and the lane that owns the guard already has an open question about whether
the retry limit should be higher. This is now a real example rather than a hypothetical.

I was careful about one thing there. A fresh version of the system was deployed at 12:06, and a job
interrupted by a deployment looks *exactly* like a job that was refused — same failed status, same
missing result. I checked the timing and the refusal happened half an hour before the deployment,
with the safety check's own reading attached. So it genuinely was refused, not interrupted. Worth
saying because reporting it the other way would have put a false black mark against another team's
fix.

**Third, I have filed the "the logo is too faint to see" bug**, as you approved. It is number 462.

While writing it up I found the argument is stronger than I first thought. The reason the existing
checks skip this is written down in the code, and it is a *good* reason: when text sits on top of a
photograph, there is no single background colour, so any contrast measurement would be guesswork,
and acting on guesswork would create churn. That is right.

But our case is the opposite of that in every respect. The logo *is* the image, and it sits on the
header's own solid colour, which the site itself declares. Nothing is guesswork — both numbers are
known exactly. So the sensible rule that excludes the hard case does not actually cover ours; we
just fall down the gap between two reasonable decisions. That is a much better bug than "somebody
forgot to check".

I also checked the header really is white rather than assuming it, because the entire measurement
depends on that one number.

**Fourth, your logo regeneration for `websitepromotion.co.uk` is queued and running now.** I gave it
three fresh attempts rather than the one it had left, because the evidence says the margin matters:
of the three sites that just went through this, one needed two attempts, one needed three, and one
ran out. With a single attempt it would probably have failed for no good reason.

If it fails entirely, the current logo simply stays — the failure mode is "no change", not "worse".
I will report what actually comes out, including if it comes back pale again, which would be bug 462
happening in front of us.

**One thing I got wrong, and caught.** Looking at the new gamedesign logo, the white areas inside the
maze looked solid to me, and I worked out a plausible reason why the safety check would miss that —
it only inspects the outer edge. I was about to pass it to the other team as a finding. I measured
it first and it was simply false: those areas are see-through, and they look white because the page
behind them is white. One command, and it saved someone a wasted investigation. A picture cannot
tell you what is transparent when everything behind it is white.

---

**2026-09-03 (end of session) — the regeneration you approved made the logo worse, and I want to be
plain about that.**

`websitepromotion.co.uk` generated a new logo on its second attempt and it is now live on the site.
**It is worse than the one it replaced.** I would not have predicted this and I am not going to
dress it up.

What is on the page now is a chevron shape drawn in **white**, with a thin **magenta** outline. On a
white header, the white part is simply not there to the eye — I measured the contrast at **1.01 to
1**, where 1.0 means identical. The previous logo was 1.43 to 1, which was already too faint. So the
mark got fainter, not clearer.

And the only part you *can* see is the wrong part. Of the pixels that are actually visible, **63% are
magenta** — and magenta is the temporary background colour the system paints behind a logo purely so
it can be cut out afterwards. It is meant to disappear entirely.

**Why this happened, and why it is nobody's mistake exactly.** The background removal worked
*perfectly* — better than before, 93% of the image is now properly see-through. But the picture the
model drew was a white shape on that magenta background. When you cut magenta away from white, the
soft edge between them turns magenta, and if the shape itself is also white, that coloured edge is
the only thing with any colour left in it. So you get a magenta outline of an invisible shape.

**Every automatic check passed.** It was generated, cut out, inspected by the safety check, stored,
published and put on all 25 pages. The one thing nothing measures — whether a person can see it —
is the one thing that got worse. That is precisely the bug I filed this morning as 462, and it has
now happened in front of us on a brand new file, which settles any question of whether the first
case was bad luck.

**Two things you should know before deciding anything.**

First, **there is no undo.** Regenerating replaces the stored logo and the old one is deleted. The
only surviving copy of the previous logo is the one I happened to download before firing, and I still
have it. If you want it put back, that is possible, but it has to come from my copy and it needs your
say-so.

Second, **regenerating again is a coin toss, not a fix.** Nothing in the pipeline asks for a mark
that reads against a white header, so the next attempt could be pale again. Two of the four sites
that regenerated today came out genuinely good and legible; this one and one other did not. Until
462 is fixed, this is luck.

**I also have to correct something I told you and another team this morning.** I reported the
leftover magenta fringe as tiny — 0.01% on one logo, 0.05% on another — and said it was not worth
chasing. That was true of those two logos, and both of them happened to have **dark** strokes, where
a thin coloured edge is just cosmetic. On this new white one the same fringe is 63% of everything
visible. So the fringe matters far more than I said, and it depends on how light the logo is. I have
told the other team not to close that item on my numbers.

I also got a count wrong three times in one hour today and each time in a slightly different way; the
detail is in the technical notes, but the short version is that I kept measuring a rule by searching
for a word that the rule itself contains. What fixed it was looking in a different table, not writing
a cleverer search.

---

**2026-09-03 — your two decisions, written down.**

**On the invisible-logo problem (462): report it afterwards, not refuse it up front.** So the system
will measure whether a logo can be seen against its header, and raise it as a fault to be fixed —
rather than refusing to save a logo that fails. Recorded properly, including the thing it costs:
under this shape a hard-to-see logo *does* go live, and gets repaired once the check spots it. There
is a gap in between.

I had ranked the other option first, so I have written down clearly that you ruled against my
ordering on purpose, with the facts in front of you. That is so nobody later "discovers" the
ranking and re-opens it as an oversight. And for what it is worth, your call has a point in its
favour I had underweighted: refusing logos up front spends attempts, and we ran out of attempts on
designblog twice today, so refusing more things is not free.

**On websitepromotion: the old logo is not coming back.** The site keeps the new white-and-magenta
mark. Which means — and I want this said out loud rather than buried — **we are now knowingly
serving one logo that a visitor essentially cannot see.** That is a decision, not an oversight, and
I have marked it as settled so the next person doesn't treat it as an outstanding repair.

The useful side effect: it gives the new check a real test case. When someone builds it, it has to
flag websitepromotion, because we already know the answer there. A check that passes that site is
broken.

I kept both copies of the logo in the project files. They are no longer a way back — they are the
evidence that shows what happened, and the before-and-after pair is the clearest explanation of the
problem we have.

**One thing still genuinely undecided, and it is a real fork, not paperwork.** "Report it
afterwards" doesn't say *where the measuring happens*. Either we do it in the browser, which sees
exactly what a visitor sees, or we do it by reading the stored image and the site's own header
colour, which is much cheaper and could check every existing logo tonight rather than only pages
that happen to get audited. I lean towards the cheap one first, because right now nobody can answer
"how many of our sites have this problem?" and that version would answer it immediately. I have not
built either — that is the next person's job, and it needs the usual code review.

**Meanwhile designblog still hasn't managed to produce a logo at all** — two more refusals this
afternoon, with one final attempt due. It is the one site I most want to see a picture from, because
it is the only one whose instructions still contain the phrase that started bug 417.
