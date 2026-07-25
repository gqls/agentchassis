# NOTES — VM estate (framework-controlled provisioning)

Running technical log, append-only, newest at the bottom. Evidence and commands
inline; missteps recorded as they happen.

---

## 2026-07-25 — walkthrough of setup.sh, and the shape of the merge

Opened on the owner's direction to bring `setup.sh` under framework control and
to converge it with the tools-api island rather than run them as separate
projects. Design record: `PLAN_2026-07-25_framework_controlled_vm_estate.md`.

**Read the whole script (585 lines) rather than skimming it.** Worth saying,
because the two findings below are both invisible to grep — one is a bash scoping
rule, the other is an absence (a function called unconditionally that should be
conditional).

### Defect A — `local` at top level, verified

```bash
$ bash -c 'set -euo pipefail; for d in a; do if [[ ! -d /nope ]]; then local x=""; fi; done; echo continued'
bash: line 1: local: can only be used in a function
# exit 1 — "continued" never prints
```

Line 496 of setup.sh sat inside the top-level TLS loop, in the branch taken when
a domain has **no certificate yet**. Consequences, stated precisely:

- **The owner's pending relojistas re-run is UNAFFECTED** — relojistas.com has a
  cert, so `[[ ! -d /etc/letsencrypt/live/$d ]]` is false and the branch never
  runs. Verified the cert directory exists via the live conf carrying `ssl_certificate`.
- It is fatal on the "add a domain and re-run" path the file header advertises,
  and it fails *between* the two nginx writes — stage 1 (HTTP) written, stage 2
  (HTTPS upgrade) never reached, engine not restarted.

Fixed in place (`extra_san=""`, comment explaining why it is not `local`);
`bash -n` clean; remaining `local` uses all confirmed inside functions
(lines 236, 253, 295, 313).

**The idea.uk original does NOT contain this bug** — it entered during the fork.
A one-fork-only defect is the cheapest possible illustration of why two copies
of a provisioning script is the problem, not the file's length.

### Defect B — per-site policy in a shared code path [OPEN]

`static_body()` calls `legacy_feed_locations()` and `api_locations()`
unconditionally, so **every** domain on the box gets relojistas' vBulletin
`/external.php` → `/feed.xml` rewrites and a `/buscar` route.

```bash
# latent, not live — verified only one domain is on the box:
ssh root@167.233.33.159 'grep -h server_name /etc/nginx/sites-enabled/vm-sites.conf | sort -u'
#     server_name relojistas.com;
```

Deliberately NOT patched. A conditional in bash would work and would also
entrench the design being replaced; per-site policy belongs in per-site DB
state (PLAN Part 2).

### Divergence, measured

```bash
A=docs/.../idea.uk/golang_files/setup.sh          # 393 lines
B=docs/.../traffic_probe/deploy_setup/vm-deploy/setup.sh   # 585 lines
diff <(sed 's/[[:space:]]*$//' $A) <(sed 's/[[:space:]]*$//' $B) | grep -c '^[<>]'   # 614
comm -12 <(sort -u $A) <(sort -u $B) | wc -l                                        # 61
```

**614 diverged lines, 61 in common.** Two scripts with a common ancestor, not one
script in two places. The island's `docker-compose.yml` + `Caddyfile` +
`backup_pg.sh` are a third lineage starting now.

### Reuse: the provisioning primitives already exist

Before designing anything, checked what the chassis can already do to a remote
machine (CLAUDE.md: reuse existing machinery). It can do all of it:

```bash
grep -n "provision" platform/orchestration/actions/registry.go
# dispatch_thunder_provision   -> instance_ip, ssh_user, ssh_key_secret_name, provisioning_id
# dispatch_thunder_ssh_exec    -> run a command on a provisioned instance; exit_code/stdout/stderr
# dispatch_thunder_ssh_status  -> reachability + status command
```
```sql
SELECT type FROM agent_definitions WHERE is_active AND type LIKE '%deploy%' OR type LIKE '%provision%';
-- asset-deployer, deployer-agent, gpu-provisioner, site-deployer, tool-deployer
```

`gpu-provisioner` + `training-launcher` + `thunder-training-monitor` is already
provision → execute → monitor, in production for GPU training. A VM estate needs
that contract against another provider, not a new concept.

**And the promise is already written down:** setup.sh's header says it serves
"the chassis service-deployer LATER". There is **no `service-deployer`** in the
codebase — `grep -rl "service-deployer\|service_deployer" platform/ internal/ cmd/`
returns nothing. It has been a comment for months.

### The design point that took the longest to see

The first sketch was "chassis holds SSH keys, pushes config to all boxes". That
is wrong for the island, and wrong in a way that would have been expensive to
discover after building it: the island's entire rationale is that *the production
cluster appears nowhere in its path* and *nothing on it holds a production
credential* (`RUNBOOK_island.md`). Naive framework control inverts that.

Resolution: **merge the generator, not the trust boundary.** Shared profile
schema + shared renderer + shared drift check; but the public estate is pushed to
over SSH while the island *pulls* its rendered profile outbound-only — the same
direction cloudflared already dials. Merged in every sense that removes
duplication, unmerged in the one sense that would spend the island's isolation.

### Framing recorded for the plan

The box's nginx conf is to the machine what `rendered_html` is to a page — and
this is not a metaphor reached for, it is the same defect already ruled on in
contract doc `003`. A surgical hand-edit to the live conf on 07-19 (legacy feed)
would have been destroyed by the next generator run; it was reconciled by hand on
07-24. That is precisely the manual repair the platform refuses to accept for
pages, running unchallenged on machines.
