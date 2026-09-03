# NOTES — static site form endpoint

Running record, append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-09-03 — session 2 picks the thread up, and the pre-plan's survey does not survive contact

The owner asked this session to pick the thread up if it could find it. It found
`PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md` (the pre-plan), the
`site_delivery_and_editor` D1 review, and today's `portfolio_positioning` CONTRIB.

### 1. The first thing I did was re-run the pre-plan's own measurements, and most of them broke

Not out of suspicion — the pre-plan is a day old and the tree moves at ~1,500 commits/week, so
re-grounding a figure before repeating it is the standing rule. Five of its load-bearing claims
came out differently.

**1a. `platform/httpguard` exists, and it is a form intake gate.** `[MEASURED 2026-09-03]`

```
$ ls platform/httpguard/
clientip.go  httpguard_test.go  intake.go  limiter.go
$ git log --format='%h %ad %s' --date=short -- platform/httpguard/ | tail -1
3632874d4 2026-07-28 feat(platform/httpguard): one client key, one banded limiter, one intake gate
```

Its exported surface is `CheckIntake(honeypotValue, elapsedMillis string, minFill time.Duration)`,
`NewLimiter(bands ...Band)/Allow(key)`, and `ClientIP(r, front)` with `Nginx()`,
`CloudflareTunnel()` and `Direct()` front-ends. That is the whole of the pre-plan's D4 — "the part
most likely to be under-scoped, and the reason a form endpoint is not a small job" — built five
weeks before the pre-plan was written.

**1b. `platform/mailer` exists, and its header names contact forms as its third intended caller.**
`1d747f5e8 2026-07-28 feat(platform/mailer): the platform's one way to send email`. From
`platform/mailer/mailer.go:11-13`:

> Every "we'll email you a link" journey needs this: idea.uk's paid report today, the gripper
> dossier next, **contact forms after that**.

**1c. The public receiver is already live.** The pre-plan calls D1(b) "a new public surface to
defend". It is not new. Probed live, with controls, so that a 404 means "not routed" rather than
"everything 404s":

```
OPTIONS https://tools.apis.uk/api/v1/tools/gauntlet/round   -> 204   (registered; app is up)
OPTIONS https://tools.apis.uk/api/v1/tools/gripper/session  -> 204   (registered)
OPTIONS https://tools.apis.uk/api/v1/tools/gauntlet/nope_…  -> 404   (negative control)
GET     https://tools.apis.uk/health                        -> 404   (outside the Caddy prefix)
```

The chain is Cloudflare tunnel → island Caddy (`gauntlet_dead_cta/infra/island/Caddyfile`:
forwards `/api/v1/tools/*` only, 1 MB body cap, everything else 404) → `tools-api`, a gin service
that already imports **both** `platform/mailer` and `platform/httpguard`.

**1d. The hosting-mode table that D1 was built on is not the deciding fact.** Two static sites
already POST to that endpoint from published markup:

```sql
SELECT s.domain, count(*) FROM page_components pc
  JOIN pages p ON pc.page_id=p.id JOIN sites s ON p.site_id=s.id
 WHERE pc.rendered_html ILIKE '%tools.apis.uk%' GROUP BY 1;
--  robot-hands.com | 1
--  vonc.com        | 1
```

