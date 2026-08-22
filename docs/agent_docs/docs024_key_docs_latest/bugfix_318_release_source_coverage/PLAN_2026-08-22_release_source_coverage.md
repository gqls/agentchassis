# PLAN — `bugs_open/318`: close the coverage gate's self-referential admission

**Lane:** `bugfix_318_release_source_coverage`, opened 2026-08-22.
**Designed with** the `fable` planning agent against this lane's measurements; its
adversarial pass over those measurements is in §7, **including the one correction of
its own that I checked and rejected**. Nothing below is repeated from that agent's
report without being verified here first — a subagent's report is another document.

---

## 1. What is actually wrong

`check-release-coverage` (register **BLD-022**) asks, of every production overlay:
*"is the image this overlay pins one that the release builds?"* — and if not, it
`continue`s. Membership of `RELEASE_IMAGES` is therefore the gate's **own admission
criterion**, so a service left out of that list at birth is not *uncovered*, it is
*out of scope*, and the gate prints "Release coverage OK" about the one case it exists
to catch. Eight services have fallen in; two of them on 2026-08-21 and 2026-08-22,
authored by sessions that had the closing owner ruling in front of them.

The remedy currently in the makefile is a comment in capitals. CLAUDE.md's owner
ruling of 2026-08-02 §2 already settles what that is worth: **"a comment is not a
control on a tree this many sessions share."**

## 2. The predicate set

Three preventive invariants and one detector. Each says what it enumerates and, just
as importantly, what it cannot see.

### P1 — birth admission (the door-closer)

**Asserts:** every production overlay pinning a `$(REGISTRY)/…` image must name an
image in `RELEASE_IMAGES`, **or** the service must appear in a new `OWN_LINEAGE`
exemption list that states why and names the target that does retag it. This
*inverts* today's `continue`: non-membership stops being the admission test and
becomes the violation.

**Enumerates:** the filesystem — `services/*/overlays/production/**/kustomization.yaml`
at **any depth**, reading `newName` when present else `name`.

**Cannot see:** a workload with no `overlays/` tree at all (base-only, applied by
hand — two exist, §7 correction B); an overlay never applied to the cluster
(`capped-schedule-ordering-check`, today); images outside our registry, which is
correct — `ollama/ollama` and `postgres:16-alpine` are not ours to roll.

**Would have caught:** all 8 birth omissions, at the first `deploy-core` after birth —
and at the *birth commit itself* once §3's advisory is in.

> **⚠ P1 goes live against a compliant tree.** All 32 production overlays currently
> pin either a `RELEASE_IMAGES` member, an upstream image, or a placeholder, so
> `OWN_LINEAGE` ships **empty** and the gate stays green. A green gate on a compliant
> tree proves nothing — it could only have passed. P1's discriminating power comes
> from the mutation table in §5 and from nowhere else.

**Why the exemption is a list and not a judgement.** `OWN_LINEAGE` is an opt-in field
whose **unsafe side (exemption) is the default OFF**, with **zero live consumers** —
enumerated, not asserted: the overlay census in RUNBOOK R5 is the query. That is
exactly the shape CLAUDE.md's RFC_022 narrowing (owner, 2026-08-11) declares **not**
architecture-scope, and exactly the shape the 2026-08-02 §2 ruling prescribes for
new authority on a shared seam.

### P2 — build completeness — **ALREADY DONE, by construction** (`95757b6c2`)

`build-backend: $(addprefix build-,$(RELEASE_IMAGES))`. Set difference is now 25/25
and cannot drift. Nothing further to build; the residual work is bookkeeping —
correct BLD-022 §(iv) ("policed by **nothing**") and append the dated correction to
the LANDMINES birth-omission entry.

> **Do not "re-verify" P2 by re-running the set difference.** Post-`95757b6c2` the two
> sides are one declaration, so 25/25 holds **by identity and could not come out
> otherwise**. The discriminating evidence is the copy-mutation
> (`No rule to make target 'build-nobody-built-this-check'`, exit 2, against exit 0 on
> the live file), recorded in that commit's message.

