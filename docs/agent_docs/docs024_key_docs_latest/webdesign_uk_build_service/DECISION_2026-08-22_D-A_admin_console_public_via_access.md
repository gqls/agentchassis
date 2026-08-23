# D-A ANSWERED — core-manager may be reachable from the internet, narrowly

**Owner ruling, 2026-08-22.** Asked whether to keep debugging the WireGuard route or move to a
web-based admin, the owner chose: *"let's go for a web based admin then"*, then chose
**`admin.apis.uk`** as the hostname and **Cloudflare Access** as the gate.

This closes **D-A**, open since 2026-08-21 and recorded in `HANDOFF_2026-08-21_continue_here.md`
§2 as one of the two decisions blocking real work. It was the `guardian` seat's **high-severity
gating objection** on council correlation `99b5af22-7150-4e91-a5e3-809fd06504c0`, and the
objection was correct: the `/c/<token>` submission had bundled *how a confirmation should work*
(ruled 2026-08-19) with *whether the service holding every site's data becomes publicly
reachable* (never asked).

## What was actually decided, stated precisely

The council's three options were **(a)** approve public reach for named paths only, **(b)** move
`/c/` and `/d/` to a service allowed to be public, **(c)** keep both behind the box.

**The ruling is (a), with the exposure narrowed further than (a) as written.** core-manager
becomes reachable from the internet **only**:

- through **`admin-dashboard`'s own gateway**, never directly;
- on **one hostname**, `admin.apis.uk`;
- **after Cloudflare Access has authenticated the request at the edge** — an unauthenticated
  request never reaches the box, let alone the cluster;
- with the box's nginx **failing closed** if the Access header is absent, so deleting the
  Access application does not silently open the API.

**`sitefacts.go`'s "NO PUBLIC EXPOSURE … ClusterIP only, no Ingress" remains literally true.**
There is still no Ingress object anywhere in the cluster (`kubectl get ingress -A` → none). The
exposure is a tunnel dialling **out** from a box, which is the estate's existing, proven
pattern — `tools.apis.uk` has run that way since 2026-07-24.

## What this does and does not unblock

**Unblocks:** the admin console (built; see `web_admin_console/RUNBOOK_deploy_admin_apis_uk.md`).

**Does NOT automatically unblock `/c/` and `/d/`.** This ruling is about an *authenticated
admin* surface. `/c/<token>` is the opposite shape — an **anonymous** link a customer clicks
from an email, which cannot sit behind Access. It needs its own answer, and the council's
options (b) and (c) are still the live ones for it. **Do not cite this decision as approval for
exposing `/c/` or `/d/`.**

The related objections on `99b5af22` also still stand and are unaffected:

- the prefetch hazard (`Sec-Purpose: prefetch`, `Purpose: prefetch`, `X-Purpose: preview`, HEAD)
  must be refused at the handler — independent of how tokens are minted;
- the `$http_cf_connecting_ip` rate-limit trust boundary must be **verified, not argued**;
- the `/d/` deferral still owes a `doc_notes` row.

## The architecture seat's accumulation warning

`architecture` approved with a low-severity note: *"a second and third publicly-proxied prefix
should trigger a boundary review."*

> **CORRECTED 2026-08-23, from the box's own `cloudflared` config.** This section first said
> `admin.apis.uk` was *"the **second** publicly-proxied surface on this box, after
> `webdesign.uk`"*. **Both halves were wrong**, and I had not looked at the tunnel config when I
> wrote it. The box serves **seven hostnames across three backend ports**: `webdesign.uk`,
> `www.webdesign.uk`, `preview.webdesign.uk` (→ :8080), `noted.co.uk`, `www.noted.co.uk`,
> `app.noted.co.uk` (→ :8082) and now `admin.apis.uk` (→ :8083). I did not know `noted.co.uk`
> was on this box at all.

**Counted correctly — by what the seat actually meant, a publicly-reachable route INTO THE
CLUSTER — `admin.apis.uk` is the FIRST, not the second.** `[MEASURED 2026-08-23]` the other two
cluster-proxying blocks are written but not reachable: `webdesign.uk/c/` and
`webdesign.uk/stripe/webhook` both sit behind a Cloudflare **page rule** that 302s
`webdesign.uk/*` to `webdesign.co.uk`, and `preview.webdesign.uk/c/x` returns 404 because that
nginx block's `server_name` is the apex, not `preview`. Control: `preview.webdesign.uk/` → 200,
so those are refusals, not failed fetches.

**So the boundary review is further away than I implied, and the count that matters is
cluster-reaching surfaces (now 1), not hostnames (now 7).**

## ⚠ AND THE THING THAT FALLS OUT OF THAT, WHICH IS A REAL HAZARD

**The only thing keeping `/c/` and `/stripe/webhook` off the internet is a Cloudflare page
rule** — a redirect that exists for *parking*, not for security. Nobody chose it as a control.

The moment `webdesign.uk` is pointed at its own shopfront — which is the whole point of that
lane, and is coming — that redirect goes, and **both routes become internet-reachable with no
code change, no deploy, and nothing to notice.** `/c/` is a GET that mutates state, and the
prefetch hazard on it is still unmitigated (`DECISION_2026-08-21b` §4).

**Whoever removes that page rule owns this.** Before it goes: mitigate the prefetch hazard, or
move `/c/` to a hostname that is not the shopfront, or gate it. Recorded in `LANDMINES.md`.

## Sources

- `HANDOFF_2026-08-21_continue_here.md` §2 (D-A as posed), §5 (the seat-by-seat table)
- `DECISION_2026-08-21b_zip_download_link_needs_a_credential_home.md` §5–6 (the full objection)
- `web_admin_console/PLAN_2026-08-22_web_admin_console.md` §1 (the three options, costed)
- `web_admin_console/RUNBOOK_deploy_admin_apis_uk.md` (the build)
