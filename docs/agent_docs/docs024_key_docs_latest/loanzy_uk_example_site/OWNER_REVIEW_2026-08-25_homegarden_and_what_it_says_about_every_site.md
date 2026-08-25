# OWNER REVIEW 2026-08-25 — `homegarden.uk`, and what it says about every site the fleet builds

**This is the owner's review, recorded verbatim in substance by the `loanzy_uk_example_site` lane,
with each point traced to a measurement at the served artefact.** It is the canonical record; other
lanes should cite this file rather than a paraphrase of it. Where I have added evidence, it is
marked. Where I have added an opinion, it says so.

**His framing, which governs everything below:** *"It's a whole category of not thinking about the
user and what they are after on a site like this."* Every specific below is an instance of that, and
he says explicitly he could not list them all — **so treat this as a sample, not a checklist.**

---

## 0. Two decisions he settled in the same message

1. **Card composition: MORE THAN FOUR card sections before something has to break them up.**
   (Answers the open question from his 2026-08-24 review — the "carousel" complaint, which was
   measured and found to be a CSS grid, not a carousel. The real quantity was never cards-per-section
   but **sections-per-page**.)
2. **A RE-PLAN OF `garden-tools.uk` IS AUTHORISED.** *"A re-plan is fine."* That site's job as a
   frozen pre-fix baseline is therefore over; its measurements are dated and recorded in this
   directory and in `bugs_open/381`.

---

## 1. The calendar: the half that worked, and the half that did not

**`[MEASURED 2026-08-25 14:2xZ, served HTTPS, control 404]`** `/index.html` carries
`<ol class="period-cal__list">` with **12 `<li>`, January–December in order**, each with a focus line
and practical detail. That is the fix `bugs_open/381` shipped and it is real.

**What is missing, in his words: *"there are no links outward, there is no imagery."*** Both measured:

| | measured |
|---|---|
| `<a href>` inside the calendar block | **0** — the twelve months are **not links** |
| `<img>` on `/index.html` | **1** (the logo) |
| `<img>` on `/april/index.html`, `/about.html`, `/garden/index.html` | **1** each — the logo |
| `<picture>`, `<svg>` anywhere | **0** |

**His instructions:**
- **The calendar must link through to the monthly guides.** The twelve month pages exist and are
  deployed; the calendar simply does not reference them.
- **The monthly guides need much more imagery, and it should sit BETWEEN PARAGRAPHS**, not just at
  the top.
- **The site could carry a background image or graphics to make it more interesting.**
- ⚠ **"All this could be default for sites generally unless determined otherwise."** That is the
  important sentence — he is not asking for a fix to one site. He is asking for the DEFAULT.

---

## 2. Links that promise one thing and deliver another `[ALL VERIFIED AT THE SERVED PAGES]`

| page | button text | target | his verdict |
|---|---|---|---|
| `/index.html` | *"See what's due in your garden and home this month"* | `/this-month/index.html` | the target **does not tell you what's due**; it talks generally — *"Why the month matters more than the job"* |
| `/index.html` | *"Look at seasonal garden maintenance guidance"* | `/garden/index.html` | the target **tells the user how we as site owners organise things** |
| `/garden/index.html` | *"Check what's due in the garden this **April**"* | `/april/index.html` | **it is August.** Should link to August, or forward to September, or both |

> **On the second one he is emphatic and it generalises: *"bad and absolutely not what a user wants
> to read"*.** And on the first: ***"we don't want any of that sort of content anywhere on any
> site."*** These are not page-level defects; they are a content CLASS to be eliminated fleet-wide.

**⚠ ONE HE DID NOT MENTION AND I AM ADDING, because it is the same class and it is on every page:**
every page carries a **`Get Started` button pointing at `/contact.html`**. On an editorial gardening
site "Get Started" is a SaaS template CTA that means nothing — get started with *what*? It is the
purest example of a component placed without asking what the reader wants. `[MEASURED: present on
all four pages sampled]`

**On the stale April link he draws the general conclusion himself:** *"Generally across all sites in
the fleet this sort of lack of second-tier planning is letting us down."*

---

## 3. Telling the reader what we do NOT do

