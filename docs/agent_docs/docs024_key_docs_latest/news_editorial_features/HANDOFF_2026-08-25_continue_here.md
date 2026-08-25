# HANDOFF — news editorial + editorial design uplift, 2026-08-25. START HERE.

Supersedes `HANDOFF_2026-08-24_continue_here.md`, which in turn superseded the
08-21 one. **Neither predecessor is dead, and this file does not restate them:**

- **08-21 §3 (the recipe, proven three times) and §9 (the ten traps) still
  govern.** Read them there before shipping a feature page.
- **08-24 §8 is the full evidence trail for the instance-scope acceptance** —
  written yesterday by this thread, and §8.5/§8.6 are the operative parts today.
  This file points at it rather than copying it.

Everything below is measured 2026-08-25 unless marked otherwise. Fleet at chassis
`v1.0.1337`, both pods started 2026-08-25 09:27Z. **Nothing in this lane is
pending a roll** — its work is docs, DB config (live on apply) and a repo-local
harness. P1 will still be this lane's first platform code.

---

## 1. ~~THE ONE THING TO DO FIRST — three SQL scripts are written and UNRUN~~ **DONE 2026-08-25 16:04Z**

> **✅ RUN AND VERIFIED — nothing here is outstanding. Do NOT re-run any of it.**
> Both pages serve `c-evidence-timeseries`: rh **94,348 B**, do **92,871 B** (measured at the
> served page, stylesheet control OK, `empty_id=0`, one occurrence each). Both rows re-locked
> `permanent` / `news_editorial_features-lane` with their **original** `locked_at` intact. Both
> `lock_blocked_change` items closed `complete` / `disposition='accepted'`. RFC_032 §9a updated
> (`evidence-timeseries` 3 → 1 unconverted; the survivor is oufe's, another lane's page).
> The `pending_sql_instance_scope_acceptance/` directory has been **deleted**, per its own README.
>
> **⚠ Two corrections this run produced, both of which outlive it:**
> 1. **The scripts as written could neither unlock nor re-lock.** Neither touched `locked_at`,
>    which is the only column `AgentWritableSQLFor` reads. The first run delivered nothing while
>    both work items reported `complete` / `success: true`. The re-lock carried the mirror defect,
>    which against a corrected unlock would have left both flagship rows agent-writable while
>    displaying as `permanent`. **The two defects cancel, so no dry run could have separated
>    them.** Full account: NOTES 2026-08-25, `WRONG_CALLS.md`, `LANDMINES.md`
>    (footprint `page_components.locked_at`).
> 2. **§2's `[PREDICTED]` byte figures were wrong by exactly 1 byte on BOTH pages** (predicted
>    −2/−11, measured −3/−12). The template change is provably id-only and the id occurs once, so
>    the extra byte is template-derived output the RUNBOOK §11 harness does not model. **P1's
>    baseline is rh 94,348 / do 92,871** — see §2's amended warning.

**This session cannot write to the DB** (the permission classifier blocks it;
reads are fine). The scripts are written, idempotent, and waiting on the owner.
**Verified unrun at 2026-08-25: both rows still `lock_type='permanent'`,
`converted=false`, and no new dispatch items exist.**

**They are IN THE REPO**, at
`news_editorial_features/pending_sql_instance_scope_acceptance/` — moved there
2026-08-25 precisely because the first draft of this handoff pointed at a session
scratchpad, and a scratchpad dies with its session, so that instruction had a
shelf life of hours. That directory's `README.md` carries the run order, the
unlocked-window warning and the publish-lag rule; this section is the summary.

```
cd docs/agent_docs/docs024_key_docs_latest/news_editorial_features/pending_sql_instance_scope_acceptance
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

! $PSQL -f - < A_unlock_and_dispatch.sql   # unlocks BOTH rows + re-dispatches
./verify.sh                                # re-run until ALL PASS
! $PSQL -f - < B_relock.sql                # restore the permanent lock
./verify.sh                                # confirm the lock came back
! $PSQL -f - < C_close_lock_items.sql      # both items -> complete
```

They are deliberately **not** in `docs/agent_docs/sql_for_agents/` — they are
operational one-offs against two named rows, and that directory is what the
migration runner sweeps. **Delete the whole directory once C has run and the ids
verify served**, in the commit that records the result: it is a pending-action
folder, not a record, and a stale one reads as outstanding work for ever.

