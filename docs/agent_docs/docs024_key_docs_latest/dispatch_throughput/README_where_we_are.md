# Where we are — dispatch throughput and the road to thousands of domains

Owner's plain-prose log. Append-only, newest at the bottom.

---

## 2026-08-19 — the workstream is claimed, and the question got bigger

You asked for a deep look at every option for increasing the system's throughput — including
whether we need more than one repo, more than one deploy path, or another cluster — because
the plan is to host and maintain several thousand domains, and when you promote the service
you may get many sign-ups per day in bursts.

What we found, in plain terms:

**The machines are not the problem.** The cluster is running at a fraction of its capacity —
the computers are mostly idle. What limits us is that the work is deliberately taken one
piece at a time in several places: the dispatcher serves one site at a time (about 83 items
an hour for the whole fleet, measured), the agents that talk to outside services mostly
handle one request at a time, and every finished page goes out through a two-slot deploy
queue. These were all sensible safety choices when they were made; they are now the ceiling.

**Money is the other ceiling.** Every piece of work costs AI calls. The account has hit its
monthly spending cap twice this month already, at a fraction of the volume you're aiming
for. Scaled naively, thousands of domains would cost more per domain per month than the
£10 domain fee brings in. So "how fast" and "how cheap per item" have to be solved together —
fewer, smarter items per domain, cheaper models for routine work, and Anthropic's batch
pricing for anything that can wait a few hours.

**Bursts change the priorities.** If a promotion brings fifty sign-ups in a day, each one is
a full site build, and today's queue serves everyone strictly in the order work arrived —
so sign-up number fifty would wait days behind routine maintenance. The fix is not "make
everything faster", it is: a priority lane for new customers, a spending governor so a burst
can't blow the monthly cap mid-build, and an honest "your site is building, ready in about
X" in the sign-up flow. We also found the Cloudflare account has an unpublished limit of
roughly a thousand zones — at fifty domains a day that arrives in weeks, so the already-
agreed "plan B" for DNS may need to be built before the first big promotion, not after.

**Your three questions, answered short:** more repos — no (you already ruled one repo per
site out, and the evidence agrees); another cluster — not for capacity, the current one is
idle; the future shape is self-contained "satellites" when a single cluster genuinely fills
up; the deploy path — yes, that one really is a bottleneck that grows with domain count, and
there is a concrete choice to make about it.

What happens next: a diagnosis run is verifying the dispatcher finding independently (our
own rules require that before we write it down as fact), and the full research document with
every option costed and a decision list for you is being written now. Nothing in the live
system has been changed.

## 2026-08-19, later — the research document is written, and we measured a real build

The full review is now in this directory (RESEARCH_2026-08-18_throughput_to_thousands_of_
domains.md). The headline addition since the morning: we measured what one new domain
actually costs, using loanzy.uk, which was built yesterday. One site: about two hundred
pieces of work, four hundred and ten agent runs (a fifth of which failed internally and
were retried — worth knowing before a burst), about twenty dollars of AI calls, and ten and
a half hours from submission to mostly-live — most of that queueing behind other sites, not
working. We also found the AI bill is dominated not by site work but by the platform's own
review council, which grows with how much we change the platform, not how many domains we
host — good news for per-domain economics, and it means the per-domain running cost still
needs one clean measurement before pricing anything. The document ends with a decision
list; the two that block everything else are: how much upkeep per domain per day do you
actually want, and how big a sign-up burst should we design for, with what promise on
"your site will be ready in…".

## 2026-08-20 — the decision list, explained in plain terms (owner asked)

