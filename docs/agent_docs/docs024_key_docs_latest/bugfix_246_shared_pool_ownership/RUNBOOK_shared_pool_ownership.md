# RUNBOOK — bugfix 246, shared `*sql.DB` pool ownership

Every command here was needed to get a fact right. The gotcha is attached to the
command; when one changes, change it HERE.

---

## 1. Is anyone else on this bug?

Three checks, blind in different ways. Run all three — any one alone is weak.

```bash
# (a) LIVE sessions. who-owns.py reads COMMITS and cannot see a session mid-fix.
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(ls -t *.jsonl | head -32); do
  nums=$(tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort -u | tr '\n' ' ')
  echo "$f :: $nums"
done
```
> **Gotcha:** `tail -c` on some of these files panics (uutils `tail`, `Os { code: 22 }`)
> on a file being written concurrently. The loop continues; a panicking file is
> UNREAD, not empty — do not treat its silent row as "this session mentions nothing".

```bash
# (b) commits
python3 scripts/who-owns.py 246
# (c) the filer's own statement of intent — often the most direct answer
grep -n "246" docs/agent_docs/docs024_key_docs_latest/bugfix_239_dispatch_fail_closed/HANDOFF_*.md
```

## 2. Prove the pool overwrite — WITH A CONTROL

No database needed: `sql.Open` builds the pool object without connecting.

```go
db, _ := sql.Open("pgx", "postgres://u:p@127.0.0.1:1/none")
db.SetMaxOpenConns(12)                                  // what agentbase does
fmt.Println(db.Stats().MaxOpenConnections)              // 12
db.SetMaxOpenConns(4)                                   // what the processor does
fmt.Println(db.Stats().MaxOpenConnections)              // 4
db.SetMaxOpenConns(9)                                   // CONTROL
fmt.Println(db.Stats().MaxOpenConnections)              // 9  <- could have come out otherwise
```
> **Gotcha:** without the control line, "it printed 4" is indistinguishable from a
> probe that always prints 4. `db.Stats().MaxOpenConnections` is the only getter —
> there is no `GetMaxOpenConns`.

> **Gotcha:** `(*sql.DB)(nil).SetMaxOpenConns(n)` **panics** (nil deref). Relevant
> because `agentbase` leaves `a.db` nil when `DatabaseURL` is empty.

## 3. Read the LIVE pool configuration (not the repo's)

```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c \
  'echo "MAXCONNS=[$CHASSIS_DB_MAX_OPEN_CONNS] INTAKE=[$CHASSIS_INTAKE_MODE] DBURL=[${DATABASE_URL:+set}]"'
```
> **Gotcha — the repo is NOT the source of truth here.** These keys render nowhere
> in the overlay:
> ```bash
> kubectl kustomize deployments/kustomize/services/agent-chassis/overlays/production/uk_001 | grep CHASSIS_
> # => nothing
> ```
> They live only on the Deployment object. **Ask the pod, never the overlay.**

