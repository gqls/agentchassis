# 024 — Consolidation programme: what actually blocks 1,600 domains

**Status:** programme proposed, owner-directed 2026-07-27. Items sequenced, not
bundled. Two items have a live consumer waiting; one is a recommended **won't-do**.

> **A2 + A3 BUILT 2026-07-28.** `platform/mailer` (commit `1d747f5e8`, 8 tests) and
> `platform/httpguard` (commit `3632874d4`, 12 tests) are in the build, gofmt and
> vet clean, stdlib only, **no service rewired**. `platform/mailer` is the first
> SMTP anywhere in the built code. `httpguard.ClientIP` carries the `bugs_open/090`
> hardening and its regression test is **verified to fail** when the original
> defect is reintroduced.
> Council **APPROVED 2026-07-28 09:25, corr `6db59c8b-829f-4e4f-8273-511e1714d6ce`**
> (`complete_approved`, no veto). The two code commits **carry no
> `Council-Reviewed:` trailer and never will** — the verdict post-dates them and
> history is forward-only, so a trailer here would be a claim the commit could not
> have honoured. Recorded instead; `098` will list them as un-reviewed and that is
> the report being accurate, not a defect. A follow-up commit in this lane may
> carry the trailer.
>
> The FIRST submission (`c43a1166`) never reached the council: it died at
> `complete_invalid` on three schema-type errors of mine — `operation: "create"`
> (not in the allowlist; a new file is `add`), `grounded_in` as objects (it is
> `[]string`), `risks` as an array (it is a `string`). That failure is near
> invisible — no verdict row reads exactly like "still queued". Fixed forward for
> everyone: `097` now type-checks all three client-side (`be0f6aa16`), each check
> verified to reject the exact shape that failed.
> **Adoption is NOT done and is the next step**: wiring `httpguard` into
> `internal/tools-api/middleware/` touches the gauntlet workstream's service, which
> has `bugs_open/083` open against it — coordinate first.

**Why now.** The owner asked for a consolidation plan towards thousands of domains
after a near-miss: a design proposed a second public API on the island VM one day
before `cmd/tools-api` shipped there doing exactly that, multi-tool and multi-site.
Caught by the owner asking an integration question — no mechanism caught it.
Incident: `WRONG_CALLS.md` 2026-07-27. Mechanism findings: `bugs_open/108` and
`architecture_review/DECISIONS_open_for_owner_2026-07-26_architecture_seat.md` §8.

**The doctrine this programme applies** (offered for ratification):

> **Divergence is allowed when it is parameterised and forbidden when it is
> copied.** A second implementation is fine as a row in a table or a profile; it is
> not fine as a second copy of the code.

Generalises `vm_estate`'s *"merge the generator, not the trust boundary."*

---

## The scale arithmetic

`platform/orchestration/actions/registry.go`: **296** entries across **25**
category strings against **10** declared in its own header; `site` alone = 107.

**9 of the 296 exist for 2 of ~1,000 sites, and 5 shipped in one week**
(`med_*` ×4 for vetcomparison; `score_grippers`, `pull_report_requests`,
`verify_report_prose`, `create_report_page`, `emit_report_status_files` for
robot-hands). A per-site action is a per-site entry in Go source requiring a
rebuild and a redeploy — at 1,600 domains that couples site count to binary size.

**This is the only true scale blocker on the list.** Everything else below is
either a capability gap or tidiness.

---

## A. Blocks scale — do these

### A1. Per-site actions become config, not Go

The pattern that already does this correctly is `CHVerticalProfile`
(`platform/orchestration/actions/companies_house_vertical_profiles.go`) — its
header reads *"Add a new entry here when onboarding a new industry vertical"*: a
table, not a package per vertical.

