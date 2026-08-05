# HANDOFF — loancalculator.co.uk · continue here (2026-08-05)

**Supersedes `HANDOFF_2026-08-04_continue_here.md`.** That file is still accurate for
everything before today and is NOT repeated: read its §3 (the offline route + the ⚠
`bugs_open/189` correction), §4 (verification assets), §5 (what was learned the hard
way) and §6/§6b (post-roll method). This file carries today's change of direction.

Read order for a cold start: this file → `HANDOFF_2026-08-04` §3/§5/§6b → `NOTES` tail
(the 2026-08-05 section) → `portfolio_positioning/VOICE_gentle_explanatory_v1.md`.

---

## 1. What changed today, in one paragraph

The owner ruled that this site's copy must be **rerun through the FRAMEWORK** in the
new "gentle explanatory" (H) voice — **explicitly not hand-written through this CLI**
(standing ruling 2026-08-04). Scope: **copy only, calculators kept.** The voice is
seeded and live. Two blockers were found underneath it; the first is fixed, **the
second is open and is a platform fix.** No copy has been rewritten yet. Nothing is
broken: the site is 26/26 and unchanged.

```
site            loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis         v1.0.1254, rolled 2026-08-05T20:41Z
live            26/26 HTTP 200 · old-footer 0 · no-canonical 0 · empty-desc 0
calculators     12 locked rows, untouched today
voice           SEEDED + LIVE in site_specs.content_direction (has_H=true in .formatted)
copy rewritten  NONE — blocked, see §3
```

## 2. ⛔ Do not touch loanandmortgagecalculator.co.uk

Session `fffe0948` (the `portfolio_positioning` lane) owns it and seeded the same
voice there at **2026-08-05T10:54:58Z**, live at the time of writing. Its prompt is
`portfolio_positioning/VOICE_gentle_explanatory_v1.md` — **read that, reuse it, do not
fork it.** Two sessions writing one `site_specs` row is how one silently wins.

## 3. THE OPEN BLOCKER — `plan_sections` cannot resolve a positional slot name

**This is `bugs_closed/182` again, in the sibling call site, and it is the whole of
what stands between this lane and the owner's instruction.**

`pages.sections` here is `["prose-0","prose-1"]` — positional slot names.
`plan_sections` resolves those against component **name/function**:
`loadComponentSchemas` (`plan_sections_action.go:1144`) indexes "by both name and
function"; `:918` does `components[sectionName]`. `prose-0` is neither, so it misses,
falls through to the selector at `:937`, and every section defers.

**Induced end to end, not argued** — one real `content_rewrite` at
`guide-how-loans-are-calculated` returned:

```
page-build-handler no-op: no sections ready to build … the target section was NOT rebuilt
```

**Measured: 0 of 57 section names on this site resolve.** Fleet: 86 unresolvable
across 5 sites (loancalculator 57/57, gaswholesalers 11, finetuning 10,
leopardessconsulting 6, oufe 2). Only this site is 100% blocked.

> **`a43be1e70` — 182's own fix — EDITED THIS FILE** to factor `componentInfoFromRaw`
> across "the three now-shared conversion sites", and still added `component_id`-first
> resolution only to the re-render path. Re-checked at v1.0.1254: `plan_sections_action.go`
> untouched since 08-04. **Open, unowned, and the precedent for the fix is already
> APPROVED** — do the same thing 182 did, one function over.

⚠ **`bugs_closed/041` will look like this bug and is CLOSED — it is not the same.**
041's cause was the RAW string (`call_to_action` missing the existing
`call-to-action`) and its fix was normalisation, live v1.0.1146. That does not reach
204: `prose-0` normalised is still `prose-0`, and no component bears that name or
function under any spelling. 041 closed the *spelling* half of this lookup's
blindness; 204 is the *identity* half. Family members: 039, 041, 095, 204.

**The fix:** resolve by `page_components.component_id` first, fall back to
name/function. 182's `loadComponentSchemasByID` / `loadContentComponentsByID` already
exist and are the reusable unit — **do not write a third resolver.** ⚠ `plan_sections`
has no `pageID` at that point in the workflow (its own comment says so, `:1140`), so
the first real design question is where the page id comes from. That is the thing to
work out before writing code.

**Platform Go ⇒ council gate applies.** Submit before or alongside the commit; use
`Council-Submitted: <corr>` if the verdict has not landed.

⚠ **It does not merely fail — it asks the fleet to manufacture junk.** My canary made
the selector file `needs_new_component` items to build components literally named
`prose-0` and `prose-1`, plus two `needs_section_data` HITL items. I cancelled all
four with an explanatory note. **After any build-path attempt on a decomposed site,
check for and cancel these:**
```sql
SELECT item_type, status, summary FROM site_work_items
WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND item_type IN ('needs_new_component','needs_section_data')
  AND status NOT IN ('complete','cancelled') ORDER BY created_at DESC;
```

### ⚠ The sibling lane has a CLI voice applier. It does NOT satisfy this lane's instruction.

At 21:54 today session `fffe0948` committed `35c1e11e9` — a decomposer, a guarded row
writer and **`voice_apply.py`** in the `loanandmortgagecalculator_couk` lane, and
re-voiced a canary guide with it. That is their lane and their call, and their toolkit
may well be worth reading (their decomposer solves the same shape of problem).

**But do not reach for it here as the shortcut past §3.** The owner's instruction for
*this* site was explicit: rerun the copy **through the framework**, *"don't build it
through this cli"*. A CLI applier writing `page_components` rows directly is precisely
the thing that instruction rules out, however good the output looks — the whole point
is that framework-written copy passes the framework's checks and CLI-written copy does
not. If the two lanes should converge on one route, that is an owner decision, not a
convenience. **Fix 204 instead.**

