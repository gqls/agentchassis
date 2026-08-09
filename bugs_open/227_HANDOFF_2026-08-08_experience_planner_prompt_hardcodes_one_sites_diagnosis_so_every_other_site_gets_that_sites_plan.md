# 227 — `experience-planner`'s prompt hardcodes ONE site's diagnosis, so an experience plan for any other site describes that site's pages

**Filed 2026-08-08** by the `loancalculator_couk` lane, which ran the planner to judge a
page-ordering question and got back a detailed, confident plan about a different site.

**Latent since 2026-07-18 and invisible until now for a simple reason: the agent had only
ever been run on the site it is hardcoded for.** `debt-difficulty-help` is the FIRST
non-vonc experience ever planned. See Measured.

> ## ⚠ CORRECTED 2026-08-08 ~23:00Z by the same lane, before writing the fix. THREE
> ## corrections, and each one changes what the fix has to do.
>
> **1. It is FIVE prompts, not one.** This file says the `compose` prompt. A
> case-INSENSITIVE census of `provocation|gauntlet|arena|vonc|spark` over the live
> row (`e0194bee-…`) returns **48 hits across five steps**: `compose` 41,
> `review_feasibility` 2, `review_honesty` 2, `review_mvp` 2, `reframe` 1.
> **My own first census missed two of them because it was case-sensitive** —
> "Gauntlet" is capitalised everywhere it appears in `reframe` and `review_honesty`,
> so a `~ 'gauntlet'` grep returns 37 hits and shows only three steps, and both
> veto-holding seats read as clean. A grep proves absence only for the spelling it
> searches; measure `lower()`.
>
> **2. The council is contaminated too, so "the review layer is not the defect
> here" (below) is too generous — and a fix to `compose` alone is a trap.** Three
> of the four seats hold vonc's specifics as their general judging criteria:
> `review_feasibility` (veto) asks whether data is "in /data/provocations.json or
> client-computable" and watches for "the daily emitter"; `review_mvp` is told the
> core loop is "land on a provocation → file a position → …; enter a real timed
> Gauntlet round"; `review_honesty` (**hard** veto) is told "vonc's evidence_base
> has ZERO facts". And `reframe` — the step that rewrites a vetoed plan, i.e. the
> one that produced the SECOND contaminated plan in the run measured below —
> carries "If the Gauntlet is what was vetoed …". The seat that correctly refused
> the loancalculator plan was itself holding another site's criteria. **So a
> CORRECT post-fix plan can still be objected to by a seat looking for a feed and a
> timer this site was never going to have, and that would look exactly like "the fix
> did not work".**
>
> **3. The hardcoded premise is not only wrong-site, it is STALE — inside a hard
> veto.** `review_honesty`'s "vonc's evidence_base has ZERO facts" was true when
> written on 2026-07-18. It is **false now**: `site_specs` aspect `evidence_base`,
> `is_current`, for vonc.com holds **4 facts**, written **2026-08-08 08:58Z**
> (loancalculator has no such spec row at all). A vonc plan run today is told by its
> own anti-fabrication seat that four verified facts do not exist. Nothing updates a
> premise pinned inside a shared prompt when the site moves underneath it — which is
> an argument for the brief being data that stands independently of this bug.
>
> **A fix must also NOT leave vonc briefless.** D1/D2 are owner rulings marked "do
> NOT relitigate" and 59 of 61 plans all-history are vonc's; deleting the text
> without rehoming it trades a contaminated plan for a de-briefed one on the only
> site that has ever used this agent in anger.
>
> **Fix candidate 1 is written, dry-run proven, and NOT applied:**
> `docs/agent_docs/sql_for_agents/345_experience_planner_site_brief_becomes_data.sql`
> (+ `_ROLLBACK`). Config only, live on apply, no image. See that file's header for
> the design; see the `loancalculator_couk` lane's
> `HANDOFF_2026-08-09_continue_here.md` for what is left to do.

## The defect

`agent_definitions` row `experience-planner` (active, one row, fleet-shared) has proper
template variables — `{{.experience_name}}`, `{{.experience_domain}}` — so it was written
to be generic. Its `compose` prompt then continues, hardcoded:

> "Write an EXPERIENCE_PLAN for the `{{.experience_name}}` experience on
> `{{.experience_domain}}`. …
> **## The diagnosis you are fixing (three broken surfaces, artifact-verified 2026-07-17)**
> 1. `/provocations/index.html` — archive entries are runtime-filled into a template whose
>    `href="#"` was never given a destination…
> 2. `/tools/arena/index.html` — the tool-arena widget does NOT fetch `/data/provocations.json`…"

