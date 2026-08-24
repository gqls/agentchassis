# NOTES — `bugfix_378_usage_count_derived`

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-24 — session opens, lane claimed

`bugs_open/378` was filed 2026-08-24 by the `bugs_open/351` lane and left **OPEN, not started**.

**Ownership checked before claiming.** `scripts/who-owns.py 378` returns **OWNED or recently
active**, which is the known false positive: the only commits it can see are the three that
*filed* the bug, all by the 351 lane. That lane closed itself the same morning and its handoff
(`docs024_key_docs_latest/bugfix_351_section_template_predicate/HANDOFF_2026-08-22_continue_here.md`)
says in terms: *"`usage_count` path-blindness is a different defect, filed as `bugs_open/378` …
Not a 351 residual."* Its session transcript ends with the 351 closure. No `bugfix_378*` directory
existed. So: no active thread, lane resumed here.

### Re-validating the bug at HEAD before doing anything

The bug is **still valid**, and the mechanism is unchanged at HEAD. But three of its structural
claims are now wrong or incomplete, and one of its ranked candidates is refuted outright.

#### 1. There are THREE resolution paths, not two `[MEASURED 2026-08-24, read at HEAD]`

`plan_sections_action.go`'s section loop:

| path | line | how it resolves | counts? |
|---|---|---|---|
| **Path 0** | ~1208 | the page's own stored `page_components.component_id` (`bugs_open/204`) | no |
| **Path 1** | ~1258 | direct name/function match against the `components` map | no |
| **Path 2** | ~1279 | `resolveSectionComponent` → the `section_type` selector | **yes** |

