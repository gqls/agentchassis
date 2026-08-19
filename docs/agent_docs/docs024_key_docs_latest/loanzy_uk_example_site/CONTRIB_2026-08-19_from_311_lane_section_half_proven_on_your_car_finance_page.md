# CONTRIB from the bugfix_311 lane — the section-half fix is PROVEN on loanzy.uk's car-finance page (2026-08-19 20:16 UTC)

Written for whoever picks up `loanzy_uk_example_site`. Full evidence:
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`
(entries from "~16:20Z" onward); recipe: `.../bugfix_311_component_keys/RUNBOOK_311_fix.md`
("The re-drive that WORKED").

**What I did on your site, and why.** Nobody had run the agreed protocol (your handoff
`HANDOFF_2026-08-19_fixing_the_one_shot_route.md` §"The protocol for the next clean-domain run")
and the portfolio sites are locked under the owner's halt, so I ran ONE case here, on
`tool-car-finance-calculator` only — the page that was serving with zero `<input>`. Nothing else
on loanzy was touched.

**Result against your three checks:**
1. Diversion worked — item `9d16951e` complete attempt 0; new base row `2e497429`
   `function=loans-car-finance-calculator-loanzy-uk`, `section_type=loans-car-finance-calculator`,
   one `COMPONENT_COLLISION_DIVERTED` row in `agent_error_log`.
2. No collateral — all eight loanandmortgagecalculator.co.uk incumbents byte-identical to the
   md5s pinned beforehand (incl. the three in your handoff).
3. The served page — `https://loanzy.uk/tools/car-finance-calculator/index.html` now 200,
   38,912 B, **4 `<input>`** (was 25,703 B, 0); no `{{` leakage; its JS asset serves (3,516 B).

**One thing your handoff's premise needs corrected, which also corrects mine:** the parked
pages do NOT converge on their own. `check_unresolved_sections` flips `build_status` to
`needs_rebuild`, but nothing consumes that status (`page-rebuild` has never run). The page
only rebuilt because I filed a `needs_page` item (`page_rerender:tool-car-finance-calculator`,
the same shape `flag_page_image_rebuild` uses). So the six remaining hollow tool pages each
need a `needs_new_component` re-drive AND a `needs_page` re-render — or your full rebuild,
which files both. I stopped at one deliberately: your site, your planned run. Expect 5 of 6
to divert; `loans-credit-health-check` fails UPSTREAM (`generate_template`, max_tokens 16000,
twice) — a different defect, not 311's.

**Your clean-domain run is no longer the first exercise, but it is still the first
GREENFIELD one** — the one-shot path filing the item itself rather than by hand. Worth running
for that reason alone.
