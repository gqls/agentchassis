# `daemon.json` — BuildKit cache eviction (owner ruling, 2026-08-25)

**Docker's build cache had NO automatic eviction at all.** That is how it reached **539 GB** on this
box — 6,052 records, 437.9 GB of it reclaimable — while Go's cache trims at 5 days and
`systemd-tmpfiles` reaps `/tmp` at 10. It was the only reaper-shaped hole here that was *absent*
rather than merely too slow.

This config evicts build cache **unused for 7 days** (`168h`), keeping a 10 GB floor so a quiet week
does not force a cold rebuild.

## Installing it — needs root, so the owner runs this

```bash
sudo dockerd --validate --config-file /home/ant/projects/agentchassis/deployments/docker/daemon.json
sudo install -m 0644 -D /home/ant/projects/agentchassis/deployments/docker/daemon.json /etc/docker/daemon.json
sudo systemctl restart docker
systemctl is-active docker && docker info >/dev/null && echo "docker is back"
```

**Rollback, if anything goes wrong:** `sudo rm /etc/docker/daemon.json && sudo systemctl restart docker`.
There is no existing file to preserve — `/etc/docker/daemon.json` did not exist as of 2026-08-25.

> ## ⚠ `dockerd --validate` DOES NOT VALIDATE VALUES — only key names
>
> `[MEASURED 2026-08-25]` it printed **`configuration OK`** for
> `{"reservedSpace": "not-a-size"}`, and rejected only an unknown *key*
> (`bogus-key-xyz` → "directives don't match any configuration option"). So a
> passing validate proves you have not misspelled a **setting**; it proves nothing
> about a misspelled **size**.
>
> **This matters because a daemon.json Docker cannot parse stops the daemon from
> starting**, and on this box that takes the whole build pipeline with it. So the
> restart above is not a formality: **run it, then confirm Docker actually came
> back** with the `is-active`/`docker info` line, and keep the one-line rollback to
> hand. Do not run the install and walk away.
>
> The control is the transferable part: **an "OK" from a validator is only evidence
> if you have watched that validator say NO to something.** It took one deliberately
> broken value to learn this one is half-blind.

## What it does NOT do

**There is no size ceiling** — the owner chose age-only over "7 days AND a 100 GB cap". So a heavy
build week can still grow the cache a long way *inside* the 7-day window; age bounds how long
anything survives, not how much accumulates. `[MEASURED 2026-08-25]` the cache went from 1.272 GB
to **16.8 GB in about 80 minutes** of ordinary fleet activity after the prune, so the regrowth rate
is real. If that becomes a problem, adding `"maxUsedSpace": "100GB"` to the policy is the knob, and
the nightly scratch report is the place to watch for it.

Related: `scripts/docker-build-retention.sh` bounds the *number of release builds* kept (the owner's
actual ask — the count lives on the image side, not in the cache) · full account in
`docs/agent_docs/docs024_key_docs_latest/tmpfs_exhaustion/`.
