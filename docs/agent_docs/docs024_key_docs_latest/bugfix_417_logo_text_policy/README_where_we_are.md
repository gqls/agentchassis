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
