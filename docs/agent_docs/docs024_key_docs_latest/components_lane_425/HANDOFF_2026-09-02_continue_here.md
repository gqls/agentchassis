# HANDOFF — components lane, `bugs_open/425`

**Cold-start: read §0o, then §0j, then §2. Everything after that is supporting detail.**

> **How to read this file.** Sections are ordered by what a new session needs first, not by when
> they were written — the `0a`–`0o` labels are the order they were *added* and are kept so the
> commit history still resolves. **Superseded readings are left in place, struck or marked**, and
> that is deliberate: roughly ten of my claims were overturned during this work, several of them
> after reaching other lanes, and the corrections are more useful to you than a tidy account would
> be. Where a section is wrong, it says so and says what replaced it.

## STATUS AT A GLANCE (2026-09-03)

| | |
|---|---|
| **Do first** | read batches `…000691` then `…000690` (§0o) — both queued, **do not re-file** |
| **Blocked on you** | `691` gates another lane's copy fix; `/index` must have **no second writer** until it reads |
| **Delivered + verified** | migration `682` (card slots), migration `721` (six hero components), two detectors, a widened lint |
| **Open, well-posed** | the producer fix does not execute on *some* rerenders — 16 candidates eliminated (§0j), model refuted (§0i) |
| **Filed from here** | `bugs_open/425`, `bugs_open/457` (live), register `PBP-050` |
| **Council** | `84b51f16` REVISE r3 — two objections actioned, round 4 unsubmitted · `cf3a052c` (721) REVISE r1 |
| **Not mine** | the component birth-gate (`bugsweep4` holds it), `bugs_open/437`, the 57-instance hero repair wave (needs §0o `690` first) |

## 0o. ⭐ THE TWO QUEUED EXPERIMENTS — read these first, do NOT re-file them

Both `triaged` as of 2026-09-03 ~12:30Z. **Find them by batch id; a long silence is the queue, not
a lost item.**

| batch | target | reason | baseline recorded before firing | what each answer means |
|---|---|---|---|---|
| **`…000691`** ⭐ | boxingonline.com **/index** | `template_changed` | instance `f01a8669`, canonical `content-listing`, `query.blog_posts`, 6 items, **all excerpt-bearing, all suffix-free**, written 12:09:49Z | **key GONE** → a rerender actively rebuilds the array with something that is not the fixed projection · **key SURVIVES** → it never touches `articles` here |
| `…000690` | remortgagecalculator.uk **/about** | `image_landed` | instance `228921ba`, **no `background_image` key**, unpinned, own hero `hero_about`, untouched since 08-23 | **key appears and matches** → the 56-page hero wave is viable · **absent** → hero class and deck class share one root cause |

**691 is the better of the two and is the one to read first.** It runs on the *exact page,
component and source* where a rerender produced the old shape **four times** — so unlike `688`
(which ran on a fork with a different source and could not discriminate), both hypotheses predict
different outcomes here.

⚠ **`691` also gates someone else's work.** The `site_delivery_and_editor` lane is holding a copy
fix for that page's CTA (*"the calendar below tells you what's coming up next"* in the last section
of a page with no calendar) — they file it if the key survives, and route it through a build if it
goes. They have excluded `/index` from their assemble batch, so **there is no second writer on that
page**; keep it that way until the reading lands.

⚠ **A `survive` reading licenses the copy fix ON INDEX only** — not a general conclusion that
rerenders are deck-safe. designblog's rerender rebuilt correctly and boxingonline's did not, with
everything a census can hold constant held constant.

**Reading them:** print the item status alongside the artefact (a timed-out watcher otherwise shows
a baseline as a result), attribute the write by `source_item_id` joined on **`page_id`**, and expect
the instance id to change — the sections path DELETE/re-INSERTs, so a new id is not evidence.

## 0j. ⭐ START HERE — the consolidated elimination list and the one pair to work from

**The open question, stated exactly:** two `page_rerender` items with `reason='section_data_resolved'`
produced the NEW item shape and a third produced the OLD one. Same code path, same reason, same
binary, same component row. **What differs is per-INSTANCE and is not any of the following.**

**ELIMINATED — sixteen candidates, each by measurement across all 17 live instances** (jointly with
the boxingonline lane, 2026-09-02/03):

| | |
|---|---|
| the render **path** | refuted — a rerender produced the NEW shape twice |
| the **reason** string | NEW and old both appear under `section_data_resolved` |
| the **component** | `content-listing` on both sides |
| the **source** declaration | `query.blog_posts` on both sides |
| the **binary** | both rolls predate instances on each side; pods unchanged since 09-02 20:56 |
| **time of run** | old-shape writes both precede and follow new-shape ones |
| **locks** (`locked_at`, `lock_type`) | unlocked, 17/17 |
| `component_version_id` | set, 17/17 — and it is **write-only**, nothing reads it |
| `content_item_id` | NULL, 17/17 |
| `schema_mode` | NULL, 17/17 |
| `build_status` | `deployed`, 17/17 |
| `built_from_plan_version` | cuts across both groups |
| `suppressed_sections` | 2 instances have it, both old — but 10 old ones do not; and suppression does not touch `planSection` |
| **per-section wiring** | `pages.sections` is a flat array of slot-name STRINGS — nothing there to differ |
| **empty resolve** (`queryListBelowContract` → `handleMissingField`) | every site has a non-zero eligible population; stored counts match what the query returns, both sides |
| **overnight config change** | `agent_definitions` 0 rows changed in the window; `content-listing` last updated 09-02 **10:43:56**, before all three rerenders |

**THE PAIR TO WORK FROM — `garden-tools.uk`, and nothing else comes close:**

| | `/index` | `/care` |
|---|---|---|
| shape | **NEW** | **old** |
| site, component, source | identical | identical |
| stored items / eligible posts | 4 / 4 | 4 / 4 |
| **written by** | `empty_section` · page-build-handler | `page_rerender` · `template_changed` |

Everything a census can hold constant is constant. Whatever the difference is, it is visible in
that pair or it is not in the database.

**The declaration everything hangs on**, so nobody re-derives it:
```json
"articles": {"type":"array","source":"query.blog_posts","required":true,
             "on_missing":"skip_section","missing_reason":"No blog posts published yet"}
```

**Method note worth keeping:** the model that had to be refuted fitted the first two data points
perfectly. It died at n=4 and would have died at n=17 immediately. **Go to the whole population
first and theorise second** — a story that fits everything and forbids nothing feels like
understanding and is the most expensive thing to carry into a handoff.

## 2. THE OPEN DEFECT — the producer fix does not execute on the rerender path

> ⚠ **THIS SECTION'S FRAMING IS SUPERSEDED — read §0i first.** It is written as though the split
> were *per PATH* (build works, rerender does not). **It is not.** Two `page_rerender` items with
> `reason='section_data_resolved'` produced the NEW shape (designblog, websitepromotion) while a
> third on boxingonline produced the old one — same reason, same binary, same component row. The
> difference is **per INSTANCE**. Everything below is accurate as *evidence* and wrong wherever it
> says "the path"; substitute "these instances". The eliminations in §0j are the current state.


