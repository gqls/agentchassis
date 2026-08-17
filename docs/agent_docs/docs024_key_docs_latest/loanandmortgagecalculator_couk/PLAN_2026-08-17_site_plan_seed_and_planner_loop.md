# PLAN 2026-08-17 — the site-spec seed and the planner loop (owner ruling D6)

**Status: opened this session, nothing dispatched yet.** This is the working doc for the
item `HANDOFF_2026-08-15b_continue_here.md` §1 names as the next work. Everything below
was measured against the live DB on 2026-08-17 unless marked otherwise.

## 1. What D6 asks

Owner ruling, 2026-08-11 (verbatim constraints, from
`HANDOFF_2026-08-11_after_track_a_decisions_pending.md`):

> Seed the site spec, let the planner plan, reseed until the plan is *reasonably close
> to where we are*. The exact combination/makeup of calculators and guides is NOT
> important; keeping the overall size, density and complexity IS — **the site must not
> shrink on rebuild**; the improvement loop growing it over time is welcome. (Site is
> new; no visitor-facing risk from URL churn.)

## 2. CORRECTION to the handoff's framing — the mechanism is two spec flags, not a hand-seeded plan

`HANDOFF_2026-08-15b` §1 frames the work as "a seeded plan must name every tool slot,
and there are now 23". That is the right *success criterion* and the wrong *mechanism*.
Nothing in this lane's docs had yet mentioned `build-site-planner` or
`plan_includes_tools` — the sibling `loancalculator.co.uk` lane built the machinery on
08-12/08-13 and ran it on 08-15, and it works by seeding the site's **structure spec**
and firing the planner, not by writing `site_plans` rows by hand.

Two site-level switches decide whether an LMC replan is safe. Both default OFF, both
live in the `structure` aspect of `site_specs`, both are read through one helper each:

| key | reader | what OFF means for LMC | evidence |
|---|---|---|---|
| `plan_includes_tools` | `build-site-planner` step `load_components` (migration `407`) | ~~the planner's component menu **excludes this site's own tool components**, so the plan cannot name a single one of the 23 calculator slots~~ **CORRECTED below — for LMC it costs exactly 3 components, not 23** | sibling has `true`; LMC has the key absent — measured below |
| `honour_realised_identity` | `siteIdentityPolicyFor` (`platform/orchestration/actions/site_identity_policy.go:81`), read by BOTH canonicalisation surfaces | a plan page derived from a realised page gets its name/URL **re-derived** from its role instead of keeping what the site serves ⇒ `mortgages-stamp-duty` comes back as `tool-stamp-duty` and is INSERTed as a second row — "the phantom, re-minted by the very pass that spotted it" (`bugs_open/215`) | file's own comment, lines 112–129 |

`url_shape` is the third key of that family and **LMC must NOT set it** (see §4).

Measured 2026-08-17 (`[MEASURED]`, the query is in the RUNBOOK):

```
domain                            plan_includes_tools   url_shape
loancalculator.co.uk              true                  flat
loanandmortgagecalculator.co.uk   (absent)              (absent)
```

> **⚠ CORRECTED 2026-08-17, same session, before anything was seeded — I had the
> `plan_includes_tools` half wrong, and the disconfirming evidence was already in my own
> notes when I wrote it.** I read the flag's absence on LMC, read what it does on the
> sibling (whose 12 calculators ARE `component_level='tool'` components), and carried that
> consequence across without checking LMC's own component levels against migration 407's
> actual SQL. Read the live query:
>
> ```sql
> -- build-site-planner, step load_components (abridged to the level test)
> AND ( component_level IN ('section','element')          -- UNCONDITIONAL
>    OR ( component_level = 'tool'
>         AND EXISTS (… ss.data->>'plan_includes_tools' = 'true')
>         AND id IN (… this site's page_components …) ) )
> ```
>
> Section- and element-level components are offered **unconditionally**; the flag gates
> only `component_level='tool'`. **LMC's 23 B2 calculator components are all
> `component_level='section'`** (measured earlier in this same session — 17 of 19 tool
> pages carry only section-level rows), so they are in the planner's menu with or without
> the flag. On LMC the flag decides the fate of exactly **three** components:
> `loans-consolidation`, `tool-affordability-complaint-checker` and
> `tool-overpayment-priority` — the only tool-level rows on the site, and two of the three
> are the improvement loop's brand-new tools. Still worth seeding, for those three and for
> every tool the generator makes from now on. **Not** the blocker.
>
> The cheap check that would have caught it: read the step's own `query` out of
> `agent_definitions` and compare it against the `component_level` census of THIS site —
> one query each, and I had already run the census. Logged in `WRONG_CALLS.md`.

