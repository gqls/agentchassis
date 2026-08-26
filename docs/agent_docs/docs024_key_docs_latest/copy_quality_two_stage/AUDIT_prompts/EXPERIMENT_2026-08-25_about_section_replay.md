# EXPERIMENT (pre-registered) — replaying homegarden's about-content section with the instructed premises removed, 2026-08-25

**Written BEFORE any arm was run.** Owner instruction of 2026-08-25 (this session): *"Let's test
the changes here, I don't think we are there yet."* This is the Finding-2 causal test
(`PHASE1_2026-08-25_findings.md`), run OFFLINE by replay — no live prompt, template, page or spec
is touched; calls go through the platform's own provider credentials from a chassis pod (the key
never leaves the pod).

## Subject

`llm_call_log` id `efcb756b-24f6-4752-9d93-668ecae5dba5` — the `about-content` section of
homegarden.uk's about page, built 2026-08-25 11:38 by `page-content-writer` on
`claude-sonnet-5`, max_tokens 16000, default temperature. Its shipped output is the body the owner
rejected. Rendered prompt: 29,252 chars, saved verbatim.

**What the prompt itself carries, found reading it whole:** the premise arrives in LAYERS —
(1) the PLANNER's page title, *"About Home Garden — Editorial Approach and What We Will Not Do"*
(line 1 and the internal-links list); (2) the writer template's three "instead" instructions (the
STRICT-RULE remedy clause; the Operating-history "method = sourcing" sentence; rule 19's
"we cannot tell you X" substitute); (3) the site brief's trust line (*"Trust is built by
acknowledging what the site does not know … Explicit honesty about uncertainty is itself a trust
signal"*) — kept intact in ALL arms, recorded here as a residual carrier; (4) the field guidance's
"(origins, approach, who it is for)" — also kept in all arms.

## Arms — 3 runs each, 12 calls total

| arm | prompt |
|---|---|
| A | verbatim replay — the VALIDITY CONTROL |
| B | A minus the writer template's three "instead" instructions (every BAN kept: no-overclaiming, no-history-claims, the "honest" word ban) |
| C | B, plus the page title neutralised to "About Home Garden | Home Garden" in both places |
| D | C, plus the candidate replacement instruction inserted where the STRICT-RULE remedy was: *"If the section calls for a statement about method or about the site, write about the reader's subject: what they can find here and what it helps them do, in one or two sentences, and stop. Do not describe the site's sourcing, editorial process, or limitations unless the brief for this page explicitly asks for it."* |

Every excision/insertion is applied by exact-substring with an asserted match count; the arm files
are diffable against A.

## Metrics, defined now

Over each output's `content` + `highlights` + `section_title` text (tags stripped):

- **m1 methodology self-description** — case-insensitive count of `sourc|dated|editorial|method|
  how this site|how the guidance|we publish|we decide|we check|principle`
- **m2 self-limiting / performed candour** — `will not|won't|cannot|can't|no product|no brands?|
  plainly|be wrong|wrong answer|does not (sell|endorse)`
- **m3** the five negation-tell shapes (as in `count_negation_tells.py`)
- **m4** the `<h3>` headings, LISTED — classified reader-facing vs site-facing by reading, the
  owner's own measure (his: 14 of 17 site-facing)

**Primary outcome: m1+m2.** Baseline: the ORIGINAL shipped response is scored with the same
regexes before any arm runs.

## Pre-registered readings

