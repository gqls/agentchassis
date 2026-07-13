# relojistas.com — manual go-live checklist (P1, first domain)

Exact commands for the operator. Grounding: the Wayback snapshot in the project
confirms relojistas.com was a Spanish watch FORUM (boards: general, marcas,
ferias, ventas/outlet) — so the probe is Spanish, search-kind, thanks page
`/gracias.html`. Deliverables referenced here: `relojistas-site/` (index +
gracias), `vm-deploy/` (setup.sh, env example, two workflows), `site-engine/`
(engine source).

Caveats before starting: repointing DNS stops any parking revenue on the domain;
the marketplace's ~1.2M visit estimate is unverified — your own nginx logs +
beacon counts will give the truth within days.

---

## 0. Repos and secrets (one-time, laptop)

```bash
# Engine repo (manual creation so YOU choose visibility; the git-adapter
# auto-creates repos as PUBLIC, so don't leave this to it)
gh repo create OWNER/site-engine --private --clone && cd site-engine
cp <outputs>/site-engine/{go.mod,env.go,store.go,service.go,main.go} .
mkdir -p .github/workflows
cp <outputs>/vm-deploy/deploy-engine-to-vm.yml .github/workflows/
git add -A && git commit -m "site-engine v1" && git push -u origin main

# Content repo
gh repo create OWNER/vm-sites --private --clone && cd ../vm-sites
mkdir -p .github/workflows
cp <outputs>/vm-deploy/deploy-to-vm.yml .github/workflows/
git add -A && git commit -m "vm-sites scaffold" && git push -u origin main
```

> LAYOUT CONFIRMED (2026-06-11): the live `gqls/sites` repo keeps domain
> folders at the repo ROOT (no `sites/` directory) — matching what the
> git-adapter produces. `vm-sites` uses the same root layout, and
> `deploy-to-vm.yml` is written for it (triggers on root changes,
> `paths-ignore` for `.github/**`/README/LICENSE). The `sites/**` variant seen
> earlier was a stale copy living in `agentchassis/.git/workflows/` — a
> location GitHub never reads; consider deleting it to avoid future confusion.

```bash
# Deploy SSH key (used by BOTH repos' Actions)
ssh-keygen -t ed25519 -f ./deploy_key -N "" -C "vm-deploy"
# In GitHub → each repo → Settings → Secrets and variables → Actions:
#   VM_HOST    = <box IP or hostname>
#   VM_USER    = deploy
#   VM_SSH_KEY = contents of ./deploy_key (the PRIVATE key)
```

## 1. Box + DNS

```bash
# Provision the smallest Ubuntu 24.04 VM (idea.uk model: Hetzner CX-class is
# ample; the engine is I/O-bound). Note its IP.

# DNS (at the registrar / Cloudflare): A record  relojistas.com -> <IP>
# Start DNS-only (grey cloud if Cloudflare) so certbot's http-01 works plainly;
# the proxied/orange option is runbook §8, switchable later.
# v1 is apex-only; www.relojistas.com handling is a listed follow-up.
```

## 2. Deploy user + provision (as root on the box)

```bash
adduser --disabled-password --gecos "" deploy
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
cat > /home/deploy/.ssh/authorized_keys   # paste deploy_key.pub, then Ctrl-D
chown deploy:deploy /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys
```

```bash
# From the laptop: build the engine and ship it + setup.sh
cd site-engine
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o site-engine .
scp site-engine root@<IP>:/tmp/site-engine
scp <outputs>/vm-deploy/setup.sh root@<IP>:/tmp/

# On the box (root). DEPLOY_USER wires webroot ownership + the sudo hook.
DOMAINS="relojistas.com" \
LETSENCRYPT_EMAIL=you@your-real-domain.tld \
DEPLOY_USER=deploy \
ENGINE_BINARY_PATH=/tmp/site-engine \
bash /tmp/setup.sh
```

## 3. Engine env (on the box)

```bash
KEY=$(openssl rand -hex 24)
sed -i "s|^INTERNAL_API_KEY=.*|INTERNAL_API_KEY=$KEY|" /etc/site-engine/site-engine.env
sed -i "s|^THANKS_PATH=.*|THANKS_PATH=/gracias.html|"   /etc/site-engine/site-engine.env
systemctl restart site-engine
echo "STATS KEY: $KEY"   # keep this; /stats requests need it
```

## 4. First content (manual rsync; the Action takes over later)

```bash
# From the laptop:
rsync -az --delete <outputs>/relojistas-site/ deploy@<IP>:/var/www/vm-sites/relojistas.com/
```

## 5. Verify end to end

```bash
curl -sS https://relojistas.com/health           # {"ok":true}
# Browser: https://relojistas.com → submit "correa Omega Seamaster"
#   → 303 to /gracias.html
curl -sS -H "X-Internal-Key: $KEY" https://relojistas.com/stats
# expect: visits >= 1, events >= 1, events_per_1k_visits computed
ssh root@<IP> 'tail -c 600 /var/lib/site-engine/intent_events.json'
```

## 6. Switch the engine to Action-driven updates (proves the seam)

```bash
# Any push to *.go / go.mod in the site-engine repo now builds, ships, and
# swaps via the sudo hook. Test with a no-op:
cd site-engine && git commit --allow-empty -m "engine deploy test" && git push
# Watch: repo → Actions; then on the box:
journalctl -u site-engine -n 20 --no-pager
```

## Follow-ups (listed, not blocking)
- www.relojistas.com: add the A record, extend `server_name`/cert (setup.sh
  change), or handle at Cloudflare.
- Cloudflare proxied mode (runbook §8): cache-bypass the API paths, nginx
  real-IP, Full (strict); gains CF-IPCountry for the `country` field.
- Content via the pipeline instead of rsync: needs the chassis patch
  (resolveGitRepoName + ensure_site_record/upsertSite), the planner gate, the
  designation UPDATE (`github_repo='vm-sites'`, deploy_config capabilities),
  and the repo-layout answer above.
- Off-box collection of intent_events (P4).
