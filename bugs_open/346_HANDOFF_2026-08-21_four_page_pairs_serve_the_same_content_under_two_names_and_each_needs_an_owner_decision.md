# 346 — four page pairs serve one page under two live names; the minting is fixed, the existing duplicates are not, and each pair needs an owner decision rather than a fix

**Filed 2026-08-21**, spun out of `bugs_closed/215` when that bug closed. 215's
defect — the framework *minting* a second identity for one page — is fixed, live
and behaviourally proven. **This file is the damage that predates the fix**, and
it is deliberately separated because what remains is not engineering: it is four
per-pair decisions only the owner can make.

Continues the "O2" remediation begun in the brochure lane. **Three of the original
seven pairs are resolved** (`gripper-payload-calculator` and
`gripper-cycle-time-estimator` on robot-hands, completed 08-14/08-15 with full
retraction; one on ai-agent-orchestration.com no longer presents as a pair). **Four
remain.**

## The four, measured 2026-08-21

All eight URLs HTTP-tested the same day: **all serve 200 with real content**, against
a fabricated-URL 404 control of 2,886b on robot-hands. So every one of these is a
live page a visitor or a search engine can reach — not a phantom.

| site | pair | components | policy | bytes served |
|---|---|---|---|---|
| finetuning.uk | `ai-readiness-quiz` `/ai-readiness-quiz.html` | 2 | generic | 39,551 |
| | `tool-ai-readiness-quiz` `/tools/tool-ai-readiness-quiz.html` | 1 | **owned** | 36,367 |
| fundamentallyai.com | `automation-savings-estimator-guide` `/blog/…` | 3 | generic | 27,854 |
| | `tool-automation-savings-estimator-guide` `/guides/…` | 3 | generic | (200, content) |
| fundamentallyai.com | `model-approach-selector-guide` `/blog/…` | 3 | generic | 27,010 |
| | `tool-model-approach-selector-guide` `/guides/…` | 3 | generic | 32,934 |
| robot-hands.com | `matchmatrix` `/matchmatrix.html` | 4 | generic | 79,520 |
| | `tool-matchmatrix` `/tools/matchmatrix/index.html` | 1 | **owned** | 100,552 |

> ⚠ **Do NOT choose the survivor by component count.** `tool-matchmatrix` has **one**
> component and serves **more** bytes than its four-component twin. A `page_components`
> count is a container count, not a content measure — the landmine is on file, and this
> pair is a textbook instance of it. Open the served page.

> ⚠ One fetch returned `HTTP 000` on the first attempt and **200 on retry**. A single
> failed curl is not evidence a page is dead; re-fetch before recording an absence.

## What the decision is, per pair

**Which name owns the page**, and therefore which URL survives. That is all. The
framework will not choose: when both spellings are realised and both are in the plan,
every twin-identity layer deliberately REFUSES (register `PLAN-048`'s landmine), because
snapping would hand the writer two entries with one name and the dedup would then evict a
live page.

**The constraint that should be read BEFORE choosing, not after:**

- **There is no redirect mechanism on this estate.** `link_registry` and `redirects` are
  empty fleet-wide. So retiring one name of a pair makes that URL a **permanent 404** for
  anyone holding the link — a bookmark, an inbound link, a search result. This is the
  single fact most likely to change a decision, and it is why this is an owner call.
- **`finetuning.uk` is doubly constrained.** It is on the decomposed-site list
  (`bugs_open/204`), so its pages carry positional slot names and it must not be replanned
  until 204 is fixed. Its pair is the one to leave until last.
- **An archive is durable now, but was not.** `bugs_closed/266` (four producers rebuilding
  archived pages) is fixed, live and behaviourally proven, so archiving one side now holds
  — that was the blocker when this remediation stalled, and it is discharged.
- **Retraction is a separate step from archiving.** Archiving removes the page from the
  rebuild population; it does **not** delete the served file. The worked, twice-validated
  procedure is the runbook's 8 steps, and step 6 is the `delete_file` retraction.

## How to execute a decision once made

`docs/agent_docs/docs024_key_docs_latest/brochure_component_library/RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md`
— eight steps, proven end-to-end on two pairs. Its two procedural corrections are load-bearing:
"open" work items means `workItemClosedStatuses` (not `workItemTerminalStatuses`), and step 4
is not durable alone because the fleet sweep re-queues a rerender per **active** page.

**Run the inbound census before mutating anything.** It carries its own positive control and
has correctly predicted both a refusal and a clean pass — it is the step that tells you whether
retiring a name will break a link that something else on the site depends on.

## How to verify this file is finished

Zero rows from the pair census (both sides `active` + `deployed_at IS NOT NULL` sharing a
role-prefix-stripped stem), **and** each retired URL HTTP-tested to 404 at the site's control
byte size, **and** the survivor unchanged to the byte. Re-run the census before believing a
zero: this file's own population dropped from 7 to 4 partly by remediation and partly by
causes nobody recorded.

## Relations

`bugs_closed/215` (the minting defect, fixed — this is its residual damage);
register `PLAN-048` (the seam that refuses these pairs by design, and why);
`bugs_open/204` (constrains finetuning.uk); `bugs_closed/266` (archive durability, the
discharged blocker); `bugs_open/098` → `RFC_011` (retiring a live page has no mechanism —
the reason a redirect does not exist).
