# COMMISSION — five owner rulings, 2026-08-10 · hand this to a fresh thread

**Authority: the owner ruled on all five items on 2026-08-10.** This doc is the brief, not a
proposal. Where it says a thing was measured, the query or the file:line is inline — re-run it
rather than trusting it, because every figure here has a date and this tree moves in hours.

**Written by the `rfc012_await_findings` lane, which is now CLOSED** — its own two jobs are done
(`HANDOFF_2026-08-10_continue_here.md`). Nothing below is blocked on that lane; it is where the
context lives, not where the work must happen.

**Suggested order: 5 → 2 → 1 → 3.** That is not arbitrary — **item 5 unblocks item 1's research**,
and item 2 produces evidence item 1 needs. Item 3 is independent and can run in parallel by a
different thread.

| # | ruling | state | rough size |
|---|---|---|---|
| 1 | Awaited-merge design: **wait, then research hard what is best** | research commissioned, **decision deliberately deferred** | large, mostly investigation |
| 2 | Make the three silent hero/logo readers **say something** | **approved, do it** | small (hours) |
| 3 | **Enable `deploy_commit` usefully** — store the git commit | approved; **bigger than it looks, see below** | medium, spans 3 layers |
| 4 | `PROCESS` trigger wording ("adds, changes or removes") | **confirmed, no work** | done |
| 5 | Let the diagnosis loop **see `orchestration_states`** | ~~approved, do it~~ **BUILT + COMMITTED 2026-08-10 (`5f8a326fc`), INERT until a fleet roll** → `diagnosis_schema_visibility/` | small–medium |

---

## 1. The awaited-merge design — WAIT, then research hard

> **Owner, 2026-08-10:** *"it is possible that another thread removed those keys thinking they
> weren't used, but I'm not sure. wait and then research hard as to what is best."*

### The hypothesis was tested before this doc was written. It is NOT supported.

Checked three ways, all clean — record this so nobody re-runs it:

- **Go, the writer:** `deploy_image_asset_action.go:369-371` still assigns `image_url`,
  `output_path`, `size_bytes` onto the action's result. Nothing removed it.
- **Go, the history:** `git log --since=2026-07-01 -S 'image_url'` over the writer and the three
  readers returns **one** commit, a bulk `auto updated …` sweep. No thread deleted this handling.
  The merge region (`coordinator.go` ~2700-2760) and `storeActionResult` have **no commits at all**
  since 2026-07-15.
- **Config:** both producing agents still declare the keys —
  `site-work-orchestrator` and `pageflow-builder` each carry `deploy_hero_image` and
  `deploy_logo_image` steps with `action: deploy_image_asset` and
  `output_field: hero_deployed` / `logo_deployed`. Live query in `bugs_open/236`.

**So nothing was removed. The value is written by code that still writes it, into a key that is
still declared, and it is absent by the time the readers run.** The loss is in the runtime path
between the two, and that is what "research hard" has to establish.

⚠ It is *not* migration 356 either: that removed `commit_from`/`output_format` (unrelated keys) and
was applied 2026-08-10 evening, while the affected rows are from 2026-08-09 13:06 and 13:50.

### What the research must settle, in this order

**(a) Where `image_url` is lost.** This is the blocker and it is still unknown. `bugs_open/236` has
the live evidence and — importantly — **names a REFUTED theory**: the obvious candidate, "the
awaited-response merge overwrites the key", is contradicted by `coordinator.go:2719-2748`, which
reads the existing map and *adds* to it. Do not re-derive that theory; do not quote it as the
cause. Remaining candidates, all `[UNVERIFIED]`, are listed in 236 §5.

⚠ **This is why item 5 comes first.** Two `090` diagnosis runs on this symptom returned
UNVERIFIABLE — the second (`074beb8a`) purely because the loop could not query
`orchestration_states` (it asked for `id`; the column is `orchestration_id`). Fix item 5 and this
question becomes answerable by the loop. Attempting (a) before item 5 means doing by hand what the
tool should do.