**This is the thing to pick up.** It is well-posed, reproducible, and three diagnosis runs
have failed on it.

### The A/B that establishes it

Same site, same component, same binary, three minutes apart:

| path | result on `content_data.articles[0]` |
|---|---|
| **build** (`needs_page` rebuild, 17:23:02Z) | `excerpt` key **PRESENT**, title suffix **STRIPPED** |
| **rerender** (`page_rerender`, `reason='template_changed'`, 17:26:52Z) | **no** `excerpt` key, suffix **INTACT** |

Reproduced three times (13:59, 17:26, 17:32), all wrote, all old shape.

**Controls, all clean:** the rerender *wrote* (`updated_at` + 8 `page_component_history` rows
at the rerender minute); **0** floor refusals fleet-wide in the window; the slot is bound to
the **canonical** component, not a fork; the markup carries no `article-card__category`, so the
post-682 template ran.

### Why the key's ABSENCE is categorical

The fixed projection writes `"excerpt"` into the item map **unconditionally** — present
whatever the value, including empty. So an absent key cannot be thin data; it can only be code
that did not run, or ran and was discarded. That is the sharpest instrument this bug produced.

### Branches ELIMINATED by reading them — do not re-walk these

`mergedContent = stored ⊕ plan.ResolvedData` with **ResolvedData last**
(`rerender_page_sections_action.go:1709-1716`), so the old shape means the key was **missing
from `ResolvedData`**, not that it lost a merge.

| branch | why it is not this |
|---|---|
| literal-markdown **strip** (~:1600) | gated on `reason == 'literal_markdown'`; dispatch used `template_changed`. Mutates values, never deletes keys |
| **`queryListBelowContract`** → `handleMissingField()` → key never written | the base returns **6** items for this site under its own `listedOnly` floor, and the field declares **no `min_items`** |
| **whole-page escalation** to the writer (STY-048) | **0** `needs_page` items created |
| **`save_page_sections` refusing** | row written *and* archived; 0 floor refusals in the window |
| **stale/cached binary** | killed by the A/B — one binary, two paths, opposite results, minutes apart. **Re-confirmed post-roll:** both pods on v1.0.1355 probed with both controls, and a rerender still produces the old shape |
| **the rerender REASON** (`template_changed` vs `section_data_resolved`) | **[MEASURED 2026-09-02 21:30Z]** a `section_data_resolved` rerender of the same page produced the **same old shape** — no `excerpt`, suffix intact. The reason is not the variable, matching the code: the only reason-gated behaviour inside `rerender_page_sections_action.go` is the literal-markdown strip, and `check_rerender_mode` routes all five recognised reasons to the same step |
| **version pinning** | the pinned version postdates the template update and carries the guard |
| a **third producer** of `articles` | only `rebuild_blog_listing` writes `"articles"` literally; the generic path writes `resolvedData[fieldName]`. Both fixed |

### ⚠ THE RERENDER PATH DELETE/re-INSERTs THE ROW — three consequences for how you verify

`[MEASURED 2026-09-02]` two rerenders of boxingonline `/index.html` produced two **new
`page_components.id`s**: `7dead3e5` at 21:22:26Z, `9e643633` at 21:30:07Z. The path replaces the
row; it does not `UPDATE` it.

1. **Never key a before/after on the instance id.** It differs on every rerender whether or not
   anything changed. Key on `page_id` + `slot_name`.
2. **A fresh instance id is NOT evidence of new content.** It is evidence the row was replaced,
   which happens on the runs that reproduce the old shape exactly.
3. **This is why `page_component_history` is reliable here** — the rows come from the **DELETE**
   trigger, not an update trigger — **and why `page_id` is the only sane join**: it survives the
   replacement, and `component_id` is NULL on 44,555 of 45,285 rows.

### STATUS 2026-09-03 ~08:35Z — the experiment is FILED AND QUEUED, not run

Item batch `…000688`, `reason='template_changed'`, on **`guides-index`** (page `2e738efd`) — the
lane's own new-shape baseline, built 17:23:02Z, `articles[0]` carrying `excerpt`, suffix-free,
untouched since. Filed 08:26:24Z, **still `triaged`**.

`[MEASURED 2026-09-03 08:35Z]` it is behind a backlog: **192 `page_rerender` triaged, 164 of them
older than mine**, draining ~29 per 30 minutes. So it is roughly **three hours out** on FIFO. It
is queued, not stuck.

> ⚠ **A TIMED-OUT WATCHER PRINTS THE CURRENT STATE, WHICH READS EXACTLY LIKE A RESULT.** My first
> watcher expired after 8 minutes and printed
> `e6b51597 | t | f | 4 | 17:22:27` under the heading "THE DISCRIMINATOR RESULT". That is the
> **unchanged baseline** — the item had not run. Nothing in the output says so. **Always print the
> terminal status alongside the reading, and treat an unchanged `updated_at` as "did not run"
> rather than "no change".** Here `17:22:27` is the give-away: it predates the dispatch by fifteen
> hours.

**TWO EXPERIMENTS ARE IN FLIGHT ON THE SAME PAGE — both queued, neither read.** Do not re-file
either; find them by these ids:

| what | id | filed | reads out as |
|---|---|---|---|
| **mine** — `template_changed` rerender on the new-shape baseline | batch `00000000-0000-0000-0000-000000000688` | 08:26:24Z | `excerpt` **gone** → the rerender path rebuilds with unfixed code · **survives** → it never touches `articles` |
| **the delivery lane's** — `needs_rerender` / stale_chrome, assemble-only | `ec92320f-3037-448a-bd55-de8385404d92` | 08:35:44Z | tests whether assemble mode really is a byte-for-byte re-ship |

⚠ **The chrome wave's per-page items are CHILDREN.** `rerender-pages` creates them once
`ec92320f` is claimed, so their `source_item_id` chain hangs off that child batch, **not** off
`ec92320f` directly. Find the children by `page_id` + creation time, not by parent id.

⚠ **THE DELIVERY LANE'S PASS CONDITION NEEDS A POSITIVE CONTROL AND I HAVE SENT THEM ONE.**
"No new history row for guides-index" cannot distinguish *assemble preserved the array* from *the
wave never reached this page* — both give zero rows. The control is to confirm a per-page
`_assemble` item for `2e738efd` exists and reached terminal BEFORE reading history as a pass.
Same family as the three absences that cost me yesterday (a 20-second log window, a 98%-NULL
filter column, a timed-out watcher).

⚠ **The two branches want OPPOSITE readings of the same table** — if assemble does not replace
rows, the ABSENCE of a history row is the pass; if it does replace them, the PRESENCE of an
`artefact_archive_trigger` row whose archived `content_data` still carries `excerpt` is the pass.
Decide which branch you are in before reading, not after.

**Attribution does NOT depend on running this before or after anything else** — which is why the
delivery lane's chrome wave was released rather than held for three hours.
`page_component_history` carries `source_item_id` on **1,169 of 1,196** `save_page_sections_overwrite`
rows in 24h, so a sections-path write names the item that caused it. Read the chain, do not police
the order:

