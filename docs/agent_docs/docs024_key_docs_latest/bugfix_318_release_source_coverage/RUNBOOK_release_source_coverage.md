# RUNBOOK — `bugs_closed/318` release source coverage

Every command this lane had to get right, with its gotcha attached. Change a command
HERE, not in your scrollback.

---

## R1 — is anyone else on this?

```bash
python3 scripts/who-owns.py 318
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT id, item_type, status, left(summary,110), created_at
    FROM site_work_items
   WHERE status NOT IN ('complete','cancelled','rejected')
     AND (summary ILIKE '%release%' OR summary ILIKE '%RELEASE_IMAGES%'
          OR summary ILIKE '%coverage gate%' OR summary ILIKE '%build-backend%')
   ORDER BY created_at DESC LIMIT 20;"
```

⚠ `who-owns.py` reads **commits**, so a session mid-fix is invisible. Check the dirty
tree too (`git status --porcelain -- makefile 'cmd/**' 'platform/**'`).

---

## R2 — the set difference that found the live defect

```bash
make -n build-backend 2>/dev/null | grep -oE 'docker build -t [^ ]+' \
  | sed 's#.*/##; s#:.*##' | sort -u > /tmp/built.txt
awk '/^RELEASE_IMAGES :=/{f=1} f{print} f&&!/\\$/{exit}' makefile \
  | sed 's/^RELEASE_IMAGES :=//; s/\\//' | tr -s ' \t' '\n' | grep -v '^$' | sort -u > /tmp/declared.txt
comm -23 /tmp/declared.txt /tmp/built.txt   # declared, never built  → release dies at push
comm -13 /tmp/declared.txt /tmp/built.txt   # built, never pushed    → wasted build minutes
```

⚠ **`tr ' ' '\n'` alone is not enough** — the `RELEASE_IMAGES` continuation lines are
indented with **tabs**, so a space-only split yields entries with a leading tab and
`comm` then reports every one of them as a difference. Use `tr -s ' \t' '\n'`. This bit
me on the first run and produced a nine-line false positive.

⚠ **Do not measure this at `make -n release`.** Since `bugs_open/249` made `release` a
`pinned_sweep` shell block, `make -n release` prints the sweep and descends into
nothing. `make -n build-backend` is the real preview (BLD-020's own landmine).

---

## R3 — does the release actually have an image to push?

```bash
docker images --format '{{.Repository}}:{{.Tag}}' | grep -E '^aqls/<service>:' | sort
```

⚠ **Always run two controls in the same breath**: a service that IS in
`build-backend` (`agent-chassis`, `verifier-remit-check` — each carries an image at
every recent tag) and the service under test. A bare "no image found" is
indistinguishable from a typo in the grep.

⚠ Local images print as `aqls/x`, the makefile pushes `docker.io/aqls/x`. Docker
normalises `docker.io` away, so these are the same image — do not read the missing
prefix as a different repository.

---

## R4 — what the cluster is actually running

```bash
kubectl -n ai-persona-system get cronjob -o custom-columns=\
'NAME:.metadata.name,SCHEDULE:.spec.schedule,SUSPEND:.spec.suspend,\
IMAGE:.spec.jobTemplate.spec.template.spec.containers[0].image,LAST:.status.lastScheduleTime'
```

⚠ `LAST=<none>` means never fired, which for a brand-new CronJob is correct and for an
old one is a defect — the column does not distinguish them. Check `metadata.creationTimestamp`.

⚠ A CronJob can exist in the cluster with **no overlay on disk**, and an overlay can
exist on disk with **no CronJob in the cluster** (`capped-schedule-ordering-check`,
today). The filesystem and the cluster are two different enumerations and neither is a
superset of the other.

---

## R5 — the overlay census (service → image → pinned tag)

```bash
for f in deployments/kustomize/services/*/overlays/production/uk_001/kustomization.yaml; do
  [ -f "$f" ] || continue
  svc=${f#deployments/kustomize/services/}; svc=${svc%%/*}
  img=$(awk '/^images:/{i=1;next} i&&/name:/{print $NF;exit}' "$f")
  tag=$(awk '/^images:/{i=1;next} i&&/newTag:/{print $NF;exit}' "$f")
  printf '%-42s %-46s %s\n' "$svc" "${img:-<none>}" "${tag:-<none>}"
done
```

⚠ **This reads the WORKING TREE, and the working tree is not what git has.**
`deploy-agents` seds `newTag` in place and nothing commits it — 26 of these files were
dirty on 2026-08-22, `agent-chassis` committed at `v1.0.1239` against a tree value of
`v1.0.1323`. So `git log -- <overlay>` does **not** tell you when the pin last moved,
and any staleness test built on that history is testing a fiction. Read the artefact
(BLD-019 OCI label / binary stamp, BLD-023 `service_binary_capabilities.git_commit`)
when you need to know what commit an image came from.

---

## R6 — the source closure of a service

```bash
grep -oE '\./cmd/[a-zA-Z0-9_-]+|cmd/[a-zA-Z0-9_/-]+\.go' build/docker/backend/<svc>.dockerfile
go list -deps ./cmd/<pkg> | grep '^github.com/gqls/agentchassis/'
grep -E '^COPY --from=builder' build/docker/backend/<svc>.dockerfile
```

⚠ Six images build `./cmd/config-key-audit` and differ only in CMD flags and a COPY'd
acks JSON. A closure computed from Go deps alone calls all six identical. The COPY set
is part of the closure.

⚠ `git-adapter` and `remote-job-spawner` build `cmd/X/main.go`, not `./cmd/X`; the
regex above covers both. `github-actions-runner` has no Go build at all.

---

## R7 — proving the gate discriminates

**Never mutate the live `makefile` to test it.** On 2026-08-22 a session did exactly
that and another session committed the file inside the window; HEAD got the restored
version by luck (`WRONG_CALLS.md`, `f016b07ec`). Either:

```bash
make check-release-coverage AGENT_DEPLOY_SERVICES="agent-chassis"   # must NAME a service
cp makefile /tmp/mf && <edit /tmp/mf> && make -f /tmp/mf check-release-coverage
```

or — the reason this lane wants the predicate in Go — a table test that can be mutated
without touching a file 40 sessions share.
