# NOTES — inline_guide_imagery (append-only, newest at the bottom)

Running technical record for the lane whose plan is
`PLAN_2026-08-14_durable_inline_guide_imagery.md`. Missteps are recorded here on purpose,
not tidied away.

---

## 2026-08-31 — picking the lane up, and what has moved under it in 17 days

Session started from the owner's instruction to take up this thread. The PLAN still opens
**"Status: design, nothing implemented"**, and two CONTRIBs landed against it today (one
from the boxingonline paid-build review, one from `editorial_design_uplift`). Both are
right that nothing shipped. But three facts the PLAN reasons from have **moved**, and one
of its load-bearing assumptions is now measurably false. All queries inline.

### 1. The component the PLAN proposed to fork ALREADY EXISTS — someone built it on 08-24

PLAN §7 Phase 1 says "Fork, don't edit: new `content_components` row
`article-body-illustrated` … + `figure_url` (`source: "site_assets.illustration"`)".

That component exists, under a different name, built by another lane
`[MEASURED 2026-08-31]`:

```sql
SELECT id, name, function, created_at, created_from FROM content_components
 WHERE function='illustrated-text-block';
-- 6322b121-2d04-4877-89ac-c1785a81ae84 | Illustrated Text Block | section | 2026-08-24 11:15Z | manual
```

Its `input_schema` is almost exactly what the PLAN specified: `heading` + `content` (llm,
required), `image_url` (**`source: site_assets.illustration`**, optional,
`on_missing: skip_field`), `image_alt`, `image_caption` (llm, optional). Its `content`
guidance even forbids the writer emitting `<img>`/`<figure>` inline — the anti-pattern
this lane exists to retire, already written into the field guidance.

**So Phase 1's "fork a component" step is DONE and was done by someone else.** Do not
build `article-body-illustrated`.

### 2. Migration 644 (applied 2026-08-26) made the planner able to SEE it

`docs/agent_docs/sql_for_agents/644_planner_sees_imagery_and_illustrated_block_sources_an_illustration.sql`
— applied 2026-08-26 11:16Z `[MEASURED]` (`schema_migrations`). Two halves: it taught
`component_expresses()` the token for an image so the planner menu can tell an illustrated
block from a plain one, and it repointed `image_url` from `site_assets.image` (which
alias-resolves to the page hero) to `site_assets.illustration`.

Its own closing note states the half it did not do, and it is this lane's half:

> "It makes the illustrated component SELECTABLE. It does not create a single asset. …
> The supply question is the bigger half and is NOT addressed here."

### 3. THE FINDING: the durable route the three lanes agreed on cannot bind an image to a SECTION

`dartsonline_traffic`'s 08-31 handoff, `news_editorial` and the PLAN's own 08-31 addendum
all converged on the same route: **one `illustrated-text-block` per h3, ordinary flat
sections, no composition**. I read the resolver to cost it and it does not work today, for
two independent reasons. Both are code reads at HEAD with the lines cited.

**(a) The resolver has no section identity at all.**
`plan_sections_action.go:481-518` (`ensureAssets`) loads the page's section-scope
`site_plan_imagery` rows — `WHERE spi.scope='section' AND spi.scope_ref LIKE $page||':%'` —
into ONE flat per-page map, `r.assets[assetKey]=url` plus a **kind first-wins** alias
`r.assets[kind]`. The `scope_ref` ordinal appears only inside that `LIKE`; nothing parses
it. And `sourceResolver.resolve` (:643) takes `(ctx, source)` — **a source string and
nothing else**. `planSection` (:2072) receives `sectionName` but no ordinal.

Consequence: six sections on one page all declaring `source: site_assets.illustration`
resolve to the **same** URL — whichever section-scope row sorts first by `(kind, ordering)`.
Distinct per-section images are not expressible through the declared source.

This is not a new discovery about the ordinal — `bugs_open/214` already recorded that
"no consumer parses the ordinal" and judged it *harmless*, which it was at one illustration
per page. What is new is that the route this lane needs is exactly the case where it stops
being harmless.

**(b) The carry-forward that would rescue stored values is switched off by repetition.**
`ensureStoredContent` (:185-245) keys the carry map by `page_components.slot_name`, and any
slot that repeats with **different** `content_data` is deleted from the map outright
("slot_name repeats with different content_data — not a carry-forward source"). And
`save_page_sections_action.go:1001,1033` writes `slot_name = section.ComponentName`. So N
sections of one component on a page ALWAYS collide, and the carry is always dropped for
them.

