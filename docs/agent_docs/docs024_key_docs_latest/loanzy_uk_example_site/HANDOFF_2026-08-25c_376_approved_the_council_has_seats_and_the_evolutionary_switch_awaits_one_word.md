# HANDOFF 2026-08-25c — 376 APPROVED (not applied), the acceptance council has its seats in the tree, and the evolutionary switch awaits one word

**Lane:** `loanzy_uk_example_site`. **Supersedes `HANDOFF_2026-08-25b_…the_council_already_exists.md`**
(same directory; its §1–§5 stand as history, its §6 to-do list is DONE except where marked).
**Read on a cold start:** this file → `PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md`
→ `architecture_review/RFC_056_…seats_are_the_benchmark.md` → `bugs_open/376` §11f.

## 1. State `[all as of 2026-08-25 evening]`

| thing | state |
|---|---|
| **`bugs_open/376` fix** (migration `618` + ROLLBACK) | **council APPROVED round 1**, corr `3d890adc-6f76-42c3-9eb2-20a76d7195f1`; committed `0d21ba356`/`9054fa7e8`; **NOT APPLIED** — the session harness declined the live write; the owner's one command is in §11f of the bug. After apply: the three §11e behavioural tests are owed. **No new greenfield domain until it ships.** |
| **`filing_mode: record`** on `write_audit_findings` | Go WRITTEN + 4 tests, commit `c440d5c5e`, **inert until a chassis roll**. Register IMP-056. |
| **Three mechanical seats** (`build_prerequisites`, `heading_promise`, `structure_floor`) | Go WRITTEN + 38 tests, same commit, registered via init(), **named in NO checks array — inert**. Register IMP-057. `verify-head-builds --with` ×10 OK vs HEAD `f3c1da996`. |
| **Migration `623_…_HOLD`** (bypass the 4 LLM seats, one edge) | written + rehearsed (apply and apply-then-rollback in rolled-back txns); **NOT applied — the owner's word is the gate** (PLAN §5 Phase 1). |
| **Migration `624_…_HOLD`** (acceptance + reader seats, record mode on 6 seats, 11 seat-failure rows, `check_seats_ran`) | written + rehearsed both ways; **NOT applied — needs the roll AND the RFC verdict AND `-v FILING_MODE_SHIPPED=<verified sha>`**. |
| **RFC_056** + council submission (Go + both HOLDs) | submitted, corr **`d1342f2a-fcfc-4fb1-b28d-cf3ee0ad492a`** — check the verdict before acting on 624. Addendum answers the vigilant lane's parked-verdict objection. |
| **`improvement-sweep`** | still `enabled=false`; the vigilant lane HOLDS the switch by agreement until the owner's word. |
| The LANDMINE ("`detected` is a QUEUE, not a shelf") | appended, synced to doc_notes; rode in the 375 lane's commit `4210764e9` (declared by them — cite the entry, not the commit). |

## 2. The one mechanism a newcomer needs

The "evolutionary aspect" = the four LLM seats' findings routed by `write_audit_findings` Rule 4
into regenerating handlers (`content_rewrite`/`needs_content_page`/`needs_content_planning` —
976/399/964 lifetime from the design-audit source alone). TWO doors dispatch them: the loop's own
`triage_findings` AND `detected-item-promoter` (900s, live since 08-15) — **26 LLM-audit rows were
promoted 08-20→08-24 with the sweep OFF; 0 sit at `detected` now**. So the sweep being off never
meant the rewrites were off, and `detected` is not a parking state. Record rows are `deferred` +
`handler ''` + provenance/`release_recipe` in spec; they self-clear via the seat's own
silence-retraction (RFC_056 addendum).

## 3. Next, in order

1. **Owner applies 618** (command in `bugs_open/376` §11f) → verify 15 steps + `check_exemplar_floor`
   → run the three §11e tests on the next research draws (one must be induced). Then the greenfield
   route unblocks.
2. **Owner's word on Phase 1** → apply `623_…_HOLD`, then `UPDATE scheduled_tasks SET enabled=true
   WHERE name='improvement-sweep'` → run PLAN §5's four verification queries within the hour
   (query 2 MUST be 0).
3. **Read the RFC_056 verdict** (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND
   body LIKE '%d1342f2a%'`). REVISE → resubmit with `RESUBMIT_CORR=d1342f2a-…`; the 4 committed Go
   files and 2 HOLD migrations are on the shared branch either way (review-after-the-fact, 07-29 ruling).
4. **After the next chassis roll**: prove the roll carries `c440d5c5e`
   (`git merge-base --is-ancestor c440d5c5e <stamp>`), then 624 becomes appliable (its probe
   demands the sha).
5. **garden-tools.uk: still HOLD** (owner). **Do not enable the offer-analyser oneshots** (spent,
   site-pinned, no stop condition — vigilant lane's measurement).

## 4. Falsifiers

- PLAN §5 query 2 non-zero after Phase 1 → an LLM seat is reachable by a route the plan did not see.
- A record row dispatched → one of the two doors changed (`write_audit_findings_filing_mode_test.go` names both).
- A record row still open after 3 audits at a new fingerprint with the finding gone → the
  silence-retraction is not matching record rows (RFC addendum falsifier).
- Bad renders continuing with the seats bypassed → look at `site-render-audit-rotation`
  (`contrast_failure` → `css-patch-agent`, 239 completions in 14 days) — the OTHER source, the 390 lane's.

---

## 5. SAME-EVENING UPDATE (post-REVISE, post-roll) — read this over §1 where they disagree

- **RFC_056 round 1: REVISE** (gating: editquality). Every objection triaged in RFC_056 **ADDENDUM 2**;
  the real defects are FIXED: HOLD migrations renumbered **619/620 → 623/624** (collision with the live
  `619_cta_bg_is_not_a_colour.sql`; `621`/`622` were also taken), `check_seats_ran` evaluator semantics
  now PROVEN (`conditional_branch_seats_gate_test.go`, 3 tests), a failed-seat sweep now stamps the
  audit ATTEMPT (cooldown) via new step `record_audit_attempt` without counting a PASS (48 steps),
  `heading_promise` uses `datahelpers.PageWantedLivePredicateFor` (was a raw `status='active'`),
  and 624 writes two `doc_notes` decision rows on apply. Resubmitted with `RESUBMIT_CORR=d1342f2a-…`.
- **The fleet rolled to v1.0.1339 at 19:07Z and it CARRIES `c440d5c5e`** — verified by the
  `service_binary_capabilities` route with a present-control and an absent-control
  (⚠ the CLAUDE.md `build provenance` log grep is REFUTED — the string exists nowhere in the Go
  source; bugs_open/395 lane's finding, confirmed by the vigilant lane). So `filing_mode` and the
  three checks are IN THE RUNNING BINARY today; 624 needs only the council verdict +
  `-v FILING_MODE_SHIPPED=a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`.
- **Gate 1c interaction (vigilant lane):** record rows never reach `complete`, so gate 1c never grades
  that producer while record mode is in force — correct on both sides, stated in ADDENDUM 2.