- **A fails to reproduce** the register (median m1+m2 < half the original's score) → the replay
  protocol is INVALID (a system-side element is missing from `prompt_rendered`); stop, publish
  the failure, no conclusions.
- **B ≈ A** (median m1+m2 > half of A's median) → the template clauses are NOT the operative
  source for this section — Finding 2's simple causal claim is REFUTED here and the planner
  title / brief trust line are the live suspects.
- **B markedly below A** (median ≤ half of A's) → template clauses causal, at least in part.
- **C below B** → the planner's title carries premise weight of its own (expected; the planner
  audit inherits it).
- **D** judged by reading plus m1+m2 near zero AND the output still being a usable about section
  — an empty or evasive section is a failure of the candidate fix, not a success.
- n=3 per arm detects only GROSS effects; no fine rate claims will be made from this. Model
  stochasticity is the reason for 3, not a power calculation.

## What this experiment cannot show

Fleet-wide generality (one section, one site); anything about the wider "methodical" register
(m-metrics are lexical); the brief-layer and guidance-layer contributions (held constant by
design). A B-refutation does not clear the template clauses elsewhere — reviewer-facing pages
with real method sections still read them.

## Results

*(appended after the runs — nothing above this line changes after they start)*

**Run 2026-08-25 evening. 12/12 calls completed (claude-sonnet-5, ~9.5K in / 2-4K out each,
platform credentials from the pod, no key extracted). Scores are m1+m2 per run; medians bold.**

| arm | runs (m1+m2) | median | "What this site will not do" heading | "How the guidance is put together" heading |
|---|---|---|---|---|
| ORIGINAL (shipped) | 10 | 10 | present | present |
| A verbatim | 15 · 10 · 7 | **10** | **3 of 3** | 3 of 3 |
| B − template clauses | 6 · 5 · 6 | **6** | **3 of 3** | 3 of 3 |
| C = B − planner title | 1 · 4 · 1 | **1** | **0 of 3** | 3 of 3 (reader-framed) |
| D = C + replacement | 2 · 3 · 3 | **3** | 0 of 3 | 3 of 3 |

**Pre-registered readings applied:**
- **A = ORIGINAL (10 = 10): the replay is VALID.** The rejected register reproduces, headings
  near-verbatim.
- **B (6) is NOT ≤ half of A (threshold 5): the "template clauses are the cause" claim is
  REFUTED as the PRIMARY cause** — narrowly, and the 10→6 drop plus the disappearance of
  "sourced/dated" phrasing shows the clauses are a real SECONDARY contributor. The
  "will not do" heading survived their removal in every run.
- **C (1) is the collapse: the PLANNER'S PAGE TITLE — "About Home Garden — Editorial Approach
  and What We Will Not Do" — is the dominant premise carrier.** With it neutralised, the
  self-limiting heading vanished 3 of 3 and the copy turned reader-facing (C3's opening:
  *"Home Garden exists to answer one question: is this job urgent, or can it wait until a
  better month?"* — the register the owner asked for, unprompted by any replacement text).
- **D ≈ C at n=3: the candidate replacement instruction neither helped nor hurt measurably.**
  Deletion plus a right-premised plan may be sufficient; the replacement is not yet earned.
- **The two faults SEPARATE.** m3 (negation tells) stayed at 2–12 per output in every arm —
  the premise cuts do not touch the register tells, because all ~60 demonstrations remain in
  every arm's prompt. Premise ← the planner's page plan (primarily) + writer clauses
  (secondarily). Register ← the demonstration stack. Different fixes, different owners.
- **Residual carriers confirmed live:** "How the guidance is put together" appeared in 3/3 of
  every arm — the field guidance's "(origins, approach, who it is for)" and the brief's trust
  line (*"Explicit honesty about uncertainty is itself a trust signal"*) are enough to keep one
  methodology section, softened, even with title and clauses gone. One such section, framed as
  what the reader gets, may be acceptable — the owner's call.

**Next question the result forces:** WHERE does the planner get a title like "Editorial Approach
and What We Will Not Do"? `build-site-planner`'s rendered prompt is already top-of-table for
candour-beat and em-dash proxies (87–118K chars, 34–59 negations). The planner audit
(phase 2 item 4) is now the priority, ahead of any writer-template migration.

---

## Follow-on: the PLANNER replay (2026-08-25/26) — testing the ruling-1 fix wording before shipping it

Subject: homegarden's real `plan_site` call (`cabfb760`, claude-opus-4-6, 86,752-char rendered
prompt, the call that planned the rejected about page). Metric: the planned about-page `title`,
plus limitation-phrase count across the whole plan. Owner had ordered the fix shipped; the wording
was still tested first — and the first draft FAILED the test:

| arm | about title per run | verdict |
|---|---|---|
| verbatim ×1 | "About Home Garden — Our Editorial Approach" | validity holds — the methodology premise reproduces |
| draft 1 (register rule) ×2 | "About Home Garden" · **"About Home Garden — Practical UK Guidance, No Products to Sell"** | **1 of 2 — the failure is the rule's own named error, produced WITH the rule in the prompt** |
| draft 2 (+ hard format clause: "'About' plus the site name and nothing more: no subtitle, no dash, no qualifier") ×3 | "About | Home Garden" · "About | Home Garden" · "About | Home Garden" | **3 of 3 clean**; limit-phrases across whole plans 0–1 (baseline 1) |

Reading: **register rules bend; format rules hold** — the same lesson as the 08-12 heading-examples
finding, at the planner layer. Draft 2 shipped as migration `630` (applied 2026-08-26, verified at
the live row, `Council-Submitted: 5f084feb`). Residue: the ~6 existing self-limiting about titles
are not rewritten by a prompt rule — per-site re-plans or title edits, per-lane work.
