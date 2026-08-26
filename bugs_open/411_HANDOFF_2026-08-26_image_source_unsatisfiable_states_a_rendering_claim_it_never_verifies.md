# 411 — `check_image_source_unsatisfiable` states a RENDERING claim it never verifies, and ~87% of its open queue is wrong about the thing it asserts

**Status: OPEN.** Diagnosed 2026-08-26, **CONFIRMED first iteration** by the diagnosis loop
(`RUN_CORRELATION_ID=8aeba0b6-0508-4059-8a10-b3e94211dd8c`), which independently re-read the same
functions and cited the same lines. Not fixed — see §5 for why the obvious fix is wrong.

**Found by:** the `vigilant_designer_offer_analysis` lane, after the `apis.uk` lane routed a single
item at it rather than cancelling it. **Neither the item nor the queue is mine**; this is filed for
whoever owns the discovery checks.

---

## 1. The defect in one sentence

The check decides **satisfiability** correctly, then emits a **reason about rendering** that it has
no code path to verify — and for most of its open queue that reason is false.

## 2. What it emits

`check_image_source_unsatisfiable.go`, `ImageSourceUnsatisfiableCheck.Run`, writes into every item's
`spec.reason`:

> `"no asset key, plan imagery row, or image-role alias can supply this source; the field renders empty or falls back to a placeholder"`

**The first clause is true.** The second is a different claim, about what the page actually shows,
and nothing in the check establishes it.

## 3. Why it cannot be true in general

The check's satisfiability test reads `literalKeys`, `pageHero`/`siteHero`/`siteLogo`,
`imageryplan.ImageRoleForPath`, and ONE site-wide `sites.content_data` `hero_url`/`logo_url` lookup
keyed on `site_id` alone.

`plan_sections_action.go`'s resolver has a **later** arm the check knows nothing about:
`carryStored` (`bugs_closed/238`) satisfies any **non-llm** field from **the page's own deployed
`content_data`** when the declared source resolves nothing — a per-component lookup by
`(slot, field)` via `storedFieldValue`.

**The two are separate functions and neither names the other.** So a field can be genuinely
unsatisfiable *and* render perfectly, and the check will assert it renders empty. **It reads
SCHEMAS; the truth of its own sentence lives in VALUES.**

## 4. Measured blast radius

`[MEASURED 2026-08-26]` open items at `needs_human_review`, joined on site + page name +
`content_components.function` + the named field. **Loose and strict joins agree.**

| source | open items | field IS populated on that page |
|---|---|---|
| `site_assets.hero` | 46 | **46** |
| `site_assets.image` | 12 | 8 |
| `site_assets.illustration` | 9 | 4 |
| **TOTAL** | **67** | **58 (87%)** |

**Every one of the 46 `site_assets.hero` items is wrong about rendering.** The diagnosis independently
surfaced one: `hero` on page `guide-how-loans-are-calculated`, field `background_image`.

⚠ **This is NOT caused by migration `644`.** That change contributed exactly ONE of the 58 (an
`illustration` row on apis.uk). The hero rows are the bulk and long predate it. 644 is only how it
was noticed.

## 5. ⚠ THE OBVIOUS FIX IS WRONG — do not "treat carry as satisfied"

Silencing the carried case would hide the thing most worth seeing. **Two states are being
conflated and they want separating, not merging:**

- **(a) unsatisfiable AND nothing carried** — renders empty. The real `bugs_closed/238` defect. The
  reason text is TRUE here. `[MEASURED 2026-08-26]` **9 of 67**.
- **(b) unsatisfiable BUT carried** — renders correctly **today**, and is fragile: the value survives
  only as long as a deployed row holds it. **A NEW page with the same component on the same site
  renders nothing.** This is a *supply* warning wearing a *rendering defect's* words. **58 of 67**.

Suppressing (b) would also erase the estate's clearest signal of the imagery supply gap
(`[MEASURED 2026-08-26]` 26 illustration assets across 5 sites against 206 heroes across 28).

**Fix candidates, ordered by what closes the door:**

1. **Split the verdict at emission** — have the check consult the page's stored `content_data` for
   the named field, and emit an accurate reason plus a distinguishing marker (`carried` vs `empty`)
   and severity. Makes the bad state unrepresentable: the reason can no longer be false. ⚠ Cost: the
   check gains a per-component stored-content read it does not have today. **Reuse `storedFieldValue`
   / the `carryStored` predicate rather than writing a second one** — two spellings of "what does
   this page already hold" is exactly the drift class this estate keeps filing.
2. **Weaken the reason text only** — drop the rendering clause, keep the satisfiability claim. Cheap,
   honest, and loses the (a)/(b) distinction that makes the queue actionable.
3. ~~Treat carry as satisfied and suppress~~ — **rejected above.** It would close 58 rows by making
   the check blind to the supply gap, and (b) is genuinely fragile.

## 6. How to verify a fix

The demand control matters: this check runs on a discovery pass, so **a drop in new items is not
evidence until a pass has actually run.**

```sql
-- (a) vs (b) split — should become readable off the item itself after a fix
WITH items AS (
  SELECT swi.id, swi.site_id, swi.spec->>'page' AS page,
         swi.spec->>'component_function' AS fn, swi.spec->>'field' AS field
    FROM site_work_items swi
   WHERE swi.item_type='image_source_unsatisfiable' AND swi.status='needs_human_review')
SELECT count(*) AS open_items,
       count(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM pages p JOIN page_components pc ON pc.page_id=p.id
           JOIN content_components c ON c.id=pc.component_id
          WHERE p.site_id=i.site_id AND p.name=i.page AND c.function=i.fn
            AND COALESCE(pc.content_data->>i.field,'') <> '')) AS renders_fine
  FROM items i;
```

⚠ **The 67 existing rows will not re-describe themselves.** A fix changes emission; the open queue
keeps its false reasons until those items are re-detected or closed. Say which, or the queue reads
as fixed while 58 wrong sentences sit in it.

## 7. Interactions worth knowing before touching this

- ⚠ **`bugs_open/033`** — the human-review queue *has no working surface*. These 67 rows are in a
  queue nobody reads, which is why an 87% error rate went unnoticed. **Fixing the reason without
  fixing the surface changes nothing a human sees.**
- ⚠ **`bugs_open/356`** — a sibling defect in the same check family (discovery checks selecting on
  the build axis, so retired pages are filed as work). Names this check only in passing; **different
  axis, do not merge them.**
- **Register `IMG-074` / LANDMINES** — the `site_assets.image` → `hero` alias trap, which is why the
  `site_assets.hero` and `site_assets.image` rows are so populated in the first place.
- **`bugs_closed/238`** — `carryStored` itself, the arm the check cannot see.

## 8. What was NOT done, and why

**The checker is untouched.** It is a shared Go discovery check: council scope, inert until a roll,
and *what the check should mean* is a decision rather than a patch — (b) is a real warning and
someone must choose whether it stays visible. **The apis.uk item is deliberately left open as
evidence; do not cancel it.**

Prior art checked before filing, by mechanism AND by field/table (`bugs_open/033`, `356`,
`bugs_closed/238`, and a grep for `image_source_unsatisfiable` across both directories): nothing
covers the carry blind spot.
