# robot-hands.com — where we are

Plain-prose running log, newest at the bottom. Append only.

---

**2026-07-19 (session: robot-hands.com site fixes R1–R6)**

You looked at the site on the 17th and found six things wrong. All six are now
fixed, and five of them I've checked on the live site rather than just in the
database. The dark look is back, the learning centre has one address instead of
three, the dead "Load More" button is gone, and the article list now shows only
the three guides that actually exist instead of nine, six of which were dead
links.

The interesting part is that the handoff I started from had the cause wrong, in
a way worth knowing about. It said a regeneration run on the 16th had wrecked
the header. Actually the blue header had been there since the 9th — I found
that by walking back through the deployed page history — and the run on the
16th only spread it to every page. Two other things in that handoff were also
wrong: it told me to restore the header from saved snapshots, and there are no
snapshots; and it blamed a colour-fixing agent that, it turns out, has never
worked at all. Every time it runs it fails, and the platform records it as
successful. I've written that up as a bug.

What was really wrong was that when the site was switched to the dark design
back on the 10th, only one of four places colour is stored actually changed.
Three stayed on the old light palette, and the header and footer were still
being built from old components that write colours directly into the page. So
every time anything regenerated, the blue came back.

Worse, a broken health check was telling the platform every few minutes that
this site had no design at all — which made the design agent invent a brand new
colour scheme, from scratch, over and over. On the 17th alone the stylesheet
was rewritten four times. One of those rewrites put a white background on a
site whose entire design is dark, and that went live before I caught it.

I've fixed that health check in the platform code, not just on this site. It
was looking for a marker that nothing in the entire system has ever written, so
it was firing on every properly-designed site, forever. That fix is now live —
another session's image roll picked it up last night — and I've confirmed it
worked: the false alarm hasn't fired once anywhere since. I also rebuilt this
site's header and footer so they carry no colours at all, only references to
the stylesheet, which means a future regeneration can't reintroduce this.

I put the platform change through the reviewer council, and I'm glad I did. It
took three rounds and finished with seven of the eight reviewers approving. It
caught a genuine mistake in my fix — I was checking whether a value existed
rather than whether it had anything in it, which would have created the exact
opposite problem — and it pushed back, correctly, on my trying to bundle a
bigger change in alongside a small one. I've split that bigger piece out as its
own job with the groundwork done.

**What still needs a decision from you.** Two tool pages — MatchMatrix and the
robot payload budget calculator — were planned but never actually built, so
five buttons on the homepage lead to "page not found". That's now the most
visible thing wrong with the site. The experience-loop work is already tackling
exactly this problem across all the sites, so my recommendation is that the
next session joins that rather than building these two by hand. If you'd rather
have it look right in the meantime, the quick fix is to point those five
buttons at the MatchMatrix page that does exist.

One caution I want to flag rather than bury. The project guidance changed
yesterday: claims about how the platform works are now meant to go through the
automated diagnosis loop before being asserted, because a session with full
context recently filed a confident claim that was refuted in ten minutes. Two
of my write-ups were done before that change. The evidence in them is cited and
checkable, but if either becomes the basis for someone else's work, it should
go through the loop first.

---

## 20 July 2026 — the MatchMatrix tool is built, and I found something worse on the way

**The short version.** MatchMatrix now exists and works. But going after the
broken buttons turned up two things that matter more than the buttons did: the
site's links were pointing almost at random, and three pages were publishing
numbers that were simply made up.

**On the buttons.** The plan was to repoint five homepage buttons away from a
dead page. When I actually looked, it was twenty buttons across eleven pages —
and only about six of them were *about* MatchMatrix. The dead MatchMatrix
address had quietly become a dumping ground: a button saying "Search the Gripper
Catalog" pointed at it, so did "Browse the Learning Center", so did "Open the
Payload Calculator", so did "Request Integration Support". Each of those had a
perfectly good page sitting there unused.

The second layer was worse and completely invisible. There are twenty more
"secondary" buttons, and essentially all of them were pointing at the wrong page
too — but none of them were *broken*, so nothing ever flagged them. Fourteen
buttons saying "Read the MatchMatrix Methodology" pointed at the services page,
while the actual methodology page sat there working fine the whole time.

The important bit: I fixed these by matching **what the button says** to where it
should go, not by find-and-replacing the dead address. The obvious quick fix — and
the one my own previous handoff suggested — would have sent "Search the Gripper
Catalog" to the MatchMatrix page and locked the mistake in permanently.

**On the made-up numbers.** This is the one I'd want you to see. The about page
was telling visitors the site indexes **"1,200+ gripper models"**. It indexes
**five**. Another page claimed 2,400+ models, 140+ manufacturers, and scoring
across 18 parameters. None of those numbers came from anywhere. There's also a
claim, repeated all over the site, that the catalogue spans six actuation
technologies — pneumatic, electric, vacuum and so on — and the database doesn't
record actuation type at all, so there's no honest version of that number at any
value.

