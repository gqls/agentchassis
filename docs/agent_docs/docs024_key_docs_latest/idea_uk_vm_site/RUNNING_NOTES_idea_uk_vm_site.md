# RUNNING NOTES — idea.uk VM site

Chronological record of the conversation, what was discovered, and what was decided.
Append; do not rewrite history. Companions: `PLAN_idea_uk_vm_site.md`, `RUNBOOK_idea_uk_vm_site.md`.

---

## A — 2026-07-14 · Opening brief

Owner asked for two threads:
1. Publish idea.uk's chassis-built static pages to the **Hetzner VM** instead of S3/B2, so it is one
   complete site that still includes the paid idea tool.
2. Continue `HANDOFF_claude_code_continue.md` (three uncomposed pages; the re-aimed CSS fixer; a
   slice-4a re-seed check; contact-form spam).

Later in the session the owner added a third: record the conversation and choices in a fresh
directory (this one), and gave two constraints that shaped everything:
- **"There will be several thousand domain names so individual repos will be ungainly."**
- The sites repo already holds **domains as subdirectories**, with a **locally hosted GitHub Action**
  posting to S3.

---

## B — What the codebase actually does (established by reading, not assumption)

**Deploy pipeline.** Chassis action `git_commit` → Kafka `system.adapter.git.requests` → git-adapter →
GitHub Git Data API → `gqls/sites`, domain folders at the repo root → self-hosted Actions runner →
`b2 sync --delete` → `b2://portfolio-sites/<domain>` → Cloudflare purge. **"Commit is deploy."** The
B2 step lives in the *external* repo's workflow, not in this codebase.

**The two halves of idea.uk have never met.** DNS (Cloudflare) points at the VM; nginx there proxies
**everything** to the Go tool on `127.0.0.1:8080`, so the tool's single embedded `page.html` is the
live homepage. The chassis's nine-page build goes to B2, **where nobody sees it** — an elaborate
staging copy. This is why "publish to the VM" is the whole ask.

**The tool is not part of the chassis.** It is a standalone stdlib-only Go module (`module idea`, zero
dependencies) parked at `docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/`. It calls
`api.anthropic.com` directly (`engine.go:198`), hand-rolls Stripe with no SDK (`billing.go:41`), and
stores orders in a **JSON file** — `/var/lib/idea/orders.json`, **no database**. It has no CI: it ships
by building a linux/amd64 binary locally, `scp`-ing it, and restarting systemd.

**The per-site deploy target was designed, written, and never wired.** `sites.github_repo` exists as a
column. `resolveGitRepoName` (`helpers.go:206`) has exactly the right logic — explicit config → the
site's own repo → default `"sites"`. **Nothing calls it.** Four wires are missing (RUNBOOK §2a).
The concept register already framed the intent: *"`sites.github_repo` selects the target."*

---

## C — 🔴 Unplanned finding: live credentials on a public repo

Not part of either thread; found while mapping the tool. `gqls/agentchassis` is **public**, and
`idea.env.example` has been on `origin/main` since **2026-06-04** (~6 weeks).

The `.example` name is misleading, so the keys were judged **by length, not by name**:

| Key | Length | Real length | Verdict |
|---|---|---|---|
| `SMTP_USER` (AWS SES) | 20 | 20 (`AKIA`+16) | 🔴 **REAL** |
| `SMTP_PASS` (AWS SES) | 44 | 44 | 🔴 **REAL** |
| `INTERNAL_API_KEY` | 64 | 64 (`rand -hex 32`) | 🔴 **REAL** |
| `ANTHROPIC_API_KEY` | 25 | ≈108 | ✅ truncated placeholder |
| `STRIPE_SECRET_KEY` | 16 | ≈107 | ✅ truncated placeholder |
| `STRIPE_WEBHOOK_SECRET` | 14 | ≈38 | ✅ truncated placeholder |

