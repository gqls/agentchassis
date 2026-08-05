# SUMMARY 2026-08-05 — operator bulk page rebuild (`features_open/021`)

**What we're trying to do.** Give an operator a real, supported way to say
"rebuild these specific pages, for this reason" — instead of hand-mutating
old work-item rows, which is silently lossy (the row lies about what
happened) and races the platform's own stale-request cleanup. This is one of
five site-quality automation ideas filed together in late July; the others
(brief-fidelity audit, component adoption, design critic) all end in "…and
then the affected pages get rebuilt," so this is also groundwork for them.

**Where we've come from.** Filed 2026-07-25 from a real, painful five-page
rebuild during the fundamentallyai.com rollout. The filing's own design
called for four things: an entry point, sequencing, a new "this is a rebuild"
work-item type, and recompose-vs-rerender intent — with a fix to the stale
reaper (`bugs_open/070`) as a stated prerequisite. It sat untouched for eleven
days. That reaper fix landed 2026-07-27 and its own closing note said the
block on this feature was lifted; nobody had come back to check.

**What we've done.** Re-verified every claim in the original filing against
the live system rather than trusting an eleven-day-old snapshot, then read
the actual mechanism in full rather than the filing's summary of it. That
turned up a genuine correction: the dormant paved road this feature is meant
to activate (`maintenance_queue` → `maintenance-triage` → `page-rebuild`)
never touches the table the stale reaper watches at all — so two of the
filing's four planned pieces (the new work-item type, and the reaper
dependency) turn out not to be needed. What was actually missing was
narrower: one operator-facing script. Built it, and testing it caught two
real problems before either reached anyone else — a command-tag parsing bug,
and a deeper design gap where the script's own "just show me, don't do it"
mode wasn't actually previewing what the operator asked for. Both are fixed
and re-verified.

**Where we are now.** The script exists, is tested safe in its preview mode
(confirmed: zero database writes, zero dispatches), and is ready to fire for
real. It has not been fired for real. The one candidate site tried so far
carries seven pages already needing rebuild for reasons this session doesn't
know, and the mechanism would sweep those in alongside whatever is explicitly
requested — a live first test needs a deliberately chosen target, not
whichever site was convenient to test the safe path against.

**Where we're going.** Pick a real target with a deliberate accounting of
"what else would ride along," fire the script for real, and verify a page
actually rebuilds end to end (`build_status` flips, the page redeploys, the
content genuinely changed for the stated reason). After one clean real run:
decide whether `intent` (recompose vs. cheap re-render) is worth building —
it isn't wired to anything today, and the original use case never needed
it — and whether dispatch volume ever justifies this getting its own Kafka
topic the way council-gate did. Neither is blocking; both are open questions
for whoever runs the first real test to revisit once there's real usage to
look at.
