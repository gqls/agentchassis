# HANDOFF — LMC lane entry point, 2026-08-14. START HERE.

**Supersedes `HANDOFF_2026-08-11b_track_b_authorised_read_253_first.md`** (which is
now five addenda deep — its content is folded in here; read it only for history).
Parent context: `HANDOFF_2026-08-11_after_track_a_decisions_pending.md` (Track A) and
`HANDOFF_2026-08-10c_continue_here.md` §2b/§6/§7 (still-valid Track B background).

---

## 0. State in one paragraph

Track A (17 prose pages): **done and stable**. Track B (22 calculator pages): **18
converted by the trackb2 session under a NEW architecture, 5 remain** — held pending
a required ruling (§3). One live arithmetic regression from that conversion was
found and **repaired 2026-08-14** (§2). The per-slot floors (text + component) are
**live on both write paths and council-APPROVED**. The whole arithmetic estate reads
**oracle 176 PASS / 0 FAIL / 0 CONVENTION**, verified today.

> **CORRECTED 2026-08-14 (afternoon): the standing oracle read is 170 / 0 / 6, not
> 176/0/0.** Six minutes after §2's hand repair, the trackb2 session rebuilt the
> same page from the clean sites-repo pin (`7e6b993ef`), restoring the ORIGINAL
> 224-fix wiring — billed-convention totals (payment rounded to the penny first).
> The six CONVENTION lines are all standard-calc, all healthy, and match the
> pre-regression estate's profile. **Do not "repair" them.** Full account: NOTES
> 2026-08-14 (afternoon) entry.

## 1. Verified on v1.0.1298 (pods `64cb9c4bb9-*`, up 08:58Z, 2026-08-14)

| check | result |
|---|---|
| floors in the binary, both replicas | `SLOT FLOOR` 1 · `COMPONENT FLOOR` 1 · `SHRINK` 1 · neg control 0 |
| `189`/`204` | `1 / 1` |
| **post-roll mirror check** (assemble `legal`, diff vs `predicted/`) | **byte-identical** |
| oracle, whole estate, controls in-session | **176 / 0 / 0** (parse OK; mutation 4 FAIL / 0 passed) — **superseded ~08:14Z: now 170 / 0 / 6, healthy (§0 correction)** |
| standard-calc on the wire | loads `calculators.js`, stale `r > 0` guard gone |

**Do the mirror check after every roll** (one rerender + one diff — §5 commands).
A `predicted/` file is only valid until the framework next writes that page.

## 2. The 224 regression on standard-calc — REPAIRED, and the pattern to watch

**What happened.** The trackb2 re-architecture (machinery in
`content_components.html_template`, copy as `input_schema` fields, tool rows
UNLOCKED) moved the FIXED arithmetic engines into `/assets/js/calculators.js` — but
`loans-standard-calc`'s template kept its own **pre-224 inline script**
(`if (P > 0 && r > 0 …)`, the stale-answer guard). The page served wrong answers at
0% APR for ~20 hours and its rerender committed the bad bytes into the sites repo.
Found by a routine oracle re-run (164/6), attributed to one tool, mechanism measured.

**The repair (owner-directed, 2026-08-14).** Template AND rendered row, one
`DO`/`RAISE`-verified transaction: stale inline replaced by the `calculators.js` tag
plus thin wiring that calls `calculateAmortization` and **always writes the DOM**
(blank inputs coerce to 0 → engine returns zeros), so the two-routes stale-answer
bug is structurally unrepresentable there. Copy placeholders untouched; row left
**unlocked** per the owner's ruling. Oracle: the tool 15/15, estate 176/0/0.

> **CORRECTED 2026-08-14 (afternoon): this repair was superseded six minutes after
> it landed.** At 08:11–08:12Z the trackb2 session rebuilt `loans-standard-calc`
> (and `mortgages-overpayment`) from clean pin `7e6b993ef` and rerendered — the
> served inline script is byte-identical to that pin's slice, i.e. the original
> 08-10 billed-convention 224 fix, restored through the framework. Still 224-safe
> (every path writes the DOM, explicit £0.00 in the guard branch). The morning's
> "still serving the repaired page" check used markers both versions carry, so it
> missed the swap; see `WRONG_CALLS.md` 2026-08-14 (lmc) and the NOTES afternoon
> entry.

**The pattern that caused it is still present**: four more loans templates carry
their own inline arithmetic (`settlement-calculator` and siblings — passing today,
but each is a second copy of arithmetic that must now be maintained twice, which is
exactly how this regression happened). Flagged to the owning session; moving them
onto the engine is a small well-shaped batch **if the owner asks**.

## 3. ⛔ THE REQUIRED RULING — blocking the last 5 conversions

Owner (2026-08-14): calculators stay **decomposed and editable — no locks that
block editing**. The demand on the trackb2/owning session, written in full in
`HANDOFF_2026-08-11b…` final addendum, condensed:

