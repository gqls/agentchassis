# HANDOFF — inline_guide_imagery. START HERE. Written 2026-09-02, rewritten 2026-09-04 16:00Z.

**Status in one line:** IMG-075 is built, council-APPROVED, and **proved end to end including the
durability test** — and as of **2026-09-03 22:05:57Z the blocker that made the result unusable has
CLEARED**, so for the first time both halves of the owner's ask are live at once and **nothing in
this lane is waiting on anything.**

> ## ⚠ WHAT CHANGED SINCE YESTERDAY — read this before the rest
>
> **1. `bugs_open/443` Stage B LANDED.** The `page-content-writer` prompt now renders
> `current_section.subject` (**14** `current_section.*` paths where yesterday there were 13 and
> `subject` was not among them). `[MEASURED 2026-09-04]` **The acceptance test this lane and
> `dartsonline_traffic` agreed on PASSES on the discriminating population** — §5a has the numbers.
> **The writer defect that made grip-styles' words wrong is fixed.** Applied **by hand at
> 2026-09-03 22:05:57Z**; `schema_migrations` carries **no row** for 640 or 641, so a filename gate
> still reads "not applied" — the case the landmine was written for.
> ⚠ **THE GATE QUERY BOTH LANES CIRCULATED WAS BROKEN — use §2a's, not `section_subject`.**
>
> **2. The supply half landed too — under `762`, not 640.** `[MEASURED 2026-09-04]` plan rows
> carrying a subject: **0% on 08-31 → 5.1% on 09-02 → 62.1% on 09-03 → 84.5% today** (125 of 148).
> ⚠ **CORRECTED — this bullet said the planner carried no subject rule and that was WRONG.**
> `762_build_site_planner_rule17_subjects_become_opening_lines.sql` applied **2026-09-03 19:22:35Z**
> and **IS in `schema_migrations`**; the live planner does carry a subject rule. **640 is both
> unapplied AND superseded** — its own idempotence anchor (`may also carry a "subject"`, exact
> substring) returns `f`. I had inferred "no rule" from the absence of a name I guessed.
> ⚠ **THE LEDGER FAILED IN OPPOSITE DIRECTIONS ON THE TWO HALVES OF ONE FIX** — 641 applied by
> hand with **no** row; 640 **correctly** recorded as unapplied while its content shipped under a
> higher number. **So "unrecorded" is only one of the two failure modes: a file can be truthfully
> absent from the ledger and live anyway.** A lane checking `schema_migrations` for 640 gets a true
> answer and the wrong conclusion. Found by `dartsonline_traffic` chasing my own question back.
>
> **3. `grip-styles` is still REVERTED and that has not changed** — 3 plan sections, 0 section-scope
> imagery rows, 3 `page_components`. **The five illustrations are still `active` assets**, so the
> retry is a plan seed plus a rebuild and the expensive half is paid for. **It is
> `dartsonline_traffic`'s to run, not this lane's** — see §7.1.
>
> **4. This lane's pre-registered prediction is LIVE AGAIN** the moment any page carries several
> section-scope figures. It was void only because the proving page lost its rows. §7.2.

**Lane docs:** `docs/agent_docs/docs024_key_docs_latest/inline_guide_imagery/` —
`PLAN_2026-08-14…`, `NOTES_…` (technical log, newest at the bottom; §17–§22 are the proof, the
revert and four corrected numbers), `RUNBOOK_…` (the queries, with their traps),
`README_where_we_are.md` (owner's plain-prose log),
`SUMMARY_2026-09-03_inline_guide_imagery.md` (milestone read-out + a correction footer), this file.
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

## 2. THE LIVE STATE, and how to re-probe it

⚠ **Re-probe before trusting anything below.** `[MEASURED 2026-09-04 15:4xZ]` chassis
**`v1.0.1360`**, both pods on commit `239ab3626`, started 22:06–22:07Z on 2026-09-03 — i.e. **the
roll that carried 641**.

> ⚠ **AND THOSE TWO FIGURES WERE SUPERSEDED WHILE THIS PARAGRAPH WAS BEING WRITTEN.** A fleet roll
> to **`v1.0.1361`** (cut **`06c0b18f2`**, dated 2026-09-04 15:22Z) was building and pushing at
> 15:29–15:44Z, restarting every agent pod. The numbers above were true when measured and are the
> wrong ones now — which is this section's own point, demonstrated on itself. **Re-run the query.**
>
> **What the roll does NOT touch, and this is the load-bearing part: the writer fix is DB config,
> not code.** 641 lives in the `agent_definitions` row for `page-content-writer`; a chassis roll
> replaces binaries and leaves it alone. **§5's result survives the roll** — re-confirm with the
> artefact query in **§2a** (with both controls), not by re-reading this file.
>
> **And IMG-075's code rides it:** `[MEASURED 2026-09-04]` all four commits (`cb698ee58`,
> `844eb3023`, `38178d549`, `4084481d7`) are ancestors of `06c0b18f2` by
> `git merge-base --is-ancestor`. They were already live since `v1.0.1351`; this only confirms the
> new cut does not regress them.
>
> ⚠ **A note on that check, because I got the reading wrong first time.** My "must NOT be an
> ancestor" control returned *ancestor*, and I briefly wrote the instrument off as broken. The
> instrument was fine — **I had mislabelled which side of the cut my own commit fell on**, because I
> recognised the cut's sha from a session-start log and assumed it was old. It is from today.
> `git show -s --format=%ci <sha>` settled it in one command. **A control that "fails" may be
> refuting your label rather than your tool; check the dates before you distrust the instrument.**