The two that size everything: D0a is how much upkeep per domain per day you actually want
once a site is built — the whole requirement swings 100× on it (one item/domain/day for
3,000 domains is nearly reachable with config changes; "hundreds of thousands of jobs a
day" needs the structural tier). D0b is the burst: the peak signups/day to design for and
the promise in the signup flow ("ready in about X hours") — the promise sets the required
drain rate, and one measured build (213 items, 410 runs, ~$20 AI, 10.5h) says fifty
signups in a day exceeds what the whole fleet currently does.

The queue (D1–D3): D1 is how many dispatch turns run in parallel — each also multiplies
spend, default stop at 2, and 2 is the safe limit until the adapter work is done. D2 is
who gets served first — today strictly oldest-item-first fleet-wide, which under a burst
puts a paying customer's build behind days of old maintenance; the priority-lane constant
is yours because any priority scheme risks starving someone. D3 is whether batch size and
the scheduler's timeout always move in lockstep (default: yes).

The AI account (D4–D7): D4 is a spending governor — nothing in the code limits AI spend,
the monthly cap was hit twice in eleven days, and without a governor a successful
promotion likely ends in a mid-burst AI outage; it's a promotion prerequisite. D5 is
whether maintenance may fail over to a second provider (the stay-on-Sonnet ruling was made
for the council, where caching makes mixing costlier). D6 is which work classes may wait a
day for half-price batch processing. D7 is the Anthropic account tier itself.

The two forks to pick before code is written: D8 — pages reach production either by the
platform writing storage directly (git becomes the audit record) or by keeping the
Actions-per-commit path with batching and more runners; mutually exclusive investments.
D9 — either fix the scheduler properly or retire polling for workers pulling from the
database queue (a pattern the chassis already proved); same exclusivity; the sibling-row
stopgap is reversible under both.

How we work: D10 — releases are owner-only, serial, no CI; decide on CI/delegation.
D11 — worktrees for code sessions, deferred in July at a quarter of today's commit rate.

The estate: D12 — DNS plan B timing is calendar-shaped: at fifty domains/day the ~1k
Cloudflare zone cap arrives about three weeks into a promotion, so plan B may need
building BEFORE it. D13 — satellites: domains per satellite, the second-satellite
trigger, and whether to build the five cheap seams now.

Hygiene: D14 spot-node floor/autoscaling; D15 a backlog ceiling and whether maintenance
pauses during bursts; D16 retention for the two database tables growing toward the 100 GB
disk. Defaults exist for everything but amount to "stay as we are" — fine for details,
risky for D0b/D4/D12 if a promotion is coming. Rough answers to D0a and D0b turn the rest
into a costed, ordered build queue.

## 2026-08-21 — your rulings are recorded, and four questions answered

Your answers to the decision list are now written into the research document and notes
(clients first; timeout moves with concurrency; Batch pricing yes; one AI provider for
now; keep GitHub Actions and add runners; spot machines fine for now; maintenance pauses
during bursts; start DNS plan B now; retention review owed; three portfolios — client-
retained, own-high-attention, own-low-attention; and a human check on every site before
it goes out, with the fixing mechanism designed in your other thread).

One of your beliefs turned out to be happily wrong: Anthropic does have higher tiers.
Start is capped at $500/month, Build at $1,000, Scale at $200,000, and Custom is
uncapped — organisations move up automatically with usage history, and there is a
"Request rate limit increase" button in the console that covers the monthly spend cap
too. Better still: the error we hit on the 17th says "you have reached your SPECIFIED
limits", which is the signature of the limit you set yourself on the Billing page — not
Anthropic's cap. So at least one recent outage is fixable today by raising your own
limit in the console, and a burst month at fifty sign-ups a day (~$30k) needs the Scale
tier requested ahead of time.

On removing the polling dispatcher (your D9 question): the short answer is that it's the
right end-state and the problems are manageable — the dispatcher's hidden second job is
pacing (it is the accidental spending brake, so the governor must land first); the
one-at-a-time-per-site rule and the clients-first ordering move into the database claim
query (which is actually where clients-first is easiest to express); the fleet-freezing
"wedged loop" failure disappears and is replaced by lease-and-reclaim, which we already
have machinery for; and the scheduler binary stays for all its other timers. It also
fits the satellite future better, as long as each satellite's workers pull only from
their own database. Do it the way the chassis worker pool was done: behind a switch,
both paths alive during cutover.

CI options (D10): the real gap is that our working branch is never pushed anywhere, so
nothing can watch it. Cheapest path: push the branch on commit, let a self-hosted runner
build and test every push (catching broken shared HEAD fast), and add a scheduled
release train — one or two fixed times a day when a release is built from committed
HEAD, provenance-verified, with you keeping a skip switch. Full continuous deployment
stays gated on the delivery-semantics work, but build-and-test CI has no such gate.

Worktrees (D11), plainly: today every session works in the same directory, so they see
each other's half-finished edits, and two sessions editing the same file take each
other's changes when either commits. A worktree gives each coding session its own
directory and branch inside the same repository — same history, no shared files. It
fixes the same-file collisions and the stash-class accidents; it does not fix cluster or
database collisions, and it makes merging a real job. Recommendation: coding sessions
get worktrees, documentation stays on the shared tree, exactly one branch deploys.

Burst scaling (D15): yes, mostly with what we have. A "burst profile" is: pause
maintenance (your ruling), temporarily raise the dispatch and worker concurrency
(config), and let the governor put the whole budget behind client builds. The machines
have roughly five-fold headroom before nodes matter; the true burst ceiling is the AI
account, which is why the governor plus the tier request are the two burst enablers.

## 2026-08-24 — the second dispatch lane is live

Today the fleet stopped taking work strictly one site at a time. The change we planned
last week (a second dispatcher slot, N=2) is applied and verified working: two dispatch
turns were observed running at the same moment, and each slot correctly books its own
completions — including the subtle idle path the original plan had missed (there were
three places stamping the ledger, not two; we found the third by counting rather than
trusting the plan). Rollback, if ever needed, is a single switch. The change went to the
review council (which now covers database migrations too) and is awaiting its verdict.
Nothing else was raised: we deliberately stop at two lanes until the spending governor
exists and the one-at-a-time adapters are addressed, because we measured things starting
to fail at five simultaneous pieces of work back in July. Still owed from here: a
deliberate test of two dispatchers picking the same site at the same instant (the design
says the loser walks away cheaply — we want to see it), a day-long before/after
throughput comparison, and then the batch-size increase as the next small step.

## 2026-08-25 — the second lane is safe and helps a little, but not for the reason we said; and there is a cheaper knob

Two things today. First, the review council sent the second-lane change back a second time,
asking for the one test we still owed: two dispatchers picking the same site at the same instant.
Second, when I measured that, the measurement showed the change works differently from how we —
I — described it to you last week. I am correcting that here in full.

**What we believed.** Each row in the scheduler's task table was a "slot": a task could not fire
again until its previous run had finished. So one row meant one dispatch turn at a time, and adding
a second row meant two at a time.

**What is actually true.** The scheduler has been "fire and forget" since March. It marks a task
complete the instant it has sent the message, without waiting for the run to finish. So the single
original row was already firing every 90 seconds regardless, and its runs overlapped — it was never
a slot. The evidence is direct: the original row overlapped with itself 361 times in a day; its
timestamps read "completed" 39 seconds after firing while the run still had a minute to go; and the
code says so in words ("we don't wait for the orchestration to finish").

**What the second row therefore did.** It doubled the number of fires — two every 90 seconds, about
one second apart. Because the second fire comes before the first turn has claimed anything (that
takes about 18 seconds), the two turns pick the SAME site 94% of the time and share its batch of
work. The result is not two sites at once; it is two workers on one site, with 39% of their claim
attempts bouncing off items the other already took. The measured gain in items claimed per hour is
roughly +10–15%, not double.

**The good news.** The safety we cared about held perfectly. Across 2,579 handler runs, no item was
ever worked twice at the same time (the 41 items with repeat runs were ordinary retries, one after
another). And two workers on the same site did not cause failures — the failure rate was actually
lower when a site had two (1.6% against 3.9%). So the change is safe and a small improvement; it is
just not the mechanism we sold it as. The two bookkeeping migrations (582, 583) were doing nothing
on this scheduler — harmless, but not a fix.

**The cheaper knob.** The task table has an interval column, and it is live. Setting the original
row's interval to 30 seconds would fire it every 60 seconds instead of 90 (a 1.5× turn rate);
setting it to 25 would fire it every 30 seconds (3×). Fires that far apart DO see the previous
turn's claim and go to a different site — which is the thing we actually wanted. It needs no
migration, no sibling row and no bookkeeping.

**What I did not do.** I did not switch anything. You authorised N=2 as a sibling row; I have found
the premise was wrong, and which lever replaces it is your call. The options:

- **A.** Leave the sibling as it is: +10–15%, two workers per site, safe, instant rollback.
- **B.** Retire the sibling and set the interval to 30 s: a turn every 60 seconds, each on a
  different site with a full batch — about 1.5× the turns and fewer wasted attempts than today.
  **My recommendation for now.**
- **C.** Same, but interval 25 s: a turn every 30 seconds, about 3× — roughly 3.3 turns alive on
  average against 1.7 today. I would hold this until the spending governor exists, per your own
  caution about running more at once.
- **D.** Keep the sibling and teach the site picker to skip a site that already has a turn running.
  This restores the one-turn-per-site rule you were told we had, at the cost of the two-workers-
  per-site gain.

**Two corrections to earlier numbers on your list.** The throughput ceiling I gave you last week
(~83 items an hour, "one site at a time") was wrong: it is about 200 claims an hour per trigger row,
and turns overlap. And the next planned step — "batch 5→8 plus timeout 300→600 in lockstep" — the
timeout half does nothing on this scheduler; only the batch half is real.

**How the error happened, in one line.** Both code readings looked at the code that *reads* the
timestamps and never at the code that *writes* them, thirty lines above. The document that stated
the "one at a time" ceiling also quoted figures — 17 runs in 25 minutes averaging 218 seconds each —
that already say 2.5 runs were alive on average. I have logged this as a wrong call so the pattern
is on record, and the check ("runs × duration ÷ window") is now in the runbook.

The council will get a third round with all of this stated plainly, so the record is honest
whichever way they rule.

## 2026-08-26 — done as you ruled: one dispatcher, faster, evenly spaced

Your decision (option B) is applied and live as of this morning, 08:51 UTC. The second dispatcher
row is switched off — kept in the table so switching back is one statement — and the original now
fires every 60 seconds instead of every 90, with each fire far enough from the last that it sees
what the previous turn claimed and goes to a different site.

Two safeguards went in with it. The daily check now enforces your ruling mechanically: it fails
loudly if anyone re-enables a second dispatcher row, and it fails loudly if anyone sets the
interval below 30 — which is option C — before the spending governor exists. So C cannot happen
by accident; when the governor lands, changing that one line is part of doing C deliberately.

One number from overnight that confirms the choice: the share of claim attempts the old pair
wasted on each other had grown to almost 60% (it was 39% when I measured on Monday). The change
went to the review council this morning as usual; first measurements of the new behaviour follow
below, and a full day's before/after comparison tomorrow.

---

**26 Aug, late morning — two hours in, the change is doing what we hoped.**

Fires are evenly spaced (median 60 seconds apart). The work is spreading: 22 different sites
picked up in two hours, where the old pair spent 94% of its time piling onto one. Wasted claim
attempts are down from roughly six in ten to about one in ten — and half of that residue was
the tail end of the overnight credit outage, not the dispatcher.

The fleet is clearing the backlog at roughly 200 claims an hour, against a theoretical ceiling
of about 300 at this cadence. Demand is not the limit — there are about 1,270 triaged items
across 30 sites waiting. The batch size (5 items per turn) is now the knob that binds, which is
exactly the next change on the queue once tomorrow's full-day reading is in.

One safety flag came up and was checked. The monitoring showed what looked like two handlers on
the same work item at once — the first time ever. On inspection it was not: one handler had died
mid-job during the credit outage and sat unnoticed until the hourly cleaner removed it; a second
handler legitimately picked up the freed item a couple of minutes before that cleanup was
stamped. The overlap was bookkeeping, not two live workers. The daily check has been taught to
recognise that shape — and to report it rather than hide it — so it doesn't cry wolf tomorrow.

The full before/after verdict on the speed change comes tomorrow morning.

## 2026-08-26 evening — the starvation bug has a fix built, held until tomorrow midday (413 session)

You asked the 413 session to pick up the starvation fix; this is where it stands tonight.

The bug, in plain terms: the dispatcher keeps two lists. One list decides WHICH SITE goes
next — it looks at each site's oldest waiting item and picks the site with the oldest. The
other list decides which items to actually do once a site is picked — it takes the most
IMPORTANT items first, up to five. Those two rules fight. A site can hold one very old but
very unimportant item: the first list keeps picking that site because of the old item, and
the second list keeps skipping the old item because of its low importance. So that site gets
served over and over on the strength of work that never happens, and every site behind it in
the queue just waits. Tonight sixteen of the twenty-five sites with waiting work were stuck
in exactly that shape, and none of our normal dashboards can see it, because a site being
quietly skipped produces no errors — just silence.

The fix is one change to the first rule: a site's place in the queue is now decided by the
oldest item the dispatcher would ACTUALLY DO next for that site, not by the oldest item full
stop. The trap becomes impossible rather than merely unlikely. We proved it against the live
database tonight: the old rule picks the worst-stuck site; the new rule picks the site whose
work has genuinely waited longest. The change is written, tested both ways (we broke it on
purpose and watched the alarm fire), reviewed by the council [corr to follow in the bug
file], and deliberately NOT switched on yet — the throughput lane's 24-hour measurement
finishes tomorrow morning and the batch-size change goes in at half nine, so this switches
on after midday, so each change can be measured on its own.

One decision is yours when you want it, no urgency: even with this fix, a site with only
YOUNG work still waits its turn behind sites with genuinely old work — that's first-come
first-served working as intended, but when the backlog is deep, "your turn" can be hours
away. If you'd rather no site ever waits more than some fixed time regardless of age order,
that's a policy choice (it would serve young sites at the expense of old work), and we now
have the "worst wait per site" meter to tell you whether it's needed. The bug file ranks it
as the remaining option; nothing is lost by waiting for the meter to speak.

Also found while in there, filed as bug 415, not urgent: the cheap "is there anything to do
at all?" check that wakes the dispatcher uses an older, narrower definition of "anything to
do" than the dispatcher itself. Today it never matters because there is always plenty of
work in the form it counts; on the day the backlog drains to only human-approved items, the
dispatcher would stop being woken at all. Cheap to fix, written up, out of tonight's change
on purpose.

2026-08-27 morning (throughput lane). The 24-hour read on the new single-trigger setup came back
good: fires land every 60 seconds as designed, work went to 32 different sites instead of one,
and wasted claim attempts are down from about 60% to about 16%. The backlog is actually
shrinking now — 716 items queued this morning versus 1,268 yesterday — even though new work kept
arriving all night. Items now wait a median of about 2 hours to be picked up, down from nearly 9.

One alarm went off and it turned out to be our own smoke detector, not a fire: the daily safety
check flagged what looked like two workers on one item. Tracing it showed the item was handled
strictly one-after-the-other — the first worker had wedged, the system correctly took the work
back and gave it to a second worker, and the stuck one was only swept up by a cleaner that
writes a DIFFERENT message than the one our check knows about. So the check was taught the
second message, and we proved the fix catches exactly that case and nothing more. Committed.

With all four gate conditions met, the batch-size increase (5 → 8 items per turn) went in at
09:15 UTC and is confirmed live. Honest expectation stays ~7% more throughput, not more — bigger
bites but longer chews. Next: a two-hour check around 11:30, then the other session applies the
starvation fix after midday. On starvation: the worst case this morning is lendzy.co.uk — 55
items waiting, the oldest for over 10 hours, exactly the queue-jumping bug (413) the midday fix
targets. One correction to yesterday's picture: the site that looked worst (adversecreditmortgage)
is actually deliberately locked/parked — not starving.

2026-08-27 early afternoon — ⚠ ONE THING NEEDS YOUR HAND. Since 11:30 UTC every AI call the
fleet makes has been refused with "You have reached your specified API usage limits" — that is
the self-imposed limit on the billing page (the same flavour as 17 August, not the empty-credit
one from Monday night). Still failing 100% as of 13:15. The machinery keeps running — pages
that don't need AI still deploy — but anything needing the AI bounces and retries. Raising or
waiting out the billing-page limit is the fix; this is also the third live case for the "spend
governor" we have queued as the next build item, which exists precisely so we shed load
deliberately instead of hitting this wall.

Otherwise a good morning despite two hiccups. The batch increase (5→8) is in and provably
working — bigger mouthfuls confirmed at full size. Two of the sites that were being
queue-jumped got unstuck just from the bigger batches. The starvation fix from the other
session got its all-clear at 13:20 (late — my own session hit ITS usage cap and sat frozen
10:40–13:10, so the midday handover slipped ~1.5h; nothing was lost, they waited as agreed).
And a nice piece of teamwork: another session reported that a stuck claim can black out a
whole site "forever"; we tested it live, found the safety net they and we had both missed
(a 2-minute sweeper that frees stuck work after 40 minutes), and both corrected our documents
the same hour. The blackouts are real but capped at ~40 minutes — happening about 90 times
today across 27 sites though, so worth keeping an eye on once the AI limit is lifted.

2026-08-27 13:36 UTC — you added credit and the AI calls came back immediately (last failure
13:34:41, clean from 13:35:13). The blackout ran just over two hours. The starvation fix went
live at 13:18 and its very first picks were the sites that had been stuck longest, which is
what we wanted to see; the afternoon readings at 15:20 and 19:20 will now measure it properly.

2026-08-27 15:25 UTC — the 3pm reading on the queue-jumping fix: PASS, comfortably. The worst
wait for any site with hour-old work is now about 68 minutes (it was 6–10 HOURS this morning).
The site that had been stuck since yesterday morning has fully drained its backlog, and the
fleet is back at full speed since the credit top-up (~270 items/hour). One design point stands
for a decision when you want it: individual low-priority rows can still sit at the back of a
busy site's queue indefinitely (two are ~16 hours old now) — the fix stops them starving OTHER
sites, but only the "age floor" option we wrote up would put a ceiling on their own wait.

2026-08-30 evening — back after the frozen days, and two headlines. The good one: the dispatch
system is in the best shape it has ever been. The whole backlog from Thursday is gone, the
queue-jumping fix has held for three days (nothing starving, no stuck work anywhere), and
wasted claim attempts are down to 3.9% — from around 60% before this workstream started. The
bad one needs your hand again: the AI account has been refusing essentially every call since
Friday (same "usage limits" message as Thursday — the top-up ran out or the limit reset). The
site machinery is idling as a result: almost no new work is being generated. Another top-up or
a limit raise brings it all back; and this is now the fourth incident for the spend-governor
case — it remains the top build item.

2026-08-31 morning — your top-up this morning worked: AI calls clean from about 09:00 UTC
(the blackout ran roughly two and a half days in the end). Work is flowing again and the
dispatch side checked out clean — nothing stuck, nothing starving, new work being picked up
within minutes. The one thing waiting on you is the spend governor's design question: when
spend nears the cap, what should be refused FIRST — routine maintenance, new site builds, or
research? Answer that in a sentence and the governor build can start; it is the thing that
turns these blackouts into deliberate slow-downs.

2026-08-31 afternoon — the spend governor's foundations are in and approved by review. The
meter is live (August ran about $2,113 of AI spend, and that's with the blackouts suppressing
it), the shedding order you ruled this morning is wired in as data, and the alarm that
announces a shedding-level change has been tested in both directions. Nothing sheds yet —
that needs the small code change (next step) plus one number from you: the monthly budget the
governor should defend. Give me that number whenever; until then it watches and stays silent.

2026-09-02 — you gave a provisional steer on the queue-ordering question: no reordering
machinery; things flow well and we scale to meet excess demand. I've applied that as a
provisional "no" to the age-ceiling option (the one that would cap how long an individual
low-priority item can wait) and told the other session that ordering decisions come through
this lane so you aren't fielding them directly. What the "no" accepts: during a heavy run of
new work on one site, that site's least-important items can wait many hours — but since the
queue now regularly drains to empty, those waits end naturally. We watch one number (the age
of the oldest waiting item) and revisit only if it trends up. Your clients-first ruling for
paid work during signup bursts is untouched by this. When you're ready to make it final,
one word does it.

2026-09-02 afternoon — the last known gap in the dispatch chain is closed. Since the earlier
fixes, three different pieces of the system each held their own definition of "is there work
to dispatch", and the first of them — the cheap check that decides whether the dispatcher
wakes up at all — was stricter than the ones behind it. In three specific situations (work a
human had approved but nothing else pending; work labelled for a pipeline other than "build";
work on a locked site that had been explicitly excepted from the lock) the system would have
had dispatchable work and simply never woken up to look, with nothing anywhere reporting a
problem. None of those situations occurs today — that's exactly why it was worth fixing
cheaply now rather than diagnosing it expensively later, because the quiet end-of-backlog
state we're steering toward is where it would have bitten. The fix makes the wake-up check
deliberately a little more generous than the dispatcher behind it (a spare wake-up costs
almost nothing; a missed one is invisible), and each of the three situations was proven
against the live database: the old check stays silent, the new one fires. Applied this
afternoon, dispatch carried on without a blip, and the review council is looking at it in
parallel. An old half-finished draft fix for the same problem from three weeks ago was also
defused — it had crept into the migrations folder where a bulk apply could have run it, and
it would have re-broken things in subtler ways.

2026-09-02 evening — your £/$2,000 monthly budget is set and the meter is armed (nothing
sheds yet — that still waits on the release and your final "enable"). One thing said out
loud so it's a choice rather than a surprise: the fleet currently spends about $99 a day
when fully busy, which is roughly $3,000 a month — so at $2,000, once enforcement is on,
the governor will actively slow things down in the second half of a typical month: routine
AI maintenance pauses around day 14, new site builds around day 17, research around day 19.
If holding spend to $2,000 is the intent, that's exactly what you'll get, in that order,
with the dashboard view showing what's paused and why. If you'd rather shedding stay rare,
a bigger number does that — it's one line to change, any time. Also: please make sure the
Anthropic console cap is set ABOVE $2,000, otherwise the account's own wall still arrives
before the governor's gentler one. September's spend will tick along on the meter either
way, so we'll see the real crossing dates within a couple of weeks.

2026-09-02 late — you deployed the fresh build; my access token expired three minutes before
I could verify it carries the governor code (the usual 3-day cycle — a fresh kubeconfig from
the Rackspace console fixes it). Everything is written up for the next session: verify the
build, apply the held configuration, then your "enable" — plus the console cap raise above
$2,000. The handoff file has the exact steps in order.

2026-09-03 morning — you said "enable" and the spend governor is now LIVE. Nothing changes
today: September stands at $373 of your $2,000, well under the first line, so everything
runs normally — but from this moment every piece of AI-bearing work is checked against your
budget before it's picked up, and the staged slow-down you ruled (maintenance first, builds
second, research last) will happen by itself if spend approaches the cap. One number to
know: the fleet is currently burning ~$124/day, a bit hotter than last week's estimate, so
the first line ($1,400 — maintenance pauses) would arrive around the 11th at this pace.
The dashboard question "is this paused or stuck?" is one query (governor_withheld_now).
Last reminder on my list: the Anthropic console cap should sit above $2,000 so the
governor's gentle brake always beats the account's hard wall.