The giveaway that nobody had ever looked at the finished page: one block was
rendering "2,400+%" and "140+ms" — leftover placeholder symbols stuck onto a
model count and a manufacturer count. It had been live like that.

I've corrected every one of these to figures that come from an actual query, and
written the query into the file so the next person can check them. I've also
filed it as a platform bug (043), because this isn't a robot-hands problem — the
content generator invented these, and nothing anywhere checks a number against
the data before publishing it. It's the same family as the fake veterinary
practices, but a different route in, so the fix for that one wouldn't have caught
this one.

**One thing I deliberately did not do.** That six-technologies claim also appears
in **forty-two** other places in the site's ordinary prose — body text, headings,
FAQ answers. Correcting a statistic is a bug fix; rewriting forty-two paragraphs
of copy is a decision about what the site says it is, and that's yours, not mine.
I've left them and recorded the count.

**On the tool itself.** I built MatchMatrix by hand rather than through the
platform's tool generator, and I want to be straight about why. The generator has
no rule against inventing data, and a gripper-matching tool is exactly the kind
that gets invented — it needs a catalogue to match against, so if you don't give
it one it makes one up. That's the bug you've currently got a hold on. Building
it by hand meant it could be honest about only having five grippers.

It does real work: you enter your part's weight, what it's made of, how hard the
robot accelerates and what safety margin you want, and it calculates the actual
clamping force needed, then tests all five grippers against it and shows you
which criterion each one passes or fails — and, importantly, where a manufacturer
simply doesn't publish a figure, it says so rather than guessing.

The most useful thing it does came out of a mistake I made. I'd assumed a gripper
advertised as "11 kg payload" would obviously handle an 8 kg part, and wrote a
test saying so. It doesn't — an 8 kg part on dry steel needs about 523 newtons of
grip and that gripper produces 140. The 11 kg figure quietly assumes very grippy
surfaces. My test was wrong and the tool was right, and if I hadn't written the
test I'd have "corrected" the working code to match my wrong assumption. The tool
now explains that trap on screen whenever it comes up, which is probably the
single most valuable thing on the page, because it's the mistake a real buyer
would make.

**Where it stands right now.** The tool is live and I've checked the real page,
not just the status — it loads, it has the working form, all five grippers are
there, and the dark theme survived. The link and number corrections are saved and
twelve pages are queued to be rebuilt with them. **Those corrections are not
visible on the site until that queue drains**, which takes a while because the
system rebuilds one page per site at a time. Anyone picking this up should check
the live pages rather than trusting the queue's own "done" marks — that caution is
in the handoff.

---

## 21 July 2026 — checked the site still holds after the new build, and cleared a fleet snag

A fresh build (v1.0.1144) went out this morning. I re-checked everything the
robot-hands work touched, against the live pages rather than the system's own
"done" marks: the MatchMatrix tool still loads and works, the made-up numbers on
the about and gripper-detail pages are still gone (5 grippers, not 1,200+), and the
dead tool link is still off the homepage. All good — nothing the earlier work fixed
came undone.

One thing worth knowing for whoever picks this up next: the build itself knocked out
the site-building queue across all the sites, the same way the last two builds did.
It's a known problem with its own write-up; robot-hands has no work waiting so it
doesn't affect us, but if you fire anything off at the cluster it may sit in a
backlog for a while. I cleared the jam once, which got one job through before it
snagged again — so it's a nudge, not a cure. The real fix is somebody else's open
task.

Also: the "listing shows a page that was never built" problem I filed (052) got
sharpened by another session — the right test for "never built" turned out to be
"was it ever deployed", not the flag I first used, because some pages that show as
needing a rebuild have in fact been live for weeks. Their version catches a worse
case on another site (a link in 28 page-footers pointing at a missing page). I've
pointed the handoff at their correction so we don't fix it the narrow way.

The handoff document is now current as of today's build, so a new chat can start
straight from it.

---

## 22 July 2026 — made the "six technologies" claim true by adding real grippers

You'll remember the one thing I flagged as your decision: the site says all over that it
covers grippers across six actuation technologies — pneumatic, electric, vacuum, magnetic,
soft-robotic and adhesive — and the catalogue behind it was five grippers that are all the
same kind (parallel-jaw, and the ones that say so are electric). I'd fixed the made-up
numbers but left that bigger claim alone because changing it is really a decision about
what the site is.

When I actually looked at the five grippers in the database, it was worse than "unsupported":
of the six technologies the site names, four of them (vacuum, magnetic, soft-robotic,
adhesive) had **zero** grippers in the catalogue at all. So the site wasn't just rounding up
— it was advertising whole categories it held nothing in. That's the same shape as the vet
site's made-up prices.

