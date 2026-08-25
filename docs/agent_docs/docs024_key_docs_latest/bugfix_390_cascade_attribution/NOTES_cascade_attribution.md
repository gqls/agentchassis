# NOTES — bugs_open/390 (append-only, newest at the bottom)

## 2026-08-25 (a) — reproducing the mechanism, and finding the bug file's own case is wrong

Started from `bugs_open/390`'s worked case, `loancash.co.uk/guides/index.html`. Reproduced its
four facts at the artefact: stylesheet link at byte 8562, offending declaration at 13408 (inside
the inline `<style>` that opens at 12080), served theme mentions `ported-page-content` **0** times,
`--color-primary: #e8f5ee` present in the theme. Control: invented path 404, real path 200.

**But the served theme also contains `css-patch-agent` 0 times**, which does not fit a story about
an appended rule losing. Checked the DB: `loancash.co.uk` has `style_collection_id IS NULL`. So
`check_has_css → mark_no_css` fires and the item parks in `needs_human_review`
(`parked_by=css_no_theme_198`) — all **10** of that page's live findings are parked, not completed.

**So 390's worked case does not demonstrate the bug 390 is filed for.** The mechanism is real; the
case is not. Found a real one instead — `vonc.com/index.html`, `A.gauntlet-btn-primary`, filed
1.76:1, completed 2026-08-24 10:46, agent appended `A.gauntlet-btn-primary { color:#ffffff }`:

- appended (theme, `<head>`): `A.gauntlet-btn-primary` = **(0,1,1)**
- winning (page inline `<style>` at byte 22285): `.gauntlet-cta-section .gauntlet-btn-primary
  { color: var(--color-primary,#1a1a2e) }` = **(0,2,0)**

Loses on class count and on source order. Item is `complete`; text still fails.

## 2026-08-25 (b) — the two [UNMEASURED] items in the bug file

**§5 blast radius.** Over live ∪ archive: **75 of 151** real-selector (page, selector) pairings
that ever reached `complete` were filed again afterwards; **97** re-filings carry byte-identical
`fg`/`bg`. At the artefact the churn is visible as accumulated dead rules — noted.co.uk 16
appended contrast fixes for **5** distinct selectors, vonc.com 53 for 39, loanzy.uk 62 for 31.

**§6 "is the inline `<style>` per-page or per-component?"** Per-page: it is
`page_components.rendered_html`, and `assemblePage` writes it inside `<main>`, always after the
`<link>` in `<head>`. **So the theme can never win on source order** — candidate (2) can only win
by out-specifying or by `!important`. That closes the question the bug file said determined
whether candidate (2) was worth anything.

**Which cause dominates** (the question §6 wanted a `090` run for). 40 random completed
real-selector findings across 7 sites, page and theme fetched for each: **33** have the winning
declaration in a page `<style>` block *out-specifying* the filed selector, **6** in a page block
the theme could out-specify, **0** in the theme, 1 not located. Near-uniform shape: `(0,1,1)` filed
vs `(0,2,0)` winning. `[MEASURED, static approximation]` — text parsing, so blind to media queries
and cascade layers; good enough to classify a population, not good enough to be a contract.

## 2026-08-25 (c) — MISSTEP: I called cv1.co.uk a third failure mode, and it is not

cv1.co.uk has three `contrast_failure` rows sitting `detected`. Its pages link
`/assets/css/styles.css`; **that URL 404s**, and the offender
(`.checklist-section h3 { color: var(--accent-color) }`) is in a page `<style>` at byte 15544,
after the link at 8179. I said in chat that this was a third, uncaught way a repair goes inert.

**Wrong, and I said it before checking the gate that governs it.** cv1's theme is the shared seed
`professional-dark` at **1,649 bytes**. Migration 542's base-integrity gate refuses unless
`css_len >= 4096 AND site_count <= 1`, so these three will park at
`css_base_integrity_guard_198` — the existing guard already covers them. The 404 stylesheet is
real and worth knowing; it is not a new hole.

