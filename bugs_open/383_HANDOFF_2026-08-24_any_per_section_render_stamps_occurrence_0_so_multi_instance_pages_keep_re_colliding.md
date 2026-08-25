# 383 — Any per-section render stamps occurrence 0, so a repaired multi-instance page re-collides within hours

**Filed** 2026-08-24 · **Status: FIX COMMITTED (`364e80b7f`), INERT until the next chassis roll — so this stays OPEN**
· lane `docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/`
· council `3fd0d026-8966-44c6-b0d8-bd8c0dfba187` **APPROVED round 1**, 2026-08-24 16:04 UTC (§9)

> **WHY THIS IS A NEW NUMBER AND NOT `283`.** `bugs_closed/283` is a DIFFERENT defect —
> interactive components could not be reused on one page because their element ids were
> **literal**. That was fixed and went live 2026-08-22 and stays closed. This bug is the
> defect that fix *exposed*: the ids are now derived from an occurrence, and two render
> paths supply the wrong occurrence. Until now the only open record of it lived inside
> `architecture_review/RFC_032…md` §9d/§10 and a PLAN file, so a reader grepping
> `bugs_open/` for duplicate element ids found **nothing**. Filed at the owner's direction
> 2026-08-24.

## 1. What the thing is, in plain terms

A component that can appear more than once on a page (a text block, an FAQ, a mechanism
flow) must give each of its copies **different** HTML element ids, or a script or a
stylesheet aimed at "the FAQ" hits whichever copy the browser found first. The estate does
that by putting a per-copy token in every id: `id="c-faq"`, `id="c-faq-2"`, `id="c-faq-3"`.

The token is `InstanceToken(function, occurrence)`. The **occurrence** is simply *how many
sections with the same function come before this one on the page*, counted in position
order. One function computes it for everybody: `InstanceCounter.Next`
(`platform/orchestration/actions/component_instance_scope.go:140`), which lowercases and
trims the function name so `"FAQ "` and `"faq"` count as the same thing.

## 2. The defect

Two render paths only ever see **one section at a time**, so they cannot count anything —
and both simply passed **0**:

| path | file:line (before the fix) | when it runs |
|---|---|---|
| `RenderComponentAction` | `v3_site_actions.go:2404` | every page **build** and every `content_rewrite` |
| `applyContentEdit` | `section_editor_actions.go:1104` | a section **edit** on a live page |
| `applyComponentSwap` | `section_editor_actions.go:1275` | a component **swap** on a live page |

So every copy on the page got occurrence 0, i.e. the **same** token — and the ids collide
again. The trigger is not "a rebuild": it is **any per-section render**, which is why the
damage kept coming back.

## 3. Evidence (all 2026-08-24, this session, at the artefact unless stated)

```bash
curl -s https://gaswholesalers.com/pricing-transparency.html \
  | grep -o 'id="c-generic-text-block[^"]*"' | sort | uniq -c
#   2 id="c-generic-text-block"          <-- two sections, one id
```
Same shape on `vetcomparison.uk/how-it-works.html`.

- **3** multi-instance pages carry a duplicated `c-` token in stored `rendered_html` (live DB).
- **30** multi-instance `(page, function)` pairs exist; **16 of 30** repeat a `slot_name`.
- **3 of the 12** pages repaired by the 2026-08-23 rerender queue were **re-collided within
  hours** by an unrelated lane's `content_rewrite` backfill. That is the defining symptom:
  *the repair works and does not hold.*
- `apis.uk/index.html` still holds `build_status='needs_rebuild'`, so it rebuilds and
  re-collides **on its own** — a repro that regenerates itself, no setup.

⚠ **A corpus count stays GREEN while this happens.** The same token twice is still two
`c-`-prefixed tokens, so any check that counts *prefixed* ids sees nothing wrong. Count
**distinct** tokens per page, or fetch the page.

## 4. Root cause