2026-09-03 late morning — thank you for the $3,000 cap; that was the last thing on my list
and it closes the gap properly. Your account's own hard wall now sits a thousand dollars
above the budget the governor defends, so the gentle staged slow-down always arrives before
the hard stop. Nothing else is outstanding from you.

I then spent half an hour watching the governor now that it is actually switched on, because
"it went live and nothing bad happened" is worth very little unless someone looks. Dispatch
is completely unchanged either side of the switch — same number of wake-ups, more sites
served if anything, no failures, and not one piece of work refused. That is what it should
look like at $386 of a $2,000 budget: the governor is watching and saying yes to everything.

Two things I found while looking, neither of them a problem, both worth you knowing.

The first: when a new build is released, the system rewrites every one of its 200-odd agent
configurations in a single stroke about a minute before the new pods start. Our governor
setting was applied by hand directly to the live configuration, so I had to check it had not
been wiped by this morning's release. It had not — it is intact — but it could have been, and
nothing would have told us. I have written the two-line check into the runbook as a
"do this after every release" habit.

The second: I proved the governor's staged slow-down actually works, at all three levels,
without pausing a single real piece of work. The trick is to pretend the budget has been
crossed inside a database transaction that is then thrown away — the live system never sees
it. At the first level it would hold back 51 routine maintenance jobs; at the second, 112 (the
extra 61 being new page builds); routine work that needs no AI keeps flowing at every level,
which is exactly the order you ruled. Sites barely drop out of the queue even when a hundred
items are held, because the holding is per type of work, not per site — so a site with a mix
carries on with the parts that are free.

