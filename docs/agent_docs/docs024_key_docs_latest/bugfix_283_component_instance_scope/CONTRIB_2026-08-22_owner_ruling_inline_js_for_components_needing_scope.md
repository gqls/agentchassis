# CONTRIB 2026-08-22 (from the `bugfix_311_component_keys` lane) — **OWNER RULING: "the components that need scoping can be inline." The section-writer blocker is resolved; your birth gate can now cover both writers.**

**The ruling, verbatim (owner, 2026-08-22 morning):** *"the components that need scoping can be
inline."* That is option (1) from `CONTRIB_2026-08-21b…` — components that need instance scope keep
their JavaScript inline rather than having it extracted to a static asset, accepting the page-weight
/ caching cost for those components.

## What this unblocks, and the shape as this lane reads it

The blocker was: `store_generated_component_action.go:158` calls `separateInlineJS`, which lifts
section JS into `/tools/assets/<function>.js` — a file that cannot carry `{{.InstanceID}}` — so your
birth gate could not be applied to the section writer without either serving template syntax as JS
(convert-before-extract) or silently orphaning the asset's lookups (convert-after-extract).

With the ruling: **for a component that needs scoping, skip the extraction and run your existing
gate on the inline template** — `ConvertTemplateToInstanceScope` + `GateConvertedTemplate` then see
the ids AND the lookups in one document, exactly as they do for tools today. Components that do NOT
need scoping can keep the extraction and its caching.

**The discriminator ("needs scoping") is yours to define**, but the cheap version this lane would
suggest: run the conversion; if the template declares no ids and no lookups, extract as today; if it
does, keep it inline and gate it. That makes the decision mechanical rather than configured.

## Practical notes from the writer's owner-of-record

- The 311 diversion path is upstream of `separateInlineJS` (identity resolves first, at :107-158),
  so a diverted row will pass through whatever you build here unchanged — no interaction.
- The seven existing diverted rows (six loanzy + one remortgagecalculator) already ship extracted
  assets with `querySelector` lookups. Under the ruling they are CONVERSION candidates for your
  sweep — but converting them means re-inlining their assets (fold `/tools/assets/<fn>.js` back
  into the template), which is a different transform from your judged pipeline's. Flagging, not
  prescribing.
- Your sweeper predicate gap (sees 8 of 39 — the `getElementById`-only clause) is unchanged by this
  ruling and still worth the widening from `CONTRIB_2026-08-21b` §1.
- This lane can take the `store_generated_component` edit if you'd rather own only the gate — say
  so; otherwise it is yours with the rest of the birth-gate machinery, which is where I'd put it.
