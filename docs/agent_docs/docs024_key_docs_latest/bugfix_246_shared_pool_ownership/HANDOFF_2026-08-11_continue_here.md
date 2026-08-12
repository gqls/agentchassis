# HANDOFF — bugfix 246 lane, 2026-08-11 evening. COLD-START READ THIS FIRST.

**Lane status: 246 is DONE, FIXED, COUNCIL-APPROVED and LIVE on `v1.0.1288`.**
Nothing is half-finished. What remains is **five owner decisions** and **three
follow-ups nobody has started**. If you are picking this up, you are starting the
follow-ups, not rescuing the fix.

---

## 1. What shipped, and how to re-prove it in 20 seconds

`NewMessageProcessor` was calling `SetMaxOpenConns(4)` on the `*sql.DB` it was
**passed**. A `*sql.DB` is a pool object, so that re-sized agentbase's pool instead of
making its own, silently discarding `CHASSIS_DB_MAX_OPEN_CONNS=12`. Deleted. The rule
now stated in the code: **you size what you OPEN, never what you were GIVEN.** The same
rule was applied to the pool that constructor opens *itself* from `DATABASE_URL`, which
had its error discarded and was never sized (Go's zero value = **unlimited**).

| | |
|---|---|
| fix commit | `039cfce84` |
| council follow-up commit | `6ba3fca28` |
| council correlation | `c94d73ac-2a15-40cb-98a9-1185a2b7435a` — **APPROVED round 1** |
| shipped in | `v1.0.1288`, revision `bb534864249117003ac758e50adc0df9176ef370` |

```bash
docker image inspect docker.io/aqls/agent-chassis:v1.0.1288 \
  --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
git merge-base --is-ancestor 039cfce84 <that>          # exit 0 = shipped
git merge-base --is-ancestor $(git rev-list <that>..HEAD | head -1) <that>  # MUST exit 1
```

**Do NOT verify this by grepping pod logs for `build provenance`.** Measured twice today:
the pods started 17:13:36Z and their logs' first available line was already 17:43:51Z —
30 minutes of startup log gone inside half an hour. An empty grep means *rotated*, not
*unstamped*. And do not probe `/proc/1/exe` for your own sha: the binary carries ONE
commit (the build point), not its ancestors, so a binary that contains your fix reports
your sha as **absent**.

---

## 2. THE FIVE DECISIONS (owner)

### D1 — the pgbouncer admin credential. **Blocks the only measurement that matters.**

The one real risk of this change is that the chassis, now able to hold 12 client
connections per pod instead of a silent 4, queues at pgbouncer. The disconfirming
observation is pgbouncer's own `SHOW POOLS` — `cl_waiting` sustained > 0, or `maxwait`
climbing. **No session can run it.** `admin_users`/`stats_users` is `pgbouncer_admin`
(`pgbouncer-configmap.yaml:73-74`), that user exists in `/etc/pgbouncer/userlist.txt`,
but **its password is not in `personae-platform-secrets`** — I enumerated the secret's
keys and there is no pgbouncer entry.

`pg_stat_activity` is **not** a substitute: every row's `client_addr` is pgbouncer, so
Postgres sees the server pool and never the client queue.

- **(a)** Add `PGBOUNCER_ADMIN_PASSWORD` to `personae-platform-secrets` → any session can
  check this and every future pool question. Cheapest, and it is a read-only stats user.
- **(b)** You run `SHOW POOLS` when asked. Works, but a measurement that needs a human is
  one that will not be taken.
- **(c)** Accept it unmeasured. Defensible — transaction pooling makes queueing unlikely —
  but then nobody can ever confirm it.

### D2 — build the pool instrumentation, or accept this class stays invisible

`db.Stats()` carries `WaitCount` and `WaitDuration`. **Nothing surfaces them.** This is
precisely why 246 sat inert for weeks with nobody noticing: there is no instrument that
can see a connection pool under strain. Until one exists, every future pool question gets
the same answer I had to give — *"unmeasurable"*.

Scope is small (export the two counters per pod). I deliberately kept it **out** of the
246 change so the blast-radius argument stayed simple enough for a reviewer to check.
It would be **registrable** in the concept register, not ratchetable.

### D3 — `default_pool_size = 15` at pgbouncer

Measured this evening: **10 of 15** server connections in use, 2 active / 8 idle. Earlier
today it was 6 of 15. **That rise is confounded** — the client cap went 4→12 *and* intake
load roughly tripled (60–100/hr → 182–228/hr) in the same window, so it is not evidence
about this fix either way. What is true: less headroom than before, and the configmap's
own sizing comment reasons from *"3 chassis replicas × 4 conns"*, which is stale on both
numbers (it is 2 replicas, intending 12).

Raise it, leave it and watch, or leave it and fix only the misleading comment. **D1 gates
this** — without `SHOW POOLS` you would be tuning blind.

### D4 — the repo does not record how the fleet is configured

`CHASSIS_DB_MAX_OPEN_CONNS=12`, `CHASSIS_INTAKE_MODE=worker_pool_all` and
`CHASSIS_RESPONSES_START_AT=latest` render **nowhere** in
`deployments/kustomize/services/agent-chassis/overlays/production/uk_001`. They exist only
on the live Deployment object.

**They are safe** — absent from `last-applied-configuration`, so `apply -k` three-way-merges
and preserves them (I checked; the alarming reading is the intuitive one and it is wrong).
The cost is not risk, it is truth: you cannot learn the fleet's real configuration by
reading the repo, and the neighbouring pgbouncer comment already went stale because of it.
Adding them to the overlay makes the repo honest; **note that `CHASSIS_INTAKE_MODE` must
NOT go into `personae-prod-config`** (`intake.go:20` is emphatic — spawned pods inherit it
and must not become intake workers).

### D5 — collapse `p.sqlDB` into `p.db`

`p.sqlDB` comes from `DATABASE_URL`, which is **unset on chassis pods**, so it is nil in
production — and all eight readers already carry a `db := p.db; if db == nil { db = p.sqlDB }`
fallback. It is a second answer to "which database handle?" that only ever resolves one
way. Deleting it is the honest end state; it is a separate change with its own review
because it touches eight sites. Related to the `bugs_open/247` dead-code family (that bug
is closed and live as of `v1.0.1288`, by another lane).

---

## 3. What I would do next, in order

1. **D1**, because D3 is unanswerable without it and it is a one-line secret addition.
2. **D2**, because it converts "unmeasurable" into "measured" for every future pool question
   and is the structural lesson of this whole bug.
3. **D4**, cheap and purely a truthfulness fix.
4. **D5**, then **D3** once D1 gives you numbers.

---

## 4. Traps this lane hit, so you do not

- **A negative control must be chosen by asking the repo, not by assuming your own commit
  is newest.** I used my own summary commit as the "must be absent" control; the build
  stamp was *later* than it, so the control reported "present" and looked like a broken
  test. On a shared branch your work is not the newest thing within minutes. Use
  `git rev-list <stamp>..HEAD | head -1`.
- **A sha is generated output, never retyped.** I published `039fcce84` for `039cfce84`.
  Caught by `git cat-file -t` and, independently, by the 239 lane. Use `git rev-parse HEAD`.
- **Grep `LANDMINES.md` for the nouns in your VERIFICATION, not only your CHANGE.** I greped
  every symbol I was editing and none of the service I was going to verify against, and
  wrote a check documented as inoperative. The council caught it.
- **A zero from an instrument needs a demand control.** `DISPATCH_LOOKUP_RETRYABLE` read 0
  before the fix, at 1–2 messages/minute — a load that cannot saturate a 4-connection pool.
  It discriminated nothing and must not be quoted as health.
- **The 090 diagnosis loop cannot read a Kubernetes env var** (repo + `clients_db` only). It
  ran 5 iterations on this bug, got `(0 rows)` twice from `agent_definitions.env_vars`, and
  emitted **no verdict**. Had it emitted one on the evidence it could reach, it would have
  said the override was harmless — the opposite of the truth. Now in `LANDMINES.md`.
- **The bug file's own code quote was wrong**, in the direction that inflates the bug.
  Re-read cited lines at HEAD even when the file is a day old.

---

## 5. The honest limit of what was proven

There is **no behavioural witness** for this change and there never could be: nothing in
the platform reports a pool's size or its wait counters. The evidence is (a) a
mutation-proven unit test for the mechanism and (b) image-label ancestry for the shipment.
**Anyone who later writes "the pool was observed at 12" is describing something the
platform cannot show.** That gap is D2.

## 6. Where everything is

- Bug: `bugs_open/246_HANDOFF_2026-08-10_processor_silently_reshrinks_the_shared_db_pool_to_four_connections.md`
  (kept in `bugs_open/` per owner ruling 2026-08-06; final section is the post-roll proof)
- Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_246_shared_pool_ownership/`
  — `PLAN_2026-08-11`, `RUNBOOK`, `NOTES`, `README_where_we_are`, `SUMMARY_2026-08-11`
- Landmine: `LANDMINES.md`, "The 090 diagnosis loop cannot read a Kubernetes env var"
- Missteps: `WRONG_CALLS.md`, two 2026-08-11 sections from this lane (5 rows)
- Ratchet: `docs026_concept_register/102_coverage_ratchet.txt`, `bugfix_246_shared_pool_ownership`
  — **names D2 and D5 as the follow-ups that would be REGISTRABLE, not ratchetable**

---

## UPDATE 2026-08-12 — D1 wired; D2 collided; **D5 was not a tidy-up and is now `bugs_open/259`**

### D1 — half done, half needs the credential holder
Committed `aee444a35`. `pgbouncer_admin_password` is declared in
`047-base-configs/variables.tf`, wired into `personae-platform-secrets` in `main.tf`, and
its value is in `terraform.tfvars.secret` (gitignored, **verified untracked**).
**Checked so nobody has to fear the apply:** the live secret holds exactly the 7 keys
Terraform declares — zero drift — so an apply adds the 8th and deletes nothing
(`kubernetes_secret.data` is authoritative, so re-check if time passes).
**Still needed:** the `pgbouncer-userlist` secret is NOT Terraform-managed, so step 1
alone does not make `SHOW POOLS` work. Its `pgbouncer_admin` line must carry the same
password. Reading it is a credential read and was refused by the permission classifier,
so its current value is **[UNVERIFIED]**. Runbook §9 has the whole sequence.

### D2 — STOPPED, not abandoned: the 239 lane is already in it
Greping live transcripts before starting showed that session hitting `db.Stats`,
`WaitCount`, `WaitDuration` and `bugs_open/246` heavily, active within the hour, nothing
committed. Messaged them offering three ways to split it and handed over what I hold
(the pool that matters is the one **agentbase** opens; `p.sqlDB` is nil so instrumenting
it measures nothing; the counters are **cumulative since process start**, so a raw gauge
misleads across a roll; my pre-roll baselines are against the OLD 4-connection pool and
are not like-for-like). **Awaiting their answer — do not start a second design.**

### D5 — REDEFINED. It is a defect, not a cleanup. See `bugs_open/259`.

D5 was written here as "collapse `p.sqlDB` into `p.db`, every reader already falls back".
**That description was wrong**, and I am correcting it rather than editing it away: not
every reader falls back. **Three sites GUARD on `p.sqlDB != nil`**, and since the handle
is nil on every chassis pod, none has ever executed in production.

- **A** (`~:351`) child-workflow completion early-return — **[UNASSESSED]**.
- **B** (`~:582`) the workflow's final result — **always the literal placeholder
  `{"status":"completed"}` instead of the orchestration's `CollectedData`.** A
  control-flow certainty. Downstream effect `[UNMEASURED]`.
- **C** (`~:1486`) the entire `bugs_open/003` two-phase dedup claim — **redundant, NOT a
  hole**: `agentbase` does the same claim on a live handle, evidenced by 449 rows / 82
  distinct writers in one hour in `processed_messages` with the lifecycle visible.

**The trap for whoever takes it:** mechanically applying 239's
`db := p.db; if db == nil { db = p.sqlDB }` to all three READS as consistency and is the
dangerous option — on C it switches on a second dedup layer that has never run, and on
A and B it changes live behaviour. **246's "this is a no-op everywhere" safety argument
does NOT transfer to 259.** That was the whole basis of 246's confidence and it is absent
here.

Also worth carrying: the log markers (`Duplicate message ignored`, `DEDUPE_CLAIM_LOST`)
read **zero** and that means nothing — they fire only when a duplicate occurs. The table
is the instrument. I tried the blind check first and rejected it; do not repeat it.

### Standing verification note
The chassis has rolled twice more since 246 shipped. Re-verified on **`v1.0.1290`**
(revision `fa078ab3d`): `039cfce84` still an ancestor, negative control holds. The method
is in §1 of this file — image label, never the pod log.

---

## UPDATE 2026-08-12 (later) — **LANE CLOSED.** 246 moved to `bugs_closed/`; D2 and 259 handed over

### The bug is CLOSED and MOVED — the bar changed under us
`bugs_closed/246_HANDOFF_2026-08-10_...`. **The owner restored the fixed-AND-live bar on
2026-08-12**: *"if it is fixed and live it should be moved"* (`2aa3014a3`), superseding the
08-06 direction that kept finished bugs in `bugs_open/`. Relayed by the 239 lane and
**verified first-hand at the commit before acting** — a relayed owner ruling that
contradicts a recorded one is not something to take on trust. Moved with **both paths named
on the commit** and verified at HEAD (`git ls-tree` returns exactly one line), per the
`git mv` pathspec landmine. The auto-memory entry still named after the old direction has
had its `description` corrected; its index line was already fixed by another session.

### D4 — DONE (`871c24665`)
The three `CHASSIS_*` keys are now in the production overlay. **The safety argument is the
diff, not care:** I rendered the overlay and diffed its `CHASSIS_*` env against the live
object — **identical**, so an apply is a no-op for these three. Do the same if you ever
touch them; a typo here CHANGES live configuration rather than documenting it.

### D2 — HANDED OVER to the `bugfix_239` lane
They have taken the `db.Stats()` instrumentation on their owner's direction. All five of
my constraints are recorded verbatim in their cold-start
(`bugfix_239_dispatch_fail_closed/HANDOFF_2026-08-12_continue_here.md`), including the
register-not-ratchet obligation. **Do not start a parallel design.**

### 259 — HANDED OVER to the same lane, as 239's continuation
They re-verified all three sites at source before accepting and confirmed the control-flow
claim on B. **Note the number is ambiguous** — an unrelated GPU-provisioning `259` was filed
the same day; resolve by the `three_processor_paths...` slug.

### Still open, and now the only things this lane leaves behind
- **D1's second half** — the `pgbouncer-userlist` secret's `pgbouncer_admin` line still
  needs the credential holder. Terraform half is done (`aee444a35`); Runbook §9 has the
  sequence. Until then `SHOW POOLS` is unreachable and 246's one real risk stays unmeasured.
- **D3** — `default_pool_size = 15`, gated on D1.

### Worth carrying into any future submission from this lane
The 239 lane relays that **RFC_023 is RULED: the architecture seat's trigger is BEHAVIOUR
(the consumer's success path), not diff/package count**, with a standing rider on code
bloat. That changes how a submission should argue scope — recorded here second-hand, so
read the RFC before relying on it.
