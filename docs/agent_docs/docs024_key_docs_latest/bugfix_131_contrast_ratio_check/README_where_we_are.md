# Where we are — the contrast check lane (plain prose, append-only, newest at the bottom)

## Friday 22 August 2026

You asked for a look at bug 131 — the list of problems you found using the vonc Gauntlet yourself
back in July. Here is where it stands, in plain terms.

Most of that bug was genuinely fixed at the time: the invisible headline, the content cut off on
phones, the button that did nothing, the cluttered page, the shareable verdict card — all done and
verified in July. But when I re-measured the live site today, two things had quietly come back.
The amber word "Gauntlet" in the headline is readable but again below the accessibility bar it was
fixed to meet, because the page's purple background has shifted shade since the fix. And the text
column on phones is now narrower than it was even when you first complained about it. Nobody
noticed, because nothing in the platform watches for this — every single time unreadable text has
been found on any of our sites (four sites now), it was a person looking at the page, usually you.

There is a machine that sweeps sites weekly and files tickets about poor contrast, but it cannot
stop a bad page going out, its ticket queue is parked awaiting your decision on a separate matter,
and the automatic repairer attached to it caused the idea.uk outage last week. What is missing is
much simpler: the acceptance test that every tool page already goes through — the one that catches
pages that don't load or content that gets cut off — cannot see colour at all. A check for "can a
person actually read this text" was proposed back in July, the day this bug was filed, and nobody
ever built it.

So that is what this lane is building: one new check, "contrast_ratio", in the same place and the
same shape as the overflow check that item B of this same bug produced (which has been quietly
catching real problems ever since). It measures the page the way a browser paints it, fails the
page if text is genuinely unreadable, and names the exact element so the fixing machinery knows
where to look. It deliberately does not fail on text sitting over photos (where the measurement
can't be trusted) — a wrong failure here would aim an automatic rewriter at a correct page, which
is worse than staying quiet.

What I am NOT doing from this lane: repainting the vonc pages (that is the gauntlet workstream's
surface, and there is a design pass queued for it — I've left them the measurements), and not
touching the parked ticket queue (that is your call, recorded elsewhere).

Two things from the old bug still need a human word from you, and I've written them into the bug
file: whether the phone column width (item D) goes to the queued design pass — it was never
actually decided, three documents disagree about it — and whether item H counts as done from the
engineering side, since what remains there is the distribution experiment you said you'd run
yourself.
