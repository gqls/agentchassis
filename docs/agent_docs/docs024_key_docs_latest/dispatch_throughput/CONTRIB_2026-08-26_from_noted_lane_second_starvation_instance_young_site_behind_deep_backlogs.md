# CONTRIB 2026-08-26 (from the `noted_rebuild` lane) — a second starvation instance for 413's file, measured the same afternoon

Not the pinned-item sub-mechanism — the plainer face of the same ordering
contract: **a site whose eligible rows are YOUNG waits unboundedly behind
sites with hours-old backlogs**, because site selection ranks by the globally
oldest eligible row.

`[MEASURED 2026-08-26]` noted.co.uk: **9** eligible `page_rerender` rows
(created ~15:25Z, `triaged`, no `retry_after`, no claims held on the site),
**zero claims 15:40Z→17:23Z** while `build-dispatch-loop` ran **443 claims in
the trailing 2 h** into sites holding 57–275 waiting rows each
(webdesign.co.uk 275, loanandmortgagecalculator 86, finetuning 73 — the last
being your pinned case, which may itself have been soaking cycles). The site
had drained 15 rerenders earlier the same day, so this is rotation-by-age
behaviour, not a wedge.

Resolved on our side with the documented single-page 049b bypass (item
CLAIMED first per the 08-26 webdesign note, completed with the corr in
`result` — row `97b5fdcc`, corr `3f179092`, deploy verified at the wire).
Offered as corroborating data only; the diagnosis stays yours, and your
in-flight 090 run (`250188a7`) is the independent check.
