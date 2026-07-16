# Summary to read out — the empty product pages, and the fix loop that lied

*2026-07-16. Plain language, written to be read aloud. No commands, no
file paths. Roughly five minutes at a normal pace.*

---

## Where this started

A page on one of our sites, robot-hands.com, was serving product pages that
were completely empty. Not broken — empty. The layout rendered, the "Add to
Cart" and "Buy Now" buttons rendered, the star ratings rendered. But every
actual value — the product name, the price, the features, the part number —
was blank. The pages were live and publicly reachable.

That was bad. But it wasn't the real problem.

The real problem was this: **our platform had already found this fault, tried
to fix it, and told us it was fixed.** There were work items describing these
exact empty sections. A handler had picked them up. And it had marked them
"complete" — on the tenth of July. Four days later, the sections were still
empty.

So we had a self-healing system that was reporting success while healing
nothing. That is worse than a system that fails loudly, because it means you
can't trust any of its good news.

## What we found

Three faults stacked on top of each other.

First, the dispatch loop marked a work item "complete" whenever the handler
finished without throwing an error. Not when the problem was actually gone —
just when nothing crashed.

Second, the handler had exit paths that looked like success but did nothing at
all. If it decided it had no work it could do, it exited through a door
labelled "completed successfully."

Third, on this particular page, that no-op exit was guaranteed every single
time. So every attempt did nothing, reported success, and moved on.

And there was a nasty sting in the tail. The system re-detected the same fault
repeatedly, and a safety rule that exists to stop infinite loops eventually
gave up and parked those repeat detections in a status that meant nobody would
ever look at them again. So the false "fixed" reports quietly generated a
backlog of invisible, permanently-ignored problems.

## What we did

**We made the loop honest.** Before anything can be marked "complete" now, the
system re-runs the exact same check that found the problem in the first place.
If the fault is still there, it refuses to call it fixed and sends it back
round. We proved this by taking the *original* work item — the one that lied to
us on the tenth of July — and running it again. It now stops and says, plainly,
"I could not do this." It cannot lie any more.

**We taught it to spot this class of fault.** A new check looks for components
where the required content is simply missing. When we ran it, it found eight
real problems on robot-hands. Just as importantly, we ran it against a
different site as a control, and it correctly found nothing — so it's precise,
not noisy.

**We added a guard against a strange failure we found.** One section had
stored the AI's own apology — text explaining why it couldn't produce the
content — and published that apology as if it were page content. There's now a
guard that blocks that.

**We fixed the actual pages.** robot-hands is a specification and comparison
site. It doesn't sell anything. So "Add to Cart" was wrong regardless of
whether we filled in the data. We replaced that with a proper specification
sheet, and we filled it with real, genuine data — five actual industrial
grippers from Schunk, OnRobot, Robotiq, Zimmer Group and Festo. Every number
was read off the manufacturer's own page, and every card on the site now shows
its source and the date it was verified. Nothing was invented. Both pages are
live and correct right now.

**We cleaned up the backlog.** Rather than blindly re-running thirty-odd stale
items, we checked each one against reality. Every single one was already dead —
either the component no longer existed, or later work had already fixed it. Not
one was a live problem. We closed them honestly. The backlog went from
thirty-six items, nineteen of them invisible zombies, down to six genuine
faults, each correctly attributed to whoever owns it.

That collapse is itself the proof. The old backlog wasn't work. It was
wreckage left behind by the lying loop.

## The mistake I made, and what it cost

I need to flag this, because it matters more than any of the fixes.

Along the way, some builds hung. I investigated, found a plausible explanation
in the code, and wrote it up as the root cause — in four separate places,
including the words "investigated, not guessed." I said it only affected manual
operations and never touched production.

**I was wrong on both counts.** A separate investigation, with proper evidence,
found the truth: it's a network fault. Certain machines in our cluster can't
reach one of our message brokers. Any job that lands on a bad machine hangs.
It affects production constantly — it was never a manual-only issue.

Here's the uncomfortable part. I had the disproving evidence in my hand. I had
already read the logs from the failing component and quoted them. I just didn't
scroll far enough to see the connection errors. I had four data points, built a
theory that explained them, and mistook "explains it" for "is the cause."

I've retracted it everywhere and pointed to the correct diagnosis. But the
lesson is sharper than that: a confident wrong answer is worse than no answer,
because it sends the next person down a dead end — in this case, into the most
dangerous code we have, hunting a bug that was never there.

And note the irony. This entire project exists because a system reported
success it hadn't earned. Then I did exactly the same thing with a diagnosis.
So the rule now written into our notes is: when something doesn't respond,
read the logs to exhaustion before theorising, prefer the boring
infrastructure explanation, and say "hypothesis" when that's what it is.

## Where we are

The core work is done and proven on the live site. The loop is honest, the
pages are real, the backlog reflects reality.

One thing we built doesn't work yet. We built a tool to automatically
re-verify those product specifications against the manufacturers' pages, so
they don't silently rot. It runs end-to-end, it fetches the pages successfully
— and then extracts nothing. The good news is it fails safely: it leaves the
existing good data untouched rather than wiping it.

I have a strong hunch why — we're likely only showing the AI the first chunk of
each page, and the specification tables probably sit below that cut. But,
having just been burned by exactly this kind of reasoning, I have deliberately
written that down as a hypothesis, not a conclusion. I also found and fixed a
blind spot in my own code that was hiding the answer: it was failing silently.
It now reports what the AI actually said. The next run should tell us the
answer rather than us having to guess it.

## Where we're going

Next, in order.

One: rebuild, re-run that spec refresher, and read what it now tells us. That's
a short job that ends either in a fix or a precise diagnosis.

Two: the network fault is the big one. It's slowing down work across the whole
fleet, it's properly diagnosed, and it belongs with whoever owns the
infrastructure.

Three: a handful of smaller, well-documented items — a failing test, one page
where two systems disagree about its layout, and a decision about another site
whose product grid will disappear the next time it rebuilds.

Every one of those is written up so someone can pick it up cold.

## The thing worth remembering

We didn't just fix some empty pages. We fixed a system that was telling us
things were fine when they weren't — and we found out the hard way, through my
own mistake, that the same failure mode applies to people and diagnoses, not
just to code. The fix in both cases is the same: check that the thing you
claimed actually happened.