Those are **vonc.com's** surfaces. Occurrence counts in the live `default_config`:
`provocation` ×24, `gauntlet` ×8, `arena` ×5.

So every invocation, for any site, is told that the diagnosis it is fixing is vonc's, in
the imperative, immediately after being told which site it is actually planning for.

## Measured

Run: correlation `a30b0c5b-7e0f-4d91-82a4-061b3af8e6c9`, parent
`271c193e-…`, child planner `8991ef52-…`.

**The inputs were correct.** `initial_request_data.input_data` carried
`site_id=0162cde4-…` (loancalculator.co.uk), `experience_key=debt-difficulty-help`,
`experience_name='getting help when you cannot keep up with a loan repayment'`.

**The loaded context was correct.** `collected_data` contains loancalculator's own pages
(`guide-hidden-loan-fees` present) and **zero** occurrences of `provocation`.

**The output was not.** Both `compose` and `reframe` produced plans whose four journeys are
`/provocations/index.html`, `/tools/arena/index.html`, `/tools/gauntlet/index.html` and
lobby cards. loancalculator has **0** pages matching provocation/arena/gauntlet; vonc.com
has **6**.

**Where vonc enters the run** — `collected_data` keys whose value mentions it:
`compose`, `reframe`, `proposal`, `review_mvp`, **`agent_config`**, `review_contracts`,
`review_feasibility`. It is absent from `load_context` and from `load_schema_hint`
(10,547 b). **`agent_config` is the agent's own definition.** That is the source.

**Population, all-history:**

```sql
SELECT subject_key, count(*),
       count(*) FILTER (WHERE body ILIKE '%provocation%' OR body ILIKE '%gauntlet%') AS vonc_surfaces
FROM doc_plans WHERE subject_type='experience' GROUP BY 1;
```

```
vonc-spark-game        59   59    <- correct for this site
debt-difficulty-help    2    2    <- the only non-vonc subject, 100% contaminated
```

**61 plans, 59 of them for the one site the prompt describes.** The blast radius is
therefore not "59 bad plans" — it is **every future experience on every site**, and the
reason it has never bitten is that nobody has run it elsewhere until today.

## What worked, and it is worth saying

**The council caught it.** Round 1, 5 reviewers, `"decision": "rejected"`,
`"decided_by": "veto from feasibility"`, `should_reframe: true`. The run ended
`complete_refused`. A feasibility seat reading a plan about pages that do not exist on the
site correctly refused it. ~~**The review layer is not the defect here.**~~

