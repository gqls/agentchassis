# SUMMARY — the owned-page door, 2026-08-26 (bug 333 CLOSED)

**What we're trying to do.** About two dozen parts of the platform notice problems on web pages and
file "fix this" tickets, choosing which worker should do the fixing. Some pages — the interactive
tools, and hand-curated pages — are marked as not the general page-writer's to touch. Nothing that
filed tickets ever checked that mark, so tickets for those pages went to the one worker that is
forbidden to act, were refused, and were recorded as "we decided not to fix this" — on real defects,
over and over. We set out to make that impossible.

**Where we've come from.** The core fix — one check at the shared write path, parking a ticket
visibly instead of routing it into a guaranteed refusal — was built, council-approved and proven
within a day. What kept the bug open was the residue: one producer wrote straight to the database
around the check, and one escalation path resolved a page's identity and then dropped it, so its
tickets passed through the check invisibly. An adversarial re-review corrected three of our own
claims about that residue before the owner ruled on the remedies; a first council round sent us back,
rightly, to measure a side-effect we had only asserted and to make a silent skip leave a durable
record.

**What we've done.** Both remedies are live: the report-writer files through the shared path, and the
escalation checks ownership before raising its alarm, leaving a per-page review record either way.
Every claim was proven at the running binary (four deployments in a row) and then on real traffic.
Along the way this lane traded eight corrections with five other work-threads — in both directions,
every one caught before a document inherited it — and the register now holds, in one place, the dated
evidence for the bigger open question this bug deliberately did not answer.

**Where we are now.** Closed, on the strict bar. Twelve real escalation attempts on owned pages since
the fix: all twelve skipped, zero refused, zero false tickets — including three repeats each on the
very pages that were burned before. Five audit findings on owned pages: none refused. One honest
caveat is written into the close: the audit path is currently protected first by a newer,
config-level mechanism (findings from opinion-seats are recorded rather than dispatched); our check
remains the durable backstop underneath it, and a daily auditor now watches the configuration the
check depends on.

**Where we're going.** Nothing further in this lane. The genuinely open question — what SHOULD write
and repair the content of owned pages, for which the parked tickets are standing, quantified demand
(48 link-less pages, 40+ parked findings, a cross-link feature whose every correct pick parks) — has
no bug of its own by design; the evidence sits in the concept register's demand block, waiting on an
owner decision to revisit the "not now" ruling.
