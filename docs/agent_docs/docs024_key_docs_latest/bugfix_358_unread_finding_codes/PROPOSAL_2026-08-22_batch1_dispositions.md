# Batch 1 — proposed dispositions for 7 of the 32 undecided finding codes

**For the owner to ratify or overrule** (his ruling, 2026-08-22: the session proposes with the
evidence attached, he ratifies in batches). **Nothing here is applied** — every one of these codes
is still `unruled` in the registry. Ratify and I will change them in one commit.

Batch chosen by volume and by expiry, not alphabetically. Every row count is
**`[MEASURED]` 2026-08-22** over the retained window.

---

## The two things to read before the table

### 1. A code can be automatically consumed and its `agent_error_log` row STILL be unread

This is a distinction the registry cannot currently express, and batch 1 found a live case.

`component_validation_rejected` **does** have an automated reader: migration `563` makes the
component-creator's prompt branch on `last_error_code`, so a rejected component is regenerated with
a remedy specific to its failure class. That is exactly what `consumed` is supposed to mean.

**But the reader does not read `agent_error_log`.** It reads `site_work_items.retry_feedback`, a
dedicated column with a single writer (`store_generated_component_action.go:1587
recordRetryFeedback`), wired to the prompt by migration `564` as `current_item.last_error_code`.
The `agent_error_log` row is a *second, parallel* record that nothing consumes.

Marking it `consumed` would give it a `reader` file:line the checker happily verifies (563 does
contain the string), a 365-day clock, and a clean bill of health — while the row itself stays
unread for ever. **That is `bugs_open/358`'s own defect wearing a green badge**, and it is a
*different* failure from the one the registry already warns about (`STRUCTURAL_KEY_CARRY_MISS`'s
"consumed is not coverage", which is about the WRITER's blindness). This one is about the reader
reading a different sink.

**Proposed remedy, for ratification with the batch:** `consumed` gains a required
`reader_sink` field naming the table the reader actually selects from. If it is not
`agent_error_log`, the row is unread and the disposition must say so rather than imply otherwise.
Cheap, and it cannot be satisfied by typing — the checker can compare it against the reader file it
already opens.

### 2. Ruling anything `operational` now has a cost, by design

Migration `567` (live) gives `operational` codes the 30-day clock and everything else 365 days, and
the parity check added in `bc45567a0` asserts the two agree. So **ruling a code `operational` is
incomplete until it is also added to the sweep's short-retention list** — otherwise the next audit
run reports `retention_parity_missing`. That friction is deliberate: the disposition and the
retention it implies should not be able to drift apart. **No code in batch 1 is proposed as
`operational`, so no migration is needed to ratify this batch.**

---

## The proposals

| code | rows | what the writer actually records | proposed | the field that disposition requires |
|---|---|---|---|---|
| `CONTENT_LINK_REPAIR_DETAIL` | 405 | what the link-repair pass **did on a build that SUCCEEDED** — deliberately distinct from the blocker code because *"the two rows answer different questions"* (`validate_page_content.go:667`). Its stated purpose is to name WHICH of three paths repaired, so *"a row that cannot say which path acted cannot be used to spot the path that stopped acting"* (`:670`) | **human-evidence** | `why`: comparing the three origins over time is a longitudinal question and it now has the 365-day window rather than 30. ⚠ this is the code most worth a real reader — three paths, count by origin, and a path dropping to zero is the `bugs_open/097` drift it exists to catch. Proposing `human-evidence` records today's truth, not the ambition |
| `VALIDATION_ERROR_DROPPED` | 206 | a message that failed validation and was **dropped** — the row is the only trace the message ever existed. Written on a detached context precisely so a cancelled request cannot take the record with it (`validation_drop.go:96`). Two writers | **human-evidence** | `why`: it has a named hand-run consumer already — `bugs_open/034`'s `VERIFY_034_post_roll.sh:38` greps for exactly this code. Accepts the 365-day window |
| `PLAN_SECTION_NAME_DROPPED` | 140 | a section the planner proposed that resolves to no active component, carrying a `remedy` string that distinguishes "the planner invented it" from "this planner cannot read its own menu" (`component_name_resolver_menu.go:168`) | **human-evidence** | `why`: it has a **documented** hand-run path — `site_db_actions.go:429` tells the reader in prose to query `agent_error_log` with this exact code. The strongest human-evidence case in the batch |
| `component_validation_rejected` | 101 | a component rejected by validation. **Its code value IS automatically consumed — from another sink.** See §1 | **human-evidence** (NOT `consumed`) | `why`: naming the parallel channel (`site_work_items.retry_feedback` → 563's prompt branch) and stating that the `agent_error_log` row itself has no reader. Revisit if `reader_sink` is adopted |
| `ARCHIVED_PAGE_DEPLOY_REFUSED` | 80 | a deploy refused because the page is archived, written **before** the return *"so a refusal is countable even though this action dispatches nothing further"* (`git_deployer_actions.go:100`). Two writers | **human-evidence** | `why`: the writer's own stated purpose is that refusals be COUNTABLE; nothing counts them today, and the honest record of that is `human-evidence` plus the 365-day window, not a `consumed` claim |
| `CONTENT_CLAIMS_FLOOR_DETAIL` | 61 | claims-guard detail on the SAVE path, deliberately distinct from the gate's code so *"which path caught this"* stays answerable — the same `097` rationale as the link-repair code (`save_sections_claims_guard.go:97`) | **human-evidence** | `why`: same longitudinal argument as `CONTENT_LINK_REPAIR_DETAIL`, and the two should be ruled together since they are one design |
| `TRUNCATION_DEGRADED_REVIEW` | 41 | a council seat's opinion damaged before the verdict counted it. **Settled by Task A** | **human-evidence** | `why`: the population is extracted and committed; the condition has been clean for 20 days with the mechanism proven live. ⚠ carry the Task A correction on the entry — **40 of its 41 rows are NOT truncation**, and the name asserts a mechanism the rows do not carry |

---

## What is NOT proposed, and why

- **No `consumed` in this batch.** Nothing automated selects any of these seven from
  `agent_error_log` and acts. That is not an impression — **0 scheduled tasks select on
  `error_code` at all** (`SELECT name FROM scheduled_tasks WHERE pre_query LIKE '%error_code%'`
  → 0 rows, 2026-08-22), and every in-process reader already registered belongs to the binary that
  wrote the row.
- **No `instrumented`.** That disposition needs an owner document and a `review_by` date that
  expires. None of these seven has an owning lane doing a time-boxed measurement; inventing one to
  fill the field would be exactly the "satisfiable by typing" failure the registry is built against.
- **Six of seven land on the same disposition, which is worth being suspicious of.** It is not
  laziness: `human-evidence` is what "a human wrote this deliberately, and nothing reads it" means,
  and that is the true state of this table for almost everything in it. The interesting content is
  in the `why` fields, which differ, and in §1 — not in the disposition column.

## If you ratify

One commit changes seven `disposition` fields and adds their required `why`. It needs no API
access, no migration, and no rebuild — the registry is a file the check reads. `25 unruled` after.
