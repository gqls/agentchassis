# HANDOFF — `bugs_open/233`, credentials logged in plaintext

**Written 2026-09-04 ~08:2xZ.** Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_233_logged_credentials/`.
Read this, then `NOTES_logged_credentials.md` for the evidence and the missteps.

---

## 1. The state, in one paragraph

Two log lines emitted live credential values at INFO: the shared S3 constructor logged the **B2
application key pair** on every client build, and the spawn path logged **`CLIENTS_DB_PASSWORD`** on
every dynamic-agent spawn. **The code fix shipped 2026-08-09, is council-APPROVED, and is live and
re-verified today on `v1.0.1360`.** The class is now fully bounded — including the one gap the
original fix honestly declined to claim. **Exactly one thing remains, and it is not code: the
credentials must be presumed disclosed and rotated.** That is an owner decision and it is the only
reason this file is still open.

## 2. What is left before it can close — the whole list

| # | item | owner | state |
|---|---|---|---|
| 1 | **Rotate the B2 application key pair** (`personae-default-secrets`, plus the GitHub-secrets copy the B2 CLI uses) | **the owner** | ⛔ **OUTSTANDING / [UNVERIFIED]** |
| 2 | **Rotate `CLIENTS_DB_PASSWORD`** (touches every service's DB wiring) | **the owner** | ⛔ **OUTSTANDING / [UNVERIFIED]** |
| 3 | Code fix live fleet-wide | this lane | ✅ verified `v1.0.1360`, 2026-09-04 |
| 4 | Class bounded incl. the `zap.Any` gap | this lane | ✅ swept clean, 2026-09-04 |
| 5 | Rotation-ordering blocker discharged | this lane | ✅ 2026-09-03 |

**Close the file when the owner confirms 1 and 2, or rules them unnecessary. Nothing else is owed.**

⚠ **Do not close it on the strength of the code fix.** The code stopped the emission; it did not
un-emit ten months of it.

## 3. Why the rotation is not optional-looking

`git log -S 'B2_APPLICATION_KEY from env' -- platform/storage/s3.go` dates the leak to `9260b86ed`,
**2025-10-28**. Both credentials have therefore been written in plaintext into pod logs for **over
ten months**, readable by anything with `kubectl logs` on `ai-persona-system` — and by any log
aggregator, if one was ever attached. The original fix deferred rotation to the owner deliberately
and said so.

⚠ **The ordering constraint that later blocked it is DISCHARGED.** This file used to instruct: roll
`render-audit-adapter` to ≥`v1.0.1274` *before* rotating, or the rotation writes the **new** key into
that pod's log. That roll has happened (it runs `v1.0.1360`, leak string absent). **A rotation today
cannot be re-leaked by any pod in the fleet**, because no image contains the emitting code.

⚠ **That instruction was stale for 15 days and nobody knew** — its blocker (`bugs_open/237`) closed
2026-08-19. See NOTES (a); it is the second recorded instance of a closed blocker still being obeyed,
and the cost was inaction on a live credential.

## 4. Why this session could not just check whether they were rotated

- The secret's `creationTimestamp` is `2025-08-02` and there are no rotation annotations — but
  ⚠ **`creationTimestamp` survives an in-place `kubectl apply`**, so it cannot distinguish "never
  rotated" from "rotated in place". It is not evidence either way.
- Reading a key value is forbidden (**owner ruling 2026-08-23** — never read a key into a session;
  probe from the pod).

So the answer has to come from the owner. There is no query that settles it from here.

## 5. Reproduce the "is it still leaking?" check in one paste

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=render-audit-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- grep -aq "NewS3Client"                 /proc/1/exe  # CONTROL: must be PRESENT
kubectl -n ai-persona-system exec $POD -- grep -aq "access_key_present"          /proc/1/exe  # FIX:     PRESENT <=> fixed
kubectl -n ai-persona-system exec $POD -- grep -aq "B2_APPLICATION_KEY from env" /proc/1/exe  # LEAK:    PRESENT <=> leaking
```
**Run all three.** A broken `grep -aq` returns ABSENT for everything and reads as "still leaking".
⚠ **Do not use `strings`** — absent from the debian-slim images, and behind the customary
`2>/dev/null` its failure is indistinguishable from "not found". The file's original recipe used it;
that recipe is superseded.

Fleet census, with the control that stops its zero being vacuous:

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
2026-09-04: **36 parsed, min = max = 1360, 0 pre-fix.** ⚠ Without min/max, "0 pre-fix" reads
identically to a regex that matched nothing.

## 6. Traps this lane paid for

1. **A stale banner causes INACTION, not confusion.** A "wait for X" line makes the correct action
   look premature, so it never happens and nothing surfaces as wrong. Whoever closes a bug owes
   `grep -rln "bugs_open/<number>" bugs_open/ docs/` — and must read each hit for an **instruction**,
   not a mention (3 hits on `237`, only 1 a real dependency).
2. **A zero from an empty log is vacuous.** The pre-fix pods return 0 leak lines out of **0 total**.
   That is "cannot be answered", not "clean".
3. **A credential-shaped KEY NAME is not a credential.** 81 agents matched a secret-shaped key with a
   long value; extracting the **names** showed all 81 were `*_env_var` keys holding variable names.
   Second sighting of that class inside this one bug. Extract names; never trust the count.
4. **`strings` is retired** on these images — see §5.

## 7. Files and trail

- `platform/storage/s3.go` (values → presence booleans, plus a pre-existing mislabel fixed) and
  `platform/orchestration/actions/spawn_actions.go` (the `CLIENTS_DB_PASSWORD` field → a boolean).
- `bugs_open/233_HANDOFF_2026-08-09_…md` — the bug, with the 2026-09-03 and 2026-09-04 updates.
- Council: `Council-Submitted: 7490388d-c945-42c0-b3c4-c452741a10cd` on `43c1801d6` — **APPROVED,
  round 1, all reviewers**. ⚠ Read a verdict by **your** correlation from `diagnosis_artifacts`,
  never `doc_notes … ORDER BY created_at DESC LIMIT 1`; that returns whoever finished last, which on
  this tree is usually somebody else (documented in `LANDMINES.md`, and hit again 2026-09-03).
- Related and CLOSED: `bugs_open/237` (the release-coverage gap that pinned `render-audit-adapter`).

## 8. Not this lane's, but adjacent and unfinished

`bugs_open/329` (orchestration takeover claim) was the other bug this session worked, and it has its
own handoff at
`docs/agent_docs/docs024_key_docs_latest/bugfix_329_takeover_claim/HANDOFF_2026-09-03_continue_here.md`.
Its fix is live on `v1.0.1360` (re-probed today) and **its durable needle is still 0** against
**6,890** orchestrations in 24 h — i.e. live and not yet exercised, which is the expected reading, not
a fault. `bugs_open/461` carries its named residual.
