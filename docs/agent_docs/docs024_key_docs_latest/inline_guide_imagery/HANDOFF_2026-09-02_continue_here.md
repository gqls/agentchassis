# HANDOFF — inline_guide_imagery. START HERE. Written 2026-09-02 (evening).

**The one-line version:** the durable per-section imagery binding is **built, council-APPROVED at
round 3, and VERIFIED LIVE on chassis `v1.0.1355`** — both halves, both replicas, both controls
behaving (§2). It has a driver (apis.uk seeded six rows) and **has still never run**, because that
page has not re-rendered since August. The next real evidence is its first re-resolving render.

**Lane docs:** `docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/`
(`PLAN_2026-08-14_durable_inline_guide_imagery.md`, `NOTES_inline_guide_imagery.md`,
`RUNBOOK_inline_guide_imagery.md`, `README_where_we_are.md`, this file).
**Register:** `docs026_concept_register/register/imagery.md` → **IMG-075** (and IMG-074, corrected).

---

## 1. What this lane is for

The owner asked (2026-08-13, restated 2026-08-31 naming ring/razor/shark grip on
`dartsonline.com/blog/grip-styles.html`) that guide articles carry explanatory imagery **inside**
the body, not just a banner. The plan reframed it correctly as a **durability** problem: in-body
`<figure>` markup lives in `article-body`'s single LLM-owned `content` field and dies on the next
regeneration.

---

## 2. Deploy verification — DONE for `v1.0.1355`, and ALREADY DATED

⚠ **A further chassis build was in progress at ~21:30 on 2026-09-02 and deploys within the hour**
(owner, at the time of writing). **So the verification below is `v1.0.1355`'s and expires with the
next roll — RE-RUN THE PROBE FIRST.** Nothing about IMG-075 should change (it is committed and was
in `v1.0.1354` and `v1.0.1355`), but "it was live on the previous build" is an inference, not a
measurement, and this lane has now been caught twice by exactly that gap in one day.

## 2a. What `v1.0.1355` measured — and how the first attempt lied

**`v1.0.1355` VERIFIED `[MEASURED 2026-09-02, after the token was refreshed]`** — both replicas
(`cd2h9`, `vppjz`), errors deliberately NOT suppressed:

```
PRESENT PlanSectionsAction          <- must-be-present control
PRESENT sectionRefForOrdinal        <- round 1, the binding
PRESENT sectionOrderAgrees          <- round 2, the drift guard
PRESENT sectionScopeRefOrdinal      <- round 2, the shared parser
PRESENT newSectionRef               <- round 2, the identity constructor
absent  sectionOrderAgreesNOTREAL   <- must-be-absent control ("exit code 1" printed, so the exec RAN)
```

⚠ **Keep the story of the first attempt, because it is the reason the control is not optional.**
Before the refresh, the identical probe returned **absent for all six — including the
must-be-present control** — and I nearly recorded a regression. `kubectl` was returning
`Unauthorized` (expired token, refreshed every ~3 days by the owner), and `2>/dev/null` had turned
a failed exec into the word "absent". **A failing command and a missing symbol produce the same
output; only the control tells them apart.** Drop the `2>/dev/null` — the visible
`command terminated with exit code 1` on the negative control is what proves the exec ran at all.

### The probe, for the next roll

I probed the new build and got **absent for every symbol including `PlanSectionsAction`, my
must-be-present control**. That does not mean the code is missing — it means the instrument
failed. `kubectl` then returned `You must be logged in to the server (Unauthorized)` for
everything, including `psql`. **The token expired** (it does every ~3 days; the owner refreshes
it), and my `2>/dev/null` turned a failed exec into the word "absent".

**This is the whole reason the control exists — a failing command and a missing symbol produce the
same output.** Do not read my last probe as evidence of anything.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers \
        -o custom-columns=NAME:.metadata.name | head -1)
for sym in PlanSectionsAction sectionRefForOrdinal sectionOrderAgrees sectionScopeRefOrdinal \
           newSectionRef sectionOrderAgreesNOTREAL; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe && echo "PRESENT $sym" || echo "absent  $sym"
