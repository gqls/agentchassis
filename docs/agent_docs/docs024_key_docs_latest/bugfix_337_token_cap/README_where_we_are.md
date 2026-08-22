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
