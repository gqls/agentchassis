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

## 2026-08-25 (h) — migration 616 APPLIED and verified at the live row

Commit `c441b3b8f`, council submission `ef5f9a0d-48a4-468e-afb4-7b6a06520f7f` (trailer is
`Council-Submitted:`, verdict not yet read — do NOT upgrade that to `Council-Reviewed:` without
reading it).

Applied 2026-08-25 ~16:47 BST. Live row after apply, read back rather than assumed:

| check | value |
|---|---|
| old 318 bullet still present? | position **0** (gone) |
| source-order correction present? | position 1109 |
| actionable instruction present? | position 1579 |
| `"css_added"` output contract intact? | position 2129 |
| prompt length | 1567 -> **2280** chars |
| re-apply | raises `390/616: already applied` (loud no-op) |

**Guards mutation-proven against the LIVE row inside rolled-back transactions, before applying** —
each mutation applied, the shipped guard block run against it, the error observed, the transaction
rolled back, and the live row re-read to confirm nothing leaked:

- bullet text altered -> `drift: the cascade bullet 318 planted is not present verbatim`
- bullet duplicated -> `drift: the cascade bullet appears more than once`
- `"css_added"` renamed after a successful replace -> `verify: the JSON output contract was damaged`
- `deploy_css.next_step` pointed at a non-existent step -> `verify: 1 workflow edge(s) point at a
  step that does not exist`

**A guard I wrote and then threw away, because it could not have failed honestly.** The
occurrence count was originally `(length(p) - length(replace(p, bullet, ''))) / 219` — and the
bullet is **214** characters. That guard would have compared a fraction to 1 and fired or passed
for a reason unrelated to the text. The constant was a second copy of the string's length and
would drift from it silently. Replaced with two `position()` searches so the string itself is the
only literal. → `WRONG_CALLS.md`.

**P1 status:** unchanged, still awaiting dispatch. **P2/P3** are now live predictions — the next
`contrast_failure` dispatched on a site that passes 542's gate is the test.

## 2026-08-25 (i) — I withdrew two thirds of my own erasure evidence

Filing `bugs_open/396` (the design run that erases appended repairs), I had written that three
sites were victims: agritec.uk (5 repairs, 0 markers left) `[MEASURED]`, dartsonline.com (16, 0)
and lendzy.co.uk (12, 0) `[INFERRED]`.

Checked the inference instead of shipping it. **Both of the inferred two have ZERO `needs_design`
items of any status** — there is no design run to attribute their loss to:

```
domain          | repairs_done | design_runs | markers_left
dartsonline.com |           16 |           0 |            0
agritec.uk      |            5 |           2 |            0
lendzy.co.uk    |           12 |           0 |            0
noted.co.uk     |           37 |           0 |           73
vonc.com        |           94 |           0 |          111
loanzy.uk       |           80 |           0 |          129
```

So the marker counts were real and the CAUSE was mine: I read an absence (no markers) as evidence
for the mechanism I had just proven elsewhere. Recorded in 396 §4 as **UNEXPLAINED, not refuted**
— `needs_design` may not be webdesign-agent's only trigger, and the theme row may have been
re-linked — and whoever takes 396 is told to settle it first, because it is the difference between
"one owner-instructed palette flip cost five repairs" and "this eats every repair on the fleet".

The general shape, which is the reusable part: **once you have proven a mechanism, every unexplained
absence starts looking like it.** The discriminating question is not "is this consistent with my
mechanism" but "what else would produce exactly this, and can I rule it out" — here, one query.

## 2026-08-25 (j) — council round 1 on migration 616: REVISE, and it found a real latent defect

Verdict `ef5f9a0d`, gated by `debug_historian` (HIGH), with `guardian` and `editquality` raising
the same point independently. **Three seats on one issue is a signal, not noise.** Both objections
accepted; one was a genuine latent defect I had already made once in this same file.