**(c) The live instance, and what it predicts.** apis.uk `/index.html` is the only page in
the fleet serving a distinct illustration per prose section `[MEASURED 2026-08-31]`:

```sql
SELECT pc.position, pc.slot_name, pc.content_data->>'image_url'
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE s.domain='apis.uk' AND p.url='/index.html' ORDER BY pc.position;
-- 6 rows, ALL slot_name='generic-text-block', six DIFFERENT illustration-*.jpg values
```

Its `page_components.component_id` points at `illustrated-text-block` while its
**current plan says `generic-text-block`** for all six — a plan/cache divergence, i.e. the
page is hand-maintained, not framework-produced. And its imagery rows are `scope='page'`,
not `scope='section'`, so `site_assets.illustration` resolves to **nothing** there.

Put (a)+(b)+(c) together and the prediction is: **on the next SAVE-path build of that page
(a content write, not a re-render), all six `image_url` values are dropped** — source
resolves nothing, carry is conflicted-and-deleted, `on_missing: skip_field` removes the
key. A re-render is safe (it merges stored ⊕ fresh, and fresh has no key to overwrite
with). Migration 644's header asserts the opposite — *"carryStored then preserves the six
good values"* — and that claim does not survive the conflict rule at :233-245.

⚠ `[UNVERIFIED AT THE ARTEFACT]` — this is a code-read prediction, not an observed loss. It
is filed to the diagnosis loop rather than asserted (below). The apis.uk lane's own NOTES
record that this page lost its images once already, on 2026-08-23, to a `page-content-writer`
sweep four minutes after they were verified live — same class, different route.

### 4. Filed to the diagnosis loop rather than asserted

Per CLAUDE.md ("always file when … you suspect it is cross-cutting / platform-wide"), the
mechanism above went to `090` before being written into any bug file:

- intake `CORRELATION_ID=90167d51-7e97-49ac-b510-dfbca7224298`
- **run `RUN_CORRELATION_ID=351a9b33-74bb-4e52-a027-2f7881b44412`** ← artifacts join on this

Queue checked first (`needs_diagnosis` / `awaiting_diagnosis` → empty), and
`bugs_open/`+`bugs_closed/` grepped for the mechanism (214 and 114 are the adjacent files;
neither covers the N-per-page case).

### 5. State of the motivating pages, today

`dartsonline.com/blog/grip-styles.html` — the owner's own example and the canary the other
lanes nominated `[MEASURED 2026-08-31]`: three sections (`hero`, `article-body`,
`call-to-action`), the whole article in one 8.4 KB `article-body` blob, zero content images.
Its plan agrees (`site_plan_sections`: hero / article-body / call-to-action).

Supply, fleet-wide `[MEASURED 2026-08-31]`: **4 section-scope `illustration` rows across 3
sites** (fundamentallyai `about:2`, idea.uk `index:1`, vonc `about:2` + `index:2`) and **1
section-scope `infographic`** (mortgagecalculator `scorecard-simulator:1`). None on
dartsonline. So the supply half of the PLAN is untouched as well.

### 6. Coordination checked before touching anything

`dartsonline.com` has an active lane that worked its heroes today (three regeneration passes,
committed this afternoon) and open items on the site including two
`save_refused_incomplete` needs_human_review rows and a failed pair of asset-landed
re-renders `[MEASURED 2026-08-31]`. Their handoff explicitly hands the per-section imagery
mechanism to this lane ("mechanism is NOT ours and NOT built"), so the split of work is
agreed, but **nothing is to be dispatched at their pages without saying so first**.

---

## 2026-08-31 (later) — the diagnosis came back NOT CONFIRMED, and it was right to

`RUN_CORRELATION_ID=351a9b33-74bb-4e52-a027-2f7881b44412` → **UNVERIFIABLE (stopped:
scope-not-narrowing)**. Recorded here in full because a refuted or unnarrowed verdict is
the cheapest place to be wrong, and this one found a real defect in my own symptom.

**What it caught, and it is my error, not the loop's.** I wrote one symptom containing TWO
mechanisms and named apis.uk as "the live instance" of both. It is not. The loop queried
`site_plan_imagery` for that page and found **zero `scope='section'` rows** — so the
kind-first-wins collision *cannot* be happening there, and it said so:

> "the data_request against site_plan_imagery for it returned zero scope='section' rows, so
> there is nothing there for the kind-first-wins map to collide on"

