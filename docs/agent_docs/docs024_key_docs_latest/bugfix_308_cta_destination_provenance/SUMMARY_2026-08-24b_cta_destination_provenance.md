# SUMMARY — CTA destination provenance (bugs_open/308) — 2026-08-24b

**What we're trying to do.** Buttons on client sites sometimes promise one page and deliver
another. We want the platform to notice this by itself, fix it by itself, and prove the fix on
the page a visitor actually receives.

**Where we've come from.** This morning the fix was proven on exactly one page, and the open
question was a backlog of 215 old repair jobs, believed stuck since July across eleven sites.
Re-measuring against today's pages — independently double-checked by a second model — showed
that picture was wrong: the jobs were never stuck, they were "gave up after two tries" labels;
78% of them describe buttons since fixed, rewritten, or deliberately declined; and only 65 of
the estate's 301 actually-wrong buttons were in that backlog at all.

**What we've done.** Dropped the backlog release, and instead swept the worst sites one at a
time with the fixed platform, each after a safety check of its recent history. Four sites went
through today: gaswholesalers, leopardess, ai-agent-orchestration, lendzy.

**Where we are now.** 88 of those sites' 90 machine-fixable wrong buttons are verified corrected
on the live pages — including "Contact Our Sales Team" buttons that had been sending fuel
customers to a calculator. Each site took under an hour, unattended. The two stragglers are
ancient components with no data behind them, which no re-render can touch.

**Where we're going.** Dartsonline gets its sweep after Friday, once the platform's two-strikes
cooldown expires. The rest of the estate's wrong buttons (~130) sit in prose and ported pages a
machine shouldn't rewrite — they flow to the human-review queue. The next build work is Phase C:
a verifier so a repair that changes nothing can no longer report success, which is also what
retires all three "completed but unchanged" classes found today.
