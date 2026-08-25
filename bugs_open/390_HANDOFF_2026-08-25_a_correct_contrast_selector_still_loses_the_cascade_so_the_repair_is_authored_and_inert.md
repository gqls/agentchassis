# 390 — a CORRECT contrast selector still loses the cascade: css-patch-agent appends a rule that cannot win, and the item completes

> ## STATUS 2026-08-25 — OPEN, LIVE, REPRODUCIBLE, AND NOT YET DESIGNED
>
> **This is arm 2 of `bugs_closed/352`, split out at the owner's direction on 2026-08-25.** 352's arm 1
> — the producer inventing a selector that matched nothing — is **fixed, live and proven**, and 352 is
> closed on that basis. This file carries the half that is still biting.
>
> **The one-line version:** even when the filed selector is right, the CSS rule css-patch-agent
> appends is **outranked**, so it is authored, deployed, and marked `complete` while the text stays
> unreadable. The failure mode is identical to 352's from the outside; the cause is unrelated.
>
> ⚠ **The mechanism below is verified FIRST-HAND at the artefact and at the live agent config, not
> inherited from 352's sketch — and the verification CORRECTED the remedy 352 proposed** (§4). Per the
> owner ruling of 2026-07-31, this is the declared substitute for a `090` run: every load-bearing
> claim here is a served-page or live-row observation with its command attached, and the one claim I
> could not settle that way is marked `[UNMEASURED]`. A `090` run is still worth its credits before
> anyone *designs* the fix — see §6.

---

## 1. The mechanism, measured [MEASURED 2026-08-25 ~10:00 UTC]

Worked case: `loancash.co.uk/guides/index.html`, a live `contrast_failure` filed **by the fixed
producer** at 17:33 UTC on 2026-08-24 — so this is not a legacy row, it is what the corrected pipeline
produces today.

The finding says: elements matching `.ported-page-content A` render `rgb(232, 245, 238)` on
`rgb(234,244,239)` — **1.00:1** where 4.5:1 is needed. The selector is correct: an independent parse
of the served page counts **15** matching elements, exactly as the producer recorded.

**Four facts, in the order that makes the failure inevitable:**

1. **The agent appends to the end of one file.** Live `css-patch-agent.save_css_to_db`, read from
   `agent_definitions` (not a seed):
   ```sql
   UPDATE css_themes SET css_content = css_content || E'\n\n' || '/* css-patch-agent … */' || E'\n' || $2 …
   ```
   and `load_current_css` reaches it by `sites → style_collections → css_themes`. Its prompt says
   verbatim: *"The platform APPENDS your rules to the END of the stylesheet above"* and *"Repeat the
   offending selector exactly as it appears above (or more specifically) so your override wins."*
2. **That file is served as `/assets/css/styles.css`, and it is linked at byte offset 8562** of the
   page.
3. **The offending declaration is NOT in it.** `curl` the served stylesheet and grep: it mentions
   `ported-page-content` **0 times**. The declaration lives in an **inline `<style>` block at offset
   12080** — i.e. *after* the stylesheet link:
   ```css
   .ported-page-section .ported-page-content a { color: var(--color-primary); … }
   ```
4. **Specificity, which is the part 352's sketch got wrong.** The offender is
   `.ported-page-section .ported-page-content a` = **(0,2,1)**. The filed selector is
   `.ported-page-content A` = **(0,1,1)**. An appended rule that repeats the filed selector *exactly
   as instructed* is **lower specificity** — it loses before source order is even consulted, and it
   would lose on source order too.

**So the appended rule cannot win, for two independent reasons at once**, and the agent is following
its own instructions correctly when it writes one.

### How to reproduce, end to end

```bash
curl -sS -L https://loancash.co.uk/guides/index.html -o page.html
# 1. the site stylesheet is linked here:
grep -bo '/assets/css/styles.css' page.html
# 2. the offending declaration is emitted AFTER it:
grep -bo 'ported-page-content a {' page.html
# 3. and it is NOT in the file the agent edits:
curl -sS -L https://loancash.co.uk/assets/css/styles.css | grep -c ported-page-content   # 0
```
⚠ **Control first** (parked-domain landmine): an invented path on the same domain must 404.
Confirmed 2026-08-24.

## 2. ⚠ THE THIRD FACT, WHICH CHANGES THE FIX — the offending VALUE *is* in the editable file

`/assets/css/styles.css` contains:

```css
--color-primary: #e8f5ee
```