That is correct and I should have seen it: apis.uk's illustration rows are `scope='page'`
(I had measured that myself, in §3(c) above, and then still offered the page as evidence
for the collision half). **The collision half has NO live instance at all** — nothing in the
fleet has two section-scope figures on one page — which is exactly why it has never been
noticed. One symptom, one mechanism; and a live instance must be an instance of *the
mechanism you are claiming*, not of the neighbouring one.

**What it confirmed the precondition of, and what it correctly refused to conclude.** It
verified the carry half's setup (six repeated `slot_name` values holding six distinct
`content_data` payloads → `conflicted[slot]=true` → deleted from the carry map) but refused
to conclude damage, naming two gaps: whether the section's image field actually *depends*
on that carry, and whether the conflict is even observable. Both are now closed by reading,
and the answers hold:
- `illustrated-text-block.image_url` is `source: site_assets.illustration`, `required:
  false`, `on_missing: skip_field` — resolver-sourced, so it depends on either resolution or
  the carry, and on apis.uk resolution finds nothing.
- `handleMissingField` calls `carryStored()` **first** (`plan_sections_action.go`), which
  calls `storedFieldValue(slot, field)` → the deleted slot → `false` → `skip_field` omits
  the key. The chain is complete.

**How I proved the collision half instead.** With a test, which is stronger than the live
probe would have been:
`platform/orchestration/actions/plan_sections_section_imagery_binding_test.go`. On the
pre-fix code, two sections with two distinct section-scope illustration rows both resolve
`illustration-ring-grip.jpg`. Watched to fail before the fix was written; its negative
control (one row, two sections → both take the page-wide alias) **passes on the pre-fix
code**, so it pins today's population rather than the one being added.

⚠ **A misstep worth keeping: the first run of that test failed for the WRONG REASON.** My
mock declared four columns while the (then) query selected three, so every row failed
`Scan` and was silently skipped — both tests failed, including the control that should have
passed. A test that fails is not thereby evidence of the defect it was written for: the
control failing was the tell.

## 2026-08-31 (later still) — the fix, and what it deliberately does not do

Committed `cb698ee58`, council `Council-Submitted: 2979c27f-1545-47c5-b28d-f8a700bb1cb0`,
registered as **IMG-075**. Go only, so **inert until the next chassis roll**.

The ordinal is now read. It is translated once, in `ensureAssets`, into a
`sectionRef{Name, Occurrence}` against the plan's own section order, and both render paths
count occurrences of a slot name in their own order to look it up. Not a position integer —
`site_plan_sections.ordering` is 0-based and counts site-level slots while
`page_components.position` is 1-based on 847 of 1,065 pages and neither on 128
`[MEASURED 2026-08-31]`, and the build path filters header/footer names out of the list.

**Both paths, deliberately.** The re-render merges stored ⊕ fresh *resolved last*, so
binding only the build path would have the next re-render overwrite the per-section figures
the build had just got right. That was nearly the shape of this fix, and it would have been
worse than the defect.

**Four stand-down cases**, each keeping today's behaviour: no current plan row; an
out-of-range/malformed ordinal (logged with the `scope_ref`); a failed scan of the plan
order (fails CLOSED — a gap makes every later ordinal name the section *before* the one it
meant); and a page with locked sections the plan does not name (`LoadLockedPageSlots` +
`MergeLockedPageSlots` asked, deliberately NOT merged — the ordinal indexes the plan's list,
so merging would bind figures one section out).

Two estate ratchets fired on the first cut and both were right: the silent-scan-loss
ratchet (`bugs_open/410`) on a `continue`-on-scan-error, now fail-closed; and the
lock-blind-plan-reader coverage test (RFC_033/LOCK-008), which is why the lock question
above was answered rather than skipped.

**Still open after this, and it is the supply half:** nothing yet *writes* a section-scope
illustration row for a guide. The binding is the join; the plan rows and the images are the
next job, and `grip-styles` (7 h2 + 6 h3, zero content images) is the canary the other lanes
nominated. apis.uk/index also stays exposed until it gets `scope='section'` rows — its six
figures live only in stored `content_data`, which the carry cannot protect.

---

## 2026-09-02 — the code shipped while the review was open, and the review found two real defects

Picked the lane up again after two days. Three things had happened, and the order matters.

### 1. The binding is LIVE, and the prescribed way to check that did not work

`[MEASURED 2026-09-02]` chassis `v1.0.1351`, pods up 2026-09-01 21:00, carries round 1's code.

