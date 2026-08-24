# HANDOFF — `bugfix_367_router_remit`, 2026-08-24

> # ✅✅ LANE CLOSED 2026-08-24. `bugs_open/367` → `bugs_closed/367`.
> Fixed, live, survived the **v1.0.1334** roll, and **observed on a real item in production**.
> **Nothing here needs picking up.** Read this file for the traps in §5 and the reasoning, not
> for state.
>
> **Two things are open and named, and neither belongs to this lane:**
> `bugs_open/375` (filed by this lane — architecture-scope, unowned) and `bugs_open/333`
> (owned by the 277 lane; contributed into, not competed with).

---

## 1. What this lane fixed, in one paragraph

`required-fields-missing-handler` is a triage **router**: it repairs nothing, it classifies each
`required_fields_missing` item and then converts, parks, or closes it. It resolved the offending
component with `WHERE pc.build_status = 'deployed'` — mirroring the post-deploy check it was
built for (`bugs_closed/277`). When `bugs_closed/342` added a **second** producer at render time,
whose entire stated purpose is reaching components that never reach a deployed row, every finding
it filed resolved nothing, fell to route `stale`, and was **closed `complete` with no error**. A
true finding scored as a success — and quieter than the three-failed-attempts behaviour it
replaced.

**The rule now installed:** *a disposer may close only on POSITIVE evidence of absence.* The page
row is gone, the component is locked (277's accept-as-is resolution), or a
`build_status='removed'` row is actually at that slot. A failed lookup — or a real but
non-deployed target — routes to `park_not_dispatchable` at `needs_human_review`, holding its
dedup key.

---

## 2. State — what is live, and how to re-check it in one command

| thing | state |
|---|---|
| migration `574` | **APPLIED 2026-08-23**, config only, live on apply |
| migration `576` (tomb empty-slot guard) | **APPLIED 2026-08-23**, from a council objection |
| council `d48c0a89-9ff8-4286-bfe9-2690dc13d5bc` | **APPROVED round 1**, 14 advisory objections, none high |
| survived roll | **yes** — v1.0.1334 (2026-08-24 15:39Z); the `_VERIFY` sidecar passes after it |
| observed on a real item | **yes** — orchestration `e2a6bb94`, 2026-08-24 16:08Z |
| `bugs_open/367` | **CLOSED** → `bugs_closed/367` |

**The one command that answers "is it still right?"** — read-only, re-runnable, and **proven
non-vacuous** (run against a deliberately reverted row it fails loudly):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/574_required_fields_router_stops_closing_what_it_cannot_resolve_VERIFY.sql
```

⚠ **Its two POSITIVE CONTROLS are the point, not the first check.** A change that simply stopped
this router closing anything would satisfy the defect case and break a real route. C2 (retired
component) and C3 (page gone) must **still close**.

### ⚠ One thing changed UNDER this lane after it closed (2026-08-24)

`bugs_open/333` shipped its owned-page door (`6ab0b3434`, **INERT until the next chassis
roll**). A finding whose page is `rebuild_policy='owned'` and whose handler declares
`refuse_owned_page` is now parked **at creation** — `status='deferred'`, `handler_agent=''`,
`error` prefixed `OWNED_PAGE_GUARD`. **`page-build-handler` declares it** (verified live
2026-08-24 — it is the only agent that does), and this router's `file_rewrite`/`file_recreate`
go through `create_work_item` → `writeWorkItem`, so **this router's conversions are inside that
door's coverage.**

Three things follow, and they matter to whoever watches this router next:

1. **The `28 of 31 failed` figure quoted throughout this lane is now FROZEN.** New owned-page
   conversions will be `deferred`, not `failed`. **A health query counting `failed` on this
   item type will show an improvement that neither lane made** — the finding is still not
   repaired, it is merely no longer misfiled. Add a
   `status='deferred' AND error LIKE 'OWNED_PAGE_GUARD:%'` arm.
2. ⚠ **That `error` prefix now has TWO producers** — the door (on `deferred` rows) and the
   handler (on refused ones). A census meaning *refusals* needs `AND status <> 'deferred'`.
3. **`close_converted`'s note is rosier than the truth**, and now more visibly so. It closes
   the original `complete`/`converted` saying *"repair filed as a follow-on item at
   page-build-handler"*, and after the roll that follow-on may be born undispatchable.
   **This is not a regression** — the pre-door behaviour was a conversion that FAILED, and a
   `deferred` row is *more* legible (the roadmap sweep reads it) — but the ORIGINAL's note
   reads like a dispatched repair either way. `create_work_item`'s step result now carries
   `row_status` and `owned_page_parked`; `row_status` is the field that tells the truth.
   `handler_agent` in that result still reports what the config ASKED for.

The 333 lane also confirmed, independently, that **this lane's volume prediction held**: nothing
new was filed at `page-build-handler` on an owned page since 08-19, so the 28 are all pre-`574`
and nothing rose. That was the stated failure signal for `574`, and it did not fire.

They **declined** the `triage.component_id` item this lane handed them, with reasons recorded in
`bugs_open/333`'s fix section as a named non-scope item — it re-opens `CQ-023`'s
council-settled key design for a population of 3 items that no longer reaches the broken shape.
Correct call, and it is a decision on the record rather than a silence.

---

## 3. The proof, because it is the most reusable thing here

Two orchestration records, same item, same component, one day apart — each other's control:

| orchestration | when | `route` | `target_state` | `component_id` | `html_len` |
|---|---|---|---|---|---|
| `ab2cf74e` | 08-23 17:09Z | **`stale`** → closed `complete` | *(none)* | *(empty)* | 0 |
| `e2a6bb94` | 08-24 16:08Z | **`target_not_dispatchable`** → parked | `pending` | `0a1498b3…` | **9220** |

Lifecycle observed live: `triaged` → `claimed` → `needs_human_review` in ~100 seconds. And the
parked row **HOLDS its dedup key** (1 non-terminal row on the key) where the close had released
it — the anti-churn property as data rather than intent.

Before that, and before anything was applied: the whole population (all 65 items ever filed of
this type) was re-classified under the old and new queries in one plpgsql loop — **exactly one
route changed** — and apply-then-rollback returned `default_config` **byte-identical**.

---

## 4. What this did NOT do — say it plainly, it is the claim most likely to drift

**The population is visible and HONEST, not REPAIRED.** It parks for a human with three named
resolutions on the row (deploy/rebuild the component, lock it, or retire it). Repair needs
`bugs_open/333` *plus* a producer that writes the convert arm's read-set. **Do not let anyone
write that 367 made render-time findings repairable.**

Also deliberately not done, each with a reason recorded in `PLAN_2026-08-23_367_router_remit.md`:
no Go change and no chassis build; no re-keying of `file_rewrite` to `triage.component_id`; no
verifier registration (`CQ-023` warns one would fail-close the router's own `converted` arm); no
component-axis predicate family — the 19 other hand-typed `pc.build_status='deployed'` reads are
all **producers**, whose failure mode is under-detection.

---

## 5. Traps worth carrying out of this lane

- **Read the router's SQL from the LIVE ROW, never the seed.** `410` was edited in place three
  times and its comments say "v3"; the live row is **`version = 1`**. A subagent reported "v3"
  from the file and was wrong about the row.
- **`pages.sections` is an array of TEXT**, not objects. `s->>'name'` returns NULL for every
  element, and a census filtering on it reads a confident **0 of 0** while 746 rows sit there.
  The real figure — 336 of 2,160 named slots have no component row — is what makes
  "no longer exists" a false claim.
- **`orchestration_states` retains ~2 days**, but `min(created_at)` reads five weeks back
  because the purge **exempts** stuck `CANCELLED` rows. A minimum over a purged table tells you
  what the purge exempts, not how far back it keeps.
- **Never quote a count off a query carrying a `LIMIT`.** I told a reviewer "all 8 conversions
  failed"; the real census was 31 (28 failed). The direction held, which is what makes it hard
  to notice.
- **`snapshot_agent` is overloaded** — a bare literal gives `function snapshot_agent(unknown) is
  not unique`. And `to_jsonb()` over adjacent string literals is `unknown`; cast with `::text`.
- **Test a migration without touching production, in three levels**: parse + its own verify
  (`sed 's/^COMMIT;$/ROLLBACK;/'`); behaviour against the *patched* row inside a transaction;
  and apply-then-rollback requiring `default_config` byte-identical. The RUNBOOK has all three.
- **A closed bug's referrers keep being obeyed.** On the day you close, grep `bugs_open/NNN` and
  retract the ones asserting it is live. Four documents here still called 367 open.
- **The window between appending to a shared doc and committing it belongs to everybody.** My
  three `WRONG_CALLS` entries were swept into another lane's commit (`bb1e144b5`) in that gap —
  and the very next commit on that file is a different lane writing up the same thing happening
  to them, the same afternoon.

⚠ **All of these are in `LANDMINES.md` (3 entries), `WRONG_CALLS.md` (3 entries) and `016b` §9
already.** They are repeated here only because a cold start reads the handoff first.

---

## 6. Residuals — real, and none of them this lane's to take

1. **`bugs_open/375`** *(filed by this lane, UNOWNED)* — `update_work_item_status` stamps
   `complete` without ever consulting the verifier framework. That is **why** the silent close
   was reachable at all. Architecture-scope under the 2026-07-29 ruling, so deliberately filed
   rather than fixed inside a bug patch. **Its blast radius is explicitly NOT established** —
   §3 of the file says so and gives the query. Do not size it from 367, which is one router
   found by accident.
2. **Seed `410` can silently revert this.** It is an `ON CONFLICT DO UPDATE` whole-config seed
   and its verify block asserts branch wiring and park statuses but **never the resolution
   predicate**, so a hand re-run reverts `574`/`576` and passes its own checks. It carries a
   header pointer, which is mitigation, not a control. The honest fix — 410 asserting the
   predicate it itself writes — belongs to whoever next touches that file. The `_VERIFY` sidecar
   detects the reversion loudly, which is the interim.
3. **`file_rewrite`'s read-set is WIDER than `classify`'s**, and that gap is *unexercised, not
   closed*. It needs `spec.component_id`, `.page_id`, `.component_function`, `.reason`; the
   render-time producer writes none of them (0 of 3 items vs 62 of 62 for the post-deploy
   producer), and both `spec_paths` and `item_key_suffix_field` are deliberate hard errors. A
   **third producer** that satisfies the classifier and omits those will route correctly and die
   one step later. Masked today only because such targets park first.
4. **`bugs_open/333`** *(OWNED by the 277 lane — contribute, do not compete)* — until owned-page
   routing is fixed, this population can only ever park.

---

## 7. Where everything is

| what | where |
|---|---|
| the bug | `bugs_closed/367_HANDOFF_2026-08-23_…the_render_time_producer_exists_to_reach.md` |
| the fix | `docs/agent_docs/sql_for_agents/574_required_fields_router_stops_closing_what_it_cannot_resolve.sql` (+ `_ROLLBACK`, `_VERIFY`) |
| the guard | `docs/agent_docs/sql_for_agents/576_tomb_cte_cannot_match_on_an_empty_slot.sql` (+ `_ROLLBACK`) |
| the canary | `…/bugfix_367_router_remit/CANARY_2026-08-24_reopen_the_wrongly_closed_item.sql` (guarded, re-runnable) |
| council submission | `…/bugfix_367_router_remit/submission_367_router_remit.json` |
| the decision, travelling | `doc_notes`: `subject_type='decision'`, `subject_key='required-fields-missing-handler:disposer_closes_only_on_positive_evidence'` |
| the register | `docs026_concept_register/register/content-quality.md` → **CQ-023** |
| the transferable rule | `016b` §9 — *"a DISPOSER may close only on POSITIVE evidence of absence"* |
| the traps | `LANDMINES.md` ×3, `WRONG_CALLS.md` ×3 |
| this lane | `docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/` — the standing five, plus this handoff |

**Commits** (all on `087_towards_multiple_domains`): `4111bcb8f` the fix · `8caa995ec` the
standing five · `21b88a3d0` the record + landmines + 016b + CQ-023 · `fe0df0cd0` bug 375 ·
`d51f02a4b` the 333 contribution · `598612438` mig 576 + `_VERIFY` · `7071fc99f` the council
round + the `bugs_closed/032` precedent · `313421727` the close · `89adae0b1` the retractions.
