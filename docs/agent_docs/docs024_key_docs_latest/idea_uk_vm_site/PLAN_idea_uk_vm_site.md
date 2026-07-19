# PLAN — idea.uk: one site, one origin, behind the VM nginx

> **STATUS 2026-07-19 — the plan below is EXECUTED. It is retained as the design record.**
> Phases 0–3 are done: credentials rotated, site completed, per-site deploy target wired + guarded +
> activated, pull-sync live on the box, and the **nginx cutover shipped 2026-07-18**. idea.uk now
> serves the chassis static site with all 16 tool paths proxied on one origin.
>
> **What the plan did not anticipate, and what now matters more than anything left in it:**
> 1. **The cutover orphaned the tool's entry forms** (`/bugs_open/017`). The tool served its own
>    landing page at `/`, and that page carried the audience-check and report-request forms.
>    Giving `/` to the static site removed the funnel. **Fixed** — both forms are now authored as
>    chassis sections and live (`sql/p2_01`, `p2_02`).
> 2. **The static site's chrome is broken on every page** (`/bugs_open/018`): 31 of 33 homepage
>    links are `href=""` — the whole nav, every CTA — plus an empty logo `src`. Almost certainly
>    true since the pages were first built; invisible until the cutover put them in front of the
>    public. **Open, unstarted, and now the top job.**
> 3. Discovery auditors had **never been run** against this site (`/bugs_open/002 F`).
>
> Read `RUNNING_NOTES §T–§W` for the execution record and the collected missteps; `BRIEFING` for the
> plain-English version; `HANDOFF_RESUME` to start a fresh chat.

**Status:** design agreed 2026-07-14; executed 2026-07-14 → 2026-07-19.
**Site:** idea.uk, `site_id 1244516d-014d-421c-88c6-090bb1e9552a`.
**Box:** Hetzner (Nuremberg) `116.203.204.115`. Live, earning. Do not break the money path.

---

## 1. The problem in one paragraph

idea.uk is two half-sites that cannot see each other. The chassis plans and builds a nine-page
static site and deploys it git → GitHub Actions → Backblaze B2 — where **nobody ever sees it**,
because DNS points at the VM, not at B2. Meanwhile the VM runs the £29 report tool (a standalone
stdlib-only Go binary under systemd on `127.0.0.1:8080`) and nginx proxies **everything** to it, so
the tool's single embedded `page.html` is the live homepage. The built site is, in effect, an
elaborate staging copy. The goal is one complete site at idea.uk — the chassis pages *and* the
tool — behind one nginx, on one origin.

---

## 2. The architectural decision (and the two options rejected)

### Decided: the static artefact stays; the VM is a second **sink**, not a second renderer

Three hosting classes, of which we adopt two:

| Class | What it is | Who it's for |
|---|---|---|
| **A — static → B2** *(default)* | Chassis builds HTML, commits to a shared repo, an Action syncs to B2. Serverless. | The thousands. Unchanged. |
| **B — static → VM + backend** *(idea.uk)* | The **same** static artefact lands on a box; nginx serves it and proxies the backend's reserved paths. | The handful of sites that sell something. |
| **C — dynamic rendering on the box** | ❌ **Rejected.** | Nobody. |

**Why C is wrong.** The chassis *is* a static-site generator: rendering happens in Kubernetes, out
of `clients_db`, and its output is HTML. The VM has no route to that database and the tool binary
has no DB driver at all (`go.mod` → `module idea`, zero dependencies). Serving pages dynamically
from the box would mean either giving every box a path into the cluster database — across a fleet,
a coupling and security problem — or re-implementing rendering locally against a copy of the data,
which is the static artefact with extra steps. It would also cost three things we currently get
free:

- **Availability.** Static files survive a tool crash. *Today a tool crash takes the whole site
  down*, because nginx proxies everything to `:8080`. Class B is a strict improvement.
- **Cacheability.** Static is Cloudflare-cacheable; dynamic bypasses.
- **Fleet uniformity.** One artefact, two sinks. Two rendering paths would drift.

**Why one origin beats split-origin** (static on B2 + tool on the VM, stitched by path). Same-origin
means the funnel, forms and cookies simply work — the tool carries an `ALLOWED_ORIGINS` CORS setting
precisely because it anticipated the split, and we get to not need it. One cert, one DNS record, no
Cloudflare Worker routing two origins by path. The box is already running and already paid for.

**The honest cost, recorded so nobody is surprised later.** idea.uk's marketing pages now depend on
one Hetzner box rather than on B2's effectively-infinite availability, and the box carries an ops
burden (patching, certs, disk, backups). That is acceptable *because* nginx serves the files with no
application in the request path — the box has to be genuinely down, not merely the tool. It is also
exactly why **Class B must stay the exception**: B2-static remains the default for the fleet. Do not
scale a box-per-site to thousands.

### Decided: **pull**, not push, for getting files onto the box

Each VM syncs *itself* — a systemd timer that sparse-checkouts its own domain folder from the shared
sites repo into `/var/www/<domain>`.