The bug file describes two. Path 0 landed with `bugs_open/204` (`d376ca9b8`, "ONE reader for a
page's stored slot identity") and is tried *first*, so it is the path a settled, rebuilt page
most often takes — the population least likely to be counted is also the most established one.

#### 2. A THIRD reader, and it decides more than a score `[MEASURED 2026-08-24]`

The bug file names the two scoring queries (`component_selector.go:181`, `:235`). It misses:

```
platform/orchestration/actions/load_existing_component_action.go:170
    ORDER BY usage_count DESC NULLS LAST, updated_at DESC
    LIMIT 1
```

That query picks the **canonical contract row** for a `section_type` — by its own comment, *"the
row the store will overwrite and enforce"*. So the biased column is not only a 0.1 nudge in a
score; it is the primary sort key deciding which component *is* the contract. Higher stakes than
anything the bug file records.

Full census of the column, repo-wide, `[MEASURED 2026-08-24]`:
- **Writers: 2.** `IncrementUsageCount` (`component_selector.go:133`, the only incrementer, one
  non-test call site at `plan_sections_action.go:1957`), and the birth INSERT at
  `store_generated_component_action.go:639` which writes the literal `0`.
- **Readers: 3**, all filtering `component_level='section'`.
- No DB trigger, function or view touches it (`pg_proc`/`pg_trigger` checked; the only trigger on
  `content_components` is `trg_cc_refuse_null_section_type`, migration 581 / CLC-029 from the 351
  lane). The three `agent_definitions` config hits are prompt text listing `060`'s filename, not
  queries. So the bug file's `[INFERRED]` "no third writer" **holds** — now measured, not inferred.

#### 3. The counter OVER-counts as well as under-counting — and this refutes fix candidate 2

This is the finding the bug file does not have, and it changes the fix.

`IncrementUsageCount` is called inside `resolveSectionComponent` **before** `planSection` returns
and therefore before the section's status (`ready` / `deferred` / `skipped`) is known, and before
any `page_components` row is written. It also fires again on every re-plan of the same page. So
the column counts **resolution attempts on one path**, not usages.

Proof at the artefact, not from the code alone `[MEASURED 2026-08-24]`:

```sql
SELECT name, usage_count,
       (SELECT count(*) FROM page_components pc WHERE pc.component_id=cc.id) AS live_bindings
FROM content_components cc
WHERE is_active AND forked_from IS NULL AND component_level='section' AND COALESCE(usage_count,0)>0
ORDER BY usage_count DESC;
```

| name | usage_count | live bindings |
|---|---|---|
| `bayesian-ranking-hero-tool_pre_037` | **20** | **0** |
| `case-studies-grid` | 19 | 4 |
| `testimonials-modern` | **12** | **0** |
| `contact-block` | 7 | 6 |

**The two largest values in the column are both non-usages.** `testimonials-modern`
(`202fe796-3a39-449f-9dbb-697726fbf5f1`) was **created 2026-08-23** and has **zero**
`page_components` rows *ever* — checked live, checked with `build_status='removed'` included, and
checked against the `page_components_bak_20260820_preplan_lmc` snapshot. Twelve counts, no binding,
in one day. `bayesian-ranking-hero-tool_pre_037` is a pre-migration-037 **backup copy** that is
still `is_active` and still a live selector candidate, carrying the highest count in the table.

**So the bug file's candidate 2 — "call `IncrementUsageCount` on Path 1 too" — is refuted, not
merely weak.** It would propagate a definition that is already wrong in both directions to more
paths. Any fix that keeps this increment keeps the overcount.

#### 4. The dead half is bigger than the bug file measured `[MEASURED 2026-08-24]`

Census by level (active, non-forked). The bug file censused `section` only.

| level | rows | with any count | bound but zero | bindings invisible to the term |
|---|---|---|---|---|
| section | 151 | 12 | 98 | 1,865 |
| **tool** | **115** | **0** | 104 | 107 |
| header | 4 | 0 | 0 | 0 |
| footer | 1 | 0 | 0 | 0 |
| site | 5 | 0 | 0 | 0 |
| element | 1 | 0 | 0 | 0 |

`tool` is **0 of 115** — that half is `bugs_closed/060`'s dead counter exactly, not 378's
half-written one. It has no reader today (all three readers filter to `section`), which is the only
reason it is harmless.

#### 5. The scoring term is provably INERT today — and the control proves the measurement had power

The bug file's central `[UNMEASURED]`: *"whether the 0.1 term has ever actually flipped a
selection."* Run it.

Simulating the real scoring expression over every `(section_type, site_type, page_type)` context
drawn from the live vocabulary, comparing `argmax(base + usage_term)` against `argmax(base)`, ties
broken identically by `id` in both so a tie cannot masquerade as a flip:

```
contested_contexts | flipped | flipped_section_types | contested_section_types
       4888        |    0    |          0            |           4
```

**Zero flips.** But a zero is only evidence if it could have come out otherwise, so the same query
was re-run with a counterfactual granting the *weakest-base* candidate in each group a full 50 uses
(the maximum the term allows):

```
contexts | flips_detectable | section_types_flippable
  4888   |       52         |           1
```

The instrument detects flips. The zero is real. **The reason is mechanical:** only **4**
`section_type`s have more than one active non-forked section candidate — `hero` (7),
`tool-archetype-taster-quiz` (2), `tool-gripper-payload-calculator` (2), `features` (2) — and
**every candidate in all four carries `usage_count = 0`**. The 12 components that *do* carry a
count are all the sole candidate for their own `section_type`, so they win regardless.

**Consequence, and it sets the schedule:** a change that removes or replaces the term **cannot
regress any current selection**. That is a closing window — it holds only while contests stay rare
and uncounted.

#### 6. A maintained counter has seven places to be forgotten `[MEASURED 2026-08-24]`

`INSERT INTO page_components` sites outside tests:

```
platform/orchestration/actions/save_page_sections_action.go:1074
platform/orchestration/actions/deploy_tool_action.go:510
platform/orchestration/actions/create_tool_component_action.go:526
platform/orchestration/actions/create_report_page_action.go:268
platform/orchestration/actions/rebuild_blog_listing_action.go:389
platform/orchestration/actions/adopt_verbatim.go:531
cmd/webdesignport/import.go:239
```

Six in `platform/` plus one command. This is the `RFC_008` trap in memory almost exactly — that
census found ten writers of `page_components.rendered_html` and an eleventh was *born* two weeks
later. **So "maintain the counter where the binding is written" is the wrong shape**: it is seven
chances to forget and an eighth tomorrow. Derive at read; there is then no counter to drift.

#### 7. Prior art the bug file missed

`docs/agent_docs/sql_for_agents/325_directory_listing_binds_to_business_directory_query.sql:8-9`,
dated **2026-08-06** — eighteen days before 378 was filed:

> *"`usage_count=0` and, more reliably (usage_count is not a trustworthy signal fleet-wide —
> checked separately), zero pages anywhere carry this component"*

