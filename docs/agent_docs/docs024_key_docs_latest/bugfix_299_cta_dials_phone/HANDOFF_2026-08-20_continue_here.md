> # ⚠ LANE CLOSED 2026-08-20 — this file is HISTORY, not instructions.
> **Read `OWNER_ATTENTION_2026-08-20.md` and `SUMMARY_2026-08-20_lane_closed.md` instead.**
> The one live thread is housekeeping (a roll picks up the capability-table prune; one-line
> interim in the RUNBOOK) and it belongs to whoever runs the next release, not to this lane.
>
> # ⚠ SUPERSEDED 2026-08-20 (later the same day) — most of this is DONE.
> **`bugs_open/299` is CLOSED** (moved to `bugs_closed/`, fixed AND live at the served page:
> `href="/faq.html"` with copy naming the FAQ). 477 APPLIED; the canary SATISFIED;
> RFC_040 stages 1-2 BUILT, council-APPROVED (`06ce6aab`), LIVE on v1.0.1319, and a leak
> in it found and fixed. **What actually remains is §5 at the bottom of this file — read
> that, not the middle.** The rest is kept as the record of how it went.

# HANDOFF 2026-08-20 — continue here (bugs_open/299 slug `home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead` + `bugs_open/312` + RFC_040)

> **Refer to 299 BY SLUG.** `bugs_closed/299` is the unrelated skipped-render-audit case.

**Read this first, then `NOTES_cta_dials_phone.md` from its 2026-08-19 entries down.**
Everything below is committed. Nothing is uncommitted anywhere.

---

## 1. What is DONE and PROVEN (do not redo any of it)

| thing | state | evidence |
|---|---|---|
| Go fix (keeps, normaliser, detector) | **LIVE** on chassis **v1.0.1317** | capability probe, both pods, negative control absent |
| migration **475** (arm `cta_nonpage_destination`) | **APPLIED** 2026-08-19 20:18Z, ledger-recorded | completeness-discovery-agent checks 43 → 44 |
| migration **476** (arm destination stamp) | **APPLIED** 2026-08-19 20:20Z, ledger-recorded | `stamp_cta_destination_guidance = true` |
| migration **477** (the wiring / `bugs_open/312`) | **APPLIED** 2026-08-20 **07:17:41Z**, ledger-recorded | path now `resolved_links.response.sections_ready` |
| `bugs_open/312` itself | **FIXED, PROVEN AT THE ARTEFACT** | before: 41 runs, 33 minted titles, **0 survived**; after: 4 runs, 4 minted, **4 survived**, all byte-identical |
| the detector actually works | **PROVEN LIVE** | induced run corr `ee07fd81` COMPLETED, filed 6 true findings incl. 299's own button |

**The owner's decisions, received 2026-08-20 — all four:**
1. Apply 477 with leopardessconsulting.co.uk watched. ✅ done.
2. **The phone button was NOT intentional**, but may be used to verify the fix.
3. `+44 7934 524 911` **CONFIRMED** ⇒ `tel:+447934524911`.
4. Build **RFC_040's small half** as recommended (scope ruling now written into RFC §0).

---

## 2. What is NOT done — in priority order

### (a) The canary is ARMED but NOT SATISFIED — this is the one open safety item

477 is live fleet-wide. The four builds since were all `webdesign.co.uk` tool-guide pages, and
**that site carries zero authored contact/tel CTAs**, so the KEEP branches were never exercised
post-477. A `Monitor` was watching for the first build on a keep-carrying site; **it dies with
that session — re-arm or just run the query.**

Sites that carry the CTAs the keeps protect (censused 2026-08-20):
`webdesign.uk 10 · leopardessconsulting.co.uk 7 · ai-agent-orchestration.com 5 ·
gaswholesalers.com 3 · fundamentallyai.com 1 · robot-hands.com 1`

