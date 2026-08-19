# RFC_040 — a migration cannot ask the binary what it can do, so every config-ahead-of-binary interlock on this estate is enforced by prose

**Status: DRAFT** · raised 2026-08-19 by the `bugfix_299_cta_dials_phone` lane
· motivating cases `bugs_open/299` (slug
`home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead`) and
`bugs_open/312` · prompted by a **medium-severity council objection** on this
lane's own approved submission (corr `1f1fecc9`, `bug_historian` seat).

> **Refer to 299 BY SLUG.** The number is ambiguous — `bugs_closed/299` is the
> unrelated skipped-render-audit case.

---

## 1. Problem + evidence

### 1.1 The structural fact everything below follows from

CLAUDE.md states it in one line: **"Go changes are inert until an image is rebuilt
and rolled; DB config is live immediately."** So for every change whose config half
names a behaviour in its code half, there is a window in which the config is ahead
of the binary. The estate has two names for what happens in that window, and they
are different failures:

- **fail-fast** — the binary rejects the config it cannot honour. Live example, and
  it is worse than it sounds: `discovery_checks.go:198-216`, an unregistered check
  name in a `checks` array returns
  `fmt.Errorf("discovery check %q is not registered …")` unless the step sets
  `allow_unregistered_checks: true` (default false). The `return` at `:208` happens
  **before** `tx.Commit()` at `:284`, inside `defer tx.Rollback()` — so one bad name
  discards **every earlier check's findings in the same run**, not just its own.
- **silent no-op** — the binary ignores the config it does not understand. This is
  the "380 trap" already in `LANDMINES.md`: a config key naming behaviour the binary
  lacks *reads as applied and does nothing*, which is worse than either state alone
  because the applied-ness is visible and the inertness is not. The memory index
  carries the same class three separate ways ("a dead config key looks like a live
  one", "grep the config key before calling it a win", "the error echoing the OLD
  number means the key is UNREAD, not too small").

### 1.2 The class is large, and it is entirely unenforced

Measured 2026-08-19 over `docs/agent_docs/sql_for_agents/` (560 non-sidecar files):

| | count |
|---|---|
| migrations whose header asserts a binary precondition in prose (`pod-verif`, `until the image`, `merge-base --is-ancestor`, `image carrying commit`) | **32** |
| `_HOLD`-suffixed files (the filename convention for "do not apply yet") | **16** |
| migrations that mechanically **verify** such a precondition | **0** |

Zero, and it cannot be otherwise: SQL running in `postgres-clients-0` has no way to
reach a pod. The `_HOLD` suffix is honoured by `SIDECAR_RE` in
`scripts/migration/run-migrations.sh:65`, which keeps a held file out of the pending
list — a real and useful control over the *runner*, but it constrains nothing about
**when a human renames the file**, which is the actual decision. The council seat put
it exactly: *"a recorded user decision with no enforcement point is decorative"* —
raised against this lane's fleet-wide interlock, which is documented in three places
(the migration header, `bugs_open/312`, `LANDMINES.md`) and enforced in none.

### 1.3 The prescribed human check is frequently IMPOSSIBLE — measured, first-hand, today

This is the part that turns an ordinary "we should automate this" into an RFC. The
procedure CLAUDE.md prescribes is: read the service's own `build provenance` log line,
then `git merge-base --is-ancestor <your-commit> <the stamp>`. Attempting exactly that
today, for this lane's own hold:

1. **The stamp had scrolled.** The fleet rolled to `v1.0.1316` at 17:13Z; the pods'
   earliest retained log line is 20:08Z. CLAUDE.md already warns the line is a startup
   line and that an empty result means "not in range", not "unstamped" — so the
   documented primary check simply had no answer available, three hours after a roll.
2. **The documented fallback does not answer this question either.** The binary carries
   `buildinfo.GitCommit` — **one** string (`pkg/buildinfo/buildinfo.go`), not its
   ancestry. So `grep -aq <my-commit> /proc/1/exe` returns *absent* on a binary that
   certainly contains that commit's code. I ran precisely this and got three confident
   "absent" results for commits I had already proven were ancestors of the previous
   stamp. A session that stopped there would have concluded its fix had not shipped.
3. **The remaining fallback is forbidden.** `LANDMINES.md` rules out a discovery grep
   for "some 40-hex string" (it matches Go's internal digit table and returns the same
   wrong answer on every service).

> **CORRECTED 2026-08-19, before this RFC was ratified — I wrote the above as a fresh
> discovery and it is at least the SECOND occurrence.** The `bugs_open/215` lane hit the
> same wall on 2026-08-11, chassis `v1.0.1288`, and wrote it up: probing `/proc/1/exe` for
> four of its own commit shas **plus a fabricated control** returned absent for all five —
> "a check with a working negative control and **no positive control**, i.e. no information
> at all" — and on the same day its `build provenance` log route "failed in both of its
> documented ways at once" (1.4MB of council payloads quoting the phrase, and the startup
> line already rotated out of both pods).
>
> **This changes the argument's weight, and in the RFC's favour.** A one-off is a session
> being unlucky; a recurrence across two independent lanes, eight days apart, on different
> chassis versions, is the mechanism failing by design. It also means the estate has already
> paid for this twice and *still* has no enforceable answer — the 215 lane's remedy (probe a
> literal your last commit ADDED, with a one-letter near-miss) is a better marker, but it is
> still a marker: it needs a human to choose a good literal per change, it cannot be asserted
> by a migration, and it silently stops working for a change that adds no literal. §1.4's
> capability probe generalises it; §2 makes it mechanical. **Recorded here rather than
> quietly fixed, because an RFC that overstates its novelty invites exactly the
> prior-art objection this estate's council seat exists to raise.**