> **CORRECTED 2026-07-28 — the "free first move" was not free, and the premise was
> never checked.** This section said: *"First move, free and self-evidencing:
> `med_export_json` sits ten lines above `directory_export_json` … It is the generic
> version of the one above it. Both registered, both live, nobody saw it. Merge and
> deprecate."* **Wrong.** Verified live before touching anything:
>
> - `med_export_json` is named by a **live, active** agent (`med-json-exporter`).
> - The two files are 634 vs 537 lines and share almost nothing: the med exporter
>   has **16 functions the generic one does not**, including
>   `filterMedExportProvenance` — a **fail-closed provenance gate** ("no price
>   without a source URL and capture date") added two commits ago in `f82f8b425`,
>   directly serving vetcomparison's P1 provenance work — plus letter-bucketing, a
>   medicine index and variant counting, which are a price-comparison data model,
>   not boilerplate.
>
> Merging would either break a live lane or drag all sixteen into the "generic"
> exporter, which would stop it being generic. **They share a purpose, not an
> implementation.** The claim came from a sub-agent audit line I repeated without
> opening either file.

**The genuine finding in this family, and it is the opposite shape.**
`firecrawl_map` and `med_map_urls` both call Firecrawl's `/map`. But:

| action | style | callers |
|---|---|---|
| `med_map_urls` | direct in-process HTTP, own key handling | **live agent `med-url-mapper`** |
| `firecrawl_map` | wraps `WebscrapeAction` (adapter, async) | **none — anywhere** |

`firecrawl_map` appears in `registry.go:1641` and **nowhere else**: no Go caller, no
seed, no agent config. So this is not "consolidate the bespoke one onto the generic
one" — the generic one is **dead registered code**, and the live one went direct
deliberately (the adapter path is async; these callers are synchronous).

The available action is therefore to **decide `firecrawl_map`'s fate** — wire it or
delete it — not to merge anything into it. That is small, safe, and real. Whoever
takes it should check `ListDeprecatedActions` semantics first; the registry already
supports `DeprecatedBy`.

> ### ⛔ CORRECTED 2026-07-28 — the scoring engine is a **WON'T-DO**, and the number that justified it was wrong
>
> This section said: *"`score_grippers` becomes a scoring engine with its rule
> table in `site_specs`"*, justified by *"9 of 296 registry entries serve 2 of
> ~1,000 sites"*, citing `CHVerticalProfile` as the config-table exemplar.
> **All three claims fail on inspection.** This is the third premise in this
> ticket adopted from an unverified sweep, after `med_export_json` and the
> health servers — the pattern is now recorded in `WRONG_CALLS.md`.
>
> **The exemplar is not a config table.** `companies_house_vertical_profiles.go`
> is `var chVerticalProfiles = map[string]*CHVerticalProfile{…}` — a **Go map
> compiled into the image**, with one populated entry and a commented-out
> template for the next. Onboarding a vertical through it costs a build and a
> roll. It demonstrates "one action, many verticals"; it demonstrates nothing
> about DB-driven config. *(The genuine DB-driven precedents are `voice_gate` on
> the `voice` site_specs aspect, `evidence_base`, `growth_config`, and
> `directory_export_json` via `scheduled_tasks.input_data`.)*
>
> **The count is 1, not 9.** Opening all nine: `pull_report_requests` selects its
> sites with `WHERE … deploy_config ? 'report_island'` — already fleet-generic and
> already per-site config-driven. `emit_report_status_files` is pure plumbing.
> `create_report_page` carries three display literals in ~480 lines of generic
> page persistence. `verify_report_prose` carried two, **and this session removed
> both**. Measured by grep, gripper/robot-hands mentions in code (not comments)
> are: `score_grippers` 41, `create_report_page` 3, `report_request_pull` 1,
> `verify_report_prose` **0**. The four `med_*` serve one *vertical*, which this
> ticket has already conceded is a data model rather than a duplicate.
> **One action is irreducibly single-site.**
>
> **And N=2 already exists and refutes the abstraction.** idea.uk is the
> platform's other live Tier-3 signature operation. Its scorer
> (`idea.uk/golang_files/engine.go`) is an **LLM-produced 1–5 rubric** —
> `Defensibility / Willingness / Buildability / Reuse / Durability / Risk`, gated
> on `Advances` and `Risk ≤ 2` — with no candidate index, no published-figure
> semantics, no units, no capacity/need ratio and no verdict ladder. Intersect it
> with the config this section proposed to extract (materials→μ, the 1.25 band,
> cycle-rate→safety tiers, the physics defaults): **the empty set.** Not low
> overlap — zero fields. A schema shaped around μ and headroom cannot hold "rate
> a business idea and drop anything in regulated-profession territory".
>
> **What actually generalises is the pipeline, not the scorer** — intake → pull →
> verify → prose → deterministic gate → render → deliver → status. Four of the
> five gripper actions are already in that generic layer, and
> `pull_report_requests` already serves site two with **zero changes** the moment
> an operator adds a `report_island` block to `sites.deploy_config`. A1 picked
> the one member of the family that is irreducibly domain-specific.
>
> **Recorded as won't-do rather than deferred**, deliberately: "generalise after
> the pilot" leaves a refuted idea on the schedule with a date attached, which is
> how it gets built by someone who does not reread this. Reopen only if a second
> site genuinely wants a *physics* scorer — and note that would be N=3, with two
> worked examples to abstract from instead of one.

*Guard against regression:* the registry has a parity test for *unregistered*
actions (`registry_parity_test.go`, from `bugs_open/017`) but nothing that notices
a near-duplicate registration. Nothing would have flagged `med_map_urls` beside
`firecrawl_map`.

### A2. An SMTP mailer that lives in the build

```
grep -rn "net/smtp" --include=*.go platform/ internal/ cmd/     →  nothing
```

**There is no mailer anywhere in the code we build and deploy.** The only working
one is in idea.uk's VM app in the docs tree — outside `go build`, untested by CI,
undeployable by the image pipeline. `send_notification`
(`basic_actions.go:134`) is not email; it produces a Kafka message.

Every "we email you a link" journey depends on this, including the gripper dossier
next. Promote it to `platform/mailer` **once**. Smallest item here and the only one
with a consumer already queued. Do it *before* the gripper island half, or it forks.

### A3. `platform/httpguard` — one limiter, one CORS policy, one intake guard

Today the public API's limiter is the **weakest** of the three that exist:
`internal/tools-api/middleware/ratelimit.go` (token bucket, per-pod, no
forged-`X-Forwarded-For` test) guards the only public endpoint, while idea.uk's
better one — banded per endpoint, returns retry-after, and has an explicit test
proving a forged `X-Forwarded-For` cannot escape it — is unreachable in the docs
tree. Four different CORS postures exist (tools-api DB-driven 403 vs auth-service
static allowlist that silently continues, plus two VM ones).

Fold in idea.uk's honeypot + timing gate (`service.go:359-400`), which is the only
abuse protection of its kind in the estate and which the gripper design currently
plans to **copy**.

Again: before the island half, not after.

---

## B. Cheap ride-along

**B1. Three HITL action pairs → one.** `await_approval`/`process_approval_decision`,
`create_approval_request`/`wait_for_approval_response`, and
`request_human_input`/`process_human_input_response` — 1,549 lines across three
files for one capability. The third is already the superset (confirmation /
questionnaire / review, skip conditions, field defaults). Consolidate onto it and
deprecate the other two pairs. Worth doing while already in the registry.

---

## C. Measure before acting

**C1. Six Firecrawl client constructions** — two adapter providers plus four
direct in-process HTTP calls (`vet_med_*` ×3, `refresh_product_specs`). The
`vet_med_*` headers say so in their own source: *"Follows the same direct-HTTP-call
pattern as ch_fetch_accounts_action.go"* — duplication by imitation, named as such.

Tidiness **if** they share retry/timeout config; a **cost** bug that scales with
domain count if they don't, because six clients means six independent spend paths
against one vendor. Measure first; the answer decides which list this belongs on.

---

## D. Recommended WON'T-DO — record it so it stops looking available

**D1. The eight `StartHealthServer` copies.** This looks like the tidiest,
most countable win on the list. It is a trap.

A sweep reported them as "8 byte-identical copies." Hashing the bodies gives **8
distinct hashes**, serving **1–3 endpoints each** — some use `http.HandleFunc` on
the default mux, some build their own, `internal/adapters/git/adapter.go:871`
serves three endpoints. And `platform/health/server.go` is not the shared version
of them: it uses gorilla and registers a different surface, and is imported by
`cmd/agent-chassis` alone.

So it is eight behavioural migrations against eight live binaries, on the liveness
path Kubernetes uses to decide whether to restart them, to save a few dozen lines
— for **zero benefit at any domain count**, because health endpoints do not scale
with site count. Close it as won't-do rather than leave it sitting there.

*(This entry exists because the wrong premise nearly made it the first task. See
`WRONG_CALLS.md` 2026-07-27: a sweep's figure carries no measurement date.)*

---

## E. The estate — already planned elsewhere, unblocked and free

`vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md` P1→P2. Its own
measurement: the two `setup.sh` copies **share 61 lines and differ on 614**, with
Defect A present in one fork only. P2 (`render_vm_config` + byte-for-byte diff
against the live box) is **read-only and free**, and proves the DB description is
complete before anything is automated. Owner already ruled pull-only (Q1).

Not this programme's to schedule — flagged so it is not duplicated.

---

## F. Unowned and needed

`features_open/015` (staged site maturity ladder) is the **stated methodology** for
fleet scale — named rungs, stepped reference examples, so a site climbs one rung at
a time against a worked example. It has **no workstream directory, no PLAN, no
owner**. It is the difference between "make every site as good as idea.uk" (a
cliff) and "move every site up one rung" (a staircase).

Flagged, not adopted — starting it inside this entry would be exactly the
divergence this programme exists to stop.

---

## Sequencing

Ship these as **separate commits**. Bundled, they make a 16-file commit that
`scripts/commit-scope-report.sh` will correctly print and nobody will correctly
read — which is the failure mode that script was written for.

Order: **A2 → A3** (both gate the gripper island half) → **A1 first move**
(`med_export_json`, free) → **B1** → **C1 measurement** → **A1 remainder** after
the pilot ships.

## What will bite us when we do

- **A1 remainder touches a pilot in flight.** Do not start it before the gripper
  E2E fixtures pass, or the pilot loses its evidence.
- **A2/A3 touch `tools-api`, which the gauntlet thread owns** and which has
  `bugs_open/083` open against its error handling. Coordinate; do not start a
  competing fix (`scripts/who-owns.py`).
- **Deprecating `med_export_json` needs a live check first** — some
  `agent_definitions` workflow may still name it. Query before deprecating.
- **The registry's category drift (25 vs 10 declared)** makes "does this already
  exist?" hard for humans as well as machines. Worth fixing alongside A1, but it is
  a rename across 296 entries — its own task, not a ride-along.
