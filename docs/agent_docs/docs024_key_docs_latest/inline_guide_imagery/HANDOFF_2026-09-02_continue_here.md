# HANDOFF — inline_guide_imagery. START HERE. Written 2026-09-02 (evening).

**The one-line version:** the durable per-section imagery binding is **built, council-APPROVED at
round 3, and live** (last verified on chassis `v1.0.1354`) — but **a fresh build `v1.0.1355`
deployed at ~20:56 and I could NOT verify it, because the kubeconfig token expired mid-probe.**
First job for the next session is that verification. The command is in §2.

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

## 2. ⚠ FIRST JOB: verify `v1.0.1355`. My probe is VOID, not negative.

I probed the new build and got **absent for every symbol including `PlanSectionsAction`, my
must-be-present control**. That does not mean the code is missing — it means the instrument
failed. `kubectl` then returned `You must be logged in to the server (Unauthorized)` for
everything, including `psql`. **The token expired** (it does every ~3 days; the owner refreshes
it), and my `2>/dev/null` turned a failed exec into the word "absent".

**This is the whole reason the control exists — a failing command and a missing symbol produce the
same output.** Do not read my last probe as evidence of anything.

Once the token is refreshed:

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

- **apis.uk/index — the first and only driver. ARMED, NOT EXERCISED.** They seeded six
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
- **`bugs_open/114` — the fleet finding, two CONTRIBs from me today.** The second is the big one:
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

1. **Verify `v1.0.1355`** (§2) once the token is refreshed. If the guards are absent, the binding
   is running unguarded — exposure is currently zero pages, but say so rather than assume it.
2. **Watch apis.uk/index for its first re-resolving render**, and read the served bytes. That is
   the mechanism's first real evidence, and it does not exist yet.
3. **Leave the guide recompose with `dartsonline_traffic`** — it is their site and their call, and
   the justification is the owner's 08-31 ask, explicitly NOT "give the mechanism a driver".
4. **Do not build the Phase-4 detector yet** (`check_unrendered_section_imagery`). ⚠ The PLAN's
   "discovery has no driver" blocker is **stale** — discovery IS driven daily now (corrected in
   the PLAN) — but the hand query in the RUNBOOK already answers it, and it found the two live
   misses now in `bugs_open/114`. Build it when there is more than one consumer to protect.