So the estate's standing instruction for the most safety-critical ordering decision it
makes resolves, in the common case, to *no available check* — and the honest sessions
then improvise a substitute each time, undocumented and unverified. **That is the
defect this RFC exists to close.** It is not that sessions are careless; it is that
the mechanism they are told to use expires within hours of the roll it describes.

### 1.4 What the substitute turned out to be, and why it is better than the thing it replaced

Forced to improvise, this lane probed the binary for **the capability the hold exists
to guarantee** rather than for a commit that proxies it — on both pods, each with a
negative control that came out absent:

```
cta_nonpage_destination          PRESENT (both pods)     <- the name 475 puts in a checks array
cta_names_nonpage_destination    PRESENT
cta_tel_malformed                PRESENT
cta_nonpage_destination_NOTREAL  absent                  <- control
stamp_cta_destination_guidance   PRESENT                 <- the key 476 sets
"Destination (fixed)"            PRESENT                 <- the literal the code emits when it acts on it
```

This is strictly stronger than commit ancestry on three counts, and the third is the
one that matters most on this estate:

- it asserts the property itself, not a proxy for it;
- it is immune to the same-tag-rebuild trap (`MEMORY.md`'s first banner: **"a FRESH
  BUILD CAN SHIP NO NEW CODE"** — 203 commits unshipped on 08-17 while pods looked new);
- **it has no shelf life.** A log line scrolls; a stamp needs git and a reachable ref;
  a compiled symbol is in the artefact for as long as the artefact runs.

### 1.5 The data already exists — it just cannot leave the pod

Nothing here needs computing. The registries are already enumerable in every binary:

- `checks.Names()` (`discovery_checks/registry.go`) — the fail-fast error message at
  `discovery_checks.go:208-211` **already prints the entire registered set** to a log
  nobody can query an hour later;
- the action registry (`actions/registry.go`) — the same question in the form CLAUDE.md
  already warns about under *"Image first, then seeds — a seed naming an unregistered
  action fails at runtime"*;
- `buildinfo.GitCommit`, imported by all 14 backend `cmd/*/main.go` targets and, today,
  **only ever passed to a logger**.

The gap is not knowledge. It is that the knowledge is written to a stream with hours of
retention instead of to the one place every migration, CronJob check and council seat
can already read: the database.

---

## 2. Design

### 2.1 One table, written by the binaries

```sql
CREATE TABLE service_binary_capabilities (
  service       text        NOT NULL,           -- 'agent-chassis', 'core-manager', …
  pod_name      text        NOT NULL,
  image_tag     text        NOT NULL,
  git_commit    text        NOT NULL,           -- buildinfo.GitCommit, verbatim
  kind          text        NOT NULL,           -- 'discovery_check' | 'action' | 'config_key'
  name          text        NOT NULL,
  started_at    timestamptz NOT NULL,
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (service, pod_name, kind, name)
);
```

Each service, at startup, writes its enumerable registries in one transaction and
deletes its own previous rows for that `pod_name`. A lightweight periodic touch of
`last_seen_at` (the existing agent heartbeat is the obvious carrier) is what makes
staleness detectable — **without it a dead pod's rows vouch for a binary that is no
longer running**, which is the same class of error this RFC is trying to end and must
not be reintroduced by its own mechanism.

`kind='config_key'` is **opt-in and explicitly partial**: it carries only keys a call
site chose to declare. A key's absence therefore proves nothing, and the assertion
helper below must refuse to answer for `config_key` unless the service has declared it
supports declaration at all. Stated here because an assertion that silently means
"unknown" is exactly the decorative-control failure this RFC objects to.

### 2.2 One function, called by the migrations

```sql
SELECT assert_live_capability('agent-chassis', 'discovery_check', 'cta_nonpage_destination');
```

Raises unless **every** pod of that service seen within the freshness window reports
the name. Three properties, each load-bearing:

- **every pod, not any** — during a partial roll two replicas differ, and the run that
  fails is the one that lands on the old one. `logs deploy/X reads one pod of N` is
  already a `LANDMINES.md` entry; this helper must not repeat it.
- **fails closed on no rows.** No rows means "nothing told me", which must raise, not
  pass. A helper that passes when uninformed is worse than no helper.
- **it raises, it does not return a boolean** — `ON_ERROR_STOP` ignores a non-empty
  result set, so a verify block of `SELECT`s cannot stop a `COMMIT`. That is already a
  memory-index entry (RFC_006's lesson); this helper is built to the same shape.

### 2.3 What changes for an author

The `_HOLD` convention **stays** — it records human intent and it is what keeps a file
out of the runner's pending list. What changes is that the migration's own opening block
asserts the precondition, so renaming the file early fails loudly at apply time instead
of succeeding into a fleet-wide clobber:

```sql
BEGIN;
SELECT assert_live_capability('agent-chassis','discovery_check','cta_nonpage_destination');
SELECT snapshot_agent('completeness-discovery-agent','475…: pre-update');
…
```

---

## 3. Alternatives considered

1. **Do nothing; keep prose interlocks.** Ruled out by §1.2 (32 stated, 0 enforced) and
   by the council objection. But note honestly what "do nothing" has actually cost so
   far: **no incident is on record from a prematurely-applied hold.** The convention is
   working *today* because sessions are careful. The argument for acting is §1.3 — the
   prescribed check is expiring under them, so the current safety rests on improvisation
   whose quality is invisible to review.
2. **Reuse the existing `agent_capabilities` table** (it exists; 16 rows). Ruled out:
   it is a different concept — agent *skills* for routing, with a `strength numeric(3,2)`
   weight and no pod, image or commit dimension. Overloading it would make a routing
   table load-bearing for deploy safety and would give one word two meanings on a shared
   seam, which is the drift class this track exists to prevent. Named here so the reuse
   seat need not ask.
3. **Have the migration runner (not the migration) do the probe** — `run-migrations.sh`
   already shells out to `kubectl exec`. Rejected as the *primary* mechanism: it protects
   only files applied through the runner, and this lane has just demonstrated the common
   escape hatch — with 25+ other lanes' files pending, `--apply` was not available to me
   and I applied out of band with `psql -f` plus `--record-only`. A control that the
   normal workflow routinely bypasses is not a control. **Worth doing as well**, as
   defence in depth.
4. **Make provenance durable instead** (write `git_commit` to the DB at startup, keep
   asserting ancestry). This is a strict subset of the proposal — the table above carries
   `git_commit` anyway. Rejected as the *whole* answer because ancestry still needs a git
   ref resolvable from wherever the check runs, and it still answers a proxy question:
   §1.4's same-tag-rebuild case is one where the commit is right and the code is old.
5. **A CI/pre-commit gate.** Ruled out by precedent already ratified here: RFC_006 —
   **a pre-commit hook cannot gate live config**, because at commit time the migration is
   unapplied. Same answer, same reason.

---

## 4. Blast radius, named

- **Writes (behaviour changes):** all 14 backend `cmd/*/main.go` targets gain a startup
  write. They already import `pkg/buildinfo`, so the import graph does not change shape;
  the change is that a value already computed is now persisted. A service that cannot
  write must **log and continue, never fail to start** — a capability registry that can
  take the fleet down is a worse bargain than the problem it solves.
- **Reads (new dependency):** only migrations that opt in by calling the helper. Existing
  migrations are untouched and keep their prose headers.
- **Schema:** one new table, one new function. No existing table altered, so the previous
  binary tolerates it (§6).
- **NOT touched:** orchestrator, Kafka, dispatch. No wire contract, no dedupe key, no
  state machine. [MEASURED for the read half — the helper is called only from SQL files
  that name it. The write half's per-binary `go list -deps` delta is **[UNMEASURED]** and
  is owed before implementation.]

---

## 5. Staged rollout plan

1. **Table + function only.** No writers, no callers. Inert by construction.
2. **One service writes** — `agent-chassis`, the busiest and the one carrying the
   discovery-check registry that motivates this. Watch: rows appear for both pods within
   a roll; `last_seen_at` advances; a deliberately killed pod's rows go stale and the
   helper starts failing closed for it. **Induced-fault test, not a happy-path grep:**
   assert on a name that is NOT in the binary and confirm the helper raises; assert with
   all rows deleted and confirm it raises rather than passes.
3. **First real caller** — a new migration, deliberately chosen to be one whose hold is
   already discharged, so the assertion's *first* live run is one we can predict the
   answer to. If it raises, the mechanism is wrong, not the estate.
4. **Remaining services**, then optionally the runner-side probe (§3.3).

**Canary:** none needed for stages 1-2 (no behaviour depends on them). Stage 3's canary
is the migration itself.

---

## 6. Rollback plan

Stage 1-2: drop the function, drop the table; nothing reads them. The startup write is
best-effort by design (§4), so an older binary that does not write is not a failure state —
it simply reports no rows, and the helper **fails closed**, which is the safe direction and
is why fail-closed is specified rather than fail-open. Stage 3+: revert the individual
migration's assertion line. Schema tolerates the previous binary at every stage.

---

## 7. Acceptance evidence

The measurements that retire this RFC's risk (none are collectable yet):

1. Rows present for every live pod of a writing service, `image_tag` matching
   `kubectl get deploy -o jsonpath` for that service, **with a negative control**: a name
   that must be absent, asserted absent in the same query.
2. The induced-fault results from §5.2 — the helper raising on an unknown name and on an
   empty table. A helper that has only ever been observed passing is untested.
3. A stale-row proof: kill a pod, confirm its rows stop being counted within the window.
4. One week on: the count of migrations carrying a mechanical assertion, against the 32
   that today carry prose. That number is the whole point, and it is disconfirmable.

---

## 8. What this RFC does NOT propose

- It does not propose enforcing council review, changing the `_HOLD` convention, or
  gating commits. Only: giving an existing, already-computed fact a durable home, and
  giving migrations a way to assert on it.
- It does not touch `bugs_open/312`'s own fix (migration 477). 477 remains held on the
  owner's decision and its canary; if this RFC were implemented first, 477 would be its
  ideal first caller — but 477 must not wait for it.

## 9. Pointers

- Motivating lane + all measurements: `docs/agent_docs/docs024_key_docs_latest/bugfix_299_cta_dials_phone/`
  (`NOTES_cta_dials_phone.md`, 2026-08-19 entries).
- The council objection that prompted this: `doc_notes` id `55c8a047-d70c-400a-ad72-16c4d340fa66`,
  `bug_historian` seat, submission corr `1f1fecc9-3502-4757-8929-fd173fca2dc6`.
- Precedent for "a hook cannot gate live config": `RFC_006`.
- The traps this mechanism is built not to repeat: `LANDMINES.md` — the 380 trap,
  `logs deploy/X reads one pod of N`, the `strings`/discovery-grep entries, and
  `MEMORY.md`'s "a FRESH BUILD CAN SHIP NO NEW CODE".
- **The prior occurrence, and the reason this is a recurrence rather than a discovery:**
  `bugs_open/215` lane, 2026-08-11, chassis v1.0.1288 — memory
  `a-commit-sha-probe-has-no-positive-control`, indexed under
  `prove-a-deploy-at-the-artefact-index` (14 lessons in that family already, which is itself
  evidence about how often this seam is walked).

## 10. Live evidence added after filing (2026-08-19)

The capability probe was not only reasoned about — it gated a real config change the same day
and the change then behaved as the probe predicted. Migration 475 armed
`cta_nonpage_destination` on `completeness-discovery-agent` after both pods were probed for the
check name with a negative control; an induced discovery run (`webdesign.uk`, corr `ee07fd81`)
then **COMPLETED**, filing 6 findings from the new check alongside the existing ones. Had the
probe been wrong, the fail-fast arm would have failed the whole step and rolled back every
other check's findings in that run. **That is one worked instance of the proposed gate's
predicate being both checkable in advance and correct** — a sample of one, stated as such.
