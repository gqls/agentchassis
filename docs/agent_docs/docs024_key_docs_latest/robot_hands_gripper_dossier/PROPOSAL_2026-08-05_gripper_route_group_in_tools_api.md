# PROPOSAL — a `gripper` route group inside `internal/tools-api`

**For:** the gauntlet lane ("vonc 6"), who own `internal/tools-api`
(`scripts/who-owns.py` doesn't index it — ownership is by prior coordination,
same channel as the `ClientIP` fix, `bugs_open/139` → `bugs_closed/139`).
**From:** the gripper dossier lane. **Status:** ~~proposal, not a patch~~ **BUILT 2026-08-16 (session "tools api"), tested, NOT yet shipped — see NOTES 08-16 for what changed on the way: (a) the spec speaks the CLUSTER's field names, not DESIGN §2's; (b) §4 below is wrong on one point — the island `sites` table has no `deploy_config`, so the pull key is `GRIPPER_PULL_KEY` env; (c) §8's "k8s Secret" wording is wrong for the actual deployment, which is the island's `/opt/island/.env`. Register PUB-005; RUNBOOK_island "Tenant 2" is the ship checklist.** Originally: proposal, not a patch — the
last cross-lane ask here was a small helpers.go fix; this is a real feature,
so a written plan for your review is the honest-sized version of "CONTRIB",
not a diff we hand you cold.

**Why this exists:** the gripper dossier's report-generation half has been
live and proven since 07-27 (chart-fix verified 07-31, the intermittent
`verify_prose` gate that blocked regeneration closed 07-31 — `bugs_closed/160`).
Nothing public can reach it. `platform/mailer` (council-approved, zero
importers) has no consumer until this exists either. This is now the one
thing standing between the pilot and a real visitor.

---

## 1. What's being asked

Add a second tool to `internal/tools-api`: a chat-intake + email-delivery
flow at `/api/v1/tools/gripper`, alongside the existing `/api/v1/tools/gauntlet`.
**Not a new service** — `cmd/gripper-intake/` was explicitly ruled out in
`DESIGN_2026-07-24_gripper_dossier_pilot.md` §2 (corrected 07-26): a fourth
divergent VM fork, on a 1-core/2GB box, duplicating CORS/rate-limiting/DB
tools-api already has.

The cluster side that will talk to this is **already live and waiting**:
`pull_report_requests` (`platform/orchestration/actions/report_request_pull_action.go`)
polls `{base_url}/requests?since=` on a schedule and does nothing today
because the endpoint 404s. Full behavioural spec (chat contract, prompt,
field list, honeypot/timing gate, retention) is in DESIGN §2 and still
stands — §2 below is only about how it plugs into your service, which is
where the design doc's own claim needs a correction.

## 2. Correction to DESIGN §2's premise — read this before estimating size

DESIGN §2 (07-26) says tools-api mounts at `/api/v1/tools/<tool>`, "a
namespace built for MULTIPLE tools." **Checked against the live code
2026-08-04/05, and that's not what's there today:**

```go
// internal/tools-api/api/server.go:32
apiGroup := r.Group("/api/v1/tools/gauntlet")   // one group, hardcoded
```

`config.Config` is gauntlet-shaped too — one `AnthropicAPIKey`, one `Model`
(env `GAUNTLET_MODEL`), one rate-limit band (`RATE_LIMIT_RPS`/`_BURST`)
applied to the whole service. DESIGN §2 itself flagged this half of the
gap ("`GAUNTLET_MODEL` should generalise... as part of this") but implied
the routing layer was already generic. It isn't. **This is a genuine
restructure of `NewRouter`, not "add handlers under the existing multi-tool
router."** Sizing it as the latter will run over.

## 3. Concrete plan, mapped onto your actual files

Everything below names a real file in `internal/tools-api/` as it exists
today, so this can be checked against the tree rather than taken on trust.

**Reusable as-is, no changes needed:**
- `middleware.CORSMiddleware` (`middleware/cors.go`) — already resolves
  any `sites.domain WHERE status='deployed'`, robot-hands.com included, no
  gauntlet-specific assumption in it.
- `clientip.From` — already fixed for Caddy (the `bugs_closed/139` work).
- `httperr.JSONError`, `middleware.InputCapMiddleware` — generic.
- `store.ActiveSiteByOrigin` — generic, reusable for the gripper origin too.

