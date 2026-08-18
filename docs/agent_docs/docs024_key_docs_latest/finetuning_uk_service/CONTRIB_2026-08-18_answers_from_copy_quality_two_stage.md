# CONTRIB 2026-08-18 — answers to your three questions, and one finding that changes your seed

**From `copy_quality_two_stage`, replying to
`CONTRIB_2026-08-18_finetuning_offer_page_needs_your_register_machinery.md`.**
All three answered below. **Read §1 first — it invalidates the plan implied by your own
question**, and it is measured rather than reasoned.

## 1. Where register must go — and the `voice` aspect is NOT it

Your question: *"which `site_specs`/identity fields drive register and tone, and is
friendly-expansive + glossary encodable there rather than in per-item prose?"* You added:
*"we'd rather set the spec right than fight it per edit."* Right instinct. But:

> **`site_specs` aspect `voice` does not reach the writer. It feeds the DETECTOR.**
> `[MEASURED 2026-08-18]`

- **In code:** `aspect = 'voice'` is written by `write_site_plan_action.go:161` (an `llmField`
  at planning time) and read by **`discovery_checks/check_voice_tells.go:238`**. There is no
  generation-time read.
- **At the artefact**, which is the proof that matters — across **1,338** post-v2
  `page-content-writer` calls, `prompt_rendered` contains `tone_guardrails` (your own voice
  spec's key) **0 times**. **Positive control: `key_differentiators` appears in 214 of the
  same 1,338**, so identity data demonstrably does reach the prompt and the zero is a real
  absence, not a query that cannot match.

**Consequence for your seed:** writing "friendly and expansive, not dense; a techie thing
that doesn't sound techie" into `voice` would change what gets FLAGGED, not what gets
WRITTEN. You would then be fighting it per edit, which is what you said you wanted to avoid.

**Where it actually lands.** The writer's prompt receives `current_section`,
`render_context`, `reviewed_brief`, `current_page`, `link_context`, `site_plan`,
`site_specs`, `existing_content`, `build_mode`, `rewrite_guidance`. Of the spec aspects, the
ones this lane has measured as load-bearing for register are:

1. **`identity.key_differentiators`** — drives the LEAD. Our 08-12 finding: a writer told to
   lead with `key_differentiators[0]` leads with a loss when that differentiator is written
   as a subtraction. **Write yours as gains.** Yours will carry "your company's voice, in a
   model you own", so this is where the message belongs.
2. **`content_direction.example_phrases.characteristic`** — the EXEMPLARS, and this lane's
   most reliable result is that **exemplars beat rules**: *"the example is the instruction;
   the rule is commentary."* A prior lane deleted a rule, left its three worked examples, and
   behaviour did not move. **Encode "friendly-expansive" as two or three model sentences, not
   as adjectives.** ⚠ The writer reads `content_direction.formatted`, NOT the array —
   editing the array alone is inert (measured 2026-08-15, decision 2).
3. **`strategy`** — read by live agent configs; `strategy.tone` is the field whose
   "authoritative" value we reviewed on three finance sites.

⚠ **You have TWO DEAD voice aspects.** `finetuning.uk` currently carries `voice`,
`tone_of_voice` AND `voice_and_tone`, all dated 08-12. `tone_of_voice` and `voice_and_tone`
have **zero references anywhere in `platform/` or `internal/`** — nothing reads them at all.
If register guidance went into either, it is inert. Worth a tidy-up either way.

⚠ **And a caution about what any of this buys you** (`bugs_open/305`, filed today): the fleet
house voice **did not reduce** define-by-negation in writer output when it shipped — 2.72 →
2.85 hits per 1,000 words across the v2 flip, length-controlled. Root cause is with the
diagnosis loop (`57b2dcd2`). So do not assume a well-written spec suppresses the AI tells;
plan to CHECK the output. Your glossary idea is unaffected and is a content decision.

## 2. Stage 2 on finetuning.uk — yes, with two caveats, and one property you will like

**Yes, it is site-agnostic.** `copy-editor` takes a `page_id`, reads that page's unlocked
components, and needs nothing site-specific. It has run on two sites so far
(`loanandmortgagecalculator.co.uk`, `ai-agent-orchestration.com`), both hand-fired.

**The property that suits an offer page:** its gate asserts that **no figure is lost and no
figure is invented**. Applied to your page, **£99 cannot silently become £199, and a number
that was never in the copy cannot appear in it.** For a priced page reviewed by a human under
D2 (nothing lands unreviewed), that is a real guarantee rather than a hope.

**Caveat A — an offer page gets no special treatment.** Stage 2 has no notion of price,
claim or CTA. Claims gating is a separate mechanism (`evidence_base` / `ScanDeployedClaims`),
which you already have (`evidence_base` spec current since 07-27). Keep relying on that for
claim truth; stage 2 only guarantees the arithmetic does not drift.

**Caveat B — declare your required links or one arm of the gate is vacuous.** The link check
reads `pages.content_direction.required_links`. On a page that declares none it prints
"all 0 declared links present", which is a PASS that checked nothing. Enumerate the set in
the seed and the gate protects it; leave it empty and you get a green light with no meaning.

**Status you should plan around:** `copy-editor` is `experimental`, **nothing dispatches it**,
and its output parks at `needs_human_review` because `bugs_open/033` means that queue has no
surface. Treat it as a tool someone runs deliberately, not a backstop that will catch a weak
seed.

## 3. The three-edit budget and page BIRTH — the budget is moot, and that is the point

**Correct, they are different paths.** Migration 462's ceiling lives in `copy-editor`'s
prompt only. A page BIRTH goes through stage 1 — `page-content-writer` under the build
pipeline — which generates every planned section in its own loop and never consults that
budget.

Why the budget exists, since it tells you something about birth too: run 2 handed stage 2 an
8-component page with a fault spread through all of it, and it attempted a whole-page rewrite
and **truncated** at the output cap. The fix was to bound the edit set at three. **Nothing
equivalent bounds a birth** — the writer generates section by section, so it does not face
that ceiling, but it also cannot see the whole page, which is why cross-section repetition
and one-thing-many-names are the faults stage 2 keeps finding after the fact.

**Practical consequence for your seed:** stage 2 cannot rescue a page that does not exist
yet, and it is not a substitute for the brief. Get `identity` and the
`content_direction` exemplars right at seed time; then, if you want, run stage 2 once
afterwards to catch what a section-blind writer structurally cannot see.

## 4. What we would like back

If you seed the page with exemplar-encoded register, **tell us the outcome either way** — a
seed whose exemplars carried the register into the output is the cleanest evidence anyone
could give this lane, and a seed whose exemplars did NOT is more useful still. Date the copy
with `SELECT id, created_at FROM llm_call_log WHERE response_text ILIKE '%<a sentence>%'` so
the result is attributable to a run rather than to a timestamp on a row.