### P3 — deploy-side validity (the reverse direction BLD-022 §(iv) names)

**Asserts:** every `AGENT_DEPLOY_SERVICES` entry's resolved image is in
`RELEASE_IMAGES`, and every `RETAG_EXEMPT` / `OWN_LINEAGE` entry names a service that
has an overlay. **Enumerates:** the makefile declarations only — pure set logic.

**Would have caught:** none of the 8 historical incidents. It exists because BLD-022
§(iv)'s hazard is still unpoliced: drop `github-actions-runner` from `RELEASE_IMAGES`
while leaving both runners in `AGENT_DEPLOY_SERVICES` and both `ImagePullBackOff`
**together**, taking CI with them. P2's derivation did **not** close this — a name
deleted from `RELEASE_IMAGES` consistently stops being built *and* pushed, but the
deploy loop still retags the orphaned entry. Nearly free alongside P1, so it ships
with it.

### P4 — daily cluster census (detection for what no filesystem gate can see)

One `doc_notes` row per run, **clean or not** (the WFA-013 rule: a MISSING row means
the job did not run, and must not read as "nothing is wrong"). Three comparisons:

- **C1 stragglers** — any cluster workload on a `$(REGISTRY)/` image whose tag differs
  from the modal `$(REGISTRY)/` tag in the cluster. Would have seen the runners
  (v1.0.948 / v1.0.1126 against a modal 12xx) within a day, and sees
  `optional-explicit-wires-check` at v1.0.1321 today.
- **C2 declared-but-absent** — an `AGENT_DEPLOY_SERVICES` entry with no workload.
  Catches `capped-schedule-ordering-check`, the phantom a filesystem gate is
  structurally incapable of seeing.
- **C3 running-but-undeclared** — a workload on a `$(REGISTRY)/` image mapping to no
  declared service.

**Cannot see:** whether a running image's *content* matches its tag — that is
BLD-019/BLD-020 territory, deliberately not duplicated here. It **detects**, it does
not prevent; a cluster-side absence has no commit to gate.

### Candidate 1 (the content-change trigger) — re-aimed, and the re-aim goes to the owner

The owner's stated intent of 2026-08-18 was *"fail the release when a service's pinned
tag predates the last commit to the sources that service is built from."* As worded it
is **uncomputable on this tree**: the pin's history is fiction (26 dirty overlays;
`agent-chassis` committed at v1.0.1239 against v1.0.1323 on disk and in the cluster).
The durable fact is the mechanism, not the number — `deploy-agents` seds `newTag` in
place and committing it is nobody's step. Now in `LANDMINES.md`.

The honest anchor is the **artefact**: BLD-019's OCI `revision` label / binary stamp
gives the commit *C* an image was built from, and staleness is then
`git log C..HEAD -- <closure>` non-empty, with the closure being the dockerfile's Go
target's `go list -deps` **plus `go.mod`, `go.sum` and the dockerfile itself** (a
dependency or base-image bump changes the artefact with no diff under `cmd/`) **plus
every file the dockerfile `COPY --from=builder`s besides the binary** — for
`commit-sha-exposure-check` that is its own acks list, i.e. its definition of a
permitted exception.

