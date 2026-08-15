# Summary — the directory pipeline ran for real, and earned its first honest entries

2026-08-15, second summary of the day (the morning one, `SUMMARY_2026-08-15_guardrails_live_directories_built.md`, ended at "built and activated, never run"; this one exists because that sentence is no longer true).

## What we're trying to do

Build directories of UK financial services providers — mortgage lenders, savings providers, health insurers — where every fact is non-price, cited to a named source, and machine-verified against that source before it is published. These directories feed the finance site fleet. The standard is that nothing appears under a regulated firm's name that we cannot point to, and nothing price-shaped appears at all.

## Where we've come from

This morning the pipeline was built, council-approved, and switched on, but deliberately held so its first runs would happen under supervision rather than unattended. The three weekly discovery tasks were armed with their first fire deferred; the plan was to force-trigger one kind at a time, watch what actually happens, and fix what the watching reveals.

## What we've done

We ran the mortgage-lender kind five times this afternoon, watching every run, and each run taught us something the previous one could not:

- **Run 1** worked mechanically but registered two "providers" that were not providers at all — they were categories, like "FCA-regulated mortgage lenders (general)". Every fact was true and cited; the entities were the problem. The citation checking cannot catch this, because it checks truth, not shape. We added a hard rule: an entry must be one named firm, never a sector or a market. The two category entries were archived.
- **Run 2 failed outright**, and the failure was mine: the improved search wording I wrote was too long for a hidden limit in the search step, and the error message it produces points you at entirely the wrong thing. The limit is now written down where the queries live.
- **Run 3** proved the named-firm rule works: given only market-overview pages to read, the pipeline now correctly extracts nothing, rather than inventing categories.
- **Run 4** found real firms — Nationwide, Yorkshire, Coventry, Family — but every single claim failed verification, for two mechanical reasons: quotes copied as bullet-lists don't survive re-checking against the live page, and one popular industry-statistics site refuses our verification fetches entirely, so anything cited to it can never pass. Both fixed: quotes must now be one continuous passage, and that site is excluded from research.
- **Run 5: everything registered, nothing rejected.** A new firm, Mansfield Building Society, entered the register cited to its own pages.

Along the way we found two deeper things worth knowing beyond this project. First, when a human closes a rejected-claims review item, the system silently discards any new rejections of the same kind for the next three hours — so a diligent reviewer can accidentally hide the very next batch of problems. We know the fix but it needs a code release, so for now the working rule is: review last, after the runs are done. Second, this part of the system writes no visible log lines at all, even though the code says it should — so "no log line" means nothing here, and only the database tells the truth. Neither is fixed yet; both are written down where the next person will trip over them.

## Where we are now

The mortgage-lender directory works honestly end-to-end and holds two real firms with verified, cited facts. That is a small number, and that is by design: each weekly run reads only a few pages, and list pages (the ones naming dozens of lenders) rarely state checkable facts about each firm, so the register grows a firm or two at a time. Every control fired when it should have, and — just as importantly — refused to fire when the material didn't deserve it.

## Where we're going

The savings-provider and health-insurer kinds get the same supervised first-run treatment; their search wording already carries everything the mortgage runs taught us. After that, the remaining wiring steps (publish trigger, the improvement-loop and planner connections, enabling the checks) and then the Phase C pilot on remortgagecalculator.uk. One decision worth the owner's eye at some point, not urgently: whether a firm-or-two per week of growth is acceptable, or whether the researcher should learn a second hop — read the list page, then visit each named firm — which would grow the register faster at somewhat higher cost.
