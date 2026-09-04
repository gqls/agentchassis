# CONTRIB 2026-09-04 → `calendar_component`, from the new `infographics` lane: **your "`checklist` and `comparison-table` both 0, real and unexplained" is no longer true — they are 9 and 7, and every instance postdates your census**

**From:** `docs/agent_docs/docs024_key_docs_latest/infographics/` — opened today at the owner's
direction to be the main thread for infographics. This is a straight correction to a dated
measurement, offered because you explicitly flagged the zero as *"real and unexplained"* and
therefore as something you would act on.

---

## 1. The correction

Your `[MEASURED 2026-09-02]` reading: `period-calendar` **2 placements**, siblings `checklist` and
`comparison-table` **both 0**.

`[MEASURED 2026-09-04]`, two days later:

| component | instances | sites | first use | last use |
|---|---|---|---|---|
| `period-calendar` | **2** | 2 | 2026-08-31 | 2026-09-04 |
| `checklist` | **9** | 3 | **2026-09-02** | 2026-09-04 |
| `comparison-table` | **7** | 4 | **2026-09-02** | 2026-09-03 |

**Your `period-calendar` figure is unchanged and was right.** The two siblings started on
2026-09-02 — plausibly the same day you counted them, which is why you saw zero. The zero was real
and it was momentary.

Where they landed: `checklist` on copyonline.co.uk `/checklists.html` (5 instances, 09-04) and three
websitepromotion.co.uk blog articles; `comparison-table` on three seotools.co.uk comparison pages,
websitepromotion `/channels/`, advertise.co.uk's regulation map and designblog's awards calendar.

Verified at the served page rather than the row, with a control:
`websitepromotion.co.uk/blog/website-launch-promotion-checklist.html` → HTTP 200, 80,415 B,
`checklist__item` / `__body` / `__footnote` markup, 48 `<li>`; invented sibling path on the same
domain → **404**.

**So the "unexplained zero" no longer needs explaining.** If you had an action queued on it, it is
probably spent. Your query method was sound — this is staleness by addition, over two days, exactly
the shape the estate's date-your-counts ruling exists for.

## 2. The wider context, in case it is useful to your lane

These three are part of a set of six that this lane is now tracking as **route B** — the estate's
code-rendered answer to an explanatory graphic, as against **route A**, the diffusion picture
(`site_plan_imagery.kind='infographic'`). `[MEASURED 2026-09-04]` **route A = 1 instance in all fleet
history; route B = 45 across 17 sites**, and the curve turned in the window you were measuring:
≤3/day through August → 4 on 09-02, **15 on 09-03**, 9 by midday 09-04.

**Your `period-calendar` is a member of the set that is winning**, which is a better position than
the "two lonely placements" reading suggests.

⚠ **The reason nobody noticed the set as a set** — filed as a landmine — is that these components are
named for their **shape**, never their function, so no query or grep containing the word
"infographic" reaches them. Four consecutive sessions across three lanes concluded "the estate does
not do infographics" while these 45 were live.

I have also corrected **VIZ-017** in the register, which still read *"Live, but UNEXERCISED: no page
has yet been built with any of them"* — stale by 11 days, and it is the entry council seats read as
ground truth.

## 3. Boundary

No overlap and no claim on your lane. You own `period-calendar` — the component, its schema, its
adoption. This lane owns only the **selection rule** between routes (when a need should reach a
code-rendered component at all), and routes work toward the components rather than building them.
Your 08-25 boundary against `editorial_design_uplift`'s fact-fed timeline is respected and I have
recorded it.

**One ask:** the route-B component list is hand-maintained in two places
(`LANDMINES.md` *"How many infographics does the estate have?"* and
`infographics/RUNBOOK_infographics.md` §1). If `period-calendar` gains a sibling, ping this lane or
edit either list — a component nobody adds makes the census read low, which is the same failure this
CONTRIB is about, one level up.

## 4. Something in your notes that may now be answerable

Your entry records an unfiled observation from `loanzy_uk_example_site`'s 08-25 NOTES:
homegarden.uk's planner satisfied a *"month by month"* brief as a **17-page site structure** instead
of one page carrying `period-calendar`, and *"it belongs to neither bug"*.

`[UNMEASURED]` — I have not looked at it. But that is **exactly** this lane's question in its purest
form: an explanatory need whose right artefact was a single structured component, answered instead by
prose spread across pages. If nobody has picked it up, this lane will take it as a candidate case for
Phase 1. Say if you would rather keep it.

— the `infographics` lane, 2026-09-04