**The money path is safe** — Stripe and Anthropic were never exposed. What *is* exposed: the ability
to send email as idea.uk (SES reputation, phishing, the AWS bill), and `INTERNAL_API_KEY`, which gates
`/confirm` `/decline` `/approve` `/internal/run` — order approval and Claude-billing runs on the live
earning service.

**DECISION:** scrub the repo now; **owner rotates**. Rotation is what actually closes it — the values
are in pushed history, so deleting the file is necessary but not sufficient.

---

## D — DECISION: keep the static build; the VM is a second *sink*, not a second renderer

Owner asked directly: *"Do we need a static site for the dynamic sites — would it be better or worse to
host the whole site behind the VM nginx?"* Two questions, answered separately.

**Rendering: keep it static. Rejected dynamic rendering on the box.** The chassis *is* a static-site
generator — rendering happens in Kubernetes out of `clients_db`, and the artefact is HTML. The VM has
no route to that database and the tool binary has no DB driver at all. Serving pages dynamically from
the box would mean either giving every box a path into the cluster database (across a fleet: a coupling
and security problem) or re-implementing rendering locally against a copy of the data — which is the
static artefact with extra steps. It would also throw away three free properties:
- **Availability** — static files survive a tool crash. *Today a tool crash takes the whole site down*,
  because nginx proxies everything to `:8080`. The cutover is a strict improvement here.
- **Cacheability** — static is Cloudflare-cacheable; dynamic bypasses.
- **Fleet uniformity** — one artefact, two sinks. Two rendering paths would drift.

**Origin: yes, put the whole site behind the VM nginx.** Same-origin means the funnel, forms and
cookies simply work (the tool carries an `ALLOWED_ORIGINS` CORS setting precisely because it
anticipated a split — we get to not need it). One cert, one DNS record, no Cloudflare Worker stitching
two origins by path. The box is already running and already paid for.

**The cost, recorded honestly.** idea.uk's marketing pages now depend on one Hetzner box rather than on
B2's effectively-infinite availability, and the box carries an ops burden. Acceptable *because* nginx
serves the files with no application in the request path — the box must be genuinely down, not merely
the tool. And it is precisely why **this stays the exception**: B2-static remains the default class for
the thousands. Do not scale a box-per-site.

Resulting fleet model:

| Class | Shape | For |
|---|---|---|
| **A — static → B2** *(default)* | serverless, no box | the thousands |
| **B — static → VM + backend** | same artefact, one nginx, backend paths proxied | the handful that sell something |
| **C — dynamic rendering on the box** | ❌ rejected | nobody |

---

## E — DECISION: pull, not push

How the built files reach the box. Owner chose **pull: each box syncs itself** — a systemd timer that
sparse-checkouts its own domain folder from the shared repo into `/var/www/<domain>`.

Rejected the alternative (extend the existing `vm-sites` rsync Action with a domain→host map): it holds
**one SSH deploy key on the self-hosted runner, authorised on every box** — compromise the runner and
you reach the fleet. Pull inverts the blast radius: a compromised box holds a **read-only** repo deploy
key and cannot reach its siblings. A `--filter=blob:none` + sparse checkout keeps a
thousands-of-domains repo cheap, since each box fetches only its own folder.

Accepted cost: deploy latency = the timer interval, not instant-on-commit; and it is a new mechanism
rather than the one already proven for relojistas.com.

**⚠️ Hazard this created, caught before it bit.** `gqls/vm-sites`'s existing Action rsyncs *every*
changed domain folder to a single `VM_HOST` secret — **relojistas' box, `167.233.33.159`**. The moment
idea.uk lands in that repo, that Action would push idea.uk onto the **wrong machine**. Before pointing
`sites.github_repo` at `vm-sites`, the Action must gain an explicit allowlist (`deploy-targets.json`,
hosts as data not secrets) from which idea.uk is simply absent. RUNBOOK §2b.

---

## F — The site's real state (read from the DB, not assumed)

