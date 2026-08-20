# SUMMARY — bug 321, 2026-08-20 (lane complete)

**What we're trying to do.** When the system asks an AI which interactive tools would
help a website, each suggestion is supposed to become a work item that gets built. A
bookkeeping flaw meant almost all of them were silently thrown away — about 72% of
every batch — and nothing anywhere reported it. We set out to cure that loss, for the
whole framework rather than the one agent it was spotted on, and to make sure the
mistake cannot quietly come back.

**Where we've come from.** The bug was filed on 2026-08-19 by a neighbouring lane:
the label used to file each suggestion was built from the website's name alone, so on
any one site every suggestion after the first collided with an open duplicate and was
dropped. This exact class of mistake — "the label is coarser than the finding" — had
already been written up as a known trap two weeks earlier, and it happened again
anyway, which told us a written warning was not enough. Two other agents in the fleet
had the identical flaw waiting to fire, one of them about to be switched on by
another team's fix.

**What we've done.** Three things, all in one day. First, the cure: a database
migration gave all four affected filing steps a per-item label (the mechanism already
existed in the code, proven by two other steps that used it), plus a tolerance
setting so one malformed suggestion costs one item rather than the whole batch.
Second, the standing guard: a daily automated check that scans every agent definition
for this shape and reports every morning — including an explicit "all clean" line, so
silence means the check is broken, never that all is well. Third, a small runtime
warning in the shared filing code that notices the one variant the daily check cannot
see, reviewed and approved by the council. Everything was proven by demonstration:
the check was fired at yesterday's flawed configuration and caught all of it, and a
live re-run on the same site that lost six of seven suggestions in the morning
produced five items from five suggestions in the evening.

**Where we are now.** Finished, and everything verified on the live system this
morning: all five suggested tools actually got built (first attempt, about fifty
minutes end to end, zero failures — the build-volume increase the owner approved);
the runtime warning is confirmed inside the running binaries on both replicas; the
council approved the code change; the daily check ran on schedule this morning and
reported clean across all 193 agents; and no filing step has hit an error since the
fix. The bug file has moved to the closed pile with the evidence written in.

**Where we're going.** Nothing further is planned on this lane. The daily check now
owns the future: any new loop-filed work item missing its per-item label will be
named in the morning report. Two small things live elsewhere: the neighbouring lane
that revives the internal-linker will use our shared query to confirm its first real
run files one item per planned link, and the owner may revisit the no-throttle
decision once a few more suggestion runs show the real build cost per run.