> **Gotcha:** `DATABASE_URL` is NOT set on the chassis; `CLIENTS_DATABASE_URL` is.
> They are different variables and `p.sqlDB` (from `DATABASE_URL`) is nil in
> production. `p.db` and `p.sqlDB` are **not** interchangeable — a test fixture with
> `sqlDB` set exercises a shape production does not have (239 lane's warning).

## 4. Would a deploy strip live-only env keys? (it would not)

```bash
kubectl -n ai-persona-system get deploy agent-chassis \
  -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration}' \
  | python3 -c "import json,sys; d=json.load(sys.stdin); \
      print([e['name'] for c in d['spec']['template']['spec']['containers'] for e in c.get('env',[]) if 'CHASSIS' in e['name']] or 'NONE')"
```
> **Gotcha:** the intuition "live-only means the next `apply -k` deletes it" is
> **wrong**. `kubectl apply` three-way-merges: absent from both `last-applied` and
> the incoming config ⇒ **preserved**. Check the annotation before raising an alarm.
> (Different from the `kubectl scale` landmine, where replicas ARE in the overlay.)

## 5. Baseline the lookup-fault instrument — and its demand control

```bash
kubectl -n ai-persona-system logs <chassis-pod> --tail=200000 | grep -c "DISPATCH_LOOKUP_RETRYABLE"
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns=NAME:.metadata.name,START:.status.startTime,RESTARTS:.status.containerStatuses[0].restartCount --no-headers
```
```sql
-- the demand control: was there anything for the instrument to observe?
SELECT date_trunc('hour', received_at) AS hr, kind, status, count(*)
FROM chassis_intake_events
WHERE received_at > now() - interval '6 hours'
GROUP BY 1,2,3 ORDER BY 1 DESC;
```
> **Gotcha — a zero here is not a health check.** Report the pod uptime and the
> available line count alongside the count, and the intake volume alongside both.
> At the observed ~1–2 messages/minute a 4-connection pool cannot saturate, so a
> zero is consistent with the pool being fine AND with it freezing under load. It
> discriminates neither. Say so.

## 6. Pool saturation — there is currently NO instrument

`db.Stats()` carries `WaitCount` and `WaitDuration`, and **nothing surfaces them**.
Any claim about contention today is unmeasurable. Do not assert one.

## 7. pgbouncer's side of the seam

```bash
grep -n "pool_mode\|max_client_conn\|default_pool_size\|reserve_pool_size" \
  deployments/kustomize/services/pgbouncer/pgbouncer-configmap.yaml
```
> **Gotcha:** `default_pool_size` is per **user/database pair** and shared by every
> client of `clients_user`/`clients_db`, not per client pod. `max_client_conn` is the
> one the chassis's client-side cap counts against. The configmap's comment reasons
> from "3 chassis replicas × 4 conns" — stale on both numbers.

## 8. Postgres-side connection counts do NOT show the client pool

```sql
SELECT client_addr, application_name, state, count(*) FROM pg_stat_activity
WHERE datname='clients_db' GROUP BY 1,2,3 ORDER BY 4 DESC;
```
> **Gotcha:** every row's `client_addr` is **pgbouncer**, so this measures
> pgbouncer's server-side pool, never the chassis's client-side one. To see client
> connections you need pgbouncer's own `SHOW CLIENTS` / `SHOW POOLS`. And with
> `MaxIdleConns(1)` the pods hold ~1 connection at rest regardless of the cap, so
> an at-rest count cannot distinguish 4 from 12 either way.

## 9. Activating the pgbouncer admin console (owner decision D1, wired 2026-08-12)

**Why:** `SHOW POOLS` / `SHOW CLIENTS` are the ONLY way to see client-side pool queueing
(`cl_waiting`, `maxwait`). §8 explains why `pg_stat_activity` cannot substitute.

**The whole point of this section: the password lives in TWO places that must AGREE**, and
Terraform owns only one of them.

| half | where | managed by |
|---|---|---|
| what a client sends | `personae-platform-secrets.PGBOUNCER_ADMIN_PASSWORD` | **Terraform** — `047-base-configs` (`variables.tf` + `main.tf`, committed) |
| what PgBouncer checks it against | `pgbouncer-userlist` secret → `/etc/pgbouncer/userlist.txt` | **not Terraform** — hand-applied `kustomize/services/pgbouncer/pgbouncer-secret.yaml`, whose repo copy is the literal placeholder `PGBOUNCER_ADMIN_PASSWORD_HERE` |

### Step 1 — Terraform half (done; applies on the next roll)

`pgbouncer_admin_password` is declared in `variables.tf`, wired into the
`personae-platform-secrets` resource in `main.tf`, and a freshly generated 32-char value
is in `terraform.tfvars.secret` (**gitignored and untracked — verified**).

> **Checked before you apply, so you don't have to fear it:** the live secret holds
> **exactly the 7 keys Terraform declares** — zero drift — so an apply ADDS the 8th and
> deletes nothing. Re-check if time has passed, because `kubernetes_secret.data` is
> authoritative and an apply WILL remove any key that has since been added out of band:
> ```bash
> kubectl -n ai-persona-system get secret personae-platform-secrets -o json \
>   | python3 -c "import json,sys; print(sorted(json.load(sys.stdin)['data'].keys()))"
> ```

### Step 2 — the userlist half (NOT done; needs the owner)

Terraform does not manage `pgbouncer-userlist`, so **step 1 alone does not make
`SHOW POOLS` work.** The `pgbouncer_admin` line must carry the same password. Patch only
that line — the `clients_user` and `templates_user` lines are live credentials and must
not be regenerated or reordered.

> **Not attempted by this lane, deliberately.** Reading the live userlist is a credential
> read and was refused by the permission classifier — correctly. So the current value of
> the `pgbouncer_admin` line is **[UNVERIFIED]**: it may still be the literal placeholder,
> or a real value nobody has recorded. Whoever holds the credential should check before
> overwriting, in case something else already authenticates with it.

### Step 3 — make PgBouncer re-read it

The userlist is read at startup. After patching the secret, either `RELOAD;` on the admin
console (chicken-and-egg — only works if admin auth already succeeds) or restart the pod:
`kubectl -n ai-persona-system rollout restart deploy/pgbouncer`. **A restart drops every
pooled connection**, so do it deliberately, not during a build sweep.

### Step 4 — the query this was all for

```bash
kubectl -n ai-persona-system exec <pgbouncer-pod> -- \
  psql -h 127.0.0.1 -p 6432 -U pgbouncer_admin -d pgbouncer -c "SHOW POOLS;"
```
Read `cl_waiting` (clients queued for a server connection) and `maxwait` (how long the
oldest has waited). **Sustained `cl_waiting > 0` or a climbing `maxwait` is the
disconfirming observation for `bugs_open/246`'s pgbouncer risk** — the one measurement
that change has never had. `cl_waiting = 0` with `maxwait = 0` means the client-side
queue is empty, which is the result we expect and have never been able to confirm.

> **Gotcha:** authenticate as `pgbouncer_admin`, not `clients_user`. `clients_user` is in
> the userlist and connects fine to `clients_db`, but the admin console refuses it with
> a bare `FATAL: not allowed` — which reads like a broken connection rather than a
> permissions answer. Measured 2026-08-11.

> **Known wart, recorded rather than hidden:** two owners of one value is the same defect
> class as `bugs_open/246` itself. The structural fix is to have Terraform render the whole
> userlist from the same variables that populate `personae-platform-secrets`, which would
> make the mismatch unrepresentable. It needs the other two passwords moved into that
> resource and was outside what decision D1 authorised.
