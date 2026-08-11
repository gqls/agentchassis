# NOTIFY 2026-08-11 — owner directive: deploy the gaswholesalers.com logo properly; the 404 is a week old and is blocking acceptance work

From: `staged_component_build` (carrying an owner directive given 2026-08-11, in chat).

**The defect.** Every page on gaswholesalers.com 404s `/assets/images/logo.png`. First
recorded 2026-08-05, called out in the 08-09 handoff ("4+ days"), re-verified HTTP/2 404
at 2026-08-11 10:07Z. It has been blocking Tier-4 acceptance of `fuel-budget-forecaster`
for over a week (its fence's chrome checks cannot pass against a page whose logo 404s).

**The directive: deploy the real logo to the stable path, via the proper deploy route.**
Two cautions from `bugs_open/248` (read both 248 files before touching asset deploys):
the automated undeployed-asset repair path deploys every asset as a HERO under a
placeholder name — do not let it "fix" this; and the generic-topic
`deploy_image_asset` route fails outright on the standing chassis (no storage client) —
the working route is the work item via `build-dispatch-loop`.

When the logo serves 200, tell `staged_component_build` (append here or in their
handoff) so fuel-budget-forecaster's fence can finally run S6.
