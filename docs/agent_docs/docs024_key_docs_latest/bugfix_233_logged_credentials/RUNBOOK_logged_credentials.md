# RUNBOOK — bugfix 233

Commands that were hard to get right, with the gotcha attached.

## Is anything still emitting the credential?

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=render-audit-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq "NewS3Client"                 /proc/1/exe   # CONTROL
kubectl -n ai-persona-system exec $POD -- grep -aq "access_key_present"          /proc/1/exe   # FIX
kubectl -n ai-persona-system exec $POD -- grep -aq "B2_APPLICATION_KEY from env" /proc/1/exe   # LEAK
```
⚠ **Run all three or none.** A broken `grep -aq` returns ABSENT for everything, which reads exactly
like "still leaking".
⚠ **Never `strings`** — absent from the debian-slim images, and behind the customary `2>/dev/null`
its failure is indistinguishable from "not found". This bug file's *original* recipe used `strings`;
it is superseded.

## Fleet census — and the control that stops the zero being vacuous

```bash
kubectl -n ai-persona-system get deploy,cronjobs -o json | python3 -c "
import json,sys,re
d=json.load(sys.stdin); v=[]
for it in d['items']:
    sp=it['spec'].get('template') or it['spec']['jobTemplate']['spec']['template']
    for c in sp['spec']['containers']:
        m=re.search(r':v1\.0\.(\d+)', c['image'])
        if m: v.append(int(m.group(1)))
print('parsed:',len(v),'min:',min(v),'max:',max(v),'pre-fix(<1274):',sum(1 for x in v if x<1274))"
```
⚠ **Print min/max, not just the count of old images.** "0 older than v1.0.1274" reads identically
whether the fleet is clean or the regex matched nothing. 2026-09-04: 36 parsed, min = max = 1360.
⚠ Third-party images (`postgres`, `ollama`, `pgbouncer`, `wireguard`, `kubectl`) do not parse. Record
them as **unjudged**; do not fold them into the zero.

## Counting a credential in a log WITHOUT reading it

```bash
kubectl -n ai-persona-system logs <pod> --tail=100000 | grep -c "B2_APPLICATION_KEY from env"
kubectl -n ai-persona-system logs <pod> --tail=100000 | wc -l          # <-- the control
```
⚠ **Always print the total too.** A `0` out of `0` total lines is **vacuous** — the pre-fix pods'
logs have expired, so their zero means "cannot be answered", not "clean".
⚠ **Never print a matching line.** Owner ruling 2026-08-23: never read a key into the session.

## Is a secret sitting inline in any live agent config?

```sql
-- names only, never values
SELECT DISTINCT m[1] AS key_name, count(*) OVER (PARTITION BY m[1]) AS hits
FROM agent_definitions,
LATERAL regexp_matches(default_config::text,
  '"([a-z_]*(?:password|secret|api_key|apikey|token|credential|access_key)[a-z_]*)"\s*:\s*"[^"]{12,}"','gi') AS m
WHERE is_active AND deleted_at IS NULL ORDER BY 2 DESC;
```
⚠ **Extract the NAMES; never report the raw match count.** The count was **81 agents** and every one
was a false positive — `api_key_env_var` ×160, `secret_key_env_var` ×2, `access_key_env_var` ×2, all
holding environment variable *names*. A credential-shaped key name is not a credential.
⚠ Pair it with an existence control (e.g. `"prompt_template"` → 67 agents) so a zero cannot come from
a broken pattern.

## Has the rotation happened?

**You cannot answer this from a session.** `kubectl get secret … -o jsonpath='{.metadata.creationTimestamp}'`
returns `2025-08-02`, but ⚠ **`creationTimestamp` survives an in-place `kubectl apply`**, so it cannot
distinguish "never rotated" from "rotated in place". Reading the value is forbidden. **Ask the owner.**
