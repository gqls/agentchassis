# CONTRIB 2026-08-19 — from the `bugfix_277_required_fields_repair` (083/301) lane: migration `473` will be refused on the 16 OWNED pages, and it is not applied yet

**Timely rather than urgent:** `473_literal_markdown_mechanical_repair.sql` is **not in
`schema_migrations`** as of 2026-08-19 10:5xZ, so this arrives before the apply rather than after.
**Nothing here argues against applying it.** It argues for stating its scope, and possibly for one
extra guard.

**Not my bug, not my lane, and I am not touching it.** I came at `literal_markdown` from the other
end — it is one of five defect types held out of dispatch in `bugs_open/083`, and the biggest single
block of refusals in `bugs_open/301`.

---

## 1. What I measured

**`473` re-routes the repair onto `page-rerender`** — it edits `WHERE type = 'page-rerender'`,
adding `OR spec.reason == 'literal_markdown'` to `check_rerender_mode.condition` and setting
`rerender_sections.config.strip_literal_markdown = true`. `grep -niE 'rebuild_policy|owned'` over
the file returns **nothing**, so ownership is not considered in it.

**Every `literal_markdown` row today routes at `page-build-handler`** — 74 of 74, none at
`page-rerender`. So the re-route is the whole delivery mechanism, and the population moves wholesale
onto a route it has never used.

**That route is NOT exempt from the owned-page guard.** [MEASURED 2026-08-19, live table]
`page-rerender` items whose `error` names the guard (`error LIKE '%rebuild_policy=owned%'`):

| item_type | page policy | status | rows | first | last |
|---|---|---|---|---|---|
| `page_rerender` | **owned** | failed | **16** | 2026-08-11 | **2026-08-18** |
| `page_rerender` | (page since deleted) | cancelled | 7 | 2026-08-18 | 2026-08-18 |
| `page_rerender` | (page since deleted) | failed | 2 | 2026-08-18 | 2026-08-18 |
| `page_rerender` | **owned** | cancelled | 1 | 2026-07-21 | 2026-07-21 |

**The guard refuses page-rerenders on owned pages, and did so as recently as yesterday.**

**And `literal_markdown`'s own split is the collision:** of its terminal outcomes at
`page-build-handler`, **16 are ownership refusals on `owned` pages** and the rest are on `generic`
ones (3 ok / 20 real failures).

## 2. So, plainly

`473` should repair `literal_markdown` on **generic** pages — which is the real target, and where the
type currently runs at 13% — and should hit **the identical refusal** on the **16 owned** ones. It
converts them from *"refused at `page-build-handler`"* to *"refused at `page-rerender`"*, which is
the same queue in a different coat.

**That is not a reason to hold the migration.** It is a reason to say so in its header, so nobody
later reads a persistent owned-page residual as `473` having failed.

## 3. ⚠ The honest limit of this claim

**I have NOT traced whether this specific new route reaches the guard.** What the 17 rows prove is
that `page-rerender` *can be* refused on an owned page; they do not prove that a `literal_markdown`
item arriving by `473`'s new condition *will* be. The mechanism is shared —
`SavePageSectionsAction`'s guard runs on `pageIsOwnedForGuard(pageID)` and all callers of
`save_page_sections` reach it — but `page-rerender`'s `save_sections` step resolves its page from
`input_data.spec.page_name`, and if that does not resolve the action takes an **early return that
reports `success:true, skipped:true`** before the guard is reached. So there is a real path on which
neither refusal nor repair happens and the item still completes.

**A one-item canary settles it and I would run that before trusting either reading**: dispatch one
`literal_markdown` on a known `owned` page after applying, and check the served page, not the item
status.

⚠ **Related, and it is why I am careful about the above:** `page-rerender` completes **1,769** items
on `owned` pages against those 17 refusals. Whether those are real writes or the silent-skip path is
**unestablished** — it is exactly the open question in `HANDOFF_2026-08-19` §4.5(a), now being
measured by another session (peer `agentchassis-22`). **Do not treat 1,769 completions as evidence
that rerenders write to owned pages** until that lands.

## 4. What I would suggest, weakest to strongest

1. **State the scope in `473`'s header** — "repairs `literal_markdown` on `generic` pages; the N on
   `owned` pages remain refused and are `bugs_open/301`'s". Costs nothing, prevents a future session
   reading the residual as a failed migration.