done
```

**Read it like this:** `PlanSectionsAction` PRESENT and `sectionOrderAgreesNOTREAL` ABSENT means
the instrument works — only then are the middle four meaningful. **Run it on BOTH replicas.**

⚠ `kubectl logs … | grep 'build provenance'` does **not** work on this service: the phrase appears
in LLM prompt text the chassis logs, so a careless grep returns a hit shaped exactly like a stamp.
Already a LANDMINE; do not re-derive it.

⚠ **A SECOND cause of "absent with clean controls", filed by another lane the same evening**
(`LANDMINES.md`, *"A capability probe for INERT code reads ABSENT … the linker's dead-code
elimination strips uncalled functions"*). It does **not** explain tonight's result — mine had a
FAILING control, which is the token — but it qualifies the recipe: a symbol with **zero callers**
can be stripped from the binary and probe absent on a build that genuinely contains the commit.
That is not a risk for these four symbols (`sectionOrderAgrees` is called by `planSectionOrder`,
which is called by `ensureAssets`, which every section render reaches) — but if you ever probe for
something inert, **verify by ANCESTRY instead**: read the pod's `git_commit` from
`service_binary_capabilities` and `git merge-base --is-ancestor <commit> <stamp>`. Two different
failure modes, one identical symptom; tell them apart by whether your control passed.

**Last VALID measurement:** `v1.0.1354` (pods 15:39/15:53), both replicas, controls both ways —
`sectionRefForOrdinal` PRESENT, `sectionOrderAgrees` PRESENT, `sectionScopeRefOrdinal` PRESENT,
`PlanSectionsAction` PRESENT, synthetic control ABSENT. Both halves of IMG-075 were live then.

---

## 3. What shipped (all committed, HEAD builds)

**IMG-075 — a section-scope `site_plan_imagery` row now binds to the ONE section its `scope_ref`
ordinal names.** Before it, every section on a page declaring `site_assets.illustration` resolved
the *same* URL (kind first-wins); the ordinal was filtered on and thrown away.

| commit | what |
|---|---|
| `cb698ee58` | the binding + register entry IMG-075 |
| `844eb3023` | **fix HEAD** — a third `resolve()` caller I left off the pathspec |
| `38178d549` | round 2: one occurrence rule, one ordinal parser, the drift guard |
| `4084481d7` | round 3 advisories discharged (mutation-proven identity test; probe list fixed) |

**Council: `Council-Reviewed: 2979c27f-1545-47c5-b28d-f8a700bb1cb0` — APPROVED round 3**, 12 seats,
1 advisory none high. Three rounds; each REVISE found something real (see §6).

**Design, in one paragraph:** the ordinal is translated ONCE, in `ensureAssets`, into a
`sectionRef{Name, Occurrence}` against the plan's own section order — **never a position integer**,
because `site_plan_sections.ordering` is 0-based counting site-level slots while
`page_components.position` is 1-based on 847 of 1,065 pages and neither on 128. Both render paths
count occurrences with the estate's shared `InstanceCounter`. `sectionOrderAgrees` stands the whole
binding down — rather than mis-binding — when the plan's order and the live order disagree.

---

## 3b. ⚠ THE LAYER BELOW EVERYTHING ELSE: ~84% of articles are NOT IN THE PLAN AT ALL

**Found by `dartsonline_traffic` 2026-09-03, independently reproduced here with the join control
they asked for.** This bounds IMG-075 harder than the composition gap, and it is **this lane's own
first stated degrade case**: no `site_plan_sections` row for the page → `planSectionOrder` returns
nil → binding disabled → page-wide resolution. It cannot engage at all.

`[MEASURED 2026-09-03, active pages on the 33 sites that have a current plan with sections]`:

| page_type | built | in plan | absent | % absent |
|---|---|---|---|---|
| tool | 294 | 49 | 245 | **83%** |
| blog-post | 262 | 40 | 222 | **85%** |
| guide | 96 | 25 | 71 | **74%** |
| content | 149 | 126 | 23 | 15% |
| section-index | 56 | 46 | 10 | 18% |
| landing | 49 | 48 | 1 | **2%** |

**The split is by page TYPE, not site health** — structural pages are essentially always planned,
content pages essentially never. So a site passes every site-level "is it planned?" check while
sending every article it owns down the fallback path.

✅ **Join control (they flagged their own join as verified on one site only):** `pages.name =
site_plan_sections.page_name` — **0 of 33 sites have zero joining pages**, so the
naming-mismatch confound does not apply anywhere in this population. The `landing` row at 2%
absent is a second built-in control: a broadly broken join could not produce it.

**The mechanism, now READ FIRST-HAND `[MEASURED 2026-09-03]`** (the relayed version had a stale
line number, an overstatement and an unchecked inference — all three corrected by that lane when
asked whether they had opened the file; the claim survived, in a better form):

- `create_blog_posts_action.go:212` — the triple is a **FALLBACK, not a hardcoded layout**:
  `sections := post.Sections; if len(sections) == 0 { sections = []string{"hero","article-body","call-to-action"} }`.
  A caller MAY supply its own.
- `grep -n site_plan_sections` on that file: **no occurrence.** It writes
  `INSERT INTO pages (…, sections, page_spec)` — **tier 3, the materialised cache** — and never
  tier 1, the authority.
- **And the positive half — CORRECTED 2026-09-03, because I censused Go and said "fleet-wide".**
  `[MEASURED 2026-09-03]` there are **THREE** writing populations, not one:

  | population | count | writes? |
  |---|---|---|
  | Go (`platform/`) | **2** — `write_site_plan_action.go:668`, `apply_gap_plan_action.go:1067` | yes; **neither on the article-creation path** |
  | live `agent_definitions` config SQL naming the table | 2 rows (`build-site-planner`, `required-fields-missing-handler`) | **READS ONLY** |
  | **operator SQL in the repo** (`INSERT INTO site_plan_sections`, `*.sql`) | **15 files** | **YES — a real third path** |

  ✅ **The config row is a negative control that could have broken the claim and didn't** —
  config-embedded SQL is this estate's documented blind spot for code-only censuses (it is why
  `cmd/config-key-audit` is in council scope), so "2 Go writers" surviving it is worth more than
  the number. ⚠ **My original wording was a Go-and-`platform/`-scoped grep reported as
  fleet-wide** — same shape as the other unit errors this lane logged: the population I measured
  was not the population my sentence named.
  **And the 15 include `dartsonline_traffic/SQL_2026-07-29d_article_sections.sql` — the single
  file that is the entire reason any dartsonline article has a plan row.**

⚠ **PHASE ONE THEREFORE HAS THREE POSSIBLE SHAPES, AND THE THIRD IS A TRAP.** (1) Change one of
the two Go writers so article creation writes the authority. (2) Add a third writer. (3) Backfill
by operator SQL — **which fixes the pages that exist and nothing about the route.** dartsonline is
the worked example: `SQL_2026-07-29d` gave nine articles plan rows in July, and the **14 articles
created there since have none**, because every future article still arrives the same way.
**Backfill is a cheap unblocker for a canary; it is not phase one** — and if it reaches the owner
as one it will look done and then quietly stop being true with the next article. (That framing is
`dartsonline_traffic`'s, who are their own cautionary example for it.)

⚠ **THE SENTENCE THAT STOPS PHASE ONE BEING MISTAKEN FOR A CHEAP FIX** (theirs, and it is the
reason separating fallback-from-inference mattered): **"just pass richer `post.Sections`" is NOT a
fix.** It changes what the page is composed OF and leaves the plan empty — so the `scope_ref`
ordinal still has nothing to name, and this lane's binding stands down exactly as before. **Any
fix must write the authority**, i.e. go through one of those two writers or add a third.

⚠ **AND IT CORRECTS SOMETHING I PUBLISHED.** I reported "9 of dartsonline's 13 `/blog/*` pages bind
today" and implied that was the framework working. **Confirmed here:** all 27 plan rows for those
nine pages share **one timestamp, `2026-07-29 13:28:03.58521Z`** — a single hand-written backfill
by that lane, from a seed whose header says *"Nobody ever decided what blocks a guide page should
contain."* The 14 articles created there since have no plan rows. **So my figure was a fact about
one seed, not about the platform**, and reading it as the latter badly overestimates how many
pages anywhere are ready for this mechanism.

**So the ordering is three deep, and this lane owns only the top:**
`get articles into the plan` → `compose them with illustrated blocks` → `bind figures per section`.
The third is IMG-075 and is done. The second is `editorial_design_uplift`'s. **The first looks like
nobody's.**

## 4. Where the actual ask stands: the binding was NOT the bottleneck

`[MEASURED 2026-09-02]`

```
active content pages (blog/guide/article) fleet-wide   442
  ...carrying ANY illustration-capable section           9
  ...carrying MORE THAN ONE (can host a figure SET)      1   (apis.uk/index, hand-built)
  ...that are blog-post or guide                         0
```

**The plumbing is finished and the composition it needs exists on one page.** But the sharper
correction, made an hour later and easy to get wrong in both directions:

**Migration 644 WORKED — the planner selects the illustrated component, six sites in seven days**
(webdesign.uk 5h after 644, idea.uk/tools, lendzy, oufe, remortgagecalculator, robot-hands — the
last two on 09-02). Three pages pre-date it. **So "nobody drives what we build" is false here.**
The gap is the PATTERN: 8 of 9 are `landing` pages with exactly ONE illustrated section as an
accent, and **zero are blog-post or guide**. The planner has never been asked to compose an
ARTICLE out of several.

**And the reason nothing REQUESTS illustrations (answered tonight, NOTES §15):** the live
`build-site-planner` prompt carries the full `kind` enum and then says *"Use sparingly in v1 —
most plans will have zero section-scope entries"*, sets a minimum naming only `logo` and `hero`,
and gives a worked example whose `sections` block contains **only icons**. `hero 399 / icon 211 /
logo 50 / illustration 25 / infographic 1` is the prompt working as written — **obedience, not a
defect.** The `designblog.co.uk` lane took this to the owner, who ruled to change it: **migration
718 applied 19:59Z tonight** (their work, my §15 as its evidence base), flipping all four
suppressors. ⚠ It changes what is REQUESTED; it does **not** put images inside articles, which
needs the composition half.

---

## 5. Open threads, and who owns them

- **⚠⚠ THE SHARPEST OPEN RISK TO THIS MECHANISM, and it is not in the binding code.** A genuine
  sections-path run on `dartsonline.com/tool-brand-comparator` (`section_data_resolved`,
  attributed by `source_item_id`, 2026-09-03 00:40:40Z) **did not write a newly-declared
  `site_assets.hero` field**, on a page that HAS a resolvable hero (`content_hero_tool_brand_comparator`,
  active — arm 2 of `ensureAssets`). I tried to dissolve it with the mundane explanation ("no hero
  to find, so omitting is correct") and **it survived** — see `bugs_open/114`, this date.
  **IMG-075 resolves `site_assets.illustration` on that same path.** So apis.uk's six seeded
  figures may fail to bind when that page re-renders **for a reason unrelated to this lane's
  code**, and the natural reading on the day will be "IMG-075 does not work". `[UNVERIFIED]` — one
  instance on `site_assets.*`, one on `query.*` (`bugs_open/425` §2, four reproductions). The
  components lane owns the discriminating test (batch 690, `remortgagecalculator.uk/about`); its
  result decides whether this lane has a problem at all.
  > **A control worth offering if anyone runs it:** apis.uk/index is a page where the field is NOT
  > newly declared (`image_url` has been on `illustrated-text-block` since 08-24, values stored).
  > Firing `section_data_resolved` there vs their newly-declared case separates "never re-resolves"
  > from "never writes a field added after the row was last built". Not this lane's page to fire.
- **apis.uk/index — the first and only driver. ARMED, AND ITS ONE ATTEMPTED TEST FAILED.**
  ⚠ **Updated 2026-09-03 — this is stronger than "not exercised yet".** `[MEASURED]` that lane
  filed a `reason='section_data_resolved'` rerender at **the identical microsecond** as the six
  imagery rows (16:47:03.788197Z — one transaction, exactly right) and **it FAILED** at 18:21:41Z
  with `result = {}` and no detail recorded. A second item completed 09-03 07:06 carrying **no
  reason**, i.e. the assemble path, which asks the resolver for nothing. Rows still
  `updated_at = 2026-08-24`. ⚠ **WHY it failed, supplied by that lane 2026-09-03 (I could not see it — `result` is `{}`):
  the SAVE GUARD refused it** — *"overwrite: REFUSED for page index — re-confirmed too little of
  what is stored (prune_floor…)"*, observed in the pod logs and recorded in
  `apis_uk_bees_homepage/NOTES`. **So it is NOT the resolver and NOT the locks per se.** The page
  is wedged between its own protections: assemble mode completes and redeploys stale bytes;
  re-resolve mode **resolves successfully and is then refused at save**. All seven rows are also
  `lock_type='permanent'`, `locked_by='apis-uk-bees-lane'`.
  ⚠⚠ **THEREFORE DO NOT COUNT apis.uk TOWARD THE 425 SHARED-ROOT-CAUSE PILE.** It fails at a
  *different stage* from the dartsonline case (where the write completed and simply lacked the
  field). Binning it with the others would corrupt a hypothesis three lanes are accumulating
  evidence for. Their open route is the site-level `rerender-pages` fan-out, which served this
  page successfully on 08-24 with the locks in place. **I withdrew my own offer of it as the "existing declared field" control and told
  that lane their test had failed** (nothing about the failure is visible on the item).
  **The existing-field half of the experiment therefore has NO VENUE yet** — any page with a
  resolver-sourced field declared *before* its last build, with values stored, would serve.
  ⚠ **PARK THAT HUNT UNTIL BATCH 690 LANDS** (agreed with `designblog.co.uk`, 2026-09-03): if 690
  writes the newly-declared field, the pairing question mostly dissolves and no control is needed;
  only if it fails is the venue worth finding, and the components lane's census machinery is the
  fastest way to derive that population. **Do not spend the query before the discriminator runs.** They seeded six
  section-scope illustration rows at **16:47:03Z** on my CONTRIB (`apis_uk_bees_homepage/CONTRIB_2026-09-02_…`).
  But that page's `page_components` still read `updated_at = 2026-08-24` — **nothing has
  re-resolved, so the branch has never actually run.** Evidence arrives at its next re-resolving
  render or, decisively, its next `content_rewrite`. ⚠ An assemble-only re-render changes nothing
  (the images are already in `content_data`) and must not be read as failure.
- **dartsonline `grip-styles` — the owner's actual case. Theirs to recompose.** 9 of their 13
  `/blog/*` pages would bind today, grip-styles among them; 4 have **no plan sections at all**
  (traced to their own content batch — theirs to chase). ⚠ I gave them a recipe and had to correct
  it: it would be the **first page in the estate** composed that way, so it is an experiment, not
  a pattern. Sequencing they need: **recompose → seed rows → rebuild → verify → then re-render**
  (a re-render in the gap correctly stands down).
- **⚠ `bugs_open/114` — the component fix LANDED at 20:15:47Z and has repaired ZERO pages.** Six of
  the seven components gained an asset-sourced image field in one transaction; `component_expresses`
  now returns `{image}` for all six. **The damage is unchanged: still 61 of 65 orphaned, and NINE
  pages have re-rendered since the fix without recovering.** The field is necessary and not
  sufficient — only `image_landed`/`section_data_resolved` re-resolve, so a page can rebuild all
  night without asking the resolver for anything. **Do not let anyone report this class closed on
  the component change alone**; read `spec->>'reason'` and the served bytes.
  > ⚠ **CORRECTED 2026-09-03 — my "only two reasons re-resolve" was WRONG, and it is mine.** I
  > quoted `rerender_page_sections_action.go:47`'s comment; the LIVE `page-rerender` config gates
  > on **FIVE** — `image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`,
  > `literal_markdown`. The comment has drifted from the config it describes. **And the deeper
  > claim is UNDER TEST:** whether the sections path re-resolves `site_assets.*` at all is
  > unsettled (`bugs_open/425` §2 reports it does not for `query.*`, four reproductions; the one
  > page traced as recovering did so via the BUILD path). **AND I have since RETRACTED the "nine
  > re-rendered, none recovered" evidence itself (2026-09-03):** ten of the twelve pages whose
  > `updated_at` moved are `seotools.co.uk` tool pages with **no `page_rerender` item near the
  > write** — BUILD-path writes on a site being built out, not re-renders. `updated_at` moved ≠ a
  > re-render happened ≠ the resolver was asked. The sections path has essentially never been
  > exercised against this class, so **nothing measured so far says whether it re-resolves
  > `site_assets.*`** — the components thread's `image_landed` batch is the first real test.
  > ⚠ Their trap worth carrying: **"currently correct" is a STATE, not evidence of a transition**
  > (10 of 66 read as recovered in a naive sweep; all were already-correct pages).
  > One free data point to check first: `dartsonline.com/tool-brand-comparator` moved at 00:40Z
  > with a `section_data_resolved` beside it — the only qualifying reason in the set.
- **A mechanical section-compatibility guard is ROUTED HERE and NOT BUILT** (718's `bug_historian`
  advisory, architecture concurring as a monitoring item). Verdict on feasibility: **sound** —
  `component_expresses`' predicate is `source LIKE 'site_assets.%' AND type IN ('url','image','image_url')`,
  which is exactly "can this section receive a server-resolved image". ⚠ **Anyone validating it
  from now measures a REPAIRED population** (the six were fixed at 20:15), so a clean result is not
  evidence the guard is useless. Sequencing: let the first post-718 plan set the base rate, and
  consider `bugfix_114`'s offer of a section-scope arm on `check_unrendered_page_imagery` (IMG-077)
  before building anything new.
- **`bugs_open/114` — the fleet finding, three CONTRIBs from me today.** The second is the big one:
  **61 pages** have a hero generated, deployed and `active` specifically for them and render
  something else. Cause: seven components whose template reads `{{or .hero_url .background_image}}`
  while their schema declares no `site_assets` source (`hero-tool` biggest at 76 instances).
  ⚠ **Not uniform — 4 of 65 render their own correctly** via another writer
  (`leopardessconsulting.co.uk/tool-automation-savings-estimator`); **identify that route before
  editing seven components, it may be the cheap fix.** Component-library thread has it.
- **Round-2/3 guards vs the roll:** committed, live on `v1.0.1354`, **unverified on `v1.0.1355`** (§2).

---

## 6. Traps this session paid for — read before trusting a number

- **A count of a population says nothing about whether it is GROWING.** I read "9 of 442" as
  "nothing selects it" and published it. The refutation was `created_at`, in the table I had
  already queried.
- **A `LIMIT` counted section ROWS while my claim counted PAGES.** One page contributed 6 of 12
  rows; two older pages fell off the end. `GROUP BY` the page before capping.
- **An asset-reference census matched the component NAME, not the filename** — `LIKE '%hero-about%'`
  finds `class="hero-about"`. Anchor on the extension, and run a control. **Now a LANDMINE.**
- **A remedy is fitted to a POPULATION, not to a defect** (a peer rolled back a migration for this
  the same day: 292 of 301 pages already showed the same image). Ask what the HEALTHY instances do.
- **A positive control turns a broken instrument into a readable result** — §2 is tonight's
  instance, and it is the only reason I am not reporting the new build as regressed.
- **`git stash` is forbidden; commit by pathspec — and build the pathspec from `git status`, not
  memory.** I broke HEAD for eleven minutes by naming two of three callers (`WRONG_CALLS.md`).

---

## 7. What I would do next, in order

1. ~~**Verify `v1.0.1355`**~~ **DONE** (§2) — both halves live on both replicas, controls clean.
   Re-run it after the NEXT roll, not this one.
2. **Watch apis.uk/index for its first re-resolving render** — re-checked after the roll and it is
   STILL unexercised (`updated_at` 2026-08-24 on all seven rows), so the roll did not rebuild it, and read the served bytes. That is
   the mechanism's first real evidence, and it does not exist yet.
3. **Leave the guide recompose with `dartsonline_traffic`** — it is their site and their call, and
   the justification is the owner's 08-31 ask, explicitly NOT "give the mechanism a driver".
4. **An open offer from `bugfix_114_imagery_wiring`, which is a decision, not a task.** They
   shipped `check_unrendered_page_imagery` (**IMG-077**, inert until the roll) covering the
   ContentHeroKey population, and deliberately left the **section-scope** cases out of the first
   cut because they are this lane's vocabulary. If a section-scope arm is wanted once that is
   live, say so in `bugfix_114_imagery_wiring/NOTES_imagery_wiring.md` and it rides their
   machinery — which is cheaper than building ours. Also from them: migration **709** deletes
   `sites.content_data.illustration_url` on 4 sites and **cannot touch this lane's resolution**
   (the content_data fallback covers `hero_url`/`logo_url` only) — so do not re-derive that when
   the key disappears.
5. **Do not build the Phase-4 detector yet** (`check_unrendered_section_imagery`). ⚠ The PLAN's
   "discovery has no driver" blocker is **stale** — discovery IS driven daily now (corrected in
   the PLAN) — but the hand query in the RUNBOOK already answers it, and it found the two live
   misses now in `bugs_open/114`. Build it when there is more than one consumer to protect.
