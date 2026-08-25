# PLAN 2026-08-25 — the delivery-only listener (RFC_054 Q2)

**Status: IN BUILD.** Owner ruled **BUILD** on 2026-08-25 (`architecture_review/RFC_054_public_cluster_exposure_boundary_review.md` §5 Q2): *"BUILD the delivery-only listener. Its own plan + council round; aim to land it before the first `customer_access_tokens` row exists."* This is that plan.

**Lane:** site_delivery_and_editor (Phase 4 infrastructure). Joint lane with webdesign.uk build service; the box vhost this touches lives in that lane's directory and is applied by the owner's box steps.

---

## 1. What is wrong today, stated plainly

`links.webdesign.uk` is a public hostname. Its nginx proxies customer clicks into the cluster, to **`core-manager:8088`**. That is the same port that serves the whole admin API — every site's data, the work-item queue, the pipeline controls, the agent-definition writes.

The only thing standing between the internet and that API is a **regex in a config file on a box**:

```nginx
location ~ "^/c/[A-Za-z0-9_-]{20,128}$" { ... proxy_pass http://core-manager...:8088; }
```

Widen that location to `/` — a one-character mistake, the exact class `LANDMINES.md` keeps collecting — and the admin API is on the internet. Nothing in the binary would notice or refuse.

CLAUDE.md's own 2026-08-02 §2 ruling is the argument, one level along: *"a comment is not a control on a tree this many sessions share."* Neither is a box's nginx config, for the service holding every site's data.

## 2. What the change is

A **second HTTP listener inside core-manager**, on its own port, whose router registers **only** the delivery routes. The box and the egress fence point at that port instead. The delivery routes are **removed** from the main router — moved, not copied, because a copy leaves the hole open.

After this, widening the box location to `/` exposes `/c/` and (later) `/d/` and nothing else. The blast radius is capped **in the binary**, which is what the RFC asked.

### The pieces

| # | File | Change |
|---|---|---|
| 1 | `platform/config/loader.go` | `ServerConfig` gains `DeliveryPort string` (`mapstructure:"delivery_port"`). **No default — empty means OFF.** |
| 2 | `internal/core-manager/api/server.go` | second `*http.Server` + its own `gin.Engine` registering only `deliveryHandler.RegisterRoutes`; delivery routes **removed** from the main router; startup assertion (§4) |
| 3 | `cmd/core-manager/main.go` | start/shutdown the second listener alongside the first |
| 4 | `deployments/.../core-manager/base/service.yaml` + `deployment.yaml` | expose port `8090`, named `delivery` |
| 5 | `deployments/.../core-manager/overlays/production/uk_001/patch-env.yaml` | `SERVICE_SERVER_DELIVERY_PORT: "8090"` — **this is the opt-in** |
| 6 | `deployments/.../networkpolicy-wireguard-egress.yaml` | add `core-manager:8090` to the allowlist |
| 7 | `webdesign_uk_build_service/box/links.webdesign.uk.nginx` | `proxy_pass` → `:8090`, header note updated |

### Opt-in, unsafe default OFF

Per CLAUDE.md 2026-08-02 §2 and RFC_022: the new authority here is *a second listener accepting public traffic*, so the unsafe side is ON and the default is **OFF**. An unset `delivery_port` starts no second listener and mounts the delivery routes nowhere — the door is shut, not silently re-opened on 8088. Production opts in explicitly, in the overlay, where a reviewer of the deployment can see it.

