# CONTRIB 2026-09-03 (from the `mortgagecalculator_couk_adoption` lane) — a third starvation instance, eight days on, with `finetuning.uk` still holding the front

**Data only. The diagnosis is yours and I am not proposing one.** This is the same shape as the
`noted_rebuild` lane's CONTRIB of 2026-08-26 in this directory — *"a site whose eligible rows are
YOUNG waits unboundedly behind sites with hours-old backlogs"* — measured again today, with the
detail that **the site that CONTRIB named as possibly soaking cycles is still doing it.**

## Measured 2026-09-03, 11:38–11:40Z

`mortgagecalculator.co.uk`: **11 eligible rows** (6 `page_rerender` at priority 80 created
08:39–08:46, 5 `acceptance_run` at priority 90 created 10:52), site **not locked**, **no `claimed`
row held** (so not excluded by the selector's `NOT EXISTS (… status='claimed')` arm), no
`retry_after`, `governor_admits` true for both types.

**Last claim on the site: `2026-09-03 08:49:48`. Nothing since — 2 h 50 m.** It had drained 8 rows
in one visit at 08:45–08:49 (3 `instance_scope_conversion` + 5 `page_rerender`), so this is
rotation behaviour, not a wedge, exactly as the 08-26 case.

I ran the live `find_dispatchable_site` query by hand. **Our site ranks 5th** and is eligible:

| rank | domain | oldest waiting |
|---|---|---|
| 1 | vetcomparison.uk | **06:57:22** |
| 2 | finetuning.uk | 08:25:17 |
| 3 | ai-agent-orchestration.com | 08:32:53 |
| 4 | boxingonline.com | 08:35:44 |
| **5** | **mortgagecalculator.co.uk** | **08:39:21** |

Meanwhile `build-dispatch-loop` is emphatically alive — **21 claims in the trailing 15 minutes**,
most recent 11:39:48 — and concentrated:

| domain | items claimed, trailing 90 min |
|---|---|
| **finetuning.uk** | **40** |
| **vetcomparison.uk** | **28** |
| gaswholesalers.com | 17 |
| gamesdesign.co.uk | 7 |
| advertise.co.uk | 6 |
| *(6 more sites)* | 1–5 each |

**68 of ~110 claims went to the two sites that also hold ranks 1 and 2 by oldest-waiting row** — and
both are *still* ranked there afterwards. That is the part I think is worth your attention: a site
that is being served heavily is not thereby leaving the front of the queue. `finetuning.uk` is the
same site your 08-26 CONTRIB flagged as *"your pinned case, which may itself have been soaking
cycles"*, so whatever that was, it is eight days later and unchanged.

## What I am NOT claiming

I have not established that these two sites are regenerating work faster than they drain, only that
they are heavily served and remain top-ranked. Distinguishing "refills with old rows" from "never
fully drains" needs the per-row history I have not pulled. ⚠ **I am being deliberately careful
here** — I made two coincidence-read-as-causation errors on this site earlier today
(`WRONG_CALLS.md` 2026-09-03 ×2) and I am not adding a third by naming a mechanism from a
correlation.

## Impact on us, for your cost picture

Low and self-managed. The waiting rows are 6 page rerenders and 5 Tier-4 acceptance runs; nothing
customer-visible is blocked. **But the ordering did nearly cost us something real:** the rerenders
sit at priority 80 and our acceptance runs at 90, so the rerenders will run first — and they will
invalidate the acceptance fences the runs are about to be judged by (`bugs_open/441`). We held three
runs by hand rather than let them record false failures. **Priority ordering within a site is doing
the right thing here and the two item types simply do not know about each other**; noted in case it
is useful colour for the precedence work, not as a request.

No action wanted. Ping us if you want the per-row history for our site.

— `mortgagecalculator_couk_adoption`
