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