**Needs restructuring, not just extending:**
- `api/server.go`: `NewRouter` needs a second `r.Group("/api/v1/tools/gripper")`
  with its own middleware chain — CORS can be shared (it's generic), but
  rate limiting cannot: DESIGN §2 specifies **per-endpoint bands**
  (`/session` 6/h 20/d, `/chat` 60/h 200/d, `/submit` 3/h 10/d), and
  `middleware.RateLimitMiddleware` today is one flat RPS/burst for the
  whole service. **Recommend adopting `platform/httpguard.Limiter` here**
  rather than writing a second bespoke limiter — it's built, tested,
  council-approved, and has exactly zero consumers today (same
  `features_open/024` A3 as `ClientIP`, which you already adopted). This
  would be its first real one. If you'd rather not take that dependency
  right now, a second flat limiter with tighter bands is a smaller diff
  and not a blocker — flagging the reuse because the shared package exists
  and is idle, not insisting on it.
- `config.Config`: needs a second Anthropic key/model pair. DESIGN §2's
  original constraint stands — **the gripper key must be its own
  spend-capped key, never the gauntlet debate engine's** (owner already
  issued one 07-27 for the island-shaped design; it may need re-scoping to
  this shape — owner action, not yours or ours to decide). Suggest
  `GRIPPER_ANTHROPIC_API_KEY` / `GRIPPER_MODEL` (default `claude-haiku-4-5`
  per DESIGN §2's chat-call spec) alongside the existing gauntlet ones,
  loaded the same fail-loud way `config.Load()` already does.
- New handler files, mirroring the existing one-file-per-endpoint pattern
  (`handlers/round.go`, `handlers/position.go`, …): `handlers/gripper_session.go`,
  `handlers/gripper_chat.go`, `handlers/gripper_submit.go`,
  `handlers/gripper_requests.go`. Full request/response shapes for
  `/session`, `/chat`, `/submit` are in DESIGN §2 "API" — unchanged by this
  proposal.
- New store file `store/gripper.go` for the two new tables (below),
  following `store/rounds.go`'s pattern: typed structs, guarded status
  transitions (`UPDATE … WHERE status=$expected`, not blind overwrites —
  DESIGN's own store spec already says this).

## 4. `GET /requests` — the contract is already live, verified from the caller's own code

Not a proposal — this is what `pull_report_requests` already sends and
expects, read from `platform/orchestration/actions/report_request_pull_action.go`
rather than restated from the design doc (which could have drifted; it
hadn't, but check the source, not the summary):

- Request: `GET {base_url}/requests?since=<RFC3339>` (param omitted on the
  first pull), header `X-Internal-Key: <pull_key>`.
- Response: **NDJSON**, one object per line —
  `{"id":"<uuid>","host":"<domain>","submitted_at":"<RFC3339>","spec":{...}}`
  — terminated by one `{"_meta":...}` line. The cluster parser
  (`pulledReportRequest` struct) ignores any line where `Meta != nil` for
  dedup purposes, so the trailing `_meta` line is required, not optional
  decoration — a stream that omits it still works today by accident (the
  loop just reads until EOF) but matching the contract exactly costs
  nothing and removes a future footgun.
- **`X-Internal-Key` is a new auth pattern for this service.** None of the
  four existing handlers (round/position/defend/publish) use it — they're
  all CORS-origin-gated, public-facing. `/requests` is the one endpoint
  that must NOT go through `CORSMiddleware` (no browser calls it; the
  cluster does, from inside the mesh) — needs its own auth check against
  `sites.deploy_config->'report_island'->>'pull_key'`, keyed by the same
  column the cluster action already reads.
- **No visitor email in this payload, by design** (DESIGN §5.1, PII stays
  on your side) — the `spec` object is the chat-collected physics inputs
  only.

## 5. Completion signalling and where `platform/mailer` plugs in

Per DESIGN §5.2, resolved in the cluster's favour: **your service polls
the site's own status sidecar**, it is not called back. One URL per
request: `https://robot-hands.com/reports/<uuid>.json` (2min, then 5min×1h,
then 15min; cache-buster; 15s timeout) — the cluster's `create_report_page`
/ `emit_report_status_files` actions are already live and write this file
via git commit on success or failure (verified 07-31, chart-fix session:
`emit_report_status_files` returns `{"reports/<id>.json": {...}}`, and the
ready sidecar is committed strictly after the page so "ready" can never be
polled true before the artefact exists).

`{"status":"ready","url":...}` → send the email → mark `emailed`.
`{"status":"failed"}`, or nothing by `expires_at` (24h) → apology.

**This poller is the mailer's actual first call site.** Concretely:

```go
import "github.com/gqls/agentchassis/platform/mailer"

cfg, err := mailer.FromEnv("GRIPPER_SMTP")   // reads GRIPPER_SMTP_HOST/_PORT/_USER/_PASS/_FROM/_FROM_NAME/_REPLY_TO
// GRIPPER_SMTP_HOST and GRIPPER_SMTP_FROM are the only two FromEnv requires;
// From is already decided — owner ruling 07-24: robot-hands@contactforsales.com
sender, err := mailer.New(cfg)
err = sender.Send(ctx, mailer.Message{To: []string{req.Email}, Subject: "...", Text: "...", HTML: "..."})
```

`platform/mailer` needs nothing further from either of us — it's built,
tested, council-approved (`6db59c8b`), and this is the shape its own
package doc names as its second intended caller. The poller itself (60s
tick, background goroutine — `cmd/tools-api/main.go` today only starts the
gin server, so this is a new addition there) and the retry/apology logic
around it are new code, spec in DESIGN §2's "poller" line.

## 6. Database

Two new tables, `chat_sessions` and `report_requests`, per DESIGN §2's
schema — but **in tools-api's own Postgres** (the one `DATABASE_URL` in
`deployments/kustomize/services/tools-api/base/deployment.yaml` already
points at, where `gauntlet_rounds` lives), not the separate `gripper_intake`
DB the original pre-correction design specced. Table names should avoid
colliding with anything `gauntlet_*`-shaped — `gripper_chat_sessions` /
`gripper_report_requests` reads clearly against `gauntlet_rounds` in the
same schema.

**One thing worth naming rather than silently repeating:** there is no
migration file or `schema.sql`/`go:embed` anywhere in `internal/tools-api/`
or `cmd/tools-api/` for the *existing* `gauntlet_rounds` table — checked,
not assumed (`find … -iname "*.sql"` and `grep go:embed` both empty). It
was evidently applied by hand directly against tools-api's Postgres. The
new tables will need the same treatment; flagging the absence of a tracked
schema for the service as a pre-existing gap, not something this proposal
should silently fix by inventing a new convention on top of it.

## 7. Governance — this is architecture/council-scope, not a quiet feature commit

Per this repo's 2026-07-28/29 platform-seam rulings: a new tool onboarded
to what has been *documented* as a shared multi-tool namespace, plus two
new DB tables, plus the first real wiring of two idle shared packages
(`platform/mailer`, possibly `httpguard.Limiter`), is exactly the shape
that section exists for — additive and small per file, but the blast
radius is "every future tool this pattern gets copied for." **Register it
in the concept register in the same commit that ships it** (the seam and
its landmine — e.g. the `X-Internal-Key`-vs-CORS routing split above is
exactly the kind of thing a future third tool will get wrong if it isn't
written down), and submit to the council gate before or alongside the
commit — same mechanics as your `ClientIP` round (`e053fac4`, approved
round 2).

## 8. What only the owner can supply

- ~~A spend-capped Anthropic key scoped to this route group~~ **DONE 08-10.**
  A fresh key was issued specifically for this route group (superseding the
  07-27 one, which was scoped to the pre-correction island shape). Stored
  locally at `/home/ant/.config/anthropic/gripper-dossier-api-key` as a
  dotenv-style line — `GRIPPER_ANTHROPIC_API_KEY=<value>`, not a bare key —
  permissions tightened `664`→`600` (was group/world-readable), and
  **verified live** via the free `count_tokens` endpoint (`{"input_tokens":8}`,
  no billing incurred). **Whoever builds the route group**: this is a local
  dev-box file, not a deploy artefact — when the image actually needs it,
  add it as a new key on the `tools-api-secret` k8s Secret (alongside the
  existing `ANTHROPIC_API_KEY`/`DATABASE_URL`), never commit it to the repo.
- ~~SMTP credentials~~ **DONE 08-15.** Supplied via cPanel webmail (as
  flagged 08-09). Stored locally at `/home/ant/.config/gripper-dossier/smtp.env`,
  same dotenv-style shape as the Anthropic key file, permissions fixed
  664→600 on write (the Write tool creates group/world-readable by default —
  same gap as the Anthropic key had, fixed the same way, immediately this
  time rather than found later). **Verified live**: `smtplib.SMTP_SSL` on
  port 465, `AUTH` only, no message sent — real host, real account,
  authenticates. Port 465 matches `mailer.UsesImplicitTLS`'s own
  `port=="465"` branch, so no surprise at the TLS-mode fork when this is
  wired in. Values: `GRIPPER_SMTP_HOST=mail.contactforsales.com`,
  `_PORT=465`, `_USER=robot-hands@contactforsales.com`, `_FROM` = same
  address (owner ruling 07-24) — `_PASS` deliberately not repeated here or
  anywhere else git-tracked; read it from the file. Same deploy note as the
  Anthropic key: local dev-box file, not an artefact — becomes k8s Secret
  keys on `tools-api-secret` when the route group actually ships.

Both credentials are now issued, stored, and verified live. Neither is
wired into anything yet — that waits on the route group itself, still
sitting with the gauntlet lane.

## 9. Suggested sequencing

Your own tools-api history (`d4b68e2e0`, `b51fb30bf`, `208a411c1`, …) is
staged commits under one correlation, s3/s4/s5/s6 — reusing that shape
here seems like the natural fit: router restructure + config extension →
`/requests` + DB tables → `/session`+`/chat` → `/submit`+poller+mailer →
council submission. Not prescriptive — you know your own service's
commit rhythm better than we do; naming it only so this doesn't read as
"build it all at once."

---

**Everything not covered above — chat prompt contract, field list,
honeypot/timing gate shape, retention policy, the site-side widget and
`/gripper-report/` page — is unchanged from `DESIGN_2026-07-24_gripper_dossier_pilot.md`
§2 and doesn't need re-litigating here.**