**Ask the pod what it is running (CLAUDE.md was REORDERED 2026-09-04 — this table is now FIRST,
ahead of the log line and the binary probe):**

```sql
SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
 WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
```

⚠ **Filter on `pod_name`, not the `service` column** — that column also carries
`agent-landmine-verifier-*` and friends sharing the image, which may have rolled at a different
time. ⚠ **It is a TWO-HOUR WINDOW, not a history** (`RetentionWindow`): it answers *what is running
now*, and it answers a question about the past with today's survivors, silently. Dating anything
older than two hours needs a source that is not pruned (`kubectl get rs -l app=agent-chassis
--sort-by=.metadata.creationTimestamp`).

**The capability probe, if you need the symbols themselves** (unchanged, still valid):

```bash
PODS=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=NAME:.metadata.name)
for POD in $PODS; do echo "== $POD =="; for sym in PlanSectionsAction sectionRefForOrdinal \
    sectionOrderAgrees sectionScopeRefOrdinal newSectionRef sectionOrderAgreesNOTREAL; do
  timeout 40 kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe \
    && echo "PRESENT $sym" || echo "absent  $sym"; done; done
```

**Read it like this:** `PlanSectionsAction` **PRESENT** and `sectionOrderAgreesNOTREAL` **ABSENT**
means the instrument works — only then are the middle four meaningful. Use the per-exec `timeout`:
without it the loop over two pods exceeds a 2-minute tool limit and the partial answer looks like a
partial deploy.

⚠ **Do NOT suppress stderr.** On 2026-09-02 the identical probe returned "absent" for all six
*including the must-be-present control* because `kubectl` was returning `Unauthorized` (token
expires ~3 days; the owner refreshes it) and `2>/dev/null` had turned a failed exec into the word
"absent". **A failing command and a missing symbol are the same output; only the control separates
them.**

⚠ **`kubectl logs … | grep 'build provenance'` does not work on this service** — the phrase appears
in LLM prompt text the chassis logs. Already a LANDMINE.

⚠ **A capability probe cannot see code nothing CALLS** — the linker drops it, so a genuinely inert
symbol probes absent on a build that contains the commit. Not a risk for these four. For an inert
symbol verify by ANCESTRY (`git merge-base --is-ancestor <commit> <the stamp>`).

---

## 2a. THE GATE: is the per-section subject live? (and the broken version to stop using)

