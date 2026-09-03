# HANDOFF — inline_guide_imagery. START HERE. Written 2026-09-02, rewritten 2026-09-03 15:15Z.

**Status in one line:** the per-section imagery binding (**IMG-075**) is **PROVED END TO END on the
owner's own page**, including the decisive save-path test — and the same page shows that the WORDS
beside those figures are wrong, for a reason that is not this lane's code and has another owner.

**Lane docs:** `docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/` —
`PLAN_2026-08-14…`, `NOTES_…` (technical log, newest at the bottom; today is **§17**),
`RUNBOOK_…` (the queries, with their traps), `README_where_we_are.md` (owner's plain-prose log),
`SUMMARY_2026-09-03_inline_guide_imagery.md` (**first summary — the milestone read-out**), this file.
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

⚠ **Re-probe before trusting anything below.** Every artefact claim expires with the next roll.
`[MEASURED 2026-09-03 12:2xZ and re-run at session start]` on **`v1.0.1358`**: all four symbols
PRESENT on **both** replicas, controls clean.

```bash
PODS=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=NAME:.metadata.name)
for POD in $PODS; do echo "== $POD =="; for sym in PlanSectionsAction sectionRefForOrdinal \
    sectionOrderAgrees sectionScopeRefOrdinal newSectionRef sectionOrderAgreesNOTREAL; do
  timeout 40 kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe \
    && echo "PRESENT $sym" || echo "absent  $sym"; done; done
```

**Read it like this:** `PlanSectionsAction` **PRESENT** and `sectionOrderAgreesNOTREAL` **ABSENT**
means the instrument works — only then are the middle four meaningful. **Run it on BOTH replicas**,
and use the per-exec `timeout`: without it the loop over two pods exceeds a 2-minute tool limit and
you get a partial answer that looks like a partial deploy.

⚠ **Do NOT suppress stderr.** On 2026-09-02 the identical probe returned "absent" for all six
*including the must-be-present control*, and I nearly recorded a regression. `kubectl` was returning
`Unauthorized` (token expires ~3 days; the owner refreshes it) and `2>/dev/null` had turned a failed
exec into the word "absent". **A failing command and a missing symbol are the same output; only the
control separates them.**

⚠ **`kubectl logs … | grep 'build provenance'` does not work on this service** — the phrase appears
in LLM prompt text the chassis logs. Already a LANDMINE; don't re-derive it.

⚠ **A second cause of "absent with clean controls":** Go's linker strips uncalled functions, so a
genuinely INERT symbol probes absent on a build that contains the commit (LANDMINE, 2026-09-02).
Not a risk for these four. For an inert symbol verify by ANCESTRY instead (pod's `git_commit` from
`service_binary_capabilities`, then `git merge-base --is-ancestor`).

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

## 4. ✅ THE MECHANISM IS PROVED — this is the new news, and it happened on the owner's page

`dartsonline.com/blog/grip-styles.html`, 2026-09-03, on `v1.0.1358`. Full account: **NOTES §17**.

| time (UTC) | what | item |
|---|---|---|
| 11:39 | darts lane recomposed the plan to 11 sections + seeded 5 figures | `SEED_2026-09-03` |
| 12:41–12:47 | the five illustrations generated, went `active` | 5× `needs_imagery` |
| 12:47→13:02 | **rebuild through the writer** | `d5edd37b` `needs_content_page` |
| 14:00→14:11 | **a SECOND full regeneration**, fired automatically by the last asset landing | `8bd71ef8` `needs_page` `reason=image_landed` |

**Run 1 — the binding engaged, visible before the page deployed.** Writer `837bd4ea`,
`process_sections_loop_item_N.resolved_data`: items 2–6 resolved
`illustration-ring-grip` / `-razor-grip` / `-shark-grip` / `-smooth-barrel` / `-combination-grip`
— five ordinals, five distinct URLs, in plan order. ⚠ **The pre-IMG-075 result AND a stand-down
both look like five IDENTICAL URLs**, so the failure shape is a run of identical URLs, not an error.
The query is in the RUNBOOK; it grades the binding minutes before the deploy.

**Run 2 is the decisive test and it PASSED.** The `image_landed` item routed to `page-build-handler`,
which spawned a second `page-content-writer` (`74d6b7e4`) and rewrote every heading and paragraph.
Measured on the two runs' own `section_output_2`: **prose differs, `illustration-ring-grip.jpg` is
unchanged.** A full body rewrite happened and the figures were untouched — the durability property
this lane exists for, observed rather than argued.

**At the served bytes, 14:11:46Z:** 11 sections, five `<figure>` blocks, five distinct files, each
**200 at 1071×800**, invented sibling → **404**. All five opened and visually correct (ring grooves,
razor cuts, raked shark cuts, a smooth polished barrel, two distinct zones); no feathered flights,
no screw threads — the darts lane's guide-level anatomy clauses held.

⚠ **STILL UNPROVEN: the RE-RENDER path on a multi-figure page.** Both grip-styles runs were the
build/save path. `rerender_page_sections` takes its live section list from stored `page_components`
slots rather than `pages.sections`, so it feeds `sectionOrderAgrees` a **different list** and is a
genuinely separate arm. The four `page-rerender` runs on that site at 13:55–13:58 were **other
pages** (guides-index, index, tool-brand-comparator, tool-setup-builder).

⚠ **apis.uk/index is still armed and still unexercised** — six rows since 2026-09-02 16:47Z, its
`page_components` had not re-resolved. **grip-styles proving the mechanism says nothing about that
page**; the register bullet is fenced accordingly.

---

## 5. ⚠ THE FIGURES ARE RIGHT AND THE WORDS BESIDE THEM ARE WRONG — and it is not our code

**This is the half of the owner's ask that is NOT delivered, on the page he asked about.**

| section | figure bound (correct) | run 1 heading | run 1 `image_alt` |
|---|---|---|---|
| 2 | ring | "The ring grip: a light touch with a clear edge" | ring bands |
| 3 | **razor** | "Ring grip gives you texture without taking over the release" | ring grooves |
| 4 | **shark** | "What a ring grip actually does to your release" | ring-cut bands |
| 5 | **smooth** | "The ring grip: bands that stop the dart sliding forward" | ring-style knurling |
| 6 | **combination** | "The ring grip: bands of shallow cuts" | ring, two bands |

Five sections written about the ring grip under five different **correct** photographs. Run 2
replaced them with five near-identical *"what your fingers feel"* headings, none naming its own
grip, alt text still describing knurling on the smooth barrel.

**Cause, measured against the live config with controls in one predicate.**
`plan_sections_action.go`'s `Subject` doc comment says *"Rides to the writer as
current_section.subject; the v5 prompt renders it only when non-empty."* **First clause TRUE** (all
nine slots carry distinct subjects on both runs). **Second clause FALSE at HEAD-of-live:** the active
non-snapshot `page-content-writer` references **13** distinct `current_section.*` paths, `subject` is
not one, and the string `subject` appears nowhere in that config in any casing — the single step
that references `resolved_data` never mentions it. So the writer is handed the resolved **URL** and
never told what the section is about, nor what is in the picture.