**The cheap check that would have caught it, and now lives in RUNBOOK §3:** read 542's gate
condition against the actual `css_themes` row before predicting any outcome for a site. I reasoned
from the served artefact and skipped the gate. → `WRONG_CALLS.md`.

## 2026-08-25 (d) — a second, DESIGNED way the repair goes inert: erasure

agritec.uk completed 5 contrast repairs on 2026-08-24 20:54. Its theme row now holds **0**
`css-patch-agent` markers, and the served stylesheet holds no `bl-read-link` rule. What happened at
**2026-08-25 12:09:57** was a theme rewrite, one minute before a `needs_design` item
(`reason: palette_changed`, handler `webdesign-agent`) completed at 12:10:29.

This is migration `543_webdesign_agent_persists_rendered_css_to_theme_row.sql` working as designed:
it adds `persist_css_to_theme`, which writes the freshly-rendered CSS into the theme row
byte-for-byte. Its own header says *"whichever agent ran last owns the file entirely"* — it simply
never draws the consequence for contrast repairs. **Every theme-appended repair expires at that
site's next design run** (543 puts that at roughly weekly).

Fleet: 3 of the 11 sites holding completed contrast repairs now have zero surviving markers —
dartsonline.com (16 completed repairs), lendzy.co.uk (12), agritec.uk (5). `[MEASURED]` on agritec,
where I have the timestamps a minute apart; `[INFERRED]` on the other two, where I have only the
absence. **Filed as its own bug, not folded into 390.**

## 2026-08-25 (e) — the `!important` census, and why it is the load-bearing number

Commit 1 hands the agent `!important`, which is only sound if nothing it must beat already carries
it. Page-level colour declarations carrying `!important`, across 8 sites' homepages: **0 of 812**
(noted.co.uk 0/157, idea.uk 0/210, cookly.uk 0/120, vonc 0/94, remortgagecalculator 0/78, oufe
0/72, loanzy 0/42, loancash 0/39). It could easily have come out otherwise, which is what makes it
evidence. ⚠ It is a **state** measurement and expires: re-run before quoting it.

## 2026-08-25 (f) — the tree moved under me mid-session

At session start the completion-gate files were dirty with the `bugs_open/395` lane's work, and I
wrote a collision warning into the plan on that basis. That lane has since committed;
`complete_work_item_verification.go`, `complete_work_item_acceptance_predicate.go`,
`load_work_item_actions.go`, `write_render_audit_findings_action.go` and `render_audit_action.go`
are all clean. A Fable design agent caught this and corrected me. Only `platform/livespec/livespec.go`
is still dirty, and that is a different lane's rename.

**The general form:** a `git status` taken at session start is a claim about a moment, and this
tree moves in minutes. Re-run it immediately before each commit, not once at the start.

## 2026-08-25 (g) — PRE-REGISTERED PREDICTIONS, written before any change ships

Recorded now so they can be wrong. A prediction written after the run is not a prediction.

| # | subject | prediction | how it is settled |
|---|---|---|---|
| P1 | cv1.co.uk's 3 `detected` rows | park at `css_base_integrity_guard_198`; the LLM never runs; no rule appended | `status='needs_human_review'` and `result->>'parked_by'` |
| P2 | the next `contrast_failure` dispatched on a site passing 542's gate, AFTER migration 616 | the appended rule carries `!important` on exactly one property | curl the served stylesheet |
| P3 | that same pairing at the next render audit | **retracted** (`resolved_by='render_audit'`), not re-filed | the item's `result` |
| P4 | cascade attribution, once built, on vonc/loanzy/noted | `repair_surface='theme'` with `strictly_greater=true` on the large majority; `unreachable` rare | the filed `spec.repair_surface` distribution |
| P5 | agritec.uk | its 5 completed pairings re-file at its next audit, because the palette re-render erased the repairs | new rows with the same selector+page |

**Disconfirming results, stated in advance:** P2 fails if the appended rule has no `!important` or
carries it on several properties. P3 fails on a re-filing with byte-identical `fg`/`bg`. P4 fails
if `unattributed` dominates — which would mean the remove-and-remeasure verification is rejecting
its own attributions and the probe is blind, not that the pages are unusual.