⚠ **CLAUDE.md's `kubectl logs … | grep 'build provenance'` matched NOTHING on this service.**
The only hits were prompt text *describing* the check — an LLM prompt quoted back through the
logs, which is a false positive shaped exactly like a true one. (The text itself says so: *"the
deploy check CLAUDE.md prescribes — matches NOTHING on a backend service"*.)

What worked was the binary probe with **both** controls in one breath:

```bash
POD=agent-chassis-5bd89cf49-t4wdl
for sym in PlanSectionsAction sectionRefForOrdinal planSectionOrder sectionRefForOrdinalNOTREAL; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe && echo "PRESENT $sym" || echo "absent  $sym"
done
# PRESENT PlanSectionsAction   <- must-be-present control
# PRESENT sectionRefForOrdinal
# PRESENT planSectionOrder
# absent  sectionRefForOrdinalNOTREAL   <- must-be-absent control
```

Probe the **capability**, not the commit: a symbol answers "can this binary do the thing", which
is the question, and it has no shelf life the way a startup log line does.

### 2. The council returned REVISE, and both gating objections were real

`bug_historian` (HIGH, gating) and `reuse_agent` (HIGH). Neither was paperwork.

- **The ordinal's identity had THREE separately-written implementations** over three different
  orderings — write-time range check, build-time plan-order walk, rerender-side occurrence count.
  That is *the same shape my own rationale used to reject a position integer*. Reading my
  submission back, the argument refuted the design I then wrote.
- **The drift population was untested and unguarded.** Round 1 stood down only for locked-slot
  insertions. Every other way the plan's order and the live order come apart — a manual reorder, a
  deleted section, an unreconciled replan, a renamed component — went unnoticed, and in that
  population the binding does not fall back, it binds a **real figure to the wrong section**,
  silently, on a page that renders and deploys clean.
- **`NewInstanceCounter` already existed for exactly this question** and, on the rerender path,
  sat **two lines above** my hand-rolled `map[string]int`. I did not look at it. It is not a
  cosmetic duplication either: the counter lower-cases and trims its key, so a raw-key map
  disagrees with it precisely where a plan and a stored row spell one slot differently — the
  apis.uk shape.

### 3. What round 2 changed, and how it is proven

One occurrence rule (`InstanceCounter.NextOccurrence`, now exported, with `Next` reimplemented on
top of it), one ordinal parser (`sectionScopeRefOrdinal`, living beside the write-time range check
that already parsed it), one identity constructor (`newSectionRef`, normalising as the counter
keys), and `sectionOrderAgrees` — bind only when the plan's section list and the list the calling
path iterates describe the same sequence of slots.

**Both guards mutation-proven, run and watched to fail:**

```
# force sectionOrderAgrees -> true
--- FAIL: TestPlanSections_DriftedLiveOrderStandsDownToPageWide
    section 1: image_url = "/assets/images/illustration-shark-grip.jpg",
    want the page-wide "/assets/images/illustration-ring-grip.jpg"
# revert newSectionRef to the raw name
--- FAIL: TestSectionIdentity_MatchesTheEstatesInstanceTokenRule
    section 1: identity name "Illustrated-Text-Block" is not normalised the way the counter keys it
```

The first is the predicted damage reproduced on demand — the wrong grip's photograph under the
wrong heading — which is what turns "a drift risk" into a defect with a failing test attached.

Committed `38178d549`, resubmitted on the same trail correlation `2979c27f`.

⚠ **Misstep of the day, and it is the same one twice:** the package suite failed on
`apply_theme_kit` having no registry entry. Not mine — another session's untracked
`apply_theme_kit_action.go`. Checked before assuming, this time (`git status` on the file that
registers it), rather than "fixing" a symbol that is somebody's plan.

### 4. The first real consumer is apis.uk, and it is not ours to change

`[MEASURED 2026-09-02]` apis.uk `/index.html` still serves six `illustrated-text-block` sections
with six distinct illustrations, all six URLs living **only** in `page_components.content_data`,
with **zero** `scope='section'` imagery rows. Its plan order (`hero`, `generic-text-block` ×6,
`site-footer`) and its live slots (`hero`, `generic-text-block` ×6) **agree**, so the binding would
engage there today if the rows existed. Filed as a CONTRIB into `apis_uk_bees_homepage` with the
SQL, the pairing read off their own live rows, and the verification command — theirs to run.

That CONTRIB also carries the correction to IMG-074's "carryStored preserves the six values",
which is false for the reason in §3 of the 08-31 entry.

### 5. Is the new guard so cautious that nothing binds? Counted, not assumed

A stand-down rule that stands everything down is a dead feature which still passes every test —
the "your own action can silence your own detector" shape. So I counted the eligible population
before claiming the guard was cheap. `[MEASURED 2026-09-02]`

```
pages carrying ANY section-scope imagery row : 31
  would bind (plan order == live order)      : 21
  would stand down                           : 10
     of which: no built rows at all          :  4   (unbuilt / internal pool sites — moot)
               plan and page differ in LENGTH:  5   (dartsonline index: plan 6, live 4 —
                                                    a re-plan the page has not been built from)
               same length, different order  :  1   (loancalculator index)
```

Two thirds eligible, and every refusal individually explicable. **This could have come out 0/31**,
which would have meant the guard had quietly killed the feature — that is why it is worth the
query rather than a sentence saying the guard is conservative. The full SQL (a plan-sequence vs
live-sequence comparison, site-level slots filtered) is in the scratch of this session; it is
worth re-running rather than quoting, because the sequences move on every re-plan and rebuild.

### 6. My own blast-radius caution named the wrong population — a row census is not a reachability census

I had written, in the register, in the council submission and in the commit message, that `icon`
was "the case to watch" because it reaches 10 section-scope rows on one page against illustration's
maximum of one. That was inferred from a count of ROWS. The branch does not read rows; it reads a
component field whose `source` names a kind. So I asked which components can actually walk through
it `[MEASURED 2026-09-02]`:

```sql
SELECT v->>'source', count(*) AS fields, count(DISTINCT cc.id) AS components,
       string_agg(DISTINCT cc.function, ', ')
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') AS f(k,v)
 WHERE cc.is_active AND v->>'source' IN ('site_assets.illustration','site_assets.icon','site_assets.infographic')
 GROUP BY 1;
-- site_assets.illustration | 2 | 2 | brief-explanation, illustrated-text-block
-- (no rows for icon; no rows for infographic)
```

**Zero components source `site_assets.icon` or `site_assets.infographic`.** Icon fields are reached
by their literal asset key (`site_assets.icon_speed`), and the per-section map holds **kind keys
only** — so a literal-key lookup falls straight through to the unchanged page-wide map. The icon
population is not merely unlikely to be affected; it is unreachable.

And then the second half, which is the one that matters:

```sql
SELECT s.domain, p.name, count(*) FROM page_components pc
  JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE cc.function IN ('brief-explanation','illustrated-text-block') GROUP BY 1,2 ORDER BY 3 DESC;
-- apis.uk index 6 · idea.uk index 1 · idea.uk tools 1 · lendzy 1 · oufe 1 · remortgagecalculator 1
-- · robot-hands 1 · vonc 1 · webdesign.uk 1
```

**One page in the estate has more than one instance of a component that can reach this binding.**
On a single-instance page, per-section and page-wide resolution are identical by construction. So
the live blast radius of IMG-075 is exactly `apis.uk/index`, and only once section-scope rows are
seeded there.

⚠ **The lesson, and it is worth more than the correction:** a row census answers *how much data
exists*, never *which code path can read it*. I cautioned about the wrong population for two days —
in a register entry, a council submission and a commit message — on the strength of a number that
was accurate and irrelevant. The check that would have caught it is the one above and it is a
single query: **before naming a population as the risk, ask what would have to name it in a schema
for the branch to fire.**

### 7. The round-2 council run was KILLED by a chassis roll, and it reads as a slow one

`[MEASURED 2026-09-02]` The run froze at `review_prior_art`, `updated_at` **12:28:23**, `error`
NULL. The chassis pods were created at **12:28:03 and 12:28:24** (`agent-chassis-96c48f448-*`,
v1.0.1352). Round 1 of the same submission had completed in nine minutes, so 47 minutes on one
step was the tell that this was not queue latency.

**The control that makes it conclusive came from a peer lane** (`editorial_design_uplift`), whose
own run died in the same second and who checked the half I had not: **every run submitted AFTER
the roll completed normally** (3m48s, 5m29s, 5m48s, 19m55s). So the gate is healthy and it is
specifically the runs in flight across 12:28 that died — three submitters, `error` NULL on all
three. This is the existing LANDMINE ("a council run killed by a chassis roll looks exactly like a
slow one"), confirmed rather than newly found; no new entry, the ledger already says it.

Resubmitted on the same trail (`RESUBMIT_CORR=2979c27f`, new envelope `931921c6`).

### 8. My round-2 commit missed that build by two and a half minutes — probed, not assumed

Round 2 committed **12:25:41 UTC** (`38178d549`); pods up **12:28:03**. Probed at the binary,
BOTH replicas, controls in both directions:

```
PRESENT PlanSectionsAction        <- must-be-present control
PRESENT sectionRefForOrdinal      <- round 1
absent  sectionOrderAgrees        <- round 2, the drift guard
absent  NextOccurrence            <- round 2, the shared occurrence rule
absent  sectionOrderAgreesNOTREAL <- must-be-absent control
```

**So the fleet runs half of this change: the binding without its guards.** Worth being precise
about the exposure rather than alarmed by it — it is currently **zero pages**, because only two
components can reach the binding (§6) and the single page with more than one instance of either
has no section-scope rows. It closes on the next roll.

⚠ **The lesson is the timing, not the outcome:** a commit made minutes before a roll is not in
that roll, and "it went out with the build" is exactly the inference that feels safe and is
unfalsifiable without the probe. Two and a half minutes was enough.

### 9. Round 2 verdict: REVISE again, and the gating objection was me repeating a mistake I had just confessed

`[2026-09-02 13:38Z]` **8 approve / 4 object, gated by `editquality` (HIGH).** The gating
objection is not about the design at all:

> "The rationale explicitly states a NINTH file … needs 'the byte-identical two-token mechanical
> change' … but it is not included in the edit list, only named in prose 'to stay inside the
> 8-edit budget.' This is the exact failure the rationale itself confesses to for round 1."

Round 1 was objected to for naming call-site files in prose instead of listing them. I answered
that objection **about the file it named**, and then did the identical thing with the ninth file.
The class, not the instance — the same error `MEMORY.md` records as "an objection naming one file
is naming a CATEGORY", which I have apparently learned to quote and not to apply. An 8-edit budget
is not a reason to submit a non-compiling diff; the fix was to drop something else, and what came
out is the round-1 test file that is **unchanged this round** — listing it showed a reviewer a diff
that does not exist.

**The three substantive objections, and what each one actually found:**

- **`bug_historian` (medium) — "count ROWS, not names" (016b §9).** Correct, and the estate holds
  the worked case `[MEASURED 2026-09-02]`: loancalculator.co.uk/index has plan ordering 1 =
  `tool-loan-repayment` while live position 2 is `slot_name='tool-3'`, component function
  `tool-loan-repayment`. Plan name and live slot differ; composition identical. Two consequences,
  and only one is the guard's: the guard stands that page down (a false negative that costs the
  feature, never correctness) — but **per-section binding could not have worked there anyway**,
  because the map key IS the name and the two sides spell it differently. The guard makes an
  existing miss explicit. The improvement it points at — key by the resolved component FUNCTION,
  which `planSection` already holds — is named in round 3 and deliberately NOT rushed: the plan
  side would need `component_name`→function resolution, i.e. `componentNameResolver`'s five arms,
  and a hasty by-name resolution in this exact file is `bugs_closed/044`.
- **`reuse_agent` (low) — `check_section_source_drift` already compares plan against live.** Read
  it: it compares the authoritative plan against `pages.sections` (the CACHE), and its header says
  both sides are viewed through `MergeLockedPageSlots` so that **locked rows are NOT drift**. Mine
  compares the plan against the list the render path iterates, and must treat a locked row the
  plan does not name as a reason to stand DOWN. Same shape, different second operand, **inverted
  polarity** — merging is what makes that check correct and would make this guard agree exactly
  when it must not.
- **`debug_historian` (medium) — there is a LANDMINE on the exact symbols this rewrites.** True,
  and **this lane filed it** on 2026-08-31 as part of this work. I did not cite my own entry in
  `grounded_in`, which is why the seat had to point at it. Now quoted.

Also `debug_historian` (low): my "1,082 active pages" never named the status predicate. Enumerated
— `pages.status` has exactly **two** spellings today, `active` 1105 / `archived` 91, so the
documented four-spelling trap does not apply here. And the figures **moved during the review**:
1,082→1,105 active, 31→32 pages carrying section imagery, in one afternoon.

Round 3 submitted on the same trail (envelope `d2dab332`).

### 10. Round 3: APPROVED — and what three rounds actually bought

`[2026-09-02 ~14:20Z]` **APPROVED**, 12 seats, one advisory objection none high, plus three
non-gating from `prior_art_librarian`. `Council-Reviewed: 2979c27f-1545-47c5-b28d-f8a700bb1cb0`.

The tally is worth writing down because I would have called round 1 finished work:

| round | gating objection | what it actually found |
|---|---|---|
| 1 | `bug_historian` HIGH | the ordinal's identity had **three** implementations over three orderings — the exact shape my own rationale used to reject a position integer — and the drift population mis-bound silently rather than falling back |
| 1 | `reuse_agent` HIGH | I hand-rolled an occurrence counter **two lines below** `NewInstanceCounter()` in the same loop |
| 2 | `editquality` HIGH | I answered round 1's file-list objection about the file it NAMED, then repeated it with a ninth file — the "an objection names a CATEGORY" lesson, quoted and not applied |
| 3 | — | approved; two advisories that named real gaps |

**The two advisories are discharged (`4084481d7`), not filed away:**

- `guardian`: the Next/NextOccurrence arithmetic-identity claim was "plausible from the sketch"
  and untested while a third caller outside this change (`assemble_from_library.go:295`) depends
  on it. Now `TestInstanceCounter_NextIsExactlyTokenOfNextOccurrence`, **mutation-proven** — add
  `+1` to the occurrence inside `Next` and it fails at position 0 naming all three callers.
- `debug_historian`: the pod probe recorded in the RUNBOOK listed only **round 1's** symbols.
  That is worse than incomplete — `sectionRefForOrdinal` and `planSectionOrder` are PRESENT on
  both `v1.0.1351` and `v1.0.1352` while round 2/3's guards are absent from both, so the old
  list returns **all-present on a binary carrying half the change** and reads as a clean deploy.

`prior_art_librarian`'s two mediums ask for verification that seat cannot perform (the call-site
enumeration; the icon reachability census). Both had already been re-run independently after
submission — the census dialect-independently over all 432 active components, same answer — and
that is in the register entry rather than only in the submission.

