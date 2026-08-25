# Where we are — the register that guards our words, and now our sums

Plain prose, append-only, newest at the bottom. Written for the owner.

## 2026-08-16

**The short version.** You asked me to take the next unworked bug in `bugs_open`.
That was 225 — the stamp duty calculator that charged first-time buyers using a rule
that stopped being law in March 2025. The good news is it was already fixed, and had
been since the 9th; I re-checked the live pages today to be sure rather than taking
the file's word. So the ticket itself was just tidying, and it has now moved to the
closed pile.

**The bad news is what the file said about itself.** Buried in it was a section
titled "Why no existing check could ever have caught this", and it was right. Every
check we own was blind to that mistake, and would be blind to the next one.

**Here is the shape of it, and it is worth a minute.** We keep, for each site, a
register of facts — every tax band, every rate, every threshold, each with a link to
the GOV.UK page it came from, and a robot re-checks those links every morning. It is
good machinery. But that register only ever governed what a page could **say**. It
never governed what a calculator **works out**. So the site could have the correct
figure sitting in its register, freshly verified that very morning, while the
calculator three feet away quietly used a figure that expired sixteen months ago. And
nothing anywhere would notice, because nothing was ever asked to compare the two.

There were three separate reasons nothing noticed, and the awkward part is that all
three are sensible decisions on their own. Our text checks ignore anything inside a
program — quite right, code is not prose. Our number checks skip calculator pages —
also right, a calculator's own help text is full of numbers that aren't claims. And
our number checks ignore money amounts entirely, because otherwise every price on
every site would trip them. Each is defensible. Together they leave a hole shaped
exactly like this bug, and nobody could see it because nothing was broken.

**What I've done about it.** I found that this had already been thought through —
there is a plan in the mortgagecalculator work, written on the 9th, that you have
seen, which sets out four pieces of a fix. The first piece is already live: when we
rebuild a calculator, the builder is now shown the register's facts. Pieces two and
three had been designed and then left, because that team moved on to logos and hero
images. So rather than invent something new alongside their plan, I built their
pieces two and three.

In plain terms: **a calculator can now declare which registered facts it relies on,
and when one of those facts changes, the morning check tells that calculator's owner
the same day.** If the Chancellor moves a stamp duty threshold, the register notices
overnight — as it always has — and now the calculators that encode that threshold get
named, instead of the change stopping at a note nobody reads.

**One decision I want to flag, because it is a judgement rather than a technicality.**
When a figure moves, I have made it *hard* for the system to fix a calculator by
itself. It will only hand the job to the automatic fixer when the calculator owns its
own code and its own settings say auto-fixing is allowed. In every other case — and
that includes every case where it's the *evidence* that changed rather than the
number, such as a GOV.UK page disappearing — it stops and asks a person. There are two
scars behind that. Our automatic fixer has twice rewritten a shared template and
changed a hundred-odd pages when it was asked to fix one. And a false alarm once
pointed the fixer at a page's legal disclaimer. Neither of those is something I want
happening to arithmetic on a page that quotes tax law.

The honest consequence: on today's sites, *every* route ends with a person, because
both stamp duty calculators are set to "don't auto-fix" and neither owns its own code.
The automatic path exists and is tested, but it has never run in anger. I'd rather
tell you that than let it read as more automatic than it is.

**Two more things I should say plainly.**

The first is a limit. This tells us when a figure has **moved**. It cannot tell us
whether a figure is **right**. If the register and the calculator are both wrong in
the same direction, they agree, and this says nothing. Answering that properly is the
fourth piece of that plan — a thing that works out the correct answer independently
from the published rules — and it needs its own review before anyone builds it. The
tool that actually caught this bug does exactly that, but it lives in a folder and
only runs when a person runs it.

The second is that none of this is switched on yet. The code goes live with the next
release, and after that somebody has to tell the stamp duty calculator which facts it
relies on — one line in its settings. I've written that up and handed it to the team
that owns the site, since it's theirs, not mine. Until that line exists, this machinery
is real but idle, and I've said so in the tracking file rather than letting a green
tick imply otherwise.

**Also worth knowing:** I got something wrong today and logged it. I quoted two fleet
statistics in a commit message from a research assistant's summary rather than
checking them myself; they were about ten percent out. The conclusion they supported
was still true, which is precisely what makes that kind of error easy to leave in
place. It's written up in our wrong-calls log.

---

## 2026-08-24 — what we found and what we built

The problem this lane exists for, in one line: the site's list of checked facts governs
what a page can **say**, and it has never governed what a **calculator works out**. The
case that started it was a stamp duty tool that used a tax threshold which expired
sixteen months earlier, while the correct figure sat in the register a few feet away,
re-checked that very morning, and every check we own passed the page.

Back in August we built the first half of the answer: a calculator can now declare which
registered facts it uses, and the nightly sweep tells it when one of them moves. That
part works and is proven. I picked the lane up after it had been quiet for a week.

**The first thing I did was re-count, and the counts had moved a lot.** When the bug was
filed there were 143 facts across 12 sites. Today there are 294 across 15 — the register
has more than doubled in eight days. In that same period the number of calculators that
actually declare anything went from nought to **one**, and that one only because we asked
another team to add it by hand. There are 178 calculator pages sitting on sites that have
a register. So the machinery works and almost nothing is plugged into it, and the gap is
getting wider rather than narrower.

**The second thing was that the nightly sweep had never once looked at a calculator.** It
asks "has this figure changed, and who says they use it". It does not ask "does this
calculator actually contain that figure". On the one tool that had adopted it, the sweep
filed thirteen requests on 17 August asking a person to confirm the figures by hand.
Seven days later all thirteen were still sitting there, untouched. So the honest position
was: we had built something that asks a question nobody answers.

