# HANDOFF 2026-08-26 — everything this lane owed is LIVE and proven; one test waits on the owner

**Supersedes `HANDOFF_2026-08-25_continue_here.md` entirely.** Depth on demand:
`NOTES_web_admin_console.md` (2026-08-25 and 26 entries — probe tables, missteps, the census) ·
`README_where_we_are.md` (the owner's plain-prose log) ·
`../webdesign_uk_build_service/NOTE_2026-08-25_second_click_built_by_web_admin_console_lane.md`
(what the other lane needs from us) · `PLAN_2026-08-24_build_steps_screen.md` §6 (the
measurements the Builds screen is built on).

## 0. State in one paragraph

**Nothing in this lane is blocked on a session.** The second-click confirmation page is LIVE on
the public internet and proven end to end from outside; the Builds screen is live and has had
two rounds of fixes; the deep-merge follow-up was closed as REFUTED by its own census; RFC_054's
three questions are all RULED and Q2's listener is built, shipped and live. `customer_access_tokens`
= **0**, `transfer_confirmed_at` = **0**, `handed_over_at` = **0** `[MEASURED 2026-08-26]`.
**The one outstanding item is a test that needs the owner's authorisation** (§2). ⚠ Another
session is actively working this same screen and shipped two rounds overnight — read §4 before
touching `App.tsx` or the admin handlers.

## 1. Verified LIVE this morning, at the artefact

Fresh build `v1.0.1341`; core-manager provenance stamp **`2fb40a960`**, one distinct stamp across
all pods (a release can straddle — check per service, never fleet-wide).

| thing | proof `[MEASURED 2026-08-26]` |
|---|---|
| second-click page, from the **public internet** | `GET https://links.webdesign.uk/c/<43-char>` → **200**, `text/html`, `cache-control: no-store`, `referrer-policy: no-referrer`, `x-frame-options: DENY`, body `<h1>Confirm you have moved your site</h1>` + `<form method="post">` (no `action`, so the token never enters the HTML) |
| the button's route | `POST` same path, unmatched token → **200** "That link is no longer active." — routing, handler and DB lookup exercised, nothing mutated |
| **CONTROL** the box still refuses everything else | `POST …/c/<token>/confirm` → **404 with nginx's own HTML body** (died at the box, never reached the service); `GET /other` → 404 |
| my three commits still ship | `24b63120d`, `d1a4bdcdf`, `d30917150` all ancestor-proven against `2fb40a960`, reversed control fails as it must |
| delivery listener wiring survived the roll | `SERVICE_SERVER_DELIVERY_PORT` = `8090` in the pod; live egress policy allows `core-manager: 8088 8090` |
| Builds screen in the served bundle | dashboard `v1.0.1341`, bundle greps for this lane's strings **and** the other session's new terminate copy, control string absent |

## 2. THE ONE OPEN ITEM — needs the owner, not a session

**The SUCCESS arm of the confirmation flow has never been run at the live endpoint:** minting a
real token, pressing the button, and seeing `transfer_confirmed_at` stamped. Everything around it
is proven (above), and the redemption SQL itself was verified against real Postgres in a
rolled-back transaction on 2026-08-20 — but the endpoint's success path has not been exercised in
production.

**Why it is not done:** it mutates a real `sites` row. Blast radius today is genuinely small
(`handed_over_at` = 0 fleet-wide, so nothing is scheduled against that stamp) and the stamp can be
cleared in seconds, but it is a production write on a customer-facing record.
**The owner was offered it and has not yet chosen a site.** Do not pick one unilaterally.

Recipe when authorised: mint one token via `delivery.MintToken`-shaped INSERT with
`created_by='verify-second-click-<date>'`; `GET` twice and assert `use_count`/`used_at`/
`transfer_confirmed_at` unchanged; `POST` once and assert all three move; `POST` again and assert
the spent-link page; then delete the token row and clear the stamp. ⚠ **Say so in NOTES BEFORE
doing it** — a non-zero `customer_access_tokens` is a documented falsifier meaning "a delivery
email may be about to go out", and a later session will read its own alarm.

## 3. Closed since the last handoff, so nobody re-opens them

- **Second-click page** — built, council `ea99befa` APPROVED r1 (all seats), shipped, proven.
- **RFC_054, all three questions** — RULED by the owner 2026-08-25. Q1 = the two-door pattern is
  standing (register **SYS-094**). Q2 = build the delivery-only listener — **built, shipped and
  live** (**SYS-095**). Q3 = SYS-094 plus a header line in every `box/*.nginx` and
  `box/*.cloudflared-ingress.yml`; **verified present in all five box files**. The RFC also stands
  as the recorded boundary review for `/stripe/webhook`, so unpark day owes no further review.
- **`WriteSiteSpecAction` deep-merge follow-up — REFUTED, do NOT build the guard.** No agent
  writes `evidence_base` through that action (0 of 20 live steps, 0 across snapshots, 0 templated
  — control: 157 templated steps elsewhere); the scheduled refresher shrank a register **0 times
  in 222 writes**; the one emptying in all history lasted **59 seconds** and was one session's own
  two-part write. Residual, different from the item: 8 of 19 sites have no refresher coverage, and
  the only door that has ever emptied a register is hand-written SQL. Full census in NOTES.
- **Builds screen owner feedback** — orchestration list truncated, running work pinned, "Nothing
  running…" line added, misleading `(50)` heading fixed.

## 4. ⚠ ANOTHER SESSION IS IN THIS SCREEN — read before editing

`48e75aad2` and `f24743358` (overnight, both LIVE in `2fb40a960`) changed the terminate endpoint,
the workflow detail query and the SPA's confirm copy. **Their finding lands on our code and is
the single most useful fact about this screen:**

> **`correlation_id` is NOT unique.** One correlation is a TREE of orchestration rows.
> `[MEASURED 2026-08-26]` 6,031 site-tagged rows; one site's single correlation is shared by **27**
> rows (then 25, 25, 22, 21). `/admin/workflows` returns them raw, with **no unique id in the
> payload**.

Consequences already handled — do not redo:
- Their fix: terminate now carries `AND status NOT IN (COMPLETED, FAILED)` (it was relabelling
  finished siblings as FAILED), and the detail SELECT pins root-first oldest-first LIMIT 1 (it was
  answering with an arbitrary sibling, so the step-error panel appeared and disappeared between two
  clicks on the same row).
- **Our fix, `c016b3fb4`:** the list groups by correlation, one row per correlation with `×N` in
  the screen's existing idiom. Before it, React was keyed on a repeating value on a list that
  re-polls every 10s. **Inert until the next dashboard image.**

**So: `git log -3 -- frontends/admin-dashboard/src/App.tsx` before you edit it, every time.**

## 5. Falsifiers — re-run these before trusting anything above

- **`customer_access_tokens` non-zero** expires every "nothing at risk" line here. If it goes
  non-zero and §2 has not been done deliberately, raise it with the owner at once.
- Tags roll daily and this handoff names `v1.0.1341` / `2fb40a960`. **Re-ask each pod its own
  provenance stamp, per service.** An empty `grep 'build provenance'` means "scrolled out of
  range", not "unstamped".
- The dashboard bundle hash changes on every rebuild — re-grep the served bundle, never cite a hash.
- **`committed` and `applied` are independent facts for CONFIG, and config has no image tag to give
  it away.** This bit us on 2026-08-25: `d30917150` added `port: 8090` to
  `networkpolicy-wireguard-egress.yaml` and **nobody applied it**; the live policy allowed `8088`
  only, so the repo read correct while the cluster refused the very port the new vhost proxies to.
  Applied 2026-08-25 with the owner's go-ahead. **Read the live object and prove reachability from
  the pod that must do the reaching, with a must-fail control:**
  ```bash
  WG=$(kubectl -n ai-persona-system get pods -l app=wireguard -o jsonpath='{.items[0].metadata.name}')
  kubectl -n ai-persona-system exec $WG -- sh -c '
    for p in 8090 8088 9999; do timeout 4 nc -z core-manager.ai-persona-system.svc.cluster.local $p \
      && echo "$p OPEN" || echo "$p blocked"; done'
  ```
  Want: 8090 OPEN, 8088 OPEN (proves the probe works), 9999 blocked (proves the fence still
  discriminates). Also re-check postgres is still blocked from that pod — that is what the fence is for.
- **When every arm of a probe returns the same status, the status is not the instrument.** All
  three external probes returned 404 while the box was mid-change; only the response BODY
  distinguished "still pointing at the old port" (gin's `text/plain 404 page not found`) from
  "died at the box" (nginx's HTML page).
- kubectl token expires on the 3-day cycle (last refreshed ~2026-08-24 21:00, so due ~2026-08-27).
- A newer handoff here or in `../webdesign_uk_build_service/`.

## 6. Where the missteps are written down

Two from this lane, both in `WRONG_CALLS.md` (2026-08-25) because the checks they produce are
worth more than the findings: a test for an ABSENCE that could not fail (the mutation that should
have caught it PASSED, and the file's own "assert the effect, never the absence of a call" rule is
what produced the vacuous test), and declaring my own LANDMINES entry clobbered on three checks
that were one impossible grep pattern (the heading had backticks; the entry was in HEAD the whole
time, and the guard in my re-add script inherited the same bug and let me duplicate it).
New landmine added: `POST /c/<token>` must live on the SAME path as the GET.