Not "the paths disagree about the rule" — they use the same rule. They were given the
wrong **input**, and the constant was documented as safe on a measurement that was never
true of these templates: the binder's comment claimed the component "appears once per
page, which is every interactive component on every live page today (measured
2026-08-15)". That measurement was about `getElementById` tool components. It was never
true of the RFC_032 §8 repeatable templates — they are precisely the ones that repeat, up
to ×6.

## 5. The fix, committed `364e80b7f` (Go — INERT until the roll)

Both paths now feed the canonical rule its real input; nothing invents a second rule.

- **Build path**: it is always a loop iteration. Loop expansion already injects
  `loop_item_index` / `loop_name` into each substep's own config and parks every item in
  `CollectedData`, so the occurrence is counted from the sections **already rendered in
  this pass**. That is `InstanceCounter`'s arithmetic over the same list, one section at a
  time — and it is therefore right on a page's **first build**, when no `page_components`
  rows exist to count.
- **Editor path**: no loop, but it holds the stored row, so it counts same-function
  predecessors in the DB, position-exact, with a `(position, id)` tie arm — matched by
  tightening `loadStoredSections`' `ORDER BY`, because that query *is* the canonical walk.
- **Everything else**: occurrence 0, exactly as before. The derivation is an
  **input-improver, never a gate** — it cannot fail a render.
- New: `component_instance_occurrence.go`, `datahelpers/loop_keys.go`, two test files.

No migration and no workflow-config key, deliberately: the fix is live at the image roll
with no activation step.

## 6. How to verify — at the artefact, never at a status

**Before believing any of this is fixed, confirm the roll actually shipped it:**
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 364e80b7f <the stamp>
```

Then:
```bash
# 1. the two standing repros — expect the bare token x1 and -2 x1
curl -s https://gaswholesalers.com/pricing-transparency.html \
  | grep -o 'id="c-generic-text-block[^"]*"' | sort | uniq -c
