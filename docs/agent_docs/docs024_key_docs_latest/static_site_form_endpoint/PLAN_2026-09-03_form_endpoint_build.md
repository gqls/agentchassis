# PLAN 2026-09-03 — build the form endpoint

**Supersedes the decision space of `PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md`, which
stays in place as history.** The pre-plan was written to be extended in another thread; this is
that extension. Its D1–D6 are answered below — four by the owner today, the rest by measurement.

Corrections to the pre-plan's premises are recorded as corrections, in the pre-plan itself and in
`NOTES_static_site_form_endpoint.md` §1. They are not edited away, because the reasoning that was
wrong is the part worth keeping: it was wrong about *what already exists*, which is the failure
mode this estate keeps paying for.

---

## 1. What we are building, in one paragraph

A static site can already POST to a real backend in this estate — `robot-hands.com` and
`vonc.com` do it today. What is missing is a **general** receiver: one route group on the existing
public `tools-api`, which stores each submission and then emails a recipient held in the database
rather than baked into the page. A site opts in by having a row; without one, nothing changes and
forms keep behaving exactly as they do now.

## 2. The decisions, and why

### D1 — where the receiver lives: **(b), as a route group on `tools-api`**

Settled by evidence rather than preference. The pre-plan called (b) "a new public surface to
defend"; the surface is not new. `https://tools.apis.uk/api/v1/tools/*` is live (probe recipe in
the RUNBOOK §3), fronted by a Cloudflare tunnel and an island Caddy that forwards **only** that
prefix with a 1 MB body cap, and served by a gin service that already imports both
`platform/mailer` and `platform/httpguard`.

D1(c) — the Cloudflare edge — was opposed by the seam owner on four cited grounds
(`site_delivery_and_editor/REVIEW_2026-09-02_form_endpoint_preplan_D1_vs_publish_seam.md`) and is
not revisited here. D1(a), the box, is additionally weakened by a fact the pre-plan did not have:
the box's order intake is a Go service living **under `docs/`**, outside `go build`
(`webdesign_uk_build_service/box/chat-service/orders_http.go`) — the exact out-of-build shape that
`platform/mailer` was created to end.

**The route group must be `/api/v1/tools/forms`.** The prefix is load-bearing, not cosmetic: the
island Caddy 404s everything outside `/api/v1/tools/*`, so this path needs **no edge change**, and
a path chosen for tidiness instead would need one.

### D2 — what happens to a submission: **store, then email a DB-held recipient** (owner)

The row is the durable record; notification reads its recipient at delivery time. This is what
makes the CONTRIB's requirement satisfiable — *"the recipient is expected to CHANGE without the
page changing… that is a routing concern, not a form concern, and it is the thing we would most
like designed in rather than bolted on."*

The pre-plan deferred D2 to `bugs_open/420`'s identity replumb, on the grounds that the estate
does not model who a site's owner is. **That deferral is now discharged, and by a route the
pre-plan did not anticipate.** 420's class fix separated the two contracts that shared
`sites.email`: the payer's address is delivery-only and stays in
`build_queue.direction.customer_email`; `sites.email` is now *only* the published contact, written
only from an explicit `direction.published_contact` opt-in. Neither is a form recipient. So this
lane does not wait on 420 — it adds the **third** identity 420's split makes room for, in its own
table, where changing it is a config update.

### D3/D4 — site identity and abuse: **a per-site token stamped at build** (owner)

`site_id` in the existing handlers comes from middleware, which is the right shape — but the
middleware derives it from the **`Origin` header** (`middleware/cors.go:17` →
`store/sites.go:34`), and an attacker sets `Origin` freely.

For the gauntlet and the gripper that is an attribution and rate-limit-bucket concern. For an
endpoint that emails a recipient it is a spam relay wearing the estate's name: forge the Origin
and a submission attributed to any estate site lands in that site's mailbox. The pre-plan's own
D3 review point states the rule — *derive the site from something the visitor cannot choose* —
and the existing code predates it.

So: the render seam stamps a per-site token into the form; the endpoint resolves the site from the
token; `Origin` stays a CORS check only. The rest of D4 is already built —
`httpguard.CheckIntake` (honeypot + time-to-fill), `httpguard.NewLimiter` (banded), a 1 MB cap at
Caddy — and is reused, not rewritten.

### D5 — what the visitor sees: **a framework-built thank-you page, reached by a 303**

The seam review's point stands: a thank-you page is just a page, so D5 creates no pressure to move
the receiver to the edge. The endpoint answers a form-encoded POST with `303 → the site's
thank-you page`, and a JSON POST with `{"accepted":true}`. A plain cross-origin form POST needs no
JS, and `CheckIntake` fails open on a missing `_elapsed`, so a JS-less visitor degrades correctly
rather than being scored as a bot.

**The thank-you page goes through the framework.** No hand-authored HTML, however small (owner
ruling 2026-08-04).

### D6 — retrofit: **forward-only, plus fix the broken ones** (owner)

The pre-plan framed this as "22 live decorations vs. a fleet migration". The real fleet state is
neither. Measured at the served layer on 2026-09-03: **21 of 27 contact forms deliver a real
`mailto:`** — the render seam repairs them, per the owner's 2026-07-17 ruling — and only **6
components on 6 address-less sites** still serve `#contact`.

So there is no fleet migration to argue about. Working `mailto:` forms are left alone; the
endpoint is opt-in. What *is* owed:

- the 6 address-less sites, which the seam correctly refuses to guess an address for and which the
  endpoint can now answer;
