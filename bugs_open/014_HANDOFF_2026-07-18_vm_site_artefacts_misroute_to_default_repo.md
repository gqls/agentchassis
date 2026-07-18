# 014 — VM-hosted site artefacts silently deploy to the default `sites` repo (FIXED, two causes)

**Found:** 2026-07-18, during the relojistas.com rebuild (first real exercise of per-site
deploy targeting). **Status:** both causes fixed; recorded because the *second* cause is a
trap that will recur, and because the pattern generalises beyond deploy routing.

## Symptom

relojistas.com was correctly marked as a VM-hosted site (`sites.github_repo='vm-sites'`,
`deploy_config.target='vm'`) before anything deployed. Every artefact nevertheless landed in
**`gqls/sites`** (→ B2, invisible) instead of **`gqls/vm-sites`** (→ the box): news JSON,
page HTML, and all nine generated images + `styles.css`. The live box kept serving the old
hand-made probe page. Nothing errored — every commit reported success, to the wrong place.

## Evidence

- Feed commit result: `"repo_url": "https://github.com/gqls/sites", "repo_name": "sites"` —
  from `orchestration_states.collected_data->'news_commit_result'`, for a site whose
  `github_repo` was `vm-sites`.
- Page deploys likewise: `repo_name='sites'` on the `page-rerender` orchestration.
- `SELECT ... collected_data ? 'site_record'` → **false** for `page-rerender` and
  `build-dispatch-loop` orchestrations.
- After fixing cause 1 and shipping it (v1.0.1126, symbol verified in-pod), page rerenders
  **still** committed to `sites`.

## Root cause — TWO independent causes, second only visible after fixing the first

**Cause 1 — resolution depended on workflow state that most workflows never populate.**
`resolveGitRepoName` (`platform/orchestration/actions/helpers.go`) resolved:
explicit step config → `site_record.github_repo` **from CollectedData** → default `"sites"`.
Only planner-tier workflows run `ensure_site_record`; `page-rerender`,
`build-dispatch-loop` and `content-feed-orchestrator` do not. So for those the middle term
was always absent and every VM site fell through to the B2 default. The per-site target was
therefore a property of *which workflow happened to be committing*, not of the site.

**Cause 2 — three agent definitions pinned `repo_name: "sites"` in their git_commit step
config**, which correctly outranks the fallback. Fixing cause 1 changed nothing for them:
- `page-rerender` → step `deploy_page`
- `site-deployer` → step `deploy_to_git`
- `deployer-agent` → step `commit_to_git`
The pin was invisible because it names the same value as the default — it looked like a
no-op restating the obvious, and it silently defeated the new resolution.

## Fix (both applied)

1. **`resolveGitRepoNameDB`** (helpers.go) — same precedence plus a DB fallback:
   explicit config → collected `site_record.github_repo` → `SELECT github_repo FROM sites
   WHERE domain=$1` → `"sites"`. Wired into `git_commit` (domain now extracted *before*
   resolution) and `deploy_image_asset`. Deploy target is now a property of the site row,
   independent of the calling workflow. Shipped v1.0.1126. Tests:
   `platform/orchestration/actions/git_repo_resolution_test.go` (7 subtests: precedence,
   NULL repo, unknown domain, DB error → safe default, nil DB/empty domain).
2. **Removed the three pins** via jsonb `#-` on `agent_definitions.default_config`
   (DB config, live immediately, no image needed). Diagnosis/fix agents pinning
   `agentchassis` were deliberately left alone.

## How to verify

```sql
-- must be 0: no git_commit step may pin a repo
SELECT count(*) FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') s(k,v)
WHERE deleted_at IS NULL AND v->>'action'='git_commit' AND v->'config' ? 'repo_name';
```
```bash
# in-pod symbol (never the tag)
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c resolveGitRepoNameDB'
```
Then rerender a page on a VM site and confirm `collected_data->'deploy_result'->'response'
->'data'->>'repo_name'` is the site's own repo. Confirmed live: all four relojistas pages
plus the feed committed to `vm-sites`, and the box now serves the built site.

## Collateral trap worth knowing (cost nothing only because it was caught first)

The vm-sites Action rsyncs with `--delete`. The repo held only `assets/js/snippets.js` for
relojistas while the **box** held the hand-made `index.html`/`gracias.html` from the manual
go-live. The first pipeline commit would have deleted the live pages. Mitigation applied
before the first deploy: mirror the live webroot into the repo (checksum-verified), commit,
*then* let the pipeline take over. **Any repo→box sync with `--delete` needs the box's
current state committed first.**

## Transferable pattern (also filed to 016b §9)

A per-entity routing value resolved from *workflow state* is only as reliable as the
workflows that populate that state — resolve it from the entity row instead. And when you
add a fallback, **grep for explicit config that already sets the same key**: a pinned value
identical to the old default is indistinguishable from documentation and silently outranks
the new logic.
