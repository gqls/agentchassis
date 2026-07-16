# REPORT — vm-sites as a separate repo, and the static→dynamic migration path

**Date:** 2026-07-16. Requested: think hard about vm-sites as a separate repo, given several
sites will move from static to dynamic and the difference is primarily in the deploy stage;
find previous discussions; report back.

---

## 1. What previous discussions decided (the paper trail)

**June 2026 (traffic_probe running notes 10–13 Jun).** `gqls/vm-sites` was created by hand as
a private sibling of `gqls/sites` for the probe go-live. Reasons then: (a) the git-adapter's
`createOrGetRepo` creates repos **public** — a hand-made private repo avoided that landmine;
(b) same root layout as `sites` (domain folders at root) so the chassis could target it
unchanged; (c) "designating a VM/probe site = set `sites.github_repo` to the VM-sites repo".
The per-site chassis patch (P3: `resolveGitRepoName` + call-sites + `upsertSite` plumbing)
was **specified but not landed**.

**HANDOFF Thread B (13 Jun).** The vision: a VM/backend site becomes a *normal* chassis build
that "differs only by (a) deploying to a VM and (b) optionally including backend components."
This is exactly the premise of the current question — the difference should live in the
deploy stage.

**idea.uk workstream (14–16 Jul, PLAN/RUNNING_NOTES/RUNBOOK).** The most developed discussion,
with the owner's constraint on record: *"There will be several thousand domain names so
individual repos will be ungainly."* Decisions:
- **Three hosting classes.** A: static→B2 (default, the thousands). B: static→VM+backend (the
  handful that sell something). C: dynamic rendering on the box — **rejected** (would need
  fleet-wide DB access from boxes or a second render path; loses availability/cacheability/
  uniformity). "The VM is a second **sink**, not a second renderer."