Rejected the alternative (the existing `vm-sites` GitHub Action that rsyncs over SSH) because it holds
**one SSH deploy key on the self-hosted runner, authorised on every box** — compromise the runner and
you reach the fleet. Pull inverts that: a compromised box holds a read-only repo deploy key and cannot
reach its siblings. It also sidesteps the fact that the Action targets a single `VM_HOST` secret
(currently relojistas' box, `167.233.33.159` — a *different* machine), which does not scale to a fleet.

Cost of pull, accepted: deployment latency equals the timer interval rather than being instant on
commit, and it is a new mechanism rather than the one already proven for relojistas.com.

---

## 3. What exists already (do not rebuild)

- `sites.github_repo` — the column exists, `varchar(500)`.
- `resolveGitRepoName` — `platform/orchestration/actions/helpers.go:206`. Correct logic: explicit step
  config → the site's own repo → default `"sites"`. **Nothing calls it.** The per-site-target patch
  was written and never wired.
- The shared-repo layout (domain folders at the repo root) is right for thousands of domains and stays.
- A per-IP sliding-window rate limiter (3/hr + 20/day) and a `clientIP()` extractor already exist in the
  tool — `audience_check.go:31-95` and `:100-113` — wired **only** to the free taster.
- nginx already forwards `X-Real-IP` / `X-Forwarded-For` (`setup.sh:229-230, 302-303`) and already
  rate-limits (`setup.sh:86, 226, 299`).

## 4. Landmines

- **`createOrGetRepo` creates repos PUBLIC** (`internal/adapters/git/github_client.go:307`,
  `"private": false`). If `repo_name` points at a repo that doesn't exist, the adapter will happily
  create it public and start committing sites into it. **Create target repos by hand.**
- **`updateRef` is `force: false` with no retry** (`github_client.go:420`). Concurrent commits for
  different sites can non-fast-forward and the loser fails quietly. Adding a second write stream makes
  this *more* reachable, not less.
- **Reserved-path completeness is the whole cutover risk.** The tool serves **16 routes**
  (`service.go:527-543`). The existing cutover runbook's example nginx block lists **7**. Anything
  missing is served as a static 404 — a missing `/audience-check` kills the free taster; a missing
  `/op`, `/confirm`, `/approve`, `/decline` kills the operator flow.
- **`/privacy` genuinely collides.** The static build generates `/privacy.html`; the tool serves
  `/privacy`. `try_files $uri $uri.html` would hand it to the static page unless a `location = /privacy`
  proxy block wins. There is **no** static `/terms` or `/refund-policy`, so those are uncontested.
- **Go changes are inert until the chassis image is rebuilt** (`make quick-agent-update`), unlike DB
  workflow config which takes effect immediately.
- **The tool has no CI.** It ships by building a linux/amd64 binary locally, `scp`-ing it, and
  `systemctl restart idea`.

---

## 5. The phases

### Phase 0 — Rotate the leaked credentials *(urgent, independent of everything else)*
`gqls/agentchassis` is a **public** repo and `idea.env.example` has been on `origin/main` since
**2026-06-04** carrying two real secrets: the **AWS SES SMTP credentials** (`SMTP_USER` len 20 = an
`AKIA` key id; `SMTP_PASS` len 44 = an SES password) and **`INTERNAL_API_KEY`** (len 64 = `openssl rand
-hex 32`), which gates `/confirm`, `/decline`, `/approve` and `/internal/run` — order approval and
Claude-burning runs on the live earning service. The Stripe and Anthropic keys in that file are
truncated placeholders and are **not** exposed; the money path is safe.
Repo-side scrub is ours; **rotation is the owner's and is what actually closes it**, because the values
are in pushed history.

### Phase 1 — Complete the site *(blocks the cutover)*
Build the three catalogued-but-uncomposed pages — `/guides/index.html`, `/news/index.html`,
`/tools/audience-check/index.html`. They have `pages` rows (`build_status=planned`) but `sections=[]`
and no `site_plan_sections`, so their nav links 404. Fix via the tested route: re-run
`build-site-planner`, whose `normaliseRealisedToPlanPage` unions the realised pages with the LLM
proposal so the six built pages are preserved. **Do not hand-write `site_plan_sections`.**
Cutting over before this puts the 404s on the live site.

### Phase 2 — Take idea.uk off B2
Wire the four dead wires so `sites.github_repo` actually selects the deploy target, then point idea.uk
at the VM class. See RUNBOOK §2. Requires a chassis image rebuild.

### Phase 3 — Serve it from the box
Provision `/var/www/idea.uk` + the pull timer, then the nginx cutover: static root for general pages,
all 15 reserved paths proxied to `127.0.0.1:8080`. DNS unchanged. **Rollback is nginx-only** — the
tool's binary and service are never touched. Prove `/stripe/webhook` through the new config *before*
cutting over.

### Phase 4 — Spam
Not what the existing handoff says (see RUNNING_NOTES §4 — it names the wrong process and the wrong
datastore). Order: restore the real client IP in nginx first (Cloudflare currently masks it, so any IP
ban would ban Cloudflare), then harden `/request` with a honeypot + timing check + the existing rate
limiter + an IP field on the order, then remove the existing spam from `orders.json`. Honeypot and
timing beat an IP list — spammers rotate IPs.

---

## 6. Out of scope, but found and worth knowing

Chassis-generated sites emit a contact form that **posts into a void**:
`apply_gap_plan_action.go:465` emits a `contact-form` section whose stored HTML is
`<form class="contact-form" action="/contact" method="POST">`, and the generated sites are static —
so `POST /contact` resolves to nothing and every submission is silently lost, fleet-wide. idea.uk has a
deployed `/contact.html`. This is a *dead form* problem, not a spam problem, and it wants its own
thread. Whatever contact backend is eventually built should be born with the honeypot and rate limit
that Phase 4 is retrofitting onto `/request`.
