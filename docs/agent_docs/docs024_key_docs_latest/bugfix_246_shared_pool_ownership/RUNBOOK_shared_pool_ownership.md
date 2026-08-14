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

### §9 CORRECTED 2026-08-13 — the Terraform half landed, the console still fails, and the plan was wrong

**State:** `PGBOUNCER_ADMIN_PASSWORD` **is now in `personae-platform-secrets`** (8 keys —
the apply happened). The console still refuses:

```
psql -h 127.0.0.1 -p 6432 -U pgbouncer_admin -d pgbouncer -c "SHOW POOLS;"
FATAL:  password authentication failed
```

**That error is informative and worth distinguishing from the earlier one.** As
`clients_user` the console answers `FATAL: not allowed` — authentication *succeeded* and
the admin check refused. As `pgbouncer_admin` it answers `password authentication failed` —
the user is an admin, the password is wrong. So the roster is right and only the value is
out of step.

**Why:** pgbouncer's userlist stores **plaintext** (no `auth_query`/`auth_user` is
configured — `auth_type = md5`, `auth_file` only, so the userlist is the sole authority).
The existing `pgbouncer_admin` entry is **20 characters**; the value Terraform generated is
**32**. A length comparison is enough — they cannot be the same string.

> **MY ERROR, corrected here rather than quietly.** I wired D1 by **generating a fresh
> password** for an account that **already had a working one**. I did flag the existing
> value as `[UNVERIFIED]` and wrote "check before overwriting" — the right instinct — and
> then generated anyway, which guaranteed a mismatch. The console was never going to work
> after step 1 alone, and not for the reason §9 originally gave.
>
> **A second, worse error, retracted.** I first compared the userlist's `clients_user`
> entry against `CLIENTS_DB_PASSWORD` **as whole lines** — including quoting and
> whitespace — got "DIFFER", and reported that a rewrite would have broken fleet-wide
> auth. **That conclusion was not supported.** Whole-line inequality does not imply value
> inequality, and the decisive evidence points the other way: the chassis authenticates
> through pgbouncer as `clients_user` with `CLIENTS_DB_PASSWORD` continuously, and with
> no `auth_query` the userlist plaintext is the only thing it can be matching. **The
> values almost certainly agree and my alarm was an artefact of the comparison.** The
> lesson is the one this lane keeps relearning: *compare the VALUE you mean, never the
> line that contains it* — the same shape as the whole-line/prefix traps already in
> `LANDMINES.md`.

### The corrected plan: RECORD the existing password, do not impose a new one

**Option B (recommended).** Put the **existing 20-char userlist value** into
`terraform.tfvars.secret` as `pgbouncer_admin_password`, replacing the generated one.

- **No pgbouncer restart** — the userlist is untouched, and pgbouncer already accepts it.
- **No risk to `clients_user`/`templates_user`** — their lines are never rewritten.
- Reversible, and it makes the platform secret *describe* reality rather than contest it.
- Needs one credential read + one tfvars edit, both of which need the credential holder:
  this session's attempts to read the userlist value were **refused by the permission
  classifier**, correctly, and were not worked around.

**Option A (not recommended).** Write Terraform's 32-char value into the `pgbouncer-userlist`
secret and restart pgbouncer. Costs a restart that **drops every pooled connection
fleet-wide**, and puts the other two users' lines at risk in the rewrite for no gain.

> **Note what this means for the original D1 framing:** a working admin password has
> existed in the userlist all along, so `SHOW POOLS` was never blocked by the *absence* of
> a credential — only by its absence **from any sanctioned place a session could find it**.
> The Terraform wiring still earns its keep (the credential now has a recorded home), but
> the honest description is "record the existing secret", not "create the missing one".

### §9 RESOLVED 2026-08-14 — the password is recorded, the console works, and D3 has its answer

**Done:** `terraform.tfvars.secret` now holds the **existing** 20-character userlist value
instead of the generated 32-character one. Verified without printing it: exactly one
`pgbouncer_admin_password` key in the file, written length 20, byte-equal to the userlist
value — and **proven to authenticate before it was written**, which is the order that
matters:

```
kubectl -n ai-persona-system exec <pgbouncer-pod> -- \
  env PGPASSWORD="$ADMIN" psql -h 127.0.0.1 -p 6432 -U pgbouncer_admin -d pgbouncer -c "SHOW POOLS;"
```
returned rows. **No pgbouncer restart, no userlist rewrite, nothing at risk for
`clients_user`/`templates_user`.**

**⏳ Still owed by the machine, not by a person:** `047-base-configs` applies as part of
`make release`, so until the next release `personae-platform-secrets.PGBOUNCER_ADMIN_PASSWORD`
still carries the old generated string. **It self-corrects on the next roll.** The console is
reachable today using the userlist value directly.

**Pre-apply drift check — run it, and parse it properly.** All 9 live keys are declared, so
the apply deletes nothing.

> **⚠ My first run of this check produced a FALSE POSITIVE that would have read as an
> incident.** I isolated the `personae_platform_secrets` block by splitting the file on the
> string `'resource '` — which occurs **inside a comment in that very block** (*"this
> resource reconciles the whole secret…"*), truncating the block before `SITE_FACTS_TOKEN`
> and reporting that an apply would delete it. That is the key whose deletion **took the
> site-facts relay down on 2026-08-13**, so the false alarm landed on exactly the sore
> spot. **Do not parse HCL by splitting on keywords.** Ask the simple question instead —
> is each live key assigned anywhere in the file?
> ```bash
> kubectl -n ai-persona-system get secret personae-platform-secrets -o json | python3 -c "
> import json,sys,re
> live=sorted(json.load(sys.stdin).get('data',{}).keys())
> src=open('deployments/terraform/environments/production/uk001/047-base-configs/main.tf').read()
> missing=[k for k in live if not re.search(r'^\s*'+re.escape(k)+r'\s*=', src, re.M)]
> print('undeclared (an apply DELETES these):', missing or 'NONE')"
> ```

## 10. THE MEASUREMENT 246 NEVER HAD — `SHOW POOLS`, 2026-08-14

The disconfirming observation for `bugs_closed/246`'s pgbouncer risk, taken at last:

| database | user | cl_active | **cl_waiting** | sv_active | sv_idle | **maxwait** | pool_mode |
|---|---|---|---|---|---|---|---|
| clients_db | clients_user | 17 | **0** | 3 | 2 | **0** | transaction |
| pgbouncer | pgbouncer | 1 | 0 | 0 | 0 | 0 | statement |

**The risk is not materialising.** `cl_waiting = 0` and `maxwait = 0` — no client is queued
for a server connection, and none has waited. **5 server connections in use of
`default_pool_size = 15`**, with 17 client connections multiplexed onto them, which is
transaction pooling doing exactly what the 246 submission argued it would.

**D3 answered: `default_pool_size = 15` does not need raising.** The chassis's cap going
4 → 12 has not pushed pgbouncer into queueing.

> **This is ONE SAMPLE, and the honest limits are:** `cl_waiting` is instantaneous, and
> pgbouncer's `maxwait` is *the current longest wait, not a high-water mark* — it returns
> to 0 when the waiting client is served, so a zero here cannot rule out queueing between
> samples. It was also taken at ~17 client connections; a burst is exactly when the answer
> could differ. To make this a real result rather than a snapshot, sample repeatedly
> (`SHOW POOLS` on a loop, or `SHOW STATS`) across a busy period. **Do not quote this table
> as "pgbouncer is fine under load" — it says "pgbouncer was not queueing at this moment,
> at this load".**
