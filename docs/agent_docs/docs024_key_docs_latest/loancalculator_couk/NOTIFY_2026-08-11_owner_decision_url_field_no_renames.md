# NOTIFY 2026-08-11 — owner decision: your 8 tool pages KEEP THEIR NAMES; acceptance testing adopts the url_field route

From: `staged_component_build` (carrying an owner decision made 2026-08-11, in chat).

**The decision.** The eight loancalculator.co.uk tools whose page slugs differ from their
component functions (`tool-car-finance-pcp-hp`→`tool-car-finance-calculator`,
`tool-compare-loan-offers`→`tool-compare-loans`, `tool-consolidation-risk`→
`tool-consolidation`, `tool-early-settlement`→`tool-settlement-calculator`,
`tool-loan-repayment`→**`index`**, `tool-overpayment-impact`→`tool-overpayment-calculator`,
`tool-rate-stress-test`→`tool-interest-rate-stress-test`,
`tool-return-damage-checker`→`tool-damage-checker`) will NOT be renamed. Instead the
Tier-4 acceptance dispatcher learns to name the page it means: `url_field` in the
`request_browser_run` step config (already supported by the code, checked BEFORE the name
lookup — `tool_acceptance_actions.go:163-166`), with the work-item spec carrying
`page_url`. Additive and inert until a work item carries the field; goes through the
normal council gate.

**What changes for your lane: nothing on your site.** No URL changes, no redirects, no
page renames. Your golden-values harness (`toolgolden.py`) keeps working against the same
pages. What you GAIN: once the config lands, all eight tools (including the homepage
calculator, which no rename could ever have fixed) become acceptance-testable — real
clicks + screenshot/vision checks on a schedule, raising visible work items on failure.

**Who does the work**: `staged_component_build` (the config change + the spec producers).
Tracked in that lane's `HANDOFF_2026-08-11_continue_here.md` §3b. Questions/objections:
append here or in that handoff — the written claim is the coordination channel.
