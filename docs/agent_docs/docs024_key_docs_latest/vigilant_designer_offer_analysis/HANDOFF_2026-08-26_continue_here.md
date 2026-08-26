# HANDOFF — vigilant designer + offer analyser (2026-08-26)

**COLD-START = this file + `PLAN_2026-08-25b_imagery_slot_and_visual_designer_dispatch.md` §8 +
`bugs_open/395` §9 + register `WII-033` / `CLM-024` + `../loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md`.**

**Supersedes `HANDOFF_2026-08-25b_continue_here.md`**, which is correct on history and **wrong on
three headline facts**: the sweep switch is no longer held, the gate is no longer inert, and
"5 of 28 sites" is now **13 of 31**.

> **Re-run every number here before acting.** This branch takes hundreds of commits a day and the
> improvement sweep is now running on a 900s tick, so the fleet-state figures below move hourly.
> Verify against `git archive <resolved-sha>`, never the working tree (measured twice on 2026-08-25:
> another lane's untracked test failed the package mid-run, and a `mapKeys` redeclaration broke the
> build from a file I had not touched).

---

## §0 — THE THREE THINGS THAT CHANGED OVERNIGHT

**1. THE SWEEP IS ON.** `improvement-sweep` was enabled 2026-08-25 21:18:19Z **by the owner's
explicit word**, executed by the `loanzy_uk_example_site` lane — not by this lane, which held it as
instructed until released. `[VERIFIED 2026-08-26 08:55Z]` `enabled=true`, ticking.

**2. GATE 1C IS LIVE** on `v1.0.1339` (rolled 2026-08-25 19:07:18Z), capability-probed PRESENT on
**both** replicas with a must-be-present and a must-be-absent control. ⚠ **Do not re-verify it with
CLAUDE.md's prescribed check — that check cannot succeed** (see Watch-outs).

**3. RECORD MODE IS LIVE ON THIS LANE'S SEATS** (migration 624, 21:27:02Z). `offer-analyser`'s
`write_audit_findings` step carries `filing_mode: record`; **6** seats total. Findings park as
`deferred` verdict rows for a person to release.

### ⚠ AND THE CONSEQUENCE THAT WILL MISLEAD YOU IF NOBODY SAYS IT

`[MEASURED 2026-08-26 09:0xZ]`

| | |
|---|---|
| sites audited since 624 | **23** |
| offer-analysis findings since 624 | **28**, of which **5** carry an `acceptance_predicate` |
| status of those 5 | **`deferred` ×5** |
| **gate 1c verdicts recorded, ever** | **ZERO** |
| sites carrying `offer_ordering` | **13** (was 5) of **31** live sites |
| `homegarden.uk` enrolled | **YES** — the site the owner reviewed is now audited |

**The zero is CORRECT and EXPECTED, not a fault.** A record-mode finding never reaches `complete`, so
gate 1c never grades it. ⚠ **An empty `result->'_verification'->'acceptance_predicate'` census has now
meant THREE different true things in 24 hours** — *"the gate was never switched on"* (pre-roll),
*"no completed item carried a predicate"* (post-roll), and now *"this producer's findings are parked
by design"*. Same query, same empty result, opposite conclusions. **Read the meaning off `WII-033`,
not off the number.**

## What is DONE — do not re-take any of this

| | state |
|---|---|
| **completion gate 1c** | **LIVE**, `69479bcf6`; council **APPROVED r1** `064841bd` (14 seats, 5 advisory); objections acted on in `74829da90` |
| `PromotionOwes` | **rewritten twice on 2026-08-25** — the negative control is **UNREACHABLE**, not absent (`864e73d8a`). See §2 |
| `bugs_open/395` §8 + §9 | what shipped, and the correction that it is a **recurrence of `bugs_open/320` §5** |
| `016b` §9 | my 08-25 entry **corrected in place** — the verifier-seam sentence was gate 1b's argument and does not transfer |
| register `WII-033` + index row | live status, record-mode interaction, the three-readings warning |
| `RFC_055` (completion-gate accumulation) · `RFC_057` (`HandlerCanWriteField` as a shared contract) | **FILED, awaiting the owner.** Neither asks to reverse anything |
| `LANDMINES` ×3, `WRONG_CALLS` ×2 | incl. the joint **prose-rules-have-a-half-life** entry with the 395 lane |
| `PLAN_2026-08-25b` §8 | ⚠ **the imagery design is CORRECTED — see §1. Do not build a component.** |
| the sweep's consumed-`audit_due` defect | found here, fixed by migration `625` the same minute; **0** bypass-era stamps remain |

## What the next session should do

### 1. THE IMAGERY FIX IS ONE `UNION` ARM. This is the work, and it is small.

**Read `PLAN_2026-08-25b` §8 first — §2a is SUPERSEDED and would have you build a duplicate.**