```sql
SELECT h.created_at, h.source, h.source_item_id,
       (h.content_data->'articles'->0 ? 'excerpt') AS had_excerpt_BEFORE_this_write
  FROM page_component_history h JOIN pages p ON p.id = h.page_id
 WHERE p.id = '2e738efd-...' AND h.created_at > now() - interval '12 hours'
 ORDER BY h.created_at;
```

That table archives **the state being REPLACED**, so each row says what the page looked like
immediately *before* that write — a per-write before/after chain, already attributed.

⚠ Assemble-path writes may arrive as `artefact_archive_trigger` rows with **no** `source_item_id`.
For the delivery lane's own prediction that is cleaner, not worse: a byte-identical re-ship trips
**neither** trigger (both test `IS DISTINCT FROM`), so **an absent row is their pass and a present
row is their finding**.

### ⭐ THE EXPERIMENT TO RUN FIRST — it breaks the ambiguity that blocked this all evening

Proposed by the `site_delivery_and_editor` lane as a question ("would a rebuild of boxingonline
`/index.html` destroy a repro you still need?"). **It would not — it makes the repro strictly
better**, and it is now the highest-value next action.

**The ambiguity, stated plainly:** today the stored array is the OLD shape, and a rerender
produces the OLD shape. Those two are *byte-identical*, so I cannot tell whether the rerender
**overwrote stored with an old-shape rebuild** or **never touched the key and stored simply
survived**. Every reading this evening ran into that wall.

**A build-path rebuild removes the wall by changing the baseline:**

1. Fire a `needs_page` rebuild of boxingonline `/index.html` (the build path is the known-working
   route — `guides-index` did exactly this at 17:23:02Z and produced `excerpt` present, suffix
   stripped). Confirm the new shape landed: `content_data->'articles'->0 ? 'excerpt'` → true.
2. Then fire a `page_rerender`, `reason='template_changed'`, `spec.page_name='index'`.
3. **Read the key again. The two outcomes now say opposite things:**

| after the rerender | conclusion |
|---|---|
| `excerpt` key **GONE** (reverted to old shape) | the rerender path **actively rebuilds** the array with code that is not the fixed projection — it overwrites good data with old-shape data |
| `excerpt` key **SURVIVES** | the rerender path **never touches** `articles`; `ResolvedData` lacks the key and stored wins by default |

Either answer halves the remaining search space, and neither is available while stored and
produced are identical.

⚠ **Confirm at `page_component_history` keyed on `page_id`** (NOT `component_id`, which is NULL
on 44,555 of 45,285 rows) that the rerender actually wrote — and take a same-window positive
control from another component, or an absence proves nothing.

⚠ **Cost of being wrong:** none to the repro. The defect reproduced on demand three times by
re-filing a rerender, so the old state is one dispatch away whichever result comes back.

### What has NOT been established

- Whether `planSection` is reached for this field on this path at all. `plan_sections_action.go:2695-2709`
  resolves any `query.*` source **unconditionally** — no "only if missing" gate — so the read
  says it should be. The behaviour disagrees. **The behaviour is what ships.**
- An instrumented run (live tails on both pods during a rerender) captured **no**
  `queryresolve:` or `plan_sections:` message — **recorded as INCONCLUSIVE, not as evidence.**
  The pods serve every agent so surrounding lines cannot be attributed without correlation
  filtering, and there is no positive control that the resolver's line would appear if it ran.

### Diagnosis runs — all three failed

`afbf8544` capped UNVERIFIABLE · `fe4b8537` NOT CONFIRMED (0-row data requests; one behind an
`EXPLAIN` erroring `syntax error at or near $`, a parameterised query handed to `EXPLAIN`
unbound — a concrete lead) · `c755b0be` UNVERIFIABLE **even with the four eliminations stated**.
Three failures on a well-posed symptom is a signal about the loop's reach here, not only about
phrasing. **A fourth run in the same shape is probably not the move.**

---

## 0i. ⛔ THE PATH-SPLIT MODEL IS REFUTED — a rerender DID produce the new shape

**Read this before §2 and §0h; both are framed on a model this refutes.**

The boxingonline lane built a fleet discriminator — all 17 `content-listing` instances with an
`articles` array — and noted the split is **not chronological**: `boxingonline/index` (old shape)
was written LATER than two new-shape instances. They asked me to attribute each write, since that
is the side of the line I can reach. `[MEASURED 2026-09-03]`, via
`page_component_history.source_item_id`:

| shape | instance | wrote_via | handler | reason |
|---|---|---|---|---|
| **NEW** | designblog.co.uk/index | **page_rerender** | **page-rerender** | **section_data_resolved** |
| **NEW** | websitepromotion.co.uk/index | **page_rerender** | **page-rerender** | **section_data_resolved** |
| NEW | garden-tools.uk/index | empty_section | page-build-handler | (no reason) |
| NEW | boxingonline/guides-index | needs_page | page-build-handler | rebuild_cleared_component |
| NEW | advertise.co.uk/index | needs_page | page-build-handler | image_landed |
| old | **boxingonline/index** | **page_rerender** | **page-rerender** | **section_data_resolved** |
| old | homegarden ×6, dartsonline ×2, garden-tools/care, idea.uk | page_rerender | page-rerender | template_changed / cta_links_stale / section_data_resolved |

**Two `page_rerender` items with `reason='section_data_resolved'` produced the NEW shape, and a
third produced the OLD one.** Verified tightly on designblog: history row `05:25:28` matching
`page_components.updated_at` exactly, item `03304c6a`, complete, and the row now carries the key.

**So the rerender path CAN re-resolve.** It is not "the sections path never resolves listings".
The difference is **per-instance, not per-path**, which is the "completely different hunt" the
boxingonline lane named as the alternative — and it is the one we are in.

**Also eliminated in the same measurement:** it is not the component or the source. `content-listing`
with `articles ← query.blog_posts` appears on **both** sides — 4 of the 5 new-shape instances and
all 12 old ones. And `garden-tools.uk` has one of each (`/index` NEW, `/care` old).

**Everything built on the path split needs re-reading**, including §0h's "the rerender renders the
current template against stale data" — which remains true *of boxingonline/index* but is not a
property of the path.

**What is NOT yet eliminated:** whatever differs per instance. Not the component, not the source,
not the reason, not the handler, not the ordering in time. Both the 12:28 and 20:56 rolls predate
several instances on each side, so it is not the binary either.

### Eliminated since the refutation — three more, and the tightest pair available

**`pages.sections` is a FLAT ARRAY OF SLOT-NAME STRINGS, not objects.** The boxingonline lane
flagged it as the most obvious unexamined per-instance surface after their `e->>'slot_name'` and
`e->>'component'` matches returned nothing on every page. That is why: the entries are scalars —
`["featured-content","content-listing","advertising","info-card-grid","call-to-action"]`. **There
is no per-section wiring there to differ.** Surface closed, and their query failed for a reason
rather than by mistake.

**The empty-resolve hypothesis is dead**, and it was the best-shaped one I had:
`queryListBelowContract` → `handleMissingField` → key never written → stored survives would
explain the outcome exactly. `[MEASURED 2026-09-03]` every site has a non-zero eligible blog-post
population, and stored item counts match what the query would return on **both** sides.

**THE TIGHTEST PAIR, and it is the thing to work from:** `garden-tools.uk` carries one instance of
each — `/index` NEW, `/care` old — with the **same site, same component, same source, same 4
stored items, same 4 eligible posts.** Everything a census can hold constant is constant. The only
recorded difference is how each was written: `/index` by `empty_section` (build path), `/care` by
`page_rerender`/`template_changed`.

> **[UNTESTED OBSERVATION, flagged rather than claimed]** Among **rerender-written** instances
> only, the two NEW ones were written 09-03 (04:18, 05:25) and every old one 09-02 or earlier
> (21:30, 13:5x, 03:10, 08-28). That looks chronological — **but the binary did not change in that
> window**: both pods started 09-02 20:56/20:57 and are still running, so `boxingonline/index` at
> 21:30 and `designblog/index` at 05:25 ran the same code. If the boundary is real, something
> other than the binary changed between 21:30 and 04:18. I have not identified it and am not
> proposing it — recorded so it is neither lost nor believed.

## 0h. ⭐ THE SHARPEST NARROWING YET — the template renders FRESH while the data stays STALE

`[MEASURED 2026-09-03, verified independently at the stored artefact]` boxingonline
`/index.html`, content-listing slot, rendered 09-02 21:30:07 by a `section_data_resolved`
rerender:

| | count | means |
|---|---|---|
| `article-card__excerpt` elements | **0** | 682's guard **took** — the slot collapsed |
| `article-card__date` elements | **0** | 682's guard **took** |
| `article-card__title` elements | 6 | six cards, so the section rendered |
| `" \| Boxing Online"` occurrences | **12** | 6 headlines + 6 alt texts — the producer did **NOT** take |

**Through ONE render, the config half applied and the code half did not.** So the rerender path
is not inert and is not "emitting the old shape wholesale": it **re-renders the current template**
against **stale data**.

**AND THAT RULES OUT THE 'WRONG PROJECTION' READING** — including the one the boxingonline lane
proposed when they found this. Their framing was *"the divergence is not at re-resolution, it is
at which projection the resolved items are passed through."* It cannot be, because **there is no
old projection left to pass through**: commit `f57f5ad1f` fixed *both* producers of this shape,
so any re-resolution — through `resolvePagesWhereType` or through `scanBlogArticles` — would
necessarily emit `excerpt` and a stripped title. The items carry neither.

**Therefore the array was never resolved at all.** These are the STORED items passing through a
freshly rendered template, which is exactly `mergedContent = stored ⊕ ResolvedData` with
`ResolvedData` lacking the `articles` key. The hunt narrows to: **why is `articles` absent from
`ResolvedData` on a path that demonstrably rendered the template it belongs to?**

> **⚠ A MEASUREMENT CAUTION THAT CUTS BOTH WAYS — theirs, and it is a good one.** Counting
> `article-card__excerpt"></p>` (the EMPTY element) returns **0** both when the slot collapsed and
> when the deck is filled. Post-682 those are indistinguishable on that predicate. Count the
> **element itself** and measure its **inner length** separately: `0 elements` = collapsed,
> `6 elements with content` = filled, `6 elements empty` = the pre-682 defect. The delivery lane
> gave the inverse caution earlier — do not file filled decks as a 682 regression — which is the
> same ambiguity from the other side.

### `data_path` / `schema_mode` are a DEAD END — the schema comes from the component, not the row

The boxingonline lane measured that the affected `page_components` row has `data_path` NULL and
`schema_mode` NULL, and that the experience-loop lane found `data_path` **empty on every row
fleet-wide**. Good question, and it is answerable by reading rather than testing: **neither column
participates in resolution.**

The chain, traced 2026-09-03:

```
rerender_page_sections_action.go:1430  comp, _, haveComp := resolveComponent(s)
                        :361-393       resolveComponent → byID[s.componentID]  (live component by id)
                                                        → else schemas[s.slotName]
                        :354           byID  ← loadComponentSchemasByID(...)
plan_sections_action.go :2110-2133     → loadContentComponentsByID → componentInfoFromRaw
                                          → componentInfo.InputSchema  (from content_components)
                        :1450          planSection(..., comp, ...)
                        :2460          fieldsRaw, ok, _ := SchemaContentFields(comp.InputSchema)
                        :2695-2709     for each field: if source has prefix "query." → Resolve(...)
```

**The schema `planSection` iterates comes from `content_components` — the LIVE library row —
looked up by `page_components.component_id` or by slot name.** `data_path` and `schema_mode` are
never consulted on this path. So their being NULL here and everywhere explains nothing, and is not
worth a test.

**And that deepens the puzzle rather than solving it**, which is worth stating plainly: every link
in that chain reads correctly. The component resolves live, the schema is the current one (it must
be — 682's guards rendered), the `query.*` branch is unconditional, and `ResolvedData` merges
last. **The code says the array must be re-resolved; the artefact says it was not.** That
contradiction, not a missing gate, is the actual open question — and it is why three diagnosis
runs have failed on it and why a fourth in the same shape is not the move.

## 1. What is DONE and verified

### `bugs_open/425` — card components rendering empty slots (the owner's "cards need better designs")

**Root cause:** two producers write the same standard list-item shape for one shared
component and disagreed. The deck was *already in the row* under `meta_description` while
the template read `.excerpt`; the headline carried the site-name suffix one producer
stripped and the other did not.

| half | state |
|---|---|
| **template** — migration `682` guards four per-item slots + the section header | **DELIVERED on 10 of 14 pages.** Verified at the served markup *and* corroborated by `page_component_history` at the exact rerender minutes. Two independent instruments agree |
| **producer** — `queryresolve.ListItemTitle` / `ListItemExcerpt`, both producers sharing one spelling | **IN THE BINARY, NOT EXECUTING on the rerender path.** See §2 |
| **detectors** — `check_card_slot_guards.py` (new), `check_list_empty_states.py` (widened) | LIVE, advisory, `--self-test` passes |

**4 of the 14 pages were REFUSED** by the section COMPONENT floor (`bugs_open/253`) and
**cancelled** with the reason written into each row. This fix trips that floor *by design* —
every collapsed empty element carries layout classes, so the least-fed sites flatten hardest.
⚠ **NARROWED 2026-09-03: a SCOPED override is possible and my earlier "fleet-wide" was too
strong.** Each agent has its own `save_sections` step config — `[MEASURED]`
`page-build-handler` carries `section_shrink_floor = 0.1` (owner-ruled, migration 725) while
`page-rerender`'s is unset, and they do not touch each other. The standing caution is narrower:
setting it on **`page-rerender`'s** step reaches the fleet's highest-volume pipeline, and even a
per-agent override is per-pipeline-for-the-duration rather than per-page — pair it with a
monitored rollback, as that lane did.

⚠ **Do not set `section_component_floor` casually** as its error invites: it is read from
`params.StepConfig.Config`, i.e. the `page-rerender` **step**, so setting it weakens the guard
for every rerender in the fleet to land one page.

### Migration `721` — six page-scope hero components declare their image field

Applied, ledgered, verified at the artefact, committed `d9e18e2e1`, council `cf3a052c`
**submitted, verdict not yet read**.

Six components render `{{or .hero_url .background_image}}` while declaring no image field, so
their generated page hero was orphaned and the page wore the site-wide homepage image —
passing every "has an image" check while doing it. `hero-tool` 77 instances, `about-hero` 43,
`contact-hero` 25, `services-hero` 6, `case-studies-hero` 5, `use-cases-hero` 2 = **158 live**.

⚠ **The `type` is the fix.** `sectionHasImageField` (`plan_sections_action.go:2936`) gates the
authoritative hero aliasing on `type == "image" || "image_url"`. As `"text"` the gate stays
false, the aliasing never runs, `hero_url` still wins, and the migration applies cleanly and
changes nothing.

**Owed:** `gamedesign.uk` is verifying at its own artefact and will report the `url()`s back.
Nothing has re-rendered yet as far as I know.

---

## 0d. MIGRATION 721 — working mechanically, ONE page improved so far, and my own count went stale in 12 hours

`[MEASURED 2026-09-03 ~08:50Z]` instances of the six hero components carrying
`content_data.background_image`:

| era | rows | window |
|---|---|---|
| **before 721** | **5** | 08-25 20:18 → 09-02 14:00 |
| **after 721** (applied 20:15:47Z) | **23** | 09-02 20:20 → 09-03 05:50 |

**So the field resolves and lands.** 721 is not inert — 23 instances have gained
`background_image` through resolution since it applied, which is the mechanism working.

**But the visible improvement is ONE page.** Of those 23: **22 have no page-scope hero asset**, so
they take the declared fallback or the site hero — correct behaviour, no change to what a reader
sees. **1 renders its own hero.** Only a page that BOTH has its own asset AND has re-rendered
since 20:15Z can improve, and most overnight re-renders were on pages without one.

**The designblog lane predicted exactly this and was right**: the fix is *necessary and not
sufficient*; closing the class needs re-renders carrying a re-resolving reason across the affected
pages, then verification at the SERVED bytes, filename-anchored. The component row now reads
correct everywhere while the pages still serve the old image.

> ⚠ **AND MY OWN "5 of 158" WENT STALE IN TWELVE HOURS — by ADDITION, caused by my own change.**
> I published that count in migration 721's header and in its council submission, and used it to
> reject a proposed cheaper route. It was correct when measured (2026-09-02, pre-migration) and is
> now 28. Nothing about it was wrong; it simply counted a population my own edit then grew. This
> is precisely why CLAUDE.md requires a count to carry the date it was counted — the header says
> `[MEASURED 2026-09-02]`, so the staleness is visible rather than silent, and
> `--since 2026-09-02` re-derives it. **A census does not go wrong; it goes stale, and yours can
> be stale because of you.**

## 0l. NEW BUG FILED FROM THIS LANE — `bugs_open/457`, and it is live

`rebuild_blog_listing_action.go:403-407` appends an orphan `page_components` row on **every run**
where `findBlogListingSlot` fails: `position` hard-coded to `3`, `component_id` never set,
unconditional INSERT with no upsert or existence check.

`[MEASURED 2026-09-03]` boxingonline `articles-index` carries **6 NULL-component rows at position
3** accumulated 08-31→09-02, beside a legitimate `call-to-action` at the same position. Fleet-wide:
8 such rows on 3 pages; **12 pages carry more than one row at one position**.

**It only surfaced now because migration 316's `uq_page_components_no_byte_identical_duplicate`
refuses the insert once the listing renders to bytes an orphan already holds. The constraint is
not the bug — it is what finally reported two days of silent accumulation.** Before it, the action
appended junk and returned success.

**Live blast radius:** the failure aborts the action *before* `create_rerender_items`, so a
boxingonline chrome refresh created **none** of its ~18 child rerenders and retries identically.
Same file, same carelessness about component identity as `425` fix-candidate 5 (`loadListingTemplate`
looks its template up by NAME rather than following the `component_id` it is writing).

## 0c. A FLEET FINDING made while measuring this bug — 30 rerenders carrying PROSE as their reason

`[MEASURED 2026-09-02]` completed `page_rerender` items over 24 hours, by reason:

| reason | completed |
|---|---|
| **(none — assemble mode)** | **1,418** |
| `section_data_resolved` | 60 |
| `cta_links_stale` | 39 |
| `template_changed` | 17 |
| **a human-readable SENTENCE** | **30** |

Those 30 were filed by two lanes via three migrations — 701 (16), 696 (11), 693 (3) — with
`spec.reason` set to explanations like *"FCA rule citation corrected by migration 696 (owner
decision 2026-09-02)"*. `check_rerender_mode` tests **equality against five literals**, so a
sentence takes `else_step` exactly as NULL does. All 30 completed and re-shipped stored HTML.

**Both lanes told.** The `bugs_open/357` lane then supplied the important qualification: for
*their* pages the two routes converge **byte-identically by construction** (the adopted template
IS the stored bytes), so their artefacts are correct and only their *mechanism* claim was wrong.
My "has not reached those pages" was too strong for their case.

**Recorded in `LANDMINES.md`** as an append to the existing reason entry, which also needed
correcting: its quoted condition listed **three** values where the live config has **five**
(`template_changed` and `literal_markdown` were missing), and its remedy tests for NULL, which
cannot see a set-but-unrecognised value. The check now tests **membership, not presence**.

### ⚠ AND THE INSIGHT THAT BEARS DIRECTLY ON §2

The 357 lane reached, from the opposite direction, the property that has blocked this bug all
evening: **byte-identical output is not evidence about which path ran.** Their md5-unchanged
result proves only that concatenation reproduces its input; my unchanged array proves only that
*something* left the same bytes. Neither of us can read the artefact alone and say which branch
executed.

That is exactly why the §2 ⭐ experiment matters: it does not try to read harder, it **changes the
baseline** so the two hypotheses stop producing the same bytes.

## 0k. 688 RAN — and it is INCONCLUSIVE, by a flaw I flagged when choosing the page

`[MEASURED 2026-09-03 10:45:53Z]` batch `688` (`template_changed`, `guides-index`) **completed and
wrote**: `excerpt` key **present**, title **suffix-free**, 4 items, new instance id (the path
DELETE/re-INSERTs).

**It does not discriminate.** The baseline was already NEW-shape, so *"re-resolved through the
fixed projection"* and *"never touched `articles`, stored survived"* produce the **same** result.
I noted that risk when selecting the page and accepted it because it was the only new-shape
baseline available; it has now cost the experiment. **`690` — old-shape baseline, key absent — is
the one that discriminates, and it is still queued.**

### ⚠ AND IT CORRECTS AN INSTRUMENT I GAVE ANOTHER LANE

I told the delivery lane: *"`page_component_history` row present = content changed; absent = it did
not."* **The first half is wrong.** `save_page_sections_action.go:875-889` snapshots
**unconditionally** before every overwrite:

```sql
INSERT INTO page_component_history (...) SELECT ... 'save_page_sections_overwrite', $2
  FROM page_components pc JOIN pages p ON pc.page_id = p.id
 WHERE pc.page_id = $1 AND pc.rendered_html IS NOT NULL AND LENGTH(pc.rendered_html) > 0
```

No `IS DISTINCT FROM`. **A history row proves a SAVE RAN, not that content CHANGED.** The trigger-written
rows (`artefact_archive_trigger`) *are* change-gated; the code-written ones are not, and they are
the ones carrying `source_item_id`.

**And the same read explains my earlier zero.** That INSERT writes `pc.id` — the
**`page_components` row id** — into the column named `component_id`. So
`page_component_history.component_id` does **not** hold a `content_components` id, and filtering
it by one returns nothing. My "join on `page_id`, not `component_id`" advice was right; the reason
is sharper than "the column is mostly NULL" — **the column holds a different entity than its name
suggests.**

## 0e. THE HERO CLASS AND THE DECK CLASS MAY SHARE ONE ROOT CAUSE — one-page test filed

`[MEASURED 2026-09-03]` **57 instances across 24 sites** have their own page-scope hero and are not
rendering it (5 already correct, 104 with no own hero — nothing to repair). Derived independently;
the `bugs_open/114` census says 61, same class.

**The proposed repair — "re-renders carrying a re-resolving reason" — is an UNTESTED premise, and
my one data point contradicts it.** I traced the single page that has visibly improved since 721
(`garden-tools.uk/contact`, 09-02 23:18) via `page_component_history.source_item_id`:

```
item 726aa1e5 · type unbuilt_internal_link · handler page-build-handler · reason (NONE)
```

**The BUILD path.** Not a re-render, and carrying no reason at all. So the route we were about to
dispatch 57 items down has **zero** confirmed successes, while the one success came via the route
nobody proposed. The delivery lane has marked its own relayed condition `[UNTESTED → under test]`
— it was a hypothesis wearing the voice of a fact.

**ONE DISPATCH, NOT 57.** Filed batch `00000000-0000-0000-0000-000000000689`:
`advertise.co.uk/about`, `reason='image_landed'`, chosen because that site has **0 open work
items** so attribution cannot be muddied. Baseline recorded before it runs: instance
`4e681d76`, `hero-about`, **no `background_image` key at all**, `updated_at` 09-02 17:02:11.
Its own hero asset is `hero_about`.

> **CHECKED BEFORE READING THE TEST — every post-721 rerender on this class was ASSEMBLE MODE.**
> `[MEASURED 2026-09-03]` **66** completed `page_rerender` items since 20:15:47Z on pages carrying
> one of the six hero components with their own hero asset. **All 66 carry `reason = (none)`.
> Zero qualifying reasons. Not one routed to the sections path.**
>
> So the "nine re-rendered since 721, none recovered" evidence **proves nothing about the sections
> path** — it was never exercised. That makes the one-page test *necessary* rather than
> confirmatory: batch `689` will be the **first** time the sections path runs against this class
> since the field became declarable.
>
> ⚠ **And beware the word RECOVERED in that sweep.** Ten rows read as recovered, all on
> `leopardessconsulting.co.uk` and `garden-tools.uk/contact` — but those are the pages that were
> ALREADY correct before 721 (the pre-existing 5) plus the one the build path fixed. Assemble mode
> preserved bytes that were already right. **"Currently correct" is a STATE; it is not evidence of
> a transition**, and a column labelled RECOVERED invites reading it as one. **No page has been
> fixed by a re-render. Zero.**

| result | conclusion |
|---|---|
| `background_image` appears **and matches `hero_about`** | the wave is viable → prepare the other 56 as a HELD migration in the 683 shape, hand firing to the site owners |
| **no key, or the site hero instead** | **the hero class and the deck class share one root cause — `425` §2** — and the fix is the rerender path itself, not a wave |

**The second outcome is the more valuable one**, because it unifies two classes worked separately
by three lanes, and it means a 57-item wave would have completed, stamped fresh timestamps and
changed nothing.

⚠ **QUEUE POSITION — rank with the SELECTOR'S OWN QUERY, never a proxy census.** A proxy count
(`status='triaged' AND pipeline='build'`) said 21 sites / 270 items ahead and "hours". Replicating
the trigger's full eligibility SQL puts **boxingonline 2nd** and `remortgagecalculator` (batch 690)
**~10th** — minutes, not hours. The proxy was wrong because it ignored the eligibility CTE's first
clause: `s.locked_at IS NULL OR wi.id = ANY(lock_except_item_ids)`. `adversecreditmortgage.co.uk`
held the fleet's oldest window minimum (09-01) and is **LOCKED** (`locked_at` 2026-08-18,
`locked_by` "portfolio_positioning: owner HALT"), so its 22 items are invisible to the trigger —
which is why it sat untouched for two days while 99 build items an hour completed elsewhere.
**A census that omits one clause of the selector's WHERE ranks you against a population the
selector cannot see.** (Found by handing the delivery lane a measured inconsistency rather than
planning around their estimate; they resolved it in one query and logged the check.)