So, corrected: **the blocker is `honour_realised_identity`, and it is measured on LMC's own
data in §2a.** Firing the planner at LMC today would keep the calculators in the menu and
still re-mint most of the site under names it does not use.

## 2a. The blocker, measured on LMC — 38 of 45 pages are not fixed points of the canonicaliser

`[MEASURED 2026-08-17]` by calling the real `datahelpers.CanonicalisePage` over the live
page list, through the descriptor the write path actually builds
(`write_site_plan_action.go:487` — `Role`=stored `page_type`, `Slug`=`firstNonEmpty(slug,
name)` which for a realised page is the stored NAME, `ParentSection`=
`parentSectionFromURL(url)`, `FlatURLs`=false because LMC has no `url_shape` key).
Harness in `identcheck/` (scratch, reproducible from the RUNBOOK), and it **self-tests
with a positive and a negative control** — a canonical tool page must come back unchanged,
a legacy one must move — so "nothing moved" cannot be reported by an inert harness.

```
45 active pages: 7 fixed points, 38 moved (name 17, url 38, type 0)
```

- **17 pages move by NAME**, every one of them a calculator:
  `mortgages-stamp-duty` → `tool-mortgages-stamp-duty` at
  `/mortgages/mortgages-stamp-duty/index.html`. Note the doubled segment — the slug is the
  stored *name*, which already carries its section, so the derived URL is
  `/loans/loans-standard-calc/index.html`, `/mortgages/mortgages-repayment/index.html`,
  and so on. A name that collides with nothing on `(site_id, name)` is INSERTed: 17 new
  rows, and the real calculators left behind them.
- **38 move by URL**, including all 13 guides (`/guides/x.html` →
  `/guides/x/index.html`) and both section hubs (`/loans/index.html` → `/loans.html`).
- **7 are fixed points**: `index`, `guides-index`, `legal`, and the four pages the
  improvement loop created on 08-15 (two tools + two guides). Everything the framework
  itself has made recently is canonical; everything from the 07-31 adoption is not.

`sync_pages` (`sync_pages_to_db`) is step 12 of 14 in the live `build-site-planner`
workflow, so this is not a plan-table-only concern — the run writes `pages`.

**The sibling did not need this flag** (`honour_realised_identity` is absent there too) and
that is not evidence it is unnecessary here: its pages were already canonical in name and
dir-flat in URL, so it had nothing to lose. **LMC would be the flag's first live consumer**,
which is exactly the "turn it on where the population has been measured" case its own
author's comment asks for.

## 3. THE FLOOR — captured before the first reseed, as §1 of the handoff requires

`[MEASURED 2026-08-17]`, site `ed633ada-f8af-424b-b4d4-8af79160dbcd`. **These numbers are
four pages higher than `HANDOFF_2026-08-15b` §0 records, and the handoff was not wrong**
— see §5.

| floor quantity | value | how |
|---|---|---|
| active pages | **45** (+1 archived = 46 total) | `SELECT status, count(*) FROM pages … GROUP BY 1` |
| pages carrying a `tool-*` slot | **24** | `sections::text LIKE '%tool-%'` |
| of which the B2 calculator slots (`tool-0`/`tool-1`) | **23** | the list is in the RUNBOOK; do not take it from memory |
| `page_type='tool'` pages | **19** | census below |
| arithmetic calculators the oracle proves | **23 on 18 pages** — PASS 170 / FAIL 0 / CONVENTION 6 / N/A 0, `--mutate expectation` → 0 pass / 161 fail / 15 N/A, both this session | `oracle.py` |
| required homepage links | 16, of which **6 deliberately missing** (owner ruling, stage-2 proof case) | `gate_page_links.py` exits 1 on purpose |

