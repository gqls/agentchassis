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

## N — 2026-07-16 (later) · Phase 2 EXECUTED: guard, runner, activation, seed

- **Secrets guard now tracked** (30b6aa30d). The handoff's "guard + hook are UNTRACKED" was caused by
  bulk doc commits sweeping `scripts/check-secrets.sh` and `.githooks/` into `.gitignore` — removed
  those two lines and committed guard + hook.
- **§2b executed** — `deploy-targets.json` (`{"relojistas.com": "167.233.33.159"}`) + allowlist
  workflow pushed to `gqls/vm-sites` (d0fb7a1). Hosts are data; `VM_HOST` secret no longer read;
  unmapped domains are skipped. Simulated all four domain scenarios locally before pushing.
- **🔴 Discovery: the vm-sites Action had NEVER run.** Two runs ever, both stuck `queued`: no
  self-hosted runner was registered on the repo (the cluster runner serves only `gqls/sites`), AND
  the shared runner image (`aqls/github-actions-runner`) had **no ssh/ssh-keyscan/rsync** — the
  'Set up SSH' step dies exit-127, invisibly, because its stderr goes to /dev/null. "Commit is
  deploy" was only ever true for B2. relojistas' box content must have been hand-rsynced at go-live.
- **Defused before fixing**: the stale queued run (11:44, old unguarded workflow) would have
  `rsync --delete`'d the repo's then-1-file `relojistas.com/` over the live webroot the moment a
  runner appeared → **cancelled it** (restorable: `gh run rerun 29495521330`; now superseded — the
  relojistas session independently mirrored the live webroot into the repo, 19debed, then deployed).
- **Runner provisioned**: image rebuilt with `openssh-client` + `rsync` as
  `aqls/github-actions-runner:v1.0.1126` (same tag family as the day's chassis build — different
  image repo, no clash); new one-replica deployment `github-actions-runner-vmsites`
  (`deployments/kustomize/services/github-actions-runner-vmsites/`, GITHUB_REPO=vm-sites, same PAT).
  Landed on the healthy node. The `sites` runner was not touched.
- **Verified LIVE, both directions**: run 29499847029 → `Changed domains: idea.uk` →
  `Skipping idea.uk (no mapped host)`; run 29522849364 (relojistas news feed) →
  `Deploying relojistas.com -> 167.233.33.159` → success. Two further green skip-runs followed
  (marker removal, seed).
- **§2c executed**: `UPDATE sites SET github_repo='vm-sites' WHERE domain='idea.uk'` (was NULL;
  `github_branch` already `main`, which matches vm-sites). Rollback = set NULL.
- **Repo seeded** (4cbaf2a): copied `idea.uk/` verbatim from `gqls/sites` master (the built B2
  artefact) into `gqls/vm-sites` — 8 built pages + assets; the pointer page correctly has no
  artefact. Future chassis builds now commit here via the activated per-site target.
- **⚠️ §3b premise correction**: static `terms.html` AND `refund-policy.html` DO exist in the
  artefact (the RUNBOOK said they didn't) and the built footers link all three legal pages **with
  the `.html` extension**, so exact-match proxy locations alone won't stop the static copies being
  user-visible. Recommend at cutover: `location = /terms.html { return 301 /terms; }` (same for
  refund-policy + privacy) so the tool's versions — the terms the buyer agreed to — stay canonical.
- Housekeeping: commit c9eafa3c8 (overlay tag bump) accidentally swept in a pre-staged empty doc
  file from another session (`fixloop_eg_dartsonline/SUMMARY_of_the_json_leak.md`) — harmless.
  Concurrent sessions bulk-committed the dockerfile edit (6880c669e) and kustomize base (87d13b864).

## O — 2026-07-17 · Phase 0 CLOSED: credentials rotated by the owner

Owner initially believed the leak was resolved because the *current* `idea.env.example` is clean —
the session demonstrated (lengths/prefixes only, values never printed) that the real values sat in
pushed public history from 2026-06-10/11 to the 2026-07-14 scrub (SMTP_USER 20 chars `AKIA…`,
SMTP_PASS 44 chars, INTERNAL_API_KEY 64-hex), retrievable by anyone via `git log -p`. Also caught
that the scrubbed file's header prematurely claimed "and rotated" — it had NOT been (verified via
IAM key creation date, per the walkthrough).

**Rotation executed 2026-07-17 by the owner:** fresh SES SMTP IAM user created
(`ses-smtp-user.20260717-…`) and sending verified; the old June user **deleted** (not merely
deactivated); `INTERNAL_API_KEY` regenerated; env updated; `idea` restarted healthy. The history
values are now dead. `/op` links: old ones no longer verify; issue fresh on next use. History
rewrite deliberately NOT done — rotation makes it pointless, and a force-push would break clones
and every SHA in the docs. Stale `IAM_USER` comment removed from `idea.env.example` (don't publish
the new user name in a public repo).

## P — 2026-07-17 · REVIEW email spam-blocked: subject carried the whole business description

First real order flow after rotation: the "New report request" email arrived, but the operator
REVIEW email (draft report + /op approve link) was blocked by Clook/MailChannels as "Spam Content".
Cause: `service.go` built the subject as `[idea.uk] REVIEW <id> (<o.Domain>)` — and `Order.Domain`,
despite the name, holds the requester's **free-text business description** (here ~480 chars of
sales-sounding copy → keyword-stuffed subject). The request email survived because its subject is
short (`New report request <id>`); only the REVIEW subject interpolated the full text.

Fix (in the tool source, tested): `subjectSnippet()` — whitespace collapsed, cut ≤60 runes at a word
boundary + ellipsis — applied at the one call site; full text stays in the body.
`TestSubjectSnippet` added (3 subtests PASS); build/vet clean; `TestReviewBeforePayFlow` still the
known pre-existing red. **Inert until the binary ships** — rides the same
build+scp+restart as the pending /request hardening.

Owner-side deliverability (not code, still open): (1) release the blocked email via "Not Spam" —
the order sits `awaiting_review` on the box, nothing lost; (2) whitelist operator mail in Clook
(sender or `[idea.uk]` subject prefix); (3) check SES: the From domain (leopardess.uk) should be a
verified **domain** identity with Easy DKIM + ideally a custom MAIL FROM, else DMARC alignment fails
and buyer-facing report emails will hit other people's filters too — the quarantine listing showing
the `…@eu-west-2.amazonses.com` envelope sender is a hint alignment is missing.

## Q — 2026-07-17 · contact-form → mailto (owner decision) + a stale email caught

Owner chose **"Convert form → mailto"** for the dead contact form. Traced the seam: the fleet
`contact-form` component (`content_components name='contact-form', forked_from IS NULL`) renders
`<form action="{{.form_action}}" method="POST">` — a template variable, filled from the section's
**inline** `page_components.content_data` (`data_path` empty; no `site_specs` aspect carries
`form_action`; `content_hash` NULL; the about/contact recovery rebuild reused this data verbatim,
which is why `#contact` persisted). So the correct per-site lever is that inline value — **the fleet
template is untouched** (the fleet-wide dead-form fix stays its own thread, aaa_fails_to_mend/006 §B).

Also caught while there: the form's `form_description` still named the **old** email
`idea-uk@leopardess.uk` — the `contact-info` block above it was aligned to
`idea.uk@contactforsales.com` in p1_05/p1_06 but this description was missed. Fixed both in one edit.

`sql/p1_07_contact_form_mailto.sql` (idempotent, guarded on `form_action='#contact'`):
- `form_action`: `#contact` → `mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry`
- `form_description`: old email → new. Phone `+44 (0) 7934 524 911` left as-is (no decision to change it).
Applied & verified (BEFORE/AFTER in the run). `rendered_html` deliberately NOT hand-edited — the
renderer regenerates it from this corrected `content_data` (editing the rendered artifact is the
documented revert trap).

**Publish state:** the fix is staged at the source. It reaches the deployed artifact (and gqls/vm-sites)
on the **next contact-page build**. Not forced now because: (a) single-page rebuilds of already-`deployed`
pages bounce straight to `needs_human_review` at attempt 0 (observed emitting a no-op `needs_page:privacy`
proof item — cancelled), and (b) the site isn't live (nothing pulls vm-sites until the owner does §3a),
so there is no user-visible gap to close in a hurry — the pre-cutover build carries it. On next render the
form becomes `<form action="mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry" method="POST">`.

Aside (deploy-path proof, task): not done via a forced rebuild — instead **code-verified** that
`resolveGitRepoName` is present in the running pod `v1.0.1134` (3×) with `sites.github_repo='vm-sites'`
set, so the chassis→vm-sites routing is live. The live commit will be observed on the first natural build.

## R — 2026-07-17/18 · Box artifacts, email subject-line block, SES/DKIM, briefing + docs

- **REVIEW email spam-block diagnosed and fixed** (§P). Root cause was the tool stuffing the full
  ~480-char business description into the Subject; `subjectSnippet()` bounds it to 60 runes at a word
  boundary, body keeps the full text. `TestSubjectSnippet` PASS. Rides the next tool binary deploy
  with the /request hardening. Commit 08e767b7c.
- **SES/DKIM walked through with the owner.** Domain identity + Easy DKIM were already in place and
  passing (verified the live selector `ly4vxsi…._domainkey.leopardess.uk` resolves to a valid key);
  the one missing piece is a custom MAIL FROM (`bounce.leopardess.uk` — MX + SPF TXT, both absent in
  DNS). Read the blocked email's own X-SPAM-LEVEL table: the 4.4 CustomCheck (subject) was the block;
  everything else (DMARC_NONE, HEADER_FROM_DIFFERENT_DOMAINS, the resolver-hiccup DKIM_INVALID) is
  small and net reputation is negative (ham). So the subject fix is the ballgame; MAIL FROM is
  worthwhile hardening for the buyer-facing report emails. Two owner DNS records outstanding.
- **contact-form → mailto executed** (§Q) — `sql/p1_07`, source-only, fleet template untouched;
  stale `idea-uk@leopardess.uk` in the form description also aligned. Publishes on next contact build.
- **Box artifacts committed** (`box/`, commit 2b5a797a8): `provision-pullsync.sh`, `sitesync` +
  `.service`/`.timer`, `proxy_tool.conf`, `idea.uk.nginx` (16 tool paths + the 3 legal-page `.html→`
  301s + loud-404 static root), `README`. Everything §3a–§3e as ready-to-paste payloads.
- **Pull-sync explained to the owner in depth** (what/why); the explanation is now written up two
  ways: `BRIEFING_idea_uk_vm_site.md` §7 (read-aloud), and a step-by-step `provision-pullsync.sh`
  walkthrough folded into **RUNBOOK §3a** (per-stage: dirs → key → the deploy-key PAUSE → sparse
  clone → install → verify; plus by-hand verify + journalctl). Commit 795a34b84 (briefing).
- **Deliverable added:** `BRIEFING_idea_uk_vm_site.md` — a ~3,600-word plain-English, jargon-unpacked
  narrative (goal → done → status → next → why pull-sync → risks/decisions), safe to read aloud
  (no IPs/usernames/secret values).

## S — 2026-07-18 · §3a EXECUTED on the box: pull-sync is LIVE

Ran `box/provision-pullsync.sh` on 116.203.204.115. Clean:
```
== pre-flight ==  -- OK: deploy key accepted.
== sparse clone ==  Already on 'main' / up to date with 'origin/main'
== install script + units ==  symlink → sitesync.timer
== first sync + verify ==  8/8 OK   (timer next in 5min)
```
`/var/www/idea.uk` now holds the built site; `sitesync.timer` re-syncs every 5 min. **nginx untouched
— nothing public changed.** The site is on the box but not yet served; §3b–§3e is the next step.

