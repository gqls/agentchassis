# PLAN — 2026-07-28 — bugs_open/131 (og-card slug): every page advertises a social preview that 404s

Case file: `bugs_open/131_HANDOFF_2026-07-28_og_image_points_at_a_card_that_was_never_generated.md`.
**NB `131` is one of the documented ambiguous numbers** — the other 131 is the vonc gauntlet
usability audit, owned elsewhere. Resolve by slug; `git log` the file path, never the number.

Grew out of the relojistas / fleet-discoverability lane
(`traffic_probe/HANDOFF_2026-07-28_continue_here.md` §4 item 2).

## The defect, restated

`render_site_components_action.go:448` writes `og:image` (and `:451` `twitter:image`)
unconditionally, pointing at `/assets/images/og-card.png`. On 11 of 14 live sites that file
does not exist, so every share of every page on those sites renders with no preview at all.

## What measuring changed — the fix ORDER in the case file is wrong for this estate

The case file ranks fix 1 (suppress the tag unless the asset exists) ahead of fix 2 (generate
the card), on "what closes the door" grounds. Measured against the live system on 2026-07-28,
that ordering costs more and delivers less:

| | fix 1 — suppress the tag | fix 2 — generate the card |
|---|---|---|
| outcome | *no* preview | a *working* preview |
| code change | Go, `render_site_components_action.go` | none |
| needs council + build + roll | yes | no |
| needs chrome re-render on 14 sites | **yes** — head is a stored artefact (`bugs_open/117`) | no |
| needs page redeploy | yes | only the asset commit |
| available today | yes | **yes — all 14 live sites have an active `logo`** |

The decisive measurement: **every live site already has an active `logo` asset**, which is
`derive_brand_head_assets`'s only precondition. So fix 2 is available fleet-wide *right now*
and needs no code at all — the tag already points at the right path; the file is simply absent.

**Both still belong.** Fix 1 remains the structural guard for a future site whose logo is
missing or whose derivation failed. It is second, not first.

## Decisions

1. **Fix 2 first, piloted.** Queue `needs_brand_head_assets` for one site (relojistas — this
   lane's own), verify the card on the wire end-to-end, then decide on the remaining 11.
2. **Do NOT re-derive leopardessconsulting.co.uk.** Its card serves 200 and was hand-made from
   an owner-approved logo (`docs/leopardessconsulting/RUNBOOK.md` H4, resolved 2026-07-10).
   Re-deriving would overwrite an approved brand artefact. Its missing `og_card` provenance row
   is a bookkeeping gap, to be backfilled, not a reason to regenerate.
3. **The gate for fix 1 cannot be "an `assets` row exists".** That is the obvious design and it
   is wrong: leopardess has a working card and *no* row, so a row-gate would regress the one
   site whose preview works. Whatever fix 1 keys on must not have that false negative.
4. **Fix 1 follows the sprites.css precedent** already in the same function — Phase I2 gates
   `<link rel="stylesheet" href="/assets/css/sprites.css">` on an active sprite-sheet asset
   "otherwise the `<link>` would 404 on sites without one" (`:704-712`). The og-card case is
   the same question, and the comment at `:701` explicitly waved it away as "harmless if they
   404 until derivation runs". It is not harmless; that is this bug.

## Phasing

- **P1 (done first)** — pilot fix 2 on relojistas; verify card 200 on the wire.
- **P2** — roll fix 2 to the remaining 11 (owner check first: it puts a generated image on 11
  live sites).
- **P3** — fix 1, the code gate, through the council. Separate commit.
- **P4** — backfill leopardess's `og_card` provenance row.
- **Out of scope, filed not fixed** — the `og:title` / `og:description` fallback (case file
  "second defect"); and the `undeployed_asset` detector blind spot found below.

## Found while measuring — a second, separate defect

The `undeployed_asset` detector has fired 5 times for `og_card`, **every one of them on
robot-hands.com — the one site whose card serves 200.** It has never fired for any of the 11
sites that genuinely lack a card, because its denominator is the `assets` table and those sites
have no `og_card` row at all. A detector that can only see assets that were *generated* is
structurally blind to the ones that were *never generated* — which is the actual failure here.
Not fixed in this lane; see NOTES for the evidence.
