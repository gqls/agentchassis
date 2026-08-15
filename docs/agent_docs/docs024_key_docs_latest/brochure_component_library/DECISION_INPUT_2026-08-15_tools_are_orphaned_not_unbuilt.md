# DECISION INPUT 2026-08-15 — the tools are ORPHANED, not unbuilt, and the council was right to block the plan that assumed otherwise

**For the owner.** Raised by the contrast front while executing item 2b of
`HANDOFF_2026-08-12_contrast_front_continue_here.md`. The experience-planner run
**escalated** (correlation `cf8923ab-2d5a-462b-89eb-0e97c72d1bea`, `complete_escalated`,
5 revise rounds), and per `RUNBOOK_experience_loop.md` an escalation **is** the
round-boundary decision menu — surface it and stop the phase. This is that surfacing.

The escalation turned out to be worth far more than the plan would have been.

---

## 1. What you asked for

> *"we need links to the tools from the platform log and elsewhere, and a tools entry in
> the top nav would probably be nice."*

## 2. What the planner proposed, and why the council refused it

The plan proposed building a **new tools-hub row on the index page**, reusing the
existing `portfolio-showcase` component to render tool cards.

The `contracts` seat blocked it, twice, on the ground that `portfolio-showcase` is an
**existing deployed component** whose `html_template` was not in the plan's context — so
the claimed `title` → anchor-text and `live_url` → href mapping was *"unproven by
source"*, and a field-presence guard might blank the new cards. It refused to approve a
contract claim on reasoning alone and **named the two queries that would settle it.**

**I ran its two queries. The seat was right to block, and the answers are decisive:**

| its worry | answer |
|---|---|
| a guard could blank the whole card | **REFUTED.** Every guard (`{{if .logo_url}}`, `.live_url`, `.domain`, `.built_with`, `.build_time`) wraps only its own fragment. No card-level guard exists; a card with blank optionals still renders. |
| `title` → anchor text | **REFUTED, and worse than it thought.** `.title` renders inside an `<h3>` and is **not a link at all**. The only anchor is a hard-coded literal **"Visit Site →"**. The plan's acceptance criterion expecting the title as anchor text could never have passed. |
| stored titles unverified | **CONFIRMED.** The portfolio spec holds exactly three projects — Relojistas, idea.uk, Leopardess. There is **no** LLM-Cost-Calculator or Review-Council-Simulator entry to match against. |

And one the seat could not see from its context: `portfolio-showcase` renders
`target="_blank" rel="noopener"` — it is a component for linking **out** to other
people's sites. It would have opened your own tools in a new tab, described by a button
saying "Visit Site".

**So the council prevented a real defect.** This is the second time in this lane a seat's
objection that looked procedural turned out to be load-bearing.

## 3. THE ACTUAL DIAGNOSIS — nothing needs building

The plan, and our own 08-12 handoff, both assumed the tools hub had to be created.
**It already exists and it already works.**

`[MEASURED 2026-08-15, at the served artefact and the live DB]`

- **`/tools.html` serves 200, 27,163 bytes.** It is `active`, `deployed`, and its
  `nav_label` is `Tools / Index`.
- The site **already uses `tool-cta`** — the purpose-built component for linking a page
  to its tool — on **six** pages. It is not missing machinery; it has the right
  machinery already installed.

> ### ⚠ CORRECTED 2026-08-15, before anything was changed — I named the wrong table
>
> This section first said *"the top nav is generated from `site_plan_pages WHERE
> in_header`"* and that the fix was five missing plan rows. **That is wrong.** The nav
> is built from **`pages.in_header` / `pages.in_footer`** — there is a dedicated index
> for it, `idx_pages_nav btree (site_id, in_header, nav_order) WHERE status='active'` —
> and materialised into `site_nav_items`.
>
> **How I got it wrong:** I queried `site_plan_pages`, found exactly six `in_header`
> rows, saw they matched the six items in the served nav, and concluded the plan was the
> source. Both tables carry the same flags for those pages, so the match was a
> **coincidence, not a mechanism.** One matching count is not a causal chain, and I
> should have read the builder before asserting it.
>
> **What caught it:** reading a completed `nav_drift` item, whose own `fix` string says
> *"rebuild `site_nav_items` **from pages**"*. The correction makes the fix smaller and
> removes the collision risk in §5 entirely.
>
> The plan-row absence below is still factually true, and may matter to something else,
> but it is **not** why the nav is missing a Tools entry. Do not act on it.

**The defect is two boolean fields on one row.**

`[MEASURED 2026-08-15]` at `pages`:

| page | `in_header` | `in_footer` | `nav_order` | `nav_label` |
|---|---|---|---|---|
| `tools` | **false** | **false** | 203 | `Tools / Index` |
| `llm-cost-calculator` | **true** | false | 201 | `Tools / LLM Provider Cost Comparison Calculator` |

**`/tools.html` has both flags false, so it appears nowhere.** That is the entire nav
gap, and it is the answer to your first ask.

And the second row is its own small finding: `llm-cost-calculator` is declared
`in_header=true` and yet materialised into the **footer** group only. The stored nav
disagrees with the declared flags — nav drift in the literal sense, and the reason your
one existing tools link is in the footer.

`site_nav_items` holds twelve rows in two groups of six, which is exactly what the site
serves:

```
header:  Home · About · Contact · Platform Log · News · Capabilities
footer:  Review Council · Fine-Tuning · Backend Engineering · Asset Recovery ·
         Private Search · Llm Cost Calculator
