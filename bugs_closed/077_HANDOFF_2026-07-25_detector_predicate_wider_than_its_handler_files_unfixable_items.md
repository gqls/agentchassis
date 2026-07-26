# HANDOFF — the `hardcoded_section_colors` detector files items its own handler cannot fix

**Filed:** 2026-07-25, by the thread closing `bugs_open/021`. Split out of 021's
INSTANCE 2 rather than left inside a closing file.
**Severity:** LOW. Not an outage, no data loss, no credit spend. It is a
permanent, bounded population of work items that misdescribe reality.
**Status:** **CLOSED 2026-07-26** — code committed `ce4adfac4`, INERT until the
next chassis roll; migration `221` written and deliberately NOT applied (see
SEQUENCING).

**Council: NO VERDICT, and therefore NO `Council-Reviewed:` trailer on any commit
here.** Submission `346500db-89ca-47f3-bc5a-e1c099d6f4f8` was fired twice and
never returned one. Round 1 vanished with no orchestration row at all, published
2–4 minutes after a chassis pod restart (`startTime 18:35:07Z`) — CLAUDE.md's
documented ~300s drop window, which the 097 trigger does not warn about. Round 2
landed ("Lane is clear, LAG 1"), started at 19:23:14, and **froze one second
later** at `review_editquality` with `awaited_steps = []` and no error. A
different correlation started after it and completed twice while it sat there, so
the lane was healthy — this is `bugs_open/029`, and the instance is recorded in
that file. A third round was not fired: another full council into a lane that is
demonstrably losing runs buys a likely-identical hang. The trailer is earned by an
APPROVED verdict only, so its absence here is correct rather than an omission, and
the `098` coverage report will list these commits as un-reviewed — accurately.

> **NOW LIVE AND VERIFIED — v1.0.1171, 2026-07-26 21:02:56Z.** This block
> previously warned that the fix was inert and unobserved; that no longer holds and
> the original wording is superseded rather than deleted, because the sequencing it
> describes is what made the cleanup safe.
>
> **Deploy proven discriminatingly**, not by the vacuous grep: the OLD summary
> string `"hardcoded hex colors in inline styles instead of CSS variables"` returns
> **0** in the running binary, the three new markers return 1 each, and the
> positive control `"unresolved after %d attempts"` returns 1 (so `strings` works).
>
> **Migration 221 applied AFTER the roll, in that order** — `UPDATE 3`, exactly the
> three rows predicted, recorded in `schema_migrations` via `--record-only`.
> Applied by hand, NOT via `run-migrations.sh --apply`, which would have swept four
> other threads' pending files into this task.
>
> Live behavioural evidence is in **RESULTS** below.

---

## RESOLUTION (2026-07-26)

**Owner's ruling on the design call this file refused to make for itself:** keep
the wide detection, **split** the output, and **queue the handler work**. That is
candidate **C**, implemented so that it also opens the door to **B** — and
explicitly not **A**, because the detector's breadth is the only thing that sees
light / 3-digit / inline-attribute colours at all.

### What shipped

A check now **partitions its own population by the handler's literal transform**:

```
population
  ├── in remit  → the normal dispatchable item, counted HONESTLY
  └── residue   → ONE capability_gap: "found work I have no handler for"
```

`capability_gap` is not a new item type. It is the platform's existing durable
record of exactly this, emitted since long before by `WriteBuildItemsAction`
(`load_work_item_actions.go:250-280`) for page types whose builders do not exist
yet, already classified in the build-enforced coverage guard
(`verifier_coverage_test.go:278`), and already read as a **roadmap** view by
`diagnose_triage_action.go:355-378` and `fixloop_digest_action.go:358`. So the
residue is queued for a handler to be built by machinery that already exists —
and this answers the standing objection recorded at
`durable_write_guard/PLAN_2026-07-21:158-163` that a second list drifts from the
one the build enforces. **No 78th item type was added.**

The residue item is undispatchable **twice over**: `status='deferred'` (in
neither `claim_work_item_action.go:102` nor `load_work_item_actions.go:559`) and
an empty `handler_agent` (a row that somehow reached claim is blocked, not
handed to a fixer that cannot fix it). It is bounded for free — `deferred` is not
in `workItemTerminalStatuses`, so `idx_swi_dedup` holds **one open row per site
per check** — and `stale-work-item-reaper` only touches
`status='triaged' AND pipeline='build'`, so it cannot churn them.