`Illustrated Text Block` already exists: section-level, active, prose plus a gated `<figure>`, lazy
loading, caption, responsive CSS, and — the part that matters — `image_url`/`image_alt` carry
`source: "site_assets.image"`, so paths resolve **server-side**. `[MEASURED 2026-08-26]` **6 live
instances in the whole estate, all on ONE site.**

**Nothing chooses it because the planner cannot tell it apart from `Generic Text Block`:**
`component_expresses` returns `[html-block, list, table]` for **both**. The function derives four
tokens and **none of them is an image**. `bugs_open/381` taught the planner to see lists, tables and
item sets; imagery was never added to the vocabulary.

The fix, derived from the SCHEMA not the template:

```sql
UNION
SELECT 'image' WHERE EXISTS (
  SELECT 1 FROM jsonb_each(COALESCE(p_input_schema->'fields', '{}'::jsonb)) f
   WHERE f.value->>'source' = 'site_assets.image')
```

⚠ **A template grep (`<img`) would report 47 components — every header, hero and card thumbnail,
chrome the writer cannot influence.** The schema predicate reports **9, of which 8 are active and
section-level**. That precision is the design.

**Owed with it:** council round (it changes what every planner is shown, fleet-wide, on every
subsequent build); a register entry in the same commit; and **the control that distinguishes a
widening from a reshuffle — `component_expresses` output for all 8 affected components before and
after, plus `Generic Text Block` proven byte-identical.**

### 2. ⚠ DO NOT GO LOOKING FOR GATE 1C'S NEGATIVE CONTROL. It is UNREACHABLE.

`bugs_open/395` §6 asks for one item whose predicate is SATISFIED after its fix. **It cannot occur
today**, and two earlier wordings of `PromotionOwes` — including one of mine — named routes to it
that do not exist. `[MEASURED 2026-08-25]` every predicate this producer has emitted grades
`pages.meta_description`, and a **non-empty** meta description is structurally immutable on the route
these findings take: `page-build-handler` has **zero** steps touching the column; the only
unconditional writer is reachable from one agent whose selector takes **empty** values only; the other
two writers are `COALESCE(NULLIF(EXCLUDED,''), existing)`.

**Read `PromotionOwes` in `complete_work_item_acceptance_predicate.go` before touching promotion.** It
carries all three conditions and both superseded wordings, deliberately.

### 3. The supply question — the bigger half, and nobody owns it

`[MEASURED 2026-08-25]` assets per page, 27 sites with >10 components: **median under 1.0**;
**13 sites at 0.5–1.0**; **webdesign.co.uk at 0.10** (149 pages, 15 assets); **THREE sites at ZERO**
(adversecreditmortgage.co.uk, loancash.co.uk, loanandmortgagecalculator.co.uk — 88 pages between them).

**§1 does not touch this.** `on_missing: skip_field` means a site with no assets renders the
illustrated component as plain prose, silently. **Shipping §1 alone and reporting the imagery ask as
answered would be this lane's own `bugs_open/395` shape** — a green record for work that did not
happen. Say the supply figure alongside it.

### 4. Ask the owner one question rather than assuming

He asked for imagery **"between paragraphs"**. `Illustrated Text Block` places its figure **above**
the prose, once per section. Alternating sections gives interspersed imagery *at page level*; within a
section it does not. **If he meant strictly mid-prose, §1 answers the wrong reading** — and that is a
different component and a different plan.

### 5. Carried forward, still not done

- **The v2 batch (a)+(b)+(c)** — config-only, migration `602` unwritten. Traps in `features_open/030`
  §10. ⚠ v2(a) GROWS the offer surface and **widens what a predicate can address** (body-text shapes
  are excluded today precisely because the surface carries no content).
- **`features_open/034`** — claims audit over `site_specs` prose. Owner-approved 2026-08-14, still not
  designed.
- **The `HandlerCanWriteField` drift audit** — named in `RFC_057` §4, **unbuilt and unclaimed by
  either lane**. ⚠ Read `RFC_057` §3/§3a first: it must resolve steps through the **action registry**,
  not by grepping config text, or it goes green while wrong about two thirds of the writers — and its
  instrument must err in the direction its claim errs in.

## Watch-outs — new ones first

- **⚠ CLAUDE.md's PRESCRIBED DEPLOY CHECK CANNOT SUCCEED.** It says read a `build provenance` startup
  line; `[MEASURED 2026-08-25]` that string is emitted **nowhere** in this repo's Go source (the only
  hits are prose in comments and an unrelated box service). CLAUDE.md also says an empty result means
  *"not in range"*, so **the documented failure mode absorbs the real one**. Use a capability probe
  with **both** controls, or `SELECT git_commit FROM service_binary_capabilities WHERE
  service='agent-chassis' AND kind='build' ORDER BY last_seen_at DESC LIMIT 1` + `git merge-base
  --is-ancestor` with a control each way.
