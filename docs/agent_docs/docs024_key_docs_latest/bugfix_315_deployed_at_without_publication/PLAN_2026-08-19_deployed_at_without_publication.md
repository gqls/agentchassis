# PLAN — bug 315: make `pages.deployed_at` mean something, and make publication checkable

**Written 2026-08-19.** Design, phasing, decisions and their reasons. Corrections live here, marked.

## The problem, restated after measuring it

`pages.deployed_at` is read across the estate as "this page has shipped". It is written by five
agents. **Two of them write it before the deploy is even dispatched**; the other three write it
after a commit **whose result they discard**. So there is no arrangement of the current workflows
under which the column could be evidence of publication.

Underneath that sits a second fact that decides the design: **"commit is deploy" is asynchronous and
batched.** The origin is rewritten per changed domain directory by one `b2 sync` on a self-hosted
runner, tens of minutes after the commit (measured: 40 sampled pages sharing a single three-second
`last-modified` window while their stamps spread across the following hour). A synchronous stamp
*cannot* honestly assert publication, however well guarded — at the moment it is written, the claim
is reliably false and becomes true later if nothing goes wrong.

And a third, from register `DGH-009`: **an unchanged file commits as an EMPTY commit** and the
adapter still reports success. So neither "the step succeeded" nor even "a commit exists" implies
bytes moved. Only content does.

## Decisions

### D1 — Split the fact. Keep `deployed_at` as the ever-shipped floor; carry publication evidence separately.

`deployed_at` keeps its current meaning — *this page's artefact was committed to the deploy repo* —
strengthened so it is only written downstream of a real commit result. Publication evidence goes
into a content hash plus a commit sha, and publication itself is confirmed **asynchronously**, by a
sweep, never inside the stamp.

**Why not just redefine/narrow it in the docs:** it leaves 315's actual question permanently
unanswerable, and even the narrowed reading is false for the two agents that stamp pre-deploy.

**Why not make it mean "served bytes confirmed":** the delivery half is outside this repo's process
boundary and minutes away in time. Worse, ~15 consumers depend on the current semantic as an
*ever-shipped floor* (`queryresolve.go` eligibility, `check_componentless_pages`,
`maintenance_actions` staleness). Redefining it would let a transient CDN failure start delisting
working pages — worse than the bug.

**Why the split is cheap:** it reuses `pages.content_hash` (dead, 0/786, `varchar(64)` — exactly a
sha256 hex), `page_components.deploy_commit` (dead, 0/1,775), the `publish_site` served-bytes
acceptance idiom, and the discovery-check framework. The site-level publish seam already made this
exact design choice one level up; this is the same idea at page granularity, with acceptance
detached in time because delivery is detached in time.

> **Note, and it is the strongest argument for the shape:** the platform already *retired a config
> key for this feature*. `UpdatePageStatusInputSpec.RemovedConfigKeys["commit_from"]` says the intent
> was *"recording which git commit a page's content was DEPLOYED in, from the git_commit step's
> output … unimplemented — pages has no such column. **Implement it as a feature if wanted**, do not
> re-add the key."* And the page-rerender seed's output contract already promises
> `"deploy_result": "git commit result with commit_sha"` — a promise the adapter has never kept.
> This plan is that feature, under a new name as the retirement note requires.

### D2 — Make the layer that knows say what it knows.