| file | what |
|---|---|
| `discovery_checks/remit.go` | **new.** `PartitionByRemit`, `CapabilityGapItem`, `HandlerStepConfig` — the shared seam |
| `check_hardcoded_section_colors.go` | this bug. Detector and verifier now share ONE population SQL constant *and* one remit predicate |
| `check_forced_text_colors.go` | handler agent **never existed** → capability_gap only |
| `check_component_standards.go` | `missing_site_metadata` — handler agent **never existed** → capability_gap only |
| `check_broken_nav_links.go` | partitions on the TEMPLATE its handler edits; reads the LIVE seeded patterns |
| `fix_nav_link_templates_action.go` | transform + patterns rehomed into `discovery_checks`, one copy |
| `handler_coverage_test.go` | **new.** Build-enforced: no check may route at a non-existent agent |
| `sql_for_agents/221_…sql` | **new, UNAPPLIED.** Retires the 3 lying rows |

### Three things found while fixing it that were not in this file

1. **The degenerate case is the common one, and nothing was looking for it.** A
   remit of zero because the handler *does not exist*. Two were live:
   `forced-text-color-fixer` (its check IS enabled in `design-discovery-agent`,
   with one live match waiting on webdesign.co.uk) and `site-metadata-fixer`
   (severity **high**, priority **3** — near the front of the queue). Every item
   either filed was destined for `blocked` at claim after occupying a dedup slot.
   Nothing in Go, in config, or in any test connected `HandlerAgent: "x"` to
   "does x exist?". `handler_coverage_test.go` found both on the day it was
   written, and now fails the build on the third.
2. **A remit is not always all in Go.** `nav-link-fixer`'s find/replace patterns
   are seeded in `agent_definitions` and **override** the action's Go defaults
   entirely; the live row carries **three** where `DefaultNavLinkPatterns`
   declares **four**. A check partitioning on the source would credit the handler
   with a replacement it does not make — this same defect one level down, and
   invisible to any test that only reads the repository. So the check reads the
   live config.
3. **The artefact the handler edits is not always the one the detector read.**
   `broken_nav_links` detects on rendered `site_components`; its handler rewrites
   `content_components.html_template`. "Is this fixable?" is a question about the
   template.

### On the numbers in this file — NOT a correction

Re-measured 2026-07-26, the population is **33 components across 9 sites** (this
file said 32 / 8; vonc.com has since joined). The per-site remit column below was
computed here with a SQL predicate *strictly wider* than the Go transform, which
is a different instrument from the one that produced the table further down — and
the two **do not disagree**. A superset can prove zero; it can never disprove it,
so a superset of 1 around a true value of 0 is exactly what a superset is for. An
earlier draft of migration `221` claimed this file's webdesign.co.uk figure needed
correcting. It did not, the claim was wrong, and it is logged in `WRONG_CALLS.md`
(2026-07-26, *"my wider SQL predicate contradicts 077's remit table"*).

What the wider predicate **does** establish is which rows can be retired by SQL
alone. Of the sites carrying a work item: ai-agent-orchestration.com, finetuning.uk
and gaswholesalers.com are provably zero → **3 rows retired**. webdesign.co.uk and
dartsonline.com carry detector matches but **no work item**, so there was nothing
to retire on either.

### SEQUENCING — the one thing to get right

Migration `221` is written, lint-clean, probe-clean and **deliberately unapplied**.
Retiring a row frees its `idx_swi_dedup` slot. Applied while the OLD detector is
still deployed, the next discovery pass over those sites re-files the same
dishonest item and the cleanup is undone. **Roll the chassis first, verify against
the pod, then apply.**

---

## The mechanism

The **detector** and the **handler** for `hardcoded_section_colors` implement two
different predicates, and the detector's is strictly wider:

| | predicate |
|---|---|
| detector (`countHardcodedColorComponents`, `check_hardcoded_section_colors.go:95`) | `rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'` **AND** `rendered_html LIKE '%<style%'`, `locked_at IS NULL` — ANY hex, 3/4/6/8-digit, light or dark, **anywhere in the component** including inline `style=""` attributes; the `<style>` test is on the component, not on the colour's location |
| handler (`ReplaceHardcodedColors`, same file, used by `fix_hardcoded_colors` / agent `color-variable-fixer`) | dark 6-digit only (`#[0-4][0-9a-fA-F]{5}`) on `background`/`background-color`, plus two-colour `linear-gradient(Ndeg, …)` — and **only inside `<style>…</style>` blocks** |

