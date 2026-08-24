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

---

## 2026-08-24 — 357 concedes, and hands over the one thing I could not have derived

They accepted the decline and logged their own wrong call: the claim *"excludes the phantom votes by
construction"* was **true**, which is what made it dangerous — they had measured what the filter
**rejects** and never what it **keeps**. Worth carrying as a general check, and it is close to this
lane's own family: **a filter justified by what it removes needs the count of what survives it.**

### CORRECTION to my own entry above

The previous entry says `578_..._HOLD.sql` retypes the 22 *"after which the derived score
self-corrects with no guard needed here"*. **That is too strong and I am correcting it rather than
editing it away.** Their phase 2 (which stops the mint) is confirmed live in the running binary but
ships **default-OFF and unarmed**, and arming is the owner's decision. So phase 3 would be repairing
a population that **refills**. My contamination figure is a **floor with a growth rate**, not a fixed
1.1%. What I wrote implied a one-off cleanup; it is not one while phase 2 is unarmed.

### Their claim about the age spread, re-measured here — and it strengthens the decline

They report 6 of the 22 born in June (`gamesdesign.co.uk`, `rebuild_policy='owned'`, stable because
the owned-page guard refuses automated saves). I cut it by birth month instead:

```
 born       | count
 2026-06-01 |     8
 2026-08-01 |    14
```

`[MEASURED 2026-08-24, this lane]` — **8** June, **14** August, against their 6/16. Not a
contradiction: theirs is a cut by `rebuild_policy='owned'`, mine by birth month, and 6 owned rows can
sit inside 8 June rows. Both support the same conclusion, which is the point that matters:

**Age does not separate the dishonest rows either.** Unstamped-and-old and unstamped-and-recent each
contain honest and dishonest rows. So no recency-derived filter — including the stamp at today's
coverage — can do the job; only the stated-and-bounded framing survives it.

### The insight I could not have derived, and it changes the helper's comment

**Stamp coverage will approach but never reach 100%, and that is by design.** A row is stamped when
it is *written*, and nothing backfills. So coverage is a pure function of **rebuild cadence** — a
component's bindings convert only as its pages are re-rendered, and **a page that is never rebuilt is
never stamped**. The population that converts LAST is the long-stable component, which is exactly the
population the primary signal exists to credit.

**Therefore the switch condition is not "coverage is high".** It is *"the unstamped remainder is small
enough to be a stated exception rather than the majority"*. That distinction cannot be recovered from
a coverage percentage alone, so **it goes in the helper's own comment when the helper is written**,
next to the definition it governs — not in a lane doc nobody will open.

### Routing the switch through the durable record instead of a session

They asked — correctly — not to be the channel: sessions here are ephemeral and will not be around to
answer. The stamp's register entry is **CLC-026**
(`docs/agent_docs/docs026_concept_register/register/component-lifecycle.md`), with **CLC-028** as the
carriage contract that made it actually arrive. A note there naming this lane as a downstream consumer
means the next session to touch the stamp finds us. Done in this commit.

⚠ Note while reading CLC-026: its own `verify-later` already anticipates this population —
*"`bugs_open/357`'s population is rewritten roughly daily, so within a day its re-minted rows must
still read `component_version_id IS NULL`; a population row appearing WITH a hero stamp means the
splice hygiene failed"*. So a future stamped hero row is **their** alarm, not a windfall for my
signal — do not read it as coverage improving.

---

## 2026-08-24 — the fix, and the misstep that produced it