> **CORRECTED 2026-08-08 ~23:00Z — the last sentence is wrong; see correction 2 at the
> top.** The verdict was right, but the seat that produced it carries vonc's criteria in
> its own prompt, as do two of the other three. The right verdict here is
> over-determined: a vonc-shaped plan on a non-vonc site fails a generic check ("these
> pages do not exist") *and* would fail a contaminated one. **What this run cannot tell
> us is which check did the work** — so it is not evidence that the review layer is
> sound, and I should not have written that it was. What caught the correction was
> re-measuring case-insensitively, one query, after the fix was already drafted against
> the three steps a case-sensitive census had shown.

## ⚠ SECOND, SEPARABLE DEFECT — a REJECTED plan is persisted as `is_current`

Sequence in the same run:

```
18:21:43  persist  13,721 b   (compose output)      is_current -> f  (superseded later)
          council  REJECTED, veto from feasibility
18:25:45  persist  13,142 b   (reframe output)      is_current -> TRUE
18:25:53  run ends complete_refused
```

**The plan the council vetoed is now the plan of record for `debt-difficulty-help`.**
Anything reading `doc_plans` for that subject — a build round, an acceptance ladder, a
future thread — gets a rejected, fabricated document with nothing marking it as refused.
`persist_plan` runs before the refusal path and nothing demotes it afterwards.

Demoted by hand at filing time (`is_current=false`, note recorded) so nothing builds from
it. **That is a one-off cleanup, not a fix.**

## Fix candidates, ordered by what closes the door

1. **Move the worked case out of the prompt and into the input.** The diagnosis of the
   site being planned should arrive as data — the loader already fetches pages,
   components, work items and JS. A `## The diagnosis you are fixing` section belongs in
   `load_context`'s output, per site, or not at all. This makes the bad state
   unrepresentable: there is no site-specific text left in the shared row.
2. **If an illustrative example is genuinely wanted, fence it as one** — label it
   explicitly as an example from a different site and instruct that it must not appear in
   the output. Weaker: today's evidence across three separate experiments in the
   `loancalculator_couk` lane is that **a worked example beats the instruction around it**,
   so a fenced example may still leak.
3. **Refuse to persist a rejected plan as current** (the second defect). At minimum,
   `persist_plan` should set `is_current=false` when the council decision is `rejected`,
   or the refusal path should demote what it just wrote.

Candidate 1 is the real fix; 3 is independent and cheap.

## How to verify a fix

- Re-run `092_TRIGGER_experience_plan.sh loancalculator.co.uk debt-difficulty-help "…"`
  and assert the resulting plan names **loancalculator's own** pages. Negative control:
  `body ILIKE '%provocation%'` must return **false**, which it has never done for a
  non-vonc subject.
- `SELECT count(*) FROM agent_definitions WHERE type='experience-planner' AND is_active
  AND default_config::text ILIKE '%provocation%'` → **0**. ⚠ **Insufficient on its own,
  and case-sensitively it is worse than insufficient** — use
  `default_config::text !~* 'provocation|gauntlet|arena|vonc|spark'`, which is what
  caught the two steps this file originally missed.
- **The behavioural check needs a POSITIVE CONTROL, or a fix that loads no brief at all
  passes it.** `llm_call_log.prompt_rendered` is the only record of what the model was
  actually handed — the field that would have caught this bug on day one. On the
  `compose` step of a post-fix run: for `debt-difficulty-help`, `prompt_rendered` must
  NOT match `provocation` and MUST contain the no-brief sentinel; for `vonc-spark-game`
  it must come out the **other way** (matches `provocation`, no sentinel), because vonc's
  brief legitimately contains the word. One direction alone cannot tell a working channel
  from a channel that silently loads nothing.

  > **⚠ CORRECTED 2026-08-09 by the applying session — "MUST contain the no-brief
  > sentinel" is a check that CANNOT FAIL, and I ran it before I noticed.** The phrase
  > "no brief on file" also occurs **once in the static `compose` template** (the
  > instruction covering the no-brief case), so `prompt_rendered LIKE '%no brief on
  > file%'` is TRUE on every run of every site — including one where `load_brief` was
  > never wired. It came out TRUE for the vonc control too, where this file demands
  > FALSE, and **that looked like the fix failing when it was the assertion failing.**
  > The disconfirmable form is the COUNT (2 = template + rendered fallback, 1 = template
  > only) or the substring only the COALESCE emits, `'(no brief on file for this
  > experience — there is no prior diagnosis'`. What caught it was the control
  > disagreeing with the prediction; what would have caught it a step earlier is
  > grepping the template I had just installed for the phrase I was about to assert on.

- **PROVEN LIVE 2026-08-09**, applied then run in both directions with the corrected
  assertion — same step, opposite outcomes, keyed only on `subject_key`:

  | run | `no brief on file` hits | COALESCE fallback rendered | leaked | prompt |
  |---|---|---|---|---|
  | loancalculator / `debt-difficulty-help` (`c3976aab`) | 2 | TRUE | **FALSE** | 24,721 b |
  | vonc / `vonc-spark-game` (`72f540d3`) | 1 | FALSE | TRUE (correctly) | 70,427 b |

  The resulting `debt-difficulty-help` plan of record is clean —
  `body ~* 'provocation|gauntlet|arena|vonc|spark'` → **false**, 11,442 b, names loan
  and debt subjects — the first non-vonc experience plan in this system's history that
  does not describe vonc's pages.
- After a rejected round, `SELECT is_current FROM doc_plans …` → **false**.

## Filing basis

**CLAUDE.md requires a cross-cutting structural claim to go through the `090` loop, or the
filing session to state why it substituted equivalent first-hand verification. This is
that statement.** Substituted, because every link was read directly rather than inferred
and the failure was induced rather than predicted: the live agent config carrying the
hardcoded diagnosis; the run's own recorded input; the run's own loaded context proven to
contain the right site and none of the wrong one; the output; the per-key search locating
`agent_config` as the only plausible source; the all-history population; and the council's
own recorded verdict. A `090` run would add an independent reader, which is worth having
if the fixing thread wants it — the symptom to file would be *"experience-planner produces
plans describing vonc.com's pages for every site, and its prompt contains vonc.com's
diagnosis"*.
