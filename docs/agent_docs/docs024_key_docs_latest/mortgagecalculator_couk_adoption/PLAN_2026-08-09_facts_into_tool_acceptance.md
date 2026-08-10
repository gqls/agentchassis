# PLAN 2026-08-09 — wiring the registered facts into tool acceptance

**Status: DESIGN ONLY. No code written by this session.** Written by Fable for
Opus to implement. Every figure below was measured against the live system in
this session and carries its check; inferences are marked.

> **EXECUTION STATUS — do not read the phase list below as untouched.**
> `HANDOFF_2026-08-10_continue_here.md` carries live state and supersedes §4 here.
> As of 2026-08-10: **A2 done** (first sweep ran and re-verified all four facts),
> **A5 done** (comparator re-run; all three supply-both rebuilds landed),
> **A3 part done** — migration `366` applied, so Piece 1 is live for
> `tool-recreation-handler` and its effect on a real rebuild is still unproven.
> A1 and A4 remain as written.

The one-line problem: **the evidence register guards copy, not code.** A fact
constrains what a page may *say*; nothing connects a fact to the constants
inside a calculator's JavaScript. mortgagecalculator.co.uk's original stamp-duty
tool ran an expired tax rule for sixteen months and every check we own passed it.

---

## 0. What "wiring" has to mean, and the principle that decides the design

Three distinct failures hide behind one phrase. Naming them separately is what
keeps the design from collapsing into "add a check":

| # | failure | who could catch it today |
|---|---|---|
| F1 | a tool is **built** encoding a wrong or stale threshold | nobody |
| F2 | legislation **moves**; the register notices; the tool is never told | the register notices (daily); nothing propagates |
| F3 | a golden captured from an already-wrong tool **pins** the wrong answer and then defends it | nobody — and the runner's own source says so |

F3 is not speculation. `internal/adapters/browserrunner/run_checks_action.go:775-781`
states it in the code that does it: *"a golden captured from an already-wrong
tool pins the wrong answer"*. A green `computed_values` today means **"the
arithmetic has not moved since capture"** — which is a real and useful claim,
and is not the claim we need.

**The principle, inherited rather than invented.** `PBP-037` settled the same
question for prose and its wording is exact:

> the assignment pins WHICH facts a section states, **never their values** — a
> re-verified fact's new number flows through on the next build.

Apply it verbatim to tools. A tool declares **which facts it encodes**; the
values are resolved from the current register at build time and at check time.
Anything that pins a value into an artefact re-creates F3 one rung higher.

---

## 1. What exists — measured this session

Four seams, all live, none joined to the next.

**(a) The register — per site.** `site_specs.aspect='evidence_base'`, one current
row per site. `mortgagecalculator.co.uk` carries 4 SDLT citation facts seeded
2026-08-09, each `{id, kind, unit, claim, value, source.citation{url,quote,
publisher,title,accessed,published}, verified_at, writer_line, staleness_days}`.
[MEASURED — keys enumerated with `jsonb_object_keys`, not read off a seed file;
there is no `.sql` for these facts in the repo, they were seeded direct.]

The daily `evidence-freshness` task is enabled, `interval_seconds=86400`, and
**last ran 2026-08-09 08:58Z — before these facts were seeded (~12:30)**. So the
first sweep over them has NOT happened yet; due ~08:58Z on 2026-08-10.
[MEASURED: `scheduled_tasks` row + zero `stale_evidence`/`citation` items for
this site. This is the check `RUNBOOK §11` sets up, and it is still owed.]