A cross-origin POST from a static page to a real backend is not a thing this estate needs to
invent; it is a thing it has been doing for over a month. Separately, the pre-plan's mode table
describes 5 of 39 live sites — 34 have `github_repo` **and** `publish_target` NULL, because
`publish_target` is itself an opt-in seam (`publish_site_action.go:120`: *"publish_target not set
for <domain> (seam is opt-in, default OFF)"*).

**1e. "Every form the estate has ever built is a decoration" is false at the served layer, and
this is the correction that matters most.** The pre-plan's census read `content_data`. The render
seam rewrites the value on the way out — `deliverableFormAction`
(`platform/orchestration/actions/component_library.go:1495`) replaces a non-delivering action with
`mailto:<sites.email>`, the pattern **the owner chose on 2026-07-17**. Same day, both layers:

| layer | result |
|---|---|
| `content_data->>'form_action'` | `#contact` × **27** (the pre-plan counted 22 the day before — it grew by 5 overnight) |
| `rendered_html`, contact-form components | **21 of 27 serve a real `mailto:`**; **6 components on 6 sites** still serve `#contact` |

The 6 are exactly the address-less sites: `boxingonline.com`, `cv1.co.uk`, `farmerinsurance.uk`,
`garden-tools.uk`, `relojistas.com`, and the `pool-ai-agents.internal` pool row. All have an empty
`sites.email`, which is precisely the case `deliverableFormAction` refuses to guess at.

> **This is a `[MEASURED]` marker doing no work.** The pre-plan's figure was marked measured,
> dated, and had its query written out — and it still could not have come out any other way,
> because the query was aimed at the layer the defect had already been repaired *below*. Recorded
> in `WRONG_CALLS.md`.

### 2. What I found that nobody had: a dead form the detector structurally cannot see

While enumerating form shapes I looked at the forms whose action is neither `mailto:` nor `#`:

```
gamesdesign.co.uk  /premium.html        action="/request"
gamesdesign.co.uk  /contact/index.html  action=""
idea.uk            /tools.html          action="/audience-check"
idea.uk            /report.html         action="/request"
relojistas.com     /index.html          action="/intent"
```

Probed each with the two controls the estate's own `scripts/probe-page-url.sh` insists on — a
known-good sibling that must be 200, and an invented URL that must be non-200:

| domain | sibling (must 200) | invented (must 404) | **target** | reading |
|---|---|---|---|---|
| `idea.uk` | `/report.html` 200 | 404 | `/request` → **405** | real POST handler (VM site) |
| `relojistas.com` | `/index.html` 200 | 404 | `/intent` → **405** | real POST handler (VM site) |
| `gamesdesign.co.uk` | `/premium.html` 200 | 404 | `/request` → **404** | **no handler — the form is dead** |

The 405s are the positive control that makes the 404 mean something: the same probe, on the same
day, distinguishes a live handler from an absent one. `gamesdesign.co.uk` is not a `vm-sites` row,
so there is no application that could answer `/request` at all.

**`check_contact_form_undeliverable` cannot see this, and that is a structural property, not an
oversight.** Its predicate is an enumeration of known-bad literals — no action, empty, `#…`,
`/contact` (`check_contact_form_undeliverable.go:99-104`, mirrored by `nonDeliveringFormActions`
at `component_library.go:1448-1454`). A *plausible-looking* action passes unprobed. The check
knows whether the value is on a list; it never asks whether anything answers.

### 3. What the reuse actually is

`internal/tools-api/handlers/gripper.go:227` `GripperSubmitHandler` is already a
browser-facing form intake, in production, and it does every one of the things D4 asked for:
`httpguard.CheckIntake` runs **first**, a bot verdict gets the success body and stores nothing
("a distinguishable rejection tells a spammer which gate to tune"), the IP is hashed, and the
logging is structural only. The forms endpoint is that handler with a different payload.

### 4. The one place the reuse must NOT be copied

`site_id` in those handlers comes from `c.GetString("site_id")`, which reads as "identity from
middleware, not from the request" — the right shape. But the middleware that sets it derives the
site from the **`Origin` header** (`middleware/cors.go:17` → `store.ActiveSiteByOrigin` →
`store/sites.go:34`). An attacker sets `Origin` freely.

For the gauntlet and the gripper that is a rate-limit-bucket and attribution concern. For an
endpoint that **emails a recipient** it is a spam relay wearing the estate's name: forge the
Origin, and a submission attributed to any estate site lands in that site's mailbox. The
pre-plan's own D3 review point says exactly this; the existing code predates it.

Owner decision this session: **per-site token stamped at build**; Origin stays a CORS check only.

Second, smaller thing in the same query: `ActiveSiteByOrigin` scopes on `status = 'deployed'`
alone. Yesterday's `744`/CLM-033 ruling established that the estate's liveness convention is
`IN ('active','deployed')`, and that a narrower predicate re-creates a blind spot one status value
over. Today that is **latent, not live** — 39 `deployed`, 0 `active` — so it is a note, not a
finding, and the new middleware will use the wider predicate rather than inherit the narrower one.

---

## 2026-09-03 (later) — MY OWN WRONG CALL: I applied the schema to the wrong database

**I said `tools-api` runs in the cluster and reads `clients_db`. It does not.** Migration 756 is
applied to `clients_db`, and the public receiver cannot see it.

### What I assumed, and what is actually true

I found `deployments/kustomize/services/tools-api/` — a `ClusterIP` Deployment on :8083 with a
`DATABASE_URL` from `tools-api-secret` — and treated that as *the* tools-api. It is **a** tools-api.
The one serving `tools.apis.uk`, the one I probed and got 204s from, is a **different deployment on
a different machine with a different database**:

```yaml
# gauntlet_dead_cta/infra/island/docker-compose.yml — Mythic Beasts VM toolsapisuk
postgres:
  environment: { POSTGRES_DB: tools_api, POSTGRES_USER: tools_api }
  ports: ["127.0.0.1:5432:5432"]   # host-local only (pg_dump/debugging); never public
```

and, in the same file, the line that names the consequence exactly:

> CORS comes from **the island DB's minimal sites table**, NOT an env allowlist

Migration `436`'s header says it in capitals — *"**TARGET**: the ISLAND Postgres, **NOT**
clients_db"* — and ledgers into the island's own `island_migrations`. The island's `sites` table is
a **hand-seeded mirror**: `RUNBOOK_island.md` records that it "currently holds only `vonc.com`",
that `robot-hands.com` is absent until 436 runs, and that until then "CORS answers 403 to the
widget". Confirmed in `clients_db` too: `gripper_chat_sessions` and `gripper_report_requests` do
not exist there.

### What caught it

Reading `internal/tools-api/store/gripper.go`'s package comment on the way to copying its shape —
*"three tables created by migration 436 … **on the ISLAND Postgres**, beside gauntlet_rounds"*.
Not a check I ran; a sentence I happened to read while doing something else.

**The check that would have caught it deliberately, and which I skipped:** I probed the endpoint
over HTTP and confirmed it was live, then reasoned about its database from the *kustomize
manifest*. Those are two different artefacts describing two different processes, and nothing
forced me to notice. **The cheap check is to ask the artefact you actually probed what it is
connected to** — the deployment manifest is a claim about a deployment, not about the process
answering your request. I had already written the "prove it at the artefact, per service" rule into
my own plan and then applied it to liveness only, not to identity.

> This is the second census-shaped error in this lane in one day, and it has the same shape as the
> first: **an answer taken from the layer that was convenient rather than the layer that decides.**
> The pre-plan read `content_data` where the visitor reads `rendered_html`; I read a kustomize
> manifest where the request reads a `DATABASE_URL` on another machine.

### What this costs, and what it does not

Migration 756 is **not wasted and not wrong** — `site_form_routes` and `form_submissions` belong in
`clients_db`, because that is where the render seam and the mailer live. What was wrong is the
assumption that the *receiver* could read them. Nothing is broken: both tables are inert and empty,
and no code references them.

What it does change is the architecture, and for the better.

### The design this forces, which is the estate's own proven shape

The two readers cannot share one database, so **stop trying to make them**:

| where | what it holds | why there |
|---|---|---|
| **island** (`tools_api`) | `form_submissions_inbox` — whatever the browser posted, plus the token as *presented* | it is the only database the public receiver can reach |
| **cluster** (`clients_db`) | `site_form_routes` (756) — token → site, intent, recipient, enabled | the render seam stamps the token and the mailer reads the recipient; neither runs on the island |
| **cluster** (`clients_db`) | `form_submissions` (756) — the durable record | the lead is a business record and belongs where the business data is |

and a **collector** pulls island → cluster, exactly as `GET /api/v1/tools/gripper/requests` +
its poller already does, and as `order-intake-collector` does for the shopfront.

**This is what the seam reviewer told us to carry across whatever D1 decided**, and I did not
recognise it as load-bearing until the database split forced it:

> the shopfront's **poll-collector** shape (receiver stores; `order-intake-collector` pulls) means
> the cluster exposes no inbound surface for submissions and receipt survives cluster downtime. A
> (b) receiver that stores-and-is-polled keeps that property while shedding (a)'s single-box
> coupling.

**And it makes the security property stronger than the design I submitted to council.** The island
never decides whose submission it is: it records the token it was handed. **Identity is resolved in
`clients_db`, against a table the island cannot see**, at ingest. So a forged token cannot reach a
mailbox — it is stored on the island, fails to resolve, and is discarded. That materially answers
risk (3) of `submission_756_form_endpoint_storage.json` (a bearer token published in page markup):
holding a site's token lets you inject into that site's own queue, which is what its form does
anyway, and nothing else.

### The consequence I cannot engineer away

**The island is deployed by hand and I cannot reach it.** `docker save | ssh load` onto a Mythic
Beasts VM, with secrets in a root-only `/opt/island/.env` that is not in this repo. So this lane
can write the receiver, the inbox migration, the collector and the render-seam branch, and can get
all of them committed and reviewed — and **"live" is a step only the owner or the gauntlet lane can
take.** Committed and live are separate facts here in a stronger sense than usual, and no amount of
care on my side collapses them.