**The near-miss worth telling you about.** I set out to make the machine answer that
question itself, by looking for the registered number inside the calculator's code. My
first check said all thirteen figures were present — which sounds like the idea works.
It doesn't prove anything at all. Four of those thirteen "figures" are 5, 2, 10 and 12,
and any page of HTML contains those digits somewhere. I only found out because I then
went looking for numbers that ought to be **missing**: the expired threshold from the
original bug, and two numbers I invented. All three were absent, which is what made the
present ones mean something — and two more digits I picked at random turned out to be
present too, despite not being registered facts at all.

**And then a trap that would have wasted the whole thing.** The obvious way to look for
the figure is to search the page. But our own system writes the registered figure into
the page's **wording** — that is what the register is for. So on the original bug's page,
the text says "£500,000" (correct, because the register put it there) while the code
underneath still says the old number. A check that searches the page finds the right
figure in the prose and declares the calculator healthy. **It would have passed the exact
bug it was built to catch, every day, for sixteen months.** The fix is to read only the
code, and to ignore the prose entirely.

**What is now built, in four pieces:**

1. A declaration that can no longer fail silently. We had a rule that was supposed to
   reject a badly-written declaration — it turned out that rule had never once run on
   these documents, in the place they are actually written. And a formatting slip in the
   declaration made the whole thing quietly do nothing, *including* the warning whose job
   was to say it was doing nothing. Both fixed.
2. The one existing tool that reads a calculator's raw code can now be pointed at the
   facts that matter. It could not be before, and its address was a reference that dies
   whenever a page gets rebuilt.
3. A check that looks in the calculator's code for the registered figure. **It only
   reports for now — it changes no decisions.** It runs for a month, we see how often it
   is right, and only then do we let it act. That is deliberate: an earlier plan rejected
   this kind of check precisely because nobody had measured how often it cries wolf.
4. The adoption piece. Rather than asking people to hand-write 178 declarations, the
   sweep now proposes them: it finds calculators whose code already contains registered
   figures and files a ready-to-paste suggestion. Fifteen of these exist today across
   three calculators — one of which is our **second** stamp duty calculator, which is
   currently protected by nothing at all.

**What this still does not do, and I would rather say it plainly:** none of it can tell a
correct figure from a confidently wrong one. Everything here assumes the register is
right. If the register and the calculator are wrong in the same direction they agree, and
every check stays silent. That remains a separate piece of work, and it needs a proper
design review before it is built.

**One thing I got wrong three times in a day**, because it is the sort of thing worth
recording: I wrote tests for each fix, they all passed, and then when I deliberately
broke the code to check the tests would notice — they didn't. Three times. Each time the
test was checking the piece of machinery in isolation and could not see whether anything
actually called it. The third time was the worst, because it was the test guarding the
most important rule in the whole change. It is fixed, and the pattern is written down.

**Nothing is live yet.** All of this is code, and code on this system does nothing until
the next fleet build goes out. The first real test is the day after that.

---

## 2026-08-25, later — it ran for real, it worked, and I need a build to go out

The thing above did go live, and this morning at 09:06 it ran across the whole estate
for the first time properly. Nineteen sites, no errors. It found five calculators whose
code already contains figures from their own register and wrote a ready-to-paste
suggestion for each. Seven of those are on our second stamp duty calculator, which until
today was watched by nothing at all.

So the mechanism works. The line at the end of the last entry — "nothing is live yet" —
is out of date, and I have left it there rather than tidy it away.

**It also found a mistake of mine, which is the part I would rather report than bury.**
On one site it proposed *two* different register entries for the same single number in
the calculator's code. Nobody can act on that: you cannot declare two facts for one
constant. I had written a comment claiming the code refused exactly this case, and the
refusal had never been written. That is fixed. A second, separate problem turned up the
same day: if someone writes one of these checks in a slightly different — and arguably
more sensible — place in the file, nothing reads it and nothing complains. The sweep now
says so out loud. Also fixed.

**Both fixes are committed and neither is in the running system.** I checked rather than
assumed: I asked the live service directly whether it contains the new code, with a
control string that must be there and a nonsense string that must not, and the answer is
no on both counts. The machines happened to restart eight minutes *before* I wrote the
first fix, which is why.

**So the ask is a build.** I have bumped the tag to `v1.0.1338`, because re-releasing the
tag that is already out there just serves the cached copy of the old code — that trap is
written into the makefile itself. When you have a moment:

```
date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date
```

I will check afterwards, at the service itself rather than at the tag, and report back.

Until that build goes out, one thing is worth knowing: the "someone put the check in the
wrong place" warning reads as empty everywhere. That is because the code isn't running,
not because everyone has got it right — so it is not evidence of anything yet.

**Meanwhile I am not waiting.** The seven bindings on the second stamp duty calculator do
not need the build, so I am doing that adoption now. That matters beyond one calculator:
the measuring half of this work only gathers data when somebody newly declares, and the
one site that has declared so far uses figures too small for the check to say anything
about. This one is the first that will produce a real answer.

While setting that up I found that the note I sent that lane last night was wrong in a
way worth owning. I warned them about a problem in their fence installer — except I had
read the *neighbouring* lane's installer, not theirs. Theirs does not have that problem.
It has a worse one I had not spotted: it rebuilds the whole file from scratch every time
it runs, so anything added by hand afterwards is silently deleted the next time anyone
uses it. I am fixing their installer properly rather than pasting something in that would
quietly vanish, and I have corrected the note so they read the true version first.
