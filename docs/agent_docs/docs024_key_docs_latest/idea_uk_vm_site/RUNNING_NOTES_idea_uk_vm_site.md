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