*"We start telling people that we don't do this or that like 'won't find brand comparisons or price
tables here' — it's just absolutely not necessary to say this on the site."*

`[MEASURED]` Confirmed and worse than one instance. On `/about.html` the heading **"What this site
will not do" appears TWICE**, plus *"Editorial approach and what we will not do"*, *"No product
endorsements"*, *"No brands, no fabricated tests"*. `/index.html` carries *"this site will not tell
you…"*.

---

## 4. The `about.html` page — his sharpest criticism, and it is measurable

*"The about.html page is especially bad and sounds like an AI created it."* His points:

- **The premise of the page is wrong.** *"Stop talking about us in such a technical way."* The
  readers *"are interested in homes and gardens and not in our technical prowess or lack of it"*.
- **It should be brief and about what we are trying to do FOR THEM** — his own example of the right
  register: *"We're hoping you can get a lot of useful tips from this site…"*
- **Remove almost all of it.**

**`[MEASURED 2026-08-25 14:2xZ]` — the page has 17 content headings and FOURTEEN of them are about
the site's own methodology, not about homes or gardens:**

> How Home Garden decides what to tell you, and what it leaves out · How this site is put together ·
> How the guidance is worked out · **What this site will not do** · Sourced and dated · No product
> endorsements · Timing stated plainly · The principles behind what we publish · Sources named ·
> No brands, no fabricated tests · Editorial approach and what we will not do · Where the detail
> comes from · **What this site will not do** *(again)* · Why the plain answer matters more than a
> confident one

Only three are reader-facing: *What brought you here*, *Who it's written for*, *DIY jobs versus
professional ones*.

**The specific copy patterns he named, all confirmed present:**

| pattern | his objection |
|---|---|
| *"and names the source behind any figure it uses"* | nobody wants this |
| *"Most visitors arrive with one specific question:"* | **presumptive** |
| *"say plainly"* / *"Timing stated plainly"* | *"people just don't say that"* |
| *"How the guidance is worked out"* + the whole shape-of-an-article paragraph | *"Who would want to know that, really? (no one reading this site)"* |
| *"plain answer"*, *"flattened"*, *"one date that suits nowhere in particular"* | AI register |
| *"What this site will not do"* | *"no one cares"* |

**⚠ He is explicit that this list is incomplete: *"I can't list it all because the whole page needs a
revisit, it's all bad."*** Do not treat the table above as the defect set.

---

## 5. The comparisons page

*"There are no comparisons on the comparisons page only copy that shouldn't be there at all
describing comparisons as a concept."* `[MEASURED]` `/comparisons/index.html`: **0 `<table>`**,
headings are *"What these comparisons set out to do"*, *"What each comparison covers"*, *"What a
comparison cannot settle for you"*. Its own `<h1>` asks *"Decking or paving. Fence panels or gravel
boards. Which one actually suits your garden."* and the page never answers it.

**Mechanism already established by this lane:** the page's `content-listing` section had nothing to
list and skipped silently (`source: query.blog_posts`, `on_missing: skip_section`), so the index has
no index. Routed to `bugs_open/384`. **But the copy that remains is a separate defect and is his
point here** — the page fills the hole by describing the concept of comparison.

---

## 6. What he proposes, and what already exists

He offers three ideas. **All three map onto agents that ALREADY EXIST**, which reframes the question
from "should we build this" to "why did the existing one not fire, or not help":

| his proposal | live agent(s) `[MEASURED 2026-08-25]` |
|---|---|
| *"a user experience agent that takes the view of a happy user"* | `experience-planner`, `experience-approval-council`, and the `experience_loop` / `experience_register` workstreams |
| *"the offer and benefit analysis agent… make it much more clear what we can offer the customer"* | `offer-analyser` |
| *"let the visual designer know that it hasn't in any way made this site good"* | `visual-designer`, plus `visual-design-auditor`, `brand-designer`, `feature-designer`, `design-audit-agent` |

**His verdict on the designer is unambiguous: *"It hasn't done its job."*** `[MEASURED]` one `<img>`
per page (the logo), zero `<picture>`, zero `<svg>`, one background-image declaration.