`#e8f5ee` is `rgb(232, 245, 238)` — **exactly the `fg` the finding recorded**. The declaration is
out of reach; **the token it resolves is not.**

**This refutes the remedy 352 proposed for this arm.** 352's sketch says: *"grep `css_themes` for a
declaration governing the filed selector's property; if the offending declaration is not in the file
the agent can edit, refuse and park."* Run here, that precondition finds no `ported-page-content`
declaration, concludes the fix is impossible, and **parks a finding that one token change would
repair.** A precondition phrased on the *declaration* is the wrong test; the reachable thing is the
*computed value's source*.

⚠ And note what a pale-green link on a pale-green background actually is: a **palette** defect, not a
cascade one. See the `webdesign` colour-churn landmine (`generic_theme` misfires fleet-wide; pin via
`design_intent.palette.reference_values`). **A rule appended to beat the cascade would paper over a
bad token** — which is why candidate (3) below is ranked where it is.

## 3. Why this survives every existing check

- The item completes **honestly** by the workflow's own lights — a rule *was* authored, appended and
  deployed. `bugs_closed/198`'s migrations 542/546 stopped refusals and failures minting `complete`;
  they do not and cannot cover this, where the write genuinely happened.
- **Each spec already carries an `acceptance_test`** naming the exact re-measurement: *"computed
  contrast for elements matching X on Y is at least 4.5:1 — a single-selector, single-page
  measurement, not a site re-audit"*. Confirmed present on the live 2026-08-24 rows. **It is written
  by the audit and read by nothing** — `write_audit_findings_verifier_join_test.go:85` confirms it.
- The next render audit re-measures the same pairing, files it again, and the promoter routes it back
  to the same agent. **The symptom is a finding that keeps returning, not one that fails.**
- 352's fix does not help and was never meant to: the selector is *correct* here. **A correct address
  and an unreachable rule look identical in every log.**

## 4. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Complete only on measured improvement — use the spec's own `acceptance_test`.** Re-measure the
   one pairing at the served page after deploy, at the `checks.GetVerifier` / `verifyBeforeComplete`
   choke point, and complete the item only if contrast actually rose. **This is the only candidate
   that closes the door**: it catches this cause, 352's cause, and every "authored but inert" cause
   nobody has thought of yet, because it asserts the outcome rather than the method. It is also the
   one the data already supports — the acceptance test is written on every row and read by nothing.
2. **Make the agent's rule able to win, deliberately rather than by luck.** It already has the
   ingredients: `load_current_css` returns the whole theme, so the agent can compute the offending
   declaration's specificity from the page and emit at least that, or a `body`-prefixed override (the
   `fixloop_eg_dartsonline` lane's workaround) instead of `!important`. ⚠ This narrows the failure;
   it does not close it, because source order still beats specificity ties and page-level CSS is
   emitted after the theme by construction.
3. **Fix the token, not the rule, when the value is reachable.** Where the offending declaration
   resolves a custom property that IS defined in the editable theme — as here — the honest repair is
   the token. ⚠ **Blast radius**: a token is site-wide by design, so this needs the same "prove it"
   discipline 352's arm 1 got, and it collides with the palette lane. Do not do this blind.
4. **A measurable precondition that refuses and parks** (352's sketch, `mark_base_unsafe`'s shape).
   ⚠ **Ranked last and NOT as written** — §2 shows the declaration-based version parks repairable
   findings. If built, the test must be *"is the computed value's source reachable from the theme"*,
   not *"is the declaration in the theme"*.

## 5. Blast radius — how much of the backlog is this?

[UNMEASURED] The `~1.0x:1` family was 352's stated example, and `bugs_open/296` §10.5 reaches the same
finding from the other end, noting it may explain a subset of its durable parked findings directly:
*processed, the fix was correct, and it never applied.* **Nobody has counted it.** The census is
cheap and is the first thing to do:

```sql
-- how many contrast_failure rows completed against a selector whose offending declaration
-- is not in the editable theme? Start by sizing the ~1.0:1 family, which is the tell.
SELECT status, count(*),
       count(*) FILTER (WHERE (spec->>'ratio')::numeric < 1.1) AS near_1_to_1
  FROM site_work_items WHERE item_type='contrast_failure' GROUP BY 1 ORDER BY 2 DESC;
```
⚠ `site_work_items` is a **rolling window** and `site_work_items_archive` exists (25,281 rows back to
2026-02-22). Any "how many ever" question must UNION it. ⚠ And date the count: a census here goes
stale by addition.

