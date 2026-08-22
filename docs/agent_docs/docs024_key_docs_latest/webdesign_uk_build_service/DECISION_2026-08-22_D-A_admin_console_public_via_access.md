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

## The architecture seat's accumulation warning is now LIVE

`architecture` approved with a low-severity note: *"a second and third publicly-proxied prefix
should trigger a boundary review."* `admin.apis.uk` is the **second** publicly-proxied surface
on this box, after `webdesign.uk`. **The third triggers the review** — that is not a formality,
it is the condition the seat set, and it is now one step away.

## Sources

- `HANDOFF_2026-08-21_continue_here.md` §2 (D-A as posed), §5 (the seat-by-seat table)
- `DECISION_2026-08-21b_zip_download_link_needs_a_credential_home.md` §5–6 (the full objection)
- `web_admin_console/PLAN_2026-08-22_web_admin_console.md` §1 (the three options, costed)
- `web_admin_console/RUNBOOK_deploy_admin_apis_uk.md` (the build)