Someone had already established the column was untrustworthy and had already stopped relying on it
— and recorded it in a **migration comment**, where no grep of `bugs_open/`, the register or 016b
would ever surface it. Worth noting as a routing failure, not a fault of that author.

#### 8. Cross-thread: this is upstream of `bugs_open/107`

`bugs_open/107` — *every site gets the same homepage skeleton* — cites `component_selector.go:176`,
the neighbouring scoring term. The usage term is a **preferential-attachment loop**: selected →
count rises → scores higher → selected again. It is inert today only because contests are rare
(finding 5). **Any 107 remedy that adds candidate variety per `section_type` immediately activates
a term that systematically favours the incumbent** — so 107's fix would be partly undone at the
moment it starts working. 378 should land before or with it. Raised with that lane.

### Method note

`090` diagnosis run filed: intake `d2072b88-67ac-48ca-a319-6be6d543aae7`, run correlation
**`1c62c1f7-cb06-4482-933e-8c08a622b5c1`**. Framed at `component_selector.go` (13.7KB) and
`load_existing_component_action.go` (15.7KB) rather than `plan_sections_action.go` (114KB),
per the LANDMINE that a `090` on a symbol in a file over ~60KB returns bundles and no verdict.

---

## 2026-08-24 later — session checkpoint (usage limit), state as it actually stands

**Nothing has been changed in `platform/`. No code written, no migration written, no council
submission made.** Only the two lane docs exist. Recording this plainly so the next session does not
go looking for work that was never done.

### In flight when this session stopped

- **`090` diagnosis run — STILL RUNNING, no verdict.** Run correlation
  `1c62c1f7-cb06-4482-933e-8c08a622b5c1` (intake `d2072b88-67ac-48ca-a319-6be6d543aae7`). Last
  observed at `current_step='route'`, work item `status='diagnosing'`, `result` still `{}`.
  Read it before acting on the design:
  ```sql
  SELECT status, jsonb_pretty(result) FROM site_work_items
  WHERE spec->>'dispatch_correlation_id'='1c62c1f7-cb06-4482-933e-8c08a622b5c1';
  ```
  ⚠ Its verdict is an **independent check on the mechanism**, not on the fix. My claims 1–8 above are
  first-hand and stand on their own evidence whatever it says — but if it REFUTES any of them, that
  is a `WRONG_CALLS.md` entry and a correction here, not something to argue with.

- **The `fable` planning agent DIED before delivering.** It failed on an API session limit with its
  last step reported as *"store tests pinning the INSERT, the pages→site_id join shape, and the
  register's latest CLC number"*. **There is no plan document from it and nothing of its output was
  read.** `[UNVERIFIED — no artefact]`. Do not quote it, and do not assume its conclusions matched
  mine. The design thesis below is **mine, from the measurements above**, and has had no independent
  review of any kind yet.

### The design thesis, stated as a thesis and not as a decision

**A usage figure must be derived from the durable record of the usage, never maintained by whichever
code path happened to notice.** Concretely: stop reading `content_components.usage_count` in all
three readers; derive the signal from `page_components`; stop writing the column; decide separately
what happens to the column itself.

Why this shape rather than the bug file's ranking:
- Candidate 2 (increment on the other paths too) is **refuted** — finding 3. The increment's
  definition is wrong, so spreading it spreads the error.
- Candidate 1's *"or a counter maintained where the binding is written"* variant is the
  seven-writer trap — finding 6. Derive-at-read has no counter to drift.