## 0f. ⚠ A RETRACTED FINDING — "33 of 57 are version-pinned and unreachable" WAS WRONG

`[MEASURED 2026-09-03]` I found that 33 of the 57 repairable hero instances carry a
`component_version_id` pinned to a version predating migration 721, and that those pinned
versions lack the `background_image` field. I reported that as **33 instances unreachable by any
rerender**, proposed it as a distinct repair needing its own owner, and a peer lane was routing an
ownership decision to the owner on it.

**It is wrong. `page_components.component_version_id` is WRITE-ONLY.**
`save_sections_component_version.go:40`, verbatim:

> *"THIS FILE ONLY WRITES. Nothing reads component_version_id, so this change is inert by
> construction: it cannot alter what any page serves."*

Its header adds, measured 2026-08-22: *"0 of 1,930 rows populated, no Go code writing it, no Go
code reading it — dormant machinery."* The only reader is `loadStoredSections`, which SELECTs it
to carry the stamp forward, **not** to resolve a schema. Component resolution keys on
`component_id` against the **live** `content_components` row, so a pinned page resolves exactly
like an unpinned one.

**What that costs and returns:**
- there is **no unreachable category** — the split is 57 repairable, full stop;
- batch `689`'s cancellation was for a wrong reason (the page was always a valid test). The row's
  error now carries a dated correction; `690` supersedes it and is equally valid;
