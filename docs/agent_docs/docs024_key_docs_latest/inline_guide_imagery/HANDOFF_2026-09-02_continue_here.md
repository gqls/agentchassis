# HANDOFF — inline_guide_imagery. START HERE. Written 2026-09-02, rewritten 2026-09-03 12:30Z.

**Status in one line:** the per-section imagery binding (**IMG-075**) is built, council-APPROVED at
round 3, and **verified live on `v1.0.1358`** — and for the first time it has two real consumers,
one of them the owner's own page, **both mid-transition and neither yet exercised.**

**Lane docs:** `docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/` — `PLAN_2026-08-14…`,
`NOTES_…` (technical log, newest at the bottom), `RUNBOOK_…` (the queries, with their traps),
`README_where_we_are.md` (owner's plain-prose log), this file.
**Register:** `docs026_concept_register/register/imagery.md` → **IMG-075** (also IMG-074, corrected).

---

## 1. What this lane exists for

The owner asked (2026-08-13, restated 2026-08-31 naming ring/razor/shark grip on
`dartsonline.com/blog/grip-styles.html`) that guide articles carry explanatory imagery **inside**
the body. The plan reframed it as a **durability** problem — in-body `<figure>` markup lives in
`article-body`'s single LLM-owned `content` field and dies on the next regeneration — and IMG-075
is the answer: a figure planned in `site_plan_imagery` re-resolves on every build and re-render
instead of living in the prose.

---

## 2. THE LIVE STATE, and the one thing to do first

⚠ **Re-probe before trusting anything below.** The fleet has rolled four times in two days
(`1351 → 1352 → 1354 → 1355 → 1358`), and every artefact claim expires with the next roll.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers \
        -o custom-columns=NAME:.metadata.name | head -1)
for sym in PlanSectionsAction sectionRefForOrdinal sectionOrderAgrees sectionScopeRefOrdinal \
           newSectionRef sectionOrderAgreesNOTREAL; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe && echo "PRESENT $sym" || echo "absent  $sym"