Nine pages. Six deployed; three `planned` with **zero sections** — exactly the trio the handoff
described, and the reason their nav links 404:

```
/index.html                       landing        deployed   6 sections
/tools.html                       content        deployed   3
/report.html                      landing        deployed   4
/guides/index.html                section-index  planned    0   ← 404
/news/index.html                  section-index  planned    0   ← 404
/about.html                       content        deployed   4
/contact.html                     content        deployed   3
/tools/audience-check/index.html  tool           planned    0   ← 404
/privacy.html                     content        deployed   1
```

This settles the path-collision question the cutover depends on:
- **`/privacy` collides.** Static generates `/privacy.html`; the tool serves `/privacy`. `try_files
  $uri $uri.html` hands it to the static page unless a `location = /privacy` block wins. **Open
  decision** (RUNBOOK §3b); default is that the tool keeps it, alongside its other legal pages.
- **`/terms` and `/refund-policy` do not collide** — the static build has no such pages. Tool keeps them.
- **`/tools/audience-check/` does not collide** with the tool's `/audience-check` endpoint. Different
  paths.

**The cutover's real risk, quantified.** The tool serves **16 routes** (`service.go:527-543`). The
pre-existing cutover runbook warns that "reserved-path completeness is the whole risk" — and then its
own example nginx block lists **7**. A missing `/audience-check` kills the free taster; missing `/op`
`/confirm` `/approve` `/decline` kills the operator flow. The full list is in RUNBOOK §3b.

---

## G — The spam handoff is built on a false premise

`HANDOFF_spam_and_ip_blocklist.md` is wrong on three counts, and following it would burn a session:

1. It says the `/request` handler is **"the chassis Go process"**. It is not — it is the standalone
   `idea` binary, unrelated to the chassis.
2. It says the submissions are *"almost certainly into a store we can query"*, and instructs the next
   session to hunt harder in `clients_db`, explicitly dismissing an earlier 0-row result as *"a search
   miss, not a 'no data' result."* **The 0 rows were correct.** idea.uk has no database; orders are a
   JSON file (`store.go:3-5`, `setup.sh:150`). `spam_read.sql` is therefore **void** — it should be
   discarded, not extended with more `ILIKE` needles.
3. It describes nginx as already serving static pages and proxying reserved paths. That is the *future*
   state — the cutover has not happened.

What it gets right: *"a honeypot + timing check usually outperforms an IP list"* — spammers rotate IPs.

**Two facts that constrain any fix.**
- The `Order` struct (`store.go:17-30`) has **no IP field**. The existing spam rows therefore **cannot
  seed a blocklist retroactively**. The only historical IP source is the nginx access log.
- idea.uk sits behind Cloudflare, but `setup.sh` never sets `set_real_ip_from` / `real_ip_header
  CF-Connecting-IP`. So nginx sees a **Cloudflare edge IP**: the existing `limit_req` zone buckets all
  Cloudflare traffic as one, and any nginx `geo` deny or fail2ban jail would ban **Cloudflare, not the
  spammer**. This exact trap is documented in the traffic_probe runbook for the *other* box and was
  never back-ported. It gates every IP-blocking approach.

The ingredients already exist, merely unwired: a per-IP sliding-window `rateLimiter` (3/hr + 20/day) and
a `clientIP()` extractor, both in `audience_check.go`, both used **only** by the free taster.
`handleRequest` calls neither, and discards `ParseForm`'s error.

---

## H — Out of scope, but found: the fleet's contact forms post into a void

`apply_gap_plan_action.go:465` emits a `contact-form` section whose stored HTML is
`<form class="contact-form" action="/contact" method="POST">`. The generated sites are **static**, so
`POST /contact` resolves to nothing and **every submission is silently lost, fleet-wide**. idea.uk has a
deployed `/contact.html`. This is a *dead form* problem, not a spam problem. It wants its own thread.
Whatever contact backend is eventually built should be born with the honeypot and rate limit that
Phase 4 retrofits onto `/request`.