- **`gamesdesign.co.uk/premium.html`, whose form posts to `/request` and gets a 404** (NOTES §2) —
  a dead form on a static site, found by this lane and filed by it;
- widening `check_contact_form_undeliverable`, whose predicate is an enumeration of known-bad
  literals and therefore cannot see a plausible-looking action. **A list knows whether a value is
  on it; it never asks whether anything answers.** The fix is to probe the target — the same
  method the RUNBOOK §2 recipe uses, via `platform/fetchguard`.

## 3. Phasing

| phase | what | why in this order |
|---|---|---|
| **0** | coordinate; open these docs; record the corrections in the pre-plan | `tools-api` is the gauntlet lane's (`features_open/024`), and a CONTRIB is owed before a route group appears in their service |
| **1** | migration `750`: `site_form_routes`, `form_submissions` | DB config is live on apply; Go is inert until a roll. Schema first means the seam has something to read |
| **2** | the receiver: token middleware, handler, store, mail | reuses `GripperSubmitHandler`'s proven shape |
| **3** | the render seam: one branch in `deliverableFormAction` | the shared-mechanism change; goes to council and, on its own merits, architecture review |
| **4** | pilot on copyonline; the 6 sites; widen the detector | pilot last, because it is the thing that proves the rest |

### Phase 1 — schema

`docs/agent_docs/sql_for_agents/750_form_submissions_and_site_form_routes.sql` (+ `_ROLLBACK`,
`_VERIFY`).

- **`site_form_routes`** — `(site_id, intent, token, recipient_email, reply_to, enabled)`. One
  table carrying both the opt-in switch and the movable recipient. `intent` is copyonline's second
  axis: a commercial enquiry and a directory removal request have different recipients and
  different urgency, and the CONTRIB asks whether that widens the extensible axis. It does, and
  this is where it lives.
  **No row means no endpoint** — opt-in, with the unsafe side default OFF, visible in the config
  rather than licensed by a comment (owner ruling 2026-08-02 §2).
- **`form_submissions`** — `(id, site_id, intent, payload jsonb, ip_hash, user_agent, created_at,
  notified_at)`. `notified_at` is what makes a failed send re-deliverable, and what lets anyone ask
  whether the endpoint is actually working.

### Phase 2 — the receiver

Copy the shape of `internal/tools-api/handlers/gripper.go:227`: bot gates **first**, an
indistinguishable rejection, hashed IP, structural-only logging.

- `internal/tools-api/middleware/formtoken.go` *(new)* — resolve the site from the token; set
  `site_id`/`site_domain`; set CORS headers from the **resolved** domain. Liveness predicate
  `IN ('active','deployed')`, not the `status='deployed'` that `sites.go:34` still carries.
- `internal/tools-api/handlers/forms.go` *(new)* — `CheckIntake`, validate, store, notify.
- `internal/tools-api/store/forms.go`, `internal/tools-api/forms/` *(new)*.
- `internal/tools-api/api/server.go` — a `WithForms(...)` `RouterOption`, mirroring
  `WithPlayground`; wired in `cmd/tools-api/main.go` as the gripper is.
- SMTP under a **`FORMS_SMTP`** prefix (`mailer.FromEnv`), alongside the existing `GRIPPER_SMTP`.
  **If it is unset the submission still stores and the absence is reported** — a notification
  channel that silently drops is the failure this whole lane exists to end.

### Phase 3 — the render seam

`deliverableFormAction` (`component_library.go:1495`) gains one branch ahead of the mailto branch:
an enabled `site_form_routes` row → the endpoint URL and the stamped token; otherwise, behave
exactly as today. Both render paths already funnel through this one function, and
`component_library_form_action_test.go` already covers it.

## 4. Governance

- **Council gate** before or alongside the commit: `internal/`, `platform/` and an appliable
  migration are all in scope. `Council-Submitted: <corr>` if the verdict has not landed;
  `Council-Reviewed:` only on a verdict actually read.
- **Architecture review is owed and is not this lane's to waive.** A new public data-accepting
  surface, plus a change to what the render seam *guarantees* — from "mailto, or leave it alone"
  to "a third destination" — is the 2026-07-29 §1 trigger, not the additive-and-inert case
  RFC_022 narrowed. Route it on its own merits; do not let a build carry it.
- **Register the seam in the same commit that ships it** (2026-07-28 condition (2), which the
  2026-07-29 ruling left standing), naming the producer set and the key shape.
- **Tell the other consumers** (2026-07-29 §3): the render-seam change touches every site's
  contact form, so `site_delivery_and_editor` and `portfolio_positioning` are told what changed
  about their guarantee — not handed a list of new keys.
- A `LANDMINES.md` entry for the Origin-vs-token trap, then `./scripts/landmines-verify-dispatch.sh`.

## 5. Verification

Full recipes in the RUNBOOK. The three that decide whether this worked:

1. **Default-OFF proof** — a site with no `site_form_routes` row still renders `mailto:`
   unchanged. Without this the opt-in is a claim.
2. **The forged-Origin case, in the same run as the valid one** — a valid token with a forged
   `Origin` must attribute to the token's site; an invalid token must be refused **with no row
   written**. One arm alone is uninformative.
3. **The mail actually sends** — `notified_at` set and a delivery observed, with a demand control
   (a send that must succeed beside one that must not). A zero from a channel that has never
   worked looks exactly like a zero from a channel with nothing to send.

And, because the island is **not** the chassis roll: confirm how the island picks up a new
`tools-api` image before saying the endpoint is live. Committed, approved and not live is a real
state here.
