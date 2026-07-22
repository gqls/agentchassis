# SUMMARY — vonc gauntlet: dead CTAs fixed, detector hole closed (2026-07-22)

## What we're trying to do
Make the vonc.com "Gauntlet" tool actually work (the owner reported its link did
nothing), and make the fix generic so any new site is protected from shipping the
same kind of broken, placeholder control.

## Where we've come from
The gauntlet page loaded and its widget (timer, checkable objectives, progress,
counters) worked — but its two hero buttons, "Enter the Gauntlet" and "Preview
Rules", were both `href="#"`: dead by construction (the component's schema had button
*labels* but no button *URLs*). Its headline stats and leaderboard were fabricated
placeholders — there was no real gauntlet behind the button. The platform already had
a detector built for exactly this ("dead_controls"), whose own source names the vonc
gauntlet as its proof case — but it never fired on it, because it only scanned pages
flagged `build_status='deployed'` and the gauntlet serves live while flagged
`needs_rebuild` (one of ~34 such fleet pages). The detector missed its own poster child.

## What we've done
1. **Fixed the gauntlet, honestly and live.** Removed both dead CTAs; "Enter the
   Gauntlet" is now a real button that starts the clock and drops you into the
   challenge, "Preview Rules" reveals a "How the Gauntlet works" card. Stripped the
   invented stats and the fake leaderboard entirely; the page now presents an honest
   self-paced solo challenge ("no sign-up, nothing is scored or shared"). Delivered
   via the sanctioned owned-page path (section-editor) + an assemble-only rerender to
   republish the JS asset. Verified live on vonc.com.
2. **Fixed the generic detector.** Changed `dead_controls` to judge liveness by the
   component that actually serves (`pc.build_status='deployed'`) instead of the
   drifting page-level flag. Council-reviewed and APPROVED (corr 1834a349); committed
   (01e18019a); ships with the next chassis image.
3. **Carried the owner's directive to the council** — "we shouldn't be creating
   placeholders that don't work" — as the rationale the seats reviewed the fix against,
   so it's on record with the reviewers who judge platform changes.

## Where we are now
The gauntlet is genuinely functional and honest, live. The detector fix is committed
and council-approved, inert until the next image roll (no fleet rebuild forced for a
detection-only change). Working docs + reusable delivery scripts + the two delivery
landmines (section-editor doesn't republish JS; the bare 049b page-rerender envelope
doesn't ingest) are recorded.

## Where we're going
Optional, owner-gated: give the gauntlet one real backend action — let a visitor submit
their "Position" via the existing contact-form delivery — a small, honest feature. The
full competitive backend (accounts + a live leaderboard) waits until there's real
traffic to populate it, so we never ship a simulated crowd again. Post-image-roll:
confirm the detector now flags a live `needs_rebuild` page's dead control.
