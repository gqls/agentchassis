# Where we are — the vonc gauntlet

Plain-prose running log. Newest at the bottom.

## 2026-07-22 (start)

You reported that `vonc.com/tools/gauntlet/index.html#` doesn't work and weren't sure
we had a working gauntlet. Here's what I found.

The page itself loads fine, and most of it actually works — the countdown timer, the
tick-off objectives, the progress bar and the counting-up numbers are all real
JavaScript and behave. The two things that DON'T work are the two big buttons at the
top: "Enter the Gauntlet" and "Preview Rules". Both are wired to `href="#"`, which
means "go nowhere" — clicking just adds a `#` to the address bar. That `#` on the end
of your URL is exactly that. They're dead by design: the button *text* is configurable
but there is no field anywhere for the button's *destination*, so nothing could ever
fill it in.

Two more honesty problems on that page: the headline stats ("12,847 competitors",
"94,210 challenges completed", "38% win rate") and the whole "Top Competitors"
leaderboard (AxonFury, ZeroRush…) are invented placeholders. There is no real gauntlet
behind the button — no actual competitors, no real challenge to enter.

The good news: the platform already built a detector for exactly this ("dead controls")
and its source code literally names *our vonc gauntlet* as the example it was built to
catch. The bad news: it never actually caught it. It only looks at pages marked
"deployed", and our gauntlet page is marked "needs rebuild" even though it's serving
live — so the detector skipped its own poster child, and about 34 other live pages with
the same quirk.

You made three calls: make the gauntlet genuinely work (don't fake it); fix the
detector so any new site is covered; and send the council a message that we shouldn't
be creating placeholders that don't work. You also asked whether we should give it a
real backend like idea.uk or relojistas. My judgement: not a full one yet — a real
"gauntlet" needs real competitors and a live leaderboard, and with no real users a live
leaderboard is just another fake. So I'm making the front end genuinely work now,
reusing our existing form-delivery plumbing for the one real action (letting someone
actually submit their take to you), and leaving the full competitive backend as
something to switch on once there's real traffic.

Done so far: I fixed the detector (a one-line change — judge "is it live" by the
component that's actually serving, not the page's stale flag) and sent it to the
council with your placeholder message as the reasoning it reviews against. Waiting on
the council's verdict (~30 min). Next I'm rebuilding the gauntlet page itself so the
buttons do real things and the fake numbers come out.

## 2026-07-22 (done for today)

Both jobs are done. The gauntlet page now works honestly: the two dead buttons are
gone — "Enter the Gauntlet" is a real button that starts the clock and drops you into
the challenge, and "Preview Rules" now scrolls to a short "How the Gauntlet works"
card. I stripped the invented numbers and the fake leaderboard entirely; the page no
longer pretends there's a crowd of competitors. It's honestly a self-paced solo
challenge now, and it says so ("no sign-up, nothing is scored or shared"). All of this
is live on vonc.com right now — verified.

The generic detector fix went to the council and was APPROVED, and I've committed it.
It won't actually take effect until the next time the backend image is rebuilt (I
didn't force a fleet-wide rebuild just for a detection tweak) — but from then on, any
new site that ships a dead button on a page in this "serving-but-flagged-for-rebuild"
state will get caught, where before it slipped through.

And I sent the council your message — that we shouldn't be creating placeholders that
don't work — as the reasoning it reviewed this fix against, so it's on record with the
seats that judge platform changes.

On the backend question: I've deliberately kept this to an honest front-end for now, for
the reason above (a live leaderboard with no real users would just be a new fake). If
you'd like, the natural next step is to let a visitor actually submit their "Position"
to you using the same contact-form delivery we already fixed — a small, real feature —
and only build the full competitive backend once there's real traffic to populate it.