```sql
-- has a keep-carrying site built since 477?
SELECT os.created_at, os.collected_data->'input_data'->>'domain' AS dom,
       os.collected_data->'input_data'->'current_page'->>'name' AS page
FROM orchestration_states os
WHERE os.collected_data ? 'resolved_links'
  AND os.created_at > '2026-08-20 07:17:41.761737+00'
  AND os.collected_data->'input_data'->>'domain' IN
      ('webdesign.uk','leopardessconsulting.co.uk','ai-agent-orchestration.com',
       'gaswholesalers.com','fundamentallyai.com','robot-hands.com');
```
When one appears: diff its CTA urls against the **7-row leopardess baseline recorded inside
migration 477 itself** (and the full baseline in `RUNBOOK`, canary section). Survival of the
authored `/contact.html` values is the control that the keeps, not luck, made 477 safe.

**⚠ Do NOT force it with the `page-rebuild` agent.** It selects on
`build_status='needs_rebuild'` and leopardess already has **11 pages** in that state from
another lane's queue — firing it rebuilds all 11 live client pages, not one canary.

**What is already established (so this is a confirmation, not an open risk):** the last real
build of webdesign.uk `index` (2026-08-19 10:12, pre-477) shows the resolver's own output as
`primary_cta_url = /contact.html` and `secondary_cta_url = tel:+447934524911` with
`secondary_cta_target_title = "a phone call to +44 (0) 7934 524 911"` — the stored value was
the **unnormalised** `tel:+44 (0) 7934 524 911`, so KEEP #3 demonstrably fired, kept and
normalised on the motivating page. The keeps were right all along; 477 stopped their answer
being discarded.

### (b) `bugs_open/299` is still OPEN — the served page is unchanged

The bar is the served page. webdesign.uk `index` has not rebuilt since 477, so the button still
dials the phone. Two steps, **in this order**:

1. **Verify** — let/make `index` rebuild through the normal pipeline (**never hand-build**,
   owner ruling 2026-08-04). Expect `secondary_cta_url` → `tel:+447934524911` (normalised) with
   its target title, and the copy regenerated to name a phone call, because 476 now feeds
   "Destination (fixed): …" to the writer.