**Objection 1 (HIGH) — the UPDATE and the drift guard scoped on `type=… AND is_active AND NOT
is_snapshot AND deleted_at IS NULL` with no version tie-break**, against the landmine that some
agent types carry two active rows and only the higher version loads.

I checked the landmine myself rather than taking the verdict's word, and **both halves are true**:

```
type                     | active_rows | versions
chief-strategist         |           2 | {1,2}
content-creator          |           2 | {1,2}
content-creator-contact  |           2 | {1,2}
site-component-architect |           2 | {1,2}
-- and css-patch-agent:  |           1 | {1}
```

So the objection is right about the estate and **did not bite this agent**. It is still right about
the SHAPE: my guards assumed a single row without asserting it, and that assumption fails silently
the day css-patch-agent gains a second. Fixed with a row-count assertion before any read or write,
plus `GET DIAGNOSTICS … ROW_COUNT` after the UPDATE. Mutation-proven: insert a second active row in
a transaction and the migration raises `expected exactly ONE live css-patch-agent row, found 2`.

**Objection 2 (LOW) — and it is the better catch.** The drift guard's `v_old` literal and the
UPDATE's `$old$…$old$` literal were **two separately-authored copies of the same string**. Diverge
by one character and the guard passes while `replace()` no-ops: `UPDATE 1`, nothing changed.

Fixed **by construction rather than by detection**: guard and edit are now one `DO` block, the
literals are declared once as variables, and both use the variable. Divergence is unrepresentable.

⚠ **This is the SECOND time one restated literal has bitten this single migration** — the first was
the `219` I typed for a 214-character string, which I caught myself. Same shape, twice, in one
file: *a literal written down twice is a literal that will disagree with itself.* That is the
transferable rule, and it is worth more than either fix.

**Two round-1 premises that are FALSE, recorded so nobody inherits them:**

1. `editquality` (medium) and `guidelines` both said a submission naming two files (migration +
   `_ROLLBACK`) is refused server-side. **It is not.** This submission is exactly those two files;
   it passed `DRY_RUN` admission and round 1 itself was accepted and fully reviewed. The landmine
   they cite does not describe this gate's current behaviour.
2. Several seats' read-only checks returned `has_old_text: false`, which reads as "the precondition
   is already gone, so the replace would no-op". **That is my sequencing, not drift** — I applied at
   ~16:47 BST, before the verdict, so the council observed the post-apply world.
   `prior_art_librarian` asked specifically for that check to run *before* the replace executes.

**The process lesson, which is the part I would otherwise not have noticed: applying before the
verdict lands makes the council's own precondition checks report the precondition as ABSENT.** The
answers then look like evidence of drift and are nothing of the kind. If you apply early, say so in
the submission so the seats can read their own checks correctly.

**What no seat objected to on the merits:** the `!important` decision. `constitution`,
`architecture`, `render_guardian` and `reuse_agent` each examined it and approved, three noting
explicitly that the append-only structural position leaves no alternative and that the measured
0/812 premise is the right kind of grounding. `bug_historian` (approve) made one medium note — that
migration 543 will erase these wins at the next design run — and recommended the separate filing
cross-reference this migration number so the two are not resolved independently. **Done**, in
`bugs_open/396`.

Round 2 resubmitted on the same correlation `ef5f9a0d`, so the trail accumulates.

## 2026-08-26 (k) — the roll landed, both approvals landed, and commit 3 shipped

**Deploy proven at the artefact, per service, with controls** — never at the tag:

- `browser-runner-adapter` and `render-audit-adapter` both log `build provenance 2fb40a96…`
  (started 23:11–23:12 on 08-25).
- `agent-chassis`: startup line had scrolled (the known shelf-life), so binary probe:
  `grep -aq <stamp> /proc/1/exe` → present; absent-sha control → absent.
- `git merge-base --is-ancestor ea64845e0 2fb40a96` → YES; control `HEAD` → NO.