- **⚠ "GREP BEFORE YOU FILE" DID NOT SAVE THIS LANE.** I grepped `/bugs_open/` for the MECHANISM
  (completion gating, acceptance tests) — correctly, thoroughly, nothing found — and `bugs_open/320`
  was filed under *meta descriptions are never asked for*, six days earlier, saying **do not file the
  item I then filed**. **Grep for the mechanism AND for the column/table/field.** Full entry in
  `LANDMINES.md`; the general form is that a rule existing only as prose in a bug file has a half-life.
- **⚠ A `$.workflow.steps` WALK UNDER-REPORTS.** It misses nested `sub_workflow` steps. Mine returned
  **ZERO** agents for `save_page_meta_description` and the truth is one — which would have been
  *more alarming than reality AND false*. Use whole-config `default_config::text LIKE '%…%'`.
- **⚠ A CONTROL PROVES YOUR INSTRUMENT WORKS; IT CANNOT TELL YOU IT IS POINTED AT THE WRONG THING.**
  The 395 lane answered a council objection with a SQL check carrying a correct demand control that
  asked *"does this config mention the column"* and read the answer as *"can this agent write it"*.
  The two differ by two thirds of the population.
- **⚠ `pattern-check.py` READS YOUR PROSE**, string literals and comments included, six lines from any
  log sink — and the comment you write to explain the false positive re-fires it. **And a no-argument
  run reads `git diff --cached`**: unstaged means it scans ZERO files and prints a clean result.
  Confirm the denominator.
- **⚠ A STORED `acceptance_predicate` CANNOT BE FED TO `EvaluateAcceptancePredicate`** — the emission
  provenance stamp breaks its closed key set, so every live predicate returns `inapplicable`. Use
  `predicateForEvaluation(stored)`. `TestStampAndStripAreInverses` catches a third stamped key.
- **⚠ `verifyBeforeComplete` RETURNS FOUR VALUES**, and `verification != nil` **no longer means
  blocked** — a completion that proceeds can carry a verdict. Stamp, then read `mayComplete`.
- **⚠ REGISTER FILES ARE HIGH-TRAFFIC BETWEEN THIS LANE AND `bugs_open/395`'s.** Their WII-035 entry
  rode into my commit `6ef02131a` as a same-file passenger. Nothing lost; use narrow pathspecs.

### Carried forward, still true

- **⚠ CONFIRMING THAT THE PROMOTER YOU THOUGHT OF IS OFF IS NOT A SAFETY ARGUMENT.** The cheap check
  is the inverse query — `SELECT name, enabled FROM scheduled_tasks WHERE enabled`. ⚠ **And
  `detected-item-promoter` (900s, live since 08-15) dispatches `detected` rows regardless of the
  sweep** — `detected` is a queue, not a shelf.
- **⚠ `pages.in_header` IS NOT THE RENDERED NAV.** Nav stays out of the predicate vocabulary.
- **⚠ A ROLL KILLS AN IN-FLIGHT COUNCIL**; a lone casualty is the expected shape.
- **⚠ MIGRATION NUMBER COLLISIONS ARE ROUTINE** — two 601s, and 619 collided again on 08-25. Resolve
  by slug, never by number.
- **⚠ `site_work_items` has no `audit_source` COLUMN** (it is `spec->>'audit_source'`, and the column
  form ERRORS); **`orchestration_states` has no `agent_type`** (it is `owner_agent_type`).
- **⚠ psql prints UTC, your shell prints BST** — always toward alarm. Make the DATABASE subtract.

## Residuals, stated plainly

1. **Gate 1c's refusal arm has never fired**, and its negative control is unreachable (§2). Third
   instance of CLM-023's residual, stated in the roster entry rather than buried.
2. **Gate 1c currently grades nothing**, by design, while record mode stands (§0).
3. **The imagery supply gap (§3) is unowned.**
4. **The `HandlerCanWriteField` drift audit is unbuilt** and its roster is a negative claim that goes
   stale by addition.
5. **The truncation asymmetry, unmeasured:** predicates are authored against a meta description
   truncated at 160 chars; the evaluator reads the full column.
6. **Why the handler produces content failing its own criterion** is the 395 lane's feed-forward work,
   not this lane's.

## Who owns what nearby

**`bugs_open/395`'s lane** owns the routing seam (**WII-035**), `HandlerCanWriteField`, and
feed-forward. **The split is agreed: they own routing, I own the emit-side stamp** — and my half
**STAMPS `field_writable:false`, it does not reject**, because rejecting would delete the field their
guard reads and the two would blind each other. Not yet built.
**`loanzy_uk_example_site`** owns the owner review, `RFC_056`, the improvement-loop seats and the
sweep. **The `bugfix_381`** lane owns `component_expresses` and the planner menu — **§1 is their
mechanism and they should be told before it is widened.** `bugfix_390_cascade_attribution` has the
css-patch-agent fix; the **leopardess** lane still holds 123 items at `needs_human_review`.