So the detector can file an item on a site where the handler provably has nothing
to do. The handler then runs, changes nothing, and the item is eventually parked.

## Live evidence (2026-07-25, all counts re-run today)

Detector population, with "inside the handler's remit" computed by running
`ReplaceHardcodedColors` verbatim over each component's `rendered_html`:

| site | detector matches | inside handler's remit |
|---|---|---|
| robot-hands.com | 3 | 3 |
| gamesdesign.co.uk | 4 | 1 |
| leopardessconsulting.co.uk | 4 | 1 |
| **finetuning.uk** | **8** | **0** |
| **gaswholesalers.com** | **6** | **0** |
| **ai-agent-orchestration.com** | **4** | **0** |
| **webdesign.co.uk** | **2** | **0** |
| **dartsonline.com** | **1** | **0** |

**On 5 of 8 sites the handler's remit is empty.** 32 components matched, 5 of them
fixable.

Work items of this type, live today: **13** (4 `complete`, 8 `unresolved`, 1
`detected`). The 8 `unresolved` were parked by the two-strike rule
(`insertWorkItem`, `load_work_item_actions.go:1041`) — whose label is *"handler
had 2 chances and the issue persists"*. On the zero-remit sites that label is
wrong about the cause: the handler did not fail, it was never able to succeed.
Oldest is 2026-04-08.

## What is NOT happening — measured, because the obvious guess is wrong

