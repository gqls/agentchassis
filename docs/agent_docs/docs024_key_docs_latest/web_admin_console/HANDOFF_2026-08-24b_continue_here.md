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

6. **Link expiry + more protection between the links host and the cluster (owner asked
   2026-08-24 evening; answers measured at the code, `platform/delivery/handover.go` +
   `handlers/delivery.go`):**
   - **Expiry is ALREADY BUILT IN** — every token row carries `expires_at`, enforced
     atomically inside `RedeemToken`'s single UPDATE (`AND expires_at > now`), alongside
     single-use (race-safe: one statement, `used_at IS NULL` predicate), sha256-at-rest
     (a DB leak yields no working links), 256-bit random plaintext, and a 128-char
     length bound refused before the DB is touched. **The confirm-link TTL is UNPINNED**:
     nothing mints confirm tokens yet (the delivery-email builder is unbuilt), so the
     number is a free owner dial. Recommendation put to the owner: short (7–14 days) —
     cheap because there is deliberately no resend path, and each weekly chase email can
     mint a fresh link; pin it as a named constant beside `LiveLinkWindow` (six weeks,
     which governs the ZIP link by the 2026-08-20 ruling) when the email builder is built.
   - **"Disguise" is largely designed in already**: every failure (unknown/expired/
     revoked/spent/wrong-purpose) is ONE page at HTTP 200 — no oracle, nothing for an
     enumerator to learn, and probing tells you nothing distinguishable. Added at the
     vhost 2026-08-24: `X-Robots-Tag: noindex,nofollow` on every response (the hostname
     exists only inside emails), `server_tokens off`.
   - **Vhost hardened** (committed; the owner applies the CURRENT file — re-copy if an
     older copy was already staged on the box): the `/c/` location is now an exact
     token-shape regex (`^/c/[A-Za-z0-9_-]{20,128}$`) so traversal/encoding/junk paths
     die at the box and never cross WireGuard; methods beyond GET/HEAD/POST refused at
     the box; `client_max_body_size 4k`. **Consequence pinned in the decision doc: the
     second-click POST must be `POST /c/<token>` (same path) or the box 404s it.**
   - **Edge layer (owner, Cloudflare dashboard, optional ~5 min):** a rate-limiting rule
     for `links.webdesign.uk` (e.g. >30 req/min per IP → block) keeps floods off the box
     entirely; a Configuration Rule raising Security Level for that hostname is the
     lighter variant. Not applied — dashboard is owner-side.
   - **Structural candidate — feed it to the owed boundary review (§1.3):** a dedicated
     delivery-only LISTENER in core-manager (e.g. `:8090` serving only `/c/`+`/d/`),
     with the WireGuard egress fence allowlisting 8090 instead of 8088. Then even a
     widened nginx location could reach nothing but delivery routes — containment that
     survives a box misconfiguration. Moderate effort, its own council round; NOT built.
   - **Layers that already exist and count**, so the review doesn't re-invent them:
     tunnel-only box (no inbound ports, ufw deny, loopback binds), WireGuard egress
     fence (peers reach exactly kube-dns/core-manager:8088/auth-service:8081/
     admin-dashboard:8080 — postgres proven blocked, with control), zero Ingress
     objects in the cluster, the prefetch guard, and the second-click page once built.