**Shipped in code, inert until the next chassis roll.** Commit `5074367f7`, council
`ca01b81a` (`Council-Submitted:`, verdict unread). HEAD verified to build with
`scripts/verify-head-builds.sh` (HEAD had already moved to `7d18c8c83` — another session
committed in between; the working tree was red from a third lane's WIP, which is what the
commit hook's "the tree does not build" line was about, **not** this change).

### The misstep, which is the most useful thing in this file

I measured *"what happens if the term is REMOVED"* — 0 of 4,888 contexts — and then wrote,
in three places, *"a change that removes **or replaces** the term cannot regress any current
selection."* The measurement compared `base + term` against `base`. **It licenses deletion and
nothing else.** When I finally ran the comparison against the replacement I had actually built,
it was **3,246 of 4,888** winners across 3 section_types.

What caught it was not a test and not a review — it was **running the query I had just written
against the live database and reading the output**. The five `hero` candidates came back at
27 / 22 / 18 / 6 / 4 sites where the old column had them all at `0`, and the spread was visible on
the row. `go build` passed throughout. Logged in `WRONG_CALLS.md` with the cheap check: **state the
two versions the number compares, inside the claim** — "removing changes 0 winners" cannot be
stretched; "removing or replacing" only reads as supported because both verbs sit in one sentence.

### And it inverted the design, for the better

Being forced to measure the replacement honestly is what killed it. **A *working* usage term is a
preferential-attachment loop** — selected → count rises → scores higher → selected again — and
`bugs_open/107` is the standing complaint about exactly that outcome, citing this file. So the
accurate version of the term makes the estate *more* homogeneous. Repairing the measurement would
have closed this bug while making an open one worse, under the smaller change's evidence.

**So the fix is bug-file candidate 3 (stop scoring on it), not candidate 1.**

| | winners changed, of 4,888 contested contexts |
|---|---|
| remove the term | **0** |
| repair the term with the derived number | **3,246**, across 3 section_types |

### What actually shipped

1. `IncrementUsageCount` **deleted**, and its only call site in `resolveSectionComponent` replaced
   with the reason (it fired before `planSection` decided anything).
2. `ComponentUsageSitesSQL` — **one** named constant, `count(DISTINCT p.site_id)` over
   `page_components` joined to `pages`, excluding `build_status='removed'`. Distinct **sites**, not
   binding rows: `[MEASURED 2026-08-24]` raw bindings run max 414 / median 1 (a near-binary signal
   reporting "is this on a big site"), distinct sites max 27 / p90 9.
3. **The scoring term is gone** from both selector queries. The derived figure is still SELECTed as
   `usage_count` so it is logged and honest — nothing scores on it.
4. `load_existing_component_action.go`'s **contract-row** `ORDER BY` uses the constant. This is the
   reader the bug file missed and the only place with a real behavioural change: **2 of 4** contested
   section_types move, **both corrections** — `about-hero`→`hero`, and
   `archetype-taster-quiz`→`tool-archetype-taster-quiz`.

**Both rewritten queries were executed against the live DB via `PREPARE`/`EXECUTE`, not merely
compiled** — `go build` cannot parse SQL, and in this change the SQL *is* the change.

### Owed, and deliberately not done

- **The column still exists**, written by nothing and read by nothing in Go. Dropping it is a
  follow-up migration *after* this code is live, so the code cannot roll back onto a missing column.
  ⚠ **Until then it still reads as a maintained figure.**
- **Read the council verdict** (`ca01b81a`) and act on a REVISE/REJECTED — the code is already on the
  shared branch. The commit hook also raised an **architecture signal** for the removed exported
  symbol; the submission argues this is a point fix (`IncrementUsageCount` had one non-test call site
  and no other consumer, verified by grep across `platform/ internal/ pkg/ cmd/`), but the
  architecture seat may disagree and that is its call, not mine.
- **The `090` run never returned a verdict** — `1c62c1f7-cb06-4482-933e-8c08a622b5c1`, still
  `diagnosing` with an empty `result` after ~1 hour. So this fix rests on first-hand evidence only,
  which the 2026-07-31 ruling permits provided the substitute is stated plainly: the writer claim is
  an exhaustive grep of four Go trees plus `pg_proc`/`pg_trigger`/views/agent config, the reader
  claims are the three queries quoted verbatim, and every behavioural claim is a simulation over the
  live library with a disconfirming control named. **If that verdict lands and refutes anything, it
  is a `WRONG_CALLS.md` entry and a correction here.**
- A `LANDMINES.md` entry for the "increments before the outcome is decided, and again on every
  re-plan" shape — held deliberately until the fix is live so the entry can name the remedy
  (agreed with the 357 lane).

---

## 2026-08-24 — council verdict APPROVED, and what I owe against the objections

**`ca01b81a` — APPROVED round 1**, "approved with 2 advisory objection(s) — none high-severity".
Trailer already on the commit as `Council-Submitted:`, so `098` credits it automatically; no amend
(forward-only). Reading the full report, more seats objected than the headline counts. Answering
each, because an approved verdict is not a reason to leave a medium objection unanswered.

### ANSWERED — `editquality` (low): "removing IncrementUsageCount may break test call sites"

**No test ever called it.** `grep -rn "IncrementUsageCount" --include="*.go" .` across the whole repo
returns exactly one hit today: my own replacement comment. `go vet ./platform/orchestration/actions/`
(which type-checks the test files too) is clean apart from a pre-existing `unreachable code` warning
in `load_component_library_actions.go:207`, which is not mine. The objection was reasonable — my
submission said *"the only non-test call site"*, which describes the grep filter I used and reads as
implying test callers exist. Wording fault, not a defect.

### ANSWERED — `bug_historian` (medium): the contract-row swap re-shapes the enforced schema

The sharpest objection in the round, and it is **partly right**. The schemas genuinely differ:

| component | fields |
|---|---|
| `hero` (new contract) | **7** |
| `about-hero` (old contract) | **2** |
| `tool-archetype-taster-quiz` (new) | **22** |
| `archetype-taster-quiz` (old) | **1** |

So this is not a cosmetic re-ordering. **What bounds it:** `LoadExistingComponentAction`'s only
consumer is the `component-creator` agent (confirmed — one `agent_definitions` row names it), and the
action feeds a **prompt** (`field_names`, `function`, `field_count`, behind a
`{{if .existing_component.field_names}}` guard). Its own doc comment says it must
*"never block generation on a lookup problem (the store-time guard is the backstop)"*. So it steers
what the creator generates; it does not itself overwrite a stored component or re-type a bound page.
The change points that guidance at the row most sites actually use rather than the most recently
touched one, which is the intent.

⚠ **Residual, stated not closed:** steering the creator at a 7-field contract where it previously saw
a 2-field one can still change what gets generated for pages bound to the old winner. I have not
traced the store-time guard's behaviour on that mismatch. **This is the first thing to verify after
the roll**, and it is recorded as owed rather than argued away.

### BLOCKED, NOT ANSWERED — `guardian` (medium): correlated-subquery cost on a hot path

Fair objection: this converts a stored-column read into `count(DISTINCT p.site_id)` over a join,
per candidate, in the path every page build takes. I tried to answer it with `EXPLAIN (ANALYZE)`
and **could not, for a reason that has nothing to do with this change** — see the incident note
below. `[UNMEASURED — instrument blocked]`. **Owed before this is called done.**

What is known and does not need the DB: `page_components` is 2,038 rows with
`idx_page_components_template btree(component_id)`, `queryCandidates` is `LIMIT 5`, and the batch
query is bounded by the section_types on one page. The earlier live `PREPARE`/`EXECUTE` of the real
selector query returned in well under a second, and the 4,888-context simulation (which evaluates the
subquery far more often than production will) completed inside 300s. That is suggestive, **not** the
plan the guardian asked for. If the plan turns out badly, the fix is a pre-aggregated CTE or a
`LATERAL`, not a revert.

### ANSWERED — `prior_art_librarian` (medium): the CLC-026 coverage figures are unverified

They are mine and they carry their own control: `[MEASURED 2026-08-24]` **0 of 1,795** pre-roll rows
stamped vs **245 of 249** post-roll. And the register claim is now true because I made it true —
CLC-026 carries the downstream-consumer bullet as of commit `1103c5cbd`. Checkable, and checked.

### ALREADY SATISFIED — `architecture` (low): "add an in-code TODO about CLC-026"

The seat read the plan sketch, not the code. `ComponentUsageSitesSQL`'s doc comment already carries
the whole CLC-026 handoff, including the switch condition. Nothing owed.
**`ARCHITECTURE_SIGNAL: point_fix`** — so the commit hook's RFC prompt on the removed exported symbol
is answered by the seat itself.

### OPEN — `reuse_agent` (low): was an existing "sites per component" aggregation missed?

Not searched before authoring. Cheap to check and honest to record as not done.

---

## 2026-08-24 — ⚠ LIVE DB INCIDENT, not caused by this lane, blocking every reader of `pages`

Found while trying to run the `EXPLAIN` above. **`SELECT count(*) FROM pages` alone times out at
15s**, which is what told me the problem was not my query.

```
pid 2007330  active  ClientWrite   01:01:30  COPY (SELECT translate(encode(convert_to(row_to_json(t)...base64...
pid 2016837  active  Lock relation 00:23:46  ALTER TABLE pages ADD COLUMN noindex boolean NOT NULL DEFAULT false;
pid 2016865  active  Lock relation 00:23:36  SELECT ... FROM pages p JOIN page_components pc ...
   ... and every other reader of `pages`, all 21-24 minutes, all waiting on Lock/relation
```

**The chain:** a `COPY` export has been running **over an hour** stuck on `ClientWrite` — it is
blocked writing to a client that has stopped reading, so it is not progressing and not releasing.
Behind it an `ALTER TABLE pages ADD COLUMN` is queued for an ACCESS EXCLUSIVE lock, and **behind that
every single reader of `pages` is queued**, because a pending ACCESS EXCLUSIVE request blocks new
shared locks even though the ALTER has not started.

Neither is mine — this lane has run no DDL and no export. Not touched, not killed: cancelling another
session's hour-long export or migration is destructive and is the owner's call, not a side-effect of
my measurement. Surfaced to the owner instead.

---

## 2026-08-24 — the guardian's cost objection, now MEASURED (it was the last thing owed)

The DB cleared, so the `EXPLAIN` that was blocked above has been run. **The objection is answered and
the answer is favourable, but the honest version includes the ratio, not just the absolute.**

`EXPLAIN (ANALYZE, BUFFERS)` against the live database, `[MEASURED 2026-08-24]`:

| query | execution |
|---|---|
| single-section (`queryCandidates`, `section_type='hero'`, 7 candidates) | **11.3 ms** |
| batch (`queryCandidatesBatch`, 10 section_types — a full homepage — 15 candidates, 903 page lookups) | **10.7 ms** |
| the OLD stored-column read, same batch | **0.36 ms** |

**So it is ~30x the old query in relative terms, and ~10 ms in absolute terms.** Both numbers belong
in the record: quoting only "10 ms" would hide a real regression ratio, and quoting only "30x" would
imply a problem that does not exist at this scale.

**Why 10 ms is the number that matters here:** the batch form runs **once per page build**, and a page
build is dominated by LLM calls measured in seconds. The plan is fully index-backed on both sides —
`Bitmap Index Scan on idx_page_components_template` for the binding lookup and
`Index Scan using pages_pkey` for the site join — so it scales with *bindings for the candidates on
this page*, not with the table. The guardian's specific question ("confirm candidate-set sizes stay
small enough that the subquery is a non-issue") is answered: **15 candidates and 903 page lookups for
a ten-section homepage, in 10.7 ms.**

⚠ **What would change this, i.e. the disconfirming condition to watch:** the cost is linear in
bindings-per-candidate, and `hero` alone already carries 105 bindings. If `page_components` grows by
an order of magnitude, or if a single component reaches many thousands of bindings, re-measure. The
remedy would then be a pre-aggregated CTE or `LATERAL` — **not** a revert to a stored counter, which
is the defect this bug is about. Recorded so a future session re-measures rather than reverts.

### Council objections — final state

| seat | objection | state |
|---|---|---|
| `editquality` | test call sites would break | **ANSWERED** — none exist; `go vet` clean |
| `bug_historian` | contract-row swap re-shapes enforced schema | **PARTLY ANSWERED** — bounded (consumer is a prompt + store-time backstop); residual owed post-roll |
| `guardian` | correlated subquery cost on a hot path | **ANSWERED** — 10.7 ms batch, index-backed, once per build |
| `prior_art_librarian` | CLC-026 coverage figures unverified | **ANSWERED** — mine, with control; register bullet is commit `1103c5cbd` |
| `architecture` | wants an in-code CLC-026 TODO | **ALREADY SATISFIED** in the shipped comment; seat read the sketch |
| `reuse_agent` | was an existing "sites per component" query missed? | **OPEN** — not searched; low |

**Still owed after the roll:** the `bug_historian` residual (does the store-time guard fail loudly or
silently when the new contract row's schema differs from what a bound page carries?), and the
`reuse_agent` search. Neither blocks the fix; both are written down rather than closed by assertion.

### Note on the DB incident above — I did NOT fix it

For the record, because it would be easy to read the sequence as a fix: with the owner's approval I
ran `pg_cancel_backend(2007330)` and it returned **false** — *"PID is not a PostgreSQL backend
process"*. The blocker had already gone between my measurement and my command. The pile-up cleared
(**83 of 91** backends waiting → **0 of 9**), **but not by my action.** Either it ended on its own or
another session cancelled it first. Claiming the fix here would be exactly the kind of
post-hoc-ergo-propter-hoc this lane spent the day arguing against.

---

## 2026-08-24 (post-roll) — LIVE and proven at the binary; the last two objections resolved, one of them INVERTED

### The fix is live — proven three ways, each with a control

Chassis build stamp `48f55f21834ac3e2d95aa43716f6e63e40ac12ee` (pod started 18:55:21Z), read from
`service_binary_capabilities` (`kind='build'`) because the `build provenance` startup line had
already scrolled out of `--tail=400`.

1. **Ancestry.** `git merge-base --is-ancestor 5074367f7 48f55f218` → **YES**.
   ⚠ **My first control was worthless and I nearly recorded it** — I picked `5f7d32e4f`, which also
   predates the build, so both arms returned YES and the test proved nothing. Replaced with
   `4ad5b10fb` (committed 19:55, after the build), which correctly reports **NOT an ancestor**. The
   test discriminates.
2. **The new SQL is in the binary.** `grep -aq 'count(DISTINCT p.site_id)' /proc/1/exe` → **PRESENT**.
3. **The old SQL is gone.** `grep -aq 'usage_count, 0)::float / 50.0' /proc/1/exe` → **ABSENT**,
   with a positive control (`component_selector: selected` → PRESENT) proving the grep is not blind
   on this binary.

### ⚠ NOT demand-proven, and the control is what says so

`usage_count` values are **byte-identical** to the 13:30Z snapshot — nothing has incremented. That is
the expected post-fix observation and **it is currently worth nothing**, because the demand control
fails: `SELECT count(*) FROM page_components WHERE created_at > '2026-08-24 18:55:21+00'` → **0**.
No page has been built since the roll, so the old code would also have incremented nothing.
`[UNMEASURED — no demand yet]`. **The frozen counter becomes evidence only once a page build has run.**

### `bug_historian`'s medium objection — INVERTED, not merely answered

The objection: changing `load_existing_component`'s ORDER BY switches which row is authoritative and
could silently re-shape the enforced schema for pages bound to the old winner. Right family
(the schemas really do differ — `hero` 7 fields vs `about-hero` 2), **but the direction is backwards,
and reading the file's own fallback is what showed it.**

`resolveContractViaStorageIdentity`'s doc comment states the design intent: the function name is
derived *"exactly as `store_generated_component_action.go` derives it"* so that **"the prediction and
the enforcement agree by construction rather than by coincidence"**. The store resolves what it will
overwrite by **function name** = `NormaliseToKebab(section_type)`. So the right question is not "does
the new winner differ from the old" but **"does the winner's `function` equal the section_type the
store will enforce"**.

`[MEASURED 2026-08-24]`, over all 117 section_types with a section-level candidate:

| ordering | predictions agreeing with what the store enforces |
|---|---|
| OLD (`usage_count DESC, updated_at DESC`) | **88** of 117 |
| NEW (derived sites DESC, `updated_at DESC`) | **90** of 117 |

And both changed types moved **from disagree to agree**:

| section_type | old predicted `function` | new predicted `function` | store enforces |
|---|---|---|---|
| `hero` | `hero-about` ❌ | `hero` ✅ | `hero` |
| `tool-archetype-taster-quiz` | `archetype-taster-quiz` ❌ | `tool-archetype-taster-quiz` ✅ | `tool-archetype-taster-quiz` |

**So the change REDUCED a pre-existing prediction/enforcement mismatch; it did not create one.** The
old ordering was the mismatched one.

**New finding falling out of that, unrelated to this bug and worth someone's time: 27 of 117
section_types (117 − 90) still predict a contract the store would not enforce.** The advisory tells
`component-creator` to preserve one row's field names while the store overwrites a different row.
That is latent, pre-existing, and not created by this change — **not filed as a bug yet.**

### `reuse_agent`'s objection — VALID, I had not checked, and prior art does exist

They asked whether an equivalent "sites using component X" aggregation already existed. I had not
looked. It does:

- `component_write_guard.go:437` — `SELECT count(DISTINCT pc.page_id), count(DISTINCT p.site_id)`
- `store_generated_component_action.go:1179` — `SELECT DISTINCT p.site_id::text`

**Not reused, deliberately, and the reason is stated rather than assumed:** those compute a *write
fence* blast radius (how much would this overwrite affect) over all rows; mine is a *merit* signal
that excludes `build_status='removed'` and counts sites only. Same SQL shape, different question and
a different predicate — sharing them would couple a scoring input to a safety guard. **But the seat
was right that I did not check, and "I didn't look" is the answer, not "there was nothing".**
