# HANDOFF — `bugfix_351_section_template_predicate`, 2026-08-22 (cold start: read this file top to bottom, then §"What to do next")

> ## ✅ LANE COMPLETE, 2026-08-24 — `bugs_open/351` is CLOSED. Read this, then stop.
>
> **Both halves are live and proven at the artefact. There is no outstanding 351 work.** The
> 2026-08-23 block below and the whole body are kept as the record of what was believed on the way;
> nothing in them is a to-do any more.
>
> - **Predicate half** — live on chassis `v1.0.1332` (`0b262ed5e`, ancestry checked with a control).
>   Council **`7b662d65` APPROVED**. **28 rescued / 0 regressed** fleet-wide as of 2026-08-23, after
>   an objection found 12 rows at other component levels the original calibration never covered
>   (6 of them were also being wrongly dropped). Demand-proven: 3 cross-site reuses on `loanzy.uk`.
> - **Birth half** — migration **`581`** applied and ledger-recorded 2026-08-24, register
>   **CLC-029**, council **`f0cd2420` APPROVED** (2 medium advisories acted on: the UPDATE mutation
>   path closed, idempotency fixed). Proven *behaviourally* on the live table, not merely present.
> - **The file moved**: `bugs_closed/351_HANDOFF_2026-08-21_…`. `git log` on the old `bugs_open/`
>   path will not find the closure.
>
> **The two things that are deliberately NOT done, so nobody re-opens them as gaps:**
> 1. The **25 standing NULL rows are not backfilled**, by the ruling of 2026-08-23.
>    ⚠ The reason the bug file *originally* gave for declining it is **spent** — read the ruling
>    section, not the old paragraph, or you will conclude the backfill is now safe. It is not.
> 2. **`usage_count` path-blindness** is a *different* defect, filed as **`bugs_open/378`** with its
>    own evidence and `[UNMEASURED]` list. Not a 351 residual.
>
> ~~**Still owed by someone, unrelated to 351:** §6.3's migration `541` (the `stylesheet_gutted`
> check), which this lane built but never released.~~ **DISCHARGED 2026-08-26 by this lane,
> resumed:** `541` applied and live (commit `6531e694b`; ledger row record-only; both hold
> conditions met with controls — 217/217 live pods self-report the capability, negative control
> 0 — and re-calibrated first with the check's OWN `Run()` over all **31** deployed sites:
> **0 filed / 29 resolved / 2 declined-to-judge**). Details in the migration's discharge header
> and register IMP-055. Nothing from this lane remains owed.