done
```

**Read it like this:** `PlanSectionsAction` **PRESENT** and `sectionOrderAgreesNOTREAL` **ABSENT**
means the instrument works — only then are the middle four meaningful. **Run it on BOTH replicas.**
`[MEASURED 2026-09-03 12:2xZ]` on `v1.0.1358`: all four present on both, controls clean.

⚠ **Do NOT suppress stderr.** On 2026-09-02 the identical probe returned "absent" for all six
*including the must-be-present control*, and I nearly recorded a regression. `kubectl` was
returning `Unauthorized` (token expires ~3 days; the owner refreshes it) and `2>/dev/null` had
turned a failed exec into the word "absent". **A failing command and a missing symbol are the same
output; only the control separates them** — the visible `command terminated with exit code 1` on
the negative control is what proves the exec ran.

⚠ **`kubectl logs … | grep 'build provenance'` does not work on this service** — the phrase appears
in LLM prompt text the chassis logs, so it returns a hit shaped exactly like a stamp. Already a
LANDMINE; don't re-derive it.

⚠ **A second cause of "absent with clean controls":** Go's linker strips uncalled functions, so a
genuinely INERT symbol probes absent on a build that contains the commit (LANDMINE, 2026-09-02).
Not a risk for these four — `sectionOrderAgrees` is reached by every section render — but for an
inert symbol verify by ANCESTRY instead (pod's `git_commit` from `service_binary_capabilities`,
then `git merge-base --is-ancestor`).

---

## 3. What shipped

**IMG-075 — a `scope='section'` `site_plan_imagery` row now binds to the ONE section its
`scope_ref` ordinal names.** Before it, every section on a page declaring `site_assets.illustration`
resolved the *same* URL (kind first-wins); the ordinal was filtered on and thrown away.

| commit | what |
|---|---|
| `cb698ee58` | the binding + register entry IMG-075 |
| `844eb3023` | **fix HEAD** — a third `resolve()` caller left off the pathspec |
| `38178d549` | round 2: one occurrence rule, one ordinal parser, the drift guard |
| `4084481d7` | round 3 advisories discharged (mutation-proven identity test; probe list fixed) |

**`Council-Reviewed: 2979c27f-1545-47c5-b28d-f8a700bb1cb0` — APPROVED round 3**, 12 seats, 1
advisory none high.

**Design, in one paragraph.** The ordinal is translated ONCE, in `ensureAssets`, into a
`sectionRef{Name, Occurrence}` against the plan's own section order — **never a position integer**
(`site_plan_sections.ordering` is 0-based counting site-level slots; `page_components.position` is
1-based on 847 of 1,065 pages and neither on 128). Both render paths count occurrences with the
shared `InstanceCounter`. `sectionOrderAgrees` **stands the binding down** — rather than
mis-binding — when the plan's order and the live order disagree.

---

## 4. ⚠ TWO LIVE CONSUMERS, BOTH MID-TRANSITION — this is the new news

`[MEASURED 2026-09-03 12:2xZ]` section-scope illustration/infographic rows fleet-wide have gone
**5 → 20** in two days. The interesting ones are both from today:

**(a) `dartsonline.com/grip-styles` — the owner's own case, and it is half-built.**
- **Plan RE-COMPOSED:** 11 sections — `hero`, `Generic Text Block`, **5× `Illustrated Text Block`
  (ordinals 2–6)**, 3× `Generic Text Block`, `call-to-action`.
- **5 imagery rows seeded** `source='manual'` at exactly `grip-styles:2 … :6` —
  `illustration_ring_grip`, `_razor_grip`, `_shark_grip`, `_smooth_barrel`, `_combination_grip`.
  **2 of 5 assets active**; three still generating.
- **The page has NOT been rebuilt** — `page_components` still the old three (`hero`,
  `article-body`, `call-to-action`, updated 2026-09-01).

⚠ **So the binding correctly STANDS DOWN right now**: plan says 11 sections, live says 3, they
disagree. That is the designed behaviour and **it is not a failure** — the ordinals name a
composition the page does not have yet. **Sequence: recompose → seed → REBUILD → verify → then
re-render freely.** They are between steps 2 and 3.

**(b) `gamedesign.uk/index` — the first PLANNER-WRITTEN section imagery, i.e. 718 working.**
4 rows `source='llm'`, created 2026-09-03 10:40Z, hours after migration **718** flipped the
planner's "use sparingly — most plans will have zero" instruction. 3 of 4 assets active.
⚠ **But note the SHAPE, which is not the one this binding optimises for:** three of the four sit at
the **same ordinal** (`index:2`), because 718's decomposition rule tells the planner to emit one
entry per image for a card grid. The per-section map is **kind-first-wins within an ordinal**, so
that section gets ONE of the three through `site_assets.illustration`; the other two are reachable
only by **literal asset key** (`site_assets.illustration_article_card_balance`), which the resolver
does support but which needs per-key schema fields on the component. **Worth watching as the first
real-world shape 718 produces.**

**(c) `apis.uk/index` — the original driver. Armed since 16:47Z 09-02, and its one attempted test
was REFUSED AT SAVE.** Six rows seeded; the lane filed a `section_data_resolved` rerender in the
same transaction and it **failed** — that lane supplied the reason I could not see (`result` is
`{}`): the **save guard** refused it, *"re-confirmed too little of what is stored (prune_floor…)"*.
**Not the resolver, and not the locks per se.** All seven rows are also `lock_type='permanent'`.
⚠ **Do NOT bin apis.uk with the `bugs_open/425` evidence** — it never reached the write, whereas
the dartsonline hero case did. Different failure stages; folding them together inflates a
hypothesis three lanes are building.

---

## 5. ⚠ THE OPEN RISK TO THIS MECHANISM, and it is not in the binding code

A genuine sections-path run on `dartsonline.com/tool-brand-comparator` (`section_data_resolved`,
attributed by `source_item_id`, 2026-09-03 00:40:40Z) **did not write a newly-declared
`site_assets.hero` field**, on a page that HAS a resolvable hero (arm 2, `content_hero_…`, active).
**I tried to dissolve it with the obvious mundane explanation and it survived** (`bugs_open/114`,
this date).

**IMG-075 resolves `site_assets.illustration` on that same path.** So a seeded page may fail to
bind **for a reason unrelated to this lane's code**, and the natural reading on the day would be
"IMG-075 does not work". `[UNVERIFIED]` — one instance on `site_assets.*`, one on `query.*`
(`bugs_open/425` §2, four reproductions). **The components lane owns the discriminating test
(batch 690).** Its result decides whether this lane has a problem at all.

---

## 6. Where the ask really stands: THREE layers, and this lane owns only the top

1. **Can a figure survive regeneration?** ✅ Done, reviewed, live.
2. **Does anything compose an article out of illustrated sections?** Until today, no — 1 page in
   442, hand-made. **dartsonline just did it by hand for grip-styles**, and 718 has started the
   planner doing it for landing pages. `editorial_design_uplift` owns this.
3. **Are articles even IN THE PLAN?** ⚠ **Mostly not — and this is the floor.**
   `[MEASURED 2026-09-03]` on the 33 sites with a current plan: **tool 83%, blog-post 85%,
   guide 74%** have NO `site_plan_sections` row, against **landing 2%**, content 15%. The split is
   by page TYPE, not site health. **No plan row → `planSectionOrder` returns nil → binding
   disabled.** This lane's own first stated degrade case, and nobody owns fixing it.

**Mechanism, read first-hand:** `create_blog_posts_action.go:212` — the article layout triple is a
**fallback** (`post.Sections` may be supplied), and the action writes `pages.sections` (**tier 3,
the cache**) and never `site_plan_sections` (**tier 1, the authority**).

**Who can write the authority** `[MEASURED 2026-09-03]` — three populations, because a Go-only
grep sees one and I published that mistake once:

| population | count | writes? |
|---|---|---|
| Go | **2** (`write_site_plan_action.go:668`, `apply_gap_plan_action.go:1067`) | yes; **neither on the article path** |
| live `agent_definitions` config SQL | 2 rows | **reads only** (negative control that could have broken it) |
| **operator SQL in the repo** | **15 files** | **yes — a real third path** |

⚠ **The third is the trap.** Backfilling by hand fixes the pages that exist and nothing about the
route — dartsonline's nine July plan rows all share ONE timestamp (`2026-07-29 13:28:03.58521Z`),
and the 14 articles created there since have none. **It is a cheap unblocker for a canary, not
phase one**, and "just pass richer `post.Sections`" is not a fix either — it fills the cache and
leaves the authority empty, so the ordinal still has nothing to name.

---

## 7. What I would do next, in order

1. **Re-probe the current build** (§2). Everything else here assumes it.
2. **Watch `dartsonline.com/grip-styles` through its REBUILD.** It is the owner's own case, it is
   two steps from finished, and it will be **IMG-075's first real test**. Read the served bytes,
   not the rows; expect the binding to stand down until the page is rebuilt to match its plan.
3. **Watch `gamedesign.uk/index`** as 718's first planner-written set, and specifically whether the
   three-rows-at-one-ordinal shape needs per-key component fields.
4. **Do not build the Phase-4 detector** (`check_unrendered_section_imagery`) yet. The PLAN's
   "discovery has no driver" blocker is **stale** (corrected there — discovery runs daily now), and
   the RUNBOOK's hand query already does the job. `bugfix_114` has offered a section-scope arm on
   their `check_unrendered_page_imagery` (IMG-077), which is cheaper than a new mechanism.
5. **Leave phase 3 (article planning) where it is** — nobody owns it, it is not this lane's, and it
   should not be closed by a backfill.

---

## 8. Traps this lane paid for — read before trusting a number

- **A count of a population says nothing about whether it is GROWING** — I read "9 of 442" as
  "nothing selects it"; the refutation was `created_at`, in the table I had already queried.
- **A `LIMIT` counted section ROWS while my claim counted PAGES** — one page contributed 6 of 12.
- **A reference census matched the component NAME, not the filename** (`LIKE '%hero-about%'` finds
  `class="hero-about"`). Anchor on the extension; run a control. **LANDMINE.**
- **`updated_at` moved ≠ a re-render happened ≠ the resolver was asked** — three events, one word.
- **Filed ≠ ran** — an item can exist, be quoted as evidence, and have FAILED with `result = {}`.
- **A Go-only grep is not a fleet-wide census** — config SQL and operator SQL are two more writer
  populations.
- **I quoted a Go COMMENT as live config** — it said two re-render reasons; `agent_definitions`
  says five. **Cite the row.**
- **Ask for PROVENANCE, not just correctness** — *"have you read it, or is it relayed?"* caught a
  29-line-drifted citation, an overstatement and an unchecked inference inside a claim that was
  true. Keep your own `[UNVERIFIED BY ME]` until **you** have read it.
- **`git stash` is forbidden; commit by pathspec — and build the pathspec from `git status`, not
  memory.** I broke HEAD for eleven minutes naming two of three callers.
