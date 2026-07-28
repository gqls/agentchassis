# HANDOFF — experience register, 2026-07-28b (supersedes HANDOFF_2026-07-28)

**Read this, then `NOTES_experience_register.md` from the `2026-07-28e` entry onward.**
Owner rulings are in `PLAN_2026-07-24` §2 — do not relitigate them.
The previous handoff's §1–§3 (what this is, the harness-gap ranking, IMPLEMENTED≠SATISFIABLE)
still hold and are not repeated.

---

## 0. THE ONE THING BLOCKING EVERYTHING: the image is built and pushed, NOT rolled

`v1.0.1195` is **built and pushed** (image id `98ae7405f91b`, digest
`sha256:f9958349d2dd…`, a real rebuild — distinct id from 1194's `46bf8f4ec3b6`, 94 minutes
apart). **The fleet is still on v1.0.1194.** The roll was blocked by a permission gate at the end
of the session, not by anything technical.

```bash
# roll it (checks first that no council is in flight — a roll KILLS one):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT count(*) FROM orchestration_states WHERE status='EXECUTING_STEP' AND current_step LIKE 'review_%';"
# then, only when that returns 0:
sed -i 's/newTag:.*/newTag: v1.0.1195/' deployments/kustomize/services/agent-chassis/overlays/production/uk_001/kustomization.yaml
kubectl apply -k deployments/kustomize/services/agent-chassis/overlays/production/uk_001
```

Verify against the POD, never the tag:
```bash
kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "attribute_matches"'
# positive control, expect > 0. Negative control, expect 0:
kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "attribute_nonsense_xyz"'
```

`[PROVENANCE, stated]` The image is built from commit `b0c2da010`. Two later commits
(`992652730` and this handoff) add a doc comment, an unused-at-runtime classification table and a
test — **no behavioural change**, so 1195 is behaviourally complete. Do not pod-grep for
`experienceStaticConfirming`: it is referenced only from tests, so the linker may drop it.

## 1. What changed today, and what it bought

**Attribute assertion exists.** `attribute_absent` and `attribute_matches` at Tier 2, in the shape
six entries had already authored before any code existed. This was the register's largest measured
gap — 13 of 38 deferred clauses, across 9 of 9 entries — and it is the `no-inert-control`
invariant, the rule the register was built to enforce and could not check.

Proven against six live pages through the real exported evaluator: **6 checks PASS over 18 real
elements**, 1 correctly SKIPS, 1 FAILS on a genuinely served `href="#"`.

**`executable_checks` stopped lying.** The validator counted any check some tier implements; the
consumer runs Tier 2 only. The two were separate copies of one rule and had drifted, so the number
included exactly the checks that would never run — and that number is what migration 230's approval
constraint rests on. Now one shared function. Migration **264 applied and recorded** (comment-only).

**CC-001 is revised** and validates clean (2 executable, 8 deferred, 0 errors, and **zero**
unused-binding deferrals — every binding is now referenced by a check). **Not yet seeded**: the
live binary does not know `attribute_matches`, so the write path would refuse it. See §2.

## 2. Do these next, in order

### 2.1 Roll (§0), then seed CC-001 and resubmit its approval council
```bash
cd docs/agent_docs/docs024_key_docs_latest/experience_register
./240_TRIGGER_seed_harvested_entries_v1.sh CC-001     # precheck pod-greps the binary; it will refuse before the roll
RESUBMIT_CORR=ec91c7e4-1b2c-4329-be19-4231cdfa553b ./260_TRIGGER_experience_approval_v1.sh feed-driven-teaser-list
```
Expected after seeding: `executable_checks = 2`, 8 deferred. Both executable checks were proven to
PASS against the live page before the entry was written — `list_exists` and the new
`template_row_stays_hidden`.