One honest limit: what I proved is the database half. The worker code that reads the level
has only ever been exercised at "level zero, allow everything". The only way to prove the
whole chain is to see a real slow-down — which at the current spend arrives around the 11th —
or to induce one deliberately: drop the budget for about five minutes, watch the first level
fire and the held-work list fill up, then put it back. Nothing is lost either way; the held
jobs simply wait and resume. Say the word if you want that done today rather than waiting a
week, and I will do it in a controlled window and write up what happened.

2026-09-03 midday — I induced the slow-down you asked for, and it works. Between 11:17 and
11:29 I dropped the budget so the governor thought you were 90% spent, watched it hold work
back, then put the budget straight. Everything behaved: 115 jobs were held (61 new page
builds, 54 routine AI maintenance), the AI-free work carried on the whole time — the machine
actually did MORE of that work while the AI half was held — and when I restored the budget
everything released and resumed. Nothing was lost, nothing failed, and no job burned a retry.
The staged brake you ruled in August is now a thing we have seen happen rather than a thing we
believe about the design.

Two numbers worth having. The governor is about twice as slow to react as its design says —
roughly four minutes to start holding after a crossing, and about four minutes to let go again
after the budget is put right. That is the scheduler's pace, not a fault, but it means the
brake is gentle at both ends. And your spend is lumpier than the daily average suggests: it
moved about $35 in the hour I was working, against a daily average implying $5. September
stands at $404.