- **Pull, not push**, for file delivery: each box sparse-checkouts **its own folder from
  `gqls/vm-sites`** with a read-only deploy key. The vm-sites push Action was rejected for
  *fleet* use because the runner holds one SSH key authorised on every box (compromise the
  runner → reach the fleet) and targets a single `VM_HOST` (relojistas' box). Push survives
  as a legacy mechanism for relojistas only, behind a **`deploy-targets.json` allowlist**
  (RUNBOOK §2b — designed, not yet applied).
- **The four dead wires** (`upsertSite` RETURNING, `EnsureSiteRecord` return map,
  `git_commit`→`resolveGitRepoName`, `deploy_image_asset` hardcode) — wired and shipped in
  chassis v1.0.1123 (16 Jul).
- Open decision recorded: does relojistas migrate push→pull, or do both coexist? (Coexist is
  fine; the allowlist enables it.)

**Concept register (`dynamic-applications`).** Tier model: Tier 1 static+dynamic components;
Tier 2 thin per-site backends (business logic stays in agents); Tier 3 full generated
applications with "one site one repo one deployment". So the long-range picture is *three*
storage shapes: `sites` (Class A), `vm-sites` (Class B), per-app repos (Tier 3) — vm-sites is
the coherent middle.

## 2. Live evidence from today (relojistas, the first real exercise of the wiring)

relojistas.com was submitted fresh today with `github_repo='vm-sites'` set before any deploy.
Running image v1.0.1125 (includes the wiring). Result:
- **News feed commit → `gqls/sites` (wrong).** The content-feed-orchestrator workflow never
  loads the site record, so `resolveGitRepoName` found no `site_record.github_repo` in
  collected data and defaulted.
- **Page deploys → `gqls/sites` (wrong).** Verified: `page-rerender` / `build-dispatch-loop`
  orchestrations have **no `site_record` key at all** — only planner-tier workflows run
  `ensure_site_record`.
So despite correct data and shipped wiring, **every relojistas artefact is landing in the B2
repo, invisible**, and the box webroot still holds the probe page. The per-site target
resolution is structurally incomplete: it depends on *which workflow* commits, when it should
depend only on *the site row*.

## 3. Analysis — is a separate vm-sites repo right?

**Yes — keep it.** The premise ("difference is primarily in the deploy stage") is correct,
and the repo boundary is the cleanest available encoding of that difference:

1. **Deploy keys are repo-scoped, not path-scoped.** The pull model gives each box a
   read-only key. If Class-B sites lived inside `gqls/sites`, every box's key would read the
   *entire* portfolio (thousands of domains, including unlaunched/staging content) — an
   enumeration and scraping gift on any box compromise. A separate vm-sites repo caps a box's
   read blast-radius at the handful of Class-B sites. GitHub offers no path-scoped deploy
   keys, so this argument cannot be engineered away within one repo.
2. **Sink separation by construction.** The `sites` Action B2-syncs every changed folder. VM
   sites inside `sites` would perpetually double-deploy to B2 (the "elaborate staging copy"
   pathology idea.uk suffered) or force exclusion lists into the fleet Action. Separate repos
   make each Action total over its repo — no filters, no drift.
3. **Operational blast-radius.** The B2 path currently runs through a half-crash-looping
   runner (idea.uk notes §L). vm-sites is insulated from that, and vice versa.
4. **The repo flip is the migration primitive.** Class A→B = flip one column. That is as
   close to "the difference is only the deploy stage" as it gets.

**The costs (and their remedies):**
1. **A default-repo misroute bug-class** — any committing workflow that doesn't carry
   `site_record.github_repo` silently deploys to `sites`. Proven twice today. *Remedy:
   resolve from the DB, not from workflow state (recommendation 1).* Note the alternative
   (one repo) would NOT fully avoid this — routing would just move into the Action layer.
2. **Two flags that must agree** — `github_repo='vm-sites'` and `deploy_config.target='vm'`
   both mark Class B; they can drift. *Remedy: treat `deploy_config.target` as semantic
   truth, derive/validate `github_repo` against it (recommendation 3).*
3. **Migration mechanics** — moving a site A→B leaves a stale copy in `sites`/B2 and its git
   history behind. Acceptable (history is a build log, not an asset — `clients_db` is the
   source of truth; the artefact is regenerable), but the stale-copy cleanup must be part of
   a scripted migration (recommendation 4).

## 4. Recommendations

1. **Make repo resolution workflow-independent (the structural fix).** In `git_commit` (and
   `deploy_image_asset`): when neither step config nor `site_record.github_repo` supplies a
   repo, **query `sites.github_repo` by domain** before defaulting to `"sites"`.
   `ActionParams` already carries `DB`; the domain is already extracted. One query per
   commit; kills the entire misroute class for all ~40 workflows, present and future. Interim
   data-only fix for relojistas news: add a load-site step to the content-feed-orchestrator
   workflow (agent_definitions JSON — takes effect immediately, no image rebuild).
2. **Apply the vm-sites Action allowlist NOW** (`deploy-targets.json`, RUNBOOK §2b as
   written: `{"relojistas.com": "167.233.33.159"}`). Today the Action rsyncs *every* changed
   folder to relojistas' box; the moment a second domain (idea.uk) lands in vm-sites it
   deploys to the wrong machine. Hosts become data, not secrets.
3. **One source of truth for the class.** `deploy_config.target='vm'` is the semantic flag;
   `github_repo` is the routing consequence. Either derive it or add a consistency check
   (discovery check: `target='vm'` XOR `github_repo='vm-sites'` → flag).
4. **Script the A→B migration** (the "several sites will move" path), roughly:
   (a) provision/extend box (setup.sh multi-vhost, DNS); (b) `UPDATE sites SET
   github_repo='vm-sites', deploy_config.target='vm'`; (c) trigger full re-render/redeploy so
   artefacts land in vm-sites; (d) add domain to `deploy-targets.json` (push) or provision
   the sparse-checkout pull timer (preferred); (e) delete the domain folder from `gqls/sites`
   + B2 (stop the staging-copy drift); (f) verify serving from the box. Reversible by
   flipping back (B2 artefacts regenerate on next deploy).
5. **Converge delivery on pull.** idea.uk pioneers the sparse-checkout pull; migrate
   relojistas' box from push→pull when convenient; retire the push Action after the last
   push domain moves. Coexistence in the meantime is safe *only* behind recommendation 2.
6. **Keep Class B the exception.** Reaffirm the idea.uk principle: B2-static remains the
   default for the thousands; a box per site does not scale and isn't meant to.

## 5. One-line answer

The separate repo is correct and already settled by the idea.uk decisions — it is the
security and sink boundary that makes "same artefact, different deploy stage" real; what's
missing is not repo consolidation but **workflow-independent repo resolution (rec 1)** and
the **Action allowlist (rec 2)**, both exposed by relojistas' first live run today.