**(b) The criteria document — fleet-global, NOT per site.** Criteria live as a
```` ```criteria ```` fence inside `doc_plans.body`, keyed
`(subject_type='tool', subject_key)` with a **unique index on that pair where
`is_current`** — and `doc_plans` **has no `site_id` column at all**.
[MEASURED: `\d doc_plans`.] This is the single most important constraint in this
plan and §5 turns it into a design rule.

**(c) `computed_values` — the only rung that judges arithmetic.** Tier 4 only, by
deliberate judgement (`experience_criteria.go:73-88`); drives `steps`, then
asserts the **exact text** of each selector (`runComputedValues`,
`run_checks_action.go:792`). Whitespace is the only latitude.

**(d) The tool-building agents are blind to the register.** Of eight agent types
checked, `page-content-writer` and `build-site-planner` reference `evidence_base`;
`tool-generator`, `tool-deployer`, `tool-recreation-handler`, `tool-improver`,
`tool-suggester` and `tool-acceptance-agent` do **not**.
[MEASURED, and disconfirmable — the same query returned `t` for two of the eight,
so a blanket false was not baked into the check.]

**And the detail that makes Phase A nearly free:** `tool-recreation-handler`
already runs a `load_site_specs` step — `read_site_spec` with **no `aspect`
config**, which returns *all* current aspects keyed by aspect name
(`site_spec_actions.go:457-490`). So `{{.site_specs.specs.evidence_base.facts}}`
— the exact template path `build-site-planner` already uses — **already resolves
inside the recreation agent's context.** It always received the facts; it was
never shown them. That is PBP-037's finding, repeating.

### Where this site actually stands

- **All five improvement-loop items filed this morning are `complete`** (built
  11:08–11:19Z): the three recreations and both `add_tool` companions.
  [MEASURED by row id.] The handoff's next action — re-run the replay comparator
  — is now unblocked and owed.
- The site has **14 tool pages**; the two generator-built companions
  (`tool-bridging-compound`, `tool-rate-scenarios`) have `doc_plans` PLANs,
  **the twelve recreated tools have none.** No PLAN ⇒ no criteria ⇒ no Tier 2
  and no Tier 4. Zero `acceptance_run`, `improve_tool`, `audit_tool` or
  `acceptance_stuck` items have ever existed for this site. [MEASURED.]
  This is `TL-032` biting exactly as written: the recreation handler never
  creates a component row, and it does not write a PLAN either.
- **A concurrent lane is moving in this area right now.** loanandmortgagecalculator
  installed **9 `computed_values` fences at 14:33–14:40Z today** (`created_by =
  operator:bugfix224-session`), taking the fleet to 19 fences carrying that check
  across 59 current tool PLANs. Its commit message (`5dbd47653`, 14:25Z) says
  "NOT installed" — true when written, overtaken minutes later. **Do not treat
  that commit as the current state; re-measure `doc_plans` before touching a
  fence.** Nothing in this plan may change the meaning of a fence they just shipped.

---

## 2. The design — four pieces, and why in this order

Ordered by what each makes *unrepresentable*, not by size.

### Piece 1 — SHOW THE BUILDER THE FACTS (config only, no Go, no image)

Inject the site's fact roster into the three tool-building prompts, the same way
`page-content-writer` gets its writer block.

- `tool-recreation-handler` → **prompt seed only**; the data is already in context.
- `tool-generator`, `tool-improver` → add a `read_site_spec` step (an action that
  already exists and is registered) plus the prompt clause. Still config.

Effect on F1: a rebuilt stamp-duty tool encodes £500,000 because the register
says £500,000 — instead of because a session hand-typed the formula into a work
item spec, which is what happened this morning (NOTES 08-09).

This is the cheapest piece and the largest single reduction in F1. It is also the
piece that makes every later phase *repairable*: once the builder reads the
register, "fix the tool" is a rebuild, not a patch.

**It does not close any door on its own** — an LLM shown a fact may still ignore
it. That is what Pieces 2–4 are for. Say so in the register entry; do not let
this ship described as a guarantee.

### Piece 2 — THE DECLARATION: which facts does this tool encode?

Add a `facts: ["sdlt-ftb-relief-cap", …]` list to the tool's criteria document.
Extracted by the existing `load_doc_context` fence reader; validated by
`ValidateExperienceCriteria` (`experience_criteria.go:237`), which must be taught
the key or its P10 "no inert fields" rule will reject it — that rule is the
reason the key must be *declared*, not merely tolerated.

**The `doc_plans`-has-no-`site_id` rule.** A fact id is per-site; a PLAN is
fleet-global. So: **fact ids are resolved against the register of the site whose
page is being driven**, never against the PLAN. The acceptance agent already
resolves the tool's deployed URL from `pages`, so the site is in hand. An id that
does not resolve in that site's register is **inert and logged, never fatal** —
PBP-037's rule, and the only one that survives the same tool function existing on
two sites. Today `mortgages-stamp-duty` (loanandmortgage) and this site's
`tool-stamp-duty` are the same calculator under two keys; that is luck, not
design, and the resolution rule is what makes the luck stop mattering.

### Piece 3 — THE FAN-OUT: a fact that moves reaches the tools that encode it

`refresh_evidence_base` already raises `stale_evidence`, and the citation half
already classifies outcomes three ways (quote found / **citation lost** /
fetch error = UNKNOWN). Extend the drift path to emit one item per tool whose
current PLAN declares the drifted id, on that site.

**The split that must not be collapsed:**

- **value drift** (a re-derived number differs) → `improve_tool`, carrying the
  fact's new value and `writer_line` in the spec. The tool must *change*.
- **`citation_lost` / fetch error** → **human review, never an automated
  rewrite.** A 404 at GOV.UK is not evidence the number moved. Pointing a
  rewriter at a calculator on that evidence is `bugs_open/126`'s failure —
  a false failure aimed an automated fixer at a tool's disclaimer — with
  arithmetic as the target instead of a disclaimer.

Dedup key `fact_drift:<fact_id>:<subject_key>:<site>`. Per the owner ruling of
2026-08-02 §1, converging a new producer onto an existing `item_type` needs no
RFC **provided the producer set and the `item_key` shape are named in the concept
register entry** — so that naming is part of the commit, not a follow-up.

This is the literal answer to the owner's legislation question: legislation moves
→ the daily sweep sees it → the calculators that encode it are named that day.

### Piece 4 — THE ORACLE: expectations computed from the register

The strongest piece and the only one that closes F3. Two routes; take the first.

- **4a (recommended) — emit at criteria time.** A Go action computes expected
  outputs at chosen vectors *from the register's values* and writes the fence.
  The browser runner stays dumb. Needs an oracle library in Go — annuity and
  SDLT bands to start; `loanandmortgagecalculator/oracles.py` is the worked
  reference, and its four-outcome reporting (PASS / CONVENTION / FAIL / N/A) is
  the part worth porting, not just the formulas.
- **4b (rejected) — resolve at run time in the browser runner.** The adapter
  imports no database driver at all (`adapter.go`); giving it one is a large new
  authority on a shared seam, for no gain over 4a.

**Use a NEW check type; do not overload `computed_values`.** Nineteen fences carry
that type today, nine installed within the last two hours. Feeding them
register-derived values would silently change what every existing green fence
*claims* — from "unchanged since capture" to "matches the published rule" —
which is precisely RFC_002's stated trigger for an architecture round. A distinct
type leaves the old claim intact and makes the new one legible. §6 routes it.

**The register must first be shaped so an oracle can read it.** Today
`sdlt-standard-bands` is `value: 12` (the top rate) with the bands in prose. No
oracle can derive a band table from a `claim` string. Two ways out:

- **(i) recommended — one fact per threshold and rate.** SDLT becomes ~12 facts
  instead of 4. **No schema change at all**, each band edge gets its own verbatim
  GOV.UK quote (better evidence, not worse), `value` stays scalar — which is what
  `numberSupported` and `writer_line` already assume — and every band number
  becomes a registered number the *copy* gate can support for free.
- (ii) a structured `parameters` key on a fact. It would survive (both write
  paths use `map[string]interface{}`), but it invents a second shape for
  "a number we vouch for" and buys nothing (i) doesn't.

---

## 3. What we deliberately do NOT build

- **No static scan of tool JavaScript for unregistered numbers.** Every constant
  in a calculator — 12, 100, 0.01, a viewport width — would need whitelisting.
  The false-positive rate trains people to ignore the alarm, and an ignored alarm
  is worse than none.
- **No LLM judging arithmetic.** The whole value here is that the check is
  deterministic and can be induced to fail.
- **No Tier-2 form of any of this.** `experience_criteria.go:73-88` explains why
  a statically-green arithmetic check is worse than no check; that judgement
  stands and this plan does not reopen it.

---

## 4. Sequencing

**Phase A — data and config. No Go, no image roll, live on apply.**

- A1. Re-shape the SDLT facts to one-per-threshold (§2 Piece 4(i)). Extract
  quotes **programmatically from the fetched page**, never retyped — the
  emission-rewrite trap that NOTES 08-09 already records.
- A2. Check the first `evidence-freshness` sweep, due ~08:58Z 2026-08-10.
  **A `citation_lost` on day one means our quote extraction differs from
  `VisibleTextFromHTML`, not that legislation moved** — fix the quote.
- A3. Piece 1: the three prompt seeds.
- A4. Create tool PLANs for the twelve recreated tools (prerequisite for
  everything in Phase B, and it turns Tier 2/4 on for this site for the first time).
- A5. Re-run the replay comparator now the five builds have landed — bridging and
  forecaster should now match the golden. **Owed from the previous handoff.**

**Phase B — the edge.** Pieces 2 and 3. Go + validator + one seed.

**Phase C — the oracle.** Piece 4a behind an RFC (§6).

---

## 5. Landmines the implementation must respect

1. **`doc_plans` has no `site_id`.** Resolve fact ids against the driven page's
   site, never against the PLAN. §2 Piece 2.
2. **Never round-trip the register through the typed struct.** `EvidenceBase` /
   `EvidenceFact` do **not** model `citation`, `writer_line`, `unit`,
   `staleness_days` or `writer_block`. Both live write paths
   (`refresh_evidence_base_action.go`, `evidence_citations.go`) use
   `map[string]interface{}` precisely so those survive. A new consumer that
   parses typed and writes back would silently delete every citation on the site.
3. **`ParseEvidenceBase` returns `nil` for "not opted in"** (no facts *and* no
   banned claims). A new consumer must treat nil as a silent no-op.
4. **`computed_values` compares text verbatim.** A register-derived expectation
   must be formatted the way the tool formats. The neighbouring lane already
   spent a round here — six mismatches that were *its own* rounding convention,
   not the site's (`WRONG_CALLS`, 08-09). State the formula/convention
   distinction before changing any checker to agree with a page.
5. **SKIPPED IS NOT PASSED.** A fact-derived check that cannot resolve its facts
   must fail or skip *loudly*. A verdict needs ≥1 passed and 0 failed.
6. **A concurrent lane owns 9 of the 19 live fences and is still moving.**
   Re-measure `doc_plans` at implementation time; do not carry today's counts.

---

## 6. Council and register routing

| piece | scope | route |
|---|---|---|
| 1 (prompts) | config only, no platform paths | no gate required; register entry if it becomes a shared mechanism |
| 2 (`facts` key) | additive, inert until a document names it — RFC_002 §9 Q1 says this is **not** an RFC trigger | **council gate**, consumers named |
| 3 (fan-out) | new producer converging on existing item types | **council gate** + concept-register entry naming the producer set and `item_key` shape (owner ruling 2026-08-02 §1) |
| 4 (oracle) | changes what a green arithmetic check *claims* | **RFC to `architecture_review/` first**, then council. Using a new type name is what keeps the existing fences' claim intact — say so in the RFC |

Per the 2026-07-29 ruling §3, name the other consumers and **tell them**: the
loanandmortgagecalculator lane (19 fences, 9 installed today), the experience
register (`experience_patterns.criteria_template`, the second document reaching
the same evaluator), and `verify_site_experience`.

---

## 7. Verification — and how each check could come out false

- **Before the roll:** pod-grep the new symbol = 0, *with a positive control in
  the same exec*, so the 0 is a real absence and not a broken grep.
- **The behavioural proof is an INDUCED one, not an observed one.** Change
  `sdlt-ftb-relief-cap` on a canary register from 500000 to 550000 and assert the
  pipeline names the stamp-duty tool; roll it back. Waiting for real legislation
  to move is not a test plan. (The finetuning lane's 08-09 verify block was
  proven exactly this way.)
- **A control that must come out red**, `oracle.py`'s discipline: mutate one
  register-derived expectation and require the fence to fail. A green run under
  mutation means the check is inert.
- **The NO-OP case, not only the damage case:** a site with no `evidence_base`
  row must behave byte-identically. Six live sites have no register; that is the
  population this must not disturb.
- **Disconfirming shape for Phase A:** if, after A3, a rebuilt calculator still
  encodes a number absent from the register, Piece 1 has failed and Pieces 2–4
  are load-bearing rather than belt-and-braces. Record which it was.

---

## 8. The honest limit of all four pieces

Every mechanism here rests on the register being **right**. Nothing above can
tell a correct threshold from a confidently-wrong one; the citation check proves
a source *said* it, which is provenance, not correctness — the distinction the
council's compliance seat forced into `content-quality` CQ-022 twice before the
owner ruled on it. So the register's SDLT facts deserve a human read before
Phase C makes them authoritative over arithmetic, and the RFC should say so
rather than let the mechanism imply a guarantee it cannot carry.