The 021 INSTANCE 2 note called this **churn** ("correct completions keep
re-detecting; the cycle repeats"). That is **not** what the data shows, and the
correction matters because it changes the severity:

- `idx_swi_dedup` is UNIQUE on `(site_id, item_key)` excluding only terminal
  statuses — and `detected` is **not** terminal. So one open item per site blocks
  any re-file. robot-hands.com has carried a `detected` item since 2026-07-17;
  a design-discovery sweep ran over that very site on 2026-07-24 20:46 (it filed
  `undeployed_asset` ×21, `needs_sprite_css` ×3, `needs_imagery` ×4 — so the
  agent and this check's list both ran) and filed **no** new
  `hardcoded_section_colors` item.
- `hardcoded_section_colors` appears **zero** times in 7 days of discovery output
  fleet-wide.

So the volume is bounded at roughly one open item per site, indefinitely. There is
no repeated dispatch and no repeated spend — and the handler is not LLM-driven
(3 workflow steps, no `execute_llm_prompt`), so a wasted run costs compute only.
**This is a correctness/legibility defect in the backlog, not a cost defect.**

## Why it is worth fixing anyway

1. **The backlog lies.** A human or an agent reading `site_work_items` sees 8
   unresolved "hardcoded colours" items and reasonably infers 8 sites with a
   colour problem a fixer keeps failing on. On 5 of them there is nothing the
   assigned fixer was ever going to change.
2. **`unresolved` is load-bearing elsewhere.** It is the two-strike parking state
   used fleet-wide to mean "needs investigation". Poisoning it with items that
   are *correctly* unfixable devalues the signal for every other type.
3. **The completion gate is now scoped to the handler and discovery is not.** As
   of `34adb171c` (live v1.0.1159) `VerifyHardcodedSectionColorsResolved` asks
   "would the fixer's own transform still change anything?" — deliberately the
   HANDLER's remit (see `016b` §9, *"A 'did the fix work?' check must assert the
   HANDLER's remit, not the DETECTOR's predicate"*). Detection still asks the
   wider question. The two ends of the same item now disagree by design, which is
   defensible as a stopgap and poor as a resting state.

## Candidate fixes (a design call — do not just pick one)

- **A. Narrow the detector to the handler's remit.** Count only what
  `ReplaceHardcodedColors` would change — ideally by *calling it*, as the verifier
  does, so the three ends cannot drift. Cheapest and makes item counts honest.
  **Cost:** the platform stops noticing light/3-digit/inline-attribute hardcoded
  colours entirely. If anyone wants those found, they lose their only detector.
- **B. Widen the handler.** Teach the fixer light hexes, 3/4/8-digit forms and
  inline `style=""` attributes. **Cost:** materially riskier — inline styles and
  light colours are where legitimate design intent lives, and a wrong rewrite is
  a visible site regression. This is the branch that wants a real review.
- **C. Split the item type.** `hardcoded_section_colors` (fixable, dispatch to
  `color-variable-fixer`) and a report-only finding for the rest. Honest, but adds
  a type to a registry that already carries 77.

**Before picking:** the 5 zero-remit sites are the test set. Whatever ships must
leave them with either zero items or items truthfully labelled as report-only.

## How to verify a fix — UNRUN, and it is the next person's job

Nothing below has been executed: the Go is inert until an image roll. A green
happy path would only prove deployment anyway, so **step 2 is the load-bearing
one** — without it, "no dishonest item was filed" is indistinguishable from "the
check stopped filing anything".

**0. Roll the chassis** past `ce4adfac4`, then pod-grep a string this change
CREATED (not one it uses), with a positive control:

```
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "capability gap, not a handler failure"'   # expect >0
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "unresolved after %d attempts"'            # control, expect >0
```

**1. Fire design discovery at a provably ZERO-remit site** — finetuning.uk (8
matches, 0 in remit). Template:
`scripts/initial_messages/290_design_discovery/081_design_discovery_agent_robot_hands.sh`.
Expect **one** `capability_gap` row (`status='deferred'`, `handler_agent=''`,
`spec->>'residue'='8'`, `spec->>'gap_kind'='handler_remit'`) and **no** new
`hardcoded_section_colors` row.

**2. POSITIVE CONTROL — fire at robot-hands.com** (3 matches, all 3 in remit).
Expect the opposite: a `hardcoded_section_colors` item with
`spec->>'components_found'='3'` and **no** capability_gap.

**3. Apply migration `221`** (only now), and confirm 3 rows moved to `wont_fix`
carrying `spec->>'retired_by'`.

**4. Re-run the remit table.** For every dispatchable item filed, "detector
matches" and "inside remit" must now agree — this file's original acceptance
criterion. Method in
`docs024_key_docs_latest/durable_write_guard/RUNBOOK_durable_write_guard.md`,
§"Know the expected verdict": dump the detector population with `row_to_json` and
run the shipped transform over it.

**5. Then the capability gaps are intake, not decoration.** Grouped by
`spec->>'builder_needed'`, they are the queue for the feature builder
(`0NN_TRIGGER_feature_designer_v1.sh <work_item_id>`), whose spec gate needs an
`owner_approval` stamp — a human act, deliberately. `forced-text-color-fixer` is
the cheapest first candidate: its action `fix_forced_text_colors` is **already
registered**, so the gap is a seed plus a remit decision, not a build. Do not seed
it blind — that action bails out entirely below its WCAG contrast floor and only
rewrites text-element selectors, so seeding it without partitioning its check
re-creates this bug under a different item type.

## References

- **`016b` §9 — *"A detector must PARTITION its population by the handler's remit,
  and file the residue as a capability gap"*** (2026-07-26), the transferable
  pattern written on closing this. It is the deliberate other half of the
  2026-07-25 entry below.
- **`bugs_open/083` (`detected_findings_never_reach_a_handler`) — SEQUENCING.**
  That bug names this one as a blocker it must clear first: *"Do not enable (1) or
  (2) without checking that, or the pile moves from `detected` to `failed` and
  nothing improves."* 98 rows sit at `status='detected'` fleet-wide because the
  only promoter lives in the disabled `improvement-sweep`. **That constraint is
  now satisfied for the three checks fixed here** — their residue never reaches
  `detected` at all. It is NOT satisfied for the other 18 item types in that pile,
  and this fix says nothing about them. 083 is a *delivery* gap; this was a *remit*
  gap. They are different bugs and the distinction is load-bearing.
- `bugs_open/071` / `079` — earlier stages of the same detect→persist→promote→fix
  pipeline (a gate that detects and discards; a finding that never blocks). Also
  not this class. 079's own fix candidate 4 already cites this file as a trap to
  avoid: *"do not file items whose handler has no remit to fix them."*
- `WRONG_CALLS.md` 2026-07-26 — the false correction this thread nearly published
  against this file's own figures.
- `bugs_open/021` (now `/bugs_closed/`) — INSTANCE 2, where this was found and
  where the verifier's remit-scoping is argued in full.
- `016b` §9 — *"A 'did the fix work?' check must assert the HANDLER's remit, not
  the DETECTOR's predicate"*, the transferable pattern.
- `WRONG_CALLS.md` 2026-07-20 — the held `page_rerender` verifier, the near-miss
  that taught the distinction.
- `platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors.go`
  — detector, handler transform and verifier, all three in one file, deliberately.
