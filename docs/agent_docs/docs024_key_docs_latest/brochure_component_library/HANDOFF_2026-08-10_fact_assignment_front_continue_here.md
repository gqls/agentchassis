# HANDOFF 2026-08-10 — fact-assignment front (bug 151 / RFC_016): cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-09c_…`.** Written ~15:00 BST, after chassis
**v1.0.1279** and after the Slice B council round came back **REVISE**.

Candidate **1b is BUILT, both halves** — (ii) is committed and **pod-verified
live**; (i) is seed `362`, written and `_HOLD`-held. The front's next move is
**answering the REVISE**, and §3 is the answering material: two of the six
objecting seats are already settled by measurements taken today, one is a real
code hole, and one is a real evidence defect that changes the round.

**This is ONE OF TWO fronts in this directory. Do not confuse them.**
- **This file = the fact-assignment front** (bug 151 candidate 1 + 1b, RFC_016,
  planner/writer prompts, seeds 327/328/329/330/333/**362**).
- **`HANDOFF_2026-08-09_sweep_front_continue_here.md` = the fundamentallyai
  sweep front** — different live thread, same site, same directory. Read it
  before touching the site.

Site id, needed everywhere: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## 1. Verified live state (all re-checked 2026-08-10 afternoon)

| thing | state |
|---|---|
| Chassis | **v1.0.1279**, both replicas (`agent-chassis-8496665bb8-f6svp`, `-sskxd`), rolled 13:42Z |
| **1b (ii) — the fact carry** | **LIVE + ARTEFACT-VERIFIED on BOTH replicas.** POS `FACT_CARRY_UNMATCHED_SECTION`→1, `section_facts_carried`→1, the remedy sentence→1; NEG `FACT_CARRY_MISSING_SECTION`→0; CTRL (pre-existing literal) →1 |
| **1b (i) — seed 362** | written, `_HOLD`-named, **NOT applied**. Dry-run proven; live row untouched (18,738 bytes, no `\| sections: `) |
| Slice B round | **REVISE**, corr **`a06ff850-aff6-4ed0-8e0a-93d57b0cbc45`**, 14 seats, 4 abstained, gated by `editquality` (HIGH). Report pinned in §3 |
| Seeds 328/330 | still `_HOLD` at HEAD, **and one of them now needs re-evidencing — see §3.4** |
| Migration 327 / seeds 329, 333 | applied, unchanged |
| Consumption of assignments | **still zero**, re-measured today with a positive control: `section_facts`/`facts_scoped`/`assigned_fact_ids` in `agent_definitions` = 0/0/0, against 185 live agents, predicate proved able to match (`workflow` 186, `evidence_base` 9) |
| `platform/orchestration/actions` | **does not compile in the working tree** — another session has `load_work_item_actions.go` mid-flight (`undefined: checks`). Test via a `git archive HEAD` overlay; see §5 |

## 2. What was built and committed today

- **`f611dde6a` — 1b (ii)**, the carry. `carrySectionFactsOntoRealised` re-attaches
  assignments onto restored entries by component name, at **both** loss sites:
  Pass B2 (same name, composition changed) and **Pass B** (renamed page — the
  08-09c handoff listed this as unestablished; it loses assignments the same way,
  in one of its two branches). Misses recorded durably as
  `FACT_CARRY_UNMATCHED_SECTION` in `agent_error_log`. Return type is now
  `reconcileCounts` instead of five positional ints.
  ⚠ **It also carries another session's 7-line `bugs_open/240` fix** at
  `expectedItemFieldsFromComponentSchema` — a same-file passenger a pathspec
  commit cannot exclude. Declared in the commit message.
- **`e5ed4d536` — 1b (i)**, seed `362` + the `PBP-037` register update.
- **`sameSectionList` corrected in the same commit as (ii)** — see §4, this is
  the finding of the day and it was already live.

**Neither commit carries a council trailer, and that is my error, recorded so it
is not repeated:** I committed before submitting, so `Council-Submitted:` could
not be written (the correlation did not exist yet). Both will list as un-reviewed
in the `098` report even though corr `a06ff850` covers exactly them. **Do NOT
put the trailer on a later unrelated commit to fix this** — that is the MISMATCH
the report exists to surface. The rule for next time: **build → submit → commit
with the trailer**, not build → commit → submit.

## 3. The REVISE, and how much of it is already answered

Full report: `diagnosis_artifacts` kind=`council_report`, correlation
`a06ff850-aff6-4ed0-8e0a-93d57b0cbc45`. ⚠ **`diagnosis_artifacts` carries
`expires_at` — pin anything you cite into the repo, do not link it.** Six seats
objected; `reuse_agent`, `guidelines`, `tooling_provenance`, `diagnosis_guardian`,
`constitution`, `mission`, `prior_art_librarian`, `architecture` approved.

### 3.1 The gating HIGH objection is a DOCUMENTATION defect in my submission, not a defect in the seed — settled
`editquality` objected that seed 330 targets the top-level
`{workflow,steps,generate_content,config,prompt_template}`, while a landmine
warns `page-content-writer` keeps that prompt **nested inside
`process_sections_loop`'s `sub_workflow`** — so the seed would write where
nothing reads and report success.

**Measured today: the seat's premise about the agent is RIGHT and its premise
about the seed is WRONG.** The live row has **no** top-level `generate_content`
and **does** have `process_sections_loop` + a `sub_workflow`. And seed 330
already targets
`{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}`
— the correct nested path. **The seat read my SKETCH, which said only
"generate_content prompt_template".** I carried that sketch verbatim from the
08-08 draft without opening the seed.
**Fix for the resubmission: quote the full jsonb path in the sketch.** Do not
change the seed.

### 3.2 The duplicate-active-rows objection is settled — measured 1/1/1
`build-site-planner`, `page-build-handler`, `page-content-writer` each have
**exactly one** active non-snapshot row (`page-content-writer` at version 2).
So "the seed updates by `type` and may miss the loaded row" does not apply here.
State the measurement in the resubmission rather than the reassurance.

### 3.3 Two objections are already discharged by evidence taken AFTER submission
- `debug_historian`: "no post-deploy pod-grep of the running binary." **Done
  today on v1.0.1279, both replicas, with a negative control** — table in §1.
- `debug_historian`: "the plan never states the mutants still COMPILED, so a
  'caught' mutation may be a build failure." **They compiled.** A build failure
  in this package prints `FAIL … [build failed]` and **no** `--- FAIL: TestX`
  lines — a distinction this session hit for real earlier (a test-helper name
  collision). All four mutation runs printed per-test `--- FAIL:` lines, i.e.
  the binary built and the logic was exercised. Say exactly that.

### 3.4 ⚠ ONE OBJECTION IS UPHELD AND IT CHANGES THE ROUND — re-measure before resubmitting
`debug_historian` challenged the claim *"exactly one live agent wires
`spec_sections` into `plan_sections`"*, which is the whole low-blast-radius
argument for seed 328.

**Re-measured today, by `action`, not by step name — the claim is wrong in both
directions:**

```sql
SELECT type, k AS step,
       (default_config->'workflow'->'steps'->k->'config'->>'spec_sections')  AS spec_sections_wired,
       (default_config->'workflow'->'steps'->k->'config'->>'section_facts')  AS section_facts_wired
FROM agent_definitions ad, jsonb_object_keys(default_config->'workflow'->'steps') k
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->k->>'action' = 'plan_sections';
```
→ **TWO** agents run the action (`page-build-handler`, `page-content-writer`),
and **NEITHER** wires `spec_sections` (both NULL).

So: seed 328 wires `section_facts` on `page-build-handler` only. **Establish
which agent's `plan_sections` actually runs on the writer path before
resubmitting** — if it is `page-content-writer`'s, 328 targets the wrong agent
and the consumption half would be inert for the second time in this round's
history. This is the single most valuable next query.

### 3.5 A real code hole in 1b (ii) worth fixing in the resubmission
`bug_historian`, MEDIUM, and it is correct: in `carrySectionFactsOntoRealised`,
a proposed entry whose `facts` key is **absent or malformed** is skipped before
`pending`, so it never reaches `unmatched` and never produces a
`FACT_CARRY_UNMATCHED_SECTION` row. Under seed 333 the `facts` key is
**mandatory on every section**, so an omission is exactly the disobedience this
round's measurement strategy depends on catching — and it is currently
indistinguishable from a page correctly emitting none.
**Fix:** count entries that resolve a name but carry no usable `facts` value,
and record them under a distinct code (e.g. `FACT_ASSIGNMENT_ABSENT`) or a
distinct field on the same row. Mutation-test it like the rest.

### 3.6 Two objections are design questions for the OWNER, not for me — see §6
`guardian` on the 176-line window and on the three-pipeline apply order;
`compliance` on what the human read of the v4 plaintext must explicitly check.

## 4. THE FINDING OF THE DAY — a counter changed meaning with no code change

`sameSectionList` compared whole entries with `fmt.Sprintf("%v", …)`. Seed 333
(applied 08-08) made the planner emit **objects** against a realised list of
plain **strings**. `%v` of a map never equals `%v` of a string, so **from 08-08
every composed page on every re-plan compared "changed"**:

1. `snapped_sections` silently stopped counting composition changes and started
   counting SHAPE differences — any figure spanning 08-08 is two measurements
   with one name;
2. every composed page was pointlessly restored over a list already identical to
   the realised one — a no-op in content, and a fact-assignment killer in effect.

**No Go changed to cause that. A row in `agent_definitions` did.** Fixed by
comparing section NAMES. Now in `LANDMINES.md` (footprinted on the symbol) and
inline in `PBP-037`, because the general form recurs: on this platform Go is
inert until a roll and DB config is live instantly, so **a prompt edit can
redefine a Go metric between two reads of it.**

Pleasant side effect: once seed 362 applies, the planner re-emits the realised
names, the composition compares EQUAL, nothing is restored — and assignments
survive **with no carry at all**. The carry becomes the safety net, not the
mechanism.

## 5. Do these, in this order

1. **Run the §3.4 query and settle which agent's `plan_sections` serves the
   writer path.** Everything else in the round is contingent on it.
2. **Fix the §3.5 hole** (absent/malformed `facts` is invisible). Commit narrowly.
3. **Resubmit** with `RESUBMIT_CORR=a06ff850-aff6-4ed0-8e0a-93d57b0cbc45` so the
   trail accumulates. Carry over: the full jsonb path in 330's sketch (§3.1), the
   1/1/1 row measurement (§3.2), the pod-grep + mutants-compiled evidence (§3.3),
   the corrected `plan_sections` census (§3.4), the new hole's fix (§3.5). Drop
   seed 333 from `edits` and cite it in `grounded_in` only (`editquality`, LOW —
   it is applied and changes nothing now).
   **Submit BEFORE committing** and use `Council-Submitted: <corr>`.
4. **Then**, and only after the owner decisions in §6: apply `362` → `328` → `330`
   → replan + rebuild fundamentallyai's flagged pages → census. Overlap pairs
   must fall on engaged pages; the five fact-blind sites must **not** move (the
   disconfirming half).

## 6. OWNER DECISIONS OWED (nothing below is mine to settle)

1. **The 215 lossy-merge policy** — carried forward unresolved from 08-09c. When
   two *composed* pages collide, the shipped fix keeps the richer and logs the
   other at Warn: silent partial data loss. The alternative is failing the write,
   i.e. today's whole-replan loss. My position is unchanged (proceeding is right;
   the branch needs a collision **and** both entries composed, and the observed
   shape is composed-plus-stub, which loses nothing) — but "how much silent loss
   is acceptable" belongs to the owning pipeline.
2. **The human/compliance read of the v4 writer plaintext**
   (`brochure_component_library/sql/page_content_writer_prompt_v4_2026-08-06.txt`)
   must happen **before seed 330 applies** — the compliance seat's round-1 ask,
   restated this round with an addition: the read should explicitly check the
   three-way branch (scoped / factless / unscoped) for **overclaimed-reliability
   phrasing**, especially what the writer is told to say when a section has "no
   verified facts". A person has to do this; I cannot self-certify it.
3. **Should seed 362's redesign escape be a FIELD rather than prose?** 362 tells
   the planner to re-emit a built page's realised sections verbatim, with an
   escape ("only when the briefing explicitly asks for a page to be redesigned").
   `recompose_pages` is the real mechanism for that, and the planner is **not
   told which pages are on it** — so the escape is prose in a prompt.
   **This is precisely the shape of OWNER RULING 2026-08-02 (RFC_010 §2): new
   authority on a shared seam ships as an opt-in FIELD, not a documented
   contract, because "a comment is not a control on a tree this many sessions
   share."** Plumbing `recompose_pages` into the prompt is a bigger change than
   this round should carry. Your call whether it blocks 362 or becomes a
   follow-up.
4. **The guardian's three-pipeline apply-order concern.** The round touches
   config on `build-site-planner` (362), `page-build-handler` (328) and
   `page-content-writer` (330). Order is documented (362→328→330) but enforced
   only by a `_HOLD` filename, and a partial apply leaves the planner re-emitting
   realised sections with nothing consuming the assignments. That is inert rather
   than harmful, but if you want it enforced by tooling rather than by discipline,
   say so and it becomes work.

## 7. Traps (this front's, still live)

- **`platform/orchestration/actions` does not compile in the working tree** —
  another session's WIP. Build and test like this:
  ```bash
  git archive HEAD | tar -x -C <scratch>/overlay
  cp platform/orchestration/actions/{v3_site_actions.go,v3_site_reconcile_test.go,v3_site_fact_carry_test.go} \
     <scratch>/overlay/platform/orchestration/actions/
  cd <scratch>/overlay && go test ./platform/orchestration/actions/ -count=1
  ```
- **Pre-existing red at HEAD, not this front's:** `discovery_checks`
  `TestEveryCheckProducedItemTypeIsClassified` (`decision_regression` has no
  verifier). Confirmed on a pure HEAD archive. The 08-09c line "HEAD tests
  clean" is expired for that package.
- **`orchestration_states` prunes at ~24h**, failures and completed rows alike.
  Never use an error census over it as proof a fix worked — it reads 0 either
  side. `diagnosis_artifacts` carries `expires_at` too.
- **Read the verdict by CORRELATION, never `doc_notes … LIMIT 1`** — that
  returns whichever lane finished most recently, and it did today (it handed
  back an unrelated APPROVED round for a code-index change).
- **A roll is not evidence.** Grep a string your change ADDED **and** a
  plausible-but-absent one, same exec, every replica.
- **Line numbers in this directory's docs go stale within hours.** Cite by
  symbol. As of this commit: `reconcilePlanWithRealised` `:5479`,
  `carrySectionFactsOntoRealised` `:5232`, `sameSectionList` `:5909`,
  normalise pass `:3277`, reconcile call `:3101`.
- **A replan on this site IS a build dispatch** and, until the quiet mode is
  fixed, still a phantom-page generator. Co-ordinate with the sweep front.
- **Never `run-migrations.sh --apply`** — the pending list carries other lanes'
  files plus the three `_HOLD` slices (328, 330, 362).
- **The 097 trigger validates `operation`** against
  `modify|add|remove|config_change`.
- Rollback for the prompt seeds: `agent_definitions_bak_329` / `_bak_333`; 362
  creates `_bak_362` when it applies.

## 8. Commit trail (this front)

`d6e9dcf06` decisions + seed 333 · `47620cb53` bug 215 filed · `c589779a3` Pass
B2 correction · `9b61d04b1` 08-08 handoff · `f58357515` 08-09 re-check ·
`14b1cff28` the 215 dedup fix · `90414d055` bug-file + docs · `fa483dcdc`
landmine footprint fix · `1c854175b` council verdict pinned · `?` 08-09c handoff ·
**`f611dde6a` 1b (ii) the carry + the `sameSectionList` correction + LANDMINE** ·
**`e5ed4d536` 1b (i) seed 362 + register** · this file's commit (REVISE
disposition + the three re-measurements + this handoff).
