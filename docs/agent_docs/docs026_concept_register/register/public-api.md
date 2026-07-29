# Register — public-api

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit
U01. **Plus 2 post-freeze additions (PUB-002, PUB-003) made 2026-07-29** for the
two shared packages the consolidation programme built on 07-28 — both
council-approved, both with zero importers at the time of writing.

**Stage-2 duplicate resolution (2026-07-14):** PUB-001 is the same underlying
plan as `admin-dashboard-and-api.md`'s ADM-007 + ADM-008, extracted under a
distinct category tag (`new:public-api`) because P2 (its source doc) is a
later/live version of the plan that ADM-007/008's source (`007b_public_api_plan_v2.md`)
was an earlier snapshot of. Full entry retained below for its distinct source
citation (P2 vs 007b); treat ADM-007 (endpoint plan) + ADM-008 (site_ownership
table) as the canonical pair — this entry is the pointer.

### PUB-001 — Public API plan: site_ownership junction + user-facing build/HITL endpoints (duplicate — see ADM-007 + ADM-008)
- **status:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes — the user-facing half was never built.
- **stage2-verified (2026-07-14):** confirmed unbuilt (0 hits for site_ownership or /api/v1/sites in any .go/.sql file) and confirmed duplicate of ADM-007 + ADM-008 — see those entries for the canonical write-up.
- **what:** `site_ownership` junction table (site/client/user/role) rather than columns on `sites` (shared sites; 15+ FKs untouched); all public queries scope through it. `POST /sites` writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2_public_api_plan.md (full, docs024_key_docs_latest)
- **relations:** duplicate-of: ADM-007 (admin-dashboard-and-api.md, public REST API plan), ADM-008 (admin-dashboard-and-api.md, site_ownership table). admin API (ADM-002); needs_human_review status; build_queue.
- **verify-later:** site_ownership table; which blocks landed

---

*Added 2026-07-28/29 by the consolidation programme (`features_open/024` A2/A3),
post-dating the 2026-07-13 extraction freeze. Both packages were built and
council-approved on 07-28 and neither was registered at the time — the omission
this file's own banner and `bugs_open/106` exist to catch, recorded here rather
than left as folklore.*