Then the useful part of doing a rehearsal: I found something broken. **The governor sheds
correctly but never announces it.** It is supposed to write a line saying "shedding level
changed" each time it moves. Two changes happened today and it wrote nothing. I have traced it
precisely — a small database change made on 31 August, for a good reason, silently killed the
announcement, and the test that had proved the announcement worked belonged to the previous
change and was never re-run. It is a one-token cause and I can show it failing and working on
demand. I have written it up as bug 459 with three ways to fix it.

I have deliberately not fixed it today. The fix edits the live task that recomputes your
spend every couple of minutes — the one thing in this system I least want to break in a hurry
— and that class of change goes through review first. It is the obvious next job.

One thing to know either way, which the rehearsal made visible: the governor only holds back
work on the main build queue. There are three other, much smaller work queues (diagnosis,
reports, deliverables) which were deliberately left outside it, so they keep spending at every
level. That is about 1% of jobs by count, but diagnosis jobs are individually expensive, so I
would not assume it is 1% of the money. Worth measuring before the real crossing around the
11th.

2026-09-03 early afternoon — I need to correct something I told you yesterday, and it matters
for the budget.

Yesterday I said that at $2,000 the governor would slow things down in the second half of a
typical month — maintenance around day 14, builds day 17, research day 19. The pauses are real
and I proved them this morning. **But the implication that this holds your spend to $2,000 is
wrong, and I should have measured it before saying it.**