1. **Name what protects the MACHINERY.** Floors cover `save_page_sections` and
   `ApplySectionEditAction`; **nothing guards `content_components.html_template`**
   — and the incident above is the proof that a template can carry stale arithmetic.
2. **State the seam rule**: templates carry wiring only, engines carry arithmetic
   (the repaired `loans-standard-calc` template is the worked example) — or the
   stated alternative.
3. **Supersede the "tool row born locked" language** in every Track B brief,
   visibly, dated.
4. **Explain `page_rerender_…trackb2-b1fix`** (08-13 16:37, complete) — it
   completed and the page stayed wrong.

The last 5 verbatim pages (`rate-forecaster`, `equity-release`, `bridging-loan`,
`fee-analyser`, `damage-checker`) **wait for this ruling**. Not yet answered as of
14 Aug ~10:00Z — check `git log` on this lane before assuming.

## 4. The floors (bugs_open/253 framework_rewrite slug) — done, live, approved

- Component floor (whole-page) + both floors on the section editor: **live since
  v1.0.1295/1297**, verified again on 1298. Council **APPROVED round 3** (10/11,
  correlation `b30ac52c`; commits carry `Council-Submitted:` and `098` credits them).
- The **coverage test** (`page_component_writer_coverage_test.go`) is the class fix:
  every file that `UPDATE`s `page_components.rendered_html` must enforce a floor or
  sit in `exemptWriters` with a reason. It exists because an induction failed —
  unwiring the guard broke nothing — and it caught a writer the manual audit
  misclassified. Its stated weakness: proves wiring EXISTS, not that it EXECUTES.
- Production audit (08-14): 0 refusals / 77 writes, and 0 in-scope flattenings
  slipped through (archived-prior comparison via `page_component_history`,
  `op='overwrite'`, join on `component_id`).
- **⚠ `253` is an ambiguous number** — two unrelated bugs. Use the slug.

## 5. Commands (all exercised today)

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk

# post-roll: guards in the binary (positive + negative controls)
P=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $P -- sh -c \
 'strings /app/agent-chassis | grep -c "SLOT FLOOR REFUSED"; strings /app/agent-chassis | grep -c "zzz_cannot_exist"'

# post-roll: mirror check (file an assemble-only page_rerender on legal, then)
diff <(curl -s -A Mozilla/5.0 https://loanandmortgagecalculator.co.uk/legal.html) \
     <DECOMP_WORK>/predicted/legal.html

# arithmetic, controls FIRST, same session — a green run without them is not evidence
cd $LANE && python3 oracle.py --selftest-parse && python3 oracle.py --mutate expectation --tools simple
cd $LANE && python3 oracle.py            # expect 170/0/6 (six CONV = standard-calc billed convention — healthy; CORRECTED 08-14 afternoon)

# floor refusals, fleet (0 is ambiguous — pair it with the archived-prior audit)
# site_work_items WHERE spec->>'reason' LIKE '%FLOOR REFUSED%'
```

## 6. Everything else open, with owners

| item | state | owner |
|---|---|---|
| the ruling (§3) | **awaited** | trackb2 session |
| last 5 Track B pages | held on §3 | trackb2 session |
| 4 loans templates with inline arithmetic | flagged, passing | trackb2 (batch move if owner asks) |
| `bugs_open/251` canonical → then `252` og:/lang | scheduled, ordered | unowned — platform work |
| `bugs_open/250` last box: loancalculator restore round-trip | open | loancalculator lane |
| colour-fixer exemptions REASONED→MEASURED | **DONE 08-14 pm** — `colour_fixer_class_preservation_test.go`, commit `bb894e312`, council **APPROVED round 1** (corr `d4b08e11`) | done |
| loancash `complaint-deadline-calculator` oracle | workstream exists (`loancash_couk_fca_validation/`) | unstarted — note the trackb2 session's transcript mentioned starting it; check before duplicating |
| Track C (loancash decomposition) | after Track B | — |

## 7. Traps this lane has paid for (short list; LANDMINES has the full entries)

- **A stale template is invisible to every floor** — floors compare row-vs-row;
  arithmetic wrongness needs the ORACLE, and the oracle needs its controls run in
  the same session.
- **A rerender COMMITS to the sites repo** — a bad page poisons `origin/master`,
  and byte-checks then validate the wrong bytes. Post-224 reference: `e69b5b275`.
- **`decompose_lmc.py` refuses a stale pin** (would have reverted 16 calculators);
  `ALLOW_STALE_PIN=1` overrides loudly. Re-point to a concrete SHA, never a branch.
- **Zero refusals from a guard is ambiguous** — pair it with the archived-prior
  audit before calling it healthy.
- **The actions package may be red from another session's work** — check WHICH
  tests before reading a red as yours.
- Number collisions: `253` (two bugs), `146`, `083`… — **resolve by slug**.
