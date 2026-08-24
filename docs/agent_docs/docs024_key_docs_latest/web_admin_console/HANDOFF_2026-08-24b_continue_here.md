# HANDOFF 2026-08-24b — Builds screen built; deploy ordering is the live constraint

Supersedes `HANDOFF_2026-08-24_continue_here.md` for STATE; that file's §1 table (edge
infrastructure proofs) and §2/§3 (owner box steps) remain the reference — read it after this.

## 0. What changed since the morning handoff

1. **The build-steps screen (§4 there) is BUILT and COMMITTED, not live.**
   - Backend `e6350e74b` (`internal/core-manager/admin/`): `GET /admin/workflows` gains
     `site_id` filter (indexed column, PLAN §6a) + per-row `has_step_error`
     (`collected_data ? '__step_error'`, exact per §6b) + `site_id`; terminate's
     `orchestrator_state` → `orchestration_states` (ADM-002 B2 — register entry updated);
     `HandleUpdateSiteSpec` evidence_base guard (PLAN §6h hazard: parse-before-write, 409
     `EMPTY_EVIDENCE_BASE` unless `confirm_empty`, stored counts + `superseded` returned).
     Six sqlmock tests. `Council-Submitted: 45b3c93f-7937-474d-8234-31c39bab033b` —
     **verdict unread as of this writing; read it and act on a REVISE** (find the run by
     payload: `collected_data->'input_data'->>'fix_correlation_id' = '45b3c93f-…'`; a chassis
     roll was in flight around submission time, so if NO orchestration row exists after the
     queue latency window, the dispatch may have hit the ~300s post-restart drop — that is
     the one case where resubmission with the same file + `RESUBMIT_CORR` is warranted).
   - Frontend `b3fbfdd02` (`frontends/admin-dashboard/src/App.tsx`): `BuildsView` — per-site
     stage timeline over `site_work_items` in the 082 cascade order with durations (§6d
     shape), other-activity rollup, orchestration drill-down surfacing `__step_error` in red
     whatever `status` says, resume/terminate behind confirms; §6g `⚠ overwritten ×N`
     badges (red when unlocked); §6h ENFORCED/advisory chips, counts echoed on
     evidence_base saves, 409→confirm→`confirm_empty` flow, prohibition nudge. Vite build
     proven via the Dockerfile builder stage.

2. **⚠ DEPLOY ORDERING — the one live constraint.** New SPA against old core-manager is the
   misleading combo (gin ignores the unknown `site_id` param → BuildsView shows the WHOLE
   FLEET's workflows labelled as the site's; no `has_step_error`; `confirm_empty` ignored).
   So: **(1)** wait until core-manager runs a build carrying `e6350e74b` —
   `kubectl -n ai-persona-system logs -l app=core-manager --tail=300 | grep -m1 'build provenance'`
   then `git merge-base --is-ancestor e6350e74b <stamp>`; **(2)** then
   `make admin-dashboard` (build+push+deploy; it builds the WORKING TREE — ensure App.tsx is
   at/after `b3fbfdd02` and bump `IMAGE_TAG`). Old SPA + new backend is harmless.

3. **Council verdict on the `/c/` prefetch guard (`6b1726ab…`): APPROVED round 1**
   (2026-08-23 12:07Z, one advisory objection, none high-severity). Morning handoff §5.2
   CLOSED. `0e9cb31ee`'s `Council-Submitted:` trailer is credited automatically by 098.

4. **LANDMINES gained the evidence_base wrong-shape silent-off entry** (footprint
   `site_specs`/`ParseEvidenceBase`/SQL seeds; the check is read-back-the-counts). Synced +
   verifier dispatched. The FILE is uncommitted — it carries the 333 lane's WIP entries too;
   whoever commits takes both, append-only, say so in the message.

