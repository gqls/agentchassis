# Where we are — 462, can a visitor actually see the logo

Plain prose, append-only, newest at the bottom. The owner maintains this too — **append, never
rewrite or reorder**, and never edit his words; add a dated note below instead.

---

## 2026-09-04 — this became its own job, and the routing answer is not the one anyone expected

**The bug, in one line:** a site's logo can be generated, matted, checked, stored, deployed and
served — every gate green — and a visitor still cannot see it, because nothing in the estate ever
measures the mark against what is behind it.

**Where it had got to before today.** You ruled on 2026-09-03 that we should *report it afterwards*
rather than refuse a pale logo at the moment it is made, and that websitepromotion's invisible logo
stays where it is as the test case. The 417 lane then built the check and ran it over every site.
It found two logos below the accessibility floor — and, more importantly, that **only 7 of our 34
logos can be judged at all**. The other 22 have a background baked into the picture, so the header
is not their backdrop and the question does not apply to them yet. Those are the old ones, made
before the transparency fix; the judgeable number will grow as they get replaced.

**What changed today.** You put 462 on its own session. The 417 lane stopped, handed over cleanly,
and — this turned out to matter — told me about a measurement it had taken **after** its last
commit. I re-ran the check myself before trusting it, on the real artefacts rather than test
pictures, and it does what it says.

**The piece that was left: who fixes it.** The check reports; nothing is passed to anybody. So I
went to find the right place to send a finding. I expected a short job. It was not, and the answer
is worth your attention because it points away from building more machinery:

- **The obvious destination — hand it to the thing that makes logos — would make a new logo by
  drawing another one, with no rule about whether it can be seen.** There is no such rule anywhere in
  generation; adding one is the option you ruled against, deliberately. And we have already watched
  that go wrong once: when websitepromotion's logo was regenerated on 03 September, every measurement
  improved and the logo became *less* visible. A regeneration also overwrites the file everywhere,
  instantly, with nothing kept — there is no undo and no chance to look at it first.

- **The obvious "just tell a person" destination is a queue nobody works.** Our own code records 370
  items sitting there unactioned on 25 July. I counted it today: **1,439**, the oldest from March.
  Sending 462's findings there would be this same bug wearing a row number.

- **The one safe remedy — a person uploading a replacement — exists and works, and cannot be
  automated.** It needs someone with a file.

**And then the measurement that settles the sequencing.** If we built the automatic filer today, how
many logos would it legitimately act on? **None.** There are two findings. One is websitepromotion,
which you have ruled stays. The other is mortgagecalculator — and it turns out **that logo was
uploaded by a person, not generated**; there is no prompt behind it to make another one from, and
looking at it, a person *can* see it: it is a gold mark on cream that sits under the accessibility
bar rather than being invisible. Read as a score for our image pipeline, it is one failure in six,
and that one is the case you have already ruled on.

**So my recommendation is to make the check standing and stop there for now** — run it daily, have it
report every time including when it finds nothing, and build the filing machinery when there is
something it may actually act on. That is your "report it afterwards" ruling delivered in full; what
it leaves out is only the automatic dispatch, which today has nothing to dispatch and one dangerous
way to do it.

**What I need from you** is a yes/no on that, or a different call. The three options, and the trade
in each, are written up as §9e of the bug file. I am building the scheduled check meanwhile, because
it is needed under either answer.

**One thing to keep in view regardless.** This check reads each site's header colour from the theme
as *declared*, which is a snapshot. Our colours do get rewritten without anyone asking, so a logo
that passes today can quietly stay "passed" against a palette that no longer exists. The version
that stays correct measures the colour from the page as it actually renders. It is still unbuilt,
and it is still the destination.

**Later the same day — the scheduled check is built, and it is switched off.**

It is done and committed: a daily job at 08:15, which will look at every logo and write a short
report every time, *including on the days it finds nothing*. That last part is deliberate — if it
only spoke up when something was wrong, we could never tell the difference between "all fine" and
"the job quietly stopped weeks ago", and that confusion is this whole bug in miniature.

I proved the pieces rather than assuming them: I ran the check inside the exact container image it
will run in, and it passed all six of its own tests there, including on the two saved copies of the
websitepromotion logo.

**I have not switched it on.** Turning it on is one command, and it is the same thing as choosing
option (A), which is your call rather than mine. So it sits there, committed and doing nothing, and
I have said so in five places — a check that is built but not running, while looking to a reader as
though it is running, is a mistake this estate has made before and always notices too late.

**And then you said yes, so it is on.**

The check now runs by itself every morning at 08:15 and writes a short report each time. Its first
real run agreed exactly with what we had measured by hand: 34 logos, two below the line, five clearly
fine, and 22 we still cannot judge because they have a background baked in.

Switching it on immediately found two things wrong with how I had set it up — neither of them
anything to do with whether the check works, both about the difference between a tool a person runs
and a job that runs itself.

The first would have quietly wasted the whole exercise. The check reports a problem by exiting with
an error, which is right when you are sitting at a terminal. As a scheduled job it means the job is
marked *failed* every single morning — because there will always be something on the list, including
the websitepromotion logo you have deliberately decided to keep. A job that is red every day is one
everybody stops looking at within a fortnight, and then the next real problem arrives somewhere
nobody reads. It was also re-running the entire five-minute sweep on failure and filing its report
twice, which breaks the one thing the report is for.

The second is smaller but the same shape: it printed nothing at all for five minutes, so a job that
was working looked exactly like a job that had hung — inside the check we built precisely because a
broken thing and a fine thing look identical.

Both fixed, and the second run confirms it. I have written the general lesson into the shared
landmines file, because any of the other twenty-odd scheduled checks could be sitting on the same
trap.