- **the dartsonline result returns to its face value** — a `section_data_resolved` rerender,
  sections path, attributed by `source_item_id`, that did **not** resolve the newly-declared
  field. I had explained it away. It is unexplained again, and it is a **second** data point
  pointing where §2 points, on a *different source type* (`site_assets.*`, not `query.*`).

> **The error, because it is the seventh on this thread and the pattern has not varied:** I found
> a **correlation** — pinned AND missing field — and reported **causation**, without checking that
> the mechanism exists. Two greps would have caught it, and I only ran them because the estate's
> landmine list pointed me at the pin, which made me ask what the pin actually does. **Before
> attributing an outcome to a mechanism, confirm something READS the thing you are blaming.**

## 0m. A SWEEP OF MY OWN COUNTS — the decorative ones had already drifted

Prompted by the `bugsweep4` lane's diagnosis of why three of my figures reached a shared record
unchallenged: **the wrong numbers survived precisely because nothing in the argument depended on
them.** Their suggested first pass — *"the counts you would not notice changing"* — run against
this lane's own documents:

| recorded here | re-run 2026-09-03 |
|---|---|
| `check_card_slot_guards.py`: 77 slots / 46 wrappers / **47 of 55** | **79 / 48 / 48 of 57** |
| `check_list_empty_states.py` docstring: 29 of 72 across **55** | **30 of 74 across 57** |

