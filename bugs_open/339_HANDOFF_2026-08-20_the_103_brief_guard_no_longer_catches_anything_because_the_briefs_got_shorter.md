# 339 — `bugs_closed/103`'s brief guard catches nothing any more, because the briefs got shorter

**Filed** 2026-08-20 by the `meta_description_never_backfilled` lane. **Status: OPEN.**
A live recurrence of a closed bug, in its own guard's blind spot.

> **Resolve by SLUG** (`103_brief_guard_catches_nothing`) — bug numbers collide on this
> tree, and `git log` the FILE PATH, not the number.

---

## 1. What a reader needs first

`bugs_closed/103` was this: **tool pages published their internal build brief as the
public meta description**, so Google printed generator instructions under the search
result. The worst live case was 1,206 characters of *"no fetch calls, no backend"* on
vonc.com's Arena page.

Its fix is `datahelpers.PublicMetaDescription(candidate, composed)`
(`platform/orchestration/datahelpers/meta_description.go`), a shared gate with **two
signals**:

1. **Length** — anything over `maxPublicMetaDescription = 320` chars reads as internal.
   Chosen because *"the shortest observed brief in the bugs_open/103 census was 449
   characters and the longest 1,206"*.
2. **`briefMarkers`** — a regex of phrases measured in that census:
   `no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)`.
   Its own comment says why it exists: *"length alone would miss a SHORT brief, which is
   the failure this guard would otherwise still allow through."*

**Both signals now miss, and the failure is back.**

## 2. The measurement

`[MEASURED 2026-08-20]` over `pages` where `status='active'` and the column is non-empty
(693 pages):

| | |
|---|---|
| descriptions **over 320** chars (signal 1 would fire) | **0** |
| descriptions in the **200-320** band | **11** |
| of those 11, matched by **`briefMarkers`** (signal 2) | **0** |
| page types | **9 `tool`**, 2 `blog-post` |

So the population that used to be 449-1,206 characters now sits at 200-320. **It moved
underneath the guard.** Signal 1 fires on nothing; signal 2 was the designed backstop for
exactly this and matches none of them.

## 3. What is actually being published

Verbatim from `pages.meta_description`, live today:

- `gamesdesign.co.uk` / `tool-wave-difficulty-ramp` (301 chars) —
  *"Companion to the Spawn Rate Balancer. Designers input player power growth per wave
  (DPS scaling, healing, item unlocks) against enemy health and spawn…"*
- `gamesdesign.co.uk` / `tool-drop-rate-tuner` (302) —
  *"Tune loot drop rates against player experience. Set a base drop chance,…"*
- `gamesdesign.co.uk` / `tool-probability-curve-visualiser` (280) —
  *"Lets designers plot and compare multiple probability distributions (uniform, binomial,
  geometric, custom weighted) on a single chart. Directly support…"*
- `gamesdesign.co.uk` / `tool-stat-budget-allocator` (272) —
  *"Lets designers define a total stat budget for a character or item tier,…"*
- `webdesign.co.uk` / `tool-css-unit-converter` (296) —
  *"Converts between px, rem, em, vw, and vh units given a base font size an…"*

These are specifications addressed to a builder. *"Lets designers…"*, *"Designers
input…"*, *"Companion to the Spawn Rate Balancer"* — third person, describing inputs and
outputs. Not one of them is written to a visitor, which is what a meta description is for.

Full list:
```sql
SELECT s.domain, p.name, p.page_type, length(p.meta_description) AS len, p.meta_description
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND length(COALESCE(p.meta_description,'')) BETWEEN 200 AND 320
ORDER BY len DESC;
```

## 4. Provenance — what is established and what is not

**Established:** the **planner did not write these.** `site_plan_pages` for the current
plan has **no rows** for those tool page names — tool pages are created by the tool-deploy
path, not the site plan. So this is not a consequence of migration `485` (which taught
`build-site-planner` to write descriptions) and 8 of the 11 predate `485` anyway.

**Established by elimination:** it is not the **composed** fallback either.
`composedToolMetaDescription` emits a fixed sentence — *"An interactive %s, free to run in
the browser. The companion guide sets out the method behind it, so you can check the
working."* — about 120 characters and nothing like these. `PublicMetaDescription` returns
either the candidate or the composed fallback, so by elimination **these are the CANDIDATE
side, passed by the guard.** That is 103's mechanism exactly: the candidate for a
`component_level='tool'` row is the brief.