**(b) Only then, which merge design.** `CENSUS_2026-08-07_rfc012_await_step_readers.md` §7 sets out
Design 1 (reply nested under `response`) vs Design 2 (additive, reply keys at top level) and
recommends **Design 2 sharply**.

⚠ **Do not take that recommendation at face value — its premise is contradicted by production.**
The census scores Design 2 as *"Go side: 0 breaks (the three direct accesses still find `image_url`
at the top level)"*. Live rows show `image_url` is **not** there today. So either the census's
baseline was wrong, or (a)'s unknown mechanism removes it — and until you know which, "0 breaks"
may really mean "0 *additional* breaks on an already-broken path". **Re-establish the baseline
before scoring either design.**

**(c) `(a′) storeActionResult` is explicitly NOT covered by that census** (different write path, own
reader set — census §7 says so). If the ruling is to extend to it, that is its own census.

**Deliverable:** a recommendation with the baseline re-measured, not a patch. The design decision
returns to the owner.

---

## 2. Make the three hero/logo readers say something — APPROVED, do it

> **Owner:** *"2. yes."*

> **STATUS 2026-08-11 (added by the lane that did it): BUILT, COUNCIL-APPROVED
> (`c80ea1d7-ce1e-493f-8175-877501d895e6`, round 1, no high-severity objections),
> COMMITTED, and INERT until a fleet roll.** Work, evidence and the standing five:
> `docs024_key_docs_latest/silent_hero_logo_readers/`.
>
> **Four corrections to this section, all established by doing it:**
>
> 1. **The line numbers below are stale.** The hero/logo readers are at
>    `v3_site_actions.go:1125` and `:1136`, not `:1020`/`:1031`. Grep the symbol.
> 2. **A `Warn` is NOT sufficient, and this section's own verify note is wrong
>    about it.** It says the change "will be greppable, because a log string is a
>    real literal" — true, and irrelevant to whether the finding survives. It does
>    not: `agent-chassis` does not retain the line (its own startup line was
>    measured absent from `--tail=3000` hours after a roll), and the run record
>    holding the shape is pruned after **4 hours**, because `AWAITING_RESPONSES` is
>    the shortest-lived status in `orchestration_states` and is exactly the state
>    these keys live in (measured first-hand off the live
>    `scheduled_tasks.pre_query`; the repo seed says 24h and disagrees). So the
>    change also writes an `agent_error_log` row via the existing `LogActionError`
>    door — `error_code DEPLOYED_IMAGE_RESULT_MISSING_URL`. Declared as the review
>    question in the submission rather than slipped in; every seat that examined it
>    endorsed it.
> 3. **The demand gate is the load-bearing part and this section does not mention
>    it.** `BuildRenderContextAction` runs for every page build and most pages
>    deploy no image, so an `else` on the outer guard files a row per page
>    fleet-wide. Absence is silent; only present-but-unusable records.
> 4. **"Three silent readers" is right, and "three of a larger silent class" is
>    NOT.** The council's `bug_historian` seat pressed on exactly that. Censused:
>    64 occurrences of the shape in `platform/orchestration/actions`, of which ~50
>    are config/input reads, 5 are loaded records, and **all 4 genuine
>    awaited-result readers were opened and none has this defect** — they either
>    fail loudly (`generate_image_actions.go:777`) or fall through to a legitimate
>    alternative (`call_agent.go:736`/`:375`, `spawn_actions.go:1804`). The
>    discriminator is not "an `ok` guard with no `else`" but **"a miss whose only
>    consequence is a quietly worse artefact"**. 4 dynamic-key readers stay
>    `[UNVERIFIED]` — they resolve a key from step config.
>
> **The `precedent to copy` below (`f7111f4d8`) was read and is a `Warn`-only
> pattern; it is the right shape for a terminal error and the wrong one here,
> for reason 2.**
>
> **Item 1(a) now has a forward evidence path** that does not race the 4-hour
> prune. See `bugs_open/236`'s 2026-08-11 contributions — including a **candidate**
> for §5's open root cause (the park path writes a state re-loaded from the DB and
> never copies `CollectedData` onto it), filed to the diagnosis loop as
> `dbcc4259-ab84-494b-a48b-1df647209a40` rather than asserted. **Verdict not yet
> read at the time of writing.**

