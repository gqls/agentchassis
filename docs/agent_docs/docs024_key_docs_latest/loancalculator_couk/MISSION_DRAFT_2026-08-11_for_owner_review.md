# MISSION DRAFT — loancalculator.co.uk fresh submission (step 7) — FOR OWNER REVIEW, do not fire

> The text between the markers is what `082_submit_domain_unified.sh --mission-file`
> would receive. Drafted from the live strategy spec (domain-strategist, 2026-08-08)
> plus the owner's framing of 2026-08-10. The rebuild's audit purpose is deliberately
> NOT in the mission — the mission weights the classifier, and telling the pipeline it
> is being audited would distort the very behaviour we want to observe.

<!-- MISSION START -->
loancalculator.co.uk is the definitive free UK loan calculator site: an authority
portal whose primary product is a suite of interactive calculators, supported by
plain-English guides that answer the questions the numbers raise.

The site already serves eleven calculators (standard repayment, overpayment,
consolidation, car finance PCP-vs-HP, loan comparison, credit health check, credit
roadmap, early settlement, interest-rate stress test, loan-vs-savings, application
tracker plus a damage checker), thirteen guides, and a legal page. Keep this page
set and keep every live URL exactly as served — the URLs are indexed and must not
move or change shape.

Every calculator must actually compute what it claims, client-side, with the
arithmetic visible in the page. Never state a capability a tool does not have: if a
calculator shows three headline figures and no month-by-month schedule, no copy
anywhere on the site may say it shows a breakdown. Accuracy claims are promises about
other people's financial decisions.

Voice: plain English for UK borrowers. Explain the mechanism, not just the figure —
why early payments are mostly interest, what a settlement figure includes, what the
Consumer Credit Act actually entitles a borrower to. Cite UK legislation and named
institutions where relevant. Explicit non-advice positioning throughout: the site
shows numbers and rights, it does not recommend products. No promotional tone, no
urgency, no invented statistics, no testimonials.

Revenue is affiliate click-through to lenders and comparison partners at the moment
a completed calculation makes the next step obvious, plus display advertising —
but trust converts, so the editorial layer must never read as marketing. A visitor
is satisfied when they have run the calculation they came for, understood what the
output means, and left with a clear next step.
<!-- MISSION END -->

## What the owner should check before this fires

1. **The page-set-and-URLs sentence** — it instructs preservation. The url_shape flag
   (seeded, live) enforces the SHAPE mechanically; this sentence is the belt to that
   brace at the planning layer. If you would rather let the planner propose a different
   page set, strike it, and the pre/post URL-diff verification (step 8) becomes a
   report instead of a gate.
2. **The capability-truth paragraph** — this is the direct answer to the two false
   claims that started this. It is a standing instruction to the writer, not a fix to
   the two sentences (those die with the rebuild anyway).
3. **What is deliberately absent**: any mention of the audit, the rebuild, the
   framework, or the hand-authored history. The submission must look like any owner
   brief so the framework's ordinary behaviour is what we observe.

## Fire sequence, once approved (from the handoff + this session's state)

```bash
# 1. release the 20 locks (SQL from the 20-row state in NOTES — 17 pc + 3 sc)
# 2. fire:
scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh \
  loancalculator.co.uk --email uk@websy.uk \
  --mission-file <this file's MISSION block, extracted to a plain .txt>
# 3. monitor children via parent_orchestration_id (the printed id logs nothing)
# publish→start can be ~30 min; no dispatch within ~300s of any chassis restart
```