**Every figure drifted inside 24 hours, by addition, and I would not have noticed any of them.**
The library grew 55 → 57 range components; nothing in either document loads the numbers, so
nothing tested them. The lint's figure was the worse case — it sat in the **tool's own docstring**,
where a reader takes it as current.

**Fixed by dating it and saying it drifts**, not by updating it: an updated number is stale again
next week. The command is the answer, the figure is an illustration.

> **The rule, sharper than the one I had:** `[MEASURED <date>]` makes a count **re-derivable, not
> correct**. The estate already records that a marker proves a measurement was *claimed*, not
> *complete*, and that the deeper test is whether it could have come out otherwise. This adds the
> part that was missing: **whether any inference in the document ever LOADS it.** A figure nothing
> rests on is never tested by the argument carrying it, however carefully taken and however well
> marked — so it is simultaneously the most likely to be wrong and the least likely to be caught.

## 0n. METHOD — a claim tested against a population YOU chose is the weak form

From the `bugs_open/437` lane via `bugsweep4`, and it is the most transferable thing this lane
produced today, so it is recorded even though the finding itself is theirs.

Their fix rested on *"exactly 1 live component qualifies"* — established by their own sweep over
their own population. **My five legacy-dialect components were a population they did not choose**,
found for an unrelated reason (I was answering a CGV-030 question about per-item slots). Re-running
their structured-property test over exactly those five: of **21 declared item properties, exactly
ONE is structured** (`mechanism-flow.steps.branches`, array); the other 20 are string or number.
Same answer, someone else's population — and it **could have falsified them**: a nested structured
property on `checklist`, `comparison-table`, `evidence-timeseries` or `period-calendar` would have
meant the fix silently under-covered.