You chose to make the claim true rather than water it down, and I've done that. I added one
real, genuine product for each of the missing technologies — a Festo pneumatic gripper, an
OnRobot vacuum gripper, an OnRobot soft silicone gripper, an OnRobot "Gecko" adhesive
gripper, and a Schmalz magnetic gripper. These are all real products you can buy; I read
each one's specifications off the manufacturer's own page and saved the web address next to
it, so every number on the site now traces back to a real datasheet, not to something the
content generator invented. That was the whole discipline of this — the temptation with a
job like this is to type in plausible-looking specs, and plausible-looking is exactly the
disease. I didn't add a single figure I couldn't point at a source for.

The catalogue now genuinely holds ten grippers across all six technologies, and I've wired
the site's counters ("Gripper Models Indexed", "Manufacturers Covered") to count the
catalogue directly, so they can't drift away from the truth again the way the old made-up
numbers did. Those counters will tick from 5 to 10 (and manufacturers from 5 to 6) once the
site rebuilds — and the rebuild queue is still backed up behind that platform build problem
I mentioned, so it may be a little while. In the meantime the pages show 5, which is now an
undercount rather than a lie, so nothing dishonest is live.

Two honest limitations I want to name rather than bury. First, the MatchMatrix tool still
only compares the parallel-jaw grippers — it works out clamping force, which is a
jaw-gripper idea and doesn't apply to a vacuum or magnetic gripper — so any sentence that
implies the *tool* itself weighs up all six technologies is still ahead of what the tool
does. Fixing the catalogue fixed the catalogue claim, not that one. Second, the new grippers
don't yet get their own browsable pages; they back the claim and the counts, but the
"catalogue" page is still written as prose rather than a real list. Both are follow-ups, not
part of what you decided today, and I've written them down so they don't get lost.

---

## 24 July 2026 — the tool now really does all six technologies, and the fabrication engine got its muzzle

You asked me to carry on with the leftovers, and made three calls: make the tool
match the claims rather than water the claims down, clean up the other sites'
made-up numbers properly, and put a stop in the machine that invents them.

**The tool.** MatchMatrix now tests all ten grippers — including the vacuum,
magnetic, soft and adhesive ones — and it does it honestly, which took some care,
because you can't ask a suction cup how hard it squeezes. Each type is judged on
the numbers its own manufacturer publishes: the jaw grippers keep the force
calculation (including the trap explainer about optimistic payload ratings), the
magnet is checked against its published holding force and only on steel, and the
suction, gecko and soft grippers are checked against their payload ratings with
your acceleration and safety factor applied — so a "15 kg" vacuum cup correctly
fails an 8 kg part that's being flung around. Thirty logic tests pass, and one of
them failed the right way first: I'd written a test assuming a small gripper
could open 10 mm when its own datasheet says 6 — the tool was right and my test
was wrong, again. That's twice now, and it's why the tests exist.

**The write-ups.** Checking the claims against the new tool turned up something
worse than stale numbers: the methodology page described a scoring system that
has never existed — weighted six-dimension scoring, adjustable weightings, spec
"normalisation" with a worked example that misuses the name of a standard. I
rewrote those pages to describe what the tool actually does. The honest version
is genuinely stronger: every result now prints its own formula, so an engineer
can check it by hand — the invented version could never say that.

**The catalogue page** finally lists the grippers — all ten, pulled live from
the database each time the page rebuilds, so it can't drift.

**The other sites.** The made-up stats you'd have found embarrassing are gone:
vonc's "14,203 takes filed today" (there is no counter — and it claimed nine
archetypes when its own site documents eight), gamesdesign's "PRD accuracy gap"
(no tool on the site does any such calculation), and the agency site's "70
agents, 8 departments, 1000 concurrent" — which, funnily enough, UNDERSOLD the
real platform: it actually runs 170 agents across 13 live sites and has
completed over 1,200 automated jobs. Those real numbers are on the page now.

**The muzzle.** The important one. While I was working, the platform itself
demonstrated the disease: a routine page rebuild re-invented "2,400+ gripper
models" on robot-hands' homepage — a number I'd removed four days ago. The cause
turned out to be precise: the writing engine is handed a form that says "this
stat field is required, give me a number like 2.4M", its rulebook says "never
invent statistics", and given no data, the demand beats the rule. The fix gives
it a legal way out: if you weren't given a figure, leave the field empty — an
empty stat shows nothing, an invented one publishes a lie. And each of the four
cleaned sites now carries a short "verified facts" card — the only numbers the
writer may use, with instructions to recount them from the database before
changing them. One site found this morning still carries invented claims
(finetuning.uk's "clients served 11+, satisfaction 100%") — I've left it, because
inventing a "true" replacement without knowing that site's real story would be
the same sin; it needs a decision from you.
