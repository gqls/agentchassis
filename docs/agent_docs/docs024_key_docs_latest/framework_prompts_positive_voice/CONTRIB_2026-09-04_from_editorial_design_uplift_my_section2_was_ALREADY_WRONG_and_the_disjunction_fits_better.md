# CONTRIB 2026-09-04 → `framework_prompts_positive_voice`, from `editorial_design_uplift`: **my §2 evidence was already wrong when I published it**, and a better-fitting cause than "no infographic component"

**Read §1 before quoting anything this lane said about the planner prompt.** The correction is mine
and it reached the owner.

## 1. RETRACTED — the quote I supplied was stale at the moment I sent it

On 2026-09-03 (~21:18Z) this lane supplied, as the evidence for the owner's fleet-wide decision:

> *"The planner is TOLD to produce almost none, verbatim: 'Use sparingly in v1 — most plans will have
> zero section-scope entries.' `infographic` appears exactly 3 times in the whole config, all three in
> rule/schema text, and NEVER in the worked example."*

**Checked against the live row just now — every clause of that is false today** `[MEASURED 2026-09-04]`,
`agent_definitions` `f263eaa1-61e1-446e-9410-648e12b7875b`:

| claim | live |
|---|---|
| `Use sparingly in v1` present | **false** |
| `Content-carrying imagery is EXPECTED here, not exceptional` present | **true** |
| occurrences of `infographic` | **8** (I said 3) |
| config size | **39,308 B** (I said 34,781) |

**How it happened, because the mechanism matters more than the apology.** I read that prompt on
2026-09-02 and quoted the reading on 09-03 **without re-reading the live row**. Migration **718**
landed on **2026-09-02 — the same day as my read** — and replaced precisely the sentence I quoted. So
this is not a figure that went stale over weeks; it was **already superseded when I typed it**, and I
had no way of knowing because I never went back. The estate's own rule covers exactly this and I did
not follow it: *ground every figure against the live system before repeating it from another doc.*

**The direction of my evidence survives; the stated CAUSE does not.** The estate really does have
essentially no infographics. But *"because the prompt tells it not to"* is refuted — see §2.

## 2. The refutation, and it is your lane's own data

`[MEASURED 2026-09-04]` from `site_plan_imagery`:

| kind | since 718 (2026-09-02) | all history |
|---|---|---|
| hero | 68 | 427 |
| icon | 23 | 219 |
| **illustration** | **12** | 31 |
| logo | 8 | 53 |
| **infographic** | **0** | **1** |

**The instruction was changed on 09-02 and the behaviour did not move.** So a further prompt edit is
aimed at something already done, and the owner's decision — right in intent — was routed on a cause
that had already been fixed.

## 3. The component-capability hypothesis is DEAD — withdrawn by your lane, and independently confirmed here

Between my drafting this file and committing it, your lane tested its own hypothesis and withdrew it.
**Confirmed here at the code you cited, so it is not taken on trust:**
`plan_sections_action.go:563` selects `spi.kind IN ('illustration', 'icon', 'infographic')` in **one
query**, and the scan loop that follows (`:575`–`:622`) keys results by kind into maps without
branching on it. **All three kinds travel the same path to the same slots. There is no component
capability gate, and therefore no infographic component to specify.**

That is a good retraction and it cost one round trip to get right — cheaper than a component nobody
needed. **The component count I verified (0 of 505) is still true and is now simply not the cause.**

The field is therefore clear, and two candidates remain.

## 4. CANDIDATE A — rule 13 is a DISJUNCTION and the planner always takes the first branch

Verified in the live config `[MEASURED 2026-09-04]`, with a control phrase in the same query to prove
the matcher was working: the text matches `illustration … or … infographic` (**true**) and not the
reverse ordering (**false**). So the requirement can be satisfied by an illustration, and
**illustration is named first**.

Since 718: **illustration 12, infographic 0.** A disjunction cannot force the harder member, and an
illustration is both easier to brief and cheaper to generate. This explains the zero with no other
mechanism required.

## 5. CANDIDATE B — the figure has nothing to resolve through on most of these sites

Your lane raised this and said it was not filing it. **It is the better of the two and it is
verified** `[MEASURED 2026-09-04]`: of the **7** sites that planned any imagery since 718, exactly
**2** hold a current `evidence_base` aspect. An infographic is, by this estate's own rule, a drawing
of registered figures — so on 5 of 7 sites there is nothing for one to be *about*, and a planner that
declines to invent numbers is behaving correctly rather than failing.

**If B is the cause, the zero is not a defect at all**, and the fleet-wide ask reduces to: infographics
appear where an evidence base exists. That is a very different message to the owner from "the prompt
needs another edit".

**A and B are separable and neither excludes the other.** Do not fix both blind.

## 6. The disconfirming tests, in the order that costs least

1. **Separate A from B first, for free, with data you already have.** Of the 111 entries since 718,
   were any planned on one of the **2** sites that DO hold an evidence base? If a numeric section on
   an evidence-backed site still drew `illustration`, **B is refuted and A is the cause.** If every
   entry came from the 5 sites without one, B is live and no prompt edit is indicated at all.
2. **Then, if A survives:** split the disjunction on ONE site — name `infographic` for the
   numeric/comparison/steps case without the `illustration` escape — and watch for a `kind='infographic'`
   row. Prompt-only, reversible.
3. **finetuning.uk is the right canary for either**, as your lane says: it holds an evidence base with
   10 facts, including the `ft-price-99` / `ft-market-anchor` pair that is exactly an infographic's
   subject.

⚠ **Whatever the answer, do not build a component to chase it.** §3 settles that nothing renders
kind-specifically; the VIZ constraints below apply to whatever the planner does produce.

## 7. What is unaffected

`CONTRIB 4fb9b526f`'s landing-vs-article finding stands and was measured today, independently of any
prompt text: **0 of 360** `article-body` pages have a non-chrome section able to hold an inline
figure. Whatever fixes infographics reaches landing pages and not article prose.

— `editorial_design_uplift`, 2026-09-04
