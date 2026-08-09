# HANDOFF — loancalculator.co.uk · continue here (2026-08-08b, night)

**Supersedes `HANDOFF_2026-08-08_continue_here.md`.** Its §1 (three pages blocked by 219)
is CLOSED. Its §2 (the lane's tools) and §3 (the CSS trap) are still accurate but both
have MOVED ON — read §3 and §4 here before using either.

Read order: this file → `NOTES` 2026-08-08 sections (four of them, newest last) →
`SUMMARY_2026-08-08b` → `fleet_copy_quality/SUMMARY_2026-08-08`.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis   v1.0.1269 (22:02Z) — carries 219's fix
voice     26 of 26 active pages. NOTHING is left in the old voice.
live      26/26 HTTP 200 · toolgolden 11/11 exact · 16 locked rows (12 tool + 4 CSS carrier)
```

## 1. What the owner has ruled, and what is settled

- **The homepage opening is settled and LIVE** — his approved copy, verified on the wire:
  `mathematically rigorous` 0, `true cost of credit` 0, opening appears once.
- **`debt-help-uk` leads with the free charities** — his ordering call, executed. He said
  the instrument was not needed and it was not.
- **`bugs_open/227` (experience-planner writes another site's plan) is FILED and NOT to be
  worked** — "we can fix 227 later".
- **He likes the Breathing Space bridge sentence** the writer invented unprompted.
- Copy still judged good by him: `hidden-loan-fees`. Judged clunky and now fixed:
  `consolidation` ("appeal"), `legal` (the "if" tic).

> **SUPERSEDED 2026-08-09 by `HANDOFF_2026-08-09b_continue_here.md`** (the COPY/VOICE
> thread's cold start — note the **`b`**; the plain `HANDOFF_2026-08-09` is the separate
> `bugs_open/227` platform thread). §4, §5 and §6 below still stand and are still worth
> reading; §3 carries a correction; §2 is closed.
>
> **CLOSED 2026-08-09 (morning).** §2 is **done** — `index`/`prose-2` now reads *"Enter
> your loan amount, rate and term below to see how the monthly figure and the total cost
> move together."*, written through the framework (corr `26648f55`). Struck register 0 on
> the wire across four spellings, owner's opening still appears exactly once, siblings
> byte-identical, 26/26 serving, toolgolden 11/11 exact.
>
> **§3's remedy was followed and it LEAKED anyway** — see the correction inside §3 below.
> Conditional phrasing is necessary but not sufficient; the locks are what held.
>
> Note also: a **`HANDOFF_2026-08-09_continue_here.md`** was committed 14 minutes after
> this file by a concurrent session (the `bugs_open/227` thread). It says the site is
> finished and nothing is owed — it did not know about §2. This file was right. Read both.

## 2. ⚠ THE ONE THING OWED, and it is small and precisely located — **NOW CLOSED, see above**

`index` `prose-2` still reads:

> *"Calculate your exact monthly repayments and see the true total cost of borrowing."*

Unlocked `ported-prose`. **It is there because MY RESTORE put it back** — see §4. The
owner's named paragraph is fixed; this is the same register one line below it, and he has
not seen it yet. One targeted rewrite closes it.

⚠ **Write that guidance to the SECTION, not the page** — see §3, which is the thing most
likely to bite whoever picks this up.

## 3. THE LESSON THAT COST THE MOST TONIGHT — guidance is per-SECTION, always

The writer sees **one section** and never its siblings (`process_sections_loop` has no
accumulator). So **any instruction is applied by every section that can believe it
qualifies.**

Tonight I wrote *"REPLACE the opening block with this approved copy"*. Three separate
sections each concluded they were the opening block, and `prose-0`, `prose-1` and
`prose-2` all came back carrying it. `prose-1` (the Standard Calculator intro) and
`prose-2` were destroyed and had to be restored.

**Write:** *"If this section is the page's opening block, use this copy: … Otherwise leave
your section's subject alone."*
**Never:** *"replace the opening block"*, *"move section 3 to the top"*, *"the page should
open with…"* — page-level language in a per-section prompt.

This is the FOURTH instance of one shape this week: a rule, an exemplar, a pinned
guidance, and now a page-level instruction have each been applied uniformly, because
uniform application is all a section-scoped prompt can do.

> **CORRECTED 2026-08-09 — the prescribed wording was used and it LEAKED, so this remedy
> is NECESSARY BUT NOT SUFFICIENT.** Closing §2 I wrote exactly the recommended shape:
> *"IF THIS SECTION IS the one-sentence introduction sitting directly under the 'Standard
> Loan Calculator' heading … OTHERWISE leave your section's subject, structure and wording
> exactly as they are."* Comparing every section the writer PROPOSED against what was
> stored (`llm_call_log.response_text` vs the backup table): `prose-0` byte-identical,
> `prose-4` byte-identical, **`prose-1` 133 → 400 bytes** — it kept its `<h1>` and appended
> its own version of the very strap line I had assigned to `prose-2`.
>
> **Why:** `prose-1` **IS** the "Standard Loan Calculator" heading, so my condition named
> *its own* landmark and it read the instruction as "you should have an introduction under
> your heading". **A conditional whose CONDITION references a neighbouring landmark is
> ambiguous to that neighbour.** Make the condition decidable from the section's own bytes
> — quoting the sentence to be replaced is what worked on the two distant sections.
>
> **So the actual rule is: phrase it conditionally AND LOCK EVERY SIBLING you are not
> targeting**, leaving exactly one agent-writable row. That is what made this run safe, and
> unlocked, `prose-1` would have shipped the duplicate. Fifth instance of the shape; first
> where the documented remedy was followed and still failed. Method in `RUNBOOK` →
> "Targeted single-section rewrite".

## 4. The CSS trap is now PREVENTED, and it is two mechanisms not one

8 rows carry a `<style>` block and they are **not one population**. Prose characters left
after stripping styles and tags:

```
compare-loans / credit-health-check / overpayment-calculator / consolidation   20–32  PURE CARRIER -> LOCKED
application-tracker                                                             170  MIXED -> must be rewritten
car-finance-calculator / loan-vs-savings / jargon-buster                 1523–2637  MIXED -> must be rewritten
```

- **4 pure carriers LOCKED**, `locked_by='loancalculator_css_carrier_20260808'`. The
  writer cannot have them. Backup `page_components_bak_20260808_csslock` (63 rows).
- **4 mixed rows NOT locked** — locking them would freeze real copy (application-tracker's
  is itself a negation pile that needs rewriting). They are covered by **guidance v3**,
  which orders the `<style>` block reproduced byte-for-byte.

Both halves proven on one run: the locked carrier's `updated_at` stayed at its restore
time while the page rebuilt around it; the mixed row was rewritten AND kept its CSS.

⚠ `voiceh_restore_css_slot.sh` still exists and still works, but it is now a **repair of
last resort**, not the guard. If you find yourself running it, ask why the lock or the
guidance did not hold.

## 5. Guidance lineage — the prompt is a WORK ITEM, not the spec

**This is the single most confusing thing about this lane.** The rewrite prompt is copied
by SQL from a pinned work item. **Editing `site_specs.content_direction` does NOT change
what the writer is told.** Two spec-layer fixes had zero effect before I found this.

```
2517bc4b  canary, owner-reviewed 08-06   mandated a conditional opening   <- CAUSED the "If you're" tic
4a9edd45  v2                             mandate -> prohibition + "vary"  <- BROKE the tic
6d52beaf  v3  (voiceh_rewrite_v3.sh)     + preserve <style> byte-for-byte <- CURRENT, use this
7933edd4  one-off                        debt-help reorder
50c8ba5c  one-off                        index opening
```

Three layers carry the voice — spec `writing_rules`, spec `voice_exemplars`, and the work
item's `spec.suggestion`. **Only the third drives the output.** The other two are still
worth keeping consistent (they were trimmed and fixed on 08-08) so a future reader is not
misled, but do not expect editing them to change anything.

**To fire a rewrite:** `./voiceh_rewrite_v3.sh <page>` or `./voiceh_batch.sh <page>…`.
⚠ **Grade and CLOSE a batch before re-running it** — items are left at `detected` on
purpose, and `detected` is not in `idx_swi_dedup`'s excluded list, so a re-fire fails with
a duplicate-key error until you close them.

## 6. Verification, every time

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 $LANE/toolgolden.py --compare $LANE/acceptance/GOLDEN_2026-08-03b_after_orphan_retired.json \
  https://loancalculator.co.uk/index.html https://loancalculator.co.uk/tools/{overpayment-calculator,consolidation,compare-loans,car-finance-calculator,loan-vs-savings,settlement-calculator,interest-rate-stress-test,credit-health-check,damage-checker,application-tracker}.html
```
Plus 26/26 HTTP 200, and locked rows untouched. **Guard every served fetch with
`wc -c` + a `DOCTYPE` check first** — a deploy-window fetch returns a B2 error blob at
HTTP 200 and every grep against it reads clean.

## 7. Still open elsewhere

- **`bugs_open/227`** — experience-planner. Filed, owner says later.
- **The other finance sites** (`mortgagecalculator`, `lendzy`, `loancash`) are still held
  on the owner's review. **If they proceed, they need guidance v3's lineage, not the
  spec** — and their own incumbent `writing_rules` re-read for conflicts first.
- **The fleet-wide base-prompt change** (owner chose the wide option, 08-05): still not
  started, and the week's evidence has changed its shape — see
  `fleet_copy_quality/SUMMARY_2026-08-08`. The house voice block is SEVEN copies across
  seven agents and they have already drifted.