**NOT established, and stated so rather than asserted:** I did not trace each string back
to its specific writer row. A `LIKE` join to `content_components.description` returned no
match, which is inconclusive rather than contrary — those rows may have been regenerated
or edited since the page was created. **Whoever fixes this should establish the exact
writer before changing the guard**, because the remedy differs depending on whether the
brief arrives from `content_components.description`, from a tool spec, or from somewhere
a census has not looked.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Re-derive the length threshold from the CURRENT population, not the 2026-07 census.**
   320 was chosen against briefs of 449-1,206 chars. Today's briefs are 272-302. The
   number is stale, not wrong-headed. ⚠ But it cannot simply drop to ~250: legitimate
   descriptions run to 177 chars today and the band is narrow, so a naive drop trades this
   failure for false refusals of good copy. **Measure the two populations before picking a
   number, and say what the disconfirming result would have looked like.**
2. **Strengthen `briefMarkers` from the CURRENT corpus.** It was measured once, in July,
   and has not been re-derived. *"Lets designers"*, *"Designers input"*, *"Converts
   between"*, *"Companion to the"*, *"Set a base"* — second-person-absent, imperative-to-a-
   builder constructions — are the shape now. This is the signal the guard's own author
   said was the backstop for short briefs; it just needs re-fitting.
3. **Stop the brief being the candidate at all.** The most durable fix: don't hand
   `content_components.description` to `PublicMetaDescription` as a candidate for a
   `component_level='tool'` row, because for those rows it is *by definition* the brief.
   The guard exists to catch a mistake that the call site could avoid making. This makes
   the bad state unrepresentable rather than detectable.
4. **Repair the 11 live rows.** Separate from the class fix, and note they are *tool*
   pages — the framework has a composer for exactly these
   (`composedToolMetaDescription`), so this is a re-derivation, not authoring, and does
   not need an LLM.

⚠ **Whatever the fix, re-run it over what the blind guard already cleared.** 693 live
descriptions passed a check that could not fire; a fixed guard should be swept back over
all of them, not just applied going forward. (MEMORY: *a PASS from a BLIND check outlives
the blindness*.)

## 6. How to verify a fix

```sql
-- must go to 0, and quote the denominator beside it
SELECT count(*) FILTER (WHERE length(meta_description) BETWEEN 200 AND 320) AS brief_band,
       count(*) AS total_with_desc
FROM pages WHERE status='active' AND COALESCE(meta_description,'')<>'';
```

**Induce both arms or the test proves nothing:**
1. A 280-char brief-shaped candidate must be **REFUSED**.
2. A legitimate 177-char human description (there are live examples) must still be
   **ACCEPTED**. A guard tested only on its refusing side is indistinguishable from one
   that refuses everything.

## 7. Why this is filed rather than fixed here

It is not this lane's defect. None of the 11 was written by the meta-description
backfiller (`SEO-004`) — that action *reuses* `MetaDescriptionLooksInternal` and its
descriptions run 65-177 chars. This was found while measuring the column for
`bugs_open/320` and is flagged rather than annexed: changing a shared guard used by two
tool-creation paths is a different blast radius from filling a blank column, and
re-deriving both of its signals needs the tool-page population's owner.

`scripts/who-owns.py 103` names the `gauntlet_dead_cta` lane, whose recent commits are
unrelated (gripper/SMTP), so it may need rehoming rather than routing.

## 8. Provenance of this file

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and declaring the substitution rather than omitting it:** every figure is a
direct query over `pages` (quoted in §2 and §6), the published strings are verbatim from
the column, the guard's two signals were read in source
(`datahelpers/meta_description.go`) and its `briefMarkers` regex was run against the
population to get the 0, and the planner was eliminated by querying `site_plan_pages`.
§4 marks explicitly what is **not** established. The claim is bounded — a guard's two
signals no longer match a population that moved — and asserts no cause outside the
symptom.

Related: `bugs_closed/103` (the original), `bugs_open/320` §13 (where this was found),
register **SEO-004**.
