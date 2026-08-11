# NOTIFY 2026-08-11 — dartsonline setup-builder has invisible option text (1.06:1), and the owner has decided the fix happens at the SHARED COMPONENT, not your site

From: `staged_component_build` (carrying an owner decision made 2026-08-11, in chat).

**The defect on your site.** On `tool-setup-builder`, the pre-selected option in each
group ("Beginner", "Smooth and fluid", "Pinch grip") and the "Get my recommendation" CTA
render at **1.06:1 contrast** — invisible, desktop and mobile. Found by the acceptance
agent's vision pass on its first-ever run (2026-08-11, run `0ee53904…`; every selector
check passes because the text is present). Correct pairing on the same page would be
14.65:1.

**Why it is NOT your palette.** Both tokens hold the values dartsonline intends. The
shared component rule `.db-option input:checked + label { background:
var(--color-primary); color: var(--color-surface); }` ASSUMES surface contrasts with
primary; on dartsonline both are near-identical dark navy. The idiom is live in 9
components across 8 domains — 6 are fine, dartsonline (1.06:1) and mortgagecalculator
(2.95:1) fail. Full working: `staged_component_build/NOTES` `## 2026-08-11 (parallel
session)`.

**The decision (owner, 2026-08-11): BOTH halves.** (1) Fix the shared component template
(a proper on-primary pairing) and rerender affected pages — your page gets its text back
without you changing anything; (2) add a build-time palette-contract check so a future
site cannot regress silently. `staged_component_build` carries the work; your page will
be rerendered when the template fix lands. Nothing is asked of your lane except: do not
hand-patch the page CSS in the meantime — a local patch would be overwritten by the
rerender and would mask the template-level proof.
