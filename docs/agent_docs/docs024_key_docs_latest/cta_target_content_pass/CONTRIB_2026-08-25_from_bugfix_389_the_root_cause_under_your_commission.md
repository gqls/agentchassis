# CONTRIB 2026-08-25 — from the `bugfix_389_cta_relevance` lane: the cause under your commission, and it should change your Phase 1 scope BEFORE you run it

**Told, not merely measured** (owner ruling 2026-07-29 §3). Nothing here criticises the plan — your
2026-08-15 read of the population was right, and your mechanism note (*"the positional fallback
chose the site's top-ranked interactive page"*) is exactly correct. What follows is **why** that
page is top-ranked, which turns out to be actionable in a way that changes what your pass has to do.

## 1. The owner has withdrawn the floor

On **2026-08-25** the owner reported `/tools/password-entropy.html` as a CTA and said it is *"not
deliberate and actually wrong."* Your plan records the owner **accepting it as a floor** on
08-15 and commissioning your pass to raise it. Read the new report as the floor being withdrawn:
the off-topic case is now a defect (`bugs_open/389`), separate from the repetition your pass is
really about.

## 2. The cause is a fossil integer, not a content problem

`chooseCTATargets` (`resolve_internal_links_action.go:651`) sorts candidates by
**`COALESCE(nav_order,100)` then `name`** and returns `[0]`. On the three sites where
password-entropy is the modal target it carries **`nav_order = 1`**, set when the page was created
on **2026-03-13**, while every genuinely relevant tool sits at **6–204**:

| site | password-entropy | relevant tools it beats |
|---|---|---|
| `ai-agent-orchestration.com` | **1** | ROI estimator, LLM cost calculator, build-vs-buy (200–202) |
| `finetuning.uk` | **1** | AI readiness quiz, GDPR risk assessment, model-approach selector (200–204) |
| `leopardessconsulting.co.uk` | **1** | ROI estimator (6), LLM cost calculator (7), vendor trust (8) |

## 3. Why that matters to YOUR phasing specifically

Your Phase 1 canary is **finetuning.uk — the worst offender at 39 rows**. Those 39 rows share one
cause, and it is one number. **Correcting `nav_order` moves the positional fallback for every
label-less CTA on the site in a single step**, with no LLM involved and no wording changed.

So the canary as designed would measure the wrong thing: it would spend a `content_rewrite` over
the site's pages and attribute the improvement to reworded labels, when a one-row `UPDATE` would
have moved most of them. **Suggested reordering** — and it makes your pass cheaper and its result
attributable:

1. correct the ranking input on the three fossil sites (or take `389`'s opt-out option);
2. **re-measure your 16-site / ≥6-rows population** — it is 10 days old and step 1 changes it
   substantially;
3. scope the content pass against what actually remains, which is the genuine repetition problem
   your commission is for.

## 4. Two things that constrain step 1, so you do not inherit a new bug

- ⚠ **`nav_order` is overloaded.** It orders the visible nav menu *and* ranks CTA candidates.
  On `ai-agent-orchestration.com` the page is `in_header = true`, so changing its `nav_order`
  **moves the menu item too**. On the other two it is `in_header = false` and the change is
  invisible to visitors. That coupling is `389`'s core finding and the reason a data fix is not
  obviously the right answer.
- ⚠ **`in_header` is not read by the chooser at all.** Someone already hid this page from one
  site's nav (`docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql:29`, comment: *"a password
  tool doesn't belong in the primary nav"*) and **it changed nothing**. Do not expect hiding to
  help.

## 5. It is still minting, which your plan could not have known

The `__cta_minted` stamp (LNK-035) only shipped **2026-08-22**, a week after your measurement.
Using it now: of the 80 CTA url fields pointing at this tool, **17 are stamped as resolver-minted
and dated 2026-08-23 → 2026-08-25 (today)**. So this is live behaviour, not inherited state —
any repair, yours or `389`'s, must land the mechanism decision first or the next run re-mints.
⚠ And read the stamp carefully: **NULL means "not recorded", never "authored"** (no backfill, by
design), so the 39 unstamped fields are unattributable rather than human-written.

## 6. Your step 2's known gap, restated because it now matters more

Your own NOTES already record that the detector computes the right destination into
`suggested_target` and **nothing reads it**, and that the rerender re-derives from the narrower
set. If step 1 above moves the positional fallback to a sensible tool, that gap costs you less —
one more reason to do the ranking first.

**Full finding, with the served-bytes evidence and the four owner decisions:**
`bugs_open/389_HANDOFF_2026-08-25_cta_destination_is_ranked_by_nav_order_alone_so_an_off_topic_tool_wins_every_primary_button.md`.
**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_389_cta_relevance/`.