```sql
SELECT bool_or(default_config::text LIKE '%current\_section.subject%') AS capability_live,
       bool_or(default_config::text LIKE '%current\_section%')          AS control_present,
       bool_or(default_config::text LIKE '%current\_section.subjectNOTREAL%') AS control_absent
FROM agent_definitions WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

`[MEASURED 2026-09-04]` → `t / t / f`. **That is the gate for the grip-styles retry and for any
claim that the writer defect is fixed.**

⚠ **THE VERSION THIS LANE PUBLISHED YESTERDAY WAS BROKEN, AND IT RETURNED THE RIGHT ANSWER ANYWAY.**
It was `default_config::text ILIKE '%section_subject%'`. **`_` is a single-character WILDCARD in
`LIKE`/`ILIKE`**, so it matched the unrelated path `section.subject`. `[MEASURED 2026-09-04]` the
escaped literal `LIKE '%section\_subject%'` is **FALSE** — that string is not in the config at all
— while the capability *is* live, under a different name. **A broken instrument and a renamed fix
cancelled out: right verdict, wrong mechanism, and nothing about it looked partial.** Caught by
`dartsonline_traffic` re-testing the gate they had published to me. Corrected in `LANDMINES.md` too,
where it was the fleet-wide worked example for "ask the running agent for the capability".

**Two rules out of it:** escape every `_` in a `LIKE` pattern, and **prefer testing the
interpolation the template actually performs over a key name you expect it to contain.**

⚠ **And a control can be inapplicable rather than informative.** The same query run against
`build-site-planner` returns `f` for the capability *and* `f` for `control_present` — that agent has
no `current_section` at all, so its control cannot fire and the `f` is uninformative rather than
evidence. **640-did-not-land is still true**, but establish it from the planner's own vocabulary,
not from a control borrowed from a different agent.

**What the live template actually does** (read at the row, 2026-09-04) — it interpolates the
section's own subject *and* enumerates the siblings':

```
{{if .current_section.subject}}## This section
{{.current_section.subject}}
{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
```

So it also addresses **`bugs_open/151`** — the writer having no memory of what sibling sections
said — in the same change. ⚠ **This is why §5a's naive test could not fail:** every sibling subject
appears in every prompt *by design*, so "does this subject appear in this prompt" is `6/6` for all
six. The discriminating read is the `## This section` block alone.

---

## 3. What shipped

**IMG-075 — a `scope='section'` `site_plan_imagery` row binds to the ONE section its `scope_ref`
ordinal names.** Before it, every section on a page declaring `site_assets.illustration` resolved
the *same* URL (kind first-wins); the ordinal was filtered on and thrown away.

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

## 4. ✅ THE MECHANISM WAS PROVED, both halves — on 2026-09-03, on a page since reverted

`dartsonline.com/blog/grip-styles.html`, on `v1.0.1358`. Full account: **NOTES §17**, §19–§20.

| time (UTC) | what | item |
|---|---|---|
| 11:39 | darts lane recomposed the plan to 11 sections + seeded 5 figures | `SEED_2026-09-03` |
| 12:41–12:47 | the five illustrations generated, went `active` | 5× `needs_imagery` |
| 12:47→13:02 | **rebuild through the writer** | `d5edd37b` `needs_content_page` |
| 14:00→14:11 | **a SECOND full regeneration**, fired automatically by the last asset landing | `8bd71ef8` `needs_page` `reason=image_landed` |

**Binding:** run 1's writer (`837bd4ea`) resolved **five distinct URLs in plan order** on
`process_sections_loop_item_N.resolved_data` — ring / razor / shark / smooth-barrel / combination.
⚠ **The pre-IMG-075 result AND a stand-down both look like five IDENTICAL URLs**, so the failure
shape is a run of identical URLs, not an error.

**Durability — the decisive test, and it passed.** Run 2 spawned a second `page-content-writer`
(`74d6b7e4`) that rewrote every heading and paragraph. `section_output_2` prose differs between the
runs; `illustration-ring-grip.jpg` does not. `carried_fields` was **none**, and the carry could not
have supplied them anyway — five sections sharing one `slot_name` are dropped from the carry map by
`ensureStoredContent`'s conflict rule — **so the figures were re-derived from `site_plan_imagery`.**
**Confirmed independently by `dartsonline_traffic`** at their own 16:49Z page snapshot (five
distinct illustrations, one per section), after they initially read the property as unexercised and
traced the error to their own query window. Served bytes at 14:11:46Z: 11 sections, five `<figure>`
blocks, five distinct files, each **200 at 1071×800**, invented sibling **404**, all five visually
correct.

⚠ **"Survive a `content_rewrite`" names the EVENT** — prose rewritten over a built page — **not the
`item_type`.** No item of `item_type='content_rewrite'` has been fired at a multi-figure page, so
that **dispatch path** is untested; do not cite this result as covering it.

⚠ **The RE-RENDER path is still unproven on a multi-figure page.** Both runs were the build/save
path. `rerender_page_sections` takes its live list from stored `page_components` slots rather than
`pages.sections`, so it feeds `sectionOrderAgrees` a **different list**.

---

## 5. ✅ THE WRITER DEFECT IS FIXED — this is the day's news

Yesterday this section read *"the figures are right and the words beside them are wrong"*. That is
no longer true of the mechanism, though it is still true of the reverted page's stored copy.

**What was wrong:** `plan_sections_action.go`'s `Subject` doc comment claimed *"the v5 prompt
renders it"*; the live `page-content-writer` referenced 13 `current_section.*` paths and `subject`
was not one. So five sections with five distinct subjects received effectively one brief, and
grip-styles was written about the **ring** grip five times under five different correct photographs.
That was `bugs_open/443` Stage B, predicted in writing by that lane before this page existed.

**What changed:** `[MEASURED 2026-09-04]` `page-content-writer` `version=2`,
`updated_at 2026-09-03 22:05:57Z`, and `current_section.subject` is now among **14** rendered paths.
The roll that carried it is the one both chassis pods are running (`239ab3626`, 22:06–22:07Z).

### 5a. The acceptance test, run on the population that can actually fail it

**The test both lanes agreed:** after the fix, **N sections must show N DISTINCT prompt hashes**,
scoped by `orchestration_id`, never by a time window.

⚠ **My first pass was NON-DISCRIMINATING and I nearly reported it as a pass.** Every post-fix run
showed N sections → N distinct prompts — but **on a page whose sections are all different
components, distinct prompts prove nothing**; they would differ anyway. The test only discriminates
where a component **repeats**, which is exactly the shape that failed before.

**The discriminating case** `[MEASURED 2026-09-04]`, run `be79d5a2` (copyonline.co.uk):
**6 sections, only 2 distinct component names — 4 repeats — all 6 carrying a subject.**

- ~~**6 sections → 6 distinct prompt hashes.**~~ ⚠ **WITHDRAWN AS EVIDENCE 2026-09-04 — and this
  retires the acceptance test both lanes agreed yesterday.** The template's sibling list is
  `{{range}}…{{if ne $s.subject $.current_section.subject}}`, i.e. **each prompt enumerates every
  sibling subject EXCEPT its own** — so every prompt differs from every other **structurally, even
  if the `## This section` block were empty**. Distinctness is guaranteed by the design and cannot
  fail. It is still true that the pre-fix shape collapsed 11 sections onto three prompts; it is no
  longer a test. Caught by `dartsonline_traffic`.
- ⚠ **A naive "does this subject appear in this prompt" test returns 6/6 for EVERY subject** — the
  prompt carries the whole page outline, so all six subjects appear in all six prompts. **That test
  cannot fail and is not evidence.**
- - **THE ONLY DISCRIMINATING FORM, and it is the whole test now:** each prompt carries a
  `## This section` block, and its content must be *that* section's own subject. **6 of 6 MATCH**,
  index for index, against the plan. **Grade a retry on this, per prompt — not on hash counts.**
- Controls: `ZZNOTREAL` absent from all 6; the plan subjects present.

**So the writer now receives a per-section brief.** The supply half is live too — **84.5%** of plan
rows created today carry a subject (125 of 148), against **0%** on 08-31.

⚠ **Do not credit 640 for that.** `build-site-planner` carries no per-section-subject rule
`[MEASURED 2026-09-04]`, so the subjects are arriving from somewhere else (443's Stage A columns are
the likely route). **Check the plan row for the page you care about rather than assuming.**
⚠ **But do not establish that from the writer's gate** — see §2a: run against the planner, that
query's own control cannot fire, so its `f` is uninformative rather than evidence.

⚠ **The served outcome of that run is not readable** — copyonline.co.uk's recorded page URL
(`/checklists.html`, read from `pages.url`, not composed) returns 404, and so does an invented
sibling, so the site is not published. **The proof above is at the prompt, not at the artefact.**
The artefact-level proof still wants a published multi-section page.

---

## 6. Where the ask really stands: THREE layers, and this lane owns only the top

1. **Can a figure survive regeneration?** ✅ Done, reviewed, live, **proved at the artefact**.
2. **Does anything compose an article out of illustrated sections?** grip-styles was the first, by
   hand, and is reverted. `editorial_design_uplift` owns this.
3. **Are articles even IN THE PLAN?** ⚠ **Mostly not — the floor.** `[MEASURED 2026-09-03]` on the
   33 sites with a current plan: **tool 83%, blog-post 85%, guide 74%** have NO `site_plan_sections`
   row, against **landing 2%**. **No plan row → `planSectionOrder` returns nil → binding disabled.**
   ⚠ **Re-measure before quoting — the subject figures moved 0%→84.5% in four days, so this
   population is moving too.**

**Mechanism, read first-hand:** `create_blog_posts_action.go:212` — the article layout triple is a
**fallback**, and the action writes `pages.sections` (**tier 3, the cache**) and never
`site_plan_sections` (**tier 1, the authority**). Writers of the authority `[MEASURED 2026-09-03]`:
**2** Go call sites (neither on the article path), 2 config rows that only read, and **15 operator
SQL files** — the third is the trap, because backfilling by hand fixes the pages that exist and
nothing about the route.

---

## 7. What I would do next, in order

1. **The grip-styles retry is UNBLOCKED and it is `dartsonline_traffic`'s to run, not ours.** They
   recorded it as four steps gated on the artefact query (§5), with our acceptance test as the
   grading criterion, and **committed to reporting the run either way, including on failure**.
   Stage 2 is skippable — the five illustrations are still active assets. **Do not fire it at their
   page yourself**; the gate is now open, so the right action is to tell them if they have not
   noticed, and read the report.
2. **The re-render pre-registration is LIVE AGAIN** (it was void only while grip-styles had no
   rows). The prediction: on a page whose plan and stored slot lists agree, a re-resolving re-render
   **BINDS per-section**; **the disconfirming result is all five sections showing ONE image.**
   Pre-flight query in the RUNBOOK. ⚠ Grade on the run's `resolved_data`, never the served bytes —
   an assemble-only re-render produces identical bytes whether the binding engaged or did nothing.
3. **`apis.uk/index` is now the strongest candidate in the estate and nobody has touched it.**
   `[MEASURED 2026-09-04]` six `scope='section'` illustration rows since 2026-09-02 16:47Z, and its
   `page_components` still read **2026-08-24** — armed, never exercised. It is six repeats of one
   component, which is both the discriminating shape for the binding **and** for the subject fix.
   Its lane's call; offer, do not fire.
4. **Do not build the Phase-4 detector** (`check_unrendered_section_imagery`). The PLAN's "discovery
   has no driver" blocker is stale, the RUNBOOK's hand query does the job, and `bugfix_114` has
   offered a section-scope arm on `check_unrendered_page_imagery` (IMG-077).
5. **Leave phase 3 (article planning) where it is** — nobody owns it, and it should not be closed by
   a backfill.

⚠ **Fleet supply moved under us** `[MEASURED 2026-09-04]`: section-scope illustration/infographic
rows are now **apis.uk 6, vonc.com 2, and five sites with 1 each** (`fundamentallyai`, `idea.uk`,
`mortgagecalculator`, `gamedesign.uk`, `copyonline.co.uk` — the last two dated today).
**gamedesign.uk fell from 4 rows to 1**, so yesterday's note about its three-at-one-ordinal shape no
longer describes it. Re-run the census before quoting any of this.

---

## 8. Traps this lane paid for — read before trusting a number

- **A test that cannot fail is not a pass.** §5a: N-distinct-prompts on a page of N *different*
  components is guaranteed; and "does this subject appear in this prompt" returns 6/6 for every
  subject because the outline is shared. **Name the discriminating population before you report.**
- **A step's collected value is its RESULT, not its prompt.** `llm_call_log.prompt_rendered` holds
  what was actually sent — reach for it first, and **scope by `orchestration_id`, never by a time
  window you chose from when you expected the work to run** (that cost the peer lane a wrong
  conclusion: their run was 69 minutes outside the window).
- **Read the recorded `pages.url`, never compose one.** Did it again today on copyonline; the
  composed URL and the recorded one both 404, and only the invented-sibling control showed the site
  is simply unpublished.
- **A served-page census on `data-component` UNDERCOUNTS** — `generic-text-block` emits none. Count
  `class="section` families.
- **`alt` text is not evidence of what an image shows** — LANDMINE. Open the image.
- **A count of a population says nothing about whether it is GROWING** — "9 of 442" read as "nothing
  selects it"; the refutation was `created_at`, in the table already queried. Subjects went 0% →
  84.5% in four days.
- **A `LIMIT` counted section ROWS while my claim counted PAGES** — one page contributed 6 of 12.
- **`updated_at` moved ≠ a re-render happened ≠ the resolver was asked** — three events, one word.
- **Filed ≠ ran** — an item can exist, be quoted as evidence, and have FAILED with `result = {}`.
- **A Go-only grep is not a fleet-wide census** — config SQL and operator SQL are two more writer
  populations.
- **I quoted a Go COMMENT as live config twice** — the re-render reason list (five, not two) and the
  `Subject` field's "the v5 prompt renders it". **Cite the row.**
- **Do not gate on "is migration NNN applied?"** — `_HOLD` files are applied by hand, and **129
  migration numbers name two or more unrelated files** (`[MEASURED 2026-09-03]`, ~1 in 6). Ask the
  running agent for the capability. LANDMINE.
- **The author of a query cannot catch its own encoding error.** Four errors across two lanes in one
  day, **all four caught by the other lane re-running it**. Publish the exact command next to any
  number; when figures differ, chase the predicates rather than deciding whose error it was — twice
  both of us were wrong in different directions.
- **`git stash` is forbidden; commit by pathspec — and build the pathspec from `git status`, not
  memory.** I broke HEAD for eleven minutes naming two of three callers.
