# RFC_054 — the cluster's public exposure boundary, reviewed at the second door

**Status: RULED 2026-08-25 — all three questions answered by the owner; see §5. Filed
2026-08-24 by the web_admin_console lane, the evening `links.webdesign.uk` went live.** This fulfils the architecture seat's stated approval
condition from the `admin.apis.uk` round: *"a second and third publicly-proxied prefix
should trigger a boundary review."* The second deliberate prefix is now live; this RFC is
that review's brief, written while every figure below was freshly measured.

## 1. The public surface, measured 2026-08-24 (evening)

**Two doors reach the cluster from the internet. Everything else that looks public does
not touch it.**

| door | reaches | gate | state |
|---|---|---|---|
| `admin.apis.uk` | admin-dashboard:8080 (SPA + core-manager/auth-service APIs) | Cloudflare Access (OTP, 2 emails) + box nginx fail-closed on the Access header | LIVE, owner using it |
| `links.webdesign.uk/c/<token>` | core-manager:8088, ONE handler | token IS the credential (256-bit, hashed, expiring, single-use) + token-shape regex at the box + edge rate limit (429s proven) + prefetch guard | LIVE 2026-08-24, verified 404/404/200/429; **0 tokens exist** |
| ~39 portfolio zones (noted.co.uk, idea.uk, robot-hands.com, apis.uk …) | **nothing in the cluster** | — | static serving from the B2 bucket / git route via the Worker |
| `webdesign.uk/c/` + `/stripe/webhook` | (would reach core-manager / auth-service) | **parking 302 to webdesign.co.uk swallows both** (measured) | inert until the shopfront unparks; `/c/` also now removed from the committed vhost |

Shared plumbing behind both doors: cloudflared tunnel (box dials out; no inbound ports),
box nginx on loopback, WireGuard leg fenced by `networkpolicy-wireguard-egress.yaml` to
exactly kube-dns / core-manager:8088 / auth-service:8081 / admin-dashboard:8080 (postgres
proven blocked, with control). Zero Ingress objects in the cluster.

## 2. What the review is FOR — three questions, in rising order of consequence

**Q1 — Is the two-door pattern the standing pattern?** Both doors were built the same
way: hostname-scoped vhost on the box, exact-path exposure, tunnel-only, gate at the
edge appropriate to the audience (Access for the owner; token-in-link for customers).
If this is the pattern, say so once — the next door (`/d/` downloads, the Phase-5 editor
session, a future `/stripe/webhook` unparking) copies it instead of re-deriving it, and
the seat's condition converts from "review each door" to "review deviations".

**Q2 — The containment candidate: a delivery-only listener.** Today the links door
proxies to core-manager:8088 — the SAME port that serves the whole admin API. The
defence is the nginx location regex, i.e. configuration on a box. The candidate: a
second HTTP listener in core-manager (e.g. :8090) serving ONLY the delivery routes
(`/c/`, later `/d/`), with the box and the egress fence pointed at it. Then a box
misconfiguration — the exact class the LANDMINES file keeps collecting — could expose
nothing but delivery routes. Cost: a listener + config + fence change + roll; a council
round of its own. **The trade to rule on: is box-nginx config an acceptable last line,
or must the blast radius be capped in the binary?** (CLAUDE.md 2026-08-02 §2's own
logic — "a comment is not a control" — leans toward the binary; the counter-argument is
a second listener is a second thing to keep in step.)

**Q3 — What makes door number three automatic?** The `/stripe/webhook` door becomes
reachable the day the shopfront unparks — nobody will run a review that day unless a
mechanism fires. Candidate: a LANDMINES/register entry stating the rule from Q1 plus a
one-line check in the box files' headers ("adding a hostname or location that proxies
into the cluster ⇒ cite the Q1 pattern or file an RFC"). Cheap, prospective, and it is
the difference between this review being a norm and being an event.

## 3. Facts a reviewer should not have to re-derive

- The mail-scanner hazard on `/c/` is CLOSED BY RULING, not accepted: GET becomes
  render-only, confirm is a POST from the page, no delivery email before it ships
  (`webdesign_uk_build_service/DECISION_2026-08-24_confirmation_needs_a_second_click.md`).