**Between A and B the two rows are UNLOCKED**, exposed to every sweep on the
estate — this lane was already hit once by an improvement-loop misfire (08-22).
Do not leave that gap open longer than the verification takes, and do not start
the sequence if you cannot finish it.

### What the change is, in one paragraph

The **283 / RFC_034** lane converted the shared `evidence-timeseries` template on
08-23 12:33:33Z so its section id comes from `{{.InstanceID}}` instead of
`{{.ComponentID}}`. `component-template-fixer` fanned the re-render out; both our
instances refused it (our `permanent` lock) and filed `lock_blocked_change` items.
**Owner ruled ACCEPT on 2026-08-25.** The lane had recommended honouring — the id
is inert (class-based CSS, one reference per page, nothing selects on it) — and
the owner chose consistency with the fleet-wide convention. Executed as: unlock →
re-dispatch → verify at the served page → re-lock → close both items `complete`.

**The dry run is already done and passed** (RUNBOOK §11 is the recipe):

| page | id before | id after | delta |
|---|---|---|---|
| robot-demand-step-change | `evidence-timeseries-ifr` | `c-evidence-timeseries` | −2 B |
| darts-calendar-density | `evidence-timeseries-pdc-calendar` | `c-evidence-timeseries` | −11 B |

`InstanceToken(function, occurrence)` returns `"c-" + function` for occurrence ≤ 0
(`component_instance_scope.go:102-115`) — **`c-evidence-timeseries`**, not
`evidence-timeseries-0`. Both pages get the same token; different pages, nothing
collides.

### The trap that will bite you here

**A `complete` item and correct stored bytes are BOTH consistent with the page
still serving the old version.** Measured by the 283 lane 2026-08-24: three
repairs `complete`, stored bytes correct, one page served the old version for
hours. So verify at the **served page**, never at the row — `verify.sh` enforces
the order and fails loudly. If steps 1 and 2 pass and step 3 still shows the old
id, that is a **publish lag, not a failed edit**: wait and re-check. Do NOT
re-dispatch — a second delivery on top of a pending publish races two writes at an
unlocked flagship row. Full detail: 08-24 §8.5.

### When it lands

~~Tell **`bugs_open/283`** the resulting served ids~~ — **DONE, and the pointer was
already stale: 283 CLOSED on 2026-08-25 (`291607d40`) and now lives in
`bugs_closed/`.** The ids were recorded where the count actually lives — a dated
UPDATE block in **RFC_032 §9a**
(`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_032_three_render_context_builders_disagree_about_what_an_instance_is.md`),
moving `evidence-timeseries` 3 → 1 unconverted and the fleet residual 48 → 46, both
marked `as of 2026-08-25` and flagged as re-derived by subtraction rather than
re-run. And
**write the close-out as CONSISTENCY, not repair** — our pages carry one ev-ts
each, so they were never in the population where a literal id can actually bite.
A close-out implying a repair makes the next reader think a defect was there.

## 2. P1 (035 read path) — the seams are re-read, constraints written down

The 08-24 handoff said "re-read the seams FIRST". **Done 2026-08-25; the results
are 08-24 §8.6.** Do not re-derive them. In short:

1. The **AST test** (`render_seam_one_spelling_test.go`) forbids a second
   exported `RenderTemplate*` symbol — **do not name the walk
   `RenderTemplateComposite`**. Only `RenderTemplate` may call
   `executeGoTemplate`. A walk calling `RenderTemplate` per node needs no test
   change.
2. `RenderTemplate(templateStr, ctx, logger) → (html, missing, inURLAttr, err)`
   (`component_library.go:1060`). Discarding the two reports must be **visible**
   in the diff — hiding it is how `bugs_open/238` shipped five `<img src="">` to
   a live homepage.
3. `deriveRenderMode` (`store_generated_component_action.go:2024`, called :732
   INSERT / :822 UPDATE) must test `slots` **BEFORE** its llm-field loop —
   D3's own example declares `slots` *and* an llm field, and the current order
   would derive `agent` and never route to the composition build.

**Two new hazards were added to `035` §6 this session — read them before writing
the walk:**