The three readers take a value they need, and when it is absent they do nothing at all — no log, no
error, no work item. That is why a page shipping with no hero and no logo looked identical to a page
that never wanted one, for weeks.

**The sites** (all the same shape: `ok`-guarded map access, no `else`):

- `platform/orchestration/actions/v3_site_actions.go:1020` — `hero_deployed` → `hero_url`
- `platform/orchestration/actions/v3_site_actions.go:1031` — `logo_deployed` → `logo_url`
- `platform/orchestration/actions/assemble_from_library.go:452` — the third instance

**What to add:** on the miss, a `Warn` naming the key it wanted, and **what the map actually
contained** (its keys — not the whole value, which can be large). The keys are the diagnostic: a
map holding `response/response_status/response_received_at` and no `image_url` tells the reader
immediately that this is the 236 shape.

**Scope discipline — this is observability ONLY.** Do not "fix" the read by also accepting
`response.data.file_path`. That is a design decision belonging to item 1, and encoding the merged
shape at three call sites is exactly the patch-one-site-by-hand pattern the census already found
evidence of at `unified_extractor.go:200`.

**Precedent to copy:** the same fix was made for `fallback_url_field` in `webscrape_actions.go`
(commit `f7111f4d8`) — a `Warn` on the empty branch plus naming the key in the terminal error.

**Verify:** it is inert until a chassis roll; then the first page build that hits the 236 shape logs
it. Council gate as usual (`platform/` change); pod-grep your new phrase after the roll — and note
it *will* be greppable, because a log string is a real literal.

---

## 3. Enable `deploy_commit` usefully — store the git commit

> **Owner:** *"3. enable the column usefully. store the git commit."*

**Target column: `page_components.deploy_commit`** — TEXT, exists, documented as "Git commit
reference when deployed", **1,329 rows, all NULL, no Go writer**. (`pages.deploy_commit` was
deliberately dropped by `sql_for_tables/003` as "belongs in page_components" — do not resurrect it.)

### ⚠ This is NOT a wiring job. The sha is computed and then thrown away.

Measured 2026-08-10, and it is the finding that sets the scope:

- `internal/adapters/git/github_client.go:256` computes `newCommitSHA` from `createCommit(...)`,
  uses it at `:261` to update the ref — and the function then **returns `repo.HTMLURL`**, not the
  sha. The value is discarded inside the client.
- `internal/adapters/git/adapter.go:459-465` builds the response payload:
  `repo_url`, `files_count`, `commit_message`, `file_path`… **no sha field.**
- Confirmed against a live row: `logo_deployed.response.data` on
  `3e46be5b-8788-447b-9643-e32ae33f601b` carries `files, domain, success, repo_url, file_path,
  repo_name, timestamp, files_count, commit_message` — **no sha**.