**This is `bugs_open/443` Stage B, which that lane predicted in writing** (§8: *"the writer prompt is
v4; seed 641 (v5, renders the subject) is owner-read gated and NOT applied"*; §9: *"the subject
reaches the writer's DATA and not yet its PROMPT. Stage B is exactly 641 and nothing else."*).
**641's applier is the `framework_prompts_positive_voice` lane, per the owner. DO NOT work it.**
CONTRIB filed into the 443 file 2026-09-03 with three things they did not have:

1. **The damage class is CONTRADICTION, not repetition.** Their censused damage is verbatim-repeated
   `h2`s — dull, misleads nobody. Here the framework's half worked, so identical specification became
   **false captioning of a correct artefact**, `image_alt` included, which is the accessibility
   surface.
2. **The page degraded UNATTENDED.** Run 1's grip-naming headings came from the operator's
   `suggestion` in the `needs_content_page` spec, not from the plan (`[MEASURED]` run 1's handler
   input contains *"five illustrated blocks"*; run 2's whole spec is
   `{"reason":"image_landed",…}`). A routine asset landing undid hand-crafted copy in 70 minutes.
   **That is IMG-075's own durability argument, one field along.**
3. **The two mechanisms are COUPLED and only one shipped.** `[MEASURED 2026-09-03]` **2** active
   pages fleet-wide carry >1 instance of a component pairing an `llm` alt with a resolver
   `image_url`; **73** carry exactly one; **13** `llm` `*alt*` fields across **9** active components
   (unchanged from 2026-08-26), **6** paired with a resolver URL. **The contradiction class grows
   with every page this lane converts.**

**A LANDMINE was filed for the transferable half:** *alt text is written by a model that has never
seen the image, so grading a figure by its alt confirms the prose, not the picture.* Verified
dispatch run.

---

## 6. Where the ask really stands: THREE layers, and this lane owns only the top

1. **Can a figure survive regeneration?** ✅ **Done, reviewed, live, and now PROVED at the artefact.**
2. **Does anything compose an article out of illustrated sections?** grip-styles is the first, done
   by hand 2026-09-03; 718 has started the planner doing it for landing pages.
   `editorial_design_uplift` owns this.
3. **Are articles even IN THE PLAN?** ⚠ **Mostly not — and this is the floor.**
   `[MEASURED 2026-09-03]` on the 33 sites with a current plan: **tool 83%, blog-post 85%,
   guide 74%** have NO `site_plan_sections` row, against **landing 2%**, content 15%. The split is
   by page TYPE, not site health. **No plan row → `planSectionOrder` returns nil → binding
   disabled.** Nobody owns fixing it.

**Mechanism, read first-hand:** `create_blog_posts_action.go:212` — the article layout triple is a
**fallback** (`post.Sections` may be supplied), and the action writes `pages.sections` (**tier 3,
the cache**) and never `site_plan_sections` (**tier 1, the authority**).

**Who can write the authority** `[MEASURED 2026-09-03]` — three populations, because a Go-only grep
sees one and I published that mistake once:

| population | count | writes? |
|---|---|---|
| Go | **2** (`write_site_plan_action.go:668`, `apply_gap_plan_action.go:1067`) | yes; **neither on the article path** |
| live `agent_definitions` config SQL | 2 rows | **reads only** |
| **operator SQL in the repo** | **15 files** | **yes — a real third path** |

⚠ **The third is the trap.** Backfilling by hand fixes the pages that exist and nothing about the
route — dartsonline's nine July plan rows share ONE timestamp, and the 14 articles created since
have none. **A cheap unblocker for a canary, not phase one**, and "just pass richer `post.Sections`"
fills the cache and leaves the authority empty, so the ordinal still has nothing to name.

---

## 7. What I would do next, in order

1. **Re-probe the current build** (§2). Everything else assumes it.
2. **The RE-RENDER arm is PRE-REGISTERED, not fired — grade the next natural re-render against the
   prediction, do not fire your own.** `[MEASURED 2026-09-03 15:1xZ]` pre-flight on grip-styles: the
   plan's 11 site-level-filtered names and the 11 stored `slot_name`s **agree at every position**,
   0 locked slots, so **the prediction on record is that it BINDS per-section**; the disconfirming
   result is **all five sections showing ONE image**. Reasons I did not fire one, and they still
   hold: it is `dartsonline_traffic`'s freshly-finished page, the decisive save-path test has
   already passed, that page carries **12** `unresolved` `cta_links_stale` rerender items (all
   predating the 11:39Z replan), and `item_key='page_rerender:grip-styles'` was a live
   `idx_swi_dedup` collision risk. **NOTES §18** has the full reasoning. ⚠ Grade on the run's
   `resolved_data`, never the served bytes — an assemble-only re-render produces identical bytes
   whether the binding engaged or did nothing.
3. **Offer grip-styles to the 443/641 lane as Stage B's canary.** Five same-component instances,
   distinct subjects, distinct **correct** images — the images are independent ground truth for
   whether each heading is right, which no other page in the estate provides. Offer made in the
   CONTRIB; theirs to take.
4. **Watch `gamedesign.uk/index`** as 718's first planner-written set (4 rows `source='llm'`, three
   at the SAME ordinal `index:2` because 718's decomposition rule emits one entry per card image).
   The per-section map is **kind-first-wins within an ordinal**, so that section gets ONE of the
   three; the others are reachable only by literal asset key, which needs per-key component fields.
   ⚠ Its plan (`hero`, `featured-content`, `content-listing`, `generic-text-block`) and its live page
   (2 components) **disagree**, so the binding correctly stands down there today.
5. **Do not build the Phase-4 detector** (`check_unrendered_section_imagery`). The PLAN's "discovery
   has no driver" blocker is stale, the RUNBOOK's hand query does the job, and `bugfix_114` has
   offered a section-scope arm on `check_unrendered_page_imagery` (IMG-077) — cheaper than a new
   mechanism.
6. **Leave phase 3 (article planning) where it is** — nobody owns it, it is not this lane's, and it
   should not be closed by a backfill.

---

## 8. Traps this lane paid for — read before trusting a number

- **A step's collected value is its RESULT, not its prompt.** I tested
  `…_iter_2_generate_content` for the subject string, got ABSENT, and nearly filed it. That key
  holds `{result, type}` — the model's output — so the subject could not have been there whatever
  the truth was. **To ask what a model was told, read the agent config.** (2026-09-03)
- **A served-page census on `data-component` UNDERCOUNTS** — `generic-text-block` emits no such
  attribute. I read 7 sections on a page serving 11 and briefly had a defect. Count `class="section`
  families. (2026-09-03)
- **`alt` text is not evidence of image content** — LANDMINE, §5. Open the image.
- **A count of a population says nothing about whether it is GROWING** — I read "9 of 442" as
  "nothing selects it"; the refutation was `created_at`, in the table I had already queried.
- **A `LIMIT` counted section ROWS while my claim counted PAGES** — one page contributed 6 of 12.
- **A reference census matched the component NAME, not the filename** (`LIKE '%hero-about%'` finds
  `class="hero-about"`). Anchor on the extension; run a control. **LANDMINE.**
- **`updated_at` moved ≠ a re-render happened ≠ the resolver was asked** — three events, one word.
- **Filed ≠ ran** — an item can exist, be quoted as evidence, and have FAILED with `result = {}`.
- **A Go-only grep is not a fleet-wide census** — config SQL and operator SQL are two more writer
  populations.
- **I quoted a Go COMMENT as live config** — twice now: the re-render reason list (five, not two)
  and today the `Subject` field's "the v5 prompt renders it". **Cite the row.**
- **Ask for PROVENANCE, not just correctness** — *"have you read it, or is it relayed?"* caught a
  29-line-drifted citation, an overstatement and an unchecked inference inside a claim that was
  true. Keep your own `[UNVERIFIED BY ME]` until **you** have read it.
- **`git stash` is forbidden; commit by pathspec — and build the pathspec from `git status`, not
  memory.** I broke HEAD for eleven minutes naming two of three callers.