Page-type census (active): `tool` 19 · `guide` 13 · `content` 9 · `blog-post` 2 ·
`section-index` 1 · `landing` 1.

**A floor detail worth carrying into the plan comparison:** 5 pages carry a calculator
slot but are NOT `page_type='tool'` (`loans-application-tracker`,
`loans-credit-health-check`, `loans-damage-checker`, `mortgages-fact-finder`,
`mortgages-portfolio`). A plan that types them `tool` would move their URLs; a plan that
drops the slot would shrink the site. Neither is caught by counting pages.

## 4. URL shape — LMC is MIXED, and copying the sibling's spec would move 39 URLs

`nestedOrFlatURL` (`datahelpers/page_canonical.go:272`): **flat** = `/dir/bare.html`,
**nested** = `/dir/bare/index.html`. Absent key = nested (`siteUsesFlatURLs`, and its
comment states the default explicitly).

LMC today `[MEASURED]`: **39 flat** (17 tool + 13 guide + 7 content + 2 blog-post) and
**6 nested** (2 tool + 2 content + landing + section-index). The two nested tool pages are
the improvement loop's own, created 08-15.

So there is no single correct value of `url_shape` for this site, and that is fine:
with `honour_realised_identity=true` a realised page keeps its stored URL, and the flag
then governs only **new** pages. The decision this plan takes is therefore:

- **leave `url_shape` absent** (new pages get the modern nested shape, matching the two
  the improvement loop just made), and
- **carry the existing 39 flat URLs by identity, not by shape.**

⚠ Do NOT seed `url_shape:'flat'` by copying the sibling: it is right for
loancalculator.co.uk (26 flat pages, `bugs_open/241` was exactly this) and wrong here.

## 5. Why the floor moved between the handoff and this session — and what it means for D6

`HANDOFF_2026-08-15b` recorded 41 active / 23 tool slots at 19:18 on 08-15. Both were
correct then. Between 19:28 and 19:35 **the same evening**, the improvement loop created
four pages — `tool-overpayment-priority` + its guide, and
`tool-affordability-complaint-checker` + its guide — and the whole site was re-rendered
between 20:57 and 21:13. Separately, four site_specs were rewritten that evening by
`domain-research-classifier`, `tool-suggester`, `domain-strategist` and
`vertical-exemplar-researcher` (18:18–18:45), and the daily `evidence-refresher` has
written a fresh `evidence_base` row each morning since.

**This is the ruling working as intended** ("growth from the improvement loop is
welcome"), and it has a consequence for how D6 must be executed: *the floor is a moving
target measured in hours, not days.* So the comparison "is the plan reasonably close to
where we are?" has to be made against a floor captured in the **same session as the
plan**, and any figure quoted from a handoff must be re-measured before it is used as a
pass/fail. The four queries in the RUNBOOK exist for that.

Two follow-ons this drift exposed, neither of them blocking D6:

1. **The two new tools have no oracle coverage.** `oracle.py` covers the 23 B2
   calculators on 18 pages; `tool-overpayment-priority` and
   `tool-affordability-complaint-checker` are outside it. `[UNVERIFIED]` whether either
   computes anything the oracle should be proving — that is a read of their templates,
   not an assumption, and it is the first thing to do after D6's first dispatch.
2. **The 08-15 classifier rewrite of `content_direction` did NOT clobber the lane's
   seeded voice** — checked, because it looked like it might have. The current row
   (`d655009c`, `domain-research-classifier`, 08-15 18:27) carries the lane's own heading
   rule and negativity-reframe rules verbatim inside it. It does contain one internal
   contradiction (`heading_style.format` allows "a genuine question"; `writing_rules`
   says headings are "never questions") — recorded here, not fixed, not this lane's item.

## 6. Phasing

**Phase 0 — before anything is dispatched (this session).**
- [x] floor captured (§3), oracle + control green in one session
- [x] the two flags' state measured on LMC and on the sibling (§2)
- [x] **identity population measured** — §2a: 7 fixed points, 38 moved, 17 by name, with
      both controls firing. Done by calling the real helper, not by re-deriving the rule
      in SQL, which is the drift this platform has been bitten by twice.
- [x] confirm no other session has LMC replan work in flight — `who-owns.py 263` shows
      only this lane's own commits; grepped the live `.jsonl` transcripts too (four
      sessions mention the domain today, all incidentally: a fleet register census, the
      `283`/RFC_034 lane, and portfolio-positioning seeding a different domain).
      Chassis pods are ~14 h old, so the ~300 s post-restart dispatch rule is clear.