### PUB-002 — `platform/httpguard`: inbound-abuse primitives for public HTTP endpoints, with the trusted proxy stated per deployment
- **status:** built, unit-tested, council-approved (`6db59c8b`) — **and called by NOTHING.** Zero importers, re-measured 2026-07-29. Adoption is owner-ruled in remit ("get it adopted", 07-29) but has not happened.
- **status-evidence:** `grep -rn 'agentchassis/platform/httpguard' --include=*.go .` returns only the package itself. `go test ./platform/httpguard/` passes; the `bugs_closed/090` regression test is verified to fail when the original defect is reintroduced.
- **what:** One shared implementation of three things that had diverged across the estate: `ClientIP` (a per-client key the client cannot choose), `Limiter` (in-memory sliding-window, multi-band, returns retry-after — replaces a single token bucket that could express only one ceiling and no retry-after), and `CheckIntake` (honeypot + client-side timing gate, fail-open on missing JS). net/http only, no gin, so a gin service writes a three-line adapter. Written because three per-IP limiters existed and **the weakest guarded the only public endpoint**, and four different CORS postures existed.
- **the seam, and the landmine it exists to close (2026-07-29):** `ClientIP` takes a **required `FrontEnd` argument** naming which headers *this deployment's* proxy is known to **write** (as opposed to merely forward). Pre-declared: `Nginx()`, `CloudflareTunnel()`, `Direct()`. It previously hard-coded nginx's rules — prefer `X-Real-IP`, else the rightmost `X-Forwarded-For` — justified in its own docstring by nginx's `proxy_set_header` behaviour. **That justification is false on the estate's other front-end.** Measured 07-29: Caddy does not set `X-Real-IP` and forwards a client-supplied one verbatim, and it *overwrites* `X-Forwarded-For` with its own peer instead of appending. So on the island both old rules resolve to either user input or a constant, and adopting the old default into `tools-api` would have keyed every visitor on the docker bridge gateway — **83 of 83 rows, one distinct value** — while reading like a fix. The argument is required precisely so that a caller cannot inherit an assumption it never stated.
- **open review question:** the peer gate (`loopback || RFC1918`) is what keeps `CloudflareTunnel()` honest — trusting `CF-Connecting-IP` is trusting Cloudflare to be in front, and the gate is what makes the header revert to being ignored if the origin is ever reachable without the tunnel. Nobody has tested that reversion against a real direct-exposure path, only in unit tests. **The `architecture` seat ruled on this at review (2026-07-29, corr `49392838`, severity low): *"Fine to ship, but the open question should not go stale - the next thing to land against this package should close it, not add a fourth FrontEnd."* Treat that as the package's next obligation, ahead of any new front-end.**
- **council (round 1, 2026-07-29, corr `49392838-5ada-4c8e-baeb-94b01e5855b4`):** **APPROVED** — *"approved with 1 advisory objection(s) — none high-severity"*. 9 seats fired, 8 abstained on relevance, none unreadable. The `architecture` seat returned `ARCHITECTURE_SIGNAL: point_fix`, explicitly settling the venue question the submission raised itself: *"hardening a shared mechanism's own contract before its first real consumer, which is the cheapest point in its life to do it, not a shared-mechanism change smuggled in via a symptom fix."* Two medium objections, both answerable with evidence rather than argument, and both answered: `guardian` asked that the register entry be **confirmed** to ship in the same commit rather than assumed — it did, `31c684124` carries this entry and the code together; `prior_art_librarian` asked that the load-bearing zero-importer claim be re-grepped at merge time rather than trusted from the submission — re-run after the verdict, still zero. `reuse_agent` left a standing question this entry inherits: whether the three divergent per-IP limiters httpguard was built to replace are actually decommissioned. **They are not — that is what adoption means, and it has not happened.**
- **sources:** `platform/httpguard/{clientip,limiter,intake}.go` + `httpguard_test.go`; `features_open/024` A3; `bugs_closed/090` (the proven production incident it hardens against); `bugs_open/139` (the measurement that produced the seam); `016b` §9 *"Who is the client" is decided by the PROXY CHAIN*.
- **relations:** replaces-on-adoption: `internal/tools-api/middleware/ratelimit.go` (token bucket keyed on gin's `c.ClientIP()`), idea.uk's banded limiter + honeypot/timing gate. Adoption into `tools-api` is the **gauntlet_dead_cta** thread's service — see `gauntlet_dead_cta/CONTRIB_2026-07-29_…`. Sibling: PUB-003.

### PUB-003 — `platform/mailer`: the first SMTP sender inside the built code
- **status:** built, unit-tested (8 tests), council-approved (`6db59c8b`) — **and called by NOTHING.** Zero importers, re-measured 2026-07-29.
- **status-evidence:** before it landed, `grep -rn "net/smtp" --include=*.go platform/ internal/ cmd/` returned **nothing** — there was no mailer anywhere in the code we build and deploy. `send_notification` (`basic_actions.go:134`) is not email; it produces a Kafka message.
- **what:** SMTP send promoted into `platform/` so that "we email you a link" journeys stop depending on code outside the build. The only working sender was in idea.uk's VM app in the docs tree — outside `go build`, untested by CI, undeployable by the image pipeline.
- **landmine carried from the operational record:** see `email-infrastructure.md` — cloud boxes generally cannot use outbound 25/465, Go's `smtp.SendMail` does STARTTLS not implicit TLS, and shared-host relays content-filter legitimate transactional mail. Those constraints decide deployment, not this package.
- **sources:** `platform/mailer/mailer.go` + `mailer_test.go`; `features_open/024` A2.
- **relations:** operational constraints: `email-infrastructure.md`. First queued consumer: the gripper dossier's public half (`robot_hands_gripper_dossier/`). Sibling: PUB-002.