---

## I — 2026-07-14/15 · Phase 1 executed, and the handoff's "safe re-plan" premise proved FALSE

Ran the handoff's prescribed fix — emit a `needs_site_plan`, re-run `build-site-planner`, let its
union preserve the built pages while composing the three empty ones. It did not work as promised, in
both directions. **This is a fleet-wide landmine, not an idea.uk quirk.**

**What the re-plan actually did** (plan `ff03bdef`, superseding `32be2797`):
- ✅ composed `guides-index` and `news-index` (2 sections each) — the LLM re-proposed them *with*
  sections, so these built and deployed cleanly.
- ❌ left `tool-audience-check` empty — see below, it structurally cannot be composed this way.
- ⚠️ **invented 10 pages nobody asked for** — 5 more tools + 5 blog posts, all uncomposed — turning 3
  empty pages into 11.
- ⚠️ **regressed 4 built pages** it was supposed to preserve: `index` (dropped info-card-grid),
  `about` (dropped hero-about + info-card-grid, swapped in a generic hero), `contact` (dropped
  hero-contact for a generic hero), `report` (dropped generic-text-block + info-card-grid), plus nav
  churn. `about` and `contact` actually **rebuilt and deployed** the regressed versions to B2 before I
  caught it (proven via `page_components`: both rendered a generic `hero`, not their specific one).

**Why — the two-way failure of the "union is safe" claim:**
1. `normaliseRealisedToPlanPage`'s union (`v3_site_actions.go:4630`) only protects pages the LLM does
   **not** re-propose. For a page the LLM re-proposes, the LLM's section list wins and the realised
   composition is overwritten. `reconcilePlanWithRealised` force-preserves **only adoption-locked**
   pages — and idea.uk's are not adopted. So "the union preserves the built pages" is false whenever
   the LLM re-proposes them, which it did for the four flat pages.
2. The union carries `sections: []` **faithfully**. A catalogued-but-uncomposed page is therefore
   preserved *as empty* — a re-plan can never fill it. `guides-index`/`news-index` were composed only
   because the LLM happened to re-propose them with sections; `tool-audience-check` was unioned empty
   and stayed empty.

**Recovery (all reversible, nothing live-visible — DNS still points at the VM, so B2 is unseen):**
- Paused the queue: 23 items `triaged`→`detected` (ids in `sql/paused_item_ids.txt`).
- Backed up idea.uk's pages/plan/work-item state to `_ideauk_bak_20260714_*`.
- `sql/p1_02_replan_rollback.sql`: restored the 6 built pages' sections + `pages.sections` + nav
  metadata from the old plan (verified **0 section diffs**), deleted the 10 invented pages and their
  work items, kept the two wins, re-triaged only what genuinely needed building.
- `sql/p1_03_rebuild_about_contact.sql`: `about` + `contact` had deployed the regressed artefact and
  their build items were already terminal, so the rollback's re-triage couldn't catch them — emitted
  fresh `needs_page` rebuilds from the restored composition.

**Net after recovery:** 9 pages. `guides-index` + `news-index` composed and deployed (2 of 3 targets
done). `about` + `contact` rebuilding from restored composition. `index`/`tools`/`report`/`privacy`
untouched-correct. `tool-audience-check` still empty, held at `detected`.

**Correction to the doctrine:** "re-run build-site-planner to compose missing pages" is unsafe on any
site whose built pages are not adoption-locked. To compose a *single* named page without disturbing
the rest, the real writer is `write_site_plan_action.go` (only INSERT into `site_plan_sections`), and
the intended single-page retrigger is the `sectionless_pages` discovery check → `needs_content_page` →
page-build-handler sibling fallback (`discovery_checks/check_sectionless_pages.go`). **But that check
only fires for a sectionless page with a same-role sibling that has sections** — its self-healing loop
is scoped to what the sibling fallback can repair.