### 2.2 `apply_experience_verdict` — still the gate that is shut
Unchanged from the last handoff and still the register's most important missing piece: the council
records a verdict and **nothing can write `status='approved'`**, so every entry is `draft`, every
fork `proposed`, and nothing reaches `verified`/`proven` except via `dry_run`. Build it to promote
only if decision is `approved` AND `unreadable == 0` AND the entry's `updated_at` is unchanged since
submission (migration 259's header states that last gap). Registered as the open item in **PLAN-046**.

### 2.3 Wire bind + verify
Unchanged from the last handoff §4.2, including the real vonc selectors, except one correction:
`feed_path` is **not** used by a Tier 2 check any more (`feed_loads` is now `tier: 4` — it is a
network claim, not a text match).

## 3. THE THING THAT NEEDS A HUMAN, not another submission

**The council gate's `architecture` seat ruled that these check types ARE architecture-scope:**

> *"new reserved keys added to experienceCheckTiers, a capability table with two read-sites across
> two systems. Per the plan's own cited 2026-07-28 seam ruling, a new key on a shared vocabulary is
> architecture-scope even when additive, small, well tested, and measured at zero current
> collision."*

The `guardian` seat separately **declined to veto**, calling it *"a constrained, well-fenced
addition rather than a redesign"*, and asked that tool-acceptance be acknowledged as **a live second
consumer that did not review this change**, not merely a measured one.

CLAUDE.md is explicit that a scope objection **is not answered by resubmitting with better
measurements** — it is a judgement about *how* a capability reached production. So: recorded in
**TL-031** where the change lives, and flagged here. **It needs an owner call.** Do not resubmit it.

Facts for whoever picks it up (all measured, not asserted):
- 78 criteria fences in `doc_plans` fleet-wide; **0** use either type; 1 `experience_patterns` row.
- `evaluateStaticCriteria` has **exactly two callers**: `check_tool_acceptance.go:212` and
  `exported_static_criteria.go:53` (reached only from `verify_site_experience_action.go:224`).
  Their criteria sources are `SELECT body FROM doc_plans` + `extractCriteriaFence`, and
  `experience_patterns.criteria_template` — so those two tables are the complete inventory.
- Both types previously hit the switch's `default:` and were SKIPPED.

**Council trail:** `99f2a5e6-e934-4ca1-addb-f16a29b38b0f`, two rounds.
Round 1: 11 reviewers, 8 approve, 3 object, **1 unreadable, and the REVISE was decided BY the
unreadable seat** — the harness, not the change. Round 2: 13 reviewers, **0 unreadable**, 10 approve,
3 object, no veto, decided by architecture on scope.

> **CORRECTED, and it was my error.** I first wrote that the architecture seat had never fired in
> its life and that this submission's explicit scope question drew it. Both halves are wrong.
> ANOTHER session seated it on `fix-proposer` + `council-gate` earlier the same evening (owner
> reversal of decision D9, register FIX-054), and its genuine first review — on a different
> submission, citing `bugs_closed/129` by name — landed at 22:11Z, one minute before my round-2
> report. Round 1 was submitted before that seating and drew 11 reviewers; round 2 after it and
> drew 13. **So a resubmission is NOT judged by the same panel as the original** — the roster is
> shared, mutable state that changed five times in 18 hours. Read the live seat list per round
> before explaining a verdict by anything about your own submission.

## 4. Open, and honest about it

- **`bugs_open/137`** — two mechanisms judge "is this control alive" in the same function and
  disagree about a specific live element: the `shell-dead-controls` sweep (with its page-wide
  `data-runtime-fill` exemption) versus `attribute_absent`. Filed at `reuse_agent`'s request.
  **My reading is that CC-001's clause is mis-tiered, and I have recorded that this reading is the
  one which makes my own red result disappear** — which is why it is filed for someone else rather
  than closed by me.
- **The approval council's REVISE on CC-001 is answered but unproven** — the revision validates
  locally; no council has seen it.
- **Three distinct harness gaps are now named separately** where one used to be counted three times:
  event-listener assertion, fault injection at the fetch boundary, and per-row conditionals tied to
  source data. Attribute assertion is done; these are the next tier of the §3.1 ranking.

## 5. Landmines added today (the previous handoff's §5 all still hold)

- **A doc comment is not an enforcement mechanism.** Tier 2's confirm/refute guarantee is now a
  classification every handled type must appear in, with the build failing otherwise
  (`TestEveryStaticCheckTypeIsClassified`, proved by induced fault). If you add a check type you
  will be made to classify it — that is the point, not an obstacle.
- **A validation harness that under-feeds the validator invents failures.** Running the nine
  entries through `ValidateExperienceCriteria` with only `criteria_template` + `binding_schema`
  reported 7 failures; they vanished when the `contract`/`states`/`data_contract` documents were
  passed as `extra`, because their placeholders also close. Caught before it was written up as a
  finding. Feed it everything the write path feeds it.
- **A Cyrillic homoglyph compiles.** A test constant was named `vonсTemplateRow` with U+0441; it
  builds, and `grep vonc` never finds it. `cat -A` on the identifier is the check.
- **A ratchet line covers a WORKSTREAM, not a mechanism.** Adding one concept-register entry made
  `102_CHECK_register_coverage.py` treat the whole workstream as covered while four callable
  mechanisms were still absent. Dropping the line obliges you to register the mechanisms (now
  PLAN-043/044/045/046 + TL-031).
- **A roll kills an in-flight council.** One was running when this session tried to deploy; the
  right move was to wait ~5 minutes, not to take the round off another session.