**Two bugs found and fixed getting here** (both cost a real cycle, both now trapped in the docs):
1. **`ssh` ignores `$HOME`** — expands `~` from the passwd entry, so `env HOME=…` configured git but
   left ssh in `/var/www/.ssh` (www-data's passwd home, unwritable) → "Host key verification failed",
   then "Permission denied (publickey)" with the deploy key showing **"Never used"**. Fixed by naming
   identity + known_hosts explicitly via `GIT_SSH_COMMAND`, in the provisioner **and** in `sitesync`
   (the 5-minutely fetch runs as the same account — a hand-fixed clone would still have left a dead
   sync). Filed `/bugs_open/016` + pattern in `016b §9`. My first diagnosis (ssh-keyscan exiting 0 on
   an empty scan) was **wrong** and is recorded as such; the hardening was kept as a real latent flaw.
2. **`scp -r` nests when the destination exists** — the fixed script landed at
   `/root/idea-uk-box/box/` while the **stale** one at `/root/idea-uk-box/` re-ran, reproducing the
   identical failure and reading as "the fix didn't work". RUNBOOK now says `rm -rf` first + grep the
   copy for a new-version string before running. Also reordered the installer to **authenticate
   first, prompt only on failure** — the old `read` pause could never work under
   `ssh host 'bash script'` (no TTY), so re-runs are now fully non-interactive.

**§3b premise re-confirmed on the box:** the webroot listing shows `terms.html`,
`refund-policy.html` **and** `privacy.html` present — so the three `.html →` tool 301s in
`box/idea.uk.nginx` are required, not theoretical.

**Unattended sync tick PROVEN.** `systemctl status sitesync.service` after the first timer-driven run:
`Process: ExecStart=/usr/local/bin/sitesync (code=exited, status=0/SUCCESS)`, 142ms, triggered by
`sitesync.timer`. This is the run that matters — it proves the deploy key + `GIT_SSH_COMMAND` work
under **systemd as www-data**, not merely under an interactive root shell.

**Route coverage verified + legal-page decision settled (2026-07-18).** `/opt/idea/*.go` is not on the
box (binary only), so the reserved-path set was re-confirmed from `service.go:596-612`: **16 routes**
plus `/`. `box/idea.uk.nginx` covers all 16 in 15 location blocks (12 exact + 3 prefix; `^~ /order/`
covers `success`+`cancel`), 15 `proxy_tool.conf` includes, no gaps. **Owner confirmed the tool keeps
all three legal pages**, closing that open decision — the staged config already matched.

**Box housekeeping noted (not blocking):** the box reports `*** System restart required ***` and 19
pending updates (1 security). Worth scheduling deliberately — a reboot during/just after cutover would
muddy diagnosis, and the reboot also exercises `sitesync.timer`'s `OnBootSec`.

## T — 2026-07-18 · 🎉 CUTOVER DONE — idea.uk is one site

Reboot verified first (uptime 3min, `idea` active, 8 pages, `/health` 200, and the timer's last run
1 min after boot = `OnBootSec` proven). Then `nginx -t` + reload → `CUTOVER_RELOADED` at 14:55 UTC.

**Live conf was `idea.conf`, not `idea.uk`** — chasing that (via `setup.sh`, which provisioned the box)
caught that the staged config was a **downgrade**: it dropped `proxy_read_timeout 930s` ("the engine
can take minutes" — reports would have died at nginx's 60s default, invisible to every smoke test),
`limit_req`, the port-80 ACME+redirect block, IPv6 listeners, ssl_protocols, four security headers,
`client_max_body_size`, and used the wrong ACME webroot. Rebuilt as a faithful superset (`c5357595b`),
plus a separate `proxy_stripe.conf` because setup.sh deliberately exempts the webhook from
rate-limiting (Stripe retries in bursts; a 503 reads as an outage). Live preamble confirmed as the
plain `limit_req_zone $binary_remote_addr zone=idea_rl:10m rate=10r/s;` branch.

**Verified live from outside, post-cutover:**
- **Homepage is now the chassis site** — `<title>idea.uk — Where You Take an Idea Seriously</title>`
  + `assets/css/styles.css`, not the tool's embedded page. *The workstream's goal, achieved.*
- All **16 tool paths reach the tool**: 200 health/capacity/order·success/cancel/terms/refund/privacy,
  405 audience-check + request, 400 subscribe + stripe/webhook, 401 confirm/approve/decline/internal·run.
- **`/op` 404 is the TOOL's**, not a static miss — body is its branded `<title>Link not valid ·
  idea.uk</title>`; a real static 404 is nginx's plain page (checked `/definitely-not-here`).
  *Distinguishing these by body, not status, is the check that makes a 404 in this list safe.*
- Static: `/`, index/about/tools/report/contact, `/guides/`, `/news/` all 200; `/nonexistent-xyz` 404
  (loud, as designed). Legal 301s: `/terms.html`→`/terms`, `/refund-policy.html`, `/privacy.html`. ✓
- Security headers intact (HSTS, nosniff, SAMEORIGIN, referrer-policy).

**Process note (honest):** the port-8443 rehearsal was skipped — the swap went straight to live. It
came out clean and the post-cutover checks above cover the same ground, but the rehearsal existed so
a mistake would have been free rather than public. Rollback stayed one command away throughout.

**Still open after cutover:** (1) confirm the deployed `proxy_tool.conf` really carries
`proxy_read_timeout 930s` (the earlier stale-copy incident makes this worth an explicit grep — a
missing timeout is invisible until a real paid run); (2) **the money path is not yet proven** — send a
Stripe test event and watch `journalctl -u idea`; (3) purge the Cloudflare cache; (4) tool binary
deploy; (5) SES bounce records. Cosmetic: `listen … http2` deprecation warnings, inherited from
setup.sh's template, predate this change.

**Standing landmine recorded:** `idea.conf` is headed "managed by setup.sh — do not edit by hand".
Re-running `setup.sh` now **reverts the site to tool-only** and does `ufw --force reset`. Port these
server blocks into its stage-2 template before any re-provision.

## U — 2026-07-18/19 · The cutover broke the funnel; forms authored as sections; auditors finally run

**The break (`/bugs_open/017`).** Owner: *"/audience-check produces page: POST only … I can't find the
tool."* The tool served its OWN landing page at `/`, and **that page carried the entry forms**.
The cutover gave `/` to the static site; `/audience-check` and `/request` are POST-only *handlers*,
so the two surviving `href="/audience-check"` links were GET requests. Verified: the whole static
site contained **one** form (a newsletter box, no action) and **nothing** posting to `/request`. The
tool was healthy, reachable and unusable — no way to buy. RUNBOOK §3b had noted *"`/` (the landing
page it loses)"* as a routing fact; nobody asked what was ON that page.

**Correction to §T's confidence:** the post-cutover smoke tests all passed *because nothing errored*.
A funnel can be absent without a single non-2xx. Status codes cannot see a missing form.

**The auditors were right; they had never been run.** I first reported they'd missed this — wrong.
`SELECT source…` for the site showed 13 sources, none `discovery`. Once a run completed it caught the
owner's symptoms exactly: `empty_internal_href` in **site_component (header)** and **(footer)** (the
blank logo `src=""` and `href=""` buttons — correctly traced to the shared chrome, not per-page),
9 × `phantom_internal_link` (`/about`, `/report`, `/tools`, `/how-it-works`, `/method` — the site uses
`.html`, so all dead-end), 7 × `cta_names_unknown_destination`, 9 × `misdirected_cta`, and
**`needs_rerender`: 9 pages missing header/footer**. The genuine gap is narrower than I claimed: no
check models the *backend*, so nothing noticed a GET link to a POST-only route (`017` proposes
`backend_entry_orphaned`).
- ⚠️ **I wrote a duplicate trigger before searching.** `scripts/initial_messages/060improvement_loop/
  076_improvement_loop_trigger.sh` already spawns all three discovery agents — and passes `domain`,
  which is exactly why my hand-rolled envelope died. Deleted mine. Dispatch problems recorded as
  **error F in `bugs_open/002`**, including a well-formed envelope that produced *no orchestration at
  all* (unexplained; not the 300s rule).

**The fix — forms authored as chassis sections** (`sql/p2_01`): `audience-check-form` → tools page,
`report-request-form` → report page. Field names read from the tool source, never guessed
(`audience_check.go:159-160`; `service.go:327-355`), including the `company_url` honeypot and
`_elapsed` timing field. Appended to existing pages — no new pages, so the re-plan landmine is
untouched.
- **Own-goal caught before it shipped** (`sql/p2_02`): a raw SQL INSERT **bypasses
  `separateInlineJS`**, so `report-request-form` landed in exactly the shape 016b §9 calls broken
  (`js_len=0, src_ref=f, raw_inline=t`). `collectJSAssets` only publishes
  `/tools/assets/{function}.js` when `js_content` is non-empty — so `_elapsed` would never be set and
  the timing gate would **fail open silently while appearing present**. Also `function` had defaulted
  to `generic-text-block`, which would have published the JS under a colliding name. Both corrected.

**Outcome so far:**
- ✅ **Free taster LIVE** — `<form class="ac-form" action="/audience-check" method="POST">` is on
  idea.uk/tools.html.
- ✅ **The chassis→vm-sites→box pipeline is now PROVEN end-to-end by a real build** (the item that
  §T left outstanding): the build committed to `gqls/vm-sites`, the box's 5-minute sync pulled it,
  nginx served it. Previously only code-verified.
- ⚠️ *"Trust the artefact, not the status"* again: the tools work item reads **`failed` — "Claim
  timed out — handler pod likely died"** while the page rendered and deployed correctly. Known
  claim-timeout churn; the artefact is authoritative.
- ⚠️ The first report build failed `validate_content … 0 blockers, 1 errors` →
  `needs_human_review` (by design: error severity routes to a human rather than deploying). Detail
  was unrecoverable — the orchestration rows had been reaped. Most likely cause: that build ran
  against the **pre-p2_02 component** (raw inline `<script>`, `function='generic-text-block'`); the
  shape fix landed after it started. A straight re-drive on the corrected component passed
  `complete` first attempt, with no other change.
- ✅ **FUNNEL RESTORED — both forms live and verified end-to-end:**
  - `report.html` → `<form class="rr-form" action="/request">` carrying **all seven** fields:
    `name, email, business, audience, notes` + `company_url` (honeypot) + `_elapsed` (timing).
  - `tools.html` → `<form class="ac-form" action="/audience-check">` with `business, audience`.
  - `/tools/assets/report-request-form.js` serves **200** — so `collectJSAssets` published it and
    the timing gate is genuinely armed, not merely declared. This is the exact failure p2_02
    prevented, now positively confirmed rather than assumed.

**Still open on this site:** the `needs_rerender` finding — **9 pages missing header/footer** — plus
the header/footer `empty_internal_href` (blank logo `src`, dead buttons) and the phantom `.html`-less
links. All are chrome-level, affect every page, and want their own pass. The forms work regardless.

## V — 2026-07-19 · The chrome is broken on every page (`/bugs_open/018`)

Chasing the auditors' header/footer findings to the deployed artefact produced a worse number than
expected. Counted on the live homepage: **31 of 33 links are `href=""`** — the entire primary nav
(8 × `nav-link`), every page CTA, all three social links; only the two *literal* logo hrefs (`/`)
work. `<img class="header-logo-img" src="">` too, though `/assets/images/logo.jpg` serves 200. Some
footer links have empty text as well as empty href, so they are invisible rather than merely dead.
**The site is effectively unnavigable**, and has presumably been so since it was first built — it
only became visible when the cutover put it in front of the public.

Shape of the evidence: the chrome **templates** render correctly (classes, structure, `site-footer.js`,
the nav element). Only **resolved values** are missing. Both working links are template literals;
everything requiring data is empty. Combined with the 9 `phantom_internal_link` findings — the nav
model uses `/about`, the built site is `/about.html` — a URL-shape mismatch between the nav data and
`pages.url` would explain the empties and the phantoms together. Root cause NOT established; not
guessed. Fleet check included in 018 — the filler is shared code, so fix the filler, not the rows.

**Verification trap recorded** (I fell into it): the footer is emitted as `<section class="footer-…">`,
**not** `<footer>`. `grep '<footer'` returns 0 and *appears* to confirm the auditors'
`missing_structure` finding ("9 pages deployed without header/footer"). That finding is **false for
the artefact** — the chrome is all there. I reported "footer missing" to the owner before checking
properly, then corrected it. Anyone acting on `missing_structure` as written would re-run a rerender,
see chrome "return", and record a fix that fixed nothing.

## W — Missteps in this workstream, collected (so they are not repeated)

Kept together deliberately: each cost real time or nearly shipped a defect.

1. **Diagnosed from a message about a path as if it described contents** (`/bugs_open/016`).
   "Host key verification failed" means the host key was not found *where ssh looked*; I inferred an
   empty file and "fixed" a file ssh would never read. The actual cause — ssh expands `~` from the
   passwd entry, not `$HOME` — was only visible in the *second* error. `ssh -v` prints the paths and
   would have ended it in seconds.
2. **Re-ran a stale copy and read it as "the fix didn't work."** `scp -r box host:/root/idea-uk-box`
   **nests** when the destination exists, so the fixed script landed in `…/box/` while the old one
   ran. Symptom is indistinguishable from a failed fix. Now: `rm -rf` first, then grep the copy for a
   string only the new version contains.
3. **Staged an nginx config that was a silent downgrade.** The live conf is `idea.conf` (not
   `idea.uk`), and my version dropped `proxy_read_timeout 930s` — "the engine can take minutes" —
   along with `limit_req`, the port-80 ACME block, IPv6, ssl_protocols and four security headers.
   Every smoke test would still have passed; reports would have died at nginx's 60s default. Caught
   by reading `setup.sh`, the script that provisioned the box, instead of assuming.
4. **Declared the deploy path "code-verified" against the wrong symbol.** I grepped the pod for
   `resolveGitRepoName`; `/bugs_open/014` (filed the same day by another thread) showed the real fix
   is `resolveGitRepoNameDB`, because the old resolver read workflow state most workflows never
   populate. Right answer, wrong evidence.
5. **Cut over without the rehearsal I had written.** The port-8443 dry run existed precisely so a
   mistake would be free; the swap went straight to live. It came out clean, and the post-cutover
   checks covered the same ground — but that was luck, not method.
6. **Trusted green smoke tests over the funnel** (`/bugs_open/017`). All 16 routes returned the
   tool's codes, so the cutover was called a success — while the tool's entry forms had vanished with
   the landing page it lost. **A funnel can be absent without a single non-2xx.**
7. **Said the auditors had missed it. They had not — they had never been run.** Worse, I wrote a
   duplicate trigger without searching; `076_improvement_loop_trigger.sh` already existed and passes
   `domain`, which is exactly why my envelope died. Deleted; dispatch faults filed as `002 F`.
8. **Nearly shipped a silent security regression** (`sql/p2_02`). A raw SQL INSERT bypasses
   `separateInlineJS`, so the request form landed in the shape `016b §9` calls broken — no JS asset
   published, `_elapsed` never set, and the `/request` timing gate **failing open while appearing
   present**. Caught by checking my own work against the guide before deploy.
9. **Reported "footer missing" from a grep that could not see it** (§V above).

Common thread in 1, 4, 6, 9: *a check that cannot see the thing it is asked about returns a clean
answer.* The guide's "0 rows is not decisive" generalises further than SQL — to filesystem paths,
symbol greps, HTTP status codes and tag names.

## Open decisions

- ~~`/privacy` — tool or static?~~ **RESOLVED 2026-07-18 (owner): the tool keeps all three legal
  pages** (`/terms`, `/refund-policy`, `/privacy`); the static `.html` copies 301 onto them. Matches
  what `box/idea.uk.nginx` already stages.
- ~~`/contact.html` form~~ **RESOLVED 2026-07-17 §Q** — converted to a mailto (owner's choice); fix
  staged at source (`sql/p1_07`), publishes on the next contact build.
- **Cloudflare proxied (orange) or DNS-only (grey)?** Unverifiable from the repo; decides whether the
  real-IP problem is live and whether Cloudflare WAF/Turnstile is reachable as the blocking layer.
- **Does relojistas.com migrate from push to pull too**, or do the two mechanisms coexist? (Coexist is
  fine and is what §2b's allowlist enables.)

## §X — 018 root-caused; the Go fix put to the council (2026-07-19)

**Fleet check first, per 018's own instruction.** idea.uk is the ONLY affected site.
```sql
SELECT s.domain, count(*) FILTER (WHERE sc.rendered_html LIKE '%href=""%') AS empty_href
FROM site_components sc JOIN sites s ON s.id=sc.site_id
WHERE sc.rendered_html IS NOT NULL GROUP BY s.domain ORDER BY empty_href DESC;
-- idea.uk 2 (of 3 components); all ten other domains 0.
```

**Root cause (established, not guessed).** `render_site_components_action.go:222-262` builds
`RenderContext.ContentData` as a hardcoded literal map and `:530` renders the component's
`html_template` against it. `input_schema` appears **nowhere** in that file; `site_components.
content_data` (`{}` fleet-wide) is not read either. idea.uk's two per-site fork components
(`site-header` f420f3fa, `site-footer` 4238e467, created 2026-05-06, each used by exactly one site)
declare a completely different vocabulary — `nav_link_1_url`…`nav_link_4_label`, `cta_primary_url`,
`nav_aria_label`, `col1_link1_url`…`col3_link4_label`. None is in the map, so each gets `""`.

**The confirming detail.** `company_name` IS in the map, and is the *only* chrome value that
rendered (`<span class="header-logo-name">idea.uk</span>`, `aria-label="idea.uk home"`). Everything
absent from the map is blank. That is the whole bug in one line.

**Why it was silent.** `component_library.go:544-559` — Go's `missingkey=zero` yields `<no value>`
for a missing key; the renderer counts them, `strings.ReplaceAll` them to `""`, and logs a **count
only** with a 100-char template preview. 30 dead controls → one unread log line.

**MISSTEP / CORRECTION — 018's stated theory is wrong and I nearly inherited it.** 018 proposed a
URL-shape mismatch (`/about` vs `/about.html`). Checked: idea.uk's `site_nav_items` holds 6 primary
+ 1 utility + 1 legal, all `status='active'`, all correctly `.html`-suffixed. `GetNavItems` would
return every one. The data is fine; nothing consumes it. Corrected in `/bugs_open/018` (commit
`46feb0f74`) with the superseded block kept and labelled, not deleted.

**MISSTEP — a regression I drafted and only caught via prior-art search.** My first submission
applied each field's declared `fallback` whenever resolution missed. `header-bold-gradient.cta_url`
is `{source: pages.contact, fallback: /contact.html}` — documented in
`NOTES_cta_link_integrity.md:333-336` as *"the literal fossil of the 143 of 144 buttons point at
/contact.html bug"*. That rule would have re-fabricated the phantom **LNK-007 was deployed to kill,
across nine live sites**. Fixed before submitting: fallbacks apply **only** to `source:static`;
a missed data-source resolution leaves the field unset, which is LNK-005 (correct-or-absent) by
construction. Lesson: *a schema's `fallback` is not a safe default — on a URL field it is a
fabrication licence.*

**Prior art that reframed the submission** (owner asked for it, and he was right — it changed the
argument, not just the citations):
- **LNK-007** (`docs026_concept_register/register/link-management.md:52-58`, status *deployed*)
  fixed this same hardcoded-ContentData surface by hardcoding **more**. **LNK-016** (`:127`) records
  why: *"Its scope excludes ContentData values and literal anchors — which is why the header/footer
  phantoms (LNK-007) had to be fixed at source in Go instead."* So the submission argues with a
  deployed decision and must say so up front.
- `PLAN_2026-07-19_cta_link_integrity.md:86-89` names *"one derivation function, three call sites:
  plan_sections, resolve_internal_links, applyCTARecompute"*. **The chrome renderer is absent.**
  The submission is therefore framed as *adding the missing fourth call site* to an
  already-owner-approved principle, not as a new idea.
- `RUNBOOK_cta_link_integrity.md:40-42` — the CTA census is `page_components` only and *"does not
  cover site_components (header/footer)"*. That is why chrome was never measured.
- `093_component_creator.sql:1185` — the component-creator agent is contractually told to emit
  TIER C `site_specs.*` / `site_assets.*` sources. **The generator produces, by design, schemas the
  chrome renderer cannot honour** — which is why this recurs rather than being one bad component.

**Reuse (the council has a reuse seat; this is the answer).** `sourceResolver`
(`plan_sections_action.go:65-475`) already resolves `site_specs.*`, `site_assets.*` (incl. image-role
alias), `pages.*`, `config.*`, and already refuses to fabricate URLs (`:449-453`). Same package —
callable with no export or refactor. Zero new resolution code. `newSourceResolver`'s own comment
says an empty `pageName` *"degrades safely"*, which is exactly the site-wide chrome case.

**Submission.** 3 edits, all in `render_site_components_action.go`; two deliberate limits chosen to
preserve LNK-007 — (a) resolved values **fill gaps only**, the hardcoded map stays authoritative
wherever it supplies a value, so the nine working sites render byte-identically; (b) fallbacks
**static-only**, per the missteps above.
```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/council_submission_chrome_schema_driven.json
SUBMISSION_CORR=7152c7cf-5c4d-41b3-8ab4-0c3d8d40fbd5
RUN_ORCH_ID=3e7f7507-e1eb-4a40-b39b-9b2c5c593a69   # council-gate-132453
```
Submission JSON kept in this directory so a resubmission starts from the reviewed text, not memory.

**Known and stated in the submission, not hidden:** this Go change does **not** by itself make
idea.uk navigable. Both its templates have **zero** `{{if}}` gates (`grep -c '{{if'` → 0), so a
legitimately-unresolved field still emits an empty attribute. Gating them is a separate DB-only
change. Also flagged as the one visible change to a working site: `sites.logo_url` is empty for
vonc.com too, and `header-bold-gradient.logo_url` declares `site_assets.logo`, so vonc may newly
render a logo where it currently shows its `{{else}}` glyph.

### §X.1 — MISSTEP: I called a queued message a dropped one, and paid for it (2026-07-19)

**What I claimed, and it was wrong.** After submitting (12:24:53 UTC) I found no
`orchestration_states` row and no `orchestration_requests` row for the run, and nothing in the
chassis log for my ids. I checked the topic, found my message intact, saw *other* council runs
completing in the same window, and concluded: *"the message was skipped, not delayed."* I then
re-submitted.

**That was wrong, and the reasoning error is worth naming.** The evidence I used — my message
appearing FIRST in a `kcat -o -60` window while later council notes were being written — is exactly
what an **in-order backlog** looks like on a single-partition topic. The council notes I saw landing
were from messages produced *before* mine. First-in-window means *oldest unprocessed*, not *skipped*.
I read a queue as a hole.

**What is actually true** (`kafka-consumer-groups.sh --describe --group generic-requests-group`):
```
GROUP                  TOPIC                          PARTITION CURRENT-OFFSET LOG-END-OFFSET LAG
generic-requests-group system.agent.generic.requests  0         93403          93465          62
```
- **One partition, one consumer** (`agent-chassis-5c568b8c74-2f4qv`, segmentio/kafka-go). Strict
  in-order, serial processing.
- Across ~2.5 min of observation `CURRENT-OFFSET` advanced by **1** while `LAG` grew **41 → 62**.
  Production is outpacing consumption by roughly an order of magnitude.
- The message at the consumer head (offset 93403) carries timestamp **12:24:43 UTC**, i.e. the
  consumer is **~26 minutes behind wall-clock**.

**The tell that settles it:** the head-of-queue message is timestamped 12:24:43 and my submission
was 12:24:53 — mine is roughly *at* the head. It was always going to run; it just had not got there
yet. Nothing was dropped.

**Cost of the misstep.** The resubmission (`RESUBMIT_CORR` kept the trail id stable, so both land on
`7152c7cf`) means the same plan will now be reviewed **twice** and spend council credits twice. It
cannot be recalled — the message is already in the topic. Recorded rather than quietly absorbed.

**The transferable rule.** *On a single-partition topic, "my message hasn't been processed but later
ones have" is not evidence of a drop — verify against consumer-group LAG before concluding anything.*
Check `CURRENT-OFFSET` vs `LOG-END-OFFSET` and the head message's timestamp FIRST; it is one command
and it is decisive:
```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```
(Gotcha: the broker pod is `personae-kafka-cluster-combined-pool-prod-0`; the CLI is under
`/opt/kafka/bin/`, not on `$PATH`. `--all-groups` across every group takes >120s — describe the one
group.)

**This also retro-explains an earlier entry in these notes.** §S recorded that on-demand discovery
dispatches "produced no orchestration row at all" and that "the pod was six hours old, so the
documented 300s-post-restart drop doesn't explain it." That thread almost certainly hit *this* —
a backlog, not a drop — and it cost a run and an unresolved TODO. The 300s-post-restart rule in
CLAUDE.md is real but is not the only reason a dispatch appears to vanish; **queue depth is the
commoner one, and it is now measurable.**

**Wider significance (not idea.uk's problem alone).** Many sessions fire triggers at this one topic
concurrently. A serial consumer at roughly one orchestration per couple of minutes, with a backlog
that grew while being watched, means *every* thread's dispatch latency is a function of how many
other threads are firing. Worth its own bug file; noted here because it changed how this task ran.

### §X.2 — chrome templates fixed and driven; both council rounds voided by bugs_open/019 (2026-07-19)

> **CORRECTION to §X.1.** I wrote there that "production is outpacing consumption by roughly an
> order of magnitude", from a 2.5-minute sample (lag 41→62, offset +1). **Too strong.** Over the
> following ~35 minutes lag went 41 → 62 → **24**, and the consumer cleared 133 messages. The queue
> is **bursty, not diverging** — it drains, but head-of-queue age (~26 min) is the real symptom.
> Corrected figures live in `/bugs_open/030`. What caught it: continuing to sample instead of
> filing the first reading.

**THE FIX IS APPLIED (DB, live at next render).** `sql/p3_01_chrome_templates_gated.sql`:
both templates rewritten against the renderer's actual vocabulary, **every anchor gated**, CSS
preserved verbatim except the footer grid (`2fr 1fr 1fr 1fr` → `2fr 1fr 1fr`, since 3 hardcoded
link columns became 2 data-driven ones). `input_schema` rewritten to match. `sites.logo_url` set to
`/assets/images/logo.jpg` (asset served 200; column was empty). Originals backed up beside the SQL.

**Verified BEFORE applying, not after** — rendered both templates through `text/template` with
`Option("missingkey=zero")`, the same engine as `executeGoTemplate` (`call_agent.go:1150-1152`),
against live-shaped data (`scratchpad/tplcheck/main.go`):

| case | empty href/src |
|---|---|
| full data (6 nav items, cta_url, logo set) | **0** |
| logo_url empty (today's data) | **0** |
| **renderer supplies NOTHING at all** | **0** — only the literal `/` logo link survives |

That last row is the point: correct-or-absent (LNK-005) now holds **by construction**, not by the
data happening to be present.

**Driving it.** Reused the existing `needs_rerender` item `dce4e4ac` (spec already carried
`refresh_site_components:true` + all 9 pages) rather than inserting a duplicate — it was stuck at
`status='detected'` and the loader only consumes `('triaged','approved')`
(`load_work_item_actions.go:559`). Promoted it and **corrected its false stated reason** in the same
write (`sql/p3_02`), keeping the old text as `superseded_reason` so nobody re-acts on it. Also fired
an explicit dispatch, `scripts/initial_messages/001_assemble_all_pages_rerender/081e_rerender_pages_for_idea.uk.sh`
(orchestration `c6179a53`, published 13:32:37Z).

⚠️ **`refresh_site_components:true` is load-bearing.** Without it the run reassembles from the
STORED `site_components.rendered_html`, which still holds the broken chrome —
`renderAndStoreSiteComponent` skips a slot that already has `rendered_html` unless forced
(`render_site_components_action.go:468-483`). A rerender without that flag would look like it
worked and change nothing.

**NOT YET VERIFIED LIVE.** At the time of writing `site_components` still hold the 2026-07-02
`rendered_html` with `href=""`. The deployed pages have chrome **baked in**, so nothing changes on
idea.uk until the rerender runs, deploys to `gqls/vm-sites`, and the box's `sitesync.timer` pulls
(≤5 min). Verify against the artefact, never the item status:
```bash
curl -s https://idea.uk/ | grep -oE '<a href="[^"]*"' | sort | uniq -c | sort -rn   # want: no href="" rows
curl -s https://idea.uk/ | grep -o 'header-logo-img[^>]*'                            # want: non-empty src=
```

**BOTH COUNCIL ROUNDS VOIDED — and not on the merits.** Runs `3e7f7507` (13:01) and `b8a0c8a5`
(13:14) both ended `complete_invalid`. Cause, from `collected_data->>'__step_error'`:
`response truncated: stop_reason=max_tokens (output_tokens=8000 reached the configured cap)`.
That is `/bugs_open/019` — one truncated reviewer voids the whole round. Token log:

| run | seat | ok | output | % of 8000 |
|---|---|---|---|---|
| 3e7f7507 | editquality | t | 7,296 | **91%** |
| 3e7f7507 | bug_historian | t | 5,293 | 66% |
| 3e7f7507 | reuse_agent | t | 3,352 | 42% |
| 3e7f7507 | **guidelines** | **f** | — | truncated → void |
| b8a0c8a5 | **editquality** | **f** | — | truncated → void |

**A different seat failed each time on the byte-identical submission** — so this is several seats
sitting just under the ceiling, not one pathological writer, and "shorten it until it passes" is not
a reliable workaround. Appended to `/bugs_open/019` as its second reproduction. The panel *had*
selected 6 relevant seats (render, debugging, compliance, guidelines, reuse_agent, bug_historian),
so the relevance filter worked; three reviews completed and were then discarded.

**Consequence for this task: the Go proposal has no verdict, and cannot get one until 019 is
resolved or the plan is cut.** The submission JSON is committed
(`sql/../council_submission_chrome_schema_driven.json`) so a resubmission starts from the reviewed
text. Not worked around — filed.

### §X.3 — chrome VERIFIED LIVE; the taster's two remaining defects fixed (2026-07-19/20)

**The chrome fix worked. Verified against the deployed artefact, not the item status:**
```
$ curl -s https://idea.uk/ | grep -oE '<a href="[^"]*"' | sort | uniq -c | sort -rn
      4 <a href="/about.html"     3 <a href="/tools.html"    3 <a href="/report.html"
      3 <a href="/news/index.html" 3 <a href="/index.html"   3 <a href="/guides/index.html"
      3 <a href="/contact.html"   2 <a href="/"              2 <a href="#"
      1 <a href="/privacy.html"
$ curl -s https://idea.uk/ | grep -o 'header-logo-img[^>]*'
  header-logo-img" src="/assets/images/logo.jpg" alt="idea.uk logo" /
```
**Zero `href=""` — down from 30.** `site_components` re-rendered 13:55:44, work item `complete`.
The logo renders. The 2 remaining `href="#"` are the `brief-explanation` CTAs — page components,
already on 018's list as `dead_control`, NOT chrome, and untouched by this work.

**Then the owner reported two things that are NOT the chassis at all — both the tool.**

**(1) "The audience check page has no chrome, just the text."** `/audience-check` is an **AJAX
fragment endpoint by design**: `audience_check.go:135-138` is POST-only and
`renderAudienceHTML`'s own comment says it *"produces the HTML fragment the taster widget drops
into its result div"*. There was no widget. p2_01 seeded the form as a plain native POST with no
JS, so the browser NAVIGATED to the endpoint and rendered the bare fragment.

The diagnosis came from putting the two forms side by side on the live pages:
```
/report.html : <form class="rr-form" action="/request">  + <script src="/tools/assets/report-request-form.js">
/tools.html  : <form class="ac-form" action="/audience-check">   ← and no script at all
```
Measured cause: `collectJSAssets` (`rerender_single_page_action.go:156-176`) publishes
`/tools/assets/{function}.js` **only when `js_content` is non-empty**.
```
report-request-form  js_content = 266 bytes → /tools/assets/report-request-form.js  HTTP 200
audience-check-form  js_content =   0 bytes → /tools/assets/audience-check-form.js  HTTP 404
```
So this is **the same defect p2_02 fixed for the other form, never applied to this one** — and the
fix is that established mechanism, not a new one. `sql/p3_03` adds the interceptor
(`ev.preventDefault()` is the whole bug), a result div, and styles for the tool's own classes.
Post-apply shape check passes: `js_len=1463, src_ref=t, result_div=t, raw_inline=f`.

**(2) "The tool link still shows POST only."** Live: `GET /audience-check` → **HTTP 405**. The
tool-list cards on index + tools pointed at it. Crucially the card URLs are **not hand-written** —
`tool-list.items` is `source: "query.pages_where_type:tool"`, derived from the `pages` table. So the
pointer page `tool-audience-check` (`url='/audience-check'`) is the source of truth, and that is
what p3_03 corrects, to `/tools.html#audience-check` (the id the seeded form section already
carries). The already-resolved `page_components.content_data` copies are patched too — transient by
nature (`bugs_open/001`), which is exactly why the `pages` row is fixed as well rather than instead.

**MISSTEP AVOIDED, worth recording.** My first instinct was to patch the two `content_data` rows,
which would have "fixed" it until the next plan pass silently reverted it. Reading the component's
`input_schema` — `source: query.pages_where_type:tool` — is what showed the URLs are derived. *A
rendered value that looks hand-authored may be a resolved query; check the source before editing
the copy.*

**Stopgap flagged, not hidden.** The tool's fragment ends with `<a href="#request">`, and **no page
on this site has `id="request"`** (the real form is `/report.html#request-a-report`). That href
comes from the box binary, so it cannot be fixed from the DB. The injector retargets that one known
anchor after injection, and says so in a comment — **delete it once the tool emits the right href**.

**NEW STRUCTURAL FINDING — `/tools/assets/site-header.js` 404s.** The chrome template references it
(inherited from the original), but `collectJSAssets` reads **`page_components` only** (`:157`),
never `site_components`. So a site component can reference a JS asset that is **never published**,
with no error anywhere. The hamburger/mobile menu is therefore dead on every page. Note
`render_js_snippets_for_site_action.go:203-219` DOES union both tables for the `snippets.js` bundle
— so the two JS paths disagree about whether chrome exists. Fleet-wide; not fixed here.

**Pending:** rerender `4a66f0bd` fired 10:05:05Z to publish `audience-check-form.js` and rebuild
the pages. Expect ~25-35 min queue (`bugs_open/030`) plus ≤5 min box sync. Verify with:
```bash
curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk/tools/assets/audience-check-form.js   # want 200
curl -s https://idea.uk/tools.html | grep -c 'audience-check-form.js'                          # want 1
curl -s https://idea.uk/tools.html | grep -oE 'href="/tools.html#audience-check"' | head       # cards retargeted
```

### §X.4 — the rerender that reported success, deployed, and changed nothing (2026-07-20)

**This is the CLAUDE.md rule biting in a form I had not seen: `complete` was true, the deploy was
real, and the change still did not reach the page.**

After p3_03 I fired `rerender-pages`. Result: 9/9 `page_rerender` items `complete`, a real deploy to
`gqls/vm-sites` per page, and the JS asset genuinely published —
`/tools/assets/audience-check-form.js` → **HTTP 200, 1469B**. Every status surface said done.

The page markup was **unchanged**: no `<script src=…>` ref, no `#ac-result` div, tool cards still
`href="/audience-check"`.

**Two separate traps here, and I fell into the first one.**

1. **I checked the wrong artefact and nearly called a real deploy a no-op.** I looked at
   `page_components.rendered_html`, saw 18/19 July timestamps, and concluded "reported complete,
   did nothing". Wrong — the rerender path assembles and deploys directly and does not necessarily
   write back to `page_components`. What settled it was reading the work item's `result` JSON, which
   lists the deployed files verbatim:
   `"files": ["/tools.html", "/tools/assets/audience-check-form.js"]`.
   *A stale `page_components.rendered_html` is not evidence that a rerender did nothing.*

2. **The real cause — the section re-render is gated behind `spec.reason`.**
   `rerender_page_sections_action.go:47-51`:
   ```
   check_rerender_mode (conditional: reason==image_landed OR reason==section_data_resolved)
     -> rerender_sections -> check_escalated -> save_sections -> render_page
   else_step (no/other reason) -> render_page      ← ASSEMBLE-ONLY
   ```
   The items `rerender-pages` creates carry **no `spec.reason`** (verified: `{"domain","page_id",
   "filename","page_name"}` and nothing else). So all 9 took the assemble-only branch: pages
   rebuilt from **stored section HTML** and deployed. **A `content_components.html_template` edit
   can never reach a page down that path.**

**Why the JS asset landed anyway, which is what made this confusing.** `collectJSAssets` reads
`content_components.js_content` **directly**, not the rendered HTML — so the asset published while
the `<script>` tag that references it (which lives in `html_template`) did not. The result is the
worst possible shape to debug: the asset exists, returns 200, and nothing loads it.

**Fix: `sql/p3_04`** inserts `page_rerender` items for tools + index with
`reason='section_data_resolved'`, the documented route into `rerender_page_sections` — "re-render
ALL of a page's sections from their STORED content_data plus FRESHLY re-resolved dynamic fields,
WITHOUT invoking the content writer (no LLM)". Exactly right here: copy unchanged, only the template
and one query-resolved URL need refreshing. Scoped to the 2 affected pages, not all 9.

⚠️ **Guarded, because this path can rewrite live copy.** If ANY section has NULL/empty
`content_data`, `rerender_page_sections` escalates the WHOLE page to the LLM content writer. p3_04
refuses with a `RAISE EXCEPTION` rather than risk that on live sales copy. The guard passed (0
sections affected), so the insert proceeded — but the check is in the file permanently.

**Transferable rule for the next thread:** *a rerender has two modes, and the default one cannot
see template changes.* If you edited `content_components.html_template`, a plain `rerender-pages`
will deploy and report success without applying it. You need `spec.reason='section_data_resolved'`
(or `image_landed`). Worth noting `rerender-pages` offers no way to set that — hence the manual
insert.

### §X.5 — VERIFIED LIVE end-to-end (2026-07-20 11:05)

`p3_04` (`reason='section_data_resolved'`) was the right route and is now proven. Final state,
checked against the deployed site and the running tool, not against work-item status:

```
1. JS asset          /tools/assets/audience-check-form.js   HTTP 200, 1469B
2. /tools.html        audience-check-form.js  ×1   id="ac-result"  ×1
                      href="/tools.html#audience-check" ×1   href="/audience-check" ×0
3. /index.html        href="/tools.html#audience-check" ×1   href="/audience-check" ×0
4. site-wide sweep    href="/audience-check" → 0 on /, /tools.html, /index.html,
                                                  /report.html, /about.html
5. FUNCTIONAL         POST /audience-check → HTTP 200, 2537B, body begins
                      "<h4>Your stated audience</h4><p>independent vets in the UK</p>…"
                      and contains exactly 1 × href="#request" — the dead anchor the
                      injector retargets to /report.html#request-a-report
```

So the taster now runs in place: the form intercepts, POSTs by fetch, and injects that 2537B
fragment into `#ac-result` **without leaving the page**, so chrome is retained; and no route
anywhere on the site still advertises the POST-only endpoint to a browser.

**One stall en route, recorded in `/bugs_open/006` C addendum.** The `tools` item sat `claimed` by
`build-dispatch-loop` for 16 minutes, `attempt_count=0`, `result='{}'`, no log line — while its
same-second sibling (`index`) completed in ~90s. There is **no requeue predicate on `claimed_at`
anywhere in `platform/`**, so the "claim-timeout sweep" 006 describes is not in the Go tree and its
window is >16 min. Reset to `triaged` by hand and it completed in under a minute. Safe here
*because* nothing had been done — `attempt_count=0`, empty result, no handler log — so there was
nothing to duplicate. That condition is the whole justification; do not generalise the reset to the
churn case, where the work HAS succeeded and the right action is to mark it `complete`.

**Still outstanding on this page (not regressions — pre-existing, and named so they are not
rediscovered):**
- `/tools/assets/site-header.js` **404** — hamburger/mobile menu dead on every page. Cause:
  `collectJSAssets` reads `page_components` only (`rerender_single_page_action.go:157`), never
  `site_components`, so chrome can reference a JS asset that is never published. Note
  `render_js_snippets_for_site_action.go:203-219` DOES union both tables for `snippets.js` — the two
  JS paths disagree about whether chrome exists. Fleet-wide.
- The tool's fragment emits `href="#request"`, an id on no page. Retargeted client-side by the
  injector as a **stopgap**; delete that line once the tool binary emits
  `/report.html#request-a-report`. Tool deploy is still pending (hardened `/request` + email subject
  fix ride the same build).
- 2 × `href="#"` on index/tools — `brief-explanation` CTAs, page content, on 018's `dead_control`
  list. Untouched by this work.

### §X.6 — COUNCIL VERDICT: REVISE (2026-07-20 18:37) — and 019's fix proven live en route

**The round completed instead of voiding — that is the first thing to record.** The chassis image
deployed 17:58:20 UTC contains `tolerate_truncation` (probed in-pod: 1 occurrence, vs **0** in the
Jul 19 16:42 binary; controls `max_tokens` 15→18, `stop_reason` 3→6). The report shows
`"unreadable": ["review_editquality.result"]` — **edit-quality truncated again and the round carried
on regardless**, degrading that one seat instead of discarding everyone's work. That is exactly the
behaviour `/bugs_open/019` specifies, now proven on a real submission. Third attempt on this trail;
the first two voided.

Queue latency, for `/bugs_open/030`'s record: published 18:07:00 UTC, orchestration created
**18:23:46** — **16 min 46 s**.

**Verdict: REVISE**, `decided_by: objection from bug_historian`. Tally: **7 approve** (reuse_agent,
guidelines, compliance, render_guardian, debug_historian, constitution, mission), **4 object**
(bug_historian, guardian, tooling_provenance, prior_art_librarian), **4 abstained**, 1 unreadable.

**TWO CONCRETE DEFECTS THE COUNCIL CAUGHT THAT WOULD HAVE SHIPPED.** These are not style notes:

1. **The `{"fields": {...}}` assumption is wrong for real components.** A reviewer's own check
   enumerated live `input_schema` shapes and found several with **no `fields` wrapper at all** —
   `Hero Section` and `Call to Action` have `fields_type = NULL` and their top-level keys ARE the
   field names (`headline`, `primary_cta`, `primary_cta_url`, `secondary_cta_url`). My edit 2
   unmarshals into `struct{ Fields map[...] }`, which for those yields an **empty map with no
   error** — the resolution would silently do nothing, exactly the class of silent no-op this
   submission exists to remove. Caught by `guardian` (low) and confirmed by data, not opinion.
2. **The gap-fill precedence is wrong for non-string values.** `render_guardian` (low): my check
   `if s, isStr := existing.(string); !isStr || s != "" { continue }` treats any NON-string existing
   value (e.g. the bool `show_subscribe`) as "not present" and falls through to resolution — so a
   legitimately-set non-string hardcoded value could be overwritten. The stated guarantee
   ("wherever the hardcoded map already supplies a non-empty value it stays authoritative") is *not*
   what the code implements. My own sketch, my own bug.

**THE DECISIVE OBJECTION (bug_historian, high ×2) — and it is right.**
- *"patches ONE call site of a shared underlying mechanism while leaving the mechanism itself
  generic and exploitable elsewhere"*: `component_library.go:544-559`'s blanket `<no value>→""`
  substitution stays untouched, so any other renderer of a `content_components.html_template`
  against an incomplete map inherits the identical silent blanking with no signal at all.
- *"the plan's own corroboration documents a SECOND instance … and does not touch it"*: I cited
  `bugs_open/041` (collectJSAssets omitting `site_components`, with the correct UNION already
  written in `render_js_snippets_for_site_action.go:203-219`) as my strongest argument that this is
  a class — while my edit list leaves it alone. Using an instance as evidence and not fixing it is
  the inconsistency. Fair, and I did not see it.

**Blast radius, now measured rather than assumed.** `guardian` objected that I asserted it; the
reviewers' check answered it: **six** agent definitions reference this action —
`nav-updater`, `rerender-pages`, `site-work-orchestrator`, `rerender-site`, `nav-link-fixer`,
`pageflow-builder`. The "nine sites render byte-identically" argument has to hold across all six
callers, not the one I had in mind. That belongs in the revised risks section as fact.

**Also owed (tooling_provenance, medium):** the plan reads `PLAN_2026-07-19_cta_link_integrity.md`
et al as prior decisions but writes nothing back. Adding the fourth call site to a principle that
documents three should append a `doc_notes` entry for that pipeline — the read side of the contract
was done meticulously, the write side not at all.

**Harness limitation, not a plan defect (`prior_art_librarian`):** its Schema section omits
`site_components` and `site_specs`, so it could not verify the load-bearing "no `navigation` aspect
on any of the 11 sites" claim and flagged it as unverified. Worth knowing that this seat is
structurally unable to check the tables this class of submission rests on.

**SIDE-FINDING from a reviewer's check — a stale value this workstream believed fixed.**
`sites.content_data` for idea.uk still carries `"email": "idea-uk@leopardess.uk"`. §DONE item 5 and
`sql/p1_05`/`p1_06` record the contact email as corrected to `idea.uk@contactforsales.com` "across
all sources the validator COALESCEs" — that sweep evidently missed `sites.content_data`. Not
rendering today (the footer takes `email` from the render context), but it is a live wrong address
one code path away. **Verify before repeating the "fixed everywhere" claim.**

### §X.7 — COUNCIL ROUND 2: REVISE again (2026-07-20 20:36) — 6 approve / 5 object, and the split is real

Round 2 completed (queued ~40 min, ran ~14 min — bug 030's queue, not the council). **REVISE**,
decided_by editquality. `unreadable: ["review_guidelines.result"]` — a DIFFERENT seat truncated this
round (guidelines, vs editquality in round 1) and the round again carried on rather than voiding, so
019's fix is holding across seats. One high objection (down from two), 6 medium, 11 low.

**Two are genuine defects in MY sketch — accept:**
1. **editquality (medium), edit 2/4 wiring:** `unresolved` is declared inside `if len(inputSchemaRaw)
   > 0 { … }` but edit 4 reads it at the RenderTemplate call site OUTSIDE that block — a Go scoping
   bug as written. Must declare `var unresolved []string` at function scope. Correct, and it is the
   join between the two edits, so it matters.
2. **editquality/prior_art (low), the href/src heuristic:** `missingBareFields` only matches
   double-quoted `href="{{.x}}"`. I checked the corpus: the escalation is the only new actionable
   signal, so its precision matters. Single-quoted/whitespace variants slip through. Fixable by
   widening the regex; still a heuristic (acknowledged in risk 5).

**The recurring HIGH (bug_historian) is now a genuine design fork, not a defect — this needs an
owner call, not another silent revise:**
> the strip `ReplaceAll("<no value>","")` still fires on ANY runtime `<no value>` (e.g. `{{.Foo.Bar}}`,
> `{{.Foo | fn}}`), but the NEW log only fires when `missingBareFields` (regex on `{{.Name}}` in the
> SOURCE) matched. So detection NARROWS vs the old unconditional count-log, even as it gains field
> names. And render_guardian (medium) + the FAIL-LOUD contract: **a named log is still not
> escalation** — content is still blanked, just better-documented.

They are right that a log is not a gate. My submission argued that on purpose (bugs_open/023: the
platform's problem here is a DELIVERY gap — 34 findings unread — so another unconsumed signal helps
nobody). **That disagreement is now explicit and repeated, and it will not resolve by me revising
wording.** The council wants fail-loud to MEAN something (block or escalate); the plan deliberately
chose observability because the consumers don't exist. That is an owner decision about scope.

**Everything else the reviewers asked for, I have now MEASURED (they flagged these as asserted, and
several seats literally cannot query the tables):**
- **Log-volume risk (guardian, medium — the one real unquantified item):** components with a
  URL-bound bare placeholder = **71 total / 59 active**; of the active, **30 ungated / 29 gated**. So
  edit 3 could newly log at Error on up to ~30 active components fleet-wide when their URL fields are
  unresolved. That is the real cost of the mechanism change, and it was unquantified. NOT trivial.
- **Blast radius (guardian/prior_art, asserted):** confirmed by query — `render_site_components` is a
  step in exactly the six agents named (nav-link-fixer, nav-updater, pageflow-builder, rerender-pages,
  rerender-site, site-work-orchestrator). Attachable as a check next round.
- **The existing guard (reuse/prior_art, medium):** `store_generated_component_action.go:305-318` is
  **generation-time and blocking** — it refuses to STORE a template that already contains `<no value>`,
  and refuses a regeneration that drops/renames a field. It does NOT detect render-time blanking of a
  correctly-stored template. So it is not the same guard and edit 3 does not duplicate it; they are
  complementary (store-time contract vs render-time signal). The reviewers were right to ask; the
  answer is that they are different layers.
- **doc_notes (reuse/guardian/tooling, low):** table verified — `(subject_type, subject_key, body,
  categories jsonb, …)`. Edit 6's INSERT matches. Real table, established provenance mechanism.
- **sourceResolver signature (multiple, low):** confirmed against source — `newSourceResolver(siteID,
  db, logger, pageName)` and `(r *sourceResolver) resolve(ctx, source) (interface{}, bool)`. The
  sketch's call matches.

**So round 3 is not mechanical.** The measurable objections I can close in the plan. The load-bearing
one is a scope decision — does the fix stop at "name the dead control loudly" (my position, grounded
in 023's delivery gap) or must it escalate/gate (the council's position, grounded in FAIL-LOUD)? I am
taking that to the owner rather than revising past it a third time. See README entry.

### §X.8 — bugs_open/017 auditor gap: `backend_entry_orphaned` Finding A BUILT (2026-07-21, session "bugfix 017")

Separate strand from the chrome-renderer council above. This closes the *durable* half of
`bugs_open/017` (the static-cutover funnel break); the site half was already fixed & verified live
(§X.5) and I re-confirmed it holds today (forms present, zero GET links to `/audience-check`, taster
JS 200, `GET /audience-check` → 405 correct).

**Two premises in the 017 handoff were stale — corrected against the live DB:**
1. "discovery has never run against idea.uk (0 rows)" — FALSE now: **30 `source='discovery'` items**
   (9 phantom_internal_link, 8 cta_names_unknown_destination, 4 dead_control, …). The operational gap
   closed. Crucially **none of the 30 is the 017 symptom** — confirming the backend-orphan gap is real.
2. Proposed check gated on `deploy_config.target='vm'`. **idea.uk's `deploy_config` is `{}`** — no
   backend marker at all — so that gate (shared with `check_backend_unreachable`) would NOOP here. So
   `check_backend_unreachable` has *also* never probed idea.uk. Owner chose the **un-gated probe** path.

**Built:** `check_backend_entry_orphaned.go` (+ `_test.go`), commit `7b03f296a`. Reads deployed
`page_components`, keeps internal extensionless handler-like routes (`ExtractAnchors` +
`ClassifyLinkScope`; `.html`/assets can't 405 — a cost filter, not the decision), dedupes, probes
`GET https://<domain><path>`, flags **exactly 405** → high `needs_human_review`, no handler. Cap 40
probes/site (logged on hit). Modeled on `check_backend_unreachable`'s probe idiom.

**Induced-failing-branch evidence (live, 2026-07-21), not just wiring:** `GET /audience-check` → 405
and `GET /request` → 405 → FLAG (the exact symptom the owner saw as "POST only"); `GET /subscribe` →
**400** (a route that answers a bare GET with 400, not 405, is deliberately NOT flagged — validates
the 405-only boundary), `/tools.html` 200, `/health` 200, bogus 404, `/` 200 → all ignored. Unit test
17/17, `go vet` clean. On idea.uk today it reports clean (links removed by the fix) — so no
false-positive on the fixed site.

**Under council review** (advisory), corr `ed4851c9-e51b-446d-a4b4-bbbf516eaa60`. **Inert** until an
image roll + adding `backend_entry_orphaned` to a discovery agent's `checks` array (image-first, else
it references an unregistered check). **017 stays OPEN** until live. Finding B (`no_backend_entry`)
deferred — it needs a reliable "site has a backend" signal, which the empty `deploy_config` shows is
itself missing (a data-model gap for a separate decision).

### §X.8 — ROUND 3 built as CODE, not just a plan (2026-07-21, owner ruling applied)

Owner ruled via AskUserQuestion 2026-07-21: **ship the observability version now, block/escalate as
follow-on `bugs_open/054`.** The round-3 submission JSON already baked that ruling in and `054` was
already filed (both in commit `76aaeb72b`, a prior turn). This turn turned the five plan sketches into
committed Go — **commit `36829b07b`** (`platform/orchestration/actions/{render_site_components,component_library,rerender_single_page}_action.go` + the CTA-plan write-back). No `Council-Reviewed` trailer — round 2 was REVISE on the scope point the owner overruled, and the trailer is earned only by APPROVED.

Verified before committing:
- `go build ./platform/orchestration/actions/` → **exit 0** (compiles). No signature changes:
  `RenderTemplate` stays a thin wrapper over the new `RenderTemplateReportingMissing`, so none of its
  many callers are touched.
- `input_schema` is `jsonb` → `COALESCE(cc.input_schema,'{}'::jsonb)` scanned into `[]byte` is fine.
- **Not inert:** `contextToInterfaceMap` merges `ctx.ContentData` at `component_library.go:744-746`, so
  edit 2's writes into `renderCtx.ContentData[name]` DO reach the rendered template. Checked, not assumed.
- `newSourceResolver` is a cheap struct init (specs/pages/assets lazy-loaded); `resolve()` returns
  `(nil,false)` on a `pages.*` miss and **never** fabricates `/contact.html`. Edit 2 handles
  `static`/`llm`/`renderer`/`""` BEFORE calling `resolve`, so the fossil-fallback path is unreachable
  on a data-source miss (LNK-005 correct-or-absent holds).
- Both schema shapes handled: wrapped `{"fields":{...}}` and FLAT (top-level field names, e.g.
  "Document Head"); a top-level value that isn't a field descriptor map is skipped, so stray scalars in
  the flat shape (`required`, `version`) are ignored.
- Live idea.uk (`curl`) already clean — every link resolves to a real `.html`, logo `src` populated
  (the DB fix `d63e62aad`). **This Go change is the fleet-wide mechanism; it is inert until an image roll.**

Edit-6 write-back done: `cta_link_integrity/PLAN_2026-07-19` now records the **4th call site**; a
`doc_notes` row was inserted **live** (`subject_type='pipeline'`, `subject_key='cta_link_integrity'`,
`categories ? 'council-gate'`), idempotent via `WHERE NOT EXISTS`.

- `[UNMEASURED]` after the roll: the **~30 active ungated URL-placeholder components** (count measured
  2026-07-20) will newly log at **Error** fleet-wide when their URL fields are unresolved. That is the
  accepted observability cost — a log-volume change, NOT a behaviour change (the content was already
  blanked). Watch Error volume after the first roll; it is the signal `054` will escalate on.
- `[ASSUMED, vetted in plan]` vonc.com may newly render a logo image (its `logo_url` was empty and
  `header-bold-gradient` sources it from `site_assets.logo`) where it currently shows an `{{else}}`
  glyph — an accepted, reviewed side-effect (a fix), not a regression.

### §X.9 — LIVE & VERIFIED on v1.0.1146; 018 + 041 CLOSED (2026-07-21)

Owner deployed v1.0.1146; `36829b07b` rode it. Verified against the **running pod**, not the tag:
```
POD=agent-chassis-55bbccfdbc-xrkv6  (binary /app/agent-chassis, Jul 21 12:07)
RenderTemplateReportingMissing → 2   URL attribute rendered empty → 1   missingBareFields → 2
Cleaning up <no value>        → 0   (the DELETED string — the decisive discriminator)
Render context built          → 1   (positive control, grep works)
```

**041 verified END-TO-END** (behavioural, not just pod-grep). Chrome JS publishes only on a page
rerender, so I inserted ONE assemble-only `page_rerender` (`created_by='bugfix-018-chrome session
verify041'`, spec has NO `reason` → assemble-only path, NO LLM, markup unchanged). Item went
`triaged → claimed 15:53 → complete 15:54`.
- **Trap re-lived, then resolved:** at 15:54 (item complete) `site-header.js` was STILL 404 for ~1
  poll — looked like "complete ≠ done" again. It was **VM pull-sync lag**: idea.uk serves from the
  box, `sitesync.timer` every 5 min, so there is a lag between chassis-publish and box-serve. Re-checked
  a few minutes later → **200**. Do NOT conclude failure at the moment of `complete` on a VM-served
  site; wait one sync cycle. (Belongs in the "complete≠done" family but the cause is propagation, not
  a no-op render.)
- Final state: `site-header.js` **200** (708 B, the real `.hamburger`/`aria-expanded` handler),
  `site-footer.js` **200**; homepage `<script src>`-references both; homepage `href=""` count **0**
  (no chrome regression from the rerender).

**018 + 041 → `bugs_closed/`** (both fixed AND live AND verified). Follow-on `054` stays OPEN
(block/escalate + consumer). No `Council-Reviewed` trailer on `36829b07b` (round 2 REVISE, owner
overruled). Council round 3 NOT fired — the owner overruled its sole remaining objection, so a verdict
would change nothing; the round-3 JSON stays prepped if the audit trail is later wanted.

### §X.10 — bugs_open/054 IMPLEMENTED: drop-the-control + drain (2026-07-22, bugfix o54 session)

The follow-on `054` the owner scheduled off §X.8's ruling. Two owner rulings this session
(AskUserQuestion): **(1)** mechanism = **render-path drop-the-control** (not per-template gating,
not a hard build-block); **(2)** `bugs_open/033` **is a queue** — so the escalation goes to a
**draining** pathway, not the `needs_human_review` void that `023`/Constraint 2 warned against.

**Grounded first (and it reframed the job):** live 2026-07-22, **0** `site_components` render an
empty `href`/`src`; all 10 live-placed chrome components are gated; the 7 ungated chrome census
components (`*_pre_037`/`site-head`/`header-docs`) have **0 placements**. So the acute idea.uk fire
is out and a chrome-scoped mechanism has ~0 live blast radius — the change is **preventive**, inert
until an image roll. Constraint 1's "can't gate cold" is largely moot for this path.

**Built (commit `524b03f03`), both gated on the already-computed `deadURLFields`:**
- `DropDeadURLControls` (new file `drop_dead_url_controls.go`) — removes any anchor whose `href`
  rendered empty, blanks any empty `src`, from the rendered chrome before store (LNK-005). Wired
  into `renderAndStoreSiteComponent`. 16 unit cases incl. negative boundaries (`<area>`/`<abbr>`
  excluded, `href="#"`/non-empty attrs untouched).
- `emitChromeDeadControlItem` — files one `chrome_dead_control` item at `detected` +
  `nav-link-fixer`/`build` (the `phantom_internal_links` site_component convention), deduped by
  `item_key` against `idx_swi_dedup`. Triage promotes it, dispatch drains it; a re-render fixes the
  data-lag case, a persistently-dead field exhausts `max_attempts` → surfaces to the human queue.

**Misstep avoided, logged:** `component_library.go` was under **active concurrent edit** by another
session (a template-scanning feature — `scanTemplateFuncs`/`bareFieldName`). I first put the helper
there; caught it via `git diff --stat` (130 insertions, only ~10 mine) + a transient `go vet`
"text/template unused" (their mid-edit state). **Moved the helper to its own file** so the pathspec
commit excludes the contended file and takes no same-file passenger. The general lesson (re-derive
which insertions are yours before committing a file two sessions touch) is why the helper lives in
`drop_dead_url_controls.go` rather than beside `RenderTemplateReportingMissing`.

Council review in flight (advisory), `SUBMISSION_CORR=f54a1808-51a6-4ddd-8f60-783f9b263e37`. No
`Council-Reviewed` trailer on `524b03f03` (committed before the verdict, per the commit-per-task
rule; trailer is earned only by APPROVED). `054` stays OPEN until the change is live AND verified on
the failing branch; the general pre-existing unread pile stays `033`'s.

### §X.9 — bugs_open/017 (static-cutover) CLOSED & LIVE on v1.0.1149 (2026-07-22)

Owner rolled v1.0.1149. Verified `backend_entry_orphaned` is in the running binary — pod
`agent-chassis-7d4ff8b54-cm786`: `strings /app/agent-chassis | grep -c BackendEntryOrphanedCheck` = 10,
`handlerRouteCandidate` = 2, positive control `DeadControlsCheck` = 10 (grep discriminates). Applied
**seed 188** (snapshot `b05773e0` taken; `DO` appended the name; doc_note written; COMMIT) → the
completeness-discovery-agent checks array now carries `backend_entry_orphaned` (24 checks). Detection
was already proven on the induced 405 (live probe + httptest); NOT observed inside a live sweep (no
committed discovery-trigger, ~30 min queue, and zero findings expected today). Moved the case to
`bugs_closed/`, updated 016b §10. Cause-guard split to `features_open/011` + RUNBOOK §3d(ii); Finding
B (`no_backend_entry`) deferred pending a backend-presence signal (the empty `deploy_config` gap).

### §X.8 — COUNCIL ROUND 3: REVISE, owner ruled SHIP, parse-tree refinement LIVE on v1.0.1149 (2026-07-22)

Round 3 verdict (2026-07-21 11:17): **REVISE**, decided_by `bug_historian` (non-veto). 11/13 approve;
only `bug_historian` + `guardian` object, both explicitly "not a veto". `abstained: 2`;
`review_editquality.result` came back **unreadable** this round.

**The handoff predicted "round 4 is wording, not evidence." That was WRONG, and I checked before
acting on it.** Ran the objections' read-only checks against the real code:
- `guardian` (enumerate `RenderTemplate` callers): **~8 call sites across 5 files / multiple pipelines**
  (chrome, assemble_from_library, rerender_page_sections, section_editor ×2, v3_site_actions, + 3
  internal). "Large and diverse" is real — the fleet-wide log-behaviour change guardian feared is real.
- `bug_historian` (audit for a sibling silent-drop path): **FOUND ONE** — `RenderTemplateWithMap`
  (`rerender_pages_actions.go:757`) is a second, independent renderer that does its own parse/execute
  and (pre-fix) left `<no value>` visible, logging only on error. So the round-2 claim "fixes THE
  MECHANISM" was incomplete.
- `guardian`+`bug_historian` (regex control-flow blindness): **CONFIRMED in the working-tree code.**
  `missingBareFields`'s flat `bareFieldRe` matches a bare `{{.Name}}` textually even inside
  `{{range}}`/`{{if}}` bodies, then tests it against the TOP-LEVEL map → false-positive. Per §X.7 the
  URL-bound placeholder count is 71/59-active/**30-ungated** — so up to ~30 active components could log
  a false Error on every render. That is exactly the noisy channel `bugs_open/054` is meant to escalate
  on. [Was `[UNMEASURED]` as a "still a heuristic, risk 5" line; now measured against the parse tree.]
- `guardian` (collectJSAssets other callers): **one call site** (`rerender_single_page:121`) — safe.

**Owner ruling 2026-07-21 (already recorded for round 2's scope fork): ship the observability version,
do block/escalate as `bugs_open/054`.** Owner ruling 2026-07-22 (this thread): **ship on the standing
ruling — no round 4.** Council is advisory; it stays at REVISE, no `Council-Reviewed:` trailer (earned
by APPROVED only — same posture `bugs_open/053` took).

**What shipped (commit `78482c86b`, 2026-07-22 11:07 UTC):**
1. Rewrote `missingBareFields` as a **scope-aware `text/template/parse` walk** — reports only ungated,
   root-scope bare fields; a field inside `{{range}}`/`{{if}}`/`{{with}}` is NOT reported. The flat
   regex is kept as `missingBareFieldsRegex`, the fallback for templates that won't parse as Go.
   `an.Pos` points just INSIDE the `{{`, so back up via `strings.LastIndex(tpl[:an.Pos], "{{")` before
   testing `urlAttrRe` (debugged empirically — my first cut had the href/src test failing on offset 11
   vs the `{{` at 9). Test `missing_bare_fields_test.go` locks all scope cases + the fallback.
2. Routed the sibling `RenderTemplateWithMap` through the same detector + `<no value>` strip.

**VERIFIED LIVE on v1.0.1149** (pod `agent-chassis-7d4ff8b54-cm786`, started 2026-07-22 13:56 UTC,
built AFTER my 11:07 commit). Discriminating pod-grep (symbols created by ONLY my commit,
`git log -S` confirms):
- `missingBareFieldsRegex` = 2, `bareFieldName` = 1, `scanTemplateFuncs` = 1 → **parse-tree fix is in
  the binary and on the live `RenderTemplate` path.**
- positive control `RenderTemplate: URL attribute rendered empty` (base fix) = 1.

**CAVEAT, recorded honestly:** the sibling half has **zero runtime effect today.** Pod-grep for
`RenderTemplateWithMap` = **0** — the linker dead-code-eliminated it, because its only caller
`rerenderContactInfo` (`rerender_pages_actions.go:660`) has **no callers of its own** (`git grep`
confirms). So the contact-info render path is currently unreachable. My fix makes it correct IF it is
ever wired back up; it is not exercised now. Do NOT claim "sibling fixed and live" — claim "sibling
made correct; path is dead code today". [This is the honest read of bug_historian's audit: the sibling
mechanism exists in source but is not a live silent-drop path in the current chassis.]

**No build/roll needed from me** — my commit was already in HEAD when another session built v1.0.1149,
so it rode that roll (owner: "a new chassis image is on production"). The makefile tag churned
1147→1149 under three sessions during this thread; I never claimed a tag.

**Side-finding still open** (a reviewer's check, §X.6): `sites.content_data` for idea.uk still carries
`"email": "idea-uk@leopardess.uk"` — not rendering today (footer takes `email` from the render
context) but one code path from a live wrong address. Not fixed this thread.

### §X.11 — bugs_open/054 council: REVISE → routing fix → APPROVED (2026-07-22, bugfix o54 session)

Update to §X.10. The council REVISE'd the first plan (corr `3951e2be`) and it was the right call: it
found a real hole in **my** Half-2 routing (not the owner's intent). I had routed the finding to
`detected`+`handler=nav-link-fixer`; that agent's live workflow (`fix_nav_templates` → rerender →
`complete_workflow`) marks the item **complete without verifying the field resolved**, so a
genuinely-unresolvable chrome control would be re-dropped and marked done on every re-render and
**never reach a human** — exactly the "silent loss with a hopeful side-channel" bug_historian named.

Owner confirmed the reroute (AskUserQuestion): **`needs_human_review`, no handler**, mirroring the
sibling `check_dead_controls`. A dropped control is a human decision; the data-lag case self-heals on
the next normal re-render regardless. Revision `0132f859b` also switched the emit to the SHARED
`insertWorkItem` helper (the `idx_swi_dedup`-matched `ON CONFLICT` + two-strike label — house
standard, answers reuse_agent/guidelines) and made a dead `<img>` drop whole (editquality).

**Round 2 APPROVED** (4 advisory, none high). Acted on `render_guardian`'s `data-runtime-fill`
exemption (`2afa6531a`, carries the `Council-Reviewed: 3951e2be` trailer). The rest were the seats'
partial-schema inability to see `site_components` (it exists, 36 rows) and a preference for attached
code_check bundles over manual grep.

Trigger schema gotchas learned (both died `complete_invalid`, no credits): `plan.risks` is a STRING
not an array; edit `operation` ∈ `modify|add|remove|config_change`, new files = `add` not "create".

**Cross-dependency with your own §X.8 work:** my drop is gated on `deadURLFields` = `missingBareFields`'
`inURLAttr`. Until `78482c86b` (your scope-aware parse-tree detector, LIVE v1.0.1149) that was
control-flow-blind — dropping on its false positives would have removed live gated controls. It's an
ancestor of my commits, so next-build-from-HEAD is safe. That fix was also the concurrent
`component_library.go` edit I stepped around. `054` stays OPEN until live + verified on the failing branch.

### §X.9 — Home-page CTAs don't reach the paid tool: diagnosis + fix (2026-07-23)

**Owner report:** "the links on the home page still don't go to the paid for tool."

**Diagnosis (read the live page + DB, not theory).** `/report.html` IS the paid £29 tool — `curl`
shows `<form class="rr-form" action="/request" method="POST">`, "Request a report", `£29` ×2, the
honeypot `company_url` + `_elapsed` timing field. So the destination exists and works. The home page
(`index`, page_id `b147b925-…`, 6 sections) simply doesn't point at it. Live hrefs on the prominent
CTAs: `[Get Started]→/contact.html`, `[See how it works]→/contact.html`, `[Browse All Tools]→
/contact.html`, plus two dead `#`. **Root cause: the four CTA sections were never link-resolved** —
`page_components.content_data` has the button LABELS but no `cta_url`, and `resolved_data` is NULL on
all four. So each URL field fell back:
- hero `{{.cta_url}}`/`{{.secondary_cta_url}}` (`source: renderer`) → `contextToInterfaceMap`'s
  `/contact.html` default (`component_library.go`).
- brief-explanation → **`href="#"` HARDCODED in the shared template** (no url field declared at all).
- tool-list `{{.cta_url}}` (`source: query.section_index_for:tool`) → the query finds no tool
  section-index for idea.uk (tools is `page_type='content'`, not `section-index`) → nil → default.
- call-to-action (`source: renderer`) → skip_field.

**The deeper reason it can't self-heal:** the fleet-wide resolver (`resolve_internal_links` /
`applyCTARecompute`, owned by the cta-link-integrity workstream) only ever points CTAs at content
HUBS (`page_type='section-index'`, excluding about/contact/legal). `/report.html` is `page_type=
'landing'`, so the generic resolver would NEVER choose it. Funnelling the home page to the paid tool
is idea.uk **business intent** that has to be set explicitly — this is not a resolver bug.

**Fix (`sql/p3_05`, owner-confirmed mapping via AskUserQuestion 2026-07-23 — all tool CTAs → the
paid page):** set the url fields in each section's `content_data` (hero cta+secondary → /report.html;
call-to-action primary → /tools.html#audience-check, secondary → /report.html; tool-list cta →
/tools.html; brief pri+sec → /report.html) + a GATED edit to the shared brief-explanation template
(`href="{{if .cta_primary_url}}…{{else}}#{{end}}"`, backward-compat no-op for the 3 other instances —
verified none sets those fields) + a `reason='section_data_resolved'` rerender of the index only (no
CTA recompute, no LLM; NULL-content_data guard passed 0). Read-back confirmed all four content_data
writes. **Mechanism that makes it hold:** `applyCTARecompute` runs ONLY for `cta_links_stale`
(`rerender_page_sections_action.go:287`), so a `section_data_resolved` rerender uses stored
content_data; `contextToInterfaceMap` defaults `cta_url` then the ContentData merge OVERRIDES it;
`mergeIntoRenderContext` captures ALL content_data keys (`v3_site_actions.go:1465`); the tool-list
query returns nil so the set value survives ("assigns any NON-nil value").

**DELIVERY — the page_rerender queue was DEAD** (nothing completed in ~15h; the bugs_open/029/030
stalled-dispatch). The p3_05 work item sat `triaged`. Direct-fired a `section_data_resolved` rerender
via Kafka (the 049b pattern + `input_data.spec.reason` — the workflow's `check_rerender_mode` reads
`input_data.spec.reason`, NOT `input_data.reason`). Orchestration `335739d2-…` → COMPLETED 11:41.
(The queue later revived and completed the p3_05 work item too — a harmless idempotent 2nd rerender.)

**VERIFIED LIVE 2026-07-23** (`curl https://idea.uk/`): all four body CTAs correct —
`[See how it works]`/`[Request a verified idea report]`→/report.html (hero);
`[Get Started]`/`[Learn More]`→/report.html (brief-explanation, the dead `#` gone);
`[Run the free idea check]`→/tools.html#audience-check, `[See what a verified idea report contains]`→
/report.html (call-to-action); `[Browse All Tools]`→/tools.html (tool-list). **Sections LOCKED**
(`sql/p3_06`, lock_type=permanent) + a `doc_note` warns cta-link-integrity off a blind cta_links_stale
recompute.

**ONE REMAINING, and it is CHROME not body — flagged to owner:** the header/mobile "Get Started"
button (`site-header` component, `class="btn-primary"`, on EVERY page) still → `/contact.html`. It is
a template-var CTA resolved from the chrome value map (sites has no cta_url → /contact.html default,
the LNK-007 fossil). Fixing it → /report.html is a SITE-WIDE chrome change (all 9 pages), broader than
"the home page", so it is an owner decision (asked 2026-07-23). The `[Contact]` nav link correctly
stays /contact.html.

**§X.9 addendum — header CTA DONE & VERIFIED LIVE (2026-07-23/24).** Owner chose to point the header
"Get Started" at the paid tool too. `cta_url` is source=renderer (hard-set to the contact page in
`render_site_components_action.go:155-156`, and the renderer does NOT read `site_components.content_data`
— no per-site data override), so edited idea.uk's OWN `site-header` component (`f420f3fa-…`, 1 site) —
`href="{{.cta_url}}"` → `href="/report.html"` on both anchors (gate preserved; snapshot
`bak_ideauk_header_20260723`); `sql/p3_07`. Fired the chrome refresh
(`agent_type=rerender-pages, refresh_site_components:true`, CORR `22c993cb`, message PRODUCE-verified on
the topic) → re-renders chrome + reassembles all 9 pages. Orchestration COMPLETED; `site_components`
header rendered_html carries `/report.html`. VERIFIED LIVE: header `btn-primary` → /report.html on
`/`, `/about.html`, `/tools.html`, `/report.html`, `/guides/index.html`; the only remaining
`/contact.html` on the home page is the correct `[Contact]` nav link. **The idea.uk CTA funnel now
drives to the paid tool end-to-end (body + header).**

### §X.12 — the ideas pipeline starts: idea.uk's FIRST guide (patents) built & LIVE (2026-07-25)

Owner's ask, captured 2026-07-24 as `features_open/014`: idea.uk should grow from "marketing site
+ one paid report" into a **guided journey across the whole life of an idea** — guides and free/paid
tools from ideate → build → test → UAT → feedback → **patents** → copyright → funding ways → funding
sources → more. Patents is the owner's explicit lead. Owner then said (this session):
*"Let's carry on with idea.uk specifically in this thread"* → so this is increment 1, built.

**GROUNDING FIRST — what the site actually had.** Nine pages. `/guides/index.html` served 200 but
was an EMPTY SHELL: hero + a `content-listing` section rendering **601 bytes** (heading, no cards).
`/news/index.html` the same. **Zero pages of `page_type='guide'` anywhere on the site.** So this is
genuinely idea.uk's first guide, not an addition to a library.

**THE MECHANISM (grounded before building, not assumed).**
- Fleet precedent for the page shape: `/guides/<slug>/index.html`, `page_type='guide'`, hero +
  Generic Text Block — gamesdesign.co.uk ×5, relojistas ×4, vetcomparison ×3.
- **The hub populates itself**: `guide-list_pre_037` (`9d5e461a-…`) sources `items` from
  `query.pages_where_type:guide` → `resolvePagesWhereType` (`queryresolve.go:81`). Verified by
  reading gamesdesign's stored `content_data`: five resolved item objects, 7,758 bytes rendered.
- **Eligibility gates it**: that resolver applies `FetchablePageEligibilitySQL` —
  `deployed_at IS NOT NULL OR build_status='deployed'`. So the guide must SHIP before the hub can
  list it. Ordering is forced: guide first, hub second.
- `Generic Text Block` renders `{{.content}}` **unescaped** — verified against the live
  `/guides/rng-design/index.html` artefact, so authored HTML in `content_data` is safe.
- Deliberately did **NOT** re-run `build-site-planner` to compose the page (the RUNBOOK Phase 1
  route). Re-planning to add one page is how built pages get clobbered (`bugs_open/001`, `050`).

**AUTHORED, NOT GENERATED — and why that is a decision, not fussiness.** The body is hand-written
UK patent guidance. The platform's evidence gate for factual claims (claims-verification **V5**) is
BUILT BUT INERT, and `bugs_open/043` is a live fabricated-content lane. An LLM pass here would not
just change tone — it would emit legal assertions with nothing behind them, on a live commercial
site, under a heading inviting reliance. `sql/p4_01` sets every field of every section so a
`section_data_resolved` rerender has no reason to escalate to the content writer.

**A CORRECTION MADE WHILE DRAFTING, recorded because it would have shipped.** The first draft said
the **IPEC small claims track** makes patent enforcement affordable for small businesses. It does
not — IPEC's small claims track expressly does **not** hear patent, registered design or
semiconductor topography claims. Corrected before writing to the DB: IPEC *multi-track* (costs and
damages caps) plus the IPO's non-binding opinions service. Nothing caught this but re-reading my
own draft against what I actually know; that is not a reliable control, which is the argument for
V5 rather than against authoring.

**MISSTEP 1 — `slot_name` NULL: rendered nothing, reported COMPLETED.** `p4_01` inserted the three
sections with `component_id` set and `slot_name` NULL. The rerender ran and the workflow COMPLETED:
```
rerender_sections -> {"section_count":3,"rerendered":0,"carried":3}
render_page       -> {"skipped":true,"reason":"no components found for page"}
workflow          -> complete_skipped, status COMPLETED
```
Cause: `rerender_page_sections_action.go:249` looks components up as **`schemas[s.slotName]`** —
keyed on `slot_name`, not `component_id` (`loadStoredSections` reads `COALESCE(slot_name,'')`).
A NULL misses the map and takes the *"component not found, carrying stored HTML"* branch (`:251`);
on a new page the stored HTML is empty, so assembly found nothing. **The key is the component's
`function` column, not its `name`** (`Generic Text Block` → `generic-text-block`) — verified across
the fleet's existing guide pages and idea.uk's own home page. Fixed by `sql/p4_01b`
(`SET slot_name = cc.function`, so it cannot drift from the lookup key). This is
"trust the rendered artefact, not the status" in its purest form: **nothing failed.**

**MISSTEP 2 — called a queued dispatch a dropped spawn.** The hub rerender produced no
`orchestration_states` row for ~4 min; I called it a lost spawn (`bugs_open/003`), re-fired, and
wrote it into the RUNBOOK as a standing trap. Wrong. The message was on the topic the whole time
(read at offset 103566), and `orchestration_states` had **zero rows fleet-wide** since 08:45 — the
generic-requests consumer was stalled (`bugs_open/029/030`), every thread waiting. RUNBOOK TRAP 3
rewritten to the correct check (*are ANYONE's orchestrations starting?*) and logged in
`WRONG_CALLS.md`. I already had this note — `memory/council-queue-latency-trap`, written after the
same mistake against the council gate. It did not fire, because it is filed under "council" and
this was a page rerender. Also learned in passing: "no `agent-page-rerender` pod running" proves
nothing — they are one-shot Jobs that idle-shut-down after ~3 min (`agentbase/agent.go:1541`,
observed `idle_duration 184s`).

**RESULT — /guides/patents/index.html LIVE 2026-07-25 08:42 UTC.** `curl` verified, not job status:
HTTP 200, 39,214 bytes, `<title>Patents: how to protect an idea in the UK</title>`, h1 + 8 numbered
h3 sections + the closing CTA, full site chrome (header/footer nav present). Every CTA on the page
funnels correctly — `/report.html` ×4 (incl. the header "Get Started" from `p3_07`),
`/tools.html#audience-check` ×1, `/guides/index.html` ×1. Committed to `vm-sites` (`b253d868`, then
`b78f70c4` — a *deploy-step retry*, identical content, because the first git-adapter success
response was not consumed; "two commits" is not evidence of two edits). `build_status='deployed'`,
`deployed_at` stamped 08:40:05.

**RESIDUAL FIXED — `pages.sections` was left `[]`** (`sql/p4_01c`). The rerender path does not write
it: `save_page_sections` reported `{"sections_found":3,"sections_saved":3}` but that saves
`page_components.rendered_html`, not the page-level slot list. Every other fleet guide has it,
populated by the original build path this page bypassed. Not cosmetic: `ListedPageEligibilitySQL`
requires `jsonb_array_length(sections) > 0` and is the SHARED literal behind both the article
listing and the imagery sweep, so an empty array makes a deployed page invisible to that whole
contract — and a future "deployed but no sections" sweep could try to rebuild it, over authored
legal content. Backfilled to `["hero","generic-text-block","call-to-action"]`.

**HUB SWAP (`sql/p4_02`) — applied to the DB, render PENDING at time of writing.** idea.uk's guides
hub carried `content-listing` (`aa3e4b68-…`), whose `articles` is a **static array with no query
source** — it had always been empty and would *never* have picked up a guide page, no matter how
many were added. That is the derived-vs-static shape of `bugs_open/023` inverted: a listing that
cannot see the pages it exists to list. Swapped to `guide-list_pre_037`, moving `slot_name`
**and** `pages.sections` with it (`["hero","content-listing"]` → `["hero","guide-list"]`), with the
surrounding copy authored and `items` deliberately left unset so it stays query-resolved. Snapshot
`bak_ideauk_guideshub_20260725`. **[UNVERIFIED at time of writing]** the rerender is queued behind
the stalled consumer; the hub's `guide-list` section still holds the old 601-byte HTML until it
runs. Verify with `curl https://idea.uk/guides/index.html` and look for the patents card — do not
trust the DB swap as proof the page changed.

**STILL TO DO on this increment:** `sql/p4_03` (lock the authored sections — written, waits on the
hub being verified live, per its own guard).

### §X.13 — guide 2 (copyright) + idea.uk's first FREE tool (patent checker) (2026-07-25)

Owner, same session: *"yes for copyright and the checker"*. Both built. §X.12's recipe held — no
new traps, and the two it documented (slot_name, pages.sections) were handled up front in the SQL
rather than discovered again, which is the whole point of writing them down.

**§X.12 LOOSE END CLOSED FIRST.** The guides-hub CTA rerender that was queued at the end of §X.12
landed. VERIFIED LIVE: `<a class="guide-list-cta-btn" href="/report.html">Get a verified idea
report</a>`. Both hub rerenders from that queue ran, **including the one I had called a dropped
spawn** — so the WRONG_CALLS entry is confirmed correct by the system's own behaviour, not just by
my re-reading of it.

**GUIDE 2 — `/guides/copyright/index.html`** (`sql/p4_05`, nav_order 20 so it sorts after Patents).
Hand-authored, same policy. It leads on the thing that actually costs small businesses money and is
invisible from the patents guide: **a contractor keeps copyright unless there is a written, signed
assignment** (CDPA 1988 s.90(3)); only employees' work vests in the employer automatically
(s.11(2)). So the freelancer who built your site owns your site. Two further deliberate choices:
- The AI section says the law is **UNSETTLED** — whether AI output attracts copyright, who would
  own it, and whether training on copyright works is lawful — rather than picking a side. Writing a
  confident answer there is precisely the fabrication the authored-content policy exists to prevent,
  and it would have been the easiest paragraph in the guide to write wrongly.
- It states that copyright **can** use the IPEC small claims track. That is the exact counterpart of
  the error caught in §X.12's draft (which wrongly said patents could). Getting the pair right in
  both directions is the useful outcome of that correction.

**FIRST FREE TOOL — `/tools/patent-check/index.html`** (`sql/p4_06`), new `patent-check` component
(`37f2ca9c`). This is idea.uk's first genuine Tier-1 free probe: the existing
`/tools.html#audience-check` is a pointer at the paid tool's own backend, not a free self-contained
thing.

**REUSE WAS CHECKED AND REJECTED, with a reason worth keeping.** `ai-readiness-quiz` (`71a636a7`,
2 live instances) is the obvious candidate — a client-side 5×4 questionnaire with result tiers. It
is the wrong instrument, and not cosmetically. **It is a sum-score quiz, and patentability is not
additive.** "Have you already disclosed it publicly?" is close to dispositive on its own — the UK
has no general grace period — so under a sum, someone who has already published would score well on
the other five questions and be told they look patent-ready. That is not an imperfect UX; it is
advice that could cost a reader their rights. Same for subject-matter exclusions: a business method
is excluded however strong the commercial case.

So `patent-check` is **gated, not scored**: disclosure and subject-matter short-circuit to their own
outcomes first, and only if both pass does the commercial question (prior art, detectability,
ability to fund enforcement, shelf life) get scored into three bands. The reasoning is commented in
the template's own script so a later reader does not "simplify" it back into a sum. (Secondary
reason reuse would have been awkward: the quiz's `quiz_badge_label` is `source: static` WITH a
fallback — the p4_04 defect — so the badge would have read "AI Readiness Assessment" on a patent
checker, unoverridably.)

**Two delivery decisions, both grounded rather than assumed:**
- **JS inline in `html_template`, not an external `/tools/assets/*.js`.** The quiz references an
  external asset, which rides the publishing path that produced `bugs_open/041` (chrome JS published
  but never loaded) and the `js_content`-vs-`js_snippets` trap. An inline `<script>` is part of the
  rendered section HTML and cannot be published-but-not-loaded. Verified the template parses and
  executes under `text/template` **before** inserting it (0 `<no value>`, 0 unreplaced actions,
  `</section>` present so it passes `componentTemplateValid`) — cheaper than finding out via a
  carried section.
- **URL `/tools/patent-check/index.html` — checked against nginx first.** idea.uk's box reserves
  the tool binary's routes by **exact match** (`location = /request` …), with only `/stripe/`,
  `/internal/` and `/order/` as prefixes. `/tools/` is NOT reserved. Probed live to be sure:
  `/tools/assets/site-header.js` → 200, `/tools/anything-random/` → 404 (a static miss, not the Go
  binary). Had `/tools/` been proxied, the page would have been invisible no matter how well it
  rendered.

Schema fields are all `source: static` with **no fallback** — the shape `p4_04` established is
required for `content_data` to win. `page_type='tool'` means `/tools.html` lists it automatically
(`tool-list` sources `query.pages_where_type:tool`), the same self-listing mechanism `p4_02` gave
the guides hub — so the same ordering constraint applies: ship the page, then re-render the hub.

**RENDER RESULT:** copyright `rr=3/carried=0`, tool `rr=2/carried=0`, both COMPLETED, both
`build_status='deployed'` with `deployed_at` stamped, and both files confirmed in the git-adapter's
commit to `vm-sites`. **[UNVERIFIED at time of writing]** both still 404 on the live box — VM
sitesync is a 5-minute timer, so this is the known lag, not a failure. Waiting for the 200 before
claiming either is live, and before running `p4_07`/`p4_03`.

**QUEUE NOTE:** these two renders were picked up in under a minute, against ~12 minutes for the
three fired during the §X.12 stall. Same fire mechanism, same payload shape — the difference is
entirely consumer backlog. Worth remembering next time a dispatch looks slow: the latency is not a
property of your message.

**STILL TO DO:** `p4_07` (cross-links: patents guide → checker + → copyright; then re-render
/tools.html so its derived tool-list picks the new tool up) and `p4_03` (locks, extended to cover
all four new pages).

**§X.13 addendum — cross-links, locks, and a lock that was nearly a silent time bomb.**

`p4_07` wired the pages together (patents guide → checker + copyright guide; /tools.html re-rendered
so its derived `tool-list` picked the new tool up — no content edit, only a render). Both clean
(`rr=3/carried=0`, `rr=4/carried=0`). VERIFIED LIVE: `/tools.html` carries
`href="/tools/patent-check/index.html"`, the patents guide carries links to both the checker and the
copyright guide, and all six pipeline pages return 200.

`p4_03` then locked the authored sections — and here is the misstep, the third of the day and the
worst of them because it would not have announced itself.

> **CORRECTED 2026-07-25 — `p4_03`'s own comment was FALSE.** It said: *"the items themselves stay
> query-resolved and must NOT be frozen — the lock protects the surrounding copy, not the derived
> list."* A lock is applied to the `page_components` **ROW** and cannot separate authored copy from
> derived items, because both live in that row. `SavePageSectionsAction`
> (`save_page_sections_action.go:487-534`) preloads actively-locked rows, holds them out of the
> rebuild DELETE and re-attaches them **verbatim** — the code's own comment is *"Human-locked rows
> must survive the rebuild with copy AND row identity"*, and it logs *"preserving human-locked
> section over rebuilt copy (bugs_open/058)"*.

So for about ten minutes the guides hub's `guide-list` was locked, which would have **frozen the
listing at one card permanently**: every future guide written, deployed and silently never listed,
each render reporting success. The self-populating listing is the entire reusable contribution of
increment 1, and locking it would have killed it on the day it was built.

**What caught it was luck, and it is worth being precise about that.** The final sweep showed the
live hub with one card while two guides existed — legitimately, because copyright shipped ~25
minutes after the hub's last render. That gap made me ask why. Had I locked before writing a second
guide, or looked five minutes earlier, one card and one guide would have looked perfect. A frozen
derivation is indistinguishable from a working one until the data it tracks moves on.

**THE RULE, and it generalises well beyond this page:** *never lock a section whose component
schema has any `query.*` source.* Locks are for AUTHORED sections only. Protect a deriving section
by making its authored fields content_data-driven — which `p4_04` had already done here — so
nothing wants to regenerate them; guard the inputs, not the output. `p4_08` unlocks the hub (with a
guard that refuses unless the section really does derive) and records the rule as a `doc_note` under
`component_locks`. Logged in `WRONG_CALLS.md`.

FINAL LOCK STATE — 8 authored sections locked (patents ×3, copyright ×3, patent-check ×2), and both
deriving listings (`guide-list` on the guides hub, `tool-list` on /tools.html) deliberately free.

**Pattern across the day's three missteps, worth naming:** all three were confident statements about
platform behaviour made from intent rather than from the code or a query — the carried section
(slot_name), the "dropped" spawn, and this lock. Each was one grep or one query away. The code was
on disk in every case.

### §X.14 — owner's tools-page report: four defects, all confirmed; the paid-tool audit; funding pair built (2026-07-25)

Owner: *"The structured diagram is not showing on the tools.html page. The paid-for tool doesn't
show on the tool listing. The 'Try the tools' and the 'Browse All Tools' buttons go to the contact
form... Has the paid-for tool changed? Does it do what we say it does?"* Plus: *"please go ahead"*
(the funding pair) and *"write up the docs to this point including the summary"*.

**TOOLS-PAGE DIAGNOSES — every one grounded before fixing (`sql/p4_09`):**
1. *"Try the tools" → /contact.html*: the tools.html hero had `cta_text` but NO `cta_url` — the
   LNK-007 default. p3_05 fixed exactly this on the home page; this page's sections were never
   link-resolved. Its secondary CTA had no URL at all, so the second button was gated out.
2. *"Browse All Tools" → /contact.html*: `tool-list.cta_url` is `query.section_index_for:tool` →
   nil for idea.uk → default. AND the label is the p4_04 defect again — `source:static` WITH
   fallback `'Browse All Tools'`, unoverridable. Dropped the fallback (guard-verified no-op: all
   6 instances across 4 sites carry the value in content_data), then set label+URL to the funnel
   ("Get a verified idea report" → /report.html) — a "Browse All Tools" button ON the tools page
   is a self-reference whatever URL it gets.
3. *The missing diagram*: rendered `<img src="/assets/images/illustration.jpg">` — a file that
   does not exist (live 404, absent from vm-sites). The site HAS a purpose-made live asset:
   `/assets/images/illustration-tools.jpg` (assets row `illustration_tools`, HTTP 200). Set
   `illustration_url` in content_data. CAVEAT flagged in the SQL: the field is
   `source: site_assets.illustration` and resolved_data merges last — if the resolver re-emits
   the dead path on rerender, the fix moves upstream. Verify the rendered src, not the DB.
4. *Paid tool absent from the listing*: `tool-list` items derive from
   `query.pages_where_type:tool`; /report.html was `page_type='landing'`. Flipped to `'tool'`
   after checking every consumer: nav untouched (keys in_header/nav_order), eligibility passes,
   sections-carrying tool pages are legitimate (13/33 fleet-wide), and the CTA resolver treating
   it as interactive is what this site wants. Also set its EMPTY meta_description (the card would
   have rendered blank) and gave the audience-check pointer a real nav_label + description.
5. **UNPROMPTED FIND — a fabricated stat on the live page**: brief-explanation claimed
   **"8 Tools available free"** (there are 2) and **"Data stays on your device — Always"** (false
   for the audience check, which posts to the server, and for the paid report). The 043 class, on
   the page whose pitch is honesty. Corrected to true values; "2" costs us a bump per new tool,
   which is the price of it being true. tool-list's section_intro also claimed per-card
   browser/server labels that don't exist, and cta_supporting_text said "Each tool is free" —
   false the moment the paid report joined the list. Both reworded.

**THE PAID-TOOL AUDIT** (owner's direct question; full table in
`AUDIT_2026-07-25_paid_tool_vs_copy.md`). Read the actual source (engine.go/prompts.go/
service.go/billing.go) against the live /report.html copy. **The tool has NOT changed** — it is
and always was the ideation method v2: generate 12–24 AI-product ideas for the customer's
business across five lenses → cross-vendor cut → REAL web-search verification → gated scoring
with a separate operator-risk axis → ranked report incl. "didn't make the cut" and "set aside on
risk", with human review before every run and honest refusal outcomes ("No idea cleared the
bar"). **The COPY describes a different product**: "produced for a single idea you submit" (the
engine has no your-idea input — its own intro says "You asked us to find AI product ideas for
{domain}") and "where we cite a figure or a claim, we explain its source so you can check it
yourself" (the report renders findings with NO sources; the verify prompt explicitly suppresses
names). Delivered honestly: refusal, human review, £29 Stripe, competitor checking, cheap-test
per idea, AI disclosure in the T&Cs (not in the report itself). Also caught:
`reportContact()` falls back to the stale `idea-uk@leopardess.uk` unless CONTACT_EMAIL is set on
the box. Fix direction (copy vs engine vs both) = owner decision, presented.

**FUNDING PAIR BUILT** (`sql/p4_10`, `p4_11` — stages 8+9): /guides/funding-ways/ (the eight
mechanisms and what each really costs; "get evidence before you get money" as the spine) and
/guides/funding-sources/ (the durable UK institution map: Innovate UK/UKRI/KTP/Catapults,
British Business Bank/Start Up Loans, devolved agencies + Growth Hubs, banks, angels/UKBAA, VC,
crowdfunding classes, King's Trust + social investment, universities). **Figure policy stricter
than the patents guide**: NO amounts, rates, caps or deadlines anywhere — they go stale within a
fiscal year and a stale figure is indistinguishable from a fabricated one; every section points
at the institution's own site. **Institution policy**: durable major bodies only, no individual
funds/platforms named where a class will do (naming one is an endorsement + a staleness bomb).
Both guides funnel to /report.html; hub lists them automatically (nav_order 30/40).

**Docs written to this point**: SUMMARY_2026-07-25 (NEW file per the series rule — the 07-18
summary stands untouched), README_where_we_are entry, this section. RUNBOOK Phase 5 unchanged
(recipe held for both new guides — slot_name + pages.sections handled in-SQL, no new traps).

**§X.14 addendum — owner ruled: EXTEND THE ENGINE (option B). Built same day, INERT until deploy.**
The paid tool now does what the page says: new STEP 0 assesses the submitted idea itself
(web-verified, the copy's six areas, `is_assessable:false` = an honest rendered "too early to
assess" refusal); sources carried from assess + verify into "Check it yourself:" lists in both
email formats; AI use disclosed in the report intro, not just the T&Cs. Files:
`idea.uk/golang_files/engine.go` (+235/-44 across the three), `prompts.go`, `service_test.go`.
`go build` + `go vet` clean; **full test suite green** — including a pre-existing failure fixed in
passing (`TestReviewBeforePayFlow` asserted "pay here"; the email says "pay £29 here" — stale
wording assertion, failing on copy not behaviour). Run cost roughly doubles (6 calls, 2 long
web-search calls). **Owner deploy checklist** is at the foot of `AUDIT_2026-07-25_paid_tool_vs_copy.md`
(build/scp/restart + the CONTACT_EMAIL grep + one end-to-end report). Until the binary rolls, the
live tool keeps its old behaviour and /report.html keeps overselling — the gap is now closed in
code, not yet in production.

### §X.15 — box triage after the owner's deploy + pipeline stages 1–5 built (2026-07-25 evening)

Owner deployed the extended binary and hit three things; all diagnosed, none was the new code:

1. **The deploy itself was FINE.** The mangled output ("181 loaded units… is-active: command not
   found") was the pasted one-liner splitting at `&&` — read-only SSH confirmed the service
   restarted 15:11:50 and `/proc/<pid>/exe → /opt/idea/idea` built 15:11. The restart had already
   succeeded before the error text appeared.
2. **`at_capacity` on /confirm = five of the owner's own stale test orders** (June 10 – July 17,
   all aaa@designconsultancy.co.uk: blobber4/bubblefarm/me/2×Antony) holding every slot —
   `ActiveCount` counts `awaiting_review+awaiting_payment+paid+running` against
   `MAX_ACTIVE_ORDERS=5` (store.go:109, service.go:411). Fix = hand-edit orders.json to
   `declined` (RUNBOOK Phase 4c pattern; hand-edit sends NO emails, unlike /decline).
3. **The CONTACT_EMAIL grep found a real live bug**: TWO assignments in /etc/idea/idea.env
   (line 29 `idea@contactforsales.com`, line 75 `idea-uk@leopardess.uk`) — and in a systemd
   EnvironmentFile the LAST wins, so the STALE dead address was the effective one in every
   report email. Fix = delete line 75. Open owner question: line 29 is `idea@…` but the site DB
   uses `idea.uk@contactforsales.com` — confirm which mailbox is real.

   **Mutating the box was BLOCKED by the permission classifier** (read-only SSH allowed). Right
   call — it is the live order store. Exact fix commands (backup → python status-flip of the five
   ids → sed the env line → start + verify incl. `grep -ac "YOUR IDEA, ASSESSED" /opt/idea/idea`,
   since the box has no `strings`) handed to the owner in-chat. UNVERIFIED until the owner runs
   them: capacity should read `{"open":true,"active":0}`.

**STAGES 1–5 BUILT & VERIFIED LIVE** (`sql/p4_13`, locks `p4_14`): creating-ideas / building-it /
testing-it / user-acceptance / feedback-loops, nav_order 2–6 so the hub reads as the JOURNEY and
feedback-loops chains into patents, joining the two halves. All five: authored (method not data —
nothing checkable to fabricate), 3 sections each, rr=3/carried=0, curl-verified (~31KB each, 7
h3s), locked after the p4_08 derive-guard passed. **All 9 guides now live: 27/27 sections locked,
both hub listings deriving free.** One render batch quirk: 3 of 5 orchestrations COMPLETED fast,
2 sat in deploy_page AWAITING_RESPONSES ~10 min under git-adapter backlog — the pages were already
deployed+live while their orchestrations still showed in-flight; page state, not job state, was
the truth again.

**§X.15 addendum — one more silent cap found and fixed (`sql/p4_15`).** With nine guides live, the
hub listed SIX: `guide-list`'s schema declares `items.limit: 6`, so the first six by nav_order got
in and copyright + both funding guides silently fell off the hub that exists to list them — green
render, healthy-looking page, absent content (the "no silent caps" failure shape, as data). Raised
to the resolver's hard cap (24); guard enforces the no-op for the other guide-list sites
(gamesdesign 5 guides, relojistas 4 — both under the old cap; gamesdesign live-verified unchanged
after). VERIFIED LIVE 16:05: all 9 cards, journey order — creating-ideas → building-it →
testing-it → user-acceptance → feedback-loops → patents → copyright → funding-ways →
funding-sources.

### §X.16 — funding-fit finder (second free tool) + /report.html copy pass (2026-07-25 late)

Owner: *"please go ahead with both."* Both live.

**COPY PASS (`sql/p4_16`)** — light by design, now the engine matches the promise (extended binary
deployed & verified running). Two edits only: one paragraph appended to the body ("The report does
not stop at the idea you sent…" — the further-ideas half: cross-vendor critique, live-web checks,
ranked survivors each with its own cheap test and sources, plus what didn't make the cut); one
sentence extended in the closing CTA ("…with sources you can check — plus further ideas for your
business, tested the same way"). Sections verified UNLOCKED before editing (a locked section's
rerender is discarded by save_page_sections — the p4_08 lesson applied in the other direction).
Pre-checks that turned out fine: the info-cards' `/report` links serve 200 (a tool route, not a
404), and the stale leopardess email in the form section's ambient blob does not render.
VERIFIED LIVE: both new texts on the page.

**FUNDING-FIT FINDER (`sql/p4_17`, locks `p4_18`)** — /tools/funding-fit/index.html + new
`funding-fit` component (56548044). Second application of the gate-before-score rule, exactly as
014 predicted it would be needed: **two dispositive answers checked before anything composes** —
GATE 1 "money is for living costs" → the honest almost-nothing-funds-runway verdict (a sum-score
would route a runway-seeker to Innovate UK); GATE 2 "no evidence yet" → buy evidence cheaply,
funnelled at the testing/building guides + the report. **Past the gates the answers COMPOSE a
route map** (grants / equity / debt / customer money) rather than picking one winner — funding
routes are not mutually exclusive, which is the second reason a scored quiz is the wrong machine.
Q7 is CHECKBOXES (18–30 / social purpose / devolved nation — not mutually exclusive; validation
requires only the six radios). Tension detection included (never-equity + must-be-huge → "one of
the two normally has to give"). Same content policy as the funding guides (durable institutions,
zero amounts, free-front-doors, paid-intermediary warning) and same delivery posture as
patent-check (inline JS, static-fallback-free schema, template parse-tested before insert, nginx
already probed). VERIFIED LIVE: 7 questions, 20 radios + 3 checkboxes, gating JS present
unescaped, 0 unreplaced actions. page_type='tool' → /tools.html lists it at nav_order 20
(report → patent-check → funding-fit → audience-check).

idea.uk now: 9 guides + 3 free-ish tools (2 client-side finders + the audience-check taster) +
the paid report, everything cross-funnelled, all authored surfaces locked, all listings derived.

### §X.17 — box state re-checked: the three fixes have NOT been run yet (2026-07-26)

Owner re-ran the CONTACT_EMAIL grep and asked for a walkthrough. Verified read-only before
writing it: `curl https://idea.uk/capacity` → `{"active":5,"max":5,"open":false}` (the five
stale test orders still hold every slot); orders.json status counts unchanged (60 requested /
4 awaiting_review / 1 awaiting_payment); `grep -c CONTACT_EMAIL /etc/idea/idea.env` → 2 (both
lines still present, stale one still winning); no `.bak.*` files exist. So NONE of yesterday's
three fixes has run — the walkthrough starts from the backup step. Full owner-facing walkthrough
issued in-chat (backup → release the 5 slots → email fix with the idea@-vs-idea.uk@ decision →
restart + verify → one end-to-end report). Noted in passing: yesterday's binary deploy also took
the /request spam hardening live (honeypot + timing + limiter were in the source tree since
07-17 but never deployed until now), so the two optional items left on the box are the old spam
rows in orders.json (~50 garbage 'requested' entries, cosmetic) and the two SES custom-MAIL-FROM
DNS records (outstanding since 07-18).

Docs to this point: SUMMARY_2026-07-26 written (new file — the pipeline-complete milestone;
07-25's summary predates stages 1–5, both tools and the engine extension, so "where we are now"
genuinely differs).

> **CORRECTED 2026-07-26 (owner): `idea-uk@leopardess.uk` is the CORRECT address for the tool —
> not stale.** §X.14/X.15/X.17 and the AUDIT doc called it "the stale/dead address"; that was
> carried over from the SITE contact-email work (where the owner chose a contactforsales
> address) and stated as fact for the TOOL without evidence the mailbox was dead. The two are
> deliberately different addresses. Consequences: (1) since the LAST EnvironmentFile line wins,
> the tool's effective CONTACT_EMAIL was leopardess — i.e. **already correct all along**; the
> only real defect is a confusing inert duplicate (line 29, contactforsales), and the fix is to
> delete THAT line, the opposite of what the 07-25 walkthrough said — both of its step-3 options
> would have set the wrong address, caught by the owner before running. (2) engine.go's
> reportContact() fallback (idea-uk@leopardess.uk) is CORRECT, not a bug — the AUDIT side-find
> reverses. (3) The "stale idea-uk@leopardess.uk in sites.content_data" side-finding stands for
> the SITE only. Logged in WRONG_CALLS.

**§X.17 addendum — the slot-clear "didn't work": edit-under-a-live-writer (2026-07-26).** Owner
ran the clear (cleared: 5) but capacity stayed 5/5. Cause, confirmed read-only: the tool loads
orders.json ONCE at startup (store.go:44 NewStore→ReadFile) and rewrites the WHOLE file from
memory on every change (store.go:64 persist→WriteFile) — the service had been up since 07-25
15:11, so the hand-edit never entered memory, and the owner's next test request (requested 60→61)
persisted memory back over the edit (file mtime 12:40 = the service's own write, the five
restored). **RULE for the RUNBOOK: never edit /var/lib/idea/orders.json under a running service —
stop, edit, start.** The original walkthrough's step 1 had the stop; the owner's wrapper script
skipped it. Corrected single-command sequence (stop→clear→start→verify) issued in-chat. The env
dedupe (now 1 line, the correct leopardess address) rides the same restart.

### §X.18 — capacity: the root cause, the durable fix, and the site notice (2026-07-26)

Owner: *"think hard about how to clear these declined orders… we want to keep them on record, and
if the queue of real orders is full then we want to state that on the site."* Plus a correction on
the env var.

**WHY THE OWNER'S `systemctl start` DID NOTHING.** `start` on an already-active unit is a no-op —
the service has been up since 07-25 15:11:50 throughout. So the clear never entered memory and the
next request persisted memory back over the file (the §X.17 addendum trap, second occurrence in
one day). It needs `restart` (or stop→edit→start). Verified: file shows the five slot-holders
restored, mtime 12:40 = the service's own write.

**ROOT CAUSE, stated properly.** `ActiveCount` counts awaiting_review/awaiting_payment/paid/running
against `MaxActive` and **nothing ever aged anything out**. With MaxActive=5, five orders nobody
progressed closed the service to new work permanently, and the only remedy was hand-editing a file
the running process owns. That is a design gap, not an operational slip — hence a code fix, not a
better runbook entry.

**THE FIX (code, tests green, inert until the owner's next deploy):**
- `Store.ExpireStale(reviewAge, paymentAge, now)` — releases cold slots, returns what it released.
- New terminal status **`expired`, deliberately distinct from `declined`**: declined = the operator
  looked and said no; expired = it went cold and was released. Filing an abandoned order as a
  judgement we made would corrupt the owner's own record, which is exactly what "keep them on
  record" is asking us to protect. Both keep the row; neither deletes anything.
- Two separate ages because the waits differ in kind: `STALE_REVIEW_DAYS` (default 14 — waiting on
  US) and `STALE_PAYMENT_DAYS` (default 7 — waiting on the customer). `0` disables either.
- `App.sweepStale()` runs at startup, hourly, **and before /capacity answers** — so the number the
  site publishes can never be inflated by a six-week-old order.
- Never touches `running`/`paid`/terminal rows; falls back to CreatedAt for rows predating
  UpdatedAt; idempotent. 3 tests (`expire_stale_test.go`) lock all of it.

**THE SITE NOTICE (`sql/p4_19`) — LIVE TODAY, no deploy.** Banner on the report form, fed by the
existing public `/capacity`. **Copy says "there will be a wait", never "you cannot order"** — and
that is correctness, not tone: the capacity gate is on the operator's `/confirm`, not on
`/request`, so visitors can always submit and being full only blocks *starting*. Disabling the
form would misdescribe the system and bin real demand. Fail-open three ways (fetch error,
malformed JSON, `open:true` → render nothing); JS-disabled visitors see exactly today's page; we
deliberately never advertise "slots available".

**ENV CORRECTION (owner).** `OPERATOR_EMAIL` and `CONTACT_EMAIL` are different settings, both real:
OPERATOR_EMAIL (main.go:30) = where operator mail goes; CONTACT_EMAIL (main.go:31) = the public
support address, and `contactEmail()` (service.go:719-725) already falls back to OperatorEmail when
it is unset. So the owner is right that OPERATOR_EMAIL is the one that must be correct — and it is.
The one wart is `engine.go reportContact()` reading `CONTACT_EMAIL` **directly from os.Getenv** with
a HARDCODED `idea-uk@leopardess.uk` fallback, bypassing the config and the OperatorEmail fallback
the rest of the service uses. Folded into the same deploy: reportContact now takes the address from
config (ContactEmail → OperatorEmail), so **OPERATOR_EMAIL is the single source of truth** and the
CONTACT_EMAIL line can simply be deleted.

**§X.18 addendum — owner ran it; and the lock gate is now LIVE (2026-07-26 13:19).** Verified
read-only: orders now `{requested:60, expired:5, declined:4, delivered:3}` — the five cold slots
released using the `expired` status (not `declined`, as intended), service restarted 13:19:38,
`CONTACT_EMAIL` lines now **0**, `/capacity` → `{"active":0,"max":5,"open":true}`. The site banner
correspondingly renders nothing (open:true → silent by design). Note the removal of CONTACT_EMAIL
is correct against BOTH binaries: today's deployed one falls back to the hardcoded
`idea-uk@leopardess.uk`, and the queued one resolves it from `OPERATOR_EMAIL`.

> **CORRECTED 2026-07-26:** several entries in this file and in `p4_03`'s header say
> `bugs_open/058` (the component lock gate) is "committed but NOT YET LIVE, so lock enforcement is
> incomplete". **It is now CLOSED & LIVE on v1.0.1165** (closed by another session today,
> induced-fault proven). Consequences for this workstream: (a) the 27 locked authored sections are
> genuinely protected now, not just annotated; (b) the p4_08 rule — never lock a section with a
> `query.*` source — stops being theoretical and becomes load-bearing, since a locked derived
> listing would now really freeze; (c) editing any locked page from here needs an unlock first.

**Docs written for the handoff (2026-07-26):** `SUMMARY_2026-07-26b` (second today — justified:
the morning's summary said the remainder was "four pasted commands", and the capacity work turned
out to be a design gap plus a second deploy, so "where we are now" genuinely differs), and
`HANDOFF_RESUME_idea_uk_vm_site.md` rewritten with a current ▶ START HERE block (the 07-22 block
demoted to "PREVIOUS STATE", kept for history — the same supersede pattern the RUNBOOK uses). The
handoff leads with the one outstanding proof: **no report has yet been received in the new
format.**

### §X.19 — the extended report is PROVEN IN PRODUCTION (2026-07-26)

The week's last open item, closed. Owner submitted a real idea, confirmed it, received the draft,
judged it good, and declined it — the correct way to close a test without self-charging (declined
is terminal, releases the slot, and emails the requester politely at no cost).

**Verified against the STORED REPORT, not the owner's impression** — the discipline this workstream
has needed all week. `ord_1785069609860726188`, created 12:40, decided 13:39; report 13,227 chars
text / 20,207 chars HTML:

| marker | |
|---|---|
| `YOUR IDEA, ASSESSED` (text) + `Your idea, assessed` (HTML) | ✅ |
| `Check it yourself` source lists | ✅ — **16 http links in the HTML** |
| `A considered next step` | ✅ |
| `We use AI to research…` — disclosure in the REPORT, not only the T&Cs | ✅ |
| `FURTHER IDEAS WORTH PURSUING` — the ideation half, retitled | ✅ |
| old intro `"You asked us to find AI product ideas for…"` | ✅ **ABSENT** |

`too early to assess` absent — CORRECT, the submission was assessable. That branch stays
test-covered only (`service_test.go`); a deliberately vague submission would give the live proof
if ever wanted.

**So the whole chain is proven end to end**: request → operator confirm → step-0 web-verified
assessment → cross-vendor cut → verify with sources → score → draft → operator decision. The
07-25 engine work (assessment, sources, AI disclosure, honest-refusal branch) is delivering to
real customers, and `AUDIT_2026-07-25_paid_tool_vs_copy.md`'s two NOT-DELIVERED rows are now
DELIVERED.

Two things to watch on the next few real runs, both flagged to the owner: wall-clock (two long
web-search passes now) and spend (~2× per report at the same £29).

Docs updated: HANDOFF "DO THIS FIRST" replaced with the evidence and the deploy promoted to top
job; SUMMARY_2026-07-26b carries a visible dated UPDATE on its "unproven" paragraph (corrected in
place per the working-docs rule, not overwritten, and no 26c written — the milestone is the one
26b already anticipated).

### §X.20 — two live security defects in the paid tool, the second deploy, and a stale-handoff misstep (2026-07-26 evening, session "idea.uk vm site 6")

**Read §X.19 first.** This session started from the **15:34** version of the handoff, which still
carried "DO THIS FIRST — prove the new report format end to end". §X.19's session rewrote that
file at **15:57**. I did not re-read it, so the first half of this entry is work done against a
state that had already moved. The misstep is written up at the end and in `WRONG_CALLS.md`; the
findings below stand on their own evidence and are unaffected by it.

#### 1. `bugs_open/089` — the £29 report could be taken without paying

Found while reading `service.go` to work out which legs of the funnel I could drive myself.

`orderSuccess` honoured a local-test shortcut on the strength of a **query parameter alone**:

```go
if r.URL.Query().Get("fake") != "" { // FakeProvider local-test shortcut
```

It moves an order `awaiting_payment → paid` and delivers. The shortcut exists for `FakeProvider`
(`billing.go:134` builds exactly that URL so a local run needs no Stripe keys) and the type is
commented *"local/testing only — NEVER in production"* — but the handler never asked which
provider was configured. The box runs Stripe:

```
$ curl -s http://127.0.0.1:8080/health
{"auto_deliver":false,"ok":true,"price_gbp":29,"provider":"*main.StripeProvider"}
```

**Reachable by an ordinary buyer, which is what makes it real rather than theoretical.**
`CreateCheckout` hands Stripe both URLs with the order id embedded (`billing.go:53-54`), so
cancelling a real checkout redirects the buyer to `/order/cancel?o=<their id>` — the id is
disclosed to the one person with a motive. Then `/order/success?o=<id>&fake=1` delivers.

Fixed by gating on a type assertion the caller cannot influence, plus a refusal log line.
**Induced against pre-fix source** (scratch copy of the module + `git show HEAD:…/service.go`):

```
payment_bypass_test.go:75: PAYMENT BYPASS: status moved to "delivered" under StripeProvider;
  want it held at awaiting_payment
```

`delivered`, not merely `paid` — under `ReviewBeforePay` the stored report is emailed in the same
request. Nothing was lost to it: of 72 orders, 3 are `delivered` and none by this path.

#### 2. `bugs_open/090` — a visitor could choose the IP it was rate-limited as

Found while proving 089's fix on the live box, one endpoint later.

`clientIP` took the **first** `X-Forwarded-For` entry, commented *"first entry is the original
client"*. True of the internet; false of this deployment. nginx forwards with
`$proxy_add_x_forwarded_for` (`snippets/proxy_tool.conf:13`), which **appends** the real peer — so
the first entry is precisely the part a caller writes for itself. Proven against production:

```
$ curl -s -H 'X-Forwarded-For: 203.0.113.77' 'https://idea.uk/order/success?o=ord_xff_probe&fake=1'
Jul 26 18:32:22 idea1 idea[106548]: orderSuccess: refused fake=1 payment shortcut under
  *main.StripeProvider (order "ord_xff_probe", ip 203.0.113.77)
```

`203.0.113.77` is TEST-NET-3 — it can never be a real client. That key drives the taster limiter
(3/hour — **the only bound on LLM spend at a free, unauthenticated endpoint costing ~£0.02 a
call**), the intake limiter, and the IP stored on every order to seed a future block list.

Bounded by nginx's own `limit_req_zone $binary_remote_addr … rate=10r/s`, which keys on the real
peer and is **not** spoofable — so the box cannot be flooded off the network. But 10r/s is a flood
limit, not a spend limit: the control protecting *money* is the one that failed.

Fixed: believe forwarding headers only from a loopback/private peer (the tool port is firewalled —
`ufw` allows only OpenSSH/80/443, and `curl http://116.203.204.115:8080/health` from outside times
out), then take `X-Real-IP` (set with `proxy_set_header`, so replaced not merged) else the
**rightmost** XFF entry. `net.SplitHostPort` replaces `LastIndexByte(addr, ':')`, which returned
IPv6 peers wrapped in brackets — the live shape, since this session's own order came from IPv6.

**The part worth keeping:** `request_hardening_test.go:51` asserted the defect verbatim —
`want IP 203.0.113.7 (first XFF entry)` — and had passed every run since July. The spoofable
behaviour was not missed by the tests, it was **pinned in place by one**. Corrected with a dated
note, not deleted. A green suite says behaviour is *intended*; it never says it is *safe*.

#### 3. A standing open item, refuted rather than done

The handoff carried *"Real-client-IP in nginx — idea.uk is behind Cloudflare, so nginx would see
Cloudflare's IP"*. **It is not behind Cloudflare.** `dig NS idea.uk` → Hetzner nameservers;
`dig +short idea.uk` → `116.203.204.115` (the box itself); `curl -sI https://idea.uk/` → no
`cf-ray`, `server: nginx/1.28.3 (Ubuntu)`. No `set_real_ip_from` is needed and none should be
added. The premise was false and the real defect was in the Go, three files away.

#### 4. The second deploy — done, with the two fixes riding along

18:29 UTC. `systemctl stop` → old binary kept as `/opt/idea/idea.prev-2026-07-25` → new binary →
`start`. Orders backed up first (`orders.json.bak-2026-07-26-predeploy`). After restart: unit
active; `{"active":0,"max":5,"open":true}`; **all 72 orders intact** (60 requested / 5 expired /
4 declined / 3 delivered); discriminating pod-grep — `grep -ac "refused fake=1"` = **1** in the new
binary, **0** in `idea.prev-2026-07-25`, and `"YOUR IDEA, ASSESSED"` = 1 in both (positive control
for the 07-25 work). Refusal proven live on the deployed binary, not just in tests.

First boot swept nothing, as predicted: `ExpireStale` only touches `awaiting_review` and
`awaiting_payment`, and there were none.

#### 5. Residual found in the expiry design, NOT built

`ExpireStale` skips `running` — right while a run is genuinely in flight, wrong after a restart,
because fulfilment is an in-memory goroutine and none survives one. An order left `running` by a
restart therefore holds a slot **for ever**, which is the exact failure the expiry work was
written to end. Cheap fix: at startup, before the first sweep, any order still `running` cannot
be — reset it to `requested` or `expired`. Deliberately not built: a run was in flight and this
change is only correct if it never touches a live one.

> **CORRECTED 2026-07-27 (§X.24) — the diagnosis held, the proposed fix had a money bug.**
> "Reset it to `requested` or `expired`" is wrong for a buyer who has **paid**: under charge-first
> the payment lands before the engine runs, so `requested` lets `/confirm` issue a second pay link
> and charge them twice, and `expired` bins a paid-for report. It also would not have worked —
> `ExpireStale` skips `requested` and `paid` too, so an undiscriminating reset just renames the
> leak. Built instead with `ProviderSessionID` (written only by `sendPayLink`) as the
> paid/unbilled discriminator. Fixed, deployed and induced live — see §X.24.

#### 6. Observation for the margin question

The engine runs `GEN_MODEL=claude-opus-4-8`, `CRITIQUE_MODEL=claude-sonnet-4-6`,
`VERIFY_MODEL=claude-opus-4-8`, `SCORE_MODEL=claude-sonnet-4-6` (`engine.go:26-29`) — a full model
generation behind, and all four are plain env vars, so changing them needs no rebuild. That is a
live lever on both halves of the open margin question. **Not changed** — model choice is the
owner's call and trades cost against quality on the product's core output. `[cut]` in the logs is
a critique-step label, not a truncation marker; no call came back at its `max_tokens`.

#### 7. MISSTEP — I acted for two hours on a handoff that had been rewritten under me

I read `HANDOFF_RESUME` at 15:34. §X.19's session replaced its "DO THIS FIRST" block at 15:57 with
proof that the format run had already happened at 12:40. I never re-read it, so I: put a question
to the owner premised on a state that no longer held, and fired a second production engine run at
~2× the old per-report spend.

What caught it was opening this file to append and finding §X.19 already there. Nothing in my own
work would ever have shown it — my order ran, the engine logged normally, every check passed.

The check that would have caught it costs two seconds: `ls -la` the workstream directory, or
`git log --oneline -5`, **immediately before an expensive action** rather than once at session
start. The mtimes read 15:57/15:58 against the 15:33–15:35 I had listed an hour earlier.

CLAUDE.md already says a session-start snapshot goes stale within minutes. I apply that to `git
status` before committing and had never applied it to **the handoff itself**, which is shared
mutable state in exactly the same way. A doc reads like a fact about the world; it is a message
from a session that may still be typing. And the more valuable a "next action" looks in a handoff,
the likelier another session is already doing it.

Salvage, which is luck and not design: the 12:40 order was **declined**, so
`approve → pay link → payment → delivery` had never run in production. This run is being taken
through exactly that leg. Written up in `WRONG_CALLS.md`; a coordination block is at the top of
the handoff so §X.19's session does not redeploy into the in-flight run.

### §X.21 — third deploy: the copy fixes are live; all four markers verified (2026-07-26 21:10 UTC)

Built from the committed tree, deployed 21:10, verified 21:12. Discriminating marker for this one
is `(each out of 5)` — **2** in the new binary (text renderer + HTML renderer), **0** in the
running one before it.

All four markers present together on the deployed binary, which is the check that matters — a new
build is not evidence that it kept the previous build's fixes:

```
089 refused fake=1    : 1
090 X-Real-IP         : 1
copy (each out of 5)  : 2
07-25 YOUR IDEA, ASSESSED : 1
```

Both attacks re-run against this binary and refused: the forged `X-Forwarded-For: 203.0.113.77`
logged as my real IPv6 peer, and the bypass against the real `awaiting_payment` order held it at
`awaiting_payment`. Order intact across the restart — status, 10,109-char report, `cs_live_`
session, 73 orders.

Rollback chain now three deep: `idea.prev-2026-07-25` → `idea.prev-2026-07-26-089only` →
`idea.prev-2026-07-26-089-090`. Orders backed up before each (`…predeploy`, `…2`, `…3`).

**Note on the fresh chassis build (v1.0.1171, deployed this evening): it is irrelevant to this
tool.** The £29 tool is a standalone stdlib-only Go module on the Hetzner box with its own build
and deploy path (`go build` → `scp` → `systemctl`), no CI and no chassis coupling. Nothing this
workstream shipped tonight was chassis code, so nothing was inert awaiting that roll and nothing
is owed against it. The chassis version matters to idea.uk only for *page* builds and renders.

**Still open, unchanged from §X.20:** the `running`-order slot leak across a restart (fix named,
not built); the model generation as a margin lever (owner's call); and the payment leg, which is
now the single unproven step in the product and is waiting on a human.

### §X.22 — the engine moved to the Claude 5 family, and the model swap alone would have broken it (2026-07-26 late)

Owner: *"update all the outdated models to the most recent."* The engine ran
`GEN/VERIFY=claude-opus-4-8` and `CRITIQUE/SCORE=claude-sonnet-4-6`
(`engine.go:26-29`) — one full generation behind. Now `claude-opus-5` /
`claude-sonnet-5`. **Deployed 21:5x, verified against the live API and the live site.**

Model ids taken from the `claude-api` skill's catalogue, not from memory. That
mattered: the ids are the *visible* half of the change and the cheap half.

#### The load-bearing half: a swap alone would have 400'd every call

`usesAdaptiveThinking()` was an **allow-list** of models that take adaptive
thinking (`opus-4-7` / `opus-4-8` / `mythos`). Any model it had never heard of
fell through to the legacy `thinking:{type:"enabled",budget_tokens:N}` branch —
and the 5 family rejects that outright. **Induced against the real API before
changing anything**, using the exact bodies the Go builds:

```
=== claude-opus-5
  NEW format (adaptive + effort): 200  stop=end_turn out=4 text='OK'
  OLD format (budget_tokens)   : 400  "thinking.type.enabled" is not supported for
                                      this model. Use "thinking.type.adaptive" and
                                      "output_config.effort"
=== claude-sonnet-5   — identical on both counts
```

So the env vars I called "a live lever needing no rebuild" in §X.20 were **not**
that: changing `GEN_MODEL` on the box alone would have taken the paid product
down with a 400 on every order. Correcting that here because I wrote it in the
handoff, NOTES and memory, and it was wrong in the direction that costs money.

**The fix is an inversion, not a patch.** `usesManualThinkingBudget()` is now a
**deny-list**: only the known-old families get the legacy format, and an
unrecognised model gets the modern one — because an unknown model is far likelier
to be newer than older. The old shape turned *every future upgrade* into a
runtime 400; that class is now closed.

#### The second trap: an omitted field whose meaning changed under us

Two call sites (the free taster `runAudience`, and generate step 2) set no
`Effort`, so `callClaudeOpts` sent **no `thinking` field at all**. On Opus 4.8
that meant no thinking. On the 5 family the same omission means thinking runs
**adaptively by default**, and thinking shares the `max_tokens` cap with the
answer. Both steps return JSON, so a truncated answer is a parse failure, not a
short report — the taster (4096 cap) was the exposed one. Both now set `Effort:
"low"` explicitly with raised caps.

Chose low-effort adaptive over `thinking:{type:"disabled"}` deliberately: the
migration guidance flags that disabling thinking on Opus 5 can leak `<thinking>`
tags into the visible response, and these responses are parsed as JSON. Verified
on the deployed binary — the live taster fragment contains **0** occurrences of
`thinking`.

Sonnet steps got headroom too: Sonnet 5's tokenizer produces **~30% more tokens
for the same text** than 4.6, so caps sized against 4.6 can cut equivalent
output. `max_tokens` is a ceiling, not a reservation — headroom costs nothing
unless used.

#### Found while there: a CUT completion was being served as a finished one

The response parser never read `stop_reason`. A completion cut at `max_tokens`
came back HTTP 200, its text parsed, and nothing downstream could tell — the
exact trap CLAUDE.md names ("`output_tokens == max_tokens` means the completion
was CUT"). Harmless-ish while nothing thought; materially riskier now that
thinking competes for the same budget. `max_tokens` and `refusal` now return an
error instead of persisting a fragment into a customer's report.

Also: the pre-commit twin check flagged `callClaude` as an untouched sibling of
`callClaudeOpts`. It turned out to have **no callers at all** — its own comment
claimed it "keeps every existing call site unchanged" and there are none. It
passes no `Effort`, so it would hit the same trap if revived; documented as
correct-if-revived rather than deleted.

#### Verification, in order

1. Live API probe of both ids, new format **and** old, before editing (above).
2. `go vet` clean, full suite green, new `model_wire_format_test.go` — 5 cases
   including "an unrecognised model must default to adaptive", the regression
   that motivated the inversion.
3. Discriminating pod-grep on the deployed binary: `claude-opus-5`=1,
   `claude-sonnet-5`=1, `claude-opus-4-8`=**0**, `claude-sonnet-4-6`=**0**; and
   the three earlier fixes (089/090/copy) all still present in the same binary.
4. **The deployed binary actually calling the new models**: live taster →
   HTTP 200, 2,587 bytes, 9.1s, coherent output, 0 leaked thinking tags.

Rollback: `/opt/idea/idea.prev-2026-07-26-opus48`. Orders backed up to
`orders.json.bak-2026-07-26-predeploy4`; 73 orders and the pending unpaid order
intact across the restart.

#### Cost — [UNMEASURED], and the owner asked about margin

Not claiming a per-report figure: nobody has run a full report on the new models
yet. What is known rather than guessed: Opus 5 is **$5/$25 per MTok, the same as
Opus 4.8**; Sonnet 5 is **$3/$15 with an introductory $2/$10 through
2026-08-31**, so at or below Sonnet 4.6. Pushing the other way: Sonnet 5's
tokenizer counts ~30% more tokens for the same text, and two steps that
previously did no thinking now do a little. Net direction is genuinely unknown
until a real report runs — the `[cache]` log lines carry per-call token counts,
so the next order measures it for free.

### §X.23 — FIRST SALE: paid, delivered, slot released (2026-07-27 11:13:13 UTC)

`ord_1785090638951163875` → `delivered`. The owner paid the £29 on the live Stripe session issued
at 18:43 the previous day. **This is the first time idea.uk has taken money and delivered a report**,
and it closes the last unexecuted leg of the chain.

Verified from the box and from nginx, not from being told it worked:

```
order   : status=delivered  updated=2026-07-27T11:13:13.463Z  session=cs_live_a1j7o8uz…
          report 10,109 chars / html 14,551 — unchanged since generation
statuses: {requested 60, expired 5, delivered 4, declined 4}   ← delivered 3 → 4
capacity: {"active":0,"max":5,"open":true}                      ← slot released automatically
journal : 11:13:13 email to aaa@designconsultancy.co.uk sent: "Your idea.uk report"
nginx   : 11:13:13 POST /stripe/webhook 200  3.130.192.231  "Stripe/1.0 (+…/docs/webhooks)"
          11:13:15 GET /order/success?o=ord_1785090638951163875  ref=checkout.stripe.com
```

The payment arrived through the **signed webhook** — the source of truth — not through any client
redirect. `deliverReport` then sent the stored report without re-running the engine, exactly as the
review-before-pay design intends, and `ActiveCount` dropped the order out of the active set on its
own, freeing the slot with no operator action.

#### Two things the same access log proves for free

1. **Removing the `fake=1` bypass did not break the real payment path.** The genuine Stripe success
   redirect is `GET /order/success?o=…` with **no** `fake=1` parameter — the shortcut `bugs_closed/089`
   deleted is not part of the real flow at all. That was the reasonable fear when removing something
   the checkout appears to touch, and it is now answered with evidence rather than argument.
2. **The attack and the real payment are visible side by side.** The three `&fake=1` probes from
   26 July 18:45 / 21:12 all returned 200 and progressed nothing; the Stripe webhook on 27 July
   progressed everything. Refusal and acceptance in one file.

#### What the customer actually received — stated because it is not the newest build

The report was generated at **18:40 on 26 July**, which is before the copy fixes (deployed 21:10)
and before the Claude 5 migration (~21:55). Checked against the stored text rather than assumed:

```
"out of 5 — hard to copy" present : True     ← the malformed score line
".." (doubled full stop)  present : True
"(each out of 5)"         present : False    ← the fix, absent as expected
```

So the first paying customer received a report carrying two of the three copy defects. Nothing to
be done about that copy — it is sent — but it means **the next order is the first report on the new
models AND the first with the copy fixes**, and therefore also the first that measures per-report
cost on the 5 family. The `[cache]` log lines will carry the token counts.

#### What this retires, and the transferable bit

The standing line "nobody has ever paid for a report and received one" is gone. Worth keeping is
*why it survived so long unnoticed*: the 12:40 run on 26 July proved the report format and was then
**declined**, which is the correct way to close a test without self-charging — and it left
`approve → pay → webhook → delivery` unexecuted while every visible signal said the product was
finished. **A product can be complete, verified, and demonstrably working and still never have done
the thing it exists to do.** The question to ask of the next site declared done is not "does it
work" but "has the transaction at the end of it ever completed once".

### §X.24 — the `running` slot leak is fixed, deployed and induced live (2026-07-27, session "idea.uk vm 7")

The last open item from §X.20 §5. Built now because the precondition finally held: the box was
idle (`{"active":0,...}`, `RUNNING NOW: []`), and this change is only correct if it can never touch
a live run.

#### The defect, stated precisely

Fulfilment is an in-memory goroutine (`App.dispatch`), so no run survives a restart — **and a
deploy is a restart** (there were four on 26 July alone). The order is left `running`, which
`ActiveCount` charges against `MaxActive` and which `ExpireStale` deliberately skips, because from
inside the store a dead run and a live one are indistinguishable. The slot was then held **for
ever**, with no path back — the exact failure the expiry work was written to end, reached through a
different door.

That restraint in `ExpireStale` is **correct and must stay**: on an hourly ticker in a live process
it cannot tell a genuine 20-minute run from an abandoned one, and expiring a live run would destroy
a report the customer is owed. So the release has to come from the one place where the question is
decidable — startup, where a process cannot inherit another's goroutines, so every `running` order
is *by definition* abandoned.

#### A correction to the fix §X.20 §5 proposed — it had a money bug

> §X.20 §5 said: *"reset it to `requested` or `expired`."* **Both are wrong for a buyer who has
> paid**, and the note did not distinguish them.

Under charge-first (`REVIEW_BEFORE_PAY=false`) payment lands *before* the engine runs
(webhook → `paid` → `fulfil` → `running`). Sending such an order back to `requested` would let
`/confirm` issue a **second pay link and charge them twice**; `expired` would silently bin a report
someone had paid for. The same shape as `089`/`090`: a plausible one-line fix that quietly costs
money.

Two further things the original note missed:

- **"Any non-running status" does not fix it.** `ExpireStale` also skips `requested` *and* `paid`
  (`default: continue`), so a careless reset just renames the leak. The target has to either free
  the slot or be genuinely re-executed.
- **The discriminator already exists.** `ProviderSessionID` is written in exactly one place —
  `sendPayLink`, `service.go:235` — at the moment a checkout is created. So it answers "has this
  buyer been asked for money" exactly, per order, with no config guessing.

Final design (`Store.RecoverInterrupted`):

| stranded `running` order | goes to | why |
|---|---|---|
| no `ProviderSessionID` — never billed | `requested` | frees the slot (`ActiveCount` ignores `requested`) and is re-startable through the operator's **existing** `/confirm` link, which accepts precisely that status. Nothing durable is lost: a report is only stored on completion |
| has `ProviderSessionID` — paid | `paid`, and re-run | they paid, we owe them the report. Keeps its slot, which is correct |

Wired at the top of `StartSweeper`, **not** in `main.go`: "recover before the first sweep" is an
ordering invariant, and keeping both halves in one function means it cannot be broken by
rearranging the caller.

#### Tests, including a negative control

`recover_interrupted_test.go`, 6 cases. Two worth naming:

- The first pins **the leak itself** — that `ExpireStale` must *not* expire a running order, even
  at 99 days — so the fix cannot later be "simplified" into the sweep, where it would kill live
  runs.
- The wiring test leaves the expiry thresholds **disabled** on purpose: recovery must not inherit
  the sweep's opt-out. An operator who turns expiry off has not asked us to leak slots on restart.

**Checked against a negative control before trusting it.** With `a.recoverInterrupted()` unwired,
`TestStartSweeperRecoversInterruptedRuns` fails with `paid order not re-run on startup: status
"running"`; rewired, green. Restored by reversing the exact edit rather than copying a backup file
over `service.go`, and confirmed byte-identical — a whole-file restore could have clobbered a
concurrent session's edit.

#### Fifth deploy (2026-07-27 12:48 UTC), and the marker discipline

Commit `5c3081e3f`, tree clean at build time so the binary is exactly `HEAD`. Backups first:
`orders.json.bak-2026-07-27-predeploy`, rollback binary `idea.prev-2026-07-27-pre-recover`.

```
                          old binary   new binary
"recovered after a restart"      0            1     ← discriminates
"refused fake=1"    (089)        1            1     ← positive control
"X-Real-IP"         (090)        1            1     ← positive control
"claude-opus-5"                  1            1     ← positive control
```

73 orders intact (60 requested / 5 expired / 4 delivered / 4 declined), unit active,
`review_before_pay=true, auto_deliver=false, price=£29`.

#### The part that actually proves it: induced live

A deploy with nothing stranded exercises **none** of this — the first boot logged no recovery,
correctly, and that is a green happy path proving deployment, not correctness. So the fault was
induced on the live box against a June spam row (`requested`, no session id, nobody cares):
service stopped → status set to `running` → started.

```
recover: order ord_1780668917992469761 was interrupted mid-run and was never billed
         → back to requested, slot released
email to idea-uk@leopardess.uk sent: "[idea.uk] 1 order(s) recovered after a restart"
```

Afterwards: status `requested`, `running: []`, and the distribution **identical to before the
induction** (73; 60/5/4/4) — fully reversible, the spam row is exactly where it started. The
operator address received one real test email; that is the notification working, not a fault.

Edited only with the service **stopped**, per the standing trap: `orders.json` is read once at
startup and rewritten wholesale from memory, so an edit under a running service is both invisible
and doomed.

#### What is NOT proven live, stated because the tests are green either way

The **paid branch** (`running` + session id → `paid` + re-run) is **[UNPROVEN LIVE]** — covered by
unit tests only. Two reasons, and the first is the honest one: under the live config
(`review_before_pay=true`) `running` only ever occurs at `/confirm`, *before* any checkout exists,
so that branch **cannot currently be reached in production at all**. It is defensive correctness
for the documented charge-first config. Proving it live would also mean a real engine run — money
and ~9.5 minutes — to exercise a path that is presently unreachable.

#### Minor

- A `MarkEventSeen` doc comment had been orphaned above `ExpireStale` in `store.go`, so the file
  read as if it documented the wrong function. Moved back; noted in the commit.
- My sandbox clock read ~73 min behind the box on first contact and then self-corrected. The box is
  `Etc/UTC`, NTP-synced, and was right throughout; nothing here depended on it. **Trust the box's
  clock, not the sandbox's**, when comparing timestamps.

### §X.25 — the cost measurement: a partial number, and two checks that only LOOKED like they passed (2026-07-27)

Owner authorised one full engine run to measure per-report cost on the 5 family and confirm the
copy fixes reach a rendered artefact. Run via `idea internal` (the authenticated no-order, no-billing
path) against one of our own domains, launched detached with env sourced from `/etc/idea/idea.env`;
recipe now in the RUNBOOK. **rc=0, 445 s (7 m 25 s), 11,347-char report.**

#### The models are compiled-in, not env — checked before spending

`/etc/idea/idea.env` has **no** `GEN_MODEL`/`CRITIQUE_MODEL`/`VERIFY_MODEL`/`SCORE_MODEL` line, so
the run used `engine.go`'s defaults — `claude-opus-5` (gen/verify/assess), `claude-sonnet-5`
(critique/score). That matches the binary markers, so this measures what a real order costs.

#### The cost number is a FLOOR, not a total — the usage log is conditional

Only **two** `[cache]` lines were emitted, both `claude-opus-5`:

```
[cache] claude-opus-5: created=23905 read=122295 input=1132 output=9012
[cache] claude-opus-5: created=13128 read=48970 input=776  output=3549
```

`engine.go:315` only logs usage when `CacheReadInputTokens > 0 || CacheCreationInputTokens > 0`. All
six call sites set `CacheSystem: true`, but a call whose system prompt falls under the cacheable
minimum (**512 tokens on Opus 5, 1024 on Sonnet 5**) caches nothing and therefore reports nothing.
The two Sonnet steps (critique — its `[cut] same-vendor: Anthropic (claude-sonnet-5)` line did
appear — and score) produced no usage line at all. **So their tokens are invisible, and no amount of
reading this log will recover them.**

Costed at the rates from the `claude-api` skill (Opus 5 $5 in / $25 out; 5-minute-TTL cache write
1.25× input = $6.25, cache read 0.1× = $0.50 — the engine sets plain `ephemeral` with no `ttl`):

| | tokens | $/M | cost |
|---|---:|---:|---:|
| cache write | 37,033 | 6.25 | 0.2315 |
| cache read | 171,265 | 0.50 | 0.0856 |
| input | 1,908 | 5.00 | 0.0095 |
| output | 12,561 | 25.00 | 0.3140 |
| **measured floor** | | | **$0.641** |

**The bound is the useful part, and it survives the gap.** The two measured calls are the expensive
ones (Opus at `xhigh`, 32k `MaxTokens`); the unmeasured ones are Sonnet, on a cheaper model at lower
caps. Even if the missing calls cost as much again, the report stays **under ~$1.30 — under 4% of a
£29 sale.** Margin is not the risk here, and that conclusion does not depend on the missing data.
**[UNMEASURED]** remains the honest label for the exact total.

#### The Sonnet intro rate has an expiry, and it is close

Sonnet 5 is on an introductory $2/$10 per MTok through **2026-08-31**, reverting to $3/$15 — a 50%
rise on the critique+score half of the bill. Immaterial at these volumes, but the margin decision
should not be made on a rate that expires in five weeks.

#### The part that matters more: TWO of the three copy checks were VACUOUS

First read of the artefact looked like a clean pass — all three defects absent, every format marker
present. It is not a pass. Two of the three defects **could not have appeared in this run
regardless of whether the fix works**:

- **Doubled full stop** — the defect fires when the *submitted text already ends in a full stop* and
  the template appends another. My submission ended `…what things should cost` with **no trailing
  stop**, so the doubling had nothing to double. Absence proves nothing.
- **Score line reading "out of 5 —"** — the report hit `NO FURTHER IDEA CLEARED THE BAR`, so **no
  idea was scored and the score block never rendered at all**. The fix marker `(each out of 5)` is
  absent for the same reason. Checking for the defect in a section that does not exist is not a check.

> **So: the three copy fixes are STILL UNPROVEN in a rendered artefact.** The 21:10 deploy on 07-26
> put them in the binary (marker verified), and the unit tests cover them, but no report has yet been
> produced that exercises them. To prove them, submit an idea whose text **ends in a full stop** and
> which is strong enough that at least one further idea clears the bar.

This is [[verify-the-failing-branch]] a third time, and it is worth naming why it keeps recurring:
the failing branch here isn't a code path I control, it's a property of the *input*. A green result
from an input that cannot trigger the fault is indistinguishable from a real pass unless you go
back and ask "could this input have produced the bug?" — which is the question I nearly skipped.

#### What the run DID prove, live

- **The honest-refusal branch of the ideation half runs in production.** `NO FURTHER IDEA CLEARED
  THE BAR` is the engine declining to pad the report with weak ideas — previously test-covered only.
- The assessment half is good and unflattering where it costs us: told the submitter the demand is
  on the *seller* side, that a dozen UK agencies give the same advice away free as lead bait
  ("proof that nobody can charge for it"), and cited UK Business Forums threads and named
  directories with real prices. Sources are checkable.
- **445 s vs the 570 s (9.5 min) of the 07-26 order** — but different submission and a short-circuited
  ideation half, so this is **not** a like-for-like latency comparison and must not be quoted as one.

#### Owed: make the usage log unconditional

One line in `engine.go` — log `Usage` on every call, not only on cache activity — turns cost from
"partially observable when caching happens" into a permanent per-order fact, for real customer
orders as well as test runs. Not built this session: it needs a 6th deploy and the owner has not
asked for one. **Until it ships, every cost figure for this product is a floor.**

> **DONE 2026-07-28 — shipped as the 6th deploy (`fb10b2659`).** Unconditional, `[cache]`→`[usage]`.
> Induced live with one free taster (~1.3¢): `[usage] claude-opus-5: created=0 read=0 input=620
> output=405 stop=end_turn`. **`created=0 read=0` IS the proof** — that call cached nothing, so the
> old binary logged nothing for it. Every future order now measures itself.

### §X.26 — a broken DNS zone, a funnel with no people in it, and a hero CTA pointing away from the sale (2026-07-28)

#### 1. `bounce.leopardess.uk` is BROKEN, not pending — and would never have self-resolved

Owner reported the Clook records were added and SES had gone back to Pending. The zone edit did
land (`SOA serial 2026072801`, bumped that day). But:

| query | result |
|---|---|
| `www` / `_dmarc` / `mail` `.leopardess.uk` | `NOERROR` |
| a nonexistent name | `NXDOMAIN` — the server reports absence correctly |
| **`bounce.leopardess.uk`** | **`SERVFAIL`** on SOA, NS, A, MX, TXT, CNAME, from **both** `dns1`/`dns2.uk-noc.com`, and with `+cd` (so not DNSSEC validation) |

**The NXDOMAIN control is the load-bearing part.** A name that does not exist answers NXDOMAIN; a
name the server thinks exists but cannot serve answers SERVFAIL. So `bounce` is not missing — it is
**present and broken**, the signature of a stale delegation or an orphaned child zone. SES queries,
gets SERVFAIL, and stays Pending indefinitely. **No amount of waiting fixes this**, which is exactly
what "pending" invites you to do. Handed to the owner as a precise report for Clook rather than
"it isn't working".

#### 2. The funnel: there is essentially nobody there

Full access log (05 Jun → 28 Jul), post-cutover window 18–28 Jul, bots filtered:

```
GET /report.html   26 views, 20 unique IPs
POST /request       8,       7 unique IPs
GET /order/success  1
```

**8/26 = 31% is a garbage number and must never be quoted as a conversion rate.** Reverse-DNS on the
viewers: `googleusercontent.com` (3 IPs, 9 views), `tor-exit-8.zbau.f3netze.de`, several Tencent-range
addresses with no PTR — and **4 of the 26 views were the owner**. Of the 8 submissions, **4 were our
own test orders** and two came from Tor exits. **Genuine external prospects in ten days: between
zero and a small handful.** This is the hard evidence for the demand-not-correctness claim in
`SUMMARY_2026-07-27b`, and it is why an A/B test would have been unreadable noise.

#### 3. Report intro restructured — and BOTH renderers needed it (7th deploy, `bec6193cc`)

One block of four long sentences → three scannable paragraphs (what this is, with the reader's own
words in the opener; what is in it; how it was made). **The trap:** `reportIntro` fed a single `<p>`
in the HTML renderer, and **HTML collapses blank lines** — so changing only the string would have
fixed the text email and left the HTML one a wall, with nothing to show it. Both now split on the
blank line via a shared `introParagraphs()`.

**A test caught a real regression.** My first draft prefixed the third paragraph `"How it was made:"`,
which lowercased the W in `"We use AI to research and draft this report"` — the AI disclosure
required by the 07-25 audit and pinned by `TestRenderReadable`. Label dropped, comment added at the
site so it is not re-added. A new test pins ≥3 paragraphs and the submitted idea being in the first.

#### 4. Asked for an early in-page link; found the hero CTA pointing AWAY from the sale

The request was "the form is a long way down, add a link near the top". Measuring first
(`id="request-a-report"` at byte **31,326 of 43,224 — 72% down**) turned up something worse:

```
16191  BEFORE form   <a href="/contact.html">Request a Verified Idea Report</a>
```

**The most prominent control on the page, the one a reader reaches first, took them away from the
purchase.** The section immediately above the form (`call-to-action`, position 4) rendered **no link
at all** — a dead label.

Mechanism, confirmed from the schema rather than assumed: `cta_url` and `primary_cta_url` are
`source: renderer` with **no fallback**; `content_data` carried the CTA *text* but no URL, so the
renderer's hardcoded `/contact.html` default filled the gap. Same shape as the home-page CTA fix
(`p3_05-07`), and because those fields are **not** recomputed on a `section_data_resolved` rerender,
a value set in `content_data` holds.

**So the right fix was better than the one asked for:** point the *existing* hero button at
`#request-a-report` rather than add a second control beside a broken one. `p4_20` sets both CTAs;
`p4_21` dispatches the rerender with `reason=section_data_resolved` (a plain `rerender-pages`
assembles from stored HTML and could never apply a `content_data` change).

Guards written into both scripts, each earned by a past failure on this workstream: both rows
verified **unlocked** first (a locked section's rerender is silently discarded); **no NULL
`content_data`** or the rerender escalates to the LLM writer and rewrites live sales copy; and
`p4_21` **refuses if the CTA values are not present**, so it cannot dispatch a render that changes
nothing and reports success.

**VERIFIED LIVE, with a negative control** (~3 min later — far faster than the 16–36 min the queue
used to take, consistent with `030` being fixed):

```
16216  BEFORE form   #request-a-report   Request a Verified Idea Report   ← hero, 37% down
30449  BEFORE form   #request-a-report   Request a Verified Idea Report   ← was a DEAD LABEL
43901  after form    /contact.html       Contact                          ← footer, correct
'/contact.html' + 'Request a Verified Idea Report'  : GONE ✓
```

Two live paths to the form, the first at **37% down instead of 73%**, and the misdirect is gone.
`secondary_cta_url` left unset on both, so those labels render nothing — correct-or-absent (LNK-005)
rather than inventing a destination.

**The transferable bit:** the brief was a UX tweak; measuring the page before implementing it turned
up a live revenue leak that no scanner had reported, because *every link on the page returned 200*.
`/contact.html` is a real page. **A dead CTA and a wrong CTA look identical to a link checker** —
only reading where the link *goes against what its text promises* finds this class.