`CommitToRepo` computes `newCommitSHA` and returns `repo.HTMLURL` (a per-repo constant). Change it
to return an outcome struct carrying the sha, a `no_change` flag (new tree sha == parent's tree sha)
and, for deletions, the absent paths. The adapter's success payload then carries `commit_sha`,
`no_change` and `files_sha256` (path → sha256 of the exact bytes committed; the adapter holds them).

Cost is small and was verified rather than assumed: **3 production callers, all inside the
git-adapter**; `CreateBranch`/`CreatePullRequest` do not call it; the delete tests are `_, err :=`
and compile unchanged.

**`no_change` is REPORT-ONLY in v1.** Skipping the empty commit would change what the shared seam
guarantees for all 19 `git_commit` steps (today every success fires the deploy workflow and a cache
purge). That is architecture-scope and is deliberately excluded.

### D3 — The stamp guard reads a CONFIG-NAMED field, never a literal one.

New optional config key on `update_page_status` naming the deploy step's output field. When set, the
`deployed` branch refuses the stamp on a skip/failure and writes the evidence columns on success.

**This is not stylistic.** `[MEASURED]` the 19 live `git_commit` steps carry **nine different
`output_field` names** and two set none; `deploy_result` names only three of them. A guard
hard-coding `deploy_result` would be blind on 16 of 19 and — since it must fail open — would wave
them through silently, which is this bug reintroduced inside its own fix.

Fail-open is required (a config typo must not freeze deploy stamps fleet-wide) but must be
**countable**: a durable `agent_error_log` row whenever the field is set and unreadable. A silent
fail-open is how this bug survived four rerenders.

⚠ **The verdict must be resolved through the estate's existing resolver, not by indexing a path.**
`deploy_result` has two shapes — inline (`…response.data.success`) and, when the deploy is done by a
called sub-agent, one level deeper (`…response.deploy_result.response.data.success`). That second
shape is **57 of 744 rows (7.7%)** over 7 days. `datahelpers`' whole-tree search and
`tryUnwrapMapPatterns`/`UnwrapDeep` exist for exactly this; do not hand-roll a second unwrapper.

### D4 — The two pre-deploy stampers are a CONFIG reorder, not a guard.

`page-build-handler` and `tool-recreation-handler` stamp before any deploy exists *by construction* —
no guard can read evidence that is not there. Both then call `page-rerender`, **which already
contains its own post-commit `update_status`**. So the fix is to delete their early stamp step and
let the rerender's stamp stand.

DB-only, live immediately, and **on its own it takes "2 of 5 stamp before the deploy" to zero.**

### D5 — The divergence sweep compares HASHES, not timestamps, and only after a settle window.

A new discovery check comparing `pages.content_hash` against the sha256 of served bytes fetched with
a cache-buster.

**Not timestamps.** The bug file's §5 candidate 4 proposes `deployed_at` vs origin `last-modified`.
That check would have convicted the two *healthy* pages in the bug's own §2 table, because a
byte-identical rerender legitimately rewrites nothing. And `[MEASURED]` it flags 40 of 40 sampled
pages at any given instant, because the origin lags in batches. **It cannot separate "not synced
yet" from "will never sync"** — only elapsed time can, and the known bad case took six hours.

False-positive sources and their mitigations: mid-delivery window → require `deployed_at` older than
a configurable settle window; CDN edge → cache-buster and assert the origin-header shape; retracted
and archived pages (which deliberately keep `deployed_at`) → predicate on `status='active'`; pages
with no hash yet → predicate `content_hash IS NOT NULL`, so the check is structurally inert until D2
has shipped and reports nothing before then.

Emission: a new item type, per-page dedup key, **no handler agent in v1** — re-filing a rerender is
the loop that already failed four times; the honest first move is visibility.

## Phasing

Ordering matters in a way that is the inverse of the usual rule. `UpdatePageStatusInputSpec` is
**`StrictConfig: true`**, so a config naming a key the running binary does not declare fails
validation. **The image must ship before the config**, and that is a real, stated ordering
constraint rather than a convenience.

| phase | kind | live when | useful alone? |
|---|---|---|---|
| 0a — drop the pre-deploy stamp from the two handlers (D4) | config migration | immediately | **yes** — kills the worst half by itself |
| 0b — schedule the already-built, undriven `deployed_zero_components` detector via its existing `emit_checks` override | config | immediately | yes — detection exists and has never run |
| 1 — `CommitOutcome` + payload `commit_sha`/`files_sha256`/`no_change` (D2) | Go, **git-adapter image** | on rebuild + roll | yes — every deploy becomes auditable |
| 2 — the guard + evidence writes; migration adding `pages.deploy_commit` (D3) | Go, **chassis image** + migration | on rebuild + roll | with 1 |
| 3 — name the deploy field on the three post-commit stamping steps | config migration | **must follow phase 2's roll** (StrictConfig) | — |
| 4 — the divergence check + enabling it (D5) | Go + config | chassis roll, then config | gated on 1–3 |

## Out of scope, and why

- **The `gqls/sites` GitHub Actions workflow** — it lives outside this repo and the repo is private,
  so §5 candidate 3 ("why is this page skipped by the batch") cannot be diagnosed from here. D5 is
  deliberately designed to *detect* that failure from this side without reading the runner.
- **The 42 `planned` componentless pages** — never built, nothing served; a different defect already
  owned by `check_componentless_pages` / `nav_linked_never_built`. Phase 0b starts surfacing them
  incidentally; repairing them is the build pipeline's job.
- **`bugs_open/083`'s stranded `needs_human_review` items** — that bug's disease. D5 deliberately
  does not emit into `needs_human_review`, so it cannot inherit the stranding.

## Architecture-scope assessment

**Not architecture-scope**, with the reasons: the new config key is opt-in with the unsafe default
OFF (absent ⇒ today's behaviour) and **zero live steps name it** — the exact RFC_022 shape, and the
enumeration is the evidence, not the assertion. The adapter payload additions are
additive-and-inert (no chassis code reads any payload field today). The new item type has no
automated consumer, which is fine under the 2026-08-02 §1 ruling **provided** the producer and key
shape are named in the concept register in the same commit.

**Deliberately excluded as RFC-shaped:** suppressing the empty commit on `no_change`, and any
synchronous served-bytes gate inside the stamp.

**Consumers to TELL, not merely measure** (2026-07-29 ruling 3): the 16 agents carrying `git_commit`
steps; the `webdesign_tool_rebuilds` lane that filed this; retraction owners (`deployed_at`
semantics for non-active pages are unchanged, and D5 depends on that); and `deployment-github.md`
DGH-001, updated in the same commit as each shipping phase.

## Status

Phases are **not implemented**. This document is the design; nothing here has shipped.
The diagnosis loop produced no verdict (LLM usage cap, twice) — see NOTES for the first-hand
substitute, stated plainly per the owner ruling of 2026-07-31.

---

## Council gate

Submitted 2026-08-19 (phases 0–2 — the config reorder, the adapter return contract, the guard).

**`SUBMISSION_CORR = 377167cd-6324-4bc7-a866-87ad8c435132`**

Five edits, twelve `grounded_in` evidence quotes, risks stated including the two `[UNVERIFIED]`
items (whether the two handlers' `complete_workflow` output fields tolerate the removed
`status_updated`, and whether anything downstream of the commit rewrites page bytes — the private
`gqls/sites` workflow could not be read).

Read the verdict with:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='377167cd-6324-4bc7-a866-87ad8c435132' AND kind='council_report' ORDER BY created_at;
```
