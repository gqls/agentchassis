# RUNBOOK — release enumeration (`bugs_open/237`)

Every command this lane had to get right, with its gotcha attached. When one
changes, change it **here**.

---

## 1. The coverage gate

```bash
make check-release-coverage                                        # green on a compliant tree
make check-release-coverage AGENT_DEPLOY_SERVICES="agent-chassis"  # MUST name render-audit-adapter
make -n deploy-agents | grep -c render-audit-adapter               # the retag loop
make -n deploy-core   | grep -c "Release coverage OK"              # the gate, via update-kustomization-images
```

> ⚠ **Never measure at `release`.** Since `bugs_open/249` made `release` a
> `pinned_sweep` shell block, `make -n release | grep <service>` returns **0
> whether or not the service is in the release**. 237's own 2026-08-10 table
> recorded that recipe as *passed*; it now always says "absent". Measure at
> `deploy-agents` / `deploy-core`.

> ⚠ **`make -n` does not expand a shell `for` loop.** Counting action lines reads
> **9** where the old copy-pasted blocks read **37**, with the resolved sets
> identical. Compare the **resolved service set** (expand the make variable, or run
> the loop under `kubectl`/`sed` stubs), never the line count.

The gate is only armed for services whose image **the release builds**. A service
pinning an image outside `RELEASE_IMAGES` is invisible to it by construction —
that is exactly Decision B, not a bug in the gate.

## 2. Enumerate the frozen set (which services are not on the fleet tag)

```bash
for d in deployments/kustomize/services/*/; do
  s=$(basename $d); o=$d/overlays/production/uk_001/kustomization.yaml
  [ -f "$o" ] || { printf "%-36s NO-UK001-OVERLAY\n" "$s"; continue; }
  t=$(grep -m1 'newTag:' $o | awk '{print $2}')
  [ "$t" = "$FLEET_TAG" ] && continue
  printf "%-36s %-14s %s\n" "$s" "$t" "$(git log -1 --format='%h %ad' --date=short -- $o)"
done
```

Set `FLEET_TAG` to the current `IMAGE_TAG` from the makefile (~line 16).

- A **blank** tag column means the overlay pins no image — check what the *base*
  uses before treating it as frozen. As of 2026-08-18 every such row runs
  `postgres:16-alpine` (SQL/ConfigMap checks, a different staleness mechanism) and
  is **not** part of this class.
- The overlay's last-touched commit doubles as the **build commit**: these overlays
  were written once, on the day the service was created. That is also the evidence
  the deploy targets are never run.

## 3. Which binary does a check service actually run?

```bash
grep -nE '^(CMD|ENTRYPOINT)' build/docker/backend/<service>.dockerfile
```

Do this **before** reasoning about a check's staleness.
`shared-output-fields-check` and `removed-config-keys-check` are both
`cmd/config-key-audit` under different `CMD`s, and neither CronJob sets `args` — so
the Dockerfile `CMD` is the entire behavioural difference and nothing in the
kustomize manifests reveals it.

## 4. Is a check functionally stale? (the registry census — the decisive one)

```bash
acts () { git grep -h -A1 'RegisterActionInputSpec(' "$1" -- platform/ internal/ \
          ':(exclude)*_test.go' | grep -o '"[a-z0-9_]\+"' | tr -d '"' | sort -u; }
acts HEAD            > /tmp/head.acts
acts <build-commit>  > /tmp/built.acts
wc -l /tmp/head.acts /tmp/built.acts
comm -23 /tmp/head.acts /tmp/built.acts     # actions the running binary cannot see
```

Three gotchas, each of which produced a wrong number first:

1. **`':(exclude)*_test.go'` is mandatory.** Test-only registrations
   (`test_action_removed_keys`, `strict_marker_test_action`, …) are ~22% of the raw
   count and are in no shipped binary. Without the exclusion HEAD reads 201, not 169.
2. **`-A1` is mandatory.** Five non-test registrations wrap the name onto the next
   line, so a same-line grep undercounts by five.
3. **The delta is the finding, not the count.** Run it at both commits; if the two
   totals match, the registry was quiet and the question is closed. (They did not
   match, which is why this measurement counted.)

## 5. Does a check use the changed code, or merely link it?

```bash
grep -hn "datahelpers\.[A-Z][A-Za-z]*\|actions\.[A-Z][A-Za-z]*" cmd/<pkg>/*.go \
  | grep -v _test | grep -o "[a-z]*\.[A-Z][A-Za-z]*" | sort -u
```

Then for each symbol, ask whether its **definition** moved:

```bash
git diff <build-commit>..HEAD -- <package-dir> | grep -E "^[+-].*(func .*\b<Sym>\b|type <Sym>\b)"
```

A count inherited from a linked package is **not** a property of the service —
`verifier-remit-check` links the package holding the registry and never reads it.

## 6. Prove the freeze is live (not just in the repo)

```bash
kubectl -n ai-persona-system get cronjob -o custom-columns=\
'NAME:.metadata.name,SCHEDULE:.spec.schedule,SUSPEND:.spec.suspend,\
IMAGE:.spec.jobTemplate.spec.template.spec.containers[0].image,LAST:.status.lastScheduleTime'
```

`IMAGE` is the served tag and `LAST` proves it ran. A repo overlay is what *would*
be applied; this is what *is* applied.

## 7. Ask a service what it is running

```bash
kubectl -n ai-persona-system logs -l app=<svc> --tail=3000 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <the-stamp>      # "did my fix ship?"
```

Startup line, so it scrolls — an empty result means "not in range", **not**
"unstamped". Per **service**, not per fleet. Never `strings`; if you probe the
binary, run a control sha that must be absent alongside one that must be present.

## 8. Things that will refuse you on this lane

- **Council refuses makefile-only submissions client-side** (scope is `platform/`,
  `internal/`, `pkg/`). No credits spent, no `FORCE`, and no commit here carries a
  review trailer.
- **The landmine verifier cannot grade makefile-footprinted entries** — its index is
  Go-only, so it returns `UNVERIFIABLE`. That is not doubt; write a hand-run check
  into the entry.