## 6. What I did NOT do

- **No `090` diagnosis run.** Declared substitute under the owner ruling of 2026-07-31: the mechanism
  in §1 is four served-artefact/live-row observations, each with its command, and it corrected the
  inherited theory rather than confirming it (§2). **But §5 is unmeasured and §4 is undesigned** — a
  `090` run is the right spend before choosing between candidates 1 and 2, because the question
  *"which of these completes-without-repairing causes dominates?"* is exactly what it is for.
- **No fix.** Nothing in this file has been built.
- **[UNMEASURED]** whether the inline `<style>` at offset 12080 is emitted per-page or per-component,
  and therefore whether the theme could ever be re-ordered to win. That determines whether candidate
  (2) is worth anything.

## 7. Provenance and where the record lives

- **Split from `bugs_closed/352`** on 2026-08-25 at the owner's direction. 352's arm 1 is closed:
  fixed (`ffa6e1c3d`), council-approved (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`), live since
  `v1.0.1334`, proven at the artefact, and migration `587` withdrew the 73 unexecutable legacy rows.
- **Working record** (the whole story, including four measurement errors worth reading):
  `docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/` — start at
  `HANDOFF_2026-08-25_continue_here.md`.
- **Register:** VIZ-016 (the selector contract), WII-016 (the retraction seam and the condition it
  depends on).
- **Related:** `bugs_open/296` §10.5 (same finding, other end) · `bugs_closed/198` (css-patch-agent's
  persist path; `mark_base_unsafe` is the park shape) · `bugs_closed/352` (arm 1).

---

# APPENDED 2026-08-25 (`bugfix_390_cascade_attribution` lane) — the two [UNMEASURED] items are now measured, §1's worked case is corrected, and commit 1 of 3 is live

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_390_cascade_attribution/`
(start at `PLAN_2026-08-25_cascade_attribution.md`; the missteps are in `NOTES_…` and `WRONG_CALLS.md`).

## A. ⚠ CORRECTION — §1's worked case is PARKED, not completed, so it does not demonstrate this bug

`loancash.co.uk` has `style_collection_id IS NULL`. The agent therefore takes
`check_has_css → mark_no_css` and parks in `needs_human_review` with
`result->>'parked_by' = 'css_no_theme_198'`. **All 10 of that page's live findings are parked**,
including the 17:33 UTC row §1 is built on. The four CSS facts in §1 are all correct and I
reproduced every one of them; what does not follow is "and the item completes".

**The mechanism is real — here is a case that actually shows it.** `vonc.com/index.html`, filed
`A.gauntlet-btn-primary`, 1.76:1, completed 2026-08-24 10:46:10 with a real verified selector:

| | selector | specificity | where |
|---|---|---|---|
| appended by the agent | `A.gauntlet-btn-primary { color:#ffffff }` | **(0,1,1)** | theme, linked at byte 8883 |
| actually wins | `.gauntlet-cta-section .gauntlet-btn-primary { color: var(--color-primary,#1a1a2e) }` | **(0,2,0)** | page inline `<style>` at byte 22285 |

Loses on class count and on source order. Controls: page 200, invented path 404.

## B. §6 [UNMEASURED] SETTLED — the inline `<style>` is per-PAGE, and the theme can never be reordered to win

It is `page_components.rendered_html`, and `assemblePage`
(`platform/orchestration/actions/rerender_single_page_action.go:560-720`, sections written at
`:700`) emits those blocks inside `<main>` — always after the `<link>` in `<head>`, on every page,
by construction. `getPageSections` concatenates `rendered_html` verbatim and nothing strips or
relocates a `<style>` block. For a ported page they arrive from
`cmd/webdesignport/transform.go:404-412`.

**So candidate (2) can only ever win by out-specifying or by `!important`. Source order is not
available, and no amount of reordering the theme makes it available.**

## C. §5 [UNMEASURED] SETTLED — the blast radius, and which cause dominates

Over `site_work_items` UNION `site_work_items_archive`, excluding arm-1's invented `TAG.TAG`
selectors [MEASURED 2026-08-25 — date it; this grows by addition]:

- **75 of 151** real-selector (page, selector) pairings that ever reached `complete` were filed
  **again** afterwards;
- **97** re-filings carry **byte-identical `fg` AND `bg`** — the repair changed nothing measurable.

At the artefact, the same thing seen as accumulated dead rules: noted.co.uk holds **16** appended
contrast fixes for **5** distinct selectors, vonc.com 53 for 39, loanzy.uk 62 for 31.

**Which cause dominates** — the question §6 wanted a `090` run for. 40 random completed
real-selector findings across 7 sites, served page and served theme fetched for each, governing
`color` declaration located:

| where the winning declaration lives | n / 40 |
|---|---|
| page block, **out-specifying** the filed selector | **33** |
| page block, lower specificity | 6 |
| **the theme** | **0** |
| not located (likely `over_image`) | 1 |

Near-uniform shape: filed `(0,1,1)` against page `(0,2,0)`. ⚠ `[MEASURED, static approximation]` —
a text parse, blind to media queries and cascade layers; good enough to classify a population and
not good enough to be a contract, which is why commit 2 measures in the browser.

**So the dominant defect is not a weak rule. It is a wrongly-addressed repair**: in 83% of the
sample the platform sent the fix to a stylesheet that structurally cannot govern the pixel.

## D. What has SHIPPED (commit 1 of 3), and what it does not do

**Migration `616_css_patch_agent_prompt_stops_instructing_the_losing_move.sql` — APPLIED
2026-08-25 ~16:47 BST, verified at the live row.** Commit `c441b3b8f`,
`Council-Submitted: ef5f9a0d-48a4-468e-afb4-7b6a06520f7f` (verdict not yet read).

Since migration 318 the prompt has said, verbatim: *"Repeat the offending selector exactly as it
appears above (or more specifically) so your override wins."* Both halves are false here — the
offending selector does not appear above (the model is handed `spec.selector`), and equal
specificity does not win against a later competitor. It now states the source-order truth and
instructs `!important` on the single corrected property. `!important` is licensed on a **measured**
premise: page-level colour declarations already carrying it are **0 of 812** across 8 sites — a
state measurement that expires, so re-run it (lane RUNBOOK §5) before quoting it.

Four guards mutation-proven against the live row in rolled-back transactions.

**It does not** help the ~56% of live inflow that never reaches the prompt (no linked theme →
`css_no_theme_198`; shared or short theme → 542's base-integrity gate); it does not stop a repair
being **erased** (§E); and it does not make `complete` mean the text became readable.

## E. A SECOND, INDEPENDENT way the repair goes inert — now `bugs_open/396`

`webdesign-agent`'s `persist_css_to_theme` (migration 543) writes the freshly-rendered CSS into
`css_themes.css_content` byte-for-byte, deleting every rule appended since the last design run.
agritec.uk: 5 repairs appended 2026-08-24 20:54–20:55, row rewritten **2026-08-25 12:09:57**, one
minute before a `needs_design` item completed; 0 markers left, none of the five rules in the
served stylesheet, **all five items still `complete`**.

⚠ It also makes a live concept-register sentence false (*"that is now fixed at source"*) — a dated
correction is now beneath it in `styling-render-pipeline.md`, and there is a `LANDMINES.md` entry.
**Filed separately as `bugs_open/396`, deliberately not folded in here**: 390 is a repair that never
applies, 396 is a repair that applies and is then deleted.

## F. Owner decisions taken 2026-08-25, so the next reader does not re-open them

1. **Fix the ROUTING before the RECORD.** The completion-gate half — stop `complete` meaning "a
   rule was written", let only a fresh measurement close a contrast finding — is the right
   door-closer and is **deferred, not rejected**. Note `contrast_failure` has no verifier by
   explicit decision (`discovery_checks/verifier_coverage_test.go:216`: verification needs a
   browser on the completion path), so that half is an async design, not a `GetVerifier` entry.
2. **Leave the historic rows alone; fix forward only.** No arm-2 equivalent of migration 587. The
   damage figures in §C stay quotable as evidence rather than being edited away.

## G. Still to come in this lane

- **Commit 2** — cascade attribution in the render-audit probe (which declaration wins, where it
  lives, its specificity and importance, **verified in-page by remove-and-remeasure** so the
  walker's blind spots degrade to "unverified" rather than to a confident wrong answer), plus
  routing and an `override_requirement` in the filer. The `item_key` does not change (VIZ-016).
- **Commit 3** — the agent consumes the requirement; the one class no stylesheet can reach (an
  `!important` inline `style=` attribute) parks before any LLM spend. ⚠ Its drift guard must assert
  **616's** prompt text, not 318's.