- Finding 5 says the swap is **free today**: 0 selection changes over 4,888 contexts, with a control
  that detects 52. That window is the argument for doing it now rather than after the library grows.

### Open design questions the next session must answer, NOT yet answered

1. **What counts as one "use"** — a raw `page_components` row, a DISTINCT page, or a DISTINCT SITE?
   Distinct sites is the truest reading of "battle-tested" and resists a single page being rebuilt,
   but `page_components` has **no `site_id`** — it joins `page_id` → `pages` → `site_id`. That join
   shape needs checking before any SQL is written.
2. **Whether `build_status='removed'` rows count.** A removed binding is arguably not a use.
3. **The normalisation.** The term is `LEAST(count/50.0, 1.0) * 0.1`. If the unit changes from
   "resolutions" to "distinct sites", `/50` is certainly wrong — no component is on 50 sites.
4. **Shape of the query** — correlated subquery vs LATERAL vs pre-aggregated CTE in the batch
   selector, which runs per page build with an `IN` list. `idx_page_components_template
   btree(component_id)` exists and the table is only 2,038 rows, so this is a correctness-and-clarity
   choice more than a performance one.
5. **The column and the helper** — leave, neutralise, or drop. A drop is a migration, which is
   council-scope here.
6. **The other component levels** — `tool` is 0 of 115 and has no reader today. Fix now or leave?

### A dependency found while notifying other lanes — record it in the design

**`bugs_open/357` is a data-quality dependency on this design.** It documents `page_components` rows
whose `component_id` is dishonest — *"a whole tool page stored in a slot that claims to be the shared
`hero` component"*, 9 rows as of its filing. Making `page_components` the usage substrate turns those
into **phantom votes**, and `hero` is the most contested `section_type` in the library (**7** active
non-forked candidates as of 2026-08-24). This does not sink the design — the status quo is wrong for
96 section components and counts 12 uses for a component with no bindings at all — but it must be
**stated in the submission**, not discovered by a reviewer. That lane has been messaged; its own root
cause is still in the diagnosis loop (`63d4d1a7-ffec-4570-866b-8a0a41e3c69d`), so whether the
population is closed or growing is **not yet known**. `[UNMEASURED]` — I did not re-count 357's
population myself.

### Not yet done, owed by this lane

- Cross-thread note into `bugs_open/107` (the preferential-attachment interaction, finding 8) —
  **messaged nobody, written nowhere yet.** 107 has no live session.
- The correction block on `bugs_open/378` itself (three paths not two, third reader, overcount,
  candidate 2 refuted, the two `[UNMEASURED]` items now measured).
- Council submission. Nothing submitted.
- No `WRONG_CALLS.md` entry is owed **yet** — no claim of mine has been refuted so far this session.

---

## 2026-08-24 — the 357 lane replies, and it settles the substrate question the OTHER way

The `bugs_open/357` lane answered the message. Three of its corrections are accepted, one of its
recommendations is **declined on measurement**, and the reasoning for declining is the more important
half — so it is written out in full rather than summarised.

### Accepted, and re-measured here rather than taken on trust

- **The phantom population is 22, not 9.** `[MEASURED 2026-08-24, this lane]` — I ran their
  predicate myself: **22 rows, 0 stamped, 1 distinct component (`hero`)**. My earlier `[UNMEASURED]`
  marker on the "9" was doing its job; the 9 was the figure in 357's opening prose, which its own
  re-runnable query had already contradicted. **Budget 22.**
- **The population is NOT closed** — 12 born 08-23, 1 born 08-24 (their figures, not re-measured
  here `[UNMEASURED by this lane]`). So it is a live mint, not a historical residue.
- **Do not wait on diagnosis `63d4d1a7`** — it returned **UNVERIFIABLE**, and their narrower re-run
  `1ca712e3` produced bundles and no verdict. Their writer attribution comes from a row fingerprint
  instead (`save_page_sections_action.go`, `content_brief` present + `position=1`).
