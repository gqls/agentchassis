# BASELINE (pre-gate) — finetuning.uk `index` (the homepage), captured 2026-09-03 10:15Z

The page the owner read and called *"written by AI with all sorts of negativity"*. Pinned
because the copy gate **mutates `content_data` in place**, so the rebuild he ordered
(work item `1513b86a`, dispatched 10:13Z, still `triaged` and unclaimed at capture —
verified, so this IS the pre-gate state) destroys it permanently.

Measured on `page_components.content_data`, the surface the gate rewrites. Compare the after-
pass here, never against the served-HTML figures in `NOTES_two_stage_copy.md` — see the
surface caveat in `BASELINE_2026-09-03_technical_details_pre_gate.md`.

| shape | hits in content_data |
|---|---|
| `x_not_y` | **8** |
| `rather_than` | **1** |
| `instead_of` | **2** |
| `negative_reveal` | **1** |
| `nothing_unless(v3 cand)` | **1** |
| `so_consequence(v3 cand)` | **1** |

**Total: 14** across 5 components.

## His worked example, and the ruling that applies to it

> *"We're not tied to one provider, so you get the model that fits the task, not the model we happen to sell."*

Owner's ruling, relayed verbatim 2026-09-03: **"Keep 'We're not tied to one provider'"** —
cut at the comma. He was given both readings and chose truncation over removing the `not`
from the first half, so the surviving clause is a sanctioned negative opening frame and the
house-voice "never open by saying what something is NOT" rule does **not** override it here.
The sentence trips `x_not_y` (`", not t"`) and `so_consequence`; the register's `x_not_y`
treatment (truncate, keep the first half) already produces exactly his ruling.

## The sentences, verbatim (what the after-pass diffs against)

- **`features`** · `nothing_unless(v3 cand)`
  > Nothing runs unsupervised unless you decide it should."}, {"icon": "trending-up", "name": "Gets better after launch", "description": "Our systems run automated quality sweeps between review cycles.
- **`features`** · `x_not_y`
  > That doesn't make the output foolproof, but it does mean you can see where a claim came from."}], "headline": "What you actually get", "subheadline": "Systems built around your business and your data, not around one AI provider's product line.
- **`features`** · `negative_reveal`
  > That doesn't make the output foolproof, but it does mean you can see where a claim came from."}], "headline": "What you actually get", "subheadline": "Systems built around your business and your data, not around one AI provider's product line.
- **`differentiators`** · `x_not_y`
  > {"features": [{"name": "The right model for the job, not the only one we offer", "description": "We use whichever model does the job best.
- **`differentiators`** · `x_not_y`
  > We're not tied to one provider, so you get the model that fits the task, not the model we happen to sell."}, {"name": "Privacy comes first, not as an afterthought", "description": "Your data doesn't need to leave your building to get value from AI.
- **`differentiators`** · `so_consequence(v3 cand)`
  > We're not tied to one provider, so you get the model that fits the task, not the model we happen to sell."}, {"name": "Privacy comes first, not as an afterthought", "description": "Your data doesn't need to leave your building to get value from AI.
- **`differentiators`** · `x_not_y`
  > Ask us what running your own private AI setup would look like."}, {"name": "Open models, not a walled garden", "description": "We work with open-weight models like Llama, Mistral, and Phi, and fine-tune them on your specific domain with custom embeddings.
- **`differentiators`** · `x_not_y`
  > If a better open model comes along next year, you can move to it."}, {"name": "You stay in control of what gets automated", "description": "Automation should save you time, not hand over judgement you didn't agree to give up.
- **`differentiators`** · `x_not_y`
  > It's a mechanism for catching mistakes early, not a promise that every output is correct.
- **`differentiators`** · `instead_of`
  > You should still read what comes out before you act on it."}], "headline": "Why businesses work with us instead of building this alone"}
- **`case-studies-grid`** · `instead_of`
  > We built a pipeline that collects and structures that data automatically, so analysts start from a finished table instead of a search bar.
- **`case-studies-grid`** · `x_not_y`
  > Each project below started as a specific business problem, not a technology wish list.
- **`departments-grid`** · `rather_than`
  > We also set up the agent frameworks underneath, so your team can run these systems in production rather than just look at a demo."}, {"icon": "compass", "name": "Strategy, Sites & Implementation", "subtitle": "From first conversation to working system", "description": "Most companies don't need a lecture on AI.
- **`hero`** · `x_not_y`
  > We pick whichever model actually does the job, not whichever vendor we're tied to.

> **CORRECTED 2026-09-03 ~10:25Z, before either rebuild ran — I overstated the repair above.**
> The line "the register's `x_not_y` treatment (truncate, keep the first half) already produces
> exactly his ruling" is wrong in a way that matters for reading the after-pass. **"Truncate" is
> the register's TREATMENT STRING, not the implementation.** `rewrite_negations_action.go` asks
> the model **once** to rewrite the offending sentences directly, then judges each proposed
> rewrite with `AcceptNegationRewrite`, which **fails closed** — a rejected rewrite leaves the
> ORIGINAL sentence in place. So the gate is not guaranteed to reproduce his cut-at-the-comma
> ruling; it may rephrase instead, and it may keep the sentence untouched. Whether his sentence
> comes out as he ruled is **empirical, and the after-pass is the measurement** — do not record
> it as expected-and-confirmed.

## What a PASSING after-pass looks like (settled from the code, before the run)

Verified in the live config, so the 2026-08-21 silent no-op does **not** apply here:
`rewrite_negations.output_field = copy_gate` and **`render_section.content_from =
copy_gate.result`** (not `generated_content.result`). The repair genuinely reaches the render.
⚠ Reading that wiring needs the path `steps.process_sections_loop.config.sub_workflow.steps.*` —
querying `steps.process_sections_loop.steps.*` returns `(unset)` for every key and reads exactly
like "the repair is unwired", which is how a false alarm starts.

**The budget is inert, so residue is not automatically a failure.** `page_budget: 2` is only
spendable by MILD shapes (`rewrite_negations_action.go:263-270`: *"A sharp shape now never
consumes the budget and is always repaired; a mild one is tolerated up to `page_budget`"*), and
`mildNegationShapes` has been EMPTY since Decision A. So every non-exempt, non-headline hit is
repaired, and the known per-section-vs-per-page counter bug is moot while the set is empty.

That leaves exactly **two legitimate reasons for a surviving hit**, and both are recorded rather
than silent:
1. **Exempt** — brief-supplied (the seven `defaultBriefFields` paths) or regulatory/capability
   negations. These are counted and reported, never rewritten; fixing them means fixing the
   brief, which is the site lane's call.
2. **Rejected rewrite** — `AcceptNegationRewrite` failed the model's proposal closed, keeping the
   original. Every rejection carries its reason, and the header calls that log *"the instrument:
   it is how we find out whether the repair is fixing the copy or teaching the model a new tic"*.

**So the after-pass must read the rejection reasons, not just re-count shapes.** A count alone
cannot tell an exemption from a rejection from a repair that never ran, and those three want
three different responses.
