# HANDOFF 2026-08-24c — everything shipped and LIVE; next code task is the second-click page

**Read order, cold:** this file alone is sufficient for state. Depth on demand:
`HANDOFF_2026-08-24b_continue_here.md` (the day's full detail — census §3.3, foot-guns
§3.4, hardening §3.6, rate-limit recipe §3.7, verify expectations §3.9) · the morning
`HANDOFF_2026-08-24_continue_here.md` (edge-infrastructure proof table) ·
`PLAN_2026-08-24_build_steps_screen.md` §6 (the measurements the screen is built on) ·
`../webdesign_uk_build_service/RUNBOOK_links_host_box_steps.md` (the box bring-up, done) ·
`../webdesign_uk_build_service/DECISION_2026-08-24_confirmation_needs_a_second_click.md`
(the next build's spec) · `../architecture_review/RFC_054_public_cluster_exposure_boundary_review.md`
(the boundary review, awaiting ruling).

## 0. State in one paragraph

**Everything this lane had in flight is LIVE and verified as of 2026-08-24 ~22:15.** The
owner's console (`admin.apis.uk`, Access-gated) now serves the **Builds screen**: the
21:00 fleet roll's core-manager (`635f2d32f`) carries backend `e6350e74b`
(ancestor-proven) and the same roll's dashboard image `v1.0.1336` carries the SPA —
proven at the served bundle, not the tag (all five markers grepped in
`assets/index-Bqjp4Gs8.js` fetched from the pod, including the last commit's terminate
caveat). The deploy-ordering constraint dissolved: another session's roll shipped both
halves together, in the right order. **`links.webdesign.uk` is live** (hardened vhost +
edge rate limit; verified from outside: `/other` 404 · `/c/x` 404 by design ·
token-shaped 200 through to core-manager · 8×404 then 32×429 on the hammer loop).
`customer_access_tokens` = **0**, handed_over/confirmed = **0/0** — nothing at risk.
Council: builds-screen backend **APPROVED r1** (`45b3c93f…`, three advisory objections
answered in NOTES); prefetch guard **APPROVED r1** (`6b1726ab…`). kubectl token
refreshed by the owner ~21:00.

## 1. NEXT — in order

1. **Owner eyeball (5 min, no session needed):** admin.apis.uk → a site card → **Builds**.
   Expected for apis.uk: the ~67-minute twelve-stage chain (PLAN §6d table). While there:
   click **Terminate** on some months-old EXECUTING_STEP orchestration — expect a success
   message, NOT a 500 (this live-verifies the ADM-002 B2 table fix; sqlmock proved the
   statement, only a real click proves the endpoint; the confirm dialogue now states
   termination is a DB label, not an interrupt).
2. **BUILD: the second-click confirmation page** — the one owed code task, fully specced
   in `DECISION_2026-08-24_confirmation_needs_a_second_click.md`: GET `/c/<token>`
   becomes render-only (page + button), confirm moves to **POST `/c/<token>`** (SAME
   path — the hardened vhost's regex 404s suffix routes), prefetch guard stays, tests
   must assert GET mutates NOTHING at the DB. Files: `internal/core-manager/handlers/
   delivery.go` (+ `delivery_test.go` conventions), route in `internal/core-manager/api/
   server.go`. Council round (internal/ scope), commit with trailer, rides the next roll.
   **This gates the first delivery email** — the webdesign lane cannot send before it is
   LIVE (not merely committed).
3. **RFC_054 awaits a ruling** (owner / architecture track): Q1 is-the-two-door-pattern-
   the-pattern · Q2 delivery-only listener (cap blast radius in the binary?) · Q3 what
   makes door three automatic. Filed with the fresh census; do not re-measure, re-USE.
4. **Follow-up (own council round, not urgent):** `WriteSiteSpecAction`'s deep-merge lets
   `"banned_claims": []` empty a register wholesale (`site_spec_actions.go:554`);
   `source='scheduled'` is the top evidence_base writer (214 of 319 rows all-history,
   counted 2026-08-24). **Census the legitimate-shrink history first** — the scheduled
   refresher may shrink registers by design.
5. **Webdesign-lane dependencies now unblocked/waiting:** delivery-email builder mints on
   `links.webdesign.uk` (canonical emailed-links host) but waits on item 2; the Stripe
   webhook is still 302-swallowed by the parking rule — it cannot go live before the
   shopfront unparks or the rule excludes the path (flagged in their lane).

## 2. What went LIVE today, with its proof (all measured 2026-08-24)

| thing | proof |
|---|---|
| Builds screen backend (`e6350e74b`) in prod | `git merge-base --is-ancestor e6350e74b 635f2d32f` passes; both pods stamp `635f2d32f` |
| Builds screen SPA in prod (`v1.0.1336`) | served bundle greps: "reconstructed from outputs", "EMPTY_EVIDENCE_BASE", "needs_domain_research", "does NOT interrupt", "advisory — prompt text" — all present |
| evidence_base save guard | same ancestor proof; 6 sqlmock tests in `spec_update_guard_test.go`; council-approved |
| `links.webdesign.uk` | outside curls 404/404/200; hammer 8×404→32×429; admin.apis.uk healthy post-restart; apex still parked |
| edge rate limit (10-per-10s, `http.host eq`) | the 32×429 above |
| LANDMINES: evidence_base wrong-shape silent-off | entry synced to doc_notes, verifier dispatched (file itself was left uncommitted carrying the 333 lane's WIP — check `git log` before assuming) |

## 3. Falsifiers

- `customer_access_tokens` **0** and handed_over **0/0** — any non-zero expires every
  "nothing at risk" line above; re-run before trusting.
- The bundle hash `index-Bqjp4Gs8.js` names TODAY's image; a redeploy renames it — re-grep
  the served bundle, never cite this hash as current.
- Tags roll daily; re-ask each pod its provenance stamp per service, never fleet-wide.
- kubectl token expires on the 3-day cycle (refreshed ~2026-08-24 21:00).
- The second-click ruling means: if `customer_access_tokens` goes non-zero BEFORE item 2
  ships, raise it with the owner immediately — a delivery email may be about to go out
  against the ruling.
- A newer handoff here or in `../webdesign_uk_build_service/`.
