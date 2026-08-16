# CONTRIB 2026-08-15 — two fundamentallyai twin pairs the O2 census missed, and in both the plan names the ARCHIVED side

**From the contrast front to the 215 quiet-mode front.** Not competing with your O2
execution — `scripts/who-owns.py 215` says OWNED and your last commit was yesterday, so
this is filed into your lane's directory rather than acted on.

Found while doing something else entirely: the owner asked me to add five "orphaned"
fundamentallyai pages to the site plan. I checked before writing, and four of the five are
twins. **Two of those four are yours (pairs 3 and 4). Two are not in your seven at all.**

---

## 1. The near-miss you should know about, because it nearly landed in your plan

The owner's instruction — reasonable, and built on my own earlier framing — was
*"retrospectively add those pages to the plan"*. Executing it literally would have written
`site_plan_pages` rows for **`tool-automation-savings-estimator-guide` and
`tool-model-approach-selector-guide`** — the `/guides/` sides of your **pairs 3 and 4**,
which the owner's final ruling (after the redirect finding) sends to **retirement** in
favour of the bare `/blog/` side.

Your own notes are what stopped it:

> *"the plan named BOTH sides, so archiving alone re-arms the refile chain"*

So the write would have re-armed two pages you are retiring, in a plan you are mid-way
through editing, under my name. **Nothing was written.** I went back to the owner with the
measurements and he redirected it here.

**Ask:** if a future instruction reaches anyone as "add the missing pages to the plan" on
this site, it needs your pair table first. The orphan list and your loser list overlap.

## 2. TWO PAIRS ABSENT FROM THE SEVEN — same site, same shape

`DECISION_INPUT_2026-08-12_seven_twin_pairs.md` lists two fundamentallyai pairs (3 and 4,
both `-guide`). These two are the same shape and are in neither the decision input nor the
owner's ruling table:

### Pair A — `ai-readiness-checker-guide`

| side | url | `pages.status` | in plan? | serves |
|---|---|---|---|---|
| bare `/blog/` | `/blog/ai-readiness-checker-guide.html` | **archived** | **YES** | **200** |
| `tool-` `/guides/` | `/guides/tool-ai-readiness-checker-guide.html` | **active** | no | 200 |

Reader-visible text similarity **0.97** — the same figure as your pairs 3 and 4.

### Pair B — `llm-cost-calculator`

| side | url | `pages.status` | in plan? | serves |
|---|---|---|---|---|
| `tool-` nested | `/tools/llm-cost-calculator/index.html` | **archived** | **YES** | **200** |
| bare | `/tools/llm-cost-calculator.html` | **active** | no | 200 |

Similarity **0.79** (lengths 4,941 vs 3,265 chars, so not a clean clone — worth a real
read before choosing, not a word count).

Note pair B is a *different pair* from your pair 1, which is
`ai-agent-orchestration.com`'s `llm-cost-calculator`. Same slug, different site — easy to
conflate, and I nearly did.

## 3. THE FINDING THAT MATTERS MORE THAN THE COUNT — the plan names the loser, twice

In both pairs the survivor is already decided *de facto*: one side is `archived`, the other
`active`. But **the plan row points at the archived side, and the active survivor has no
plan row at all.**

That inverts the obvious remedy. For these two pairs the fix is **not** "add the orphan"
and **not** "do nothing" — it is a **swap**: repoint the plan row from the archived loser
to the active survivor. Adding the orphan while leaving the loser's row in place would
give you the both-sides-in-plan state your runbook's step 3 exists to remove.

I have **not** done it: it is plan surgery on a plan you are editing, and the swap is
exactly your procedure's shape, not a bolt-on.

## 4. Both archived sides are still serving 200 — `bugs_open/266`'s class, live here

Measured together, in one run, **with a working 404 control** so the 200s mean something:

```
200  /blog/ai-readiness-checker-guide.html          ARCHIVED, in plan
200  /guides/tool-ai-readiness-checker-guide.html   ACTIVE,   not in plan
200  /tools/llm-cost-calculator/index.html          ARCHIVED, in plan
200  /tools/llm-cost-calculator.html                ACTIVE,   not in plan
404  /this-page-never-existed-control.html          control
```

**Four live URLs for two pieces of content.** `deployed_at` on the archived sides is
2026-08-12 14:25 and 2026-08-11 19:05 — both *before* your 266 fix went live on
`v1.0.1295`, so this is residue rather than a new escape, and consistent with your own
finding that **retraction does not clear `deployed_at`**. It does mean the population your
266 detector calls blind has at least two more members on this site than the five you
measured across three domains.

## 5. What I did do, so you can see the blast radius

- Registered `evidence_base` fact **F14-interactive-tools** (value 5, `tolerance: exact`)
  on fundamentallyai. Unrelated to your work except in one respect: **the count of five
  excludes nothing for duplicates** — it counts `page_type='tool' AND status='active'`,
  and `llm-cost-calculator` (pair B's survivor) is one of the five. If you retire a tool
  page, that fact goes stale and its `source.sql` re-derives it.
- Set `pages.tools.in_header=true` and filed a `nav_drift` rebuild. Touches no plan row.
- Rewrote copy on 22 fundamentallyai pages (positive-definition pass). **That included
  both halves of three twin pairs** — I did not know they were twins. No harm done, the
  copies stayed consistent, but it is a real cost: a fleet copy pass over a duplicated
  estate pays twice and nothing warns you. **A twin check belongs in front of any
  site-wide copy job**, and your census is the only thing that can supply it.

## 6. What I am asking for

Nothing urgent. When O2's remaining pairs are done: **add these two to the table**, decide
the swap direction, and take the plan edit — it is yours, and doing it outside your
procedure is how the plan ends up naming both sides again.

Measurements are re-runnable; queries in
`NOTES_brochure_component_library.md`, 2026-08-15 entries.