> ## ⚠ CURRENT STATE, 2026-08-23 — read this before anything below it
>
> **Everything under §3, §4 and §6.1–6.2 has moved. The body of this file is left as written**
> (it is the record of what was believed on 08-22); this block is what is true today.
>
> **1. The predicate fix is LIVE, not inert.** §3's "Go is inert until an image is built and rolled"
> and §6.2's "after the next chassis roll" are both spent. Both `agent-chassis` pods report
> `git_commit = f5eaabe33`, which has `97c337371` as an ancestor — verified from
> `service_binary_capabilities` (`kind='build'`) with a control commit that correctly reports NOT an
> ancestor, because the `build provenance` startup line had already scrolled.
>
> **2. It is DEMAND-PROVEN at the artefact — §6.2's bar is met and 351's predicate half is done.**
> On 2026-08-23 `loanzy.uk` bound three library incumbents carrying `section_type IS NULL`, all born
> on a *different* site, with **no `needs_new_component` filed for any of them**: `loans-damage-checker`
> 13:57:41Z, `loans-credit-health-check` 14:07:15Z and 14:23:29Z. Incumbent `824e3309`'s template was
> last written 2026-08-20, so the code moved, not the data.
>
> **3. Re-calibrated 2026-08-23** — section=148, tool=129, calculators=22 (asserted); rescued=22,
> regressed=0; `endsCleanly` flip SET still exactly `{3f946437, 6c41404d}` by id. Corpus moved again
> from 08-22's 150/124, which is why §5's "re-calibrate before shipping" earns its place.
>
> **4. §4's council round has been RESUBMITTED** under the same correlation `7b662d65`
> (`RESUBMIT_CORR`), carrying today's live-and-proven evidence. The council is working again —
> verdicts landed fleet-wide on 08-23 at 12:49, 13:14, 13:42, 17:07 and 17:17Z. Submission JSON:
> `<this session's scratchpad>/council_351_resubmit.json`. **Verdict still unread.**
>
> **5. §5's isolation recipe is superseded.** Do **not** `git archive HEAD | tar -x -C /tmp/chk` —
> `/tmp` is a 16 GB tmpfs and another lane found it at 100% with 12 GB of abandoned checkouts on
> 2026-08-23. Use `scripts/verify-head-builds.sh --with <file> --test <pkg>`, or a scratchpad path on
> disk. See the 2026-08-23 note in `WRONG_CALLS.md`.
>
> **6. TWO CORRECTIONS to counts this file repeats from the bug file.** §3's "**7** diverted twins" is
> **TEN as of 2026-08-23** — the census grew by addition while the sentence stayed true-looking.
> And **`usage_count` is not the reuse signal here**: all 22 incumbents still read `0` while three are
> bound to live pages. Read `page_components`, or you will conclude the opposite of the truth.
>
> **7. ⚠ UPDATED 2026-08-24 — the `section_type` question is DECIDED and the birth door is WRITTEN.**
> The backfill is **DECLINED** (ruling recorded in `bugs_open/351`, "incumbents stay Path-1-only"):
> it adds no reachability any live caller uses, and it harms two — `load_existing_component`'s
> primary query deliberately relies on the NULL miss, and the selector's `ORDER BY score DESC` has no
> secondary key. ⚠ **The reason the bug file originally gave for declining it is SPENT** (the guard
> no longer drops them), so do not re-derive "the backfill is now safe" from it.
> Instead: **migration `581_refuse_selector_invisible_section_birth.sql` + register **CLC-029**,
> committed `a99049669`, `Council-Submitted: f0cd2420-8687-4d6d-80cd-6627dc57788d`, and
> **DELIBERATELY NOT APPLIED** — a migration is live the instant it applies with no image to roll
> back, so it waits for the verdict. Tested against the live DB with `COMMIT`→`ROLLBACK` (verify
> passed, nothing left behind) and mutation-proven both ways.
> ⚠ Two traps if you touch it: `forked_from IS NULL` is **load-bearing** (`deploy_tool`'s fork INSERT
> omits `section_type`, so a fork is legitimately born NULL — widening it breaks tool deployment at
> runtime), and it is **INSERT-only** (a CHECK, even `NOT VALID`, is enforced on UPDATE and would
> break template repairs to the 25 standing rows).
> **Still genuinely open:** apply `581` once the verdict is read; §6.3's migration `541`; and the
> `usage_count` path-blindness found on 08-23 (incremented only on Path 2, read as a Path-2 scoring
> input; 96 of 149 read 0 despite live bindings, 1,802 bindings invisible) — **not filed as a bug
> yet**, and it is a separate defect from 351.


**This lane began as "the CSS is broken on remortgagecalculator.uk" and ended up owning a platform
predicate.** Both halves of the owner's original complaint are FIXED AND LIVE. What is left is one
committed-but-unrolled platform change, one unread council verdict, and one deliberately-declined
decision.

**One-line status:** `bugs_open/351`'s predicate fix is **committed (`97c337371`) and INERT until
the next chassis roll**; its council round **died with no verdict** and must be re-run; the
`section_type` half is deliberately NOT done.

---

## 1. What the owner asked, and where it landed

| ask | state |
|---|---|
| "the CSS is broken, find why and fix it" | **DONE.** css-patch-agent had clobbered the stylesheet 17,403 → 136 B. Restored; serves 17,403 B with `:root` ×3. |
| "run a fresh rebuild through the same mechanism to check" | **DONE.** Assemble-mode rerender deployed `index.html` only; stylesheet untouched. |
| "do we have a checker/handler that would spot this?" | **NO — built one.** `stylesheet_gutted`, register **IMP-055**, council APPROVED (`d3187418`). Migration `541` is **HELD** for the roll. |
| "the site is still missing its tools — file or point at a bug" | **Pointed, not filed twice.** It was `bugs_open/345`, and `bugs_open/311`'s residual, which became **`bugs_open/351`**. |
| "implement 351 here" | **DONE** (predicate half). See §3. |