## J — 2026-07-15 · `tool-audience-check` has no composition route, and may not need one

`tool-audience-check` is the **only `role='tool'` page**, so it has no same-role sibling — the
`sectionless_pages` check explicitly excludes this case ("cannot be auto-repaired from a sibling").
So there is no automated route to compose it.

Investigating what it is *for*: it is **not in any nav** (`in_header=f, in_footer=f`). The only thing
linking to `/tools/audience-check/index.html` is the **`tool-list` card** on the `index` and `tools`
pages (`href="/tools/audience-check/index.html"`). Its nav_label is "Free Audience Check — idea.uk".

So this is not really a "compose the page" task — it is a **funnel decision**. After cutover,
`/audience-check` is the **live tool** (proxied to the VM). Two clean resolutions (RUNBOOK §3b / owner
decision):
- **A — retarget + retire:** point the `tool-list` card straight at `/audience-check` (the working
  tool) and drop the redundant static page. One click to the real thing; no interstitial; 404 gone.
  Cost: a visual jump from chassis styling into the tool's own embedded page.
- **B — compose an interstitial:** build `/tools/audience-check/` as a chassis-styled landing page
  (hero + explanation + CTA → `/audience-check`). Keeps the design consistent up to the tool
  interaction. Cost: a page that structurally can't auto-compose, so it needs a driven build or a
  role change to give it a sibling.

## K — 2026-07-15 · Decisions taken; Phase 1 essentially done