- **§6.9 — one `RenderContext` reused across the walk forges RFC_046
  provenance.** The digest is MUTATED onto the context, not returned
  (`component_library.go:1081`), and the empty-template branch returns early
  *without setting it* — so on a reused context an empty-template child inherits
  its **predecessor's** stamp. **Today's code is safe** because all FOUR live
  readers render exactly one template per context — a property held by convention
  across four files in four lanes, which nothing enforces. The walk is the first
  code that would break it. Carries a falsifier; a test with only non-empty
  children passes either way. **The primitive-level fix belongs to the 357 lane
  and is architecture-scope — do NOT bolt it into the P1 commit.**
- **§6.1 addendum — the completeness floor.** `save_page_sections` carries a
  floor (`save_sections_prune_floor.go`) refusing the whole save when incoming
  rows are too small a fraction of the existing rows AND of `pages.sections`'
  planned count. Composition moves both denominators (parents+children in
  `page_components`, parents only in `pages.sections`). **Moot for P1/P2**
  (owned/locked rows are not agent-writable, the delete never runs); it is P5's
  question, and the answer is *"what is the denominator on a composed page?"*,
  never "lower the ratio".

**⚠ P1's byte baseline moved when §1 landed — and the prediction was WRONG.**
The acceptance test is "served page byte-equivalent". Baseline 08-24 was rh 94,351 /
do 92,883. Predicted after acceptance rh 94,349 / do 92,872; **measured 2026-08-25
after the change actually landed: rh 94,348 / do 92,871.** Use those.

**The miss is 1 byte on BOTH pages, and it is not noise — it is a defect in the
RUNBOOK §11 harness that P1 depends on.** The id delta really is −2 / −11
(the old rendered id was exactly the slot_name, proven against the surviving
unconverted third instance), the template change is provably id-only
(`component_versions` v1 vs live diffs to one line), and the new id occurs once per
component. Yet each page lost **3 / 12**. Both lost the *same* byte: the two pages'
content difference is preserved exactly (569 B before and after), which a
content-dependent difference could not do. So one byte of **template-derived** output
leaves each component on re-render and the harness does not model it.
`[UNEXPLAINED]` — measured, not guessed; the open candidate is a renderer difference
between the 08-20 stored bytes and today's chassis. **A harness that under-predicts by
1 byte per re-rendered instance will fail P1's byte-equivalence test for a reason that
has nothing to do with P1.** This is §6.3's trap one step on: showing the harness
reproduces the CURRENT stored bytes does not establish that its PREDICTION of the
post-change bytes is exact. Fix or characterise the harness before leaning on it.

Rest of the P1 brief is unchanged: walk in both render paths + `deriveRenderMode`
third value + `check_render_mode` routing arm + register entry, ONE council-gated
commit; recompose one live insights page; then rewrite one prose child and prove
every sibling row byte-identical. Consumers to tell (07-29 §3): component-library
lane, 283 lane, 238/268 carry lane, inline_guide_imagery lane. **Until P1 ships:
no `composite` rows, no `parent_instance_id` values, anywhere** (035 §9 r1).

## 3. Open work

| item | state |
|---|---|
| ~~**The three SQL scripts**~~ | **DONE 2026-08-25 16:04Z — run, verified at the served page, re-locked, items closed, directory deleted. §1.** Two defects found in the scripts themselves and one in the byte prediction; see §1 and §2. |
| **P1** (035 read path) | Seams re-read, constraints in 08-24 §8.6, two new hazards in 035 §6. Next after §1. |
| **Our own feature pages are failing to rerender** | 3 × `page_rerender` failed on misdirected CTA: `electric-vs-pneumatic-economics` (×2, 08-22 and 08-24), `darts-calendar-density` (08-22). **Not yet investigated** — found while triaging, not chased. |
| `head_essentials_missing: skip_link` | detected on all four feature pages + `insights-index`. |
| `dead_internal_link_live` / `unbuilt_internal_link` | both → `/gripper-technology-comparison.html`, cited from `learning-center` and our tool guide page. |
| `owned_page_review` on `electric-vs-pneumatic-economics` | `needs_human_review` since 08-21, untouched. |
| **A2** — `compute_component_quality` | still never run. |
| **Phase B furniture** | unchanged from 08-24 §3. |
| **Phase E1 / E2** | E1 (collect dated cited events) live; **E2 is PLANNED, NOT BUILT**. See §4 — the 381 lane needed telling. |
| **Rollout site 3, cobot #4, `published_page_id` reader, bugs 349/198/296** | unchanged from 08-24 §3. |
| `claims_unverified` on robot-hands (08-24 19:05Z) | **NOT ours** — names `gripper-detail` + two tool guides, the catalogue lane's pages. Scoped out deliberately. |