## 4. What is DONE and live (do not redo)

**a. The `authored` mislabel is corrected.** `ported-prose.content.source` was
`"authored"` — which asserts *a human supplied this, do not regenerate*. **Owner
ruling 2026-08-05: false.** It was another LLM's output, written outside the
framework's checks, lifted byte-for-byte by the decomposer. Now `"llm"`, with
`llm_guidance` rewritten to be writer-facing and to carry the decomposer's invariant
(prose only; no form control, id or script; *"rewriting this cannot break a
calculator"*; preserve every fact, figure and link — voice only).

Measured before changing: `ported-prose` is used by **this site and no other** (51
rows); `authored` fields fleet-wide **2 → 1**, the negative control that exactly one
moved. ⚠ **The survivor is `ported-page.body` and it MUST stay `authored`** — that is
the `--fidelity locked` byte-preserving adoption path (ADO-037). Flipping it would let
a writer rewrite a site adopted precisely to be preserved.

**b. The H voice is seeded and live.** `seed_voice_h.py` in this lane.
`writing_rules` 15 → 23, `formatted` 20,699 → 24,556 b, `has_H=true` on the stored row.
Two properties worth preserving if you touch it:

- **The formatter gate.** The writer reads exactly ONE field —
  `{{.site_specs.specs.content_direction.formatted}}`, live in
  `page-content-writer.prompt_template`. Everything else reaches the prompt only by
  being serialised into it by `datahelpers.FormatContentDirection`. **A hand-written
  `content_direction` that does not regenerate `formatted` is invisible to the writer
  and looks applied.** The script ports that function and refuses to write unless its
  port reproduces the STORED value exactly (PASSED at 20,699 b / 141 lines).
- **Conflicts are REPLACED, not appended.** The incumbent spec said *"Avoid
  contractions in declarative or authoritative statements"* against H's *contractions
  wherever they would be spoken*. Stacking them lets the model choose. 3 replaced, 8
  added; the script **refuses to run** if an expected incumbent is missing.

**c. Backups:** `content_components_bak_20260805_prosesource` (2),
`page_components_bak_20260805_framework_rewrite` (63).

## 5. The other finance sites — HELD, deliberately

Owner decision 2026-08-05: **wait until he has reviewed loanandmortgagecalculator's
copy in the new voice.** Then seed **mortgagecalculator.co.uk, lendzy.co.uk,
loancash.co.uk** (loancalculator is done; loanandmortgagecalculator is session
`fffe0948`'s). Reuse `seed_voice_h.py` — but **per site, with that site's OWN
exemplars**, never by editing a shared prompt to hardcode one site's voice
(VOICE doc step 2). Each site's incumbent `writing_rules` must be re-read for
conflicts first; loancalculator had three and another site's will differ.

## 6. The fleet-wide voice change — owner chose the WIDE option

Owner chose folding H into the writer's base prompt (every future site, every
vertical) over the contained finance-pool option, 2026-08-05. Two findings shape it:

**a. "The base prompt" is SEVEN prompts, already drifted.** No two identical:

```
content-writer                        1046 b   563e678a…
content-creator-hero-without-research 1243 b   224a2008…
content-creator-about                 1250 b   8f117bcc…
content-creator-hero                  1252 b   ea4736ea…
simple-content-writer-with-approval   2334 b   cba1f868…
grounded-explainer                    2866 b   89221f73…
page-content-writer                   4657 b   d4b409e1…
```

So it is seven edits that will drift again, **or** one shared carrier read at
prompt-assembly time (local precedent: the `footer_compliance_lines` carrier the
portfolio lane built). Put both in the submission; do not pick silently.

**b. There is a real semantic conflict, and it is the thing the architecture seat
will ask about.** All seven carry **"Start with the fact. Never open by saying what
something is NOT"**. H rule 1 says *open where the reader is standing… before the
first assertion*. They **agree** on banning the negative-twist opener, so the
reconciliation is to revise the opening rule and keep the ban — but it is a
**revision of a live fleet-wide default**, not an addition. Say so in the submission.

This is architecture-scope: council gate, concept register entry **in the same commit
that ships it**, and tell the other consumers rather than only measuring them
(owner ruling 2026-07-29 §3).

## 7. Verification OWED

- **The v1.0.1254 post-roll byte check.** Today's roll (20:41Z) has been checked for
  *health* only — 26/26 200 and the three content checks clean. Nothing has
  re-rendered since, so served bytes are necessarily unchanged; that is **not** the
  same as proving the new image renders them identically. Do §6b of the 08-04
  handoff: re-render three pages, compare md5 to a pre-captured baseline.
  ✅ **The site is fully settled, so any page may be a baseline** — no pending change
  to contaminate it. Guard the fetch with `wc -c` + `DOCTYPE` first (a deploy-window
  fetch returns a B2 error blob at HTTP 200 and every grep against it reads clean).
- **`toolgolden --compare` against `GOLDEN_2026-08-03b`** — not re-run since the roll.

## 8. Commits (2026-08-05)

`34f169247` v1.0.1251 post-roll + the 182/189 handoff correction ·
`9a01c67f0` the canonical finding + 23-page propagation ·
`65ee614c4` landmine + wrong-call (sizing a re-render by your own change) ·
`38c46a4b0` wrong-call (the deploy-window error blob) ·
`ea63ef105` item (2) closed, 23 of 23 ·
`d8ed358f1` the voice seed + why the framework could not rewrite this prose

No `Council-Reviewed:` trailers and none owed — every file is under `docs/`, which the
gate refuses client-side. **The `plan_sections` fix in §3 WILL need one.**
