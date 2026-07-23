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

## 2026-07-22 (evening) — I was wrong; the tool still doesn't really work

You tried it and you're right. What I shipped was cosmetic. I took the dead buttons and
wired them up, but I wired them to things you can't see happen: "Enter the Gauntlet"
quietly starts the clock and scrolls to a panel that's already in front of you; "Preview
Rules" scrolls to a card that's already there. And ticking the objectives just moves a
progress bar that isn't connected to anything. So from where you sit, nothing happens —
which is the exact hollow-placeholder problem I'd just asked the council to stamp out.
I fixed the broken link and missed the actual point: the tool doesn't DO anything.

You've made the right call, and it's an honest one: build a real backend with an AI
opponent that actually plays against you — file your take, the AI challenges it, you
defend, there's a real outcome — and let it say plainly that it's an AI competitor until
real people show up. That's a genuine feature, not a fake crowd.

Good news: there's already a whole workstream built for exactly this — the "experience
loop" — and it was started because of THIS gauntlet. It can plan the whole end-to-end
experience (the journey, what's promised, what data it needs) and it's already produced
an approved plan once. The missing piece across the platform is turning such a plan into
a built, working feature — and doing that here also needs a real backend, which these
static sites don't have yet. So I'm researching three things before I commit to a
route: how the experience loop's build step works, how the "feature builder" implements
things, and how a static page like this can safely call a live AI backend. Then I'll lay
out a concrete plan for you rather than guess.

## 2026-07-23 — the real build is underway

Today the gauntlet stopped being a cosmetics job and became a real feature build, on
your decisions: a debate opponent (you file your take, the AI genuinely argues back,
you defend on the clock, it judges honestly), built by the platform's own
feature-builder, exposed through a shared API address on apis.uk with a bastion
machine guarding the cluster.

Progress so far today, in order:

1. **Unstuck the experience loop.** Its contracts reviewer had been blocking every
   plan that proposed new code (it demanded quotes from source that doesn't exist
   yet). Applied the rule split you approved — strict as ever about existing code,
   sane about genuinely new code. Config change, snapshotted, reversible.

2. **Gave the planner your new ruling.** The planner's standing instructions still
   said "client-side only, no server" — that's how decisions travel to it, so I
   updated that block with the debate design and the exact API endpoints, and fired
   a fresh planning run. It's working now (these queue behind the fleet; expected).

3. **The backend is being built by the machine, right now.** I wrote the formal
   capability request for `tools-api` (the debate engine service), the
   feature-designer turned it into a 6-stage build plan, and its review council
   approved it **unanimously on the first round**. Then I fired the
   feature-implementer — the first time it has ever run. It's writing the Go code
   stage by stage on its own branch, with a compile gate after every stage, and will
   open a pull request for you to review. **Nothing lands without your merge.**

4. **The security shell is drafted.** The tunnel config, the bastion proxy rules
   (only the tools API path gets through, with size and rate limits), and the
   cluster-side network policy are written and waiting on three things only you can
   do: name the subdomain on apis.uk, provision the bastion machine, and approve the
   WireGuard peering. No rush — the code has to land first anyway.

What I got wrong earlier still stands corrected in the record: last night's "fix"
looked right and did nothing a visitor could feel. Today's work is the opposite shape:
the machine is building something real, and the last gate is you reading a pull request.