**Still not shipped.** The approval is of code that is committed and, for rounds 2–3, not yet in
any binary. Next roll carries it; the probe list above is what to run then.

### 11. The population this lane's PLAN never measured — 432 guide pages, 0 composed the way the plan proposes

Prompted by a peer lane's rollback (below), applied to my own advice, and it bears on the PLAN's
core proposal rather than only on the CONTRIB I corrected. `[MEASURED 2026-09-02]`

```
active guide/blog pages fleet-wide              432
  ...with a hero component                      330   (76%)
  ...with MORE THAN ONE illustrated section       0
```

**Zero.** The fleet pattern is hero-section-for-imagery, article-body-for-prose. The PLAN's
route — one `illustrated-text-block` per h3 — has no instance anywhere in the estate, and
`grip-styles` would be the first. That is not an argument against it: 432 pages of banner-plus-
wall-of-text is the complaint, not the standard. But it changes the status of the recommendation
from "adopt the pattern" to "run the experiment", and the PLAN does not say so anywhere.

**Why I only measured it today.** A peer lane (`editorial_design_uplift`) applied a migration
giving `article-body` an image field, having measured the component, the assets, all three
resolver arms, the llm-dispatch mechanism, the alias map and the precedents — and then found that
**292 of 301 pages carrying that component already show the same image through their hero**, so
it would have rendered twice on 97% of the population. They rolled it back within 25 minutes,
prompted by the vonc.com case I had sent them from bug 114, and were careful to label "0 of 301
instances acquired the field" as luck rather than a control they had built.