> **The rule:** a claim checked against a population you selected is the **weak form**. The cheap
> upgrade is to re-run it against one somebody else named **for their own reasons** — precisely
> because it was not chosen to test your claim. It costs one query and converts "I looked and found
> nothing" into "a population picked by someone with different purposes also found nothing".
>
> This is the sibling of the control lesson from earlier today (*a positive control must exercise
> the same row population as the claim*) and sharper in one respect: that one is about the control
> matching the claim, this one is about the population being **independent of the claimant**.

**Also resolved, and it closes the overlap I flagged:** `240` and `437` are two defects in one
place, not one. `240` was the dialect being **misread** — keywords reaching the writer as field
names — and is fixed. `437` is one layer past it: the dialect is read **correctly**, the per-item
names come out right, but the nested **type** is lost, so a field declared array-of-objects reached
the prompt as `"branches": "..."`. The proof they are distinct was already in their evidence: the
failing prompts listed the **real** names (`body, branches, marker, note, title`), not the keywords.
Their fix re-declares **no** component schema, so converting the four legacy siblings is free of
`437` either way.

⚠ **A COORDINATION GAP WORTH KNOWING:** the `437` lane and this one **failed to find each other in
both directions**, while both live and both named after what they were working on. I guessed a
neighbour by bug number; they could not resolve me either. `ListAgents` is the only place a session
appears, an idle session still receives, and a stale listing is worse than none — **re-run
`ListAgents` before guessing, and hedge the message ("if you are not X, please forward") when you
do guess.** That hedge cost the wrong recipient one lookup and lost nothing.

## 5. Practice notes earned today — the expensive ones

- **Every instrument that survived reads the ARTEFACT; every instrument that fell reads a column,
  or a table about a column.** Six overturned claims, one pattern.
- **A positive control must exercise the same row population as the claim.** I proved
  `page_component_history` works using 389 rows from *other* components, concluded it was blind
  for mine, and told a peer to stop using it. My query was filtering `component_id`, which is
  **NULL on 44,555 of 45,285 rows**. Join on `page_id`.
- **A sound methodological rule reasoning from a false premise produces a confident, quotable,
  wrong conclusion** — and it travelled to a peer's notes inside the hour *because* it was
  well-phrased. Well-phrased + confirming is the accelerant pair.
- **A census offered as evidence about what CODE returns must replicate that code's predicate.**
  I published "the resolver would return 5" from a query missing `ListedPageEligibilitySQL`; a
  peer built a discriminating test on it that would have burned a round proving nothing.
- **Grep before filing.** Three would-be new landmines this session were already recorded
  (`pages.status`'s dead `'deployed'` literal — three entries; the roll-kills-council trap, with
  a *better* check than the one I had; the per-item slot class).
- **The chassis log's readable window is ~SECONDS**, not hours: `--since=3h` returned 318 lines
  spanning 20 seconds. Tail while inducing; never read back.

---

## 3. Council state

| correlation | subject | state |
|---|---|---|
| `84b51f16-086f-493d-b188-49231d0ca907` | 425 (Go + 682 + 683) | **REVISE at round 3.** Two real objections ACTIONED (`c1178442d`, `9f6f91325`): reuse `datahelpers.SafeCut`/`TruncateString` instead of hand-rolled rune slicing; 683's header no longer tells a human to verify a deploy with `git merge-base`. **Round 4 NOT submitted.** |
| `cf3a052c-37e3-4573-ad6c-fdddd83faeda` | 721 (hero image fields) | **Submitted, verdict unread.** |

Round 3's remaining objections were largely **sketch artefacts** — verified: the 682 ROLLBACK
carries the real pre-image (1 occurrence of the markup, 0 of the placeholder), and 682 does wrap
UPDATE+verify in `BEGIN`/`COMMIT` with a drift guard before the UPDATE. If you resubmit, fix the
**sketches**, not the files.

---

## 4. Peers, and who is waiting on what

- **`site_delivery_and_editor`** — owns the boxingonline pipeline. Has the two-probe acceptance
  test (class-count for "did the template render", key-presence for "did the producer run").
  **Waiting on a ping when the producer half is explained.** Verified the 682 delivery at the
  served markup independently.
- **`boxingonline.com`** — raised 425. Holds the before-capture. Contributed the correction that
  **a correct rerender `reason` is necessary and NOT sufficient** (rows completed twice with the
  right reason and produced the old shape).
- **`gamedesign.uk`** — routed the hero defect, handed me the migration, verifying `721` at its
  artefact and will report `url()`s back.
- **`designblog.co.uk`** — routed the owner's "every site looks the same" directive; has my
  census. Routed the counter-instance that stopped me building a two-component version of 721.
- **`theme kits`** — building `page_archetypes`. I gave them a key-collision finding
  (`content_components.function` is not unique after `forked_from IS NULL`: `site-header`,
  `site-footer`, `tool-agent-complexity-estimator` still collide, resolved by `ORDER BY name`).