**`tool-audience-check` → option A (retarget + retire).** Owner chose: the tool-list card links
straight to the live tool; no interstitial. Implemented as a **pointer page** pattern (worth reusing
for any VM-hosted tool surfaced in a chassis tool-list):
- The card href is `tool-audience-check.url` (tool-list resolves `query.pages_where_type:tool` →
  `resolvePagesWhereType`, `queryresolve.go:155`, which uses each tool page's `url`). Repointed
  `url` → `/audience-check`.
- "Retire the static page" = never emit an artefact + stop reconcile churn. Set
  `build_status='deployed'` pinned to the current plan (`built_from_plan_version`), which is the exact
  condition `decideEmit` (`reconcile_site_plan_action.go:293`) needs to return `skip_built`. Truthful:
  the tool it points to IS deployed (on the box). No sections → nothing assembles a file. Even a stray
  file would lose to nginx's exact-match `/audience-check` proxy.
- Re-rendered `index` + `tools` so their baked-in card href updates from the old static path.
- `sql/p1_04_audience_check_pointer.sql`.

**Phase 1 outcome:** all three original targets resolved. `guides-index` + `news-index` composed and
deployed; `tool-audience-check` resolved as a pointer to the live tool; the 4 regressed pages restored
(`about` + `contact` rebuilt with their correct heroes, verified via `page_components`). Site is back
to a coherent 9 pages, no invented pages, nothing live-visible touched.

**Residual (pre-existing, not caused here):** `contact` still lacks `contact-info` — a
`needs_section_data` / `needs_human_review` for a missing *business contact email*. Owner decision:
what public contact email should the static `/contact.html` show? (The tool uses
`idea-uk@leopardess.uk`.) Separately, `/contact.html`'s form posts to `/contact`, which is a **dead
form** after cutover (the tool serves no `/contact`); see the out-of-scope note in §H — the static
contact page needs either a working backend or its form replaced with a mailto/CTA to the tool's
`/request`.

## L — 2026-07-15 · Infra observations while watching the card re-render

- **Build path is churny but self-healing.** The `index` re-render hit "Claim timed out — handler pod
  likely died" (attempt 1/3) and reverted to `triaged`; `page-build-handler` pods are short-lived and
  one died mid-build. The dispatch loop (4 replicas, healthy) retries. This is the known claim-timeout
  hygiene item, not a hard failure — the card flip completes via retry.
- **Self-hosted GitHub Actions runner: 1 of 2 replicas crash-looping (`lhg9l`, 4906 restarts/18d).**
  Failure is a node container-runtime bug (`runc`: `expected cgroupsPath to be of format
  "slice:prefix:name" for systemd cgroups`), exitCode 128 StartError — the container never starts.
  The other replica (`5pqdv`) is healthy, so **B2 sync still works** (sites have been deploying). Net:
  **lost redundancy + constant restart noise**, and a latent single point of failure — if `5pqdv`
  dies, `lhg9l` cannot take over and fleet B2 deploys stop. Reschedule `lhg9l` off the bad node / fix
  that node's cgroup driver. Tangential to this workstream, but it **reinforces the migration
  rationale**: idea.uk moving to VM-pull leaves this fragile B2-via-crashy-runner path entirely.

## M — 2026-07-16 · Fresh chassis (v1.0.1123); /request hardened; contact email; error notes

- **Fresh chassis v1.0.1123 shipped** — confirmed the Phase 2 per-site-target wiring is in the tree
  the image built from (all three greps = 1). So the deploy-target code is now LIVE; activation still
  gated on the vm-sites Action guard (RUNBOOK §2b) + the `UPDATE sites` (§2c).
- **`/request` hardened** (task complete) — in the tool source (`golang_files/`): honeypot
  (`company_url`), timing gate (`_elapsed`<2500ms, skew-free client delta, fail-open for no-JS),
  intake rate limiter (5/hr+15/day), `mail.ParseAddress` + length caps, IP/UA captured on the Order,
  method guard (was missing). `requestAccepted` extracted so silent drops are indistinguishable from a
  real accept. `request_hardening_test.go` — 6 subtests PASS. `go build`(linux/amd64)+`vet` clean.
  Pre-existing unrelated red test `TestReviewBeforePayFlow` (fails on clean tree too). NOT deployed —
  owner builds+scp+restart.
- **Contact email = idea.uk@contactforsales.com.** First attempt (`sql/p1_05`) set only
  `site_specs.identity.email` (what the section-data resolver renders) → the contact rebuild FAILED
  `validate_page_content` with `invalid_email`: the validator's canonical is `loadSiteContactEmail`
  (`validate_page_content.go:735`), which COALESCEs `sites.email` FIRST — still the old
  idea-uk@leopardess.uk. Render and canonical disagreed. `sql/p1_06` set `sites.email` (+ nested
  `identity.contact.email`) to the new address; canonical now matches; contact rebuild re-queued.
  *Lesson:* two different email sources — the resolver reads `identity.email`, the validator reads
  `sites.email` first. Align both. (The tool's own OPERATOR_EMAIL in its env is a third, separate
  surface — unchanged.)
- **Three infra errors logged** to `aaa_fails_to_mend/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md`:
  (A) crash-looping runner replica (§L), (B) fleet-wide dead `/contact` form (§H), (C) claim-timeout
  churn (§L). The re-plan landmine handoff is already there as `001`.
- **Deliverables added this session:** `SUMMARY_idea_uk_vm_site.md` (presentable status) and
  `HANDOFF_RESUME_idea_uk_vm_site.md` (fresh-chat entry point).

## Open decisions

- **`/privacy`** — tool or static site? (RUNBOOK §3b; default: tool.)
- **`/contact.html`** — its form posts to a dead `/contact`; repoint to the tool's `/request` or a
  `mailto:idea.uk@contactforsales.com`? (§K, and `aaa_fails_to_mend/006` §B.)
- **Cloudflare proxied (orange) or DNS-only (grey)?** Unverifiable from the repo; decides whether the
  real-IP problem is live and whether Cloudflare WAF/Turnstile is reachable as the blocking layer.
- **Does relojistas.com migrate from push to pull too**, or do the two mechanisms coexist? (Coexist is
  fine and is what §2b's allowlist enables.)
