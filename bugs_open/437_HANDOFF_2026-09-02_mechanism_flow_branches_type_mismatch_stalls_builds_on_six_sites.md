# 437 — mechanism-flow's `steps[].branches` contract is unwritable by its own writer: 119 failed builds on six sites in 14 days, pages stuck `planned` for weeks

**Filed 2026-09-02 by the loanzy lane.** Diagnosis loop NOT run — substituted first-hand
verification, stated per the 2026-07-31 ruling: the error is the system's own, verbatim and
recurring; the spread is a one-query census; the stuck pages are lifecycle rows read directly.
Found investigating why 3 of loanzy's 30 active pages 404 (build_status `planned`/`needs_rebuild`,
deployed_at NULL since 2026-08-18) while live pages link one of them (`/your-rights.html`) inline.

## The chain, verified at the rows
1. `page-build-handler` fails, verbatim error: `failed to execute action render_component:
   component "mechanism-flow": content does not match the declared field type(s) —
   steps[0].branches: declared array (items: object), got string; steps[1].branches: …`
2. The failure repeats until the two-strike arm brands the item
   `[unresolved after 2 attempts]` — the repair queue looks handled; the page stays unbuilt.
3. **Spread [MEASURED 2026-09-02]: 119 `content does not match the declared field type`
   failures in 14 days**, mechanism-flow/branches on six sites: remortgagecalculator.uk 53,
   loanzy.uk 35, farmerinsurance.uk 24, cv1.co.uk 4, mortgagecalculator.co.uk 2,
   advertise.co.uk 1.
4. Symptom at the artefact: ACTIVE pages that never deploy — loanzy `/your-rights.html`
   (linked from at least two live pages' body copy), `/guides/index.html`,
   `/guides/tool-loans-consolidation-guide.html` — the INVERSE of closed 359
   (there: retired-but-serving; here: wanted-but-never-served, and nothing escalates it).

## Mechanism (narrowed, not proven to the line)
The `mechanism-flow` component declares `steps[].branches` as an ARRAY of objects; the
content writer emits a STRING for it. Either the schema changed under the writer, or the
writer's prompt never learned the nested shape — `bugs_closed/260` is the ancestor class
(one mistyped LLM field), whose fix made the render REFUSE loudly instead of degrade
silently; this bug is the refusal with no working repair path behind it. `bugs_open/348`
is the adjacent-but-different arm (refusals that report complete).

## Fix candidates (ordered by what closes the door)
1. Make the bad state unwritable at the WRITER: the component's field prompt/exemplar must
   produce the declared nested shape (or the schema must accept the string form and coerce)
   — component-contract family work (357/388 lanes; the brochure_component_library lane
   knows mechanism-flow's template deeply).
2. A repair path for type-mismatch refusals: today's only outcome is fail → two-strike →
   unresolved; nothing re-plans the section or regenerates the field with the error in hand.
3. Escalation gap (its own small bug if split): a page ACTIVE + `planned`/`needs_rebuild` +
   never deployed for N days, while other live pages link it, surfaces nowhere.

## Verify
- The error: `SELECT left(error,300) FROM site_work_items WHERE error LIKE '%mechanism-flow%branches%' ORDER BY updated_at DESC LIMIT 1;`
- The spread: the §3 census query (same predicate, GROUP BY site).
- The stuck pages: `SELECT url, build_status, deployed_at FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='loanzy.uk' AND p.status='active' AND p.deployed_at IS NULL;`
