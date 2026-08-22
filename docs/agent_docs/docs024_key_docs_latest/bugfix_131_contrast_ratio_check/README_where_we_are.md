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

## Friday 22 August 2026, later — the check is built and the council approved it

The new check is written, tested and committed, and the review council approved it on the second
round. I want to be straight about the first round, because it caught something real.

Round one came back "revise". The reviewer's objection was that my check would report a page as
FINE if the measurement never actually ran — if the browser handed back nothing, my code read that
as "nothing wrong found" and passed. That is exactly the failure this whole piece of work exists to
end: we already have a contrast tool elsewhere that prints "0 problems" for pages it never looked
at, and I had quoted that very fault three times in my own submission while the same fault sat in
the code underneath. It is now fixed: the measurement stamps the page with a marker and counts what
it looked at, and anything without that marker — or a scan that measured nothing at all — is
reported as a failure to measure, never as a pass. I proved the new guards by deliberately breaking
them one at a time and checking the tests caught each one.

The second round passed with the same reviewer approving outright. One reviewer asked me to prove,
rather than assert, that a JS refactor I did was harmless; I did that by comparing it byte-for-byte
against the exact text that was already running in production, which is now locked by a fingerprint
so nobody can quietly regenerate it.

I also caught myself, twice in one day, writing "measured" or "enumerated" about checks I had not
actually run. Both are written up in the shared mistakes log, the second one deliberately as its own
entry, because a repeat within a single session says something the first one does not: the urge is
strongest in the sentence that answers someone who doubts you. Nothing false shipped either time.

**What happens next, and it needs someone else's build.** The check cannot do anything until the
browser-runner service is rebuilt and rolled — that service has its own image, so a normal chassis
release does not carry it. Everything is committed and ready for that build. Deliberately, no page
or tool is set up to use the new check yet: a check the running service does not recognise is
silently skipped, and a skipped check reads as a pass, so switching it on before the roll would be
worse than useless. Once it rolls, the plan is to prove it on vonc's own Gauntlet page, which today
has text at 1.66:1 that a person genuinely struggles to read, with a known-good page alongside as a
control.

One correction worth flagging: I had been repeating that the two deployments of this service run far
apart in versions. They are both on the same version today (v1.0.1323). The gap is possible, not a
current fact, and the handoff now says so.

## Friday 22 August 2026, evening — it went live, and the first real test showed it was half-blind

The build you deployed carried the check, and I proved that properly rather than trusting the
version number: the running service reports the exact commit it was built from, both of my commits
are ancestors of it, and searching the actual binary finds my new text with a known-present control
alongside and a nonsense string that correctly isn't found.

Then I pointed it at the vonc Gauntlet page — the page this whole bug came from — and it said the
page was fine. It isn't. I had photographed the unreadable text that morning.

The reason turned out to be worth the whole day. The check found all ten pieces of unreadable text,
measured them correctly, and then threw every one away. It inherited a rule from our older contrast
sweep that says "if there's any background image behind the text, we can't trust the measurement,
so don't report it". That rule treats two very different situations as one. If text sits on a
photograph, we genuinely cannot know what's behind it. But the Gauntlet section is a solid purple
with a faint decorative gradient over it — the background is not unknown, it just varies slightly.
Because our house style puts decorative gradients on exactly the big colourful sections where
unreadable text tends to happen, the check was blind in precisely the place it was needed. It could
not have caught the original bug.

I've fixed it so the two cases are treated differently: genuinely unknown backgrounds are still left
alone, while a solid colour under a see-through gradient is now judged — and judged generously, on
the most flattering reading of the background, so it only fails when no version of the background
could rescue it. On the Gauntlet page that turns zero reported problems into nine real ones, and
notably it no longer flags the headline accent, because under that gradient the headline plausibly
does clear the bar. That felt like the right kind of honest.

This is the third mistake of mine logged today, and the most instructive: I adopted a rule that
*hides* things without ever testing what it hides. A rule that suppresses output can never be caught
by a passing test, which is why it survived two review rounds and a deployment check. The test that
would have caught it is three lines of HTML and now lives in the lane.

The fix is with the review council now. It needs one more build of that service before it takes
effect — the version running today has the check but not the fix.
