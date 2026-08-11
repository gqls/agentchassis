# NOTIFY 2026-08-11 — mortgagecalculator.co.uk fails AA on checked-option text (2.95:1) on two calculators; fix decided at the shared component

From: `staged_component_build` (carrying an owner decision made 2026-08-11, in chat).

Same mechanism as the dartsonline case (see
`fixloop_eg_dartsonline/NOTIFY_2026-08-11_setup_builder_invisible_text_owner_decision.md`):
the shared checked-option rule writes `--color-surface` text on `--color-primary`
background; on your site that is `#b59230` on `#ffffff` = **2.95:1**, below the 4.5:1 AA
floor (and below AA-large), on `tool-bridging-compound` and `tool-rate-scenarios`. Your
palette is as intended — the shared component's assumption is at fault.

**Owner decision 2026-08-11: the fix is at the shared component template plus a
build-time palette-contract check.** `staged_component_build` carries it; your two pages
get rerendered when it lands. Please do not hand-patch the page CSS in the meantime.
