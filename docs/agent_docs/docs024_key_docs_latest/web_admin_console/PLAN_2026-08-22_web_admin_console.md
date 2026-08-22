# PLAN — a web admin on a domain of ours, that the owner logs into

**Status: INITIAL PLAN, deliberately unfinished.** Owner asked (2026-08-22) for a starting
point to beef up in another thread.

**Owner's words:** *"I would like a web based admin on a domain of mine that I log into."*
Earlier the same day, the purpose: *"for me to follow and contribute to the steps of each
website build."*

---

## 0. The headline: about 80% of this is already built and running

This is not a build-from-scratch job. Measured today:

| piece | state | evidence |
|---|---|---|
| The admin web app | **built, running, 2 replicas, 158d old** | `admin-dashboard` ClusterIP `10.21.171.225:8080`; `GET /` → `200`, `/health` → `{"status":"healthy","service":"api-gateway"}` |
| Its API gateway | **built** — one pod serves the SPA *and* proxies `/api/v1/auth/` → auth-service, `/api/v1/` → core-manager | `frontends/admin-dashboard/nginx.conf:51,64,77` |
| Login | **works** — email/password, JWT, `role must == "admin"` | `App.tsx:257` `LoginScreen`; probe with bad creds → **401**, not a 5xx |
| Admin API surface | **~76 routes** — sites, specs (pin/propagate), pages, components (regenerate, restore-section, lock), assets, work items (retry/resolve/**approve**), customers, pipelines, **workflows** | `internal/core-manager/api/server.go:164-266` |
| Reachability | **ClusterIP only. No Ingress objects exist anywhere in the cluster.** | `kubectl get ingress -A` → "No resources found" |

**So the two real gaps are: (1) it is not reachable from a browser on the internet, and
(2) the one screen you actually asked for does not exist.**

### Gap 2 is small and specific

The build steps are already served by the API — `GET /admin/workflows`,
`GET /admin/workflows/:correlation_id`, `POST /admin/workflows/:correlation_id/resume` —
but **the SPA never calls them**: `grep -c "workflow" frontends/admin-dashboard/src/App.tsx`
→ **0**. So "follow the steps of each build" is a **frontend screen against an API that
already exists**. "Contribute" (edit a spec, regenerate a component, resolve a work item)
already has screens.

---

## 1. The decision this plan turns on: how it becomes reachable

There is a live, unanswered owner decision that this collides with — **D-A** in
`webdesign_uk_build_service/HANDOFF_2026-08-21_continue_here.md` §2: *may core-manager be
reachable from the internet at all?* The council gated a submission on exactly this, and
`sitefacts.go` in that same service documents *"NO PUBLIC EXPOSURE … ClusterIP only, no
Ingress"*.

**Putting the admin console on a public domain is a strictly larger version of that
decision**, because the admin API can rewrite every site's content, not just read a token.

### The three options, costed

| | shape | what it costs | what it risks |
|---|---|---|---|
| **A. VPN only (no public domain)** | keep it ClusterIP; reach it over the WireGuard that is **already running**, with `laptop`/`phone` peers **already generated** | **near zero — it works today** | Not "a domain of mine". No access from a device without the VPN. Endpoint pins one node's IP |
| **B. Public domain, tunnel + strong auth** | Cloudflare Tunnel → a small edge → admin-dashboard. **We have run exactly this for a month**: `tools.apis.uk` → cloudflared → Caddy path-allowlist → service, on a Mythic Beasts VM | moderate; the pattern is proven and the runbook exists | The admin API becomes internet-reachable. Needs a second auth factor in front, not just the app's own login |
| **C. Public domain, VPN-gated** | public DNS name, but the origin only answers over the tunnel — e.g. Cloudflare Access in front, or the edge only reachable from the VPN | moderate | Best of both, most moving parts |

**Recommendation: A now (today, no work), B or C as the real answer** — and B's groundwork is
already done and proven in production, which is the argument for it. See
`gauntlet_dead_cta/infra/island/RUNBOOK_island.md`.

⚠ **The `tools.apis.uk` precedent is a genuine one and I got it wrong once today.** I
initially reported that path as an unbuilt design; it has been live since 2026-07-24. It is
the estate's existing, working answer to "expose an HTTP surface without exposing the
cluster". Full correction: `WRONG_CALLS.md`, 2026-08-22.

---

## 2. Things that will bite, found today

- **Auth users are NOT in this cluster.** `auth-service` connects at startup to an **external
  MySQL** — `catalogu_vectordb_chassis:3306` — for identity, and to `clients_db` via pgbouncer
  for everything else (`kubectl logs -l app=auth-service`, `database/mysql.go:43`). So the
  login path has an off-cluster dependency, and any plan that says "we control the whole auth
  chain" is wrong until that is faced. `[UNVERIFIED]` whether an admin account exists for the
  owner — the endpoint 401s correctly on bad credentials, which proves the chain works, not
  that his account does.
- **`AdminOnly()` is a bare role-string check.** `role != "admin"` → 403
  (`internal/core-manager/middleware/auth.go:181`). There is no second factor, no IP
  restriction and no per-action scoping. Fine behind a VPN; **thin as the only thing between
  the internet and every site's content.** If the answer is B, something must sit in front.
- **The register lied about the approve endpoint and I corrected it today.** `ADM-004` said
  approve/reject were superseded with *"no approve/reject endpoints … anywhere"*.
  `HandleApproveWorkItem` exists and is routed; `HandleRejectWorkItem` genuinely does not.
  This matters because `DECISION_2026-08-21e`'s pre-delivery review gate needs exactly an
  approve step, and a session trusting the register would have built a second one.
- **`ADM-002` says the admin API has known bugs and mock data** — blocks A–F, including
  MySQL-syntax `CURDATE()` in the dashboard query and hardcoded values. Status *partial*,
  and the entry is from the 2026-07-13 extraction freeze, so it is **stale in an unknown
  direction**. Verify before promising a working screen.
- **The VPN endpoint pins one node.** `SERVERURL=134.213.168.37` is worker
  `prod-instance-…1148`. On **Rackspace Spot**, a reclaimed node breaks the tunnel with no
  warning. A LoadBalancer or a DNS name would fix it; today it is a single point of failure.

---

## 3. Suggested phasing

1. **Today, zero build:** connect over the existing VPN and use the console (see the
   walkthrough in `RUNBOOK_web_admin_console.md`). This tells us whether the *app* is good
   enough before we spend anything on *exposure*.
2. **Build the missing screen:** a build-steps view against `/admin/workflows/:correlation_id`,
   with the resume action. This is the actual ask and it is frontend-only.
3. **Answer D-A** — it is the same decision, and doing the console first without answering it
   is how the decision gets made by accident.
4. **If public: copy the island pattern.** Cloudflared tunnel + path allowlist + a real second
   factor in front of `AdminOnly()`. Pick the domain.
5. **Fix the endpoint fragility** (§2, last bullet) before depending on the VPN for anything
   time-critical.
6. **Re-verify `ADM-002`'s bug list** against live code, since it predates the freeze.

---

## 4. Open decisions

- **Which domain?** Not `webdesign.uk` — that is the customer shopfront, and its apex already
  302s to `webdesign.co.uk` under a Cloudflare page rule. A separate admin hostname keeps the
  blast radius and the DNS story clean.
- **A, B or C** from §1.
- **Second factor:** Cloudflare Access, a hardware key, or an allowlist? Only needed for B/C.
- **Does the console serve one cluster or two?** If the customer-build satellite happens
  (companion plan `customer_build_satellite/PLAN_2026-08-22_…`), this console has to reach
  both, or there are two consoles. **Decide this before building the screen**, because it
  changes whether the API base URL is a constant or a selector.

## 5. Falsifiers

- `ADM-002`'s bug/mock list predates the 2026-07-13 freeze — verify, do not repeat.
- The `[UNVERIFIED]` admin account in §2.
- Image tags roll several times a day; `admin-dashboard`'s "158d old" is the *service*, not
  the image.
- D-A may have been answered in another thread since this was written.
