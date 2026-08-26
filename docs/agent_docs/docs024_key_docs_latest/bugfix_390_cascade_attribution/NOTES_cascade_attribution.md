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

## 2026-08-26 (m) — state re-verified at 12:18 BST; the handoff's accounting query named a key that never exists; watch armed for the first post-roll audit

**Re-verified live, not inherited:** 635's verdict is `approved` at the DB
(`diagnosis_artifacts`, corr `fe5cbe0c`, 09:14:09 UTC — one row, no earlier verdict, consistent
with r1 dying at `complete_invalid`). Zero cascade-attributed `contrast_failure` rows exist
(`spec ? 'cascade_scheme'` → 0), and zero post-roll audits have run.

**The verification calendar, read from `last_selected_at + 3 days` (read-only — the pre_query
stamps in the same statement it selects, so never run IT to ask what's due):**
remortgagecalculator.uk due 13:16 UTC today, garden-tools.uk 17:18, cookly.uk 18:19.
Task ticks hourly (`interval_seconds=3600`, last trigger 10:49 UTC), so first selection at the
**13:49 UTC tick (~14:49 BST)**. The P2/P3-decisive sites all come due TOMORROW: vonc.com 10:27,
noted.co.uk 12:28, loanzy.uk 15:29 UTC (loancash 17:30; cv1 08-28 12:39).

**All three of today's due sites PASS 542's gate** (RUNBOOK §3 query: css_len 17,978–20,156, all
site_count=1, all linked) — so any new `detected` rows today dispatch to css-patch-agent rather
than parking, and **P2 may be gradable today**, not only P4.

**CORRECTION (appended to the handoff too): the handoff §1(b) accounting query reads
`collected_data->'audit_findings'`, a key that exists on NO orchestration row.** The cascade
counters are written into the action RESULT map
(`write_render_audit_findings_action.go:642-651`, unconditional, zeros included) and land under
the STEP key — `collected_data->'write_findings'` (grounded on the cv1 08-25 run, whose
`inserted: 3 / deduped: 5` accounting sits exactly there; all 3 existing `render-audit-agent`
orchestrations have no `audit_findings` key). The wrong query returns 0 rows for ever and reads
as "no audit yet" — the same shape as the `audit_findings` name being written from memory rather
than from the step map. Baseline for the watch: today, zero rows carry
`write_findings ? 'cascade_attributed'`, so the first hit IS the first post-roll audit.

**Prior-audit baseline for remortgagecalculator.uk:** 7 `contrast_failure` rows all `complete`
(5 filed 08-23 13:17, complete by 13:49; 2 filed 08-20). If today's audit re-files any of those
pairings byte-identical, that is the last pre-attribution evidence of the damage class; the NEW
rows should carry `cascade_scheme` + `repair_surface` (P4).

## 2026-08-26 (n) — FIRST POST-ROLL AUDIT: remortgagecalculator.uk at 13:50:31 UTC, exactly the predicted tick. P4 first read: 1 attributed, 4 unattributed — and the 4 are ONE mechanism, measured at the served CSS

**The audit fired as predicted** (13:49 UTC tick → orchestration `a5634f3a` created 13:50:31,
COMPLETED 13:51:06, 35 s). `write_findings` accounting, read at the correct key:
`inserted 5 · deduped 0 · retracted 0 · cascade_scheme_present TRUE · cascade_attributed 1 ·
cascade_unattributed 4 · cascade_unreachable 0 · cascade_unverified_by_probe 4 · cascade_capped 0
· cascade_dirty_pages 0 · retraction_scope_pages 4`. Both images are speaking the new scheme
(`spec.cascade_scheme = 'cascade/v1'`, `selector_scheme = 'verified/v1'`).

**The attributed row is the predicted shape exactly.** `.brief-explanation__heading EM` on
/index.html: winner `.brief-explanation__heading em` in a page `<style>` block
(`surface: style_block`, `decl: var(--color-accent, #f59e0b)`, `verified: true`, `candidates: 1`),
`repair_surface = theme`, `override_requirement = {min_specificity 0,1,1; strictly_greater true;
needs_important false}`, `override_example = body .brief-explanation__heading EM`.

**The four unattributed rows are one pairing on four pages** — `.footer-bottom P`,
rgb(107,114,128) on rgb(241,237,228), 4.14:1 vs 4.5 — so by PAIRING the read is 1 attributed :
1 unattributed, not 1 : 4. The chrome repeats per page; counting rows over-weights it. Any
distribution claim about P4 must be made per distinct (selector, fg, bg), not per row.

**WHY the footer is `verified:false` — measured at the served page + theme, with a 404 control
(`index-390-control.html` → 404; page 200 107 KB; css 200 18,162 B):**
- The `<p>` has NO colour rule of its own. Its colour is INHERITED from
  `.footer-bottom { … color: var(--color-text-muted); }` — a rule on the ANCESTOR, present in the
  theme (line 225) AND twice in the page (10 `<style>` blocks; the footer chrome's rules appear ×2).
- The only `color` declaration whose selector MATCHES the `<p>` is the theme's
  `p, li, blockquote { color: var(--section-text, inherit); }` (theme line 90). With no
  `--section-text` in scope that resolves to `inherit` — the SAME value the element would inherit
  anyway. Removing it cannot move the computed colour ⇒ `verified:false` ⇒ `unattributed`.
- `winningDecl` (cascade_attribution.go:146ff) only collects candidates where `el.matches(part)`;
  it never looks at ancestors, by design ("attribute the declaration that decides the element's
  colour"). For an INHERITED colour the deciding declaration is on the ancestor, so this class is
  structurally unattributable by the current probe. It is NOT probe blindness in the sense the
  disconfirming clause meant (`opaque_sheets 0`, `dirty 0`, `capped 0`) — it is the designed
  under-claim (same-value runner-up), triggered by inheritance.

**Consequence for the design, stated rather than buried:** this class is exactly one where the
theme CAN govern — an appended `.footer-bottom p { color: X }` (0,1,1) is a DIRECT declaration
on the element and beats any inherited value regardless of the ancestor rule's specificity, and it
out-specifies the theme's own `p` (0,0,1). So the conservative route sends the agent in WITHOUT a
requirement on precisely a case it could have won with one. Not a defect by the design's own rule
(never a weak yes), but a **systematic under-attribution class: inherited colour**, and chrome
(footer, header) is where it lives, so it recurs on every page of every site. Candidate
follow-up, NOT done here: when no removal moves the value AND `getComputedStyle(el).color ===
getComputedStyle(el.parentElement).color`, walk to the parent and attribute the property there —
the same removal proof applies one level up. Needs its own council round; recorded so the P4
grading on vonc/noted/loanzy tomorrow reads unattributed footers as THIS class, not as blindness.

**Old rows vs new (the damage class, closed at the artefact):** the 08-23 "completes" appended to
the theme against the INVENTED selectors `P.P` (×3, three different darkenings!) and `EM.EM` —
which match nothing — so today re-files those pairings byte-identical (fg AND bg): the third
filing of the footer pairing since 08-20. The ONE 08-23 repair against a real selector,
`SPAN.brief-explanation__eyebrow { color: #8a4e00 }`, is NOT re-filed today — the theme governed
that one and the old path worked where the surface could reach the pixel. `deduped: 0` because
the selectors changed shape (352's fix), not because the failures are new.

**Dispatch path, learned:** `build-pipeline-trigger` (30 s) spawns `build-dispatch-loop`, which
loads ≤5 items per run for the site (this run: 4 `unbuilt_internal_link` prio 45 + 1
`contrast_failure` prio 60), processes them in load order, and the contrast row sits LAST. The
4 page builds ahead of it each FAILED in ~4 min on `bugs_open/260`'s `mechanism-flow` type
mismatch (another lane's). `ef31c778` (/next-steps, unattributed) reached `claimed` at ~14:30 UTC.
**css-patch-agent orchestration rows are PURGED after completion** — cv1's park cites
`completed_by_orchestration_id 77704d81` (08-25 13:40) and that row is gone while 10,325 rows
from 12:27 that day onward survive. So the artefacts for P2 are the item `result` + the served
theme + `llm_call_log`, never the orchestration row. Watch armed on the five rows.

## 2026-08-26 (o) — P2 CONFIRMED AT THE ARTEFACT: first post-roll repair, 14:27 UTC, `!important` on exactly one property; 635's fence proven on a real row

**The repair:** `ef31c778` (`.footer-bottom P`, /next-steps.html, `repair_surface=unattributed`)
was claimed by `build-dispatch-loop dd3437af` (iter 4, last of its 5 items) and css-patch-agent
(child orchestration `7910a705`, `css-patch-agent-workflow-1427`) appended:

```css
/* css-patch-agent 2026-08-26: contrast */
/* css-patch-agent: contrast – .footer-bottom P foreground darkened to meet 4.5:1 on rgb(241,237,228) */
.footer-bottom p { color: #595f6b !important; }
```

Git adapter: commit `f2d68710` to `gqls/sites` master, `assets/css/styles.css`, "CSS fix: contrast
(theme v10)", sha256 `1b859a29…`.

**P2 — "the appended rule carries `!important` on exactly one property" — CONFIRMED**, and the
pre-stated disconfirming results (no `!important`, or on several properties) did not occur. Graded
at three artefacts, none of them the status column:
1. **Served stylesheet** (`curl`, HTTP 200, 18,361 B; control `styles-390-control.css` → 404): the
   rule is the LAST rule in the file (line 297), and the served file's sha256 **equals** the git
   adapter's `files_sha256` **equals** `sha256(css_themes.css_content)` (v10, `updated_at` 14:27:37).
   One artefact, three independent hashes agreeing — the bucket is serving the committed theme.
2. **Arithmetic**: #595f6b on rgb(241,237,228) = **5.49:1** (needs 4.5; was 4.14). Own computation
   from the WCAG formula, not the agent's claim.
3. **Mechanism**: a DIRECT declaration on the `<p>` with `!important` beats the inherited
   `.footer-bottom` value on every page — so this one rule fixes the pairing on /index, /about,
   /mortgage-lenders AND /next-steps, not just the filed page. [INFERRED from the cascade; the
   browser measurement that settles it is the 08-29 audit → P3.]

**635's template fence proven on a real unattributed row, at the rendered prompt** (`llm_call_log`
`work_item_id=ef31c778`, `plan_css_fix`, 14:27:36): `position('The declaration you must BEAT')`
= **0** (block absent, as designed for `override_requirement` nil), the general `!important`
guidance present at 19,707 of 20,895 chars, `output_tokens` 166 of `max_tokens` 8,000 (not cut).
The disconfirming result — the block rendered with empty fields, or an execute-time template
error — did not occur.

**Two things observed that are NOT this bug's, recorded so nobody re-derives them:**
- The item's `result` is the handler's SPAWN/response envelope (`bugfix-287` shape), with
  `attempt_count 0` at `complete` and no `completed_by_step` — the css_added text is inside
  `result.response.css_fix.result`. Grade repairs from there + the served file, never from the
  status.
- **N pages × one chrome pairing ⇒ N rows ⇒ N appended rules.** The other three footer rows are
  still `triaged`; each will dispatch css-patch-agent again for a pairing this rule has already
  fixed, and the 08-23 theme shows the outcome: `P.P` darkened THREE separate times (#595f6b,
  #555b67, #4a4f5a). `item_key` embeds `page_name` (VIZ-016) so dedup cannot see it. A
  pre-dispatch "does the current theme already carry a rule for this selector+property" check —
  or, better, the deferred completion gate (§4 of the handoff: complete only on measured
  improvement) — is where this belongs. Candidate for the 395 lane's template; not done here.

**P3's clock now runs:** remortgagecalculator.uk's next audit ≈ **2026-08-29 13:50 UTC**. The four
footer rows + `.brief-explanation__heading EM` must be RETRACTED (`resolved_by='render_audit'`),
not re-filed with byte-identical fg/bg. Check `css_themes.updated_at` first — a design run between
now and then (bugs_open/396) would erase the rule and fail P3 for the wrong reason.

**Still pending today:** `9b2b2ce9` — the ATTRIBUTED row (`theme`, must exceed 0,1,1, `needs_important
false`) — is the first real exercise of the requirement block. Expected artefact: a selector with
specificity > (0,1,1) (the verified example is `body .brief-explanation__heading EM`) and **no**
`!important` (the block says "Do NOT use !important … supersedes the general guidance"). Watch
re-armed with `ef31c778` excluded.

## 2026-08-26 (p) — all five rows repaired by 14:34 UTC (theme v10→v14); the ATTRIBUTED row met its requirement, copied the example verbatim, and IGNORED "do not use !important"

**Final state, all at the artefact:** five `complete` rows; `css_themes` v14, `updated_at`
14:34:27; served file HTTP 200 (19,179 B; control 404) with sha256 `f20b76a9…` = git adapter's
`files_sha256` for v14 = `sha256(css_content)`. The 2026-08-26 tail of the served theme:

```
297  .footer-bottom p { color: #595f6b !important; }                    ← ef31c778 (v10)
301  .footer-bottom p { color: #595959 !important; }                    ← 91690588 (v11)
305  .footer-bottom p { color: #595959 !important; }                    ← 762675d4 (v12)
309  body .brief-explanation__heading EM { color: #b5620a !important; } ← 9b2b2ce9 (v13, ATTRIBUTED)
313  .footer-bottom p { color: #595959 !important; }                    ← 720324cc (v14)
```
Arithmetic (own): footer #595959 on rgb(241,237,228) = **6.0:1** (needs 4.5); heading #b5620a on
rgb(247,248,246) = **4.18:1** (needs 3.0; the agent claimed "approximately 3.1" — under-claimed).

**The attributed row — the first real exercise of 635's requirement block — graded in three parts:**
1. **The block rendered** (`llm_call_log` 14:34:16: BEAT block at 137, "Do NOT use !important" at
   577, the verified example at 769; 172 of 8,000 output tokens).
2. **The requirement was MET**: `body .brief-explanation__heading EM` is (0,2,1), strictly greater
   than the winner's (0,1,1). The agent's own summary says so in those terms — it read the block.
   The selector is the platform's `override_example` **copied verbatim** (the quoted-exemplar
   effect, `a-quoted-exemplar-in-a-prompt-is-copied-verbatim`); fine here because the platform had
   CHECKED that example with `satisfiesRequirement`, which is exactly why the design ships a
   checked example and not a hope.
3. **The instruction "Do NOT use !important — this measured section supersedes the general
   guidance below" was IGNORED**: the rule carries `!important`. Functionally harmless (the rule
   wins either way) and `needs_important:false` is thereby an INERT signal today — the agent adds
   it regardless. Structural reading: the general `!important` guidance (616) sits at ~19,700 of
   ~20,900 chars, the measured block at ~140; the model followed the later, closer instruction over
   the earlier one that claimed to supersede it. **"Supersedes the guidance below" is a comment,
   not a control** — the same lesson as `a-doc-comment-is-not-an-enforcement-mechanism`, one level
   up: a prompt instruction adjudicated by the model is not a fence. The fix, if this is systematic,
   is to make the contradiction unrepresentable: wrap 616's general `!important` bullet in
   `{{if not .input_data.spec.override_requirement}}…{{end}}` so only ONE of the two instructions
   ever renders. **NOT done today: n=1.** Tomorrow's attributed rows on vonc/noted/loanzy are the
   sample; if ≥2 of them carry `!important` under `needs_important:false`, write migration 6xx
   (agent_definitions → council scope, drift-anchored on 635's shape, row-count assertion).

**The footer: one pairing, four rows, four rules, three of them redundant** — and the first
(#595f6b) is overridden by the last (#595959) on source order. Each agent run saw the theme with
the previous rule already appended and appended another anyway: the agent does not check whether
the theme already governs the selector, and the spec gives it no reason to (each row is "its"
page). Recorded as the N-pages × chrome class in §(o); the completion gate (handoff §4, deferred by
the owner) is the structural home. Cost today: 3 LLM calls + 3 git commits + 3 deploys for nothing.

**Rows are in the past tense now; what remains is the FUTURE tense.** P3 is settled by the 08-29
~13:50 UTC audit of this site; P4 by tomorrow's vonc (10:27) / noted (12:28) / loanzy (15:29 UTC)
audits, read per distinct pairing with inherited-colour footers expected `unattributed`.

## 2026-08-26 (q) — PRE-AUDIT BASELINE for tonight's two sites, recorded BEFORE their audits so the after can be graded against it

garden-tools.uk (due 17:18 UTC, ~17:50 tick) and cookly.uk (due 18:19, ~18:50 tick); both pass
542's gate (§(m)). Served themes fetched 15:0x UTC with 404 controls (files in scratchpad).

**garden-tools.uk — the CLEAN control:** zero contrast rows ever (live + archive), zero patch
markers in the served theme (20,304 B). Its audit grades the "attributed nothing"/fresh-finding
path.

**cookly.uk — the damage class in miniature, and the sharper test:** 25 historical rows, ALL
`complete`; the 08-17 batch re-filed byte-identical on 08-23 (9 pairings, fg AND bg). 17 patch
markers in the served theme (19,067 B). The patched selectors split exactly along 390's two
mechanisms:
- invented: `p.p` (×5 rules, three of them a compound `.footer-bottom p.p, .site-footer p.p, p.p`)
  — match nothing (352's class);
- **REAL: `p.note`, `p.contact-intro`, `span.label`, `span.accent-text` — and their pairings were
  STILL re-filed byte-identical on 08-23.** That is the pure 390 mechanism at a second site: a
  correct selector in the theme losing the cascade (the measured 33/40 shape), no invented-selector
  confusion. Pre-616, so none carry `!important`.

**Tonight's expectations, falsifiably:** cookly re-files its pairings a THIRD time (the repairs are
inert; anything else means something else changed) — now with real selectors and attribution;
paragraphs styled via `var(--section-text, inherit)`-type rules may land `unattributed` (the
inherited-colour class of §(n)) — read `winning_rule.surface` before calling it blindness; any
attributed `theme` rows then repaired tonight become cookly's own P3 test at its next audit
(~08-29 18:50 UTC). If ≥1 attributed repair renders WITHOUT `!important` where
`needs_important:false`, the §(p) n=1 does not grow; if ≥2 attributed repairs keep `!important`,
write the fence migration.

## 2026-08-26 (r) — tonight's two audits GRADE P4 CONFIRMED-SO-FAR; the !important sample hit 6/6; migration 655 written, proven, submitted (`ffd6952b`), APPLIED ~19:20 UTC

**garden-tools.uk (17:50 UTC, orch `09ecc0a8`):** 26 pairings measured — **24 attributed, 2
unattributed, 0 unreachable** (`unverified_by_probe 6` is the adapter's per-element count, a
DIFFERENT denominator than the filer's per-finding pair — carried through unchanged by design; do
not read 24+2 vs 6 as an inconsistency). 14 rows filed: **13 theme, 1 unattributed**, every
winner the measured shape (`.section .class` (0,2,0) / bare `.class` (0,1,0) / `.item p` (0,1,1)
in page style_blocks; one `linked` winner — `blockquote` (0,0,1) in the theme itself, still
theme-repairable). All `needs_important:false`.

**cookly.uk (18:55 UTC, orch `fc7ac552`):** **13 attributed, 0 unattributed**, 0 unreachable; 5
rows filed, ALL `theme` — the §(q) prediction (third byte-identical filing of the historic
pairings) holds, and the site whose REAL-selector repairs kept failing now carries the measured
requirement on every row. Note: cookly's paragraphs attributed fine — the inherited-colour class
is not universal; it depends on the site's CSS shape (remortgage's footer inherits; cookly's
paragraphs are directly ruled).

**P4 running tally across three sites, by the probe's own accounting: 38 attributed / 6
unattributed / 0 unreachable** — `repair_surface='theme'` dominates exactly as predicted, and the
pre-stated disconfirming condition (unattributed dominating) is dead. Formal P4 grading still
waits for vonc/noted/loanzy (tomorrow 10:27/12:28/15:29 UTC) since the prediction named them.

**The !important sample closed at 6/6** (remortgage `9b2b2ce9` + garden-tools' first five
attributed completions, counted from `css_added`) — every attributed repair met its specificity
requirement AND kept `!important` against the block's "Do NOT". The pre-registered ≥2 threshold
was crossed, so **migration 655** was written and shipped tonight:

- Literals SLICED from the live `prompt_template` bytes by script (the twice-bitten lesson —
  never retyped). ⚠ The config key is **`prompt_template`**, not `prompt` — my first extraction
  query silently returned empty on the wrong key; same trap-shape as §(m)'s `write_findings`.
- Fence: 616's three bullets wrapped in `{{if not .input_data.spec.override_requirement}}`;
  635's else-sentence shortened (its "guidance below" would dangle). Exactly one instruction
  family can render.
- Proofs, all against the live row in rolled-back txns: apply 3241→3282 chars (arithmetic checks:
  +55 fence, −14 sentence); apply-then-rollback restores md5 `a7be07f2…` exactly; second active
  row → RAISE 'found 2' (clone needed `version+1` — unique `(type,version)`); double-apply →
  refuses; one word edited in the passage → drift RAISE. Template proof ran on the EXTRACTED
  post-apply artefact under the production parse (`datahelpers.RenderPromptTemplate`,
  `data_helpers.go:1129` — `template.New("agent_prompt")` + funcmap, default options): 5 shapes,
  instruction families exclusive in every render.
- Submitted `ffd6952b`, DRY_RUN admission first; APPLIED ~19:20 UTC with the sequencing stated in
  the submission; committed `1e513f5c9` with `Council-Submitted:`.

**⚠ Tonight's sample is SPLIT by the apply:** rows whose `plan_css_fix` ran before ~19:20 UTC
rendered the old prompt (expect `!important`); after, the fenced one (expect NO `!important` on
`needs_important:false` rows). Grade by `llm_call_log.created_at`, not by site. The post-655
repairs are ALSO the live test that a strictly-greater selector wins WITHOUT `!important` — if one
fails to take at the next audit, the design wants that visible (655's risks §2).

> **CORRECTED 2026-08-26 19:05 UTC:** §(r) said 655 was "APPLIED ~19:20 UTC" — the estimate was
> written without reading the clock. The row says `agent_definitions.updated_at = 18:57:48 UTC`;
> the sample splits THERE. Caught three minutes later by the first post-apply render.

## 2026-08-26 (s) — the fence PASSED its first live exercise, 3.5 minutes after apply

`llm_call_log` for garden-tools, `plan_css_fix`, ordered by time: five renders 18:44–18:48 UTC
(pre-655) all show `general_rendered=t, fenced_sentence=f` and every reply kept `!important`;
the first post-655 render (`ecd68793`, 19:01:20 UTC) shows `general_rendered=f,
fenced_sentence=t` and shipped
`body .tool-guide-intro-section .tgi-step-desc { color: #4a5240; }` — (0,2,1) strictly above the
required (0,2,0), **no `!important`**. Same model, same row shape, same evening: the behaviour
tracked the prompt exactly, which is as clean a controlled comparison as this estate will ever
hand us. Remaining drains (garden-tools 7–8 rows, cookly 5) are all post-655 — expect no
`!important` throughout; any reappearance is a real finding, not noise.

## 2026-08-26 (t) — 655 APPROVED round 1 (2 medium advisories + assorted lows); each advisory graded against tonight's LIVE evidence; garden-tools drained 14/14 and verified at the served theme

**Verdict** (`ffd6952b`, council_report): APPROVED, 5 abstained, no highs. Advisory dispositions,
each answered by evidence rather than argument:

- *editquality M1 (css-patch-agent might be among the duplicate-active-row types; the guard would
  refuse):* the apply SUCCEEDED at 18:57:48 — the guard found exactly one row. Standing risk noted
  for the day a second row appears; the guard failing loud is the designed outcome.
- *editquality M2 (dotted path unverified) + bug_historian M2 (Go template truthiness could make
  the fence always-false):* settled in production, both branches, within 12 minutes of apply —
  `ecd68793` (row WITH `override_requirement`) rendered the fenced sentence and NOT the general
  bullet; `80d5dd85` (row WITHOUT) rendered the general bullet and kept `!important`. jsonb → map;
  absent key → nil → `not` true. The synthetic-shapes concern (guardian low) is closed by the same
  live renders.
- *prior_art_librarian (verify `needs_important` has no other reader):* grepped —
  `contrast_cascade_route.go` writes it (:110,:179,:195,:224); NOTHING in Go reads it; the prompt
  is its only consumer. The "inert signal" claim stands.
- *bug_historian M1 — the REAL one, OPEN and UNOWNED:* the same shape (a filer-computed
  specific instruction co-present with a later generic instruction, adjudicated by the model) may
  live in OTHER agent prompts; nobody has swept `agent_definitions` for it. Honest state: not
  done, not claimed. Transferable pattern for the close-time 016b §9 entry: *"two co-present
  prompt instructions are adjudicated by the model, not by precedence language — fence them so
  only one renders."* The sweep is a separate task for whoever takes it; it is not 390's close
  criterion.

**garden-tools end state, at the artefact:** 14/14 complete; served theme 22,902 B (control 404),
all 13 rules present; post-fence rules carry NO `!important` (8/8), pre-fence 5 + the
unattributed row do. Spot arithmetic: #4a5240 on rgb(238,234,227) = **6.81:1**. One wrinkle,
recorded: `P.tool-description` now has FOUR rules (one pre-fence `!important` #4a5240 + three
post-fence plain) — the pre-fence `!important` wins over all later ones, so the served colour is
#4a5240 (passes). Redundancy class as §(p); no action.
