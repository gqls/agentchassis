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

**The defect is one row, not one component.** The top nav is generated from
`site_plan_pages WHERE in_header`. That table holds exactly six rows, and they are
exactly the six items your nav shows:

```
index  Home | about  About | capabilities  Capabilities
platform-log-index  Platform Log | news-index  News | contact  Contact
```

**`tools` has no row in the plan at all.** Neither do four other live pages:

| page | serves at | nav_label |
|---|---|---|
| `tools` | `/tools.html` | `Tools / Index` |
| `llm-cost-calculator` | `/tools/llm-cost-calculator.html` | `Tools / LLM Provider Cost Comparison Calculator` |
| `tool-model-approach-selector-guide` | `/guides/…` | *(blank)* |
| `tool-automation-savings-estimator-guide` | `/guides/…` | *(blank)* |
| `tool-ai-readiness-checker-guide` | `/guides/…` | *(blank)* |

**Five of twenty-five active pages are absent from the plan.** Look at those two
nav_labels: `Tools / Index` and `Tools / LLM Provider Cost Comparison Calculator`. That
is a `Section / Page` hierarchy. **A Tools section was designed, built, labelled and
deployed — and then never entered into the plan the navigation is generated from.** It
has been live and unreachable ever since.

That is why the symptom looks the way it does: six guides that describe tools, and no
route to any of them. Not an oversight in the writing — an orphaned section.

## 4. Why this changes the size of the job

The blocked plan was going to build a new hub, wire new cards, and add portfolio entries.
Against the real diagnosis, ask 1 is **five plan rows and a label correction**, and ask 2
uses a component the site already runs on six pages.

It also means the acceptance test should be *"the tools section is reachable"*, not
*"a new hub renders"* — and a check that a new hub renders would have passed while the
existing orphan stayed just as unreachable.

## 5. THE DECISION I NEED, and why I have not just done it

The obvious fix — insert the missing `site_plan_pages` rows — is **plan surgery on
fundamentallyai.com, and another front in this same lane is doing plan surgery on this
exact site right now** (the 215 quiet-mode front; two of its seven duplicate pairs are
here, and it deletes and re-adds plan rows as part of its procedure). Two threads editing
one plan is precisely the collision this estate keeps paying for, so I have stopped.

**Option A — I add the five plan rows, coordinating with the 215 front first.**
Smallest change; fixes ask 1 outright. Needs the 215 front to confirm it is not
mid-transaction on this plan.

**Option B — I do only the two Tools rows now (`tools`, `llm-cost-calculator`), leave the
three guide pages.** Narrower blast radius on a plan someone else is editing. The three
orphan guides stay unreachable, which is a smaller version of the same bug.

**Option C — file it as a work item and let the framework do it.** Most in keeping with
"every site goes through the framework", slowest, and the queue is currently ~305 deep
with the oldest item four days old.

**Separately, and not blocked by the above:** the nav label wants correcting to `Tools`
(currently `Tools / Index`), and the footer link still renders **"Llm Cost Calculator"**
— a title-caser applied to an acronym. If that helper is generic, it is a fleet defect,
not a label defect.

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