- [ ] **known stale input, not yet resolved:** the structure spec's own `pages` list holds
      **41** entries (adoption-written 2026-07-31) against 45 active pages. `read_specs`
      hands that list to the planner prompt. Whether it biases the plan is `[UNMEASURED]`;
      it is not a blocker for the canary, and the canary's own divergence report is the
      cheapest way to find out. Do not "fix" it by hand first — that would destroy the
      evidence of what the planner does with a stale list, which is a thing we want to know
      about every adopted site, not just this one.

**Phase 1 — seed the two flags** into the `structure` aspect, supersede-then-insert:
`SEED_2026-08-17_identity_and_tools.sql` (written, **not yet applied** — the dispatch
decision in phase 2 is the owner's). It carries both keys with their measured
justifications, and a `DO`/`RAISE` verify block, because a verify block made of plain
`SELECT`s cannot stop the `COMMIT`. Modelled on the sibling's two seeds
(`loancalculator_couk/SEED_2026-08-11_url_shape_flat.sql`,
`SEED_2026-08-14_plan_includes_tools.sql`).
- `honour_realised_identity: "true"` — the blocker; population measured in §2a.
- `plan_includes_tools: "true"` — worth 3 components today (§2's correction), and every
  tool the generator makes from here.
- **NOT** `url_shape` (§4), **NOT** `twin_identity_snap`/`stem_twin_snap` — the two twin
  layers are a separate question with their own unmeasured collapse population, and
  seeding them alongside would make one canary answer four questions at once.
DB config is live immediately; no roll needed. ⚠ Two standing cautions inherited from the
sibling's seeds: `write_site_spec` ignores and drops `pinned`, so do not rely on it; and
check the keys survive after any adoption run.

> **PHASES 1–3 RAN 2026-08-17 12:03–12:30Z. Read NOTES (d) for the full account.**
> Seed applied (structure row `6ca809d6`). Canary fired (corr
> `6fe6ee93-67b9-4831-bf17-2ca473e1d30c`), COMPLETED in 3m19s.
> **Result: the plan's SHAPE passed and the write-back failed.** 45 planned pages with the
> role mix exactly matching the live census — which is D6's actual target — but the run
> INSERTED 19 phantom pages, moved 21 real URLs, cleared `sections` on 24 real pages and
> repointed 2 nav links. Repaired from the snapshot inside one guarded transaction; identity
> digest is back to the pre-fire value byte for byte, oracle 170/0/6, and the live site never
> served any of it (phantom paths 404). `rebuild_policy='owned'` protected all 17
> calculator pages — the run filed a review item per page instead of touching them.
> **Two causes, one filed and one identified:** the blanked sections are `bugs_open/204`'s
> positional-slot blindness reaching the plan-write path (NOT `282`, which is about
> tool-level functions); the phantom twins happened *with* `honour_realised_identity='true'`
> already set, so that link is **not asserted** — 090 filed, run correlation
> `33d4d7bc-62f8-4886-a8e2-7c39f0c0a302`.
> **Phase 4's first change is a correction to this plan's own phase-1 decision:** seeding the
> identity-preservation flag while deliberately withholding `twin_identity_snap` /
> `stem_twin_snap` turned on an effect without its precondition. One canary answering one
> question was the wrong economy here.

**Phase 2 — canary the planner ONCE**, reusing the sibling's script rather than writing a
new one: `loancalculator_couk/canary_replan_407.sh` (adapt the site id/domain; it already
prints its own judging queries, and it carries the two traps — `kcat -P` exits 0 having
sent nothing, so prove dispatch by the orchestration row **by correlation**; and no
dispatch within ~300s of a chassis pod restart). Budget tens of minutes for queue latency.

**Phase 3 — judge the plan against the floor.** The success criteria, in order:
1. the plan names all 23 calculator slots (§3's list, not a count);
2. no page identity moved that should not have — `md5` of the pages census before and
   after, exactly as the sibling's script does it;
3. page count did not shrink;
4. `oracle.py` 170/0/6 and the mutation control, re-run after, in one session.

**Phase 4 — reseed and repeat** until (1)–(4) hold. Record each round's divergence in
NOTES; the divergences are the product here, not the final plan.

**Phase 4, as it now stands after round 1 (2026-08-17, NOTES (e) — cause ESTABLISHED):**

1. **Seed `twin_identity_snap` and `stem_twin_snap` alongside the identity flag**, superseding
   the round-1 decision to withhold them. `honour_realised_identity` is **inert without a
   pairing layer**: `v3_site_actions.go:6476` strips `identity_authority` from every LLM page
   and it is *"re-stamped only by a snap or a union"*, so with no snap the flag is never
   reached. The stem layer is the one that matters for this site's shape — bare realised name
   vs prefixed plan name, in either direction.
2. **Put the page-level diff INSIDE the canary script.** Round 1's identity digest said
   *something* moved; the snapshot diff (which columns, which rows) had to be improvised
   afterwards. It belongs in the script, next to the digest.
3. **Expect the sections to stay empty until `bugs_open/204`'s class is fixed at this path.**
   Round 1's plan carried 10 sections, all for framework-built pages. A faithful plan for this
   site is not reachable while positional slot names resolve to nothing at the plan write — so
   judge round 2 on identity and page count, and treat sections as a known outstanding gap
   rather than a round-2 failure.
4. **Do not re-run without re-measuring the floor** (§5) and without a fresh snapshot: the
   pre-fire digest is the only thing that makes the repair assertable.

## 7. Council scope

Site config + lane tooling (spec rows, docs, a shell script) — out of gate scope, per
`HANDOFF_2026-08-15b` §"Council note". If phase 1 or 3 turns out to need a *code* change
(a new action, a schema change, a shared-seam field), that changes: register it in the
concept register in the same commit and submit to the gate.

## 8. Arrivals triaged this session (from the `register_guards_code_phase_b` lane)

`CONTRIB_2026-08-16_your_tier4_fences_are_ineligible_post_b2_and_a_facts_declaration_awaits_your_register.md`,
both items answered in `CONTRIB_REPLY_2026-08-17_…` in their lane:

1. **Their finding is CONFIRMED — and it is the THIRD independent measurement of it, not
   the first.** This lane's own NOTES already carries it: the `bugfix_281` lane measured
   it here on 08-15 ~18:30Z (4 eligible then) and wrote down both the mechanism and the
   two things to weigh before "fixing" it. Today `[MEASURED 2026-08-17]`, using the
   ladder's own predicate verbatim: **6 of 19 `page_type='tool'` pages are eligible, 13
   are not**, `mortgages-stamp-duty` among the 13, so its `computed_values` fence — a
   regression lock on `bugs_closed/225` — is never driven. Mechanism (unchanged from the
   08-15 entry): B2 decomposition raised those pages to 2–3 active components while
   none sits at `component_level='tool'`, so neither eligibility clause admits them.
   **The 4 → 6 delta is not drift in the predicate**: it is the improvement loop's two
   new tools, which the generator *does* create at `component_level='tool'`.
   ⚠ Before anyone "fixes" this by setting `component_level='tool'` on the 13, read the
   08-15 NOTES entry's own warning: `tool_health`'s fork branch files `improve_tool` for
   a tool-level component and **does not read `no_auto_fix`**, so the fix buys ladder
   coverage at the price of an automated fixer aimed at the calculators.
2. **Their §2 premise is stale, in our favour**: LMC *does* have an `evidence_base` row.
   It has had one since 08-15 22:04 (`7268d235`, seeded by the copy-quality lane), and
   the daily refresher has re-cut it twice since — current row `0c4b648a`, 13 facts, the
   same `sdlt-*` ids their Phase B expects. So the facts declaration is **unblocked**.