**He also offers the alternative himself:** *"maybe we should consider an agent to look at the detail
or improve the experience loop"*. **That is a live question for the owning lanes, not a decision this
lane may take.**

> ### ⚠ ANSWERED 2026-08-25 by the `vigilant_designer_offer_analysis` lane — AND IT CORRECTS THE
> ### OWNER'S PREMISE. The answer is "DID NOT FIRE", not "did not help". See §9.


---

## 7. What he instructed for copy, and it is the largest single item

Addressed to the `copy_quality_two_stage` lane, and quoted because the wording carries the scope:

1. *"We'll need to notify the copy quality two stage to **up their game a lot**."*
2. *"We have discussed copy at length. Please ask the copy quality two stage to do a **deep search and
   refresh their context on this** before suggesting fixes."*
3. *"They should also **audit every prompt in the database and code** and ask of it whether it is
   contributing to good readable copy or whether it is **encouraging AI styles of writing (bad)**."*

**Point 3 is the big one.** It is not a copy fix, it is an audit of the instruction surface that
produces copy — every prompt in `agent_definitions` and every prompt literal in the Go source, judged
against a single question. That is a workstream, not a task.


---

## 8. ROUTING STATUS — closed out 2026-08-25, so this file does not outlive the work it asked for

| section | routed to | status |
|---|---|---|
| §7 (copy, incl. the prompt audit) | `copy_quality_two_stage` lane | **ACKNOWLEDGED.** Deep context refresh running as four parallel sweeps; **no fixes to be proposed until the synthesis exists**, per his instruction. Point 3 scoped as a workstream with a dated census — `agent_definitions` prompt templates + config-embedded prompts + Go prompt literals + per-field `llm_guidance`. Outputs to `docs024_key_docs_latest/copy_quality_two_stage/`. They confirmed they are citing THIS file, not a paraphrase. |
| §1, §2, §6 (imagery, CTAs, offer, designer) | `offer analyser / benefit analyser / visual designer` lane | routed with measurements; no reply yet at time of writing |
| §5 mechanism (listing renders nothing) | `bugs_open/384` | contributed, with positive and negative controls |
| §0 decisions | this lane | recorded; re-plan authorisation noted for `garden-tools.uk` |

### 8a. ⚠ THE CONVERGENCE, and it is a stronger signal than this review on its own

`[VERIFIED here 2026-08-25, by reading the file rather than taking the report]` **This is the SECOND
owner escalation to the copy lane today.** The `finetuning_uk_service` lane routed one this morning:
`docs024_key_docs_latest/copy_quality_two_stage/CONTRIB_2026-08-25_OWNER_ESCALATION_finetuning_pages_fail_the_would_a_person_say_this_test_after_a_maximal_seed.md`.

**Same mechanism, arrived at independently, on a different site in a different vertical:** their own
verification checklist scored the section the owner rejected as **CLEAN**, and their session summary
had **praised** it. Their words: *"the owner's tell class is WIDER than the gate's."*

**Put beside this review's §4, the two say one thing:** the defect is not in the phrases. On
`homegarden.uk` it is that **fourteen of seventeen headings are about the site's own methodology** —
a page can contain no banned phrase at all and still be entirely about the wrong subject. On
`finetuning` a checklist enumerating patterns passed copy the owner threw out.

> **So an enumerable-pattern gate cannot catch this class, and both lanes now have evidence of it
> failing in opposite directions on the same day** — one passing bad copy, one leaving bad copy
> unflagged. **The unit of judgement has to be "what is this page FOR, and would a reader of THIS
> site want it", not "does this sentence contain a tell".** That is the owner's own framing
> — *"a whole category of not thinking about the user and what they are after on a site like this"* —
> and it is now measured twice.


---

## 9. THE §6 REFRAME, ANSWERED — and it changes what "fixing it" means

**Source:** `vigilant_designer_offer_analysis` lane, full measurements and re-run queries in
`CONTRIB_2026-08-25_two_of_the_three_agents_he_names_could_not_have_run.md` (this directory, commit
`9740425e7`). **The load-bearing claim was re-verified independently here before being carried to the
owner**, because it contradicts something he asserted.