**The site itself is healthy:** index serves **69,421 B with 6 `<input>`** (the calculator arrived
via the 311 lane's re-drive at 18:19Z on 08-21, on an **attempt-0** success), stylesheet 17,403 B.

---

## 2. THE ONE THING MOST LIKELY TO MISLEAD YOU

**`git log` on `bugs_open/311` will not find the closure, and `git log` on 345's Go half will not
find the code.** 311 was `git mv`'d to `bugs_closed/` (commit `6e2d21a70`); 345's Go half rode into
another session's commit (`0f80f5ea1`, message says `bugs_open/344`) as a same-file passenger. Both
are documented in their own files. Search by content, not by path or bug number.

---

## 3. `bugs_open/351` — what shipped, and what has NOT

**Committed `97c337371`. Go is inert until an image is built and rolled.**

- `endsCleanly` (`component_write_guard.go`) strips trailing `{{end}}` actions then requires `>`.
- `sectionTemplateValid` (`plan_sections_action.go`) is now structural
  (`UnbalancedStructuralTags` + `endsCleanly`) instead of a `</section>` substring test.

**Why the repair is at `endsCleanly` and not at the section predicate:** it has **4** callers as of
2026-08-22, and the fourth (`component_write_guard.go:260`) is a **write-time** regression check —
so the same false positive was refusing legitimate work at BIRTH. Fixing only the predicate would
leave the write guard refusing the shape the loader newly accepts.

**Live calibration, 2026-08-22:** `read: section=150 tool=124` (asserted), `rescued=22`,
`regressed=0`, `endsCleanly flips = 2` (`3f946437`, `6c41404d`), both hand-checked as complete
conditional wrappers.

### ⚠ NOT DONE, deliberately — the `section_type` half

**22** calculators (as of 2026-08-21) still carry `section_type = NULL`. **Do not backfill without
deciding the ordering question first:** after the predicate fix, Path-1's function match can surface
the incumbents while the selector's `section_type` match surfaces their **diverted twins** (seven
exist). Both would match after a backfill, and which wins depends on path order. `bugs_open/351`
§"What this does NOT need" states it: decide it in the migration or state explicitly that incumbents
stay Path-1-only. **Silence is the only wrong answer.**

---

## 4. Council `7b662d65` — NO VERDICT, must be re-run

The round completed at `complete_invalid` in **9 seconds** with **zero** `council_report` artifacts
and no `doc_notes` row. Per LANDMINES: *"A council run killed by the account's API usage limit
completes at `complete_invalid` and writes NO verdict row."* Corroborating: no council report of any
kind appeared fleet-wide after 18:15:54Z, and mine started 18:33:23Z.

**`complete_invalid` here is NOT a rejection.** The `Council-Submitted:` trailer on `97c337371`
remains honest because it asserts nothing.

**To do:** re-submit when capacity returns. The submission JSON is at
`<scratchpad>/council_351.json`; re-run with
`097_TRIGGER_council_review_v1.sh <file>` and use `RESUBMIT_CORR=7b662d65-…` so the trail
accumulates.

---

## 5. Verification recipes (each cost something to get right)

**Re-calibrate before shipping — and assert the SET, never the count.**
Export per level with a sentinel (a `COPY` tab gets escaped and yields a VACUOUS zero — it happened):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAX -c "
SELECT '@@@BEGIN:' || id::text || E'\n' || html_template || E'\n@@@END'
FROM content_components WHERE is_active AND component_level='section'
  AND html_template IS NOT NULL AND html_template <> '' ORDER BY id;" > /tmp/live_section.txt