- `customer_access_tokens` = **0** and handed_over/confirmed = **0/0** as of 2026-08-24
  evening — every "nothing at risk" statement expires the moment either is non-zero.
- The admin door's residuals are recorded in `RUNBOOK_deploy_admin_apis_uk.md` (bare
  role-string check behind Access; header-presence guard, not JWT validation).
- Full census, foot-guns and verification transcript:
  `web_admin_console/HANDOFF_2026-08-24b_continue_here.md` §3.3–§3.10.

## 4. What this RFC is not

Not a council-gate submission (prose is out of the gate's scope) and not a proposal to
build Q2's listener — that gets its own plan and council round IF this review rules for
it. It is the boundary look the architecture seat asked to happen, packaged for a human
ruling on Q1–Q3.

## 5. RULED — owner, 2026-08-25 (session "webdesign.uk live webdesign")

- **Q1 — YES: the two-door pattern IS the standing pattern.** Codified as register entry
  **SYS-094** (`docs026_concept_register/register/system-architecture.md`). The
  architecture seat's condition converts from "review each door" to "review deviations
  from SYS-094".
- **Q2 — BUILD the delivery-only listener.** Its own plan + council round; aim to land it
  before the first `customer_access_tokens` row exists. It does NOT gate the shopfront
  unparking, which no longer touches core-manager at all (`/c/` moved to
  `links.webdesign.uk`, 2026-08-24). Until the listener lands, the box nginx location
  regex is the stated — not hidden — last line.
- **Q3 — the register entry plus header lines.** SYS-094 plus a header line in every
  `box/*.nginx` and `box/*.cloudflared-ingress.yml`: a new hostname or cluster-proxying
  location cites SYS-094 or files an RFC first. **This review is RECORDED as the boundary
  review for `/stripe/webhook`** (safe by design: HMAC over the raw bytes, honest 503
  while keyless, `auth-service:8081` already the fence's narrowest grant) — so unpark day
  owes no further review, which was Q3's motivating scenario.
- **And the unparking itself was ruled GO the same day**, with a temporary hand-placed
  "Not active yet" label above the shopfront's two CTAs (vm-sites `444205b`) until
  ordering opens. The label is outside the framework by explicit owner instruction and
  any framework redeploy of the page removes it.

## 6. OPEN RESIDUAL, recorded 2026-08-25 — the facts relay keeps a second path to the admin port

**Q2 is BUILT** (`site_delivery_and_editor` lane, commit `d30917150`, register **SYS-095**):
the delivery routes now live on `core-manager:8090`, the box vhost and the wg fence point
at it, and `assertNoDeliveryRoutes` refuses to construct the server if a customer route is
ever mounted back on the admin router.

**What Q2 did NOT close, tracked here at the architecture seat's request** (council
`25cd3044` round 1, `architecture`, low: *"disclosure in a YAML comment is not a tracked
item — recommend a follow-up work item/RFC note so this residual doesn't silently become
permanent"*):

**Port 8088 remains reachable from the webdesign.uk box**, because the chat bot's facts
relay (`GET /api/v1/site-facts/:domain`, `box/chat-service/facts.go`) uses it, and the
fence's own comment records that removing it *"stops the bot STARTING, by its own design"*.

So the containment Q2 bought is exactly: **the existing customer door cannot be widened
into the admin API.** It is *not*: "the fence contains the admin port." A **new** vhost
written straight to `:8088` would still reach every admin route. The last line for that
path is still SYS-094's discipline — cite the pattern or file an RFC — and not a mechanism.

**What would close it:** move the facts relay onto a box-facing listener of its own (or
onto the delivery listener, accepting that it stops being delivery-only), then drop 8088
from `networkpolicy-wireguard-egress.yaml`. The fence then becomes the control, and a
misconfigured vhost reaches nothing that matters however it is written.

**Why it was not done in Q2:** it touches a live customer-facing bot whose failure mode is
"does not start", and the owner's ruling scoped Q2 to the delivery routes. Doing it inside
that round would have been exactly the bundling the guardian seat vetoed in `bugs_closed/124`.

**Status: OPEN.** Not scheduled, not owned. Whoever takes it should read SYS-095 first —
the delivery listener is the worked example of the same move.