**Eleven council seats approved that migration and none could have caught it.** That is not a gap
in the gate: a submission can be entirely accurate about the change and entirely silent about the
estate around it, and the estate is the one input a reviewer cannot fetch. Their sentence, which
I have adopted: **a remedy is fitted to a POPULATION, not to a defect.**

⚠ **So the check that belongs in this lane's next submission, and in the PLAN's Phase 1:** before
changing how a shared component behaves — or recommending a composition — ask what the OTHER
instances already do, and state it. I measured this change's blast radius carefully (reachable
components, instances per page, plan-vs-live agreement) and never asked what a healthy guide page
looks like, which is the same question one level out.

**Their reframe is worth carrying too, because it explains something I had filed as an oddity.**
The six boxingonline blog pages have no hero component AND no page-scope plan hero — which is
exactly the case `ContentHeroKey` exists to generate a per-article image FOR. So those images are
the system correctly compensating for a composition gap and then having nowhere to put the
result. The real question is why the planner composed those pages without a hero when 330 of 432
peers have one — page composition, upstream of any component change. Logged here because this
lane's Phase 2 would otherwise inherit the same wrong altitude.

### 12. First driver seeded (and NOT exercised), and the measurement that inverts this lane's phasing

**The offer was taken.** `[MEASURED 2026-09-02]` `apis.uk/index` now carries six `scope='section'`
illustration rows — `index:1`–`index:6`, the exact keys and ordinals from the CONTRIB,
`source='manual'`, `locked_by='apis_uk_bees_homepage'`, created **16:47:03Z** — about seven hours
after the guards shipped.