grep -c '^@@@BEGIN:' /tmp/live_section.txt     # assert this, or a zero means nothing
```

Then a throwaway `zz_tmp_*_test.go` INSIDE `platform/orchestration/actions` (the predicates are
unexported), run it, and **delete it in the same command** — the tree is shared.

- **Assert the flip SET by id, not `count==1`.** It earned itself immediately: one row left the
  rescued set (another lane fixed it at 14:14Z) and a new one joined (11:51Z) — a count read 22 both
  times and would have hidden the substitution.
- **`go test` in the working tree is NOT a clean signal.** 18 other sessions' `.go` files were dirty
  on 2026-08-22. Isolate: `scripts/verify-head-builds.sh --with <your file> --test <pkg>`.
  ⚠ `TestUpdateWorkItemStatus_RecordsRoutedStepError` fails in the tree from another lane's
  uncommitted `work_item_failure_ladder.go`. **It is not ours. Do not "fix" it.**
- **Mutation-prove every guard.** Five mutations, each caught by a named test — and note the
  restore step can be killed by a command timeout: after any mutation run, `diff` the file against
  its backup before believing the tree is clean. That happened here.

**Artefact checks:**
```bash
curl -s "https://remortgagecalculator.uk/?cb=$RANDOM" | grep -oc '<input'   # expect 6
curl -s "https://remortgagecalculator.uk/assets/css/styles.css?cb=$RANDOM" | wc -c   # expect 17403
```

---

## 6. What to do next (ranked)

1. **Re-run the 351 council round** (§4). Cheap, and the code is already on the shared branch, so a
   REVISE must be acted on.
2. **After the next chassis roll: prove 351 at the artefact.** A green test is not the bar. The
   signal is a site planning a calculator section that RESOLVES to a library incumbent with **no
   `needs_new_component` item filed at all**. Until then 351 stays OPEN.
3. **Release migration `541`** (the `stylesheet_gutted` check) once a rolled image carries
   `check_stylesheet_gutted.go` — probe the capability on every pod **with a negative control**, and
   add the name to `liveConfiguredChecks` in the SAME commit. An unregistered name fails the whole
   discovery step and discards earlier checks' findings.
4. **Decide the `section_type` question** (§3) — deliberately, in writing.
5. Nothing else here is owed. `198`'s remaining candidates (deploy-side shrink guard, birth guard,
   round-trip writer inventory) are the **bugfix 198** lane's; `345`'s third half (add `last_error?`
   to the dispatcher's `input_mapping`) and `337` are the **311 continued** lane's.

---

## 7. Cross-lane state — who owns what, as of 2026-08-22

- **`311 continued`** owns `bugs_open/345`, `337`, and closed `311`. Correspondence with them is
  recorded in both bug files; they reviewed 351's plan, caught a call site I missed, corrected my
  proposed rule, and assigned the implementation here.
- **`bugfix 198`** owns the css-patch clobber class. They shipped the prevention half (542/543/546,
  DGH-016) and seeded webdesign.uk (548) — the fleet is **22 PASS / 0 REFUSE** as of 2026-08-22.
  Their guard and our `stylesheet_gutted` check cover each other's blind spots; the note saying so
  is in `bugs_open/198` and should not be retired by someone tidying.
- **`bugs_open/309`** owns the field-source vocabulary guard that refused the original template.

## 8. My own wrong calls this lane, recorded because they will recur

All three are in `WRONG_CALLS.md` with the cheap check:

1. **Calibrated a detector with a PROXY** for its own predicate — claimed it would file on 1 site;
   the real predicate filed on **19 of 25**. Corrected before enabling.
2. **Read a SERVED-side absence as an ARTEFACT-side absence** — cleared webdesign.uk as "not damage"
   off a redirect page while its 15.5 KB stylesheet sat one finding from destruction. Caught by the
   198 lane. Class remedy: *stop asking the URL what the artefact is.*
3. **Verified both ends of a pipe and inferred the middle** — reported 345's fix live because the
   writer set the key and the prompt read it; a dispatcher allow-list dropped it in between, so the
   fix was INERT. I had the disconfirming zero and explained it away.
