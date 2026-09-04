# NOTES — static site form endpoint

Running record, append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-09-04 — session 2 picks the thread up, and the pre-plan's survey does not survive contact

The owner asked this session to pick the thread up if it could find it. It found
`PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md` (the pre-plan), the
`site_delivery_and_editor` D1 review, and today's `portfolio_positioning` CONTRIB.

### 1. The first thing I did was re-run the pre-plan's own measurements, and most of them broke

Not out of suspicion — the pre-plan is a day old and the tree moves at ~1,500 commits/week, so
re-grounding a figure before repeating it is the standing rule. Five of its load-bearing claims
came out differently.

**1a. `platform/httpguard` exists, and it is a form intake gate.** `[MEASURED 2026-09-04]`

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

## 2026-09-04 (later) — MY OWN WRONG CALL: I applied the schema to the wrong database

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

---

## 2026-09-04 — council round 1 on migration 756: APPROVED, and the objections are worth more than the verdict

`SUBMISSION_CORR=3aff429e-c08e-4302-a6a7-0b465dc5229f` → **APPROVED**, "4 advisory objection(s) —
none high-severity" in the summary line, 8 in the report. Dispositions, because an approved verdict
whose objections nobody answers is the coverage report's other dishonesty surface.

### ACCEPTED — the one that found a real gap in my process (`reuse_agent`, medium)

> This mints a brand-new bearer-token identity scheme … but the landmine register already documents
> an existing token-based public-route pattern (`customer_access_tokens`, with `/c/<token>` confirm
> handlers) … The plan's `grounded_in` never shows that mechanism was checked for reuse.

**Correct, and I had not looked.** `[MEASURED 2026-09-04]` `customer_access_tokens` is a mature
per-site, purpose-scoped token table with `token_hash` (hashed at rest), `expires_at NOT NULL`,
`revoked_at`, `single_use`/`used_at`/`use_count`, and partial indexes on
`(site_id, purpose) WHERE revoked_at IS NULL`. Six Go call sites across `platform/delivery/` and
`platform/orchestration/actions/`.

**But it is not a drop-in, and the reason is specific enough to be worth writing down rather than
either adopting or dismissing it:**

| its tokens | mine |
|---|---|
| secret; emailed to one customer as a link | **published in page markup**, visible to every visitor by design |
| `expires_at` NOT NULL | must live as long as the page is published |
| single-use / counted | used by every visitor who submits |

So **hashing at rest buys almost nothing here** — the plaintext is public by construction — and an
expiry would make a live form die silently, which is exactly the `017` shape the `bug_historian`
seat flagged in the same round. What I should take, and did not have, is the **revocation and
rotation half**: `revoked_at`, and the ability to hold two valid tokens while a re-render propagates
a new one. That is owed **before the receiver's resolve-by-token contract goes live**, which is now,
because nothing has run yet.

The seat's own framing — two parallel token schemes with divergent semantics — is the right worry.
The answer is not one table for both, it is one *vocabulary*: mine should use `customer_access_tokens`'
column names and semantics where they apply, so a reader of either recognises the other.

### PARTLY ACCEPTED, and one sub-claim is FALSE (`debug_historian`, medium)

> the migration was applied to the live DB BEFORE this council round … The mutation-testing of the
> verify block is genuinely good practice, but **it was run against an already-committed table, not
> the sketch under review.**

**The first half is right and I accept it**: I applied before submitting, so the
`information_schema` evidence in my `grounded_in` is a post-apply re-read of the thing being judged.
That is a fair hit and the ordering should have been submit-then-apply.

**The second half is not true, and the record should say so.** All three mutations ran through
`run-migrations.sh`'s doomed-transaction probe — which executes the file and rolls back — **before**
`--apply` was ever run. The probe output for each (`?? probe inconclusive: ERROR: P0001: verify:
enabled did not default to false …`) is a pre-apply artefact by construction: the runner probes
pending files, and a file it has applied is no longer pending. So the mutation testing was against
the sketch, and it is the one piece of evidence in that submission that could not have been
contaminated by the early apply.

Recording the rebuttal rather than quietly accepting the whole objection, because a wrong reason
attached to a right conclusion is how a good practice gets abandoned later by someone reading only
the objection.

### ACCEPTED ON EVIDENCE, REBUTTED ON CONCLUSION (`prior_art_librarian`, medium)

> `436_tools_api_gripper_intake.sql` … neither is mentioned in `grounded_in` or ruled out. If
> 'gripper' already implements site-scoped submission intake + routing … this migration risks being
> a second, uncoordinated implementation.