⚠ **Armed is not exercised, and I nearly wrote it up as if it were.** That page's
`page_components` still read `updated_at = 2026-08-24`. **Nothing has re-resolved, so the branch
has never actually run.** The mechanism has a driver and no evidence. What will produce evidence:
a re-resolving render (`reason=section_data_resolved` / `image_landed`), or — decisively — that
page's next `content_rewrite`, which is the exact event this was written to survive, on the exact
page whose six values it was written to protect. An assemble-only re-render in the meantime
changes nothing, because the images are already in `content_data`, and must not be read as
failure.

**And the measurement that matters more than any of the above** `[MEASURED 2026-09-02]`, prompted
by `dartsonline_traffic` asking what fraction of pages can actually host one of these:

```
active content pages (blog/guide/article) fleet-wide   442
  ...carrying ANY illustration-capable section           9
  ...carrying MORE THAN ONE (can host a figure SET)      1
```

**The population that wants per-section imagery and the population that can accept it differ by
two orders of magnitude, and no amount of seeding closes that: there is nowhere to put the rows.**
dartsonline is the worked case — all 22 of its content pages are exactly `hero` + `article-body` +
`call-to-action`, and neither illustration-capable component appears anywhere on the site. So
"re-plan grip-styles" was never the task; "re-plan the content estate" is.