**But note what P1 does to that trigger's population.** Once every our-image overlay is
covered-or-exempt, and every release moves every release image at one pinned commit
(BLD-020), "stale iff its source moved" can only ever bite an `OWN_LINEAGE` entry — a
set that is **empty today**. So the recommendation is to implement the predicate as
**the standing obligation attached to an exemption** (an `OWN_LINEAGE` entry declares
its source closure; the daily census grades that service's artefact-commit against it)
rather than as a release gate over an empty population.

**That is a re-aim of an owner's stated intent, so it is his call and not this lane's.**
P1/P3/advisory/census are built regardless; the trigger waits on a ruling. Put to him
with the two facts that force it: the fiction finding, and the empty population.

## 3. The shape, and why

| layer | fires at | hard or advisory | why there |
|---|---|---|---|
| `pkg/releaseset` + `cmd/releasecheck` | `deploy-core` → `make release` | **hard** | the release is the last moment before the estate diverges |
| `scripts/pattern-check.py` addition | `git commit` | advisory | **the only layer that fires at the moment the omission is made** |
| CronJob census | daily | reports | the cluster shapes no filesystem gate can see |

- **Package `pkg/releaseset/`** — beside `pkg/buildinfo`, same BLD family. `internal/`
  holds service implementations, `platform/` holds runtime machinery; neither fits
  estate tooling.
- **`cmd/releasecheck/`** — thin, the `cmd/regcheck` pattern exactly: every decision in
  the package, `main.go` parses flags and prints. **This layering is load-bearing for
  review, not taste:** `cmd/` is outside `scripts/council-scope.sh`'s regex, so a
  helper living wholly in `cmd/` would draw no council round at all. Putting the
  predicates in `pkg/` is what makes 318's own warning ("decide the shape before
  assuming a review is available") come out on the reviewable side.
- **The makefile recipe becomes `go run ./cmd/releasecheck …` and the shell
  implementation is deleted in the same commit.** Two implementations of one gate is
  precisely the drift class `099_SYNC_gate_roster.py` exists for. The Go side subsumes
  the shell predicate as test T3.
- **Mutation-testability is the second reason for Go.** Every predicate is a pure
  function over parsed data, so the proofs are table rows and `testdata/` fixtures and
  **the live makefile is never edited** — the binding lesson of `f016b07ec`
  (2026-08-22, this morning: a session mutated the shared makefile in place to prove
  this very gate discriminates, and another session committed it inside the window).

**Not a make evaluator.** `ParseMakefileDecls` is a literal extractor for the four
`NAME := …\` continuation blocks — the `cron_parity_test.go` idiom, one language over.
A missing block is **exit 2, "could not run"**, never a pass: a gate that passes what
it failed to measure is this estate's own blind-pass landmine, and `bugs_open/131`'s
lane shipped exactly that inside the check written to end it, four days ago.

## 4. Phasing

0. **Done** — P2 (`95757b6c2`). Bookkeeping only: BLD-022 §(iv) + LANDMINES correction.
1. **Coordination, not code** — the `capped-schedule-ordering-check` phantom belongs to
   the `316` lane; contribute the finding into their file rather than fixing it here.
   And tell the owner: the next release will **create** that CronJob (it is in
   `AGENT_DEPLOY_SERVICES`, and `deploy-agents` applies overlays), so a cluster-shape
   change arrives with it. **The release must run at a bumped tag ≥ v1.0.1325** —
   v1.0.1324 is contaminated (`commit-sha-exposure-check:v1.0.1324` was hand-built from
   an unpinned commit and three overlays already pin 1324; a same-tag re-push hits the
   node-cache landmine).
2. **The gate** — P1 + P3 + the port of the existing predicate, one coherent commit.
   Council submission before or alongside it.
3. **The advisory** — the `pattern-check.py` function.
4. **The census** — its own commit and its own council round; RBAC + overlay +
   `RELEASE_IMAGES` membership **in the commit that creates it**, which is the rule
   this whole lane is about.
5. **Owner decision** — the candidate-1 re-aim, after phase 2 lands.

Phase 2 first because it closes the recurrence mechanism that has fired twice in 48
hours. The census is last because it is detection, and this estate's own evidence
(*"detection works; schedule and dispatch do not"*) says its first scheduled run must
itself be verified at the artefact.

## 5. Verification — every case with its disconfirming result named

| case | fixture | must return | a different result means |
|---|---|---|---|
| **T1** motivating case | overlay pins `aqls/new-check`, absent from all lists | violation naming `new-check` | the rewrite reproduces the self-referential blindness — the 08-21/08-22 shape survived |
| **T2** negative control | same, with `new-check` listed | zero violations | a gate that fails a compliant tree, which gets disabled (pattern-check.py's own header records that decay) |
| **T3** ported direction | `render-audit-adapter` reconstruction: pins `aqls/browser-runner-adapter`, in no list | violation naming it | the port dropped the direction the shell gate already proved |
| **T4** runner hazard (P3) | `github-actions-runner-vmsites:github-actions-runner` in the deploy list, image removed from `RELEASE_IMAGES` | violation naming the paired-`ImagePullBackOff` consequence | BLD-022 §(iv) is still unpoliced |
| **T5a** exemption honoured | `OWN_LINEAGE` entry for T1's service | zero violations | exemptions unusable → the gate gets worked around |
| **T5b** exemption bounded | `OWN_LINEAGE` entry with no overlay, or no reason | violation | the exemption is a blanket mute, not a declaration |
| **T6** depth | overlay at `overlays/production/kustomization.yaml`, no region dir, our image, unlisted | violation | the `tools-api`-shaped hole survives (§7 correction A) |
| **T7** registry prefix | overlay pins `ollama/ollama` | never a violation | the gate polices upstream images and becomes noise |
| **T8** parser loudness | fixture makefile with no `RELEASE_IMAGES :=` block | exit 2, never 0 | the blind-pass landmine, one rung up |

Plus **`decl_parity_test.go`**: `ParseMakefileDecls(live makefile)` must equal
`make -s print-release-images`; **skip loudly** if `make` is unavailable, never
silently pass (the `optional_budget_cron_parity_test.go` idiom). And for the census's
first live run, a **demand control** — grade it against a known straggler
(`optional-explicit-wires-check` at v1.0.1321, or a scratch image with a mutated baked
list), because a clean first run on a coherent fleet could only have passed.

Hand-run control for the RUNBOOK, and it is the only one that touches a makefile:
`cp makefile $TMP/m && <inject> && make -f $TMP/m check-release-coverage` — must fail
naming the injected service, while the live tree stays green.

## 6. Deliberately NOT built, so the absence reads as a decision

- **No make evaluator.** Literal parser + parity test is the trade.
- **No git-history staleness detector, anywhere, ever in this lane.** That history is
  fiction here; anything dated from overlay commits is refused on principle.
- **No gating of development overlays.** Only 9 services have one; a hard failure would
  break every non-production deploy (BLD-022 §(i)'s precedent stands).
- **No auto-remediation.** The census writes rows; acting on them is work-item territory.
- **No fleet-wide extension of BLD-023.** Only `agent-chassis` calls it; that is the
  153/312 seam. The census reads image tags, not capability rows, specifically to avoid
  depending on it.
- **No third at-birth enforcement layer.** Advisory at commit + hard gate at release is
  the estate's settled two-layer pattern, and `.githooks/pre-commit` explicitly warns
  that a stray non-zero exit there stops the whole fleet committing.

**Risks, named:** replacing a live, mutation-proven shell gate (mitigated by T3 plus one
recorded shell-vs-Go equivalence run in NOTES *before* the swap); a Go toolchain is
needed at `deploy-core` time (present on the release machine, and its absence fails
loudly rather than silently — the recipe says so); `go run` collapses a child's exit
code, so the 1-vs-2 distinction must be printed in words as well as returned; new RBAC
surface for the census (list-only, dedicated ServiceAccount, named in the submission per
the 2026-07-29 §3 ruling that a shared mechanism's other consumers must be **told**, not
merely measured).

## 7. The planning agent's adversarial pass over this lane's measurements

Three corrections were offered. **Two hold and one does not**, and I checked all three
rather than adopting them.

**A. ACCEPTED, and it strengthens the design.** I had recorded that `tools-api` has no
production overlay. It has one — at
`deployments/kustomize/services/tools-api/overlays/production/kustomization.yaml`, with
**no region directory** — which the gate's fixed `$(OVERLAY_PATH)` glob
(`production/uk_001`) can never see. It pins a placeholder (`newName: IMAGE_TAG`,
`newTag: latest`) and has no workload in the cluster, so it is harmless today. The
corrected lesson is worth more than the correction: **the enumeration must not bake in
the region depth**, which is why T6 exists. The same overlay is also the estate's only
live use of `newName`, so the `newName`-else-`name` read is a correctness fix for a
**latent** case, not a live one — stating it as live would be the overclaim.

**B. ACCEPTED as a wording fix.** I wrote that `site-discovery-staleness-check` and
`site-locale-unset-check` have "no `services/*` overlay on disk". Both **do** have
`deployments/kustomize/services/<name>/base/`; what they lack is an `overlays/` tree.
The substance survives — a gate globbing `overlays/…` cannot see them, and both run
`postgres:16-alpine` — but the original wording would send a reader to a directory that
exists and make my claim look false.

**C. REJECTED, with the arithmetic.** The agent said the counterfactual release would
have aborted "after **23** successful pushes, not 22", because
`commit-sha-exposure-check:v1.0.1324` exists locally and would have pushed. Both halves
are wrong. `push-backend` iterates `RELEASE_IMAGES` **in declaration order**, and the
three unbuilt images are positions **23, 24, 25** — the last three — while
`build-backend` built exactly positions 1–22. Enumerated
`docker image inspect …:v1.0.1324` over all 25: only #21 `content-loss-check` and #24
`commit-sha-exposure-check` are present. So the loop pushes 1–22 (built that run),
**fails at #23 `optional-explicit-wires-check`**, and #24 is never reached — its local
presence is irrelevant to the abort. **22 successful pushes, abort on the 23rd**, as
originally recorded. The agent's secondary point does stand and is kept: that
hand-built `…:v1.0.1324` is itself a hazard — a release-form tag on an image built from
an unpinned commit — and is why phase 1 says bump to ≥ v1.0.1325.

Also accepted, as an addition rather than a correction: the source closure must include
`go.mod`, `go.sum` and the dockerfile itself. Folded into §2.

---

## 8. ⚖️ OWNER RULING 2026-08-22 — §2's candidate-1 recommendation is SUPERSEDED

> *"we can skip the 18 August staleness build."*

**§2's closing recommendation** — that the content-change predicate be re-aimed at the
artefact and implemented as the standing obligation attached to an `OWN_LINEAGE` entry —
**is not to be built.** Neither is the 2026-08-18 wording. The paragraph stays above as
the record of what was proposed and why; this section is what happened to it.

**Read it as CLOSED, not parked.** The reasoning §2 gives is the reasoning that emptied
it: uncomputable on this tree (the pin's history is fiction), and after P1 its population
is the exemption list, **one entry as of 2026-08-22**.

**What was taken instead, and it is not a consolation prize.** The risk that skipping
actually carries is not staleness — it is that `OWN_LINEAGE` becomes the next hiding
place, exactly as "not in `RELEASE_IMAGES`" was. So:

- `ExemptionBudget = 3` in `pkg/releaseset/predicates.go` — the release fails past it,
  naming every entry. Polices the **accumulation**, not the entry (the WFA-013 / RFC_022
  shape), with N lower than 10 deliberately. **A judgement, one line, the owner's to set.**
- **The gate names the standing exemptions on every GREEN run.** This is the load-bearing
  half: a threshold silent until it trips is a threshold nobody watches.

Committed `8fe69e6c6`, proven by `T10` plus a copied-tree run.

**§6's "deliberately NOT built" list is amended in one place:** *"no git-history staleness
detector"* becomes *"no staleness detector of any kind, git-history or artefact-anchored —
owner ruling, not a lane judgement"*. The rest of §6 stands, including the cluster census,
which this ruling does not touch and which remains a different question.

**§5's verification table gains T10** (silent at the budget; one over reports and names
the whole accumulated set; the remedy offers a *reviewed* raise, not a bare refusal). The
disconfirming results: silence above the budget would mean the exemption list can grow
without limit — rebuilding the hole this change closed, with better paperwork; a finding
*at* the budget would mean the guard fires on a state the owner explicitly permitted,
which is how a gate gets disabled.