5. **Falsifiers re-run this session:** `links.webdesign.uk` still does not resolve (morning
   §2 still owner-pending — the `/c/` move remains the top owner action);
   `customer_access_tokens` = 0; `www.apis.uk` 301s to apex **via the portfolio-sites-router
   Worker — NOT evidence the §3 Redirect Rule was applied** (NOTES 2026-08-24 correction;
   do not re-close that falsifier off a curl).

## 1. Open / owed (delta against the morning handoff's §5)

1. ~~**Read the `45b3c93f` council verdict**~~ **READ — APPROVED round 1** (16:35Z, 3
   advisory objections, none high-severity; all three re-checked against code/DB and
   answered in NOTES 2026-08-24 ~16:35Z — raw-bytes question moot at
   `site_admin_handlers.go:282`, guardian's terminate caveat shipped in `1a8db99f9`).
   **One real follow-up spun out, its own council round when taken:** `WriteSiteSpecAction`'s
   deep-merge lets a partial with `"banned_claims": []` empty the array (arrays overwrite
   wholesale, `site_spec_actions.go:554`), and `source='scheduled'` is the highest-volume
   evidence_base writer (214 of 319 rows all-history, counted 2026-08-24) — census the
   legitimate-shrink history BEFORE designing any guard there.
2. **Deploy pair** per §0.2 when the roll lands — then eyeball `admin.apis.uk` → a site →
   Builds against the apis.uk chain (PLAN §6d table is the expected shape, ~67 min build),
   and smoke-test terminate against a real non-terminal correlation (expect 200 + FAILED,
   not 500 — debug_historian's owed verification; sqlmock proved the shape, not the table).
3. Architecture boundary review when `links.webdesign.uk` goes live (morning §5.1) — unmet
   until the owner applies §2.
4. Morning §5.4–§5.7 unchanged (HOLD ban, webdesign lane items, VPN parked, ADM-002
   staleness — B2 now fixed-committed, B3/B4 still `[UNVERIFIED]`). ~~§5.3 mail-scanner
   residual: owner's call~~ **RULED — see §3.2 below: second click required, delivery
   email blocked on it.**

## 2. Falsifiers for THIS handoff

- Whether core-manager's stamp carries `e6350e74b` (the roll cadence is daily; the owner's
  fresh-chassis roll of 2026-08-24 afternoon may or may not have included core-manager —
  ask the pod, per service, never the fleet).
- The `45b3c93f` verdict may have landed (or the dispatch may have been dropped — §0.1).
- The dashboard image may already have been rebuilt by another session; check the running
  SPA for a "Builds" button before rebuilding.
- `customer_access_tokens` and handed_over counts — 0 as of 2026-08-24; any non-zero expires
  the "nothing at risk" claims inherited from the morning handoff.

---

## 3. EVENING ADDITIONS (2026-08-24, same session, after owner questions) — start HERE

1. **Deploy check DONE: core-manager does NOT carry `e6350e74b`.** Both pods (checked
   individually) stamp `70fd163c2`, built 15:37Z — an ancestor check fails. So: backend
   fixes committed+approved but NOT live; the dashboard image deploy stays blocked on the
   next core-manager roll (§0.2 ordering unchanged and still the first action when the
   roll lands).

2. **OWNER RULING (2026-08-24): the mail-scanner residual is NOT accepted — a second
   click is required.** *"We can't have email scanners clicking the accept button so
   we'll need a separate page."* Recorded with the full mechanics and build sketch in
   `../webdesign_uk_build_service/DECISION_2026-08-24_confirmation_needs_a_second_click.md`;
   the `links.webdesign.uk.nginx` header comment updated (its "owner's open call" line was
   stale the moment this was said). Consequences: `GET /c/<token>` must become
   render-only, the confirm moves to a POST from the page's button, and **no delivery
   email may be sent before that page is live** — this is now a hard dependency on the
   webdesign lane's delivery-email item. New build item: `handlers/delivery.go` split +
   POST route + council round. The owner's §2 box steps (links vhost + DNS) are NOT
   blocked by this — the vhost passes GET and POST alike, and exposing the hostname early
   is safe (`customer_access_tokens` = 0, prefetch guard live).

3. **Route census — the "second public cluster route" claim, measured properly**
   (owner asked whether noted.co.uk / robot-hands.com / idea.uk etc. are also cluster
   routes; they are not):
   - **Portfolio site domains (noted.co.uk, idea.uk, robot-hands.com — live and 200ing,
     apis.uk, the ~39 zones) never touch the cluster at serve time.** They are
     Cloudflare-fronted static sites — served from the B2 `portfolio-sites` bucket via
     the portfolio-sites-router Worker, or from the git-hosted route — the cluster
     BUILDS and uploads them; no visitor request reaches a cluster service.
   - **Paths that DO reach the cluster** (via box nginx → WireGuard): `admin.apis.uk` →
     admin-dashboard (Access-gated, deliberate, live); `webdesign.uk/c/` → core-manager
     and `webdesign.uk/stripe/webhook` → auth-service — **both currently swallowed by the
     parking 302 to webdesign.co.uk (measured: both return 302)**, i.e. TODAY zero
     ungated public cluster paths are actually reachable.
   - After the §2 move: `links.webdesign.uk/c/` becomes the FIRST deliberately public
     UNGATED cluster path (admin.apis.uk being the gated one). The stripe webhook becomes
     the next when the shopfront unparks — **⚠ flag for the webdesign lane: a Stripe
     webhook configured today would have its events 302-bounced by the parking redirect
     (non-2xx = failed delivery + retries); the webhook cannot go live before the parking
     rule excludes that path or the shopfront unparks.**
   - The architecture-seat boundary review (§1.3 / morning §5.1) stays owed when the
     links host goes live; this census is input to it.

4. **Foot-gun list for the links move, with consequences** (the security-relevant ones
   first; all are inline in `box/links.webdesign.uk.nginx` and the runbook, gathered here
   because the owner asked):
   - **Widening `location /c/`** (or removing the `location / { return 404; }`
     catch-all): whatever prefix is added becomes a public unauthenticated door into
     core-manager — e.g. `/api/v1/site-facts/:domain` would expose every site's evidence
     facts; the admin surface would face the internet with only its own JWT check.
     SECURITY. The prefix IS the exposure.
   - **Rate-limit key is `CF-Connecting-IP`** — trustworthy only while traffic can ONLY
     arrive via the tunnel. Any path that exposes the box directly makes the key
     attacker-chosen. SECURITY (posture-dependent). Keep loopback-bind + ufw deny.
   - **`cloudflared tunnel route dns`** puts the record in the WRONG ZONE and reports
     success (measured 2026-08-23, LANDMINES): links stays dead while believed live, plus
     a stray record in another zone. AVAILABILITY + confusion. Dashboard only.
   - **Ingress rule below the 404 catch-all**: cloudflared takes first match → every
     customer link 404s. AVAILABILITY (fail-closed).
   - **Static `proxy_pass` resolves at nginx start**: a changed core-manager ClusterIP →
     timeouts that read as core-manager down, until `systemctl reload nginx` on the box.
     AVAILABILITY.
   - **WireGuard egress fence**: core-manager:8088 must stay on
     `networkpolicy-wireguard-egress.yaml`'s allowlist; narrowing it produces the same
     "core-manager looks down" timeout. AVAILABILITY.
   - (apis.uk zone, if touched for www: do NOT delete the `*.apis.uk` wildcard — it
     feeds the island probe. The www 301 is served by the Worker regardless.)

5. **kubectl token EXPIRED mid-session (~16:50Z)** — fleet-wide `Unauthorized`, the known
   3-day expiry; owner refreshes. Every DB/pod check above predates the expiry; anything
   needing the cluster from here waits on the refresh.