```

**And that confirms the capitalisation defect's cause.** The stored label is
`Llm Cost Calculator` while that page's `nav_label` reads
`Tools / LLM Provider Cost Comparison Calculator`. So the builder derives the label by
title-casing the page **name** (`llm-cost-calculator`) and **ignores `nav_label`
entirely**. Useful consequence: for `tools` the derived label would be **"Tools"**,
which is exactly what you want — so this needs no label plumbing.

Separately and still true: five of twenty-five active pages have no `site_plan_pages`
row (`tools`, `llm-cost-calculator`, and three guide pages). Two of them carry
`Tools / …` hierarchical labels, so a Tools section was designed and labelled as a
section. That is worth understanding, but it is not the nav bug.

## 4. Why this changes the size of the job

The blocked plan was going to build a new hub, wire new cards, and add portfolio entries.
Against the real diagnosis, ask 1 is **five plan rows and a label correction**, and ask 2
uses a component the site already runs on six pages.

It also means the acceptance test should be *"the tools section is reachable"*, not
*"a new hub renders"* — and a check that a new hub renders would have passed while the
existing orphan stayed just as unreachable.

## 5. WHERE THIS IS BLOCKED — a permission, not a decision

> **SUPERSEDED 2026-08-15 by the correction in §3.** The three options that stood here
> were all about *which `site_plan_pages` rows to add*, and were built on the wrong
> mechanism. **The collision risk with the 215 front is void** — the real fix touches
> `pages`, not the plan, so the two threads never meet. You chose "file it for the
> framework"; that is still the right shape and it is what I attempted.

**The framework route needs two steps, and only the second is a work item.**

1. **Declare** the membership: `UPDATE pages SET in_header=true, nav_order=4,
   nav_label='Tools' WHERE name='tools'`. One row, two booleans.
2. **Materialise** it: a `nav_drift` item with `reason='nav_membership_declared'`, whose
   handler *"rebuilds `site_nav_items` from pages and re-renders chrome so the link
   ships"*. That item type is **23 of 23 complete**, most recently 2026-08-15 00:51 with
   a three-minute turnaround — it is the healthiest queue on the estate.

**Step 1 was refused by the harness permission classifier**, so neither has been applied.
This is an environment limit on my session, not a platform finding and not a decision
awaiting you.

**I deliberately did NOT file step 2 on its own.** The handler rebuilds the nav *from*
`pages`; with the flags still false it would rebuild the identical nav, complete
successfully, and report a fix that had not happened. A green work item over an unchanged
artefact is the exact failure this lane keeps documenting, and filing it would have
manufactured one.

**What I need from you:** either run the one `UPDATE` yourself (the full guarded SQL is
`SQL_2026-08-15_fundamentallyai_nav_membership.sql`, which also files the `nav_drift`
item in the same transaction), or grant the permission and I will run it.

**Still open regardless:** the footer link renders **"Llm Cost Calculator"**. §3 shows
why — the builder title-cases the page *name* and ignores `nav_label`. That is a generic
helper, so it is very likely a **fleet** defect rather than this site's label being
wrong, and it deserves its own item. I have not measured how many other sites carry an
acronym in a page name.

## 6. The open question I cannot answer from here

**Why are five deployed pages missing from the plan?** I have not diagnosed that, and I
am flagging it rather than guessing. If pages can reach `deployed` without a plan row,
then the plan is not the record of the site it is treated as — and the nav is only the
first thing that reads it. That may be a platform defect worth a `090` diagnosis run in
its own right, and it would explain orphans on other sites too. I have **not** checked
whether other sites have the same gap.

---

**Status of the rest of 2b:** the experience brief is written and is now the durable
record of your three asks (`doc_notes`, `subject_type='experience'`,
`subject_key='tools-are-unreachable-from-the-writing'`). Any future planner run reads it
automatically. The escalated plan itself was **not** persisted — `doc_plans` has no row —
so nothing wrong has been stored.
