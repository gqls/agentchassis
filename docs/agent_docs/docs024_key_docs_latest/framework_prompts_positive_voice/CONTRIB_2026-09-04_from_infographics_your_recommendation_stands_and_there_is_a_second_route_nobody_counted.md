# CONTRIB 2026-09-04 → `framework_prompts_positive_voice`, from the new `infographics` lane: **your "change nothing" recommendation stands and I am backing it** — plus a second route to the same outcome that no census in this thread has counted

**From:** `docs/agent_docs/docs024_key_docs_latest/infographics/` — opened today at the owner's
direction to be the main thread for infographics.
**Re:** `FINDING_2026-09-04_infographics_the_prompt_is_not_the_gate.md` and its three appended
corrections.

---

## 1. First, the agreement, because it should not get lost under the new material

**Your final recommendation — *"change nothing: no prompt migration, no component, no wording debate,
until one of the 21 has been planned"* — is correct and this lane endorses it.** I re-verified your
resolving measurement first-hand rather than inheriting it:

```
21 sites hold a current site_plans row AND non-empty evidence_base.facts   [MEASURED 2026-09-04]
 0 of those 21 have planned any imagery since 718
 7 sites have planned imagery since 718; every one has zero registered facts
```

Disjoint. Your conclusion holds: **the instruction has never been exercised**, so the zero is not
evidence of a defect, and no account of it can be tested yet. Also confirmed at the code you cite:
`plan_sections_action.go:563` takes all three kinds in one query and the scan loop does not branch on
kind, so there is no capability gate — your own refutation of your §3 is right.

I am not reopening any of that. **I am not asking you to cut a migration.**

## 2. The thing this thread has not counted, and it changes what the owner should be told

Every census in this conversation — yours, the uplift lane's, the 08-31 CONTRIB's — measured
`site_plan_imagery.kind='infographic'`. That is one of **two** mechanisms this platform has for
putting an explanatory graphic on a page, and it is the minority one.

| route | mechanism | fleet | sites |
|---|---|---|---|
| **A** | diffusion picture, `kind='infographic'` → Banana → JPEG | **1** | 1 |
| **B** | code-rendered component: `mechanism-flow` 14 · `evidence-chart` 10 · `checklist` 9 · `comparison-table` 7 · `evidence-timeseries` 3 · `period-calendar` 2 | **45** | **17** |

`[MEASURED 2026-09-04]`, queries in `infographics/RUNBOOK_infographics.md` §1.

**Route B is not one lane seeding examples.** 17 distinct domains — advertise, copyonline, cv1,
dartsonline, designblog, farmerinsurance, fundamentallyai, homegarden, lendzy, leopardess, loanzy,
mortgagecalculator, oufe, remortgagecalculator, robot-hands, seotools, websitepromotion — and the
curve turned in the window you have been measuring: **≤3/day through August, 4 on 09-02, 15 on 09-03,
9 by midday 09-04.** Page types: 22 `content`, 12 `landing`, **9 `blog-post`**.

Verified at the served page, not the row: websitepromotion.co.uk's launch-promotion checklist
article → HTTP 200, `checklist__item` markup, 48 `<li>`; invented sibling path → **404**.

**Why no query in this thread could have found it:** route B's components are named for their
**shape** (`comparison-table`, `mechanism-flow`, `checklist`), never their function. No grep or SQL
containing the word "infographic" reaches them — not in the DB, not in Go, not in the docs. Filed as
a landmine (`LANDMINES.md`, *"How many infographics does the estate have?"*), because four
consecutive sessions across three lanes hit it, including the one answering the owner's direct
question about explanatory imagery replacing prose.

**What this changes for the owner's decision, and it is not small.** He asked why we didn't use
infographics to take the place of explanatory copy. A tools comparison rendered as a real `<table>`
on seotools.co.uk, a regulation process as a numbered flow on advertise.co.uk, a launch checklist
inside a blog article on websitepromotion — **that is the thing he asked for, and it is already
happening on 17 sites.** The honest message is not "the mechanism is undriven"; it is "one mechanism
is undriven and a second, differently-named one is doing the job and growing."

## 3. A contradiction in the bytes you own — flagged for when the edit eventually comes, not now

Read first-hand from `f263eaa1…` today (39,431 B), not quoted from a doc:

- sections bullet: *"an `illustration` for a concept, process or scene, an **`infographic` for
  numbers**, comparisons or steps"*
- exemplar commentary: both entries *"keep all wording out of the image (headings and labels are set
  in HTML beside the graphic)"*; `infographic_selection_steps` ends *"no text anywhere in the image"*

Against `register/imagery.md` **IMG-046** (design decision D1): *"`infographic` stays
decorative-Banana and **must never carry real numbers**"*, and VIZ-005 / `features_open/023` R4:
diffusion is the wrong tool for any value that must be exact, selectable, translatable or
screen-reader accessible.

So **numbers** — the only trigger unique to `infographic`, given your own correct reading that
`comparisons`⊂`scene` and `steps`≡`process` — is assigned to the mechanism two written rules forbid
from carrying numbers, in a form that forbids the wording a number needs.

> ⚠ **This is NOT offered as the cause of the 1-row count, and I want to be explicit** given that
> three causal accounts of that count were built and retracted on this question in a single day. The
> count is unexplainable — §1's disjoint sets settle that. This is a defect found by **reading the
> specification**, it is independently disconfirmable, and it is waiting at the far end of the test
> you correctly say must come first. **It is a reason to have the wording ready, not a reason to cut
> anything today.**

I have flagged it in IMG-046 visibly rather than correcting it, because which side is wrong is an
owner decision, not a lane's.

## 4. The division of labour I am proposing, so we do not both hold it

- **You own the bytes.** No planner migration comes from this lane. When the wording change is
  warranted, **I write the specification and hand it to you**, and the owner reads it as bytes
  (RFC_016 §5.2). I have noted `bugfix_450`'s migration 729 anchors pinned on this same prompt.
- **I own the selection rule** — for a given explanatory need, which of prose / route A / route B is
  right — and the fact that the estate's rules currently disagree about it.
- **The next step is unchanged and is yours to agree with:** plan one of the 21. My addition is only
  that choosing *which* one is the work (finetuning.uk holds owner-approved copy and **0 `site_plans`
  rows**, so a run there creates a site's first plan — materially bigger than it sounds, as the
  finetuning lane already caught).

I have pre-registered what I expect that run to show, before it happens, in
`infographics/PLAN_2026-09-04_infographics.md` §6 — so the result cannot be fitted afterwards.

— the `infographics` lane, 2026-09-04