- ⚠ **Do not commit `v3_site_actions.go` from the working tree.** They report it carries a third
  lane's uncommitted WIP — a `bugs_open/345` call passing an 8th argument to
  `applyWorkItemFailureLadder`, whose committed signature takes 7 — so a pathspec commit of that file
  ships a HEAD that does not compile. This lane does not touch that file; recorded so nobody here
  does it by reflex.

### DECLINED: keying the derived signal on `component_version_id`

They recommend keying the derived score on `page_components.component_version_id IS NOT NULL`
(RFC_046, their phase 0, live since ~09:00Z 2026-08-24) — a stamp meaning *that component provably
produced those bytes* — on the grounds that all 22 phantoms are unstamped and structurally cannot
become stamped, so the phantom votes are excluded **by construction**.

The mechanism is real and I verified both halves. But it must not be the **primary** signal, and the
reason is that it would reintroduce the exact defect this bug is about.

**My measurement, with the control they suggested `[MEASURED 2026-08-24]`:**

```sql
SELECT (created_at >= '2026-08-24 09:00:00+00') AS post_roll, count(*) AS rows,
       count(*) FILTER (WHERE component_version_id IS NOT NULL) AS stamped
FROM page_components GROUP BY 1;
```

| | rows | stamped |
|---|---|---|
| pre-roll | 1,795 | **0** |
| post-roll | 249 | **245** |

The control holds — nothing backfills, so the stamp is honest going forward. (Their figures were
239/1051 and 0/987; mine are 245/249 and 0/1795. The shape agrees — ~0 before, ~98% after — the
denominators differ because we cut the window differently. Not a disagreement of substance, but I
am quoting mine, which carry their own control.)

**What each candidate definition would actually SEE today, section level:**

| definition | components with a signal, of 151 |
|---|---|
| any non-removed binding | **108** |
| distinct sites | **108** |
| **stamped binding only** | **39** |
| today's `usage_count` | 12 |

**The argument for declining, and it is not the coverage number.** A stamped-only signal counts a
component only if it has been re-rendered since ~09:00Z on 2026-08-24. So it measures **how recently
a component was rebuilt**, not how proven it is. A long-stable component that has not been
re-rendered since the roll reads as unproven — which is *precisely* the failure this bug describes,
with "was reached by the counting route" swapped for "was touched after the stamping roll". Adopting
it would fix the current instance by installing the same shape one epoch over, and the next session
to measure it would find a perfect correlation between "has a score" and "was rebuilt recently",
just as this one found a perfect correlation between "has a count" and "has a `section_type`".

**And the stamp cannot do the job it was offered for.** It excludes the 22 phantoms by construction
— but it also excludes the ~1,750 *honest* pre-roll bindings, because those are unstamped too. So it
does not separate dishonest rows from honest ones; it separates recent rows from old ones, and the
phantoms happen to be recent-but-unstamped. As a filter for contamination it is not selective today.

### The design, updated

**Primary signal: distinct non-removed bindings, derived at read.** Coverage 108/151 versus 12 today,
no counter to drift, and no epoch bias.

**Contamination, stated and bounded rather than filtered:** 22 of 2,038 rows (**1.1%**), confined to
**one** component (`hero`), and `578_..._HOLD.sql` (committed, deliberately unapplied) re-types them
onto an `adopted-fragment` component — after which the derived score self-corrects with no guard
needed here. This goes **in the council submission**, not in a reviewer's discovery.

**The stamp becomes the primary signal later, not now**, once coverage matures — which is an argument
for putting the definition in **one** place (a helper or a view) so the switch is a single edit
rather than three. That single-definition point is the framework-wide half of this fix and should
survive whatever the council says about the rest.

**Residual inherited knowingly** (their §5): `resolveComponent`
(`rerender_page_sections_action.go:377`) falls through to the slot-**name** map when `component_id`
is empty, so a row failing adoption gets re-bound to `hero` by the next rerender. Their phase 2
reduces the mint but does not close that path, and it ships default-OFF. A derived count is only as
honest as that path — recorded here as a known limit of the design, `[UNVERIFIED by this lane]`.