**Council: 616 round 2 APPROVED** (`ef5f9a0d`, 16:23) — the revise→approved trail on one
correlation, as designed. **Commit 2 APPROVED** (`058b59b6`, 16:22). Both my commits carry
`Council-Submitted:` trailers, which 098 resolves at report time; no amends.

**A touched row investigated before building on it.** `css-patch-agent.updated_at` moved to
23:10:35 — *after* my 616 apply. Checked rather than assumed: no snapshot after mine
(`agent_snapshots` view — note it has `snapshot_taken`, not `created_at`), version/prompt/steps
unchanged, and 23:10 is one minute before the fleet roll — a release stamping a non-config column.
Benign. The check cost three queries; assuming would have cost nothing today and everything the
day it wasn't benign.

**Migration 635 (commit 3) written, mutation-proven, submitted (`fe5cbe0c`), applied.** The two
design points worth carrying:

1. **The template fence is on `override_requirement`, not `winning_rule`** — the filer writes
   `winning_rule` for a verified *unreachable* winner too (where req is nil), and text/template
   errors at execute time on a missing sibling dereference, which would fail the whole LLM step.
2. **Template safety was proven on the exact post-apply artefact, not a hand-copy**: apply block
   run in a rolled-back txn, the resulting prompt `\copy`'d out of the DB, unescaped, parsed and
   executed against five data shapes. A hand-copied version can pass while the real one differs.

Mutation proofs: second active row → raises; 616's text removed underneath → raises;
`check_base_integrity` rewired → raises. Round trip: apply + ROLLBACK in one txn → restores 616's
shape. Post-apply read-back: gate wired, park stamps before terminal, block at 133, prompt 3,241
chars, 0 dangling edges; re-apply refuses.

**Numbers move fast on this tree:** 617–634 were taken by other lanes in one overnight window.
Re-check the next free number at write time, not at plan time.

**Predictions P1–P5: all still open.** Zero audits since the roll (rotation ticked 08-26 07:47,
no site past the 3-day window). P4's window opens with the first post-roll audit, expected by
~08-29.

## 2026-08-26 (l) — provider back; 635 resubmitted; P1 graded, and it was not a prediction

**Provider outage over** (owner topped up credit ~10:00 BST; last `LLM_API_ERROR` 09:57:46 BST).
635 resubmitted on the SAME correlation `fe5cbe0c`, run `a613549a`, ~10:02 BST.

**P1 graded: the OBSERVATION is confirmed, the PREDICTION is void.** cv1.co.uk's three `detected`
rows dispatched 14:40–14:42 BST on 08-25 and parked at `css_base_integrity_guard_198`,
attempt_count 1, no LLM run, no rule appended — exactly the claimed behaviour of 542's gate.

**But the timestamps disqualify it as a prediction**: the rows parked at 14:40–14:42 BST and my
"pre-registered" P1 was committed at 16:38 BST (`3956adc06`) — **two hours after the event**. I was
blind to the outcome (my last read of those rows, ~13:30 BST, showed `detected`, and I did not
re-check before writing), but *"I had not looked"* is not *"it had not happened"*. A prediction
registered after its event is a postdiction however blind the author, because nothing could have
made it wrong. → `WRONG_CALLS.md`.

**P2–P5 remain genuine predictions** — they concern post-roll behaviour and the roll (23:11) came
hours after the commit that registered them.

**The verification calendar moved up.** The earlier "no site past the window" reading used the
WRONG rotation stamp — `site_discovery_rotation.last_selected_at` unfiltered, which mixes agent
types. The live `pre_query` keys on `agent_type='render-audit-agent'`, and under that key the next
sites come due TODAY: remortgagecalculator.uk 13:16 UTC, garden-tools.uk 17:18, cookly.uk 18:19.
First post-roll audit expected at the first hourly tick after ~14:16 BST — not "~08-29".
(The rotation stamps the site in the same statement it selects it, so a due reading is consumed by
the read; always read `due_at`, never re-derive from a stale query.)