Here is the shape of it. The governor works by refusing to pick up website jobs. So it can only
touch money spent by website jobs. I measured where the money actually goes over the last 24
hours, by tracing every AI call back to whatever started it: **about 28% comes from the website
work queue, 3% from a smaller queue the governor deliberately doesn't cover — and 69% comes from
something the governor cannot see at all.**

That 69% is almost entirely **the review council** — the system that reviews platform changes
before they ship. It alone is 62% of everything the fleet spends on AI. That's steady, not a
one-off: 70% of Wednesday's spend, 62% of today's. On 1 September, when no reviews ran at all,
the whole day cost $55.

So the honest position: the governor will do exactly what you ruled — pause maintenance, then
builds, then research, in that order, announced on the dashboard — but even with everything it
can pause paused, roughly 70% of the spending carries on. It is a **dispatch** governor, not a
spend governor.

Two things you might reasonably want, and they are your call, not mine:

1. **Extend it to cover the review council.** That is the single change that would let a budget
   actually be defended. It's a bigger design question than a config change, so it would go
   through the architecture review first.
2. **Or leave it and treat the number differently** — knowing that the governor protects the
   site-building half, and that the review council's cost is managed separately (there is
   already a caching change that cut it by about two-thirds earlier in the summer).

One more thing worth saying plainly: this was findable all along. Our own house rules document
already records that the review council was about 85% of all AI spend before that caching fix.
Nobody put that sentence next to the governor's design — including me, until I measured it
today. I've corrected the internal record so the next person reads the governor's coverage as a
measured number rather than an assumption.

2026-09-03 evening — you said "extend it", so the governor now has the foundations to cover
the review council. The database half is built, tested three different ways, applied, and
sent to the architecture reviewers this evening. Nothing sheds council work yet — that needs
a small code change (the second half), which waits for the review verdict and the next
release, and switches on only when deliberately enabled. Same pattern as the first governor.

Two things I'd like you to know now rather than later.

First, a choice that is yours: **at which point should council reviews start being held
back?** I've set it conservatively — last, alongside research, at 95% of budget — because
"if it comes to the crunch" read to me as "late". If you'd rather it were the FIRST thing to
go (the biggest saving, and the reviews are advisory so nothing is blocked), say so and it's
a one-line change. The practical difference: at the current pace, "first" would mean no
automated review of platform changes for roughly the second half of every month.

Second, a safety choice I made on your behalf and want to be visible: any agent type that
nobody has explicitly listed is always allowed to run. That is the opposite of how the
site-work half works, and it's deliberate — a typo in a list should never be able to stop
the whole fleet.

Everything is written up; the reviewers' verdict lands in about half an hour.