⚠ **This inverts this lane's own phasing.** The PLAN treats composition as Phase-1 groundwork and
the durable binding as the hard part. The binding took three council rounds and is done; the
composition it depends on exists on **one** page in the fleet. **The constraint was never
component capability.** `editorial_design_uplift` reached the same place from the opposite
direction the same afternoon — asking why the planner composed six blog pages with no hero when
330 of 432 peers have one — and neither of us was looking for it. Two lanes, two starting points,
one answer: **what the planner composes.**

### 13. CORRECTING §12 within the hour: the component IS being selected — the gap is the pattern, not the capability

§12 reported 442 / 9 / 1 and concluded the bottleneck is page composition. The numbers stand; the
natural reading of them — *nothing selects the illustrated component* — is **false**, and the
`dartsonline_traffic` lane's "whatever gates composition onto `Illustrated Text Block` is upstream"
sent me to check rather than to assume.

`[MEASURED 2026-09-02]` migration 644 taught the three planner menus the word for an image on
2026-08-26 11:16Z. Of the 9 pages carrying an illustration-capable section, **6 were composed by a
planner AFTER that**:

```
webdesign.uk index          2026-08-26 16:52   (5h after 644)
idea.uk tools               2026-08-28 22:38
lendzy.co.uk index          2026-09-01 05:42
oufe.com index              2026-09-01 07:24
remortgagecalculator index  2026-09-02 12:45
robot-hands.com index       2026-09-02 15:26   ← today
apis.uk index ×6            2026-08-24         ← BEFORE 644, and hand-built
```

**644 worked. Adoption is live and growing — six sites in seven days, two of them today.** The
"nobody drives the machinery we build" story does not apply here and I nearly told two lanes it
did.

**What the gap actually is, and it is much cheaper than the coarse reading implies:**

| | |
|---|---|
| pages with an illustrated section | 9 |
| ...`page_type='landing'` with **exactly one** (an accent) | 8 |
| ...`content` with one | 1 |
| ...with **several** (the shape per-section binding needs) | 1 — apis.uk, hand-built |
| ...`blog-post` or `guide` — **the owner's actual subject** | **0** |

So the planner has learned to reach for an illustrated block **once, on a landing page**, and has
never composed an **article** out of them. That is a planning-behaviour question with a live and
growing precedent to build on — not a missing capability, and not the estate-wide recomposition
"1 of 442" suggests. ⚠ **Two readings of one census, an hour apart, and the difference is entirely
in whether I asked WHEN the 9 were created.** A count of a population says nothing about whether
that population is growing; the timestamps were in the same table the whole time.
