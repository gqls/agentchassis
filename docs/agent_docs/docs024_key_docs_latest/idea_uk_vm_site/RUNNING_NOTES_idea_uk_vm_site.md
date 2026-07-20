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