2. **Then fix it properly** — the owner has said the phone button was **not intentional**. So
   set the stored `secondary_cta_url` to `/tools/website-brief-starter/index.html` (a page URL,
   so KEEP #2 protects it thereafter) and let a rebuild write matching copy. The resolver
   already computes exactly that destination for this field.

**Verification, and the false-pass trap:** assert on the anchor whose TEXT names a destination.
Nav and footer already link the tool correctly, so a page-wide grep for the URL passes today
while the button stays broken.
```bash
curl -s https://preview.webdesign.uk/index.html | grep -A3 'cta-btn-secondary'
```

⚠ **Coordinate:** the `webdesign_uk_build_service` lane is ACTIVE on this site and rewriting it.
Check with them before rebuilding `index`.

### (c) The contact page's undialable number — owner has now answered

`contact/hero` carries `tel:+4407934524911` (collapsed trunk; `NormalizeTelHref` deliberately
REFUSES to guess it, and `check_cta_nonpage` filed it as `cta_tel_malformed`). The owner has
**confirmed the intended number is +44 7934 524 911**, so this is now a one-line fix:
set it to `tel:+447934524911`. Work item key:
`cta_nonpage:contact:hero:tel:+4407934524911` (status `needs_human_review`).

### (d) RFC_040 stage 1-2 — RATIFIED, scoped, NOT yet written

**Owner ratified 2026-08-20, scoped to stages 1-2 ONLY.** RFC §0 records the ruling and, more
importantly, **what is NOT authorised: do not build `assert_live_capability()` or any migration
calling it.** It waits for a second real caller — a fail-closed helper with one caller is a
mechanism that rots unexercised, which is the estate's own stated failure mode.

**Design decisions already made this session (do not re-derive):**

- **New package `platform/buildcapability`**, deliberately DUMB — it writes what it is given and
  imports nothing but `pkg/buildinfo` and `database/sql`. The caller supplies the lists. This is
  what avoids the import-cycle question entirely (`actions` imports `discovery_checks`;
  `discovery_checks` imports neither, and `agentbase` imports neither).
  ```go
  type Set struct { Kind string; Names []string }   // Kind: "discovery_check" | "action"
  func Record(ctx context.Context, db *sql.DB, service, podName string, sets ...Set) error
  ```
- **Call site: `cmd/agent-chassis/main.go`** — it is the one place that can already see both
  registries and has config in hand. `agentbase` was considered and REJECTED: it imports neither
  registry, so putting it there means plumbing the lists through anyway.
- **The two enumerations already exist** and cost nothing:
  `checks.Names()` (`discovery_checks/registry.go:171`) and
  `actions.ListActions()` (`actions/registry.go:2124`, excludes deprecated).
- **BEST-EFFORT, never fatal.** A capability registry that can stop a service starting is a
  worse bargain than the problem it solves. Log and continue on any error. RFC §4 says so.
- **Table** (next free migration number — re-derive it, do not trust this line):
  ```sql
  CREATE TABLE service_binary_capabilities (
    service text NOT NULL, pod_name text NOT NULL,
    git_commit text NOT NULL, image_tag text NOT NULL DEFAULT '',
    kind text NOT NULL, name text NOT NULL,
    started_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (service, pod_name, kind, name));
  ```
  Write = delete this pod's own rows, then insert, in one transaction.
- **Acceptance evidence owed** (RFC §7 items 1 and 3 only; 2 and 4 are deferred with the
  helper): rows for every live pod with `image_tag` matching
  `kubectl get deploy -o jsonpath`, **with a negative control**; and a stale-row proof (kill a
  pod, confirm its rows stop counting).
- ⚠ **`last_seen_at` must actually be refreshed** or a dead pod's rows vouch for a binary that
  is no longer running — that is the same class of error this whole RFC exists to end, and it
  must not be reintroduced by its own mechanism. If no cheap heartbeat carrier is found, say so
  loudly in the register entry rather than shipping a silent staleness hole.

### (e) Council submission — NOW REQUIRED, and the scope rule changed under us

**CLAUDE.md was updated 2026-08-19 (`bugs_open/314`): the council gate's scope now INCLUDES
appliable migrations** (`docs/agent_docs/sql_for_agents/NNN_name.sql`), not just
`platform/`+`internal/`+`pkg/`. Scope is single-sourced in `scripts/council-scope.sh`; test
admission for free with `DRY_RUN=1`.

So the RFC_040 implementation (Go package + call site + migration) should go through
`097_TRIGGER_council_review_v1.sh`. The 299 lane's existing verdict
(`Council-Reviewed: 1f1fecc9-3502-4757-8929-fd173fca2dc6`, APPROVED round 2) covers the CTA work
and **does not** cover this.

---

## 3. Traps this session paid for — do not re-learn them

- **A roll invalidates your verification.** The fleet rolled twice during this lane
  (v1.0.1316 → v1.0.1317). Re-probe after every roll; do not reuse yesterday's answer.
- **You cannot check "is my commit live" the documented way.** The provenance line is a STARTUP
  line (gone from a *full* `kubectl logs` 3h after a roll) and the binary carries ONE stamp, not
  its ancestry — so grepping for your own commit says *absent* for a binary that has it.
  **Probe the CAPABILITY, on every pod, with a control that must fail.** `LANDMINES.md` +
  RFC_040 §1.3.
- **`--apply` on the migration runner is not available to you** — 25+ other lanes' files are
  pending, several with drifted pre-gates. Apply out of band with `psql -f` semantics, then
  `--record-only <file> --note "<what you verified>"`.
- **Every migration touching `agent_definitions` owes `snapshot_agent(...)`** — 475, 476 and 477
  all needed it added; a bespoke `_backup_NNN` table does not satisfy the README rule.
- **Cast the smallest structure your claim is about.** A `::text LIKE` over the containing jsonb
  blob returned the exact opposite answer on the discard control (WRONG_CALLS 2026-08-19).
- **A red test may be another session's half-landed edit.** Check `git status --short -- '*.go'`
  and re-run before bisecting; it happened twice to this lane, once in each direction.

## 4. Where everything lives

- Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_299_cta_dials_phone/`
  (`NOTES` = technical log, `RUNBOOK` = the commands with their gotchas, `PLAN` = decisions +
  corrections, `README_where_we_are` = the owner's plain-prose log, `SUMMARY_2026-08-19` =
  the milestone read-out)
- Bugs: `bugs_open/299_HANDOFF_2026-08-17_home_page_cta_names_the_brief_starter…md`,
  `bugs_open/312_HANDOFF_2026-08-18_select_sections_discards…md`
- RFC: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_040_a_migration_cannot_ask_the_binary_what_it_can_do.md`
- Register: `docs026_concept_register/register/link-management.md` (LNK-033, LNK-034 — both
  status lines corrected 2026-08-19 from "awaiting roll" to LIVE)
- Migrations: `docs/agent_docs/sql_for_agents/475_…`, `476_…`, `477_…` (all released, applied,
  ledger-recorded; each carries its own discharge record and rollback recipe)

---

# 5. WHAT ACTUALLY REMAINS (written 2026-08-20, after everything above was done)

1. **`service_binary_capabilities` regrows until the next fleet roll.** The retention prune is
   committed but is Go, so it is inert until a chassis image carries it (current live:
   **v1.0.1319**, which has the writer but NOT the prune). Growth ~20k rows/hour. Interim, one
   line, in the lane RUNBOOK:
   `DELETE FROM service_binary_capabilities WHERE last_seen_at < now() - interval '2 hours';`
   After the next roll, confirm it self-limits: `count(DISTINCT pod_name)` should track live
   pods rather than climbing for ever.
2. **`bugs_open/312` stays OPEN on its own merits.** Its candidate 1 (the config repoint) is
   done and proven — that is what closed 299. Candidates **2 (a loud fallback)** and **3 (a
   lockstep test binding the writer's configured path to the resolver's actual response shape)**
   are unbuilt. That seam has now failed silently in BOTH directions twice (LNK-013/LNK-014,
   then this), so the tripwire is earned, not speculative.
3. **Two more instances of 299's class on webdesign.uk**, both filed by the new detector and
   sitting in `needs_human_review`: `faq/hero` ("See what you get for it" → a phone) and
   `how-it-works/call-to-action` ("Still deciding? The FAQ page covers the full terms…" → a
   phone). Those tel: hrefs are believed **genuine**, so the fix is the COPY, not the href —
   which is exactly what migration 476's destination stamp now feeds the writer. **They are the
   first real test of the stamp**; check `llm_call_log.prompt_rendered` for a value-shaped
   `Destination (fixed):` on their next rebuild (baseline before 476: 0 of 182 prompts).
4. **`how-it-works/call-to-action`** still carries the unnormalised `tel:+44 (0) 7934 524 911`.
   No action — it self-heals through the keep on its next rerender. Watch it as free evidence.
5. **The council's remaining advisories on `06ce6aab`** (APPROVED, 3 advisories, none high) —
   the two `editquality` ones about the sketch not showing the migration header and the Touch
   doc comment are **already false of the shipped files** (both carry them; the sketch was an
   excerpt). The `reuse_agent` note asking why `agent_definitions.capabilities` is the wrong
   home is worth one line in BLD-023 if anyone touches it: that column is per-agent-TYPE
   DECLARED config; this table is per-POD OBSERVED runtime state. Not yet written down.

## Traps this lane paid for — the full set
See §3 above, plus the two learned after it was written:
- **A page-rerender needs BOTH `spec.reason` and `spec.page_name`**, and all three failure modes
  report success while publishing the old page (`LANDMINES.md`, and the working envelope is in
  the RUNBOOK).
- **A pod COUNT is not a pod START RATE.** The chassis binary runs as ephemeral per-job pods
  (~52 starts/hour); `kubectl get pods` showed 5 and I nearly retired a correct council
  objection with it (`WRONG_CALLS.md`, 2026-08-20).