## 4. Coordination — three lanes, all answered, nothing owed

- **283 / RFC_034** — told we accepted. **They took a one-row finding of ours and
  made it much bigger:** we reported that the third `evidence-timeseries`
  placement (oufe.com) received no delivery attempt and therefore emitted **no
  signal at all**. They verified it, re-derived the census properly, corrected
  RFC_032 §9a (which recorded "(3/3/3)" — a count of the TEMPLATE that reads as
  placements), and filed the landmine. **Then they corrected their own number to
  us:** 48 of 437 unconverted is **consistency debt, not damage** — 48 → 8 on a
  multi-instance pair → 1 genuinely duplicated id → **0 reaching a visitor**, and
  that one page redirects to a different site row. We had already repeated the
  bare "48" into the 08-24 handoff; it is corrected there. Awaiting our served
  ids.
- **357 / RFC_046** — we found that their stamp's "no token beats a wrong token"
  reasoning has an unstated precondition (a fresh context). They agreed it is a
  gap in the reasoning, contributed the fourth reader our census had missed, and
  ruled the primitive fix architecture-scope and theirs to route. They also gave
  us the completeness floor (§2). Nothing owed; they may report a gate verdict.
- **381** — `period-calendar` boundary confirmed: ours is fact-fed (dated events,
  per-observation citations, fails closed without an evidence base), theirs is
  authored recurring-cycle guidance with no dates or citations. They may write
  "agreed 2026-08-25". We corrected two things in their header: our E2 is
  **planned, not built**, and is specified in the `mechanism-flow` idiom.

## 5. Session mechanics

Unchanged from 08-24 §5 (DB writes blocked → scratchpad SQL protocol; never
launder a blocked write through a peer; `content_components.updated_at` cannot
date or attribute a change; same-file passengers are real). **Confirmed again
today:** this session's `WRONG_CALLS.md` append was committed by another lane as
a passenger on `1bdcc929a` before it could commit its own — nothing was lost, and
`git show HEAD:<file> | grep` is how you confirm that rather than assuming it.

## 6. New traps, this session

1. **`orchestration_states` is a ~25-hour rolling window.** The runs behind the
   08-23 lock refusals were already gone. Work items survive; orchestrations do
   not. Trace from the item's `result` blob (it carries the full retry payload
   and headers) or from `component_versions`, not from `orchestration_states`.
2. **A URL assembled from `pages.name` returns a REAL 404 page**, from the right
   domain, that reads exactly like a finding. `pages.url` is a column, not a
   convention — oufe's page is `/cases/thames-water.html`, not
   `/thames-water.html`. Cost me three minutes and nearly a false bug report;
   logged in `WRONG_CALLS.md` 2026-08-25.
3. **A render-harness diff proves nothing until the harness is shown to
   reproduce the CURRENT stored bytes.** The control is the load-bearing half and
   it is the step that gets skipped — RUNBOOK §11 spells it out.
4. **Line numbers in handoffs go stale.** 08-24 §1 cited `deriveRenderMode` at
   `:561/:639`; it is `store_generated_component_action.go:2024`, called from
   :732/:822. Corrected in place. Cite the **symbol**; a line number is a hint.

## 7. What NOT to do

- Everything in 08-21 §10 and 08-24 §7 still stands.
- Do not hand-write `rendered_html` or hand-patch an element id — framework only.
- Do not leave the two rows unlocked past verification, and do not begin §1 if you
  cannot finish it.
- Do not re-dispatch on a stale served page — that is a publish lag (§1).
- Do not bolt the RFC_046 clear-on-entry fix into the P1 commit — 357's to route.
- Do not answer the completeness floor by lowering `prune_floor_ratio`.
- Do not treat the `claims_unverified` item as ours.

## 8. This session's commits

`b058c0504` lock trace + ruling + records · `a1680fc05` §8.5 verification order ·
`a3bf4ee3f` §8.5 correction (48 → 8 → 1 → 0) · `e273fcb2f` 035 hazard 9 ·
`62db696ca` §8.6 seam constraints + drifted line numbers · `c64f3d116` 035 four
readers + completeness floor. `WRONG_CALLS.md` rode `1bdcc929a`.