7. **Cloudflare rate-limit walkthrough for `links.webdesign.uk`** (owner asked 2026-08-24;
   dashboard is owner-side so recorded here, not applied). The rule lives in the
   **webdesign.uk ZONE** (links is a hostname inside it); it can be created before the
   DNS record exists and simply starts matching when traffic flows.
   - Dashboard → webdesign.uk zone → **Security → WAF → Rate limiting rules → Create**.
   - Name `links-host-limit`. ~~Custom filter expression: field **Hostname** · operator
     **equals** · value **links.webdesign.uk**.~~ **CORRECTED 2026-08-24 at the live
     form: the free-plan rule builder's Field dropdown has NO Hostname entry** (it
     offers URI Path / bot fields) — click **Edit expression** instead and type
     `(http.host eq "links.webdesign.uk")`, which is the same match in expression
     form. Deliberately the WHOLE host, not `/c/` — the 404 catch-all paths are
     exactly what probes hammer. (Fallback if a plan gates `http.host`: URI Path
     starts_with `/c/`, coarser but adequate.)
   - "With the same characteristics": **IP** (the only choice on the free plan).
   - "When rate exceeds": **10 requests per 10 seconds** → action **Block** (free plan
     fixes the period at 10s and a short block timeout; that is fine — a continuing
     flood re-trips the counter, and a customer clicks once, so 10-in-10s is far above
     any legitimate use, incl. a mail scanner fetching a link once). Paid plans can
     lengthen the block; not required.
   - Deploy. **Verify once DNS is live**: `for i in $(seq 1 40); do curl -s -o /dev/null
     -w "%{http_code}\n" https://links.webdesign.uk/c/x; done` — expect 404s giving way
     to **429** once the threshold trips, then recovery after the timeout. A single
     manual click in a browser must still work.
   - Where it sits: BEFORE the tunnel, so a flood never reaches the box; the vhost's own
     20 req/min per-IP limit stays as the finer layer behind it. The rule cannot touch
     webdesign.uk itself (Hostname filter) and is unrelated to the parking redirect.
   - Free plan carries ONE rate-limiting rule per zone — if it is ever spent on
     something else, the fallback is a WAF custom rule (managed challenge on that
     hostname), which is weaker but free-tier-unlimited.

8. **`SUMMARY_2026-08-24_web_admin_console.md` written** — the lane's first milestone
   read-out (Builds screen built+approved, exposure posture settled). Series rule: next
   summary is a NEW file, only at the next real inflection.

9. **Verify expectations for the links host — CORRECTED after the hardening pass**
   (2026-08-24 ~21:00; the owner ran the 429 loop and got 40× `000` — measured cause:
   `links.webdesign.uk` is NXDOMAIN at 1.1.1.1 while the zone control resolves, i.e. the
   CNAME is not created yet; `000` is DNS failure, not a rate-limit or tunnel result).
   The morning handoff §2 verify predates the hardened vhost and is now wrong in one
   place: **`curl https://links.webdesign.uk/c/x` will 404 at the BOX** — `x` fails the
   token-shape regex (min 20 chars), which is the hardening working, not a fault. The
   corrected sequence, run only AFTER the box steps AND the CNAME exist:
   - `curl -s -o /dev/null -w "%{http_code}\n" https://links.webdesign.uk/other` → **404** (box catch-all)
   - `curl -s -o /dev/null -w "%{http_code}\n" https://links.webdesign.uk/c/x` → **404** (regex refuses non-token shapes)
   - `curl -s -o /dev/null -w "%{http_code}\n" "https://links.webdesign.uk/c/$(printf 'a%.0s' {1..43})"`
     → **200** (token-SHAPED, crosses WireGuard, core-manager serves the uniform
     "no longer active" page — this is the one that proves the full path)
   - the 40× loop on any of the above → **429**s partway through (the edge rate limit
     counts every request on the hostname, 404s included)
   Order still: box nginx files (CURRENT committed copies) → cloudflared ingress above
   the catch-all → CNAME `links` → `81f59f78-dda8-40a0-984b-cfadb36bc891.cfargotunnel.com`,
   Proxied, in the DASHBOARD.

10. **`links.webdesign.uk` IS LIVE — box steps + DNS applied by the owner, VERIFIED from
    outside (2026-08-24 ~21:20, this session's own curls, not the terminal transcript):**
    `/other` → 404 · `/c/x` → 404 (token-shape regex) · 43-char token-shaped → **200**
    (core-manager's page — the full tunnel→nginx→WireGuard→cluster path works) ·
    `admin.apis.uk` → 302 to Access (healthy after the cloudflared restart) · apex
    `/c/x` → 302 parked (unchanged) · 40-loop → **8× 404 then 32× 429** (the edge rate
    limit fires at ~the 10-in-10s threshold). Morning handoff **§2 is CLOSED**.
    Consequences now live:
    - **The architecture boundary-review condition (§1.3) is MET** — run one council
      round over the exposure posture. Inputs ready: §3.3 census + §3.6 delivery-only
      listener candidate. ⚠ Blocked on the kubectl token refresh (expired ~16:50Z) —
      the 097 trigger needs the cluster.
    - `links.webdesign.uk` is the canonical emailed-links host from this moment; the
      delivery-email builder mints there — still gated on the second-click page (§3.2).
    - Lane entry added to MEMORY_workstreams (`web-admin-console-workstream.md`).