**So three layers change:** the client must return the sha it already has → the adapter must put it
in the response payload → the action must write it to the column. Plus a migration/backfill decision
for the 1,329 existing NULLs (recommend: leave them NULL and document that NULL means "deployed
before this shipped", rather than inventing history).

### Scope and review

Changing a client return signature and adding a field to a shared adapter response payload is a
**wire-shape change on a shared contract**. Under `PROCESS_architecture_review.md` (as worded — see
item 4) that is plausibly architecture-scope, not plain council-gate. **Route it deliberately:**
read the trigger, and if in doubt write the RFC — the cost is one document.

⚠ **Adding a field to a response payload is additive, but the census class applies:** find who else
consumes that payload before assuming nothing cares. `git_commit` responses are read in more than
one place.

**Also worth deciding while you are there:** whether the *page* level wants it too. `deploy_commit`
is per-component; a page is many components and could be deployed across several commits. That is a
modelling question the column's original author never answered.

**Verify:** a real deploy writes a real 40-char sha that `git cat-file -t` resolves in the target
repo. **A non-NULL value is not enough** — a wrong-but-plausible sha is the failure mode here.

---

## 4. `PROCESS` trigger wording — CONFIRMED, nothing to do

> **Owner:** *"4. your wording is fine."*

`docs024_key_docs_latest/architecture_review/PROCESS_architecture_review.md` now reads *"it **adds**,
changes or removes an exported symbol other packages depend on"*. This matches what two council
seats were already applying in practice (RFC_019 §10). **Confirmed by the owner; do not revert.**
Item 3 should be routed against this wording.

---

## 5. Let the diagnosis loop see `orchestration_states` — APPROVED, do it

> **Owner:** *"5. yes, let diagnosis loop see the orchestration states file."*

> **STATUS 2026-08-10 (added by the lane that did it): BUILT AND COMMITTED
> (`5f8a326fc`), council-submitted `df9dae6c-b7ca-4605-8dd4-26462ce4b20b`, and
> INERT until a chassis image ships.** Work, evidence and the standing five:
> `docs024_key_docs_latest/diagnosis_schema_visibility/`.
>
> **Two corrections to this section, both established by doing it:**
>
> 1. **The producer is found.** It is **`gatherSchema`**, a Go helper in
>    `diagnose_load_runtime_action.go` (returned under the `schema` key, rendered
>    by the assembler at `:306`) — *not* a `load_schema_hint`-style step. The
>    `[UNVERIFIED]` note below is settled; do not re-run that search.
> 2. **It was never one missing table — this section understates it by 5×.** The
>    include filter (`site%|page%|content%|flow%`) selects **26 of 433** live
>    tables, and **five of the six tables the gather itself renders rows from**
>    fell outside it: `agent_error_log`, `orchestration_states`,
>    `agent_definitions`, `llm_call_log`, `code_symbols`. Only `site_work_items`
>    matched. A second defect this section does not name turned out to be the
>    load-bearing one: the listing **never said it was filtered**, so a
>    filtered-out table and a non-existent table rendered identically — which is
>    why the run asked for a human instead of requerying.
>
> The design note below ("prefer deriving it") was right and was followed: the
> always-list derives from the action's own SQL, and a test re-derives it and
> fails when a new query adds an uncovered table.
> **Item 1(a) is still blocked — on the ROLL now, not on this work.**
>
> **VERDICT: APPROVED** (round 1, 1 medium + 5 low advisory objections, none
> high-severity, 2 seats abstained). Trailer `Council-Reviewed: df9dae6c-…`.
>
> **⚠ The `architecture` seat asked for one thing to be recorded HERE, and this
> is it.** The notice this change adds is *instructive prompt content read by an
> LLM* — it tells the verdicter *"you do not need a human to confirm it"* and
> points it at `information_schema.columns` via `data_request`. The seat approved
> it (`ARCHITECTURE_SIGNAL: point_fix`) because it documents an **existing**
> read-only channel rather than adding one, so nothing a shared mechanism
> guarantees has changed. Its forward-looking concern, verbatim:
>
> > *"if this phrasing pattern gets reused across other diagnosis actions it
> > could accumulate into a de facto shared vocabulary without ever passing
> > through architecture review."*
>
> **So: the SECOND diagnosis action to teach its prompt a self-service capability
> in this style is the one that needs an RFC — not this one.** Whoever writes it
> should treat this paragraph as the precedent that makes it the second, not the
> first. Nothing to do today.

### The failure, in the loop's own words

`090` run `074beb8a-adb4-4074-905a-cb0f857e7f85` (2026-08-10), verdict `UNVERIFIABLE`,
`stopped: scope-not-narrowing`:

> *"The bundle's own data_request for this row failed with `column "id" does not exist (SQLSTATE
> 42703)` — the `orchestration_states` table isn't in the bundle's Schema section, so its real
> primary-key/id column is unknown and must be confirmed by a human before requerying; guessing
> again would likely fail the same way."*

It queried `WHERE id = …`; the column is **`orchestration_id`**. The symptom text had given it the
right value, and it still could not address the table.

### Why this matters beyond one bug

The Schema section exists precisely so *"the verdict writes a correct data_request instead of
guessing a relation that does not exist"* (`diagnose_assemble_bundle_action.go:303-309`). With the
platform's central run-state table missing from it, **any hypothesis whose evidence lives in
`orchestration_states` is unfalsifiable by the loop for reasons unrelated to the hypothesis** — and
the verdict that comes back is `UNVERIFIABLE`, which reads to a casual eye like "your premise was
wrong". It is not. Two separate 236 runs died this way.

### Where to start — and what is NOT yet established

The section is rendered from `params.CollectedData` at the path in `schemaField`, default
**`runtime.schema`** (`diagnose_assemble_bundle_action.go:131`, used at `:305`).

⚠ **`[UNVERIFIED]` — I did not establish what populates `runtime.schema`.** I looked and did not
find it in the `diagnose-agent` step configs by the obvious queries; that is a gap in my check, not
evidence of absence. **Find the producer first** — likely a `load_schema_hint`-style step or a Go
runtime-bundle builder — and change the table list there. Do not patch the renderer.

**Design note worth considering rather than just adding one table:** a hard-coded table list is the
same drift class this estate keeps filing (see the `097`/`(d)` routing-key literal, disclosed in
that submission's risks). If the list is a literal, adding `orchestration_states` fixes today and
leaves the next missing table to be discovered the same expensive way. Prefer deriving it, or at
minimum leave a note naming the next-most-likely omissions.

**Verify — and make it disconfirmable.** Do not just check the table appears in the bundle text.
**Re-run a `090` whose evidence lives in `orchestration_states`** and confirm its `data_request`
now executes instead of erroring 42703. `bugs_open/236` is the ready-made case, and settling it is
item 1(a)'s blocker — so this verification does double duty.

---

## Traps this week paid for — read before touching any of the above

- **A change with no string literal cannot be pod-grepped.** Pure control flow or a one-line guard
  gives you nothing to grep, and the nearest quotable text is the doc comment, which is **not
  compiled** — the check then reports 0 on a correct binary. Date the binary with a literal from a
  **descendant** commit and prove it with `git merge-base --is-ancestor`. (`LANDMINES.md`,
  2026-08-10.)
- **A post-fix count of zero needs a DEMAND control, not a traffic control.** "Total rows and
  distinct types in the window" passes cleanly while blind to whether the path under test ran at
  all. Ask what demand there was on the exact path, and bucket the baseline by day and read its
  **tail** — a producer that went quiet before your fix shipped makes the test unfalsifiable.
- **A top-level step census can be confidently wrong.** `jsonb_each` over
  `default_config->'workflow'->'steps'` misses steps nested in a loop `sub_workflow` — it returned
  3 of 6 on the migration-356 work. Walk recursively, or cross-check with `::text LIKE`.
- **The live spelling may not be the documented one.** The HITL approval step calls its action by
  the deprecated alias `process_data`; registering only the canonical name would have opted in
  **zero** live steps while looking like a working fix.
- **`bugs_open/236` deliberately records a refuted theory.** Read §5 before forming one.
- **Migration numbering races.** Four numbers were claimed by other sessions during one afternoon's
  work. Re-check `ls docs/agent_docs/sql_for_agents/ | grep -E '^3[5-9]'` immediately before you
  name a file, and never `--apply` — it takes every pending file, including other threads'.

## Cold-start pointers

- `rfc012_await_findings/HANDOFF_2026-08-10_continue_here.md` — the closed lane, both jobs done
- `bugs_open/236_HANDOFF_2026-08-09_…` — items 1 and 2's evidence; refuted theory in §5
- `architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md` §7 — item 1(b), premise now
  in doubt
- `architecture_review/RFC_019_…` §12 — the acceptance-evidence failures, worth reading as a method
  lesson before designing any verification above
- `docs024_key_docs_latest/LANDMINES.md`, `WRONG_CALLS.md` — both appended to this week