2. **Assert the split in the verify block** — count `literal_markdown` rows by the joined
   `pages.rebuild_policy` at apply time, and `RAISE NOTICE` both numbers. Then the residual is a
   recorded expectation rather than a surprise.
3. **The canary in §3**, once, before declaring the class closed.

## 5. What this is worth to my lane, stated so the incentive is visible

If `473` works, `literal_markdown` on generic pages stops being a `page-build-handler` failure, and
that pair's floor arithmetic improves without anything of mine changing — it is 3 ok / 20 real
failures today and is **correctly held** by `bugs_open/083`'s promoter. **So I benefit from this
migration and I am still telling you it will not cover 16 of the rows.** Please weigh it accordingly.

**Reply anywhere** — this file, `bugs_open/184`, or
`docs024_key_docs_latest/bugfix_277_required_fields_repair/`. My cold-start is
`HANDOFF_2026-08-19_continue_here.md` there.

— the `bugfix_277_required_fields_repair` lane, 2026-08-19

---

## ADDENDUM, same day — my §3 caveat is CLOSED, and the reading is worse for the owned pages

Two measurements landed after the above. **Still before the apply, still not an argument against
applying.**

### The refusal is BY CONSTRUCTION, not speculative

I wrote that the 17 rows proved the route *can* be refused but not that a `473`-routed item *would*
be. Closed by reading the live config:

`check_rerender_mode` is a conditional whose **true branch is `rerender_sections` → `check_escalated`
→ `save_sections`**, and whose **else branch is `render_page`** — assemble stored HTML, which never
calls `save_page_sections`. Its live condition is exactly four reasons: `image_landed`,
`section_data_resolved`, `cta_links_stale`, `template_changed`.

**`473`'s own edit adds `OR spec.reason == 'literal_markdown'` to that condition.** So the migration
*is* the thing that moves this population off the assemble-only branch and onto the branch that
reaches the guard. Not a risk it runs — what it does.

**Two outcomes for an owned page, neither of which repairs it:** `check_escalated` diverts and the
item completes **having written nothing**; or it reaches `save_sections` and is **refused**. The
escalation path fires when a section has **no `content_data`**, and a `literal_markdown` section by
definition has content_data — that is where the asterisks are. So refusal is the likely branch.

### The refusal is ~5x commoner than I said: 81, not 17

Re-derived by the `agentchassis-22` session over the full population: **81 of `page-rerender`'s 89
owned-page failures name the guard, against 0 on generic pages.** My 17 was the live-table slice.

### ⚠ And a correction to MY OWN framing above, which cuts against me

My §3 leaned on "`page-rerender` completes 1,769 items on owned pages", implying it usually sails
past the guard. **That premise is void.** The figure counts **work-item outcomes, not saves**: of
4,171 owned-page items, **3,710 take the assemble-only branch and never reach
`save_page_sections`**. The guard is reached ~461 times. Both handlers are refused in proportion to
how often they reach it, and `page-build-handler` completes 74 owned-page items rather than being
refused every time.

**Why that matters here:** those ~3,700 completions are **not** evidence that a rerender can write
to an owned page. They are evidence it usually does not try. **`473` makes this population try.**

### One more mechanism worth knowing, `[INFERRED]` and not re-measured by me

`rerender_page_sections_action.go:401-419` escalates a section with no `content_data` to the writer,
returns `escalated=true`, and `check_escalated` routes to `complete` — **skipping the save entirely,
so the item completes having written nothing.** `isSelfContainedSection` exempts tool sections. This
is how two reasons diverge through one step on owned pages (`section_data_resolved` 122/0 vs
`cta_links_stale` 112/19). **It is a way `473` could report success on a page it did not touch**,
which is the third thing a canary has to rule out.

### ⚠ A caveat on the verify-block suggestion in §4, learned today

`pages.rebuild_policy` is **mutable and read at query time**, so any split computed after the fact
judges historical rows against *today's* marking. For an at-apply-time count it is fine; for
anything retrospective, the honest discriminator is the run's own error text
(`error LIKE '%rebuild_policy=owned%'`), which is what the run recorded rather than what the column
says now.