### 9a. *"The visual designer hasn't done its job"* — it was never asked. It has no dispatch path at all.

`[MEASURED 2026-08-25, verified twice by two lanes]`

| check | result |
|---|---|
| live agents whose `default_config` names `visual-designer` | **none** |
| `scheduled_tasks` targeting it | **none** |
| `llm_call_log` rows under `agent_type='visual-designer'` | **0**, across the log's whole span (2026-03-25 → today, 67,789 rows) |
| positive control in the same query | `page-content-writer` = **24,217** rows to 2026-08-25 — the query resolves |
| the only Go reference | `spawn_actions.go:3053`, inside `isStorageEnabledAgent` — a **capability grant**, not a dispatcher |

⚠ **And the zero was not rested on.** `agent_type` records the dispatch context, so a hand-fired run
hides under `generic`. Keying on `step_name='design'` instead gives 20 rows — 11 `generic`
(07-17→07-24), 9 `feature-designer` — **nothing since 2026-07-31**, with `run_offer_analysis` = 16
rows to 08-24 as the positive control in the same query.

**`brand-designer` and `experience-planner` are in the same position.** `design-audit-agent` and
`offer-analyser` are named only by `improvement-loop`, which is disabled.

**What actually built `homegarden.uk`**, from its own orchestration rows (21 distinct agent types):
`site-design-planner`, `webdesign-agent`, `site-asset-renderer`, `render-audit-agent`,
`image-build-handler` ×13, `image-generator` ×13. **None of §6's eight appears.**

> **So the honest sentence for the owner is that the designer did not do a bad job — it was never
> asked, and there is no path by which it could have been.** That is worse than the complaint, not
> better: a bad output can be fixed with a prompt; an unreachable agent is an estate-level gap.

### 9b. The offer analyser never saw the site, and it is switched off

`[MEASURED]` 0 `site_specs` rows with `aspect='offer_ordering'`; 0 work items with
`audit_source='offer-analysis'`; **5 of 28 sites enrolled fleet-wide**. All **three** of its scheduled
carriers are `enabled=false`, last triggered 08-14/08-15, and its only other route
(`improvement-loop` → `improvement-sweep`) is disabled too.

**His `about.html` complaint is precisely what this agent exists to catch** — 14 methodology headings
against 3 reader-facing. It landed on a catcher that was switched off.

### 9c. The imagery defect is STRUCTURAL and fleet-wide, and "make it the default" is a bigger ask than it sounds

⚠ **Including a correction that lane made to its own first reading — do not repeat the retracted
version.** It nearly reported *"13 images generated, none placed"*. **Wrong:** 21 of 45 homegarden
components reference `/assets/` and 20 carry `background-image`. **The images ARE placed — exclusively
as CSS backgrounds. There is no in-body image slot anywhere.** That is his *"between paragraphs"*
complaint stated as a mechanism.

**Fleet census, 27 sites with >10 live components: THIRTEEN have ZERO inline `<img>` in any
component** — homegarden (45 components), `gaswholesalers.com` (**115**), `mortgagecalculator.co.uk`
(75), `loanzy.uk` (63). Every site has background images (2–56). **Best inline coverage anywhere is
`loancalculator.co.uk` at 25/157 = 16%.**

**So his *"this could be the default for sites generally"* asks us to change something that is
currently near-absent BY DEFAULT.** It is a component-vocabulary and planner question, not a matter
of a designer's taste — the same shape as `bugs_open/381`'s missing list/table components.

### 9d. What is now an OWNER decision rather than a lane's

That lane deliberately changed nothing live: **no carrier enabled, no enrolment written, site
untouched** — noting that enabling a disabled carrier is a fleet behaviour change affecting every
site it selects, and that it has already paid once for firing a dispatch believing a promoter was off
(`WRONG_CALLS.md` 2026-08-25: findings promoted and a live page rebuilt 31 seconds later).

**So the decisions in front of the owner are:** whether to switch the offer-analysis carriers back
on; whether the visual designer should be given a dispatch path at all or retired; and whether an
in-body image slot enters the component vocabulary as a default.