# 2. the first-build arm, free: apis.uk/index.html rebuilds itself -> 6 DISTINCT tokens
# 3. THE ACTUAL DEFECT — STICKINESS. After repair, await or trigger ONE content_rewrite on
#    a repaired page and re-count at the SERVED page. This is the step that flapped on
#    2026-08-23 17:41, and it is the only check that distinguishes "the fix works" from
#    "the rerender queue happened to fix it".
# 4. the editor arm: one content_edit on the SECOND instance of a multi-instance page,
#    then re-count -> the -2 token must survive.
```

## 7. What is NOT fixed by the commit

- **The stored damage.** A `page_rerender` goes through the canonical walk and repairs it
  correctly *even on today's code* — that is how 9 of 12 were fixed on 2026-08-23. Repair
  items are filed separately; the commit is what makes the repair **durable**.
- **A non-loop `render_component` caller** would still bind 0. There are **zero** today —
  exactly **one** active agent definition (`page-content-writer`) runs the action, as of
  **2026-08-24**, and that count goes stale *by addition*. A `LANDMINES.md` entry warns
  whoever adds the second one.
- **Editor errs-high edge case**: the canonical walk does not advance its counter for a
  *carried* section (invalid template), but the SQL counts every stored same-function row.
  Where an earlier same-function section is unresolvable the editor's count comes out one
  high — swapping which partner it collides with, not adding a collision, and no worse in
  collision count than the constant 0 it replaces. Documented on `storedPredecessorCount`.
- **Empty element ids** (`id=""`) are a *different* class with three distinct causes and
  belong to RFC_032 Half B (committed `120131549`, council `661bcf00`) — not here.

## 8. Related

- `bugs_closed/283` — the literal-id defect this one was exposed by. Resolve by SLUG.
- `architecture_review/RFC_032…md` §9c (the ruled fix shape), §9d (the correction that the
  trigger is ANY per-section render), §10 (owner ruling), §10a (the interim exposure record).
- `PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md` — the originating plan.
  **Its Half A design is superseded**; the correction is recorded in that file.

---

## 9. Council: APPROVED round 1 (`3fd0d026`, 2026-08-24 16:04 UTC) — verdict read, four advisory objections and what was done

*"approved with 4 advisory objection(s) — none high-severity"*. 7 seats approved, 4 objected
advisorily, 6 abstained. The commit carries `Council-Submitted:`, which `098` resolves to
credited automatically now the correlation is approved — **no amend** (forward-only).

| seat | objection | disposition |
|---|---|---|
| **guardian** (med) | *"confirm `loadStoredSections` has no other callers whose behaviour depends on the pre-existing tie-break order"* | **RUN, and it is decisive: exactly ONE production caller** — `rerender_page_sections_action.go:262`, the canonical walk this change fixes. Every other hit of that symbol is a comment. `grep -rn loadStoredSections --include=*.go . \| grep -v _test.go` |
| **editquality** (med) | *"no edit entry modifies `loop_expansion_handler.go` or `loop_actions.go`, yet edit 1's rationale and mutations (E)/(F) assume the rewiring is done"* | **Correct, and it is a defect in my SUBMISSION, not in the code.** The three rewires are in the commit; I hit the 8-edit cap and folded them into edit 1's prose instead of listing them. The seat caught a real gap between what I described and what I enumerated. Nothing to change in the tree; recorded here so the trail is honest. |
| **bug_historian** (med) | *"the root constant-0 fallback stays generic; a LANDMINES entry is documentation, not a guard"* — cites 016b §9 item 7, the platform's most-repeated shape | **Fair, and half-answered.** The landmine is now filed and verifier-dispatched (`8cb37bfb`), which is what this seat and the architecture seat both asked for. The seat's deeper point stands and is not closed: this patches the two known callers, and a future non-loop caller inherits occurrence 0 silently. That is §7 above, and it is the right thing for a human to weigh. |
| **debug_historian** (med) | *"'live at the roll' names no post-roll pod verification; a same-tag rebuild ships a stale binary"* | **Already in §6 above** (build-provenance line + `git merge-base --is-ancestor`), but absent from the submission — so the objection is right about what I submitted. |
| **architecture** (low) | wants the binder retirement as a **tracked item**, not only a file-header comment — *"so it doesn't rot the way RFC_032's ComponentID binding did (two deferrals already)"* | Tracked here as §10. That precedent is exact: the binding this very commit deletes had been deferred twice. |
| **reuse_agent** (missing) | *did it reuse the existing loop-config reader rather than write new parsing?* | It reuses `datahelpers.GetIntField` (the helper `loop_actions.go:324` uses) via `placementInt`, and `datahelpers.LoopItemKey`, which this change created **by single-sourcing three existing literals**. There is no existing item-shape reader to extend. |

## 10. ~~TRACKED FOLLOW-UP~~ — **DONE the same day**: `BindSingleSectionInstanceToken` is RETIRED

> **CLOSED 2026-08-24.** The Half B lane reported its files clean, so the deferral's only
> condition was met within hours and the retirement was taken immediately rather than tracked.
> Deleted from `component_instance_scope.go` (a RETIRED note stands in its place, naming the
> false licence it shipped under); the seam regex drops it; the one remaining caller,
> `component_instance_scope_test.go:126`, is retargeted onto `DeriveAndBindInstanceToken` with an
> EMPTY placement — which is exactly the no-context fallback, so that test still asserts the
> property it always asserted. **A census of "single-section binders" now returns ONE.**
> Also fixed in the same commit: `pattern-check.py`'s finding text still told authors to *call*
> the retired function — a remediation pointing at a deleted symbol, which is the stale-advice
> shape the architecture seat's "don't let it rot in a file header" objection was about.

### The original entry, kept because the deferral is the lesson

`component_instance_scope.go` still defines it; **no production code calls it** after `364e80b7f`.
It was left in place only because that file was dirty in the concurrent Half B lane and moving a
function out from under it would repeat the same-file passenger that drew the guardian veto on
`e8c7414c`. **Do this once that lane is clean:** delete the function, drop
`BindSingleSectionInstanceToken` from `scripts/pattern-check.py`'s `INSTANCE_BIND_SEAM_RE`, and
update `component_instance_scope_test.go:126`, which is the only remaining caller. Until then a
census of "single-section binders" honestly returns two.

## 11. Repair items FILED 2026-08-24 (owner decision) — and the trap that nearly made them useless

Three `page_rerender` items filed for the three pages currently carrying a duplicated token
(`SQL_2026-08-24_repair_duplicated_instance_tokens.sql`; `INSERT 0 3`, `VERIFY: PASS`).

⚠ **A rerender repairs this on TODAY's code** — it goes through the canonical walk, which is how
9 of 12 pages were repaired on 2026-08-23. **What the roll adds is that the repair HOLDS.**

⚠ **THE `spec.reason` IS LOAD-BEARING AND MY FIRST DRAFT WOULD HAVE DONE NOTHING.**
`page-rerender`'s `check_rerender_mode` is a conditional over an allow-list of exactly five
reasons. Anything else — including an invented one, and including none — takes `else_step:
render_page`, which is `rerender_single_page`: *"simple concatenation, no template re-rendering"*.
It **re-ships the stored bytes** and completes successfully. I had invented
`reason: instance_scope_383`. `template_changed` is the correct one of the five and is the only
one with no Go branch keyed on it (`cta_links_stale` triggers a CTA recompute with its own clobber
landmine; `section_data_resolved`/`image_landed` get scoped to one component when a component_id
is present). The applied file carries a control asserting all three items route to the sections
branch, so this cannot recur silently.

---

## 12. REPAIR RESULT 2026-08-24 — 3 of 3 repaired in the DATABASE, 2 of 3 at the SERVED page

All three items completed (`build-dispatch-loop`, 17:58–18:33; backlog drained 225 → 21).
Verified at the artefact rather than at the status, and the two do not agree:

| page | stored `page_components` | **served page** |
|---|---|---|
| vetcomparison.uk `/how-it-works.html` | `c-generic-text-block` ×1, `-2` ×1 ✅ | ✅ repaired |
| gaswholesalers.com `/wholesale-pricing-explained.html` | ✅ | ✅ repaired |
| gaswholesalers.com `/pricing-transparency.html` | ✅ | ❌ **still serves `c-generic-text-block` ×2** |

**This is exactly why `complete` is not proof.** The work item says complete, the stored bytes are
right, and the page a visitor gets is still wrong.

**What is ruled out, measured:**
- **Not a bad repair.** Raw rows for that page: position 2 → `c-generic-text-block`, position 4 →
  `c-generic-text-block-2`, every row `build_status='deployed'`, all written 18:32:55. The
  canonical walk did its job.
- **Not an edge cache.** `cf-cache-status: DYNAMIC` on both pages, and a cache-busted query-string
  fetch returns the same duplicated ids.
- **Not "the deploy hasn't happened yet".** The served file's `last-modified` is **18:38:21**,
  i.e. *after* the 18:32:55 repair — so a deploy ran and shipped pre-repair bytes.

**The open question, and the lead.** Both gaswholesalers pages were also hit by **reason-less**
`page_rerender` items (`created_by='rerender-pages'`, created 17:46:44) that completed at
**18:51:03** and **18:51:38**. A reason-less item is assemble-only (`rerender_single_page`,
"simple concatenation, no template re-rendering") — see §11. `pages.updated_at` moved to 18:51:02
for the broken page, but the served file's `last-modified` stayed at 18:38:21, so **that assemble
updated the database row and did not rewrite the served file.** Its sibling, assembled 35 seconds
later, serves correctly.

**This is NOT the occurrence-0 defect** — that one is fixed at source by `364e80b7f` and the
stored bytes prove the repair worked. It is a delivery/assembly gap between correct stored HTML
and the file on the bucket, on one page of two treated identically. **Do not fold it into 383.**
Next step for whoever picks it up: establish whether the 18:38 deploy read a stale snapshot or
the 18:51 assemble silently skipped the write, and check the `single-page deploy bypasses stalled
queue` route as the direct repair (a `page_rerender` **with** `reason: template_changed` on this
one page would both re-render and re-deploy).

---

## 13. POST-ROLL VERIFICATION 2026-08-25 — the fix is LIVE and WORKING at the artefact; one non-colliding discrepancy is open

**The deploy is proven, not assumed.** Chassis `v1.0.1337`, both replicas, built from
`4c996e1b5`. `git merge-base --is-ancestor` says **SHIPPED** for `364e80b7f` (Half A),
`9ba3293e7` (binder retirement) and `120131549` (Half B). **Control**: today's HEAD is NOT an
ancestor of the build, so the check can come out false.

### The served pages — 3 of 3 now correct

| page | served tokens |
|---|---|
| gaswholesalers.com `/pricing-transparency.html` | `c-generic-text-block` ×1, `-2` ×1 ✅ |
| gaswholesalers.com `/wholesale-pricing-explained.html` | ✅ |
| vetcomparison.uk `/how-it-works.html` | ✅ |

**§12's stored-vs-served divergence is CLOSED by observation** — `pricing-transparency.html` now
serves the repaired bytes. It was a delivery lag, not a defect in the repair. Nothing was done to
it; recording that it resolved itself so nobody hunts a bug that is not there.

### The build path is demonstrably fixed — this is the evidence that matters

`apis.uk/index.html` carries **six** `illustrated-text-block` sections and is the standing
self-regenerating repro. Its `page_components` rows were rewritten at **11:27:27 today**, i.e.
**after** the 09:27 roll, by a per-section render (the build path). They carry **six DISTINCT
tokens**. Under the old code that same path stamped occurrence 0 on every one of them and the
page served `id="c-illustrated-text-block"` six times.

**That is the defect gone at its source, on the path that caused it**, and it is the stickiness
property too: a per-section render no longer re-collides the page.

### ⚠ OPEN, and NOT a collision: the tokens start at `-2`, not at the bare token

Stored and served tokens are `c-illustrated-text-block-2 … -7` = occurrences **1..6**. The
canonical walk over this page's rows (position 1 `hero`, positions 2–7 `illustrated-text-block`)
must assign **0..5** — bare, `-2` … `-6`. So the build-path count is **one high**, consistently.

**Every token is distinct, so there is no collision and no user-visible defect** — this is byte
drift against the canonical walk, which is the errs-safe direction the design documents
(`storedPredecessorCount`'s comment; PLAN §A5 blind spot 2: a ready-list item that produces no
saved row leaves later same-function instances stamped one high, self-correcting at the next full
rerender).

**But do not record that as the explanation — it is not established.** `pages.sections` plans
exactly 6 `illustrated-text-block` and 6 rows were saved, so the obvious "a 7th was dropped" story
is **refuted**. The build's orchestration has been reaped, so the ready list cannot be read back.

**The discriminating test is filed** (`page_rerender`, `reason: template_changed`, `created_by
bugs_open/383`, priority 60):
- **tokens become bare + `-2`…`-6`** → the canonical walk disagreed with the build's count, the
  build's ready list carried one extra same-function item, blind spot 2 is confirmed and it
  self-corrects. Nothing to fix.
- **tokens stay `-2`…`-7`** → the canonical walk itself produces them, which means the
  disagreement is NOT ready-list drift and the derivation needs diagnosing. In that case run `090`
  rather than guessing, and re-read `PlacementFromLoopStep` against a LIVE orchestration's
  `loop_item_index` (0-based, verified 2026-08-24 on a 5-item loop) before touching the code.
