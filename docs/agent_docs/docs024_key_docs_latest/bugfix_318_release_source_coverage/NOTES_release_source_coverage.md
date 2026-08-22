# NOTES — `bugs_closed/318` release source coverage

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-08-22 — session opens: is 318 still valid, and is anyone on it?

**Ownership check first** (`scripts/who-owns.py`, CLAUDE.md § "Before routing work AT
an existing bug"):

- `153` → **OWNED** by `bugfix_153_build_provenance` (10 commits in 14 days, last
  2026-08-17). Its headline fix is FIXED AND LIVE on `v1.0.1283`; what remains is a
  residual list, not an open defect. So this session took the fallback the owner named.
- `318` → filed 2026-08-19 by the `bugs_open/237` lane, **status "OPEN, not started"**,
  no owning lane, no commits against the file since filing. The only other commits in
  its territory are from the `staged_component_build` lane, which *tripped over* the
  blind spot on 08-21/08-22 while creating two new check services and folded them into
  the lists (`67201d125`) — a fix to two instances, explicitly not to the class.
- **Work-item queue**: nothing open matching `release` / `RELEASE_IMAGES` /
  `coverage gate` / `build-backend` / `318`. Query in the RUNBOOK (R1).

**Verdict: 318 is unowned and this session takes it.** Lane created today.

### Why `090` (the diagnosis loop) was NOT run — stated, not omitted

CLAUDE.md's owner ruling of 2026-07-31 says a `bugs_open/` file asserting a
cross-cutting root cause is not filed until it has been through the loop **or the
session states plainly why it substituted equivalent first-hand verification**.

This session is not filing a new root cause. 318's cause was established by the
`bugs_open/237` lane and is already in the concept register as **BLD-022**'s stated
blind spot. Everything added below is a **mechanical measurement with a control**, not
a theory: a set difference from `make -n`, a `docker images` census with two positive
controls, a `kubectl get cronjob` read. Each could have come out otherwise, and one of
them (finding 5) came out **against** the design I expected to recommend. That is the
substitute, declared.

### What was measured `[MEASURED 2026-08-22]`

**1. The gate's admission test is self-referential.** `check-release-coverage`
(makefile `:142`) `continue`s past any overlay whose pinned image is not already in
`RELEASE_IMAGES`:

```make
case " $(foreach i,$(RELEASE_IMAGES),$(REGISTRY)/$(i)) " in \
        *" $$img "*) ;; \
        *) continue;; \
esac;
```

Membership of the list is the gate's own admission criterion, so an image left out at
birth is not *uncovered*, it is *out of scope* — and the gate prints
"Release coverage OK". Already in `LANDMINES.md` (§ "A CHECK SERVICE OMITTED FROM
`RELEASE_IMAGES` AT BIRTH…"), and 8 services have now fallen into it: the original six
(fixed 08-18), then `optional-explicit-wires-check` (born 08-21) and
`commit-sha-exposure-check` (born 08-22).

**2. The `build-backend == RELEASE_IMAGES` invariant is BROKEN ON HEAD.** BLD-022
§(iv) records it as "verified by set equality 2026-08-18 [MEASURED] and policed by
**nothing**". Four days later it is no longer true:

```bash
make -n build-backend 2>/dev/null | grep -oE 'docker build -t [^ ]+' \
  | sed 's#.*/##; s#:.*##' | sort -u > built.txt          # 22
awk '/^RELEASE_IMAGES :=/{f=1} f{print} f&&!/\\$/{exit}' makefile \
  | sed 's/^RELEASE_IMAGES :=//; s/\\//' | tr -s ' \t' '\n' \
  | grep -v '^$' | sort -u > declared.txt                  # 25
comm -23 declared.txt built.txt
```

→ **declared but not built: `optional-explicit-wires-check`,
`commit-sha-exposure-check`, `capped-schedule-ordering-check`.**
Reverse direction (built but not declared): **empty**.

Cause: the names went into `RELEASE_IMAGES` and `AGENT_DEPLOY_SERVICES` but not into
`build-checks`, the `build-backend` prerequisite that fans out to the check images.
`git status --porcelain makefile` is empty and `git diff HEAD -- makefile` is empty, so
this is **HEAD, not local WIP**.

**3. The live consequence, measured at the artefact — the next release aborts.**
`push-backend` (makefile `:476`) loops `RELEASE_IMAGES` with
`docker push … || exit 1`. `IMAGE_TAG` is `v1.0.1324`; the live fleet is `v1.0.1323`.

| image | local tags present | in `build-backend`? |
|---|---|---|
| `optional-explicit-wires-check` | **v1.0.1321 only** | no |
| `capped-schedule-ordering-check` | **none, ever** | no |
| `commit-sha-exposure-check` | v1.0.1324 (hand-built today) | no |
| `agent-chassis` *(control)* | v1.0.1320, 1321, 1322, 1323 | yes |
| `verifier-remit-check` *(control)* | v1.0.1320, 1321, 1322, 1323 | yes |

The two controls are the discriminating half: a service the release really does build
carries an image at **every** recent tag. So `make release IMAGE_TAG=v1.0.1324` builds
22 images (~6 min), then dies on the first `docker push` of an image nobody built —
**before `deploy-core` and `deploy-agents`, so nothing deploys at all.**

**4. Cluster state** (`kubectl -n ai-persona-system get cronjob`):
`optional-explicit-wires-check` runs daily on **v1.0.1321**, three tags behind the
fleet. `capped-schedule-ordering-check` has **no CronJob in the cluster at all** —
overlay, make targets and `RELEASE_IMAGES` membership all exist for a service that was
never applied. `commit-sha-exposure-check` and `content-loss-check` are at v1.0.1324
and have never fired.

**5. ⚠ NEGATIVE RESULT, and it argues against 318's candidate 1 as worded.** Candidate
1 says: *"fail the release when a service's pinned tag predates the last commit to the
sources that service is built from."* Reading "when the pin last moved" out of the
overlay's git history **cannot work on this tree**: `deploy-agents` seds `newTag` in
place and nobody commits the result. Right now **26 production overlays are dirty**,
and `agent-chassis` is committed at `v1.0.1239` while the working tree says
`v1.0.1323` — 84 tags of divergence in one file. Any design that dates the pin from
git is dating a fiction. The honest instrument for "what commit is this artefact
from" is BLD-019's OCI label / binary stamp, or BLD-023's
`service_binary_capabilities.git_commit` — not the overlay.

**6. The source closure IS mechanically derivable.** Every backend dockerfile names its
build target (`./cmd/X`, or `cmd/X/main.go` for `git-adapter` and
`remote-job-spawner`); `go list -deps ./cmd/config-key-audit` returns 309 packages in
0.29 s. But **six images share `./cmd/config-key-audit`** and differ only in CMD flags
and a COPY'd acks JSON, so a closure computed from Go deps alone would call all six
identical and would miss the acks file — which for `commit-sha-exposure-check` is the
check's own definition of a permitted exception.

**7. "Has a dockerfile" is too wide a net.** Five backend dockerfiles have no
production overlay: `tools-api`, `migrator`, `seeder`, `platform`, `workflow-monitor`.
The right enumeration is "an overlay pins a `$(REGISTRY)/…` image".

**8. And the filesystem is not the whole cluster.** `site-discovery-staleness-check`
and `site-locale-unset-check` run as CronJobs with no `services/*` overlay on disk.
Both are `postgres:16-alpine`, so no freeze risk today — but it bounds what any
filesystem-only gate can claim.

---

## 2026-08-22 — the narrow unblock shipped, and it was mutation-proven on a copy

`95757b6c2`: `build-backend: $(addprefix build-,$(RELEASE_IMAGES))`, plus a
`build-github-actions-runner` alias for the one image whose target predates the naming
convention (`build-github-runner`).

Chose the derivation over adding three names deliberately. Adding names fixes three
instances of a defect whose shape is *"two hand-maintained enumerations with nothing
keeping them in step"* — which is the same shape `bugs_open/237` removed from the four
deploy lists, and the same shape that produced this. `$(addprefix …)` deletes the
second enumeration. Set difference now **25/25 IDENTICAL**.

**Mutation proof, on a COPY.** `cp makefile $SC/mf-mutant`, inject
`nobody-built-this-check` into `RELEASE_IMAGES`, `make -f $SC/mf-mutant -n
build-backend` → exit **2**, *"No rule to make target 'build-nobody-built-this-check',
needed by 'build-backend'"*. Live makefile, identical command → exit **0**. Never the
live file: `WRONG_CALLS.md` 2026-08-22 (`f016b07ec`) records a session doing exactly
that this morning and another session committing the file inside the window.

**MISSTEP, logged in `WRONG_CALLS.md`:** the first run of that proof read the exit code
off a pipeline (`make … | tail -3; echo "exit=$?"`) and printed a reassuring `exit=0`
beside the failure message, because `$?` was `tail`'s. That control would have said
`exit=0` whether or not the mutation took effect. It is the same defect the council's
`editquality` seat raised against `bugs_open/153`'s provenance check — a bug file I had
read forty minutes earlier in this session.

**Second misstep, same file:** `tr ' ' '\n'` over `RELEASE_IMAGES` reported **nine**
missing images against a true three, because the continuation lines are tab-indented.
`tr -s ' \t' '\n'` is now in RUNBOOK R2.

## 2026-08-22 — the LANDMINE entry was carried by another session's commit

Appended the `newTag`-is-never-committed landmine, ran `git diff --numstat` (41 lines),
and by the time `git commit` ran the file was already clean: the `361` lane had
committed it inside the window as `8cc994b12` (*"a CronJob's Job listing shows only the
last N failures…"*). Both entries survived intact; that lane noticed and recorded it in
`f0db82afc`. Nothing lost, forward-only holds — but **this lane's commit message for
that entry does not exist**, so `git log`ging the landmine returns a commit about
CronJob Job listings. Attribution lives in the entry's own `**added:**` line, which is
why that line is not decoration. Third instance in ~24 h of the trap already recorded at
`LANDMINES.md:14307`.

### The unblock was checked end-to-end, not just at the set difference

A derivation that adds three images to `build-backend` moves the release's failure from
`push-backend` to `build-backend` if any of those three cannot build from committed
`HEAD`. So the three were actually built, at a scratch tag never pushed
`[MEASURED 2026-08-22]`:

```
make build-capped-schedule-ordering-check IMAGE_TAG=scratch-318-probe   → image exported OK
make build-optional-explicit-wires-check  IMAGE_TAG=scratch-318-probe   → image exported OK
make build-commit-sha-exposure-check      IMAGE_TAG=scratch-318-probe   → image exported OK
docker rmi …:scratch-318-probe ×3                                       → removed
```

`capped-schedule-ordering-check` is the one that mattered: it had **never been built at
any tag**, so it was the only one of the three with no prior evidence that its
dockerfile works. Both acks files the sibling images `COPY --from=builder` exist at
committed `HEAD` (`git cat-file -e HEAD:<path>`), which is the specific way these
builds fail — `ref_build` archives `HEAD`, so a working-tree-only acks file would break
the build for everyone else and not for the author.

---

## 2026-08-22 — the gate is built, and the council APPROVED it round 1

Submitted `83442a5a-e66d-4772-8872-b445f521d47b` (`DRY_RUN=1` first, so admission cost
nothing). Verdict **approved, 3 advisory objections, none high**, in about 25 minutes.

**⚠ Read the verdict the safe way.** CLAUDE.md's recipe
(`doc_notes … ORDER BY created_at DESC LIMIT 1`) returns whoever finished last, which
with ~40 live sessions is routinely someone else's round — there is a LANDMINE about it
from this morning. Key on the correlation instead:

```sql
SELECT metadata->>'decision', body FROM diagnosis_artifacts
 WHERE correlation_id='83442a5a-…' AND kind='council_report';
```

### The two objections that were worth CODE, not a reply

**`bug_historian` (medium) — and it was right.** `firstImage()` stopped at the FIRST
element of an overlay's `images:` block, inheriting the shell gate's `awk … exit`. If an
overlay pins two images and OURS is the second, the scan produced no `Pin` and the gate
could never flag it. That is **this fix reproducing, in miniature, the exact shape it
exists to close** — a check reporting clean about the thing it never looked at.

Measured before acting `[MEASURED 2026-08-22]`: **no kustomization anywhere under
`deployments/` has more than one element**, so the gap was LATENT rather than live. The
cap is now **gone rather than warned about** — `blockImages()` returns every element with
its own tag — because a cap that is safe only because of today's data is the wrong kind
of safe, and a warning surface would have been the worse answer. T9 is the seat's own
missing fixture.

**`editquality` (medium).** It asked whether `RETAG_EXEMPT` is really declared today,
since requiring it risks a self-inflicted "could not run" on day one. It is — the live
gate parses and runs green, which is the proof — **but the objection was right about the
RULE rather than the fact.** `RETAG_EXEMPT` is a CLEARING list exactly like
`OWN_LINEAGE`, and this package's own argument for `OWN_LINEAGE` being optional applies
word for word. Both clearing lists are now optional; only the two JUDGING lists are
required. The new test pins the direction that makes it safe: dropping `RETAG_EXEMPT`
**exposes** what it cleared rather than clearing anything.

### The objections that were answered with a query rather than a change

**`prior_art_librarian` (medium ×2) — "these precedents are asserted, not checked".**
Fair: I had read all four, and the submission did not attach the check. `[MEASURED]`

```
cmd/regcheck/main.go                                    exists
cmd/config-key-audit/cron_parity_test.go                exists
cmd/config-key-audit/optional_budget_cron_parity_test.go exists
scripts/council-scope.sh:57  COUNCIL_SCOPE_CODE_RE='^(platform|internal|pkg)/'
```

The scoping claim is verbatim. `grep -c "kustomiz\|overlays\|images:" cmd/regcheck/main.go`
→ **0**, so it is a naming/shape precedent only, which is what was claimed.

**`reuse_agent` (medium) — "does `95757b6c2`'s derivation already parse this list?"**
No. The change is `build-backend: $(addprefix build-,$(RELEASE_IMAGES))` — **make
expanding its own variable**, with no parsing of any kind. There is nothing to reuse.
And `grep -rln kustomization --include=*.go .` returns **only** `pkg/releaseset`; the two
Go files matching `images:` are a prose comment about reference images and a JavaScript
object literal in an injected snippet. `pkg/releaseset/overlays.go` is the estate's only
Go reader of a kustomize `images:` block.

**`guardian` (medium ×2) — behaviour change on the single release choke point, and
`OWN_LINEAGE` completeness "unknown until the gate was built".** The second has a direct
answer the seat could not have had: the gate enumerates the **whole filesystem**, so its
first green run over 31 our-registry overlays **is** the completeness check — any other
out-of-band overlay would have been named in the same breath as `admin-dashboard`. The
first is real and is the owner's to watch: the first release under this gate should be
run with someone reading the output. Recorded, not argued away.

### The six mutation controls, re-run after the refactor (copied tree, never the live file)

| mutation | exit | finding |
|---|---|---|
| verbatim copy | 0 | — |
| `agent-chassis` out of `AGENT_DEPLOY_SERVICES` | 1 | `NO RELEASE PATH: agent-chassis` (**reproduces the original 237 bug state**) |
| `content-loss-check` out of `RELEASE_IMAGES` | 1 | `OUR IMAGE, NO RELEASE BUILDS IT` (**the shape the old gate passed**) |
| `OWN_LINEAGE` emptied | 1 | names `admin-dashboard` |
| `OWN_LINEAGE` entry with no target | 1 | `EXEMPTION NAMES NO RETAG TARGET` |
| `RELEASE_IMAGES` renamed away | 1 | `THE CHECK COULD NOT RUN` — never a pass |
| `RETAG_EXEMPT` renamed away | 1 | exposes `auth-service` + `core-manager`, does **not** refuse |

### The advisory's precision, measured against real history

Audited **every** commit since 2026-07-15 that added a production overlay — 21 of them.
It fires on **7**, and all 7 are the documented incidents: `render-audit-adapter` (237's
founding case), `github-actions-runner-vmsites`, `component-render-check`,
`shared-output-fields-check`, `verifier-remit-check`, `optional-explicit-wires-check`,
`commit-sha-exposure-check`. **Zero false positives** on the other 14 — including
`effd08fff`, a check service another lane created **correctly, an hour earlier**, which
is a negative control this lane did not construct.

### An unplanned live test, from another lane, within the hour

While this was being written the `309` lane shipped `component-source-vocabulary-check`
— a brand-new CronJob image — and did it right: `RELEASE_IMAGES`, `AGENT_DEPLOY_SERVICES`,
an overlay, **and** a `build-<image>` target, which the `95757b6c2` derivation now
requires or `build-backend` fails with `No rule to make target`. The gate went from
30-of-33 to **31-of-34 overlays judged, still green**, and the advisory stayed silent.
A pass that could have failed.

---

## 2026-08-22 — owner ruling: skip the staleness build, guard the exemption list

Put four options to the owner (skip; skip + guard the excused list; build the full
artefact-anchored staleness check; do the cluster census instead) with the costs and the
honest counter-argument for each. **Ruling: skip + guard.**

Recorded in the bug file, both register entries and the PLAN, each time with the *reason*
rather than just the verdict — a later reader who finds only "the content-change trigger
was never built" will re-file it, and this estate has a landmine for exactly that shape.

**The counter-argument I gave against my own recommendation, kept because it is real:**
the budget cannot fire today (one entry). This estate refuses guards with no live subject
— BLD-023 declined `assert_live_capability()` on exactly that ground, *"a fail-closed
helper with exactly ONE caller is a mechanism nobody exercises."* The distinction I
claimed, and it is the whole justification: **this one needs no caller.** It runs on every
`deploy-core` whether or not it fires. If that distinction is wrong, the guard is the
mistake, and it is written at the constant so the next reader can say so.

**The design detail worth carrying to other budgets:** the threshold is the smaller half.
`printExemptions` names the standing entries on every GREEN run, so the count is in front
of a human continuously rather than at a trip point. A threshold silent until it trips is
a threshold nobody is watching — this estate has measured that failure repeatedly
(*"detection works; schedule and dispatch do not"*).

**A misstep in the mutation proof, caught by reading the output rather than the exit
code.** Testing "at the budget is silent" on a copied makefile by fabricating two extra
exemptions gave **exit 1** — which looked like the budget firing one short. It was not:
`b` and `c` name services with no overlay, so `KindExemptionWithoutOverlay` fired and the
budget stayed quiet. The exit code alone would have had me lower N. **A composite gate's
exit code cannot tell you WHICH guard spoke** — grep the finding kind, not the status.
The clean at-budget case is the unit test, which builds real overlays.

---

## 2026-08-22 — phase 2: the cluster census, and it broke its own report on the first run

`--census` mode on the same binary, `make release-census`. Three comparisons over a
`[]Workload`; `cmd/releasecheck/cluster.go` owns the only part that touches client-go, so
the predicates stay table-testable with no cluster.

### The first live run `[MEASURED 2026-08-22]`

```
Release census — namespace ai-persona-system: 29 of 45 workloads run a
docker.io/aqls/ image; the fleet is on v1.0.1323.  5 findings.
```

| finding | service | independently confirmed by |
|---|---|---|
| AHEAD OF THE FLEET (hand-deployed) | `commit-sha-exposure-check` v1.0.1324 | the `docker images` census earlier today |
| AHEAD OF THE FLEET (hand-deployed) | `content-loss-check` v1.0.1324 | same |
| BEHIND THE FLEET TAG | `optional-explicit-wires-check` v1.0.1321 | the overlay census earlier today |
| DECLARED BUT NOT RUNNING | `capped-schedule-ordering-check` | found by hand; contributed to the 316 lane |
| DECLARED BUT NOT RUNNING | `component-source-vocabulary-check` | the 309 lane's own commit message: *"NOT yet built or deployed"* |

**Zero false positives**, and no `RUNNING BUT NOT DECLARED` at all — every `aqls` workload
in the cluster is accounted for by a declaration.

### ⚠ THE RUN EXPOSED A DEFECT IN MY OWN REPORT, and it is the useful part

The first two findings came out as **"RUNNING AN OLD FLEET TAG: commit-sha-exposure-check
… at v1.0.1324 while the fleet is on v1.0.1323"**. They are not old. They are **newer** —
hand-built and hand-deployed ahead of the fleet.

**A report that states the opposite of the truth is worse than one that stays quiet.** A
reader chasing a frozen service would have found one that was, if anything, too new, and
concluded the instrument works. This is the same family as the blind-pass landmine, one
rung along: not a check that misses, but a check that *asserts the inverse*.

Fixed three ways, each with a test:

- the kind is split by **direction** — `BEHIND THE FLEET TAG` vs
  `AHEAD OF THE FLEET TAG (hand-deployed)` — with **opposite remedies**: a straggler wants
  a release; a hand-deploy wants the tag never reused, because a same-tag re-push serves
  the node's cached image;
- tags are ordered **numerically, not lexically**. Not theoretical: this estate is on
  `v1.0.13xx`, so it has already crossed `v1.0.999 → v1.0.1000`, where a string comparison
  says 999 is the newer. `TestCompareTags` pins that boundary and the real
  `v1.0.948 → v1.0.1126` runner freeze;
- a tag that **cannot** be ordered (`latest`, a date, a different arity) gets its own
  finding kind rather than a guessed direction. A fabricated ordering is how a report
  starts asserting what it does not know.

**What would have caught it earlier:** nothing in the unit tests, because I wrote fixtures
for the case I had in mind (behind) and not its mirror. The live run was the control.
Generalisable: **when a finding has a direction, write the fixture for BOTH directions
before running it anywhere** — a one-sided fixture cannot fail on the side it omits.

### Two deliberate limits, stated so they are not read as oversights

- **Hand-run only.** No CronJob, no RBAC manifest, no `doc_notes` row. The makefile
  comment says so in capitals. Scheduling is a separate decision with its own round, and
  the split is the estate's own lesson: *"detection works; SCHEDULE and DISPATCH do not."*
  Nothing runs this unless a person does, and it must not be described as a live detector.
- **First container only.** `readWorkloads` reads `containers[0]`; a sidecar carrying one
  of our images would be missed. No workload here has one today. Named because it is the
  **same shape** the council caught in round 1 (the `images:` block cap) — if a sidecar
  ever appears, this is where to look.

Submitted `b0883c17-32a1-434d-b0ab-114df4cb04b1`.

---

## 2026-08-22 — census round 1: REVISE, and the gating objection found the same defect a THIRD time

Verdict `revise`, gating HIGH from `editquality`. It was right, and it is the most useful
thing this lane has been told all day.

**The objection.** `modalTag` broke ties with `tag > best` — a **lexical** comparison, in
the same file as, and one screen below, the numeric comparator I had just written to fix
exactly that. A tie between five workloads on `v1.0.999` and five on `v1.0.1000` picks
`v1.0.999`: lexically higher, numerically older.

**Why it is worse than the bug it mirrors.** The fleet tag is the **single value every
straggler and ahead-of-fleet finding is measured against**. Getting the comparator wrong
inverts one line; getting the **tie-break** wrong inverts the entire report.

**Three occurrences, one file, one afternoon:**

1. the shipped report that called two hand-deployed services "old" (found by the live run);
2. the first repair for it, which compared strings (caught before commit);
3. this tie-break (caught by the council).

**The lesson, written at the function so it is not just an anecdote here:**
**a helper that does the comparison correctly does not protect the call sites that do not
use it.** `TestCompareTags` was green throughout all three.

**Fixed and mutation-proven the decisive way round:** restoring `tag > best` makes the new
test fail with its own message; the fixed version passes. The test also pins three things
the objection did not ask for — the end-to-end census on that fleet, that a clear 3-2
majority still wins outright (so the tie-break does not quietly become the rule), and that
an unorderable tie is **deterministic across 20 calls**, because a report that changes run
to run is not reproducible.

### The reuse objection: answered with three measurements, and its premise was wrong

`reuse_agent` (medium) held that the CronJob check services "must already list/read cluster
workloads to do their jobs", so a shared helper probably existed. They do not — **they read
Postgres.** `config-key-audit`'s own dockerfile header says it reads *"DIRECTLY from
Postgres … no kubectl in this image, no pods/exec RBAC"*. `[MEASURED 2026-08-22]`

| question | command | answer |
|---|---|---|
| is there a shared k8s client wrapper? | `grep -rln 'kubernetes.NewForConfig' --include=*.go .` | **six inline sites, no wrapper** — and `agent_image.go` is `rest.InClusterConfig()` only, so it cannot serve a hand-run CLI |
| does anything list workloads? | `grep -rn 'Deployments(.*).List(\|CronJobs(.*).List(\|DaemonSets(.*).List('` | **zero hits** outside `cluster.go` |
| does any check service reach the API? | each service's dockerfile `CMD` | **none** — all Postgres or `postgres:16-alpine` |

**What the objection DID surface, recorded and deliberately not acted on:** six inline
`kubernetes.NewForConfig` bootstraps with no shared wrapper is a real extraction
opportunity. Doing it inside a bug-fix round means touching five unrelated call sites,
which is precisely the scope shape the `guardian` seat vetoes — and it would be right to.
**Named here as a follow-up so its absence is a decision rather than an oversight.** Whoever
takes it: `spawn_actions.go`, `diagnose_build_gate_action.go`, `agent_image.go`,
`internal/adapters/thunder/ssh/secrets.go`, `cmd/remote-job-spawner/main.go`,
`cmd/releasecheck/cluster.go`, and note the in-cluster-only/kubeconfig split is the one
real difference between them.

---

## 2026-08-22 — `v1.0.1326`: the close condition, measured

Whole-fleet release, owner-run. **20 Deployments + 11 CronJobs on one tag.**

**The gate ran.** `auth-service` and `core-manager` are on `v1.0.1326`, and both move only
through `deploy-core` → `update-kustomization-images` → `check-release-coverage`, with a
non-zero exit propagating through `pinned_sweep`'s `|| exit 1`. **Not overclaimed:** the
tree was compliant, so it *could only have passed*. Exercised, not proven — the six
mutation controls remain the discriminating evidence, which is where BLD-022's own entry
had to be corrected to put it.

**The census, before → after:**

```
before:  29 of 45 workloads, fleet v1.0.1323, 5 findings
after:   31 of 47 workloads, fleet v1.0.1326, 1 finding
```

| finding | resolved how |
|---|---|
| BEHIND `optional-explicit-wires-check` v1.0.1321 | → v1.0.1326 |
| AHEAD `commit-sha-exposure-check` v1.0.1324 | → v1.0.1326 |
| AHEAD `content-loss-check` v1.0.1324 | → v1.0.1326 |
| DECLARED NOT RUNNING `capped-schedule-ordering-check` | **created 15:09:35Z** |
| DECLARED NOT RUNNING `component-source-vocabulary-check` | **created 15:09:37Z** |

**The two creation timestamps are the strongest evidence this lane produced.** The
prediction — *"the next release will CREATE that CronJob, because it is in
`AGENT_DEPLOY_SERVICES` and `deploy-agents` applies overlays"* — was written into the bug
file and into a contribution to the `316` lane **before** the release, and it **could have
come out otherwise**: the release could have aborted at `push-backend` (which `95757b6c2`
fixed) or the overlay could have failed to apply. Two seconds apart, in the apply loop.

**The `v1.0.1324` contamination was avoided as recommended** — the fleet went to 1326, so
the two hand-built images were never re-pushed over.

**The surviving finding is the census working:** `live-declaration-drift-check`, filed by
the `363` lane (`18661b3c7`) *after* the release, declared and not yet running. Caught the
day it appeared.

### Misstep, and it is a record-integrity one

The bulk `sed` that repointed `bugs_open/318` → `bugs_closed/318` across 19 files **rewrote
history along with the pointers** — dated `**source:**` lines, `**added:**` lines,
`WRONG_CALLS` headers and one sentence in the owner's README that was true when written.
Caught by the pre-commit pattern check flagging removed lines from an append-only ledger,
whose own advice (*"if this IS a deliberate consolidation, say so in the commit message"*) I
had not followed because I had not noticed. **The warning was in the output of my own commit
and I read past it.** Corrected by appending dated notes rather than editing again; full
entry in `WRONG_CALLS.md`. The transferable half: **before a bulk repoint, split the hits —
`source:|added:|filed|lane|<a date>` is the leave-alone pile.**

### A second, smaller near-miss

Grepping for the `git mv` landmine I wrote the pattern with backticks — `` grep -n "`git
mv`" `` — and the shell **executed `git mv`**. It failed harmlessly with no arguments and
changed nothing. The estate's known form of this trap is backticks inside `git commit -m`;
a **grep pattern** is a second surface for it, and `git mv` with plausible arguments would
have moved a file on a tree ten sessions share. Single-quote any pattern containing
backticks.