### The owner's sameness directive — measured, ACKed, NOT actioned

`hero` on 37 sites (634 instances); **74–87% of every site's slots come from the same 10
components**; but **156 shared section components exist, 40 with zero live instances and 61
used on one site only.** The library is not the bottleneck — **selection is**. designblog serves
**0** content images across 50 slots (controls: garden-tools 4/43, boxingonline 5/48, so the
query detects images).

⚠ I told them `page_archetypes` is necessary and probably **not sufficient**, because
`defaultSectionsForPage` is only the **fallback** when the planner returns no sections. **The
cheap measurement nobody has done: what fraction of live pages got their sections from the
planner versus the fallback.** That decides whether the lever is the table or the planner prompt.

---

## 0a. HISTORICAL — post-roll verification, 2026-09-02 evening

> Records the per-pod binary probe (both pods, both controls) and the finding that the defect
> **survives** the v1.0.1355 roll. The "§0 items 1 and 2" it refers to were the blocked-on-token
> checklist now in §0b. Still current on one point: **probe every pod, with controls, never
> `merge-base`.**


**The defect SURVIVES v1.0.1355.** A `template_changed` rerender of boxingonline `/index.html`
filed 21:21:06Z, completed 21:22:43Z, **wrote** (`page_components.updated_at` 21:22:14) and
produced the **old shape**: no `excerpt` key, title still suffixed. So the four other commits in
that roll touching `rerender_page_sections_action.go` / `plan_sections_action.go` /
`queryresolve/` (444, 443, 137-residue, 427) did **not** fix it. That was a real measurement, not
a confirmation — the `site_delivery_and_editor` lane corrected its own "should not change the
answer" to an inference from commit subjects before I ran it.

**§0 item 1 done, and it closes the hole I could not close before.** Both pods on the current
build, both controls:

| pod (started) | `ListItemExcerpt` | `resolvePagesWhereType` (+control) | invented symbol (−control) |
|---|---|---|---|
| `8ddbf8958-cd2h9` (20:56:43Z) | PRESENT | PRESENT | absent |
| `8ddbf8958-vppjz` (20:57:10Z) | PRESENT | PRESENT | absent |

**So the current finding is not a pod-variance artefact**: two pods, both carrying the fix, and a
rerender still produces the old shape. (This does not retroactively close the 13:5x-era
one-pod-of-two hole — those pods are gone — but it makes that hole irrelevant to the live defect.)

**Next: the ⭐ experiment in §2.** It is now the only cheap thing left that discriminates.

## 0g. HISTORICAL — an earlier session-end snapshot, with STALE queue figures

> ⚠ **Superseded twice.** Its batch list predates `691`, and its queue arithmetic (192 triaged,
> "hours") was computed with a **proxy census that ignored the selector's lock clause** — see the
> corrected ranking in the same section's later note and in §0o. Kept for the reasoning, not the
> numbers.


| experiment | batch | reason | state |
|---|---|---|---|
| **deck discriminator** — `guides-index`, new-shape baseline | `…000688` | `template_changed` | **triaged**, filed 08:26 |
| **hero test** — `remortgagecalculator.uk/about`, unpinned | `…000690` | `image_landed` | **triaged**, filed 08:55 |
| ~~advertise.co.uk/about~~ | `…000689` | — | **cancelled for a WRONG reason** (see §0f); the row carries a dated correction |

`[MEASURED 2026-09-03 ~09:0xZ]` **234 `page_rerender` triaged, and only 9 completed in the last 30
minutes.** Earlier in the session it was 192 triaged draining ~29 per 30 minutes. **The backlog is
growing faster than it drains**, so neither experiment is hours away — it may be considerably
longer, and that is worth knowing before anyone re-files thinking the item was lost. **Both items
are queued, not stuck; find them by batch id, do not duplicate them.**

> **THE WATCHER FIX WORKED, which is worth one line.** After the first watcher timed out and
> printed the unchanged baseline under the heading "DISCRIMINATOR RESULT", I added the item's
> terminal status to the same output. The second watcher expired too — and printed
> `item outcome: triaged` beside the reading, so the output labels itself as "did not run" rather
> than presenting a baseline as a finding. **Any query that reads an artefact to judge a dispatch
> should print the dispatch's status in the same breath.**

## 0b. The block that was here (RESOLVED — kept for the recipe)

`kubectl` returns **`You must be logged in to the server (Unauthorized)`** fleet-wide
(checked `get ns` and `get pods -n ai-persona-system`, both fail). That is the 3-day
kubeconfig token expiry; **the owner refreshes it.**

**A fresh chassis build was deployed at ~21:00Z.** `[RELAYED, not verified by me]` the
`site_delivery_and_editor` lane reports chassis pods `8ddbf8958-cd2h9`/`vppjz` started
20:56:43/20:57:10Z on **v1.0.1355**, provenance **`0d2feee2f`** read from the pod by that lane's
previous session, and that `f57f5ad1f` is an ancestor of both v1.0.1354 and v1.0.1355 — so the
§2 A/B already had the producer fix aboard and v1.0.1355 should not change its answer. **I could
not verify any of that**: the token expired ~21:08Z, before I could probe. Treat it as a peer's
reading until item 1 below is run. The first three things to do when the token is back:

1. **Probe EVERY pod, with controls** — not one, because two pods of one ReplicaSet are not
   guaranteed identical (same-tag cached image), and probing one of two left a hole in this
   bug's evidence that could not be closed later because the unprobed pod was gone:
   ```bash
   for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
     for sym in ListItemExcerpt resolvePagesWhereType ListItemTitleXYZNOTREAL; do
       kubectl -n ai-persona-system exec "$pod" -- grep -aqs -- "$sym" /proc/1/exe \
         && echo "$pod $sym PRESENT" || echo "$pod $sym absent"
     done
   done
   ```
   `ListItemExcerpt` must be PRESENT; `resolvePagesWhereType` is the positive control (a
   probe that finds nothing must not read as an answer); the invented symbol must be absent,
   and it *contains* `ListItemTitle`, so its absence also proves the grep is not matching
   loosely. **Never `git merge-base` for this** — a same-tag rebuild serves a cached image
   while merge-base still says yes.
2. **Re-run the §2 discriminator.** The new build may have changed the answer.
3. **Check `721`'s effect** (§3) — it was applied before the roll and needs a re-render to show.

---

## 6. Files

`bugs_open/425_HANDOFF_2026-09-02_a_card_renders_four_empty_slots_because_two_producers_of_one_item_shape_disagree.md`
— the full case, with **every superseded reading left visible**. The corrections are the record.

Migrations: `682` (+ROLLBACK) applied · `683_..._HOLD` (+ROLLBACK) applied, batch
`…000683`, 10 complete / 4 cancelled · `721` (+ROLLBACK) applied.
Scripts: `scripts/check_card_slot_guards.py`, `scripts/check_list_empty_states.py`,
`scripts/component_template_lib.py`.
Register: **PBP-050**. Landmine + `WRONG_CALLS` entries committed.