**Consumers enumerated** (RFC_022's third condition — asserting this without the query is itself the objection): the delivery routes have exactly one live consumer, the `links.webdesign.uk` vhost. `agent_definitions` naming `/c/` = 0 and current `site_specs` naming `links.webdesign.uk` = 0 were measured by the web_admin_console lane on 2026-08-25 (control 21); re-measure before submitting.

## 3. The residual, stated rather than hidden

**Port 8088 stays on the egress fence, so the fence is not the control.** The chat bot's facts relay (`GET /api/v1/site-facts/:domain`, `box/chat-service/facts.go`) dials `core-manager:8088` over the same WireGuard leg, and the fence's own comment records that breaking it *"stops the bot STARTING, by its own design."*

So what this change buys is exact, and no more: **the links door cannot be widened into the admin API.** A *new* box vhost written straight to `:8088` still could.

Making the **fence** the control needs the facts relay moved onto a box-facing listener too, and 8088 dropped from the allowlist. That is a bigger change touching a live customer-facing bot, and it is outside the scope the owner ruled. **Recorded here as the follow-on question, not silently skipped.**

## 4. The one control that is a mechanism, not a comment

A future session mounting `/c/` back on the main router would silently undo the whole change, and the suite would stay green. So the invariant is enforced at startup:

`assertNoDeliveryRoutes(routes)` walks the main router's route table and **refuses to construct the server** if any path is under `/c/` or `/d/`.

Fail-closed is deliberate. The only way to trip it is to deliberately mount a customer route on the admin port, and the first roll catches it loudly rather than serving a quietly-reopened hole. **Flagged for the council to challenge** — log-and-continue is the alternative, and nobody reads it.

## 5. Ordering, and the window it opens

The routes **move** in one commit. Between the core-manager roll and the box repoint, `/c/` returns 404 at the box.

**Measured zero impact, 2026-08-25:** `customer_access_tokens` = **0**; 51 sites, `handed_over_at` **0**, `transfer_confirmed_at` **0**, `live_link_expires_at` **0**. No delivery email has ever been sent, and the webdesign lane's own ruling blocks the first one until the second-click page is live — which is waiting for the **same roll** (`24b63120d`/`d1a4bdcdf`, committed 2026-08-25, not in the running `4c996e1b5`).

So the honest sequence is: commit → roll → verify from outside → box repoint → verify from outside again.

## 6. How it will be verified — from OUTSIDE, because inside proves nothing

⚠ `LANDMINES.md` (footprint `links.webdesign.uk.nginx` · `location ~`): the box's anchored regex means **a request that works perfectly from inside the cluster can be dead on the only path a customer takes**, with no log line, no metric and no error anywhere in the cluster. A cluster-internal `curl` is not evidence here.

| # | check | expected |
|---|---|---|
| 1 | ancestry: pod's `build provenance` stamp vs this commit, **with the reversed control** | in / not-in |
| 2 | `GET https://links.webdesign.uk/c/<43-char junk>` from outside | 200, the button page |
| 3 | `POST` same path from outside | the "no longer active" page (token is junk) |
| 4 | `POST https://links.webdesign.uk/c/<junk>/confirm` (suffix control) | 404 at the box |
| 5 | **the containment itself:** from the box, `curl core-manager:8090/api/v1/admin/work-items` | 404 — the admin API is not on this port |
| 6 | **negative control for #5:** same path against `:8088` | not 404 (401/403) — proves #5's 404 is the listener, not a broken curl |
| 7 | facts relay still live: chat bot answers on the shopfront | unchanged (8088 untouched) |

Check #5 with #6 is the whole point of the change; #5 alone would pass just as well against a typo'd hostname.

## 7. Council

One round, before or alongside the commit (`Council-Submitted:` trailer, since forward-only forbids an amend). In scope: `platform/`, `internal/`. Submission will carry the measurements in §2, §3 and §5 as `grounded_in` — **including the checks actually run**, which is the defect this lane's own round 1 was caught by on 08-20: *a check you ran but did not cite is a check you did not run.*

## 8. Register

`SYS-094` is the two-door pattern (ruled Q1). This listener is a new callable mechanism and gets its **own entry in the same commit that ships it** (owner ruling 2026-07-28 condition 2), under `docs026_concept_register/register/system-architecture.md`, naming the residual in §3 so the next reader inherits it.

## 9. Corrections to the inputs this plan was built from

- **`DECISION_2026-08-21b` §1's headline is spent.** "The ZIP download link cannot be built" was true of the credential question, which the owner **resolved the same day** (use the existing `zip-deliverer`, already on `isStorageEnabledAgent`; pre-mint and refresh; `/d/` is pure DB→302). What remained was §5's exposure question, and this plan is its answer. `/d/` is not built here, but nothing in the credential story blocks it any more.
- **`SUMMARY_2026-08-25_webdesign_uk_build_service.md` lists the second-click page as "designed but not built"** and "still the one owed code task". It was built at 13:34 on 2026-08-25 by the web_admin_console lane, and that lane's own cross-lane NOTE landed in the webdesign directory at 13:33 — before the summary was written at 15:27. Correcting by cross-lane note, not by editing another lane's summary.