**The evidence objection is right**: 436 is absent from my `grounded_in` and it should not have
been. **The conclusion is wrong, and the reason is the most important thing this lane learned
today** — 436 is not a competing implementation, it is the *model*. Reading it is what revealed
that `tools-api` runs on the island with its own Postgres, which reshaped this entire design into
the receiver-stores/cluster-pulls split that 757 now implements. The submission failed to show its
work; the work was done.

### ACCEPTED, converging on one remedy (`guardian` ×2, `architecture` ×2, `bug_historian`)

- **no rotation or leak-detection story for a published bearer token** (guardian low, architecture
  medium) — same remedy as the `reuse_agent` objection above. The architecture seat asks for a note
  *before the receiver ships*; this lane's own plan already says an architecture round is owed, so
  the two agree.
- **`ON DELETE RESTRICT` couples any future site-deletion pipeline** (guardian low) — already
  risk (1) of my own submission; the seat and I agree on the shape and on it being worth the cost.
- **`notified_at`/`notify_attempts`/`notify_error` is an implicit delivery-guarantee contract**
  (architecture low) — accepted; it must be named explicitly rather than emerging from whichever
  handler is written first. It belongs in the collector's design, not in the schema's comments.
- **case `017` (`static_cutover_orphans_backend_entry_forms`)** (bug_historian low) — a form backend
  left unreachable after a routing change, which is precisely what a rotated-or-dropped token does
  to already-published markup. **Carried into the receiver review as a named precedent** rather than
  left to be rediscovered, which is what the seat asked for.

---

## 2026-09-04 — CORRECTION: everything above was originally dated 2026-09-03, and it was all written today

**The whole of this lane's first day was misdated by one day.** Filenames, headings, `[MEASURED]`
markers, both migration headers, the bug file, two `WRONG_CALLS.md` entries, the `LANDMINES.md`
entry, and eight commit messages all said **2026-09-03**. The date was and is **2026-09-04**.

**How it happened.** I never checked. The session's own environment block states the date plainly,
and I did not read it — I inferred "today" from the material I was reading: the pre-plan is
2026-09-02, the `portfolio_positioning` CONTRIB is 2026-09-03, and the most recent commits in the
log were 09-03. Every one of those is correctly dated *for what it is*, and reading them in
sequence produces a confident, wrong answer for *now*. A date inferred from the newest thing you
have read is the age of that thing, not the time.

**Why it matters here more than it would elsewhere.** CLAUDE.md's rule that a count must carry the
date it was counted exists so staleness is **mechanically checkable** —
`git log --since=<census date> --diff-filter=A -- <dir>` is the check, and it takes the date
literally. A census stamped one day early silently widens that window: it would list a day of
additions as "since my count" that were in fact *before* it, and a reader reconciling them would
be chasing changes that cannot explain a discrepancy. The error is small and it corrupts the exact
mechanism the rule was written to enable.

**What was corrected, and what deliberately was not.**

- **Corrected:** every self-dated claim in this lane's files, three filenames (`PLAN_…`, the
  gauntlet CONTRIB, `bugs_open/471_…`) and every reference to them, plus two relative-time phrases
  in `README_where_we_are.md` that were wrong by a day once the absolute dates moved ("a pre-plan
  written yesterday" → two days earlier; "a note that arrived this morning" → yesterday).
- **NOT corrected, on purpose:** `submission_756_form_endpoint_storage.json` still says
  "Applied and verified 2026-09-03". It is the artefact the council actually reviewed under
  correlation `3aff429e`, and editing a submission after its verdict misrepresents what was judged.
  The error is recorded here instead.
- **NOT correctable:** eight commit messages say 2026-09-03. Forward-only forbids an amend, and
  this note is the record. **A reader reconciling this lane's commit dates against its documents
  should trust `git log`'s own timestamps, not the dates written in the messages.**
- **Other lanes' 2026-09-03 references in these files are correct and were left alone** — the
  `portfolio_positioning` CONTRIB's own header, migration 744/CLM-033, `BRIEF_2026-09-03c`. The
  correction was applied by explicit per-string replacement with an expected-count assertion on
  each, not by a blanket date sweep, precisely because a sweep would have falsified those.

**The check, and it costs nothing:** run `date -u` at the start, before writing the first dated
thing. Not "check the environment block" — *run the command*, because the failure mode is not
missing information, it is not looking. See `WRONG_CALLS.md`.

> One thing this did **not** damage, worth stating because it was the first thing I checked: the
> measurements themselves are unaffected. Every figure was taken today and is correctly attributed
> to today's system state; only the label was wrong. The `744`/CLM-033 citation was independently
> re-verified while fixing its date and is **correct** — 744 is the widening of the site predicate
> to `IN ('active','deployed')`, it names CLM-033/migration 742, and it landed 2026-09-03, so the
> "yesterday's ruling" phrasing is now accurate rather than approximate.
